package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// mockGetAuth 用于测试的获取 Authorization 函数
func mockGetAuth() string {
	return "Bearer test-token"
}

func TestNewClient(t *testing.T) {
	Convey("TestNewClient", t, func() {
		Convey("使用默认配置创建客户端", func() {
			client := NewClient(Config{
				BaseURL: "https://example.com",
			}, mockGetAuth)

			So(client, ShouldNotBeNil)
			So(client.baseURL, ShouldEqual, "https://example.com")
			So(client.httpClient.Timeout, ShouldEqual, 30*time.Second)
		})

		Convey("使用自定义超时创建客户端", func() {
			client := NewClient(Config{
				BaseURL: "https://example.com",
				Timeout: 60 * time.Second,
			}, mockGetAuth)

			So(client, ShouldNotBeNil)
			So(client.httpClient.Timeout, ShouldEqual, 60*time.Second)
		})

		Convey("使用自定义 Headers 创建客户端", func() {
			client := NewClient(Config{
				BaseURL: "https://example.com",
				Headers: map[string]string{
					"Authorization": "Bearer token",
					"User-Agent":    "test-client",
				},
			}, mockGetAuth)

			So(client, ShouldNotBeNil)
			So(client.headers["Authorization"], ShouldEqual, "Bearer token")
			So(client.headers["User-Agent"], ShouldEqual, "test-client")
		})

		Convey("启用 InsecureSkipVerify 创建客户端", func() {
			client := NewClient(Config{
				BaseURL:            "https://example.com",
				InsecureSkipVerify: true,
			}, mockGetAuth)

			So(client, ShouldNotBeNil)
		})
	})
}

func TestClient_Do(t *testing.T) {
	Convey("TestClient_Do", t, func() {
		Convey("成功执行 GET 请求", func() {
			var capturedMethod string
			var capturedPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				capturedPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "success"}`))
			}))
			defer server.Close()

			client := NewClient(Config{BaseURL: server.URL}, mockGetAuth)
			resp, err := client.Do(context.Background(), Request{
				Method: http.MethodGet,
				Path:   "/api/test",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(string(resp.Body), ShouldEqual, `{"message": "success"}`)
			So(capturedMethod, ShouldEqual, http.MethodGet)
			So(capturedPath, ShouldEqual, "/api/test")
		})

		Convey("成功执行 POST 请求（带请求体）", func() {
			var capturedBody map[string]interface{}
			var capturedContentType string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContentType = r.Header.Get("Content-Type")
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &capturedBody)

				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id": 1}`))
			}))
			defer server.Close()

			client := NewClient(Config{BaseURL: server.URL}, mockGetAuth)
			resp, err := client.Do(context.Background(), Request{
				Method: http.MethodPost,
				Path:   "/api/create",
				Body:   map[string]string{"name": "test"},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(capturedContentType, ShouldEqual, "application/json")
			So(capturedBody["name"], ShouldEqual, "test")
		})

		Convey("请求携带动态 Authorization", func() {
			var capturedAuth string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedAuth = r.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient(Config{BaseURL: server.URL}, mockGetAuth)
			resp, err := client.Do(context.Background(), Request{
				Method: http.MethodGet,
				Path:   "/api/test",
			})

			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(capturedAuth, ShouldEqual, "Bearer test-token")
		})

		Convey("动态 Authorization 可以更新", func() {
			var capturedAuth string
			currentAuth := "Bearer initial-token"

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedAuth = r.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			dynamicGetAuth := func() string {
				return currentAuth
			}

			client := NewClient(Config{BaseURL: server.URL}, dynamicGetAuth)

			// 第一次请求
			_, _ = client.Do(context.Background(), Request{
				Method: http.MethodGet,
				Path:   "/api/test",
			})
			So(capturedAuth, ShouldEqual, "Bearer initial-token")

			// 更新 token
			currentAuth = "Bearer updated-token"

			// 第二次请求应该使用更新后的 token
			_, _ = client.Do(context.Background(), Request{
				Method: http.MethodGet,
				Path:   "/api/test",
			})
			So(capturedAuth, ShouldEqual, "Bearer updated-token")
		})

		Convey("getAuth 为 nil 时不设置 Authorization", func() {
			var capturedAuth string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedAuth = r.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient(Config{BaseURL: server.URL}, nil)
			resp, err := client.Do(context.Background(), Request{
				Method: http.MethodGet,
				Path:   "/api/test",
			})

			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(capturedAuth, ShouldEqual, "")
		})

		Convey("getAuth 返回空字符串时不设置 Authorization", func() {
			var capturedAuth string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedAuth = r.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			emptyGetAuth := func() string {
				return ""
			}

			client := NewClient(Config{BaseURL: server.URL}, emptyGetAuth)
			resp, err := client.Do(context.Background(), Request{
				Method: http.MethodGet,
				Path:   "/api/test",
			})

			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(capturedAuth, ShouldEqual, "")
		})

		Convey("HTTP 请求失败", func() {
			client := NewClient(Config{BaseURL: "http://invalid-host:99999"}, mockGetAuth)
			resp, err := client.Do(context.Background(), Request{
				Method: http.MethodGet,
				Path:   "/api/test",
			})

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "请求失败")
		})

		Convey("请求体序列化失败", func() {
			client := NewClient(Config{BaseURL: "http://example.com"}, mockGetAuth)
			resp, err := client.Do(context.Background(), Request{
				Method: http.MethodPost,
				Path:   "/api/test",
				Body:   make(chan int), // channel 无法序列化为 JSON
			})

			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "序列化请求体失败")
		})
	})
}

