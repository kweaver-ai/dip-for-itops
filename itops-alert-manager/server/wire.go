//go:build wireinject
// +build wireinject

package main

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/adapter/controller"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/adapter/repository"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/adapter/restapi/alert_analysis"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/adapter/restapi/isf"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/adapter/restapi/knowledge_network"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/adapter/restapi/uniquery"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/service"
	"github.com/google/wire"
)

func initServer() *core.RouterQuote {
	panic(wire.Build(repository.ProviderSet, uniquery.ProviderSet, alert_analysis.ProviderSet, isf.ProviderSet, knowledge_network.ProviderSet, service.ProviderSet, controller.ProviderSet))
}
