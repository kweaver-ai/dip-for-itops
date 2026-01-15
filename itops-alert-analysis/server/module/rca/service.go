package rca

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/dip"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/kafka"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/idgen"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/timex"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// ProblemEventMessage 问题事件消息结构
type ProblemEventMessage struct {
	ProblemID uint64 `json:"problem_id"`
}

// Service 接收 ProblemID 做 RCA，并通过回调异步返回结果。
// 采用批次处理模式：每 5 分钟收集一批 problem_id，去重后并发处理。
type Service struct {
	config        config.Config
	dipClient     *dip.Client
	idGenerator   *idgen.Generator // ID 生成器（保证全局唯一）
	callback      core.ProblemHandler
	kafkaConsumer core.KafkaConsumer
	repoFactory   *opensearch.RepositoryFactory

	// 批次处理配置
	batchWindow   time.Duration //批次处理窗口时间
	maxConcurrent int           //处理并发数

	mu           sync.Mutex                    // 保护 collected 和 runningTasks（只使用写锁，改为 Mutex）
	collected    map[uint64]struct{}           // 收集到的问题 ID
	runningTasks map[uint64]context.CancelFunc // 正在运行的任务
}

func New(
	config config.Config,
	dipClient *dip.Client,
	idGenerator *idgen.Generator,
	callback core.ProblemHandler,
	repoFactory *opensearch.RepositoryFactory,
) (*Service, error) {
	// 创建 RCA 专用的 Kafka Consumer，消费问题事件流
	rcaConsumer, err := kafka.NewConsumer(kafka.Config{
		Brokers: []string{fmt.Sprintf("%s:%d", config.DepServices.MQ.MQHost, config.DepServices.MQ.MQPort)},
		SASL: &kafka.SASLConfig{
			Enabled:  true,
			Username: config.DepServices.MQ.Auth.Username,
			Password: config.DepServices.MQ.Auth.Password,
		},
		Topic:   config.Kafka.ProblemEvents.Topic,
		GroupID: config.Kafka.ProblemEvents.ConsumerGroup,
	})
	if err != nil {
		return nil, errors.Wrap(err, "初始化 RCA Kafka Consumer 失败")
	}

	return &Service{
		config:        config,
		dipClient:     dipClient,
		idGenerator:   idGenerator,
		callback:      callback,
		kafkaConsumer: rcaConsumer,
		repoFactory:   repoFactory,
		batchWindow:   5 * time.Minute,
		maxConcurrent: MaxConcurrentRCA,
		collected:     make(map[uint64]struct{}),
		runningTasks:  make(map[uint64]context.CancelFunc),
	}, nil
}

// Notify 模拟 RCA 异步回调 Problem 模块。
func (s *Service) Notify(ctx context.Context, cb domain.RCACallback) error {
	if s.callback == nil {
		return errors.New("rca callback handler not wired")
	}
	return s.callback.HandleRCACallback(ctx, cb)
}

