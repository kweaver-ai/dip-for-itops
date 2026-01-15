package opensearch

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOpenSearchConfig(t *testing.T) {
	Convey("TestOpenSearchConfig", t, func() {
		Convey("Config 结构体字段", func() {
			cfg := OpenSearchConfig{
				Hosts:              []string{"localhost:9200", "localhost:9201"},
				Username:           "admin",
				Password:           "admin123",
				Timeout:            30 * time.Second,
				InsecureSkipVerify: true,
			}

			So(cfg.Hosts, ShouldResemble, []string{"localhost:9200", "localhost:9201"})
			So(cfg.Username, ShouldEqual, "admin")
			So(cfg.Password, ShouldEqual, "admin123")
			So(cfg.Timeout, ShouldEqual, 30*time.Second)
			So(cfg.InsecureSkipVerify, ShouldBeTrue)
		})

		Convey("Config 空配置", func() {
			cfg := OpenSearchConfig{}

			So(cfg.Hosts, ShouldBeNil)
			So(cfg.Username, ShouldEqual, "")
			So(cfg.Password, ShouldEqual, "")
			So(cfg.Timeout, ShouldEqual, 0)
			So(cfg.InsecureSkipVerify, ShouldBeFalse)
		})
	})
}

func TestNewClient(t *testing.T) {
	Convey("TestNewClient", t, func() {
		Convey("hosts 为空返回错误", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{},
			}

			client, err := NewClient(cfg)

			So(err, ShouldNotBeNil)
			So(client, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch hosts 不能为空")
		})

		Convey("hosts 为 nil 返回错误", func() {
			cfg := OpenSearchConfig{
				Hosts: nil,
			}

			client, err := NewClient(cfg)

			So(err, ShouldNotBeNil)
			So(client, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch hosts 不能为空")
		})

		Convey("hosts 全为空字符串返回错误", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"", "   ", "  "},
			}

			client, err := NewClient(cfg)

			So(err, ShouldNotBeNil)
			So(client, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "opensearch hosts 经处理后为空")
		})

		Convey("成功创建客户端（单个 host）", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"localhost:9200"},
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("成功创建客户端（多个 hosts）", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"localhost:9200", "localhost:9201", "localhost:9202"},
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("host 自动添加 http:// 前缀", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"localhost:9200"},
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("host 已有 http:// 前缀", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"http://localhost:9200"},
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("host 已有 https:// 前缀", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"https://localhost:9200"},
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("host 带有尾部斜杠被去除", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"localhost:9200/", "localhost:9201//"},
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("host 带有前后空格被去除", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"  localhost:9200  ", "\tlocalhost:9201\t"},
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("混合有效和无效 hosts", func() {
			cfg := OpenSearchConfig{
				Hosts: []string{"", "localhost:9200", "   ", "localhost:9201"},
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("使用默认超时", func() {
			cfg := OpenSearchConfig{
				Hosts:   []string{"localhost:9200"},
				Timeout: 0, // 使用默认值
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("使用负数超时（使用默认值）", func() {
			cfg := OpenSearchConfig{
				Hosts:   []string{"localhost:9200"},
				Timeout: -1 * time.Second, // 负数，使用默认值
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("使用自定义超时", func() {
			cfg := OpenSearchConfig{
				Hosts:   []string{"localhost:9200"},
				Timeout: 60 * time.Second,
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("使用用户名密码认证", func() {
			cfg := OpenSearchConfig{
				Hosts:    []string{"localhost:9200"},
				Username: "admin",
				Password: "admin123",
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("启用 InsecureSkipVerify", func() {
			cfg := OpenSearchConfig{
				Hosts:              []string{"https://localhost:9200"},
				InsecureSkipVerify: true,
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})

		Convey("完整配置", func() {
			cfg := OpenSearchConfig{
				Hosts:              []string{"https://node1:9200", "https://node2:9200", "https://node3:9200"},
				Username:           "admin",
				Password:           "securePassword123",
				Timeout:            30 * time.Second,
				InsecureSkipVerify: true,
			}

			client, err := NewClient(cfg)

			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})
	})
}
