package service

import (
	"context"
	"fmt"
	"sort"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/vo"
)

//go:generate mockgen -source ./problem.go -destination ../../mock/service/mock_problem_service.go -package mock
type ProblemService interface {
	List(ctx context.Context, req vo.DataViewQueryV2, accout_id string) (vo.ViewUniResponseV2, core.RestAPIError)
	Close(ctx context.Context, problemId, accountId string) core.RestAPIError
	SetRootCause(ctx context.Context, problemId string, req vo.RootCauseObjectIdParams) core.RestAPIError
	GetSubGraphByProblemId(ctx context.Context, problemId, accountId string) (vo.RcaContextResp, core.RestAPIError)
	SubGraphQuery(ctx context.Context, faultObjectResp []map[string]any, result *vo.RcaContextResp) core.RestAPIError
	GetRelationInfo(faultObjectResp []map[string]any) (map[string][]any, map[string][]float64, map[string]float64)
}

type problemService struct {
	uniQueryClient         dependency.UniQueryClient
	alertAnalysisClient    dependency.AlertAnalysisClient
	userManagementClient   dependency.UserManagementClient
	knowledgeNetworkClient dependency.KnowledgeNetworkClient
	configService          ConfigService
}

// Problem list
func (svc *problemService) List(ctx context.Context, req vo.DataViewQueryV2, accout_id string) (vo.ViewUniResponseV2, core.RestAPIError) {
	resp := vo.ViewUniResponseV2{}
	problem_views, err := svc.uniQueryClient.GetDataView(ctx, "__itops_problem", req, accout_id)
	if err != nil {
		return problem_views, dependency.NewClientRequestError(err)
	}
	if len(problem_views.Entries) == 0 {
		return problem_views, nil
	}
	for _, entry := range problem_views.Entries {
		if entry["root_cause_fault_id"].(float64) == 0 {
			entry["root_cause_entity_object_name"] = ""
			entry["root_cause_fault_description"] = ""
			continue
		}
		fault_point_req := vo.DataViewQueryV2{}
		fault_point_req.ViewQueryCommonParams = vo.ViewQueryCommonParams{
			Offset: 0,
			Limit:  1,
		}
		fault_point_req.GlobalFilters = map[string]any{
			"value":      entry["root_cause_fault_id"].(float64),
			"operation":  "==",
			"value_from": "const",
			"field":      "fault_id",
		}
		fault_point_views, err := svc.uniQueryClient.GetDataView(ctx, "__itops_fault_point_object", fault_point_req, accout_id)
		if err != nil {
			return resp, dependency.NewClientRequestError(err)
		}
		if len(fault_point_views.Entries) == 0 {
			err = fmt.Errorf("The root cause fault point id (%.17g) for problem (%v) does not exist", entry["root_cause_fault_id"], entry["problem_name"])
			return resp, dependency.NewClientRequestError(err)
		}
		entry["root_cause_entity_object_name"] = fault_point_views.Entries[0]["entity_object_name"]
		entry["root_cause_fault_description"] = fault_point_views.Entries[0]["fault_description"]
		entry["root_cause_fault_name"] = fault_point_views.Entries[0]["fault_name"]

	}
	return problem_views, nil
}

func (svc *problemService) Close(ctx context.Context, problemId, accountId string) core.RestAPIError {
	//查询用户名
	accountInfo, err := svc.userManagementClient.GetUserInfo(ctx, accountId)
	if err != nil {
		return dependency.NewClientRequestError(err)
	}

	//修改问题状态
	if err := svc.alertAnalysisClient.Close(ctx, problemId, accountInfo.Account); err != nil {
		return dependency.NewClientRequestError(err)
	}
	return nil
}

func (svc *problemService) SetRootCause(ctx context.Context, problemId string, req vo.RootCauseObjectIdParams) core.RestAPIError {
	if err := svc.alertAnalysisClient.SetRootCause(ctx, problemId, req.RootCauseObjectId, req.RootCauseFaultID); err != nil {
		return dependency.NewClientRequestError(err)
	}
	return nil
}

