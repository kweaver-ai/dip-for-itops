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

// FaultCausalStore 负责 itops_fault_causal 索引的存储操作
// 实现 core.FaultCausalRepository 接口
type FaultCausalStore struct {
	client *opensearchsdk.Client
}

// FaultCausalDocument 包装 FaultCausalObject 并补充索引所需的公共字段
type FaultCausalDocument struct {
	domain.FaultCausalObject
	Timestamp time.Time `json:"@timestamp"`
	WriteTime time.Time `json:"__write_time"`
	DataType  string    `json:"__data_type"`
	IndexBase string    `json:"__index_base"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	ID        string    `json:"__id"`
}

// NewFaultCausalStore 创建 FaultCausal 存储实例
func NewFaultCausalStore(client *opensearchsdk.Client) *FaultCausalStore {
	return &FaultCausalStore{client: client}
}

// ========== 接口实现 ==========

// Upsert 插入或更新故障因果实体
// 如果实体已存在则更新，否则插入
// 使用 CausalID 作为文档ID
func (s *FaultCausalStore) Upsert(ctx context.Context, fc domain.FaultCausalObject) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultCausalStore.Upsert",
			"index", faultCausalObjectIndex,
			"document_id", fc.CausalID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	if fc.CausalID == "" {
		return errors.New("causal_id 不能为空")
	}

	// 确定 Timestamp：优先使用 SCreateTime，如果为零值则使用当前时间
	ts := fc.SCreateTime
	if ts.IsZero() {
		ts = time.Now()
	}

	doc := FaultCausalDocument{
		FaultCausalObject: fc,
		Timestamp:         ts,
		WriteTime:         time.Now().Local(),
		DataType:          faultCausalObjectIndexBase,
		IndexBase:         faultCausalObjectIndexBase,
		Category:          "log",
		Type:              faultCausalObjectIndexBase,
		ID:                fc.CausalID,
	}

	body, err := encodeBody(doc)
	if err != nil {
		return errors.Wrapf(err, "序列化 FaultCausalObject 失败")
	}

	req := opensearchapi.IndexRequest{
		Index:      faultCausalObjectIndex,
		DocumentID: fc.CausalID,
		Body:       body,
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrapf(err, "写入 FaultCausalObject 失败")
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

// Update 更新故障因果实体信息
// 使用 CausalID 作为文档ID进行部分更新
// 只更新可修改的字段，保留创建时间
func (s *FaultCausalStore) Update(ctx context.Context, fc domain.FaultCausalObject) error {
	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}
	if fc.CausalID == "" {
		return errors.New("causal_id 不能为空")
	}

	// 构建更新文档（只更新可修改的字段，不更新创建时间和ID）
	doc := map[string]any{
		"causal_confidence": fc.CausalConfidence,
		"causal_reason":     fc.CausalReason,
		"s_update_time":     fc.SUpdateTime,
	}

	// 注意：SCreateTime 不应该在更新时修改，保持创建时间不变
	// 如果需要修复数据，应该使用 Upsert 方法

	return s.partialUpdate(ctx, fc.CausalID, doc)
}

// QueryByIDs 根据因果推理 ID 列表查询故障因果实体信息
// ids: 因果推理 ID 列表（CausalID）
func (s *FaultCausalStore) QueryByIDs(ctx context.Context, ids []string) ([]domain.FaultCausalObject, error) {
	if s.client == nil {
		return nil, errors.New("opensearch client 未初始化")
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
		Index: faultCausalObjectIndex,
		Body:  body,
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrapf(err, "mget FaultCausalObject 失败")
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

	result, err := decodeMGet[domain.FaultCausalObject](data)
	if err != nil {
		return nil, errors.Wrapf(err, "解析 FaultCausalObject 响应失败")
	}

	return result, nil
}

// ========== 私有辅助函数 ==========

// partialUpdate 部分更新文档
// 只更新指定的字段，不影响其他字段
func (s *FaultCausalStore) partialUpdate(ctx context.Context, id string, doc map[string]any) error {
	if s.client == nil {
		return errors.New("opensearch client 未初始化")
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
		Index:      faultCausalObjectIndex,
		DocumentID: id,
		Body:       body,
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrapf(err, "更新 FaultCausal %s 失败", id)
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

var _ core.FaultCausalRepository = (*FaultCausalStore)(nil)
