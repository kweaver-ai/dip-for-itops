package objectclass

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/cache"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/dip"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

// mockCache 实现 cache.Cache 接口
type mockCache struct {
	data      map[string]string
	getErr    error
	setErr    error
	delErr    error
	closeErr  error
	existsErr error
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]string),
	}
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return "", errors.New("key not found")
}

func (m *mockCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}

func (m *mockCache) Del(ctx context.Context, keys ...string) error {
	if m.delErr != nil {
		return m.delErr
	}
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *mockCache) Exists(ctx context.Context, key string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	_, ok := m.data[key]
	return ok, nil
}

func (m *mockCache) Close() error {
	return m.closeErr
}

// 确保 mockCache 实现了 cache.Cache 接口
var _ cache.Cache = (*mockCache)(nil)

func TestNew(t *testing.T) {
	Convey("TestNew", t, func() {
		cfg := &config.Config{
			DepServices: config.DepServicesConfig{
				Redis: config.DepRedisConfig{
					ConnectInfo: config.RedisConnectInfo{
						MasterGroupName:  "mymaster",
						SentinelHost:     "localhost",
						SentinelPort:     26379,
						SentinelUsername: "",
						SentinelPassword: "",
						Username:         "",
						Password:         "",
					},
				},
			},
		}
		dipClient := &dip.Client{}

		Convey("成功创建 ObjectClass", func() {
			mockCacheInstance := newMockCache()

			patches := gomonkey.ApplyFunc(cache.NewRedisCache, func(cfg cache.RedisConfig) (cache.Cache, error) {
				return mockCacheInstance, nil
			})
			defer patches.Reset()

			oc, err := New(cfg, dipClient)

			So(err, ShouldBeNil)
			So(oc, ShouldNotBeNil)
			So(oc.dipClient, ShouldEqual, dipClient)
			So(oc.cache, ShouldEqual, mockCacheInstance)
		})

		Convey("Redis 缓存初始化失败返回错误", func() {
			patches := gomonkey.ApplyFunc(cache.NewRedisCache, func(cfg cache.RedisConfig) (cache.Cache, error) {
				return nil, errors.New("redis connection failed")
			})
			defer patches.Reset()

			oc, err := New(cfg, dipClient)

			So(err, ShouldNotBeNil)
			So(oc, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "初始化 Redis 缓存失败")
		})
	})
}

