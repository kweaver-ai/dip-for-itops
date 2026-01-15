import React, { useEffect, useRef, useState } from 'react';
import { Space, Input, Button } from 'antd';
import * as echarts from 'echarts';
import { useAntdTable, useParams } from '@noya/max';
import intl from 'react-intl-universal';
import EventsList from './components/EventsList';
import styles from './index.module.less';
import FaultHeader from '@/organisms/FaultHeader';
import { ProblemLevel } from '@/constants/problemTypes';
import { CorrelatedEventProps } from '@/constants/correlatedEventTypes';
import { PaginationParams, Status } from '@/constants/commonTypes';
import { getDataViewList } from '@/services/fault-analysis';
import { getAlertFaultStatusCount } from '@/services/alert-fault';
import { DEFAULT_PAGE_SIZE, Status_Maps } from '@/constants/common';
import { transformOriginToChart } from '@/utils/transform-origin-to-chart';
import { transformOriginToStatistic } from '@/utils/transform-origin-to-statistic';
import { generateBarChartConfig } from '@/utils/generate-bar-chart-config';
import { useRefreshPage } from '@/hooks/useRefreshPage';
import {
  generateDataViewFilters,
  generateMetricModelFilters
} from '@/utils/generate-data-view-filters';
import ARIconfont5 from '@/components/ARIconfont5';
import { useStatistic } from '@/hooks/useStatistic';
import { useTableFilter } from '@/hooks/useTableFilter';

