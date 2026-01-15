package rca

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
)

// ========== Step3.2: 转换为"因果推理实体"和关系 ==========
// convertCandidatesToFaultCausals 将因果候选转换为"因果推理实体"和关系
func (s *Service) convertCandidatesToFaultCausals(ctx context.Context, candidates []domain.CausalCandidate) ([]domain.FaultCausalObject, []domain.FaultCausalRelation) {
	if len(candidates) == 0 {
		return []domain.FaultCausalObject{}, []domain.FaultCausalRelation{}
	}

	faultCausals := make([]domain.FaultCausalObject, 0, len(candidates))
	faultCausalRelations := make([]domain.FaultCausalRelation, 0, len(candidates)*2)
	now := time.Now()

	for _, candidate := range candidates {
		// 检查 Cause 和 Effect 是否为 nil
		if candidate.Cause == nil || candidate.Effect == nil {
			continue // 跳过无效的候选
		}
		cause := candidate.Cause
		effect := candidate.Effect

		// 验证故障点A和故障点B否有效
		if cause.FaultID == 0 || effect.FaultID == 0 {
			continue // 跳过无效的故障点ID
		}

		// 检查是否已存在因果关系（QueryByEntityPair 已支持双向查询）
		// 如果已存在，跳过该候选，避免重复创建
		if s.hasExistingCausalRelation(ctx, strconv.FormatUint(cause.FaultID, 10), strconv.FormatUint(effect.FaultID, 10)) {
			continue
		}

		// 创建新的因果关系和因果推理实体

		// 生成"因果推理实体" ID
		causalID := fmt.Sprintf(causalIDFormat, s.idGenerator.NextID())

		// 创建"因果推理实体"（FaultCausal）
		faultCausal := domain.FaultCausalObject{
			CausalID:         causalID,
			SCreateTime:      now,
			SUpdateTime:      now,
			CausalConfidence: candidate.Confidence, // 置信度存储在实体中
			CausalReason:     candidate.Reason,     // 原因存储在实体中
		}
		faultCausals = append(faultCausals, faultCausal)

		// 使用 ID 生成器生成起点关系ID
		sourceRelationID := fmt.Sprintf(relationIDFormatFaultCausalRelation, s.idGenerator.NextID())

		// 创建关系1：故障点实体A -> "因果推理实体"（has_cause 关系）
		relation1 := s.createFaultCausalRelation(
			sourceRelationID,
			strconv.FormatUint(cause.FaultID, 10), // 源：故障点A的ID
			FaultPointObjectClassID,               // 源类型：故障点A的类型
			causalID,                              // 目标："因果推理实体"ID
			faultCausalClass,                      // 目标类型：FaultCausal
			relationClassHasCause,
			now,
		)
		faultCausalRelations = append(faultCausalRelations, relation1)

		// 使用 ID 生成器生成终点关系ID
		targetRelationID := fmt.Sprintf(relationIDFormatFaultCausalRelation, s.idGenerator.NextID())
		// 创建关系2："因果推理实体" -> 故障点实体B（has_effect 关系）
		relation2 := s.createFaultCausalRelation(
			targetRelationID,
			causalID,                               // 源："因果推理实体"ID
			faultCausalClass,                       // 源类型：FaultCausal
			strconv.FormatUint(effect.FaultID, 10), // 目标：故障点B的ID
			FaultPointObjectClassID,                // 目标类型：故障点B的类型
			relationClassHasEffect,
			now,
		)
		faultCausalRelations = append(faultCausalRelations, relation2)
	}

	return faultCausals, faultCausalRelations
}

// 检查是否已存在因果关系
func (s *Service) hasExistingCausalRelation(ctx context.Context, causeID, effectID string) bool {
	// 查询关系（QueryByEntityPair 已支持双向查询，一次查询即可检查两个方向）
	relations, err := s.repoFactory.FaultCausalRelations().QueryByEntityPair(ctx, causeID, effectID)
	if err != nil {
		// 查询失败时返回 false（允许继续创建，避免查询失败导致数据丢失）
		return false
	}

	// 如果找到任何关系，说明已存在因果关系
	return len(relations) > 0
}

