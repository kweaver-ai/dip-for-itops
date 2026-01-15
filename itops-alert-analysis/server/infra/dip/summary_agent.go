package dip

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"github.com/pkg/errors"
)

// SummaryConfig 描述生成配置
type SummaryConfig struct {
	AppID         string
	Authorization string
	AgentKey      string
}

// extractDescriptionWithRegex 使用正则表达式从文本中提取 description 和 impact 字段
func extractDescriptionWithRegex(text string) (domain.Occurrence, error) {
	if text == "" {
		return domain.Occurrence{}, errors.New("输入文本为空")
	}

	desc := domain.Occurrence{}

	// 匹配 name 字段
	namePattern := regexp.MustCompile(`(?i)"name"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	nameMatch := namePattern.FindStringSubmatch(text)
	if len(nameMatch) > 1 {
		desc.Name = strings.TrimSpace(decodeJSONString(nameMatch[1]))
	}

	// 匹配 description 字段
	descPattern := regexp.MustCompile(`(?i)"description"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	descMatch := descPattern.FindStringSubmatch(text)
	if len(descMatch) > 1 {
		desc.Description = strings.TrimSpace(decodeJSONString(descMatch[1]))
	}

	// 匹配 impact 字段
	impactPattern := regexp.MustCompile(`(?i)"impact"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	impactMatch := impactPattern.FindStringSubmatch(text)
	if len(impactMatch) > 1 {
		desc.Impact = strings.TrimSpace(decodeJSONString(impactMatch[1]))
	}

	// 验证至少有一个字段不为空
	if desc.Name == "" && desc.Description == "" && desc.Impact == "" {
		return domain.Occurrence{}, errors.New("无法从文本中提取 name、description 或 impact 字段")
	}

	return desc, nil
}

// CallSummaryAgent 调用智能体接口进行过程描述和影响生成。
// - customQuerys: 发送给智能体的上下文（problem_info 或 fault_points 等）
// 返回值包含智能体输出中的 occurrence（包含 description 和 impact），
// 若返回格式不符合预期会给出错误。
// 当 JSON 解析失败时，会尝试使用正则表达式提取关键字段。
func (c *Client) CallSummaryAgent(ctx context.Context, summaryConfig SummaryConfig, customQuerys map[string]interface{}) (domain.AgentDescriptionPayload, error) {
	// 参数验证
	if c == nil {
		return domain.AgentDescriptionPayload{}, errors.New("client 未初始化")
	}
	if c.httpClient == nil {
		return domain.AgentDescriptionPayload{}, errors.New("http client 未初始化")
	}
	if ctx == nil {
		return domain.AgentDescriptionPayload{}, errors.New("上下文不能为 nil")
	}
	if summaryConfig.AppID == "" {
		return domain.AgentDescriptionPayload{}, errors.New("app_id 不能为空")
	}
	if summaryConfig.AgentKey == "" {
		return domain.AgentDescriptionPayload{}, errors.New("agent_key 不能为空")
	}

	path := fmt.Sprintf(agentAPIPathTemplate, summaryConfig.AppID)

	var result domain.AgentDescriptionPayload

	// 构建请求体
	reqBody := domain.AgentRequest{
		AgentKey:     summaryConfig.AgentKey,
		CustomQuerys: customQuerys,
		Query:        agentQueryDescription,
		Stream:       false,
	}

	headers := map[string]string{
		"Authorization": summaryConfig.Authorization,
		"Content-Type":  "application/json",
	}

	resp, err := c.httpClient.Post(ctx, path, reqBody, headers)
	if err != nil {
		return domain.AgentDescriptionPayload{}, errors.Wrapf(err, "发送 agent 请求失败")
	}

	// 检查响应是否为 nil
	if resp == nil {
		return domain.AgentDescriptionPayload{}, errors.New("agent 响应为空")
	}

	// 检查响应状态码
	if err := resp.Error(); err != nil {
		return domain.AgentDescriptionPayload{}, errors.Wrapf(err, "agent 请求失败")
	}

	// 解析响应
	var agentResp domain.AgentResponse
	if err := resp.DecodeJSON(&agentResp); err != nil {
		return domain.AgentDescriptionPayload{}, errors.Wrapf(err, "解析 agent 响应失败")
	}

	rawText, err := extractRawTextFromResponse(&agentResp)
	if err != nil {
		return domain.AgentDescriptionPayload{}, err
	}

	// 尝试 JSON 解析
	var payload domain.AgentDescriptionPayload
	if err := json.Unmarshal([]byte(rawText), &payload); err != nil {
		// JSON 解析失败，尝试使用正则表达式提取关键字段
		occurrence, regexErr := extractDescriptionWithRegex(rawText)
		if regexErr != nil {
			return domain.AgentDescriptionPayload{}, errors.Wrapf(err, "解析 agent 返回的描述 JSON 失败; 正则提取也失败: %v", regexErr)
		}
		// 验证正则提取的结果是否有效
		if occurrence.Name == "" && occurrence.Description == "" && occurrence.Impact == "" {
			return domain.AgentDescriptionPayload{}, errors.New("正则表达式提取结果为空")
		}
		result = domain.AgentDescriptionPayload{
			Occurrence: occurrence,
		}
		return result, nil
	}

	// 验证返回的描述是否为空
	if payload.Occurrence.Name == "" && payload.Occurrence.Description == "" && payload.Occurrence.Impact == "" {
		// 如果 JSON 解析成功但字段为空，也尝试使用正则表达式提取
		occurrence, regexErr := extractDescriptionWithRegex(rawText)
		if regexErr == nil && (occurrence.Name != "" || occurrence.Description != "" || occurrence.Impact != "") {
			result = domain.AgentDescriptionPayload{
				Occurrence: occurrence,
			}
			return result, nil
		}
		return domain.AgentDescriptionPayload{}, errors.New("agent 返回结果中未包含 name、description 或 impact 字段或为空")
	}

	result = payload

	// 验证返回的结果是否有效
	if result.Occurrence.Name == "" && result.Occurrence.Description == "" && result.Occurrence.Impact == "" {
		return domain.AgentDescriptionPayload{}, errors.New("agent 返回的描述结果为空")
	}

	return result, nil
}
