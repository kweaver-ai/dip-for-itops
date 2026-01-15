package domain

import "time"

type EventStatus string

const (
	EventStatusOccurred  EventStatus = "1" //发生
	EventStatusRecovered EventStatus = "2" //恢复
)

// RawEvent 对应索引 itops_raw_event。
// fault_id/problem_id 为冗余字段，减少下游查询关联成本。
type RawEvent struct {
	EventID           uint64      `json:"event_id"`
	RecoveryId        uint64      `json:"recovery_id"`
	EventProviderID   uint64      `json:"event_provider_id"`
	EventTimestamp    time.Time   `json:"event_timestamp"`
	EventTitle        string      `json:"event_title"`
	EventContent      string      `json:"event_content"`
	EventOccurTime    *time.Time  `json:"event_occur_time,omitempty"`
	EventRecoveryTime *time.Time  `json:"event_recovery_time,omitempty"`
	EventType         string      `json:"event_type"`
	EventStatus       EventStatus `json:"event_status"`
	EventLevel        Severity    `json:"event_level"`
	EventSource       string      `json:"event_source"`
	EntityObjectName  string      `json:"entity_object_name"`
	EntityObjectClass string      `json:"entity_object_class"`
	EntityObjectID    string      `json:"entity_object_id"`
	EntityObjectIP    string      `json:"entity_object_ip"`
	EntityObjectPort  string      `json:"entity_object_port"`
	EntityObjectMAC   string      `json:"entity_object_mac"`
	RawEventMsg       string      `json:"raw_event_msg"`
	ProblemID         uint64      `json:"problem_id"`
	FaultID           uint64      `json:"fault_id"`
}
