package rca

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/dip"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"golang.org/x/sync/errgroup"
)

// ========== 主要功能函数 ==========

// FaultPointPair 故障点对结构
type FaultPointPair struct {
	FpA      *domain.FaultPointObject
	FpB      *domain.FaultPointObject
	Priority float64 // 优先级分数
}

// 使用 Agent 进行因果推理
// 策略：1) 只过滤明显不相关的对 2) 提高并发数 3) 智能优先级排序 4) 保证所有相关对都被分析
func (s *Service) findCausalCandidates(ctx context.Context, faultPointInfos []domain.FaultPointObject, recallCtx *domain.GraphRecallContext) []domain.CausalCandidate {
	totalPairs := len(faultPointInfos) * (len(faultPointInfos) - 1) / 2
	startTime := time.Now()
	log.Infof("开始使用 Agent 进行因果推理，故障点数量: %d, 预计分析对数: %d", len(faultPointInfos), totalPairs, startTime.Format(time.RFC3339))

	// 步骤1：预过滤 - 只过滤明显不相关的故障点对（保守策略）
	filteredPairs := s.filterFaultPointPairsConservative(faultPointInfos, recallCtx)
	log.Infof("预过滤后，需要分析的故障点对数: %d (过滤掉 %d 对)",
		len(filteredPairs), totalPairs-len(filteredPairs))

	// 步骤2：按优先级排序，优先分析重要的故障点对
	sortedPairs := s.sortPairsByPriority(filteredPairs)

	// 步骤3：并发处理所有故障点对（保证准确性：所有对都分析）
	candidates := s.processAllPairsConcurrently(ctx, sortedPairs, recallCtx)

	log.Infof("因果推理完成: 分析 %d 对故障点, 发现 %d 个因果关系", len(sortedPairs), len(candidates), time.Since(startTime))

	return candidates
}

// filterFaultPointPairsConservative 保守的预过滤策略
// 只过滤明显不相关的故障点对，保证准确性
func (s *Service) filterFaultPointPairsConservative(faultPointInfos []domain.FaultPointObject, recallCtx *domain.GraphRecallContext) []FaultPointPair {
	var pairs []FaultPointPair

	for i := 0; i < len(faultPointInfos); i++ {
		for j := i + 1; j < len(faultPointInfos); j++ {
			fpA := &faultPointInfos[i]
			fpB := &faultPointInfos[j]

			// 过滤条件1：必须有拓扑关联（这是必须的，保证准确性）
			// 没有拓扑关联的故障点对不可能有因果关系
			if !s.hasTopologyRelation(fpA.EntityObjectID, fpB.EntityObjectID, recallCtx) {
				log.Debugf("故障点 %d 和 %d 无拓扑关联，跳过", fpA.FaultID, fpB.FaultID)
				continue
			}

			pairs = append(pairs, FaultPointPair{
				FpA: fpA,
				FpB: fpB,
			})
		}
	}

	return pairs
}

// sortPairsByPriority 按优先级排序故障点对
// 优先分析重要的对，但保证所有对都会被分析
func (s *Service) sortPairsByPriority(pairs []FaultPointPair) []FaultPointPair {
	// 计算每个对的优先级分数
	for i := range pairs {
		pairs[i].Priority = s.calculatePairPriorityScore(pairs[i].FpA, pairs[i].FpB)
	}

	// 按优先级降序排序（分数高的先分析）
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Priority > pairs[j].Priority
	})

	return pairs
}

// calculatePairPriorityScore 计算故障点对的优先级分数
// 优化：修复时间差计算逻辑，避免重复计算
func (s *Service) calculatePairPriorityScore(fpA, fpB *domain.FaultPointObject) float64 {
	score := 0.0

	// 因素1：严重程度（值越小越严重，分数越高）
	// 至少有一个严重程度高的故障点
	minLevel := fpA.FaultLevel
	if fpB.FaultLevel < minLevel {
		minLevel = fpB.FaultLevel
	}
	score += float64(6-minLevel) * 3.0 // 1级->15分, 2级->12分, 3级->9分

	// 因素2：时间间隔（间隔越小，分数越高，越可能是直接因果关系）
	// 优化：使用绝对值，避免重复计算
	timeDiff := fpA.FaultOccurTime.Sub(fpB.FaultOccurTime)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff <= veryShortTimeThreshold {
		score += 10.0 // 5分钟内，高优先级
	} else if timeDiff <= shortTimeThreshold {
		score += 7.0 // 30分钟内，中高优先级
	} else if timeDiff <= longTimeThreshold {
		score += 4.0 // 1小时内，中等优先级
	} else if timeDiff <= 6*time.Hour {
		score += 2.0 // 6小时内，较低优先级
	} else {
		// 注意：由于已经过滤了超过2小时的，这里不会执行
		// 但保留逻辑以防过滤条件改变
		score += 0.5 // 超过6小时，低优先级但仍分析
	}

	// 因素3：故障状态（未恢复的故障点分数更高，更可能是当前根因）
	if fpA.FaultStatus == domain.FaultStatusOccurred {
		score += 3.0
	}
	if fpB.FaultStatus == domain.FaultStatusOccurred {
		score += 3.0
	}

	// 因素4：持续时间（持续时间长的可能是持续性根因）
	if fpA.FaultDurationTime > 0 {
		hours := float64(fpA.FaultDurationTime) / 3600.0
		score += hours * 0.3 // 每小时0.3分，最多约7.2分（24小时）
	}
	if fpB.FaultDurationTime > 0 {
		hours := float64(fpB.FaultDurationTime) / 3600.0
		score += hours * 0.3
	}

	return score
}

