package dip

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
)

const (
	// agentRetryDelay 重试间隔
	// agentRetryDelay = 500 * time.Millisecond // 500毫秒
	// agentAPIPathTemplate Agent API 路径模板
	agentAPIPathTemplate = "/api/agent-app/v1/app/%s/api/chat/completion" // 格式化URL路径
	// agentQueryCausal Agent 因果推理查询文本
	agentQueryCausal = "请输出分析结果" // 因果推理查询文本
	// agentQueryDescription Agent 描述生成查询文本
	agentQueryDescription = "请输出结果" // 描述生成查询文本
)

// CausalConfig 因果推理配置
type CausalConfig struct {
	AppID         string
	Authorization string
	AgentKey      string
}

// CallCausalAgent 调用智能体接口进行因果推理。
// - customQuerys: 发送给智能体的上下文（faultPointA、faultPointB、topology_relation 等）
// 返回值只包含智能体输出中的 fault_causal 数组，若返回格式不符合预期会给出错误。
// 当 JSON 解析失败时，会尝试使用正则表达式提取关键字段。
func (c *Client) CallCausalAgent(ctx context.Context, causalConfig CausalConfig, customQuerys map[string]interface{}) ([]domain.AgentCausalEdge, error) {
	// 参数验证
	if c == nil {
		return nil, errors.New("client 未初始化")
	}
	if c.httpClient == nil {
		return nil, errors.New("http client 未初始化")
	}
	if ctx == nil {
		return nil, errors.New("上下文不能为 nil")
	}
	if causalConfig.AppID == "" {
		return nil, errors.New("app_id 不能为空")
	}
	if causalConfig.AgentKey == "" {
		return nil, errors.New("agent_key 不能为空")
	}

	path := fmt.Sprintf(agentAPIPathTemplate, causalConfig.AppID)

	var result []domain.AgentCausalEdge

	// 构建请求体
	reqBody := domain.AgentRequest{
		AgentKey:     causalConfig.AgentKey,
		CustomQuerys: customQuerys,
		Query:        agentQueryCausal,
		Stream:       false,
	}

	headers := map[string]string{
		"Authorization": causalConfig.Authorization,
		"Content-Type":  "application/json",
	}

	resp, err := c.httpClient.Post(ctx, path, reqBody, headers)
	if err != nil {
		return nil, errors.Wrapf(err, "发送 agent 请求失败")
	}

	// 检查响应是否为 nil
	if resp == nil {
		return nil, errors.New("agent 响应为空")
	}

	// 检查响应状态码
	if err := resp.Error(); err != nil {
		return nil, errors.Wrapf(err, "agent 请求失败")
	}

	// 解析响应
	var agentResp domain.AgentResponse
	if err := resp.DecodeJSON(&agentResp); err != nil {
		return nil, errors.Wrapf(err, "解析 agent 响应失败")
	}

	rawText, err := extractRawTextFromResponse(&agentResp)
	if err != nil {
		return nil, err
	}

	// 返回 text 中返回的 JSON 示例："{\n  \"fault_causal\": {\n    \"source_id\": 7235881779318784,\n    \"target_id\": 7235927773446144,\n    \"confidence\": 0.85,\n    \"reason\": \"host_007 故障点 → host_008 故障点\"\n  }\n}"

	// 尝试 JSON 解析
	var payload domain.AgentCausalPayload
	if err := json.Unmarshal([]byte(rawText), &payload); err != nil {
		// JSON 解析失败，尝试使用正则表达式提取关键字段
		edges, regexErr := extractCausalWithRegex(rawText)
		if regexErr != nil {
			return nil, errors.Wrapf(err, "解析 agent 返回的因果 JSON 失败; 正则提取也失败: %v", regexErr)
		}
		if len(edges) == 0 {
			return nil, errors.New("正则表达式提取结果为空")
		}
		return edges, nil
	}

	// 验证解析后的数据是否有效
	if payload.FaultCausal.Source == 0 || payload.FaultCausal.Target == 0 {
		// 如果 JSON 解析成功但字段无效，也尝试使用正则表达式提取
		edges, regexErr := extractCausalWithRegex(rawText)
		if regexErr == nil && len(edges) > 0 {
			return edges, nil
		}
		return nil, errors.New("agent 返回结果中 source_id 或 target_id 无效")
	}

	// 将单个对象转换为数组格式返回
	result = []domain.AgentCausalEdge{payload.FaultCausal}

	return result, nil
}

// extractRawTextFromResponse 从 Agent 响应中提取原始文本
func extractRawTextFromResponse(resp *domain.AgentResponse) (string, error) {
	if resp == nil {
		return "", errors.New("agent 响应为空")
	}

	// 检查嵌套字段是否存在
	rawText := resp.Message.Content.FinalAnswer.Answer.Text
	if rawText == "" {
		return "", errors.New("agent 响应中 final_answer.answer.text 为空")
	}

	return rawText, nil
}

