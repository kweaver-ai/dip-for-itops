package dip

import (
	"context"
	"fmt"
)

// ObjectType 对象类列表
type ObjectType struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
	Comment string   `json:"comment"`
}

// ObjectTypeListResponse 对象类列表
type ObjectTypeListResponse struct {
	Entries    []ObjectType `json:"entries"`
	TotalCount int          `json:"total_count"`
}

// GetObjectTypes 获取对象类列表。
func (c *Client) GetObjectTypes(ctx context.Context) ([]ObjectType, error) {
	path := fmt.Sprintf("/api/ontology-manager/v1/knowledge-networks/%s/object-types", c.KnID())

	resp, err := c.httpClient.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	var result ObjectTypeListResponse
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return result.Entries, nil
}
