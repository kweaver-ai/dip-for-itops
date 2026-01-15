package kafka

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewProducer(t *testing.T) {
	Convey("TestNewProducer", t, func() {
		Convey("无 SASL 认证创建生产者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				SASL:    nil,
			}

			producer, err := NewProducer(cfg)

			So(err, ShouldBeNil)
			So(producer, ShouldNotBeNil)
			p := producer.(*Producer)
			So(p.writer, ShouldNotBeNil)

			// 清理
			p.writer.Close()
		})

		Convey("使用多个 Broker 创建生产者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092", "localhost:9093", "localhost:9094"},
				Topic:   "test-topic",
			}

			producer, err := NewProducer(cfg)

			So(err, ShouldBeNil)
			So(producer, ShouldNotBeNil)

			// 清理
			producer.Close()
		})

		Convey("使用 PLAIN 认证创建生产者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				SASL: &SASLConfig{
					Enabled:   true,
					Mechanism: "PLAIN",
					Username:  "user",
					Password:  "pass",
				},
			}

			producer, err := NewProducer(cfg)

			So(err, ShouldBeNil)
			So(producer, ShouldNotBeNil)

			// 清理
			producer.Close()
		})

		Convey("使用 SCRAM-SHA-256 认证创建生产者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				SASL: &SASLConfig{
					Enabled:   true,
					Mechanism: "SCRAM-SHA-256",
					Username:  "user",
					Password:  "pass",
				},
			}

			producer, err := NewProducer(cfg)

			So(err, ShouldBeNil)
			So(producer, ShouldNotBeNil)

			// 清理
			producer.Close()
		})

		Convey("使用 SCRAM-SHA-512 认证创建生产者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				SASL: &SASLConfig{
					Enabled:   true,
					Mechanism: "SCRAM-SHA-512",
					Username:  "user",
					Password:  "pass",
				},
			}

			producer, err := NewProducer(cfg)

			So(err, ShouldBeNil)
			So(producer, ShouldNotBeNil)

			// 清理
			producer.Close()
		})

		Convey("SASL 未启用创建生产者", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				SASL: &SASLConfig{
					Enabled:   false,
					Mechanism: "PLAIN",
					Username:  "user",
					Password:  "pass",
				},
			}

			producer, err := NewProducer(cfg)

			So(err, ShouldBeNil)
			So(producer, ShouldNotBeNil)

			// 清理
			producer.Close()
		})

		Convey("不支持的 SASL 机制返回错误", func() {
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

			producer, err := NewProducer(cfg)

			So(err, ShouldNotBeNil)
			So(producer, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "构建 SASL 认证失败")
		})
	})
}

func TestProducer_Close(t *testing.T) {
	Convey("TestProducer_Close", t, func() {
		Convey("writer 为 nil 时关闭返回 nil", func() {
			producer := &Producer{writer: nil}

			err := producer.Close()

			So(err, ShouldBeNil)
		})

		Convey("成功关闭 writer", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			producer, _ := NewProducer(cfg)

			err := producer.Close()

			So(err, ShouldBeNil)
		})
	})
}

func TestProducer_PublishRawEvent(t *testing.T) {
	Convey("TestProducer_PublishRawEvent", t, func() {
		Convey("writer 为 nil 返回错误", func() {
			producer := &Producer{writer: nil}
			ctx := context.Background()

			err := producer.PublishRawEvent(ctx, "key", []byte("value"))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "kafka writer 未初始化")
		})

		Convey("成功发布消息", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			producer, _ := NewProducer(cfg)
			p := producer.(*Producer)
			defer p.Close()

			// 记录写入的消息
			var capturedMessages []kafka.Message

			// 打桩 WriteMessages 成功
			patches := gomonkey.ApplyMethod(p.writer, "WriteMessages",
				func(_ *kafka.Writer, ctx context.Context, msgs ...kafka.Message) error {
					capturedMessages = append(capturedMessages, msgs...)
					return nil
				})
			defer patches.Reset()

			ctx := context.Background()
			err := producer.PublishRawEvent(ctx, "test-key", []byte(`{"event": "test"}`))

			So(err, ShouldBeNil)
			So(len(capturedMessages), ShouldEqual, 1)
			So(string(capturedMessages[0].Key), ShouldEqual, "test-key")
			So(string(capturedMessages[0].Value), ShouldEqual, `{"event": "test"}`)
		})

		Convey("发布空 key 的消息", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			producer, _ := NewProducer(cfg)
			p := producer.(*Producer)
			defer p.Close()

			var capturedMessages []kafka.Message
			patches := gomonkey.ApplyMethod(p.writer, "WriteMessages",
				func(_ *kafka.Writer, ctx context.Context, msgs ...kafka.Message) error {
					capturedMessages = append(capturedMessages, msgs...)
					return nil
				})
			defer patches.Reset()

			ctx := context.Background()
			err := producer.PublishRawEvent(ctx, "", []byte("value"))

			So(err, ShouldBeNil)
			So(len(capturedMessages), ShouldEqual, 1)
			So(string(capturedMessages[0].Key), ShouldEqual, "")
		})

		Convey("WriteMessages 失败返回错误", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			producer, _ := NewProducer(cfg)
			p := producer.(*Producer)
			defer p.Close()

			// 打桩 WriteMessages 失败
			writeErr := errors.New("write failed")
			patches := gomonkey.ApplyMethod(p.writer, "WriteMessages",
				func(_ *kafka.Writer, ctx context.Context, msgs ...kafka.Message) error {
					return writeErr
				})
			defer patches.Reset()

			ctx := context.Background()
			err := producer.PublishRawEvent(ctx, "key", []byte("value"))

			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, writeErr)
		})

		Convey("context 取消时发布失败", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			}
			producer, _ := NewProducer(cfg)
			p := producer.(*Producer)
			defer p.Close()

			// 打桩 WriteMessages 返回 context 错误
			patches := gomonkey.ApplyMethod(p.writer, "WriteMessages",
				func(_ *kafka.Writer, ctx context.Context, msgs ...kafka.Message) error {
					return context.Canceled
				})
			defer patches.Reset()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			err := producer.PublishRawEvent(ctx, "key", []byte("value"))

			So(err, ShouldNotBeNil)
			So(errors.Is(err, context.Canceled), ShouldBeTrue)
		})
	})
}
