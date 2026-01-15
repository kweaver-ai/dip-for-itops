package opensearch

import (
	"context"
	"io"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewProblemStore(t *testing.T) {
	Convey("TestNewProblemStore", t, func() {
		Convey("成功创建 ProblemStore", func() {
			client := newMockClient(200, `{}`)
			store := NewProblemStore(client)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldEqual, client)
		})

		Convey("使用 nil client 创建", func() {
			store := NewProblemStore(nil)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldBeNil)
		})
	})
}

func TestProblemStore_Upsert(t *testing.T) {
	Convey("TestProblemStore_Upsert", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}
			p := domain.Problem{ProblemID: 1}

			err := store.Upsert(ctx, p)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("problem_id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewProblemStore(client)
			p := domain.Problem{ProblemID: 0}

			err := store.Upsert(ctx, p)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "problem_id 不能为空")
		})

		Convey("成功写入 Problem", func() {
			client := newMockClient(201, `{"result": "created"}`)
			store := NewProblemStore(client)

			now := time.Now()
			p := domain.Problem{
				ProblemID:              1,
				ProblemOccurTime:       now,
				ProblemCreateTimestamp: now,
			}

			err := store.Upsert(ctx, p)

			So(err, ShouldBeNil)
		})

		Convey("写入失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewProblemStore(client)

			p := domain.Problem{ProblemID: 1, ProblemOccurTime: time.Now()}

			err := store.Upsert(ctx, p)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "写入 Problem 失败")
		})

		Convey("响应状态码非 2xx 返回错误", func() {
			client := newMockClient(500, `{"error": {"type": "internal_error", "reason": "server error"}}`)
			store := NewProblemStore(client)

			p := domain.Problem{ProblemID: 1, ProblemOccurTime: time.Now()}

			err := store.Upsert(ctx, p)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestProblemStore_QueryByIDs(t *testing.T) {
	Convey("TestProblemStore_QueryByIDs", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			result, err := store.QueryByIDs(ctx, []uint64{1, 2})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("ids 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewProblemStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{})

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功查询 Problem", func() {
			body := `{
				"docs": [
					{"found": true, "_source": {"problem_id": 1, "problem_name": "problem1"}},
					{"found": true, "_source": {"problem_id": 2, "problem_name": "problem2"}}
				]
			}`
			client := newMockClient(200, body)
			store := NewProblemStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{1, 2})

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewProblemStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{1})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "查询 Problem 失败")
		})
	})
}

