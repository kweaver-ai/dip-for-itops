package cache

import (
	"context"
	"time"
)

// Cache 缓存接口,将来可以适配多种缓存
type Cache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) (string, error)

	// Set 设置缓存值，expiration 为 0 表示永不过期
	Set(ctx context.Context, key string, value string, expiration time.Duration) error

	// Del 删除缓存键
	Del(ctx context.Context, keys ...string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Close 关闭缓存连接
	Close() error
}
