package domain

import "time"

// FaultCausalObject 因果推理对象，对应索引 itops_fault_causal。
// 表示故障之间的因果关系实体。
type FaultCausalObject struct {
	CausalID         string    `json:"causal_id"`         // 因果实体ID（唯一标识）
	SCreateTime      time.Time `json:"s_create_time"`     // 实体创建时间
	SUpdateTime      time.Time `json:"s_update_time"`     // 实体更新时间
	CausalConfidence float64   `json:"causal_confidence"` // 因果关系置信度（0.0-1.0）
	CausalReason     string    `json:"causal_reason"`     // 因果关系原因描述
}
