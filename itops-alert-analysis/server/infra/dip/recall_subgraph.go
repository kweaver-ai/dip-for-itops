package dip

import (
	"context"
	"fmt"
	"strconv"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"github.com/pkg/errors"
)

const (
	// 子图查询URL 常量
	SubgraphURL        = "/api/ontology-query/v1/knowledge-networks/%s/subgraph"        // 业务知识网络子图查询URL
	ObjectInfoQueryURL = "/api/ontology-query/v1/knowledge-networks/%s/object-types/%s" // 业务知识网络对象信息查询URL

	// 查询字段名
	queryFieldSID         = "s_id"          // 实体ID字段
	queryFieldScreateTime = "s_create_time" // 实体创建时间字段
	queryFieldSupdateTime = "s_update_time" // 实体更新时间字段
	queryFieldName        = "name"          // 实体名称字段
	queryFieldIPAddress   = "ip_address"    // 实体IP地址字段

	queryFieldFaultID = "fault_id" // 故障ID字段
	// 查询操作符
	queryOperationEqual = "==" // 等于操作符
	queryOperationOR    = "or" // 或操作符
	// 查询方向
	queryDirectionForward       = "forward"       // 前向查询
	queryDirectionBidirectional = "bidirectional" // 双向查询
	// 查询参数
	queryPathLengthObjectSubgraph = 1     // 对象子图路径长度
	queryPathLengthCausality      = 2     // 因果关系路径长度
	queryLimitDefault             = 10000 // 默认查询限制
)

// 根据实体ID列表构建子查询条件（使用s_id字段）
func (c *Client) buildSubConditionsBySID(entityIDs []string) []domain.SubGraphSubCondition {
	if len(entityIDs) == 0 {
		return []domain.SubGraphSubCondition{}
	}

	subConditions := make([]domain.SubGraphSubCondition, 0, len(entityIDs))
	for _, entityID := range entityIDs {
		// 过滤空字符串
		if entityID == "" {
			continue
		}
		subConditions = append(subConditions, domain.SubGraphSubCondition{
			Field:     queryFieldSID,
			Operation: queryOperationEqual,
			Value:     entityID,
		})
	}
	return subConditions
}

// 根据故障ID列表构建子查询条件（使用fault_id字段）
func (c *Client) buildSubConditionsByFaultID(faultIDs []uint64) []domain.SubGraphSubCondition {
	if len(faultIDs) == 0 {
		return []domain.SubGraphSubCondition{}
	}

	subConditions := make([]domain.SubGraphSubCondition, 0, len(faultIDs))
	for _, faultID := range faultIDs {
		// 过滤无效的故障ID（0 通常表示无效ID）
		if faultID == 0 {
			continue
		}
		subConditions = append(subConditions, domain.SubGraphSubCondition{
			Field:     queryFieldFaultID,
			Operation: queryOperationEqual,
			Value:     strconv.FormatUint(faultID, 10),
		})
	}
	return subConditions
}

// 构建子图查询请求
func (c *Client) buildQueryRequest(entityClassID string, subConditions []domain.SubGraphSubCondition, direction string, pathLength int) domain.SubGraphQueryRequest {
	// 如果子条件为空，创建一个空条件以避免查询错误
	condition := &domain.SubGraphCondition{
		Operation:     queryOperationOR,
		SubConditions: subConditions,
	}

	return domain.SubGraphQueryRequest{
		SourceObjectTypeID: entityClassID,
		Condition:          condition,
		Direction:          direction,
		PathLength:         pathLength,
		NeedTotal:          false,
		Limit:              queryLimitDefault,
	}
}

// 执行子图查询（公共逻辑）
func (c *Client) executeSubgraphQuery(ctx context.Context, queryReq domain.SubGraphQueryRequest, authorization string, errorMsg string) (*domain.SubGraphQueryResponse, error) {
	// 参数验证
	if c == nil {
		return nil, errors.Errorf("%s: client 未初始化", errorMsg)
	}
	if c.httpClient == nil {
		return nil, errors.Errorf("%s: http client 未初始化", errorMsg)
	}
	if ctx == nil {
		return nil, errors.Errorf("%s: 上下文不能为 nil", errorMsg)
	}
	if c.KnID() == "" {
		return nil, errors.Errorf("%s: 知识网络 ID 不能为空", errorMsg)
	}

	// 验证查询请求
	if queryReq.SourceObjectTypeID == "" {
		return nil, errors.Errorf("%s: 源对象类型 ID 不能为空", errorMsg)
	}
	if queryReq.Condition == nil {
		return nil, errors.Errorf("%s: 查询条件不能为 nil", errorMsg)
	}
	if len(queryReq.Condition.SubConditions) == 0 {
		return nil, errors.Errorf("%s: 查询子条件不能为空", errorMsg)
	}

	subgraphURL := fmt.Sprintf(SubgraphURL, c.KnID())
	header := make(map[string]string)
	header["author"] = "applicaiton/json"
	header["x-http-method-override"] = "GET"
	header["authorization"] = authorization

	resp, err := c.httpClient.Post(ctx, subgraphURL, queryReq, header)
	if err != nil {
		return nil, errors.Errorf("%s: %v", errorMsg, err)
	}

	// 检查响应是否为 nil
	if resp == nil {
		return nil, errors.Errorf("%s: 响应为空", errorMsg)
	}

	if err := resp.Error(); err != nil {
		return nil, errors.Errorf("%s: %v", errorMsg, err)
	}

	var result domain.SubGraphQueryResponse
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, errors.Errorf("%s: 解析响应失败: %v", errorMsg, err)
	}

	return &result, nil
}