const CorrelatedEvents: React.FC<CorrelatedEventProps> = ({ onDataChange }) => {
  const trendChartRef = useRef<HTMLDivElement>(null);
  const trendChartInstance = useRef<echarts.ECharts | null>(null);
  const [timeRange, setTimeRange] = useState<number[]>([]);
  const [refreshInterval, setRefreshInterval] = useState(0);
  const [keyword, setKeyword] = useState('');
  const [problemLevelCount, setProblemLevelCount] = useState<number[]>([
    0, 0, 0, 0
  ]);
  const [problemStatusCount, setProblemStatusCount] = useState<Array<number>>([
    0, 0
  ]);
  const { id } = useParams();

  const onTimeChange = (value: number[]) => {
    setTimeRange(value);
  };

  const onRefreshChange = (value: number) => {
    setRefreshInterval(value);
  };

  // 生成问题趋势图表
  const generateChart = async (data: {
    xAxisData: string[];
    seriesData: any[];
    title: string;
  }) => {
    if (!trendChartInstance.current) return;

    const chartData = generateBarChartConfig(data);

    trendChartInstance.current.setOption(chartData, true);
  };

  useEffect(() => {
    // 问题趋势图表
    if (trendChartRef.current) {
      trendChartInstance.current = echarts.init(trendChartRef.current);
    }

    // 窗口大小变化时重新绘制图表
    const handleResize = () => {
      trendChartInstance.current?.resize();
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      trendChartInstance.current?.dispose();
    };
  }, []);

  // 获取图表数据
  const getChartData = async (params: {
    start: number;
    end: number;
    filters: any;
  }) => {
    const { filters = {} } = params;

    if (id) {
      filters.problem_id = id;
    }
    const res = await getAlertFaultStatusCount('itops_raw_event_level_sum', {
      ...params,
      ...generateMetricModelFilters(filters)
    });

    return transformOriginToChart(res, 'event_level');
  };

  const getLevelCount = async (params: {
    start: number;
    end: number;
    filters: any;
  }) => {
    const { filters = {} } = params;

    if (id) {
      filters.problem_id = id;
    }
    const res = await getAlertFaultStatusCount('itops_raw_event_level_sum', {
      ...params,
      ...generateMetricModelFilters(filters),
      instant: true
    });
    const data = transformOriginToChart(res, 'event_level');
    const levelStatistic = [0, 0, 0, 0, 0];

    data.seriesData.forEach((item) => {
      if (item.key === String(ProblemLevel.Urgent)) {
        levelStatistic[0] = item.totals;
      } else if (item.key === String(ProblemLevel.Severe)) {
        levelStatistic[1] = item.totals;
      } else if (item.key === String(ProblemLevel.Important)) {
        levelStatistic[2] = item.totals;
      } else if (item.key === String(ProblemLevel.Warning)) {
        levelStatistic[3] = item.totals;
      } else if (item.key === String(ProblemLevel.Info)) {
        levelStatistic[4] = item.totals;
      }
    });

    setProblemLevelCount(levelStatistic);
  };

  // 获取状态统计数据
  const getStatusCount = async (params: {
    start: number;
    end: number;
    filters: any;
  }) => {
    const { filters = {} } = params;

    if (id) {
      filters.problem_id = id;
    }
    const res = await getAlertFaultStatusCount('itops_raw_event_status_sum', {
      ...params,
      ...generateMetricModelFilters(filters),
      instant: true
    });

    return transformOriginToStatistic(res, 'event_status');
  };

  const getCorrelatedEvents = async (params: any) => {
    const {
      current,
      pageSize = DEFAULT_PAGE_SIZE,
      start,
      end,
      sorter,
      filters = {}
    } = params;
    const sortAndFilters = generateDataViewFilters({
      sorter,
      // 不要直接修改filters，会影响其他地方的使用
      filters: {
        ...(id ? { problem_id: id } : {}),
        ...filters,
        ...(keyword ? { event_title: keyword } : {})
      },
      likeFields: ['event_title']
    });
    const defaultSort = [
      {
        field: 'event_status',
        direction: 'asc' as const
      },
      {
        field: 'event_level',
        direction: 'asc' as const
      },
      {
        field: 'event_occur_time',
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
    const res = await getDataViewList('__itops_raw_event', formattedParams);

    const { entries = [], total_count } = res;

    return {
      list: entries,
      total: total_count
    };
  };

  const { run, tableProps } = useAntdTable(getCorrelatedEvents, {
    manual: true,
    defaultPageSize: DEFAULT_PAGE_SIZE
  });

  const { level, status, onLevelChange, onStatusChange } = useStatistic();
  const { filterLevel, filterStatus, setFilterLevel, setFilterStatus } =
    useTableFilter({
      level,
      status
    });

  const refreshPage = () => {
    if (!timeRange.length) return;

    const params = {
      start: timeRange[0],
      end: timeRange[1],
      filters: {
        event_level: level,
        event_status: status
      }
    };

    getChartData(params).then((res) => {
      generateChart({
        ...res,
        title: intl.get('history_problem')
      });
    });
    getLevelCount(params);
    getStatusCount(params).then((res) => {
      const statusCount = [] as number[];

      Object.keys(Status_Maps).forEach((key) => {
        if (key === Status.Occurred) {
          statusCount[0] = res[key] ?? 0;
        }
        if (key === Status.Recovered) {
          statusCount[1] = res[key] ?? 0;
        }
        if (key === Status.Expired) {
          statusCount[2] = res[key] ?? 0;
        }
      });

      setProblemStatusCount(statusCount);
    });
    run({
      current: 1,
      ...params
    });
  };

  useRefreshPage({
    refreshInterval,
    refreshPage,
    timeRange,
    level,
    status
  });

  return (
    <div className={styles.faultPointsContainer}>
      {/* 顶部标签统计 */}
      <FaultHeader
        data={problemLevelCount.concat(problemStatusCount)}
        showLevel={[1, 2, 3, 4, 5]}
        onTimeChange={onTimeChange}
        onRefreshChange={onRefreshChange}
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
      />

      {/* 柱状图 */}
      <div className={styles.trendChartContainer}>
        <div ref={trendChartRef} className={styles.trendChart} />
      </div>

      {/* 操作栏 */}
      <div className={styles.operationBar}>
        <Space.Compact>
          <Space.Addon>
            <span style={{ whiteSpace: 'nowrap' }}>
              {intl.get('event_title')}
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

              if (id) {
                filters.problem_id = id;
              }
              if (filterLevel) {
                filters.event_level = filterLevel;
              }
              if (filterStatus) {
                filters.event_status = filterStatus;
              }
              if (keyword) {
                filters.event_title = keyword;
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
        {/* <Button type="primary">关闭关联事件</Button> */}
      </div>

      {/* 关联事件列表 */}
      <div className={styles.faultPointsList}>
        <EventsList
          timeRange={timeRange}
          {...tableProps}
          level={filterLevel}
          status={filterStatus}
          onChange={(pagination: any, filters: any, sorter: any) => {
            setFilterLevel(filters.event_level ?? []);
            setFilterStatus(filters.event_status ?? []);
            tableProps.onChange(pagination, filters, sorter);
          }}
        />
      </div>
    </div>
  );
};

export default CorrelatedEvents;