func TestObjectClass_Warmup(t *testing.T) {
	Convey("TestObjectClass_Warmup", t, func() {
		ctx := context.Background()
		mockCacheInstance := newMockCache()
		dipClient := &dip.Client{}

		oc := &ObjectClass{
			dipClient: dipClient,
			cache:     mockCacheInstance,
		}

		Convey("成功预热缓存", func() {
			// 打桩 GetObjectTypes
			patches := gomonkey.ApplyMethod(dipClient, "GetObjectTypes",
				func(_ *dip.Client, ctx context.Context) ([]dip.ObjectType, error) {
					return []dip.ObjectType{
						{ID: "ot-1", Name: "Pod"},
						{ID: "ot-2", Name: "Node"},
					}, nil
				})
			defer patches.Reset()

			// 打桩 QueryAllObjectData
			patches.ApplyMethod(dipClient, "QueryAllObjectData",
				func(_ *dip.Client, ctx context.Context, otID string, limit int) ([]dip.ObjectInstance, error) {
					if otID == "ot-1" {
						return []dip.ObjectInstance{
							{"s_id": "obj-1", "k8s_cluster": "cluster1", "namespace": "ns1", "name": "pod1"},
							{"s_id": "obj-2", "k8s_cluster": "cluster1", "namespace": "ns2", "name": "pod2"},
						}, nil
					}
					return []dip.ObjectInstance{
						{"s_id": "obj-3", "k8s_cluster": "cluster1", "namespace": "", "name": "node1"},
					}, nil
				})

			err := oc.Warmup(ctx)

			So(err, ShouldBeNil)
			// 验证缓存数据
			So(len(mockCacheInstance.data), ShouldEqual, 3)
		})

		Convey("获取对象类列表失败返回错误", func() {
			patches := gomonkey.ApplyMethod(dipClient, "GetObjectTypes",
				func(_ *dip.Client, ctx context.Context) ([]dip.ObjectType, error) {
					return nil, errors.New("network error")
				})
			defer patches.Reset()

			err := oc.Warmup(ctx)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "获取对象类列表失败")
		})

		Convey("单个对象类预热失败不影响其他对象类", func() {
			patches := gomonkey.ApplyMethod(dipClient, "GetObjectTypes",
				func(_ *dip.Client, ctx context.Context) ([]dip.ObjectType, error) {
					return []dip.ObjectType{
						{ID: "ot-1", Name: "Pod"},
						{ID: "ot-2", Name: "Node"},
					}, nil
				})
			defer patches.Reset()

			callCount := 0
			patches.ApplyMethod(dipClient, "QueryAllObjectData",
				func(_ *dip.Client, ctx context.Context, otID string, limit int) ([]dip.ObjectInstance, error) {
					callCount++
					if otID == "ot-1" {
						return nil, errors.New("query failed")
					}
					return []dip.ObjectInstance{
						{"s_id": "obj-1", "k8s_cluster": "c1", "namespace": "ns1", "name": "node1"},
					}, nil
				})

			err := oc.Warmup(ctx)

			So(err, ShouldBeNil)
			So(callCount, ShouldEqual, 2)                   // 两个对象类都被调用
			So(len(mockCacheInstance.data), ShouldEqual, 1) // 只有 ot-2 成功
		})
	})
}

func TestObjectClass_warmupObjectType(t *testing.T) {
	Convey("TestObjectClass_warmupObjectType", t, func() {
		ctx := context.Background()
		mockCacheInstance := newMockCache()
		dipClient := &dip.Client{}

		oc := &ObjectClass{
			dipClient: dipClient,
			cache:     mockCacheInstance,
		}

		Convey("成功预热单个对象类", func() {
			patches := gomonkey.ApplyMethod(dipClient, "QueryAllObjectData",
				func(_ *dip.Client, ctx context.Context, otID string, limit int) ([]dip.ObjectInstance, error) {
					return []dip.ObjectInstance{
						{"s_id": "obj-1", "k8s_cluster": "cluster1", "namespace": "ns1", "name": "pod1"},
					}, nil
				})
			defer patches.Reset()

			err := oc.warmupObjectType(ctx, "ot-1")

			So(err, ShouldBeNil)
			// 验证缓存键
			expectedKey := cacheKeyPrefix + "k8s_cluster:cluster1,namespace:ns1,name:pod1"
			So(mockCacheInstance.data[expectedKey], ShouldNotBeEmpty)

			// 验证缓存值
			var objInfo EntityObjectInfo
			json.Unmarshal([]byte(mockCacheInstance.data[expectedKey]), &objInfo)
			So(objInfo.ObjectTypeID, ShouldEqual, "ot-1")
			So(objInfo.ObjectID, ShouldEqual, "obj-1")
			So(objInfo.Name, ShouldEqual, "pod1")
		})

		Convey("跳过没有 s_id 的对象", func() {
			patches := gomonkey.ApplyMethod(dipClient, "QueryAllObjectData",
				func(_ *dip.Client, ctx context.Context, otID string, limit int) ([]dip.ObjectInstance, error) {
					return []dip.ObjectInstance{
						{"k8s_cluster": "cluster1", "namespace": "ns1", "name": "pod1"},             // 没有 s_id
						{"s_id": "", "k8s_cluster": "cluster1", "namespace": "ns2", "name": "pod2"}, // s_id 为空
						{"s_id": "obj-3", "k8s_cluster": "cluster1", "namespace": "ns3", "name": "pod3"},
					}, nil
				})
			defer patches.Reset()

			err := oc.warmupObjectType(ctx, "ot-1")

			So(err, ShouldBeNil)
			So(len(mockCacheInstance.data), ShouldEqual, 1) // 只有一个有效对象
		})

		Convey("查询对象数据失败返回错误", func() {
			patches := gomonkey.ApplyMethod(dipClient, "QueryAllObjectData",
				func(_ *dip.Client, ctx context.Context, otID string, limit int) ([]dip.ObjectInstance, error) {
					return nil, errors.New("query failed")
				})
			defer patches.Reset()

			err := oc.warmupObjectType(ctx, "ot-1")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "查询对象数据失败")
		})

		Convey("缓存写入失败继续处理", func() {
			mockCacheInstance.setErr = errors.New("cache set failed")

			patches := gomonkey.ApplyMethod(dipClient, "QueryAllObjectData",
				func(_ *dip.Client, ctx context.Context, otID string, limit int) ([]dip.ObjectInstance, error) {
					return []dip.ObjectInstance{
						{"s_id": "obj-1", "k8s_cluster": "cluster1", "namespace": "ns1", "name": "pod1"},
					}, nil
				})
			defer patches.Reset()

			err := oc.warmupObjectType(ctx, "ot-1")

			So(err, ShouldBeNil) // 缓存失败不影响返回结果
		})

		Convey("处理空字段值", func() {
			patches := gomonkey.ApplyMethod(dipClient, "QueryAllObjectData",
				func(_ *dip.Client, ctx context.Context, otID string, limit int) ([]dip.ObjectInstance, error) {
					return []dip.ObjectInstance{
						{"s_id": "obj-1", "name": "pod1"}, // k8s_cluster 和 namespace 不存在
					}, nil
				})
			defer patches.Reset()

			err := oc.warmupObjectType(ctx, "ot-1")

			So(err, ShouldBeNil)
			// 验证缓存键包含空值
			expectedKey := cacheKeyPrefix + "k8s_cluster:,namespace:,name:pod1"
			So(mockCacheInstance.data[expectedKey], ShouldNotBeEmpty)
		})
	})
}

