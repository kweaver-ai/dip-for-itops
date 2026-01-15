package opensearch

import (
	"context"
	"io"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewFaultCausalStore(t *testing.T) {
	Convey("TestNewFaultCausalStore", t, func() {
		Convey("成功创建 FaultCausalStore", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalStore(client)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldEqual, client)
		})

		Convey("使用 nil client 创建", func() {
			store := NewFaultCausalStore(nil)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldBeNil)
		})
	})
}

func TestFaultCausalStore_Upsert(t *testing.T) {
	Convey("TestFaultCausalStore_Upsert", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalStore{client: nil}
			fc := domain.FaultCausalObject{CausalID: "causal-1"}

			err := store.Upsert(ctx, fc)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("causal_id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalStore(client)
			fc := domain.FaultCausalObject{CausalID: ""}

			err := store.Upsert(ctx, fc)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "causal_id 不能为空")
		})

		Convey("成功写入 FaultCausalObject", func() {
			client := newMockClient(201, `{"result": "created"}`)
			store := NewFaultCausalStore(client)

			now := time.Now()
			fc := domain.FaultCausalObject{
				CausalID:    "causal-1",
				SCreateTime: now,
			}

			err := store.Upsert(ctx, fc)

			So(err, ShouldBeNil)
		})

		Convey("使用默认时间戳成功写入", func() {
			client := newMockClient(201, `{"result": "created"}`)
			store := NewFaultCausalStore(client)

			fc := domain.FaultCausalObject{
				CausalID: "causal-1",
				// SCreateTime 为零值
			}

			err := store.Upsert(ctx, fc)

			So(err, ShouldBeNil)
		})

		Convey("写入失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalStore(client)

			fc := domain.FaultCausalObject{CausalID: "causal-1", SCreateTime: time.Now()}

			err := store.Upsert(ctx, fc)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "写入 FaultCausalObject 失败")
		})

		Convey("响应状态码非 2xx 返回错误", func() {
			client := newMockClient(500, `{"error": {"type": "internal_error", "reason": "server error"}}`)
			store := NewFaultCausalStore(client)

			fc := domain.FaultCausalObject{CausalID: "causal-1", SCreateTime: time.Now()}

			err := store.Upsert(ctx, fc)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestFaultCausalStore_Update(t *testing.T) {
	Convey("TestFaultCausalStore_Update", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalStore{client: nil}
			fc := domain.FaultCausalObject{CausalID: "causal-1"}

			err := store.Update(ctx, fc)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("causal_id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalStore(client)
			fc := domain.FaultCausalObject{CausalID: ""}

			err := store.Update(ctx, fc)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "causal_id 不能为空")
		})

		Convey("成功更新 FaultCausalObject", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewFaultCausalStore(client)

			fc := domain.FaultCausalObject{
				CausalID:         "causal-1",
				CausalConfidence: 0.95,
				CausalReason:     "high correlation",
				SUpdateTime:      time.Now(),
			}

			err := store.Update(ctx, fc)

			So(err, ShouldBeNil)
		})

		Convey("更新失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalStore(client)

			fc := domain.FaultCausalObject{CausalID: "causal-1"}

			err := store.Update(ctx, fc)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestFaultCausalStore_QueryByIDs(t *testing.T) {
	Convey("TestFaultCausalStore_QueryByIDs", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalStore{client: nil}

			result, err := store.QueryByIDs(ctx, []string{"id1", "id2"})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("ids 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalStore(client)

			result, err := store.QueryByIDs(ctx, []string{})

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("ids 全为空字符串返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalStore(client)

			result, err := store.QueryByIDs(ctx, []string{"", ""})

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功查询 FaultCausalObject", func() {
			body := `{
				"docs": [
					{"found": true, "_source": {"causal_id": "causal-1", "causal_confidence": 0.9}},
					{"found": true, "_source": {"causal_id": "causal-2", "causal_confidence": 0.8}}
				]
			}`
			client := newMockClient(200, body)
			store := NewFaultCausalStore(client)

			result, err := store.QueryByIDs(ctx, []string{"causal-1", "causal-2"})

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalStore(client)

			result, err := store.QueryByIDs(ctx, []string{"causal-1"})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "mget FaultCausalObject 失败")
		})

		Convey("响应错误返回错误", func() {
			client := newMockClient(404, `{"error": {"type": "index_not_found_exception", "reason": "no such index"}}`)
			store := NewFaultCausalStore(client)

			result, err := store.QueryByIDs(ctx, []string{"causal-1"})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestFaultCausalStore_partialUpdate(t *testing.T) {
	Convey("TestFaultCausalStore_partialUpdate", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalStore{client: nil}

			err := store.partialUpdate(ctx, "causal-1", map[string]any{"confidence": 0.9})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalStore(client)

			err := store.partialUpdate(ctx, "", map[string]any{"confidence": 0.9})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "文档 ID 不能为空")
		})

		Convey("doc 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalStore(client)

			err := store.partialUpdate(ctx, "causal-1", map[string]any{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "更新文档不能为空")
		})

		Convey("成功部分更新", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewFaultCausalStore(client)

			err := store.partialUpdate(ctx, "causal-1", map[string]any{"confidence": 0.95})

			So(err, ShouldBeNil)
		})

		Convey("更新失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalStore(client)

			err := store.partialUpdate(ctx, "causal-1", map[string]any{"confidence": 0.95})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "更新 FaultCausal")
		})

		Convey("响应错误返回错误", func() {
			client := newMockClient(500, `{"error": {"type": "internal_error", "reason": "server error"}}`)
			store := NewFaultCausalStore(client)

			err := store.partialUpdate(ctx, "causal-1", map[string]any{"confidence": 0.95})

			So(err, ShouldNotBeNil)
		})
	})
}
