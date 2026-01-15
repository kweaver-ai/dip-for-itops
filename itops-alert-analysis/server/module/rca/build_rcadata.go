package rca

import (
	"context"
	"fmt"
	"sort"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/dip"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"github.com/pkg/errors"
)

// -----1.构建分析上下文-----
func (s *Service) buildRcaContext(ctx context.Context, problemName string, faultPointInfos []domain.FaultPointObject, recallCtx *domain.GraphRecallContext, result *domain.CausalAnalysisResults) (domain.RcaContext, error) {
	RcaContext := domain.RcaContext{
		BackTrace: []domain.Fault{},
	}

	// 1. 生成问题现象描述（通过 Agent 生成或使用默认值）
	if err := s.buildOccurrenceDescription(ctx, &RcaContext, faultPointInfos); err != nil {
		// 生成描述失败不影响整体流程，使用默认值
		log.Warnf("生成问题现象描述失败: %v，使用默认描述", err)
		RcaContext.Occurrence = s.buildDefaultOccurrenceDescription(problemName, faultPointInfos)
	}

	// 2. 构建故障追溯，将故障点信息添加到故障追溯中，按照故障点的发生时间排序
	s.buildFaultTrace(&RcaContext, faultPointInfos)

	// 3. 构建分析网络
	RcaNetwork, err := s.GenerateRcaNetwork(ctx, faultPointInfos, recallCtx, result)
	if err != nil {
		return RcaContext, errors.Wrapf(err, "生成分析网络失败")
	}

	// 将分析网络直接赋值给 Network 字段
	// Network 字段类型为 RcaNetwork，不需要序列化
	RcaContext.Network = *RcaNetwork

	return RcaContext, nil
}

// ----- 2. 构建故障追溯 -----
// 将故障点信息添加到故障追溯中，按照故障点的发生时间排序
// 优化：改进排序稳定性
func (s *Service) buildFaultTrace(RcaContext *domain.RcaContext, faultPointInfos []domain.FaultPointObject) {
	if len(faultPointInfos) == 0 {
		RcaContext.BackTrace = []domain.Fault{}
		return
	}

	// 先排序故障点（按发生时间升序）
	sortedFaultPoints := make([]domain.FaultPointObject, len(faultPointInfos))
	copy(sortedFaultPoints, faultPointInfos)
	// 优化：如果时间相同，按故障ID排序保证稳定性
	sort.Slice(sortedFaultPoints, func(i, j int) bool {
		if sortedFaultPoints[i].FaultOccurTime.Before(sortedFaultPoints[j].FaultOccurTime) {
			return true
		}
		if sortedFaultPoints[i].FaultOccurTime.After(sortedFaultPoints[j].FaultOccurTime) {
			return false
		}
		// 时间相同，按故障ID排序
		return sortedFaultPoints[i].FaultID < sortedFaultPoints[j].FaultID
	})

	// 转换为 Fault 并添加到 FaultTrace
	faults := make([]domain.Fault, 0, len(sortedFaultPoints))
	for _, fp := range sortedFaultPoints {
		faults = append(faults, s.faultPointToFault(fp))
	}
	RcaContext.BackTrace = faults
}

