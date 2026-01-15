import React, { useEffect, useMemo, useState } from 'react';
import * as echarts from 'echarts';
import { Button, Input, Space, Affix } from 'antd';
import { useAntdTable, useParams } from '@noya/max';
import intl from 'react-intl-universal';
import classnames from 'classnames';
import PointsList from './components/PointsList';
import styles from './index.module.less';
import ARIconfont5 from '@/components/ARIconfont5';
import FaultHeader from '@/organisms/FaultHeader';
import { ProblemLevel } from '@/constants/problemTypes';
import { PaginationParams, Status } from '@/constants/commonTypes';
import { getDataViewList } from '@/services/fault-analysis';
import { getAlertFaultStatusCount } from '@/services/alert-fault';
import { transformOriginToStatistic } from '@/utils/transform-origin-to-statistic';
import { DEFAULT_PAGE_SIZE, Status_Maps } from '@/constants/common';
import {
  generateGanttChartConfig,
  transformOriginToGanttData
} from '@/utils/generate-gantt-chart-config';
import { transformOriginToChart } from '@/utils/transform-origin-to-chart';
import { useRefreshPage } from '@/hooks/useRefreshPage';
import { useStatistic } from '@/hooks/useStatistic';
import {
  generateDataViewFilters,
  generateMetricModelFilters
} from '@/utils/generate-data-view-filters';
import ARIconfont from '@/components/ARIconfont';
import { useTableFilter } from '@/hooks/useTableFilter';