func (svc *problemService) GetSubGraphByProblemId(ctx context.Context, problemId, accountId string) (vo.RcaContextResp, core.RestAPIError) {
	resp := vo.RcaContextResp{}
	req := vo.DataViewQueryV2{
		ViewQueryCommonParams: vo.ViewQueryCommonParams{Limit: 100, Offset: 0},
		GlobalFilters:         map[string]any{"value": problemId, "operation": "==", "value_from": "const", "field": "problem_id"},
	}
	problem_views, err := svc.uniQueryClient.GetDataView(ctx, "__itops_problem", req, accountId)
	if err != nil {
		return resp, dependency.NewClientRequestError(err)
	}
	if len(problem_views.Entries) == 0 {
		return resp, nil
	}
	relationFaultIds := problem_views.Entries[0]["relation_fp_ids"].([]any)
	subConditions := make([]map[string]any, 0)
	relationFaultIdsMap := make(map[float64]bool)
	for _, id := range relationFaultIds {
		faultId := id.(float64)
		subConditions = append(subConditions, map[string]any{
			"value":      faultId,
			"operation":  "==",
			"value_from": "const",
			"field":      "fault_id",
		})
		relationFaultIdsMap[faultId] = true
	}

	//查询所有关联故障点信息
	faultObjectResp := make([]map[string]any, 0)
	noExistFaultIds := make([]float64, 0)
	for i := 0; i < len(relationFaultIds); i += 1000 {
		end := i + 1000
		if end > len(relationFaultIds) {
			end = len(relationFaultIds)
		}
		chunk := subConditions[i:end]
		fault_point_req := vo.DataViewQueryV2{}
		fault_point_req.ViewQueryCommonParams = vo.ViewQueryCommonParams{
			Offset: 0,
			Limit:  1000,
		}
		fault_point_req.GlobalFilters = map[string]any{
			"operation":      "or",
			"sub_conditions": chunk,
		}
		fault_point_views, err := svc.uniQueryClient.GetDataView(ctx, "__itops_fault_point_object", fault_point_req, accountId)
		if err != nil {
			return resp, dependency.NewClientRequestError(err)
		}
		if len(fault_point_views.Entries) != 0 {
			for _, entry := range fault_point_views.Entries {
				if _, ok := relationFaultIdsMap[entry["fault_id"].(float64)]; !ok {
					noExistFaultIds = append(noExistFaultIds, entry["fault_id"].(float64))
				}
			}
			faultObjectResp = append(faultObjectResp, fault_point_views.Entries...)
		}
	}
	// 故障点 排序
	if len(faultObjectResp) > 0 {
		sort.Slice(faultObjectResp, func(i, j int) bool {
			return faultObjectResp[i]["fault_create_time"].(string) < faultObjectResp[j]["fault_create_time"].(string)
		})
	}
	if len(noExistFaultIds) != 0 {
		err = fmt.Errorf("The IDs of the associated fault points that do not exist are %v", noExistFaultIds)
		return resp, dependency.NewClientRequestError(err)
	}
	// 查询视图关系
	if errApi := svc.SubGraphQuery(ctx, faultObjectResp, &resp); errApi != nil {
		return resp, errApi
	}
	resp.BackTrace = faultObjectResp
	return resp, nil
}