func TestObjectClass_GetEntityObjectInfo(t *testing.T) {
	Convey("TestObjectClass_GetEntityObjectInfo", t, func() {
		ctx := context.Background()
		mockCacheInstance := newMockCache()

		oc := &ObjectClass{
			cache: mockCacheInstance,
		}

		Convey("hostname 为空返回错误", func() {
			result, err := oc.GetEntityObjectInfo(ctx, "")

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "hostname 不能为空")
		})

		Convey("成功获取对象信息", func() {
			// 预设缓存数据
			objInfo := EntityObjectInfo{
				ObjectTypeID: "ot-1",
				ObjectID:     "obj-1",
				Name:         "pod1",
			}
			jsonData, _ := json.Marshal(objInfo)
			hostname := "k8s_cluster:cluster1,namespace:ns1,name:pod1"
			mockCacheInstance.data[cacheKeyPrefix+hostname] = string(jsonData)

			result, err := oc.GetEntityObjectInfo(ctx, hostname)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.ObjectTypeID, ShouldEqual, "ot-1")
			So(result.ObjectID, ShouldEqual, "obj-1")
			So(result.Name, ShouldEqual, "pod1")
		})

		Convey("缓存未命中返回错误", func() {
			result, err := oc.GetEntityObjectInfo(ctx, "nonexistent-hostname")

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "未找到 hostname 对应的对象信息")
		})

		Convey("缓存数据反序列化失败返回错误", func() {
			hostname := "invalid-json-hostname"
			mockCacheInstance.data[cacheKeyPrefix+hostname] = "invalid-json"

			result, err := oc.GetEntityObjectInfo(ctx, hostname)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "反序列化对象信息失败")
		})
	})
}

