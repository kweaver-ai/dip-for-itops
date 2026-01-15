/* eslint-disable max-lines */
import { useEffect, useMemo, useRef, useState } from 'react';
import { Table, Row, Col, Radio, Button, Tag, Input, Space } from 'antd';
import * as echarts from 'echarts';
import classnames from 'classnames';
import { useNavigate, useAntdTable, useParams } from '@noya/max';
import intl from 'react-intl-universal';
import dayjs from 'dayjs';
import { ColumnsType } from 'antd/es/table';
import {
  Problem,
  ProblemLevel,
  ProblemListMode,
  ProblemStatus
} from 'Constants/problemTypes';
import { PaginationParams } from 'Constants/commonTypes';
import {
  getAlertFaultList,
  getAlertFaultStatusCount,
  closeAlertFault
} from 'Services/alert-fault';
import { CaretDownOutlined, CaretRightOutlined } from '@ant-design/icons';
import {
  EXPAND_FIELDS,
  LevelsText,
  StatusesText,
  statusStyle,
  TIME_FIELDS
} from './constants';
import styles from './index.module.less';
import FaultHeader from '@/organisms/FaultHeader';
import ARIconfont from '@/components/ARIconfont';
import ARIconfont5 from '@/components/ARIconfont5';
import { generateBarChartConfig } from '@/utils/generate-bar-chart-config';
import {
  COMMON_PAGINATION,
  DEFAULT_PAGE_SIZE,
  Level_Maps
} from '@/constants/common';
import { transformOriginToChart } from '@/utils/transform-origin-to-chart';
import { useRefreshPage } from '@/hooks/useRefreshPage';
import {
  generateDataViewFilters,
  generateMetricModelFilters
} from '@/utils/generate-data-view-filters';
import { transformOriginToStatistic } from '@/utils/transform-origin-to-statistic';
import { useStatistic } from '@/hooks/useStatistic';
import getProblemDurationTime from '@/utils/get_problem_duration_time';
import { useTableFilter } from '@/hooks/useTableFilter';

