/* eslint-disable react-hooks/exhaustive-deps */
import React, { useRef, useState, useEffect, useContext, useMemo } from 'react';
import type { IGraph } from '@antv/g6';
import G6 from '@antv/g6';
import { useSize } from 'ahooks';
import { faultAnalysisContext } from '../../index';
import styles from './index.module.less';
import SimpleLayout from '@/pages/FaultAnalysis/TraceAnalysis/TopologyGraph/SimpleLayout';
import { dagreDefaultConfig } from '@/pages/FaultAnalysis/TraceAnalysis/TopologyGraph/SimpleLayout/utils';
import g6DefaultOptions from '@/pages/FaultAnalysis/TraceAnalysis/TopologyGraph/constants/g6Oprtions';
import {
  customNodeDraw,
  customNodeAfterDraw
} from '@/pages/FaultAnalysis/TraceAnalysis/TopologyGraph/utils/overviewNode';
import { faultPointDraw } from '@/pages/FaultAnalysis/TraceAnalysis/TopologyGraph/utils/faultPointNode';
import originDataToG6Data from '@/pages/FaultAnalysis/TraceAnalysis/TopologyGraph/utils/getG6Data';

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

const TopologyGraph = () => {
  const [graphInstance, setGraphInstance] = useState<IGraph | null>(null);
  const fullDomRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const containerSize = useSize(containerRef);
  const graphInstanceRef = useRef<any>(null);
  const failurePointsObjRef = useRef<any>({});

  const {
    problemData: {
      rca_results: {
        rca_context: { backtrace = [], network: networkData = {} }
      },
      root_cause_object_id: rootCauseObjectId
    }
  } = useContext(faultAnalysisContext);

  const { data, failurePointsObj } = useMemo(
    () =>
      originDataToG6Data({
        networkData,
        backtrace,
        rootCauseObjectId,
        rootCauseFaultId: '',
        isShowFailurePoint: false
      }),
    [networkData, rootCauseObjectId]
  );

  failurePointsObjRef.current = failurePointsObj;

  const tooltip = new G6.Tooltip({
    itemTypes: ['node'],
    container: containerRef.current!,
    shouldBegin: (e: any): boolean => {
      return !!e?.item?.getModel().label;
    },
    // 自定义 tooltip 内容
    getContent: (e): string | HTMLDivElement => {
      const outDiv = document.createElement('div');

      outDiv.id = 'toolbarContainer';
      document.body.appendChild(outDiv);

      outDiv.style.width = 'fit-content';

      outDiv.innerHTML = `
                <div style="white-space: pre-wrap">${
                  e?.item?.getModel().label
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
      plugins: [tooltip],
      ...{ ...g6DefaultOptions, layout: dagreDefaultConfig }
    });
  };

  useEffect((): void => {
    const graph = initGraph();

    graphInstanceRef.current = graph;
    setGraphInstance(graph);
    graph.data(data);
    graph.render();
  }, []);

  useEffect(() => {
    if (!graphInstanceRef.current) return;
    graphInstanceRef.current.changeData(data);
    graphInstanceRef.current.render();
  }, [data]);

  return (
    <div className={styles['layout-wrapper']} id="box-graph" ref={fullDomRef}>
      <aside
        className={styles['graph-wrap']}
        ref={containerRef}
        id="impact-graph-content"
      ></aside>
      <SimpleLayout type="treeList" graphInstance={graphInstance!} />
    </div>
  );
};

export default TopologyGraph;
