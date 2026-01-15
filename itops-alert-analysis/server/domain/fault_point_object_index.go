package domain

import "time"

type FaultStatus string

const (
	FaultStatusOccurred  FaultStatus = "1" //发生：occurred
	FaultStatusRecovered FaultStatus = "2" //恢复：recovered
	FaultStatusExpired   FaultStatus = "3" //失效：expired
)

type Severity int

const (
	SeverityEmergency Severity = iota + 1 // 紧急
	SeverityCritical                      // 严重
	SeverityMajor                         // 重要
	SeverityWarning                       // 警告
	SeverityNormal                        // 正常
)

// FaultPointObject 对应索引 itops_fault_point。
// problem_id 冗余便于 Problem 直接定位。
type FaultPointObject struct {
	FaultID           uint64      `json:"fault_id"`
	FaultName         string      `json:"fault_name"`
	FaultCreateTime   time.Time   `json:"fault_create_time"`
	FaultUpdateTime   time.Time   `json:"fault_update_time"`
	FaultStatus       FaultStatus `json:"fault_status"`
	FaultOccurTime    time.Time   `json:"fault_occur_time"`
	FaultLatestTime   time.Time   `json:"fault_latest_time"`
	FaultDurationTime int64       `json:"fault_duration_time"`
	FaultRecoverTime  time.Time   `json:"fault_recover_time,omitempty"`
	EntityObjectClass string      `json:"entity_object_class"`
	EntityObjectName  string      `json:"entity_object_name"`
	EntityObjectID    string      `json:"entity_object_id"`
	RelationEventIDs  []uint64    `json:"relation_event_ids"`
	FaultMode         string      `json:"fault_mode"`
	FaultLevel        Severity    `json:"fault_level"`
	FaultDescription  string      `json:"fault_description"`
	ProblemID         uint64      `json:"problem_id"`
}
