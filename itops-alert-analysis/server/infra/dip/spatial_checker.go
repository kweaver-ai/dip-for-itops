package dip

import (
	"context"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"github.com/pkg/errors"
)

// SpatialChecker 空间相关性检查
type SpatialChecker struct {
	client *Client
}

// NewSpatialChecker 创建空间相关性检查器实例。
func NewSpatialChecker(client *Client) *SpatialChecker {
	return &SpatialChecker{
		client: client,
	}
}

// FilterCorrelatedProblems 查询空间关系子图，过滤出与故障点空间相关的问题。
func (s *SpatialChecker) FilterCorrelatedProblems(
	ctx context.Context,
	fp domain.FaultPointObject,
	problems []domain.Problem,
) ([]domain.Problem, error) {
	if s.client == nil {
		return nil, errors.New("DIP 客户端未配置")
	}

	if len(problems) == 0 {
		return nil, nil
	}

	// 如果故障点没有 EntityObjectID，无法进行空间相关性判断
	if len(fp.EntityObjectID) == 0 {
		log.Infof("故障点 %d 缺少 EntityObjectID，跳过空间相关性判断", fp.FaultID)
		return nil, errors.New("故障点缺少 EntityObjectID")
	}

	// 调用子图查询 API
	subGraphReq := SubGraphQueryRequest{
		SourceObjectTypeID: fp.EntityObjectClass,
		Condition: &Condition{
			Field:     "s_id",
			Operation: "==",
			Value:     fp.EntityObjectID,
		},
		Direction:  "bidirectional",
		PathLength: 1,
	}

	subGraphResp, err := s.client.QuerySubGraph(ctx, subGraphReq)
	if err != nil {
		return nil, errors.Wrap(err, "查询子图失败")
	}

	// 提取子图中所有对象的 s_id
	spatialObjectIDs := make(map[string]struct{})
	for _, obj := range subGraphResp.Objects {
		if len(obj.Properties.SID) == 0 {
			continue
		}
		spatialObjectIDs[obj.Properties.SID] = struct{}{}
	}

	log.Infof("子图查询返回 %d 个相关对象", len(spatialObjectIDs))

	// 判断哪些问题与故障点空间相关
	var correlatedProblems []domain.Problem
	for _, problem := range problems {
		// 检查问题的 affected_entity_ids 是否在子图结果中
		isCorrelated := false
		for _, entityID := range problem.AffectedEntityIDs {
			if _, exists := spatialObjectIDs[entityID]; exists {
				isCorrelated = true
				log.Infof("问题 %d 空间相关（匹配 entity_id: %s）", problem.ProblemID, entityID)
				break
			}
		}
		if isCorrelated {
			correlatedProblems = append(correlatedProblems, problem)
		}
	}

	log.Infof("空间相关性判断: %d 个问题中，%d 个空间相关", len(problems), len(correlatedProblems))
	return correlatedProblems, nil
}
