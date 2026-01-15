package dependency

import (
	"context"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/entity"
)

type ConfigRepo interface {
	Create(ctx context.Context, config *entity.Config) core.RepoError
	Update(ctx context.Context, config *entity.Config) core.RepoError
	ListAll(ctx context.Context) ([]*entity.Config, core.RepoError)
}
