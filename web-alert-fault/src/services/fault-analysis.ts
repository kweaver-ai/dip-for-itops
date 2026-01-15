import cookie from 'js-cookie';
import request from 'Utils/axios-http';
import { MetricModelQueryParams } from '@/constants/commonTypes';
import getMetricTimeStep from '@/utils/get-metric-time-step';

const prefix = process.env.NODE_ENV === 'development' ? '' : '';

/**
 * 关联事件-列表查询
 * 故障点-列表查询
 */
export const getDataViewList = async (id: string, params: any) => {
  return await request.axiosPost(
    `${prefix}/api/mdl-uniquery/v1/data-views/${id}`,
    {
      ...params,
      need_total: true
    },
    {
      headers: {
        'X-HTTP-Method-Override': 'GET'
      }
    }
  );
};

// 设置根因节点
export const setRootCause = async (
  problemId: string,
  rootCauseObjectId: string,
  rootCauseFaultId: string
) => {
  return await request.axiosPut(
    `${prefix}/api/itops_alert_manager/v1/problem/${problemId}/root_cause`,
    {
      root_cause_object_id: rootCauseObjectId,
      root_cause_fault_id: Number(rootCauseFaultId)
    }
  );
};

// 获取对象详情
export const getObjectDetail = async (
  kn_id: string,
  ot_id: string,
  object_id: string
) => {
  const isfToken = cookie.get('client.oauth2_token');

  return await request.axiosPost(
    `${prefix}/api/ontology-query/v1/knowledge-networks/${kn_id}/object-types/${ot_id}?include_type_info=true`,
    {
      limit: 1,
      need_total: true,
      condition: {
        field: 's_id',
        value: object_id,
        operation: '=='
      }
    },
    {
      headers: {
        'X-HTTP-Method-Override': 'GET',
        Authorization: `Bearer ${isfToken}`
      }
    }
  );
};

// 获取对象详情
export const getDirectionSubgraph = async (
  kn_id: string,
  ot_id: string,
  object_id: string,
  direction: 'forward' | 'backward'
) => {
  const isfToken = cookie.get('client.oauth2_token');

  return await request.axiosPost(
    `${prefix}/api/ontology-query/v1/knowledge-networks/${kn_id}/subgraph`,
    {
      source_object_type_id: ot_id,
      direction,
      path_length: 1,
      limit: 100,
      condition: {
        field: 's_id',
        value: object_id,
        operation: '=='
      }
    },
    {
      headers: {
        'X-HTTP-Method-Override': 'GET',
        Authorization: `Bearer ${isfToken}`
      }
    }
  );
};

// 获取对象详情
export const getSubgraph = async (
  kn_id: string,
  ot_id: string,
  object_id: string
) => {
  const [forwardRes, backwardRes] = await Promise.all([
    getDirectionSubgraph(kn_id, ot_id, object_id, 'forward'),
    getDirectionSubgraph(kn_id, ot_id, object_id, 'backward')
  ]);

  // 前向查询
  const { objects: forwardObjects = {}, relation_paths: forwardPaths = [] } =
    forwardRes || {};

  // 后向查询
  const { objects: backwardObjects = {}, relation_paths: backwardPaths = [] } =
    backwardRes || {};

  return {
    objects: { ...forwardObjects, ...backwardObjects },
    relation_paths: [
      ...forwardPaths,
      ...backwardPaths.map((item) => ({
        relations: item.relations.map((i: any) => ({
          ...i,
          target_object_id: i.source_object_id,
          source_object_id: i.target_object_id
        }))
      }))
    ]
  };
};

/**
 * 溯源-事件数统计
 */
export const getProblemEventCount = async (
  problemId: number,
  params: Pick<MetricModelQueryParams, 'start' | 'end'> & { instant?: boolean }
) => {
  const { start, end, instant = false } = params;
  const duration = end - start;
  const timeStep = getMetricTimeStep(duration);

  return await request.axiosPost(
    `${prefix}/api/mdl-uniquery/v1/metric-models/itops_raw_event_level_sum`,
    {
      ...params,
      filters: [
        {
          name: 'problem_id',
          value: problemId,
          operation: '='
        }
      ],
      instant,
      step: timeStep.step
    },
    {
      headers: {
        'X-HTTP-Method-Override': 'GET'
      }
    }
  );
};
