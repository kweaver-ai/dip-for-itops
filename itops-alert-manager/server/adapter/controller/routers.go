package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type HandlerRoute struct {
	pc ProblemController
	cf ConfigController
}

func (r *HandlerRoute) SetRouter(app *gin.Engine) {
	app.GET("/health", func(c *gin.Context) {
		rest.ReplyOK(c, http.StatusOK, nil)
	})
	group := app.Group("/api/itops_alert_manager/v1/")
	group.POST("problem", r.pc.List)
	group.PUT("problem/:problem_id/close", r.pc.Close)
	group.PUT("problem/:problem_id/root_cause", r.pc.SetRootCause)
	group.GET("problem/:problem_id/sub-graph", r.pc.GetSubGraphByProblemId)
	group.POST("config", r.cf.Create)
	group.PUT("config", r.cf.Update)
	group.GET("config", r.cf.ListByExt)

	inGroup := app.Group("/api/itops_alert_manager/v1/in/")
	inGroup.GET("config", r.cf.ListByIn)

}