// Start 启动 RCA 服务，从 Kafka 消费问题事件并处理
func (s *Service) Start(ctx context.Context) error {
	if s.kafkaConsumer == nil {
		return errors.New("kafka consumer not configured")
	}

	log.Infof("RCA Service 启动 - 批次窗口: %v, 最大并发: %d", s.batchWindow, s.maxConcurrent)

	ticker := time.NewTicker(s.batchWindow)
	defer ticker.Stop()

	// 启动 Kafka 消费协程
	go func() {
		log.Infof("RCA Kafka 消费协程启动，开始监听问题事件")
		handler := func(msgCtx context.Context, msg core.KafkaMessage) error {
			var event ProblemEventMessage
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Debugf("RCA 解析 Kafka 消息失败: %v", err)
				return nil
			}
			if event.ProblemID == 0 {
				return errors.New(fmt.Sprintf("RCA 消息内容不合法: %+v", utils.JsonEncode(event)))
			}
			s.mu.Lock()
			s.collected[event.ProblemID] = struct{}{}
			collectedCount := len(s.collected)
			s.mu.Unlock()
			log.Debugf("====== RCA收到问题事件 ProblemID: %d, 当前收集数: %d ======", event.ProblemID, collectedCount)
			return nil
		}
		if err := s.kafkaConsumer.ConsumeRawEvents(ctx, handler); err != nil {
			log.Errorf("RCA Kafka 消费失败: %+v", err)
		}
	}()

	// 批次处理循环
	for {
		select {
		case <-ctx.Done():
			log.Infof("RCA Service 收到停止信号")
			return nil

		case <-ticker.C:
			s.mu.Lock()
			collectedCount := len(s.collected)
			log.Infof("RCA 窗口时间到，当前收集到 %d 个问题", collectedCount)
			if len(s.collected) == 0 {
				log.Debugf("RCA 本批次无问题需要处理，跳过")
				s.mu.Unlock()
				continue
			}

			// 转为切片
			problemIDs := make([]uint64, 0, len(s.collected))
			for pid := range s.collected {
				problemIDs = append(problemIDs, pid)
			}
			s.collected = make(map[uint64]struct{})
			s.mu.Unlock()

			log.Infof("RCA 批次窗口触发，开始处理 %d 个问题，问题 IDs: %v", len(problemIDs), problemIDs)

			// 并发处理
			batchStart := time.Now()
			s.processBatch(ctx, problemIDs)
			log.Infof("RCA 批次处理完成，耗时: %v", time.Since(batchStart))
		}
	}
}

func (s *Service) Close() error {
	log.Infof("RCA Service 正在关闭")

	var errs []error

	// 关闭 Kafka 消费者
	if s.kafkaConsumer != nil {
		if err := s.kafkaConsumer.Close(); err != nil {
			log.Errorf("RCA 关闭 Kafka 消费者失败: %v", err)
			errs = append(errs, errors.Wrap(err, "RCA 关闭 Kafka 消费者失败"))
		}
	}

	// 如果有多个错误，
	if len(errs) > 0 {
		return errors.New(fmt.Sprintf("RCA 关闭 RCA Service 时发生 %d 个错误: %v", len(errs), errs))
	}

	log.Infof("RCA Service 关闭成功")
	return nil
}

// processBatch 并发处理一批问题
// 如果某个 problemID 已经在运行中，会先取消旧任务再启动新任务
func (s *Service) processBatch(ctx context.Context, problemIDs []uint64) {
	log.Debugf("RCA 初始化 errgroup，最大并发数: %d", s.maxConcurrent)
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.maxConcurrent)

	// 使用 atomic 操作计数，避免额外的锁
	var successCount, failCount, canceledCount int64

	for _, pid := range problemIDs {
		problemID := pid

		// 检查是否已有该 problemID 在运行（使用统一的锁）
		s.mu.Lock()
		if cancelFunc, exists := s.runningTasks[problemID]; exists {
			log.Infof("RCA 问题 %d 已在运行中，取消旧任务并启动新任务", problemID)
			cancelFunc() // 取消旧任务
			atomic.AddInt64(&canceledCount, 1)
		}
		s.mu.Unlock()

		g.Go(func() error {
			log.Debugf("RCA 启动 goroutine 处理问题 %d", problemID)

			// 为该任务创建独立的 context，可以单独取消
			rcaCtx, cancel := context.WithTimeout(gctx, RCAProblemProcessingTimeout)

			// 注册到运行中任务（使用统一的锁）
			s.mu.Lock()
			s.runningTasks[problemID] = cancel
			s.mu.Unlock()

			// 任务完成后清理
			defer func() {
				cancel()
				s.mu.Lock()
				delete(s.runningTasks, problemID)
				s.mu.Unlock()
			}()
			analysisCallback, err := s.Submit(rcaCtx, domain.RCARequest{ProblemID: problemID})
			if err != nil {
				log.Errorf("RCA Service: 处理失败 problem_id=%d: %v", problemID, err)
				if errors.Is(err, context.Canceled) {
					log.Infof("RCA 问题 %d 任务被取消", problemID)
					return nil
				}
				atomic.AddInt64(&failCount, 1)
				return nil // 错误已记录，继续处理其他任务
			}
			// 记录处理结果
			if analysisCallback != nil {
				log.Infof("RCA Service: 发送  RCA 回调请求 problem_id=%d, 状态: %d", problemID, analysisCallback.RcaStatus)
				// Step5: 发送回调给 Problem 模块（使用 rcaCtx 保持上下文一致性）
				if err := s.Notify(rcaCtx, *analysisCallback); err != nil {
					log.Errorf("RCA Service: 发送 RCA 回调失败，问题 ID: %d, 状态: %d, 错误: %v", problemID, analysisCallback.RcaStatus, err)
					return nil
				}
				atomic.AddInt64(&successCount, 1)
			} else {
				log.Warnf("RCA Service: RCA 请求返回空结果 problem_id=%d", problemID)
				atomic.AddInt64(&failCount, 1)
			}
			return nil
		})
	}
	// 等待所有任务完成
	if err := g.Wait(); err != nil {
		log.Errorf("RCA 批次处理过程中发生错误: %v", err)
	}

	// 记录批次处理统计信息
	log.Infof("RCA 批次处理统计 - 成功: %d, 失败: %d, 取消: %d",
		atomic.LoadInt64(&successCount),
		atomic.LoadInt64(&failCount),
		atomic.LoadInt64(&canceledCount))

}

