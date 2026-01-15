package vo

import (
	"database/sql"
	"time"
)

const (
	ValueFrom_Const = "const"
	ValueFrom_Field = "field"
	ValueFrom_User  = "user"
)

type SortParamsV2 struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type ValueOptCfg struct {
	ValueFrom string `json:"value_from,omitempty" mapstructure:"value_from"`
	Value     any    `json:"value,omitempty" mapstructure:"value"`
	RealValue any    `json:"real_value,omitempty" mapstructure:"real_value"`
}

// 数据视图字段
type ViewField struct {
	Name              string       `json:"name" mapstructure:"name"`
	Type              string       `json:"type" mapstructure:"type"`
	Comment           string       `json:"comment" mapstructure:"comment"`
	DisplayName       string       `json:"display_name" mapstructure:"display_name"`
	OriginalName      string       `json:"original_name" mapstructure:"original_name"`
	DataLength        int32        `json:"data_length,omitempty" mapstructure:"data_length"`
	DataAccuracy      int32        `json:"data_accuracy,omitempty" mapstructure:"data_accuracy"`
	Status            string       `json:"status,omitempty" mapstructure:"status"`
	IsNullable        string       `json:"is_nullable,omitempty" mapstructure:"is_nullable"`
	BusinessTimestamp bool         `json:"business_timestamp,omitempty" mapstructure:"business_timestamp"`
	SrcNodeID         string       `json:"src_node_id,omitempty"  mapstructure:"src_node_id"`
	SrcNodeName       string       `json:"src_node_name,omitempty" mapstructure:"src_node_name"`
	PrimaryKey        sql.NullBool `json:"-" mapstructure:"-"`

	Path []string `json:"-" mapstructure:"-"`
}

type CondCfg struct {
	Name        string     `json:"field,omitempty" mapstructure:"field"` // 传递name
	Operation   string     `json:"operation,omitempty" mapstructure:"operation"`
	SubConds    []*CondCfg `json:"sub_conditions,omitempty" mapstructure:"sub_conditions"`
	ValueOptCfg `mapstructure:",squash"`

	RemainCfg map[string]any `mapstructure:",remain"`

	NameField *ViewField `json:"-" mapstructure:"-"`
}

// 行列规则结构体
type DataViewRowColumnRule struct {
	RuleID     string   `json:"id"`
	RuleName   string   `json:"name"`
	ViewID     string   `json:"view_id"`
	Tags       []string `json:"tags"`
	Comment    string   `json:"comment"`
	Fields     []string `json:"fields"`
	RowFilters *CondCfg `json:"row_filters"`
	// CreateTime int64         `json:"create_time"`
	// UpdateTime int64         `json:"update_time"`
	// Creator    string        `json:"creator"`
	// Updater    string        `json:"updater"`

	// 操作权限
	Operations []string `json:"operations"`
}

type SearchAfterParams struct {
	SearchAfter  []any  `json:"search_after"`
	PitID        string `json:"pit_id"`
	PitKeepAlive string `json:"pit_keep_alive"`
}

// 视图数据查询请求体v2
type DataViewQueryV2 struct {
	AllowNonExistField bool `json:"-"`
	IncludeView        bool `json:"-"` // 控制是否返回视图对象，查询参数
	// GlobalFilters      *cond.CondCfg   `json:"filters"`
	GlobalFilters map[string]any  `json:"filters"`
	Sort          []*SortParamsV2 `json:"sort"`
	Timeout       time.Duration   `json:"-"` // 超时时间，查询参数
	ViewQueryCommonParams
	SearchAfterParams

	ActualCondition *CondCfg `json:"-"`
	VegaDurationMs  int64    `json:"-"`
}

// 视图查询公共参数
type ViewQueryCommonParams struct {
	Start          int64                    `json:"start"`
	End            int64                    `json:"end"`
	DateField      string                   `json:"date_field"`
	Offset         int                      `json:"offset"`
	Limit          int                      `json:"limit"`
	Format         string                   `json:"format"`
	NeedTotal      bool                     `json:"need_total"`
	UseSearchAfter bool                     `json:"use_search_after"`
	SqlStr         string                   `json:"sql"`
	RowColumnRules []*DataViewRowColumnRule `json:"row_column_rules"`
	OutputFields   []string                 `json:"output_fields"` // 指定输出的字段列表
}

type RootCauseObjectIdParams struct {
	RootCauseObjectId string `form:"root_cause_object_id" json:"root_cause_object_id" validate:"required"`
	RootCauseFaultID  uint64 `form:"root_cause_fault_id" json:"root_cause_fault_id" validate:"required"`
}
