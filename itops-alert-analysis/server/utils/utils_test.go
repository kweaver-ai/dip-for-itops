package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJsonEncode(t *testing.T) {
	Convey("TestJsonEncode", t, func() {
		Convey("编码 map 类型", func() {
			data := map[string]interface{}{
				"name": "test",
				"age":  25,
			}

			result := JsonEncode(data)

			So(result, ShouldContainSubstring, `"name":"test"`)
			So(result, ShouldContainSubstring, `"age":25`)
		})

		Convey("编码 struct 类型", func() {
			type Person struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}
			data := Person{Name: "Alice", Age: 30}

			result := JsonEncode(data)

			So(result, ShouldEqual, `{"name":"Alice","age":30}`)
		})

		Convey("编码数组类型", func() {
			data := []string{"a", "b", "c"}

			result := JsonEncode(data)

			So(result, ShouldEqual, `["a","b","c"]`)
		})

		Convey("编码 nil", func() {
			result := JsonEncode(nil)

			So(result, ShouldEqual, "null")
		})

		Convey("编码空 map", func() {
			data := map[string]interface{}{}

			result := JsonEncode(data)

			So(result, ShouldEqual, "{}")
		})

		Convey("编码空数组", func() {
			data := []int{}

			result := JsonEncode(data)

			So(result, ShouldEqual, "[]")
		})

		Convey("编码基本类型", func() {
			So(JsonEncode("hello"), ShouldEqual, `"hello"`)
			So(JsonEncode(123), ShouldEqual, "123")
			So(JsonEncode(true), ShouldEqual, "true")
			So(JsonEncode(3.14), ShouldEqual, "3.14")
		})

		Convey("编码嵌套结构", func() {
			data := map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			}

			result := JsonEncode(data)

			So(result, ShouldContainSubstring, `"outer"`)
			So(result, ShouldContainSubstring, `"inner":"value"`)
		})
	})
}
