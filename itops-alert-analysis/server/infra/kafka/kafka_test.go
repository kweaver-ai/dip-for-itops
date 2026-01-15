package kafka

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildSASLMechanism(t *testing.T) {
	Convey("TestBuildSASLMechanism", t, func() {
		Convey("SASL 配置为 nil 返回 nil", func() {
			mechanism, err := buildSASLMechanism(nil)

			So(err, ShouldBeNil)
			So(mechanism, ShouldBeNil)
		})

		Convey("SASL 未启用返回 nil", func() {
			cfg := &SASLConfig{
				Enabled:   false,
				Mechanism: "PLAIN",
				Username:  "user",
				Password:  "pass",
			}

			mechanism, err := buildSASLMechanism(cfg)

			So(err, ShouldBeNil)
			So(mechanism, ShouldBeNil)
		})

		Convey("使用 PLAIN 机制", func() {
			cfg := &SASLConfig{
				Enabled:   true,
				Mechanism: "PLAIN",
				Username:  "user",
				Password:  "pass",
			}

			mechanism, err := buildSASLMechanism(cfg)

			So(err, ShouldBeNil)
			So(mechanism, ShouldNotBeNil)
		})

		Convey("使用 plain 小写机制", func() {
			cfg := &SASLConfig{
				Enabled:   true,
				Mechanism: "plain",
				Username:  "user",
				Password:  "pass",
			}

			mechanism, err := buildSASLMechanism(cfg)

			So(err, ShouldBeNil)
			So(mechanism, ShouldNotBeNil)
		})

		Convey("机制为空默认使用 PLAIN", func() {
			cfg := &SASLConfig{
				Enabled:   true,
				Mechanism: "",
				Username:  "user",
				Password:  "pass",
			}

			mechanism, err := buildSASLMechanism(cfg)

			So(err, ShouldBeNil)
			So(mechanism, ShouldNotBeNil)
		})

		Convey("使用 SCRAM-SHA-256 机制", func() {
			cfg := &SASLConfig{
				Enabled:   true,
				Mechanism: "SCRAM-SHA-256",
				Username:  "user",
				Password:  "pass",
			}

			mechanism, err := buildSASLMechanism(cfg)

			So(err, ShouldBeNil)
			So(mechanism, ShouldNotBeNil)
		})

		Convey("使用 SCRAM-SHA-512 机制", func() {
			cfg := &SASLConfig{
				Enabled:   true,
				Mechanism: "SCRAM-SHA-512",
				Username:  "user",
				Password:  "pass",
			}

			mechanism, err := buildSASLMechanism(cfg)

			So(err, ShouldBeNil)
			So(mechanism, ShouldNotBeNil)
		})

		Convey("不支持的 SASL 机制返回错误", func() {
			cfg := &SASLConfig{
				Enabled:   true,
				Mechanism: "UNSUPPORTED",
				Username:  "user",
				Password:  "pass",
			}

			mechanism, err := buildSASLMechanism(cfg)

			So(err, ShouldNotBeNil)
			So(mechanism, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "不支持的 SASL 机制")
			So(err.Error(), ShouldContainSubstring, "UNSUPPORTED")
		})
	})
}

func TestConfig(t *testing.T) {
	Convey("TestConfig", t, func() {
		Convey("Config 结构体字段", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092", "localhost:9093"},
				SASL: &SASLConfig{
					Enabled:   true,
					Mechanism: "PLAIN",
					Username:  "user",
					Password:  "pass",
				},
				Topic:   "test-topic",
				GroupID: "test-group",
			}

			So(cfg.Brokers, ShouldResemble, []string{"localhost:9092", "localhost:9093"})
			So(cfg.SASL, ShouldNotBeNil)
			So(cfg.SASL.Enabled, ShouldBeTrue)
			So(cfg.SASL.Mechanism, ShouldEqual, "PLAIN")
			So(cfg.SASL.Username, ShouldEqual, "user")
			So(cfg.SASL.Password, ShouldEqual, "pass")
			So(cfg.Topic, ShouldEqual, "test-topic")
			So(cfg.GroupID, ShouldEqual, "test-group")
		})

		Convey("Config 无 SASL 配置", func() {
			cfg := Config{
				Brokers: []string{"localhost:9092"},
				SASL:    nil,
				Topic:   "test-topic",
			}

			So(cfg.Brokers, ShouldResemble, []string{"localhost:9092"})
			So(cfg.SASL, ShouldBeNil)
		})
	})
}
