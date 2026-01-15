package correlation

import (
	"context"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/correlation/standardizer"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/timex"
	"github.com/pkg/errors"
)

// IngestStage 负责消费 Kafka、标准化并写入原始事件。
type IngestStage struct {
	rawEventsConsumer core.KafkaConsumer
	repoFactory       *opensearch.RepositoryFactory
	fpHandler         core.FaultPointHandler
	std               standardizer.Standardizer
}

func NewIngestStage(repoFactory *opensearch.RepositoryFactory, fpHandler core.FaultPointHandler, std standardizer.Standardizer, kafkaConsumer core.KafkaConsumer) *IngestStage {
	return &IngestStage{
		rawEventsConsumer: kafkaConsumer,
		repoFactory:       repoFactory,
		fpHandler:         fpHandler,
		std:               std,
	}
}

// Start 启动顺序消费 itops_alert_raw_event。
func (s *IngestStage) Start(ctx context.Context) error {
	if s.rawEventsConsumer == nil {
		return errors.New("kafka rawEventsConsumer not configured")
	}
	if s.std == nil {
		return errors.New("standardizer not configured")
	}
	return s.rawEventsConsumer.ConsumeRawEvents(ctx, s.handleKafkaMessage)
}

// handleKafkaMessage 处理 Kafka 消息：标准化、入库、下发。
func (s *IngestStage) handleKafkaMessage(ctx context.Context, msg core.KafkaMessage) error {
	raw, err := s.std.Standardize(ctx, msg.Value)
	if err != nil {
		return errors.Wrap(err, "standardize raw event")
	}

	// 在 defer 中记录处理耗时
	defer func(t time.Time) {
		duration := time.Since(t)
		// 根据是否是恢复事件记录不同的日志
		if raw.EventStatus == domain.EventStatusRecovered {
			log.Debugw("恢复事件处理完成",
				"event_id", raw.EventID,
				"recovery_id", raw.RecoveryId,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"duration_ms", duration.Milliseconds(),
			)
		} else {
			log.Debugw("发生事件处理完成",
				"event_id", raw.EventID,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"duration_ms", duration.Milliseconds(),
			)
		}
	}(timex.NowLocalTime())

	// 事件均做标准化的处理
	if err := s.repoFactory.RawEvents().Upsert(ctx, raw); err != nil {
		return errors.Wrap(err, "persist raw event")
	}

	// 转发给 fault_point 模块处理
	if s.fpHandler != nil {
		return s.fpHandler.HandleEvent(ctx, raw)
	}
	return nil
}

// OnFaultPointLinked 将生成的 fault_id 回写到相关事件。
func (s *IngestStage) OnFaultPointLinked(ctx context.Context, faultID uint64, RelationEventIDs []uint64) error {
	return s.repoFactory.RawEvents().UpdateFaultID(ctx, RelationEventIDs, faultID)
}

// OnProblemLinked 将生成的 problem_id 回写到相关事件。
func (s *IngestStage) OnProblemLinked(ctx context.Context, problemID uint64, RelationEventIDs []uint64) error {
	return s.repoFactory.RawEvents().UpdateProblemID(ctx, RelationEventIDs, problemID)
}
