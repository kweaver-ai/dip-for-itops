package rca

import (
	"fmt"
	"sort"
	"strconv"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
)

// ========== 主要功能函数 ==========

// determineRootCause 确定根因
// 使用多因素评分机制，综合考虑因果关系、时间顺序、持续时间、严重程度和状态
func (s *Service) determineRootCause(faultPointInfos []domain.FaultPointObject, candidates []domain.CausalCandidate) *domain.FaultPointObject {
	if len(faultPointInfos) == 0 {
		return nil
	}

	// 步骤 1：构建时间线列表（按发生时间排序）
	timeline := s.buildFaultTimeline(faultPointInfos)
	if len(timeline) == 0 {
		return nil
	}

	// 如果没有因果关系，直接返回时间线上最早的故障点
	if len(candidates) == 0 {
		return timeline[0]
	}

	// 步骤 2：构建因果关系图并找到根因候选
	causalGraph := s.buildCausalGraph(faultPointInfos, candidates)
	if causalGraph == nil {
		// 如果构建失败，返回时间线上最早的故障点
		return timeline[0]
	}

	rootCauseCandidates := s.findRootCauseCandidates(timeline, causalGraph)

	// 步骤 3：从候选中选择最合适的根因
	return s.selectBestRootCause(rootCauseCandidates, timeline, causalGraph)
}

// buildFaultTimeline 构建故障时间线（按故障点发生时间排序，早的在前）
// 优化：改进排序稳定性，处理相同时间的情况
func (s *Service) buildFaultTimeline(faultPointInfos []domain.FaultPointObject) faultTimeline {
	if len(faultPointInfos) == 0 {
		return faultTimeline{}
	}

	timeline := make(faultTimeline, 0, len(faultPointInfos))
	for i := range faultPointInfos {
		timeline = append(timeline, &faultPointInfos[i])
	}

	// 使用 sort.Slice 按故障发生时间排序（早的在前）
	// 优化：如果时间相同，按故障ID排序以保证稳定性
	sort.Slice(timeline, func(i, j int) bool {
		if timeline[i] == nil || timeline[j] == nil {
			return false
		}
		// 先按时间排序
		if timeline[i].FaultOccurTime.Before(timeline[j].FaultOccurTime) {
			return true
		}
		if timeline[i].FaultOccurTime.After(timeline[j].FaultOccurTime) {
			return false
		}
		// 时间相同，按故障ID排序（保证稳定性）
		return timeline[i].FaultID < timeline[j].FaultID
	})

	return timeline
}

// buildCausalGraph 构建因果关系图
// 创建双向映射（原因->结果、结果->原因）和置信度存储
// 用于后续的根因分析和评分计算
func (s *Service) buildCausalGraph(faultPointInfos []domain.FaultPointObject, candidates []domain.CausalCandidate) *causalGraph {
	// 使用统一的初始化函数
	graph := newCausalGraph()
	if graph == nil {
		return nil
	}

	// 构建故障点ID到故障点对象的映射
	s.buildFaultPointMap(graph, faultPointInfos)

	// 构建因果关系图
	s.buildCausalRelations(graph, candidates)

	return graph
}

// buildFaultPointMap 构建故障点ID到故障点对象的映射
// 用于快速查找故障点对象
func (s *Service) buildFaultPointMap(graph *causalGraph, faultPointInfos []domain.FaultPointObject) {
	if graph == nil {
		return
	}

	for i := range faultPointInfos {
		id := s.formatFaultID(faultPointInfos[i].FaultID)
		graph.faultPointMap[id] = &faultPointInfos[i]
	}
}

