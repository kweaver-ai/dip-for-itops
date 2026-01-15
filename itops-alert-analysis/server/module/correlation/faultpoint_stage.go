package correlation

import (
	"context"
	"fmt"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/idgen"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/slice"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/timex"
	"github.com/pkg/errors"
)

// FaultPointStage 按 failure_mode 收敛事件。
type FaultPointStage struct {
	cfgManager     *config.ConfigManager
	repoFactory    *opensearch.RepositoryFactory
	problemHandler core.ProblemHandler
	genID          *idgen.Generator
}

func NewFaultPointStage(cfgManager *config.ConfigManager, repoFactory *opensearch.RepositoryFactory, problemHandler core.ProblemHandler) *FaultPointStage {
	return &FaultPointStage{
		cfgManager:     cfgManager,
		repoFactory:    repoFactory,
		problemHandler: problemHandler,
		genID:          idgen.New(),
	}
}

// HandleEvent 执行收敛：按 failure_mode
func (s *FaultPointStage) HandleEvent(ctx context.Context, event domain.RawEvent) error {
	// 恢复事件,单独处理
	if event.EventStatus == domain.EventStatusRecovered {
		return s.handleRecoveryEvents(ctx, event)
	}
	//正常的发生事件
	return s.handleAlertEvent(ctx, event)
}

// handleAlertEvent 处理告警事件（EventStatus != 2）
func (s *FaultPointStage) handleAlertEvent(ctx context.Context, event domain.RawEvent) error {
	failureMode := event.EventType // FaultMode 就是 EventType

	expirationTime := time.Now().Add(-s.cfgManager.GetConfig().AppConfig.FaultPoint.Expiration.ExpirationTime)
	log.Debugf("执行存量故障点检查最早故障时间: %s", expirationTime.Format(time.DateTime))

	existed, err := s.repoFactory.FaultPoints().FindOpenByEntityAndMode(ctx, event.EntityObjectID, failureMode, expirationTime)
	if err != nil {
		return errors.Wrap(err, "查询故障点失败")
	}

	latestTimestamp := event.EventTimestamp

	if existed != nil {
		log.Infof("当前存在收敛故障点对象:%d", existed.FaultID)
		// 命中未关闭的故障点，更新最新时间与关联事件列表。
		//if latestTimestamp.After(existed.FaultLatestTime) {
		//	existed.FaultLatestTime = latestTimestamp
		//}

		if event.EventOccurTime != nil && event.EventOccurTime.After(existed.FaultOccurTime) {
			existed.FaultLatestTime = *event.EventOccurTime
		}

		existed.RelationEventIDs = slice.AppendUniqueUint64(existed.RelationEventIDs, event.EventID)
		existed.FaultUpdateTime = timex.NowLocalTime()

		if !existed.FaultOccurTime.IsZero() {
			existed.FaultDurationTime = int64(existed.FaultLatestTime.Sub(existed.FaultOccurTime).Seconds())
		}
		// 值越小等级越高。如果存在的事件等级低于将要合并的事件等级将使用最高等级的事件
		if event.EventLevel < existed.FaultLevel {
			existed.FaultLevel = event.EventLevel
		}

		if err := s.repoFactory.FaultPoints().Upsert(ctx, *existed); err != nil {
			return errors.Wrap(err, "更新故障点失败")
		}
		if err := s.linkFaultToEvents(ctx, existed.FaultID, []uint64{event.EventID}); err != nil {
			return err
		}

		// 将故障点传递给 Problem 阶段进行关联和合并
		return s.problemHandler.HandleFaultPoint(ctx, *existed)
	}

	// 2. 未命中则创建新的故障点。
	occurTime := latestTimestamp
	if event.EventOccurTime != nil {
		occurTime = *event.EventOccurTime
	}

	fp := domain.FaultPointObject{
		FaultID:           s.genID.NextID(),
		FaultName:         event.EventTitle,
		FaultCreateTime:   timex.NowLocalTime(),
		FaultUpdateTime:   timex.NowLocalTime(),
		FaultStatus:       domain.FaultStatusOccurred,
		FaultOccurTime:    occurTime,
		FaultLatestTime:   latestTimestamp,
		FaultDurationTime: 0,
		EntityObjectClass: event.EntityObjectClass,
		EntityObjectName:  event.EntityObjectName,
		EntityObjectID:    event.EntityObjectID,
		RelationEventIDs:  []uint64{event.EventID},
		FaultMode:         failureMode,
		FaultLevel:        event.EventLevel,
		FaultDescription:  event.EventContent,
	}

	if err := s.repoFactory.FaultPoints().Upsert(ctx, fp); err != nil {
		return errors.Wrap(err, "创建故障点失败")
	}

	// 写入故障点关系
	if err := s.writeFaultPointRelation(ctx, fp, event); err != nil {
		return errors.Wrap(err, "写入故障点关系失败")
	}

	if err := s.linkFaultToEvents(ctx, fp.FaultID, []uint64{event.EventID}); err != nil {
		return err
	}

	// 将新创建的故障点传递给 Problem 阶段进行关联和合并
	return s.problemHandler.HandleFaultPoint(ctx, fp)
}

