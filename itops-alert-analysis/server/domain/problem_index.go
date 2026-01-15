package domain

import "time"

type ProblemStatus string

const (
	ProblemStatusOpen    ProblemStatus = "0" //问题打开
	ProblemStatusClosed  ProblemStatus = "1" //问题关闭
	ProblemStatusExpired ProblemStatus = "2" //问题失效
	ProblemStatusMerged  ProblemStatus = "3" //问题被合并
)

type ProblemCloseType string

const (
	ProblemCloseTypeSystem ProblemCloseType = "1" //系统关闭
	ProblemCloseTypeManual ProblemCloseType = "2" //手动关闭
	//ProblemCloseTypeUnknown ProblemCloseType = "unknown"
)

// RCAStatus RCA 分析状态枚举
type RcaStatus int

const (
	RcaStatusPending   RcaStatus = iota + 1 // 未分析：等待开始分析
	RcaStatusRunning                        // 分析中：正在进行根因分析
	RcaStatusSuccess                        // 分析完成：成功完成根因分析
	RcaStatusFailed                         // 分析失败：分析过程中出现错误
	RcaStatusCancelled                      // 已取消：分析被取消
)

// Problem 对应索引 itops_problem。
type Problem struct {
	ProblemID              uint64            `json:"problem_id"`
	ProblemName            string            `json:"problem_name"`
	ProblemCreateTimestamp time.Time         `json:"problem_create_timestamp"`
	ProblemUpdateTime      time.Time         `json:"problem_update_time"`
	ProblemOccurTime       time.Time         `json:"problem_occur_time"`
	ProblemLatestTime      time.Time         `json:"problem_latest_time"`
	ProblemDuration        uint64            `json:"problem_duration"`
	ProblemDescription     string            `json:"problem_description"`
	ProblemStatus          ProblemStatus     `json:"problem_status"`
	ProblemCloseType       *ProblemCloseType `json:"problem_close_type,omitempty"`
	ProblemCloseNotes      string            `json:"problem_close_notes,omitempty"`
	ProblemClosedBy        string            `json:"problem_closed_by,omitempty"`
	ProblemCloseTime       *time.Time        `json:"problem_close_time,omitempty"`
	ProblemLevel           Severity          `json:"problem_level"`
	AffectedEntityIDs      []string          `json:"affected_entity_ids"`
	RelationIDs            []uint64          `json:"relation_fp_ids"`
	RelationEventIDs       []uint64          `json:"relation_event_ids"`
	RootCauseObjectID      string            `json:"root_cause_object_id"`
	RootCauseFaultID       uint64            `json:"root_cause_fault_id"`
	RcaResults             string            `json:"rca_results"`
	RcaStartTime           time.Time         `json:"rca_start_time"`
	RcaEndTime             time.Time         `json:"rca_end_time"`
	RcaStatus              RcaStatus         `json:"rca_status"`
}
