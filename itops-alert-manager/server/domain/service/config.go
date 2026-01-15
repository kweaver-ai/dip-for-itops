package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/dependency"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/entity"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/vo"
	"github.com/mitchellh/mapstructure"
)

//go:generate mockgen -source ./problem.go -destination ../../mock/service/mock_config_service.go -package mock
type ConfigService interface {
	CreateConfig(ctx context.Context, req *vo.ConfigReq) core.ServiceError
	UpdateConfig(ctx context.Context, req *vo.ConfigReq) core.ServiceError
	ListConfigs(ctx context.Context, isIn bool) (vo.ConfigReq, core.ServiceError)
}

type configService struct {
	configRepo dependency.ConfigRepo
	aes        AesService
}

// CreateConfig 创建配置
func (s *configService) CreateConfig(ctx context.Context, req *vo.ConfigReq) core.ServiceError {

	if req.Platform.AuthToken != "" {
		encrypted, err := s.aes.AESEncrypt([]byte(req.Platform.AuthToken))
		if err != nil {
			fmt.Printf("Encryption failed: %v\n", err)
			return NewSvcInternalError(nil)
		}
		req.Platform.AuthToken = encrypted
	}
	reqData := map[string]interface{}{
		"platform":           req.Platform,
		"knowledge_network":  req.KnowledgeNetwork,
		"fault_point_policy": req.FaultPointPolicy,
		"problem_policy":     req.ProblemPolicy,
	}
	for key, value := range reqData {
		valueByte, err := json.Marshal(value)
		if err != nil {
			log.Errorf("Failed to create config: %v", err)
			return NewSvcInternalError(nil)
		}
		config := &entity.Config{
			ConfigKey:   key,
			ConfigValue: string(valueByte),
		}
		if err := s.configRepo.Create(ctx, config); err != nil {
			log.Errorf("Failed to create config: %v", err)
			return NewSvcInternalError(err)
		}
	}
	return nil
}

// UpdateConfig 更新配置
func (s *configService) UpdateConfig(ctx context.Context, req *vo.ConfigReq) core.ServiceError {
	reqData := map[string]interface{}{
		"knowledge_network":  req.KnowledgeNetwork,
		"fault_point_policy": req.FaultPointPolicy,
		"problem_policy":     req.ProblemPolicy,
	}
	if req.Platform.AuthToken != "" {
		encrypted, err := s.aes.AESEncrypt([]byte(req.Platform.AuthToken))
		if err != nil {
			fmt.Printf("Encryption failed: %v\n", err)
			return NewSvcInternalError(nil)
		}
		req.Platform.AuthToken = encrypted
		reqData["platform"] = req.Platform
	}

	for key, value := range reqData {
		valueByte, err := json.Marshal(value)
		if err != nil {
			log.Errorf("Failed to update config: %v", err)
			return NewSvcInternalError(nil)
		}
		config := &entity.Config{
			ConfigKey:   key,
			ConfigValue: string(valueByte),
		}

		if err := s.configRepo.Update(ctx, config); err != nil {
			log.Errorf("Failed to update config: %v", err)
			return NewSvcInternalError(err)
		}
	}
	return nil
}

// ListConfigs 获取所有配置
func (s *configService) ListConfigs(ctx context.Context, isIn bool) (vo.ConfigReq, core.ServiceError) {
	result := vo.ConfigReq{}
	configs, err := s.configRepo.ListAll(ctx)
	if err != nil {
		log.Errorf("Failed to list configs: %v", err)
		return result, NewSvcInternalError(err)
	}
	inputMap := make(map[string]interface{})
	for _, c := range configs {
		var parsedValue interface{}
		if err := json.Unmarshal([]byte(c.ConfigValue), &parsedValue); err != nil {
			log.Errorf("Warning: Could not parse value for key '%s'", c.ConfigKey, err)
			return result, NewSvcInternalError(nil)
		}
		inputMap[c.ConfigKey] = parsedValue
	}

	if err := mapstructure.Decode(inputMap, &result); err != nil {
		log.Warnf("error decoding map to struct: %v", err)
	}

	if result.Platform.AuthToken != "" {
		decrypted, err := s.aes.AESDecrypt(result.Platform.AuthToken)
		if err != nil {
			fmt.Printf("Decrypted failed: %v\n", err)
			return result, NewSvcInternalError(nil)
		}
		result.Platform.AuthToken = string(decrypted)
		if len(result.Platform.AuthToken) > 8 && isIn == false {
			result.Platform.AuthToken = strings.Repeat("*", len(result.Platform.AuthToken)-8) + result.Platform.AuthToken[len(result.Platform.AuthToken)-8:]
		}

	}
	return result, nil
}
