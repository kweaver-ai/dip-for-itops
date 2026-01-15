package rca

import (
	"context"
	"fmt"
	"slices"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
)

// 对象子图定义：只包含源和目标都在 entityIDs 中的节点和边，不包含邻居节点
func (s *Service) convertSubGraphQueryResponseToTopology(resp *domain.SubGraphQueryResponse, entityClassID string, objectInfoDatas []interface{}, affectedEntityIDs []string) *domain.Topology {
	topology := &domain.Topology{
		Nodes: make([]domain.Node, 0),
		Edges: make([]domain.Relation, 0),
	}

	// 用于去重节点的 map（使用 s_id 作为key）
	nodeMap := make(map[string]*domain.Node)

	// 用于去重边的 map（使用 RelationID 作为 key）
	edgeMap := make(map[string]*domain.Relation)
	// 构建 affectedEntityIDs 集合用于快速查找
	affectedEntityIDSet := make(map[string]bool, len(affectedEntityIDs))
	for _, id := range affectedEntityIDs {
		if id != "" {
			affectedEntityIDSet[id] = true
		}
	}

	// 1. 遍历 Objects，提取所有节点
	s.extractNodesFromObjects(resp.Objects, entityClassID, objectInfoDatas, nodeMap, affectedEntityIDSet, topology)

	// 2. 遍历关系路径，提取关系和关联的节点
	s.extractEdgesFromRelationPaths(resp.RelationPaths, resp.Objects, edgeMap, affectedEntityIDSet, topology)

	return topology
}

// 从对象中提取节点
func (s *Service) extractNodesFromObjects(objects map[string]domain.SubGraphObject, entityClassID string, objectInfoDatas []interface{}, nodeMap map[string]*domain.Node, affectedEntityIDSet map[string]bool, topology *domain.Topology) {
	if topology == nil {
		return
	}
	// 判断 objects 是否为空，如果为空，则使用 objectInfoDatas 中的数据
	if len(objects) == 0 && len(objectInfoDatas) > 0 {
		// 从 objectInfoDatas 中提取节点
		for _, objectInfoData := range objectInfoDatas {
			// 提取 objectInfoDatas 中的数据，转换为 Node 对象
			node := s.extractNodeFromObjectInfo(entityClassID, objectInfoData)
			if node != nil && node.SID != "" {
				// 使用 s_id 作为 key 去重
				if _, exists := nodeMap[node.SID]; !exists {
					nodeMap[node.SID] = node
					// 只有当 node.SID 在 affectedEntityIDSet 中时，且对象类型在 ObjectClasses 中时，才添加到 topology.Nodes
					if affectedEntityIDSet[node.SID] && slices.Contains(ObjectClasses, node.ObjectClass) {
						topology.Nodes = append(topology.Nodes, *node)
					}
				}
			}
		}
	}

	for _, obj := range objects {
		node := s.extractNodeFromSubGraphObject(obj)
		if node != nil && node.SID != "" {
			// 使用 s_id 作为 key 去重
			if _, exists := nodeMap[node.SID]; !exists {
				nodeMap[node.SID] = node
				// 只有当 node.SID 在 affectedEntityIDSet 中时，且对象类型在 ObjectClasses 中时，才添加到 topology.Nodes
				if affectedEntityIDSet[node.SID] && slices.Contains(ObjectClasses, node.ObjectClass) {
					topology.Nodes = append(topology.Nodes, *node)
				}
			}
		}
	}
}

