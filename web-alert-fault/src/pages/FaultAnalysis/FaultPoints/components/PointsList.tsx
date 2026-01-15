import { useState, useRef } from 'react';
import { Table, Space, Button, Tag, Drawer } from 'antd';
import { CaretDownOutlined, CaretRightOutlined } from '@ant-design/icons';
import { useAntdTable, useBoolean } from '@noya/max';
import intl from 'react-intl-universal';
import dayjs from 'dayjs';
import { ImpactLevel, PaginationParams, Status } from 'Constants/commonTypes';
import { FaultPoint } from '../types';
import styles from '../index.module.less';
import { LevelsText } from '@/pages/AlertFault/constants';
import EventList from '@/pages/FaultAnalysis/CorrelatedEvents/components/EventsList';
import {
  COMMON_PAGINATION,
  Impact_Level_Maps,
  Status_Maps
} from '@/constants/common';
import { getDataViewList } from '@/services/fault-analysis';
import { generateDataViewFilters } from '@/utils/generate-data-view-filters';
import getFaultDurationTime from '@/utils/get_fault_duration_time';

interface PointsListProps {
  expanded?: boolean;
  [key: string]: any;
}

const PointsList = ({
  level,
  status,
  expanded = true,
  ...tableProps
}: PointsListProps) => {
  const { pagination = {}, ...restProps } = tableProps;
  // 展开行状态管理
  const [expandedRowKeys, setExpandedRowKeys] = useState<React.Key[]>([]);
  const [visible, { setTrue, setFalse }] = useBoolean(false);
  const relationIdsRef = useRef<string[]>([]);

  const getCorrelatedEvents = async (params: any) => {
    const { current, pageSize, start, end, sorter = {}, filters = {} } = params;

    const notEmptyFilters = {} as Record<string, any>;

    Object.keys(filters).forEach((key) => {
      if (filters[key]) {
        notEmptyFilters[key] = filters[key];
      }
    });

    // 加上 event_id 过滤
    notEmptyFilters.event_id = relationIdsRef.current;

    if (relationIdsRef.current.length === 0) {
      return {
        list: [],
        total: 0
      };
    }
    const filterParams = generateDataViewFilters({
      filters: notEmptyFilters,
      sorter
    });
    const formattedParams: PaginationParams = {
      start,
      end,
      offset: (current - 1) * pageSize,
      limit: pageSize,
      ...filterParams
    };
    const res = await getDataViewList('__itops_raw_event', formattedParams);

    const { entries = [], total_count } = res;

    return {
      list: entries,
      total: total_count
    };
  };

  const { run, tableProps: eventTableProps } = useAntdTable(
    getCorrelatedEvents,
    {
      manual: true,
      defaultPageSize: 10
    }
  );

  // 查看关联事件
  const viewRelatedEvents = (record: FaultPoint) => {
    if (record.relation_event_ids?.length === 0) {
      return;
    }

    relationIdsRef.current = record.relation_event_ids;

    run({
      current: 1,
      pageSize: 10
    });

    setTrue();
  };

  // 表格列配置
  const columns = [
    {
      title: intl.get('level'),
      dataIndex: 'fault_level',
      key: 'fault_level',
      width: 90,
      filteredValue: level,
      filters: Object.keys(LevelsText).map((key) => ({
        text: intl.get(Impact_Level_Maps[key as unknown as ImpactLevel]?.name),
        value: key
      })),
      render: (level: string) => {
        return (
          <span>
            <span
              className={styles.levelIcon}
              style={{
                backgroundColor:
                  Impact_Level_Maps[level as unknown as ImpactLevel]?.color
              }}
            />
            {intl.get(Impact_Level_Maps[level as unknown as ImpactLevel]?.name)}
          </span>
        );
      }
    },
    {
      title: intl.get('fault_point_details'),
      dataIndex: 'fault_description',
      key: 'fault_description',
      render: (detail: string, record: FaultPoint) => {
        return (
          <div>
            <div className={styles.detailMeta}>
              <span className="max-w-[1366px] truncate" title={detail}>
                {intl.get('fault_point_description')}: {detail}
              </span>
              <span>
                {intl.get('status')}:{' '}
                <Tag
                  color={Status_Maps[record.fault_status as Status].color}
                  variant="filled"
                >
                  {intl.get(
                    Status_Maps[record.fault_status as unknown as Status].name
                  )}
                </Tag>
              </span>
              <span>
                {intl.get('related_events')}:{' '}
                {record.relation_event_ids?.length ?? 0}
              </span>
            </div>
            <div className={styles.detailTime}>
              <span>
                {intl.get('start_time')}:{' '}
                {dayjs(record.fault_occur_time).format('YYYY-MM-DD HH:mm:ss')}
              </span>
              {/* <span>
                {intl.get('end_time')}: {record.endTime}
              </span> */}
              <span>
                {intl.get('duration')}: {getFaultDurationTime(record)}
              </span>
              <span>
                {intl.get('entity_object')}:{' '}
                <Tag variant="filled" color="#007aff">
                  {record.entity_object_name}
                </Tag>
              </span>
            </div>
          </div>
        );
      }
    },
    {
      title: intl.get('Operation'),
      key: 'action',
      width: 120,
      render: (text: string, record: FaultPoint) => (
        <Space size={8}>
          <Button
            size="small"
            type="link"
            onClick={(e) => {
              e.stopPropagation();
              viewRelatedEvents(record);
            }}
          >
            {intl.get('View')}
          </Button>
        </Space>
      )
    }
  ];

  // 表格行展开内容
  const expandedRowRender = (record: FaultPoint) => (
    <div className={styles.expandedContent}>
      <div className={styles.expandedItem}>
        <span className={styles.expandedLabel}>
          {intl.get('fault_point_id')}：
        </span>
        <span>{record.fault_id}</span>
      </div>
      <div className={styles.expandedItem}>
        <span className={styles.expandedLabel}>
          {intl.get('fault_point_name')}：
        </span>
        <span>{record.fault_name}</span>
      </div>
      <div className={styles.expandedItem}>
        <span className={styles.expandedLabel}>{intl.get('status')}：</span>
        <Tag
          color={Status_Maps[record.fault_status as Status].color}
          variant="filled"
        >
          {intl.get(Status_Maps[record.fault_status as unknown as Status].name)}
        </Tag>
      </div>
      <div className={styles.expandedItem}>
        <span className={styles.expandedLabel}>
          {intl.get('fault_point_start_time')}：
        </span>
        <span>
          {dayjs(record.fault_occur_time).format('YYYY-MM-DD HH:mm:ss')}
        </span>
      </div>
      <div className={styles.expandedItem}>
        <span className={styles.expandedLabel}>
          {intl.get('fault_point_duration')}：
        </span>
        <span>{getFaultDurationTime(record)}</span>
      </div>
      {record.fault_status === Status.Recovered && (
        <div className={styles.expandedItem}>
          <span className={styles.expandedLabel}>
            {intl.get('recover_time')}：
          </span>
          <span>
            {dayjs(record.fault_recovery_time).format('YYYY-MM-DD HH:mm:ss')}
          </span>
        </div>
      )}
      <div className={styles.expandedItem}>
        <span className={styles.expandedLabel}>
          {intl.get('entity_object')}：
        </span>
        <Tag variant="filled" color="#007aff">
          {record.entity_object_name}
        </Tag>
      </div>
    </div>
  );

  return (
    <>
      <Table
        columns={columns}
        rowKey="fault_id"
        scroll={{ x: 1600 }}
        {...restProps}
        pagination={{
          ...pagination,
          ...COMMON_PAGINATION
        }}
        expandable={
          !expanded
            ? undefined
            : {
                expandedRowKeys,
                expandRowByClick: true,
                rowExpandable: () => true,
                expandedRowRender,
                expandIcon: ({ expanded, onExpand, record }) =>
                  expanded ? (
                    <CaretDownOutlined onClick={(e) => onExpand(record, e)} />
                  ) : (
                    <CaretRightOutlined onClick={(e) => onExpand(record, e)} />
                  ),
                onExpand: (expanded, record) => {
                  if (expanded) {
                    setExpandedRowKeys([...expandedRowKeys, record.fault_id]);
                  } else {
                    setExpandedRowKeys(
                      expandedRowKeys.filter((key) => key !== record.fault_id)
                    );
                  }
                }
              }
        }
        rowClassName={styles.tableRow}
        size="middle"
      />
      <Drawer
        title={intl.get('related_events')}
        open={visible}
        placement="bottom"
        onClose={() => setFalse()}
        size={570}
      >
        <EventList {...eventTableProps} />
      </Drawer>
    </>
  );
};

export default PointsList;
