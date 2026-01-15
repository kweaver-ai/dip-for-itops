package opensearch

import (
	"context"
	"fmt"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	opensearchsdk "github.com/opensearch-project/opensearch-go/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// FaultPointRelationStore 实现 FaultPointRelationRepository。
type FaultPointRelationStore struct {
	client *opensearchsdk.Client
}

// problemDocument 包装 Problem 并补充索引所需的公共字段。
type faultPointRelationDocument struct {
	domain.FaultPointRelation
	Timestamp time.Time `json:"@timestamp"`
	WriteTime time.Time `json:"__write_time"`
	DataType  string    `json:"__data_type"`
	IndexBase string    `json:"__index_base"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	ID        string    `json:"__id"`
}

// NewFaultPointRelationStore 创建故障点关系存储实例。
func NewFaultPointRelationStore(client *opensearchsdk.Client) core.FaultPointRelationRepository {
	return &FaultPointRelationStore{client: client}
}

// Upsert 创建或更新故障点关系。
func (s *FaultPointRelationStore) Upsert(ctx context.Context, relation domain.FaultPointRelation) error {
	defer func(start time.Time) {
		log.Debugw("OpenSearch",
			"operation", "FaultPointRelationStore.Upsert",
			"index", FaultPointRelationIndex,
			"document_id", relation.RelationId,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	if s.client == nil {
		return errors.New("opensearch client 未初始化")
	}

	doc := faultPointRelationDocument{
		FaultPointRelation: relation,
		Timestamp:          relation.RelationCreateTime,
		WriteTime:          time.Now().Local(),
		DataType:           FaultPointRelationIndexBase,
		IndexBase:          FaultPointRelationIndexBase,
		Category:           "log",
		Type:               FaultPointRelationIndexBase,
		ID:                 cast.ToString(relation.RelationId),
	}

	body, err := encodeBody(doc)
	if err != nil {
		return errors.Wrap(err, "序列化故障点关系失败")
	}

	// 使用 relation_id 作为文档 ID
	docID := fmt.Sprintf("%d", relation.RelationId)

	resp, err := s.client.Index(
		FaultPointRelationIndex,
		body,
		s.client.Index.WithContext(ctx),
		s.client.Index.WithDocumentID(docID),
		s.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return errors.Wrap(err, "写入故障点关系失败")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.IsError() {
		err := readErrorResponse(resp.Body)
		return errors.Wrapf(err, "写入故障点关系失败，状态码: %d", resp.StatusCode)
	}

	return nil
}