// 从 objectInfoData 中提取 Node
func (s *Service) extractNodeFromObjectInfo(entityClassID string, objectInfoData interface{}) *domain.Node {
	// 参数验证
	if objectInfoData == nil {
		return nil
	}

	// 类型断言（安全处理）
	objectInfo, ok := objectInfoData.(map[string]interface{})
	if !ok {
		return nil
	}

	node := &domain.Node{}

	// 提取 SID
	if sID, ok := objectInfo["s_id"].(string); ok {
		node.SID = sID
	}

	// 提取对象类型
	node.ObjectClass = entityClassID

	// 提取名称
	if name, ok := objectInfo["name"].(string); ok {
		node.Name = name
	}

	// 提取 IP 地址（可能是字符串或字符串数组）
	if ipAddr, ok := objectInfo["ip_address"].(string); ok {
		// 单个 IP 地址
		node.IPAddress = []string{ipAddr}
	} else if ipAddrs, ok := objectInfo["ip_address"].([]interface{}); ok {
		// IP 地址数组
		node.IPAddress = make([]string, 0, len(ipAddrs))
		for _, ip := range ipAddrs {
			if ipStr, ok := ip.(string); ok {
				node.IPAddress = append(node.IPAddress, ipStr)
			}
		}
	} else if ipAddrs, ok := objectInfo["ip_address"].([]string); ok {
		// 直接是字符串数组
		node.IPAddress = ipAddrs
	}

	// 提取创建时间
	if createTime, ok := objectInfo["s_create_time"].(string); ok {
		node.SCreateTime = createTime
	}

	// 提取更新时间
	if updateTime, ok := objectInfo["s_update_time"].(string); ok {
		node.SUpdateTime = updateTime
	}

	return node
}

// 从关系路径中提取边
func (s *Service) extractEdgesFromRelationPaths(relationPaths []domain.SubGraphRelationPath, objects map[string]domain.SubGraphObject, edgeMap map[string]*domain.Relation, affectedEntityIDSet map[string]bool, topology *domain.Topology) {
	if topology == nil {
		return
	}
	for _, path := range relationPaths {
		for _, subGraphRelation := range path.Relations {
			// 提取关系（edge）转换为 Relation 对象
			edge := s.extractRelationFromSubGraphRelation(subGraphRelation, objects)
			if edge != nil {
				// 只有当源和目标都在 entityIDs 中时，才添加到 topology.Edges（对象子图不包含邻居）
				if affectedEntityIDSet[edge.SourceSID] && affectedEntityIDSet[edge.TargetSID] {
					// edge 去重（使用 RelationID 作为 key）
					if _, exists := edgeMap[edge.RelationID]; !exists {
						edgeMap[edge.RelationID] = edge
						topology.Edges = append(topology.Edges, *edge)
					}
				}
			}
		}
	}
}

// 存储对象子图到召回上下文
// 为每个 entityID 存储拓扑图（如果该 entityID 对应的对象存在于响应中）
func (s *Service) storeTopologySubgraphs(recallCtx *domain.GraphRecallContext, affectedEntityIDs []string, topology *domain.Topology) {
	// 参数验证
	if recallCtx == nil || affectedEntityIDs == nil || topology == nil {
		return
	}

	// 确保 TopologySubgraphs map 已初始化
	if recallCtx.TopologySubgraphs == nil {
		recallCtx.TopologySubgraphs = make(map[string]*domain.Topology)
	}

	// 如果拓扑为空，直接返回
	if len(topology.Nodes) == 0 {
		return
	}

	// 构建 SID 到拓扑的映射，用于快速查找
	sidToTopology := make(map[string]*domain.Topology)

	// 遍历拓扑节点，构建 SID 映射
	for _, node := range topology.Nodes {
		if node.SID != "" {
			sidToTopology[node.SID] = topology
		}
	}

	// 为每个 entityID 存储拓扑图
	for _, entityID := range affectedEntityIDs {
		if entityID == "" {
			continue
		}

		// 检查该 entityID 是否在拓扑中（通过 SID 匹配）
		if _, exists := sidToTopology[entityID]; exists {
			// 如果已存在拓扑图，合并而不是覆盖
			if existingTopology, exists := recallCtx.TopologySubgraphs[entityID]; exists {
				// 合并拓扑图：合并节点和边，去重
				mergedTopology := s.mergeTopologies(existingTopology, topology)
				recallCtx.TopologySubgraphs[entityID] = mergedTopology
				log.Debugf("为 entityID %s 合并拓扑子图（已存在，合并）", entityID)
			} else {
				// 不存在，直接存储
				recallCtx.TopologySubgraphs[entityID] = topology
				log.Debugf("为 entityID %s 存储拓扑子图", entityID)
			}
		}
	}
}