// 1.1 召回问题关联拓扑对象的子图（只包含查询对象本身，不包含邻居）
func (c *Client) QueryTopologyObjectSubgraph(ctx context.Context, entityObjectClass string, entityObjectIDs []string, authorization string) (*domain.SubGraphQueryResponse, error) {
	// 参数验证
	if c == nil {
		return nil, errors.New("client 未初始化")
	}
	if ctx == nil {
		return nil, errors.New("上下文不能为 nil")
	}
	if entityObjectClass == "" {
		return nil, errors.New("实体对象类型不能为空")
	}

	if len(entityObjectIDs) == 0 {
		return &domain.SubGraphQueryResponse{
			Objects:       make(map[string]domain.SubGraphObject),
			RelationPaths: []domain.SubGraphRelationPath{},
			SearchAfter:   []interface{}{},
		}, nil
	}

	// 构建子查询条件
	subConditions := c.buildSubConditionsBySID(entityObjectIDs)
	if len(subConditions) == 0 {
		// 如果所有 entityID 都无效，返回空结果
		return &domain.SubGraphQueryResponse{
			Objects:       make(map[string]domain.SubGraphObject),
			RelationPaths: []domain.SubGraphRelationPath{},
			SearchAfter:   []interface{}{},
		}, nil
	}

	// 构建查询请求
	queryReq := c.buildQueryRequest(entityObjectClass, subConditions, queryDirectionForward, queryPathLengthObjectSubgraph)

	// 执行查询
	return c.executeSubgraphQuery(ctx, queryReq, authorization, "对象子图查询失败")
}

// 1.2 召回一度拓扑邻居
func (c *Client) QueryTopologyNeighbors(ctx context.Context, entityObjectClass string, entityObjectIDs []string, authorization string) (*domain.SubGraphQueryResponse, error) {
	// 参数验证
	if c == nil {
		return nil, errors.New("client 未初始化")
	}
	if ctx == nil {
		return nil, errors.New("上下文不能为 nil")
	}
	if entityObjectClass == "" {
		return nil, errors.New("实体对象类型不能为空")
	}

	if len(entityObjectIDs) == 0 {
		return &domain.SubGraphQueryResponse{
			Objects:       make(map[string]domain.SubGraphObject),
			RelationPaths: []domain.SubGraphRelationPath{},
			SearchAfter:   []interface{}{},
		}, nil
	}

	// 构建子查询条件
	subConditions := c.buildSubConditionsBySID(entityObjectIDs)
	if len(subConditions) == 0 {
		// 如果所有 entityID 都无效，返回空结果
		return &domain.SubGraphQueryResponse{
			Objects:       make(map[string]domain.SubGraphObject),
			RelationPaths: []domain.SubGraphRelationPath{},
			SearchAfter:   []interface{}{},
		}, nil
	}

	// 构建查询请求（双向查询，路径长度为1）
	queryReq := c.buildQueryRequest(entityObjectClass, subConditions, queryDirectionBidirectional, queryPathLengthObjectSubgraph)

	// 执行查询
	return c.executeSubgraphQuery(ctx, queryReq, authorization, "一度拓扑邻居查询失败")
}

// 1.3 召回历史故障因果关系
func (c *Client) QueryHistoricalCausality(ctx context.Context, entityClassID string, entityIDs []uint64, authorization string) (*domain.SubGraphQueryResponse, error) {
	// 参数验证
	if c == nil {
		return nil, errors.New("client 未初始化")
	}
	if ctx == nil {
		return nil, errors.New("上下文不能为 nil")
	}
	if entityClassID == "" {
		return nil, errors.New("实体类型 ID 不能为空")
	}

	if len(entityIDs) == 0 {
		return &domain.SubGraphQueryResponse{
			Objects:       make(map[string]domain.SubGraphObject),
			RelationPaths: []domain.SubGraphRelationPath{},
			SearchAfter:   []interface{}{},
		}, nil
	}

	// 构建子查询条件（使用fault_id字段）
	subConditions := c.buildSubConditionsByFaultID(entityIDs)
	if len(subConditions) == 0 {
		// 如果所有 faultID 都无效，返回空结果
		return &domain.SubGraphQueryResponse{
			Objects:       make(map[string]domain.SubGraphObject),
			RelationPaths: []domain.SubGraphRelationPath{},
			SearchAfter:   []interface{}{},
		}, nil
	}

	// 构建查询请求（前向查询，路径长度为2：entityID -> FaultCausal -> anotherEntityID）
	queryReq := c.buildQueryRequest(entityClassID, subConditions, queryDirectionForward, queryPathLengthCausality)

	// 执行查询
	return c.executeSubgraphQuery(ctx, queryReq, authorization, "历史因果关系查询失败")
}

