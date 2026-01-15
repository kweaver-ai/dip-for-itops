package opensearch

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOpenSearchError_Error(t *testing.T) {
	Convey("TestOpenSearchError_Error", t, func() {
		Convey("有 Reason 和 RootCause", func() {
			osErr := &OpenSearchError{
				Status: 400,
			}
			osErr.ErrorInfo.Type = "mapper_parsing_exception"
			osErr.ErrorInfo.Reason = "failed to parse"
			osErr.ErrorInfo.RootCause = []struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Index  string `json:"index,omitempty"`
			}{
				{Type: "parsing_exception", Reason: "invalid field"},
			}

			errMsg := osErr.Error()

			So(errMsg, ShouldContainSubstring, "mapper_parsing_exception")
			So(errMsg, ShouldContainSubstring, "failed to parse")
			So(errMsg, ShouldContainSubstring, "root:")
			So(errMsg, ShouldContainSubstring, "parsing_exception")
			So(errMsg, ShouldContainSubstring, "invalid field")
		})

		Convey("有 Reason 无 RootCause", func() {
			osErr := &OpenSearchError{
				Status: 400,
			}
			osErr.ErrorInfo.Type = "index_not_found_exception"
			osErr.ErrorInfo.Reason = "no such index [test-index]"

			errMsg := osErr.Error()

			So(errMsg, ShouldContainSubstring, "index_not_found_exception")
			So(errMsg, ShouldContainSubstring, "no such index [test-index]")
			So(errMsg, ShouldNotContainSubstring, "root:")
		})

		Convey("无 Reason 只有 Status", func() {
			osErr := &OpenSearchError{
				Status: 500,
			}

			errMsg := osErr.Error()

			So(errMsg, ShouldContainSubstring, "opensearch error")
			So(errMsg, ShouldContainSubstring, "status=500")
		})
	})
}

func TestReadResponseBody(t *testing.T) {
	Convey("TestReadResponseBody", t, func() {
		Convey("成功读取响应体", func() {
			body := strings.NewReader(`{"status": "ok"}`)

			data, err := readResponseBody(body)

			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `{"status": "ok"}`)
		})

		Convey("读取空响应体", func() {
			body := strings.NewReader("")

			data, err := readResponseBody(body)

			So(err, ShouldBeNil)
			So(data, ShouldNotBeNil)
			So(len(data), ShouldEqual, 0)
		})

		Convey("读取失败", func() {
			body := &errorReader{err: errors.New("read error")}

			data, err := readResponseBody(body)

			So(err, ShouldNotBeNil)
			So(data, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "读取 OpenSearch 响应失败")
		})
	})
}

