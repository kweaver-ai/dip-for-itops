package kafka

import (
	"context"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

// Producer 基于 kafka-go 实现 KafkaProducer。
type Producer struct {
	writer *kafka.Writer
}

func NewProducer(cfg Config) (core.KafkaProducer, error) {
	mechanism, err := buildSASLMechanism(cfg.SASL)
	if err != nil {
		return nil, errors.Wrap(err, "构建 SASL 认证失败")
	}

	// 为 SASL_PLAINTEXT 协议配置 Transport（不使用 TLS）
	transport := &kafka.Transport{
		SASL: mechanism,
		// TLS 未配置，使用 SASL_PLAINTEXT 协议
	}

	if mechanism != nil {
		log.Infof("Kafka Producer: 使用 SASL_PLAINTEXT 协议，mechanism=%T, brokers=%v", mechanism, cfg.Brokers)
	} else {
		log.Infof("Kafka Producer: 使用 PLAINTEXT 协议（无认证），brokers=%v", cfg.Brokers)
	}

	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Brokers...),
			Topic:                  cfg.Topic,
			Balancer:               &kafka.Hash{},
			AllowAutoTopicCreation: true,
			Transport:              transport,
			RequiredAcks:           kafka.RequireOne,
			Async:                  true,
			BatchSize:              10,
			WriteTimeout:           10 * time.Second,
			ReadTimeout:            10 * time.Second,
			Compression:            kafka.Snappy,
		},
	}, nil
}

func (p *Producer) PublishRawEvent(ctx context.Context, key string, value []byte) error {
	if p.writer == nil {
		return errors.New("kafka writer 未初始化")
	}
	msg := kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now().Local(),
	}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