// handleRecoveryEvents 处理恢复事件（EventStatus == 2）
// 恢复事件已写入 raw_event_index，有新的 event_id
func (s *FaultPointStage) handleRecoveryEvents(ctx context.Context, event domain.RawEvent) error {
	log.Infof("处理恢复事件: event_id=%d, RecoveryId=%d", event.EventID, event.RecoveryId)

	//event_provider_id 找到对应的告警事件
	providerID := fmt.Sprintf("%d", event.RecoveryId)
	rawEvents, err := s.repoFactory.RawEvents().QueryByProviderID(ctx, []string{providerID})
	if err != nil {
		return errors.Wrap(err, "查询告警事件失败")
	}
	if len(rawEvents) == 0 {
		log.Infof("恢复事件未找到对应的告警事件，忽略: event_provider_id=%s", providerID)
		return nil
	}

	for _, rawEvent := range rawEvents {
		log.Infof("恢复事件匹配到告警事件: event_id=%d", rawEvent.EventID)

		//通过告警事件的 event_id 找到关联的故障点
		faultPoint, err := s.repoFactory.FaultPoints().FindByEventID(ctx, rawEvent.EventID)
		if err != nil {
			return errors.Wrap(err, "查询故障点失败")
		}
		if faultPoint == nil {
			log.Infof("告警事件 %d 未关联到故障点，跳过", rawEvent.EventID)
			continue
		}

		log.Infof("告警事件 %d 关联到故障点: fault_id=%d", rawEvent.EventID, faultPoint.FaultID)

		//将恢复事件的 event_id 写入故障点的 RelationEventIDs
		faultPoint.RelationEventIDs = slice.AppendUniqueUint64(faultPoint.RelationEventIDs, event.EventID)
		faultPoint.FaultUpdateTime = timex.NowLocalTime()

		if event.EventRecoveryTime != nil && event.EventRecoveryTime.After(faultPoint.FaultLatestTime) {
			faultPoint.FaultLatestTime = *event.EventRecoveryTime
		}

		if event.EventOccurTime != nil {
			faultPoint.FaultDurationTime = int64(faultPoint.FaultLatestTime.Sub(*event.EventOccurTime).Seconds())
		}

		if err := s.repoFactory.FaultPoints().Upsert(ctx, *faultPoint); err != nil {
			return errors.Wrap(err, "更新故障点失败")
		}

		//最新事件是恢复事件，标记故障点为 已恢复
		log.Infof("故障点 %d 最新事件是恢复状态，标记故障点为 recovered", faultPoint.FaultID)
		recoveryTime := timex.NowLocalTime()
		if event.EventRecoveryTime != nil {
			recoveryTime = *event.EventRecoveryTime
		}
		if err := s.repoFactory.FaultPoints().MakeRecovered(ctx, faultPoint.FaultID, recoveryTime); err != nil {
			return errors.Wrap(err, "更新故障点状态失败")
		}

		// 将 fault_id 回写到恢复事件
		if err := s.linkFaultToEvents(ctx, faultPoint.FaultID, []uint64{event.EventID}); err != nil {
			return err
		}

		// 如果故障点已关联问题，也回写 problem_id 到恢复事件
		if faultPoint.ProblemID != 0 {
			if err := s.linkProblemToEvents(ctx, faultPoint.ProblemID, []uint64{event.EventID}); err != nil {
				log.Infof("回写 problem_id=%d 到恢复事件 event_id=%d 失败: %v", faultPoint.ProblemID, event.EventID, err)
			} else {
				log.Infof("已回写 problem_id=%d 到恢复事件 event_id=%d", faultPoint.ProblemID, event.EventID)
			}
		}

		//通知 ProblemStage 检查问题是否可以恢复
		if err := s.problemHandler.HandleFaultPointRecovered(ctx, faultPoint.FaultID); err != nil {
			return err
		}
	}
	return nil
}