// processAllPairsConcurrently 并发处理所有故障点对
// 优化：修复错误处理和边界情况
func (s *Service) processAllPairsConcurrently(ctx context.Context, pairs []FaultPointPair, recallCtx *domain.GraphRecallContext) []domain.CausalCandidate {
	if len(pairs) == 0 {
		return []domain.CausalCandidate{}
	}

	// 使用 errgroup 控制并发数量
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentAnalysis)

	// 共享数据结构，使用互斥锁保护
	var (
		candidatesMu sync.Mutex
		candidates   []domain.CausalCandidate
		processed    int64
		agentSuccess int64
		localSuccess int64
	)

	// 并发处理所有故障点对
	for _, pair := range pairs {
		// 捕获循环变量（重要：避免闭包问题）
		fpA := pair.FpA
		fpB := pair.FpB
		priority := pair.Priority

		g.Go(func() error {
			// 检查上下文是否已取消
			if gctx.Err() != nil {
				return gctx.Err()
			}

			atomic.AddInt64(&processed, 1)

			// 策略：根据优先级决定是否优先使用Agent
			// 高优先级的对使用Agent，低优先级的对直接使用本地规则（更快）
			// 但保证所有对都被分析
			useAgent := priority > 10.0 // 优先级超过10的对使用Agent

			var pairCandidates []domain.CausalCandidate
			if useAgent {
				// 尝试使用 Agent 进行因果推理
				agentCandidates := s.tryAgentCausalAnalysis(gctx, fpA, fpB, recallCtx)

				if len(agentCandidates) > 0 {
					pairCandidates = agentCandidates
					atomic.AddInt64(&agentSuccess, 1)
					log.Infof("Agent 因果推理成功: 故障点A ID=%d, 故障点B ID=%d, 发现 %d 个因果关系, 优先级: %.2f",
						fpA.FaultID, fpB.FaultID, len(agentCandidates), priority)
				} else {
					// Agent失败，使用本地规则作为兜底（保证准确性）
					fallbackCandidate := s.calculateCausalCandidate(fpA, fpB, recallCtx)
					if fallbackCandidate != nil {
						pairCandidates = []domain.CausalCandidate{*fallbackCandidate}
						log.Infof("Agent失败，使用本地规则: 故障点A ID=%d, 故障点B ID=%d, 置信度: %.2f, 优先级: %.2f",
							fpA.FaultID, fpB.FaultID, fallbackCandidate.Confidence, priority)
					}
				}
			} else {
				// 低优先级的对直接使用本地规则（更快，但准确性由本地规则保证）
				fallbackCandidate := s.calculateCausalCandidate(fpA, fpB, recallCtx)
				if fallbackCandidate != nil {
					pairCandidates = []domain.CausalCandidate{*fallbackCandidate}
					atomic.AddInt64(&localSuccess, 1)
					log.Infof("本地规则分析: 故障点A ID=%d, 故障点B ID=%d, 置信度: %.2f, 优先级: %.2f",
						fpA.FaultID, fpB.FaultID, fallbackCandidate.Confidence, priority)
				}
			}

			// 添加结果（线程安全）
			if len(pairCandidates) > 0 {
				candidatesMu.Lock()
				candidates = append(candidates, pairCandidates...)
				candidatesMu.Unlock()
			}

			return nil
		})
	}

	// 等待所有任务完成
	if err := g.Wait(); err != nil {
		log.Errorf("并发分析过程中发生错误: %v", err)
		// 即使有错误，也返回已收集的结果
	}

	log.Infof("分析完成: 处理 %d 对, Agent成功 %d 对, 本地规则 %d 对, 发现 %d 个因果关系",
		atomic.LoadInt64(&processed), atomic.LoadInt64(&agentSuccess), atomic.LoadInt64(&localSuccess), len(candidates))

	return candidates
}

// 尝试使用 Agent 进行因果推理，添加 token 长度控制和拓扑数据精简，失败直接返回空列表（不重试）
func (s *Service) tryAgentCausalAnalysis(ctx context.Context, fpA, fpB *domain.FaultPointObject, recallCtx *domain.GraphRecallContext) []domain.CausalCandidate {
	// 参数验证
	if fpA == nil || fpB == nil {
		return []domain.CausalCandidate{}
	}

	// 为每个调用创建独立的超时上下文，避免单个调用阻塞整个流程
	// 提前创建，可以更早检测超时
	agentCtx, cancel := context.WithTimeout(ctx, agentCallTimeout)
	defer cancel()

	// 检查上下文是否已取消或超时
	if agentCtx.Err() != nil {
		log.Debugf("Agent 调用上下文已取消或超时: 故障点A ID=%d, 故障点B ID=%d", fpA.FaultID, fpB.FaultID)
		return []domain.CausalCandidate{}
	}

	// 提取与当前这对故障点相关的拓扑关系（只保留两个实体之间的直接关系）
	relevantTopologySubgraph := s.extractRelevantTopology(fpA, fpB, recallCtx)

	// 构建 Agent 请求（带 token 长度控制）
	customQuerys, err := s.buildAgentCustomQuerysWithTokenLimit(fpA, fpB, relevantTopologySubgraph)
	if err != nil {
		log.Warnf("构建 Agent 请求失败（token 超限）: 故障点A ID=%d, 故障点B ID=%d, 错误: %v, 使用精简版本",
			fpA.FaultID, fpB.FaultID, err)
		// Token 超限时，使用精简版本
		customQuerys = s.buildAgentCustomQuerysMinimal(fpA, fpB)
	}

	// 验证配置
	causalConfig, err := s.buildCausalConfig()
	if err != nil {
		log.Warnf("Agent 配置无效: 故障点A ID=%d, 故障点B ID=%d, 错误: %v",
			fpA.FaultID, fpB.FaultID, err)
		return []domain.CausalCandidate{}
	}

	// 调用 Agent 进行因果推理（使用 agentCtx 以确保超时控制生效）
	edges, err := s.dipClient.CallCausalAgent(agentCtx, causalConfig, customQuerys)

	// 处理 Agent 调用结果（handleAgentCausalResult 会处理 err 和空结果的情况）
	// 失败直接返回空列表，由调用方使用本地规则
	return s.handleAgentCausalResult(agentCtx, edges, err, fpA, fpB)
}

// extractRelevantTopology 提取与当前这对故障点相关的拓扑关系
// 辅助函数，简化主函数逻辑
func (s *Service) extractRelevantTopology(fpA, fpB *domain.FaultPointObject, recallCtx *domain.GraphRecallContext) map[string]*domain.Topology {
	if recallCtx == nil || recallCtx.TopologySubgraphs == nil {
		return make(map[string]*domain.Topology)
	}
	return s.extractRelevantTopologySubgraphOptimized(fpA, fpB, recallCtx.TopologySubgraphs)
}

// buildCausalConfig 从配置构建 CausalConfig，并验证配置有效性
func (s *Service) buildCausalConfig() (dip.CausalConfig, error) {
	config := dip.CausalConfig{
		AppID:         s.config.Platform.Agents.CausalAnalysis.AppID,
		AgentKey:      s.config.Platform.Agents.CausalAnalysis.AgentKey,
		Authorization: s.config.AppConfig.Credentials.Authorization,
	}

	// 验证配置
	if config.AppID == "" {
		return config, fmt.Errorf("AppID 不能为空")
	}
	if config.AgentKey == "" {
		return config, fmt.Errorf("AgentKey 不能为空")
	}
	if config.Authorization == "" {
		return config, fmt.Errorf("Authorization 不能为空")
	}

	return config, nil
}

