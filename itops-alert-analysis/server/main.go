package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/app"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"golang.org/x/sync/errgroup"
)

// 程序入口：读取 YAML 配置并装配各模块。
func main() {
	// 创建配置管理器
	cfgManager, err := config.NewConfigManager("config/config.yaml")
	if err != nil {
		log.Fatalf("创建配置管理器失败: %v", err)
	}

	cfg := cfgManager.GetConfig()

	// 初始化日志
	log.SetDefaultLog(&log.LogCfg{
		Filepath:    cfg.Log.Filepath,
		Level:       cfg.Log.Level,
		MaxSize:     cfg.Log.MaxSize,
		MaxAge:      cfg.Log.MaxAge,
		MaxBackups:  cfg.Log.MaxBackups,
		Compress:    cfg.Log.Compress,
		Development: cfg.Log.Development,
	})
	defer func() {
		_ = log.Sync()
	}()

	application, err := app.New(cfgManager)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := application.Close(shutdownCtx); err != nil {
			log.Errorf("close application: %v", err)
		}
		cfgManager.Stop()
	}()

	log.Infof("告警分析服务启动，source=%s, kafka_topic=%s, api_port=%d", cfg.AppConfig.Ingest.Source.Type, cfg.Kafka.RawEvents.Topic, cfg.API.Port)

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := cfgManager.Start(egCtx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		if err := application.Start(egCtx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatalf("start application: %v", err)
	}
}
