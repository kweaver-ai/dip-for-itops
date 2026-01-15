package dip

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClient_QuerySubGraph(t *testing.T) {
	Convey("TestClient_QuerySubGraph", t, func() {
		Convey("成功查询子图", func() {
			var capturedMethod string
			var capturedPath string
			var capturedReqBody SubGraphQueryRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				capturedPath = r.URL.Path
				json.NewDecoder(r.Body).Decode(&capturedReqBody)

				resp := SubGraphResponse{
					Objects: map[string]SubGraphObject{
						"obj-1": {
							ID:             "obj-1",
							ObjectTypeID:   "Server",
							ObjectTypeName: "服务器",
							Display:        "server-01",
							Properties: SubGraphObjectProperties{
								SID:  "entity-1",
								Name: "server-01",
							},
						},
						"obj-2": {
							ID:             "obj-2",
							ObjectTypeID:   "Application",
							ObjectTypeName: "应用",
							Display:        "app-01",
							Properties: SubGraphObjectProperties{
								SID:  "entity-2",
								Name: "app-01",
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			req := SubGraphQueryRequest{
				SourceObjectTypeID: "Server",
				Condition: &Condition{
					Field:     "s_id",
					Operation: "==",
					Value:     "entity-1",
				},
				Direction:  "bidirectional",
				PathLength: 1,
			}

			result, err := client.QuerySubGraph(context.Background(), req)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 2)
			So(result.Objects["obj-1"].ID, ShouldEqual, "obj-1")
			So(result.Objects["obj-1"].ObjectTypeID, ShouldEqual, "Server")
			So(result.Objects["obj-1"].Properties.SID, ShouldEqual, "entity-1")
			So(result.Objects["obj-2"].ID, ShouldEqual, "obj-2")
			So(capturedMethod, ShouldEqual, http.MethodPost)
			So(capturedPath, ShouldEqual, "/api/ontology-query/v1/knowledge-networks/test-kn/subgraph")
			So(capturedReqBody.SourceObjectTypeID, ShouldEqual, "Server")
			So(capturedReqBody.Direction, ShouldEqual, "bidirectional")
			So(capturedReqBody.PathLength, ShouldEqual, 1)
		})

		Convey("查询子图返回空结果", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := SubGraphResponse{
					Objects: map[string]SubGraphObject{},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			req := SubGraphQueryRequest{
				SourceObjectTypeID: "Server",
				Direction:          "outgoing",
				PathLength:         2,
			}

			result, err := client.QuerySubGraph(context.Background(), req)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 0)
		})

		Convey("响应状态码非 2xx", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal error"}`))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			req := SubGraphQueryRequest{
				SourceObjectTypeID: "Server",
				Direction:          "bidirectional",
				PathLength:         1,
			}

			result, err := client.QuerySubGraph(context.Background(), req)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "500")
		})

		Convey("JSON 解析失败", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`invalid json`))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			req := SubGraphQueryRequest{
				SourceObjectTypeID: "Server",
				Direction:          "bidirectional",
				PathLength:         1,
			}

			result, err := client.QuerySubGraph(context.Background(), req)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "解析 JSON 失败")
		})

		Convey("HTTP 请求失败", func() {
			client := newTestClient("http://invalid-host:99999")
			req := SubGraphQueryRequest{
				SourceObjectTypeID: "Server",
				Direction:          "bidirectional",
				PathLength:         1,
			}

			result, err := client.QuerySubGraph(context.Background(), req)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("无条件查询子图", func() {
			var capturedReqBody SubGraphQueryRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&capturedReqBody)

				resp := SubGraphResponse{
					Objects: map[string]SubGraphObject{
						"obj-1": {
							ID:           "obj-1",
							ObjectTypeID: "Server",
							Properties:   SubGraphObjectProperties{SID: "entity-1"},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			req := SubGraphQueryRequest{
				SourceObjectTypeID: "Server",
				Condition:          nil, // 无条件
				Direction:          "incoming",
				PathLength:         3,
			}

			result, err := client.QuerySubGraph(context.Background(), req)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 1)
			So(capturedReqBody.Condition, ShouldBeNil)
			So(capturedReqBody.Direction, ShouldEqual, "incoming")
			So(capturedReqBody.PathLength, ShouldEqual, 3)
		})
	})
}
