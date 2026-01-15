package vo

import "time"

type AccountInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// DataScopeNode 表示数据作用域图中的节点
type DataScopeNode struct {
	ID              string                `json:"id"`
	Title           string                `json:"title"`
	Type            string                `json:"type"`
	InputNodes      []string              `json:"input_nodes"`
	Config          map[string]any        `json:"config"`
	OutputFields    []*ViewField          `json:"output_fields"`
	OutputFieldsMap map[string]*ViewField `json:"-"` // 存储输出字段列表（对应metadata全部字段）
}

type ExcelConfig struct {
	SheetName        string `json:"sheet_name"`          // sheet页，逗号分隔
	StartCell        string `json:"start_cell"`          // 起始单元格
	EndCell          string `json:"end_cell"`            // 结束单元格
	HasHeaders       bool   `json:"has_headers"`         // 是否首行作为列名
	SheetAsNewColumn bool   `json:"sheet_as_new_column"` // 是否将sheet作为新列
}

type DataView struct {
	ViewID            string                `json:"id" mapstructure:"id"`
	ViewName          string                `json:"name" mapstructure:"name"`
	TechnicalName     string                `json:"technical_name" mapstructure:"technical_name"`
	GroupID           string                `json:"group_id" mapstructure:"group_id"`
	GroupName         string                `json:"group_name" mapstructure:"group_name"`
	Type              string                `json:"type" binding:"required,oneof=atomic custom" mapstructure:"type"`
	QueryType         string                `json:"query_type" binding:"required,oneof=SQL DSL" mapstructure:"query_type"`
	Tags              []string              `json:"tags" mapstructure:"tags"`
	Comment           string                `json:"comment" mapstructure:"comment"`
	Builtin           bool                  `json:"builtin" mapstructure:"builtin"`
	CreateTime        int64                 `json:"create_time" mapstructure:"create_time"`
	UpdateTime        int64                 `json:"update_time" mapstructure:"update_time"`
	DataSourceType    string                `json:"data_source_type,omitempty" mapstructure:"data_source_type"`
	DataSourceID      string                `json:"data_source_id,omitempty" mapstructure:"data_source_id"`
	DataSourceName    string                `json:"data_source_name,omitempty" mapstructure:"data_source_name"`
	DataSourceCatalog string                `json:"data_source_catalog,omitempty" mapstructure:"data_source_catalog"`
	FileName          string                `json:"file_name,omitempty" mapstructure:"file_name"`
	Status            string                `json:"status,omitempty" mapstructure:"status"`
	Operations        []string              `json:"operations" mapstructure:"operations"`
	Fields            []*ViewField          `json:"fields" mapstructure:"fields"`
	FieldScope        string                `json:"field_scope" mapstructure:"field_scope"`
	FieldsMap         map[string]*ViewField `json:"fields_map" mapstructure:"fields_map"`
	ModuleType        string                `json:"module_type" mapstructure:"module_type"`
	Creator           AccountInfo           `json:"creator" mapstructure:"creator"`
	Updater           AccountInfo           `json:"updater" mapstructure:"updater"`
	DataScope         []*DataScopeNode      `json:"data_scope,omitempty" mapstructure:"data_scope"`
	ExcelConfig       *ExcelConfig          `json:"excel_config,omitempty" mapstructure:"excel_config"`
	MetadataFormID    string                `json:"metadata_form_id,omitempty" mapstructure:"metadata_form_id"`
	PrimaryKeys       []string              `json:"primary_keys" mapstructure:"primary_keys"`
	SQLStr            string                `json:"sql_str,omitempty" mapstructure:"sql_str"`
	MetaTableName     string                `json:"meta_table_name,omitempty" mapstructure:"meta_table_name"`
	DataScopeAdvancedParams
}

// 简单的视图结构，列表查询接口使用
type DataScopeAdvancedParams struct {
	HasDataScopeSQLNode bool `json:"-"` // 是否包含sql节点
	HasStar             bool `json:"-"` // 是否有 *
}

