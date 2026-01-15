package idgen

import (
	"sync"
	"time"
)

const (
	// 自定义纪元：2025-01-01 00:00:00 UTC（Unix 时间戳：秒）
	// 从这个时间开始计算，可以显著缩短 ID 长度
	customEpoch = 1735689600 // time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	// 序列号位数：6位（每秒最多生成 64 个 ID）
	// 对于告警场景，每秒 64 个 ID 足够使用
	seqBits = 6
	seqMask = (1 << seqBits) - 1 // 63
)

type Generator struct {
	mu     sync.Mutex // 保护并发访问
	lastTs int64      // 上一次生成 ID 的时间戳（秒）
	seq    int64      // 当前秒内的序列号（0-63）
}

// New 创建 ID 生成器实例。
func New() *Generator {
	return &Generator{}
}

func (g *Generator) NextID() uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 获取当前时间戳（秒，相对于自定义纪元）
	ts := time.Now().Unix() - customEpoch

	if ts == g.lastTs {
		// 同一秒内，序列号递增
		g.seq = (g.seq + 1) & seqMask
		if g.seq == 0 {
			// 当前秒的序列号用完了（超过64个），等待下一秒
			for ts <= g.lastTs {
				time.Sleep(time.Millisecond)
				ts = time.Now().Unix() - customEpoch
			}
		}
	} else {
		// 新的一秒，重置序列号
		g.seq = 0
	}

	g.lastTs = ts

	// 生成 ID：时间戳左移6位 + 序列号
	return uint64(ts)<<seqBits | uint64(g.seq)
}

// NewWithCache 为了向后兼容保留的方法（但不使用 cache）。
// 实际上会忽略 cache 参数，使用纯时间戳方案。
func NewWithCache(cache interface{}) *Generator {
	return New()
}

// SetCache 为了向后兼容保留的方法（但不使用 cache）。
func (g *Generator) SetCache(cache interface{}) {
	// 纯时间戳方案不需要 cache，此方法为空实现
}
