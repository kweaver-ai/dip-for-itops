package config

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	mu            sync.RWMutex
	config        *Config // 当前生效的配置（基础配置 + 业务配置合并后）
	configPath    string  // config.yaml 文件路径（只读）
	appConfigPath string  // app_config.yaml 文件路径（可写）

	// 远程配置
	httpClient    *http.Client
	lastRemoteApp *RemoteAppConfig // 上次远程获取的配置（用于变更检测）

	// watch 相关
	watcher *fsnotify.Watcher
	stopCh  chan struct{}
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string) (*ConfigManager, error) {
	// 初始加载基础配置
	cfg, err := Load(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "初始加载配置失败")
	}

	// 计算 app_config.yaml 路径（与 config.yaml 同目录下的 data 子目录）
	configDir := filepath.Dir(configPath)
	appConfigPath := filepath.Join(configDir, "data", "app_config.yaml")

	// 尝试加载业务配置
	appCfg, err := LoadAppConfig(appConfigPath)
	if err != nil {
		// 如果文件不存在，使用默认配置
		log.Warnf("加载业务配置失败: %v，使用默认配置", err)
		appCfg = defaultAppConfig()

		// 确保 data 目录存在
		dataDir := filepath.Dir(appConfigPath)
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Warnf("创建 data 目录失败: %v", err)
		} else {
			// 写入默认配置
			if err := SaveAppConfig(appConfigPath, appCfg); err != nil {
				log.Warnf("写入默认业务配置失败: %v", err)
			}
		}
	}
	cfg.AppConfig = *appCfg

	// 创建文件 watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "创建文件 watcher 失败")
	}

	// 添加配置文件到 watch 列表
	if err := watcher.Add(configPath); err != nil {
		watcher.Close()
		return nil, errors.Wrap(err, "添加配置文件到 watch 列表失败")
	}

	// 尝试添加 app_config.yaml 到 watch 列表（可能不存在）
	if _, err := os.Stat(appConfigPath); err == nil {
		if err := watcher.Add(appConfigPath); err != nil {
			log.Warnf("添加业务配置文件到 watch 列表失败: %v", err)
		}
	}

	return &ConfigManager{
		config:        cfg,
		configPath:    configPath,
		appConfigPath: appConfigPath,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		watcher: watcher,
		stopCh:  make(chan struct{}),
	}, nil
}

// defaultAppConfig 返回默认业务配置
func defaultAppConfig() *AppConfig {
	return &AppConfig{
		Credentials: CredentialsConfig{
			Authorization: "",
		},
		KnowledgeNetwork: KnowledgeNetworkConfig{
			KnowledgeID: "",
		},
		Ingest: IngestConfig{
			Source: Source{Type: "zabbix_webhook"},
		},
		FaultPoint: FaultPointExpirationCfg{
			Expiration: LocalExpirationConfig{
				Enabled:        true,
				ExpirationTime: 1 * time.Hour,
			},
		},
		Problem: ProblemExpirationCfg{
			Expiration: LocalExpirationConfig{
				Enabled:        true,
				ExpirationTime: 1 * time.Hour,
			},
		},
	}
}

// Start 启动配置管理（文件 watch + 定时刷新）
func (m *ConfigManager) Start(ctx context.Context) error {
	// 启动时立即尝试拉取远程配置
	if m.config.AppConfigService.Enabled {
		if err := m.fetchAndWriteRemoteConfig(); err != nil {
			log.Warnf("启动时拉取远程配置失败: %v，使用本地默认配置", err)
		}
	}

	// 启动文件 watch 协程
	go m.watchConfigFile(ctx)

	// 启动远程配置定时刷新协程
	if m.config.AppConfigService.Enabled {
		go m.runRemoteConfigRefresher(ctx)
	}

	<-ctx.Done()
	return ctx.Err()
}

// Stop 停止配置管理
func (m *ConfigManager) Stop() {
	close(m.stopCh)
	if m.watcher != nil {
		m.watcher.Close()
	}
}

