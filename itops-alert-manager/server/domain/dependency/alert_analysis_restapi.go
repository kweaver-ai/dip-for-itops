package dependency

import (
	"context"
)

type ProblemCloseBody struct {
	CloseType string `json:"close_type"`
	ClosedBy  string `json:"closed_by"`
	Notes     string `json:"notes"`
}

type RootCauseObjectIdParams struct {
	RootCauseObjectId string `json:"root_cause_object_id"`
	RootCauseFaultID  uint64 `json:"root_cause_fault_id"`
}

//go:generate mockgen -source ./uniquery_restapi.go -destination ../../mock/adapter/restapi/mock_uniquery_restapi.go -package mock
type AlertAnalysisClient interface {
	Close(ctx context.Context, problemId, closeBy string) error
	SetRootCause(ctx context.Context, problemId, rootCauseObjectId string, rootCauseFaultID uint64) error
}
