package isf

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/DE_go-lib/rest"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewUserManagementClient)

func NewUserManagementClient() dependency.UserManagementClient {
	return &userManagementClient{
		domain: "http://user-management-private:30980",
		//domain: "http://10.4.174.97:30981",
		httpClient: rest.NewHTTPClientWithOptions(rest.HttpClientOptions{
			TimeOut: 300,
		}),
	}
}