// buildOccurrenceDescription 构建问题现象描述
// 优化：添加 fallback 机制，当 token 超限时使用最小化版本
func (s *Service) buildOccurrenceDescription(ctx context.Context, RcaContext *domain.RcaContext, faultPoints []domain.FaultPointObject) error {
	// 构建发送给 Agent 的自定义查询参数（带 token 长度控制）
	customQuerys, err := s.buildAgentCustomQuerysForDescriptionWithTokenLimit(faultPoints)
	if err != nil {
		// Token 超限时，使用最小化版本作为 fallback
		log.Warnf("构建 Agent 请求失败（token 超限）: %v，使用最小化版本", err)
		customQuerys = s.buildAgentCustomQuerys(faultPoints)
	}

	// 将 AgentConfig 转换为 SummaryConfig
	summaryConfig := dip.SummaryConfig{
		AppID:         s.config.Platform.Agents.ProblemSummary.AppID,
		AgentKey:      s.config.Platform.Agents.ProblemSummary.AgentKey,
		Authorization: s.config.AppConfig.Credentials.Authorization,
	}

	// 调用 Agent 生成描述（设置超时上下文）
	agentCtx, cancel := context.WithTimeout(ctx, agentCallTimeout)
	defer cancel()

	payload, err := s.dipClient.CallSummaryAgent(agentCtx, summaryConfig, customQuerys)
	if err != nil {
		// 检查是否超时
		if agentCtx.Err() == context.DeadlineExceeded {
			return errors.Wrapf(err, "调用 Agent 生成描述超时")
		}
		return errors.Wrapf(err, "调用 Agent 生成描述失败")
	}

	// 验证返回结果（payload 是值类型，不可能是 nil，需要检查内容是否有效）
	if payload.Occurrence.Name == "" && payload.Occurrence.Description == "" && payload.Occurrence.Impact == "" {
		// 所有字段都为空，视为无效结果
		return errors.New("Agent 返回空结果（所有字段为空）")
	}

	// 设置到 RcaContext.Occurrence
	RcaContext.Occurrence = payload.Occurrence
	return nil
}

// buildAgentCustomQuerysForDescriptionWithTokenLimit 构建 Agent 描述生成的自定义查询参数（带 token 长度限制）
// 优化：控制 token 长度，避免超出大模型限制
func (s *Service) buildAgentCustomQuerysForDescriptionWithTokenLimit(faultPoints []domain.FaultPointObject) (map[string]interface{}, error) {
	// 步骤1：按优先级排序故障点（优先发送重要的故障点）
	sortedFaultPoints := s.sortFaultPointsByPriority(faultPoints)

	// 步骤2：逐步添加故障点，直到接近 token 限制
	faultPointsPayload := make([]map[string]interface{}, 0)

	for _, fp := range sortedFaultPoints {
		// 构建测试负载（优化：避免重复构建）
		testPayload := make([]map[string]interface{}, len(faultPointsPayload), len(faultPointsPayload)+1)
		copy(testPayload, faultPointsPayload)

		// 构建当前故障点的负载（只构建一次）
		fpPayload := s.buildAgentPayloadOptimized(&fp)
		testPayload = append(testPayload, fpPayload)

		testCustomQuerys := map[string]interface{}{
			agentCustomQueryKeyProblemInfo: testPayload,
		}
		testTokenCount := s.estimateTokenCount(testCustomQuerys)

		if testTokenCount <= maxInputTokens {
			// Token 未超限，添加故障点
			faultPointsPayload = append(faultPointsPayload, fpPayload)
		} else {
			log.Debugf("故障点 token 超限，跳过故障点 ID=%d (当前: %d, 限制: %d)",
				fp.FaultID, testTokenCount, maxInputTokens)
			break
		}

		// 限制最大故障点数量（双重保护）
		if len(faultPointsPayload) >= maxFaultPointsForSummary {
			log.Debugf("达到最大故障点数量限制: %d", maxFaultPointsForSummary)
			break
		}
	}
	if len(faultPointsPayload) == 0 {
		return nil, errors.New("没有可用的故障点数据（token 超限或数据为空）")
	}

	// 最终检查
	customQuerys := map[string]interface{}{
		agentCustomQueryKeyProblemInfo: faultPointsPayload,
	}
	finalTokenCount := s.estimateTokenCount(customQuerys)

	if finalTokenCount > maxInputTokens {
		return nil, fmt.Errorf("token 数量超限: %d > %d", finalTokenCount, maxInputTokens)
	}

	log.Debugf("Agent 描述生成请求 token 估算: %d/%d, 故障点数量: %d/%d",
		finalTokenCount, maxInputTokens, len(faultPointsPayload), len(faultPoints))

	return customQuerys, nil
}

