package domain

import (
	"time"
)

// RCARequest 异步发送给 RCA 模块。
type RCARequest struct {
	ProblemID uint64 `json:"problem_id"` // 问题ID
}

// RCACallback 携带 RCA 结果回传 Problem 模块。
type RCACallback struct {
	ProblemID          uint64    `json:"problem_id"`                     // 问题ID
	RootCauseObjectID  string    `json:"root_cause_object_id,omitempty"` // 根因对象ID，可为空
	RootCauseFaultID   uint64    `json:"root_cause_fault_id,omitempty"`  // 根因故障点ID，可为空
	RcaResults         string    `json:"rca_results"`                    // 分析结果详情（原 RCAResults）
	RcaStartTime       time.Time `json:"rca_start_time"`                 // 分析开始时间
	RcaEndTime         time.Time `json:"rca_end_time"`                   // 分析结束时间
	RcaStatus          RcaStatus `json:"rca_status"`                     // 分析状态
	InProgress         bool      `json:"in_progress"`                    // 是否正在进行中
	ProblemName        string    `json:"problem_name"`                   // 问题名称
	ProblemDescription string    `json:"problem_description"`            // 问题详细描述
}

// ProblemCreatedEvent 问题创建事件
type ProblemCreatedEvent struct {
	ProblemID uint64
}

const (
	SourceZabbixWebhook = "zabbix_webhook"
)

// RcaResults 分析结果详情
// 包含完整分析上下文和网络信息
type RcaResults struct {
	AdpKnID    string     `json:"adp_kn_id"`   // 知识网络ID
	RcaID      string     `json:"rca_id"`      // 分析ID
	RcaContext RcaContext `json:"rca_context"` // 分析上下文
}

// RcaContext 分析上下文
// 包含问题现象描述、故障回溯和分析网络
type RcaContext struct {
	Occurrence Occurrence `json:"occurrence"`        // 问题过程（事实层）
	BackTrace  []Fault    `json:"backtrace"`         // 故障回溯（按故障ID索引）
	Network    RcaNetwork `json:"network,omitempty"` // 分析网络（JSON 格式，可选）
}

// Occurrence 问题现象描述
// 包含问题的过程描述和影响说明
type Occurrence struct {
	Name        string `json:"name"`        // 问题名称
	Description string `json:"description"` // 问题发生过程描述
	Impact      string `json:"impact"`      // 问题发生过程影响
}

// 用于 FaultTrace，记录故障点信息
type Fault struct {
	FaultID           uint64      `json:"fault_id"`                     // 故障点ID
	FaultName         string      `json:"fault_name"`                   // 故障点名称
	FaultCreateTime   time.Time   `json:"fault_create_time"`            // 故障点创建时间
	FaultUpdateTime   time.Time   `json:"fault_update_time"`            // 故障点更新时间
	FaultStatus       FaultStatus `json:"fault_status"`                 // 故障状态（occurred/recovered/expired）
	FaultOccurTime    time.Time   `json:"fault_occur_time"`             // 故障发生时间
	FaultLatestTime   time.Time   `json:"fault_latest_time"`            // 故障最新时间
	FaultDurationTime int64       `json:"fault_duration_time"`          // 故障持续时间（秒）
	FaultRecoverTime  time.Time   `json:"fault_recover_time,omitempty"` // 故障恢复时间（可选）
	EntityObjectClass string      `json:"entity_object_class"`          // 关联对象类型
	EntityObjectName  string      `json:"entity_object_name"`           // 关联对象名称
	EntityObjectID    string      `json:"entity_object_id"`             // 关联对象ID
	RelationEventIDs  []uint64    `json:"relation_event_ids"`           // 关联的事件ID列表
	FaultMode         string      `json:"fault_mode"`                   // 故障模式
	FaultLevel        Severity    `json:"fault_level"`                  // 故障级别（1-5：紧急/严重/重要/警告/正常）
	FaultDescription  string      `json:"fault_description"`            // 故障描述
}

// RcaNetwork 分析网络
// 包含对象实体、故障点、关系等完整的分析网络结构
// 用于展示根因分析的完整过程和结果
type RcaNetwork struct {
	Nodes []RcaNode  `json:"nodes"` // 分析节点列表（包含对象节点和故障点节点）
	Edges []Relation `json:"edges"` // 关系列表（包含拓扑关系和因果关系）
}

