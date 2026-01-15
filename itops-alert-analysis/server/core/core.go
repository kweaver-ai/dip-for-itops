package core

import (
	"context"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
)

// KafkaProducer 生产 Kafka 消息。
type KafkaProducer interface {
	PublishRawEvent(ctx context.Context, key string, value []byte) error
	Close() error
}

// KafkaConsumer 顺序消费 topic itops_alert_raw_event。
type KafkaConsumer interface {
	ConsumeRawEvents(ctx context.Context, handler func(ctx context.Context, msg KafkaMessage) error) error
	Close() error
}

// KafkaMessage 表示消费到的 Kafka 消息。
type KafkaMessage struct {
	Key       string
	Value     []byte
	Partition int32
	Offset    int64
	Timestamp time.Time
}

// RawEventRepository 管理 itops_raw_event 索引。
type RawEventRepository interface {
	Upsert(ctx context.Context, event domain.RawEvent) error
	UpdateFaultID(ctx context.Context, eventIDs []uint64, faultID uint64) error
	UpdateProblemID(ctx context.Context, eventIDs []uint64, problemID uint64) error
	QueryByIDs(ctx context.Context, ids []uint64) ([]domain.RawEvent, error)
	QueryByProviderID(ctx context.Context, providerIDs []string) ([]domain.RawEvent, error)
}

// FaultPointRepository 管理 itops_fault_point 索引。
type FaultPointRepository interface {
	FindOpenByEntityAndMode(ctx context.Context, entityObjectID, failureMode string, t time.Time) (*domain.FaultPointObject, error)
	Upsert(ctx context.Context, fp domain.FaultPointObject) error
	MakeExpired(ctx context.Context, faultID uint64) error
	UpdateProblemID(ctx context.Context, faultIDs []uint64, problemID uint64) error
	MakeRecovered(ctx context.Context, faultID uint64, recoveryTime time.Time) error
	QueryByIDs(ctx context.Context, ids []uint64) ([]domain.FaultPointObject, error)
	FindInWindow(ctx context.Context, entityID string, faultMode string, start, end time.Time) ([]domain.FaultPointObject, error)
	FindByEventID(ctx context.Context, eventID uint64) (*domain.FaultPointObject, error)
	FindExpiredOccurred(ctx context.Context, expirationTime time.Time) ([]domain.FaultPointObject, error)
}

// FaultPointRelationRepository 管理 itops_fault_point_relation 索引。
type FaultPointRelationRepository interface {
	Upsert(ctx context.Context, relation domain.FaultPointRelation) error
}

// ProblemRepository 管理 itops_problem 索引。
type ProblemRepository interface {
	FindCorrelated(ctx context.Context, fp domain.FaultPointObject, t time.Time) ([]domain.Problem, error)
	FindPendingRCA(ctx context.Context, maxAge time.Duration) ([]domain.Problem, error)
	FindExpiredOpen(ctx context.Context, expirationTime time.Time) ([]domain.Problem, error)
	Upsert(ctx context.Context, p domain.Problem) error
	UpdateRootCause(ctx context.Context, problemID uint64, cb domain.RCACallback) error
	UpdateRootCauseObjectID(ctx context.Context, problemID uint64, objectID string, faultID uint64) error
	UpdateRelationEventIDs(ctx context.Context, problemID uint64, eventIDs []uint64) error
	MarkClosed(ctx context.Context, problemID uint64, closeType domain.ProblemCloseType, closeStatus domain.ProblemStatus, duration uint64, notes string, by string) error
	MarkExpired(ctx context.Context, problemID uint64) error
	QueryByIDs(ctx context.Context, ids []uint64) ([]domain.Problem, error)
	ClearMergedProblemData(ctx context.Context, problemID uint64) error // 清空被合并问题的关联数据
}

// FaultCausalRepository 管理 itops_fault_causal 索引。
type FaultCausalRepository interface {
	Upsert(ctx context.Context, fc domain.FaultCausalObject) error
	Update(ctx context.Context, fc domain.FaultCausalObject) error
	QueryByIDs(ctx context.Context, ids []string) ([]domain.FaultCausalObject, error)
}

// FaultCausalRelationRepository 管理 itops_fault_causal_relation 索引。
type FaultCausalRelationRepository interface {
	Upsert(ctx context.Context, fcr domain.FaultCausalRelation) error
	Update(ctx context.Context, fcr domain.FaultCausalRelation) error
	QueryByIDs(ctx context.Context, ids []string) ([]domain.FaultCausalRelation, error)
	QueryByEntityPair(ctx context.Context, sourceID, targetID string) ([]domain.FaultCausalRelation, error)
}

// FaultPointHandler 是 ingest 的下游处理器。
type FaultPointHandler interface {
	HandleEvent(ctx context.Context, event domain.RawEvent) error
	OnProblemLinked(ctx context.Context, problemID uint64, faultIDs []uint64) error
}

// ProblemHandler 是故障点处理的下游处理器。
type ProblemHandler interface {
	HandleFaultPoint(ctx context.Context, fp domain.FaultPointObject) error
	HandleRCACallback(ctx context.Context, cb domain.RCACallback) error
	CloseProblem(ctx context.Context, problemID uint64, closeType domain.ProblemCloseType, closeState domain.ProblemStatus, notes string, by string) error
	HandleFaultPointRecovered(ctx context.Context, faultID uint64) error
}

// RCAClient 异步调用 RCA 模块。
type RCAClient interface {
	Submit(ctx context.Context, req domain.RCARequest) error
}

type RepositoryFactory interface {
	RawEvent() RawEventRepository
	FaultPoint() FaultPointRepository
	FaultPointRelation() FaultPointRelationRepository
	Problem() ProblemRepository
	FaultCausal() FaultCausalRepository
	FaultCausalRelation() FaultCausalRelationRepository
}
