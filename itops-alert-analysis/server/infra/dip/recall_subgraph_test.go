package dip

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestClient_QueryTopologyObjectSubgraph(t *testing.T) {
	Convey("TestClient_QueryTopologyObjectSubgraph", t, func() {
		Convey("成功查询对象子图", func() {
			var capturedPath string
			var capturedReqBody domain.SubGraphQueryRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				json.NewDecoder(r.Body).Decode(&capturedReqBody)

				resp := domain.SubGraphQueryResponse{
					Objects: map[string]domain.SubGraphObject{
						"obj-1": {
							ID:           "obj-1",
							ObjectTypeID: "Server",
							Display:      "server-01",
						},
					},
					RelationPaths: []domain.SubGraphRelationPath{},
					SearchAfter:   []interface{}{},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryTopologyObjectSubgraph(
				context.Background(),
				"Server",
				[]string{"entity-1", "entity-2"},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 1)
			So(capturedPath, ShouldEqual, "/api/ontology-query/v1/knowledge-networks/test-kn/subgraph")
			So(capturedReqBody.SourceObjectTypeID, ShouldEqual, "Server")
			So(capturedReqBody.Direction, ShouldEqual, "forward")
			So(capturedReqBody.PathLength, ShouldEqual, 1)
		})

		Convey("entityObjectIDs 为空返回空结果", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryTopologyObjectSubgraph(
				context.Background(),
				"Server",
				[]string{},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 0)
		})

		Convey("entityObjectIDs 全部为空字符串返回空结果", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryTopologyObjectSubgraph(
				context.Background(),
				"Server",
				[]string{"", ""},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 0)
		})

		Convey("entityObjectClass 为空返回错误", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryTopologyObjectSubgraph(
				context.Background(),
				"",
				[]string{"entity-1"},
				"Bearer test-token",
			)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "实体对象类型不能为空")
		})

		Convey("client 为 nil 返回错误", func() {
			var client *Client
			result, err := client.QueryTopologyObjectSubgraph(
				context.Background(),
				"Server",
				[]string{"entity-1"},
				"Bearer test-token",
			)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "client 未初始化")
		})

		Convey("ctx 为 nil 返回错误", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryTopologyObjectSubgraph(
				nil,
				"Server",
				[]string{"entity-1"},
				"Bearer test-token",
			)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "上下文不能为 nil")
		})
	})
}

