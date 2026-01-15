export const generateBarChartConfig = (data: {
  xAxisData: string[];
  seriesData: any[];
  title?: string;
}) => {
  const { xAxisData, seriesData, title } = data;
  const trendOption = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e8e8e8',
      borderWidth: 1,
      textStyle: {
        color: '#333'
      }
    },
    ...(title
      ? {
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
          }
        }
      : {}),
    grid: {
      left: 88,
      right: 16,
      top: 60,
      bottom: 5
    },
    xAxis: [
      {
        type: 'category',
        data: xAxisData,
        axisLine: {
          lineStyle: {
            color: '#e8e8e8'
          }
        },
        axisLabel: {
          color: '#666',
          fontSize: 12
        },
        axisTick: {
          alignWithLabel: true,
          lineStyle: {
            color: '#e8e8e8'
          }
        }
      }
    ],
    yAxis: [
      {
        type: 'value',
        name: '问题数',
        nameLocation: 'center',
        nameGap: 50,
        minInterval: 1,
        nameTextStyle: {
          color: 'rgba(0, 0, 0, .45)'
        },
        axisLine: {
          lineStyle: {
            color: '#e8e8e8'
          }
        },
        axisLabel: {
          color: '#666',
          fontSize: 12
        },
        splitLine: {
          lineStyle: {
            color: '#f0f0f0'
          }
        }
      }
    ],
    series: seriesData
  };

  return trendOption;
};
