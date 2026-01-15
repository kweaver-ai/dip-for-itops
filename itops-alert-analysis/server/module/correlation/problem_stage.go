package correlation

import (
	"context"
	"encoding/json"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/timex"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/dip"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/idgen"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/slice"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

const (
	ProblemEventTopic = "itops_alert_problem_event"
)

// ProblemEventMessage 问题事件消息
type ProblemEventMessage struct {
	ProblemID uint64 `json:"problem_id"`
}

// ProblemStage 负责故障点合并、问题生命周期与 RCA。
type ProblemStage struct {
	cfgManager     *config.ConfigManager
	repoFactory    *opensearch.RepositoryFactory
	kafkaProducer  core.KafkaProducer
	genID          *idgen.Generator
	spatialChecker *dip.SpatialChecker
}

func NewProblemStage(cfgManager *config.ConfigManager, repoFactory *opensearch.RepositoryFactory, kafkaProducer core.KafkaProducer, spatialChecker *dip.SpatialChecker) *ProblemStage {
	return &ProblemStage{
		cfgManager:     cfgManager,
		repoFactory:    repoFactory,
		kafkaProducer:  kafkaProducer,
		genID:          idgen.New(),
		spatialChecker: spatialChecker,
	}
}

// HandleFaultPoint 接收故障点，合并/创建问题。
func (s *ProblemStage) HandleFaultPoint(ctx context.Context, fp domain.FaultPointObject) error {
	// 1. 计算时间窗口
	windowEnd := fp.FaultLatestTime
	expirationStartTime := windowEnd.Add(-s.cfgManager.GetConfig().AppConfig.Problem.Expiration.ExpirationTime)
	log.Infof("执行问题收敛，最早时间点: %s", expirationStartTime.Format(time.DateTime))

	// 2. 在时间窗口内查找可关联的问题
	correlatedProblems, err := s.repoFactory.Problems().FindCorrelated(ctx, fp, expirationStartTime)
	if err != nil {
		return errors.Wrap(err, "查找相关问题失败")
	}
	log.Infof("为故障点:%d，查询到问题数量:%d", fp.FaultID, len(correlatedProblems))

	var targetProblems []domain.Problem
	if len(correlatedProblems) > 0 {
		// 3.1 空间相关性判断
		//spatiallyCorrelatedProblems 表示该故障点已经关联了哪些问题？
		spatiallyCorrelatedProblems, err := s.spatialChecker.FilterCorrelatedProblems(ctx, fp, correlatedProblems)
		if err != nil {
			return errors.Wrap(err, "空间相关性判断失败")
		}
		// 存在空间相关性的并且开启了语义相关性的那么就开启语义相关性的处理
		if len(spatiallyCorrelatedProblems) > 0 {
			// 未开启语义相关性，直接使用空间相关的问题
			targetProblems = spatiallyCorrelatedProblems
			log.Infof("语义相关性未启用，直接使用 %d 个空间相关的问题", len(targetProblems))
		}
	}

	var problemID uint64
	var eventsToUpdate []uint64 // 需要回写 problem_id 的事件列表

	if len(targetProblems) > 0 {
		// 4. 合并到已有问题
		log.Infof("故障点 %d 合并到 %d 个问题", fp.FaultID, len(targetProblems))
		mergedProblem, err := s.mergeFaultPointIntoProblems(ctx, fp, targetProblems)
		if err != nil {
			return err
		}
		problemID = mergedProblem.ProblemID
		// 合并场景：需要更新合并后问题的所有事件（因为主问题原有的事件也需要更新 problem_id）
		eventsToUpdate = mergedProblem.RelationEventIDs
	} else {
		// 5. 创建新问题
		problemID = s.genID.NextID()
		log.Infof("为故障点 %d 创建新问题 %d", fp.FaultID, problemID)

		newProblem := domain.Problem{
			ProblemID:              problemID,
			ProblemName:            fp.FaultName,
			ProblemCreateTimestamp: timex.NowLocalTime().Local(),
			ProblemUpdateTime:      timex.NowLocalTime().Local(),
			ProblemOccurTime:       fp.FaultOccurTime,
			ProblemLatestTime:      fp.FaultLatestTime,
			ProblemDuration:        timex.AbsSecondsBetween(fp.FaultLatestTime, fp.FaultOccurTime),
			ProblemStatus:          domain.ProblemStatusOpen,
			ProblemLevel:           fp.FaultLevel,
			AffectedEntityIDs:      []string{fp.EntityObjectID},
			RelationIDs:            []uint64{fp.FaultID},
			RelationEventIDs:       fp.RelationEventIDs,
		}

		if err := s.repoFactory.Problems().Upsert(ctx, newProblem); err != nil {
			return errors.Wrap(err, "创建问题失败")
		}
		// 创建场景：只需更新当前故障点的事件
		eventsToUpdate = fp.RelationEventIDs
	}
	// 发布问题创建事件到 Kafka，由 RCA 模块订阅处理
	if err := s.publishProblemEvent(ctx, problemID); err != nil {
		log.Infof("发布问题事件失败 problem_id=%d: %v", problemID, err)
	}

	// 6. 将 problem_id 回写到故障点
	log.Debugf("回写故障点索引:%d,问题id:%d", fp.FaultID, problemID)
	if err := s.repoFactory.FaultPoints().UpdateProblemID(ctx, []uint64{fp.FaultID}, problemID); err != nil {
		return errors.Wrap(err, "回写 problem_id 到故障点失败")
	}

	// 7. 将 problem_id 回写到所有关联的事件
	log.Debugf("回写原始事件索引:%+v,问题id:%d", eventsToUpdate, problemID)
	if err := s.repoFactory.RawEvents().UpdateProblemID(ctx, eventsToUpdate, problemID); err != nil {
		return errors.Wrapf(err, "回写 problem_id 到事件失败,事件ID：%+v", fp.RelationEventIDs)
	}

	log.Infof("故障点 %d 已关联到问题 %d，关联事件数: %d", fp.FaultID, problemID, len(fp.RelationEventIDs))
	return nil
}

// mergeFaultPointIntoProblems 将故障点合并到问题
// 如果有多个问题，会将它们合并为一个主问题
func (s *ProblemStage) mergeFaultPointIntoProblems(ctx context.Context, fp domain.FaultPointObject, problems []domain.Problem) (*domain.Problem, error) {
	if len(problems) == 0 {
		return nil, errors.New("没有问题可合并")
	}

	// 主问题和被合并的问题
	mainProblem := problems[0]
	otherProblems := problems[1:]

	log.Infof("选择问题 %d 作为主问题", mainProblem.ProblemID)

	// 合并其他问题到主问题
	for _, problem := range otherProblems {
		log.Infof("合并问题 %d 到主问题 %d", problem.ProblemID, mainProblem.ProblemID)

		// 合并故障点列表
		for _, fpID := range problem.RelationIDs {
			mainProblem.RelationIDs = slice.AppendUniqueUint64(mainProblem.RelationIDs, fpID)
		}

		// 合并事件列表
		for _, eventID := range problem.RelationEventIDs {
			mainProblem.RelationEventIDs = slice.AppendUniqueUint64(mainProblem.RelationEventIDs, eventID)
		}

		// 合并受影响的实体列表
		for _, entityID := range problem.AffectedEntityIDs {
			mainProblem.AffectedEntityIDs = slice.AppendUniqueString(mainProblem.AffectedEntityIDs, entityID)
		}

		// 更新时间范围
		if problem.ProblemOccurTime.Before(mainProblem.ProblemOccurTime) {
			mainProblem.ProblemOccurTime = problem.ProblemOccurTime
		}
		if problem.ProblemLatestTime.After(mainProblem.ProblemLatestTime) {
			mainProblem.ProblemLatestTime = problem.ProblemLatestTime
		}

		// 更新问题等级（值越小等级越高）
		if problem.ProblemLevel < mainProblem.ProblemLevel {
			mainProblem.ProblemLevel = problem.ProblemLevel
		}
	}

	// 将当前故障点也加入主问题
	mainProblem.RelationIDs = slice.AppendUniqueUint64(mainProblem.RelationIDs, fp.FaultID)
	for _, eventID := range fp.RelationEventIDs {
		mainProblem.RelationEventIDs = slice.AppendUniqueUint64(mainProblem.RelationEventIDs, eventID)
	}
	mainProblem.AffectedEntityIDs = slice.AppendUniqueString(mainProblem.AffectedEntityIDs, fp.EntityObjectID)

	if fp.FaultLatestTime.After(mainProblem.ProblemLatestTime) {
		mainProblem.ProblemLatestTime = fp.FaultLatestTime
	}
	if fp.FaultLevel < mainProblem.ProblemLevel {
		mainProblem.ProblemLevel = fp.FaultLevel
	}

	// 更新持续时间
	mainProblem.ProblemDuration = uint64(mainProblem.ProblemLatestTime.Sub(mainProblem.ProblemOccurTime).Seconds())

	// 更新更新时间
	mainProblem.ProblemUpdateTime = timex.NowLocalTime().Local()

	// 保存主问题
	if err := s.repoFactory.Problems().Upsert(ctx, mainProblem); err != nil {
		return nil, errors.Wrap(err, "保存问题失败")
	}

	// 处理被合并的问题（更新关联关系并关闭）
	for _, problem := range otherProblems {
		if err := s.repoFactory.FaultPoints().UpdateProblemID(ctx, problem.RelationIDs, mainProblem.ProblemID); err != nil {
			log.Infof("更新故障点 problem_id 失败（问题 %d）: %v", problem.ProblemID, err)
		}
		if err := s.repoFactory.RawEvents().UpdateProblemID(ctx, problem.RelationEventIDs, mainProblem.ProblemID); err != nil {
			log.Infof("更新事件 problem_id 失败（问题 %d）: %v", problem.ProblemID, err)
		}

		//清空被合并问题的关联数据（故障点、事件、RCA结果等）
		if err := s.repoFactory.Problems().ClearMergedProblemData(ctx, problem.ProblemID); err != nil {
			log.Infof("清空被合并问题 %d 的关联数据失败: %v", problem.ProblemID, err)
		} else {
			log.Infof("已清空被合并问题 %d 的关联数据", problem.ProblemID)
		}

		//关闭被合并的问题
		if err := s.repoFactory.Problems().MarkClosed(ctx, problem.ProblemID, domain.ProblemCloseTypeSystem, domain.ProblemStatusMerged, 0, "合并到问题"+cast.ToString(mainProblem.ProblemID), "system"); err != nil {
			log.Infof("关闭被合并问题 %d 失败: %v", problem.ProblemID, err)
		}
	}

	if len(otherProblems) > 0 {
		log.Infof("已将 %d 个问题合并为问题 %d", len(problems), mainProblem.ProblemID)
	}
	return &mainProblem, nil
}

// HandleRCACallback 处理 RCA 模块的异步回调，更新问题根因。
func (s *ProblemStage) HandleRCACallback(ctx context.Context, cb domain.RCACallback) error {
	log.Debugf("收到rca回调,问题id:%d,内容:%s", cb.ProblemID, utils.JsonEncode(cb))
	if cb.InProgress {
		// RCA 仍在运行，无需更新。
		return nil
	}
	return s.repoFactory.Problems().UpdateRootCause(ctx, cb.ProblemID, cb)
}

// CloseProblem 关闭问题，并按需更新冗余字段。
func (s *ProblemStage) CloseProblem(ctx context.Context, problemID uint64, closeType domain.ProblemCloseType, closeStatus domain.ProblemStatus, notes string, by string) error {
	return s.repoFactory.Problems().MarkClosed(ctx, problemID, closeType, closeStatus, 0, notes, by)
}

// HandleFaultPointRecovered 处理故障点恢复，检查问题是否可以恢复
func (s *ProblemStage) HandleFaultPointRecovered(ctx context.Context, faultID uint64) error {
	//查询故障点获取关联的 problem_id
	faultPoints, err := s.repoFactory.FaultPoints().QueryByIDs(ctx, []uint64{faultID})
	if err != nil {
		return errors.Wrap(err, "查询故障点失败")
	}
	if len(faultPoints) == 0 {
		log.Infof("故障点 %d 不存在，跳过问题恢复检查", faultID)
		return nil
	}

	fp := faultPoints[0]
	if fp.ProblemID == 0 {
		log.Infof("故障点 %d 未关联问题，跳过问题恢复检查", faultID)
		return nil
	}

	log.Infof("故障点 %d 已恢复，检查问题 %d 是否可以恢复", faultID, fp.ProblemID)

	//查询问题
	problems, err := s.repoFactory.Problems().QueryByIDs(ctx, []uint64{fp.ProblemID})
	if err != nil {
		return errors.Wrap(err, "查询问题失败")
	}
	if len(problems) == 0 {
		log.Infof("问题 %d 不存在，跳过恢复检查", fp.ProblemID)
		return nil
	}

	problem := problems[0]
	if problem.ProblemStatus != domain.ProblemStatusOpen {
		log.Infof("问题 %d 状态不是 open (%s)，跳过恢复检查", problem.ProblemID, problem.ProblemStatus)
		return nil
	}
	// 更新问题的 RelationEventIDs
	allEventIDs := problem.RelationEventIDs
	for _, d := range fp.RelationEventIDs {
		allEventIDs = slice.AppendUniqueUint64(allEventIDs, d)
	}
	if err := s.repoFactory.Problems().UpdateRelationEventIDs(ctx, problem.ProblemID, allEventIDs); err != nil {
		return errors.Wrap(err, "更新问题 RelationEventIDs 失败")
	}
	log.Infof("已更新问题 %d 的 RelationEventIDs，总事件数: %d", problem.ProblemID, len(allEventIDs))

	isRecovered, err := s.checkFaultPointIsRecovered(ctx, problem.RelationIDs)
	if err != nil {
		return err
	}
	//所有故障点都已恢复，标记问题为 closed (system)
	if !isRecovered {
		return nil
	}
	log.Infof("问题 %d 的所有故障点(%d个)均已恢复，标记问题为 closed", problem.ProblemID, len(problem.RelationIDs))

	if err := s.repoFactory.Problems().MarkClosed(ctx, problem.ProblemID, domain.ProblemCloseTypeSystem, domain.ProblemStatusClosed, uint64(problem.ProblemLatestTime.Sub(problem.ProblemOccurTime).Seconds()), "所有故障点已恢复", "system"); err != nil {
		return err
	}

	return nil
}

// checkFaultPointIsRecovered 检查故障点是否完全关闭
func (s *ProblemStage) checkFaultPointIsRecovered(ctx context.Context, relatedFpIDs []uint64) (bool, error) {
	//查询问题关联的所有故障点
	relatedFaultPoints, err := s.repoFactory.FaultPoints().QueryByIDs(ctx, relatedFpIDs)
	if err != nil {
		return false, errors.Wrap(err, "查询问题关联的故障点失败")
	}

	// 4. 检查是否所有故障点都已恢复
	allRecovered := true
	for _, relatedFP := range relatedFaultPoints {
		if relatedFP.FaultStatus != domain.FaultStatusRecovered {
			allRecovered = false
			log.Debugf("问题 %d 的故障点 %d 尚未恢复 (status=%s)", relatedFP.ProblemID, relatedFP.FaultID, relatedFP.FaultStatus)
			break
		}
	}
	return allRecovered, nil
}

// Run 启动问题失效检查定时器
func (s *ProblemStage) Run(ctx context.Context) error {
	//if !s.cfg.AppConfig.Problem.Expiration.Enabled {
	//	log.Info("问题失效检查未启用")
	//	return nil
	//}

	//cfg := s.cfgManager.GetConfig().AppConfig.Problem.Expiration
	//log.Infof("启动问题失效检查器，检查间隔: %v, 失效时间: %v", defaultCheckInterval, cfg.ExpirationTime)

	ticker := time.NewTicker(defaultCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("问题失效检查器收到停止信号")
			return ctx.Err()
		case <-ticker.C:
			if err := s.checkAndExpireProblems(ctx); err != nil {
				log.Errorf("执行失效检查失败: %v", err)
			}
		}
	}
}

