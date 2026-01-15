/* eslint-disable react-hooks/exhaustive-deps */
import React, { useEffect, useState, useRef } from 'react';
import { Flex, Slider, Tooltip } from 'antd';
import * as echarts from 'echarts';
import dayjs from 'dayjs';
import intl from 'react-intl-universal';
import { useParams } from '@noya/max';
import ARIconfont5 from 'Components/ARIconfont5';
import styles from './index.module.less';
import { getProblemEventCount } from '@/services/fault-analysis';

type TimeGraphProps = {
  value: number;
  onPlay: () => void;
  onPause: () => void;
  onSliderChange: (value: number) => void;
  startTime: number;
  endTime: number;
};

function TimeGraph(props: TimeGraphProps) {
  const { value, onPlay, onPause, onSliderChange, startTime, endTime } = props;
  const [tempValue, setTempValue] = useState(value);
  const urlParams = useParams();
  const trendChartInstance = useRef<any>(null);

  const durationObj = dayjs.duration(endTime - startTime);
  const formatDate = durationObj.days() > 0 ? 'DD HH:mm:ss' : 'HH:mm:ss';

  // 获取图表数据
  const getEnventChartData = async (params: { start: number; end: number }) => {
    const problemId = Number(urlParams.id);
    const { datas = [] } = await getProblemEventCount(problemId, params);

    if (!datas.length) {
      return [];
    }
    const xAxisData = datas[0].times || [];
    const showData: [number, number][] = [];

    xAxisData.forEach((item: number, index: number) => {
      const v = datas.reduce((pre: number, cur: any) => {
        if (cur.values) {
          return pre + (cur.values[index] || 0);
        }

        return pre;
      }, 0);

      showData.push([item, v]);
    });

    return showData;
  };

  const initChart = async () => {
    const data = await getEnventChartData({
      start: startTime,
      end: endTime
    });

    const chart = echarts.init(
      document.getElementById('time-chart') as HTMLElement
    );

    trendChartInstance.current = chart;

    const maxValue = Math.max(...data.map((item) => item[1]));

    chart.setOption({
      grid: {
        left: 34,
        right: 34,
        top: 0,
        bottom: 0,
        containLabel: true
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow'
        }
      },
      xAxis: {
        type: 'time',
        show: true,
        position: 'top',
        axisLabel: {
          interval: 'auto',
          formatter: (value?: number): string =>
            dayjs(value).format(formatDate),
          margin: 2
        },
        splitLine: {
          show: false
        },
        axisTick: {
          show: false
        }
      },
      yAxis: {
        type: 'value',
        show: false,
        splitLine: {
          show: false
        }
      },
      series: [
        {
          name: intl.get('event_count'),
          data,
          type: 'bar',
          barCategoryGap: '2px',
          itemStyle: {
            color: '#1890FF'
          }
        },
        {
          name: 'background',
          data: data.map((item) => [item[0], maxValue]),
          type: 'bar',
          barCategoryGap: '2px',
          barGap: '-100%',
          itemStyle: {
            color: '#e8e8e8'
          },
          z: 0,
          tooltip: {
            show: false
          }
        }
      ]
    });
  };

  useEffect(() => {
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

  useEffect(() => {
    if (!startTime || !endTime) {
      return;
    }

    initChart();
  }, [startTime, endTime]);

  useEffect(() => {
    setTempValue(value);
  }, [value]);

  return (
    <Flex className={styles['time-content']} vertical>
      <Flex gap={2} align="center" className={styles['time-control']}>
        <div className={styles['icon-wrapper']}>
          <Tooltip title={intl.get('play')}>
            <ARIconfont5 type="icon-yunhangzhong1" onClick={onPlay} />
          </Tooltip>
        </div>
        <div className={styles['icon-wrapper']}>
          <Tooltip title={intl.get('pause')}>
            <ARIconfont5 type="icon-zanting2" onClick={onPause} />
          </Tooltip>
        </div>
        <Slider
          className={styles.slider}
          onChange={setTempValue}
          onChangeComplete={onSliderChange}
          value={tempValue}
          min={startTime}
          max={endTime}
          tooltip={{
            formatter: (value?: number): string =>
              value ? dayjs(value).format(formatDate) : ''
          }}
          styles={{
            track: { backgroundColor: '#126EE3' },
            rail: { color: 'rgba(0,0,0,0.15)' },
            handle: {
              color: '#126EE3',
              backgroundColor: '#126EE3',
              width: '4px',
              height: '6px'
            }
          }}
        />
      </Flex>
      <div id="time-chart" className={styles['time-chart']}></div>
    </Flex>
  );
}

export default TimeGraph;
