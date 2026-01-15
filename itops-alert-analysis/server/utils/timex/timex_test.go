package timex

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNowLocalTime(t *testing.T) {
	Convey("TestNowLocalTime", t, func() {
		Convey("返回本地时间", func() {
			before := time.Now().Local()
			result := NowLocalTime()
			after := time.Now().Local()

			So(result.Location(), ShouldEqual, time.Local)
			So(result.After(before) || result.Equal(before), ShouldBeTrue)
			So(result.Before(after) || result.Equal(after), ShouldBeTrue)
		})
	})
}

func TestParseTime(t *testing.T) {
	Convey("TestParseTime", t, func() {
		Convey("解析标准日期时间格式", func() {
			result, err := ParseTime("2025-01-15 10:30:00", "2006-01-02 15:04:05")

			So(err, ShouldBeNil)
			So(result.Year(), ShouldEqual, 2025)
			So(result.Month(), ShouldEqual, time.January)
			So(result.Day(), ShouldEqual, 15)
			So(result.Hour(), ShouldEqual, 10)
			So(result.Minute(), ShouldEqual, 30)
			So(result.Second(), ShouldEqual, 0)
		})

		Convey("解析日期格式", func() {
			result, err := ParseTime("2025-06-20", "2006-01-02")

			So(err, ShouldBeNil)
			So(result.Year(), ShouldEqual, 2025)
			So(result.Month(), ShouldEqual, time.June)
			So(result.Day(), ShouldEqual, 20)
		})

		Convey("解析时间格式", func() {
			result, err := ParseTime("14:30:45", "15:04:05")

			So(err, ShouldBeNil)
			So(result.Hour(), ShouldEqual, 14)
			So(result.Minute(), ShouldEqual, 30)
			So(result.Second(), ShouldEqual, 45)
		})

		Convey("解析 RFC3339 格式", func() {
			result, err := ParseTime("2025-01-15T10:30:00", "2006-01-02T15:04:05")

			So(err, ShouldBeNil)
			So(result.Year(), ShouldEqual, 2025)
		})

		Convey("格式不匹配返回错误", func() {
			result, err := ParseTime("2025-01-15", "2006/01/02")

			So(err, ShouldNotBeNil)
			So(result.IsZero(), ShouldBeTrue)
		})

		Convey("无效日期返回错误", func() {
			result, err := ParseTime("2025-13-45", "2006-01-02")

			So(err, ShouldNotBeNil)
			So(result.IsZero(), ShouldBeTrue)
		})

		Convey("空字符串返回错误", func() {
			result, err := ParseTime("", "2006-01-02")

			So(err, ShouldNotBeNil)
			So(result.IsZero(), ShouldBeTrue)
		})

		Convey("返回本地时区时间", func() {
			result, err := ParseTime("2025-01-15 10:30:00", "2006-01-02 15:04:05")

			So(err, ShouldBeNil)
			So(result.Location(), ShouldEqual, time.Local)
		})
	})
}

func TestAbsSecondsBetween(t *testing.T) {
	tests := []struct {
		name string
		t1   time.Time
		t2   time.Time
		want uint64
	}{
		{
			name: "t1 > t2",
			t1:   time.Date(2025, 1, 1, 12, 0, 30, 0, time.Local),
			t2:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.Local),
			want: 30,
		},
		{
			name: "t1 < t2",
			t1:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.Local),
			t2:   time.Date(2025, 1, 1, 12, 0, 30, 0, time.Local),
			want: 30,
		},
		{
			name: "t1 == t2",
			t1:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.Local),
			t2:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.Local),
			want: 0,
		},
		{
			name: "large difference",
			t1:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
			t2:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.Local),
			want: 86400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AbsSecondsBetween(tt.t1, tt.t2)
			if got != tt.want {
				t.Errorf("AbsSecondsBetween() = %v, want %v", got, tt.want)
			}
		})
	}
}
