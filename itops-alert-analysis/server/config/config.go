package config

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config 映射 config.yaml 的 YAML 结构。
// 所有时间窗口与策略都由配置提供，代码中不硬编码业务阈值。
type Config struct {
	API              APIConfig              `yaml:"api"`
	Log              LogConfig              `yaml:"log"`                // 日志配置
	Kafka            KafkaConfig            `yaml:"kafka"`              // Kafka 配置
	Platform         PlatformConfig         `yaml:"platform"`           // 统一的平台配置（知识网络 + AI 能力）
	DepServices      DepServicesConfig      `yaml:"depServices"`        // 依赖服务配置
	AppConfigService AppConfigServiceConfig `yaml:"app_config_service"` // 远程配置服务
	AppConfig        AppConfig              `yaml:"app_config"`         // 业务配置（本地默认值 + 远程接口合并）
}

// ========== API 配置 ==========

// APIConfig API 服务配置
type APIConfig struct {
	Port int `yaml:"port"`
}

// ========== 日志配置 ==========

// LogConfig 日志配置
type LogConfig struct {
	Filepath    string `yaml:"filepath"`    // 日志文件路径
	Level       string `yaml:"level"`       // 日志级别 info warning error
	MaxSize     int    `yaml:"max_size"`    // 每个日志文件最大空间(单位：MB)
	MaxAge      int    `yaml:"max_age"`     // 文件最多保留多少天
	MaxBackups  int    `yaml:"max_backups"` // 文件最多保留多少备份
	Compress    bool   `yaml:"compress"`    // 是否压缩
	Development bool   `yaml:"development"` // 是否开启开发模式
}

// ========== Kafka 配置 ==========

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	RawEvents     KafkaStreamConfig `yaml:"raw_events"`     // 原始事件流（HTTP -> Correlation）
	ProblemEvents KafkaStreamConfig `yaml:"problem_events"` // 问题事件流（Correlation -> RCA）
}

// KafkaStreamConfig Kafka 流配置
type KafkaStreamConfig struct {
	Topic         string `yaml:"topic"`
	ConsumerGroup string `yaml:"consumer_group"`
}

// ========== 平台配置 ==========

// PlatformConfig 统一的平台配置（技术配置）
type PlatformConfig struct {
	BaseURL            string        `yaml:"base_url"`             // 平台服务基础地址
	Timeout            time.Duration `yaml:"timeout"`              // 默认请求超时时间
	InsecureSkipVerify bool          `yaml:"insecure_skip_verify"` // 是否跳过SSL验证
	Agents             AgentsConfig  `yaml:"agents"`               // AI Agent 服务配置
}

// AgentsConfig AI Agent 服务配置
type AgentsConfig struct {
	ProblemSummary AgentConfig `yaml:"problem_summary"` // 问题摘要 Agent
	CausalAnalysis AgentConfig `yaml:"causal_analysis"` // 因果分析 Agent
}

// AgentConfig 单个 Agent 配置
type AgentConfig struct {
	Enabled  bool   `yaml:"enabled"`   // 是否启用
	AppID    string `yaml:"app_id"`    // Agent 应用 ID
	AgentKey string `yaml:"agent_key"` // Agent 密钥
}

// DIPConfig 知识网络配置（向后兼容，从 Platform 派生）
type DIPConfig struct {
	Host               string
	KnID               string
	Authorization      string
	InsecureSkipVerify bool
	Timeout            time.Duration
}

// ========== 依赖服务配置 ==========

// DepServicesConfig 依赖服务配置
type DepServicesConfig struct {
	Class443   IngressClassConfig  `yaml:"class-443"`  // Ingress 类配置
	MQ         MQConfig            `yaml:"mq"`         // 消息队列配置
	OpenSearch DepOpenSearchConfig `yaml:"opensearch"` // OpenSearch 配置
	Redis      DepRedisConfig      `yaml:"redis"`      // Redis 配置
}

// IngressClassConfig Ingress 类配置
type IngressClassConfig struct {
	IngressClass string `yaml:"ingressClass"` // Ingress 类名
}

// MQConfig 消息队列配置
type MQConfig struct {
	Auth     MQAuthConfig `yaml:"auth"`     // 认证配置
	MQHost   string       `yaml:"mqHost"`   // 消息队列主机地址
	MQPort   int          `yaml:"mqPort"`   // 消息队列端口
	MQType   string       `yaml:"mqType"`   // 消息队列类型（如 kafka）
	Protocol string       `yaml:"protocol"` // 协议（如 sasl_plaintext）
	Tenant   string       `yaml:"tenant"`   // 租户
}

// MQAuthConfig 消息队列认证配置
type MQAuthConfig struct {
	Mechanism string `yaml:"mechanism"` // 认证机制（如 PLAIN）
	Password  string `yaml:"password"`  // 密码
	Username  string `yaml:"username"`  // 用户名
}

// DepOpenSearchConfig 依赖的 OpenSearch 配置
type DepOpenSearchConfig struct {
	Host     string `yaml:"host"`     // OpenSearch 主机地址
	Port     int    `yaml:"port"`     // OpenSearch 端口
	Protocol string `yaml:"protocol"` // 协议（http/https）
	User     string `yaml:"user"`     // 用户名
	Password string `yaml:"password"` // 密码
}

// DepRedisConfig 依赖的 Redis 配置
type DepRedisConfig struct {
	ConnectInfo RedisConnectInfo `yaml:"connectInfo"` // 连接信息
	ConnectType string           `yaml:"connectType"` // 连接类型（如 sentinel）
}

// RedisConnectInfo Redis 连接信息
type RedisConnectInfo struct {
	MasterGroupName  string `yaml:"masterGroupName"`  // Master 组名（Sentinel 模式）
	Password         string `yaml:"password"`         // Redis 密码
	SentinelHost     string `yaml:"sentinelHost"`     // Sentinel 主机地址
	SentinelPassword string `yaml:"sentinelPassword"` // Sentinel 密码
	SentinelPort     int    `yaml:"sentinelPort"`     // Sentinel 端口
	SentinelUsername string `yaml:"sentinelUsername"` // Sentinel 用户名
	Username         string `yaml:"username"`         // Redis 用户名
}

// Load 从指定路径读取 YAML 配置。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read config")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Wrap(err, "unmarshal config")
	}
	return &cfg, nil
}

// LoadAppConfig 从指定路径读取业务配置
func LoadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read app config")
	}

	var appCfg AppConfig
	if err := yaml.Unmarshal(data, &appCfg); err != nil {
		return nil, errors.Wrap(err, "unmarshal app config")
	}
	return &appCfg, nil
}

// SaveAppConfig 将业务配置写入指定路径
func SaveAppConfig(path string, appCfg *AppConfig) error {
	data, err := yaml.Marshal(appCfg)
	if err != nil {
		return errors.Wrap(err, "marshal app config")
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrap(err, "write app config")
	}
	return nil
}
