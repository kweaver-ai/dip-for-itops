package rca

import (
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
)

const (
	NeighborFaultWindow         = 12 * time.Hour    // 邻居故障点召回时间窗口
	confidenceWeightMultiplier  = 10.0              //  置信度权重倍数
	maxDurationScore            = 10.0              // 持续时间最大分数
	baseConfidence              = 0.3               //基础置信度
	minConfidence               = 0.1               //最小置信度
	maxConfidence               = 0.9               //  最大置信度
	veryShortTimeThreshold      = 5 * time.Minute   // 用于判断两个故障点之间的时间间隔是否很短
	shortTimeThreshold          = 30 * time.Minute  // 用于判断两个故障点之间的时间间隔是否较短（≤30分钟）
	longTimeThreshold           = time.Hour         // 用于判断两个故障点之间的时间间隔是否较长（≤1小时）
	maxHistoricalBoost          = 0.2               // 用于置信度计算，限制历史因果关系对最终置信度的最大贡献
	relationClassHasCause       = "has_cause"       // 表示"有原因"关系，用于连接故障点和因果推理实体
	relationClassHasEffect      = "has_effect"      // 表示"有结果"关系，用于连接故障点和因果推理实体
	faultCausalClass            = "FaultCausal"     // 因果推理实体类型，用于表示两个故障点之间的因果关系
	RCAProblemProcessingTimeout = 120 * time.Minute // RCA 问题处理超时时间
)

// ========== agent analysis 相关常量定义 ==========
const (
	// agentCallTimeout Agent 调用超时时间（单个故障点对的因果推理）
	// 大模型接口调用有时间消耗，设置 15 分钟超时以避免长时间阻塞
	agentCallTimeout = 15 * time.Minute // 15分钟

	// 并发控制常量
	maxConcurrentAnalysis = 15 // 最大并发分析数量（避免过多并发导致资源耗尽）

	// 性能优化常量（保守策略，保证准确性）
	// 注意：不设置最大分析对数限制，确保所有有拓扑关联的故障点对都被分析
	// maxTimeWindowForAnalysis = 2 * time.Hour // 放宽时间窗口到24小时，避免遗漏长期因果关系

	// Token 长度限制（大模型支持 32K tokens，预留 12K 给 prompt，20K 给输入数据）
	maxInputTokens     = 28000 // 最大输入 token 数（预留安全边界）
	estimatedTokenSize = 4     // 每个字符大约对应 4 个 token（中文和英文混合）
	// maxTopologyNodes   = 80    // 拓扑子图最大节点数（20K tokens 支持约 80 个节点）
	// maxTopologyEdges   = 150   // 拓扑子图最大边数（20K tokens 支持约 150 条边）

	// 时间间隔相关的置信度调整
	confidenceVeryShortTime = 0.2  // 很短时间间隔（≤5分钟）的置信度调整
	confidenceShortTime     = 0.15 // 较短时间间隔（≤30分钟）的置信度调整
	confidenceLongTime      = 0.1  // 较长时间间隔（≤1小时）的置信度调整
	confidenceVeryLongTime  = -0.1 // 很长时间间隔（>1小时）的置信度调整
	// 持续时间相关的置信度调整
	confidenceLongerDuration = 0.1  // 原因故障持续时间更长的置信度调整
	confidenceNormalDuration = 0.05 // 原因故障有持续时间的置信度调整
	// 状态相关的置信度调整
	confidenceRecoveredStatus = -0.1 // 原因故障已恢复但结果故障未恢复的置信度调整
	// 历史因果关系相关的置信度调整
	confidencePerHistoricalOccurrence = 0.05 // 每次历史出现增加 0.05 的置信度调整
	// Agent 请求字段名（用于构建发送给 Agent 的自定义查询参数）
	agentRequestFieldFaultPointA      = "faultPointA"       // 故障点A字段名
	agentRequestFieldFaultPointB      = "faultPointB"       // 故障点B字段名
	agentRequestFieldTopologyRelation = "topologyRelation"  // 拓扑关系字段名
	agentRequestFieldEntityAID        = "entity_a_id"       // 对象实体A ID字段名
	agentRequestFieldEntityBID        = "entity_b_id"       // 对象实体B ID字段名
	agentRequestFieldTopologySubgraph = "topology_subgraph" // 拓扑子图字段名
)

