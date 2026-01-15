package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// RedisCache redis客户端
type RedisCache struct {
	client redis.UniversalClient
}

// RedisConfig Redis 配置，支持 Standalone 和 Sentinel 两种模式。
// 如果配置了 MasterName 和 SentinelAddrs，则使用 Sentinel 模式；否则使用 Standalone 模式。
type RedisConfig struct {
	Host string `yaml:"host"` // Redis 地址（如 "localhost:6379"）

	Username string `yaml:"username"` // Redis 密码
	Password string `yaml:"password"` // Redis 密码
	DB       int    `yaml:"db"`       // 数据库索引

	// Sentinel 模式配置
	MasterName       string   `yaml:"master_name"`       // Sentinel 主节点名称
	SentinelAddrs    []string `yaml:"sentinel_addrs"`    // Sentinel 地址列表
	SentinelUsername string   `yaml:"sentinel_username"` // Sentinel 用户名（Redis 6.2+）
	SentinelPassword string   `yaml:"sentinel_password"` // Sentinel 密码
}

// NewRedisCache 创建 Redis 实例
func NewRedisCache(cfg RedisConfig) (Cache, error) {
	var client redis.UniversalClient

	if cfg.MasterName != "" && len(cfg.SentinelAddrs) > 0 {
		client = newSentinelClient(cfg)
	} else {
		client = newStandaloneClient(cfg)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, errors.Wrap(err, "连接 redis 失败")
	}

	return &RedisCache{client: client}, nil
}

// newSentinelClient Sentinel 模式
func newSentinelClient(cfg RedisConfig) redis.UniversalClient {
	return redis.NewFailoverClient(&redis.FailoverOptions{
		// 基础配置
		MasterName:       cfg.MasterName,
		SentinelAddrs:    cfg.SentinelAddrs,
		SentinelUsername: cfg.SentinelUsername,
		SentinelPassword: cfg.SentinelPassword,
		Username:         cfg.Username,
		Password:         cfg.Password,
		DB:               cfg.DB,

		// 连接池配置
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,

		// 超时配置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})
}

// newStandaloneClient Standalone 模式
func newStandaloneClient(cfg RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		// 基础配置
		Addr:     cfg.Host,
		Password: cfg.Password,
		DB:       cfg.DB,

		// 连接池配置
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,

		// 超时配置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})
}

// Get 获取缓存值。
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", errors.Errorf("key not found: %s", key)
	}
	if err != nil {
		return "", errors.Wrap(err, "redis get")
	}
	return value, nil
}

// Set 设置缓存值。
func (r *RedisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return errors.Wrap(err, "redis set")
	}
	return nil
}

// Del 删除缓存键。
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return errors.Wrap(err, "redis del")
	}
	return nil
}

// Exists 检查键是否存在。
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, errors.Wrap(err, "redis exists")
	}
	return n > 0, nil
}

// Close 关闭 Redis 连接。
func (r *RedisCache) Close() error {
	return r.client.Close()
}