// 找到最高严重程度的故障等级（数值最小的）
// 优化：修复初始值设置错误
func (s *Service) findMaxSeverity(activeFaultPoints []domain.FaultPointObject) domain.Severity {
	if len(activeFaultPoints) == 0 {
		return domain.SeverityNormal
	}

	// 修复：初始化为正常级别（5），这样任何实际的故障级别都会小于它
	// Severity 值越小表示越严重：1=紧急, 2=严重, 3=重要, 4=警告, 5=正常
	maxSeverity := domain.SeverityNormal
	for _, fp := range activeFaultPoints {
		// 找到最严重的故障等级（数值最小的）
		if fp.FaultLevel < maxSeverity {
			maxSeverity = fp.FaultLevel
		}
	}
	return maxSeverity
}

// sortFaultPointsByPriority 按优先级排序故障点
// 优先发送重要的故障点（严重程度高、未恢复、持续时间长）
func (s *Service) sortFaultPointsByPriority(faultPoints []domain.FaultPointObject) []domain.FaultPointObject {
	sorted := make([]domain.FaultPointObject, len(faultPoints))
	copy(sorted, faultPoints)

	sort.Slice(sorted, func(i, j int) bool {
		fpI := sorted[i]
		fpJ := sorted[j]

		// 优先级1：严重程度（值越小越严重）
		if fpI.FaultLevel != fpJ.FaultLevel {
			return fpI.FaultLevel < fpJ.FaultLevel
		}

		// 优先级2：故障状态（未恢复的优先）
		if fpI.FaultStatus != fpJ.FaultStatus {
			if fpI.FaultStatus == domain.FaultStatusOccurred {
				return true
			}
			if fpJ.FaultStatus == domain.FaultStatusOccurred {
				return false
			}
		}

		// 优先级3：发生时间（越早的优先，可能是根因）
		if !fpI.FaultOccurTime.Equal(fpJ.FaultOccurTime) {
			return fpI.FaultOccurTime.Before(fpJ.FaultOccurTime)
		}

		// 优先级4：持续时间（持续时间长的优先）
		if fpI.FaultDurationTime != fpJ.FaultDurationTime {
			return fpI.FaultDurationTime > fpJ.FaultDurationTime
		}

		// 最后按故障ID排序（保证稳定性）
		return fpI.FaultID < fpJ.FaultID
	})

	return sorted
}

// buildAgentCustomQuerys 构建最小化的 Agent 请求（当 token 超限时使用）
// 优化：按优先级排序后再选择前N个
func (s *Service) buildAgentCustomQuerys(faultPoints []domain.FaultPointObject) map[string]interface{} {
	if len(faultPoints) == 0 {
		return map[string]interface{}{
			agentCustomQueryKeyProblemInfo: []map[string]interface{}{},
		}
	}

	// 按优先级排序，选择最重要的故障点
	sortedFaultPoints := s.sortFaultPointsByPriority(faultPoints)

	// 只包含最核心的信息，最多10个故障点
	maxCount := 10
	if len(sortedFaultPoints) < maxCount {
		maxCount = len(sortedFaultPoints)
	}

	minimalPayload := make([]map[string]interface{}, 0, maxCount)
	for i := 0; i < maxCount; i++ {
		fp := sortedFaultPoints[i]
		minimalPayload = append(minimalPayload, map[string]interface{}{
			"fault_id":         fp.FaultID,
			"fault_name":       fp.FaultName,
			"fault_level":      fp.FaultLevel,
			"fault_status":     string(fp.FaultStatus),
			"fault_occur_time": fp.FaultOccurTime.Format(time.RFC3339),
		})
	}

	return map[string]interface{}{
		agentCustomQueryKeyProblemInfo: minimalPayload,
	}
}

// buildAgentPayloadOptimized 构建优化的故障点负载（精简字段，减少 token 消耗）
func (s *Service) buildAgentPayloadOptimized(fp *domain.FaultPointObject) map[string]interface{} {
	if fp == nil {
		return map[string]interface{}{}
	}

	// 只包含关键字段，减少 token 消耗
	payload := map[string]interface{}{
		"fault_id":         fp.FaultID,
		"fault_name":       fp.FaultName,
		"fault_level":      fp.FaultLevel,
		"fault_status":     string(fp.FaultStatus),
		"fault_occur_time": fp.FaultOccurTime.Format(time.RFC3339),
	}

	// 可选字段：只在必要时添加
	if !fp.FaultRecoverTime.IsZero() {
		payload["fault_recovery_time"] = fp.FaultRecoverTime.Format(time.RFC3339)
	}

	return payload
}

