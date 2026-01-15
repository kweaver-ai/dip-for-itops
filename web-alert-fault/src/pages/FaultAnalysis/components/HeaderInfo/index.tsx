import { useContext } from 'react';
import { Flex, Tag } from 'antd';
import intl from 'react-intl-universal';
import dayjs from 'dayjs';
import getProblemDurationTime from 'Utils/get_problem_duration_time';
import { faultAnalysisContext } from '../../index';
import styles from './index.module.less';
import { Level_Maps } from '@/constants/common';
import { ProblemLevel, ProblemStatus } from '@/constants/problemTypes';
import { StatusesText, statusStyle } from '@/pages/AlertFault/constants';

const dateFormat = 'YYYY-MM-DD HH:mm:ss';

function HeaderInfo() {
  const {
    problemData: {
      problem_id: problemId,
      problem_level: problemLevel = '0',
      problem_status: problemStatus = '0',
      problem_start_time: problemStartTime,
      problem_occur_time: problemOccurTime,
      problem_end_time: problemEndTime,
      problem_close_time: problemCloseTime
    }
  } = useContext(faultAnalysisContext);

  const failureLevel =
    Level_Maps[`${problemLevel}` as unknown as ProblemLevel] || {};

  return (
    <Flex
      vertical={false}
      justify="space-between"
      className={styles['layout-wrapper']}
    >
      <Flex vertical={false} align="center" gap={24}>
        <Flex vertical={false} align="center" gap={8}>
          <div className={styles['item-title']}>{intl.get('problem_id')}:</div>
          <div>{problemId}</div>
        </Flex>
        <Flex vertical={false} align="center" gap={8}>
          <div className={styles['item-title']}>
            {intl.get('problem')}
            {intl.get('level')}:
          </div>
          <div>
            <span
              className={styles['alert-level']}
              style={{ backgroundColor: failureLevel.color }}
            ></span>
            {intl.get(failureLevel.name)}
          </div>
        </Flex>
        <Flex vertical={false} align="center" gap={8}>
          <div className={styles['item-title']}>
            {intl.get('problem')}
            {intl.get('status')}:
          </div>
          <div>
            <Tag
              color={
                statusStyle[problemStatus as unknown as ProblemStatus].color
              }
              variant="filled"
            >
              {intl.get(
                StatusesText[problemStatus as unknown as ProblemStatus]
              )}
            </Tag>
          </div>
        </Flex>
        <Flex vertical={false} align="center" gap={8}>
          <div className={styles['item-title']}>{intl.get('occur_time')}:</div>
          <div>{dayjs(problemOccurTime).format(dateFormat)}</div>
        </Flex>
        <Flex vertical={false} align="center" gap={8}>
          <div className={styles['item-title']}>{intl.get('close_time')}:</div>
          <div>
            {problemStatus === ProblemStatus.Closed
              ? dayjs(problemCloseTime).format('YYYY-MM-DD HH:mm:ss')
              : '--'}
          </div>
        </Flex>
        <Flex vertical={false} align="center" gap={8}>
          <div className={styles['item-title']}>
            {intl.get('problem_duration')}:
          </div>
          <div>
            <Tag color="#1677ff" variant="filled">
              {getProblemDurationTime({
                problem_status: problemStatus,
                problem_occur_time: problemOccurTime,
                problem_close_time: problemCloseTime
              })}
            </Tag>
          </div>
        </Flex>
      </Flex>
    </Flex>
  );
}

export default HeaderInfo;
