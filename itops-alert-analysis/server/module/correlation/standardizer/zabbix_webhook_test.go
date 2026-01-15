package standardizer

import (
	"context"
	"errors"
	"testing"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/correlation/objectclass"
	. "github.com/smartystreets/goconvey/convey"
)

// mockObjectClassQuerier 是 ObjectClassQuerier 接口的 mock 实现
type mockObjectClassQuerier struct {
	result *objectclass.EntityObjectInfo
	err    error
}

func (m *mockObjectClassQuerier) GetEntityObjectInfo(ctx context.Context, entityObjectName string) (*objectclass.EntityObjectInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestNewZabbixWebhookStandardizer(t *testing.T) {
	Convey("TestNewZabbixWebhookStandardizer", t, func() {
		Convey("成功创建标准化器", func() {
			cfg := config.IngestConfig{}
			querier := &mockObjectClassQuerier{}

			standardizer := NewZabbixWebhookStandardizer(cfg, querier)

			So(standardizer, ShouldNotBeNil)
		})

		Convey("使用 nil querier 创建", func() {
			cfg := config.IngestConfig{}

			standardizer := NewZabbixWebhookStandardizer(cfg, nil)

			So(standardizer, ShouldNotBeNil)
		})
	})
}

func TestZabbixStandardizer_Standardize(t *testing.T) {
	Convey("TestZabbixStandardizer_Standardize", t, func() {
		ctx := context.Background()

		Convey("成功解析完整的 Zabbix Webhook 数据", func() {
			querier := &mockObjectClassQuerier{
				result: &objectclass.EntityObjectInfo{
					ObjectTypeID: "ot-host-001",
					ObjectID:     "obj-12345",
					Name:         "test-host",
				},
			}
			standardizer := NewZabbixWebhookStandardizer(config.IngestConfig{}, querier)

			payload := []byte(`{
				"timestamp": "1704067200",
				"description": "CPU使用率过高",
				"event_id": "100001",
				"recovery_id": "100002",
				"event_name": "CPU告警",
				"occur_time": "2025-01-01 12:00:00",
				"recovery_time": "2025-01-01 12:30:00",
				"event_severity": "High",
				"event_status": "发生",
				"entity_object_name": "test-host",
				"ip": "192.168.1.100",
				"item_key": "system.cpu.util",
				"item_name": "CPU utilization",
				"item_value": "95%"
			}`)

			rawEvent, err := standardizer.Standardize(ctx, payload)

			So(err, ShouldBeNil)
			So(rawEvent.EventProviderID, ShouldEqual, uint64(100001))
			So(rawEvent.RecoveryId, ShouldEqual, uint64(100002))
			So(rawEvent.EventTitle, ShouldEqual, "CPU告警")
			So(rawEvent.EventContent, ShouldEqual, "CPU使用率过高")
			So(rawEvent.EventType, ShouldEqual, "system.cpu.util")
			So(rawEvent.EventStatus, ShouldEqual, domain.EventStatusOccurred)
			So(rawEvent.EventLevel, ShouldEqual, domain.SeverityCritical)
			So(rawEvent.EventSource, ShouldEqual, domain.SourceZabbixWebhook)
			So(rawEvent.EntityObjectName, ShouldEqual, "test-host")
			So(rawEvent.EntityObjectClass, ShouldEqual, "ot-host-001")
			So(rawEvent.EntityObjectID, ShouldEqual, "obj-12345")
			So(rawEvent.EntityObjectIP, ShouldEqual, "192.168.1.100")
			So(rawEvent.EventOccurTime, ShouldNotBeNil)
			So(rawEvent.EventRecoveryTime, ShouldNotBeNil)
			So(rawEvent.EventID, ShouldNotEqual, 0)
		})

		Convey("JSON 解析失败返回错误", func() {
			querier := &mockObjectClassQuerier{
				result: &objectclass.EntityObjectInfo{},
			}
			standardizer := NewZabbixWebhookStandardizer(config.IngestConfig{}, querier)

			payload := []byte(`{invalid json}`)

			_, err := standardizer.Standardize(ctx, payload)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "解析ZabbixWebhook数据失败")
		})

		Convey("获取对象信息失败返回错误", func() {
			querier := &mockObjectClassQuerier{
				err: errors.New("object not found"),
			}
			standardizer := NewZabbixWebhookStandardizer(config.IngestConfig{}, querier)

			payload := []byte(`{
				"event_id": "100001",
				"entity_object_name": "unknown-host"
			}`)

			_, err := standardizer.Standardize(ctx, payload)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "获取对象信息失败")
		})

		Convey("时间解析失败不影响整体解析", func() {
			querier := &mockObjectClassQuerier{
				result: &objectclass.EntityObjectInfo{
					ObjectTypeID: "ot-host-001",
					ObjectID:     "obj-12345",
					Name:         "test-host",
				},
			}
			standardizer := NewZabbixWebhookStandardizer(config.IngestConfig{}, querier)

			payload := []byte(`{
				"event_id": "100001",
				"occur_time": "invalid-time-format",
				"recovery_time": "also-invalid",
				"entity_object_name": "test-host"
			}`)

			rawEvent, err := standardizer.Standardize(ctx, payload)

			So(err, ShouldBeNil)
			So(rawEvent.EventOccurTime, ShouldBeNil)
			So(rawEvent.EventRecoveryTime, ShouldBeNil)
		})

		Convey("没有时间字段时正常解析", func() {
			querier := &mockObjectClassQuerier{
				result: &objectclass.EntityObjectInfo{
					ObjectTypeID: "ot-host-001",
					ObjectID:     "obj-12345",
					Name:         "test-host",
				},
			}
			standardizer := NewZabbixWebhookStandardizer(config.IngestConfig{}, querier)

			payload := []byte(`{
				"event_id": "100001",
				"event_name": "测试告警",
				"entity_object_name": "test-host"
			}`)

			rawEvent, err := standardizer.Standardize(ctx, payload)

			So(err, ShouldBeNil)
			So(rawEvent.EventOccurTime, ShouldBeNil)
			So(rawEvent.EventRecoveryTime, ShouldBeNil)
			So(rawEvent.EventTitle, ShouldEqual, "测试告警")
		})

		Convey("恢复状态事件解析", func() {
			querier := &mockObjectClassQuerier{
				result: &objectclass.EntityObjectInfo{
					ObjectTypeID: "ot-host-001",
					ObjectID:     "obj-12345",
					Name:         "test-host",
				},
			}
			standardizer := NewZabbixWebhookStandardizer(config.IngestConfig{}, querier)

			payload := []byte(`{
				"event_id": "100001",
				"event_status": "恢复",
				"entity_object_name": "test-host"
			}`)

			rawEvent, err := standardizer.Standardize(ctx, payload)

			So(err, ShouldBeNil)
			So(rawEvent.EventStatus, ShouldEqual, domain.EventStatusRecovered)
		})

		Convey("各种严重程度映射", func() {
			querier := &mockObjectClassQuerier{
				result: &objectclass.EntityObjectInfo{
					ObjectTypeID: "ot-host-001",
					ObjectID:     "obj-12345",
					Name:         "test-host",
				},
			}
			standardizer := NewZabbixWebhookStandardizer(config.IngestConfig{}, querier)

			testCases := []struct {
				severity string
				expected domain.Severity
			}{
				{"Disaster", domain.SeverityEmergency},
				{"High", domain.SeverityCritical},
				{"Average", domain.SeverityMajor},
				{"Warning", domain.SeverityWarning},
				{"Unknown", domain.SeverityNormal},
			}

			for _, tc := range testCases {
				payload := []byte(`{
					"event_id": "100001",
					"event_severity": "` + tc.severity + `",
					"entity_object_name": "test-host"
				}`)

				rawEvent, err := standardizer.Standardize(ctx, payload)

				So(err, ShouldBeNil)
				So(rawEvent.EventLevel, ShouldEqual, tc.expected)
			}
		})
	})
}

