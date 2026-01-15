package isf

import (
	"context"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"encoding/json"
	"fmt"
	"net/url"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/DE_go-lib/rest"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
)

type userManagementClient struct {
	domain     string
	httpClient rest.HTTPClient
}

func (um *userManagementClient) GetUserInfo(ctx context.Context, accountId string) (dependency.ISFUserInfo, error) {
	resp := dependency.ISFUserInfo{}
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	getUserInfoUrl := fmt.Sprint(um.domain, "/api/user-management/v1/users/", accountId, "/account")
	respCode, respData, err := um.httpClient.Get(ctx, getUserInfoUrl, url.Values{}, headers)
	if err != nil {
		log.Errorf("request isf User Info methodError: %v , request url:%v,post data: %v\n", err, getUserInfoUrl, respCode)
		return resp, err
	}
	if respCode != 200 {
		log.Errorf("request isf User Info failed: %v , request url:%v,post data: %v\n", err, getUserInfoUrl, respData)
		err = fmt.Errorf("request isf User Info failed,request url:%v, respCode: %v, jsonData: %v \n", getUserInfoUrl, respCode, respData)
		return resp, err
	}
	respJson, err := json.Marshal(respData)
	if err != nil {
		log.Errorf("json Marshal, error: %v \n", err)
		return resp, err
	}
	result := make([]dependency.ISFUserInfo, 0)
	err = json.Unmarshal(respJson, &result)
	if err != nil {
		log.Errorf("json Unmarshal, error: %v \n", err)
		return resp, err
	}
	return result[0], nil
}
