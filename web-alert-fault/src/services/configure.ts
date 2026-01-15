import { ITOpsConfigure } from '@/pages/Configure/types';
import request from '@/utils/axios-http';

const baseURL = '/api/itops_alert_manager/v1/config';

/**
 * 获取配置
 */
export async function getConfigure(): Promise<ITOpsConfigure> {
  return await request.axiosGet(baseURL);
}

/**
 * 更新配置
 */
export async function updateConfigure(isCreated: boolean, data: any) {
  return isCreated
    ? await request.axiosPost(baseURL, data)
    : await request.axiosPut(baseURL, data);
}

/**
 * 获取业务知识网络列表
 */
export async function getKnowledgeNetworks() {
  const params = {
    limit: -1
  };

  return await request.axiosGet(
    '/api/ontology-manager/v1/knowledge-networks',
    {
      params
    },
    {
      headers: {
        'x-business-domain': 'bd_public'
      }
    }
  );
}