// 视图查询外部接口统一返回结构 V2
type ViewUniResponseV2 struct {
	PitID          string           `json:"pit_id,omitempty"`
	SearchAfter    []any            `json:"search_after,omitempty"`
	View           *DataView        `json:"view,omitempty"`
	Entries        []map[string]any `json:"entries"`
	TotalCount     *int64           `json:"total_count,omitempty"`
	VegaDurationMs int64            `json:"vega_duration_ms,omitempty"`
	OverallMs      int64            `json:"overall_ms,omitempty"`

	// 提供给 v1 接口
	ScrollId string `json:"-"`
}
type BaseResp struct {
	Success uint8 `json:"success"` //枚举: 0,1  枚举备注: 0失败，1成功
}

// RcaContext 分析上下文
// 包含问题现象描述、故障回溯和分析网络
type RcaContextResp struct {
	BackTrace []map[string]any `json:"backtrace"`         // 故障回溯（按故障ID索引）
	Network   RcaNetworkData   `json:"network,omitempty"` // 分析网络（JSON 格式，可选）
}

// 用于 FaultTrace，记录故障点信息
type FaultData struct {
	FaultID           uint64    `json:"fault_id"`                     // 故障点ID
	FaultName         string    `json:"fault_name"`                   // 故障点名称
	FaultCreateTime   time.Time `json:"fault_create_time"`            // 故障点创建时间
	FaultUpdateTime   time.Time `json:"fault_update_time"`            // 故障点更新时间
	FaultStatus       string    `json:"fault_status"`                 // 故障状态（occurred/recovered/expired）
	FaultOccurTime    time.Time `json:"fault_occur_time"`             // 故障发生时间
	FaultLatestTime   time.Time `json:"fault_latest_time"`            // 故障最新时间
	FaultDurationTime int64     `json:"fault_duration_time"`          // 故障持续时间（秒）
	FaultRecoverTime  time.Time `json:"fault_recover_time,omitempty"` // 故障恢复时间（可选）
	EntityObjectClass string    `json:"entity_object_class"`          // 关联对象类型
	EntityObjectName  string    `json:"entity_object_name"`           // 关联对象名称
	EntityObjectID    string    `json:"entity_object_id"`             // 关联对象ID
	RelationEventIDs  []uint64  `json:"relation_event_ids"`           // 关联的事件ID列表
	FaultMode         string    `json:"fault_mode"`                   // 故障模式
	FaultLevel        int       `json:"fault_level"`                  // 故障级别（1-5：紧急/严重/重要/警告/正常）
	FaultDescription  string    `json:"fault_description"`            // 故障描述
}

// RcaNetwork 分析网络
// 包含对象实体、故障点、关系等完整的分析网络结构
// 用于展示根因分析的完整过程和结果
type RcaNetworkData struct {
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
	RelationEventIDs []any `json:"relation_event_ids,omitempty"` // 关联的事件ID列表
	// RelationObjectIDs     []string `json:"relation_object_ids,omitempty"`      // 关联的对象ID列表（拓扑邻居）
	RelationFaultPointIDs []float64 `json:"relation_fault_point_ids,omitempty"` // 关联的故障点ID列表（向后兼容，可选）
	RelationResource      []string  `json:"relation_resource,omitempty"`        // 关联的资源列表
}

// Node 基础节点
// 包含对象实体的基础信息，被 Topology 和 RcaNode 使用
// 支持多种对象类型（service, host, pod, middleware 等）
type Node struct {
	SID               string   `json:"s_id"`                // 对象实体ID（唯一标识）
	SCreateTime       string   `json:"s_create_time"`       // 对象创建时间
	SUpdateTime       string   `json:"s_update_time"`       // 对象更新时间
	Name              string   `json:"name,omitempty"`      // 对象名称
	IPAddress         []string `json:"ip_address"`          // IP地址列表（修正拼写：IPAddress）
	ObjectClass       string   `json:"object_class"`        // 对象类型（如：service, pod, host 等）
	ObjectImpactLevel float64  `json:"object_impact_level"` // 对象影响级别（1-5：紧急/严重/重要/警告/正常）
}

// Relation 关系
// 用于拓扑图和分析网络，表示对象之间的关联关系
type Relation struct {
	RelationID    string `json:"relation_id"`      // 关系ID（唯一标识）
	RelationClass string `json:"relation_class"`   // 关系类型ID（如：depends_on, connects_to 等）
	SourceSID     string `json:"source_object_id"` // 源对象实体ID
	TargetSID     string `json:"target_object_id"` // 目标对象实体ID
}
