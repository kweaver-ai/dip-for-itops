package dependency

import (
	"context"
)

type ISFUserInfo struct {
	Account string `json:"account"`
	Id      string `json:"id"`
}

//go:generate mockgen -source ./alert_analysis_restapi.go -destination ../../mock/adapter/restapi/mock_alert_analysis_restapi.go -package mock
type UserManagementClient interface {
	GetUserInfo(ctx context.Context, accountId string) (ISFUserInfo, error)
}
