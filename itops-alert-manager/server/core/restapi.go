package core

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/config"
)

type RestAPIError interface {
	GetError() error
	Type() string
	Error() string
}

type RestAPI interface {
	RestAPI() *config.RestAPI
}
type external struct {
}

func NewCoreRestAPI() *external {
	return &external{}
}

func (external *external) RestAPI() *config.RestAPI {
	return &config.Get().RestAPI
}