// 提取与当前这对故障点相关的拓扑子图信息
func (s *Service) extractRelevantTopologySubgraphOptimized(fpA, fpB *domain.FaultPointObject, allTopologySubgraphs map[string]*domain.Topology) map[string]*domain.Topology {
	relevant := make(map[string]*domain.Topology)

	// 参数验证
	if fpA == nil || fpB == nil {
		return relevant
	}

	// 如果 allTopologySubgraphs 为 nil，返回空 map
	if allTopologySubgraphs == nil {
		return relevant
	}

	// 获取故障点对应的实体对象ID
	entityA := fpA.EntityObjectID
	entityB := fpB.EntityObjectID

	// 验证 EntityObjectID 不为空
	if entityA == "" || entityB == "" {
		return relevant
	}

	// 只提取 entityA 和 entityB 之间的直接关系（边）
	// 用于去重边（避免多个拓扑子图中有相同的边）
	edgeSet := make(map[string]bool)
	relevantEdges := make([]domain.Relation, 0)

	// 遍历所有拓扑子图，查找与 entityA 和 entityB 相关的边
	for _, topology := range allTopologySubgraphs {
		if topology == nil || len(topology.Edges) == 0 {
			continue
		}

		for _, edge := range topology.Edges {
			// 只保留与 entityA 和 entityB 直接相关的边
			// 即边的源或目标是 entityA 或 entityB
			isRelevant := (edge.SourceSID == entityA && edge.TargetSID == entityB) ||
				(edge.SourceSID == entityB && edge.TargetSID == entityA) ||
				(edge.SourceSID == entityA || edge.TargetSID == entityA) ||
				(edge.SourceSID == entityB || edge.TargetSID == entityB)

			if isRelevant {
				// 生成边的唯一标识用于去重
				edgeKey := fmt.Sprintf("%s->%s:%s", edge.SourceSID, edge.TargetSID, edge.RelationID)

				// 如果边已存在，跳过（去重）
				if edgeSet[edgeKey] {
					continue
				}

				edgeSet[edgeKey] = true
				relevantEdges = append(relevantEdges, domain.Relation{
					RelationID:    edge.RelationID,
					RelationClass: edge.RelationClass,
					SourceSID:     edge.SourceSID,
					TargetSID:     edge.TargetSID,
				})
			}
		}
	}

	// 如果有相关边，创建一个只包含边的拓扑结构
	if len(relevantEdges) > 0 {
		// 使用一个固定的 key，因为只返回关系信息，不需要按子图分组
		relevant["relations"] = &domain.Topology{
			Nodes: []domain.Node{}, // 不保留节点信息，只保留关系
			Edges: relevantEdges,   // 只保留与两个实体相关的边
		}
	}

	return relevant
}

// 估算 JSON 数据的 token 数量
func (s *Service) estimateTokenCount(data interface{}) int {
	// 简单估算：将数据序列化为 JSON 字符串，然后估算 token 数
	// 实际可以使用更精确的 tokenizer，但这里使用简单估算
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		// 如果序列化失败，返回一个较大的估算值
		return maxInputTokens + 1
	}
	// 每个字符大约对应 4 个 token（中文和英文混合）
	return len(jsonBytes) * estimatedTokenSize
}

// 构建 Agent 请求（带 token 长度限制）
func (s *Service) buildAgentCustomQuerysWithTokenLimit(fpA, fpB *domain.FaultPointObject, relevantTopologySubgraph map[string]*domain.Topology) (map[string]interface{}, error) {
	// 参数验证
	if fpA == nil || fpB == nil {
		return map[string]interface{}{}, nil
	}

	// 构建故障点A和B的精简信息负载
	faultPointAPayload := s.buildAgentFaultPointPayloadOptimized(fpA)
	faultPointBPayload := s.buildAgentFaultPointPayloadOptimized(fpB)

	// 逐步添加拓扑信息，检查 token 长度
	topologyRelation := map[string]interface{}{
		agentRequestFieldEntityAID:        fpA.EntityObjectID,
		agentRequestFieldEntityBID:        fpB.EntityObjectID,
		agentRequestFieldTopologySubgraph: make(map[string]*domain.Topology),
	}

	// 精简拓扑关系：如果拓扑关系存在，只保留关系（边）信息，不保留节点详细信息
	if len(relevantTopologySubgraph) > 0 {
		topologyRelation = s.simplifyTopologyRelation(topologyRelation, relevantTopologySubgraph, fpA.EntityObjectID, fpB.EntityObjectID)
	}

	// 最终检查
	finalPayload := map[string]interface{}{
		agentRequestFieldFaultPointA:      faultPointAPayload,
		agentRequestFieldFaultPointB:      faultPointBPayload,
		agentRequestFieldTopologyRelation: topologyRelation,
	}
	finalTokenCount := s.estimateTokenCount(finalPayload)

	if finalTokenCount > maxInputTokens {
		return nil, fmt.Errorf("token 数量超限: %d > %d ", finalTokenCount, maxInputTokens)
	}

	log.Debugf("Agent 请求 token 估算: %d/%d", finalTokenCount, maxInputTokens)
	return finalPayload, nil
}

// simplifyTopologyRelation 精简拓扑关系，只保留关系（边）信息
// 从拓扑子图中提取关系信息，只保留边，不保留节点的详细信息，以减少 token 消耗
func (s *Service) simplifyTopologyRelation(topologyRelation map[string]interface{}, relevantTopologySubgraph map[string]*domain.Topology, entityA, entityB string) map[string]interface{} {
	if topologyRelation == nil {
		return topologyRelation
	}

	// 如果拓扑子图为空，直接返回精简后的结构（只保留实体ID）
	if len(relevantTopologySubgraph) == 0 {
		return map[string]interface{}{
			agentRequestFieldEntityAID:        entityA,
			agentRequestFieldEntityBID:        entityB,
			agentRequestFieldTopologySubgraph: make(map[string]*domain.Topology),
		}
	}

	// 从拓扑子图中提取关系信息，只保留边，不保留节点详细信息
	relationOnlySubgraphs := make(map[string]*domain.Topology)

	for entityID, topology := range relevantTopologySubgraph {
		if topology == nil {
			continue
		}

		// 只保留关系（边），不保留节点详细信息
		relationOnly := &domain.Topology{
			Edges: make([]domain.Relation, 0),
		}

		// 只提取与 entityA 和 entityB 相关的边
		for _, edge := range topology.Edges {
			// 只保留与 entityA 或 entityB 相关的边
			if edge.SourceSID == entityA || edge.SourceSID == entityB ||
				edge.TargetSID == entityA || edge.TargetSID == entityB {
				relationOnly.Edges = append(relationOnly.Edges, domain.Relation{
					RelationID:    edge.RelationID,
					RelationClass: edge.RelationClass,
					SourceSID:     edge.SourceSID,
					TargetSID:     edge.TargetSID,
				})
			}
		}

		// 只保留有关系的子图
		if len(relationOnly.Edges) > 0 {
			relationOnlySubgraphs[entityID] = relationOnly
		}
	}

	// 返回精简后的拓扑关系，只包含关系信息
	return map[string]interface{}{
		agentRequestFieldEntityAID:        entityA,
		agentRequestFieldEntityBID:        entityB,
		agentRequestFieldTopologySubgraph: relationOnlySubgraphs,
	}
}

