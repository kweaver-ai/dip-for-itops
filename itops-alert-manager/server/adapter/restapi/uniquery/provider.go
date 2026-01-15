package uniquery

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"github.com/google/wire"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

var ProviderSet = wire.NewSet(NewHTTPClient, NewUniQueryClient)

func NewHTTPClient() rest.HTTPClient {
	return rest.NewHTTPClientWithOptions(rest.HttpClientOptions{
		TimeOut: 300,
	})
}
func NewUniQueryClient(httpClient rest.HTTPClient) dependency.UniQueryClient {
	return &uniQueryClient{
		restapi:    core.NewCoreRestAPI(),
		httpClient: httpClient,
	}
}
