package config

import (
	"fmt"
	"time"
)

// ========== 远程配置服务 ==========

// AppConfigServiceConfig 远程配置服务配置
type AppConfigServiceConfig struct {
	Endpoint        string        `yaml:"endpoint"`         // 远程配置接口地址
	RefreshInterval time.Duration `yaml:"refresh_interval"` // 刷新间隔
	Enabled         bool          `yaml:"enabled"`          // 是否启用远程配置
}

// ========== 本地业务配置 ==========

// AppConfig 本地业务配置（config.yaml 格式）
type AppConfig struct {
	Credentials      CredentialsConfig       `yaml:"credentials" json:"credentials"`
	KnowledgeNetwork KnowledgeNetworkConfig  `yaml:"knowledge_network" json:"knowledge_network"`
	Ingest           IngestConfig            `yaml:"ingest"`
	FaultPoint       FaultPointExpirationCfg `yaml:"fault_point" json:"fault_point"`
	Problem          ProblemExpirationCfg    `yaml:"problem" json:"problem"`
}

// CredentialsConfig 认证凭据配置
type CredentialsConfig struct {
	Authorization string `yaml:"authorization" json:"authorization"` // Bearer Token
}

// KnowledgeNetworkConfig 知识网络服务配置
type KnowledgeNetworkConfig struct {
	KnowledgeID string `yaml:"knowledge_id" json:"knowledge_id"` // 知识网络 ID
}

// IngestConfig 数据摄取配置
type IngestConfig struct {
	Source Source `yaml:"source"`
}

// Source 数据源配置
type Source struct {
	Type string `yaml:"type"`
}

// ========== 失效策略配置 ==========

// FaultPointExpirationCfg 故障点失效配置（AppConfig 使用）
type FaultPointExpirationCfg struct {
	Expiration LocalExpirationConfig `yaml:"expiration" json:"expiration"`
}

// ProblemExpirationCfg 问题失效配置（AppConfig 使用）
type ProblemExpirationCfg struct {
	Expiration LocalExpirationConfig `yaml:"expiration" json:"expiration"`
}

// LocalExpirationConfig 本地失效配置（使用 Duration 格式如 "6h"）
type LocalExpirationConfig struct {
	Enabled        bool          `yaml:"enabled" json:"enabled"`
	ExpirationTime time.Duration `yaml:"expiration_time" json:"expiration_time"`
}

// FaultPointConfig 故障点配置
type FaultPointConfig struct {
	Expiration FaultPointExpirationConfig `yaml:"expiration"` // 故障点失效配置
}

// FaultPointExpirationConfig 故障点失效配置
type FaultPointExpirationConfig struct {
	Enabled        bool          `yaml:"enabled"`         // 是否启用失效检查
	ExpirationTime time.Duration `yaml:"expiration_time"` // 故障点失效时间（超过此时间未更新的故障点将被标记为失效）
}

// ProblemConfig 问题配置
type ProblemConfig struct {
	Expiration ProblemExpirationConfig `yaml:"expiration"` // 问题失效配置
}

// ProblemExpirationConfig 问题失效配置
type ProblemExpirationConfig struct {
	Enabled        bool          `yaml:"enabled"`         // 是否启用失效检查
	ExpirationTime time.Duration `yaml:"expiration_time"` // 问题失效时间（超过此时间未更新的问题将被关闭）
}

// ========== 远程 API 响应结构（alert-manager 返回格式）==========

// RemoteAppConfig 远程业务配置（alert-manager API 返回格式）
type RemoteAppConfig struct {
	Platform         RemotePlatformConfig   `json:"platform"`
	KnowledgeNetwork RemoteKnowledgeNetwork `json:"knowledge_network"`
	FaultPointPolicy RemotePolicyConfig     `json:"fault_point_policy"`
	ProblemPolicy    RemotePolicyConfig     `json:"problem_policy"`
}

// RemotePlatformConfig 远程平台配置
type RemotePlatformConfig struct {
	AuthToken string `json:"auth_token"` // 认证令牌
}

// RemoteKnowledgeNetwork 远程知识网络配置
type RemoteKnowledgeNetwork struct {
	KnowledgeID string `json:"knowledge_id"` // 知识网络 ID
}

// RemotePolicyConfig 远程策略配置
type RemotePolicyConfig struct {
	Expiration RemoteExpirationConfig `json:"expiration"`
}

// RemoteExpirationConfig 远程失效配置（time_type + time_relativity）
type RemoteExpirationConfig struct {
	TimeType       string `json:"time_type"`       // 时间类型：d-天, h-小时, m-分钟
	TimeRelativity int    `json:"time_relativity"` // 时间值
}

// defaultTimeRelativity 默认失效时间（小时）
const defaultTimeRelativity = 1

// ToAppConfig 将远程配置转换为本地配置格式
func (r *RemoteAppConfig) ToAppConfig() *AppConfig {
	faultPointTime := r.FaultPointPolicy.Expiration.TimeRelativity
	if faultPointTime <= 0 {
		faultPointTime = defaultTimeRelativity
	}

	problemTime := r.ProblemPolicy.Expiration.TimeRelativity
	if problemTime <= 0 {
		problemTime = defaultTimeRelativity
	}

	return &AppConfig{
		Credentials: CredentialsConfig{
			Authorization: fmt.Sprintf("Bearer %s", r.Platform.AuthToken),
		},
		KnowledgeNetwork: KnowledgeNetworkConfig{
			KnowledgeID: r.KnowledgeNetwork.KnowledgeID,
		},
		Ingest: IngestConfig{Source: Source{Type: "zabbix_webhook"}},
		FaultPoint: FaultPointExpirationCfg{
			Expiration: LocalExpirationConfig{
				ExpirationTime: time.Duration(faultPointTime) * time.Hour,
			},
		},
		Problem: ProblemExpirationCfg{
			Expiration: LocalExpirationConfig{
				ExpirationTime: time.Duration(problemTime) * time.Hour,
			},
		},
	}
}