// buildCausalRelations 构建因果关系
// 建立"原因->结果"和"结果->原因"的双向映射，并记录置信度
// 优化：添加重复检查，避免重复添加相同的因果关系
func (s *Service) buildCausalRelations(graph *causalGraph, candidates []domain.CausalCandidate) {
	if graph == nil {
		return
	}

	// 使用 map 记录已添加的因果关系，避免重复
	addedRelations := make(map[string]bool)

	for _, candidate := range candidates {
		if candidate.Cause == nil || candidate.Effect == nil {
			continue
		}

		causeID := s.formatFaultID(candidate.Cause.FaultID)
		effectID := s.formatFaultID(candidate.Effect.FaultID)

		// 跳过自引用
		if causeID == effectID {
			continue
		}

		cause := graph.faultPointMap[causeID]
		effect := graph.faultPointMap[effectID]

		if cause == nil || effect == nil {
			continue
		}

		// 检查是否已添加该因果关系
		key := s.buildCausalRelationKey(causeID, effectID)
		if addedRelations[key] {
			continue
		}
		addedRelations[key] = true
		// 记录 effect -> causes（反向关系，用于反向追溯）
		graph.effectToCauses[effectID] = append(graph.effectToCauses[effectID], cause)

		// 记录 cause -> effects（正向关系，用于正向扩散）
		graph.causeToEffects[causeID] = append(graph.causeToEffects[causeID], effect)

		// 记录因果关系置信度（如果已存在，取较大值）
		if existingConfidence, ok := graph.causalConfidenceMap[key]; ok {
			if candidate.Confidence > existingConfidence {
				graph.causalConfidenceMap[key] = candidate.Confidence
			}
		} else {
			graph.causalConfidenceMap[key] = candidate.Confidence
		}

		// 标记作为原因和结果的故障点
		graph.causeFaultIDs[causeID] = true
		graph.effectFaultIDs[effectID] = true
	}
}

// findRootCauseCandidates 找到根因候选
// 优化：使用多种策略逐步缩小候选范围，考虑更多因素
func (s *Service) findRootCauseCandidates(timeline faultTimeline, graph *causalGraph) []*domain.FaultPointObject {
	if graph == nil {
		// 如果图为空，返回时间线上最早的故障点
		if len(timeline) > 0 {
			return []*domain.FaultPointObject{timeline[0]}
		}
		return []*domain.FaultPointObject{}
	}

	// 方法1：找到只作为原因、不作为结果的故障点（这些是潜在的根因）
	candidates := s.findPureCauseFaultPoints(graph)

	// 方法2：如果没有找到只作为原因的故障点，找到作为原因次数最多的故障点
	if len(candidates) == 0 {
		candidates = s.findMostCausalFaultPoints(graph)
	}

	// 方法3：如果仍然没有找到，从时间线最早开始，找到第一个是其他故障点原因的故障点
	if len(candidates) == 0 {
		candidates = s.findEarliestCauseFaultPoint(timeline, graph)
	}

	// 方法4：如果仍然没有找到候选，返回时间线上最早的故障点
	if len(candidates) == 0 && len(timeline) > 0 {
		candidates = append(candidates, timeline[0])
	}

	return candidates
}

// findPureCauseFaultPoints 找到只作为原因、不作为结果的故障点
// 这些故障点只影响其他故障点，而不被其他故障点影响，是潜在的根因
func (s *Service) findPureCauseFaultPoints(graph *causalGraph) []*domain.FaultPointObject {
	if graph == nil {
		return []*domain.FaultPointObject{}
	}

	candidates := make([]*domain.FaultPointObject, 0)
	for causeID := range graph.causeFaultIDs {
		if !graph.effectFaultIDs[causeID] {
			// 这个故障点只作为原因，不作为结果，是根因候选
			if fp, ok := graph.faultPointMap[causeID]; ok && fp != nil {
				candidates = append(candidates, fp)
			}
		}
	}
	return candidates
}

// findMostCausalFaultPoints 找到作为原因次数最多的故障点
// 优化：新增方法，找到影响其他故障点最多的故障点
func (s *Service) findMostCausalFaultPoints(graph *causalGraph) []*domain.FaultPointObject {
	if graph == nil {
		return []*domain.FaultPointObject{}
	}

	maxEffectCount := 0
	candidates := make([]*domain.FaultPointObject, 0)

	for causeID := range graph.causeFaultIDs {
		effects, ok := graph.causeToEffects[causeID]
		if !ok {
			continue
		}

		effectCount := len(effects)
		if effectCount > maxEffectCount {
			maxEffectCount = effectCount
			candidates = candidates[:0] // 清空之前的候选
			if fp, ok := graph.faultPointMap[causeID]; ok && fp != nil {
				candidates = append(candidates, fp)
			}
		} else if effectCount == maxEffectCount && effectCount > 0 {
			// 如果有多个故障点的影响数量相同，都加入候选
			if fp, ok := graph.faultPointMap[causeID]; ok && fp != nil {
				candidates = append(candidates, fp)
			}
		}
	}

	return candidates
}