// ========== build_rcadata 相关常量定义 ==========

const (
	// 日期时间格式（用于格式化故障发生时间）
	timeFormatDateTime = "2006-01-02 15:04:05"
	// Agent 自定义查询中的问题信息键
	agentCustomQueryKeyProblemInfo = "problem_info"

	// Token 长度限制（大模型支持 32K tokens）
	maxFaultPointsForSummary = 60 // 最多发送给 Agent 的故障点数量（避免 token 超限）

	defaultNameNoFaultPoints = "未知问题"
	// 无故障点时的默认描述
	defaultDescriptionNoFaultPoints = "当前上下文中未识别出明确的故障点或异常模式"
	// 无故障点时的默认影响描述
	defaultImpactNoFaultPoints = "暂未识别明确业务影响，需进一步观测或补充信息"

	// 默认描述模板（包含故障点数量和时间）
	defaultDescriptionWithTime = "当前问题关联到 %d 个故障点，最早发生于 %s"
	// 默认影响描述模板（包含实体数量）
	defaultImpactTemplate = "影响 %d 个实体对象"
	// 正常级别（5）
	// ObjectImpactLevel 的值：1 紧急 2 严重 3 重要 4 警告 5 正常
	objectImpactLevelNormal = 5
)

// ========== causal analysis 相关常量定义 ==========

const (
	// causalIDFormat 因果推理实体ID格式
	// 格式：causal_{id}
	// 使用 ID 生成器生成唯一ID，每次调用都会生成新的ID
	causalIDFormat = "causal_%d"

	// relationIDFormatFaultCausalRelation 故障因果关系ID格式
	// 格式：relation_{id}
	// 使用 ID 生成器生成唯一ID，每次调用都会生成新的ID
	relationIDFormatFaultCausalRelation = "relation_%d"
)

// faultTimeline 故障时间线
// 按发生时间排序的故障点列表，用于根因分析时按时间顺序处理故障点
type faultTimeline []*domain.FaultPointObject

// causalGraph 因果关系图
// 用于构建和分析故障点之间的因果关系网络
// 提供双向映射（原因->结果、结果->原因）和置信度存储
type causalGraph struct {
	// effectToCauses 结果到原因的映射（effect -> []cause）
	// 用于反向追溯：从结果故障点找到所有可能的原因
	effectToCauses map[string][]*domain.FaultPointObject

	// causeToEffects 原因到结果的映射（cause -> []effect）
	// 用于正向扩散：从原因故障点找到所有可能的结果
	causeToEffects map[string][]*domain.FaultPointObject

	// causalConfidenceMap 因果关系置信度映射
	// 键格式：causeID->effectID，值：置信度（0-1）
	// 存储每对因果关系之间的置信度
	causalConfidenceMap map[string]float64

	// faultPointMap 故障点映射（faultID -> FaultPoint）
	// 用于快速查找故障点对象
	faultPointMap map[string]*domain.FaultPointObject

	// effectFaultIDs 作为结果的故障点ID集合
	// 用于快速判断某个故障点是否作为其他故障点的结果
	effectFaultIDs map[string]bool

	// causeFaultIDs 作为原因的故障点ID集合
	// 用于快速判断某个故障点是否作为其他故障点的原因
	causeFaultIDs map[string]bool
}

