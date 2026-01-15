package dependency

import (
	"context"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain/vo"
)

type QueryMetricBody struct {
	Method        string        `json:"method"`
	Instant       bool          `json:"instant"`
	Time          int64         `json:"time"`
	LookBackDelta string        `json:"look_back_delta"`
	Filters       []interface{} `json:"filters"`
}
type Datas struct {
	Labels map[string]string
	Times  []float64
	Values []float64
}
type MetricData struct {
	Datas      []Datas
	Step       string
	IsVariable bool
}

//go:generate mockgen -source ./uniquery_restapi.go -destination ../../mock/adapter/restapi/mock_uniquery_restapi.go -package mock
type UniQueryClient interface {
	GetDataView(ctx context.Context, viewId string, req vo.DataViewQueryV2, accout_id string) (vo.ViewUniResponseV2, error)
}
