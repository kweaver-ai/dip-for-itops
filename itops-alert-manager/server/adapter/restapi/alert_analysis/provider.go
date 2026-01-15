package alert_analysis

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/DE_go-lib/rest"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewAlertAnalysisClient)

func NewAlertAnalysisClient() dependency.AlertAnalysisClient {
	return &alertAnalysisClient{
		domain: "http://itops-alert-analysis-dip:13047",
		//domain: "http://192.168.201.15:80",
		httpClient: rest.NewHTTPClientWithOptions(rest.HttpClientOptions{
			TimeOut: 300,
		}),
	}
}
