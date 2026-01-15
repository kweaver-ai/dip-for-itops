package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/timex"
	opensearchsdk "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// FaultPointStore 负责 itops_fault_point 索引。
type FaultPointStore struct {
	client *opensearchsdk.Client
}

// faultPointDocument 包装 FaultPointObject 并补充索引所需的公共字段。
type faultPointDocument struct {
	domain.FaultPointObject
	Timestamp time.Time `json:"@timestamp"`
	WriteTime time.Time `json:"__write_time"`
	DataType  string    `json:"__data_type"`
	IndexBase string    `json:"__index_base"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	ID        string    `json:"__id"`
}

func NewFaultPointStore(client *opensearchsdk.Client) *FaultPointStore {
	return &FaultPointStore{client: client}
}

func (s *FaultPointStore) FindOpenByEntityAndMode(ctx context.Context, entityObjectID, failureMode string, t time.Time) (*domain.FaultPointObject, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.FindOpenByEntityAndMode",
			"index", FaultPointIndexObject,
			"entity_object_id", entityObjectID,
			"failure_mode", failureMode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}
	if entityObjectID == "" || failureMode == "" {
		return nil, errors.New("entityObjectID or failureMode is empty")
	}
	filters := []any{
		map[string]any{"term": map[string]any{"entity_object_id.keyword": entityObjectID}},
		map[string]any{"term": map[string]any{"fault_mode.keyword": failureMode}},
		// 仅查询未关闭的故障点（occurred）。
		map[string]any{"term": map[string]any{"fault_status.keyword": domain.FaultStatusOccurred}},
		map[string]any{
			"range": map[string]any{
				"fault_latest_time": map[string]any{"gte": t},
			},
		},
	}

	body, err := encodeBody(map[string]any{
		"size": 1,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
		"sort": []any{
			map[string]any{
				"fault_latest_time": map[string]any{
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
		Index: []string{FaultPointIndexObject},
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询 FaultPointObject 失败")
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
	result, err := decodeSearch[domain.FaultPointObject](data)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}

func (s *FaultPointStore) Upsert(ctx context.Context, fp domain.FaultPointObject) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.Upsert",
			"index", FaultPointIndexObject,
			"document_id", fp.FaultID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	if fp.FaultID == 0 {
		return errors.New("fault_id 不能为空")
	}

	ts := fp.FaultOccurTime
	if ts.IsZero() {
		ts = fp.FaultCreateTime
	}
	doc := faultPointDocument{
		FaultPointObject: fp,
		Timestamp:        ts,
		WriteTime:        time.Now().Local(),
		DataType:         FaultPointIndexObjectBase,
		IndexBase:        FaultPointIndexObjectBase,
		Category:         "log",
		Type:             FaultPointIndexObjectBase,
		ID:               cast.ToString(fp.FaultID),
	}

	body, err := encodeBody(doc)
	if err != nil {
		return err
	}
	req := opensearchapi.IndexRequest{
		Index:      FaultPointIndexObject,
		DocumentID: cast.ToString(fp.FaultID),
		Body:       body,
		Refresh:    "wait_for",
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "写入 FaultPointObject 失败")
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

func (s *FaultPointStore) UpdateProblemID(ctx context.Context, faultIDs []uint64, problemID uint64) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.UpdateProblemID",
			"index", FaultPointIndexObject,
			"problem_id", problemID,
			"fault_ids_count", len(faultIDs),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	return s.bulkUpdate(ctx, faultIDs, map[string]any{"problem_id": problemID})
}

// MakeRecovered 将故障点标记为已恢复。
func (s *FaultPointStore) MakeRecovered(ctx context.Context, faultID uint64, recoveryTime time.Time) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.MakeRecovered",
			"index", FaultPointIndexObject,
			"document_id", faultID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	doc := map[string]any{
		"fault_status":        domain.FaultStatusRecovered,
		"fault_recovery_time": recoveryTime.Local(),
	}
	return s.partialUpdate(ctx, faultID, doc)
}

// MakeExpired 将故障点标记为已失效。
func (s *FaultPointStore) MakeExpired(ctx context.Context, faultID uint64) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.MakeExpired",
			"index", FaultPointIndexObject,
			"document_id", faultID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	doc := map[string]any{
		"fault_status":      domain.FaultStatusExpired,
		"fault_update_time": timex.NowLocalTime().Local(),
	}
	return s.partialUpdate(ctx, faultID, doc)
}

func (s *FaultPointStore) QueryByIDs(ctx context.Context, ids []uint64) ([]domain.FaultPointObject, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.QueryByIDs",
			"index", FaultPointIndexObject,
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
		Index: FaultPointIndexObject,
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "mget FaultPointObject 失败")
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
	return decodeMGet[domain.FaultPointObject](data)
}

func (s *FaultPointStore) partialUpdate(ctx context.Context, id uint64, doc map[string]any) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.partialUpdate",
			"index", FaultPointIndexObject,
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
		Index:      FaultPointIndexObject,
		DocumentID: cast.ToString(id),
		Body:       body,
		Refresh:    "wait_for",
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrapf(err, "更新 FaultPointObject %d 失败", id)
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

// bulkUpdate 批量更新多个故障点文档的指定字段（使用 Bulk API，性能优化）
func (s *FaultPointStore) bulkUpdate(ctx context.Context, faultIDs []uint64, doc map[string]any) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.bulkUpdate",
			"index", FaultPointIndexObject,
			"fault_ids_count", len(faultIDs),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	if len(faultIDs) == 0 || len(doc) == 0 {
		return nil
	}

	// 构建 Bulk API 请求体
	var buf bytes.Buffer
	for _, id := range faultIDs {
		// 元数据行：指定操作和文档 ID
		meta := map[string]any{
			"update": map[string]any{
				"_index": FaultPointIndexObject,
				"_id":    cast.ToString(id),
			},
		}
		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			return errors.Wrap(err, "编码 bulk meta 失败")
		}

		// 文档行：要更新的字段
		docWrapper := map[string]any{"doc": doc}
		if err := json.NewEncoder(&buf).Encode(docWrapper); err != nil {
			return errors.Wrap(err, "编码 bulk doc 失败")
		}
	}

	// 执行 Bulk 请求
	req := opensearchapi.BulkRequest{
		Body:    &buf,
		Refresh: "wait_for", // 不强制刷新，使用默认刷新策略
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "bulk update 请求失败")
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.IsError() {
		data, _ := readResponseBody(res.Body)
		return errors.Errorf("bulk update 失败: %s", formatErrorMessage(data))
	}

	// 检查响应中是否有错误
	responseData, err := readResponseBody(res.Body)
	if err != nil {
		return errors.Wrap(err, "读取 bulk 响应失败")
	}

	var bulkResp struct {
		Errors bool             `json:"errors"`
		Items  []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(responseData, &bulkResp); err != nil {
		return errors.Wrap(err, "解析 bulk 响应失败")
	}

	if bulkResp.Errors {
		return errors.Errorf("bulk update 部分失败，总数=%d", len(faultIDs))
	}

	return nil
}

// FindByEventID 通过 event_id 查找故障点（event_id 在 RelationEventIDs 中）
func (s *FaultPointStore) FindByEventID(ctx context.Context, eventID uint64) (*domain.FaultPointObject, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.FindByEventID",
			"index", FaultPointIndexObject,
			"event_id", eventID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}
	if eventID == 0 {
		return nil, nil
	}

	// 查询 related_event_ids 数组中包含 eventID 的故障点
	body, err := encodeBody(map[string]any{
		"size": 1,
		"query": map[string]any{
			"term": map[string]any{
				"relation_event_ids": eventID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	req := opensearchapi.SearchRequest{
		Index: []string{FaultPointIndexObject},
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询 FaultPointObject 失败")
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
	result, err := decodeSearch[domain.FaultPointObject](data)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}

func (s *FaultPointStore) FindInWindow(ctx context.Context, entityID string, faultMode string, start, end time.Time) ([]domain.FaultPointObject, error) {
	defer func(t time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.FindInWindow",
			"index", FaultPointIndexObject,
			"entity_id", entityID,
			"fault_mode", faultMode,
			"duration_ms", time.Since(t).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}
	if entityID == "" {
		return nil, errors.New("entityID 不能为空")
	}
	filters := []any{
		map[string]any{"term": map[string]any{"entity_object_id": entityID}},
		map[string]any{"term": map[string]any{"fault_mode": faultMode}},
		map[string]any{"range": map[string]any{"fault_latest_time": map[string]any{"gte": start, "lte": end}}},
	}
	body, err := encodeBody(map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	req := opensearchapi.SearchRequest{
		Index: []string{FaultPointIndexObject},
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询 FaultPointObject 失败")
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
	return decodeSearch[domain.FaultPointObject](data)
}

// FindExpiredOccurred 查找所有状态为 occurred 但已超过过期时间的故障点。
func (s *FaultPointStore) FindExpiredOccurred(ctx context.Context, expirationTime time.Time) ([]domain.FaultPointObject, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointStore.FindExpiredOccurred",
			"index", FaultPointIndexObject,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}
	filters := []any{
		map[string]any{"term": map[string]any{"fault_status.keyword": domain.FaultStatusOccurred}},
		map[string]any{
			"range": map[string]any{
				"fault_latest_time": map[string]any{"lt": expirationTime},
			},
		},
	}

	body, err := encodeBody(map[string]any{
		"size": 1000, // 每次最多处理 1000 个过期故障点
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	req := opensearchapi.SearchRequest{
		Index: []string{FaultPointIndexObject},
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询过期故障点失败")
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
	return decodeSearch[domain.FaultPointObject](data)
}

var _ core.FaultPointRepository = (*FaultPointStore)(nil)