func TestClient_QueryTopologyNeighbors(t *testing.T) {
	Convey("TestClient_QueryTopologyNeighbors", t, func() {
		Convey("成功查询一度拓扑邻居", func() {
			var capturedReqBody domain.SubGraphQueryRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&capturedReqBody)

				resp := domain.SubGraphQueryResponse{
					Objects: map[string]domain.SubGraphObject{
						"obj-1": {ID: "obj-1", ObjectTypeID: "Server"},
						"obj-2": {ID: "obj-2", ObjectTypeID: "Application"},
					},
					RelationPaths: []domain.SubGraphRelationPath{},
					SearchAfter:   []interface{}{},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryTopologyNeighbors(
				context.Background(),
				"Server",
				[]string{"entity-1"},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 2)
			So(capturedReqBody.Direction, ShouldEqual, "bidirectional")
			So(capturedReqBody.PathLength, ShouldEqual, 1)
		})

		Convey("entityObjectIDs 为空返回空结果", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryTopologyNeighbors(
				context.Background(),
				"Server",
				[]string{},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 0)
		})

		Convey("entityObjectClass 为空返回错误", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryTopologyNeighbors(
				context.Background(),
				"",
				[]string{"entity-1"},
				"Bearer test-token",
			)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestClient_QueryHistoricalCausality(t *testing.T) {
	Convey("TestClient_QueryHistoricalCausality", t, func() {
		Convey("成功查询历史因果关系", func() {
			var capturedReqBody domain.SubGraphQueryRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&capturedReqBody)

				resp := domain.SubGraphQueryResponse{
					Objects: map[string]domain.SubGraphObject{
						"obj-1": {ID: "obj-1", ObjectTypeID: "Fault"},
						"obj-2": {ID: "obj-2", ObjectTypeID: "FaultCausal"},
					},
					RelationPaths: []domain.SubGraphRelationPath{},
					SearchAfter:   []interface{}{},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryHistoricalCausality(
				context.Background(),
				"Fault",
				[]uint64{1, 2, 3},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 2)
			So(capturedReqBody.Direction, ShouldEqual, "forward")
			So(capturedReqBody.PathLength, ShouldEqual, 2)
		})

		Convey("entityIDs 为空返回空结果", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryHistoricalCausality(
				context.Background(),
				"Fault",
				[]uint64{},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 0)
		})

		Convey("entityIDs 全部为 0 返回空结果", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryHistoricalCausality(
				context.Background(),
				"Fault",
				[]uint64{0, 0},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Objects), ShouldEqual, 0)
		})

		Convey("entityClassID 为空返回错误", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryHistoricalCausality(
				context.Background(),
				"",
				[]uint64{1, 2},
				"Bearer test-token",
			)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestClient_QueryObjectInfo(t *testing.T) {
	Convey("TestClient_QueryObjectInfo", t, func() {
		Convey("成功查询对象信息", func() {
			var capturedPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path

				resp := domain.ObjectInfoQueryResponse{
					Datas:           []interface{}{map[string]interface{}{"s_id": "entity-1", "name": "server-01"}},
					SearchAfter:     []interface{}{},
					SearchFromIndex: false,
					OverallMS:       10,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			result, err := client.QueryObjectInfo(
				context.Background(),
				"Server",
				[]string{"entity-1"},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(capturedPath, ShouldEqual, "/api/ontology-query/v1/knowledge-networks/test-kn/object-types/Server")
		})

		Convey("entityObjectIDs 为空返回空结果", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryObjectInfo(
				context.Background(),
				"Server",
				[]string{},
				"Bearer test-token",
			)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("entityObjectClass 为空返回错误", func() {
			client := newTestClient("http://example.com")
			result, err := client.QueryObjectInfo(
				context.Background(),
				"",
				[]string{"entity-1"},
				"Bearer test-token",
			)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func TestClient_buildSubConditionsBySID(t *testing.T) {
	Convey("TestClient_buildSubConditionsBySID", t, func() {
		client := newTestClient("http://example.com")

		Convey("正常构建子条件", func() {
			conditions := client.buildSubConditionsBySID([]string{"id-1", "id-2", "id-3"})

			So(len(conditions), ShouldEqual, 3)
			So(conditions[0].Field, ShouldEqual, "s_id")
			So(conditions[0].Operation, ShouldEqual, "==")
			So(conditions[0].Value, ShouldEqual, "id-1")
		})

		Convey("过滤空字符串", func() {
			conditions := client.buildSubConditionsBySID([]string{"id-1", "", "id-2", ""})

			So(len(conditions), ShouldEqual, 2)
			So(conditions[0].Value, ShouldEqual, "id-1")
			So(conditions[1].Value, ShouldEqual, "id-2")
		})

		Convey("空输入返回空切片", func() {
			conditions := client.buildSubConditionsBySID([]string{})

			So(len(conditions), ShouldEqual, 0)
		})
	})
}

func TestClient_buildSubConditionsByFaultID(t *testing.T) {
	Convey("TestClient_buildSubConditionsByFaultID", t, func() {
		client := newTestClient("http://example.com")

		Convey("正常构建子条件", func() {
			conditions := client.buildSubConditionsByFaultID([]uint64{1, 2, 3})

			So(len(conditions), ShouldEqual, 3)
			So(conditions[0].Field, ShouldEqual, "fault_id")
			So(conditions[0].Operation, ShouldEqual, "==")
			So(conditions[0].Value, ShouldEqual, "1")
		})

		Convey("过滤 0 值", func() {
			conditions := client.buildSubConditionsByFaultID([]uint64{1, 0, 2, 0})

			So(len(conditions), ShouldEqual, 2)
			So(conditions[0].Value, ShouldEqual, "1")
			So(conditions[1].Value, ShouldEqual, "2")
		})

		Convey("空输入返回空切片", func() {
			conditions := client.buildSubConditionsByFaultID([]uint64{})

			So(len(conditions), ShouldEqual, 0)
		})
	})
}
