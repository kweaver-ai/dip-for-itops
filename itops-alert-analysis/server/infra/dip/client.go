package dip

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	httputil "devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/http"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
)

// Client DIP 客户端
type Client struct {
	knID       string // 默认 KnowledgeID（getKnID 为 nil 时使用）
	httpClient *httputil.Client
	getKnID    func() string // 动态获取 KnowledgeID
}

// NewClient 创建 DIP 客户端实例。
// getAuth: 动态获取 Authorization
// getKnID: 动态获取 KnowledgeID（可选，为 nil 时使用 cfg.KnID）
func NewClient(cfg config.DIPConfig, getAuth func() string, getKnID func() string) *Client {
	headers := map[string]string{
		"User-Agent": "itops-alert-analysis",
	}
	httpClient := httputil.NewClient(httputil.Config{
		BaseURL:            cfg.Host,
		Timeout:            cfg.Timeout,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		Headers:            headers,
	}, getAuth).WithLogger(log.Logger)

	return &Client{
		knID:       cfg.KnID,
		httpClient: httpClient,
		getKnID:    getKnID,
	}
}

// KnID 获取当前 KnowledgeID（优先使用动态函数）
func (c *Client) KnID() string {
	if c.getKnID != nil {
		return c.getKnID()
	}
	return c.knID
}