// 构建最小化的 Agent 请求（当 token 超限时使用）
func (s *Service) buildAgentCustomQuerysMinimal(fpA, fpB *domain.FaultPointObject) map[string]interface{} {
	// 参数验证
	if fpA == nil || fpB == nil {
		return map[string]interface{}{}
	}

	// 只包含最核心的信息
	return map[string]interface{}{
		agentRequestFieldFaultPointA: map[string]interface{}{
			"fault_id":         fpA.FaultID,
			"fault_name":       fpA.FaultName,
			"fault_status":     string(fpA.FaultStatus),
			"fault_occur_time": fpA.FaultOccurTime.Format(time.RFC3339),
			"entity_object_id": fpA.EntityObjectID,
		},
		agentRequestFieldFaultPointB: map[string]interface{}{
			"fault_id":         fpB.FaultID,
			"fault_name":       fpB.FaultName,
			"fault_status":     string(fpB.FaultStatus),
			"fault_occur_time": fpB.FaultOccurTime.Format(time.RFC3339),
			"entity_object_id": fpB.EntityObjectID,
		},
		agentRequestFieldTopologyRelation: map[string]interface{}{
			agentRequestFieldEntityAID:        fpA.EntityObjectID,
			agentRequestFieldEntityBID:        fpB.EntityObjectID,
			agentRequestFieldTopologySubgraph: make(map[string]*domain.Topology),
		},
	}
}

// 构建优化的故障点负载（精简字段）
func (s *Service) buildAgentFaultPointPayloadOptimized(fp *domain.FaultPointObject) map[string]interface{} {
	if fp == nil {
		return map[string]interface{}{}
	}
	// 只包含关键字段，减少 token 消耗
	payload := map[string]interface{}{
		"fault_id":         fp.FaultID,
		"fault_name":       fp.FaultName,
		"fault_mode":       fp.FaultMode,
		"fault_level":      fp.FaultLevel,
		"fault_status":     string(fp.FaultStatus),
		"fault_occur_time": fp.FaultOccurTime.Format(time.RFC3339),
		"entity_object_id": fp.EntityObjectID,
	}
	// 可选字段：只在必要时添加
	if !fp.FaultRecoverTime.IsZero() {
		payload["fault_recovery_time"] = fp.FaultRecoverTime.Format(time.RFC3339)
	}
	if fp.FaultDurationTime > 0 {
		payload["fault_duration_time"] = fp.FaultDurationTime
	}
	return payload
}

// 处理 Agent 因果推理结果， 优化：改进错误处理和日志记录，失败直接返回空列表（不重试）
func (s *Service) handleAgentCausalResult(agentCtx context.Context, edges []domain.AgentCausalEdge, err error, fpA, fpB *domain.FaultPointObject) []domain.CausalCandidate {
	// 参数验证
	if fpA == nil || fpB == nil {
		return []domain.CausalCandidate{}
	}

	// 检查是否超时
	if agentCtx.Err() == context.DeadlineExceeded {
		log.Debugf("Agent 因果推理超时: 故障点A ID=%d (实体=%s), 故障点B ID=%d (实体=%s), 超时时间: %v",
			fpA.FaultID, fpA.EntityObjectID, fpB.FaultID, fpB.EntityObjectID, agentCallTimeout)
		return []domain.CausalCandidate{}
	}

	// 检查是否被取消
	if agentCtx.Err() == context.Canceled {
		log.Debugf("Agent 因果推理被取消: 故障点A ID=%d, 故障点B ID=%d", fpA.FaultID, fpB.FaultID)
		return []domain.CausalCandidate{}
	}

	// 检查调用错误
	if err != nil {
		log.Debugf("Agent 因果推理失败: 故障点A ID=%d (实体=%s), 故障点B ID=%d (实体=%s), 错误: %v",
			fpA.FaultID, fpA.EntityObjectID, fpB.FaultID, fpB.EntityObjectID, err)
		return []domain.CausalCandidate{}
	}

	// 将 Agent 返回的因果边映射为 CausalCandidate
	mappedCandidates := s.mapAgentEdgesToCandidates(edges, fpA, fpB)

	// 如果没有结果，返回空列表（调用方会使用 fallback）
	if len(mappedCandidates) == 0 {
		log.Debugf("Agent 因果推理无结果: 故障点A ID=%d, 故障点B ID=%d", fpA.FaultID, fpB.FaultID)
		return []domain.CausalCandidate{}
	}

	return mappedCandidates
}

// 将 Agent 返回的因果边映射为 CausalCandidate, 优化：添加更详细的验证和日志
func (s *Service) mapAgentEdgesToCandidates(edges []domain.AgentCausalEdge, fpA, fpB *domain.FaultPointObject) []domain.CausalCandidate {
	// 参数验证
	if fpA == nil || fpB == nil {
		return []domain.CausalCandidate{}
	}

	// 创建 FaultID 到 FaultPointObject 的映射
	// Agent 返回的 source_id 和 target_id 是 FaultID（uint64）
	idToFP := map[uint64]*domain.FaultPointObject{
		fpA.FaultID: fpA,
		fpB.FaultID: fpB,
	}

	var candidates []domain.CausalCandidate
	for _, e := range edges {
		// 验证 source_id 和 target_id 不为 0
		if e.Source == 0 || e.Target == 0 {
			log.Debugf("跳过无效的因果边: source=%d, target=%d", e.Source, e.Target)
			continue
		}

		// 使用 FaultID 查找对应的故障点
		causeFP := idToFP[e.Source]
		effectFP := idToFP[e.Target]
		if causeFP == nil || effectFP == nil {
			// 如果找不到对应的故障点，记录日志并跳过
			log.Debugf("无法找到对应的故障点: source_id=%d, target_id=%d, 可用故障点: fpA.FaultID=%d, fpB.FaultID=%d",
				e.Source, e.Target, fpA.FaultID, fpB.FaultID)
			continue
		}

		// 避免自引用（原因和结果不能是同一个故障点）
		if causeFP.FaultID == effectFP.FaultID {
			log.Debugf("跳过自引用的因果关系: fault_id=%d", causeFP.FaultID)
			continue
		}

		// 验证置信度范围
		confidence := e.Confidence
		if confidence < 0 {
			confidence = 0
		}
		if confidence > 1 {
			confidence = 1
		}

		candidates = append(candidates, domain.CausalCandidate{
			Cause:      causeFP,
			Effect:     effectFP,
			Confidence: confidence,
			Reason:     e.Reason,
			IsNew:      true,
		})
	}

	return candidates
}