// 构建默认的问题现象描述
// 当 Agent 配置缺失或调用失败时使用
func (s *Service) buildDefaultOccurrenceDescription(problemName string, faultPoints []domain.FaultPointObject) domain.Occurrence {
	if len(faultPoints) == 0 {
		return domain.Occurrence{
			Name:        defaultNameNoFaultPoints,
			Description: defaultDescriptionNoFaultPoints,
			Impact:      defaultImpactNoFaultPoints,
		}
	}

	// 构建默认的问题名称 如果问题名称存在，取问题名称，不存在则取故障点的第一个故障点的名称
	name := problemName
	if name == "" {
		name = faultPoints[0].FaultName
	}

	// 构建简单的描述
	description := s.buildDefaultDescription(faultPoints)

	// 统计影响范围
	impact := s.buildDefaultImpact(faultPoints)

	return domain.Occurrence{
		Name:        name,
		Description: description,
		Impact:      impact,
	}
}

// 构建默认描述文本
// 优化：改进排序稳定性
func (s *Service) buildDefaultDescription(faultPoints []domain.FaultPointObject) string {
	if len(faultPoints) == 0 {
		return defaultDescriptionNoFaultPoints
	}

	// 按发生时间排序，获取最早的故障点
	sortedFaultPoints := make([]domain.FaultPointObject, len(faultPoints))
	copy(sortedFaultPoints, faultPoints)
	sort.Slice(sortedFaultPoints, func(i, j int) bool {
		if sortedFaultPoints[i].FaultOccurTime.Before(sortedFaultPoints[j].FaultOccurTime) {
			return true
		}
		if sortedFaultPoints[i].FaultOccurTime.After(sortedFaultPoints[j].FaultOccurTime) {
			return false
		}
		// 时间相同，按故障ID排序
		return sortedFaultPoints[i].FaultID < sortedFaultPoints[j].FaultID
	})

	firstFP := sortedFaultPoints[0]
	firstTimeStr := firstFP.FaultOccurTime.Format(timeFormatDateTime)
	return fmt.Sprintf(defaultDescriptionWithTime, len(faultPoints), firstTimeStr)
}

// 构建默认影响描述
func (s *Service) buildDefaultImpact(faultPoints []domain.FaultPointObject) string {
	entityCount := s.countUniqueEntities(faultPoints)
	return fmt.Sprintf(defaultImpactTemplate, entityCount)
}

// 统计唯一实体数量
func (s *Service) countUniqueEntities(faultPoints []domain.FaultPointObject) int {
	if len(faultPoints) == 0 {
		return 0
	}

	entitySet := make(map[string]bool, len(faultPoints))
	for _, fp := range faultPoints {
		if fp.EntityObjectID != "" {
			entitySet[fp.EntityObjectID] = true
		}
	}
	return len(entitySet)
}

// ----- 4. 分析网络生成 -----

// 生成分析网络（包含对象实体、故障点、关系）
func (s *Service) GenerateRcaNetwork(ctx context.Context, faultPointInfos []domain.FaultPointObject, recallCtx *domain.GraphRecallContext, result *domain.CausalAnalysisResults) (*domain.RcaNetwork, error) {
	network := &domain.RcaNetwork{
		Nodes: []domain.RcaNode{},
		Edges: []domain.Relation{},
	}

	// 用于去重的映射
	nodeMap := make(map[string]*domain.RcaNode) // entityID -> RcaNode (对象实体节点)
	edgeMap := make(map[string]bool)            // relationID -> bool

	// 1. 收集关联的对象实体ID（用于快速查找）
	entityIDSet := s.collectEntityIDs(faultPointInfos)

	// 2. 从对象子图中提取对象节点和关系
	s.processTopologySubgraphs(network, nodeMap, edgeMap, entityIDSet, faultPointInfos, recallCtx)

	// 3. 从故障点中提取事件ID，并关联到对象节点
	s.processEventIDs(nodeMap, faultPointInfos)

	// 4. 将所有对象节点添加到网络（预分配容量以提高性能）
	network.Nodes = make([]domain.RcaNode, 0, len(nodeMap))
	for _, node := range nodeMap {
		network.Nodes = append(network.Nodes, *node)
	}

	return network, nil
}