// extractCausalWithRegex 使用正则表达式从文本中提取 fault_causal（仅支持对象格式）
func extractCausalWithRegex(text string) ([]domain.AgentCausalEdge, error) {
	if text == "" {
		return nil, errors.New("输入文本为空")
	}

	// 匹配对象格式：fault_causal: { source_id, target_id, ... }
	objectPattern := regexp.MustCompile(`(?i)"fault_causal"\s*:\s*\{([^}]+)\}`)
	objectMatch := objectPattern.FindStringSubmatch(text)
	if len(objectMatch) < 2 {
		return nil, errors.New("未找到 fault_causal 对象")
	}

	// 提取单个因果边
	edge := extractCausalEdgeFromObject(objectMatch[1])
	if edge.Source == 0 || edge.Target == 0 {
		return nil, errors.New("无法从文本中提取有效的 fault_causal 对象（source_id 或 target_id 无效）")
	}

	return []domain.AgentCausalEdge{edge}, nil
}

// extractCausalEdgeFromObject 从对象字符串中提取因果边（仅支持 source_id 和 target_id）
func extractCausalEdgeFromObject(objStr string) domain.AgentCausalEdge {
	edge := domain.AgentCausalEdge{}

	// 提取 source_id
	sourceIDPattern := regexp.MustCompile(`(?i)"source_id"\s*:\s*"?([^",}\]]+)"?`)
	if match := sourceIDPattern.FindStringSubmatch(objStr); len(match) > 1 {
		if sourceID, err := parseUint64(match[1]); err == nil {
			edge.Source = sourceID
		}
	}

	// 提取 target_id
	targetIDPattern := regexp.MustCompile(`(?i)"target_id"\s*:\s*"?([^",}\]]+)"?`)
	if match := targetIDPattern.FindStringSubmatch(objStr); len(match) > 1 {
		if targetID, err := parseUint64(match[1]); err == nil {
			edge.Target = targetID
		}
	}

	// 提取 confidence（支持字符串和数字格式，支持科学计数法）
	confidencePattern := regexp.MustCompile(`(?i)"confidence"\s*:\s*"?([0-9]+\.?[0-9]*(?:[eE][+-]?[0-9]+)?)"?`)
	if match := confidencePattern.FindStringSubmatch(objStr); len(match) > 1 {
		if conf, err := parseFloat64(match[1]); err == nil {
			edge.Confidence = conf
		}
	}

	// 提取 reason（支持多行字符串）
	reasonPattern := regexp.MustCompile(`(?i)"reason"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	if match := reasonPattern.FindStringSubmatch(objStr); len(match) > 1 {
		edge.Reason = strings.TrimSpace(decodeJSONString(match[1]))
	}

	return edge
}

// parseUint64 尝试将字符串解析为 uint64
func parseUint64(s string) (uint64, error) {
	if s == "" {
		return 0, errors.New("输入字符串为空")
	}

	s = strings.TrimSpace(s)
	// 去除首尾的引号（支持单引号和双引号）
	s = strings.Trim(s, `"'`)
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, errors.New("去除引号后字符串为空")
	}

	// 使用 strconv.ParseUint 解析
	result, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "解析 uint64 失败")
	}

	return result, nil
}

// parseFloat64 尝试将字符串解析为 float64
func parseFloat64(s string) (float64, error) {
	if s == "" {
		return 0, errors.New("输入字符串为空")
	}

	s = strings.TrimSpace(s)
	// 去除首尾的引号（支持单引号和双引号）
	s = strings.Trim(s, `"'`)
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, errors.New("去除引号后字符串为空")
	}

	// 使用 strconv.ParseFloat 解析
	result, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "解析 float64 失败")
	}

	return result, nil
}

// decodeJSONString 解码 JSON 字符串中的转义字符
func decodeJSONString(encoded string) string {
	if encoded == "" {
		return ""
	}

	var decoded string
	if err := json.Unmarshal([]byte(`"`+encoded+`"`), &decoded); err == nil {
		return decoded
	}

	// 如果解码失败，手动处理常见转义字符
	decoded = strings.ReplaceAll(encoded, "\\\"", "\"")
	decoded = strings.ReplaceAll(decoded, "\\n", "\n")
	decoded = strings.ReplaceAll(decoded, "\\t", "\t")
	decoded = strings.ReplaceAll(decoded, "\\r", "\r")
	decoded = strings.ReplaceAll(decoded, "\\\\", "\\")
	return decoded
}