func (svc *problemService) SubGraphQuery(ctx context.Context, faultObjectResp []map[string]any, result *vo.RcaContextResp) core.RestAPIError {
	// 查询auth_oken、knowledge_network
	configs, errSer := svc.configService.ListConfigs(ctx, true)
	if errSer != nil {
		log.Error(errSer.Error())
		return dependency.NewClientRequestError(fmt.Errorf("Internal service error"))
	}
	if configs.Platform.AuthToken == "" {
		return dependency.NewClientRequestError(fmt.Errorf("The authentication token has not been set"))
	}
	if configs.KnowledgeNetwork.KnowledgeID == "" {
		return dependency.NewClientRequestError(fmt.Errorf("The knowledge network ID has not been set"))
	}
	// 类别下的所属对象
	serverObjects := make(map[string][]string, 0)
	allObjectIds := make(map[string]bool)
	for _, faultObject := range faultObjectResp {
		entityObjectId := faultObject["entity_object_id"].(string)
		entityObjectClass := faultObject["entity_object_class"].(string)
		if _, ok := serverObjects[entityObjectClass]; !ok {
			serverObjects[entityObjectClass] = []string{entityObjectId}
			allObjectIds[entityObjectId] = true
			continue
		}
		if _, ok := allObjectIds[entityObjectId]; !ok {
			allObjectIds[entityObjectId] = true
			serverObjects[entityObjectClass] = append(serverObjects[entityObjectClass], entityObjectId)
		}
	}
	nodes := make([]vo.RcaNode, 0)
	relations := make([]vo.Relation, 0)
	for entityObjectClass, entityObjectIds := range serverObjects {
		subConditions := make([]dependency.SubGraphSubCondition, 0)
		for _, entityObjectId := range entityObjectIds {
			subConditions = append(subConditions, dependency.SubGraphSubCondition{
				Operation: "==",
				Field:     "s_id",
				Value:     entityObjectId,
			})
		}
		// 查询视图关系
		subGraphReq := dependency.SubGraphQueryRequest{
			Limit:              10000,
			NeedTotal:          true,
			PathLength:         1,
			Direction:          "forward",
			SourceObjectTypeID: entityObjectClass,
			Condition: dependency.SubGraphCondition{
				Operation:     "or",
				SubConditions: subConditions,
			},
		}
		subGraphResp, err := svc.knowledgeNetworkClient.SubGraphQuery(ctx, subGraphReq, configs.Platform.AuthToken, configs.KnowledgeNetwork.KnowledgeID)
		if err != nil {
			return dependency.NewClientRequestError(err)
		}

		relationEventIdsByObjectId, relationFaultIdsByObjectId, maxLevelByObjectId := svc.GetRelationInfo(faultObjectResp)

		// 处理子图返回 nodes 结果
		for _, objectId := range entityObjectIds {
			uniquerKey := fmt.Sprint(entityObjectClass, "-", objectId)
			if subGraphResp.Objects != nil {
				properties := subGraphResp.Objects[uniquerKey].Properties
				node := vo.RcaNode{
					Node: vo.Node{
						SID:               properties["s_id"].(string),
						SCreateTime:       properties["s_create_time"].(string),
						SUpdateTime:       properties["s_update_time"].(string),
						Name:              properties["name"].(string),
						ObjectClass:       entityObjectClass,
						ObjectImpactLevel: maxLevelByObjectId[objectId],
					},
					RelationEventIDs:      relationEventIdsByObjectId[objectId],
					RelationFaultPointIDs: relationFaultIdsByObjectId[objectId],
				}
				node.IPAddress = make([]string, 0)
				switch val := properties["ip_address"].(type) {
				case string:
					if val != "" {
						node.IPAddress = []string{val}
					}
				case []string:
					if val != nil {
						node.IPAddress = val
					}
				}
				nodes = append(nodes, node)
			} else {
				objectInfoReq := dependency.ObjectInfoQueryRequest{
					Limit: 10000,
					Properties: []string{
						"s_id",
						"s_create_time",
						"s_update_time",
						"name",
						"ip_address",
					},
					NeedTotal: false,
					Condition: dependency.SubGraphCondition{
						Operation:     "or",
						SubConditions: subConditions,
					},
				}
				objectInfoResp, err := svc.knowledgeNetworkClient.ObjectInfoQuery(ctx,
					objectInfoReq,
					entityObjectClass,
					configs.Platform.AuthToken,
					configs.KnowledgeNetwork.KnowledgeID)
				if err != nil {
					return dependency.NewClientRequestError(err)
				}

				for _, data := range objectInfoResp.Datas {
					parseData := data.(map[string]interface{})
					node := vo.Node{
						SID:               parseData["s_id"].(string),
						Name:              parseData["name"].(string),
						SUpdateTime:       parseData["s_update_time"].(string),
						SCreateTime:       parseData["s_create_time"].(string),
						ObjectClass:       entityObjectClass,
						ObjectImpactLevel: maxLevelByObjectId[objectId],
					}
					node.IPAddress = make([]string, 0)
					switch val := parseData["ip_address"].(type) {
					case string:
						if val != "" {
							node.IPAddress = []string{val}
						}

					case []string:
						if val != nil {
							node.IPAddress = val
						}
					}
					nodes = append(nodes, vo.RcaNode{
						Node:                  node,
						RelationEventIDs:      relationEventIdsByObjectId[node.SID],
						RelationFaultPointIDs: relationFaultIdsByObjectId[node.SID],
					})
				}
			}
		}
		// 处理子图返回 relations 结果
		relationMap := make(map[string]vo.Relation)
		if subGraphResp.RelationPaths != nil {
			for _, relationPath := range subGraphResp.RelationPaths {
				for _, relation := range relationPath.Relations {
					sourceObjectID := subGraphResp.Objects[relation.SourceObjectID].Properties["s_id"].(string)
					targetSID := subGraphResp.Objects[relation.TargetObjectID].Properties["s_id"].(string)
					relationID := fmt.Sprint(entityObjectClass, "-", sourceObjectID, "-", targetSID)
					relationMap[relationID] = vo.Relation{
						RelationID:    relationID,
						RelationClass: entityObjectClass,
						SourceSID:     sourceObjectID,
						TargetSID:     targetSID,
					}
				}
			}
			for _, relation := range relationMap {
				relations = append(relations, relation)
			}
		}
	}
	result.Network.Nodes = nodes
	result.Network.Edges = relations
	return nil
}

