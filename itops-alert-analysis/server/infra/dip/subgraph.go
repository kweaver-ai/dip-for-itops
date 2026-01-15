package dip

import (
	"context"
	"fmt"
)

// Condition 过滤条件。
type Condition struct {
	Field     string      `json:"field,omitempty"`
	Operation string      `json:"operation"`
	Value     interface{} `json:"value,omitempty"`
}

// SubGraphQueryRequest 子图查询请求（基于起点、方向和路径长度）。
type SubGraphQueryRequest struct {
	SourceObjectTypeID string     `json:"source_object_type_id"`
	Condition          *Condition `json:"condition,omitempty"`
	Direction          string     `json:"direction"`
	PathLength         int        `json:"path_length"`
}

// SubGraphObjectProperties 子图对象的属性。
type SubGraphObjectProperties struct {
	SID  string `json:"s_id"`
	Name string `json:"name"`
}

// SubGraphObject 子图中的对象。
type SubGraphObject struct {
	ID             string                   `json:"id"`
	ObjectTypeID   string                   `json:"object_type_id"`
	ObjectTypeName string                   `json:"object_type_name"`
	Display        string                   `json:"display"`
	Properties     SubGraphObjectProperties `json:"properties"`
}

// SubGraphResponse 子图查询响应。
type SubGraphResponse struct {
	Objects map[string]SubGraphObject `json:"objects"`
}

// QuerySubGraph 查询对象子图（基于起点、方向和路径长度）。
func (c *Client) QuerySubGraph(ctx context.Context, req SubGraphQueryRequest) (*SubGraphResponse, error) {
	var (
		path   = fmt.Sprintf("/api/ontology-query/v1/knowledge-networks/%s/subgraph", c.KnID())
		result SubGraphResponse
	)

	resp, err := c.httpClient.Post(ctx, path, req, nil)
	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	if err := resp.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
