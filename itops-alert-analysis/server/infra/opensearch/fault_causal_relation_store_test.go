package opensearch

import (
	"context"
	"io"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewFaultCausalRelationStore(t *testing.T) {
	Convey("TestNewFaultCausalRelationStore", t, func() {
		Convey("成功创建 FaultCausalRelationStore", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldEqual, client)
		})

		Convey("使用 nil client 创建", func() {
			store := NewFaultCausalRelationStore(nil)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldBeNil)
		})
	})
}

func TestFaultCausalRelationStore_Upsert(t *testing.T) {
	Convey("TestFaultCausalRelationStore_Upsert", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalRelationStore{client: nil}
			fcr := domain.FaultCausalRelation{RelationID: "rel-1"}

			err := store.Upsert(ctx, fcr)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("relation_id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)
			fcr := domain.FaultCausalRelation{RelationID: ""}

			err := store.Upsert(ctx, fcr)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "relation_id 不能为空")
		})

		Convey("成功写入 FaultCausalRelation", func() {
			client := newMockClient(201, `{"result": "created"}`)
			store := NewFaultCausalRelationStore(client)

			now := time.Now()
			fcr := domain.FaultCausalRelation{
				RelationID:         "rel-1",
				RelationCreateTime: now,
			}

			err := store.Upsert(ctx, fcr)

			So(err, ShouldBeNil)
		})

		Convey("使用默认时间戳成功写入", func() {
			client := newMockClient(201, `{"result": "created"}`)
			store := NewFaultCausalRelationStore(client)

			fcr := domain.FaultCausalRelation{
				RelationID: "rel-1",
				// RelationCreateTime 为零值
			}

			err := store.Upsert(ctx, fcr)

			So(err, ShouldBeNil)
		})

		Convey("写入失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalRelationStore(client)

			fcr := domain.FaultCausalRelation{RelationID: "rel-1", RelationCreateTime: time.Now()}

			err := store.Upsert(ctx, fcr)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "写入 FaultCausalRelation 失败")
		})

		Convey("响应状态码非 2xx 返回错误", func() {
			client := newMockClient(500, `{"error": {"type": "internal_error", "reason": "server error"}}`)
			store := NewFaultCausalRelationStore(client)

			fcr := domain.FaultCausalRelation{RelationID: "rel-1", RelationCreateTime: time.Now()}

			err := store.Upsert(ctx, fcr)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestFaultCausalRelationStore_Update(t *testing.T) {
	Convey("TestFaultCausalRelationStore_Update", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalRelationStore{client: nil}
			fcr := domain.FaultCausalRelation{RelationID: "rel-1"}

			err := store.Update(ctx, fcr)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("relation_id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)
			fcr := domain.FaultCausalRelation{RelationID: ""}

			err := store.Update(ctx, fcr)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "relation_id 不能为空")
		})

		Convey("成功更新 FaultCausalRelation", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewFaultCausalRelationStore(client)

			fcr := domain.FaultCausalRelation{
				RelationID:         "rel-1",
				RelationClass:      "causal",
				SourceObjectID:     "source-1",
				TargetObjectID:     "target-1",
				RelationUpdateTime: time.Now(),
			}

			err := store.Update(ctx, fcr)

			So(err, ShouldBeNil)
		})

		Convey("更新失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalRelationStore(client)

			fcr := domain.FaultCausalRelation{RelationID: "rel-1"}

			err := store.Update(ctx, fcr)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestFaultCausalRelationStore_QueryByIDs(t *testing.T) {
	Convey("TestFaultCausalRelationStore_QueryByIDs", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalRelationStore{client: nil}

			result, err := store.QueryByIDs(ctx, []string{"id1", "id2"})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("ids 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByIDs(ctx, []string{})

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("ids 全为空字符串返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByIDs(ctx, []string{"", ""})

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功查询 FaultCausalRelation", func() {
			body := `{
				"docs": [
					{"found": true, "_source": {"relation_id": "rel-1", "source_object_id": "src-1"}},
					{"found": true, "_source": {"relation_id": "rel-2", "source_object_id": "src-2"}}
				]
			}`
			client := newMockClient(200, body)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByIDs(ctx, []string{"rel-1", "rel-2"})

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByIDs(ctx, []string{"rel-1"})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "查询 FaultCausalRelation 失败")
		})

		Convey("响应错误返回错误", func() {
			client := newMockClient(404, `{"error": {"type": "index_not_found_exception", "reason": "no such index"}}`)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByIDs(ctx, []string{"rel-1"})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestFaultCausalRelationStore_QueryByEntityPair(t *testing.T) {
	Convey("TestFaultCausalRelationStore_QueryByEntityPair", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalRelationStore{client: nil}

			result, err := store.QueryByEntityPair(ctx, "source-1", "target-1")

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("sourceID 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByEntityPair(ctx, "", "target-1")

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "source_id、target_id 不能为空")
		})

		Convey("targetID 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByEntityPair(ctx, "source-1", "")

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "source_id、target_id 不能为空")
		})

		Convey("成功查询实体对关系", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"relation_id": "rel-1", "source_object_id": "source-1", "target_object_id": "target-1"}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByEntityPair(ctx, "source-1", "target-1")

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("未找到关系返回空列表", func() {
			body := `{"hits": {"hits": []}}`
			client := newMockClient(200, body)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByEntityPair(ctx, "source-1", "target-1")

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalRelationStore(client)

			result, err := store.QueryByEntityPair(ctx, "source-1", "target-1")

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestFaultCausalRelationStore_validateClient(t *testing.T) {
	Convey("TestFaultCausalRelationStore_validateClient", t, func() {
		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalRelationStore{client: nil}

			err := store.validateClient()

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("client 已初始化返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)

			err := store.validateClient()

			So(err, ShouldBeNil)
		})
	})
}

func TestFaultCausalRelationStore_partialUpdate(t *testing.T) {
	Convey("TestFaultCausalRelationStore_partialUpdate", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultCausalRelationStore{client: nil}

			err := store.partialUpdate(ctx, "rel-1", map[string]any{"class": "updated"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)

			err := store.partialUpdate(ctx, "", map[string]any{"class": "updated"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "文档 ID 不能为空")
		})

		Convey("doc 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultCausalRelationStore(client)

			err := store.partialUpdate(ctx, "rel-1", map[string]any{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "更新文档不能为空")
		})

		Convey("成功部分更新", func() {
			client := newMockClient(200, `{"result": "updated"}`)
			store := NewFaultCausalRelationStore(client)

			err := store.partialUpdate(ctx, "rel-1", map[string]any{"class": "updated"})

			So(err, ShouldBeNil)
		})

		Convey("更新失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultCausalRelationStore(client)

			err := store.partialUpdate(ctx, "rel-1", map[string]any{"class": "updated"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "更新 FaultCausalRelation")
		})

		Convey("响应错误返回错误", func() {
			client := newMockClient(500, `{"error": {"type": "internal_error", "reason": "server error"}}`)
			store := NewFaultCausalRelationStore(client)

			err := store.partialUpdate(ctx, "rel-1", map[string]any{"class": "updated"})

			So(err, ShouldNotBeNil)
		})
	})
}