func Test_mapSeverity(t *testing.T) {
	type args struct {
		zabbixSeverity string
	}
	tests := []struct {
		name string
		args args
		want domain.Severity
	}{
		{
			name: "Disaster",
			args: args{
				zabbixSeverity: "Disaster",
			},
			want: domain.SeverityEmergency,
		},
		{
			name: "High",
			args: args{
				zabbixSeverity: "High",
			},
			want: domain.SeverityCritical,
		},
		{
			name: "Average",
			args: args{
				zabbixSeverity: "Average",
			},
			want: domain.SeverityMajor,
		},
		{
			name: "Warning",
			args: args{
				zabbixSeverity: "Warning",
			},
			want: domain.SeverityWarning,
		},
		{
			name: "Other",
			args: args{
				zabbixSeverity: "Other",
			},
			want: domain.SeverityNormal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapSeverity(tt.args.zabbixSeverity); got != tt.want {
				t.Errorf("mapSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventStatus(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want domain.EventStatus
	}{
		{
			name: "发生",
			args: args{
				s: "发生",
			},
			want: domain.EventStatusOccurred,
		},
		{
			name: "恢复",
			args: args{
				s: "恢复",
			},
			want: domain.EventStatusRecovered,
		}, {
			name: "其他",
			args: args{
				s: "其他",
			},
			want: domain.EventStatusOccurred,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := eventStatus(tt.args.s); got != tt.want {
				t.Errorf("eventStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
