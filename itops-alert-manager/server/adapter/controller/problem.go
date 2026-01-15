package controller

import (
	"net/http"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/DE_go-lib/rest"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/service"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/vo"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProblemController interface {
	List(c *gin.Context)
	Close(c *gin.Context)
	SetRootCause(c *gin.Context)
	GetSubGraphByProblemId(c *gin.Context)
}

type problemController struct {
	problemService    service.ProblemService
	authVerifyService service.AuthVerifyService
	validate          *validator.Validate
}

// List 查询problem
func (p *problemController) List(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)
	// token鉴权
	vistor, errAuth := p.authVerifyService.TokenVerify(ctx, c)
	if errAuth != nil {
		httpErr := HandDomainError(ctx, errAuth)
		rest.ReplyError(c, httpErr)
		return
	}
	req := vo.DataViewQueryV2{}
	if err := c.ShouldBind(&req); err != nil {
		httpErr := NewRestHTTPError(ctx, InvalidParameter).WithErrorDetails(common.ErrorDetailBind + err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	log.Debugf("request problem from host:%s ,req:%+v", c.Request.Host, req)
	// 参数检验
	if err := p.validate.Struct(&req); err != nil {
		httpErr := HandleValidateError(ctx, err)
		log.Errorf("problem request validate err:%s", err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	// 创建
	result, err := p.problemService.List(c, req, vistor.ID)
	if err != nil {
		log.Errorf("problem request failed err:%s", err.Error())
		httpErr := dependency.NewClientRequestError(err)
		rest.ReplyError(c, httpErr)
		return
	}
	rest.ReplyOK(c, http.StatusCreated, result)
}

func (p *problemController) Close(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)
	// token鉴权
	visitor, errAuth := p.authVerifyService.TokenVerify(ctx, c)
	if errAuth != nil {
		httpErr := HandDomainError(ctx, errAuth)
		rest.ReplyError(c, httpErr)
		return
	}
	problemId := c.Param("problem_id")

	// 创建
	err := p.problemService.Close(c, problemId, visitor.ID)
	if err != nil {
		log.Errorf("problem close failed err:%s", err.Error())
		httpErr := dependency.NewClientRequestError(err)
		rest.ReplyError(c, httpErr)
		return
	}
	resp := vo.BaseResp{Success: 1}
	rest.ReplyOK(c, http.StatusCreated, resp)
}

func (p *problemController) SetRootCause(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)
	// token鉴权
	_, errAuth := p.authVerifyService.TokenVerify(ctx, c)
	if errAuth != nil {
		httpErr := HandDomainError(ctx, errAuth)
		rest.ReplyError(c, httpErr)
		return
	}
	problemId := c.Param("problem_id")
	req := vo.RootCauseObjectIdParams{}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := NewRestHTTPError(ctx, InvalidParameter).WithErrorDetails(common.ErrorDetailBind + err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	log.Debugf("request SetRootCause from host:%s ,req:%+v", c.Request.Host, req)
	// 参数检验
	if err := p.validate.Struct(&req); err != nil {
		httpErr := HandleValidateError(ctx, err)
		log.Errorf("SetRootCause request validate err:%s", err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	// 创建
	err := p.problemService.SetRootCause(ctx, problemId, req)
	if err != nil {
		log.Errorf("SetRootCause request failed err:%s", err.Error())
		httpErr := dependency.NewClientRequestError(err)
		rest.ReplyError(c, httpErr)
		return
	}
	resp := vo.BaseResp{Success: 1}
	rest.ReplyOK(c, http.StatusCreated, resp)
}

// ListByExt 更新配置
func (p *problemController) GetSubGraphByProblemId(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)
	visitor, errAuth := p.authVerifyService.TokenVerify(ctx, c)
	if errAuth != nil {
		httpErr := HandDomainError(ctx, errAuth)
		rest.ReplyError(c, httpErr)
		return
	}
	problemId := c.Param("problem_id")
	result, err := p.problemService.GetSubGraphByProblemId(ctx, problemId, visitor.ID)
	if err != nil {
		log.Errorf("GetSubGraphByProblemId request failed err:%s", err.Error())
		httpErr := dependency.NewClientRequestError(err)
		rest.ReplyError(c, httpErr)
		return
	}
	rest.ReplyOK(c, http.StatusOK, result)
}
