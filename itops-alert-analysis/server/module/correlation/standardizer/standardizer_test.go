package standardizer

import (
	"context"
	"testing"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/correlation/objectclass"
	. "github.com/smartystreets/goconvey/convey"
)

// mockStandardizer 用于测试的 mock 标准化器
type mockStandardizer struct {
	name string
}

func (m *mockStandardizer) Standardize(ctx context.Context, payload []byte) (domain.RawEvent, error) {
	return domain.RawEvent{EventTitle: m.name}, nil
}

func TestNewRegistry(t *testing.T) {
	Convey("TestNewRegistry", t, func() {
		Convey("创建空注册表", func() {
			registry := NewRegistry()

			So(registry, ShouldNotBeNil)
			So(registry.factories, ShouldNotBeNil)
			So(len(registry.factories), ShouldEqual, 0)
		})
	})
}

func TestRegistry_Register(t *testing.T) {
	Convey("TestRegistry_Register", t, func() {
		Convey("成功注册标准化器", func() {
			registry := NewRegistry()
			factory := func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "test"}, nil
			}

			registry.Register("test_source", factory)

			So(len(registry.factories), ShouldEqual, 1)
			So(registry.factories["test_source"], ShouldNotBeNil)
		})

		Convey("注册时自动转换为小写", func() {
			registry := NewRegistry()
			factory := func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "test"}, nil
			}

			registry.Register("TEST_SOURCE", factory)

			So(registry.factories["test_source"], ShouldNotBeNil)
			So(registry.factories["TEST_SOURCE"], ShouldBeNil)
		})

		Convey("注册时自动去除空格", func() {
			registry := NewRegistry()
			factory := func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "test"}, nil
			}

			registry.Register("  test_source  ", factory)

			So(registry.factories["test_source"], ShouldNotBeNil)
		})

		Convey("空 source 不注册", func() {
			registry := NewRegistry()
			factory := func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "test"}, nil
			}

			registry.Register("", factory)

			So(len(registry.factories), ShouldEqual, 0)
		})

		Convey("只有空格的 source 不注册", func() {
			registry := NewRegistry()
			factory := func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "test"}, nil
			}

			registry.Register("   ", factory)

			So(len(registry.factories), ShouldEqual, 0)
		})

		Convey("nil factory 不注册", func() {
			registry := NewRegistry()

			registry.Register("test_source", nil)

			So(len(registry.factories), ShouldEqual, 0)
		})

		Convey("覆盖已有注册", func() {
			registry := NewRegistry()
			factory1 := func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "first"}, nil
			}
			factory2 := func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "second"}, nil
			}

			registry.Register("test_source", factory1)
			registry.Register("test_source", factory2)

			So(len(registry.factories), ShouldEqual, 1)
			// 验证被覆盖
			standardizer, _ := registry.factories["test_source"](nil, nil)
			So(standardizer.(*mockStandardizer).name, ShouldEqual, "second")
		})
	})
}

func TestRegistry_Resolve(t *testing.T) {
	Convey("TestRegistry_Resolve", t, func() {
		Convey("成功解析已注册的标准化器", func() {
			registry := NewRegistry()
			registry.Register("test_source", func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "resolved"}, nil
			})

			cfg := &config.Config{
				AppConfig: config.AppConfig{
					Ingest: config.IngestConfig{
						Source: config.Source{Type: "test_source"},
					},
				},
			}

			standardizer, err := registry.Resolve(cfg, nil)

			So(err, ShouldBeNil)
			So(standardizer, ShouldNotBeNil)
			So(standardizer.(*mockStandardizer).name, ShouldEqual, "resolved")
		})

		Convey("解析时忽略大小写", func() {
			registry := NewRegistry()
			registry.Register("test_source", func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "resolved"}, nil
			})

			cfg := &config.Config{
				AppConfig: config.AppConfig{
					Ingest: config.IngestConfig{
						Source: config.Source{Type: "TEST_SOURCE"},
					},
				},
			}

			standardizer, err := registry.Resolve(cfg, nil)

			So(err, ShouldBeNil)
			So(standardizer, ShouldNotBeNil)
		})

		Convey("解析时去除空格", func() {
			registry := NewRegistry()
			registry.Register("test_source", func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
				return &mockStandardizer{name: "resolved"}, nil
			})

			cfg := &config.Config{
				AppConfig: config.AppConfig{
					Ingest: config.IngestConfig{
						Source: config.Source{Type: "  test_source  "},
					},
				},
			}

			standardizer, err := registry.Resolve(cfg, nil)

			So(err, ShouldBeNil)
			So(standardizer, ShouldNotBeNil)
		})

		Convey("未注册的 source type 返回错误", func() {
			registry := NewRegistry()

			cfg := &config.Config{
				AppConfig: config.AppConfig{
					Ingest: config.IngestConfig{
						Source: config.Source{Type: "unknown_source"},
					},
				},
			}

			standardizer, err := registry.Resolve(cfg, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unsupported source type")
			So(err.Error(), ShouldContainSubstring, "unknown_source")
			So(standardizer, ShouldBeNil)
		})

		Convey("空 source type 返回错误", func() {
			registry := NewRegistry()

			cfg := &config.Config{
				AppConfig: config.AppConfig{
					Ingest: config.IngestConfig{
						Source: config.Source{Type: ""},
					},
				},
			}

			standardizer, err := registry.Resolve(cfg, nil)

			So(err, ShouldNotBeNil)
			So(standardizer, ShouldBeNil)
		})
	})
}

func TestBuild(t *testing.T) {
	Convey("TestBuild", t, func() {
		Convey("成功构建 zabbix_webhook 标准化器", func() {
			querier := &mockObjectClassQuerier{
				result: &objectclass.EntityObjectInfo{
					ObjectTypeID: "ot-test",
					ObjectID:     "obj-test",
					Name:         "test",
				},
			}

			cfg := &config.Config{
				AppConfig: config.AppConfig{
					Ingest: config.IngestConfig{
						Source: config.Source{Type: "zabbix_webhook"},
					},
				},
			}

			standardizer, err := Build(cfg, querier)

			So(err, ShouldBeNil)
			So(standardizer, ShouldNotBeNil)
		})

		Convey("zabbix_webhook 大小写不敏感", func() {
			querier := &mockObjectClassQuerier{
				result: &objectclass.EntityObjectInfo{},
			}

			cfg := &config.Config{
				AppConfig: config.AppConfig{
					Ingest: config.IngestConfig{
						Source: config.Source{Type: "ZABBIX_WEBHOOK"},
					},
				},
			}

			standardizer, err := Build(cfg, querier)

			So(err, ShouldBeNil)
			So(standardizer, ShouldNotBeNil)
		})

		Convey("不支持的 source type 返回错误", func() {
			cfg := &config.Config{
				AppConfig: config.AppConfig{
					Ingest: config.IngestConfig{
						Source: config.Source{Type: "prometheus"},
					},
				},
			}

			standardizer, err := Build(cfg, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unsupported source type")
			So(standardizer, ShouldBeNil)
		})
	})
}
