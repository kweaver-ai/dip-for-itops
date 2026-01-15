package dip

import (
	"context"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

// mockGetAuth 用于测试的获取 Authorization 函数
func mockGetAuth() string {
	return "Bearer test-token"
}

// mockGetKnID 用于测试的获取 KnowledgeID 函数
func mockGetKnID() string {
	return "test-kn"
}

func TestNewSpatialChecker(t *testing.T) {
	Convey("TestNewSpatialChecker", t, func() {
		Convey("创建空间相关性检查器", func() {
			cfg := config.DIPConfig{
				Host:    "https://example.com",
				KnID:    "test-kn",
				Timeout: 10 * time.Second,
			}
			client := NewClient(cfg, mockGetAuth, mockGetKnID)
			checker := NewSpatialChecker(client)

			So(checker, ShouldNotBeNil)
			So(checker.client, ShouldEqual, client)
		})

		Convey("使用 nil client 创建检查器", func() {
			checker := NewSpatialChecker(nil)

			So(checker, ShouldNotBeNil)
			So(checker.client, ShouldBeNil)
		})
	})
}

func TestSpatialChecker_FilterCorrelatedProblems(t *testing.T) {
	Convey("TestSpatialChecker_FilterCorrelatedProblems", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			checker := &SpatialChecker{client: nil}
			fp := domain.FaultPointObject{FaultID: 1}
			problems := []domain.Problem{{ProblemID: 1}}

			result, err := checker.FilterCorrelatedProblems(ctx, fp, problems)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "DIP 客户端未配置")
		})

		Convey("problems 为空返回 nil", func() {
			cfg := config.DIPConfig{Host: "https://example.com", KnID: "test-kn"}
			client := NewClient(cfg, mockGetAuth, mockGetKnID)
			checker := NewSpatialChecker(client)
			fp := domain.FaultPointObject{FaultID: 1}

			result, err := checker.FilterCorrelatedProblems(ctx, fp, nil)

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("故障点缺少 EntityObjectID 返回错误", func() {
			cfg := config.DIPConfig{Host: "https://example.com", KnID: "test-kn"}
			client := NewClient(cfg, mockGetAuth, mockGetKnID)
			checker := NewSpatialChecker(client)
			fp := domain.FaultPointObject{
				FaultID:        1,
				EntityObjectID: "", // 缺少 EntityObjectID
			}
			problems := []domain.Problem{{ProblemID: 1}}

			result, err := checker.FilterCorrelatedProblems(ctx, fp, problems)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "故障点缺少 EntityObjectID")
		})

		Convey("查询子图失败返回错误", func() {
			cfg := config.DIPConfig{Host: "https://example.com", KnID: "test-kn"}
			client := NewClient(cfg, mockGetAuth, mockGetKnID)
			checker := NewSpatialChecker(client)

			// 打桩 QuerySubGraph 返回错误
			patches := gomonkey.ApplyMethod(client, "QuerySubGraph",
				func(_ *Client, _ context.Context, _ SubGraphQueryRequest) (*SubGraphResponse, error) {
					return nil, errors.New("query failed")
				})
			defer patches.Reset()

			fp := domain.FaultPointObject{
				FaultID:           1,
				EntityObjectID:    "entity-1",
				EntityObjectClass: "Server",
			}
			problems := []domain.Problem{{ProblemID: 1}}

			result, err := checker.FilterCorrelatedProblems(ctx, fp, problems)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "查询子图失败")
		})

		Convey("成功过滤相关问题", func() {
			cfg := config.DIPConfig{Host: "https://example.com", KnID: "test-kn"}
			client := NewClient(cfg, mockGetAuth, mockGetKnID)
			checker := NewSpatialChecker(client)

			// 打桩 QuerySubGraph 返回子图数据
			patches := gomonkey.ApplyMethod(client, "QuerySubGraph",
				func(_ *Client, _ context.Context, _ SubGraphQueryRequest) (*SubGraphResponse, error) {
					return &SubGraphResponse{
						Objects: map[string]SubGraphObject{
							"obj-1": {
								ID:           "obj-1",
								ObjectTypeID: "Server",
								Properties:   SubGraphObjectProperties{SID: "entity-1"},
							},
							"obj-2": {
								ID:           "obj-2",
								ObjectTypeID: "Application",
								Properties:   SubGraphObjectProperties{SID: "entity-2"},
							},
						},
					}, nil
				})
			defer patches.Reset()

			fp := domain.FaultPointObject{
				FaultID:           1,
				EntityObjectID:    "entity-1",
				EntityObjectClass: "Server",
			}
			problems := []domain.Problem{
				{ProblemID: 1, AffectedEntityIDs: []string{"entity-1"}},             // 空间相关
				{ProblemID: 2, AffectedEntityIDs: []string{"entity-2"}},             // 空间相关
				{ProblemID: 3, AffectedEntityIDs: []string{"entity-3"}},             // 不相关
				{ProblemID: 4, AffectedEntityIDs: []string{"entity-1", "entity-3"}}, // 空间相关
			}

			result, err := checker.FilterCorrelatedProblems(ctx, fp, problems)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result), ShouldEqual, 3)
			So(result[0].ProblemID, ShouldEqual, 1)
			So(result[1].ProblemID, ShouldEqual, 2)
			So(result[2].ProblemID, ShouldEqual, 4)
		})

		Convey("没有空间相关的问题", func() {
			cfg := config.DIPConfig{Host: "https://example.com", KnID: "test-kn"}
			client := NewClient(cfg, mockGetAuth, mockGetKnID)
			checker := NewSpatialChecker(client)

			// 打桩 QuerySubGraph 返回子图数据
			patches := gomonkey.ApplyMethod(client, "QuerySubGraph",
				func(_ *Client, _ context.Context, _ SubGraphQueryRequest) (*SubGraphResponse, error) {
					return &SubGraphResponse{
						Objects: map[string]SubGraphObject{
							"obj-1": {
								ID:           "obj-1",
								ObjectTypeID: "Server",
								Properties:   SubGraphObjectProperties{SID: "entity-1"},
							},
						},
					}, nil
				})
			defer patches.Reset()

			fp := domain.FaultPointObject{
				FaultID:           1,
				EntityObjectID:    "entity-1",
				EntityObjectClass: "Server",
			}
			problems := []domain.Problem{
				{ProblemID: 1, AffectedEntityIDs: []string{"entity-99"}},  // 不相关
				{ProblemID: 2, AffectedEntityIDs: []string{"entity-100"}}, // 不相关
			}

			result, err := checker.FilterCorrelatedProblems(ctx, fp, problems)

			So(err, ShouldBeNil)
			So(result, ShouldBeNil) // 空切片 append 后为 nil
		})

		Convey("子图返回空对象", func() {
			cfg := config.DIPConfig{Host: "https://example.com", KnID: "test-kn"}
			client := NewClient(cfg, mockGetAuth, mockGetKnID)
			checker := NewSpatialChecker(client)

			// 打桩 QuerySubGraph 返回空子图
			patches := gomonkey.ApplyMethod(client, "QuerySubGraph",
				func(_ *Client, _ context.Context, _ SubGraphQueryRequest) (*SubGraphResponse, error) {
					return &SubGraphResponse{
						Objects: map[string]SubGraphObject{},
					}, nil
				})
			defer patches.Reset()

			fp := domain.FaultPointObject{
				FaultID:           1,
				EntityObjectID:    "entity-1",
				EntityObjectClass: "Server",
			}
			problems := []domain.Problem{
				{ProblemID: 1, AffectedEntityIDs: []string{"entity-1"}},
			}

			result, err := checker.FilterCorrelatedProblems(ctx, fp, problems)

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("子图对象缺少 SID 属性", func() {
			cfg := config.DIPConfig{Host: "https://example.com", KnID: "test-kn"}
			client := NewClient(cfg, mockGetAuth, mockGetKnID)
			checker := NewSpatialChecker(client)

			// 打桩 QuerySubGraph 返回缺少 SID 的对象
			patches := gomonkey.ApplyMethod(client, "QuerySubGraph",
				func(_ *Client, _ context.Context, _ SubGraphQueryRequest) (*SubGraphResponse, error) {
					return &SubGraphResponse{
						Objects: map[string]SubGraphObject{
							"obj-1": {
								ID:           "obj-1",
								ObjectTypeID: "Server",
								Properties:   SubGraphObjectProperties{SID: ""}, // 缺少 SID
							},
							"obj-2": {
								ID:           "obj-2",
								ObjectTypeID: "Application",
								Properties:   SubGraphObjectProperties{SID: "entity-2"},
							},
						},
					}, nil
				})
			defer patches.Reset()

			fp := domain.FaultPointObject{
				FaultID:           1,
				EntityObjectID:    "entity-1",
				EntityObjectClass: "Server",
			}
			problems := []domain.Problem{
				{ProblemID: 1, AffectedEntityIDs: []string{"entity-2"}}, // 空间相关
				{ProblemID: 2, AffectedEntityIDs: []string{"entity-1"}}, // 不相关（SID 为空被跳过）
			}

			result, err := checker.FilterCorrelatedProblems(ctx, fp, problems)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0].ProblemID, ShouldEqual, 1)
		})
	})
}
