package opensearch

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	opensearchsdk "github.com/opensearch-project/opensearch-go/v2"
	"github.com/pkg/errors"
)

const defaultTimeout = 10 * time.Second

type OpenSearchConfig struct {
	Hosts              []string      `yaml:"hosts"`
	Username           string        `yaml:"username"`
	Password           string        `yaml:"password"`
	Timeout            time.Duration `yaml:"timeout"`
	InsecureSkipVerify bool          `yaml:"insecure_skip_verify"`
}

// NewClient 基于配置初始化官方 OpenSearch SDK 客户端。
func NewClient(cfg OpenSearchConfig) (*opensearchsdk.Client, error) {
	if len(cfg.Hosts) == 0 {
		return nil, errors.New("opensearch hosts 不能为空")
	}

	addresses := make([]string, 0, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = "http://" + host
		}
		addresses = append(addresses, strings.TrimRight(host, "/"))
	}
	if len(addresses) == 0 {
		return nil, errors.New("opensearch hosts 经处理后为空")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	dialer := &net.Dialer{
		Timeout: timeout,
	}
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: timeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
	}

	client, err := opensearchsdk.NewClient(opensearchsdk.Config{
		Addresses: addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: transport,
	})
	if err != nil {
		return nil, errors.Wrap(err, "初始化 OpenSearch SDK 失败")
	}
	return client, nil
}