// ========== fault causal store 相关常量定义 ==========
const (
	// // relationTypeHasCause has_cause 关系类型（表示"有原因"）
	// relationTypeHasCause = "has_cause"
	// // relationTypeHasEffect has_effect 关系类型（表示"有结果"）
	// relationTypeHasEffect = "has_effect"
	// // faultCausalObjectTypeID FaultCausal 对象类型ID
	// faultCausalObjectTypeID = "fault_causal"
	// // faultCausalObjectTypeName FaultCausal 对象类型名称
	// faultCausalObjectTypeName = "fault_causal"
	// propertyKeySID s_id 属性键
	propertyKeySID = "s_id"
	// propertyKeyName name 属性键
	propertyKeyName = "name"
	// propertyKeyIPAddress ip_address 属性键
	propertyKeyIPAddress = "ip_address"
	// propertyKeySCreateTime s_create_time 属性键
	propertyKeySCreateTime = "s_create_time"
	// propertyKeySUpdateTime s_update_time 属性键
	propertyKeySUpdateTime = "s_update_time"
	// // propertyKeyFaultID fault_id 属性键
	// propertyKeyFaultID = "fault_id"
	// // propertyKeyCausalConfidence causal_confidence 属性键
	// propertyKeyCausalConfidence = "causal_confidence"
	// // propertyKeyCausalReason causal_reason 属性键
	// propertyKeyCausalReason = "causal_reason"
	// // propertyKeyOccurrenceCount occurrence_count 属性键
	// propertyKeyOccurrenceCount = "occurrence_count"
	// // propertyKeyLastOccurrence last_occurrence 属性键
	// propertyKeyLastOccurrence = "last_occurrence"
	// // propertyKeyCausalUpdateTime causal_update_time 属性键
	// propertyKeyCausalUpdateTime = "causal_update_time"
)

// 对象类 ID 常量
const (
	// 对象类 ID为：service
	ServiceObjectClassID = "service"
	// 对象类 ID为：host
	HostObjectClassID = "host"
	// 对象类 ID为：pod
	PodObjectClassID = "pod"
	// 对象类 ID为：middleware
	MiddlewareObjectClassID = "middleware"
	// 对象类 ID为：database
	DatabaseObjectClassID = "database"
	// 对象类 ID为：physical_machine
	PhysicalMachineObjectClassID = "physical_machine"
	// 对象类 ID为：network_device
	NetworkDeviceObjectClassID = "network_device"
	// 对象类 ID为：fault_point
	FaultPointObjectClassID = "fault_point_object"
	// 对象类 ID为：fault_causal
	FaultCausalObjectClassID = "fault_causal"
)

// ObjectClasses 允许的对象类列表（用于过滤拓扑邻居）
var ObjectClasses = []string{
	ServiceObjectClassID,
	HostObjectClassID,
	PodObjectClassID,
	MiddlewareObjectClassID,
	DatabaseObjectClassID,
	PhysicalMachineObjectClassID,
	NetworkDeviceObjectClassID,
}

// ========== root cause 相关常量定义 ==========

const (
	// rootCauseScoreInitialMax 初始最大分数（用于比较）
	// 设置为 -1.0，确保任何有效的分数都会大于初始值
	rootCauseScoreInitialMax = -1.0

	// 时间评分相关常量
	timeScorePerLaterFault = 1.0  // 每早于一个故障点的时间分数
	timeScorePerHourEarly  = 0.1  // 每早于1小时的时间分数（额外奖励）
	maxTimeScore           = 20.0 // 时间分数最大值

	// 持续时间评分相关常量
	secondsPerHour = 3600.0 // 秒转小时的除数

	// 严重程度评分相关常量
	severityScoreWeight = 5.0  // 严重程度权重（值越小越严重，分数越高）
	maxSeverityScore    = 10.0 // 严重程度最大分数

	// 状态评分相关常量
	recoveredStatusPenalty = -5.0 // 已恢复故障的惩罚分数
	occurredStatusBonus    = 2.0  // 未恢复故障的奖励分数

	// 因果关系评分相关常量
	causalRelationKeyFormat = "%s->%s" // 因果关系键格式
)

// ========== service 相关常量定义 ==========

const (
	MaxConcurrentRCA = 10 // MaxConcurrentRCA 最大并发 RCA 处理数量
)

// newCausalGraph 创建并初始化因果关系图
// 返回一个所有字段都已初始化的因果关系图实例
func newCausalGraph() *causalGraph {
	return &causalGraph{
		effectToCauses:      make(map[string][]*domain.FaultPointObject),
		causeToEffects:      make(map[string][]*domain.FaultPointObject),
		causalConfidenceMap: make(map[string]float64),
		faultPointMap:       make(map[string]*domain.FaultPointObject),
		effectFaultIDs:      make(map[string]bool),
		causeFaultIDs:       make(map[string]bool),
	}
}
