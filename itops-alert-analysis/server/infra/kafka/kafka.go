package kafka

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const (
	defaultGroupID = "itops-alert-analysis-consumer" //默认消费者
	minBytes       = 1                               //
	maxBytes       = 10 * 1024 * 1024
)

type Config struct {
	Brokers []string    `yaml:"brokers"`
	SASL    *SASLConfig `yaml:"sasl,omitempty"`

	// 内部使用字段（用于创建 Kafka 客户端）
	Topic   string `yaml:"-"` // 由代码根据 RawEvents/ProblemEvents 填充
	GroupID string `yaml:"-"` // 由代码根据 RawEvents/ProblemEvents 填充
}

type SASLConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Mechanism string `yaml:"mechanism"` // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}

// buildSASLMechanism 根据配置构建 SASL 认证机制。
func buildSASLMechanism(saslCfg *SASLConfig) (sasl.Mechanism, error) {
	if saslCfg == nil || !saslCfg.Enabled {
		log.Infof("SASL 认证未启用")
		return nil, nil
	}

	log.Debugf("构建 SASL 认证")

	switch saslCfg.Mechanism {
	case "PLAIN", "plain", "":
		mechanism := plain.Mechanism{
			Username: saslCfg.Username,
			Password: saslCfg.Password,
		}
		log.Infof("使用 PLAIN 认证机制")
		return mechanism, nil
	case "SCRAM-SHA-256":
		mechanism, err := scram.Mechanism(scram.SHA256, saslCfg.Username, saslCfg.Password)
		if err != nil {
			return nil, errors.Wrap(err, "创建 SCRAM-SHA-256 认证失败")
		}
		log.Infof("使用 SCRAM-SHA-256 认证机制")
		return mechanism, nil
	case "SCRAM-SHA-512":
		mechanism, err := scram.Mechanism(scram.SHA512, saslCfg.Username, saslCfg.Password)
		if err != nil {
			return nil, errors.Wrap(err, "创建 SCRAM-SHA-512 认证失败")
		}
		log.Infof("使用 SCRAM-SHA-512 认证机制")
		return mechanism, nil
	default:
		return nil, errors.Errorf("不支持的 SASL 机制: %s", saslCfg.Mechanism)
	}
}
