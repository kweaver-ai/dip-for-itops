import { ProblemLevel as EventLevel } from 'Constants/problemTypes';

export enum EventStatus {
  Occurred = '1',
  Recovered = '2'
}

// 定义事件对象
export interface CorrelatedEvent {
  event_id: string;
  event_provider_id: string;
  event_timestamp: string;
  event_title: string;
  event_content: string;
  event_occur_time: string;
  event_recovery_time: string;
  event_type: string;
  event_status: EventStatus;
  event_level: EventLevel;
  event_source: string;
  entity_object_name: string;
  entity_object_class: string;
  entity_object_ip: string;
  entity_object_port: string;
  entity_object_mac: string;
  raw_event_msg: string;
  problem_id: string;
  fault_id: string;
}

export interface CorrelatedEventProps {
  onDataChange?: (eventsNum: number) => void;
}
