package kafka

import (
	"context"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConsumer(t *testing.T) {
	Convey("TestNewConsumer", t, func() {
		Convey("使用默认 GroupID 创建消费者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				GroupID: "", // 使用默认值
			}

			consumer, err := NewConsumer(cfg)

			So(err, ShouldBeNil)
			So(consumer, ShouldNotBeNil)
			c := consumer.(*Consumer)
			So(c.reader, ShouldNotBeNil)

			// 清理
			c.reader.Close()
		})

		Convey("使用自定义 GroupID 创建消费者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				GroupID: "custom-group",
			}

			consumer, err := NewConsumer(cfg)

			So(err, ShouldBeNil)
			So(consumer, ShouldNotBeNil)

			// 清理
			consumer.Close()
		})

		Convey("使用 SASL 认证创建消费者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				GroupID: "test-group",
				SASL: &SASLConfig{
					Enabled:   true,
					Mechanism: "PLAIN",
					Username:  "user",
					Password:  "pass",
				},
			}

			consumer, err := NewConsumer(cfg)

			So(err, ShouldBeNil)
			So(consumer, ShouldNotBeNil)

			// 清理
			consumer.Close()
		})

		Convey("SASL 机制不支持返回错误", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				SASL: &SASLConfig{
					Enabled:   true,
					Mechanism: "UNSUPPORTED",
					Username:  "user",
					Password:  "pass",
				},
			}

			consumer, err := NewConsumer(cfg)

			So(err, ShouldNotBeNil)
			So(consumer, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "构建 SASL 认证失败")
		})
	})
}

func TestConsumer_Close(t *testing.T) {
	Convey("TestConsumer_Close", t, func() {
		Convey("reader 为 nil 时关闭返回 nil", func() {
			consumer := &Consumer{reader: nil}

			err := consumer.Close()

			So(err, ShouldBeNil)
		})

		Convey("成功关闭 reader", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			consumer, _ := NewConsumer(cfg)

			err := consumer.Close()

			So(err, ShouldBeNil)
		})
	})
}

