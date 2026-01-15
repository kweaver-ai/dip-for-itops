package dip

import (
	"context"
	"fmt"
)

// ObjectInstance 对象实例基本信息
type ObjectInstance map[string]interface{}

// ObjectDataResponse 对象数据响应。
type ObjectDataResponse struct {
	Datas       []ObjectInstance `json:"datas"`
	TotalCount  int              `json:"total_count"`
	SearchAfter []interface{}    `json:"search_after"`
}

// QueryRequest 查询请求体。
type QueryRequest struct {
	NeedTotal   bool          `json:"need_total,omitempty"`
	Limit       int           `json:"limit,omitempty"`
	SearchAfter []interface{} `json:"search_after,omitempty"`
}

// QueryObjectData 查询指定对象类的对象详细数据
func (c *Client) QueryObjectData(ctx context.Context, otID string, limit int, searchAfter []interface{}) (*ObjectDataResponse, error) {
	path := fmt.Sprintf("/api/ontology-query/v1/knowledge-networks/%s/object-types/%s", c.KnID(), otID)

	reqBody := QueryRequest{
		Limit: limit,
	}
	if len(searchAfter) > 0 {
		reqBody.SearchAfter = searchAfter
	} else {
		reqBody.NeedTotal = true
	}
	c.httpClient.SetHeader("X-HTTP-Method-Override", "GET")
	resp, err := c.httpClient.Post(ctx, path, reqBody, nil)
	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	var result ObjectDataResponse
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// QueryAllObjectData 查询指定对象类的所有对象数据
func (c *Client) QueryAllObjectData(ctx context.Context, otID string, limit int) ([]ObjectInstance, error) {
	var allData []ObjectInstance
	var searchAfter []interface{}

	for {
		resp, err := c.QueryObjectData(ctx, otID, limit, searchAfter)
		if err != nil {
			return nil, err
		}

		allData = append(allData, resp.Datas...)

		// 如果没有更多数据，退出循环
		if len(resp.Datas) == 0 || len(resp.SearchAfter) == 0 {
			break
		}

		searchAfter = resp.SearchAfter
	}

	return allData, nil
}
