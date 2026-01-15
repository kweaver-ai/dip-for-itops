package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/kafka"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/opensearch"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/slice"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// Server 提供 HTTP 入口：事件接收、查询、问题关闭。
type Server struct {
	cfg            *config.Config
	kafkaProducer  core.KafkaProducer
	repoFactory    *opensearch.RepositoryFactory
	problemHandler core.ProblemHandler
	router         *gin.Engine
	httpServer     *http.Server
}

func New(cfg *config.Config, repoFactory *opensearch.RepositoryFactory, problemHandler core.ProblemHandler) (*Server, error) {
	kafkaProducer, err := kafka.NewProducer(kafka.Config{
		Brokers: []string{fmt.Sprintf("%s:%d", cfg.DepServices.MQ.MQHost, cfg.DepServices.MQ.MQPort)},
		SASL: &kafka.SASLConfig{
			Enabled:  true,
			Username: cfg.DepServices.MQ.Auth.Username,
			Password: cfg.DepServices.MQ.Auth.Password,
		},
		Topic: cfg.Kafka.RawEvents.Topic,
	})
	if err != nil {
		return nil, errors.Wrap(err, "初始化 Kafka Producer 失败")
	}

	return &Server{
		cfg:            cfg,
		kafkaProducer:  kafkaProducer,
		repoFactory:    repoFactory,
		problemHandler: problemHandler,
	}, nil
}

// Start 使用 gin 注册 /v1 接口并启动 HTTP Server。
func (s *Server) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// 第一层：/api/itops-alert-analysis
	api := engine.Group("/api/itops-alert-analysis")

	// 第二层：v1 版本
	v1 := api.Group("/v1")
	{
		v1.POST("/events", s.postEvent)
		v1.GET("/events/info/:event_ids", s.queryEvents)
		v1.GET("/fault-points/info/:fault_ids", s.queryFaultPoints)
		v1.GET("/problems/info/:problem_ids", s.queryProblems)
		v1.POST("/problems/:problem_id/close", s.closeProblem)
		v1.POST("/problems/:problem_id/root-cause", s.setRootCause)
	}

	// 调试接口
	debug := v1.Group("/debug")
	{
		debug.GET("/problem/:problem_id/tree", s.problemTree)
	}

	addr := fmt.Sprintf(":%d", s.cfg.API.Port)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	s.router = engine
	s.httpServer = httpSrv

	errCh := make(chan error, 1)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// Stop 优雅关闭 HTTP 服务。
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) postEvent(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20) // 限制 10MB
	defer func() {
		if c.Request.Body != nil {
			_ = c.Request.Body.Close()
		}
	}()

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求失败"})
		return
	}
	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体不能为空"})
		return
	}

	log.Debugf("收到Webhook发送数据，内容:%s", string(body))

	key := fmt.Sprintf("%d", time.Now().UnixNano())

	if err := s.kafkaProducer.PublishRawEvent(c.Request.Context(), key, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("写入 Kafka 失败: %v", err)})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted", "key": key})
}

func (s *Server) queryEvents(c *gin.Context) {
	eventIDsParam := c.Param("event_ids")
	providerIDsParam := c.Param("provider_ids")

	var (
		items []domain.RawEvent
		err   error
	)
	if len(eventIDsParam) == 0 && len(providerIDsParam) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	switch {
	case len(eventIDsParam) > 0:
		eventIDs := slice.SplitToUint64s(eventIDsParam)
		items, err = s.repoFactory.RawEvents().QueryByIDs(c.Request.Context(), eventIDs)
	case len(providerIDsParam) > 0:
		providerIDs := slice.SplitToStrings(providerIDsParam)
		items, err = s.repoFactory.RawEvents().QueryByProviderID(c.Request.Context(), providerIDs)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 event_ids 或 provider_ids 参数"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) queryFaultPoints(c *gin.Context) {
	faultIDsParam := c.Param("fault_ids")
	if len(faultIDsParam) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fault_ids 参数必填"})
		return
	}

	faultIDs := slice.SplitToUint64s(faultIDsParam)
	if len(faultIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fault_ids 参数格式错误"})
		return
	}

	items, err := s.repoFactory.FaultPoints().QueryByIDs(c.Request.Context(), faultIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) queryProblems(c *gin.Context) {
	problemIDsParam := c.Param("problem_ids")
	if len(problemIDsParam) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem_ids 参数必填"})
		return
	}

	problemIDs := slice.SplitToUint64s(problemIDsParam)
	if len(problemIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem_ids 参数格式错误"})
		return
	}

	items, err := s.repoFactory.Problems().QueryByIDs(c.Request.Context(), problemIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) closeProblem(c *gin.Context) {
	if s.problemHandler == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "problem handler 未配置"})
		return
	}

	problemIDStr := c.Param("problem_id")
	if len(problemIDStr) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem_id 不能为空"})
		return
	}

	problemID := cast.ToUint64(problemIDStr)
	if problemID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem_id 必须是有效的数字"})
		return
	}

	var req closeProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数验证失败: %v", err)})
		return
	}

	// 查询问题状态
	problems, err := s.repoFactory.Problems().QueryByIDs(c.Request.Context(), []uint64{problemID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("查询问题失败: %v", err)})
		return
	}
	if len(problems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "问题不存在"})
		return
	}

	// 检查问题状态是否为 open
	if problems[0].ProblemStatus != domain.ProblemStatusOpen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("问题状态为 %s，不允许关闭", problems[0].ProblemStatus)})
		return
	}

	problem := problems[0]

	if err := s.problemHandler.CloseProblem(c.Request.Context(), problemID, domain.ProblemCloseTypeManual, domain.ProblemStatusClosed, req.Notes, req.ClosedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"problem_id":    problemID,
		"status":        "closed",
		"events_closed": len(problem.RelationEventIDs),
		"faults_closed": len(problem.RelationIDs),
	})
}

