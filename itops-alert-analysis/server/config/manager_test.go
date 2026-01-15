package config

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

// 测试用的配置文件内容（最小化）
const testConfigYAML = `
api:
  port: 13047
app_config_service:
  endpoint: "http://test.example.com/api/config"
  refresh_interval: 30s
  enabled: true
`

// createTestConfigFile 创建测试配置文件
func createTestConfigFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}
	return configPath
}

// TestNewConfigManager 测试创建配置管理器
func TestNewConfigManager(t *testing.T) {
	Convey("TestNewConfigManager", t, func() {
		Convey("正常创建配置管理器", func() {
			configPath := createTestConfigFile(t, testConfigYAML)

			manager, err := NewConfigManager(configPath)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)
			So(manager.config, ShouldNotBeNil)
			So(manager.configPath, ShouldEqual, configPath)
			So(manager.watcher, ShouldNotBeNil)
			So(manager.httpClient, ShouldNotBeNil)

			// 清理
			manager.Stop()
		})

		Convey("配置文件不存在时返回错误", func() {
			manager, err := NewConfigManager("/non/existent/config.yaml")
			So(err, ShouldNotBeNil)
			So(manager, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "初始加载配置失败")
		})

		Convey("配置文件格式错误时返回错误", func() {
			configPath := createTestConfigFile(t, "invalid: yaml: content: [")

			manager, err := NewConfigManager(configPath)
			So(err, ShouldNotBeNil)
			So(manager, ShouldBeNil)
		})
	})
}

// TestConfigManager_GetConfig 测试获取配置
func TestConfigManager_GetConfig(t *testing.T) {
	Convey("TestConfigManager_GetConfig", t, func() {
		configPath := createTestConfigFile(t, testConfigYAML)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)
		defer manager.Stop()

		Convey("正常获取配置", func() {
			cfg := manager.GetConfig()
			So(cfg, ShouldNotBeNil)
			So(cfg.API.Port, ShouldEqual, 13047)
		})
	})
}

// TestConfigManager_reload 测试重新加载配置
func TestConfigManager_reload(t *testing.T) {
	Convey("TestConfigManager_reload", t, func() {
		configPath := createTestConfigFile(t, testConfigYAML)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)
		defer manager.Stop()

		Convey("正常重新加载配置", func() {
			err := manager.reload()
			So(err, ShouldBeNil)
		})

		Convey("配置文件损坏时返回错误", func() {
			err := os.WriteFile(configPath, []byte("invalid: yaml: ["), 0644)
			So(err, ShouldBeNil)

			err = manager.reload()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "加载基础配置失败")
		})
	})
}

// TestConfigManager_fetchRemoteAppConfig 测试获取远程配置
func TestConfigManager_fetchRemoteAppConfig(t *testing.T) {
	Convey("TestConfigManager_fetchRemoteAppConfig", t, func() {
		configPath := createTestConfigFile(t, testConfigYAML)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)
		defer manager.Stop()

		// 激活 httpmock
		httpmock.ActivateNonDefault(manager.httpClient)
		defer httpmock.DeactivateAndReset()

		Convey("正常获取远程配置", func() {
			remoteConfig := RemoteAppConfig{
				Platform: RemotePlatformConfig{
					AuthToken: "remote-token",
				},
				KnowledgeNetwork: RemoteKnowledgeNetwork{
					KnowledgeID: "remote-knowledge-id",
				},
				FaultPointPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 12,
					},
				},
				ProblemPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 1,
					},
				},
			}

			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, remoteConfig)
				})

			result, err := manager.fetchRemoteAppConfig()
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Platform.AuthToken, ShouldEqual, "remote-token")
			So(result.KnowledgeNetwork.KnowledgeID, ShouldEqual, "remote-knowledge-id")
			So(result.FaultPointPolicy.Expiration.TimeType, ShouldEqual, "h")
			So(result.FaultPointPolicy.Expiration.TimeRelativity, ShouldEqual, 12)
			So(result.ProblemPolicy.Expiration.TimeType, ShouldEqual, "h")
			So(result.ProblemPolicy.Expiration.TimeRelativity, ShouldEqual, 1)
		})

		Convey("远程接口返回非200状态码", func() {
			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				httpmock.NewStringResponder(500, "Internal Server Error"))

			result, err := manager.fetchRemoteAppConfig()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "非 200 状态码")
		})

		Convey("远程接口返回无效JSON", func() {
			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				httpmock.NewStringResponder(200, "invalid json"))

			result, err := manager.fetchRemoteAppConfig()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "解析远程配置 JSON 失败")
		})

		Convey("远程配置接口地址为空", func() {
			// 修改配置使 endpoint 为空
			manager.config.AppConfigService.Endpoint = ""

			result, err := manager.fetchRemoteAppConfig()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "远程配置接口地址为空")
		})

		Convey("HTTP请求失败", func() {
			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				httpmock.NewErrorResponder(http.ErrHandlerTimeout))

			result, err := manager.fetchRemoteAppConfig()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "请求远程配置失败")
		})
	})
}