// findEarliestCauseFaultPoint 从时间线最早开始，找到第一个是其他故障点原因的故障点
// 如果存在因果关系，优先选择时间线上最早且是其他故障点原因的故障点
func (s *Service) findEarliestCauseFaultPoint(timeline faultTimeline, graph *causalGraph) []*domain.FaultPointObject {
	if graph == nil {
		return []*domain.FaultPointObject{}
	}

	for _, fp := range timeline {
		if fp == nil {
			continue
		}

		fpID := s.formatFaultID(fp.FaultID)
		if graph.causeFaultIDs[fpID] {
			// 这个故障点是其他故障点的原因，且是时间线上最早的
			return []*domain.FaultPointObject{fp}
		}
	}
	return []*domain.FaultPointObject{}
}

// selectBestRootCause 从候选中选择最合适的根因
// 优化：使用多因素评分机制，综合考虑更多因素
func (s *Service) selectBestRootCause(candidates []*domain.FaultPointObject, timeline faultTimeline, graph *causalGraph) *domain.FaultPointObject {
	if len(candidates) == 0 {
		return nil
	}

	// 过滤掉 nil 候选
	validCandidates := make([]*domain.FaultPointObject, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate != nil {
			validCandidates = append(validCandidates, candidate)
		}
	}

	if len(validCandidates) == 0 {
		return nil
	}

	if len(validCandidates) == 1 {
		return validCandidates[0]
	}

	return s.findHighestScoreCandidate(validCandidates, timeline, graph)
}

// findHighestScoreCandidate 找到评分最高的候选
// 遍历所有候选，计算每个候选的评分，返回评分最高的候选
func (s *Service) findHighestScoreCandidate(candidates []*domain.FaultPointObject, timeline faultTimeline, graph *causalGraph) *domain.FaultPointObject {
	if len(candidates) == 0 {
		return nil
	}

	var bestRootCause *domain.FaultPointObject
	maxScore := rootCauseScoreInitialMax

	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}

		score := s.calculateRootCauseScore(candidate, timeline, graph)
		if score > maxScore {
			maxScore = score
			bestRootCause = candidate
		}
	}

	return bestRootCause
}

// calculateRootCauseScore 计算根因候选的评分
// 优化：综合考虑五个因素，提升准确性
func (s *Service) calculateRootCauseScore(candidate *domain.FaultPointObject, timeline faultTimeline, graph *causalGraph) float64 {
	if candidate == nil {
		return 0.0
	}

	score := 0.0

	// 因素1：作为原因的置信度总和（权重 ×10）
	score += s.calculateConfidenceScore(candidate, graph)

	// 因素2：时间越早，分数越高（改进：考虑时间间隔）
	score += s.calculateTimeScore(candidate, timeline)

	// 因素3：持续时间越长，分数越高
	score += s.calculateDurationScore(candidate)

	// 因素4：严重程度越高，分数越高（新增）
	score += s.calculateSeverityScore(candidate)

	// 因素5：故障状态（已恢复的降低分数，未恢复的增加分数）（新增）
	score += s.calculateStatusScore(candidate)

	return score
}

// calculateConfidenceScore 计算置信度分数
// 计算该故障点作为原因的所有因果关系的置信度总和，乘以权重倍数
func (s *Service) calculateConfidenceScore(candidate *domain.FaultPointObject, graph *causalGraph) float64 {
	if candidate == nil || graph == nil {
		return 0.0
	}

	candidateID := s.formatFaultID(candidate.FaultID)
	effects, ok := graph.causeToEffects[candidateID]
	if !ok {
		return 0.0
	}

	confidenceScore := 0.0
	for _, effect := range effects {
		if effect == nil {
			continue
		}

		effectID := s.formatFaultID(effect.FaultID)
		key := s.buildCausalRelationKey(candidateID, effectID)
		if confidence, ok := graph.causalConfidenceMap[key]; ok {
			confidenceScore += confidence * confidenceWeightMultiplier
		}
	}
	return confidenceScore
}

