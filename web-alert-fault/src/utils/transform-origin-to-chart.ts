import { Level_Maps } from '@/constants/common';
import { ProblemLevel } from '@/constants/problemTypes';
// @ts-ignore
import { groupBy } from '@noya/max';
import dayjs from 'dayjs';
import intl from 'react-intl-universal';

export const transformOriginToChart = (originData: any, keyLevel: string) => {
  if (originData.datas && originData.datas.length) {
    const seriesGroup = groupBy(originData.datas, (item: any) =>
      JSON.stringify(item.labels)
    );

    const xAxisData = Object.values<any>(seriesGroup)[0][0].times.map(
      (item: any) => dayjs(item).format('YYYY-MM-DD HH:mm:ss')
    );
    const seriesData = Object.keys(seriesGroup).map((key) => {
      const realKey = JSON.parse(key)[keyLevel];

      return {
        key: realKey,
        // @ts-ignore
        name: intl.get(Level_Maps[realKey as unknown as ProblemLevel]?.name),
        stack: 'event_num',
        type: 'bar',
        itemStyle: {
          // @ts-ignore
          color: Level_Maps[realKey as unknown as ProblemLevel]?.color
        },
        label: {
          show: false
        },
        data: seriesGroup[key][0].values,
        totals: seriesGroup[key][0].values.reduce(
          (acc: number, cur: number) => acc + cur,
          0
        )
      };
    });

    return {
      xAxisData,
      seriesData
    };
  }

  return {
    xAxisData: [],
    seriesData: []
  };
};