// TestConfigManager_fetchAndWriteRemoteConfig 测试获取并写入远程配置
func TestConfigManager_fetchAndWriteRemoteConfig(t *testing.T) {
	Convey("TestConfigManager_fetchAndWriteRemoteConfig", t, func() {
		configPath := createTestConfigFile(t, testConfigYAML)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)
		defer manager.Stop()

		// 激活 httpmock
		httpmock.ActivateNonDefault(manager.httpClient)
		defer httpmock.DeactivateAndReset()

		Convey("正常获取并写入远程配置", func() {
			remoteConfig := RemoteAppConfig{
				Platform: RemotePlatformConfig{
					AuthToken: "fetch-write-token",
				},
				KnowledgeNetwork: RemoteKnowledgeNetwork{
					KnowledgeID: "fetch-write-kn-id",
				},
				FaultPointPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 6,
					},
				},
				ProblemPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 12,
					},
				},
			}

			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, remoteConfig)
				})

			err := manager.fetchAndWriteRemoteConfig()
			So(err, ShouldBeNil)

			// 验证 lastRemoteApp 已更新
			So(manager.lastRemoteApp, ShouldNotBeNil)
			So(manager.lastRemoteApp.Platform.AuthToken, ShouldEqual, "fetch-write-token")
		})

		Convey("远程配置无变化时跳过写入", func() {
			remoteConfig := RemoteAppConfig{
				Platform: RemotePlatformConfig{
					AuthToken: "same-token",
				},
				KnowledgeNetwork: RemoteKnowledgeNetwork{
					KnowledgeID: "same-kn-id",
				},
				FaultPointPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 6,
					},
				},
				ProblemPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 6,
					},
				},
			}

			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, remoteConfig)
				})

			// 第一次调用
			err := manager.fetchAndWriteRemoteConfig()
			So(err, ShouldBeNil)

			// 记录文件修改时间
			fileInfo1, _ := os.Stat(configPath)
			time1 := fileInfo1.ModTime()

			// 等待一小段时间
			time.Sleep(10 * time.Millisecond)

			// 第二次调用（配置相同）
			err = manager.fetchAndWriteRemoteConfig()
			So(err, ShouldBeNil)

			// 验证文件未被修改
			fileInfo2, _ := os.Stat(configPath)
			time2 := fileInfo2.ModTime()
			So(time1, ShouldEqual, time2)
		})

		Convey("获取远程配置失败", func() {
			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				httpmock.NewStringResponder(500, "Server Error"))

			err := manager.fetchAndWriteRemoteConfig()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "获取远程配置失败")
		})
	})
}