const ProblemPage = () => {
  const navigate = useNavigate();
  const [displayMode, setDisplayMode] = useState(ProblemListMode.List);
  const [timeRange, setTimeRange] = useState<number[]>([]);
  const [refreshInterval, setRefreshInterval] = useState(0);
  const [problemLevelCount, setProblemLevelCount] = useState<number[]>([
    0, 0, 0, 0
  ]);
  const [problemStatusCount, setProblemStatusCount] = useState<Array<number>>([
    0, 0, 0
  ]);
  const [keyword, setKeyword] = useState('');
  const { level, status, onLevelChange, onStatusChange } = useStatistic();
  const { filterLevel, filterStatus, setFilterLevel, setFilterStatus } =
    useTableFilter({
      level,
      status
    });

  const urlParams = useParams();

  const updateStatus = (params: any) => {
    getStatusCount(params).then((res) => {
      const statusCount = [] as number[];

      Object.keys(statusStyle).forEach((key) => {
        if (key === String(ProblemStatus.Open)) {
          statusCount[0] = res[key] ?? 0;
        }
        if (key === String(ProblemStatus.Closed)) {
          statusCount[1] = res[key] ?? 0;
        }
        if (key === String(ProblemStatus.Invalid)) {
          statusCount[2] = res[key] ?? 0;
        }
      });
      setProblemStatusCount(statusCount);
    });
  };

  const closeProblem = async (status: ProblemStatus, problemId: string) => {
    if (status !== ProblemStatus.Open) return;
    const res = await closeAlertFault(problemId);

    if (!res.error_code) {
      run({ current: 1, pageSize: 10, start: timeRange[0], end: timeRange[1] });
      updateStatus({
        start: timeRange[0],
        end: timeRange[1],
        filters: {}
      });
    }
  };

  const commonColumn = [
    {
      title: intl.get('Operation'),
      key: 'action',
      dataIndex: 'action',
      width: 90,
      // @ts-ignore
      render: (_, record: Problem) => (
        <div>
          <a
            className={styles.operationButton}
            onClick={() => navigate(`/fault-analysis/${record.problem_id}`)}
          >
            {intl.get('View')}
          </a>
          <a
            className={classnames(styles.operationButton, {
              [styles.disableButton]:
                record.problem_status !== ProblemStatus.Open
            })}
            onClick={(e) => {
              e.stopPropagation();
              closeProblem(record.problem_status, record.problem_id);
            }}
          >
            {intl.get('Closed')}
          </a>
        </div>
      )
    }
  ];

  // 表格列表模式配置
  const columns = [
    {
      title: intl.get('level'),
      dataIndex: 'problem_level',
      key: 'problem_level',
      width: 90,
      filteredValue: filterLevel,
      filters: Object.keys(LevelsText).map((item) => ({
        text: intl.get(
          Level_Maps[item as unknown as keyof typeof Level_Maps].name
        ),
        value: item
      })),
      render: (text: number) => {
        return (
          <span>
            <span
              className={styles.levelIcon}
              style={{
                backgroundColor:
                  Level_Maps[text as unknown as keyof typeof Level_Maps].color
              }}
            />
            {intl.get(
              Level_Maps[text as unknown as keyof typeof Level_Maps].name
            )}
          </span>
        );
      }
    },
    {
      title: intl.get('problem_name'),
      dataIndex: 'problem_name',
      key: 'problem_name',
      ellipsis: true
    },
    {
      title: intl.get('status'),
      dataIndex: 'problem_status',
      key: 'problem_status',
      width: 80,
      filteredValue: filterStatus,
      filters: Object.keys(StatusesText).map((item) => ({
        text: intl.get(StatusesText[item as unknown as ProblemStatus]),
        value: item
      })),
      render: (text: string) => {
        return (
          <Tag
            variant="filled"
            color={statusStyle[text as unknown as ProblemStatus].color}
          >
            {intl.get(StatusesText[text as unknown as ProblemStatus])}
          </Tag>
        );
      }
    },
    {
      title: intl.get('impacted_objects'),
      dataIndex: 'affected_entity_ids',
      key: 'affected_entity_ids',
      width: 80,
      render: (text: string[], record: Problem) =>
        record.affected_entity_ids.length
    },
    {
      title: intl.get('related_events'),
      dataIndex: 'relation_event_ids',
      key: 'relation_event_ids',
      width: 80,
      render: (text: string[], record: Problem) =>
        record.relation_event_ids?.length ?? 0
    },
    {
      title: intl.get('root_cause_objects'),
      dataIndex: 'root_cause_entity_object_name',
      key: 'root_cause_entity_object_name',
      ellipsis: true,
      width: 150,
      render: (text: string) => (
        <Tag variant="filled" color="#007aff" className="max-w-full truncate">
          {text}
        </Tag>
      )
    },
    {
      title: intl.get('root_cause_fault_points'),
      dataIndex: 'root_cause_fault_description',
      key: 'root_cause_fault_description',
      width: 200,
      ellipsis: true
    },
    {
      title: intl.get('occurred_time'),
      dataIndex: 'problem_occur_time',
      key: 'problem_occur_time',
      width: 180,
      sorter: true,
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss')
    },
    {
      title: intl.get('end_time'),
      dataIndex: 'problem_latest_time',
      key: 'problem_latest_time',
      width: 180,
      sorter: true,
      render: (text: string, record: Problem) =>
        record.problem_status === ProblemStatus.Closed
          ? dayjs(record.problem_close_time).format('YYYY-MM-DD HH:mm:ss')
          : '--'
    },
    {
      title: intl.get('duration'),
      dataIndex: 'problem_duration',
      key: 'problem_duration',
      width: 120,
      ellipsis: true,
      render: (text: number, record: Problem) => getProblemDurationTime(record)
    }
  ].concat(commonColumn as any);
  // 表格详情模式配置
  const detailColumns: ColumnsType<Problem> = [
    {
      title: intl.get('level'),
      dataIndex: 'problem_level',
      key: 'problem_level',
      width: 80,
      filters: Object.keys(LevelsText).map((item) => ({
        text: intl.get(
          Level_Maps[item as unknown as keyof typeof Level_Maps].name
        ),
        value: item
      })),
      render: (level: number) => {
        return (
          <span>
            <span
              className={styles.levelIcon}
              style={{
                backgroundColor:
                  Level_Maps[level as unknown as keyof typeof Level_Maps]?.color
              }}
            />
            {intl.get(
              Level_Maps[level as unknown as keyof typeof Level_Maps]?.name
            )}
          </span>
        );
      }
    },
    {
      title: intl.get('problem_details'),
      dataIndex: 'detail',
      key: 'detail',
      render: (detail: string, record: Problem) => {
        return (
          <div>
            <div className={styles.detailMeta}>
              <span>
                {intl.get('problem_name')}: {record.problem_name}
              </span>
              <span>
                {intl.get('status')}:{' '}
                <Tag
                  variant="filled"
                  color={
                    statusStyle[record.problem_status as ProblemStatus].color
                  }
                >
                  {intl.get(
                    StatusesText[record.problem_status as ProblemStatus]
                  )}
                </Tag>
              </span>
              <span>
                {intl.get('impacted_objects')}:{' '}
                {record.affected_entity_ids.length}
              </span>
              <span>
                {intl.get('related_events')}:{' '}
                {record.relation_event_ids?.length ?? 0}
              </span>
              <span>
                {intl.get('start_time')}:{' '}
                {dayjs(record.problem_occur_time).format('YYYY-MM-DD HH:mm:ss')}
              </span>
              <span>
                {intl.get('end_time')}:{' '}
                {record.problem_status === ProblemStatus.Closed
                  ? dayjs(record.problem_close_time).format(
                      'YYYY-MM-DD HH:mm:ss'
                    )
                  : '--'}
              </span>
              <span>
                {intl.get('duration')}: {getProblemDurationTime(record)}
              </span>
            </div>
            <div className={styles.detailTime}>
              <span>
                {intl.get('root_cause_objects')}:{' '}
                <span className={styles.rootCauseObject}>
                  {record.root_cause_entity_object_name}
                </span>
              </span>
              <span
                className="max-w-[1366px] truncate"
                title={record.root_cause_fault_description}
              >
                {intl.get('root_cause_fault_points')}:{' '}
                <span className={styles.entityLink}>
                  {record.root_cause_fault_description}
                </span>
              </span>
            </div>
          </div>
        );
      }
    }
  ].concat(commonColumn as any);

  // ECharts实例引用
  const trendChartRef = useRef<HTMLDivElement>(null);
  const pieChartRef = useRef<HTMLDivElement>(null);
  const trendChartInstance = useRef<echarts.ECharts | null>(null);
  const pieChartInstance = useRef<echarts.ECharts | null>(null);

  const generateColumnConfig = (data: {
    xAxisData: string[];
    seriesData: any[];
    title?: string;
  }) => {
    // 问题趋势图表
    if (trendChartInstance.current) {
      const trendOption = generateBarChartConfig(data);

      trendChartInstance.current.setOption(trendOption, true);
    }
  };

  const generatePieConfig = (data: any[]) => {
    // 根因分布饼图
    if (pieChartInstance.current) {
      const seriesData = data || [
        { value: 323, name: '不可用', itemStyle: { color: '#1890ff' } },
        { value: 132, name: '性能下降', itemStyle: { color: '#52c41a' } },
        { value: 123, name: '资源不足', itemStyle: { color: '#faad14' } },
        { value: 123, name: '错误', itemStyle: { color: '#ff4d4f' } }
      ];
      const pieOption = {
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(255, 255, 255, 0.95)',
          borderColor: '#e8e8e8',
          textStyle: {
            color: '#333'
          },
          formatter: '{b}: {c} ({d}%)'
        },
        legend: {
          orient: 'vertical',
          right: 0,
          left: '70%',
          top: 'center',
          formatter(name: string) {
            // 截断函数，确保文本不超过指定宽度
            const truncateText = (text: string, maxLength: number) => {
              if (text.length <= maxLength) return text;

              return `${text.substring(0, maxLength - 1)}...`;
            };

            // 获取对应的数值
            const item = seriesData.find((item) => item.name === name);
            const value = item?.value ?? '-';
            // 图例名称最大显示6个字符，剩余空间显示数值
            const truncatedName = truncateText(name, 6);

            return `${truncatedName}: ${value}`;
          },
          textStyle: {
            color: '#666',
            fontSize: 12,
            overflow: 'truncate',
            width: 80 // 限制图例文本的最大宽度
          },
          itemWidth: 8,
          itemHeight: 8
        },
        series: [
          {
            name: '问题根因分布',
            type: 'pie',
            radius: ['50%', '70%'],
            center: ['35%', '55%'],
            avoidLabelOverlap: false,
            label: {
              show: false,
              position: 'center'
            },
            emphasis: {
              disabled: true
            },
            labelLine: {
              show: false
            },
            data: seriesData
          },
          {
            type: 'pie',
            radius: ['50%', '70%'],
            center: ['35%', '55%'],
            avoidLabelOverlap: false,
            emphasis: {
              disabled: true
            },
            label: {
              show: true,
              position: 'center',
              formatter() {
                const total = seriesData.reduce(
                  (acc, cur) => acc + cur.value,
                  0
                );

                return [`{val|${total}}`, '{field|问题总数}'].join('\n');
              },
              rich: {
                val: {
                  fontSize: 24,
                  fontWeight: 'bold',
                  lineHeight: 30,
                  color: '#333'
                },
                field: {
                  fontSize: 12,
                  color: '#666'
                }
              }
            },
            zlevel: -99,
            tooltip: {
              extraCssText: 'display: none;'
            },
            labelLine: {
              show: false
            },
            data: seriesData
          }
        ]
      };

      pieChartInstance.current.setOption(pieOption, true);
    }
  };

  // 初始化图表
  useEffect(() => {
    if (trendChartRef.current) {
      trendChartInstance.current = echarts.init(trendChartRef.current);
    }
    if (pieChartRef.current) {
      pieChartInstance.current = echarts.init(pieChartRef.current);
    }
    // 窗口大小变化时重新绘制图表
    const handleResize = () => {
      trendChartInstance.current?.resize();
      pieChartInstance.current?.resize();
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      trendChartInstance.current?.dispose();
      pieChartInstance.current?.dispose();
    };
  }, []);

  // 表格行展开内容
  const expandedRowRender = (record: Problem) => (
    <div className={styles.expandedContent}>
      {EXPAND_FIELDS.map((key) => {
        // 关闭状态
        if (key === 'problem_close_type') {
          return (
            <div className={styles.expandedItem} key={key}>
              <span className={styles.expandedLabel}>
                {intl.get('problem_close_type')}：
              </span>
              <span>
                {record.problem_close_type === '1'
                  ? intl.get('problem_close_status_system')
                  : record.problem_close_type === '2'
                  ? intl.get('problem_close_status_manual')
                  : intl.get('problem_close_status_unknown')}
              </span>
            </div>
          );
        }

        // 关联故障点数量
        if (key === 'relation_fp_ids') {
          return (
            <div className={styles.expandedItem} key={key}>
              <span className={styles.expandedLabel}>
                {intl.get('fault_points')}：
              </span>
              <span>{record.relation_fp_ids?.length ?? 0}</span>
            </div>
          );
        }

        return (
          <div className={styles.expandedItem} key={key}>
            <span className={styles.expandedLabel}>{intl.get(key)}：</span>
            <span>
              {TIME_FIELDS.includes(key)
                ? record[key as keyof Problem]
                  ? dayjs(record[key as keyof Problem] as string).format(
                      'YYYY-MM-DD HH:mm:ss'
                    )
                  : '--'
                : record[key as keyof Problem]}
            </span>
          </div>
        );
      })}
    </div>
  );

  // 初始化获取数据
  const getTableList = async (params: any) => {
    const {
      current,
      pageSize = DEFAULT_PAGE_SIZE,
      start,
      end,
      sorter,
      filters
    } = params;
    const sortAndFilters = generateDataViewFilters({
      sorter,
      // 不要直接修改filters，会影响其他地方的使用
      filters: {
        ...filters,
        ...(keyword ? { problem_name: keyword } : {})
      },
      likeFields: ['problem_name'],
      extraOriginFilters: [
        // 问题列表去掉被合并状态的问题
        {
          field: 'problem_status',
          operation: '!=',
          value: '3',
          value_from: 'const'
        }
      ]
    });
    const defaultSort = [
      {
        field: 'problem_status',
        direction: 'asc' as const
      },
      {
        field: 'problem_level',
        direction: 'asc' as const
      },
      {
        field: 'problem_occur_time',
        direction: 'desc' as const
      }
    ];
    const formattedParams: PaginationParams = {
      start,
      end,
      offset: (current - 1) * pageSize,
      limit: pageSize,
      filters: sortAndFilters.filters,
      ...(sortAndFilters.sort
        ? {
            sort: [
              ...sortAndFilters.sort,
              ...defaultSort.filter((item) =>
                sortAndFilters.sort?.findIndex(
                  (sort) => sort.field === item.field
                )
              )
            ]
          }
        : {
            sort: defaultSort
          })
    };
    const res = await getAlertFaultList(formattedParams);
    const { entries = [], total_count } = res;

    return {
      list: entries,
      total: total_count
    };
  };

  const getLevelFilters = (originFilters: any) => {
    const sortAndFilters = generateMetricModelFilters({
      filters: originFilters
    });

    return [
      {
        name: 'problem_status',
        operation: '!=' as const,
        value: '3'
      }
    ].concat(sortAndFilters.filters as any);
  };

  // 获取图表数据
  const getChartData = async (params: {
    start: number;
    end: number;
    filters: any;
  }) => {
    const filters = getLevelFilters(params.filters);
    const res = await getAlertFaultStatusCount('itops_problem_level_sum', {
      ...params,
      filters
    });

    return transformOriginToChart(res, 'problem_level');
  };

  // 获取等级统计数据
  const getLevelCount = async (params: {
    start: number;
    end: number;
    filters: any;
  }) => {
    const filters = getLevelFilters(params.filters);
    const res = await getAlertFaultStatusCount('itops_problem_level_sum', {
      ...params,
      filters,
      instant: true
    });
    const data = transformOriginToChart(res, 'problem_level');
    const levelCount = [0, 0, 0, 0];

    data.seriesData.forEach((item) => {
      if (item.key === String(ProblemLevel.Urgent)) {
        levelCount[0] = item.totals;
      }
      if (item.key === String(ProblemLevel.Severe)) {
        levelCount[1] = item.totals;
      }
      if (item.key === String(ProblemLevel.Important)) {
        levelCount[2] = item.totals;
      }
      if (item.key === String(ProblemLevel.Warning)) {
        levelCount[3] = item.totals;
      }
    });
    setProblemLevelCount(levelCount);
  };

  // 获取状态统计数据
  const getStatusCount = async (params: {
    start: number;
    end: number;
    filters: any;
  }) => {
    const res = await getAlertFaultStatusCount('itops_problem_status_sum', {
      ...params,
      ...generateMetricModelFilters(params.filters),
      instant: true
    });

    return transformOriginToStatistic(res, 'problem_status');
  };

  const { run, tableProps, pagination } = useAntdTable(getTableList, {
    manual: true,
    defaultPageSize: DEFAULT_PAGE_SIZE
  });

  const refreshPage = () => {
    if (!timeRange.length) return;

    const params = {
      start: timeRange[0],
      end: timeRange[1],
      filters: {
        problem_level: level,
        problem_status: status
      }
    };

    run({ current: 1, ...params });
    getChartData(params).then((data) => {
      const columnData = data.seriesData.map((item) => {
        const { totals, key, ...rest } = item;

        return rest;
      });

      generateColumnConfig({
        xAxisData: data.xAxisData,
        seriesData: columnData
      });
      const pieData = data.seriesData.map((item) => ({
        key: item.key,
        name: item.name,
        value: item.totals,
        itemStyle: item.itemStyle
      }));

      generatePieConfig(pieData);
    });
    getLevelCount(params);
    updateStatus(params);
  };

  const onTimeChange = (value: number[]) => {
    setTimeRange(value);
  };

  const onRefreshChange = (value: number) => {
    setRefreshInterval(value);
  };

  useRefreshPage({
    refreshInterval,
    timeRange,
    level,
    status,
    refreshPage
  });

  const paginationProps = useMemo(
    () => ({
      ...pagination,
      ...COMMON_PAGINATION
    }),
    [pagination]
  );

  return (
    <div className={styles.problemContainer}>
      {/* 顶部导航栏 */}
      <FaultHeader
        data={problemLevelCount.concat(problemStatusCount)}
        dataType="problem"
        onLevelChange={onLevelChange}
        onStatusChange={onStatusChange}
        extra={[
          <Button
            key="timeRange"
            shape="square"
            icon={<ARIconfont5 type="icon-quanjingtu-huifu" />}
            onClick={() => {
              refreshPage();
            }}
          />
        ]}
        onTimeChange={onTimeChange}
        onRefreshChange={onRefreshChange}
      />

      {/* 图表区域 */}
      <Row gutter={[16, 16]} className={styles.chartRow}>
        <Col span={20}>
          <div className={styles.chartCard}>
            <div className={styles.chartTitle}>
              <span>{intl.get('problem_trend')}</span>
            </div>
            <div ref={trendChartRef} className={styles.chartContainer}></div>
          </div>
        </Col>
        <Col span={4}>
          <div className={styles.chartCard}>
            <div className={styles.chartTitle}>
              <span>{intl.get('root_cause_distribution')}</span>
            </div>
            <div ref={pieChartRef} className={styles.chartContainer}></div>
          </div>
        </Col>
      </Row>

      {/* 操作栏 */}
      <div className={styles.operationBar}>
        <Space.Compact>
          <Space.Addon>
            <span style={{ whiteSpace: 'nowrap' }}>
              {intl.get('problem_name')}
            </span>
          </Space.Addon>
          <Input
            placeholder={intl.get('Input')}
            value={keyword}
            onChange={(e) => {
              setKeyword(e.target.value);
            }}
            onPressEnter={() => {
              const filters: Record<string, any> = {};

              if (filterLevel.length > 0) {
                filters.problem_level = filterLevel;
              }
              if (filterStatus.length > 0) {
                filters.problem_status = filterStatus;
              }
              if (keyword) {
                filters.problem_name = keyword;
              }

              run({
                current: 1,
                start: timeRange[0],
                end: timeRange[1],
                filters
              });
            }}
          />
        </Space.Compact>
        <div className={styles.tableTitle}>
          <Radio.Group
            value={displayMode}
            onChange={(e) => setDisplayMode(e.target.value)}
          >
            <Radio.Button value={ProblemListMode.List}>
              <ARIconfont type="icon-wuxuliebiao1" />
            </Radio.Button>
            <Radio.Button value={ProblemListMode.Detail}>
              <ARIconfont type="icon-liebiao-icon" />
            </Radio.Button>
          </Radio.Group>
        </div>
      </div>

      {/* 问题列表 */}
      <div className={styles.tableCard}>
        <Table
          columns={
            displayMode === ProblemListMode.List ? columns : detailColumns
          }
          {...tableProps}
          onChange={(pagination, filters, sorter) => {
            setFilterLevel(filters.problem_level ?? []);
            setFilterStatus(filters.problem_status ?? []);
            tableProps.onChange(pagination, filters, sorter);
          }}
          expandable={{
            // expandedRowKeys,
            expandRowByClick: true,
            rowExpandable: () => true,
            expandedRowRender,
            expandIcon: ({ expanded, onExpand, record }) =>
              expanded ? (
                <CaretDownOutlined onClick={(e) => onExpand(record, e)} />
              ) : (
                <CaretRightOutlined onClick={(e) => onExpand(record, e)} />
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
          }}
          rowKey="problem_id"
          pagination={paginationProps}
          scroll={{ x: 1600 }}
          size="middle"
          className={styles.tableWrapper}
        />
      </div>
    </div>
  );
};

export default ProblemPage;
