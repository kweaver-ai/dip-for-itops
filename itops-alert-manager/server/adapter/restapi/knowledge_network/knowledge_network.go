package knowledge_network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type knowledgeNetworkClient struct {
	domain     string
	httpClient rest.HTTPClient
}

// Subgraph 子图查询
func (c *knowledgeNetworkClient) SubGraphQuery(ctx context.Context, queryReq dependency.SubGraphQueryRequest, authorization, knowledgeId string) (*dependency.SubGraphQueryResponse, error) {
	result := &dependency.SubGraphQueryResponse{}
	headers := map[string]string{
		"Content-Type":           "application/json",
		"x-http-method-override": "GET",
		"authorization":          authorization,
	}
	// 将结构体转换为JSON格式
	jsonData, err := json.Marshal(queryReq)
	if err != nil {
		log.Errorf("Sub GraphQuery Error: %v", err)
		return result, err
	}
	url := fmt.Sprintf("%s/api/ontology-query/v1/knowledge-networks/%s/subgraph", c.domain, knowledgeId)
	respCode, respData, err := c.httpClient.Post(ctx, url, headers, jsonData)
	if err != nil {
		log.Errorf("search Subgraph post request methodError: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		return result, err
	}
	if respCode != 200 {
		log.Errorf("search Subgraph post request methodError: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post request method failed,request url:%v, respCode: %v,params: %v,post data: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return result, err
	}
	respDataByte, err := json.Marshal(respData)
	if err != nil {
		log.Errorf("search Subgraph post request parseError1: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post request parse1 failed,request url:%v, respCode: %v,params: %v,post data: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return result, err
	}
	if err := json.Unmarshal(respDataByte, result); err != nil {
		log.Errorf("search Subgraph post request parseError2: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post request parse2 failed,request url:%v, respCode: %v,params: %v,post data: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return result, err
	}
	return result, nil
}

// Subgraph 子图查询
func (c *knowledgeNetworkClient) ObjectInfoQuery(ctx context.Context, queryReq dependency.ObjectInfoQueryRequest, entityObjectClass, authorization, knowledgeId string) (*dependency.ObjectInfoQueryResponse, error) {
	result := &dependency.ObjectInfoQueryResponse{}
	headers := map[string]string{
		"Content-Type":           "application/json",
		"x-http-method-override": "GET",
		"authorization":          authorization,
	}
	// 将结构体转换为JSON格式
	jsonData, err := json.Marshal(queryReq)
	if err != nil {
		log.Errorf("Object Info Query Error: %v", err)
		return result, err
	}
	url := fmt.Sprintf("%s/api/ontology-query/v1/knowledge-networks/%s/object-types/%s", c.domain, knowledgeId, entityObjectClass)
	respCode, respData, err := c.httpClient.Post(ctx, url, headers, jsonData)
	if err != nil {
		log.Errorf("search Object Info post request methodError: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		return result, err
	}
	if respCode != 200 {
		log.Errorf("search Object Info post request methodError: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post Object Info request method failed,request url:%v, respCode: %v,params: %v,post data: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return result, err
	}
	respDataByte, err := json.Marshal(respData)
	if err != nil {
		log.Errorf("search Object Info post request parseError1: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post Object Info request parse1 failed,request url:%v, respCode: %v,params: %v,post data: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return result, err
	}
	if err := json.Unmarshal(respDataByte, result); err != nil {
		log.Errorf("search Object Info post request parseError2: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post Object Info request parse2 failed,request url:%v, respCode: %v,params: %v,post data: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return result, err
	}
	return result, nil
}