// createFaultCausalRelation 创建因果推理关系
func (s *Service) createFaultCausalRelation(
	relationID string,
	sourceID, sourceClass,
	targetID, targetClass,
	relationClass string,
	createTime time.Time,
) domain.FaultCausalRelation {
	// 参数验证
	if sourceID == "" || targetID == "" || relationClass == "" {
		// 返回一个无效的关系（调用者应该检查）
		return domain.FaultCausalRelation{}
	}

	return domain.FaultCausalRelation{
		RelationID:         relationID,
		RelationClass:      relationClass,
		RelationCreateTime: createTime,
		SourceObjectID:     sourceID,
		SourceObjectClass:  sourceClass,
		TargetObjectID:     targetID,
		TargetObjectClass:  targetClass,
	}
}

// ========== Step3.3: 检测并解决与 OpenSearch 中存储的因果关系冲突 ==========

// 检测并解决与 OpenSearch 中存储的因果关系冲突
func (s *Service) detectAndResolveOpenSearchCausalityConflicts(ctx context.Context, faultCausals *[]domain.FaultCausalObject, faultCausalRelations *[]domain.FaultCausalRelation) error {
	// 如果没有数据，直接返回
	if faultCausals == nil || faultCausalRelations == nil {
		return nil
	}
	if len(*faultCausals) == 0 && len(*faultCausalRelations) == 0 {
		return nil
	}

	// 1. 构建"因果推理实体"到关系的映射（用于按单元处理）
	causalToRelationsMap := s.buildCausalToRelationsMapForConflict(*faultCausalRelations)

	// 如果映射为空，直接返回
	if len(causalToRelationsMap) == 0 {
		return nil
	}

	// 2. 按单元处理：对于每个"因果推理实体"，检查其对应的两个关系
	now := time.Now()
	var saveErrors []error

	for causalID, relations := range causalToRelationsMap {
		// 验证 causalID 不为空
		if causalID == "" {
			saveErrors = append(saveErrors, errors.New("因果推理实体 ID 为空，跳过处理"))
			continue
		}

		// 验证 relations 不为空（一个因果推理单元应该有两个关系：has_cause 和 has_effect）
		if len(relations) == 0 {
			saveErrors = append(saveErrors, errors.New(fmt.Sprintf("因果推理单元 %s 的关系列表为空，跳过处理", causalID)))
			continue
		}

		// 查找对应的"因果推理实体"索引
		causalIdx := s.findCausalIndexByID(*faultCausals, causalID)
		if causalIdx == -1 {
			saveErrors = append(saveErrors, errors.New(fmt.Sprintf("未找到因果推理实体 %s 对应的实体对象", causalID)))
			continue
		}
		newCausal := &(*faultCausals)[causalIdx]

		// 通过关系中的实体对来判断是否已存在相同的因果关系
		existingCausal, hasCauseExists, hasEffectExists := s.findExistingCausalityByEntityPairs(ctx, relations)

		// 处理已存在的因果关系
		if existingCausal != nil && hasCauseExists && hasEffectExists {
			// 使用已存在的 CausalID（不更新 ID）
			newCausal.CausalID = existingCausal.CausalID

			// 更新关系中的 RelationID 以匹配已存在的关系（使用旧的 ID）
			s.updateRelationsWithExistingIDs(ctx, relations, existingCausal.CausalID, faultCausalRelations)

			if s.shouldUpdateCausal(newCausal, existingCausal) {
				// 需要更新：保留 SCreateTime，更新 SUpdateTime
				newCausal.SCreateTime = existingCausal.SCreateTime
				newCausal.SUpdateTime = now
			} else {
				// 不需要更新：保留已存储的数据，跳过保存操作
				*newCausal = *existingCausal
				continue
			}
		} else {
			// 关系不存在，新建因果关系
			if newCausal.CausalID == "" {
				newCausal.CausalID = causalID
			}
			if newCausal.SCreateTime.IsZero() {
				newCausal.SCreateTime = now
			}
			if newCausal.SUpdateTime.IsZero() {
				newCausal.SUpdateTime = now
			}
		}

		// 验证必要字段
		if newCausal.CausalID == "" {
			saveErrors = append(saveErrors, errors.New(fmt.Sprintf("因果推理单元 %s 的 CausalID 为空，跳过保存", causalID)))
			continue
		}

		// 使用 Upsert 操作：如果存在则更新，不存在则插入
		if err := s.upsertCausalUnitAtomically(ctx, newCausal, relations); err != nil {
			saveErrors = append(saveErrors, errors.New(fmt.Sprintf("保存因果推理单元 %s 失败: %v", causalID, err)))
			continue
		}
	}
	// 如果有错误，返回汇总错误信息
	if len(saveErrors) > 0 {
		errMsg := fmt.Sprintf("部分因果推理单元处理失败，共 %d 个单元失败", len(saveErrors))
		for i, err := range saveErrors {
			if err == nil {
				continue
			}
			if i == 0 {
				errMsg += fmt.Sprintf(": %v", err)
			} else {
				errMsg += fmt.Sprintf("; %v", err)
			}
		}
		return errors.New(errMsg)
	}
	return nil
}