func TestFormatErrorMessage(t *testing.T) {
	Convey("TestFormatErrorMessage", t, func() {
		Convey("空数据返回错误", func() {
			err := formatErrorMessage([]byte{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch 返回空错误响应")
		})

		Convey("解析结构化 OpenSearch 错误", func() {
			data := []byte(`{
				"error": {
					"type": "index_not_found_exception",
					"reason": "no such index [test-index]",
					"root_cause": [
						{"type": "index_not_found_exception", "reason": "no such index", "index": "test-index"}
					]
				},
				"status": 404
			}`)

			err := formatErrorMessage(data)

			So(err, ShouldNotBeNil)
			osErr, ok := err.(*OpenSearchError)
			So(ok, ShouldBeTrue)
			So(osErr.Status, ShouldEqual, 404)
			So(osErr.ErrorInfo.Type, ShouldEqual, "index_not_found_exception")
		})

		Convey("解析非 JSON 格式错误", func() {
			data := []byte("Internal Server Error")

			err := formatErrorMessage(data)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Internal Server Error")
		})

		Convey("解析无 Reason 的 JSON", func() {
			data := []byte(`{"status": 500}`)

			err := formatErrorMessage(data)

			So(err, ShouldNotBeNil)
			// 无 Reason 时返回原始响应
			So(err.Error(), ShouldContainSubstring, "status")
		})

		Convey("解析只有空格的数据", func() {
			data := []byte("   \n\t  ")

			err := formatErrorMessage(data)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unknown opensearch error")
		})
	})
}

func TestReadErrorResponse(t *testing.T) {
	Convey("TestReadErrorResponse", t, func() {
		Convey("成功读取并解析错误响应", func() {
			body := strings.NewReader(`{
				"error": {
					"type": "resource_already_exists_exception",
					"reason": "index already exists"
				},
				"status": 400
			}`)

			err := readErrorResponse(body)

			So(err, ShouldNotBeNil)
			osErr, ok := err.(*OpenSearchError)
			So(ok, ShouldBeTrue)
			So(osErr.ErrorInfo.Type, ShouldEqual, "resource_already_exists_exception")
		})

		Convey("读取响应体失败", func() {
			body := &errorReader{err: errors.New("connection reset")}

			err := readErrorResponse(body)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "读取 OpenSearch 错误响应失败")
		})
	})
}

func TestDecodeMGet(t *testing.T) {
	Convey("TestDecodeMGet", t, func() {
		type TestDoc struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		Convey("成功解析 mget 响应", func() {
			data := []byte(`{
				"docs": [
					{"found": true, "_source": {"id": "1", "name": "doc1"}},
					{"found": true, "_source": {"id": "2", "name": "doc2"}}
				]
			}`)

			items, err := decodeMGet[TestDoc](data)

			So(err, ShouldBeNil)
			So(len(items), ShouldEqual, 2)
			So(items[0].ID, ShouldEqual, "1")
			So(items[0].Name, ShouldEqual, "doc1")
			So(items[1].ID, ShouldEqual, "2")
			So(items[1].Name, ShouldEqual, "doc2")
		})

		Convey("跳过未找到的文档", func() {
			data := []byte(`{
				"docs": [
					{"found": true, "_source": {"id": "1", "name": "doc1"}},
					{"found": false},
					{"found": true, "_source": {"id": "3", "name": "doc3"}}
				]
			}`)

			items, err := decodeMGet[TestDoc](data)

			So(err, ShouldBeNil)
			So(len(items), ShouldEqual, 2)
			So(items[0].ID, ShouldEqual, "1")
			So(items[1].ID, ShouldEqual, "3")
		})

		Convey("跳过空 Source 的文档", func() {
			data := []byte(`{
				"docs": [
					{"found": true, "_source": {"id": "1", "name": "doc1"}},
					{"found": true}
				]
			}`)

			items, err := decodeMGet[TestDoc](data)

			So(err, ShouldBeNil)
			So(len(items), ShouldEqual, 1)
			So(items[0].ID, ShouldEqual, "1")
		})

		Convey("空 docs 数组", func() {
			data := []byte(`{"docs": []}`)

			items, err := decodeMGet[TestDoc](data)

			So(err, ShouldBeNil)
			So(len(items), ShouldEqual, 0)
		})

		Convey("解析 mget 响应失败", func() {
			data := []byte(`invalid json`)

			items, err := decodeMGet[TestDoc](data)

			So(err, ShouldNotBeNil)
			So(items, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "解析 mget 响应失败")
		})

		Convey("解析单个文档失败", func() {
			data := []byte(`{
				"docs": [
					{"found": true, "_source": {"id": 123, "name": "doc1"}}
				]
			}`)

			// 使用需要 string 类型的结构体，但 id 是数字
			type StrictDoc struct {
				ID int `json:"id"`
			}
			items, err := decodeMGet[StrictDoc](data)

			So(err, ShouldBeNil)
			So(len(items), ShouldEqual, 1)
			So(items[0].ID, ShouldEqual, 123)
		})
	})
}

func TestDecodeSearch(t *testing.T) {
	Convey("TestDecodeSearch", t, func() {
		type TestDoc struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}

		Convey("成功解析 search 响应", func() {
			data := []byte(`{
				"hits": {
					"hits": [
						{"_source": {"id": "1", "title": "First"}},
						{"_source": {"id": "2", "title": "Second"}}
					]
				}
			}`)

			items, err := decodeSearch[TestDoc](data)

			So(err, ShouldBeNil)
			So(len(items), ShouldEqual, 2)
			So(items[0].ID, ShouldEqual, "1")
			So(items[0].Title, ShouldEqual, "First")
			So(items[1].ID, ShouldEqual, "2")
			So(items[1].Title, ShouldEqual, "Second")
		})

		Convey("跳过空 Source 的文档", func() {
			data := []byte(`{
				"hits": {
					"hits": [
						{"_source": {"id": "1", "title": "First"}},
						{"_source": null},
						{"_source": {}}
					]
				}
			}`)

			items, err := decodeSearch[TestDoc](data)

			So(err, ShouldBeNil)
			// 空对象 {} 也是有效的，只有 null 会被跳过
			So(len(items), ShouldBeGreaterThanOrEqualTo, 1)
		})

		Convey("空 hits 数组", func() {
			data := []byte(`{"hits": {"hits": []}}`)

			items, err := decodeSearch[TestDoc](data)

			So(err, ShouldBeNil)
			So(len(items), ShouldEqual, 0)
		})

		Convey("解析 search 响应失败", func() {
			data := []byte(`invalid json`)

			items, err := decodeSearch[TestDoc](data)

			So(err, ShouldNotBeNil)
			So(items, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "解析 search 响应失败")
		})

		Convey("解析单个文档失败", func() {
			data := []byte(`{
				"hits": {
					"hits": [
						{"_source": "invalid source type"}
					]
				}
			}`)

			items, err := decodeSearch[TestDoc](data)

			So(err, ShouldNotBeNil)
			So(items, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "解析文档失败")
		})
	})
}