const FaultPoints: React.FC = () => {
  const chartRef = React.useRef<HTMLDivElement>(null);
  const ganttChart = React.useRef<echarts.ECharts | null>(null);
  const [timeRange, setTimeRange] = useState<number[]>([]);
  const [levelCount, setLevelCount] = useState<number[]>([0, 0, 0, 0]);
  const [statusCount, setStatusCount] = useState<Array<number>>([0, 0]);
  const [refreshInterval, setRefreshInterval] = useState(0);
  const [keyword, setKeyword] = useState('');
  const [isEmptyGantt, setIsEmptyGantt] = useState(true);
  const [collapsed, setCollapsed] = useState(true);
  const chartDataRef = React.useRef<any>([]);
  const scrollRef = React.useRef<HTMLDivElement>(null);
  const { id } = useParams();

  const onTimeChange = (value: number[]) => {
    setTimeRange(value);
  };

  const onRefreshChange = (value: number) => {
    setRefreshInterval(value);
  };

  const generateChart = (data: any) => {
    if (ganttChart.current) {
      const config = generateGanttChartConfig(
        intl.get('history_problem'),
        data
      );
      let newConfig = { ...config };

      if (collapsed) {
        newConfig = { ...config };
      } else {
        newConfig.dataZoom = undefined;
      }

      ganttChart.current.setOption(newConfig, true);
    }
  };

  // 初始化甘特图
  useEffect(() => {
    const chart = echarts.init(chartRef.current);

    ganttChart.current = chart;

    // 响应式调整
    const resize = () => chart.resize();

    window.addEventListener('resize', resize);
    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (entry.target === chartRef.current) {
          chart.resize();
        }
      }
    });

    if (chartRef.current) {
      resizeObserver.observe(chartRef.current);
    }

    return () => {
      window.removeEventListener('resize', resize);
      resizeObserver.disconnect();
      chart.dispose();
      ganttChart.current = null;
    };
  }, []);

  const getFaultList = async (params: any) => {
    const {
      current,
      pageSize = DEFAULT_PAGE_SIZE,
      start,
      end,
      filters = {}
    } = params;
    const sortAndFilters = generateDataViewFilters({
      // 不要直接修改filters，会影响其他地方的使用
      filters: {
        ...(id ? { problem_id: id } : {}),
        ...filters,
        ...(keyword ? { fault_description: keyword } : {})
      },
      likeFields: ['fault_description']
    });
    const formattedParams: PaginationParams = {
      start,
      end,
      offset: (current - 1) * pageSize,
      limit: pageSize,
      sort: [
        {
          field: 'fault_status',
          direction: 'asc'
        },
        {
          field: 'fault_level',
          direction: 'asc'
        },
        {
          field: 'fault_occur_time',
          direction: 'desc'
        }
      ],
      ...sortAndFilters
    };
    const res = await getDataViewList(
      '__itops_fault_point_object',
      formattedParams
    );

    const { entries = [], total_count } = res;

    return {
      list: entries,
      total: total_count
    };
  };

  // 获取图表数据
  const getChartData = async (params: {
    start: number;
    end: number;
    filters?: any;
    limit?: number;
  }) => {
    const { start, end, filters = {}, limit = 10000 } = params;

    if (id) {
      filters.problem_id = id;
    }
    const res = await getDataViewList('__itops_fault_point_object', {
      start,
      end,
      limit,
      ...generateDataViewFilters({ filters })
    });

    setIsEmptyGantt(!(res?.entries?.length > 0));

    const data = transformOriginToGanttData(res?.entries ?? []);

    chartDataRef.current = data;

    return data;
  };

  const getLevelCount = async (params: {
    start: number;
    end: number;
    filters?: any;
  }) => {
    const { filters = {} } = params;

    if (id) {
      filters.problem_id = id;
    }
    const res = await getAlertFaultStatusCount('itops_fault_point_level_sum', {
      ...params,
      ...generateMetricModelFilters(filters),
      instant: true
    });
    const { seriesData = [] } = transformOriginToChart(res, 'fault_level');
    const levelOriginCount = [0, 0, 0, 0];

    seriesData.forEach((item) => {
      if (item.key === String(ProblemLevel.Urgent)) {
        levelOriginCount[0] = item.totals;
      } else if (item.key === String(ProblemLevel.Severe)) {
        levelOriginCount[1] = item.totals;
      } else if (item.key === String(ProblemLevel.Important)) {
        levelOriginCount[2] = item.totals;
      } else if (item.key === String(ProblemLevel.Warning)) {
        levelOriginCount[3] = item.totals;
      }
    });

    return levelOriginCount;
  };

  // 获取状态统计数据
  const getStatusCount = async (params: {
    start: number;
    end: number;
    filters?: any;
  }) => {
    const { filters = {} } = params;

    if (id) {
      filters.problem_id = id;
    }
    const res = await getAlertFaultStatusCount('itops_fault_point_status_sum', {
      ...params,
      ...generateMetricModelFilters(filters),
      instant: true
    });

    return transformOriginToStatistic(res, 'fault_status');
  };

  const { run, tableProps } = useAntdTable(getFaultList, {
    manual: true,
    defaultPageSize: DEFAULT_PAGE_SIZE
  });

  const { level, status, onLevelChange, onStatusChange } = useStatistic();
  const { filterLevel, setFilterLevel } = useTableFilter({
    level,
    status
  });

  const refreshPage = () => {
    if (!timeRange.length) return;

    const params = {
      start: timeRange[0],
      end: timeRange[1],
      filters: {
        fault_level: level,
        fault_status: status
      }
    };

    getChartData(params).then((data) => {
      generateChart(data);
    });
    getLevelCount(params).then((res) => {
      setLevelCount(res);
    });
    getStatusCount(params).then((res) => {
      const statusCount = [] as number[];

      Object.keys(Status_Maps).map((key) => {
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

      setStatusCount(statusCount);
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

  useEffect(() => {
    generateChart(chartDataRef.current);
  }, [collapsed]);

  const ganttHeight = useMemo(() => {
    const total = tableProps?.pagination?.total || 0;
    const barHeight = 25;
    const barMargin = 2;
    const allDataHeight = total * barHeight + (total - 1) * barMargin;

    return collapsed
      ? '234px'
      : `${allDataHeight > 234 ? allDataHeight : 234}px`;
  }, [collapsed, tableProps?.pagination?.total]);

  return (
    <div className={styles.faultPointsContainer} ref={scrollRef}>
      {/* 顶部标签统计 */}
      <FaultHeader
        data={levelCount.concat(statusCount)}
        onTimeChange={onTimeChange}
        onLevelChange={onLevelChange}
        onStatusChange={onStatusChange}
        onRefreshChange={onRefreshChange}
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

      {/* 甘特图 */}
      <div className={styles.ganttContainer}>
        {/* <div
          ref={chartRef}
          className={styles.ganttChart}
          style={{
            height: '234px'
          }}
        ></div> */}
        <div className="flex flex-col">
          <div
            ref={chartRef}
            className={styles.ganttChart}
            style={{
              height: ganttHeight
            }}
          />
          <div
            className={`flex justify-center items-center ${
              tableProps?.pagination?.total > 6 ? '' : 'hidden'
            }`}
          >
            <Affix offsetBottom={0} target={() => scrollRef.current || window}>
              <span
                className="cursor-pointer flex justify-center items-center w-[42px] h-[13px] rounded-t-[5px] bg-[#e5e5e5] border border-solid border-[#e5e5e5]"
                onClick={() => {
                  setCollapsed(!collapsed);
                }}
              >
                <ARIconfont
                  className="text-[8px]"
                  type={collapsed ? 'icon-double-down' : 'icon-double-up'}
                />
              </span>
            </Affix>
          </div>
        </div>
        <div
          className={classnames(styles.emptyGantt, {
            [styles.hidden]: !isEmptyGantt
          })}
        >
          <ARIconfont type="icon-a-bianzu6" />
          <span>{intl.get('NoData')}</span>
        </div>
      </div>

      {/* 操作栏 */}
      <div className={styles.operationBar}>
        <Space.Compact>
          <Space.Addon>
            <span style={{ whiteSpace: 'nowrap' }}>
              {intl.get('fault_point_description')}
            </span>
          </Space.Addon>
          <Input
            placeholder={intl.get('Input')}
            value={keyword}
            onChange={(e) => {
              setKeyword(e.target.value);
            }}
            onPressEnter={() => {
              const filters: Record<string, any> = {
                problem_id: id
              };

              if (filterLevel.length > 0) {
                filters.fault_level = filterLevel;
              }

              if (status.length > 0) {
                filters.fault_status = status;
              }
              if (keyword) {
                filters.fault_description = keyword;
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

      {/* 故障点列表 */}
      <div className={styles.faultPointsList}>
        <PointsList
          {...tableProps}
          level={filterLevel}
          onChange={(pagination: any, filters: any, sorter: any) => {
            setFilterLevel(filters.fault_level ?? []);
            tableProps.onChange(pagination, filters, sorter);
          }}
        />
      </div>
    </div>
  );
};

export default FaultPoints;
