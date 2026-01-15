package slice

import (
	"strings"

	"github.com/spf13/cast"
)

// AppendUniqueUint64 向 uint64 切片追加元素，如果元素已存在则不追加。
func AppendUniqueUint64(list []uint64, v uint64) []uint64 {
	for _, item := range list {
		if item == v {
			return list
		}
	}
	return append(list, v)
}

// AppendUniqueString 向 string 切片追加元素，如果元素已存在则不追加。
func AppendUniqueString(list []string, v string) []string {
	for _, item := range list {
		if item == v {
			return list
		}
	}
	return append(list, v)
}

// ContainsUint64 检查 uint64 切片是否包含指定元素。
func ContainsUint64(list []uint64, v uint64) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// ContainsString 检查 string 切片是否包含指定元素。
func ContainsString(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// SplitToStrings 将逗号分隔的字符串解析为字符串切片。
func SplitToStrings(value string) []string {
	var result []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			result = append(result, part)
		}
	}
	return result
}

// SplitToUint64s 将逗号分隔的字符串解析为 uint64 切片。
func SplitToUint64s(value string) []uint64 {
	var result []uint64
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			id := cast.ToUint64(part)
			if id != 0 {
				result = append(result, id)
			}
		}
	}
	return result
}
