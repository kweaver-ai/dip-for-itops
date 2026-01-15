// @ts-ignore
import { groupBy } from '@noya/max';

export const transformOriginToStatistic = (
  originData: any,
  keyStatus: string
) => {
  if (originData.datas && originData.datas.length) {
    const seriesGroup = groupBy(originData.datas, (item: any) =>
      JSON.stringify(item.labels)
    );

    const seriesData = Object.keys(seriesGroup).reduce((prev, key) => {
      const realKey = JSON.parse(key)?.[keyStatus];

      prev[realKey] = seriesGroup[key][0].values.reduce(
        (acc: number, cur: number) => acc + cur,
        0
      );

      return prev;
    }, {} as Record<string, number>);

    return seriesData;
  }

  return {};
};