// OnProblemLinked 冗余回写 problem_id 到故障点。
func (s *FaultPointStage) OnProblemLinked(ctx context.Context, problemID uint64, faultIDs []uint64) error {
	return s.repoFactory.FaultPoints().UpdateProblemID(ctx, faultIDs, problemID)
}

// linkFaultToEvents 将 fault_id 回写到事件，保持冗余。
func (s *FaultPointStage) linkFaultToEvents(ctx context.Context, faultID uint64, RelationEventIDs []uint64) error {
	return s.repoFactory.RawEvents().UpdateFaultID(ctx, RelationEventIDs, faultID)
}

// linkProblemToEvents 将 problem_id 回写到事件，保持冗余。
func (s *FaultPointStage) linkProblemToEvents(ctx context.Context, problemID uint64, eventIDs []uint64) error {
	return s.repoFactory.RawEvents().UpdateProblemID(ctx, eventIDs, problemID)
}

// writeFaultPointRelation 写入故障点关系。
func (s *FaultPointStage) writeFaultPointRelation(ctx context.Context, fp domain.FaultPointObject, event domain.RawEvent) error {
	if s.repoFactory.FaultPointRelations() == nil {
		return nil // 如果未配置关系存储，跳过
	}

	relation := domain.FaultPointRelation{
		RelationId:         s.genID.NextID(),
		RelationClass:      event.EntityObjectClass,
		RelationCreateTime: timex.NowLocalTime(),
		RelationUpdateTime: timex.NowLocalTime(),
		SourceObjectId:     fp.EntityObjectID,
		SourceObjectClass:  fp.EntityObjectClass,
		TargetObjectId:     fmt.Sprintf("%d", fp.FaultID),
		TargetObjectClass:  "fault_point",
	}

	return s.repoFactory.FaultPointRelations().Upsert(ctx, relation)
}

// Run 在 errgroup 中运行失效检查器，定期检查并标记过期故障点。
func (s *FaultPointStage) Run(ctx context.Context) error {
	//if !s.cfg.AppConfig.FaultPoint.Expiration.Enabled {
	//	log.Info("故障点失效检查未启用")
	//	return nil
	//}

	//log.Infof("启动故障点失效检查器，检查间隔: %v, 失效时间: %v", defaultCheckInterval, s.cfg.AppConfig.FaultPoint.Expiration.ExpirationTime)

	ticker := time.NewTicker(defaultCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("故障点失效检查器收到停止信号")
			return ctx.Err()
		case <-ticker.C:
			if err := s.checkAndExpireFaultPoints(ctx); err != nil {
				log.Errorf("执行故障点失效检查失败: %v", err)
			}
		}
	}
}

// checkAndExpireFaultPoints 检查并标记过期的故障点。
func (s *FaultPointStage) checkAndExpireFaultPoints(ctx context.Context) error {
	expirationTime := time.Now().Add(-s.cfgManager.GetConfig().AppConfig.FaultPoint.Expiration.ExpirationTime)
	log.Infof("执行故障点失效检查，失效时间点: %s", expirationTime.Format(time.DateTime))

	expiredFPs, err := s.repoFactory.FaultPoints().FindExpiredOccurred(ctx, expirationTime)
	if err != nil {
		log.Errorf("查询过期故障点失败: %v", err)
		return errors.Wrap(err, "查询过期故障点失败")
	}

	if len(expiredFPs) == 0 {
		log.Info("未发现过期故障点")
		return nil
	}

	log.Infof("发现 %d 个过期故障点，开始处理", len(expiredFPs))

	expiredCount := 0
	for _, fp := range expiredFPs {
		if err := s.repoFactory.FaultPoints().MakeExpired(ctx, fp.FaultID); err != nil {
			log.Errorf("标记故障点 %d 为失效失败: %v", fp.FaultID, err)
			continue
		}
		log.Debugf("故障点 %d 已过期，标记为失效 (最新时间: %s, 过期阈值: %s)",
			fp.FaultID, fp.FaultLatestTime.Format(time.DateTime), expirationTime.Format(time.DateTime))
		expiredCount++
	}

	log.Infof("故障点失效检查完成，共发现 %d 个，成功标记 %d 个", len(expiredFPs), expiredCount)
	return nil
}

var _ core.FaultPointHandler = (*FaultPointStage)(nil)
