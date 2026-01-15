package controller

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewValidator, NewProblemController, NewConfigController, NewHandlerRoute, NewRouterQuote)

func NewValidator() *validator.Validate {
	va := validator.New()
	return va
}

// NewHandlerRoute 返回模板的路由
func NewHandlerRoute(problemController ProblemController, configController ConfigController) core.HttpRouter {
	return &HandlerRoute{
		pc: problemController,
		cf: configController,
	}
}

// NewRouterQuote 返回路由引用列表
func NewRouterQuote(handlerRoute core.HttpRouter) *core.RouterQuote {
	return &core.RouterQuote{Routes: []core.HttpRouter{
		handlerRoute,
	}}
}

// NewProblemController 返回problem控制器
func NewProblemController(validate *validator.Validate, problemService service.ProblemService, authVerifyService service.AuthVerifyService) ProblemController {
	return &problemController{
		problemService:    problemService,
		authVerifyService: authVerifyService,
		validate:          validate,
	}
}

// NewConfigController 返回problem控制器
func NewConfigController(validate *validator.Validate, authVerifyService service.AuthVerifyService, configService service.ConfigService) ConfigController {
	return &configController{
		configService:     configService,
		authVerifyService: authVerifyService,
		validate:          validate,
	}
}