func TestObjectClass_Close(t *testing.T) {
	Convey("TestObjectClass_Close", t, func() {
		Convey("cache 为 nil 返回 nil", func() {
			oc := &ObjectClass{cache: nil}

			err := oc.Close()

			So(err, ShouldBeNil)
		})

		Convey("成功关闭缓存", func() {
			mockCacheInstance := newMockCache()
			oc := &ObjectClass{cache: mockCacheInstance}

			err := oc.Close()

			So(err, ShouldBeNil)
		})

		Convey("关闭缓存失败返回错误", func() {
			mockCacheInstance := newMockCache()
			mockCacheInstance.closeErr = errors.New("close failed")
			oc := &ObjectClass{cache: mockCacheInstance}

			err := oc.Close()

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "close failed")
		})
	})
}

func TestObjectClass_Run(t *testing.T) {
	Convey("TestObjectClass_Run", t, func() {
		mockCacheInstance := newMockCache()
		dipClient := &dip.Client{}

		oc := &ObjectClass{
			dipClient: dipClient,
			cache:     mockCacheInstance,
		}

		Convey("context 取消时退出运行", func() {
			// 打桩 GetObjectTypes 返回空列表（快速完成预热）
			patches := gomonkey.ApplyMethod(dipClient, "GetObjectTypes",
				func(_ *dip.Client, ctx context.Context) ([]dip.ObjectType, error) {
					return []dip.ObjectType{}, nil
				})
			defer patches.Reset()

			ctx, cancel := context.WithCancel(context.Background())

			// 在另一个 goroutine 中运行
			done := make(chan error, 1)
			go func() {
				done <- oc.Run(ctx)
			}()

			// 短暂等待后取消 context
			time.Sleep(100 * time.Millisecond)
			cancel()

			// 等待 Run 返回
			select {
			case err := <-done:
				So(err, ShouldEqual, context.Canceled)
			case <-time.After(2 * time.Second):
				t.Fatal("Run 没有在预期时间内退出")
			}
		})

		Convey("初始预热失败不阻止运行", func() {
			callCount := 0
			patches := gomonkey.ApplyMethod(dipClient, "GetObjectTypes",
				func(_ *dip.Client, ctx context.Context) ([]dip.ObjectType, error) {
					callCount++
					if callCount == 1 {
						return nil, errors.New("initial warmup failed")
					}
					return []dip.ObjectType{}, nil
				})
			defer patches.Reset()

			ctx, cancel := context.WithCancel(context.Background())

			done := make(chan error, 1)
			go func() {
				done <- oc.Run(ctx)
			}()

			// 等待初始预热完成
			time.Sleep(100 * time.Millisecond)
			cancel()

			select {
			case err := <-done:
				So(err, ShouldEqual, context.Canceled)
				So(callCount, ShouldBeGreaterThanOrEqualTo, 1)
			case <-time.After(2 * time.Second):
				t.Fatal("Run 没有在预期时间内退出")
			}
		})
	})
}

func TestEntityObjectInfo(t *testing.T) {
	Convey("TestEntityObjectInfo", t, func() {
		Convey("JSON 序列化和反序列化", func() {
			objInfo := EntityObjectInfo{
				ObjectTypeID: "ot-1",
				ObjectID:     "obj-1",
				Name:         "test-pod",
			}

			// 序列化
			jsonData, err := json.Marshal(objInfo)
			So(err, ShouldBeNil)
			So(string(jsonData), ShouldContainSubstring, `"object_type_id":"ot-1"`)
			So(string(jsonData), ShouldContainSubstring, `"object_id":"obj-1"`)
			So(string(jsonData), ShouldContainSubstring, `"name":"test-pod"`)

			// 反序列化
			var decoded EntityObjectInfo
			err = json.Unmarshal(jsonData, &decoded)
			So(err, ShouldBeNil)
			So(decoded.ObjectTypeID, ShouldEqual, "ot-1")
			So(decoded.ObjectID, ShouldEqual, "obj-1")
			So(decoded.Name, ShouldEqual, "test-pod")
		})
	})
}
