/* eslint-disable max-lines */
/* eslint-disable react-hooks/exhaustive-deps */
import React, {
  useRef,
  useState,
  useEffect,
  forwardRef,
  useImperativeHandle,
  useContext,
  useMemo
} from 'react';
import { Flex, message } from 'antd';
import { useBoolean, useSize, useUpdateEffect } from 'ahooks';
import { useParams } from '@noya/max';
import { omit } from 'lodash';
import type { IGraph } from '@antv/g6';
import G6 from '@antv/g6';
import intl from 'react-intl-universal';
import ARIconfont from 'Components/ARIconfont';
import { Impact_Level_Maps } from 'Constants/common';
import { setRootCause, getSubgraph } from 'Services/fault-analysis';
import { AIChatContext } from 'Components/AIChatButton/context';
import { faultAnalysisContext } from '../../index';
import ToolHeader from './ToolHeader';
import styles from './index.module.less';
import SimpleLayout from './SimpleLayout';
import g6DefaultOptions from './constants/g6Oprtions';
import DetialDrawer from './DetialDrawer';
import { customNodeDraw, customNodeAfterDraw } from './utils/overviewNode';
import { faultPointDraw } from './utils/faultPointNode';
import { iconList } from './utils/index';
import originDataToG6Data from './utils/getG6Data';
import delAssociateData from './utils/delAssociateData';
import { customEdgeDraw } from './utils/customEdgs';
import { comboForceDefaultConfig } from './SimpleLayout/utils';

G6.registerNode(
  'overview-node',
  {
    draw: customNodeDraw,
    afterDraw: customNodeAfterDraw,
    update: undefined
  },
  'single-node'
);
G6.registerNode(
  'fault-point-node',
  {
    draw: faultPointDraw,
    update: undefined
  },
  'single-node'
);

G6.registerEdge(
  'custom-edge',
  {
    draw: customEdgeDraw,
    update: undefined
  },
  'line'
);

