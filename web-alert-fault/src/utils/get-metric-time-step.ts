import { IntervalType } from 'Constants/commonTypes';

const s = 1000;
const m = 60 * s;
const h = 60 * m;
const d = 24 * h;
// const w = 7 * d;
const M = 31 * d;
// const q = 3 * M;
const y = 12 * M;

// 系统步长集
export const fixdStepsList: string[] = [
  '15s',
  '30s',
  '1m',
  '2m',
  '5m',
  '10m',
  '15m',
  '20m',
  '30m',
  '1h',
  '2h',
  '3h',
  '6h',
  '12h',
  '1d'
];

export const calendarStepsList: string[] = ['minute', 'hour', 'day', 'week', 'month', 'quarter', 'year'];

/**
 * @description 计算指标模型时间步长
 * @param stepTimes 步长对应的毫秒数
 * @param type 步长类型
 * @returns 修正后的步长
 */
export const getMetricTimeStep = (timeInterval: number): { type: IntervalType; step: string; format: string } => {
  if (timeInterval <= h) return { type: IntervalType.Fixed, step: '15s', format: 'HH:mm:ss' };
  if (timeInterval <= 6 * h) return { type: IntervalType.Fixed, step: '1m', format: 'HH:mm' };
  if (timeInterval <= 24 * h) return { type: IntervalType.Fixed, step: '1h', format: 'HH:mm' };
  if (timeInterval <= 7 * d) return { type: IntervalType.Fixed, step: '6h', format: 'MM-DD HH:mm' };
  if (timeInterval <= M) return { type: IntervalType.Fixed, step: '1d', format: 'MM-DD' };
  if (timeInterval <= 3 * M) return { type: IntervalType.Fixed, step: '1d', format: 'MM-DD' };
  if (timeInterval <= 3 * M) return { type: IntervalType.Fixed, step: '1d', format: 'MM-DD' };
  if (timeInterval <= 6 * M) return { type: IntervalType.Fixed, step: '1d', format: 'MM-DD' };
  if (timeInterval <= y) return { type: IntervalType.Fixed, step: '1y', format: 'MM-DD' };
  if (timeInterval <= 5 * y) return { type: IntervalType.Fixed, step: '1y', format: 'YYYY-MM' };

  return { type: IntervalType.Fixed, step: '1y', format: 'YYYY-MM' };
};

export const getTimeStepType = (stepString: string): IntervalType => {
  if (fixdStepsList.includes(stepString)) return IntervalType.Fixed;

  if (calendarStepsList.includes(stepString)) return IntervalType.Calendar;

  return IntervalType.Fixed;
};

export const metricTimeRangeProcessing = (end: number): number => (String(end).slice(-3) === '999' ? end + 1 : end);


export default getMetricTimeStep;