// 收集实体ID集合
func (s *Service) collectEntityIDs(faultPointInfos []domain.FaultPointObject) map[string]bool {
	entityIDSet := make(map[string]bool, len(faultPointInfos))
	for _, fp := range faultPointInfos {
		if fp.EntityObjectID != "" {
			entityIDSet[fp.EntityObjectID] = true
		}
	}
	return entityIDSet
}

// 处理对象子图，提取对象节点和关系
// 优化：添加错误处理和边界检查
func (s *Service) processTopologySubgraphs(network *domain.RcaNetwork, nodeMap map[string]*domain.RcaNode, edgeMap map[string]bool, entityIDSet map[string]bool, faultPointInfos []domain.FaultPointObject, recallCtx *domain.GraphRecallContext) {
	if recallCtx == nil || recallCtx.TopologySubgraphs == nil {
		return
	}

	for _, topology := range recallCtx.TopologySubgraphs {
		if topology == nil {
			continue
		}

		// 2.1 处理对象节点
		for _, topologyNode := range topology.Nodes {
			if topologyNode.SID == "" {
				continue
			}

			// 获取或创建节点
			node := s.getOrCreateNode(nodeMap, topologyNode)

			// 如果该对象关联了故障点，添加故障点信息
			if entityIDSet[topologyNode.SID] {
				s.addFaultPointsToNode(node, faultPointInfos, topologyNode.SID)
				// 更新 ObjectImpactLevel（根据 RelationFaultPointIDs 过滤出相关故障点）
				relatedFaultPoints := s.filterFaultPointsByIDs(faultPointInfos, node.RelationFaultPointIDs)
				node.ObjectImpactLevel = s.calculateObjectImpactLevel(relatedFaultPoints)
			} else if node.ObjectImpactLevel == 0 {
				// 如果该对象没有关联故障点，且 ObjectImpactLevel 未设置，设置为正常（5）
				node.ObjectImpactLevel = objectImpactLevelNormal
			}
		}

		// 2.2 处理对象之间的关系（边）
		s.processTopologyEdges(network, edgeMap, topology.Edges)
	}
}

// 获取或创建对象节点
func (s *Service) getOrCreateNode(nodeMap map[string]*domain.RcaNode, topologyNode domain.Node) *domain.RcaNode {
	node, exists := nodeMap[topologyNode.SID]
	if !exists {
		node = &domain.RcaNode{
			Node: domain.Node{
				SID:               topologyNode.SID,
				SCreateTime:       topologyNode.SCreateTime,
				SUpdateTime:       topologyNode.SUpdateTime,
				IPAddress:         topologyNode.IPAddress,
				Name:              topologyNode.Name,
				ObjectClass:       topologyNode.ObjectClass,
				ObjectImpactLevel: topologyNode.ObjectImpactLevel,
			},
			RelationFaultPointIDs: []uint64{},
			RelationEventIDs:      []string{},
		}
		nodeMap[topologyNode.SID] = node
	}
	return node
}

// 添加故障点到节点
// 优化：改进去重逻辑，添加参数验证
func (s *Service) addFaultPointsToNode(node *domain.RcaNode, faultPointInfos []domain.FaultPointObject, entityID string) {
	if node == nil || entityID == "" {
		return
	}

	// 使用 map 进行快速去重检查
	faultPointIDSet := s.buildFaultPointIDSet(node.RelationFaultPointIDs)

	for _, fp := range faultPointInfos {
		if fp.EntityObjectID != entityID {
			continue
		}

		// 添加到 RelationFaultPointIDs（去重）
		if !faultPointIDSet[fp.FaultID] {
			node.RelationFaultPointIDs = append(node.RelationFaultPointIDs, fp.FaultID)
			faultPointIDSet[fp.FaultID] = true
		}
	}
}

