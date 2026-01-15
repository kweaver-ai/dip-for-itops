import { ProblemLevel, ProblemStatus } from 'Constants/problemTypes';

export const LevelsText = {
  [ProblemLevel.Urgent]: 'Urgent',
  [ProblemLevel.Severe]: 'Severe',
  [ProblemLevel.Important]: 'Important',
  [ProblemLevel.Warning]: 'Warning'
};

export const levelColors = {
  [ProblemLevel.Urgent]: '#FF4D4F',
  [ProblemLevel.Severe]: '#FF9900',
  [ProblemLevel.Important]: '#FFCC00',
  [ProblemLevel.Warning]: '#1890FF'
};

export const StatusesText = {
  [ProblemStatus.Open]: 'Open',
  [ProblemStatus.Closed]: 'Closed',
  [ProblemStatus.Invalid]: 'Invalid'
};

export const statusStyle: Record<ProblemStatus, React.CSSProperties> = {
  [ProblemStatus.Open]: {
    color: '#FF4D4F',
    background: 'rgba(255,77,79,0.15)',
    border: '1px solid rgba(255,77,79,0.3)'
  },
  [ProblemStatus.Closed]: {
    color: '#07C393',
    background: 'rgba(7,195,147,0.15)',
    border: '1px solid rgba(7,195,147,0.3)'
  },
  [ProblemStatus.Invalid]: {
    color: '#000000a6',
    background: 'rgba(0,0,0,0.05)',
    border: '1px solid rgba(0,0,0,0.15)'
  }
};

export const TIME_FIELDS = [
  'problem_create_timestamp',
  'problem_close_time',
  'problem_update_time'
];
export const NOT_SHOW_FIELDS = [
  'rca_results',
  'rca_start_time',
  'rca_end_time',
  'rca_status'
];
export const EXPAND_FIELDS = [
  'problem_id',
  'problem_name',
  'problem_create_timestamp',
  'problem_update_time',
  'problem_description',
  'problem_close_type',
  'problem_close_notes',
  'problem_closed_by',
  'problem_close_time',
  'relation_fp_ids'
  // 'root_cause_object_id',
  // 'root_cause_fault_id'
];