func TestProblemStore_FindCorrelated(t *testing.T) {
	Convey("TestProblemStore_FindCorrelated", t, func() {
		ctx := context.Background()
		fp := domain.FaultPointObject{FaultID: 1}

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			result, err := store.FindCorrelated(ctx, fp, time.Now())

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功查询关联问题", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"problem_id": 1, "problem_status": "open"}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewProblemStore(client)

			result, err := store.FindCorrelated(ctx, fp, time.Now())

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewProblemStore(client)

			result, err := store.FindCorrelated(ctx, fp, time.Now())

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestProblemStore_FindPendingRCA(t *testing.T) {
	Convey("TestProblemStore_FindPendingRCA", t, func() {
		ctx := context.Background()
		maxAge := 24 * time.Hour

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			result, err := store.FindPendingRCA(ctx, maxAge)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功查询待 RCA 问题", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"problem_id": 1, "problem_status": "open"}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewProblemStore(client)

			result, err := store.FindPendingRCA(ctx, maxAge)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewProblemStore(client)

			result, err := store.FindPendingRCA(ctx, maxAge)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestProblemStore_FindExpiredOpen(t *testing.T) {
	Convey("TestProblemStore_FindExpiredOpen", t, func() {
		ctx := context.Background()
		expirationTime := time.Now().Add(-24 * time.Hour)

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			result, err := store.FindExpiredOpen(ctx, expirationTime)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功查询过期问题", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"problem_id": 1, "problem_status": "open"}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewProblemStore(client)

			result, err := store.FindExpiredOpen(ctx, expirationTime)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewProblemStore(client)

			result, err := store.FindExpiredOpen(ctx, expirationTime)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestProblemStore_UpdateRootCause(t *testing.T) {
	Convey("TestProblemStore_UpdateRootCause", t, func() {
		ctx := context.Background()

		Convey("RCA 状态非成功返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewProblemStore(client)

			cb := domain.RCACallback{
				RcaStatus: domain.RcaStatusFailed,
			}

			err := store.UpdateRootCause(ctx, 1, cb)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rca 状态失败")
		})

		Convey("成功更新根因", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewProblemStore(client)

			cb := domain.RCACallback{
				RootCauseObjectID:  "entity1",
				RootCauseFaultID:   100,
				RcaStatus:          domain.RcaStatusSuccess,
				RcaStartTime:       time.Now(),
				RcaEndTime:         time.Now(),
				ProblemName:        "test problem",
				ProblemDescription: "test description",
			}

			err := store.UpdateRootCause(ctx, 1, cb)

			So(err, ShouldBeNil)
		})
	})
}

func TestProblemStore_UpdateRootCauseObjectID(t *testing.T) {
	Convey("TestProblemStore_UpdateRootCauseObjectID", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			err := store.UpdateRootCauseObjectID(ctx, 1, "entity1", 100)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功更新根因对象 ID", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewProblemStore(client)

			err := store.UpdateRootCauseObjectID(ctx, 1, "entity1", 100)

			So(err, ShouldBeNil)
		})
	})
}

func TestProblemStore_UpdateRelationEventIDs(t *testing.T) {
	Convey("TestProblemStore_UpdateRelationEventIDs", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			err := store.UpdateRelationEventIDs(ctx, 1, []uint64{100, 200})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功更新关联事件 IDs", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewProblemStore(client)

			err := store.UpdateRelationEventIDs(ctx, 1, []uint64{100, 200})

			So(err, ShouldBeNil)
		})
	})
}

func TestProblemStore_MarkClosed(t *testing.T) {
	Convey("TestProblemStore_MarkClosed", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			err := store.MarkClosed(ctx, 1, domain.ProblemCloseTypeManual, domain.ProblemStatusClosed, 3600, "test notes", "admin")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功标记为关闭", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewProblemStore(client)

			err := store.MarkClosed(ctx, 1, domain.ProblemCloseTypeManual, domain.ProblemStatusClosed, 3600, "test notes", "admin")

			So(err, ShouldBeNil)
		})

		Convey("成功标记为关闭（无持续时间）", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewProblemStore(client)

			err := store.MarkClosed(ctx, 1, domain.ProblemCloseTypeManual, domain.ProblemStatusClosed, 0, "test notes", "admin")

			So(err, ShouldBeNil)
		})
	})
}

func TestProblemStore_MarkExpired(t *testing.T) {
	Convey("TestProblemStore_MarkExpired", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			err := store.MarkExpired(ctx, 1)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("问题不存在返回错误", func() {
			body := `{"docs": []}`
			client := newMockClient(200, body)
			store := NewProblemStore(client)

			err := store.MarkExpired(ctx, 999)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "不存在")
		})
	})
}

func TestProblemStore_ClearMergedProblemData(t *testing.T) {
	Convey("TestProblemStore_ClearMergedProblemData", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			err := store.ClearMergedProblemData(ctx, 1)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功清空合并问题数据", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewProblemStore(client)

			err := store.ClearMergedProblemData(ctx, 1)

			So(err, ShouldBeNil)
		})
	})
}

func TestProblemStore_partialUpdate(t *testing.T) {
	Convey("TestProblemStore_partialUpdate", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &ProblemStore{client: nil}

			err := store.partialUpdate(ctx, 1, map[string]any{"status": "updated"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("id 为 0 返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewProblemStore(client)

			err := store.partialUpdate(ctx, 0, map[string]any{"status": "updated"})

			So(err, ShouldBeNil)
		})

		Convey("doc 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewProblemStore(client)

			err := store.partialUpdate(ctx, 1, map[string]any{})

			So(err, ShouldBeNil)
		})

		Convey("成功部分更新", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewProblemStore(client)

			err := store.partialUpdate(ctx, 1, map[string]any{"problem_name": "updated"})

			So(err, ShouldBeNil)
		})

		Convey("更新失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewProblemStore(client)

			err := store.partialUpdate(ctx, 1, map[string]any{"status": "updated"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "更新 Problem")
		})
	})
}
