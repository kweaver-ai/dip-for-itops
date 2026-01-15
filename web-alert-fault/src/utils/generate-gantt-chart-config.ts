import dayjs from 'dayjs';
import * as echarts from 'echarts';
import formatTimeDifference from './format_date_duration';
import occurredIcon from '@/assets/images/occurred.svg';
import expiredIcon from '@/assets/images/expired.svg';
import recoveredIcon from '@/assets/images/recovered.svg';
import { FaultPoint } from '@/pages/FaultAnalysis/FaultPoints/types';
import { Level_Maps } from '@/constants/common';
import { Status } from '@/constants/commonTypes';
import { ProblemLevel } from '@/constants/problemTypes';

const Status_Icon_Maps = {
  [Status.Occurred]: occurredIcon,
  [Status.Recovered]: recoveredIcon,
  [Status.Expired]: expiredIcon
};

export const generateGanttChartConfig = (
  title: string,
  dataConfig: { minTime: number; maxTime: number; data: any[] }
) => {
  const { data = [] } = dataConfig;
  const ganttData = data.map((item, index) => {
    return {
      name: item.name, // 使用name作为系列名称
      value: [
        index,
        item.startTime,
        item.startTime + item.duration,
        item.duration,
        item.text,
        item.status
      ],
      startTime: item.startTime,
      endTime: item.endTime,
      duration: item.duration,
      text: item.text,
      itemStyle: { color: item.color, borderRadius: '50%' }
    };
  });
  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'item',
      formatter(params: any) {
        return `
          <div style="font-size: 12px;width: 300px;overflow: hidden">
            <div>开始时间: ${echarts.format.encodeHTML(
              dayjs(params.data.startTime).format('YYYY-MM-DD HH:mm:ss')
            )}</div>
            <div>结束时间: ${echarts.format.encodeHTML(
              dayjs(params.data.endTime).format('YYYY-MM-DD HH:mm:ss')
            )}</div>
            <div>持续时间: ${echarts.format.encodeHTML(
              formatTimeDifference(params.data.duration)
            )}</div>
            <div style="height: 42px;line-height: 21px;white-space: normal;">故障点描述：${echarts.format.encodeHTML(
              params.data.text
            )}</div>
          </div>
          `;
      }
    },
    dataZoom: [
      /*
       * {
       *   type: 'slider',
       *   xAxisIndex: 0,
       *   filterMode: 'weakFilter',
       *   height: 20,
       *   bottom: 0,
       *   start: 0,
       *   end: 26,
       *   handleIcon:
       *     'path://M10.7,11.9H9.3c-4.9,0.3-8.8,4.4-8.8,9.4c0,5,3.9,9.1,8.8,9.4h1.3c4.9-0.3,8.8-4.4,8.8-9.4C19.5,16.3,15.6,12.2,10.7,11.9z M13.3,24.4H6.7V23h6.6V24.4z M13.3,19.6H6.7v-1.4h6.6V19.6z',
       *   handleSize: '80%',
       *   showDetail: false
       * },
       * {
       *   type: 'inside',
       *   id: 'insideX',
       *   xAxisIndex: 0,
       *   filterMode: 'weakFilter',
       *   start: 0,
       *   end: 26,
       *   zoomOnMouseWheel: false,
       *   moveOnMouseMove: true
       * },
       */
      {
        type: 'slider',
        yAxisIndex: 0,
        zoomLock: true,
        width: 10,
        right: 0,
        minValueSpan: 6,
        maxValueSpan: 6,
        handleSize: 0,
        showDetail: false
      },
      {
        type: 'inside',
        id: 'insideY',
        yAxisIndex: 0,
        minValueSpan: 6,
        maxValueSpan: 6,
        zoomOnMouseWheel: false,
        moveOnMouseMove: true,
        moveOnMouseWheel: true
      }
    ],
    title: {
      show: true,
      text: title,
      left: 0,
      top: 0,
      padding: [8, 0, 8, 16],
      textStyle: {
        color: 'rgba(0,0,0,0.85)',
        fontSize: 14
      }
    },
    grid: {
      left: 88,
      right: 16,
      top: 60,
      bottom: 5
    },
    xAxis: {
      type: 'time',
      position: 'top',
      axisLabel: {
        formatter: {
          year: '{yyyy}',
          month: '{MM}',
          day: '{MM}-{dd}',
          hour: '{HH}:{mm}',
          minute: '{HH}:{mm}',
          second: '{HH}:{mm}:{ss}',
          millisecond: '{hh}:{mm}:{ss} {SSS}',
          // @ts-ignore
          none: '{yyyy}-{MM}-{dd} {hh}:{mm}:{ss} {SSS}'
        },
        fontSize: 12
      },
      axisLine: {
        lineStyle: {
          color: '#d9d9d9'
        }
      },
      splitLine: {
        show: false
      }
    },
    yAxis: {
      type: 'category',
      data: ganttData.map((item) => item.name),
      inverse: true,
      axisLabel: {
        fontSize: 12,
        width: 180,
        overflow: 'truncate'
      },
      axisLine: {
        show: false
      },
      splitLine: {
        show: true,
        lineStyle: {
          type: 'dashed',
          color: 'rgb(0,0,0)',
          opacity: 0.15
        }
      }
    },
    // @ts-ignore
    series: [
      {
        type: 'custom',
        data: ganttData,
        encode: {
          x: [1, 2],
          y: 0
        },
        renderItem: (params, api) => {
          const categoryIndex = api.value(0);
          const text = api.value(4);
          // 字符串数字变成了数值类型
          const status = api.value(5);
          const start = api.coord([api.value(1), categoryIndex]);
          const end = api.coord([api.value(2), categoryIndex]);

          /*
           * @ts-ignore
           * const height = api.size([0, 1])[1] * 0.6;
           */
          const height = 15;
          const barLen = end[0] - start[0];
          const x = start[0];
          const y = start[1] - height / 2;
          const rectShape = echarts.graphic.clipRectByRect(
            {
              x,
              y,
              width: barLen,
              height
            },
            {
              // @ts-ignore
              x: params.coordSys.x,
              // @ts-ignore
              y: params.coordSys.y,
              // @ts-ignore
              width: params.coordSys.width,
              // @ts-ignore
              height: params.coordSys.height
            }
          );

          if (rectShape) {
            const rectWidth =
              rectShape.width < height
                ? rectShape.width * 100 + height + 2
                : rectShape.width;

            return {
              type: 'group' as const,
              children: [
                {
                  type: 'circle' as const,
                  shape: {
                    cx: start[0] + height / 2,
                    cy: start[1],
                    r: height / 2
                  },
                  style: api.style()
                },
                {
                  type: 'rect' as const,
                  shape: {
                    ...rectShape,
                    width: rectWidth - height,
                    x: rectShape.x + height / 2
                  },
                  style: api.style()
                },
                {
                  type: 'circle' as const,
                  shape: {
                    cx: rectShape.x - height / 2 + rectWidth,
                    cy: end[1],
                    r: height / 2
                  },
                  style: api.style()
                },
                {
                  type: 'image' as const,
                  style: {
                    image: Status_Icon_Maps[status as Status],
                    x: start[0] + 2,
                    y: start[1] - height / 2 + 1,
                    width: height - 2,
                    height: height - 2
                  }
                }
              ]
            };
          }
        }
      }
    ]
  };

  return option;
};

export const transformOriginToGanttData = (originData: FaultPoint[]) => {
  const targetData = originData.map((item: any) => {
    const start = dayjs(item.fault_occur_time).valueOf();
    const duration =
      item.fault_duration_time > 0
        ? item.fault_duration_time
        : Math.floor((Date.now() - start) / 1000);

    return {
      id: item.fault_id,
      name: item.entity_object_name,
      startTime: start,
      endTime: dayjs(item.fault_occur_time).add(duration, 's').valueOf(),
      duration,
      color: Level_Maps[item.fault_level as ProblemLevel]?.color,
      text: item.fault_description,
      status: item.fault_status
    };
  });

  return {
    data: targetData
  };
};
