// 问题状态
export enum ProblemStatus {
  Open = '0',
  Closed = '1',
  Invalid = '2',
}

// 问题等级
export enum ProblemLevel {
  Unknown = 0,
  Urgent = 1,
  Severe = 2,
  Important = 3,
  Warning = 4,
  Info = 5,
}

// 问题列表展示模式
export enum ProblemListMode {
  List = '列表模式',
  Detail = '详情模式',
}

// 问题详情
export interface Problem {
  problem_id: string;
  problem_name: string;
  problem_create_timestamp: string;
  problem_occur_time: string;
  problem_latest_time: string;
  problem_duration: number;
  problem_description: string;
  problem_status: ProblemStatus;
  problem_close_type: '1' /* 系统关闭 */ | '2' /* 手动关闭 */;
  problem_close_notes: string;
  problem_closed_by: string;
  problem_close_time: string;
  problem_level: ProblemLevel;
  affected_entity_ids: string[];
  relation_fp_ids: string[];
  relation_event_ids: string[];
  root_cause_entity_object_name: string;
  root_cause_fault_description: string;
}
