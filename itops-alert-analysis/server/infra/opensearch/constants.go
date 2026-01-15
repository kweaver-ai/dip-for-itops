package opensearch

// 基础索引名称
const (
	RawEventIndexBase            = "itops_raw_event"
	FaultPointIndexObjectBase    = "itops_fault_point_object"
	FaultPointRelationIndexBase  = "itops_fault_point_relation"
	ProblemIndexBase             = "itops_problem"
	faultCausalObjectIndexBase   = "itops_fault_causal"
	faultCausalRelationIndexBase = "itops_fault_causal_relation"

	maxQuerySize = 5000
	indexPrefix  = "mdl-"
)

// 实际索引名称
var (
	RawEventIndex            = indexPrefix + RawEventIndexBase
	FaultPointIndexObject    = indexPrefix + FaultPointIndexObjectBase
	FaultPointRelationIndex  = indexPrefix + FaultPointRelationIndexBase
	ProblemIndex             = indexPrefix + ProblemIndexBase
	faultCausalObjectIndex   = indexPrefix + faultCausalObjectIndexBase
	faultCausalRelationIndex = indexPrefix + faultCausalRelationIndexBase
)
