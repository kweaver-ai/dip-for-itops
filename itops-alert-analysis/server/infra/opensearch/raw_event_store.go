package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	opensearchsdk "github.com/opensearch-project/opensearch-go/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// RawEventStore 负责 itops_raw_event 索引的全部操作。
type RawEventStore struct {
	client *opensearchsdk.Client
}

// rawEventDocument 包装 RawEvent 并补充索引所需的公共字段。
type rawEventDocument struct {
	domain.RawEvent
	Timestamp time.Time `json:"@timestamp"`
	WriteTime time.Time `json:"__write_time"`
	DataType  string    `json:"__data_type"`
	IndexBase string    `json:"__index_base"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	ID        string    `json:"__id"`
}

func NewRawEventStore(client *opensearchsdk.Client) *RawEventStore {
	return &RawEventStore{client: client}
}

func (s *RawEventStore) Upsert(ctx context.Context, event domain.RawEvent) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "RawEventStore.Upsert",
			"index", RawEventIndex,
			"document_id", event.EventID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	if event.EventID == 0 {
		return errors.New("event_id 不能为空")
	}

	ts := event.EventTimestamp
	if event.EventOccurTime != nil {
		ts = *event.EventOccurTime
	}
	doc := rawEventDocument{
		RawEvent:  event,
		Timestamp: ts,
		WriteTime: time.Now().Local(),
		DataType:  RawEventIndexBase,
		IndexBase: RawEventIndexBase,
		Category:  "log",
		Type:      RawEventIndexBase,
		ID:        cast.ToString(event.EventID),
	}

	body, err := encodeBody(doc)
	if err != nil {
		return err
	}

	req := opensearchapi.IndexRequest{
		Index:      RawEventIndex,
		DocumentID: cast.ToString(event.EventID),
		Body:       body,
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "写入 RawEvent 失败")
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

func (s *RawEventStore) UpdateFaultID(ctx context.Context, eventIDs []uint64, faultID uint64) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "RawEventStore.UpdateFaultID",
			"index", RawEventIndex,
			"fault_id", faultID,
			"event_ids_count", len(eventIDs),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	return s.bulkUpdate(ctx, eventIDs, map[string]any{"fault_id": faultID})
}

func (s *RawEventStore) UpdateProblemID(ctx context.Context, eventIDs []uint64, problemID uint64) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "RawEventStore.UpdateProblemID",
			"index", RawEventIndex,
			"problem_id", problemID,
			"event_ids_count", len(eventIDs),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	return s.bulkUpdate(ctx, eventIDs, map[string]any{"problem_id": problemID})
}

func (s *RawEventStore) QueryByIDs(ctx context.Context, ids []uint64) ([]domain.RawEvent, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "RawEventStore.QueryByIDs",
			"index", RawEventIndex,
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
		Index: RawEventIndex,
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "查询 RawEvent 失败")
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
	return decodeMGet[domain.RawEvent](data)
}

func (s *RawEventStore) QueryByProviderID(ctx context.Context, providerIDs []string) ([]domain.RawEvent, error) {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "RawEventStore.QueryByProviderID",
			"index", RawEventIndex,
			"provider_ids_count", len(providerIDs),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
	}
	if len(providerIDs) == 0 {
		return nil, nil
	}
	if len(providerIDs) > maxQuerySize {
		providerIDs = providerIDs[:maxQuerySize]
	}

	filters := []any{
		map[string]any{
			"term": map[string]any{"event_status": domain.EventStatusOccurred},
		},
		map[string]any{
			"terms": map[string]any{
				"event_provider_id": providerIDs,
			}},
	}

	body, err := encodeBody(map[string]any{
		"size": len(providerIDs),
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
		"sort": []any{
			map[string]any{
				"event_occur_time": map[string]any{
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
		Index: []string{RawEventIndex},
		Body:  body,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "搜索 RawEvent 失败")
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
	return decodeSearch[domain.RawEvent](data)
}

//func (s *RawEventStore) partialUpdate(ctx context.Context, eventID uint64, doc map[string]any) error {
//	if s.client == nil {
//		return errors.New("opensearch client 未初始化")
//	}
//	if eventID == 0 || len(doc) == 0 {
//		return nil
//	}
//	body, err := encodeBody(map[string]any{"doc": doc})
//	if err != nil {
//		return err
//	}
//	req := opensearchapi.UpdateRequest{
//		Index:      RawEventIndex,
//		DocumentID: cast.ToString(eventID),
//		Body:       body,
//		Refresh:    "wait_for",
//	}
//	res, err := req.Do(ctx, s.client)
//	if err != nil {
//		return errors.Wrapf(err, "更新 RawEvent %d 失败", eventID)
//	}
//	defer func() {
//		_ = res.Body.Close()
//	}()
//	if res.IsError() {
//		data, _ := readResponseBody(res.Body)
//		return formatErrorMessage(data)
//	}
//	return nil
//}

// bulkUpdate 批量更新多个文档的指定字段（使用 Bulk API，性能优化）
func (s *RawEventStore) bulkUpdate(ctx context.Context, eventIDs []uint64, doc map[string]any) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "RawEventStore.bulkUpdate",
			"index", RawEventIndex,
			"event_ids_count", len(eventIDs),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	if len(eventIDs) == 0 || len(doc) == 0 {
		return nil
	}

	// 构建 Bulk API 请求体
	var buf bytes.Buffer
	for _, id := range eventIDs {
		// 元数据行：指定操作和文档 ID
		meta := map[string]any{
			"update": map[string]any{
				"_index": RawEventIndex,
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
		Refresh: "wait_for", // 等待刷新完成，确保立即可查询
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
		return errors.Errorf("bulk update 部分失败，总数=%d", len(eventIDs))
	}

	return nil
}

var _ core.RawEventRepository = (*RawEventStore)(nil)
