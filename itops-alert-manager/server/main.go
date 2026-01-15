package main

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/migrations/0.1.0"
)

func main() {
	// 初始化服务配置
	config.InitPremise()
	__1_0.InitDataBase()
	router := initServer()
	core.InitHttpServer(router.Routes...)
}
