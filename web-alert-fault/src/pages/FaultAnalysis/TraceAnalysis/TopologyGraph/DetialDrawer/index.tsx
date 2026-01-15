import React, { useContext, useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Drawer, Tabs } from 'antd';
import { useAntdTable } from 'ahooks';
import { useParams } from '@noya/max';
import EventsList from '../../../CorrelatedEvents/components/EventsList';
import DetialInfo from './DetialInfo';
import styles from './index.module.less';
import { getDataViewList, getObjectDetail } from '@/services/fault-analysis';
import { transformTimeToMoment } from '@/components/ARTimePicker/TimePickerWithType';
import { generateDataViewFilters } from '@/utils/generate-data-view-filters';
import { faultAnalysisContext } from '@/pages/FaultAnalysis';

interface DetialDrawerProps {
  defaultActiveKey?: string;
  visible: boolean;
  onCancel: () => void;
  nodeData: any;
  knId: string;
  failurePointNodes: any[];
}

const DetialDrawer: React.FC<DetialDrawerProps> = ({
  visible,
  onCancel,
  defaultActiveKey,
  nodeData,
  knId,
  failurePointNodes
}) => {
  const [detialData, setDetialData] = useState({});
  const [isFailurePoint, setIsFailurePoint] = useState(false);
  const { id: problemId } = useParams();

  const { isGraphFullScreen } = useContext(faultAnalysisContext);

  const getDetailData = async () => {
    if (nodeData.isFailurePoint) {
      setIsFailurePoint(true);
      setDetialData(failurePointNodes.find((item) => item.id === nodeData.id));

      return;
    }

    if (knId && nodeData?.id) {
      const objectClass = nodeData.object_class || nodeData.object_type_id;
      const res = await getObjectDetail(knId, objectClass, nodeData?.id);

      setDetialData(res);
    }
  };

  const getCorrelatedEvents = async (params: any) => {
    const timeRange = sessionStorage.getItem('arTimeRange') || '[]';
    const dayjsTime = transformTimeToMoment(JSON.parse(timeRange));
    const [start, end] = dayjsTime.map((item) => item.valueOf());
    const { current, pageSize, sort = {}, filters = {} } = params;
    const sortAndFilters = generateDataViewFilters({
      sorter: sort,
      filters
    });

    const curFilters = {
      operation: 'and',
      sub_conditions: [
        {
          field: 'problem_id',
          operation: '==',
          value: problemId,
          value_from: 'const'
        }
      ]
    };

    if (nodeData?.isFailurePoint) {
      curFilters.sub_conditions.push({
        field: 'fault_id',
        operation: '==',
        value: nodeData?.fault_id,
        value_from: 'const'
      });
    } else {
      curFilters.sub_conditions.push({
        field: 'entity_object_id',
        operation: '==',
        value: nodeData?.id,
        value_from: 'const'
      });
    }

    const formattedParams: any = {
      filters: {
        ...curFilters,
        sub_conditions: [
          ...curFilters.sub_conditions,
          ...((sortAndFilters as any)?.sub_conditions || [])
        ]
      },
      start,
      end,
      offset: (current - 1) * pageSize,
      limit: pageSize
    };
    const res = await getDataViewList('__itops_raw_event', formattedParams);

    const { entries = [], total_count } = res;

    return {
      list: entries,
      total: total_count
    };
  };

  const { run, tableProps } = useAntdTable(getCorrelatedEvents, {
    manual: true,
    defaultPageSize: 10
  });

  useEffect(() => {
    if (visible) {
      getDetailData();

      const params = {};

      // 关联节点不展示关联事件
      if (nodeData.parentAssociateId) return;

      run({
        current: 1,
        pageSize: 10,
        ...params
      });

      return;
    }

    setDetialData({});
  }, [visible]);

  const tabsData = [
    {
      label: intl.get('detail_information'),
      key: 'detail',
      children: (
        <DetialInfo dataInfo={detialData} isFailurePoint={isFailurePoint} />
      )
    }
  ];

  if (!nodeData.parentAssociateId) {
    tabsData.push({
      label: intl.get('correlated_events2'),
      key: 'associateEvent',
      children: <EventsList scrollY={410} {...tableProps} />
    });
  }

  const otherProps: any = isGraphFullScreen
    ? {
        getContainer: false
      }
    : {};

  return (
    <Drawer
      title={nodeData?.allName || ''}
      placement="bottom"
      onClose={onCancel}
      open={visible}
      size={700}
      className={styles.drawer}
      destroyOnHidden
      {...otherProps}
    >
      <Tabs
        defaultActiveKey={defaultActiveKey || 'detail'}
        tabBarGutter={24}
        items={tabsData}
        className={styles.tabs}
      />
    </Drawer>
  );
};

export default DetialDrawer;
