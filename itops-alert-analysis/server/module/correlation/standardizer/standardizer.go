package standardizer

import (
	"context"
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"github.com/pkg/errors"
)

// Standardizer 负责将上游原始 payload 转换为标准化 RawEvent。
// 不同来源独立实现，避免与具体业务耦合。
type Standardizer interface {
	Standardize(ctx context.Context, payload []byte) (domain.RawEvent, error)
}

// Build 根据数据源类型创建对应的 Standardizer。
func Build(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
	return defaultRegistry().Resolve(cfg, querier)
}

// Factory 创建具体标准化器。
type Factory func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error)

// Registry 管理不同数据源的标准化器。
type Registry struct {
	factories map[string]Factory
}

// NewRegistry 创建空注册表。
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register 注册数据源对应的标准化器。
func (r *Registry) Register(source string, factory Factory) {
	key := strings.TrimSpace(strings.ToLower(source))
	if key == "" || factory == nil {
		return
	}
	r.factories[key] = factory
}

// Resolve 根据数据源类型返回标准化器。
func (r *Registry) Resolve(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
	key := strings.TrimSpace(strings.ToLower(cfg.AppConfig.Ingest.Source.Type))
	factory, ok := r.factories[key]
	if !ok {
		return nil, errors.Errorf("unsupported source type: %s", cfg.AppConfig.Ingest.Source.Type)
	}
	return factory(cfg, querier)
}

// 默认注册表，包含内置标准化器。
func defaultRegistry() *Registry {
	r := NewRegistry()
	r.Register("zabbix_webhook", func(cfg *config.Config, querier ObjectClassQuerier) (Standardizer, error) {
		return NewZabbixWebhookStandardizer(cfg.AppConfig.Ingest, querier), nil
	})
	return r
}
