package service

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/adapter/restapi/hydra"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewAesService, NewConfigService, NewProblemService, NewAuthVerifyService)

func NewProblemService(uniQueryClient dependency.UniQueryClient, alertAnalysisClient dependency.AlertAnalysisClient,
	userManagementClient dependency.UserManagementClient, knowledgeNetworkClient dependency.KnowledgeNetworkClient,
	configService ConfigService) ProblemService {
	return &problemService{
		uniQueryClient:         uniQueryClient,
		alertAnalysisClient:    alertAnalysisClient,
		userManagementClient:   userManagementClient,
		knowledgeNetworkClient: knowledgeNetworkClient,
		configService:          configService,
	}
}
func NewAuthVerifyService() AuthVerifyService {
	return &authVerifyService{hydra: hydra.NewHydra()}
}

func NewConfigService(configRepo dependency.ConfigRepo, aes AesService) ConfigService {
	return &configService{
		configRepo: configRepo,
		aes:        aes,
	}
}

func NewAesService() AesService {
	return &aesService{Key: []byte("b279d!4zbne4ut*5")}
}
