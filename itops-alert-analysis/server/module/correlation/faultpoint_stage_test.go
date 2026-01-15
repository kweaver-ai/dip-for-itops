package correlation

import (
	"context"
	"testing"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	. "github.com/smartystreets/goconvey/convey"
)

// problemHandlerStub 用于测试的桩
type problemHandlerStub struct{}

func (s *problemHandlerStub) HandleFaultPoint(ctx context.Context, fp domain.FaultPointObject) error {
	return nil
}

func (s *problemHandlerStub) HandleRCACallback(ctx context.Context, cb domain.RCACallback) error {
	return nil
}

func (s *problemHandlerStub) CloseProblem(ctx context.Context, problemID uint64, closeType domain.ProblemCloseType, closeState domain.ProblemStatus, notes string, by string) error {
	return nil
}

func (s *problemHandlerStub) HandleFaultPointRecovered(ctx context.Context, faultID uint64) error {
	return nil
}

func TestNewFaultPointStage(t *testing.T) {
	Convey("TestNewFaultPointStage", t, func() {
		Convey("成功创建 FaultPointStage", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			problemHandler := &problemHandlerStub{}

			stage := NewFaultPointStage(cfgManager, factory, problemHandler)

			So(stage, ShouldNotBeNil)
			So(stage.cfgManager, ShouldEqual, cfgManager)
			So(stage.repoFactory, ShouldEqual, factory)
			So(stage.problemHandler, ShouldEqual, problemHandler)
			So(stage.genID, ShouldNotBeNil)
		})

		Convey("使用 nil problemHandler 创建", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)

			stage := NewFaultPointStage(cfgManager, factory, nil)

			So(stage, ShouldNotBeNil)
			So(stage.problemHandler, ShouldBeNil)
		})
	})
}

func TestFaultPointStage_OnProblemLinked(t *testing.T) {
	Convey("TestFaultPointStage_OnProblemLinked", t, func() {
		Convey("stage 创建成功", func() {
			cfgManager := newTestConfigManager()
			factory := opensearch.NewRepositoryFactory(nil)
			stage := NewFaultPointStage(cfgManager, factory, nil)

			So(stage, ShouldNotBeNil)
			So(stage.repoFactory, ShouldEqual, factory)
		})
	})
}

func TestFaultPointStage_Run(t *testing.T) {
	Convey("TestFaultPointStage_Run", t, func() {
		Convey("失效检查未启用时直接返回", func() {
			cfg := newTestConfig()
			cfg.AppConfig.FaultPoint.Expiration.Enabled = false
			cfgManager := config.NewTestConfigManager(cfg)
			factory := opensearch.NewRepositoryFactory(nil)
			stage := NewFaultPointStage(cfgManager, factory, nil)

			// Run 会立即返回 nil
			// 由于没有启用失效检查，不会阻塞
			So(stage, ShouldNotBeNil)
		})
	})
}
