import { groupBy } from 'lodash';
import fitTextToWidth from './fitTextToWidth';

interface DataType {
  nodes: any[];
  edges?: any[];
  combos?: any[];
}

/**
 * 将原始网络拓扑数据转换为 G6 图渲染所需的数据格式，并分离故障节点
 * @param networkData 原始网络拓扑数据，包含 nodes 和 edges 数组
 * @param isShowFailurePoint 是否将故障节点也加入最终节点列表，默认 false
 * @returns 返回一个对象，包含：
 *  - data: G6 可用的图数据（nodes + edges）
 *  - failurePointNodes: 故障节点数组
 *  - failurePointObj: 以 entity_object_id 分组的故障节点映射表
 */
const originDataToG6Data = ({
  networkData = {},
  backtrace = [],
  rootCauseObjectId,
  rootCauseFaultId,
  isShowFailurePoint = false
}: {
  networkData: any;
  backtrace: any;
  rootCauseObjectId: string;
  rootCauseFaultId: string;
  isShowFailurePoint: boolean;
}) => {
  const { nodes = [], edges = [] } = networkData;
  const normalNodes: any[] = [];
  const failurePointNodes: any[] = backtrace;
  const normalEdges: any[] = [];

  backtrace.forEach((item: any) => {
    item.id = item.fault_id.toString();
    item.fault_id = item.fault_id?.toString();
    item.label = fitTextToWidth(item.fault_name, 130);
    item.allName = item.fault_name;
    item.type = 'fault-point-node';
    item.comboId = item.entity_object_id;
    item.isFailurePoint = true;
    item.rootCause = item.id === rootCauseFaultId;
  });

  nodes.forEach((item: any) => {
    normalNodes.push({
      ...item,
      id: item.s_id,
      label: fitTextToWidth(item.name, 130),
      allName: item.name,
      rootCause: item.s_id === rootCauseObjectId,
      relation_fault_point_ids: item.relation_fault_point_ids.map(
        (id: string) => id?.toString()
      )
    });
  });

  edges.forEach((item: any) => {
    normalEdges.push({
      ...item,
      source: item.source_object_id,
      target: item.target_object_id
    });
  });

  const failurePointsObj = groupBy(failurePointNodes, 'entity_object_id');
  let data: DataType = {
    nodes: [...normalNodes],
    edges: normalEdges
  };

  // 若需要展示故障节点，则显示故障节点，同时添加组合节点
  if (isShowFailurePoint) {
    data = {
      nodes: [...failurePointNodes],
      combos: normalNodes,
      edges: normalEdges
    };
  }

  console.log('data', data);
  console.log('failurePointNodes', failurePointNodes);

  return { data, failurePointNodes, failurePointsObj };
};

export default originDataToG6Data;
