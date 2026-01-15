package correlation

import (
	"context"
	"errors"
	"testing"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/kafka"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/correlation/standardizer"
	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewIngestStage(t *testing.T) {
	Convey("TestNewIngestStage", t, func() {
		Convey("成功创建 IngestStage", func() {
			factory := opensearch.NewRepositoryFactory(nil)

			stage := NewIngestStage(factory, nil, nil, nil)

			So(stage, ShouldNotBeNil)
			So(stage.repoFactory, ShouldEqual, factory)
		})
	})
}

func TestIngestStage_Start(t *testing.T) {
	Convey("TestIngestStage_Start", t, func() {
		factory := opensearch.NewRepositoryFactory(nil)

		Convey("consumer 未配置返回错误", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			// 打桩 standardizer.Build 返回一个标准化器
			patches.ApplyFunc(standardizer.Build, func(cfg *config.Config, querier standardizer.ObjectClassQuerier) (standardizer.Standardizer, error) {
				return &zabbixStandardizerStub{}, nil
			})

			stage := NewIngestStage(factory, nil, &zabbixStandardizerStub{}, nil)

			err := stage.Start(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "kafka rawEventsConsumer not configured")
		})

		Convey("standardizer 未配置返回错误", func() {
			consumer := &kafka.Consumer{}
			stage := NewIngestStage(factory, nil, nil, consumer)

			err := stage.Start(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "standardizer not configured")
		})
	})
}

// zabbixStandardizerStub 用于测试的桩
type zabbixStandardizerStub struct{}

func (s *zabbixStandardizerStub) Standardize(ctx context.Context, payload []byte) (domain.RawEvent, error) {
	return domain.RawEvent{}, errors.New("stub")
}