func (s *Server) setRootCause(c *gin.Context) {
	problemIDStr := c.Param("problem_id")
	if len(problemIDStr) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem_id 不能为空"})
		return
	}

	problemID := cast.ToUint64(problemIDStr)
	if problemID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem_id 必须是有效的数字"})
		return
	}

	var req setRootCauseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数验证失败: %v", err)})
		return
	}

	// 查询问题状态
	problems, err := s.repoFactory.Problems().QueryByIDs(c.Request.Context(), []uint64{problemID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("查询问题失败: %v", err)})
		return
	}
	if len(problems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "问题不存在"})
		return
	}

	// 检查问题状态是否为 open
	//if problems[0].ProblemStatus != domain.ProblemStatusOpen {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("问题状态为 %s，不允许设置根因", problems[0].ProblemStatus)})
	//	return
	//}

	if err := s.repoFactory.Problems().UpdateRootCauseObjectID(c.Request.Context(), problemID, req.RootCauseObjectID, req.RootCauseFaultID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"problem_id":           problemID,
		"root_cause_object_id": req.RootCauseObjectID,
		"root_cause_fault_id":  req.RootCauseFaultID,
		"status":               "updated",
	})
}

type closeProblemRequest struct {
	//CloseType domain.ProblemCloseType `json:"close_type" binding:"required,oneof=1 2"`
	Notes    string `json:"notes"`
	ClosedBy string `json:"closed_by" binding:"required"`
}

type setRootCauseRequest struct {
	RootCauseObjectID string `json:"root_cause_object_id" binding:"required"`
	RootCauseFaultID  uint64 `json:"root_cause_fault_id" binding:"required"`
}

// problemTree 调试接口：查看问题的完整树状结构。
// GET /api/itops-alert-analysis/v1/debug/problem/:problem_id/tree
func (s *Server) problemTree(c *gin.Context) {
	problemIDStr := c.Param("problem_id")
	if len(problemIDStr) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem_id 不能为空"})
		return
	}

	problemID := cast.ToUint64(problemIDStr)
	if problemID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem_id 必须是有效的数字"})
		return
	}

	// 1. 查询问题
	problems, err := s.repoFactory.Problems().QueryByIDs(c.Request.Context(), []uint64{problemID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("查询问题失败: %v", err)})
		return
	}

	if len(problems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "问题不存在"})
		return
	}

	problem := problems[0]

	// 2. 查询所有关联的故障点
	var faultPoints []domain.FaultPointObject
	if len(problem.RelationIDs) > 0 {
		faultPoints, err = s.repoFactory.FaultPoints().QueryByIDs(c.Request.Context(), problem.RelationIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("查询故障点失败: %v", err)})
			return
		}
	}

	// 3. 查询所有关联的事件
	var events []domain.RawEvent
	if len(problem.RelationEventIDs) > 0 {
		events, err = s.repoFactory.RawEvents().QueryByIDs(c.Request.Context(), problem.RelationEventIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("查询事件失败: %v", err)})
			return
		}
	}

	// 4. 构建响应
	c.JSON(http.StatusOK, gin.H{
		"problem":      problem,
		"fault_points": faultPoints,
		"events":       events,
		"statistics": gin.H{
			"problem_id":       problem.ProblemID,
			"fault_count":      len(faultPoints),
			"event_count":      len(events),
			"problem_status":   problem.ProblemStatus,
			"problem_level":    problem.ProblemLevel,
			"duration_seconds": uint64(problem.ProblemDuration),
		},
		"trace_path": buildTracePath(problem, faultPoints, events),
	})
}

// buildTracePath 构建追踪路径，展示数据流转关系。
func buildTracePath(problem domain.Problem, faultPoints []domain.FaultPointObject, events []domain.RawEvent) []gin.H {
	var path []gin.H

	// 按时间排序展示数据流
	path = append(path, gin.H{
		"step":        1,
		"stage":       "事件接收",
		"description": fmt.Sprintf("接收到 %d 个原始事件", len(events)),
		"event_ids":   problem.RelationEventIDs,
	})

	path = append(path, gin.H{
		"step":        2,
		"stage":       "故障点收敛",
		"description": fmt.Sprintf("收敛为 %d 个故障点", len(faultPoints)),
		"fault_ids":   problem.RelationIDs,
	})

	path = append(path, gin.H{
		"step":        3,
		"stage":       "问题关联",
		"description": fmt.Sprintf("关联到问题 %d", problem.ProblemID),
		"problem_id":  problem.ProblemID,
		"status":      problem.ProblemStatus,
	})

	if len(problem.RootCauseObjectID) > 0 {
		path = append(path, gin.H{
			"step":              4,
			"stage":             "根因分析",
			"description":       "已识别根因对象",
			"root_cause_object": problem.RootCauseObjectID,
		})
	}

	return path
}
