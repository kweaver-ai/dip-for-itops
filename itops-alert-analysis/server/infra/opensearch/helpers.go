package opensearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// OpenSearchError 表示 OpenSearch 返回的错误响应结构。
type OpenSearchError struct {
	ErrorInfo struct {
		Type      string `json:"type"`
		Reason    string `json:"reason"`
		RootCause []struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
			Index  string `json:"index,omitempty"`
		} `json:"root_cause,omitempty"`
		CausedBy *struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"caused_by,omitempty"`
	} `json:"error"`
	Status int `json:"status"`
}

// Error 实现 error 接口。
func (e *OpenSearchError) Error() string {
	if e.ErrorInfo.Reason != "" {
		// 如果有 root_cause，优先显示
		if len(e.ErrorInfo.RootCause) > 0 {
			return fmt.Sprintf("[%s] %s (root: %s - %s)",
				e.ErrorInfo.Type,
				e.ErrorInfo.Reason,
				e.ErrorInfo.RootCause[0].Type,
				e.ErrorInfo.RootCause[0].Reason)
		}
		return fmt.Sprintf("[%s] %s", e.ErrorInfo.Type, e.ErrorInfo.Reason)
	}
	return fmt.Sprintf("opensearch error (status=%d)", e.Status)
}

// mgetResponse 与 searchResponse 仅用于解析 OpenSearch 响应。
type mgetResponse struct {
	Docs []struct {
		Found  bool            `json:"found"`
		Source json.RawMessage `json:"_source"`
	} `json:"docs"`
}

type searchResponse struct {
	Hits struct {
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func readResponseBody(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, errors.Wrap(err, "读取 OpenSearch 响应失败")
	}
	return data, nil
}

// formatErrorMessage 解析 OpenSearch 错误响应并返回结构化的错误信息。
func formatErrorMessage(data []byte) error {
	if len(data) == 0 {
		return errors.New("opensearch 返回空错误响应")
	}

	// 尝试解析为 OpenSearchError 结构
	var osErr OpenSearchError
	if err := json.Unmarshal(data, &osErr); err == nil && osErr.ErrorInfo.Reason != "" {
		// 成功解析，返回结构化错误
		return &osErr
	}

	// 解析失败，返回原始响应（可能是非 JSON 格式的错误）
	msg := strings.TrimSpace(string(data))
	if msg == "" {
		msg = "unknown opensearch error"
	}
	return errors.New(msg)
}

// readErrorResponse 从响应体中读取并解析错误信息。
// 这是一个便捷函数，统一处理 OpenSearch 错误响应的读取和解析。
func readErrorResponse(body io.Reader) error {
	data, err := readResponseBody(body)
	if err != nil {
		// 读取响应体失败，返回读取错误
		return errors.Wrap(err, "读取 OpenSearch 错误响应失败")
	}
	// 解析错误信息
	return formatErrorMessage(data)
}

func decodeMGet[T any](data []byte) ([]T, error) {
	var resp mgetResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, errors.Wrap(err, "解析 mget 响应失败")
	}
	items := make([]T, 0, len(resp.Docs))
	for _, doc := range resp.Docs {
		if !doc.Found || len(doc.Source) == 0 {
			continue
		}
		var item T
		if err := json.Unmarshal(doc.Source, &item); err != nil {
			return nil, errors.Wrap(err, "解析文档失败")
		}
		items = append(items, item)
	}
	return items, nil
}

func decodeSearch[T any](data []byte) ([]T, error) {
	var resp searchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, errors.Wrap(err, "解析 search 响应失败")
	}
	items := make([]T, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		if len(hit.Source) == 0 {
			continue
		}
		var item T
		if err := json.Unmarshal(hit.Source, &item); err != nil {
			return nil, errors.Wrapf(err, "解析文档失败,Source:%+v", string(hit.Source))
		}
		items = append(items, item)
	}
	return items, nil
}

func encodeBody(payload any) (*bytes.Reader, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "序列化请求体失败")
	}
	return bytes.NewReader(data), nil
}