// 根据实体ID列表构建对象信息查询条件（使用s_id字段）
func (c *Client) buildObjectSubConditions(entityIDs []string) []domain.SubGraphSubCondition {
	if len(entityIDs) == 0 {
		return []domain.SubGraphSubCondition{}
	}

	subConditions := make([]domain.SubGraphSubCondition, 0, len(entityIDs))
	for _, entityID := range entityIDs {
		// 过滤空字符串
		if entityID == "" {
			continue
		}
		subConditions = append(subConditions, domain.SubGraphSubCondition{
			Field:     queryFieldSID,
			Operation: queryOperationEqual,
			Value:     entityID,
		})
	}
	return subConditions
}

// 构建对象信息查询请求
func (c *Client) buildObjectInfoQueryRequest(subConditions []domain.SubGraphSubCondition) domain.ObjectInfoQueryRequest {
	return domain.ObjectInfoQueryRequest{
		Condition: &domain.SubGraphCondition{
			Operation:     queryOperationOR,
			SubConditions: subConditions,
		},
		NeedTotal:  false,
		Limit:      queryLimitDefault,
		Properties: []string{queryFieldSID, queryFieldName, queryFieldIPAddress, queryFieldScreateTime, queryFieldSupdateTime},
	}
}

// 执行对象信息查询
func (c *Client) executeObjectInfoQuery(ctx context.Context, queryReq domain.ObjectInfoQueryRequest, entityObjectClass string, authorization string, errorMsg string) (*domain.ObjectInfoQueryResponse, error) {
	// 参数验证
	if c == nil {
		return nil, errors.New("client 未初始化")
	}
	if ctx == nil {
		return nil, errors.New("上下文不能为 nil")
	}
	if c.KnID() == "" {
		return nil, errors.New("知识网络 ID 不能为空")
	}

	// 验证查询请求
	if queryReq.Condition == nil {
		return nil, errors.New("查询条件不能为 nil")
	}
	if len(queryReq.Condition.SubConditions) == 0 {
		return nil, errors.New("查询子条件不能为空")
	}

	objectInfoURL := fmt.Sprintf(ObjectInfoQueryURL, c.KnID(), entityObjectClass)
	header := make(map[string]string)
	header["author"] = "applicaiton/json"
	header["x-http-method-override"] = "GET"
	header["authorization"] = authorization

	resp, err := c.httpClient.Post(ctx, objectInfoURL, queryReq, header)
	if err != nil {
		return nil, errors.Errorf("%s: %v", errorMsg, err)
	}

	if err := resp.Error(); err != nil {
		return nil, errors.Errorf("%s: %v", errorMsg, err)
	}
	var result domain.ObjectInfoQueryResponse
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, errors.Errorf("%s: 解析响应失败: %v", errorMsg, err)
	}

	return &result, nil
}

// 查询对象信息
func (c *Client) QueryObjectInfo(ctx context.Context, entityObjectClass string, entityObjectIDs []string, authorization string) (*domain.ObjectInfoQueryResponse, error) {
	// 参数验证
	if c == nil {
		return nil, errors.New("client 未初始化")
	}
	if ctx == nil {
		return nil, errors.New("上下文不能为 nil")
	}
	if entityObjectClass == "" {
		return nil, errors.New("实体对象类型不能为空")
	}
	if len(entityObjectIDs) == 0 {
		return &domain.ObjectInfoQueryResponse{
			Datas:           []interface{}{},
			SearchAfter:     []interface{}{},
			SearchFromIndex: false,
			OverallMS:       0,
		}, nil
	}

	// 构建子查询条件
	subConditions := c.buildObjectSubConditions(entityObjectIDs)
	if len(subConditions) == 0 {
		// 如果所有 entityID 都无效，返回空结果
		return &domain.ObjectInfoQueryResponse{
			Datas:           []interface{}{},
			SearchAfter:     []interface{}{},
			SearchFromIndex: false,
			OverallMS:       0,
		}, nil
	}

	// 构建查询请求
	queryReq := c.buildObjectInfoQueryRequest(subConditions)

	// 执行查询
	return c.executeObjectInfoQuery(ctx, queryReq, entityObjectClass, authorization, "对象信息查询失败")
}
