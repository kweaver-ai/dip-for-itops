import { Table, Tag } from 'antd';
import { CaretDownOutlined, CaretRightOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import intl from 'react-intl-universal';
import { ProblemLevel } from 'Constants/problemTypes';
import { COMMON_PAGINATION, Level_Maps, Status_Maps } from 'Constants/common';
import styles from '../index.module.less';
import { EXPAND_FIELDS, TIMESTAMP_FIELDS } from '../constants';
import { CorrelatedEvent, EventStatus } from '@/constants/correlatedEventTypes';

interface EventsListProps {
  showExpand?: boolean;
  scrollY?: number;
  timeRange?: number[];
  [key: string]: any;
}

const EventsList = (props: EventsListProps) => {
  const {
    showExpand = true,
    scrollY,
    timeRange = [],
    level,
    status,
    ...tableProps
  } = props;
  const { pagination = {}, ...restTableProps } = tableProps;

  /*
   * 展开行状态管理
   * const [expandedRowKeys, setExpandedRowKeys] = useState<React.Key[]>([]);
   */

  const columns = [
    {
      title: intl.get('level'),
      dataIndex: 'event_level',
      key: 'event_level',
      width: 90,
      filteredValue: level,
      filters: Object.keys(Level_Maps).map(key => ({
        text: intl.get(Level_Maps[key as unknown as ProblemLevel]?.name),
        value: key
      })),
      render: (level: any) => {
        return (
          <span>
            <span
              className={styles.levelIcon}
              style={{
                backgroundColor:
                  Level_Maps[level as unknown as ProblemLevel]?.color
              }}
            />
            {intl.get(Level_Maps[level as unknown as ProblemLevel]?.name)}
          </span>
        );
      }
    },
    {
      title: intl.get('title'),
      dataIndex: 'event_title',
      key: 'event_title',
      ellipsis: true
    },
    {
      title: intl.get('status'),
      dataIndex: 'event_status',
      key: 'event_status',
      width: 100,
      filteredValue: status,
      filters: Object.keys(Status_Maps).map(key => ({
        text: intl.get(Status_Maps[key as unknown as EventStatus].name),
        value: key
      })),
      render: (text: string) => {
        return (
          <Tag
            variant="filled"
            color={Status_Maps[text as unknown as EventStatus]?.color ?? '#fff'}
          >
            {intl.get(Status_Maps[text as unknown as EventStatus]?.name) ??
              text}
          </Tag>
        );
      }
    },
    {
      title: intl.get('type'),
      dataIndex: 'event_type',
      key: 'event_type',
      width: 180,
      ellipsis: true
    },
    {
      title: intl.get('object'),
      dataIndex: 'entity_object_name',
      key: 'entity_object_name',
      width: 180,
      ellipsis: true,
      render: (text: string) => (
        <Tag variant="filled" color="#007aff" className="max-w-full truncate">
          {text}
        </Tag>
      )
    },
    {
      title: intl.get('occurred_time'),
      dataIndex: 'event_occur_time',
      key: 'event_occur_time',
      width: 180,
      textAlign: 'center',
      sorter: true,
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss')
    },
    {
      title: intl.get('source'),
      dataIndex: 'event_source',
      key: 'event_source',
      width: 180,
      ellipsis: true,
      render: (text: string) => (
        <Tag variant="filled" color="rgba(0, 0, 0, 0.85)">
          {text}
        </Tag>
      )
    }
  ];

  // 表格行展开内容
  const expandedRowRender = (record: CorrelatedEvent) => (
    <div className={styles.expandedContent}>
      {EXPAND_FIELDS.map(key => {
        return (
          <div className={styles.expandedItem} key={key}>
            <span className={styles.expandedLabel}>{intl.get(key)}：</span>
            <span>
              {TIMESTAMP_FIELDS.includes(key)
                ? dayjs(record[key as keyof CorrelatedEvent]).format(
                    'YYYY-MM-DD HH:mm:ss'
                  )
                : record[key as keyof CorrelatedEvent]}
            </span>
          </div>
        );
      })}
    </div>
  );

  return (
    <Table
      columns={columns}
      rowKey="event_id"
      scroll={{
        x: 1600,
        y:
          scrollY !== undefined && tableProps.pagination?.total > 7
            ? scrollY
            : undefined
      }}
      {...restTableProps}
      pagination={{
        ...pagination,
        ...COMMON_PAGINATION
      }}
      expandable={
        !showExpand
          ? undefined
          : {
              // expandedRowKeys,
              expandRowByClick: true,
              rowExpandable: () => true,
              expandedRowRender,
              expandIcon: ({ expanded, onExpand, record }) =>
                expanded ? (
                  <CaretDownOutlined onClick={e => onExpand(record, e)} />
                ) : (
                  <CaretRightOutlined onClick={e => onExpand(record, e)} />
                )
              /*
               * onExpand: (expanded, record) => {
               *   if (expanded) {
               *     setExpandedRowKeys([...expandedRowKeys, record.event_id]);
               *   } else {
               *     setExpandedRowKeys(
               *       expandedRowKeys.filter((key) => key !== record.event_id)
               *     );
               *   }
               * }
               */
            }
      }
      rowClassName={styles.tableRow}
      size="middle"
    />
  );
};

export default EventsList;
