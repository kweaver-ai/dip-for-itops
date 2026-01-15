package app

import (
	"context"
	stderr "errors"
	"fmt"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/dip"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/api"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/correlation"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/rca"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/idgen"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// App 负责模块装配，当前实现均为占位。
type App struct {
	API         *api.Server
	Correlation *correlation.Service
	RCA         *rca.Service
}

func New(cfgManager *config.ConfigManager) (*App, error) {
	cfg := cfgManager.GetConfig()

	osClient, err := opensearch.NewClient(opensearch.OpenSearchConfig{
		Hosts:              []string{fmt.Sprintf("%s:%d", cfg.DepServices.OpenSearch.Host, cfg.DepServices.OpenSearch.Port)},
		Username:           cfg.DepServices.OpenSearch.User,
		Password:           cfg.DepServices.OpenSearch.Password,
		Timeout:            time.Second * 10,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "初始化 OpenSearch 失败")
	}

	repoFactory := opensearch.NewRepositoryFactory(osClient)

	// 初始化 DIP 客户端（动态获取 Authorization 和 KnowledgeID）
	dipClient := dip.NewClient(config.DIPConfig{
		Host:               cfg.Platform.BaseURL,
		KnID:               cfg.AppConfig.KnowledgeNetwork.KnowledgeID,
		Authorization:      cfg.AppConfig.Credentials.Authorization,
		InsecureSkipVerify: cfg.Platform.InsecureSkipVerify,
		Timeout:            cfg.Platform.Timeout,
	},
		func() string { return cfgManager.GetConfig().AppConfig.Credentials.Authorization },
		func() string { return cfgManager.GetConfig().AppConfig.KnowledgeNetwork.KnowledgeID },
	)

	// 模块装配（使用 Kafka 进行消息传递）
	corr, err := correlation.New(cfgManager, repoFactory, dipClient)
	if err != nil {
		return nil, errors.Wrap(err, "初始化 CorrelationService 失败")
	}

	apiServer, err := api.New(cfg, repoFactory, corr)
	if err != nil {
		return nil, errors.Wrap(err, "初始化 Api 失败")
	}
	// 初始化 RCA 服务
	rcaSvc, err := rca.New(
		*cfg,
		dipClient,
		idgen.New(),
		corr,
		repoFactory,
	)
	if err != nil {
		return nil, errors.Wrap(err, "初始化 Rca 失败")
	}

	return &App{
		API:         apiServer,
		Correlation: corr,
		RCA:         rcaSvc,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context 不能为空")
	}

	eg, egCtx := errgroup.WithContext(ctx)

	if a.Correlation != nil {
		eg.Go(func() error {
			if err := a.Correlation.Start(egCtx); err != nil && !errors.Is(err, context.Canceled) {
				return errors.Wrap(err, "correlation 启动失败")
			}
			return nil
		})
	}

	if a.API != nil {
		eg.Go(func() error {
			if err := a.API.Start(egCtx); err != nil && !errors.Is(err, context.Canceled) {
				return errors.Wrap(err, "api 启动失败")
			}
			return nil
		})
	}

	if a.RCA != nil {
		eg.Go(func() error {
			if err := a.RCA.Start(egCtx); err != nil && !errors.Is(err, context.Canceled) {
				return errors.Wrap(err, "rca 启动失败")
			}
			return nil
		})
	}

	log.Info("应用已启动，等待退出信号")
	return eg.Wait()
}

// Close 统一关闭持有的连接资源，需由上层在取消上下文后调用。
func (a *App) Close(ctx context.Context) error {
	var errs []error

	if a.API != nil {
		if err := a.API.Stop(ctx); err != nil && !errors.Is(err, context.Canceled) {
			errs = append(errs, errors.Wrap(err, "stop api"))
		}
	}
	if a.Correlation != nil {
		if err := a.Correlation.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "close correlation"))
		}
	}

	if a.RCA != nil {
		if err := a.RCA.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "close rca"))
		}
	}

	return stderr.Join(errs...)
}