// 计算两个故障点之间的因果候选, 优化：综合考虑时间、持续时间、状态、严重程度、历史关系等因素
func (s *Service) calculateCausalCandidate(fpA, fpB *domain.FaultPointObject, recallCtx *domain.GraphRecallContext) *domain.CausalCandidate {
	// 参数验证
	if fpA == nil || fpB == nil {
		return nil
	}

	// 确定因果关系方向（综合考虑多个因素）
	cause, effect, timeDiff := s.determineCausalDirection(fpA, fpB, recallCtx)
	if cause == nil || effect == nil {
		return nil
	}

	// 计算置信度和原因描述
	confidence, reasonParts := s.calculateConfidence(cause, effect, timeDiff, recallCtx)

	// 如果置信度过低，不返回候选
	if confidence < minConfidence {
		return nil
	}

	// 限制置信度范围
	confidence = s.clampConfidence(confidence)

	// 构建原因描述
	reason := s.buildReasonDescription(cause, effect, reasonParts)

	// 检查历史因果关系，判断是否为新关系
	isNew := s.isNewCausalRelation(cause.EntityObjectID, effect.EntityObjectID, recallCtx)

	return &domain.CausalCandidate{
		Cause:      cause,
		Effect:     effect,
		Confidence: confidence,
		Reason:     reason,
		IsNew:      isNew,
	}
}

// 确定因果关系方向,综合考虑时间、持续时间、状态、严重程度、历史关系等因素
func (s *Service) determineCausalDirection(fpA, fpB *domain.FaultPointObject, recallCtx *domain.GraphRecallContext) (*domain.FaultPointObject, *domain.FaultPointObject, time.Duration) {
	// 计算时间差
	timeDiffAB := fpB.FaultOccurTime.Sub(fpA.FaultOccurTime)
	timeDiffBA := fpA.FaultOccurTime.Sub(fpB.FaultOccurTime)

	// 如果时间差很小（可能是同时发生），使用其他因素判断
	absTimeDiff := timeDiffAB
	if absTimeDiff < 0 {
		absTimeDiff = -absTimeDiff
	}

	// 如果时间差很小（≤1分钟），主要依据其他因素判断
	if absTimeDiff <= 1*time.Minute {
		// 优先考虑：持续时间更长、未恢复、严重程度更高的作为原因
		scoreA := s.calculateDirectionScore(fpA, fpB)
		scoreB := s.calculateDirectionScore(fpB, fpA)

		// 检查历史关系
		historicalAB := s.hasHistoricalRelation(fpA.EntityObjectID, fpB.EntityObjectID, recallCtx)
		historicalBA := s.hasHistoricalRelation(fpB.EntityObjectID, fpA.EntityObjectID, recallCtx)

		if historicalAB && !historicalBA {
			return fpA, fpB, absTimeDiff
		}
		if historicalBA && !historicalAB {
			return fpB, fpA, absTimeDiff
		}

		// 如果历史关系相同，使用综合评分
		if scoreA > scoreB {
			return fpA, fpB, absTimeDiff
		} else if scoreB > scoreA {
			return fpB, fpA, absTimeDiff
		}
		// 如果评分相同，默认使用时间顺序
		if fpA.FaultOccurTime.Before(fpB.FaultOccurTime) {
			return fpA, fpB, absTimeDiff
		}
		return fpB, fpA, absTimeDiff
	}
	// 时间差较大，主要依据时间顺序，但也要考虑其他因素
	if timeDiffAB > 0 {
		// fpA 发生在 fpB 之前
		// 检查是否有反向的历史关系
		if s.hasHistoricalRelation(fpB.EntityObjectID, fpA.EntityObjectID, recallCtx) {
			// 有反向历史关系，需要进一步判断
			scoreA := s.calculateDirectionScore(fpA, fpB)
			scoreB := s.calculateDirectionScore(fpB, fpA)
			if scoreB > scoreA+0.1 { // 反向评分明显更高，使用反向
				return fpB, fpA, timeDiffBA
			}
		}
		return fpA, fpB, timeDiffAB
	} else {
		// fpB 发生在 fpA 之前
		// 检查是否有反向的历史关系
		if s.hasHistoricalRelation(fpA.EntityObjectID, fpB.EntityObjectID, recallCtx) {
			// 有反向历史关系，需要进一步判断
			scoreA := s.calculateDirectionScore(fpA, fpB)
			scoreB := s.calculateDirectionScore(fpB, fpA)
			if scoreA > scoreB+0.1 { // 反向评分明显更高，使用反向
				return fpA, fpB, timeDiffAB
			}
		}
		return fpB, fpA, timeDiffBA
	}
}

// 计算作为原因的方向评分，分数越高，越可能是原因
func (s *Service) calculateDirectionScore(candidate, other *domain.FaultPointObject) float64 {
	score := 0.0

	// 因素1: 持续时间更长（权重：0.3）
	if candidate.FaultDurationTime > other.FaultDurationTime {
		durationRatio := float64(candidate.FaultDurationTime) / float64(other.FaultDurationTime+1)
		if durationRatio > 2.0 {
			score += 0.3 // 持续时间明显更长
		} else {
			score += 0.15 * (durationRatio - 1.0) // 按比例增加
		}
	}

	// 因素2: 未恢复状态（权重：0.2）
	if candidate.FaultStatus == domain.FaultStatusOccurred && other.FaultStatus == domain.FaultStatusRecovered {
		score += 0.2
	} else if candidate.FaultStatus == domain.FaultStatusOccurred && other.FaultStatus == domain.FaultStatusOccurred {
		score += 0.1 // 都未恢复，略微加分
	}

	// 因素3: 严重程度更高（权重：0.2）
	// FaultLevel 值越小表示越严重
	if candidate.FaultLevel < other.FaultLevel {
		levelDiff := float64(other.FaultLevel - candidate.FaultLevel)
		score += 0.2 * (levelDiff / 5.0) // 归一化到 0-0.2
	}

	// 因素4: 发生时间更早（权重：0.1）
	if candidate.FaultOccurTime.Before(other.FaultOccurTime) {
		score += 0.1
	}

	// 因素5: 有恢复时间但恢复时间更晚（权重：0.1）
	if !candidate.FaultRecoverTime.IsZero() && !other.FaultRecoverTime.IsZero() {
		if candidate.FaultRecoverTime.After(other.FaultRecoverTime) {
			score += 0.1
		}
	} else if !candidate.FaultRecoverTime.IsZero() && other.FaultRecoverTime.IsZero() {
		// 候选有恢复时间，但对方没有，说明候选可能不是根因
		score -= 0.1
	}

	return score
}