// 构建故障点ID集合（用于去重）
func (s *Service) buildFaultPointIDSet(ids []uint64) map[uint64]bool {
	idSet := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	return idSet
}

// 根据故障点ID列表从故障点信息中过滤出相关的故障点
func (s *Service) filterFaultPointsByIDs(faultPointInfos []domain.FaultPointObject, faultPointIDs []uint64) []domain.FaultPointObject {
	if len(faultPointIDs) == 0 {
		return []domain.FaultPointObject{}
	}

	// 构建ID集合用于快速查找
	idSet := make(map[uint64]bool, len(faultPointIDs))
	for _, id := range faultPointIDs {
		idSet[id] = true
	}

	// 过滤出匹配的故障点
	result := make([]domain.FaultPointObject, 0, len(faultPointIDs))
	for _, fp := range faultPointInfos {
		if idSet[fp.FaultID] {
			result = append(result, fp)
		}
	}
	return result
}

// 计算对象影响级别
// 优化：改进注释和逻辑
func (s *Service) calculateObjectImpactLevel(faultPoints []domain.FaultPointObject) int {
	if len(faultPoints) == 0 {
		return objectImpactLevelNormal
	}

	// 收集所有未恢复的故障点
	activeFaultPoints := s.filterActiveFaultPoints(faultPoints)

	if len(activeFaultPoints) == 0 {
		return objectImpactLevelNormal // 所有故障点都已恢复，设为正常
	}

	// 找到最高严重程度的故障等级（数值最小的）
	maxSeverity := s.findMaxSeverity(activeFaultPoints)

	// 将 Severity 转换为 ObjectImpactLevel（1=紧急, 2=严重, 3=重要, 4=警告, 5=正常）
	return int(maxSeverity)
}

// 过滤出未恢复的故障点
func (s *Service) filterActiveFaultPoints(faultPoints []domain.FaultPointObject) []domain.FaultPointObject {
	activeFaultPoints := make([]domain.FaultPointObject, 0, len(faultPoints))
	for _, fp := range faultPoints {
		if fp.FaultStatus == domain.FaultStatusOccurred {
			activeFaultPoints = append(activeFaultPoints, fp)
		}
	}
	return activeFaultPoints
}

// 处理拓扑边
// 优化：添加参数验证
func (s *Service) processTopologyEdges(network *domain.RcaNetwork, edgeMap map[string]bool, edges []domain.Relation) {
	if network == nil || edgeMap == nil {
		return
	}

	for _, relation := range edges {
		if relation.RelationID == "" {
			continue
		}

		// 去重边
		if edgeMap[relation.RelationID] {
			continue
		}
		edgeMap[relation.RelationID] = true
		network.Edges = append(network.Edges, relation)
	}
}

// processEventIDs 处理事件ID
// 优化：添加参数验证和错误处理
func (s *Service) processEventIDs(nodeMap map[string]*domain.RcaNode, faultPointInfos []domain.FaultPointObject) {
	if nodeMap == nil {
		return
	}

	for _, fp := range faultPointInfos {
		if fp.EntityObjectID == "" {
			continue
		}

		node, exists := nodeMap[fp.EntityObjectID]
		if !exists || len(fp.RelationEventIDs) == 0 {
			continue
		}

		// 使用 map 进行快速去重检查
		eventIDSet := s.buildEventIDSet(node.RelationEventIDs)

		// 添加新的事件ID（去重）
		for _, eventID := range fp.RelationEventIDs {
			eventIDStr := fmt.Sprintf("%d", eventID)
			if !eventIDSet[eventIDStr] {
				node.RelationEventIDs = append(node.RelationEventIDs, eventIDStr)
				eventIDSet[eventIDStr] = true
			}
		}
	}
}

// buildEventIDSet 构建事件ID集合（用于去重）
func (s *Service) buildEventIDSet(ids []string) map[string]bool {
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id != "" {
			idSet[id] = true
		}
	}
	return idSet
}