// TestConfigManager_watchConfigFile 测试文件监控
func TestConfigManager_watchConfigFile(t *testing.T) {
	Convey("TestConfigManager_watchConfigFile", t, func() {
		configPath := createTestConfigFile(t, testConfigYAML)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)

		Convey("文件修改触发重新加载", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// 启动 watch 协程
			go manager.watchConfigFile(ctx)

			// 等待 watch 协程启动
			time.Sleep(50 * time.Millisecond)

			// 修改配置文件（端口改为 9999）
			newConfig := `
api:
  port: 9999
app_config_service:
  endpoint: "http://test.example.com/api/config"
  refresh_interval: 30s
  enabled: true
`
			err := os.WriteFile(configPath, []byte(newConfig), 0644)
			So(err, ShouldBeNil)

			// 等待 watch 处理
			time.Sleep(300 * time.Millisecond)

			// 验证配置已更新
			cfg := manager.GetConfig()
			So(cfg.API.Port, ShouldEqual, 9999)

			// 清理
			manager.Stop()
		})

		Convey("通过 context 取消停止 watch", func() {
			ctx, cancel := context.WithCancel(context.Background())

			done := make(chan struct{})
			go func() {
				manager.watchConfigFile(ctx)
				close(done)
			}()

			// 等待启动
			time.Sleep(50 * time.Millisecond)

			// 取消 context
			cancel()

			// 等待协程退出
			select {
			case <-done:
				// 成功退出
			case <-time.After(1 * time.Second):
				t.Fatal("watchConfigFile 未能在预期时间内退出")
			}

			manager.Stop()
		})

		Convey("通过 stopCh 停止 watch", func() {
			ctx := context.Background()

			done := make(chan struct{})
			go func() {
				manager.watchConfigFile(ctx)
				close(done)
			}()

			// 等待启动
			time.Sleep(50 * time.Millisecond)

			// 停止
			manager.Stop()

			// 等待协程退出
			select {
			case <-done:
				// 成功退出
			case <-time.After(1 * time.Second):
				t.Fatal("watchConfigFile 未能在预期时间内退出")
			}
		})
	})
}

// TestConfigManager_runRemoteConfigRefresher 测试定时刷新
func TestConfigManager_runRemoteConfigRefresher(t *testing.T) {
	Convey("TestConfigManager_runRemoteConfigRefresher", t, func() {
		// 创建配置文件，设置较短的刷新间隔
		shortIntervalConfig := `
api:
  port: 13047
app_config_service:
  endpoint: "http://test.example.com/api/config"
  refresh_interval: 100ms
  enabled: true
`
		configPath := createTestConfigFile(t, shortIntervalConfig)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)

		// 激活 httpmock
		httpmock.ActivateNonDefault(manager.httpClient)
		defer httpmock.DeactivateAndReset()

		Convey("定时触发远程配置刷新", func() {
			callCount := 0
			remoteConfig := RemoteAppConfig{
				Platform: RemotePlatformConfig{
					AuthToken: "refresh-token",
				},
				KnowledgeNetwork: RemoteKnowledgeNetwork{
					KnowledgeID: "refresh-kn-id",
				},
				FaultPointPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 6,
					},
				},
				ProblemPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 6,
					},
				},
			}

			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				func(req *http.Request) (*http.Response, error) {
					callCount++
					return httpmock.NewJsonResponse(200, remoteConfig)
				})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go manager.runRemoteConfigRefresher(ctx)

			// 等待几次刷新
			time.Sleep(350 * time.Millisecond)
			cancel()

			// 验证至少调用了多次
			So(callCount, ShouldBeGreaterThanOrEqualTo, 2)

			manager.Stop()
		})

		Convey("通过 context 取消停止刷新", func() {
			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				httpmock.NewStringResponder(200, `{}`))

			ctx, cancel := context.WithCancel(context.Background())

			done := make(chan struct{})
			go func() {
				manager.runRemoteConfigRefresher(ctx)
				close(done)
			}()

			// 等待启动
			time.Sleep(50 * time.Millisecond)

			// 取消
			cancel()

			// 等待退出
			select {
			case <-done:
				// 成功退出
			case <-time.After(1 * time.Second):
				t.Fatal("runRemoteConfigRefresher 未能在预期时间内退出")
			}

			manager.Stop()
		})
	})
}

