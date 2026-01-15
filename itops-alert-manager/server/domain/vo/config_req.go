package vo

// ConfigRequest 配置保存请求体
type ConfigReq struct {
	Platform         PlatformConfig   `mapstructure:"platform" form:"platform" json:"platform" `
	KnowledgeNetwork KnowledgeNetwork `mapstructure:"knowledge_network" form:"knowledge_network" json:"knowledge_network"`
	FaultPointPolicy Policy           `mapstructure:"fault_point_policy" form:"fault_point_policy" json:"fault_point_policy" validate:"required"`
	ProblemPolicy    Policy           `mapstructure:"problem_policy" form:"problem_policy" json:"problem_policy" validate:"required"`
}

// PlatformConfig 平台连接配置
type PlatformConfig struct {
	AuthToken string `mapstructure:"auth_token" form:"auth_token"  json:"auth_token"`
}

// KnowledgeNetwork 知识网络配置
type KnowledgeNetwork struct {
	KnowledgeID string `mapstructure:"knowledge_id" form:"knowledge_id" json:"knowledge_id"`
}

// Policy 通用策略结构（用于故障点和问题策略）
type Policy struct {
	Expiration Expiration `mapstructure:"expiration" form:"expiration" json:"expiration" validate:"required"`
}

// Expiration 过期时间配置
type Expiration struct {
	TimeType       string `mapstructure:"time_type" form:"time_type" json:"time_type" validate:"omitempty,oneof=d h m"`            // d-天, h-小时, m-分钟
	TimeRelativity int    `mapstructure:"time_relativity" form:"time_relativity" json:"time_relativity" validate:"required,gte=1"` // 超过此时间未更新将自动失效或关闭
}