// findCausalIndexByID 根据 CausalID 查找在 faultCausals 中的索引
func (s *Service) findCausalIndexByID(faultCausals []domain.FaultCausalObject, causalID string) int {
	if causalID == "" {
		return -1
	}
	for i, causal := range faultCausals {
		if causal.CausalID == causalID {
			return i
		}
	}
	return -1
}

// findExistingCausalityByEntityPairs 通过实体对查找已存在的因果关系
// 返回：已存在的因果推理实体、has_cause 关系是否存在、has_effect 关系是否存在
func (s *Service) findExistingCausalityByEntityPairs(ctx context.Context, relations []domain.FaultCausalRelation) (*domain.FaultCausalObject, bool, bool) {
	// 参数验证
	if len(relations) == 0 {
		return nil, false, false
	}

	var hasCauseRelation, hasEffectRelation *domain.FaultCausalRelation
	var causalID string

	// 分离 has_cause 和 has_effect 关系
	for i := range relations {
		relation := &relations[i]

		// 验证关系的关键字段是否有效
		if relation.RelationClass == "" {
			continue
		}
		switch relation.RelationClass {
		case relationClassHasCause:
			hasCauseRelation = relation
			// has_cause: target 是 FaultCausal
			causalID = relation.TargetObjectID
		case relationClassHasEffect:
			hasEffectRelation = relation
		}
	}

	// 如果找不到 has_cause 关系，无法确定 CausalID
	if hasCauseRelation == nil {
		return nil, false, false
	}

	// 验证存储仓库已初始化
	if s.repoFactory.FaultCausals() == nil {
		return nil, false, false
	}

	// 先通过 CausalID 查找已存在的因果推理实体
	if causalID != "" {
		causals, err := s.repoFactory.FaultCausals().QueryByIDs(ctx, []string{causalID})
		if err == nil && len(causals) > 0 {
			existingCausal := &causals[0]

			// 验证关系是否存在（通过实体对查询）
			hasCauseExists := s.findRelationByEntityPair(ctx, hasCauseRelation.SourceObjectID, hasCauseRelation.TargetObjectID, relationClassHasCause) != nil
			hasEffectExists := false
			if hasEffectRelation != nil {
				hasEffectExists = s.findRelationByEntityPair(ctx, hasEffectRelation.SourceObjectID, hasEffectRelation.TargetObjectID, relationClassHasEffect) != nil
			}

			// 只有当两个关系都存在时才返回 true
			// 修复：如果 hasEffectRelation 为 nil，说明关系列表不完整，不应该返回 true
			if hasCauseExists && hasEffectExists && hasEffectRelation != nil {
				return existingCausal, true, true
			}
		}
	}

	// 如果找不到，尝试通过实体对查询关系
	hasCauseExists := s.findRelationByEntityPair(ctx, hasCauseRelation.SourceObjectID, hasCauseRelation.TargetObjectID, relationClassHasCause) != nil
	hasEffectExists := false
	if hasEffectRelation != nil {
		hasEffectExists = s.findRelationByEntityPair(ctx, hasEffectRelation.SourceObjectID, hasEffectRelation.TargetObjectID, relationClassHasEffect) != nil
	}

	// 如果两个关系都存在，查找对应的因果推理实体
	if hasCauseExists && hasEffectExists && hasEffectRelation != nil {
		// 从 has_cause 关系中获取 CausalID
		causalID = hasCauseRelation.TargetObjectID
		if causalID != "" {
			causals, err := s.repoFactory.FaultCausals().QueryByIDs(ctx, []string{causalID})
			if err == nil && len(causals) > 0 {
				return &causals[0], true, true
			}
		}
	}

	return nil, hasCauseExists, hasEffectExists
}