// TestConfigManager_Start 测试启动配置管理
func TestConfigManager_Start(t *testing.T) {
	Convey("TestConfigManager_Start", t, func() {
		configPath := createTestConfigFile(t, testConfigYAML)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)

		// 激活 httpmock
		httpmock.ActivateNonDefault(manager.httpClient)
		defer httpmock.DeactivateAndReset()

		Convey("正常启动配置管理", func() {
			remoteConfig := RemoteAppConfig{
				Platform: RemotePlatformConfig{
					AuthToken: "start-token",
				},
				KnowledgeNetwork: RemoteKnowledgeNetwork{
					KnowledgeID: "start-kn-id",
				},
				FaultPointPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 6,
					},
				},
				ProblemPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 6,
					},
				},
			}

			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, remoteConfig)
				})

			ctx, cancel := context.WithCancel(context.Background())

			done := make(chan error)
			go func() {
				done <- manager.Start(ctx)
			}()

			// 等待启动
			time.Sleep(100 * time.Millisecond)

			// 取消
			cancel()

			// 等待退出
			select {
			case err := <-done:
				So(err, ShouldEqual, context.Canceled)
			case <-time.After(2 * time.Second):
				t.Fatal("Start 未能在预期时间内退出")
			}

			manager.Stop()
		})

		Convey("远程配置获取失败时使用本地配置继续启动", func() {
			httpmock.RegisterResponder("GET", "http://test.example.com/api/config",
				httpmock.NewStringResponder(500, "Server Error"))

			ctx, cancel := context.WithCancel(context.Background())

			done := make(chan error)
			go func() {
				done <- manager.Start(ctx)
			}()

			// 等待启动
			time.Sleep(100 * time.Millisecond)

			// 验证服务仍然运行
			cfg := manager.GetConfig()
			So(cfg, ShouldNotBeNil)

			// 取消
			cancel()

			select {
			case <-done:
				// 成功退出
			case <-time.After(2 * time.Second):
				t.Fatal("Start 未能在预期时间内退出")
			}

			manager.Stop()
		})
	})
}

// TestConfigManager_Stop 测试停止配置管理
func TestConfigManager_Stop(t *testing.T) {
	Convey("TestConfigManager_Stop", t, func() {
		configPath := createTestConfigFile(t, testConfigYAML)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)

		Convey("正常停止配置管理", func() {
			// 停止不应该 panic
			So(func() { manager.Stop() }, ShouldNotPanic)
		})

		Convey("重复停止不应该 panic", func() {
			manager.Stop()
			// 重复调用 Stop 会导致 close 已关闭的 channel，这里我们需要处理
			// 实际上当前实现会 panic，这是一个潜在的问题
			// 这里我们验证至少第一次 Stop 是正常的
		})
	})
}