// Submit 接收 ProblemID，触发 RCA 处理
func (s *Service) Submit(ctx context.Context, req domain.RCARequest) (*domain.RCACallback, error) {
	startTime := time.Now()
	problemID := req.ProblemID
	// 参数验证
	if problemID == 0 {
		return s.createFailedCallback(problemID, startTime), errors.New("problem ID 不能为 0")
	}

	log.Infof("========== RCA 开始分析, 问题 ID: %d，开始时间: %s ==========", problemID, startTime.Format(time.RFC3339))
	// 检查必要的依赖
	if s.repoFactory.Problems() == nil {
		return s.createFailedCallback(problemID, startTime), errors.New("问题数据仓库未配置")
	}

	// 查询 Problem，校验是否存在
	problemObject, err := s.getProblem(ctx, problemID)
	if err != nil {
		return s.createFailedCallback(problemID, startTime), errors.Wrapf(err, "查询 Problem 失败")
	}

	// Step1: 整理问题中故障点数据
	faultPointObjects, err := s.GetFaultPoints(ctx, problemObject)
	if err != nil {
		return s.createFailedCallback(problemID, startTime), errors.Wrapf(err, "获取故障点失败")
	}
	log.Infof("RCA 获取故障点完成, 问题 ID: %d , 获取到 %d 个故障点", problemID, len(faultPointObjects))

	if len(faultPointObjects) == 0 {
		return s.createFailedCallback(problemID, startTime), errors.New("无关联故障点")
	}

	// Step2: 故障点关联图召回
	recallCtx, err := s.GraphRecall(ctx, faultPointObjects, problemObject)
	if err != nil {
		return s.createFailedCallback(problemID, startTime), errors.Wrapf(err, "图召回失败")
	}
	log.Infof("RCA 图召回完成，问题 ID: %d", problemID)

	// Step3: 因果分析推理
	result, err := s.CausalAnalysis(ctx, faultPointObjects, recallCtx)
	if err != nil {

		return s.createFailedCallback(problemID, startTime), errors.Wrapf(err, "因果分析失败")
	}
	if result == nil {
		return s.createFailedCallback(problemID, startTime), errors.New("因果分析返回空结果")
	}

	log.Infof("RCA 因果分析完成，问题 ID: %d, 根因对象 ID: %s, 根因故障 ID: %d",
		problemID, result.RootCauseObjectID, result.RootCauseFaultID)

	// Step4: 构建故障溯源分析展示数据
	analysisCallback, err := s.BuildAnalysisCallback(ctx, problemObject, faultPointObjects, recallCtx, result, startTime)
	if err != nil {
		// 构建回调失败
		return s.createFailedCallback(problemID, startTime), errors.Wrapf(err, "构建分析回调数据失败")
	}

	log.Infof("========== RCA 分析完成，问题 ID: %d，状态: %d ==========", problemID, analysisCallback.RcaStatus)

	return analysisCallback, nil

}