// mergeTopologies 合并两个拓扑图，去重节点和边
func (s *Service) mergeTopologies(existing, new *domain.Topology) *domain.Topology {
	if existing == nil {
		return new
	}
	if new == nil {
		return existing
	}

	// 创建合并后的拓扑图
	merged := &domain.Topology{
		Nodes: make([]domain.Node, 0, len(existing.Nodes)+len(new.Nodes)),
		Edges: make([]domain.Relation, 0, len(existing.Edges)+len(new.Edges)),
	}

	// 使用 map 去重节点（使用 SID 作为 key）
	nodeMap := make(map[string]*domain.Node)

	// 添加已存在的节点
	for i := range existing.Nodes {
		node := &existing.Nodes[i]
		if node.SID != "" {
			nodeMap[node.SID] = node
		}
	}

	// 添加新节点（如果 SID 已存在，会被覆盖，保留新节点的数据）
	for i := range new.Nodes {
		node := &new.Nodes[i]
		if node.SID != "" {
			nodeMap[node.SID] = node
		}
	}

	// 将去重后的节点添加到合并后的拓扑图
	for _, node := range nodeMap {
		merged.Nodes = append(merged.Nodes, *node)
	}

	// 使用 map 去重边（使用 RelationID 作为 key）
	edgeMap := make(map[string]*domain.Relation)

	// 添加已存在的边
	for i := range existing.Edges {
		edge := &existing.Edges[i]
		if edge.RelationID != "" {
			edgeMap[edge.RelationID] = edge
		}
	}

	// 添加新边（如果 RelationID 已存在，会被覆盖，保留新边的数据）
	for i := range new.Edges {
		edge := &new.Edges[i]
		if edge.RelationID != "" {
			edgeMap[edge.RelationID] = edge
		}
	}

	// 将去重后的边添加到合并后的拓扑图
	for _, edge := range edgeMap {
		merged.Edges = append(merged.Edges, *edge)
	}

	return merged
}

// 从 SubGraphObject 中提取 Node
func (s *Service) extractNodeFromSubGraphObject(obj domain.SubGraphObject) *domain.Node {
	node := &domain.Node{}

	// 提取 SID（从 UniqueIdentities 或 Properties）
	if obj.UniqueIdentities.SID != "" {
		node.SID = obj.UniqueIdentities.SID
	} else if obj.Properties != nil {
		if sid, ok := obj.Properties[propertyKeySID].(string); ok {
			node.SID = sid
		}
	}

	// 提取对象类型
	node.ObjectClass = obj.ObjectTypeID
	if node.ObjectClass == "" {
		node.ObjectClass = obj.ObjectTypeName
	}

	// 提取名称
	if obj.Display != "" {
		node.Name = obj.Display
	} else if obj.Properties != nil {
		if name, ok := obj.Properties[propertyKeyName].(string); ok {
			node.Name = name
		}
	}

	// 提取 IP 地址
	if obj.Properties != nil {
		if ipAddr, ok := obj.Properties[propertyKeyIPAddress].(string); ok {
			node.IPAddress = []string{ipAddr}
		} else if ipAddrs, ok := obj.Properties[propertyKeyIPAddress].([]interface{}); ok {
			node.IPAddress = make([]string, 0, len(ipAddrs))
			for _, ip := range ipAddrs {
				if ipStr, ok := ip.(string); ok {
					node.IPAddress = append(node.IPAddress, ipStr)
				}
			}
		}
	}

	// 提取创建时间（从 properties 中）
	if obj.Properties != nil {
		if createTimeStr, ok := obj.Properties[propertyKeySCreateTime].(string); ok {
			node.SCreateTime = createTimeStr
		}
	}

	// 提取更新时间（从 properties 中）
	if obj.Properties != nil {
		if updateTimeStr, ok := obj.Properties[propertyKeySUpdateTime].(string); ok {
			node.SUpdateTime = updateTimeStr
		}
	}

	return node
}

