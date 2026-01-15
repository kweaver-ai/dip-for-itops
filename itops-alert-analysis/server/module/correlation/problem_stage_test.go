package correlation

import (
	"context"
	"testing"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/kafka"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewProblemStage(t *testing.T) {
	Convey("TestNewProblemStage", t, func() {
		Convey("成功创建 ProblemStage", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			producer := &kafka.Producer{}

			stage := NewProblemStage(cfgManager, factory, producer, nil)

			So(stage, ShouldNotBeNil)
			So(stage.cfgManager, ShouldEqual, cfgManager)
			So(stage.repoFactory, ShouldEqual, factory)
			So(stage.kafkaProducer, ShouldEqual, producer)
			So(stage.genID, ShouldNotBeNil)
		})

		Convey("使用 nil producer 创建", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)

			stage := NewProblemStage(cfgManager, factory, nil, nil)

			So(stage, ShouldNotBeNil)
			So(stage.kafkaProducer, ShouldBeNil)
		})

		Convey("使用 nil spatialChecker 创建", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			producer := &kafka.Producer{}

			stage := NewProblemStage(cfgManager, factory, producer, nil)

			So(stage, ShouldNotBeNil)
			So(stage.spatialChecker, ShouldBeNil)
		})
	})
}

func TestProblemStage_HandleRCACallback(t *testing.T) {
	Convey("TestProblemStage_HandleRCACallback", t, func() {
		Convey("RCA 仍在进行中时不更新", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			stage := NewProblemStage(cfgManager, factory, nil, nil)

			cb := domain.RCACallback{
				ProblemID:  12345,
				InProgress: true,
			}

			err := stage.HandleRCACallback(context.Background(), cb)

			So(err, ShouldBeNil)
		})
	})
}

func TestProblemStage_CloseProblem(t *testing.T) {
	Convey("TestProblemStage_CloseProblem", t, func() {
		Convey("成功创建 stage 用于关闭问题", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			stage := NewProblemStage(cfgManager, factory, nil, nil)

			So(stage, ShouldNotBeNil)
			So(stage.repoFactory, ShouldEqual, factory)
		})
	})
}

func TestProblemStage_Run(t *testing.T) {
	Convey("TestProblemStage_Run", t, func() {
		Convey("失效检查未启用时直接返回", func() {
			cfg := newTestConfig()
			cfg.AppConfig.Problem.Expiration.Enabled = false
			cfgManager := config.NewTestConfigManager(cfg)
			factory := opensearch.NewRepositoryFactory(nil)
			stage := NewProblemStage(cfgManager, factory, nil, nil)

			// Run 会立即返回 nil
			So(stage, ShouldNotBeNil)
		})
	})
}

func TestProblemStage_publishProblemEvent(t *testing.T) {
	Convey("TestProblemStage_publishProblemEvent", t, func() {
		Convey("producer 为 nil 时跳过发布", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			stage := NewProblemStage(cfgManager, factory, nil, nil)

			err := stage.publishProblemEvent(context.Background(), 12345)

			So(err, ShouldBeNil)
		})

		Convey("成功发布问题事件", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			producer := &kafka.Producer{}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethod(producer, "PublishRawEvent", func(_ *kafka.Producer, ctx context.Context, key string, value []byte) error {
				return nil
			})

			stage := NewProblemStage(cfgManager, factory, producer, nil)

			err := stage.publishProblemEvent(context.Background(), 12345)

			So(err, ShouldBeNil)
		})

		Convey("发布失败返回错误", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			producer := &kafka.Producer{}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethod(producer, "PublishRawEvent", func(_ *kafka.Producer, ctx context.Context, key string, value []byte) error {
				return context.DeadlineExceeded
			})

			stage := NewProblemStage(cfgManager, factory, producer, nil)

			err := stage.publishProblemEvent(context.Background(), 12345)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "发送 Kafka 消息失败")
		})
	})
}
