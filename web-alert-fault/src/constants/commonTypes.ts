// 排序方向
export type SortDirection = 'asc' | 'desc';

// 过滤参数
interface FilterParams {
  operation: 'and' | 'or' | '==' | 'in' | 'match_phrase' | 'like';
  field: string;
  value: any;
  value_from: 'const';
}

// 分页查询参数
export interface PaginationParams {
  start: number;
  end: number;
  offset: number;
  limit: number;
  filters?:
    | FilterParams
    | {
        operation: 'and' | 'or';
        sub_conditions: FilterParams[];
      };
  sort?: Array<{
    field: string;
    direction: SortDirection;
  }>;
}

// 指标模型查询参数
export interface MetricModelQueryParams {
  start: number;
  end: number;
  instant: boolean;
  step: string;
  filters?: Array<{
    name: string;
    value: any;
    operation: 'in' | '=' | '!=';
  }>;
}

/**
 * 失效点状态颜色
 * 失效：expired    灰#BFBFBF
 * 发生 occurred    红#FF4D4F
 * 恢复 recovered   橙#52C41B
 */
export enum Status {
  // Unknown = '0',
  Occurred = '1',
  Recovered = '2',
  Expired = '3'
}

/**
 * 等级等级对应颜色：
 * 未知  0    灰#BFBFBF  默认值
 * 紧急  1    红#FF4D4F
 * 严重  2    橙#F8AD14
 * 重要  3    黄#FFD500
 * 告警  4    蓝#1890FF
 * 正常  5    绿#52c41b
 */
export enum Level {
  Unknown = 0,
  Urgent = 1,
  Severe = 2,
  Important = 3,
  Warning = 4,
  Normal = 5
}

/**
 * 影响等级等级对应颜色：
 * 未知  0    灰#BFBFBF  默认值
 * 紧急  1    红#FF4D4F
 * 严重  2    橙#F8AD14
 * 重要  3    黄#FFD500
 * 告警  4    蓝#1890FF
 * 正常  5    绿#52c41b
 */
export enum ImpactLevel {
  Unknown = 0,
  Urgent = 1,
  Severe = 2,
  Important = 3,
  Warning = 4,
  Normal = 5
}

/**
 * 问题等级对应颜色：
 * 未知  0    灰#BFBFBF  默认值
 * 紧急  1    红#FF4D4F
 * 严重  2    橙#F8AD14
 * 重要  3    黄#FFD500
 * 告警  4    蓝#1890FF
 * 信息  5    绿#52c41b
 */
export enum ProblemLevel {
  Unknown = 0,
  Urgent = 1,
  Severe = 2,
  Important = 3,
  Warning = 4,
  Info = 5
}

// 时间间隔类型
export enum IntervalType {
  Fixed = 'fixed', // 固定步长
  Calendar = 'calendar' // 日历步长
}