// 检查是否存在历史因果关系
func (s *Service) hasHistoricalRelation(causeObjectID, effectObjectID string, recallCtx *domain.GraphRecallContext) bool {
	if recallCtx == nil || recallCtx.HistoricalCausality == nil {
		return false
	}
	if causeObjectID == "" || effectObjectID == "" {
		return false
	}

	historical, ok := recallCtx.HistoricalCausality[causeObjectID]
	if !ok {
		return false
	}

	for _, h := range historical {
		if h.EffectObjectID == effectObjectID {
			return true
		}
	}
	return false
}

// 计算置信度和原因描述,优化：更精细的置信度计算逻辑
func (s *Service) calculateConfidence(cause, effect *domain.FaultPointObject, timeDiff time.Duration, recallCtx *domain.GraphRecallContext) (float64, []string) {
	// 参数验证
	if cause == nil || effect == nil {
		return 0, nil
	}

	confidence := baseConfidence
	var reasonParts []string

	// 因素1: 时间间隔（间隔越小，置信度越高）
	timeConfidence, timeReason := s.calculateTimeIntervalConfidence(cause, effect, timeDiff)
	confidence += timeConfidence
	if timeReason != "" {
		reasonParts = append(reasonParts, timeReason)
	}

	// 因素2: 故障持续时间（持续时间长的更可能是根因）
	durationConfidence, durationReason := s.calculateDurationConfidence(cause, effect)
	confidence += durationConfidence
	if durationReason != "" {
		reasonParts = append(reasonParts, durationReason)
	}

	// 因素3: 故障状态（已恢复的故障可能不是根因）
	statusConfidence, statusReason := s.calculateStatusConfidence(cause, effect)
	confidence += statusConfidence
	if statusReason != "" {
		reasonParts = append(reasonParts, statusReason)
	}

	// 因素4: 故障严重程度（严重程度高的更可能是根因）
	severityConfidence, severityReason := s.calculateSeverityConfidence(cause, effect)
	confidence += severityConfidence
	if severityReason != "" {
		reasonParts = append(reasonParts, severityReason)
	}

	// 因素5: 历史因果关系（历史数据支持提高置信度）
	historicalConfidence, historicalReason := s.calculateHistoricalConfidence(cause, effect, recallCtx)
	confidence += historicalConfidence
	if historicalReason != "" {
		reasonParts = append(reasonParts, historicalReason)
	}

	return confidence, reasonParts
}

// 计算时间间隔相关的置信度，优化：更精细的时间间隔处理
func (s *Service) calculateTimeIntervalConfidence(cause, effect *domain.FaultPointObject, timeDiff time.Duration) (float64, string) {
	// 参数验证
	if cause == nil || effect == nil {
		return 0, ""
	}

	roundedTimeDiff := timeDiff.Round(time.Second)
	absTimeDiff := timeDiff
	if absTimeDiff < 0 {
		absTimeDiff = -absTimeDiff
	}

	// 处理边界情况：如果时间差为 0 或非常小（≤1分钟），可能是同时发生
	if absTimeDiff <= 1*time.Minute {
		// 同时发生的故障，置信度较低，主要依赖其他因素
		return confidenceVeryShortTime * 0.5, fmt.Sprintf("故障点ID：%d 和 故障点ID：%d 几乎同时发生（间隔 %v）", cause.FaultID, effect.FaultID, roundedTimeDiff)
	}

	if absTimeDiff <= veryShortTimeThreshold {
		return confidenceVeryShortTime, fmt.Sprintf("故障点ID：%d 发生在 故障点ID：%d 之前（间隔 %v，时间间隔很短）", cause.FaultID, effect.FaultID, roundedTimeDiff)
	} else if absTimeDiff <= shortTimeThreshold {
		return confidenceShortTime, fmt.Sprintf("故障点ID：%d 发生在 故障点ID：%d 之前（间隔 %v，时间间隔较短）", cause.FaultID, effect.FaultID, roundedTimeDiff)
	} else if absTimeDiff <= longTimeThreshold {
		return confidenceLongTime, fmt.Sprintf("故障点ID：%d 发生在 故障点ID：%d 之前（间隔 %v）", cause.FaultID, effect.FaultID, roundedTimeDiff)
	} else {
		// 时间间隔过大，降低置信度
		return confidenceVeryLongTime, fmt.Sprintf("故障点ID：%d 发生在 故障点ID：%d 之前（间隔 %v，时间间隔较大）", cause.FaultID, effect.FaultID, roundedTimeDiff)
	}
}

// 计算故障持续时间相关的置信度,优化：更精细的持续时间比较逻辑
func (s *Service) calculateDurationConfidence(cause, effect *domain.FaultPointObject) (float64, string) {
	// 参数验证
	if cause == nil || effect == nil {
		return 0, ""
	}

	// 如果原因故障持续时间明显更长（>2倍），增加置信度
	if cause.FaultDurationTime > 0 && effect.FaultDurationTime > 0 {
		durationRatio := float64(cause.FaultDurationTime) / float64(effect.FaultDurationTime)
		if durationRatio > 2.0 {
			return confidenceLongerDuration * 1.5, fmt.Sprintf("原因故障持续时间明显更长（%.1f倍）", durationRatio)
		} else if durationRatio > 1.5 {
			return confidenceLongerDuration, "原因故障持续时间更长"
		} else if durationRatio < 0.5 {
			// 原因故障持续时间明显更短，降低置信度
			return -confidenceLongerDuration, "原因故障持续时间明显更短"
		}
	}

	// 如果只有原因故障有持续时间，而结果故障没有
	if cause.FaultDurationTime > 0 && effect.FaultDurationTime == 0 {
		return confidenceNormalDuration, "原因故障有持续时间"
	}

	// 如果原因故障没有持续时间，但结果故障有，降低置信度
	if cause.FaultDurationTime == 0 && effect.FaultDurationTime > 0 {
		return -confidenceNormalDuration, "原因故障无持续时间"
	}

	// 如果持续时间相同或都无持续时间，返回 0
	return 0, ""
}

// 计算故障状态相关的置信度,优化：考虑更多状态组合
func (s *Service) calculateStatusConfidence(cause, effect *domain.FaultPointObject) (float64, string) {
	// 参数验证
	if cause == nil || effect == nil {
		return 0, ""
	}

	// 情况1: 原因已恢复但结果未恢复（降低置信度）
	if cause.FaultStatus == domain.FaultStatusRecovered && effect.FaultStatus != domain.FaultStatusRecovered {
		return confidenceRecoveredStatus, "原因故障已恢复但结果故障未恢复"
	}

	// 情况2: 原因未恢复但结果已恢复（可能不合理，降低置信度）
	if cause.FaultStatus != domain.FaultStatusRecovered && effect.FaultStatus == domain.FaultStatusRecovered {
		return -0.15, "原因故障未恢复但结果故障已恢复（可能不合理）"
	}

	// 情况3: 都未恢复（略微增加置信度）
	if cause.FaultStatus == domain.FaultStatusOccurred && effect.FaultStatus == domain.FaultStatusOccurred {
		return 0.05, "两个故障都未恢复"
	}

	// 情况4: 都已恢复（略微降低置信度，因为可能不是真正的因果关系）
	if cause.FaultStatus == domain.FaultStatusRecovered && effect.FaultStatus == domain.FaultStatusRecovered {
		return -0.05, "两个故障都已恢复"
	}

	return 0, ""
}

