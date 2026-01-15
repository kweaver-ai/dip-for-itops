/**
 * AISHU Data Agent API 配置
 */
export const DATA_AGENT_CONFIG = {
  /** 服务端基础地址,需包含 /api/agent-app/v1 前缀; 开发环境使用 Vite 代理的本地前缀可避免 CORS */
  baseUrl: '/data-agent',
  /** Agent ID */
  agentId: '01KAZKS30H0X0D8Z8K25VKSJ98',
  /** Token (不带 Bearer 前缀) */
  token: 'ry_at_SL8UpKSQ6hcQhVylz_Zl6TaUY4hLXJXkzWiKddCJ6rQ.UFX6wqEPTJHGDar_xo1HFKQiH6l-y3apGHbSP50Go0o',
  /** 业务域 */
  businessDomain: 'bd_public'
};

/**
 * 使用说明:
 * - baseUrl 请替换为你的 Data Agent 服务网关地址
 * - 需要保证 bearerToken 中包含 Bearer 前缀
 */
