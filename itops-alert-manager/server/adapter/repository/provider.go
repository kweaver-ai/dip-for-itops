package repository

import (
	"database/sql"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/infrastructure/db"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(db.NewDBAccess, NewConfigRepo)

func NewConfigRepo(db *sql.DB) dependency.ConfigRepo {
	return &configRepo{
		Repo:      core.Repo{DB: db},
		TableName: "t_config",
	}
}
