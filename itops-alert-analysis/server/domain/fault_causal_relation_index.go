package domain

import "time"

// FaultCausalRelation 因果推理关系（has_cause / has_effect）, 对应索引 itops_fault_causal_relation。
type FaultCausalRelation struct {
	RelationID         string    `json:"relation_id"`
	RelationClass      string    `json:"relation_class"`       // has_cause / has_effect
	RelationCreateTime time.Time `json:"relation_create_time"` // 实体创建时间
	RelationUpdateTime time.Time `json:"relation_update_time"` // 实体更新时间
	SourceObjectID     string    `json:"source_object_id"`
	SourceObjectClass  string    `json:"source_object_class"` // 源对象类
	TargetObjectID     string    `json:"target_object_id"`    // 目标对象ID
	TargetObjectClass  string    `json:"target_object_class"` // 目标对象类
}
