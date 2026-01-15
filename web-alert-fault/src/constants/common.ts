import intl from 'react-intl-universal';
import { ProblemLevel } from 'Constants/problemTypes';
import { ImpactLevel, Status } from 'Constants/commonTypes';

// 影响等级, 实体对象、故障点
export const Impact_Level_Maps = {
  [ImpactLevel.Unknown]: {
    name: 'Unknown',
    color: '#BFBFBF'
  },
  [ImpactLevel.Urgent]: {
    name: 'Urgent',
    color: '#FF4D4F'
  },
  [ImpactLevel.Severe]: {
    name: 'Severe',
    color: '#FF9900'
  },
  [ImpactLevel.Important]: {
    name: 'Important',
    color: '#FFCC00'
  },
  [ImpactLevel.Warning]: {
    name: 'Warning',
    color: '#1890FF'
  },
  [ImpactLevel.Normal]: {
    name: 'Normal',
    color: '#52c41b'
  }
};

// 问题等级
export const Level_Maps = {
  [ProblemLevel.Unknown]: {
    name: 'Unknown',
    color: '#BFBFBF'
  },
  [ProblemLevel.Urgent]: {
    name: 'Urgent',
    color: '#FF4D4F'
  },
  [ProblemLevel.Severe]: {
    name: 'Severe',
    color: '#FF9900'
  },
  [ProblemLevel.Important]: {
    name: 'Important',
    color: '#FFCC00'
  },
  [ProblemLevel.Warning]: {
    name: 'Warning',
    color: '#1890FF'
  },
  [ProblemLevel.Info]: {
    name: 'Info',
    color: '#52c41b'
  }
};

// 失效点状态映射表
export const Status_Maps = {
  [Status.Occurred]: {
    name: 'Occurred',
    color: '#FF4D4F'
  },
  [Status.Recovered]: {
    name: 'Recovered',
    color: '#52c41b'
  },
  [Status.Expired]: {
    name: 'Invalid',
    color: '#BFBFBF'
  }
};

// 公共分页配置
export const COMMON_PAGINATION = {
  // hideOnSinglePage: true,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => intl.get('TotalCount', { total }),
  onChange: undefined,
  changePageSize: undefined
};

// 默认每页数量
export const DEFAULT_PAGE_SIZE = 20;

// 默认统计状态
export const DEFAULT_STATISTIC_STATUS = [
  { text: 'Happened', value: Status.Occurred },
  { text: 'Recovered', value: Status.Recovered }
];
