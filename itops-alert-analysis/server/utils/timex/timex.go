package timex

import (
	"math"
	"time"
)

func NowLocalTime() time.Time {
	return time.Now().Local()
}

func ParseTime(s string, f string) (time.Time, error) {
	t, err := time.ParseInLocation(f, s, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// AbsSecondsBetween 计算 t1 和 t2 之间的绝对秒差值
func AbsSecondsBetween(t1, t2 time.Time) uint64 {
	return uint64(math.Abs(t1.Sub(t2).Seconds()))
}