// checkAndExpireProblems 检查并关闭过期的问题
func (s *ProblemStage) checkAndExpireProblems(ctx context.Context) error {
	cfg := s.cfgManager.GetConfig().AppConfig.Problem.Expiration
	expirationTime := time.Now().Add(-cfg.ExpirationTime)
	log.Infof("执行问题失效检查，失效时间点: %s", expirationTime.Format(time.DateTime))

	expiredProblems, err := s.repoFactory.Problems().FindExpiredOpen(ctx, expirationTime)
	if err != nil {
		log.Errorf("查询过期问题失败: %v", err)
		return errors.Wrap(err, "查询过期问题失败")
	}

	if len(expiredProblems) == 0 {
		log.Info("未发现过期问题")
		return nil
	}

	log.Infof("发现 %d 个过期问题，开始处理", len(expiredProblems))

	expiredCount := 0
	failedCount := 0
	for _, problem := range expiredProblems {
		if err := s.repoFactory.Problems().MarkExpired(ctx, problem.ProblemID); err != nil {
			log.Errorf("标记问题 %d 为失效状态失败: %v", problem.ProblemID, err)
			failedCount++
			// 继续处理其他问题，不中断
			continue
		}
		expiredCount++
	}

	log.Infof("失效检查完成，共处理 %d 个问题，成功标记 %d 个失效，失败 %d 个", len(expiredProblems), expiredCount, failedCount)
	return nil
}

// publishProblemEvent 发布问题事件到 Kafka
func (s *ProblemStage) publishProblemEvent(ctx context.Context, problemID uint64) error {
	if s.kafkaProducer == nil {
		log.Infof("Kafka Producer 未配置，跳过发布问题事件 problem_id=%d", problemID)
		return nil
	}

	msg := ProblemEventMessage{
		ProblemID: problemID,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "序列化问题事件失败")
	}

	key := cast.ToString(problemID)
	if err := s.kafkaProducer.PublishRawEvent(ctx, key, body); err != nil {
		return errors.Wrap(err, "发送 Kafka 消息失败")
	}

	log.Infof("已发布问题事件到 Kafka，problem_id=%d, topic=%s", problemID, ProblemEventTopic)
	return nil
}

var _ core.ProblemHandler = (*ProblemStage)(nil)
