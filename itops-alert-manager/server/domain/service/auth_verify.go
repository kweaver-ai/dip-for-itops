package service

import (
	"context"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/adapter/restapi/hydra"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source ./auth_verify.go -destination ../../mock/service/mock_auth_verify_service.go -package mock
type AuthVerifyService interface {
	TokenVerify(ctx context.Context, c *gin.Context) (hydra.Visitor, core.ServiceError)
}

type authVerifyService struct {
	hydra hydra.Hydra
}

func (r *authVerifyService) TokenVerify(ctx context.Context, c *gin.Context) (hydra.Visitor, core.ServiceError) {
	visitor, err := r.hydra.VerifyToken(ctx, c)
	if err != nil {
		log.Errorf("isf Unauthorized err:%s", err.Error())
		return visitor, NewSvUnauthorizedError(nil)
	}

	return visitor, nil
}
