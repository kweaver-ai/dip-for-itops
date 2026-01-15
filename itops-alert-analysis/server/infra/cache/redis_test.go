package cache

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	. "github.com/smartystreets/goconvey/convey"
)

// TestNewRedisCache_Standalone 测试 Standalone 模式创建
func TestNewRedisCache_Standalone(t *testing.T) {
	Convey("TestNewRedisCache_Standalone", t, func() {
		Convey("Standalone 模式创建成功", func() {
			db, mock := redismock.NewClientMock()
			mock.ExpectPing().SetVal("PONG")

			// 打桩 newStandaloneClient 返回 mock client
			patches := gomonkey.ApplyFunc(newStandaloneClient, func(cfg RedisConfig) *redis.Client {
				return db
			})
			defer patches.Reset()

			cfg := RedisConfig{
				Host:     "localhost:6379",
				Password: "test",
				DB:       0,
			}

			cache, err := NewRedisCache(cfg)
			So(err, ShouldBeNil)
			So(cache, ShouldNotBeNil)

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Standalone 模式 Ping 失败", func() {
			db, mock := redismock.NewClientMock()
			mock.ExpectPing().SetErr(redis.ErrClosed)

			patches := gomonkey.ApplyFunc(newStandaloneClient, func(cfg RedisConfig) *redis.Client {
				return db
			})
			defer patches.Reset()

			cfg := RedisConfig{
				Host: "localhost:6379",
			}

			cache, err := NewRedisCache(cfg)
			So(err, ShouldNotBeNil)
			So(cache, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "连接 redis 失败")
		})
	})
}

// TestNewRedisCache_Sentinel 测试 Sentinel 模式创建
func TestNewRedisCache_Sentinel(t *testing.T) {
	Convey("TestNewRedisCache_Sentinel", t, func() {
		Convey("Sentinel 模式创建成功", func() {
			db, mock := redismock.NewClientMock()
			mock.ExpectPing().SetVal("PONG")

			// 打桩 newSentinelClient 返回 mock client
			patches := gomonkey.ApplyFunc(newSentinelClient, func(cfg RedisConfig) redis.UniversalClient {
				return db
			})
			defer patches.Reset()

			cfg := RedisConfig{
				MasterName:    "mymaster",
				SentinelAddrs: []string{"localhost:26379"},
				Password:      "test",
				DB:            0,
			}

			cache, err := NewRedisCache(cfg)
			So(err, ShouldBeNil)
			So(cache, ShouldNotBeNil)

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Sentinel 模式 Ping 失败", func() {
			db, mock := redismock.NewClientMock()
			mock.ExpectPing().SetErr(redis.ErrClosed)

			patches := gomonkey.ApplyFunc(newSentinelClient, func(cfg RedisConfig) redis.UniversalClient {
				return db
			})
			defer patches.Reset()

			cfg := RedisConfig{
				MasterName:    "mymaster",
				SentinelAddrs: []string{"localhost:26379"},
			}

			cache, err := NewRedisCache(cfg)
			So(err, ShouldNotBeNil)
			So(cache, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "连接 redis 失败")
		})
	})
}

// TestRedisCache_Get 测试 Get 方法
func TestRedisCache_Get(t *testing.T) {
	Convey("TestRedisCache_Get", t, func() {
		db, mock := redismock.NewClientMock()
		cache := &RedisCache{client: db}
		ctx := context.Background()

		Convey("获取存在的 key", func() {
			mock.ExpectGet("test_key").SetVal("test_value")

			value, err := cache.Get(ctx, "test_key")
			So(err, ShouldBeNil)
			So(value, ShouldEqual, "test_value")
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("获取不存在的 key", func() {
			mock.ExpectGet("not_exist").RedisNil()

			value, err := cache.Get(ctx, "not_exist")
			So(err, ShouldNotBeNil)
			So(value, ShouldEqual, "")
			So(err.Error(), ShouldContainSubstring, "key not found")
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Redis 错误", func() {
			mock.ExpectGet("error_key").SetErr(redis.ErrClosed)

			value, err := cache.Get(ctx, "error_key")
			So(err, ShouldNotBeNil)
			So(value, ShouldEqual, "")
			So(err.Error(), ShouldContainSubstring, "redis get")
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// TestRedisCache_Set 测试 Set 方法
func TestRedisCache_Set(t *testing.T) {
	Convey("TestRedisCache_Set", t, func() {
		db, mock := redismock.NewClientMock()
		cache := &RedisCache{client: db}
		ctx := context.Background()

		Convey("设置成功", func() {
			mock.ExpectSet("test_key", "test_value", time.Hour).SetVal("OK")

			err := cache.Set(ctx, "test_key", "test_value", time.Hour)
			So(err, ShouldBeNil)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("设置失败", func() {
			mock.ExpectSet("error_key", "value", time.Hour).SetErr(redis.ErrClosed)

			err := cache.Set(ctx, "error_key", "value", time.Hour)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "redis set")
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("设置无过期时间", func() {
			mock.ExpectSet("no_expire_key", "value", 0).SetVal("OK")

			err := cache.Set(ctx, "no_expire_key", "value", 0)
			So(err, ShouldBeNil)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// TestRedisCache_Del 测试 Del 方法
func TestRedisCache_Del(t *testing.T) {
	Convey("TestRedisCache_Del", t, func() {
		db, mock := redismock.NewClientMock()
		cache := &RedisCache{client: db}
		ctx := context.Background()

		Convey("删除单个 key", func() {
			mock.ExpectDel("key1").SetVal(1)

			err := cache.Del(ctx, "key1")
			So(err, ShouldBeNil)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("删除多个 key", func() {
			mock.ExpectDel("key1", "key2", "key3").SetVal(3)

			err := cache.Del(ctx, "key1", "key2", "key3")
			So(err, ShouldBeNil)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("删除空 key 列表", func() {
			err := cache.Del(ctx)
			So(err, ShouldBeNil)
			// 不应该有 Redis 调用
		})

		Convey("删除失败", func() {
			mock.ExpectDel("error_key").SetErr(redis.ErrClosed)

			err := cache.Del(ctx, "error_key")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "redis del")
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// TestRedisCache_Exists 测试 Exists 方法
func TestRedisCache_Exists(t *testing.T) {
	Convey("TestRedisCache_Exists", t, func() {
		db, mock := redismock.NewClientMock()
		cache := &RedisCache{client: db}
		ctx := context.Background()

		Convey("key 存在", func() {
			mock.ExpectExists("exist_key").SetVal(1)

			exists, err := cache.Exists(ctx, "exist_key")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("key 不存在", func() {
			mock.ExpectExists("not_exist_key").SetVal(0)

			exists, err := cache.Exists(ctx, "not_exist_key")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Redis 错误", func() {
			mock.ExpectExists("error_key").SetErr(redis.ErrClosed)

			exists, err := cache.Exists(ctx, "error_key")
			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(err.Error(), ShouldContainSubstring, "redis exists")
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// TestRedisCache_Close 测试 Close 方法
func TestRedisCache_Close(t *testing.T) {
	Convey("TestRedisCache_Close", t, func() {
		db, _ := redismock.NewClientMock()
		cache := &RedisCache{client: db}

		Convey("关闭成功", func() {
			err := cache.Close()
			So(err, ShouldBeNil)
		})
	})
}