// calculateTimeScore 计算时间分数（时间越早，分数越高）
// 优化：不仅统计晚于该故障点的数量，还考虑时间间隔
func (s *Service) calculateTimeScore(candidate *domain.FaultPointObject, timeline faultTimeline) float64 {
	if candidate == nil {
		return 0.0
	}

	timeScore := 0.0
	candidateTime := candidate.FaultOccurTime

	for _, fp := range timeline {
		if fp == nil {
			continue
		}

		if fp.FaultOccurTime.After(candidateTime) {
			// 基础分数：每早于一个故障点加1分
			timeScore += timeScorePerLaterFault

			// 额外分数：根据时间间隔给予奖励（越早奖励越多）
			timeDiff := fp.FaultOccurTime.Sub(candidateTime)
			if timeDiff > 0 {
				hoursDiff := timeDiff.Hours()
				timeScore += hoursDiff * timeScorePerHourEarly
			}
		}
	}

	// 限制时间分数最大值
	if timeScore > maxTimeScore {
		timeScore = maxTimeScore
	}

	return timeScore
}

// calculateDurationScore 计算持续时间分数（持续时间越长，分数越高）
// 将持续时间从秒转换为小时，最大限制为 maxDurationScore
func (s *Service) calculateDurationScore(candidate *domain.FaultPointObject) float64 {
	if candidate == nil {
		return 0.0
	}

	// 如果持续时间为负数，返回 0
	if candidate.FaultDurationTime < 0 {
		return 0.0
	}

	durationScore := float64(candidate.FaultDurationTime) / secondsPerHour
	if durationScore > maxDurationScore {
		durationScore = maxDurationScore
	}
	return durationScore
}

// calculateSeverityScore 计算严重程度分数（严重程度越高，分数越高）
// 优化：新增方法，FaultLevel 值越小表示越严重
func (s *Service) calculateSeverityScore(candidate *domain.FaultPointObject) float64 {
	if candidate == nil {
		return 0.0
	}

	// FaultLevel 值范围通常是 1-5，1 最严重，5 最轻微
	// 转换为分数：1 -> 10分, 2 -> 8分, 3 -> 6分, 4 -> 4分, 5 -> 2分
	severityScore := severityScoreWeight * (6.0 - float64(candidate.FaultLevel))
	if severityScore < 0 {
		severityScore = 0
	}
	if severityScore > maxSeverityScore {
		severityScore = maxSeverityScore
	}
	return severityScore
}

// calculateStatusScore 计算状态分数（已恢复的降低分数，未恢复的增加分数）
// 优化：新增方法，考虑故障状态对根因判断的影响
func (s *Service) calculateStatusScore(candidate *domain.FaultPointObject) float64 {
	if candidate == nil {
		return 0.0
	}

	switch candidate.FaultStatus {
	case domain.FaultStatusRecovered:
		// 已恢复的故障可能不是根因，降低分数
		return recoveredStatusPenalty
	case domain.FaultStatusOccurred:
		// 未恢复的故障更可能是根因，增加分数
		return occurredStatusBonus
	default:
		// 其他状态（如 expired）不加减分
		return 0.0
	}
}

// ========== 公共辅助函数 ==========

// formatFaultID 将故障点ID格式化为字符串
// 用于在因果关系图中作为键使用
func (s *Service) formatFaultID(id uint64) string {
	return strconv.FormatUint(id, 10)
}

// buildCausalRelationKey 构建因果关系键
// 格式：causeID->effectID，用于在因果关系图中唯一标识一对因果关系
func (s *Service) buildCausalRelationKey(causeID, effectID string) string {
	return fmt.Sprintf(causalRelationKeyFormat, causeID, effectID)
}
