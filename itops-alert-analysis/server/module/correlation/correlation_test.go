package correlation

import (
	"context"
	"errors"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/kafka"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
)

func newTestConfig() *config.Config {
	return &config.Config{
		AppConfig: config.AppConfig{
			FaultPoint: config.FaultPointExpirationCfg{
				Expiration: config.LocalExpirationConfig{
					Enabled:        true,
					ExpirationTime: 1 * time.Hour,
				},
			},
			Problem: config.ProblemExpirationCfg{
				Expiration: config.LocalExpirationConfig{
					Enabled:        true,
					ExpirationTime: 2 * time.Hour,
				},
			},
		},
	}
}

// newTestConfigManager 创建测试用 ConfigManager
func newTestConfigManager() *config.ConfigManager {
	return config.NewTestConfigManager(newTestConfig())
}

func TestService_Close(t *testing.T) {
	Convey("TestService_Close", t, func() {
		Convey("关闭所有资源成功", func() {
			producer := &kafka.Producer{}
			consumer := &kafka.Consumer{}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethod(producer, "Close", func(_ *kafka.Producer) error {
				return nil
			})
			patches.ApplyMethod(consumer, "Close", func(_ *kafka.Consumer) error {
				return nil
			})

			service := &Service{
				kafkaProducer: producer,
				kafkaConsumer: consumer,
			}

			err := service.Close()

			So(err, ShouldBeNil)
		})

		Convey("关闭 producer 失败", func() {
			producer := &kafka.Producer{}
			consumer := &kafka.Consumer{}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethod(producer, "Close", func(_ *kafka.Producer) error {
				return errors.New("producer close error")
			})
			patches.ApplyMethod(consumer, "Close", func(_ *kafka.Consumer) error {
				return nil
			})

			service := &Service{
				kafkaProducer: producer,
				kafkaConsumer: consumer,
			}

			err := service.Close()

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "close kafkaProducer")
		})

		Convey("关闭 consumer 失败", func() {
			producer := &kafka.Producer{}
			consumer := &kafka.Consumer{}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethod(producer, "Close", func(_ *kafka.Producer) error {
				return nil
			})
			patches.ApplyMethod(consumer, "Close", func(_ *kafka.Consumer) error {
				return errors.New("consumer close error")
			})

			service := &Service{
				kafkaProducer: producer,
				kafkaConsumer: consumer,
			}

			err := service.Close()

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "close kafkaConsumer")
		})

		Convey("所有资源为 nil 时正常关闭", func() {
			service := &Service{}

			err := service.Close()

			So(err, ShouldBeNil)
		})

		Convey("多个资源关闭失败", func() {
			producer := &kafka.Producer{}
			consumer := &kafka.Consumer{}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethod(producer, "Close", func(_ *kafka.Producer) error {
				return errors.New("producer error")
			})
			patches.ApplyMethod(consumer, "Close", func(_ *kafka.Consumer) error {
				return errors.New("consumer error")
			})

			service := &Service{
				kafkaProducer: producer,
				kafkaConsumer: consumer,
			}

			err := service.Close()

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "2 个错误")
		})
	})
}

func TestService_HandleRCACallback(t *testing.T) {
	Convey("TestService_HandleRCACallback", t, func() {
		Convey("InProgress 为 true 时直接返回", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			problemStage := NewProblemStage(cfgManager, factory, nil, nil)

			service := &Service{
				problem: problemStage,
			}

			cb := domain.RCACallback{
				ProblemID:  12345,
				InProgress: true,
			}

			err := service.HandleRCACallback(context.Background(), cb)

			So(err, ShouldBeNil)
		})
	})
}
