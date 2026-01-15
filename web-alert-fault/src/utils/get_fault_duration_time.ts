import dayjs from 'dayjs';
import formatTimeDifference from './format_date_duration';
import { Status } from '@/constants/commonTypes';

const getFaultDurationTime = (problem: any) => {
  const { fault_status, fault_occur_time, fault_recovery_time } = problem;

  if (fault_status === Status.Expired) {
    return '--';
  }

  if (fault_status === Status.Occurred) {
    const durationTime =
      dayjs().valueOf() - dayjs(fault_occur_time).valueOf();

    return formatTimeDifference(durationTime);
  }

  if (fault_status === Status.Recovered) {
    const durationTime =
      dayjs(fault_recovery_time).valueOf() - dayjs(fault_occur_time).valueOf();

    return formatTimeDifference(durationTime);
  }
};

export default getFaultDurationTime;