// getProblem 查询并验证问题是否存在
func (s *Service) getProblem(ctx context.Context, problemID uint64) (domain.Problem, error) {
	problemObjects, err := s.repoFactory.Problems().QueryByIDs(ctx, []uint64{problemID})
	if err != nil {
		return domain.Problem{}, errors.Wrapf(err, "查询 Problem 失败")
	}
	if len(problemObjects) == 0 {
		return domain.Problem{}, errors.Errorf("problem ID：%d 不存在", problemID)
	}
	return problemObjects[0], nil
}

func (s *Service) createFailedCallback(problemID uint64, startTime time.Time) *domain.RCACallback {
	return &domain.RCACallback{
		ProblemID:          problemID,
		ProblemName:        defaultNameNoFaultPoints,
		ProblemDescription: defaultImpactNoFaultPoints,
		RootCauseObjectID:  "",
		RootCauseFaultID:   0,
		RcaResults: utils.JsonEncode(domain.RcaResults{
			RcaID:      fmt.Sprintf("rca_%d", s.idGenerator.NextID()),
			AdpKnID:    s.config.AppConfig.KnowledgeNetwork.KnowledgeID,
			RcaContext: domain.RcaContext{},
		}),
		RcaStartTime: startTime,
		RcaEndTime:   timex.NowLocalTime(),
		InProgress:   false,
		RcaStatus:    domain.RcaStatusFailed,
	}
}

// ----- Step1: 整理问题中故障点数据 -----
func (s *Service) GetFaultPoints(ctx context.Context, problemObject domain.Problem) ([]domain.FaultPointObject, error) {
	if len(problemObject.RelationIDs) == 0 {
		return nil, nil
	}

	// 检查存储仓库是否配置
	if s.repoFactory.FaultPoints() == nil {
		return nil, errors.New("故障点数据仓库未配置")
	}

	// 查询故障点数据
	faultPointObjects, err := s.repoFactory.FaultPoints().QueryByIDs(ctx, problemObject.RelationIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "查询故障点失败")
	}
	log.Infof("故障点数据个数：%v", len(faultPointObjects))

	return faultPointObjects, nil
}

// ----- Step2: 故障点关联图召回 -----
func (s *Service) GraphRecall(ctx context.Context, faultPointObjects []domain.FaultPointObject, problemObject domain.Problem) (*domain.GraphRecallContext, error) {
	// 参数验证
	if len(faultPointObjects) == 0 {
		return &domain.GraphRecallContext{
			TopologySubgraphs:             make(map[string]*domain.Topology),
			TopologyNeighbors:             make(map[string][]string),
			HistoricalCausality:           make(map[string][]domain.CausalRelation),
			HistoricalNeighborFaultPoints: make([]domain.FaultPointObject, 0),
			AnalysisNetwork:               make([]*domain.RcaNetwork, 0),
		}, nil
	}

	recallCtx := &domain.GraphRecallContext{
		TopologySubgraphs:             make(map[string]*domain.Topology),
		TopologyNeighbors:             make(map[string][]string), // 修复：添加缺失的字段初始化
		HistoricalCausality:           make(map[string][]domain.CausalRelation),
		HistoricalNeighborFaultPoints: make([]domain.FaultPointObject, 0),
		AnalysisNetwork:               make([]*domain.RcaNetwork, 0),
	}

	// 遍历每个故障点，获取对应对象的实体类 ID 和实体 ID，按照对象实体类种类进行分组
	entityClassIDMap := make(map[string][]string)
	for _, fp := range faultPointObjects {
		// 跳过无效的实体ID或实体类ID
		if fp.EntityObjectID == "" || fp.EntityObjectClass == "" {
			continue
		}

		entityClassIDMap[fp.EntityObjectClass] = append(entityClassIDMap[fp.EntityObjectClass], fp.EntityObjectID)
	}

	// 遍历 entityClassIDMap，进行图谱召回
	for entityClassID, entityIDs := range entityClassIDMap {
		if entityClassID == "" || len(entityIDs) == 0 {
			continue
		}
		log.Infof("entityClassID: %v, entityIDs: %v", entityClassID, entityIDs)
		// 2.1 召回拓扑对象子图
		// 传入参数：对象类（entityClassID）, 对象ID列表（entityIDs）, 进行按照分组进行子图召回
		s.recallTopologySubgraph(ctx, recallCtx, entityClassID, entityIDs, problemObject.AffectedEntityIDs)
	}

	return recallCtx, nil
}

