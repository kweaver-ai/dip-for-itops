package opensearch

import (
	"context"
	"fmt"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/timex"
	opensearchsdk "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// ProblemStore 负责 itops_problem 索引。
type ProblemStore struct {
	client *opensearchsdk.Client
}

// problemDocument 包装 Problem 并补充索引所需的公共字段。
type problemDocument struct {
	domain.Problem
	Timestamp time.Time `json:"@timestamp"`
	WriteTime time.Time `json:"__write_time"`
	DataType  string    `json:"__data_type"`
	IndexBase string    `json:"__index_base"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	ID        string    `json:"__id"`
}

func NewProblemStore(client *opensearchsdk.Client) *ProblemStore {
	return &ProblemStore{client: client}
}

func (s *ProblemStore) FindCorrelated(ctx context.Context, fp domain.FaultPointObject, t time.Time) ([]domain.Problem, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.FindCorrelated",
			"index", ProblemIndex,
			"fault_id", fp.FaultID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}
	filters := []any{
		map[string]any{
			"range": map[string]any{
				"problem_latest_time": map[string]any{"gte": t.Local()},
			},
		},
		map[string]any{"term": map[string]any{"problem_status": domain.ProblemStatusOpen}},
	}

	body, err := encodeBody(map[string]any{
		"size": maxQuerySize,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
		"sort": []any{
			map[string]any{
				"problem_latest_time": map[string]any{
					"order":         "desc",
					"unmapped_type": "date",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	req := opensearchapi.SearchRequest{
		Index: []string{ProblemIndex},
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询 Problem 失败")
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.IsError() {
		data, _ := readResponseBody(res.Body)
		return nil, formatErrorMessage(data)
	}
	data, err := readResponseBody(res.Body)
	if err != nil {
		return nil, err
	}
	return decodeSearch[domain.Problem](data)
}

func (s *ProblemStore) FindPendingRCA(ctx context.Context, maxAge time.Duration) ([]domain.Problem, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.FindPendingRCA",
			"index", ProblemIndex,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}

	// 查询条件：Open 状态 + RCA 未完成 + 创建时间在允许范围内
	cutoffTime := time.Now().Add(-maxAge)
	filters := []any{
		map[string]any{"term": map[string]any{"problem_status": domain.ProblemStatusOpen}},
		map[string]any{
			"range": map[string]any{
				"problem_create_timestamp": map[string]any{"gte": cutoffTime.Local()},
			},
		},
	}

	// RCA 未完成：root_cause_fault_id 为 0 或不存在
	mustNot := []any{
		map[string]any{
			"range": map[string]any{
				"root_cause_fault_id": map[string]any{"gt": 0},
			},
		},
	}

	body, err := encodeBody(map[string]any{
		"size": maxQuerySize,
		"query": map[string]any{
			"bool": map[string]any{
				"filter":   filters,
				"must_not": mustNot,
			},
		},
		"sort": []any{
			map[string]any{"problem_create_timestamp": map[string]any{"order": "asc"}},
		},
	})
	if err != nil {
		return nil, err
	}

	req := opensearchapi.SearchRequest{
		Index: []string{ProblemIndex},
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询待 RCA 的问题失败")
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.IsError() {
		data, _ := readResponseBody(res.Body)
		return nil, formatErrorMessage(data)
	}
	data, err := readResponseBody(res.Body)
	if err != nil {
		return nil, err
	}
	return decodeSearch[domain.Problem](data)
}

// FindExpiredOpen 查询超过指定时间未更新且状态为打开的问题。
func (s *ProblemStore) FindExpiredOpen(ctx context.Context, expirationTime time.Time) ([]domain.Problem, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.FindExpiredOpen",
			"index", ProblemIndex,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}

	// 查询条件：Open 状态 + 结束时间早于失效时间（problem_latest_time 表示最后一次更新时间）
	filters := []any{
		map[string]any{"term": map[string]any{"problem_status": domain.ProblemStatusOpen}},
		map[string]any{
			"range": map[string]any{
				"problem_latest_time": map[string]any{"lt": expirationTime.Local()},
			},
		},
	}

	body, err := encodeBody(map[string]any{
		"size": maxQuerySize,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
		"sort": []any{
			map[string]any{"problem_latest_time": map[string]any{"order": "asc"}},
		},
	})
	if err != nil {
		return nil, err
	}

	req := opensearchapi.SearchRequest{
		Index: []string{ProblemIndex},
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询过期问题失败")
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.IsError() {
		data, _ := readResponseBody(res.Body)
		return nil, formatErrorMessage(data)
	}
	data, err := readResponseBody(res.Body)
	if err != nil {
		return nil, err
	}
	return decodeSearch[domain.Problem](data)
}

func (s *ProblemStore) Upsert(ctx context.Context, p domain.Problem) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.Upsert",
			"index", ProblemIndex,
			"document_id", p.ProblemID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	if p.ProblemID == 0 {
		return errors.New("problem_id 不能为空")
	}

	ts := p.ProblemOccurTime
	if ts.IsZero() {
		ts = p.ProblemCreateTimestamp
	}
	doc := problemDocument{
		Problem:   p,
		Timestamp: ts,
		WriteTime: time.Now().Local(),
		DataType:  ProblemIndexBase,
		IndexBase: ProblemIndexBase,
		Category:  "log",
		Type:      ProblemIndexBase,
		ID:        cast.ToString(p.ProblemID),
	}

	body, err := encodeBody(doc)
	if err != nil {
		return err
	}
	req := opensearchapi.IndexRequest{
		Index:      ProblemIndex,
		DocumentID: cast.ToString(p.ProblemID),
		Body:       body,
		Refresh:    "wait_for",
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "写入 Problem 失败")
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.IsError() {
		data, _ := readResponseBody(res.Body)
		return formatErrorMessage(data)
	}
	return nil
}

func (s *ProblemStore) UpdateRootCause(ctx context.Context, problemID uint64, cb domain.RCACallback) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.UpdateRootCause",
			"index", ProblemIndex,
			"document_id", problemID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	var (
		entityID           = cb.RootCauseObjectID
		faultID            = cb.RootCauseFaultID
		rcaResults         = cb.RcaResults
		rcaStartTime       = cb.RcaStartTime
		rcaEndTime         = cb.RcaEndTime
		rcaStatus          = cb.RcaStatus
		problemName        = cb.ProblemName
		problemDescription = cb.ProblemDescription
	)

	if rcaStatus != domain.RcaStatusSuccess {
		return errors.New(fmt.Sprintf("rca 状态失败:problemID:%d,cb:%+v", problemID, utils.JsonEncode(cb)))
	}

	doc := map[string]any{
		"root_cause_object_id": entityID,
		"root_cause_fault_id":  faultID,
		"rca_results":          rcaResults,
		"rca_start_time":       rcaStartTime,
		"rca_end_time":         rcaEndTime,
		"rca_status":           rcaStatus,
		"problem_name":         problemName,
		"problem_description":  problemDescription,
	}

	return s.partialUpdate(ctx, problemID, doc)
}

func (s *ProblemStore) UpdateRootCauseObjectID(ctx context.Context, problemID uint64, objectID string, faultID uint64) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.UpdateRootCauseObjectID",
			"index", ProblemIndex,
			"document_id", problemID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	doc := map[string]any{
		"root_cause_object_id": objectID,
		"root_cause_fault_id":  faultID,
	}
	return s.partialUpdate(ctx, problemID, doc)
}

func (s *ProblemStore) UpdateRelationEventIDs(ctx context.Context, problemID uint64, eventIDs []uint64) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.UpdateRelationEventIDs",
			"index", ProblemIndex,
			"document_id", problemID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	doc := map[string]any{
		"relation_event_ids":  eventIDs,
		"problem_update_time": timex.NowLocalTime().Local(),
	}
	return s.partialUpdate(ctx, problemID, doc)
}
func (s *ProblemStore) MarkClosed(ctx context.Context, problemID uint64, closeType domain.ProblemCloseType, closeStatus domain.ProblemStatus, duration uint64, notes string, by string) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.MarkClosed",
			"index", ProblemIndex,
			"document_id", problemID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	doc := map[string]any{
		"problem_status":      closeStatus,
		"problem_close_type":  closeType,
		"problem_close_notes": notes,
		"problem_closed_by":   by,
		"problem_close_time":  timex.NowLocalTime().Local(),
		"problem_update_time": timex.NowLocalTime().Local(),
	}
	if duration > 0 {
		doc["problem_duration"] = duration
	}
	return s.partialUpdate(ctx, problemID, doc)
}

func (s *ProblemStore) MarkExpired(ctx context.Context, problemID uint64) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.MarkExpired",
			"index", ProblemIndex,
			"document_id", problemID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	// 查询问题获取创建时间，用于计算持续时间
	problems, err := s.QueryByIDs(ctx, []uint64{problemID})
	if err != nil {
		return errors.Wrapf(err, "查询问题 %d 失败", problemID)
	}
	if len(problems) == 0 {
		return errors.Errorf("问题 %d 不存在", problemID)
	}

	now := timex.NowLocalTime()

	doc := map[string]any{
		"problem_status":      domain.ProblemStatusExpired,
		"problem_update_time": now.Local(),
		//失效是没有关闭时间的
	}
	return s.partialUpdate(ctx, problemID, doc)
}

func (s *ProblemStore) QueryByIDs(ctx context.Context, ids []uint64) ([]domain.Problem, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.QueryByIDs",
			"index", ProblemIndex,
			"ids_count", len(ids),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}
	if len(ids) == 0 {
		return nil, nil
	}

	// 将 uint64 转换为 string 用于查询
	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = cast.ToString(id)
	}

	body, err := encodeBody(map[string]any{"ids": strIDs})
	if err != nil {
		return nil, err
	}
	req := opensearchapi.MgetRequest{
		Index: ProblemIndex,
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询 Problem 失败")
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.IsError() {
		data, _ := readResponseBody(res.Body)
		return nil, formatErrorMessage(data)
	}
	data, err := readResponseBody(res.Body)
	if err != nil {
		return nil, err
	}
	return decodeMGet[domain.Problem](data)
}

// ClearMergedProblemData 清空被合并问题的关联数据
// 用于问题合并后，清除被合并问题的故障点、事件、RCA结果等数据
func (s *ProblemStore) ClearMergedProblemData(ctx context.Context, problemID uint64) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.ClearMergedProblemData",
			"index", ProblemIndex,
			"document_id", problemID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	doc := map[string]any{
		"relation_ids":         []uint64{},                   // 清空故障点列表
		"relation_event_ids":   []uint64{},                   // 清空事件列表
		"affected_entity_ids":  []string{},                   // 清空受影响实体列表
		"root_cause_object_id": "",                           // 清空根因对象ID
		"root_cause_fault_id":  0,                            // 清空根因故障ID
		"rca_results":          "",                           // 清空RCA结果
		"rca_status":           "",                           // 清空RCA状态
		"problem_update_time":  timex.NowLocalTime().Local(), // 更新时间
	}

	return s.partialUpdate(ctx, problemID, doc)
}

func (s *ProblemStore) partialUpdate(ctx context.Context, id uint64, doc map[string]any) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "ProblemStore.partialUpdate",
			"index", ProblemIndex,
			"document_id", id,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	if id == 0 || len(doc) == 0 {
		return nil
	}
	body, err := encodeBody(map[string]any{"doc": doc})
	if err != nil {
		return err
	}
	req := opensearchapi.UpdateRequest{
		Index:      ProblemIndex,
		DocumentID: cast.ToString(id),
		Body:       body,
		Refresh:    "wait_for",
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrapf(err, "更新 Problem %d 失败", id)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.IsError() {
		data, _ := readResponseBody(res.Body)
		return formatErrorMessage(data)
	}
	return nil
}

var _ core.ProblemRepository = (*ProblemStore)(nil)
