package opensearch

import (
	"context"
	"io"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewFaultPointRelationStore(t *testing.T) {
	Convey("TestNewFaultPointRelationStore", t, func() {
		Convey("成功创建 FaultPointRelationStore", func() {
			client := newMockClient(200, `{}`)
			store := NewFaultPointRelationStore(client)

			So(store, ShouldNotBeNil)
		})

		Convey("使用 nil client 创建", func() {
			store := NewFaultPointRelationStore(nil)

			So(store, ShouldNotBeNil)
		})
	})
}

func TestFaultPointRelationStore_Upsert(t *testing.T) {
	Convey("TestFaultPointRelationStore_Upsert", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &FaultPointRelationStore{client: nil}
			relation := domain.FaultPointRelation{RelationId: 1}

			err := store.Upsert(ctx, relation)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("成功写入故障点关系", func() {
			client := newMockClient(201, `{"result": "created"}`)
			store := NewFaultPointRelationStore(client)

			now := time.Now()
			relation := domain.FaultPointRelation{
				RelationId:         1,
				RelationCreateTime: now,
			}

			err := store.Upsert(ctx, relation)

			So(err, ShouldBeNil)
		})

		Convey("写入失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewFaultPointRelationStore(client)

			relation := domain.FaultPointRelation{
				RelationId:         1,
				RelationCreateTime: time.Now(),
			}

			err := store.Upsert(ctx, relation)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "写入故障点关系失败")
		})

		Convey("响应状态码非 2xx 返回错误", func() {
			client := newMockClient(500, `{"error": {"type": "internal_error", "reason": "server error"}}`)
			store := NewFaultPointRelationStore(client)

			relation := domain.FaultPointRelation{
				RelationId:         1,
				RelationCreateTime: time.Now(),
			}

			err := store.Upsert(ctx, relation)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "写入故障点关系失败")
		})
	})
}
