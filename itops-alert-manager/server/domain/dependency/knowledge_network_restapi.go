package dependency

import (
	"context"
)

// SubGraphQueryRequest 1. 业务知识网络子图查询请求结构
// 用于查询业务知识网络中的对象子图、拓扑邻居和历史因果关系
type SubGraphQueryRequest struct {
	SourceObjectTypeID string            `json:"source_object_type_id"` // 源对象类型ID（如：service, fault_point）
	Condition          SubGraphCondition `json:"condition"`             // 查询条件（支持简单条件和复合条件）
	Direction          string            `json:"direction"`             // 查询方向：bidirectional（双向）, forward（前向）, backward（后向）
	PathLength         int               `json:"path_length"`           // 路径长度（1：一度邻居，2：二度邻居等）
	NeedTotal          bool              `json:"need_total"`            // 是否需要总数
	Limit              int               `json:"limit"`                 // 限制返回数量
}

// SubGraphCondition 子图查询条件
// 支持两种格式：
// 1. 简单条件：只有 field, operation, value（用于单个字段查询）
// 2. 复合条件：有 operation 和 sub_conditions（用于 or/and 逻辑组合）
type SubGraphCondition struct {
	// 简单条件字段
	// Field     string      `json:"field,omitempty"`     // 字段名（如：s_id, fault_id）
	Operation string `json:"operation,omitempty"` // 操作符：==, !=, >, <, >=, <=, in, not_in 等
	// Value     interface{} `json:"value,omitempty"`     // 字段值

	// 复合条件字段
	SubConditions []SubGraphSubCondition `json:"sub_conditions,omitempty"` // 子条件列表（用于 or/and 逻辑）
}

// SubGraphSubCondition 子图查询子条件
// 用于构建复合查询条件
type SubGraphSubCondition struct {
	Field     string `json:"field"`     // 字段名
	Operation string `json:"operation"` // 操作符：==, !=, >, <, >=, <=, in, not_in 等
	Value     string `json:"value"`     // 字段值
}

// SubGraphQueryResponse 1. 业务知识网络子图查询响应结构
// 包含查询到的对象、关系路径和分页信息
type SubGraphQueryResponse struct {
	Objects           map[string]SubGraphObject `json:"objects"`             // 对象映射表，key 为对象ID（如 "physical_machine-pm_002"）
	RelationPaths     []SubGraphRelationPath    `json:"relation_paths"`      // 关系路径列表
	SearchAfter       []interface{}             `json:"search_after"`        // 搜索游标（用于分页）
	CurrentPathNumber int                       `json:"current_path_number"` // 当前路径编号
	OverallMS         int64                     `json:"overall_ms"`          // 总耗时（毫秒）
}

// SubGraphObject 子图查询返回的对象信息
// 包含对象的唯一标识、类型、显示名称和属性
type SubGraphObject struct {
	ID               string                   `json:"id"`                // 对象ID（如 "physical_machine-pm_002"）
	UniqueIdentities SubGraphUniqueIdentities `json:"unique_identities"` // 唯一标识（包含 s_id）
	ObjectTypeID     string                   `json:"object_type_id"`    // 对象类型ID（如：service, host）
	ObjectTypeName   string                   `json:"object_type_name"`  // 对象类型名称
	Display          string                   `json:"display"`           // 显示名称
	Properties       map[string]interface{}   `json:"properties"`        // 对象属性（包含各种字段，如 ip_address, name, s_id, fault_id 等）
}

// SubGraphUniqueIdentities 对象的唯一标识
type SubGraphUniqueIdentities struct {
	SID string `json:"s_id"` // 对象实体ID（唯一标识）
}

// SubGraphRelationPath 子图查询返回的关系路径
// 包含一条完整的关系路径（可能包含多个关系）
type SubGraphRelationPath struct {
	Relations []SubGraphRelation `json:"relations"` // 关系列表（路径中的每个关系）
	Length    int                `json:"length"`    // 路径长度（关系的数量）
}

// SubGraphRelation 子图查询返回的关系信息
// 表示两个对象之间的关联关系
type SubGraphRelation struct {
	RelationTypeID   string `json:"relation_type_id"`   // 关系类型ID（如：depends_on, has_cause）
	RelationTypeName string `json:"relation_type_name"` // 关系类型名称
	SourceObjectID   string `json:"source_object_id"`   // 源对象ID（如 "physical_machine-pm_002"）
	TargetObjectID   string `json:"target_object_id"`   // 目标对象ID（如 "network_device-net_002"）
}

// ObjectInfoQueryRequest 对象信息查询请求
type ObjectInfoQueryRequest struct {
	Condition  SubGraphCondition `json:"condition"`  // 查询条件（支持简单条件和复合条件）
	NeedTotal  bool              `json:"need_total"` // 是否需要总数
	Limit      int               `json:"limit"`      // 限制返回数量
	Properties []string          `json:"properties"` // 输出字段列表
}

// ObjectInfo 对象信息返回
type ObjectInfoQueryResponse struct {
	Datas           []interface{} `json:"datas"`             // 对象列表
	SearchAfter     []interface{} `json:"search_after"`      // 搜索游标（用于分页）
	SearchFromIndex bool          `json:"search_from_index"` // 是否从索引开始搜索
	OverallMS       int64         `json:"overall_ms"`        // 总耗时（毫秒）
}

//go:generate mockgen -source ./uniquery_restapi.go -destination ../../mock/adapter/restapi/mock_uniquery_restapi.go -package mock
type KnowledgeNetworkClient interface {
	SubGraphQuery(ctx context.Context, queryReq SubGraphQueryRequest, authorization, knowledgeId string) (*SubGraphQueryResponse, error)
	ObjectInfoQuery(ctx context.Context, queryReq ObjectInfoQueryRequest, entityObjectClass, authorization, knowledgeId string) (*ObjectInfoQueryResponse, error)
}