// findRelationByEntityPair 通过实体对查找关系
// QueryByEntityPair 已支持双向查询（sourceID -> targetID 或 targetID -> sourceID）
// 如果提供了 relationClass，只返回匹配该类型的关系
func (s *Service) findRelationByEntityPair(ctx context.Context, sourceID, targetID, relationClass string) *domain.FaultCausalRelation {
	// 参数验证
	if sourceID == "" || targetID == "" {
		return nil
	}

	// 如果存储仓库未初始化，返回 nil
	if s.repoFactory.FaultCausalRelations() == nil {
		return nil
	}

	// 查询关系（QueryByEntityPair 已支持双向查询）
	relations, err := s.repoFactory.FaultCausalRelations().QueryByEntityPair(ctx, sourceID, targetID)
	if err != nil {
		// 查询失败时返回 nil（不记录错误，因为这是正常的查询流程）
		return nil
	}

	if len(relations) == 0 {
		return nil
	}

	// 如果提供了 relationClass，过滤匹配的关系
	if relationClass != "" {
		for i := range relations {
			if relations[i].RelationClass == relationClass {
				return &relations[i]
			}
		}
		// 没有找到匹配 relationClass 的关系
		return nil
	}

	// 如果没有提供 relationClass，返回第一个关系
	return &relations[0]
}

// shouldUpdateCausal 判断是否需要更新因果推理实体
func (s *Service) shouldUpdateCausal(newCausal, existingCausal *domain.FaultCausalObject) bool {
	if newCausal == nil || existingCausal == nil {
		return false
	}
	// 如果新数据的置信度更高，需要更新
	if newCausal.CausalConfidence > existingCausal.CausalConfidence {
		return true
	}
	// 如果新数据的原因更详细（非空且与已存储的不同），需要更新
	if newCausal.CausalReason != "" && newCausal.CausalReason != existingCausal.CausalReason {
		return true
	}
	return false
}

// updateRelationsWithExistingIDs 更新关系中的 RelationID（使用已存在的旧 ID）
func (s *Service) updateRelationsWithExistingIDs(ctx context.Context, relations []domain.FaultCausalRelation, existingCausalID string, faultCausalRelations *[]domain.FaultCausalRelation) {
	// 参数验证
	if faultCausalRelations == nil || len(relations) == 0 || existingCausalID == "" {
		return
	}

	// 构建关系索引映射，优化查找性能（O(1) 查找）
	relationKeyMap := make(map[string]int, len(*faultCausalRelations))
	for i := range *faultCausalRelations {
		rel := &(*faultCausalRelations)[i]
		key := fmt.Sprintf("%s|%s|%s", rel.SourceObjectID, rel.TargetObjectID, rel.RelationClass)
		relationKeyMap[key] = i
	}

	now := time.Now()

	// 遍历新生成的关系，查找已存在的关系并使用旧的 RelationID
	for _, newRelation := range relations {
		// 通过实体对查找已存在的关系
		existingRelation := s.findRelationByEntityPair(ctx, newRelation.SourceObjectID, newRelation.TargetObjectID, newRelation.RelationClass)
		if existingRelation == nil {
			continue // 未找到已存在的关系，保留新生成的关系 ID
		}

		// 通过索引快速查找并更新对应关系（使用旧的 RelationID）
		key := fmt.Sprintf("%s|%s|%s", newRelation.SourceObjectID, newRelation.TargetObjectID, newRelation.RelationClass)
		if idx, exists := relationKeyMap[key]; exists {
			relation := &(*faultCausalRelations)[idx]
			// 使用已存在的关系 ID（不更新 ID）
			relation.RelationID = existingRelation.RelationID
			// 保留创建时间（创建时间不变）
			relation.RelationCreateTime = existingRelation.RelationCreateTime
			// 关系内容可能变化，更新 RelationUpdateTime（更新时间需要变）
			relation.RelationUpdateTime = now
		}
	}
}

