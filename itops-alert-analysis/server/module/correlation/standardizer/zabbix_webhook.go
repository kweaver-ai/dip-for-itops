package standardizer

import (
	"context"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/module/correlation/objectclass"
	"encoding/json"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/idgen"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils/timex"
	"github.com/spf13/cast"

	"github.com/pkg/errors"
)

// ObjectClassQuerier 对象类查询接口。
type ObjectClassQuerier interface {
	GetEntityObjectInfo(ctx context.Context, entityObjectName string) (*objectclass.EntityObjectInfo, error)
}

// zabbixStandardizer 占位实现：仅声明接口，不做实际字段映射。
type zabbixStandardizer struct {
	cfg                config.IngestConfig
	genID              *idgen.Generator
	objectClassQuerier ObjectClassQuerier
}
type ZabbixWebhook struct {
	Timestamp        string `json:"timestamp"`
	Description      string `json:"description"`
	EventId          string `json:"event_id"`
	RecoveryId       string `json:"recovery_id"`
	EventName        string `json:"event_name"`
	OccurTime        string `json:"occur_time"`
	RecoveryTime     string `json:"recovery_time"`
	EventSeverity    string `json:"event_severity"`
	EventStatus      string `json:"event_status"`
	EntityObjectName string `json:"entity_object_name"`
	Ip               string `json:"ip"`
	Itemkey          string `json:"item_key"`
	ItemName         string `json:"item_name"`
	ItemValue        string `json:"item_value"`
}

// NewZabbixWebhookStandardizer 基于 ingest 配置创建 zabbix webhook 标准化器。
func NewZabbixWebhookStandardizer(cfg config.IngestConfig, querier ObjectClassQuerier) Standardizer {
	return &zabbixStandardizer{
		cfg:                cfg,
		genID:              idgen.New(),
		objectClassQuerier: querier,
	}
}

func (s *zabbixStandardizer) Standardize(ctx context.Context, payload []byte) (domain.RawEvent, error) {
	var zabbixWebhook ZabbixWebhook
	var occurTime, recoveryTime time.Time
	var err error
	if err := json.Unmarshal(payload, &zabbixWebhook); err != nil {
		return domain.RawEvent{}, errors.Wrap(err, "解析ZabbixWebhook数据失败")
	}

	if len(zabbixWebhook.OccurTime) > 0 {
		occurTime, err = timex.ParseTime(zabbixWebhook.OccurTime, time.DateTime)
		if err != nil {
			log.Warnf("解析事件发生时间失败: event_id=%s, occur_time=%s, err=%v",
				zabbixWebhook.EventId, zabbixWebhook.OccurTime, err)
		}
	}

	if len(zabbixWebhook.RecoveryTime) > 0 {
		recoveryTime, err = timex.ParseTime(zabbixWebhook.RecoveryTime, time.DateTime)
		if err != nil {
			log.Warnf("解析事件恢复时间失败: event_id=%s, recovery_time=%s, err=%v",
				zabbixWebhook.EventId, zabbixWebhook.RecoveryTime, err)
		}
	}

	// 从缓存获取实体对象信息（包含 object_type_id、object_id、name）
	objInfo, err := s.objectClassQuerier.GetEntityObjectInfo(ctx, zabbixWebhook.EntityObjectName)
	if err != nil {
		return domain.RawEvent{}, errors.Wrap(err, "获取对象信息失败")
	}

	var rawEvent = domain.RawEvent{
		EventID:           s.genID.NextID(),
		RecoveryId:        cast.ToUint64(zabbixWebhook.RecoveryId),
		EventProviderID:   cast.ToUint64(zabbixWebhook.EventId),
		EventTimestamp:    timex.NowLocalTime(),
		EventTitle:        zabbixWebhook.EventName,
		EventContent:      zabbixWebhook.Description,
		EventType:         zabbixWebhook.Itemkey,
		EventStatus:       eventStatus(zabbixWebhook.EventStatus),
		EventLevel:        mapSeverity(zabbixWebhook.EventSeverity),
		EventSource:       domain.SourceZabbixWebhook,
		EntityObjectName:  objInfo.Name,         // 使用缓存中的 name
		EntityObjectClass: objInfo.ObjectTypeID, // 使用缓存中的 object_type_id
		EntityObjectID:    objInfo.ObjectID,     // 使用缓存中的 object_id
		EntityObjectIP:    zabbixWebhook.Ip,
		RawEventMsg:       string(payload),
	}

	if !occurTime.IsZero() {
		rawEvent.EventOccurTime = &occurTime
	}
	if !recoveryTime.IsZero() {
		rawEvent.EventRecoveryTime = &recoveryTime
	}

	return rawEvent, nil
}

func eventStatus(s string) domain.EventStatus {
	eventStatusMaping := map[string]domain.EventStatus{
		"发生": domain.EventStatusOccurred,
		"恢复": domain.EventStatusRecovered,
	}
	if severity, ok := eventStatusMaping[s]; ok {
		return severity
	}
	return domain.EventStatusOccurred
}

func mapSeverity(zabbixSeverity string) domain.Severity {
	severityMaping := map[string]domain.Severity{
		"Disaster": domain.SeverityEmergency,
		"High":     domain.SeverityCritical,
		"Average":  domain.SeverityMajor,
		"Warning":  domain.SeverityWarning,
	}
	if severity, ok := severityMaping[zabbixSeverity]; ok {
		return severity
	}
	return domain.SeverityNormal
}