// RcaNode 分析节点
// 嵌入基础节点（Node），并包含对象实体节点的关联信息
// 一个节点可以关联多个故障点、事件和对象
type RcaNode struct {
	// 嵌入基础节点（对象实体字段）
	Node `json:",inline"`

	// 对象实体节点的关联信息
	RelationEventIDs []string `json:"relation_event_ids,omitempty"` // 关联的事件ID列表
	// RelationObjectIDs     []string `json:"relation_object_ids,omitempty"`      // 关联的对象ID列表（拓扑邻居）
	RelationFaultPointIDs []uint64 `json:"relation_fault_point_ids,omitempty"` // 关联的故障点ID列表（向后兼容，可选）
	RelationResource      []string `json:"relation_resource,omitempty"`        // 关联的资源列表
}

// Node 基础节点
// 包含对象实体的基础信息，被 Topology 和 RcaNode 使用
// 支持多种对象类型（service, host, pod, middleware 等）
type Node struct {
	SID               string   `json:"s_id,omitempty"`                // 对象实体ID（唯一标识）
	SCreateTime       string   `json:"s_create_time,omitempty"`       // 对象创建时间
	SUpdateTime       string   `json:"s_update_time,omitempty"`       // 对象更新时间
	Name              string   `json:"name,omitempty"`                // 对象名称
	IPAddress         []string `json:"ip_address,omitempty"`          // IP地址列表（修正拼写：IPAddress）
	ObjectClass       string   `json:"object_class,omitempty"`        // 对象类型（如：service, pod, host 等）
	ObjectImpactLevel int      `json:"object_impact_level,omitempty"` // 对象影响级别（1-5：紧急/严重/重要/警告/正常）
}

// Relation 关系
// 用于拓扑图和分析网络，表示对象之间的关联关系
type Relation struct {
	RelationID    string `json:"relation_id"`      // 关系ID（唯一标识）
	RelationClass string `json:"relation_class"`   // 关系类型ID（如：depends_on, connects_to 等）
	SourceSID     string `json:"source_object_id"` // 源对象实体ID
	TargetSID     string `json:"target_object_id"` // 目标对象实体ID
}

// ========== 业务知识网络结构 ==========

// Topology 拓扑图
type Topology struct {
	Nodes []Node     `json:"nodes"` // 节点列表（对象实体）
	Edges []Relation `json:"edges"` // 边列表（对象之间的关系）
}

// ========== 业务知识网络查询结构 ==========

