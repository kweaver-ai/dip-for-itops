package knowledge_network

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"github.com/google/wire"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

var ProviderSet = wire.NewSet(NewKnowledgeNetworkClient)

func NewKnowledgeNetworkClient() dependency.KnowledgeNetworkClient {
	return &knowledgeNetworkClient{
		domain: "https://nginx-ingress-class-443:443",
		//domain: "https://192.168.201.15",
		httpClient: rest.NewHTTPClientWithOptions(rest.HttpClientOptions{
			TimeOut: 300,
		}),
	}
}
