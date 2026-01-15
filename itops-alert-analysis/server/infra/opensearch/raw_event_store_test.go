package opensearch

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	opensearchsdk "github.com/opensearch-project/opensearch-go/v2"
	. "github.com/smartystreets/goconvey/convey"
)

// mockTransport 实现 http.RoundTripper 接口，用于 mock HTTP 响应
type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

// newMockClient 创建带有 mock transport 的 OpenSearch 客户端
func newMockClient(statusCode int, body string) *opensearchsdk.Client {
	transport := &mockTransport{
		response: &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		},
	}
	transport.response.Header.Set("Content-Type", "application/json")

	client, _ := opensearchsdk.NewClient(opensearchsdk.Config{
		Transport: transport,
		Addresses: []string{"http://localhost:9200"},
	})
	return client
}

// newMockClientWithError 创建返回错误的 mock 客户端
func newMockClientWithError(err error) *opensearchsdk.Client {
	transport := &mockTransport{
		err: err,
	}
	client, _ := opensearchsdk.NewClient(opensearchsdk.Config{
		Transport: transport,
		Addresses: []string{"http://localhost:9200"},
	})
	return client
}

func TestNewRawEventStore(t *testing.T) {
	Convey("TestNewRawEventStore", t, func() {
		Convey("成功创建 RawEventStore", func() {
			client := newMockClient(200, `{}`)
			store := NewRawEventStore(client)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldEqual, client)
		})

		Convey("使用 nil client 创建", func() {
			store := NewRawEventStore(nil)

			So(store, ShouldNotBeNil)
			So(store.client, ShouldBeNil)
		})
	})
}

func TestRawEventStore_Upsert(t *testing.T) {
	Convey("TestRawEventStore_Upsert", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &RawEventStore{client: nil}
			event := domain.RawEvent{EventID: 1}

			err := store.Upsert(ctx, event)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("event_id 为空返回错误", func() {
			client := newMockClient(200, `{}`)
			store := NewRawEventStore(client)
			event := domain.RawEvent{EventID: 0}

			err := store.Upsert(ctx, event)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "event_id 不能为空")
		})

		Convey("成功写入 RawEvent", func() {
			client := newMockClient(201, `{"result": "created"}`)
			store := NewRawEventStore(client)

			now := time.Now()
			event := domain.RawEvent{
				EventID:        1,
				EventTimestamp: now,
			}

			err := store.Upsert(ctx, event)

			So(err, ShouldBeNil)
		})

		Convey("写入失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewRawEventStore(client)

			event := domain.RawEvent{EventID: 1, EventTimestamp: time.Now()}

			err := store.Upsert(ctx, event)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "写入 RawEvent 失败")
		})

		Convey("响应状态码非 2xx 返回错误", func() {
			client := newMockClient(500, `{"error": {"type": "internal_error", "reason": "server error"}}`)
			store := NewRawEventStore(client)

			event := domain.RawEvent{EventID: 1, EventTimestamp: time.Now()}

			err := store.Upsert(ctx, event)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestRawEventStore_QueryByIDs(t *testing.T) {
	Convey("TestRawEventStore_QueryByIDs", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &RawEventStore{client: nil}

			result, err := store.QueryByIDs(ctx, []uint64{1, 2})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("ids 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewRawEventStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{})

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功查询 RawEvent", func() {
			body := `{
				"docs": [
					{"found": true, "_source": {"event_id": 1, "event_title": "event1"}},
					{"found": true, "_source": {"event_id": 2, "event_title": "event2"}}
				]
			}`
			client := newMockClient(200, body)
			store := NewRawEventStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{1, 2})

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("查询失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewRawEventStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{1})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "查询 RawEvent 失败")
		})

		Convey("响应错误返回错误", func() {
			client := newMockClient(404, `{"error": {"type": "index_not_found_exception", "reason": "no such index"}}`)
			store := NewRawEventStore(client)

			result, err := store.QueryByIDs(ctx, []uint64{1})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestRawEventStore_QueryByProviderID(t *testing.T) {
	Convey("TestRawEventStore_QueryByProviderID", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &RawEventStore{client: nil}

			result, err := store.QueryByProviderID(ctx, []string{"p1"})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("providerIDs 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewRawEventStore(client)

			result, err := store.QueryByProviderID(ctx, []string{})

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功搜索 RawEvent", func() {
			body := `{
				"hits": {
					"hits": [
						{"_source": {"event_id": 1, "event_title": "test event"}}
					]
				}
			}`
			client := newMockClient(200, body)
			store := NewRawEventStore(client)

			result, err := store.QueryByProviderID(ctx, []string{"p1"})

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("搜索失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewRawEventStore(client)

			result, err := store.QueryByProviderID(ctx, []string{"p1"})

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestRawEventStore_UpdateFaultID(t *testing.T) {
	Convey("TestRawEventStore_UpdateFaultID", t, func() {
		ctx := context.Background()

		Convey("client 为 nil 返回错误", func() {
			store := &RawEventStore{client: nil}

			err := store.UpdateFaultID(ctx, []uint64{1}, 100)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch client 未初始化")
		})

		Convey("eventIDs 为空返回 nil", func() {
			client := newMockClient(200, `{}`)
			store := NewRawEventStore(client)

			err := store.UpdateFaultID(ctx, []uint64{}, 100)

			So(err, ShouldBeNil)
		})

		Convey("成功批量更新 fault_id", func() {
			body := `{"errors": false, "items": [{"update": {"status": 200}}]}`
			client := newMockClient(200, body)
			store := NewRawEventStore(client)

			err := store.UpdateFaultID(ctx, []uint64{1, 2}, 100)

			So(err, ShouldBeNil)
		})

		Convey("批量更新失败返回错误", func() {
			client := newMockClientWithError(io.ErrUnexpectedEOF)
			store := NewRawEventStore(client)

			err := store.UpdateFaultID(ctx, []uint64{1}, 100)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "bulk update 请求失败")
		})

		Convey("批量更新部分失败返回错误", func() {
			body := `{"errors": true, "items": [{"update": {"status": 404}}]}`
			client := newMockClient(200, body)
			store := NewRawEventStore(client)

			err := store.UpdateFaultID(ctx, []uint64{1}, 100)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "bulk update 部分失败")
		})
	})
}

func TestRawEventStore_UpdateProblemID(t *testing.T) {
	Convey("TestRawEventStore_UpdateProblemID", t, func() {
		ctx := context.Background()

		Convey("成功批量更新 problem_id", func() {
			body := `{"errors": false, "items": [{"update": {"status": 200}}]}`
			client := newMockClient(200, body)
			store := NewRawEventStore(client)

			err := store.UpdateProblemID(ctx, []uint64{1, 2}, 200)

			So(err, ShouldBeNil)
		})
	})
}
