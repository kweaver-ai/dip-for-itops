package opensearch

import (
	"context"
	"io"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewFaultPointStore(t *testing.T) {
	Convey("TestNewFaultPointStore", t, func() {
		Convey("成功创建 FaultPointStore", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldEqual, client)
		})

		Convey("使用 nil client 创建", func() {
			store := NewFaultPointStore(nil)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldBeNil)
		})
	})
}

func TestFaultPointStore_Upsert(t *testing.T) {
	Convey("TestFaultPointStore_Upsert", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}
			fp := domain.FaultPointObject{FaultID: 1}

			err := store.Upsert(ctx, fp)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("fault_id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)
			fp := domain.FaultPointObject{FaultID: 0}

			err := store.Upsert(ctx, fp)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "fault_id 不能为空")
		})

		Convey("成功写入 FaultPointObject", func() {
			client := newMockClient(201, `{"result": "created"}`)
			store := NewFaultPointStore(client)

			now := time.Now()
			fp := domain.FaultPointObject{
				FaultID:         1,
				FaultOccurTime:  now,
				FaultCreateTime: now,
			}

			err := store.Upsert(ctx, fp)

			So(err, ShouldBeNil)
		})

		Convey("写入失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointStore(client)

			fp := domain.FaultPointObject{FaultID: 1, FaultOccurTime: time.Now()}

			err := store.Upsert(ctx, fp)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "写入 FaultPointObject 失败")
		})

		Convey("响应状态码非 2xx 返回错误", func() {
			client := newMockClient(500, `{"error": {"type": "internal_error", "reason": "server error"}}`)
			store := NewFaultPointStore(client)

			fp := domain.FaultPointObject{FaultID: 1, FaultOccurTime: time.Now()}

			err := store.Upsert(ctx, fp)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestFaultPointStore_QueryByIDs(t *testing.T) {
	Convey("TestFaultPointStore_QueryByIDs", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}

			result, err := store.QueryByIDs(ctx, []uint64{1, 2})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("ids 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{})

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功查询 FaultPointObject", func() {
			body := `{
				"docs": [
					{"found": true, "_source": {"fault_id": 1, "fault_mode": "mode1"}},
					{"found": true, "_source": {"fault_id": 2, "fault_mode": "mode2"}}
				]
			}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{1, 2})

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{1})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "mget FaultPointObject 失败")
		})

		Convey("响应错误返回错误", func() {
			client := newMockClient(404, `{"error": {"type": "index_not_found_exception", "reason": "no such index"}}`)
			store := NewFaultPointStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{1})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestFaultPointStore_FindOpenByEntityAndMode(t *testing.T) {
	Convey("TestFaultPointStore_FindOpenByEntityAndMode", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}

			result, err := store.FindOpenByEntityAndMode(ctx, "entity1", "mode1", time.Now())

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("entityObjectID 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			result, err := store.FindOpenByEntityAndMode(ctx, "", "mode1", time.Now())

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "entityObjectID or failureMode is empty")
		})

		Convey("failureMode 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			result, err := store.FindOpenByEntityAndMode(ctx, "entity1", "", time.Now())

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "entityObjectID or failureMode is empty")
		})

		Convey("成功找到故障点", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"fault_id": 1, "fault_mode": "mode1", "entity_object_id": "entity1"}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			result, err := store.FindOpenByEntityAndMode(ctx, "entity1", "mode1", time.Now())

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.FaultID, ShouldEqual, 1)
		})

		Convey("未找到故障点返回 nil", func() {
			body := `{"hits": {"hits": []}}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			result, err := store.FindOpenByEntityAndMode(ctx, "entity1", "mode1", time.Now())

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointStore(client)

			result, err := store.FindOpenByEntityAndMode(ctx, "entity1", "mode1", time.Now())

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestFaultPointStore_UpdateProblemID(t *testing.T) {
	Convey("TestFaultPointStore_UpdateProblemID", t, func() {
		ctx := context.Background()

		Convey("成功批量更新 problem_id", func() {
			body := `{"errors": false, "items": [{"update": {"status": 200}}]}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			err := store.UpdateProblemID(ctx, []uint64{1, 2}, 100)

			So(err, ShouldBeNil)
		})

		Convey("faultIDs 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			err := store.UpdateProblemID(ctx, []uint64{}, 100)

			So(err, ShouldBeNil)
		})
	})
}

func TestFaultPointStore_MakeRecovered(t *testing.T) {
	Convey("TestFaultPointStore_MakeRecovered", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}

			err := store.MakeRecovered(ctx, 1, time.Now())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("faultID 为 0 返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			err := store.MakeRecovered(ctx, 0, time.Now())

			So(err, ShouldBeNil)
		})

		Convey("成功标记为已恢复", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewFaultPointStore(client)

			err := store.MakeRecovered(ctx, 1, time.Now())

			So(err, ShouldBeNil)
		})

		Convey("更新失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointStore(client)

			err := store.MakeRecovered(ctx, 1, time.Now())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "更新 FaultPointObject")
		})
	})
}

func TestFaultPointStore_MakeExpired(t *testing.T) {
	Convey("TestFaultPointStore_MakeExpired", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}

			err := store.MakeExpired(ctx, 1)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("faultID 为 0 返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			err := store.MakeExpired(ctx, 0)

			So(err, ShouldBeNil)
		})

		Convey("成功标记为已失效", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewFaultPointStore(client)

			err := store.MakeExpired(ctx, 1)

			So(err, ShouldBeNil)
		})
	})
}

func TestFaultPointStore_FindByEventID(t *testing.T) {
	Convey("TestFaultPointStore_FindByEventID", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}

			result, err := store.FindByEventID(ctx, 1)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("eventID 为 0 返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			result, err := store.FindByEventID(ctx, 0)

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功找到故障点", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"fault_id": 1, "relation_event_ids": [100, 200]}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			result, err := store.FindByEventID(ctx, 100)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.FaultID, ShouldEqual, 1)
		})

		Convey("未找到故障点返回 nil", func() {
			body := `{"hits": {"hits": []}}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			result, err := store.FindByEventID(ctx, 999)

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointStore(client)

			result, err := store.FindByEventID(ctx, 1)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestFaultPointStore_FindInWindow(t *testing.T) {
	Convey("TestFaultPointStore_FindInWindow", t, func() {
		ctx := context.Background()
		now := time.Now()
		start := now.Add(-time.Hour)
		end := now

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}

			result, err := store.FindInWindow(ctx, "entity1", "mode1", start, end)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("entityID 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			result, err := store.FindInWindow(ctx, "", "mode1", start, end)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "entityID 不能为空")
		})

		Convey("成功查询时间窗口内的故障点", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"fault_id": 1, "entity_object_id": "entity1"}},
						{"_source": {"fault_id": 2, "entity_object_id": "entity1"}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			result, err := store.FindInWindow(ctx, "entity1", "mode1", start, end)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointStore(client)

			result, err := store.FindInWindow(ctx, "entity1", "mode1", start, end)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestFaultPointStore_FindExpiredOccurred(t *testing.T) {
	Convey("TestFaultPointStore_FindExpiredOccurred", t, func() {
		ctx := context.Background()
		expirationTime := time.Now().Add(-24 * time.Hour)

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}

			result, err := store.FindExpiredOccurred(ctx, expirationTime)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功查询过期故障点", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"fault_id": 1, "fault_status": "occurred"}},
						{"_source": {"fault_id": 2, "fault_status": "occurred"}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			result, err := store.FindExpiredOccurred(ctx, expirationTime)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("未找到过期故障点", func() {
			body := `{"hits": {"hits": []}}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			result, err := store.FindExpiredOccurred(ctx, expirationTime)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointStore(client)

			result, err := store.FindExpiredOccurred(ctx, expirationTime)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestFaultPointStore_bulkUpdate(t *testing.T) {
	Convey("TestFaultPointStore_bulkUpdate", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointStore{client: nil}

			err := store.bulkUpdate(ctx, []uint64{1}, map[string]any{"status": "updated"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("faultIDs 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			err := store.bulkUpdate(ctx, []uint64{}, map[string]any{"status": "updated"})

			So(err, ShouldBeNil)
		})

		Convey("doc 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointStore(client)

			err := store.bulkUpdate(ctx, []uint64{1}, map[string]any{})

			So(err, ShouldBeNil)
		})

		Convey("成功批量更新", func() {
			body := `{"errors": false, "items": [{"update": {"status": 200}}]}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			err := store.bulkUpdate(ctx, []uint64{1, 2}, map[string]any{"problem_id": 100})

			So(err, ShouldBeNil)
		})

		Convey("bulk 请求失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointStore(client)

			err := store.bulkUpdate(ctx, []uint64{1}, map[string]any{"status": "updated"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "bulk update 请求失败")
		})

		Convey("bulk 部分失败返回错误", func() {
			body := `{"errors": true, "items": [{"update": {"status": 404}}]}`
			client := newMockClient(200, body)
			store := NewFaultPointStore(client)

			err := store.bulkUpdate(ctx, []uint64{1}, map[string]any{"status": "updated"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "bulk update 部分失败")
		})
	})
}
