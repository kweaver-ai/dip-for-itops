package opensearch

import (
	"context"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	opensearchsdk "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/pkg/errors"
)

// ========== 结构体定义 ==========

// FaultCausalRelationStore 负责 itops_fault_causal_relation 索引的存储操作
type FaultCausalRelationStore struct {
	client *opensearchsdk.Client
}

// FaultCausalRelationDocument 包装 FaultCausalRelation 并补充索引所需的公共字段
type FaultCausalRelationDocument struct {
	domain.FaultCausalRelation
	Timestamp time.Time `json:"@timestamp"`
	WriteTime time.Time `json:"__write_time"`
	DataType  string    `json:"__data_type"`
	IndexBase string    `json:"__index_base"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	ID        string    `json:"__id"`
}

// NewFaultCausalRelationStore 创建 FaultCausalRelation 存储实例
func NewFaultCausalRelationStore(client *opensearchsdk.Client) *FaultCausalRelationStore {
	return &FaultCausalRelationStore{client: client}
}

// ========== 接口实现 ==========

// Upsert 插入或更新故障因果关系
// 如果关系已存在则更新，否则插入
// 使用 RelationID 作为文档ID
func (s *FaultCausalRelationStore) Upsert(ctx context.Context, fcr domain.FaultCausalRelation) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultCausalRelationStore.Upsert",
			"index", faultCausalRelationIndex,
			"document_id", fcr.RelationID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if err := s.validateClient(); err != nil {
		return err
	}

	if fcr.RelationID == "" {
		return errors.New("relation_id 不能为空")
	}

	ts := fcr.RelationCreateTime
	if ts.IsZero() {
		ts = time.Now()
	}

	doc := FaultCausalRelationDocument{
		FaultCausalRelation: fcr,
		Timestamp:           ts,
		WriteTime:           time.Now().Local(),
		DataType:            faultCausalRelationIndexBase,
		IndexBase:           faultCausalRelationIndexBase,
		Category:            "log",
		Type:                faultCausalRelationIndexBase,
		ID:                  fcr.RelationID,
	}

	body, err := encodeBody(doc)
	if err != nil {
		return errors.Wrapf(err, "序列化 FaultCausalRelation 失败")
	}

	req := opensearchapi.IndexRequest{
		Index:      faultCausalRelationIndex,
		DocumentID: fcr.RelationID,
		Body:       body,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrapf(err, "写入 FaultCausalRelation 失败")
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

// Update 更新故障因果关系信息
// 使用 RelationID 作为文档ID进行部分更新
// 只更新可修改的字段，保留创建时间
func (s *FaultCausalRelationStore) Update(ctx context.Context, fcr domain.FaultCausalRelation) error {
	if err := s.validateClient(); err != nil {
		return err
	}

	if fcr.RelationID == "" {
		return errors.New("relation_id 不能为空")
	}

	// 构建更新文档（只更新可修改的字段）
	doc := map[string]any{
		"relation_update_time": fcr.RelationUpdateTime,
	}
	// 更新其他可修改字段（如果提供了值）
	updateIfNotEmpty := func(key string, value string) {
		if value != "" {
			doc[key] = value
		}
	}

	updateIfNotEmpty("relation_class", fcr.RelationClass)
	updateIfNotEmpty("source_object_id", fcr.SourceObjectID)
	updateIfNotEmpty("source_object_class", fcr.SourceObjectClass)
	updateIfNotEmpty("target_object_id", fcr.TargetObjectID)
	updateIfNotEmpty("target_object_class", fcr.TargetObjectClass)

	return s.partialUpdate(ctx, fcr.RelationID, doc)
}

// QueryByIDs 根据关系 ID 列表查询故障因果关系信息
// ids: 关系 ID 列表（RelationID）
func (s *FaultCausalRelationStore) QueryByIDs(ctx context.Context, ids []string) ([]domain.FaultCausalRelation, error) {
	if err := s.validateClient(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	// 过滤掉空字符串的 ID
	validIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			validIDs = append(validIDs, id)
		}
	}

	if len(validIDs) == 0 {
		return nil, nil
	}

	body, err := encodeBody(map[string]any{"ids": validIDs})
	if err != nil {
		return nil, errors.Wrapf(err, "构建查询请求体失败")
	}

	req := opensearchapi.MgetRequest{
		Index: faultCausalRelationIndex,
		Body:  body,
	}

	return s.executeMGet(ctx, req)
}

// QueryByEntityPair 通过实体对查询关系
// 支持两个方向：sourceID -> targetID 或 targetID -> sourceID
func (s *FaultCausalRelationStore) QueryByEntityPair(ctx context.Context, sourceID, targetID string) ([]domain.FaultCausalRelation, error) {
	if err := s.validateClient(); err != nil {
		return nil, err
	}

	if sourceID == "" || targetID == "" {
		return nil, errors.New("source_id、target_id 不能为空")
	}

	// 构建 OpenSearch 查询
	// 查询条件：(source_object_id == sourceID 或 target_object_id == sourceID)
	//           或 (source_object_id == targetID 或 target_object_id == targetID)
	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					// 条件1: source_object_id == sourceID 或 target_object_id == sourceID
					{
						"bool": map[string]any{
							"should": []map[string]any{
								{"term": map[string]any{"source_object_id": sourceID}},
								{"term": map[string]any{"target_object_id": sourceID}},
							},
							"minimum_should_match": 1, // 至少满足一个条件
						},
					},
					// 条件2: source_object_id == targetID 或 target_object_id == targetID
					{
						"bool": map[string]any{
							"should": []map[string]any{
								{"term": map[string]any{"source_object_id": targetID}},
								{"term": map[string]any{"target_object_id": targetID}},
							},
							"minimum_should_match": 1, // 至少满足一个条件
						},
					},
				},
				"minimum_should_match": 1, // 至少满足一个外层 should 条件
			},
		},
		"size": maxQuerySize,
	}

	body, err := encodeBody(query)
	if err != nil {
		return nil, errors.Wrapf(err, "构建查询请求体失败")
	}

	req := opensearchapi.SearchRequest{
		Index: []string{faultCausalRelationIndex},
		Body:  body,
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrapf(err, "查询 FaultCausalRelation 失败")
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
		return nil, errors.Wrapf(err, "读取响应失败")
	}

	result, err := decodeSearch[domain.FaultCausalRelation](data)
	if err != nil {
		return nil, errors.Wrapf(err, "解析 FaultCausalRelation 响应失败")
	}

	return result, nil
}

// ========== 私有辅助函数 ==========

// validateClient 验证 OpenSearch 客户端是否已初始化
func (s *FaultCausalRelationStore) validateClient() error {
	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	return nil
}

// executeMGet 执行批量获取请求
func (s *FaultCausalRelationStore) executeMGet(ctx context.Context, req opensearchapi.MgetRequest) ([]domain.FaultCausalRelation, error) {
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrapf(err, "查询 FaultCausalRelation 失败")
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
		return nil, errors.Wrapf(err, "读取响应失败")
	}

	result, err := decodeMGet[domain.FaultCausalRelation](data)
	if err != nil {
		return nil, errors.Wrapf(err, "解析 FaultCausalRelation 响应失败")
	}

	return result, nil
}

// partialUpdate 部分更新文档
// 只更新指定的字段，不影响其他字段
func (s *FaultCausalRelationStore) partialUpdate(ctx context.Context, id string, doc map[string]any) error {
	if err := s.validateClient(); err != nil {
		return err
	}

	if id == "" {
		return errors.New("文档 ID 不能为空")
	}

	if len(doc) == 0 {
		return errors.New("更新文档不能为空")
	}

	body, err := encodeBody(map[string]any{"doc": doc})
	if err != nil {
		return errors.Wrapf(err, "序列化更新文档失败")
	}

	req := opensearchapi.UpdateRequest{
		Index:      faultCausalRelationIndex,
		DocumentID: id,
		Body:       body,
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrapf(err, "更新 FaultCausalRelation %s 失败", id)
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

// ========== 接口实现验证 ==========

var _ core.FaultCausalRelationRepository = (*FaultCausalRelationStore)(nil)