// GetRelationInfo 收集故障点中对象关联的事件、故障点、故障点级别信息
func (svc *problemService) GetRelationInfo(faultObjectResp []map[string]any) (map[string][]any, map[string][]float64, map[string]float64) {
	// 对象关联的事件、故障点
	relationEventIdsByObjectId := map[string][]any{}
	allRelationEventIds := make(map[float64]bool)

	relationFaultIdsByObjectId := map[string][]float64{}
	allFaultIds := make(map[float64]bool)

	maxLevelByObjectId := map[string]float64{}
	for _, faultObject := range faultObjectResp {
		entityObjectId := faultObject["entity_object_id"].(string)
		relationEventIds := faultObject["relation_event_ids"].([]any)
		if _, ok := relationEventIdsByObjectId[entityObjectId]; !ok {
			relationEventIdsByObjectId[entityObjectId] = relationEventIds
			for _, relationEventId := range relationEventIds {
				allRelationEventIds[relationEventId.(float64)] = true
			}
			continue
		}
		for _, relationEventId := range relationEventIds {
			if _, ok := allRelationEventIds[relationEventId.(float64)]; !ok {
				allRelationEventIds[relationEventId.(float64)] = true
				relationEventIdsByObjectId[entityObjectId] = append(relationEventIdsByObjectId[entityObjectId], relationEventId)
			}
		}
	}

	for _, faultObject := range faultObjectResp {
		entityObjectId := faultObject["entity_object_id"].(string)
		faultId := faultObject["fault_id"].(float64)
		if _, ok := relationFaultIdsByObjectId[entityObjectId]; !ok {
			relationFaultIdsByObjectId[entityObjectId] = []float64{faultId}
			allFaultIds[faultId] = true
			continue
		}
		if _, ok := allFaultIds[faultId]; !ok {
			allFaultIds[faultId] = true
			relationFaultIdsByObjectId[entityObjectId] = append(relationFaultIdsByObjectId[entityObjectId], faultId)
		}
	}

	for _, faultObject := range faultObjectResp {
		entityObjectId := faultObject["entity_object_id"].(string)
		faultLevel := faultObject["fault_level"].(float64)
		if _, ok := maxLevelByObjectId[entityObjectId]; !ok {
			maxLevelByObjectId[entityObjectId] = faultLevel
			continue
		}
		if maxLevelByObjectId[entityObjectId] < faultLevel {
			maxLevelByObjectId[entityObjectId] = faultLevel
		}
	}
	return relationEventIdsByObjectId, relationFaultIdsByObjectId, maxLevelByObjectId
}