// 计算故障严重程度相关的置信度,新增：考虑故障严重程度
func (s *Service) calculateSeverityConfidence(cause, effect *domain.FaultPointObject) (float64, string) {
	// 参数验证
	if cause == nil || effect == nil {
		return 0, ""
	}

	// FaultLevel 值越小表示越严重
	// 如果原因故障更严重，增加置信度
	if cause.FaultLevel < effect.FaultLevel {
		severityDiff := effect.FaultLevel - cause.FaultLevel
		confidence := 0.1 * float64(severityDiff) / 5.0 // 归一化到 0-0.1
		if severityDiff >= 2 {
			return confidence, fmt.Sprintf("原因故障严重程度更高（严重程度差：%d）", severityDiff)
		}
		return confidence, "原因故障严重程度更高"
	}

	// 如果结果故障更严重，略微降低置信度
	if cause.FaultLevel > effect.FaultLevel {
		severityDiff := cause.FaultLevel - effect.FaultLevel
		if severityDiff >= 2 {
			return -0.05, "结果故障严重程度明显更高"
		}
		return -0.02, "结果故障严重程度更高"
	}

	return 0, ""
}

// 计算历史因果关系相关的置信度,优化：检查双向历史关系
func (s *Service) calculateHistoricalConfidence(cause, effect *domain.FaultPointObject, recallCtx *domain.GraphRecallContext) (float64, string) {
	// 参数验证
	if cause == nil || effect == nil {
		return 0, ""
	}
	if recallCtx == nil {
		return 0, ""
	}
	if recallCtx.HistoricalCausality == nil {
		return 0, ""
	}
	if cause.EntityObjectID == "" || effect.EntityObjectID == "" {
		return 0, ""
	}

	// 检查正向历史关系（cause -> effect）
	historical, ok := recallCtx.HistoricalCausality[cause.EntityObjectID]
	if ok {
		for _, h := range historical {
			if h.EffectObjectID == effect.EntityObjectID {
				// 历史上存在相同因果关系
				historicalBoost := float64(h.OccurrenceCount) * confidencePerHistoricalOccurrence
				if historicalBoost > maxHistoricalBoost {
					historicalBoost = maxHistoricalBoost
				}
				return historicalBoost, fmt.Sprintf("历史因果关系支持（出现%d次）", h.OccurrenceCount)
			}
		}
	}

	// 检查反向历史关系（effect -> cause），如果存在，降低置信度
	reverseHistorical, ok := recallCtx.HistoricalCausality[effect.EntityObjectID]
	if ok {
		for _, h := range reverseHistorical {
			if h.EffectObjectID == cause.EntityObjectID {
				// 历史上存在反向关系，降低置信度
				return -0.1, fmt.Sprintf("历史数据支持反向关系（出现%d次）", h.OccurrenceCount)
			}
		}
	}

	return 0, ""
}

// 限制置信度范围
func (s *Service) clampConfidence(confidence float64) float64 {
	if confidence < minConfidence {
		return minConfidence
	}
	if confidence > maxConfidence {
		return maxConfidence
	}
	return confidence
}

// 构建原因描述
func (s *Service) buildReasonDescription(cause, effect *domain.FaultPointObject, reasonParts []string) string {
	// 参数验证
	if cause == nil || effect == nil {
		return ""
	}

	if len(reasonParts) == 0 {
		return fmt.Sprintf("故障点ID：%d → 故障点ID：%d: 存在拓扑关联", cause.FaultID, effect.FaultID)
	}

	// 使用分号连接多个原因
	reasonDetail := ""
	for i, part := range reasonParts {
		if i > 0 {
			reasonDetail += "; "
		}
		reasonDetail += part
	}
	return fmt.Sprintf("故障点ID：%d → 故障点ID：%d: %s", cause.FaultID, effect.FaultID, reasonDetail)
}

// 检查是否为新的因果关系,优化：检查双向关系
func (s *Service) isNewCausalRelation(causeObjectID, effectObjectID string, recallCtx *domain.GraphRecallContext) bool {
	if recallCtx == nil {
		return true
	}
	if recallCtx.HistoricalCausality == nil {
		return true
	}
	if causeObjectID == "" || effectObjectID == "" {
		return true
	}

	// 检查正向关系
	historical, ok := recallCtx.HistoricalCausality[causeObjectID]
	if ok {
		for _, h := range historical {
			if h.EffectObjectID == effectObjectID {
				return false
			}
		}
	}

	return true
}

// ========== 辅助方法 ==========

// hasTopologyRelation 检查两个实体是否有拓扑关联
// 优化：改进逻辑判断和性能
func (s *Service) hasTopologyRelation(entityA, entityB string, recallCtx *domain.GraphRecallContext) bool {
	if recallCtx == nil || recallCtx.TopologySubgraphs == nil {
		return false
	}

	// 验证 EntityObjectID 不为空
	if entityA == "" || entityB == "" {
		return false
	}

	// 如果两个实体ID相同，认为有关联（自关联）
	if entityA == entityB {
		return true
	}

	// 检查对象子图中是否有直接关系
	// 首先检查 entityA 和 entityB 是否作为 key 存在于 TopologySubgraphs 中
	topologyA, hasA := recallCtx.TopologySubgraphs[entityA]
	topologyB, hasB := recallCtx.TopologySubgraphs[entityB]

	// 如果两个实体都在 TopologySubgraphs 中（通过 key 匹配）
	if hasA && hasB {
		// 检查它们是否在同一个子图中（通过比较节点集合或指针）
		// 优化：使用辅助函数判断是否为同一子图
		if s.isSameTopology(topologyA, topologyB) {
			return s.checkDirectEdgeConnection(entityA, entityB, topologyA)
		}

		// 如果它们在不同的子图中，检查两个子图是否有交集（共享节点或边）
		// 这种情况可能发生在实体属于不同的子图但通过其他节点连接
		return s.checkCrossSubgraphConnection(topologyA, topologyB)
	}

	// 如果只有一个实体在 TopologySubgraphs 中，遍历所有子图查找
	// 这种情况可能发生在实体ID与存储的key不完全匹配
	if hasA || hasB {
		return s.searchInAllSubgraphs(entityA, entityB, recallCtx.TopologySubgraphs)
	}

	// 如果两个实体都不在 TopologySubgraphs 的 key 中，遍历所有子图查找
	return s.searchInAllSubgraphs(entityA, entityB, recallCtx.TopologySubgraphs)
}

