package domain

import "time"

type FaultPointRelation struct {
	RelationId         uint64    `json:"relation_id"`
	RelationClass      string    `json:"relation_class"`
	RelationCreateTime time.Time `json:"relation_create_time"`
	RelationUpdateTime time.Time `json:"relation_update_time"`
	SourceObjectId     string    `json:"source_object_id"`
	SourceObjectClass  string    `json:"source_object_class"`
	TargetObjectId     string    `json:"target_object_id"`
	TargetObjectClass  string    `json:"target_object_class"`
}