// SubGraphQueryRequest 1. 业务知识网络子图查询请求结构
// 用于查询业务知识网络中的对象子图、拓扑邻居和历史因果关系
type SubGraphQueryRequest struct {
	SourceObjectTypeID string             `json:"source_object_type_id"` // 源对象类型ID（如：service, fault_point）
	Condition          *SubGraphCondition `json:"condition"`             // 查询条件（支持简单条件和复合条件）
	Direction          string             `json:"direction"`             // 查询方向：bidirectional（双向）, forward（前向）, backward（后向）
	PathLength         int                `json:"path_length"`           // 路径长度（1：一度邻居，2：二度邻居等）
	NeedTotal          bool               `json:"need_total"`            // 是否需要总数
	Limit              int                `json:"limit"`                 // 限制返回数量
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

// SubGraphSort 子图查询排序规则
type SubGraphSort struct {
	Field     string `json:"field"`     // 排序字段
	Direction string `json:"direction"` // 排序方向：asc（升序）, desc（降序）
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

// SubGraphRelationPath 子图查询返回的关系路径
// 包含一条完整的关系路径（可能包含多个关系）
type SubGraphRelationPath struct {
	Relations []SubGraphRelation `json:"relations"` // 关系列表（路径中的每个关系）
	Length    int                `json:"length"`    // 路径长度（关系的数量）
}

// SubGraphUniqueIdentities 对象的唯一标识
type SubGraphUniqueIdentities struct {
	SID string `json:"s_id"` // 对象实体ID（唯一标识）
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
	Condition  *SubGraphCondition `json:"condition"`  // 查询条件（支持简单条件和复合条件）
	NeedTotal  bool               `json:"need_total"` // 是否需要总数
	Limit      int                `json:"limit"`      // 限制返回数量
	Properties []string           `json:"properties"` // 输出字段列表
}

// ObjectInfo 对象信息返回
type ObjectInfoQueryResponse struct {
	Datas           []interface{} `json:"datas"`             // 对象列表
	SearchAfter     []interface{} `json:"search_after"`      // 搜索游标（用于分页）
	SearchFromIndex bool          `json:"search_from_index"` // 是否从索引开始搜索
	OverallMS       int64         `json:"overall_ms"`        // 总耗时（毫秒）
}

// AI Agent 相关

// AgentRequest Agent 请求结构（内部使用）
// 用于调用 AI Agent 进行因果推理和描述生成
type AgentRequest struct {
	AgentKey     string                 `json:"agent_key"`     // Agent 密钥
	CustomQuerys map[string]interface{} `json:"custom_querys"` // 自定义查询参数（包含故障点信息等）
	Query        string                 `json:"query"`         // 查询类型（causal_Rca 或 description）
	Stream       bool                   `json:"stream"`        // 是否使用流式响应
}

// AgentResponse Agent 响应结构（内部使用）
// 用于解析 AI Agent 的响应
type AgentResponse struct {
	Message struct {
		Content struct {
			FinalAnswer struct {
				Answer struct {
					Text string `json:"text"` // 响应文本内容
				} `json:"answer"`
			} `json:"final_answer"`
		} `json:"content"`
	} `json:"message"`
}

// AgentCausalPayload Agent 返回的因果负载（内部使用）
// 用于解析 AI Agent 返回的因果关系数据
type AgentCausalPayload struct {
	FaultCausal AgentCausalEdge `json:"fault_causal"` // 因果关系对象
}

// AgentCausalEdge Agent 返回的因果边
// 表示 AI Agent 分析得出的两个故障点之间的因果关系
type AgentCausalEdge struct {
	Source     uint64  `json:"source_id"`  // 产生影响的故障点ID（原因）
	Target     uint64  `json:"target_id"`  // 受影响/被导致的故障点ID（结果）
	Confidence float64 `json:"confidence"` // 关系置信度（0.0-1.0）
	Reason     string  `json:"reason"`     // 因果关系描述（说明为什么存在这种关系）
}

// AgentDescriptionPayload Agent 返回的描述负载（内部使用）
// 用于解析 AI Agent 返回的问题描述数据
type AgentDescriptionPayload struct {
	Occurrence Occurrence `json:"occurrence"` // 问题现象描述
}

// ========== 因果推理数据结构 ==========

// CausalAnalysisResults 因果分析结果
// 包含完整的因果分析结果，用于分析结果回调
type CausalAnalysisResults struct {
	CausalRelations      []CausalCandidate     `json:"causal_relations"`               // 因果关系候选列表
	FaultCausals         []FaultCausalObject   `json:"fault_causals"`                  // 故障因果实体列表
	FaultCausalRelations []FaultCausalRelation `json:"fault_causal_relations"`         // 故障因果关系列表
	RootCauseObjectID    string                `json:"root_cause_object_id,omitempty"` // 根因对象ID（可为空）
	RootCauseFaultID     uint64                `json:"root_cause_fault_id,omitempty"`  // 根因故障点ID（可为空）
}

// CausalCandidate 因果候选
// 表示两个故障点之间可能存在的因果关系（待确认和存储）
type CausalCandidate struct {
	Cause      *FaultPointObject `json:"cause"`      // 原因故障点
	Effect     *FaultPointObject `json:"effect"`     // 结果故障点
	Confidence float64           `json:"confidence"` // 置信度（0.0-1.0）
	Reason     string            `json:"reason"`     // 原因描述
	IsNew      bool              `json:"is_new"`     // 是否为新建的因果关系（相对于历史数据）
}

// CausalRelation 历史因果关系
// 从业务知识网络中查询到的历史因果关系信息
type CausalRelation struct {
	CausalID        string    `json:"causal_id"`        // 因果实体ID
	CauseObjectID   string    `json:"cause_object_id"`  // 原因对象ID
	EffectObjectID  string    `json:"effect_object_id"` // 结果对象ID
	Confidence      float64   `json:"confidence"`       // 置信度（0.0-1.0）
	OccurrenceCount int       `json:"occurrence_count"` // 发生次数
	LastOccurrence  time.Time `json:"last_occurrence"`  // 最后发生时间
	Reason          string    `json:"reason"`           // 原因描述
}

// GraphRecallContext 图召回上下文
// 存储从业务知识网络中召回的各种数据，用于后续的因果分析和根因定位
type GraphRecallContext struct {
	// 注意：子图仅包含查询对象本身，不包含邻居节点
	TopologySubgraphs map[string]*Topology `json:"topology_subgraphs"` // 包含故障点关联对象及其直接关系的拓扑子图
	// 用于后续查询邻居对象的历史故障点
	TopologyNeighbors map[string][]string `json:"topology_neighbors"` // 存储每个对象的一度拓扑邻居ID列表
	// 用于提升因果推理的置信度
	HistoricalCausality map[string][]CausalRelation `json:"historical_causality"` // 存储每个对象的历史因果关系列表
	// 用于扩展因果分析的范围
	HistoricalNeighborFaultPoints []FaultPointObject `json:"historical_neighbor_fault_points"` // 存储一度拓扑邻居对象在指定时间窗口内发生的历史故障点
	// 用于构建最终的分析结果
	AnalysisNetwork []*RcaNetwork `json:"analysis_network"` // 完整的分析网络，包含所有对象节点、故障点节点及其关系
}
