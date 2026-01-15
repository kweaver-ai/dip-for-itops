package alert_analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type alertAnalysisClient struct {
	domain     string
	httpClient rest.HTTPClient
}

func (uc *alertAnalysisClient) Close(ctx context.Context, problemId, closeBy string) error {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	// 将结构体转换为JSON格式
	jsonData, err := json.Marshal(dependency.ProblemCloseBody{
		ClosedBy: closeBy,
		Notes:    "",
	})
	if err != nil {
		log.Errorf("Close Problem Error: %v", err)
		return err
	}
	url := fmt.Sprint(uc.domain, "/api/itops-alert-analysis/v1/problems/", problemId, "/close")
	respCode, respData, err := uc.httpClient.Post(ctx, url, headers, jsonData)
	if err != nil {
		log.Errorf("Close Problem post request methodError: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		return err
	}
	if respCode != 200 {
		log.Errorf("Close Problem post request methodError: %v , request url:%v,params: %v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post request method failed,request url:%v, respCode: %v,params: %v,post data: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return err
	}
	return nil
}

func (uc *alertAnalysisClient) SetRootCause(ctx context.Context, problemId, rootCauseObjectId string, rootCauseFaultID uint64) error {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	// 将结构体转换为JSON格式
	jsonData, err := json.Marshal(dependency.RootCauseObjectIdParams{
		RootCauseObjectId: rootCauseObjectId,
		RootCauseFaultID:  rootCauseFaultID,
	})
	if err != nil {
		log.Errorf("json Marshal Error: %v", err)
		return err
	}
	url := fmt.Sprint(uc.domain, "/api/itops-alert-analysis/v1/problems/", problemId, "/root-cause")
	respCode, respData, err := uc.httpClient.Post(ctx, url, headers, jsonData)
	if err != nil {
		log.Errorf("Set Problem Root Cause request methodError: %v ,params: %v, request url:%v,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		return err
	}
	if respCode != 200 {
		log.Errorf("Set Problem Root Cause request failed: %v , request url:%v ,params: %v ,post data: %v\n", err, url, bytes.NewBuffer(jsonData), respData)
		err = fmt.Errorf("Post request method failed,request url:%v, respCode: %v, params:%v,post data: %v \n", url, respCode, bytes.NewBuffer(jsonData), respData)
		return err
	}
	return nil
}