const TopologyGraph = forwardRef((props: { isPlaying?: boolean }, ref) => {
  const { isPlaying = false } = props;
  const [graphInstance, setGraphInstance] = useState<IGraph | null>(null);
  const [
    drawerVisible,
    { setTrue: openDetailDrawer, setFalse: closeDetailDrawer }
  ] = useBoolean(false);
  const [isShowLegend, setIsShowLegend] = useState(true);
  const [isShowFailurePoint, setIsShowFailurePoint] = useState(false);
  const [detailActiveKey, setDetailActiveKey] = useState<string>('1');
  const [detailNodeData, setDetailNodeData] = useState<any>({});
  const [containerOriginSize, setContainerOriginSize] = useState<
    | {
        width: number;
        height: number;
      }
    | undefined
  >();
  const fullDomRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const containerSize = useSize(containerRef);
  const graphInstanceRef = useRef<any>(null);
  const knIdRef = useRef<any>(null);
  const failurePointsObjRef = useRef<any>({});
  const layoutRef = useRef<any>(comboForceDefaultConfig);
  const urlParams = useParams<{ id: string }>();

  const { send } = useContext(AIChatContext);
  const {
    problemData: {
      rca_results: {
        adp_kn_id: knId,
        rca_context: { backtrace = [], network: networkData = {} }
      },
      root_cause_object_id: rootCauseObjectId,
      root_cause_fault_id: rootCauseFaultId
    },
    onSetRootCaseNode
  } = useContext(faultAnalysisContext);

  knIdRef.current = knId;

  const { data, failurePointNodes, failurePointsObj } = useMemo(
    () =>
      originDataToG6Data({
        networkData,
        backtrace,
        rootCauseObjectId,
        rootCauseFaultId: rootCauseFaultId?.toString(),
        isShowFailurePoint
      }),
    [networkData, rootCauseObjectId, rootCauseFaultId, isShowFailurePoint]
  );

  failurePointsObjRef.current = failurePointsObj;

  const onLayoutChange = (value: any) => {
    layoutRef.current = value;
  };

  // 更新图数据,重新设置布局
  const updateGraphData = (newData: any) => {
    graphInstanceRef.current.updateLayout(layoutRef.current);
    graphInstanceRef.current.changeData(newData);
    graphInstanceRef.current.render();
    graphInstanceRef.current.fitView(20);
  };

  // 获取关联的子节点
  const getSubNodes = async (selNode: any) => {
    const res = await getSubgraph(
      knIdRef.current,
      selNode.object_class,
      selNode.id
    );

    const { objects = {}, relation_paths = [] } = res || {};
    const {
      nodes: currentNodes = [],
      edges: currentEdges = [],
      combos: currentCombos = []
    } = graphInstanceRef.current.save();
    const nodes = Object.values(objects).map((item: any) => ({
      ...item,
      ...(item?.properties || {}),
      label: item.properties?.name,
      parentAssociateId: selNode.id,
      id: item.properties?.s_id
    }));

    const edges = relation_paths
      .map((item: any) => {
        const relation = item?.relations?.[0];

        return {
          ...relation,
          source: objects[relation.source_object_id]?.properties?.s_id,
          target: objects[relation.target_object_id]?.properties?.s_id,
          parentAssociateId: selNode.id
        };
      })
      .filter(
        (item: any) =>
          !currentEdges.some(
            (edge: any) =>
              edge.source === item.source && edge.target === item.target
          )
      );

    const showNodes = nodes.filter(
      ({ id }) =>
        id !== selNode.id &&
        !currentNodes.some((node: any) => node.id === id) &&
        !currentCombos.some((combo: any) => combo.id === id)
    );

    return { nodes: showNodes, edges };
  };

  const handleMenuClick = async (key: string, item: any) => {
    const {
      nodes: currentNodes = [],
      edges: currentEdges = [],
      combos: currentCombos = []
    } = graphInstanceRef.current.save();

    if (key === 'detail' || key === 'associateEvent') {
      setDetailNodeData(item);
      setDetailActiveKey(key);
      openDetailDrawer();
    }

    if (key === 'expandAssociateObject') {
      const { nodes, edges } = await getSubNodes(item);
      const newData = {
        combos: [...currentCombos],
        nodes: [...currentNodes, ...nodes],
        edges: [...currentEdges, ...edges]
      };

      updateGraphData(newData);
      graphInstanceRef.current.updateItem(item.id, {
        ...item,
        expandAssociate: true
      });
    }

    if (key === 'collapseAssociateObject') {
      const newData = {
        nodes: currentNodes.filter(
          (node: any) => node.parentAssociateId !== item.id
        ),
        combos: currentCombos,
        edges: currentEdges.filter(
          (edge: any) => edge.parentAssociateId !== item.id
        )
      };

      updateGraphData(newData);
      graphInstanceRef.current.updateItem(item.id, {
        ...item,
        expandAssociate: false
      });
    }

    if (key === 'expandFaultPoint') {
      const newNodes = currentNodes.filter((node: any) => node.id !== item?.id);
      const currentNodeFaultPoints =
        failurePointsObjRef.current[item?.id] || [];

      currentNodeFaultPoints.forEach((node: any) => {
        node.comboId = item?.id;
      });
      const newData = {
        nodes: [...newNodes, ...currentNodeFaultPoints],
        edges: currentEdges,
        combos: [...currentCombos, { ...item, expandFaultPoint: true }]
      };

      updateGraphData(newData);
    }

    if (key === 'collapseFaultPoint') {
      const newNodes = currentNodes.filter(
        (node: any) => node.comboId !== item?.id
      );
      const newData = {
        nodes: [
          ...newNodes,
          { ...omit(item, ['style', 'type']), expandFaultPoint: false }
        ],
        edges: currentEdges,
        combos: currentCombos.filter((combo: any) => combo.id !== item?.id)
      };

      updateGraphData(newData);
    }

    if (key === 'setRootCause') {
      if (!urlParams?.id || !item?.id) return;
      const res = await setRootCause(
        urlParams.id,
        item.entity_object_id,
        item.fault_id
      );

      if (res?.success === 1) {
        onSetRootCaseNode(item?.entity_object_id, item.fault_id);
        message.success(intl.get('set_root_cause_success'));
      } else {
        message.error(intl.get('set_root_cause_failed'));
      }
    }

    if (key === 'aiInterpretation') {
      const name = item?.fault_id ? 'fault_points' : 'entity_object';
      const dataFacts = item?.fault_id
        ? {
            ...item,
            relation_fault_point_ids:
              failurePointsObjRef.current[item?.id] || []
          }
        : item;

      const params: any = {
        title: '',
        data: {
          data_facts: dataFacts
        }
      };

      send(
        `${intl.get('interpretation')}${item.label}${intl.get(name)}`,
        params
      );
    }
  };

  const menu = new G6.Menu({
    offsetX: 6,
    offsetY: 10,
    itemTypes: ['node', 'combo'],
    getContent(evt: any) {
      const {
        expandAssociate = false,
        isFailurePoint,
        parentAssociateId,
        rootCause
      } = evt?.item?.getModel() || {};

      const type = evt?.item?.getType();
      const expandFaultPoint = type !== 'node';
      const setRootCauseMenu = !rootCause
        ? `<li attr='setRootCause'>${intl.get('set_root_cause')}</li>`
        : '';
      // 对象节点特有菜单项
      const objectNodeMenu = `
        ${
          expandAssociate
            ? `<li attr="collapseAssociateObject" >${intl.get(
                'collapse_related_objects'
              )}</li>`
            : `<li attr="expandAssociateObject">${intl.get(
                'expand_related_objects'
              )}</li>`
        }
       ${
         expandFaultPoint
           ? `<li attr="collapseFaultPoint">${intl.get(
               'collapse_failure_point'
             )}</li>`
           : `<li attr="expandFaultPoint">${intl.get(
               'show_failure_point'
             )}</li>`
       }`;
      const outDiv = document.createElement('div');
      const nodeMenuHtml = `<ul class="${styles['menu-wrap']}">
        <li attr='detail'>${intl.get('view_details')}</li>
        <li attr='associateEvent' class=${styles['menu-divider']}>${intl.get(
        'correlated_events2'
      )}</li>
      <li attr='aiInterpretation'>${intl.get('ai_interpretation')}</li>
        ${!isFailurePoint ? objectNodeMenu : setRootCauseMenu}
      </li>
      </ul>`;

      const associateNode = `<ul class="${styles['menu-wrap']}">
        <li attr='detail'>${intl.get('view_details')}</li>
      </ul>`;

      outDiv.style.width = '100px';
      outDiv.style.padding = '0';
      outDiv.innerHTML = parentAssociateId ? associateNode : nodeMenuHtml;

      return outDiv;
    },
    handleMenuClick(target, item) {
      handleMenuClick(target.getAttribute('attr') as string, item?._cfg?.model);
    }
  });

  const tooltip = new G6.Tooltip({
    itemTypes: ['node', 'combo'],
    container: containerRef.current!,
    shouldBegin: (e: any): boolean => {
      return !!e?.item?.getModel().allName;
    },
    // 自定义 tooltip 内容
    getContent: (e): string | HTMLDivElement => {
      const outDiv = document.createElement('div');

      outDiv.id = 'toolbarContainer';
      document.body.appendChild(outDiv);

      outDiv.style.width = 'fit-content';

      outDiv.innerHTML = `
                <div style="white-space: pre-wrap">${
                  e?.item?.getModel().allName
                }</div>
                `;

      return outDiv;
    }
  });

  const initGraph = (): any => {
    const { width = 0, height = 0 } = containerSize || {};

    return new G6.Graph({
      container: containerRef.current!,
      width,
      height,
      plugins: [tooltip, menu],
      ...g6DefaultOptions
    });
  };

  useEffect((): void => {
    const boxGraph: any = document.getElementById('graph-content');
    const height = boxGraph?.scrollHeight;
    const width = boxGraph?.scrollWidth;

    setContainerOriginSize({ height, width });
    const graph = initGraph();

    graphInstanceRef.current = graph;
    setGraphInstance(graph);
    graph.data(data);
    graph.render();

    graph.on('node:click', (evt: any) => {
      const { item } = evt;

      if (item.getModel().type === 'combo') return;

      item.toFront();
      graph.updateItem(item, {
        ...item.getModel(),
        checked: true
      });
    });
    graph.on('edge:click', (evt: any) => {
      const { item } = evt;

      graph.updateItem(item, {
        ...item.getModel(),
        checked: true
      });
    });
    graph.on('canvas:click', () => {
      graph.getNodes().forEach((node: any) => {
        graph.updateItem(node, {
          ...node.getModel(),
          checked: false
        });
      });
      graph.getEdges().forEach((edge: any) => {
        graph.updateItem(edge, {
          ...edge.getModel(),
          checked: false
        });
      });
    });
  }, []);

  // 播放前重置所有节点状态
  const onBeforePlay = (): void => {
    if (!graphInstanceRef.current) return;
    const {
      nodes: currentNodes = [],
      edges: currentEdges = [],
      combos: currentCombos = []
    } = graphInstanceRef.current.save();

    const newNodes = currentNodes.map((node: any) => ({
      ...node,
      checked: false,
      isPlaying: true,
      active: false,
      expandAssociate: false
    }));

    const newEdges = currentEdges.map((edge: any) => ({
      ...edge,
      checked: false,
      isPlaying: true,
      active: false,
      expandAssociate: false
    }));

    graphInstanceRef.current.destroyLayout();
    graphInstanceRef.current.changeData({
      nodes: delAssociateData(newNodes),
      edges: delAssociateData(newEdges),
      combos: delAssociateData(currentCombos)
    });
    graphInstanceRef.current.fitView(20);
  };

  // 播放时激活节点
  const onPlayActive = (id: string): void => {
    if (!graphInstanceRef.current) return;

    const { nodes: currentNodes = [], edges: currentEdges = [] } =
      graphInstanceRef.current.save();

    const activeNodes = currentNodes.filter(
      (node: any) =>
        id === node.id || node?.relation_fault_point_ids?.includes(id)
    );

    if (!activeNodes.length) return;

    activeNodes.forEach((node: any) => {
      graphInstanceRef.current.updateItem(node.id, {
        ...node,
        isPlaying: false,
        active: true
      });
    });

    const activeEdges = currentEdges.filter((edge: any) =>
      activeNodes.some(
        (v: any) => v.id === edge.target || v.entity_object_id === edge.target
      )
    );

    if (!activeEdges.length) return;

    activeEdges.forEach((edge: any) => {
      const graphEdge = graphInstanceRef.current.findById(edge.id);

      graphInstanceRef.current.updateItem(graphEdge, {
        isPlaying: false,
        active: true
      });
    });
  };

  const onSelectPatchActive = (ids: string[]): void => {
    if (!graphInstanceRef.current) return;
    const { nodes: currentNodes = [] } = graphInstanceRef.current.save();

    currentNodes.forEach((node: any) => {
      graphInstanceRef.current.updateItem(node.id, {
        ...node,
        checked: false,
        active: ids.some(
          (v) => v === node.id || node?.relation_fault_point_ids?.includes(v)
        )
      });
    });
  };

  // 重置图
  const onRestGraph = (): void => {
    if (!graphInstanceRef.current) return;
    const {
      nodes: currentNodes = [],
      edges: currentEdges = [],
      combos: currentCombos = []
    } = graphInstanceRef.current.save();

    graphInstanceRef.current.changeData({
      nodes: currentNodes.map((node: any) => ({
        ...node,
        checked: false,
        isPlaying: false,
        active: false
      })),
      edges: currentEdges.map((edge: any) => ({
        ...edge,
        checked: false,
        isPlaying: false,
        active: false
      })),
      combos: currentCombos
    });
  };

  useImperativeHandle(ref, () => ({
    onBeforePlay,
    onPlayActive,
    onSelectPatchActive,
    onRestGraph
  }));

  useUpdateEffect(() => {
    if (!graphInstanceRef.current) return;
    updateGraphData(data);
  }, [data]);

  return (
    <div className={styles['layout-wrapper']}>
      <Flex
        className={styles['content-wrapper']}
        vertical
        id="box-graph"
        ref={fullDomRef}
      >
        <ToolHeader
          graphInstance={graphInstance!}
          fullDom={fullDomRef}
          originalWidth={containerOriginSize?.width || 0}
          originalHeight={containerOriginSize?.height || 0}
          isShowLegend={isShowLegend}
          setIsShowLegend={setIsShowLegend}
          setIsShowFailurePoint={setIsShowFailurePoint}
          isShowFailurePoint={isShowFailurePoint}
          isPlaying={isPlaying}
        />
        <aside
          className={styles['graph-wrap']}
          ref={containerRef}
          id="graph-content"
        ></aside>
        <DetialDrawer
          visible={drawerVisible}
          onCancel={closeDetailDrawer}
          defaultActiveKey={detailActiveKey}
          knId={knId}
          nodeData={detailNodeData}
          failurePointNodes={failurePointNodes}
        />
        <SimpleLayout
          type="treeList"
          graphInstance={graphInstance!}
          onChange={onLayoutChange}
          isCombo
        />
        {isShowLegend && (
          <div className={styles['type-box']}>
            {!isShowFailurePoint &&
              iconList
                .filter(
                  (item) =>
                    data.nodes.some(
                      (node) => node.object_class === item.type
                    ) && item.icon
                )
                .map((val) => (
                  <dl key={val.name} className={styles['type-box-item']}>
                    <dt>
                      <ARIconfont type={val.value} />
                    </dt>
                    <dd>{val.name}</dd>
                  </dl>
                ))}
            {Object.values(Impact_Level_Maps).map((item) => (
              <dl key={item.name} className={styles['type-box-im']}>
                <dt style={{ backgroundColor: item.color }}></dt>
                <dd>{intl.get(item.name)}</dd>
              </dl>
            ))}
          </div>
        )}
      </Flex>
    </div>
  );
});

export default TopologyGraph;
