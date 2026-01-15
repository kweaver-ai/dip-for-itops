package idgen

import (
	"sync"
	"testing"
	"time"
)

// TestIDGen_Next 测试 ID 生成器的基本功能
func TestIDGen_Next(t *testing.T) {
	gen := New()

	// 生成 100 个 ID
	ids := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		id := gen.NextID()
		if id == 0 {
			t.Fatalf("生成的 ID 为 0，第 %d 次", i+1)
		}
		if ids[id] {
			t.Fatalf("ID 重复: %d", id)
		}
		//t.Log(id)
		ids[id] = true
	}

	t.Logf("成功生成 100 个唯一的 ID")
}

// TestIDGen_Concurrent 测试并发生成 ID
func TestIDGen_Concurrent(t *testing.T) {
	gen := New()

	const goroutines = 10
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	idChan := make(chan uint64, goroutines*idsPerGoroutine)

	// 启动多个 goroutine 并发生成 ID
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id := gen.NextID()
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// 检查是否有重复的 ID
	ids := make(map[uint64]bool)
	for id := range idChan {
		if id == 0 {
			t.Fatalf("生成的 ID 为 0")
		}
		if ids[id] {
			t.Fatalf("并发环境下 ID 重复: %d", id)
		}
		ids[id] = true
	}

	expected := goroutines * idsPerGoroutine
	if len(ids) != expected {
		t.Fatalf("期望生成 %d 个 ID，实际生成 %d 个", expected, len(ids))
	}

	t.Logf("成功并发生成 %d 个唯一的 ID", len(ids))
}

// TestIDGen_Incremental 测试 ID 是否递增
func TestIDGen_Incremental(t *testing.T) {
	gen := New()

	var lastID uint64
	for i := 0; i < 200; i++ {
		id := gen.NextID()
		if id == 0 {
			t.Fatalf("生成的 ID 为 0，第 %d 次", i+1)
		}
		if i > 0 && id <= lastID {
			t.Fatalf("ID 不是递增的: 上一个 %d，当前 %d", lastID, id)
		}
		lastID = id
	}

	t.Logf("成功验证 ID 递增，最后一个 ID: %d", lastID)
}

// TestIDGen_Restart 模拟重启后 ID 继续递增
func TestIDGen_Restart(t *testing.T) {
	// 第一次启动
	gen1 := New()

	// 生成多个 ID
	var lastID uint64
	for i := 0; i < 10; i++ {
		lastID = gen1.NextID()
	}

	// 等待至少 1 秒，确保时间推进到下一秒
	// 这样才能真正模拟"重启"场景（时间已经流逝）
	time.Sleep(1100 * time.Millisecond)

	// 模拟重启（创建新实例）
	// 由于时间已经推进，新 ID 必然大于重启前的 ID
	gen2 := New()
	newID := gen2.NextID()

	// 重启后的 ID 应该 > 重启前（时间已推进）
	if newID <= lastID {
		t.Fatalf("重启后 ID 应该递增: 重启前=%d, 重启后=%d", lastID, newID)
	}

	t.Logf("重启验证成功: 重启前最后ID=%d, 重启后首ID=%d", lastID, newID)
}

// BenchmarkIDGen_NextID 性能测试
func BenchmarkIDGen_NextID(b *testing.B) {
	gen := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.NextID()
	}
}

// BenchmarkIDGen_Concurrent 并发性能测试
func BenchmarkIDGen_Concurrent(b *testing.B) {
	gen := New()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gen.NextID()
		}
	})
}
