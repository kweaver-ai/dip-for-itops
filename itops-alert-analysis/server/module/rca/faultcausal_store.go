package rca

import (
	"context"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"github.com/pkg/errors"
)

// ----- 故障因果存储接口实现 -----

// SaveFaultCausal 保存故障因果实体到 OpenSearch
// 如果实体已存在则更新，否则插入
func (s *Service) SaveFaultCausal(ctx context.Context, faultCausal *domain.FaultCausalObject) error {
	if err := s.validateFaultCausal(faultCausal); err != nil {
		return err
	}

	// 设置时间戳（仅在未设置时设置，保持创建时间不变）
	now := time.Now()
	if faultCausal.SCreateTime.IsZero() {
		faultCausal.SCreateTime = now
	}
	if faultCausal.SUpdateTime.IsZero() {
		faultCausal.SUpdateTime = now
	}

	// 调用存储仓库的 Upsert 方法
	return s.repoFactory.FaultCausals().Upsert(ctx, *faultCausal)
}

// SaveFaultCausalRelation 保存故障因果关系到 OpenSearch
// 如果关系已存在则更新，否则插入
func (s *Service) SaveFaultCausalRelation(ctx context.Context, relation *domain.FaultCausalRelation) error {
	if err := s.validateFaultCausalRelation(relation); err != nil {
		return err
	}

	// 设置时间戳（仅在未设置时设置，保持创建时间不变）
	now := time.Now()
	if relation.RelationCreateTime.IsZero() {
		relation.RelationCreateTime = now
	}
	if relation.RelationUpdateTime.IsZero() {
		relation.RelationUpdateTime = now
	}

	// 调用存储仓库的 Upsert 方法
	return s.repoFactory.FaultCausalRelations().Upsert(ctx, *relation)
}

// UpdateFaultCausal 更新故障因果实体到 OpenSearch
// 只更新可修改的字段，保留创建时间
func (s *Service) UpdateFaultCausal(ctx context.Context, faultCausal *domain.FaultCausalObject) error {
	if err := s.validateFaultCausal(faultCausal); err != nil {
		return err
	}

	// 设置更新时间（更新时间需要变）
	faultCausal.SUpdateTime = time.Now()

	// 调用存储仓库的 Update 方法
	return s.repoFactory.FaultCausals().Update(ctx, *faultCausal)
}

// UpdateFaultCausalRelation 更新故障因果关系到 OpenSearch
// 只更新可修改的字段，保留创建时间
func (s *Service) UpdateFaultCausalRelation(ctx context.Context, relation *domain.FaultCausalRelation) error {
	if err := s.validateFaultCausalRelation(relation); err != nil {
		return err
	}

	// 设置更新时间（更新时间需要变）
	relation.RelationUpdateTime = time.Now()

	// 调用存储仓库的 Update 方法
	return s.repoFactory.FaultCausalRelations().Update(ctx, *relation)
}

// ========== 私有辅助函数 ==========

// validateFaultCausal 验证故障因果实体
func (s *Service) validateFaultCausal(faultCausal *domain.FaultCausalObject) error {
	if faultCausal == nil {
		return errors.New("故障因果实体不能为 nil")
	}
	if s.repoFactory.FaultCausals() == nil {
		return errors.New("故障因果实体存储仓库未配置")
	}
	if faultCausal.CausalID == "" {
		return errors.New("故障因果实体 CausalID 不能为空")
	}
	return nil
}

// validateFaultCausalRelation 验证故障因果关系
func (s *Service) validateFaultCausalRelation(relation *domain.FaultCausalRelation) error {
	if relation == nil {
		return errors.New("故障因果关系不能为 nil")
	}
	if s.repoFactory.FaultCausalRelations() == nil {
		return errors.New("故障因果关系存储仓库未配置")
	}
	if relation.RelationID == "" {
		return errors.New("故障因果关系 RelationID 不能为空")
	}
	return nil
}
