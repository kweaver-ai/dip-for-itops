package uniquery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/vo"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type uniQueryClient struct {
	restapi    core.RestAPI
	httpClient rest.HTTPClient
}

func (uc *uniQueryClient) GetDataView(ctx context.Context, viewId string, req vo.DataViewQueryV2, accout_id string) (vo.ViewUniResponseV2, error) {
	resp := vo.ViewUniResponseV2{}
	headers := map[string]string{
		"Content-Type":           "application/json",
		"x-account-id":           accout_id,
		"x-account-type":         "user",
		"x-http-method-override": "GET",
	}
	req.ViewQueryCommonParams.NeedTotal = true
	// 将结构体转换为JSON格式
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Error marshaling data: %v", err)
		return resp, err
	}
	url := uc.restapi.RestAPI().UniQueryDomain + "/api/mdl-uniquery/in/v1/data-views/" + viewId + "?include_view=true&timeout=5m"
	respCode, respData, err := uc.httpClient.Post(ctx, url, headers, jsonData)
	if err != nil {
		log.Errorf("post request method error: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		return resp, err
	}
	if respCode != 200 {
		log.Errorf("Post request method failed,request url:%v, respCode: %v,params: %v, jsonData: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post request method failed,request url:%v, respCode: %v ,params: %v, jsonData: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return resp, err
	}
	respJson, err := json.Marshal(respData)
	if err != nil {
		log.Errorf("json Marshal, error: %v \n", err)
		return resp, err
	}
	err = json.Unmarshal(respJson, &resp)
	if err != nil {
		log.Errorf("json Unmarshal, error: %v \n", err)
		return resp, err
	}
	return resp, nil
}
