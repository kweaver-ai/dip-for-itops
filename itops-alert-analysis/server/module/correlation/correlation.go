package correlation

import (
	"context"
	"fmt"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/dip"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/kafka"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/correlation/objectclass"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/correlation/standardizer"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const defaultCheckInterval = 5 * time.Minute //多长时间运行一次

// Service
type Service struct {
	ingest           *IngestStage
	faultPoint       *FaultPointStage
	problem          *ProblemStage
	objectClassCache *objectclass.ObjectClass
	kafkaProducer    core.KafkaProducer
	kafkaConsumer    core.KafkaConsumer
}

// New 初始化数据收敛
func New(
	cfgManager *config.ConfigManager,
	repoFactory *opensearch.RepositoryFactory,
	dipClient *dip.Client,

) (*Service, error) {
	var err error
	var cfg = cfgManager.GetConfig()

	// 创建对象类缓存
	objectClassCache, err := objectclass.New(cfg, dipClient)
	if err != nil {
		return nil, errors.Wrap(err, "初始化 objectclass 失败")
	}

	// 初始化 Kafka Producer
	kafkaProducer, err := kafka.NewProducer(kafka.Config{
		Brokers: []string{fmt.Sprintf("%s:%d", cfg.DepServices.MQ.MQHost, cfg.DepServices.MQ.MQPort)},
		SASL: &kafka.SASLConfig{
			Enabled:  true,
			Username: cfg.DepServices.MQ.Auth.Username,
			Password: cfg.DepServices.MQ.Auth.Password,
		},
		Topic: cfg.Kafka.ProblemEvents.Topic,
	})
	if err != nil {
		return nil, errors.Wrap(err, "创建kafka生产者失败")
	}

	// 创建原始事件流的 Kafka Consumer
	kafkaConsumer, err := kafka.NewConsumer(kafka.Config{
		Brokers: []string{fmt.Sprintf("%s:%d", cfg.DepServices.MQ.MQHost, cfg.DepServices.MQ.MQPort)},
		SASL: &kafka.SASLConfig{
			Enabled:  true,
			Username: cfg.DepServices.MQ.Auth.Username,
			Password: cfg.DepServices.MQ.Auth.Password,
		},
		Topic:   cfg.Kafka.RawEvents.Topic,
		GroupID: cfg.Kafka.RawEvents.ConsumerGroup,
	})
	if err != nil {
		return nil, errors.Wrap(err, "创建kafka消费者失败")
	}

	// 创建标准化器（使用对象类缓存）
	std, err := standardizer.Build(cfg, objectClassCache)
	if err != nil {
		return nil, errors.Wrap(err, "初始化 standardizer 失败")
	}

	// 创建空间相关性检查器
	spatialChecker := dip.NewSpatialChecker(dipClient)

	// 创建问题阶段
	problemStage := NewProblemStage(cfgManager, repoFactory, kafkaProducer, spatialChecker)
	faultStage := NewFaultPointStage(cfgManager, repoFactory, problemStage)
	ingestStage := NewIngestStage(repoFactory, faultStage, std, kafkaConsumer)

	return &Service{
		ingest:           ingestStage,
		faultPoint:       faultStage,
		problem:          problemStage,
		objectClassCache: objectClassCache,
		kafkaProducer:    kafkaProducer,
		kafkaConsumer:    kafkaConsumer,
	}, nil
}

// Start 启动关联分析模块（启动 Ingest 消费）。
func (c *Service) Start(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	// 启动对象类缓存刷新器
	if c.objectClassCache != nil {
		eg.Go(func() error {
			if err := c.objectClassCache.Run(egCtx); err != nil && !errors.Is(err, context.Canceled) {
				return errors.Wrap(err, "对象类缓存刷新器启动失败")
			}
			return nil
		})
	}

	// 启动事件消费（Ingest Stage）
	eg.Go(func() error {
		if err := c.ingest.Start(egCtx); err != nil && !errors.Is(err, context.Canceled) {
			return errors.Wrap(err, "ingest stage 启动失败")
		}
		return nil
	})

	// 启动问题失效检查器（内置在 ProblemStage）
	if c.problem != nil {
		eg.Go(func() error {
			if err := c.problem.Run(egCtx); err != nil && !errors.Is(err, context.Canceled) {
				return errors.Wrap(err, "问题失效检查器启动失败")
			}
			return nil
		})
	}

	if c.faultPoint != nil {
		eg.Go(func() error {
			if err := c.faultPoint.Run(egCtx); err != nil && !errors.Is(err, context.Canceled) {
				return errors.Wrap(err, "问题失效检查器启动失败")
			}
			return nil
		})
	}

	return eg.Wait()
}

// HandleFaultPoint 实现 ProblemHandler 接口 - 处理故障点。
func (c *Service) HandleFaultPoint(ctx context.Context, fp domain.FaultPointObject) error {
	return c.problem.HandleFaultPoint(ctx, fp)
}

// HandleRCACallback 实现 ProblemHandler 接口 - 处理 RCA 回调。
func (c *Service) HandleRCACallback(ctx context.Context, cb domain.RCACallback) error {
	return c.problem.HandleRCACallback(ctx, cb)
}

// CloseProblem 实现 ProblemHandler 接口 - 关闭问题。
func (c *Service) CloseProblem(ctx context.Context, problemID uint64, closeType domain.ProblemCloseType, closeStatus domain.ProblemStatus, notes string, by string) error {
	return c.problem.CloseProblem(ctx, problemID, closeType, closeStatus, notes, by)
}

// HandleFaultPointRecovered 实现 ProblemHandler 接口 - 处理故障点恢复。
func (c *Service) HandleFaultPointRecovered(ctx context.Context, faultID uint64) error {
	return c.problem.HandleFaultPointRecovered(ctx, faultID)
}

// Close 关闭 CorrelationService 持有的资源。
func (c *Service) Close() error {
	var errs []error
	if c.kafkaConsumer != nil {
		if err := c.kafkaConsumer.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "close kafkaConsumer"))
		}
	}
	if c.kafkaProducer != nil {
		if err := c.kafkaProducer.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "close kafkaProducer"))
		}
	}

	// 关闭对象类缓存
	if c.objectClassCache != nil {
		if err := c.objectClassCache.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "close object class cache"))
		}
	}
	if len(errs) > 0 {
		return errors.New(fmt.Sprintf("关闭 correlationService 时发生 %d 个错误: %v", len(errs), errs))
	}

	return nil
}

// 确保 CorrelationService 实现了 ProblemHandler 接口
var _ core.ProblemHandler = (*Service)(nil)