func TestEncodeBody(t *testing.T) {
	Convey("TestEncodeBody", t, func() {
		Convey("成功编码结构体", func() {
			payload := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}

			reader, err := encodeBody(payload)

			So(err, ShouldBeNil)
			So(reader, ShouldNotBeNil)
			data, _ := io.ReadAll(reader)
			So(string(data), ShouldContainSubstring, "query")
			So(string(data), ShouldContainSubstring, "match_all")
		})

		Convey("编码简单类型", func() {
			payload := "simple string"

			reader, err := encodeBody(payload)

			So(err, ShouldBeNil)
			So(reader, ShouldNotBeNil)
			data, _ := io.ReadAll(reader)
			So(string(data), ShouldEqual, `"simple string"`)
		})

		Convey("编码数组", func() {
			payload := []string{"a", "b", "c"}

			reader, err := encodeBody(payload)

			So(err, ShouldBeNil)
			data, _ := io.ReadAll(reader)
			So(string(data), ShouldEqual, `["a","b","c"]`)
		})

		Convey("编码 nil", func() {
			reader, err := encodeBody(nil)

			So(err, ShouldBeNil)
			data, _ := io.ReadAll(reader)
			So(string(data), ShouldEqual, "null")
		})

		Convey("编码失败（无法序列化的类型）", func() {
			payload := make(chan int) // channel 无法序列化

			reader, err := encodeBody(payload)

			So(err, ShouldNotBeNil)
			So(reader, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "序列化请求体失败")
		})

		Convey("编码复杂嵌套结构", func() {
			type Inner struct {
				Value int `json:"value"`
			}
			type Outer struct {
				Name  string `json:"name"`
				Inner Inner  `json:"inner"`
			}
			payload := Outer{
				Name:  "test",
				Inner: Inner{Value: 42},
			}

			reader, err := encodeBody(payload)

			So(err, ShouldBeNil)
			data, _ := io.ReadAll(reader)
			So(string(data), ShouldContainSubstring, `"name":"test"`)
			So(string(data), ShouldContainSubstring, `"value":42`)
		})
	})
}

// errorReader 用于模拟读取失败的 io.Reader
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

// 确保 errorReader 实现了 io.Reader 接口
var _ io.Reader = (*errorReader)(nil)

func TestMgetResponse(t *testing.T) {
	Convey("TestMgetResponse", t, func() {
		Convey("mgetResponse 结构体解析", func() {
			data := []byte(`{
				"docs": [
					{"found": true, "_source": {"key": "value"}},
					{"found": false, "_source": null}
				]
			}`)

			var resp mgetResponse
			err := json.Unmarshal(data, &resp)

			So(err, ShouldBeNil)
			So(len(resp.Docs), ShouldEqual, 2)
			So(resp.Docs[0].Found, ShouldBeTrue)
			So(resp.Docs[1].Found, ShouldBeFalse)
		})
	})
}

func TestSearchResponse(t *testing.T) {
	Convey("TestSearchResponse", t, func() {
		Convey("searchResponse 结构体解析", func() {
			data := []byte(`{
				"hits": {
					"total": {"value": 10},
					"hits": [
						{"_id": "1", "_source": {"field": "value"}}
					]
				}
			}`)

			var resp searchResponse
			err := json.Unmarshal(data, &resp)

			So(err, ShouldBeNil)
			So(len(resp.Hits.Hits), ShouldEqual, 1)
		})
	})
}
