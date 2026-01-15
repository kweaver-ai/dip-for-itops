package slice

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAppendUniqueUint64(t *testing.T) {
	Convey("TestAppendUniqueUint64", t, func() {
		Convey("追加不存在的元素", func() {
			list := []uint64{1, 2, 3}

			result := AppendUniqueUint64(list, 4)

			So(len(result), ShouldEqual, 4)
			So(result[3], ShouldEqual, 4)
		})

		Convey("追加已存在的元素", func() {
			list := []uint64{1, 2, 3}

			result := AppendUniqueUint64(list, 2)

			So(len(result), ShouldEqual, 3)
		})

		Convey("追加到空切片", func() {
			var list []uint64

			result := AppendUniqueUint64(list, 1)

			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, 1)
		})

		Convey("追加零值", func() {
			list := []uint64{1, 2, 3}

			result := AppendUniqueUint64(list, 0)

			So(len(result), ShouldEqual, 4)
			So(result[3], ShouldEqual, 0)
		})

		Convey("追加零值到已包含零值的切片", func() {
			list := []uint64{0, 1, 2}

			result := AppendUniqueUint64(list, 0)

			So(len(result), ShouldEqual, 3)
		})
	})
}

func TestAppendUniqueString(t *testing.T) {
	Convey("TestAppendUniqueString", t, func() {
		Convey("追加不存在的元素", func() {
			list := []string{"a", "b", "c"}

			result := AppendUniqueString(list, "d")

			So(len(result), ShouldEqual, 4)
			So(result[3], ShouldEqual, "d")
		})

		Convey("追加已存在的元素", func() {
			list := []string{"a", "b", "c"}

			result := AppendUniqueString(list, "b")

			So(len(result), ShouldEqual, 3)
		})

		Convey("追加到空切片", func() {
			var list []string

			result := AppendUniqueString(list, "a")

			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "a")
		})

		Convey("追加空字符串", func() {
			list := []string{"a", "b"}

			result := AppendUniqueString(list, "")

			So(len(result), ShouldEqual, 3)
			So(result[2], ShouldEqual, "")
		})

		Convey("追加空字符串到已包含空字符串的切片", func() {
			list := []string{"", "a", "b"}

			result := AppendUniqueString(list, "")

			So(len(result), ShouldEqual, 3)
		})
	})
}

func TestContainsUint64(t *testing.T) {
	Convey("TestContainsUint64", t, func() {
		Convey("包含指定元素", func() {
			list := []uint64{1, 2, 3, 4, 5}

			So(ContainsUint64(list, 3), ShouldBeTrue)
		})

		Convey("不包含指定元素", func() {
			list := []uint64{1, 2, 3, 4, 5}

			So(ContainsUint64(list, 6), ShouldBeFalse)
		})

		Convey("空切片", func() {
			var list []uint64

			So(ContainsUint64(list, 1), ShouldBeFalse)
		})

		Convey("检查第一个元素", func() {
			list := []uint64{1, 2, 3}

			So(ContainsUint64(list, 1), ShouldBeTrue)
		})

		Convey("检查最后一个元素", func() {
			list := []uint64{1, 2, 3}

			So(ContainsUint64(list, 3), ShouldBeTrue)
		})

		Convey("检查零值", func() {
			list := []uint64{0, 1, 2}

			So(ContainsUint64(list, 0), ShouldBeTrue)
		})
	})
}

func TestContainsString(t *testing.T) {
	Convey("TestContainsString", t, func() {
		Convey("包含指定元素", func() {
			list := []string{"apple", "banana", "cherry"}

			So(ContainsString(list, "banana"), ShouldBeTrue)
		})

		Convey("不包含指定元素", func() {
			list := []string{"apple", "banana", "cherry"}

			So(ContainsString(list, "orange"), ShouldBeFalse)
		})

		Convey("空切片", func() {
			var list []string

			So(ContainsString(list, "apple"), ShouldBeFalse)
		})

		Convey("检查空字符串", func() {
			list := []string{"", "a", "b"}

			So(ContainsString(list, ""), ShouldBeTrue)
		})

		Convey("大小写敏感", func() {
			list := []string{"Apple", "Banana"}

			So(ContainsString(list, "apple"), ShouldBeFalse)
			So(ContainsString(list, "Apple"), ShouldBeTrue)
		})
	})
}

func TestSplitToStrings(t *testing.T) {
	Convey("TestSplitToStrings", t, func() {
		Convey("正常分割", func() {
			result := SplitToStrings("a,b,c")

			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, "a")
			So(result[1], ShouldEqual, "b")
			So(result[2], ShouldEqual, "c")
		})

		Convey("带空格的分割", func() {
			result := SplitToStrings("a , b , c")

			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, "a")
			So(result[1], ShouldEqual, "b")
			So(result[2], ShouldEqual, "c")
		})

		Convey("过滤空字符串", func() {
			result := SplitToStrings("a,,b,  ,c")

			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, "a")
			So(result[1], ShouldEqual, "b")
			So(result[2], ShouldEqual, "c")
		})

		Convey("空字符串输入", func() {
			result := SplitToStrings("")

			So(result, ShouldBeNil)
		})

		Convey("只有空格的输入", func() {
			result := SplitToStrings("   ")

			So(result, ShouldBeNil)
		})

		Convey("单个元素", func() {
			result := SplitToStrings("single")

			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "single")
		})

		Convey("前后有空格", func() {
			result := SplitToStrings("  first , last  ")

			So(len(result), ShouldEqual, 2)
			So(result[0], ShouldEqual, "first")
			So(result[1], ShouldEqual, "last")
		})
	})
}

func TestSplitToUint64s(t *testing.T) {
	Convey("TestSplitToUint64s", t, func() {
		Convey("正常分割", func() {
			result := SplitToUint64s("1,2,3")

			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, uint64(1))
			So(result[1], ShouldEqual, uint64(2))
			So(result[2], ShouldEqual, uint64(3))
		})

		Convey("带空格的分割", func() {
			result := SplitToUint64s("1 , 2 , 3")

			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, uint64(1))
			So(result[1], ShouldEqual, uint64(2))
			So(result[2], ShouldEqual, uint64(3))
		})

		Convey("过滤空字符串和零值", func() {
			result := SplitToUint64s("1,,2,  ,3")

			So(len(result), ShouldEqual, 3)
		})

		Convey("过滤非数字字符串", func() {
			result := SplitToUint64s("1,abc,2,def,3")

			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, uint64(1))
			So(result[1], ShouldEqual, uint64(2))
			So(result[2], ShouldEqual, uint64(3))
		})

		Convey("空字符串输入", func() {
			result := SplitToUint64s("")

			So(result, ShouldBeNil)
		})

		Convey("单个元素", func() {
			result := SplitToUint64s("123")

			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, uint64(123))
		})

		Convey("大数值", func() {
			result := SplitToUint64s("18446744073709551615")

			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, uint64(18446744073709551615))
		})

		Convey("过滤零值", func() {
			result := SplitToUint64s("0,1,0,2")

			// 0 会被过滤掉因为 cast.ToUint64("0") == 0 且 if id != 0 会跳过
			So(len(result), ShouldEqual, 2)
			So(result[0], ShouldEqual, uint64(1))
			So(result[1], ShouldEqual, uint64(2))
		})
	})
}