// buildCausalToRelationsMapForConflict 构建"因果推理实体"到关系的映射（用于冲突检测）
// 返回映射：map[CausalID][]FaultCausalRelation
// 一个因果关系单元通常包含两个关系：has_cause 和 has_effect
func (s *Service) buildCausalToRelationsMapForConflict(faultCausalRelations []domain.FaultCausalRelation) map[string][]domain.FaultCausalRelation {
	if len(faultCausalRelations) == 0 {
		return make(map[string][]domain.FaultCausalRelation)
	}

	// 预分配容量：每个因果关系单元有2个关系
	causalToRelations := make(map[string][]domain.FaultCausalRelation, len(faultCausalRelations)/2)

	for i := range faultCausalRelations {
		relation := &faultCausalRelations[i]

		// 验证关系的关键字段是否有效
		if relation.RelationClass == "" {
			continue
		}

		// 通过检查 SourceObjectClass 和 TargetObjectClass 来判断哪个是 FaultCausal
		// 不依赖关系类型，更通用和健壮
		var causalID string
		if relation.SourceObjectClass == faultCausalClass {
			// Source 是 FaultCausal（has_effect 关系）
			causalID = relation.SourceObjectID
		} else if relation.TargetObjectClass == faultCausalClass {
			// Target 是 FaultCausal（has_cause 关系）
			causalID = relation.TargetObjectID
		} else {
			// 如果 Source 和 Target 都不是 FaultCausal，跳过
			continue
		}

		// 验证 causalID 不为空后再添加到映射
		if causalID != "" {
			causalToRelations[causalID] = append(causalToRelations[causalID], *relation)
		}
	}

	return causalToRelations
}

// upsertCausalUnitAtomically 原子性保存（更新或插入）一个因果推理单元
func (s *Service) upsertCausalUnitAtomically(ctx context.Context, faultCausal *domain.FaultCausalObject, relations []domain.FaultCausalRelation) error {
	// 参数验证
	if faultCausal == nil {
		return errors.New("因果推理实体不能为 nil")
	}

	// 验证 CausalID 不为空
	if faultCausal.CausalID == "" {
		return errors.New("因果推理实体 CausalID 不能为空")
	}

	// 验证存储仓库已初始化
	if s.repoFactory.FaultCausals() == nil {
		return errors.New("故障因果实体存储仓库未配置")
	}
	if s.repoFactory.FaultCausalRelations() == nil {
		return errors.New("故障因果关系存储仓库未配置")
	}

	// 验证 relations 不为空
	if len(relations) == 0 {
		return errors.New("因果推理关系列表不能为空")
	}

	// 步骤1：保存"因果推理实体"（使用 Upsert：存在则更新，不存在则插入）
	if err := s.repoFactory.FaultCausals().Upsert(ctx, *faultCausal); err != nil {
		return errors.Wrapf(err, "保存因果推理实体失败")
	}

	// 步骤2：保存两个关系（使用 Upsert：存在则更新，不存在则插入）
	for _, relation := range relations {
		// 验证关系的必要字段
		if relation.RelationID == "" {
			return errors.New("因果推理关系 RelationID 不能为空")
		}
		if relation.SourceObjectID == "" || relation.TargetObjectID == "" {
			return errors.New("因果推理关系的 SourceObjectID 或 TargetObjectID 不能为空")
		}

		if err := s.repoFactory.FaultCausalRelations().Upsert(ctx, relation); err != nil {
			return errors.Wrapf(err, "保存因果推理关系 %s (%s) 失败", relation.RelationID, relation.RelationClass)
		}
	}

	return nil
}