// isSameTopology 判断两个拓扑是否为同一个（通过比较内容）
// 优化：避免指针比较的不准确性
func (s *Service) isSameTopology(topologyA, topologyB *domain.Topology) bool {
	// 如果都是 nil，认为是同一个
	if topologyA == nil && topologyB == nil {
		return true
	}
	// 如果一个是 nil，另一个不是，不是同一个
	if topologyA == nil || topologyB == nil {
		return false
	}
	// 指针比较（如果确实是同一个对象）
	if topologyA == topologyB {
		return true
	}
	// 内容比较：检查节点数量是否相同（简单判断，避免完整比较的性能开销）
	// 如果节点数量差异很大，肯定不是同一个
	if len(topologyA.Nodes) != len(topologyB.Nodes) || len(topologyA.Edges) != len(topologyB.Edges) {
		return false
	}
	// 如果节点和边数量都相同，且指针不同，可能是内容相同的不同对象
	// 为了性能，这里假设如果节点和边数量相同，且指针不同，认为是不同的子图
	// 如果需要精确判断，可以进一步比较内容，但会带来性能开销
	return false
}

// checkDirectEdgeConnection 检查两个实体在同一个子图中是否有直接边连接
// 优化：先检查边，避免不必要的节点集合构建
func (s *Service) checkDirectEdgeConnection(entityA, entityB string, topology *domain.Topology) bool {
	if topology == nil {
		return false
	}

	// 优化：先检查边，如果找到直接连接就返回，避免构建节点集合
	// 检查是否有边连接这两个实体（双向检查）
	for _, edge := range topology.Edges {
		if (edge.SourceSID == entityA && edge.TargetSID == entityB) ||
			(edge.SourceSID == entityB && edge.TargetSID == entityA) {
			return true
		}
	}

	// 如果没有直接边，检查两个节点是否都在子图中（可能通过其他节点间接连接）
	// 这里只做简单检查，因为调用方已经确认它们在同一个子图中
	// 如果两个节点都在子图中，即使没有直接边，也认为有关联（间接关联）
	nodeSet := s.buildNodeSet(topology.Nodes)
	return nodeSet[entityA] && nodeSet[entityB]
}

// buildNodeSet 构建节点集合（辅助函数，减少代码重复）
func (s *Service) buildNodeSet(nodes []domain.Node) map[string]bool {
	nodeSet := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if node.SID != "" {
			nodeSet[node.SID] = true
		}
	}
	return nodeSet
}

// checkCrossSubgraphConnection 检查两个实体在不同子图中是否有连接
// 优化：改进性能和逻辑
func (s *Service) checkCrossSubgraphConnection(topologyA, topologyB *domain.Topology) bool {
	if topologyA == nil || topologyB == nil {
		return false
	}

	// 优化：使用辅助函数构建节点集合
	nodeSetA := s.buildNodeSet(topologyA.Nodes)
	nodeSetB := s.buildNodeSet(topologyB.Nodes)

	// 检查两个子图是否有共享节点（交集）
	// 优化：遍历较小的集合以提高性能
	if len(nodeSetA) <= len(nodeSetB) {
		for sid := range nodeSetA {
			if nodeSetB[sid] {
				// 有共享节点，说明两个实体通过共享节点连接
				return true
			}
		}
	} else {
		for sid := range nodeSetB {
			if nodeSetA[sid] {
				// 有共享节点，说明两个实体通过共享节点连接
				return true
			}
		}
	}

	// 检查两个子图的边是否有连接（通过边的源或目标节点）
	// 如果 topologyA 的边的目标节点在 topologyB 中，或反之，说明有连接
	for _, edge := range topologyA.Edges {
		// 检查边的源节点和目标节点是否在另一个子图中
		if nodeSetB[edge.SourceSID] || nodeSetB[edge.TargetSID] {
			return true
		}
	}

	for _, edge := range topologyB.Edges {
		// 检查边的源节点和目标节点是否在另一个子图中
		if nodeSetA[edge.SourceSID] || nodeSetA[edge.TargetSID] {
			return true
		}
	}

	return false
}

// searchInAllSubgraphs 在所有子图中搜索两个实体的连接关系
// 优化：改进性能和逻辑，支持更早退出
func (s *Service) searchInAllSubgraphs(entityA, entityB string, subgraphs map[string]*domain.Topology) bool {
	if len(subgraphs) == 0 {
		return false
	}

	// 遍历所有对象子图，查找包含这两个实体的子图，并检查它们之间是否有边连接
	for key, topology := range subgraphs {
		if topology == nil {
			continue
		}

		// 优化：先快速检查边，如果找到直接连接就返回
		for _, edge := range topology.Edges {
			if (edge.SourceSID == entityA && edge.TargetSID == entityB) ||
				(edge.SourceSID == entityB && edge.TargetSID == entityA) {
				log.Debugf("在子图 key=%s 中找到实体 %s 和 %s 的直接连接", key, entityA, entityB)
				return true
			}
		}

		// 如果没有直接边，检查两个实体是否都在这个子图中
		// 优化：使用辅助函数构建节点集合
		nodeSet := s.buildNodeSet(topology.Nodes)
		hasA := nodeSet[entityA]
		hasB := nodeSet[entityB]

		// 如果两个实体都在这个子图中，即使没有直接边，也认为有关联（间接关联）
		if hasA && hasB {
			log.Debugf("在子图 key=%s 中找到实体 %s 和 %s（间接关联）", key, entityA, entityB)
			return true
		}
	}

	return false
}

// faultPointToFault 将 FaultPoint 转换为 Fault
func (s *Service) faultPointToFault(fp domain.FaultPointObject) domain.Fault {
	// 如果传入零值，会返回零值的 Fault，这是合理的
	return domain.Fault{
		FaultID:           fp.FaultID,
		FaultName:         fp.FaultName,
		FaultMode:         fp.FaultMode,
		FaultStatus:       fp.FaultStatus,
		FaultDescription:  fp.FaultDescription,
		FaultOccurTime:    fp.FaultOccurTime,
		FaultLatestTime:   fp.FaultLatestTime,
		FaultDurationTime: fp.FaultDurationTime,
		FaultCreateTime:   fp.FaultCreateTime,
		FaultUpdateTime:   fp.FaultUpdateTime,
		FaultRecoverTime:  fp.FaultRecoverTime,
		FaultLevel:        fp.FaultLevel,
		EntityObjectClass: fp.EntityObjectClass,
		EntityObjectName:  fp.EntityObjectName,
		EntityObjectID:    fp.EntityObjectID,
		RelationEventIDs:  fp.RelationEventIDs,
	}
}
