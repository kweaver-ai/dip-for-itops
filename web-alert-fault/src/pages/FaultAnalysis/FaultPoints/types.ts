import { Status } from "@/constants/commonTypes";
import { ProblemLevel } from "@/constants/problemTypes";

export interface FaultPoint {
  fault_id: string;
  // 故障点名称
  fault_name: string;
  fault_level: ProblemLevel;
  fault_status: Status;
  fault_description: string;
  relation_event_ids: string[];
  entity_object_name: string;
  entity_object_id: string;
  fault_duration_time: number;
  fault_occur_time: string;
  fault_recovery_time: string;
}
