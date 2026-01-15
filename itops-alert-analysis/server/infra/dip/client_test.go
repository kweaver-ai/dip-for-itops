package dip

import (
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewClient(t *testing.T) {
	Convey("TestNewClient", t, func() {
		Convey("带 Authorization 创建客户端", func() {
			cfg := config.DIPConfig{
				Host:               "https://example.com",
				KnID:               "test-kn-id",
				Authorization:      "Bearer test-token",
				InsecureSkipVerify: true,
				Timeout:            30 * time.Second,
			}

			client := NewClient(cfg, mockGetAuth, mockGetKnID)

			So(client, ShouldNotBeNil)
			So(client.knID, ShouldEqual, "test-kn-id")
			So(client.httpClient, ShouldNotBeNil)
		})

		Convey("不带 Authorization 创建客户端", func() {
			cfg := config.DIPConfig{
				Host:               "https://example.com",
				KnID:               "test-kn-id",
				Authorization:      "",
				InsecureSkipVerify: false,
				Timeout:            10 * time.Second,
			}

			client := NewClient(cfg, mockGetAuth, mockGetKnID)

			So(client, ShouldNotBeNil)
			So(client.knID, ShouldEqual, "test-kn-id")
			So(client.httpClient, ShouldNotBeNil)
		})

		Convey("空配置创建客户端", func() {
			cfg := config.DIPConfig{}

			client := NewClient(cfg, mockGetAuth, mockGetKnID)

			So(client, ShouldNotBeNil)
			So(client.knID, ShouldEqual, "")
			So(client.httpClient, ShouldNotBeNil)
		})

		Convey("不传 getKnID 函数创建客户端", func() {
			cfg := config.DIPConfig{
				Host: "https://example.com",
				KnID: "default-kn-id",
			}

			client := NewClient(cfg, mockGetAuth, nil)

			So(client, ShouldNotBeNil)
			So(client.knID, ShouldEqual, "default-kn-id")
			So(client.getKnID, ShouldBeNil)
		})
	})
}

func TestClient_KnID(t *testing.T) {
	Convey("TestClient_KnID", t, func() {
		Convey("使用动态 getKnID 函数获取 KnowledgeID", func() {
			dynamicKnID := "dynamic-kn-id"
			getKnID := func() string {
				return dynamicKnID
			}

			cfg := config.DIPConfig{
				Host: "https://example.com",
				KnID: "default-kn-id",
			}
			client := NewClient(cfg, mockGetAuth, getKnID)

			So(client.KnID(), ShouldEqual, "dynamic-kn-id")

			// 模拟动态更新
			dynamicKnID = "updated-kn-id"
			So(client.KnID(), ShouldEqual, "updated-kn-id")
		})

		Convey("getKnID 为 nil 时使用默认 knID", func() {
			cfg := config.DIPConfig{
				Host: "https://example.com",
				KnID: "default-kn-id",
			}
			client := NewClient(cfg, mockGetAuth, nil)

			So(client.KnID(), ShouldEqual, "default-kn-id")
		})

		Convey("getKnID 和默认 knID 都为空", func() {
			cfg := config.DIPConfig{
				Host: "https://example.com",
			}
			client := NewClient(cfg, mockGetAuth, nil)

			So(client.KnID(), ShouldEqual, "")
		})
	})
}
