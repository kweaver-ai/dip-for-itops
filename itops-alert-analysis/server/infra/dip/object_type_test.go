package dip

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClient_GetObjectTypes(t *testing.T) {
	Convey("TestClient_GetObjectTypes", t, func() {
		Convey("成功获取对象类列表", func() {
			var capturedMethod string
			var capturedPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				capturedPath = r.URL.Path

				resp := ObjectTypeListResponse{
					Entries: []ObjectType{
						{ID: "ot-1", Name: "Server", Tags: []string{"infra"}, Comment: "服务器"},
						{ID: "ot-2", Name: "Application", Tags: []string{"app"}, Comment: "应用"},
					},
					TotalCount: 2,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.GetObjectTypes(context.Background())

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result), ShouldEqual, 2)
			So(result[0].ID, ShouldEqual, "ot-1")
			So(result[0].Name, ShouldEqual, "Server")
			So(result[0].Tags, ShouldResemble, []string{"infra"})
			So(result[0].Comment, ShouldEqual, "服务器")
			So(result[1].ID, ShouldEqual, "ot-2")
			So(capturedMethod, ShouldEqual, http.MethodGet)
			So(capturedPath, ShouldEqual, "/api/ontology-manager/v1/knowledge-networks/test-kn/object-types")
		})

		Convey("空对象类列表", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := ObjectTypeListResponse{
					Entries:    []ObjectType{},
					TotalCount: 0,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.GetObjectTypes(context.Background())

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("响应状态码非 2xx", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal error"}`))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.GetObjectTypes(context.Background())

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
			result, err := client.GetObjectTypes(context.Background())

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "解析 JSON 失败")
		})

		Convey("HTTP 请求失败", func() {
			client := newTestClient("http://invalid-host:99999")
			result, err := client.GetObjectTypes(context.Background())

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
