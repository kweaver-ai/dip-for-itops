package dip

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	. "github.com/smartystreets/goconvey/convey"
)

// newTestClient 创建连接到测试服务器的客户端
func newTestClient(serverURL string) *Client {
	cfg := config.DIPConfig{
		Host:               serverURL,
		KnID:               "test-kn",
		Authorization:      "Bearer test-token",
		InsecureSkipVerify: true,
		Timeout:            5 * time.Second,
	}
	return NewClient(cfg, mockGetAuth, mockGetKnID)
}

func TestClient_QueryObjectData(t *testing.T) {
	Convey("TestClient_QueryObjectData", t, func() {
		Convey("成功获取数据（无 searchAfter）", func() {
			var capturedMethod string
			var capturedPath string
			var capturedHeader string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				capturedPath = r.URL.Path
				capturedHeader = r.Header.Get("X-HTTP-Method-Override")

				resp := ObjectDataResponse{
					Datas: []ObjectInstance{
						{"id": "1", "name": "object1"},
						{"id": "2", "name": "object2"},
					},
					TotalCount:  2,
					SearchAfter: []interface{}{"2"},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryObjectData(context.Background(), "test-ot", 10, nil)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 2)
			So(result.TotalCount, ShouldEqual, 2)
			So(result.Datas[0]["id"], ShouldEqual, "1")
			So(capturedMethod, ShouldEqual, http.MethodPost)
			So(capturedPath, ShouldEqual, "/api/ontology-query/v1/knowledge-networks/test-kn/object-types/test-ot")
			So(capturedHeader, ShouldEqual, "GET")
		})

		Convey("成功获取数据（有 searchAfter）", func() {
			var capturedReqBody QueryRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&capturedReqBody)

				resp := ObjectDataResponse{
					Datas: []ObjectInstance{
						{"id": "3", "name": "object3"},
					},
					TotalCount:  0,
					SearchAfter: []interface{}{},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryObjectData(context.Background(), "test-ot", 10, []interface{}{"2"})

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(capturedReqBody.SearchAfter, ShouldNotBeNil)
			So(capturedReqBody.NeedTotal, ShouldBeFalse)
		})

		Convey("响应状态码非 2xx", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal error"}`))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryObjectData(context.Background(), "test-ot", 10, nil)

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
			result, err := client.QueryObjectData(context.Background(), "test-ot", 10, nil)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "解析 JSON 失败")
		})

		Convey("HTTP 请求失败", func() {
			client := newTestClient("http://invalid-host:99999")
			result, err := client.QueryObjectData(context.Background(), "test-ot", 10, nil)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestClient_QueryAllObjectData(t *testing.T) {
	Convey("TestClient_QueryAllObjectData", t, func() {
		Convey("成功获取所有数据（单页）", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := ObjectDataResponse{
					Datas: []ObjectInstance{
						{"id": "1", "name": "object1"},
						{"id": "2", "name": "object2"},
					},
					TotalCount:  2,
					SearchAfter: []interface{}{}, // 空表示没有更多数据
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryAllObjectData(context.Background(), "test-ot", 10)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("成功获取所有数据（多页分页）", func() {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				var resp ObjectDataResponse

				if callCount == 1 {
					// 第一页
					resp = ObjectDataResponse{
						Datas: []ObjectInstance{
							{"id": "1", "name": "object1"},
							{"id": "2", "name": "object2"},
						},
						TotalCount:  5,
						SearchAfter: []interface{}{"2"},
					}
				} else if callCount == 2 {
					// 第二页
					resp = ObjectDataResponse{
						Datas: []ObjectInstance{
							{"id": "3", "name": "object3"},
							{"id": "4", "name": "object4"},
						},
						TotalCount:  5,
						SearchAfter: []interface{}{"4"},
					}
				} else {
					// 第三页（最后一页）
					resp = ObjectDataResponse{
						Datas: []ObjectInstance{
							{"id": "5", "name": "object5"},
						},
						TotalCount:  5,
						SearchAfter: []interface{}{}, // 没有更多数据
					}
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryAllObjectData(context.Background(), "test-ot", 2)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 5)
			So(callCount, ShouldEqual, 3)
		})

		Convey("空数据", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := ObjectDataResponse{
					Datas:       []ObjectInstance{},
					TotalCount:  0,
					SearchAfter: []interface{}{},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryAllObjectData(context.Background(), "test-ot", 10)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("查询失败", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "server error"}`))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryAllObjectData(context.Background(), "test-ot", 10)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
