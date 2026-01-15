import dayjs from 'dayjs';
import formatTimeDifference from './format_date_duration';
import { ProblemStatus } from '@/constants/problemTypes';

const getProblemDurationTime = (problem: any) => {
  const { problem_status, problem_occur_time, problem_close_time } = problem;

  if (problem_status === ProblemStatus.Invalid) {
    return '--';
  }

  if (problem_status === ProblemStatus.Open) {
    const durationTime =
      dayjs().valueOf() - dayjs(problem_occur_time).valueOf();

    return formatTimeDifference(durationTime);
  }

  if (problem_status === ProblemStatus.Closed) {
    const durationTime =
      dayjs(problem_close_time).valueOf() - dayjs(problem_occur_time).valueOf();

    return formatTimeDifference(durationTime);
  }
};

export default getProblemDurationTime;
