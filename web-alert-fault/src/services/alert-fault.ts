import request from 'Utils/axios-http';
import { MetricModelQueryParams } from '@/constants/commonTypes';
import getMetricTimeStep, {
  metricTimeRangeProcessing
} from '@/utils/get-metric-time-step';

const prefix = process.env.NODE_ENV === 'development' ? '' : '';

/**
 * 问题概览-列表
 */
export const getAlertFaultList = async (params: any) => {
  return await request.axiosPost(
    `${prefix}/api/itops_alert_manager/v1/problem`,
    params
  );
};

/**
 * 问题概览-关闭
 */
export const closeAlertFault = async (id: string) => {
  return await request.axiosPut(
    `${prefix}/api/itops_alert_manager/v1/problem/${id}/close`
  );
};

/**
 * 问题概览-状态统计
 */
export const getAlertFaultStatusCount = async (
  metricId: string,
  params: Pick<MetricModelQueryParams, 'start' | 'end' | 'filters'> & {
    instant?: boolean;
  }
) => {
  const { start, end, instant = false } = params;
  const duration = end - start;
  const timeStep = getMetricTimeStep(duration);
  const newEnd = metricTimeRangeProcessing(end);
  const body: Record<string, any> = {
    ...params
  };

  if (instant || timeStep.step === '1y') {
    body.time = newEnd;
    body.look_back_delta = `${Math.floor(newEnd - start)}ms`;
    body.instant = true;
    body.method = 'GET';
    body.start = undefined;
    body.end = undefined;
  } else {
    body.start = start;
    body.end = newEnd;
    body.instant = false;
    body.step = timeStep.step;
  }

  return await request.axiosPost(
    `${prefix}/api/mdl-uniquery/v1/metric-models/${metricId}`,
    body,
    {
      headers: {
        'X-HTTP-Method-Override': 'GET'
      }
    }
  );
};