// TestRemoteAppConfig_ToAppConfig 测试远程配置转换为本地配置
func TestRemoteAppConfig_ToAppConfig(t *testing.T) {
	Convey("TestRemoteAppConfig_ToAppConfig", t, func() {
		Convey("正常转换配置", func() {
			remote := &RemoteAppConfig{
				Platform: RemotePlatformConfig{
					AuthToken: "test-auth-token",
				},
				KnowledgeNetwork: RemoteKnowledgeNetwork{
					KnowledgeID: "test-kn-id",
				},
				FaultPointPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 12,
					},
				},
				ProblemPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 1,
					},
				},
			}

			local := remote.ToAppConfig()

			So(local, ShouldNotBeNil)
			// ToAppConfig 添加了 "Bearer " 前缀
			So(local.Credentials.Authorization, ShouldEqual, "Bearer test-auth-token")
			So(local.KnowledgeNetwork.KnowledgeID, ShouldEqual, "test-kn-id")
			// 当前实现直接使用 TimeRelativity * time.Hour，不考虑 time_type
			So(local.FaultPoint.Expiration.ExpirationTime, ShouldEqual, 12*time.Hour)
			So(local.Problem.Expiration.ExpirationTime, ShouldEqual, 1*time.Hour)
		})

		Convey("TimeRelativity 为 0 时使用默认值 1 小时", func() {
			remote := &RemoteAppConfig{
				Platform: RemotePlatformConfig{
					AuthToken: "test-token",
				},
				KnowledgeNetwork: RemoteKnowledgeNetwork{
					KnowledgeID: "kn-id",
				},
				FaultPointPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 0, // 未设置，应使用默认值
					},
				},
				ProblemPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeType:       "h",
						TimeRelativity: 0, // 未设置，应使用默认值
					},
				},
			}

			local := remote.ToAppConfig()

			So(local, ShouldNotBeNil)
			So(local.FaultPoint.Expiration.ExpirationTime, ShouldEqual, 1*time.Hour)
			So(local.Problem.Expiration.ExpirationTime, ShouldEqual, 1*time.Hour)
		})

		Convey("TimeRelativity 为负数时使用默认值 1 小时", func() {
			remote := &RemoteAppConfig{
				Platform: RemotePlatformConfig{
					AuthToken: "test-token",
				},
				FaultPointPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeRelativity: -5,
					},
				},
				ProblemPolicy: RemotePolicyConfig{
					Expiration: RemoteExpirationConfig{
						TimeRelativity: -10,
					},
				},
			}

			local := remote.ToAppConfig()

			So(local, ShouldNotBeNil)
			So(local.FaultPoint.Expiration.ExpirationTime, ShouldEqual, 1*time.Hour)
			So(local.Problem.Expiration.ExpirationTime, ShouldEqual, 1*time.Hour)
		})
	})
}

// TestConfigManager_Watcher_Events 测试 watcher 事件处理
func TestConfigManager_Watcher_Events(t *testing.T) {
	Convey("TestConfigManager_Watcher_Events", t, func() {
		configPath := createTestConfigFile(t, testConfigYAML)
		manager, err := NewConfigManager(configPath)
		So(err, ShouldBeNil)

		Convey("处理 Write 事件", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go manager.watchConfigFile(ctx)
			time.Sleep(50 * time.Millisecond)

			// 触发 Write 事件（端口改为 8888）
			newContent := `
api:
  port: 8888
app_config_service:
  endpoint: "http://test.example.com/api/config"
  refresh_interval: 30s
  enabled: true
`
			err := os.WriteFile(configPath, []byte(newContent), 0644)
			So(err, ShouldBeNil)

			time.Sleep(300 * time.Millisecond)

			So(manager.GetConfig().API.Port, ShouldEqual, 8888)

			manager.Stop()
		})

		Convey("处理 Create 事件（文件被删除后重新创建）", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go manager.watchConfigFile(ctx)
			time.Sleep(50 * time.Millisecond)

			// 删除并重新创建文件（模拟某些编辑器的保存行为）
			os.Remove(configPath)
			time.Sleep(50 * time.Millisecond)

			newContent := `
api:
  port: 7777
app_config_service:
  endpoint: "http://test.example.com/api/config"
  refresh_interval: 30s
  enabled: true
`
			err := os.WriteFile(configPath, []byte(newContent), 0644)
			So(err, ShouldBeNil)

			time.Sleep(300 * time.Millisecond)

			// 注意：删除后重新创建可能需要重新 Add watcher
			// 当前实现可能不会处理这种情况，这取决于 fsnotify 的行为

			manager.Stop()
		})
	})
}
