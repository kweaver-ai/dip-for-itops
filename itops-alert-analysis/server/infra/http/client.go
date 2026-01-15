package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"github.com/pkg/errors"
)

// Client 通用 HTTP 客户端封装。
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	logger     *log.Log      // 实例级别的 logger
	getAuth    func() string // 动态获取 Authorization
}

// Config HTTP 客户端配置。
type Config struct {
	BaseURL            string            // 基础 URL
	Timeout            time.Duration     // 请求超时时间
	Headers            map[string]string // 默认 Header
	InsecureSkipVerify bool              // 是否跳过SSL验证
}

// NewClient 创建 HTTP 客户端实例。
func NewClient(cfg Config, getAuth func() string) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	transport := &http.Transport{}
	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		headers: cfg.Headers,
		getAuth: getAuth,
	}
}

func (c *Client) WithLogger(logger *log.Log) *Client {
	c.logger = logger
	return c
}

// Request 通用请求结构。
type Request struct {
	Method  string            // HTTP 方法：GET, POST, PUT, DELETE 等
	Path    string            // 请求路径（相对于 BaseURL）
	Headers map[string]string // 额外的 Header（会合并到默认 Header）
	Body    interface{}       // 请求体（会自动序列化为 JSON）
}

// Response 通用响应结构。
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Do 执行 HTTP 请求。
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	// 构建完整 URL
	url := c.baseURL + req.Path

	// 序列化请求体
	var bodyReader io.Reader
	var requestBodyBytes []byte
	if req.Body != nil {
		var err error
		requestBodyBytes, err = json.Marshal(req.Body)
		if err != nil {
			return nil, errors.Wrap(err, "序列化请求体失败")
		}
		bodyReader = bytes.NewReader(requestBodyBytes)
	}

	// Debug 日志：在 defer 中统一记录请求/响应/耗时
	var statusCode int
	var respBody []byte
	defer func(start time.Time) {
		if c.logger == nil {
			return
		}

		c.logger.Debugw("HTTP",
			"method", req.Method,
			"url", url,
			"request_body", string(requestBodyBytes),
			"status_code", statusCode,
			"response_body", string(respBody),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}(time.Now())

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, "创建请求失败")
	}

	// 设置默认 Header
	for key, value := range c.headers {
		httpReq.Header.Set(key, value)
	}

	// 设置额外 Header（会覆盖默认 Header）
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 如果有请求体，设置 Content-Type
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 动态获取最新的 Authorization
	if c.getAuth != nil {
		auth := c.getAuth()
		if auth != "" {
			httpReq.Header.Set("Authorization", auth)
		}
	}

	// 执行请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "请求失败")
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	// 读取响应体
	respBody, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "读取响应失败")
	}
	statusCode = httpResp.StatusCode

	return &Response{
		StatusCode: httpResp.StatusCode,
		Body:       respBody,
		Headers:    httpResp.Header,
	}, nil
}

// Get 执行 GET 请求。
func (c *Client) Get(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method:  http.MethodGet,
		Path:    path,
		Headers: headers,
	})
}

// Post 执行 POST 请求。
func (c *Client) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method:  http.MethodPost,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Put 执行 PUT 请求。
func (c *Client) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method:  http.MethodPut,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Delete 执行 DELETE 请求。
func (c *Client) Delete(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method:  http.MethodDelete,
		Path:    path,
		Headers: headers,
	})
}

// DecodeJSON 将响应体解析为 JSON。
func (r *Response) DecodeJSON(v interface{}) error {
	if err := json.Unmarshal(r.Body, v); err != nil {
		return errors.Wrap(err, "解析 JSON 失败")
	}
	return nil
}

// IsSuccess 检查响应是否成功（2xx）。
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// Error 返回错误响应信息。
func (r *Response) Error() error {
	if r.IsSuccess() {
		return nil
	}
	return errors.Errorf("请求失败，状态码: %d, 响应: %s", r.StatusCode, string(r.Body))
}

// SetHeader 设置默认 Header。
func (c *Client) SetHeader(key, value string) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
}

// SetHeaders 批量设置默认 Header。
func (c *Client) SetHeaders(headers map[string]string) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	for key, value := range headers {
		c.headers[key] = value
	}
}
