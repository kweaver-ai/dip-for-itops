package kafka

import (
	"context"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

// Consumer 基于 kafka-go Reader 实现顺序消费。
type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(cfg Config) (core.KafkaConsumer, error) {
	groupID := cfg.GroupID
	if groupID == "" {
		groupID = defaultGroupID
	}

	mechanism, err := buildSASLMechanism(cfg.SASL)
	if err != nil {
		return nil, errors.Wrap(err, "构建 SASL 认证失败")
	}

	// 为 SASL_PLAINTEXT 协议配置 Dialer（不使用 TLS）
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}

	readerCfg := kafka.ReaderConfig{
		Brokers:       cfg.Brokers,
		Topic:         cfg.Topic,
		GroupID:       groupID,
		MinBytes:      minBytes,
		MaxBytes:      maxBytes,
		QueueCapacity: 1,
		Dialer:        dialer,
	}
	return &Consumer{
		reader: kafka.NewReader(readerCfg),
	}, nil
}

func (c *Consumer) ConsumeRawEvents(ctx context.Context, handler func(ctx context.Context, msg core.KafkaMessage) error) error {
	if c.reader == nil {
		return errors.New("kafka reader 未初始化")
	}
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return ctx.Err()
			}
			return err
		}

		if err := handler(ctx, core.KafkaMessage{
			Key:       string(msg.Key),
			Value:     msg.Value,
			Partition: int32(msg.Partition),
			Offset:    msg.Offset,
			Timestamp: msg.Time,
		}); err != nil {
			log.Errorf("kafka handler 处理失败，partition=%d offset=%d err=%v,body=%+v", msg.Partition, msg.Offset, err, string(msg.Value))
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return errors.Wrap(err, "commit kafka offset")
		}
	}
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