func TestClient_Get(t *testing.T) {
	Convey("TestClient_Get", t, func() {
		Convey("成功执行 GET 请求", func() {
			var capturedMethod string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": "value"}`))
			}))
			defer server.Close()

			client := NewClient(Config{BaseURL: server.URL}, mockGetAuth)
			resp, err := client.Get(context.Background(), "/api/resource", nil)

			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(capturedMethod, ShouldEqual, http.MethodGet)
		})
	})
}

func TestClient_Post(t *testing.T) {
	Convey("TestClient_Post", t, func() {
		Convey("成功执行 POST 请求", func() {
			var capturedMethod string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				w.WriteHeader(http.StatusCreated)
			}))
			defer server.Close()

			client := NewClient(Config{BaseURL: server.URL}, mockGetAuth)
			resp, err := client.Post(context.Background(), "/api/resource", map[string]string{"key": "value"}, nil)

			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(capturedMethod, ShouldEqual, http.MethodPost)
		})
	})
}

func TestClient_Put(t *testing.T) {
	Convey("TestClient_Put", t, func() {
		Convey("成功执行 PUT 请求", func() {
			var capturedMethod string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient(Config{BaseURL: server.URL}, mockGetAuth)
			resp, err := client.Put(context.Background(), "/api/resource", map[string]string{"key": "value"}, nil)

			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(capturedMethod, ShouldEqual, http.MethodPut)
		})
	})
}

func TestClient_Delete(t *testing.T) {
	Convey("TestClient_Delete", t, func() {
		Convey("成功执行 DELETE 请求", func() {
			var capturedMethod string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := NewClient(Config{BaseURL: server.URL}, mockGetAuth)
			resp, err := client.Delete(context.Background(), "/api/resource", nil)

			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			So(capturedMethod, ShouldEqual, http.MethodDelete)
		})
	})
}

func TestResponse_DecodeJSON(t *testing.T) {
	Convey("TestResponse_DecodeJSON", t, func() {
		Convey("成功解析 JSON", func() {
			resp := &Response{
				StatusCode: 200,
				Body:       []byte(`{"id": 1, "name": "test"}`),
			}

			var result struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}
			err := resp.DecodeJSON(&result)

			So(err, ShouldBeNil)
			So(result.ID, ShouldEqual, 1)
			So(result.Name, ShouldEqual, "test")
		})

		Convey("解析无效 JSON 失败", func() {
			resp := &Response{
				StatusCode: 200,
				Body:       []byte(`invalid json`),
			}

			var result map[string]interface{}
			err := resp.DecodeJSON(&result)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "解析 JSON 失败")
		})
	})
}

func TestResponse_IsSuccess(t *testing.T) {
	Convey("TestResponse_IsSuccess", t, func() {
		Convey("2xx 状态码返回 true", func() {
			testCases := []int{200, 201, 202, 204, 299}
			for _, code := range testCases {
				resp := &Response{StatusCode: code}
				So(resp.IsSuccess(), ShouldBeTrue)
			}
		})

		Convey("非 2xx 状态码返回 false", func() {
			testCases := []int{100, 199, 300, 400, 404, 500, 502}
			for _, code := range testCases {
				resp := &Response{StatusCode: code}
				So(resp.IsSuccess(), ShouldBeFalse)
			}
		})
	})
}

func TestResponse_Error(t *testing.T) {
	Convey("TestResponse_Error", t, func() {
		Convey("成功响应返回 nil", func() {
			resp := &Response{
				StatusCode: 200,
				Body:       []byte(`{"data": "ok"}`),
			}

			So(resp.Error(), ShouldBeNil)
		})

		Convey("失败响应返回错误", func() {
			resp := &Response{
				StatusCode: 500,
				Body:       []byte(`{"error": "internal error"}`),
			}

			err := resp.Error()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "500")
			So(err.Error(), ShouldContainSubstring, "internal error")
		})
	})
}

func TestClient_SetHeader(t *testing.T) {
	Convey("TestClient_SetHeader", t, func() {
		Convey("设置单个 Header", func() {
			client := NewClient(Config{BaseURL: "https://example.com"}, mockGetAuth)
			client.SetHeader("X-Custom-Header", "custom-value")

			So(client.headers["X-Custom-Header"], ShouldEqual, "custom-value")
		})

		Convey("在空 headers 上设置", func() {
			client := &Client{baseURL: "https://example.com"}
			client.SetHeader("X-Custom-Header", "custom-value")

			So(client.headers["X-Custom-Header"], ShouldEqual, "custom-value")
		})
	})
}

func TestClient_SetHeaders(t *testing.T) {
	Convey("TestClient_SetHeaders", t, func() {
		Convey("批量设置 Headers", func() {
			client := NewClient(Config{BaseURL: "https://example.com"}, mockGetAuth)
			client.SetHeaders(map[string]string{
				"X-Header-1": "value-1",
				"X-Header-2": "value-2",
			})

			So(client.headers["X-Header-1"], ShouldEqual, "value-1")
			So(client.headers["X-Header-2"], ShouldEqual, "value-2")
		})

		Convey("在空 headers 上批量设置", func() {
			client := &Client{baseURL: "https://example.com"}
			client.SetHeaders(map[string]string{
				"X-Header-1": "value-1",
			})

			So(client.headers["X-Header-1"], ShouldEqual, "value-1")
		})
	})
}
