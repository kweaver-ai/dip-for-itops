// repository/config_repo_impl.go

package repository

import (
	"context"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/entity"
	"github.com/Masterminds/squirrel"
)

type configRepo struct {
	core.Repo
	TableName string
}

// Create 创建配置项
func (repo *configRepo) Create(ctx context.Context, config *entity.Config) core.RepoError {
	query := squirrel.Insert(repo.TableName).
		Columns("f_config_key", "f_config_value").
		Values(config.ConfigKey, config.ConfigValue)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		log.Errorf("Failed to build SQL for create config: %v", err)
		return dependency.NewRepoExecuteSqlError(err)
	}

	_, err = repo.DB.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		log.Errorf("Failed to insert config: %v", err)
		return dependency.NewRepoExecuteSqlError(err)
	}
	return nil
}

// Update 更新配置项（按 ConfigKey）
func (repo *configRepo) Update(ctx context.Context, config *entity.Config) core.RepoError {
	query := squirrel.Update(repo.TableName).
		SetMap(map[string]interface{}{
			"f_config_value": config.ConfigValue,
		}).
		Where("f_config_key = ?", config.ConfigKey)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		log.Errorf("Failed to build SQL for update config: %v", err)
		return dependency.NewRepoExecuteSqlError(err)
	}

	_, err = repo.DB.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		log.Errorf("Failed to update config: %v", err)
		return dependency.NewRepoExecuteSqlError(err)
	}

	return nil
}

// ListAll 获取所有配置项
func (repo *configRepo) ListAll(ctx context.Context) ([]*entity.Config, core.RepoError) {
	query := squirrel.Select("*").From(repo.TableName)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		log.Errorf("Failed to build SQL for list all configs: %v", err)
		return nil, dependency.NewRepoExecuteSqlError(err)
	}

	rows, err := repo.DB.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		log.Errorf("Failed to query configs: %v", err)
		return nil, dependency.NewRepoExecuteSqlError(err)
	}
	defer rows.Close()

	var configs []*entity.Config
	for rows.Next() {
		var config entity.Config
		err := rows.Scan(&config.ConfigKey, &config.ConfigValue)
		if err != nil {
			log.Errorf("Failed to scan config row: %v", err)
			return nil, dependency.NewRepoExecuteSqlError(err)
		}
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		log.Errorf("Rows iteration error: %v", err)
		return nil, dependency.NewRepoExecuteSqlError(err)
	}

	return configs, nil
}