func TestConsumer_ConsumeRawEvents(t *testing.T) {
	Convey("TestConsumer_ConsumeRawEvents", t, func() {
		Convey("reader 为 nil 返回错误", func() {
			consumer := &Consumer{reader: nil}
			ctx := context.Background()

			err := consumer.ConsumeRawEvents(ctx, func(ctx context.Context, msg core.KafkaMessage) error {
				return nil
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "kafka reader 未初始化")
		})

		Convey("context 取消时返回 context 错误", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			consumer, _ := NewConsumer(cfg)
			c := consumer.(*Consumer)
			defer c.Close()

			// 打桩 FetchMessage 返回 context.Canceled
			patches := gomonkey.ApplyMethod(c.reader, "FetchMessage",
				func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
					return kafka.Message{}, context.Canceled
				})
			defer patches.Reset()

			ctx, cancel := context.WithCancel(context.Background())
			cancel() // 立即取消

			err := consumer.ConsumeRawEvents(ctx, func(ctx context.Context, msg core.KafkaMessage) error {
				return nil
			})

			So(err, ShouldNotBeNil)
		})

		Convey("FetchMessage 失败返回错误", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			consumer, _ := NewConsumer(cfg)
			c := consumer.(*Consumer)
			defer c.Close()

			// 打桩 FetchMessage 返回错误
			fetchErr := errors.New("fetch failed")
			patches := gomonkey.ApplyMethod(c.reader, "FetchMessage",
				func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
					return kafka.Message{}, fetchErr
				})
			defer patches.Reset()

			ctx := context.Background()
			err := consumer.ConsumeRawEvents(ctx, func(ctx context.Context, msg core.KafkaMessage) error {
				return nil
			})

			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, fetchErr)
		})

		Convey("成功消费消息并提交", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			consumer, _ := NewConsumer(cfg)
			c := consumer.(*Consumer)
			defer c.Close()

			// 记录处理的消息
			var handledMessages []core.KafkaMessage
			callCount := 0

			// 创建可取消的 context
			ctx, cancel := context.WithCancel(context.Background())

			// 打桩 FetchMessage 返回消息然后取消 context
			patches := gomonkey.ApplyMethod(c.reader, "FetchMessage",
				func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
					callCount++
					if callCount > 1 {
						cancel() // 取消 context
						return kafka.Message{}, context.Canceled
					}
					return kafka.Message{
						Key:       []byte("key-1"),
						Value:     []byte(`{"data": "test"}`),
						Partition: 0,
						Offset:    100,
						Time:      time.Now(),
					}, nil
				})
			defer patches.Reset()

			// 打桩 CommitMessages 成功
			patches.ApplyMethod(c.reader, "CommitMessages",
				func(_ *kafka.Reader, ctx context.Context, msgs ...kafka.Message) error {
					return nil
				})

			err := consumer.ConsumeRawEvents(ctx, func(ctx context.Context, msg core.KafkaMessage) error {
				handledMessages = append(handledMessages, msg)
				return nil
			})

			// context.Canceled 时返回 ctx.Err()
			So(err, ShouldNotBeNil)
			So(errors.Is(err, context.Canceled), ShouldBeTrue)
			So(len(handledMessages), ShouldEqual, 1)
			So(handledMessages[0].Key, ShouldEqual, "key-1")
			So(string(handledMessages[0].Value), ShouldEqual, `{"data": "test"}`)
			So(handledMessages[0].Partition, ShouldEqual, int32(0))
			So(handledMessages[0].Offset, ShouldEqual, int64(100))
		})

		Convey("handler 处理失败继续消费", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			consumer, _ := NewConsumer(cfg)
			c := consumer.(*Consumer)
			defer c.Close()

			callCount := 0
			handlerCallCount := 0

			// 创建可取消的 context
			ctx, cancel := context.WithCancel(context.Background())

			// 打桩 FetchMessage
			patches := gomonkey.ApplyMethod(c.reader, "FetchMessage",
				func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
					callCount++
					if callCount > 2 {
						cancel() // 取消 context
						return kafka.Message{}, context.Canceled
					}
					return kafka.Message{
						Key:   []byte("key"),
						Value: []byte("value"),
					}, nil
				})
			defer patches.Reset()

			// 打桩 CommitMessages 成功
			patches.ApplyMethod(c.reader, "CommitMessages",
				func(_ *kafka.Reader, ctx context.Context, msgs ...kafka.Message) error {
					return nil
				})

			err := consumer.ConsumeRawEvents(ctx, func(ctx context.Context, msg core.KafkaMessage) error {
				handlerCallCount++
				return errors.New("handler error") // handler 返回错误
			})

			So(err, ShouldNotBeNil)
			So(errors.Is(err, context.Canceled), ShouldBeTrue)
			So(handlerCallCount, ShouldEqual, 2) // 即使 handler 失败也继续消费
		})

		Convey("CommitMessages 失败返回错误", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			consumer, _ := NewConsumer(cfg)
			c := consumer.(*Consumer)
			defer c.Close()

			// 打桩 FetchMessage 成功
			patches := gomonkey.ApplyMethod(c.reader, "FetchMessage",
				func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
					return kafka.Message{
						Key:   []byte("key"),
						Value: []byte("value"),
					}, nil
				})
			defer patches.Reset()

			// 打桩 CommitMessages 失败
			patches.ApplyMethod(c.reader, "CommitMessages",
				func(_ *kafka.Reader, ctx context.Context, msgs ...kafka.Message) error {
					return errors.New("commit failed")
				})

			ctx := context.Background()
			err := consumer.ConsumeRawEvents(ctx, func(ctx context.Context, msg core.KafkaMessage) error {
				return nil
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "commit kafka offset")
		})
	})
}