// GetConfig 获取当前配置（线程安全）
func (m *ConfigManager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetAppConfigPath 获取业务配置文件路径
func (m *ConfigManager) GetAppConfigPath() string {
	return m.appConfigPath
}

// watchConfigFile 监控配置文件变动
func (m *ConfigManager) watchConfigFile(ctx context.Context) {
	log.Info("启动配置文件 watch 协程")

	for {
		select {
		case <-ctx.Done():
			log.Info("配置文件 watch 协程收到停止信号")
			return
		case <-m.stopCh:
			log.Info("配置文件 watch 协程收到停止信号")
			return
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			// 只处理写入和创建事件
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				log.Infof("检测到配置文件变动: %s", event.Name)
				// 延迟一下，确保文件写入完成
				time.Sleep(100 * time.Millisecond)
				if err := m.reload(); err != nil {
					log.Errorf("重新加载配置失败: %v", err)
				} else {
					log.Info("配置重新加载成功")
				}
			}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Errorf("配置文件 watch 错误: %v", err)
		}
	}
}

// runRemoteConfigRefresher 定时刷新远程配置
func (m *ConfigManager) runRemoteConfigRefresher(ctx context.Context) {
	interval := m.config.AppConfigService.RefreshInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}

	log.Infof("启动远程配置定时刷新协程（间隔: %v）", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("远程配置定时刷新协程收到停止信号")
			return
		case <-m.stopCh:
			log.Info("远程配置定时刷新协程收到停止信号")
			return
		case <-ticker.C:
			log.Debug("开始定时拉取远程配置...")
			if err := m.fetchAndWriteRemoteConfig(); err != nil {
				log.Errorf("定时拉取远程配置失败: %v", err)
			}
		}
	}
}

// fetchAndWriteRemoteConfig 获取远程配置并写入 app_config.yaml
func (m *ConfigManager) fetchAndWriteRemoteConfig() error {
	// 获取远程配置
	remoteApp, err := m.fetchRemoteAppConfig()
	if err != nil {
		return errors.Wrap(err, "获取远程配置失败")
	}

	// 检查是否有变更
	if m.lastRemoteApp != nil && reflect.DeepEqual(m.lastRemoteApp, remoteApp) {
		log.Debug("远程配置无变化，跳过写入")
		return nil
	}

	// 转换为本地配置格式并写入文件
	appConfig := remoteApp.ToAppConfig()
	if err := m.writeAppConfig(appConfig); err != nil {
		return errors.Wrap(err, "写入业务配置失败")
	}

	m.lastRemoteApp = remoteApp
	log.Info("远程配置已更新并写入 app_config.yaml")

	// 直接重新加载配置，不依赖 watcher（避免 watcher 还未启动或事件延迟）
	if err := m.reload(); err != nil {
		log.Warnf("写入后重新加载配置失败: %v", err)
	}

	return nil
}

// fetchRemoteAppConfig 获取远程业务配置（alert-manager API 格式）
func (m *ConfigManager) fetchRemoteAppConfig() (*RemoteAppConfig, error) {
	endpoint := m.config.AppConfigService.Endpoint
	if endpoint == "" {
		return nil, errors.New("远程配置接口地址为空")
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "创建请求失败")
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "请求远程配置失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("远程配置接口返回非 200 状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "读取响应体失败")
	}

	var remoteApp RemoteAppConfig
	if err := json.Unmarshal(body, &remoteApp); err != nil {
		return nil, errors.Wrap(err, "解析远程配置 JSON 失败")
	}

	return &remoteApp, nil
}

// writeAppConfig 写入业务配置到 app_config.yaml
func (m *ConfigManager) writeAppConfig(appConfig *AppConfig) error {
	// 确保目录存在
	dataDir := filepath.Dir(m.appConfigPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return errors.Wrap(err, "创建 data 目录失败")
	}

	// 写入文件
	if err := SaveAppConfig(m.appConfigPath, appConfig); err != nil {
		return err
	}

	// 尝试添加到 watcher（如果是新创建的文件）
	if m.watcher != nil {
		_ = m.watcher.Add(m.appConfigPath)
	}

	return nil
}

// reload 重新加载配置
func (m *ConfigManager) reload() error {
	// 加载基础配置
	cfg, err := Load(m.configPath)
	if err != nil {
		return errors.Wrap(err, "加载基础配置失败")
	}

	// 加载业务配置
	appCfg, err := LoadAppConfig(m.appConfigPath)
	if err != nil {
		log.Warnf("加载业务配置失败: %v，保持原有业务配置", err)
		m.mu.RLock()
		appCfg = &m.config.AppConfig
		m.mu.RUnlock()
	}
	cfg.AppConfig = *appCfg

	m.mu.Lock()
	m.config = cfg
	m.mu.Unlock()

	return nil
}