// 从 SubGraphRelation 中提取 Relation
// 注意：需要从 objects map 中查找对应的对象，提取它们的 s_id 作为 SourceSID 和 TargetSID
func (s *Service) extractRelationFromSubGraphRelation(subGraphRelation domain.SubGraphRelation, objects map[string]domain.SubGraphObject) *domain.Relation {
	edge := &domain.Relation{}

	// 提取关系类型（优先使用 RelationTypeID，否则使用 RelationTypeName）
	edge.RelationClass = subGraphRelation.RelationTypeID
	if edge.RelationClass == "" {
		edge.RelationClass = subGraphRelation.RelationTypeName
	}

	// 如果关系类型为空，无法生成有效的关系
	if edge.RelationClass == "" {
		return nil
	}

	// 从 objects map 中查找源对象和目标对象，提取它们的 s_id
	sourceSID := s.extractSIDFromObjectID(subGraphRelation.SourceObjectID, objects)
	targetSID := s.extractSIDFromObjectID(subGraphRelation.TargetObjectID, objects)

	// 如果无法提取 s_id，返回 nil
	if sourceSID == "" || targetSID == "" {
		return nil
	}

	edge.SourceSID = sourceSID
	edge.TargetSID = targetSID

	// 生成关系ID（使用关系类型和 s_id 组合）
	edge.RelationID = fmt.Sprintf("%s_%s_%s", edge.RelationClass, sourceSID, targetSID)

	return edge
}

// 从对象ID中提取 s_id
// 首先尝试从 objects map 中查找对象，然后提取 s_id
func (s *Service) extractSIDFromObjectID(objectID string, objects map[string]domain.SubGraphObject) string {
	if objectID == "" || objects == nil {
		return ""
	}

	// 从 objects map 中查找对象
	if obj, exists := objects[objectID]; exists {
		// 优先从 UniqueIdentities 中提取
		if obj.UniqueIdentities.SID != "" {
			return obj.UniqueIdentities.SID
		}
		// 否则从 Properties 中提取
		if obj.Properties != nil {
			if sid, ok := obj.Properties[propertyKeySID].(string); ok {
				return sid
			}
		}
	}
	// 如果对象不存在或无法提取 s_id，返回空字符串
	return ""
}

// ----- 1.1 处理返回的数据：召回故障点关联对象的子图（只包含查询对象本身，不包含邻居） -----
func (s *Service) recallTopologySubgraph(ctx context.Context, recallCtx *domain.GraphRecallContext, entityClassID string, entityIDs []string, affectedEntityIDs []string) {
	var objectInfoResp *domain.ObjectInfoQueryResponse
	var objectInfoDatas []interface{}
	if len(entityIDs) == 0 {
		return
	}

	// 参数验证
	if recallCtx == nil {
		return
	}

	subgraphResp, err := s.dipClient.QueryTopologyObjectSubgraph(ctx, entityClassID, entityIDs, s.config.AppConfig.Credentials.Authorization)
	if err != nil || subgraphResp == nil {
		log.Infof("召回问题关联拓扑对象的子图失败: %v", err)
		return
	}

	if len(subgraphResp.Objects) == 0 {
		objectInfoResp, err = s.dipClient.QueryObjectInfo(ctx, entityClassID, entityIDs, s.config.AppConfig.Credentials.Authorization)
		if err != nil || objectInfoResp == nil {
			log.Infof("召回问题关联对象的信息失败: %v", err)
			return
		}
		objectInfoDatas = objectInfoResp.Datas
	}

	// 将 SubGraphQueryResponse 转换为 Topology
	topology := s.convertSubGraphQueryResponseToTopology(subgraphResp, entityClassID, objectInfoDatas, affectedEntityIDs)
	if topology == nil {
		log.Infof("将 SubGraphQueryResponse 转换为 Topology 失败")
		return
	}

	// 为每个 entityID 存储拓扑图（如果该 entityID 对应的对象存在于响应中）
	s.storeTopologySubgraphs(recallCtx, affectedEntityIDs, topology)
}