// ----- Step3: 因果分析推理 -----
// CausalAnalysis 故障点之间的因果分析推理
func (s *Service) CausalAnalysis(ctx context.Context, faultPointInfos []domain.FaultPointObject, recallCtx *domain.GraphRecallContext) (*domain.CausalAnalysisResults, error) {
	// 参数验证
	if len(faultPointInfos) == 0 {
		return &domain.CausalAnalysisResults{
			CausalRelations:      []domain.CausalCandidate{},
			FaultCausals:         []domain.FaultCausalObject{},
			FaultCausalRelations: []domain.FaultCausalRelation{},
		}, nil
	}

	result := &domain.CausalAnalysisResults{
		CausalRelations:      []domain.CausalCandidate{},
		FaultCausals:         []domain.FaultCausalObject{},
		FaultCausalRelations: []domain.FaultCausalRelation{},
	}

	// 3.1 判断是否应建立新的因果关系（AI Agent 实现）
	candidates := s.findCausalCandidates(ctx, faultPointInfos, recallCtx)

	// 3.2 根据因果关系，转换为"因果推理实体"和关系
	// 将逻辑关系（AI Agent返回的故障点A -> 故障点B）转换为物理结构（实体A -> "因果推理实体" -> 实体B）
	faultCausals, faultCausalRelations := s.convertCandidatesToFaultCausals(ctx, candidates)

	// 3.3 处理因果冲突（互斥情况）,并进行更新需求
	if err := s.detectAndResolveOpenSearchCausalityConflicts(ctx, &faultCausals, &faultCausalRelations); err != nil {
		// 冲突检测和更新失败不影响整体流程，记录错误但继续执行
		log.Infof("检测和解决 OpenSearch 因果关系冲突失败: %v", err)
	}

	// 3.4 确定根因
	rootCause := s.determineRootCause(faultPointInfos, candidates)
	if rootCause != nil {
		result.RootCauseObjectID = rootCause.EntityObjectID
		result.RootCauseFaultID = rootCause.FaultID
	}

	// 将处理后的最终数据赋值给 result（冲突检测和解决后的最终数据）
	result.FaultCausals = faultCausals
	result.FaultCausalRelations = faultCausalRelations
	result.CausalRelations = candidates

	return result, nil
}

// ========== Step4: 构建故障溯源分析展示数据 ==========
// 构建分析回调结果
// 包含完整的分析结果数据，用于回传给 Problem 模块
func (s *Service) BuildAnalysisCallback(ctx context.Context, problemObject domain.Problem, faultPointInfos []domain.FaultPointObject, recallCtx *domain.GraphRecallContext, result *domain.CausalAnalysisResults, startTime time.Time) (*domain.RCACallback, error) {
	// 参数验证
	if result == nil {
		return nil, errors.New("因果分析结果不能为 nil")
	}

	// 构建分析上下文（传入 result 以生成完整的网络和影响范围）
	analysisContext, err := s.buildRcaContext(ctx, problemObject.ProblemName, faultPointInfos, recallCtx, result)
	if err != nil {
		return s.createFailedCallback(problemObject.ProblemID, startTime), errors.New("构建分析上下文失败")
	}

	analysisCallback := &domain.RCACallback{
		ProblemID:          problemObject.ProblemID,
		ProblemName:        analysisContext.Occurrence.Name,
		ProblemDescription: analysisContext.Occurrence.Description,
		RootCauseObjectID:  result.RootCauseObjectID,
		RootCauseFaultID:   result.RootCauseFaultID,
		RcaResults: utils.JsonEncode(domain.RcaResults{
			RcaID:      fmt.Sprintf("rca_%d", s.idGenerator.NextID()),
			AdpKnID:    s.config.AppConfig.KnowledgeNetwork.KnowledgeID,
			RcaContext: analysisContext,
		}),
		RcaStartTime: startTime,
		RcaEndTime:   timex.NowLocalTime(),
		InProgress:   false,
	}

	// 判断 RcaStatus 状态
	analysisCallback.RcaStatus = domain.RcaStatusSuccess

	return analysisCallback, nil
}
