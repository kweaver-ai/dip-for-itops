package controller

import (
	"net/http"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/service"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/vo"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type ConfigController interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	ListByIn(c *gin.Context)
	ListByExt(c *gin.Context)
}

type configController struct {
	configService     service.ConfigService
	authVerifyService service.AuthVerifyService
	validate          *validator.Validate
}

// Create 创建配置
func (cf *configController) Create(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)
	// token鉴权
	_, errAuth := cf.authVerifyService.TokenVerify(ctx, c)
	if errAuth != nil {
		httpErr := HandDomainError(ctx, errAuth)
		rest.ReplyError(c, httpErr)
		return
	}
	req := vo.ConfigReq{}
	if err := c.ShouldBind(&req); err != nil {
		httpErr := NewRestHTTPError(ctx, InvalidParameter).WithErrorDetails(common.ErrorDetailBind + err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	log.Debugf("request create config from host:%s ,req:%+v", c.Request.Host, req)
	// 参数检验
	if err := cf.validate.Struct(&req); err != nil {
		log.Errorf("config create validate err:%s", err.Error())
		httpErr := HandleValidateError(ctx, err)
		log.Errorf("config create validate err:%s", err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	// 创建
	err := cf.configService.CreateConfig(c, &req)
	if err != nil {
		log.Errorf("config create failed err:%s", err.Error())
		httpErr := HandDomainError(ctx, err)
		rest.ReplyError(c, httpErr)
		return
	}
	resp := vo.BaseResp{Success: 1}
	rest.ReplyOK(c, http.StatusCreated, resp)
}

// Update 更新配置
func (cf *configController) Update(c *gin.Context) {
	req := vo.ConfigReq{}
	ctx := rest.GetLanguageCtx(c)
	// token鉴权
	_, errAuth := cf.authVerifyService.TokenVerify(ctx, c)
	if errAuth != nil {
		httpErr := HandDomainError(ctx, errAuth)
		rest.ReplyError(c, httpErr)
		return
	}
	if err := c.ShouldBind(&req); err != nil {
		httpErr := NewRestHTTPError(ctx, InvalidParameter).WithErrorDetails(common.ErrorDetailBind + err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	log.Debugf("request Update config from host:%s ,req:%+v", c.Request.Host, req)
	// 参数检验
	if err := cf.validate.Struct(&req); err != nil {
		httpErr := HandleValidateError(ctx, err)
		log.Errorf("config Update validate err:%s", err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	// 创建
	err := cf.configService.UpdateConfig(c, &req)
	if err != nil {
		log.Errorf("config Update failed err:%s", err.Error())
		httpErr := HandDomainError(ctx, err)
		rest.ReplyError(c, httpErr)
		return
	}
	resp := vo.BaseResp{Success: 1}
	rest.ReplyOK(c, http.StatusAccepted, resp)
}

// ListByIn 更新配置
func (cf *configController) ListByIn(c *gin.Context) {
	result, err := cf.configService.ListConfigs(c, true)
	if err != nil {
		log.Errorf("config Update failed err:%s", err.Error())
		httpErr := HandDomainError(c, err)
		rest.ReplyError(c, httpErr)
		return
	}
	rest.ReplyOK(c, http.StatusOK, result)
}

// ListByExt 更新配置
func (cf *configController) ListByExt(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)
	// token鉴权
	_, errAuth := cf.authVerifyService.TokenVerify(ctx, c)
	if errAuth != nil {
		httpErr := HandDomainError(ctx, errAuth)
		rest.ReplyError(c, httpErr)
		return
	}
	result, err := cf.configService.ListConfigs(ctx, false)
	if err != nil {
		log.Errorf("config Update failed err:%s", err.Error())
		httpErr := HandDomainError(c, err)
		rest.ReplyError(c, httpErr)
		return
	}
	rest.ReplyOK(c, http.StatusOK, result)
}
