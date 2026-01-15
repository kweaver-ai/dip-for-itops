import { ChatKitBase, ChatKitBaseProps } from './ChatKitBase';
import {
  ApplicationContext,
  ChatMessage,
  ChatMessageType,
  EventStreamMessage,
  RoleType,
  OnboardingInfo,
} from '../types';

/**
 * ChatKitDataAgent 组件的属性接口
 */
export interface ChatKitDataAgentProps extends ChatKitBaseProps {
  /** AISHU Data Agent 的 Agent ID,用作路径参数 */
  agentId: string;

  /** 访问令牌,需要包含 Bearer 前缀 (已废弃，请使用 token 属性) */
  bearerToken?: string;

  /** 服务端基础地址,应包含 /api/agent-app/v1 前缀 */
  baseUrl?: string;

  /** 是否开启增量流式返回,默认 true */
  enableIncrementalStream?: boolean;

  /** 智能体所属的业务域,用于 agent-factory API */
  businessDomain?: string;
}

/**
 * ChatKitDataAgent 组件
 * 适配 AISHU Data Agent 平台的智能体对话组件
 * 继承自 ChatKitBase，实现了 generateConversation、sendMessage 和 reduceEventStreamMessage 方法
 */
export class ChatKitDataAgent extends ChatKitBase<ChatKitDataAgentProps> {
  /** 服务端基础地址 */
  private baseUrl: string;

  /** Agent ID */
  private agentId: string;

  /** 是否开启增量流式返回 */
  private incStream: boolean;

  /** 业务域 */
  private businessDomain: string;

  constructor(props: ChatKitDataAgentProps) {
    super(props);

    this.baseUrl = props.baseUrl || 'https://dip.aishu.cn/api/agent-app/v1';
    this.agentId = props.agentId;
    this.incStream = props.enableIncrementalStream ?? true;
    this.businessDomain = props.businessDomain || 'bd_public';

    // 向后兼容：如果传入了 bearerToken 但没有 token，从 bearerToken 中提取 token
    if (props.bearerToken && !props.token) {
      // bearerToken 包含 "Bearer " 前缀，需要移除
      this.token = props.bearerToken.replace(/^Bearer\s+/i, '');
    }
  }

  /**
   * 获取开场白和预置问题
   * 调用 AISHU Data Agent 的 agent-factory API 获取智能体配置信息，提取开场白和预置问题
   * API 端点: GET /api/agent-factory/v3/agent-market/agent/{agent_id}/version/v0
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @returns 返回开场白信息，包含开场白文案和预置问题
   */
  public async getOnboardingInfo(): Promise<OnboardingInfo> {
    try {
      console.log('正在获取 Data Agent 配置...');

      // 构造 agent-factory API 的完整 URL
      // baseUrl 通常是 https://dip.aishu.cn/api/agent-app/v1 或开发环境的 /data-agent
      // 我们需要替换路径为 /api/agent-factory/v3/agent-market/agent/{agent_id}/version/v0
      let agentFactoryUrl: string;
      if (this.baseUrl.startsWith('http://') || this.baseUrl.startsWith('https://')) {
        // 生产环境：使用完整 URL
        const baseUrlObj = new URL(this.baseUrl);
        agentFactoryUrl = `${baseUrlObj.protocol}//${baseUrlObj.host}/api/agent-factory/v3/agent-market/agent/${encodeURIComponent(this.agentId)}/version/v0`;
      } else {
        // 开发环境：使用相对路径走代理
        agentFactoryUrl = `/api/agent-factory/v3/agent-market/agent/${encodeURIComponent(this.agentId)}/version/v0`;
      }

      console.log('调用 agent-factory API:', agentFactoryUrl);

      // 使用 executeWithTokenRefresh 包装 API 调用，支持 token 刷新和重试
      const result = await this.executeWithTokenRefresh(async () => {
        const response = await fetch(agentFactoryUrl, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${this.token}`,
            'Content-Type': 'application/json',
            'x-business-domain': this.businessDomain,
          },
        });

        if (!response.ok) {
          const errorText = await response.text();
          const error: any = new Error(`获取 Data Agent 配置失败: ${response.status} - ${errorText}`);
          error.status = response.status;
          error.body = errorText;
          throw error;
        }

        return await response.json();
      });

      // 从响应中提取开场白和预置问题
      // 根据 agent-factory API 文档,响应格式为: { id, name, config: {...}, ... }
      const config = result.config || {};
      const openingRemarkConfig = config.opening_remark_config || {};
      const presetQuestions = config.preset_questions || [];

      // 构造开场白信息
      let prologue = '你好！我是数据智能体助手，我可以帮你分析数据、回答问题。';
      if (openingRemarkConfig.type === 'fixed' && openingRemarkConfig.fixed_opening_remark) {
        prologue = openingRemarkConfig.fixed_opening_remark;
      }

      // 提取预置问题
      const predefinedQuestions = presetQuestions
        .map((item: any) => item.question)
        .filter((q: any) => typeof q === 'string' && q.trim().length > 0);

      const onboardingInfo: OnboardingInfo = {
        prologue,
        predefinedQuestions,
      };
      return onboardingInfo;
    } catch (error) {
      console.error('获取 Data Agent 配置失败:', error);
      // 返回默认开场白信息
      return {
        prologue: '你好！我是数据智能体助手，我可以帮你分析数据、回答问题。',
        predefinedQuestions: [],
      };
    }
  }

  /**
   * 创建新的会话
   * 调用 Data Agent API 创建新的会话，返回会话 ID
   * API 端点: POST /app/{agent_id}/conversation
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @returns 返回新创建的会话 ID
   */
  public async generateConversation(): Promise<string> {
    try {
      console.log('正在创建 Data Agent 会话...');

      // 构造创建会话的请求体
      const requestBody = {
        agent_id: this.agentId,
        agent_version: 'latest',
      };

      // 使用 executeWithTokenRefresh 包装 API 调用，支持 token 刷新和重试
      const result = await this.executeWithTokenRefresh(async () => {
        const response = await fetch(
          `${this.baseUrl}/app/${this.agentId}/conversation`,
          {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${this.token}`,
            },
            body: JSON.stringify(requestBody),
          }
        );

        if (!response.ok) {
          const errorText = await response.text();
          const error: any = new Error(`创建 Data Agent 会话失败: ${response.status} - ${errorText}`);
          error.status = response.status;
          error.body = errorText;
          throw error;
        }

        return await response.json();
      });
      // 从响应中获取会话 ID
      // 根据 ConversationResponse Schema，响应格式为 { id: string, ttl: string }
      const conversationId = result.data?.id || result.id || '';

      console.log('Data Agent 会话创建成功, conversationID:', conversationId, 'ttl:', result.data?.ttl || result.ttl);
      return conversationId;
    } catch (error) {
      console.error('创建 Data Agent 会话失败:', error);
      // 返回空字符串，允许在没有会话 ID 的情况下继续（API 可能支持自动创建会话）
      return '';
    }
  }

  /**
   * 调用 Data Agent API 发送消息(流式)
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @param text 用户输入
   * @param ctx 应用上下文
   * @param conversationID 发送的对话消息所属的会话 ID
   * @returns 返回助手消息
   */
  public async sendMessage(text: string, ctx: ApplicationContext, conversationID?: string): Promise<ChatMessage> {
    if (!this.baseUrl) {
      throw new Error('Data Agent baseUrl 不能为空');
    }

    // 构造上下文信息
    let fullQuery = text;
    if (ctx && ctx.title) {
      fullQuery = `【上下文: ${ctx.title}】\n${JSON.stringify(ctx.data, null, 2)}\n\n${text}`;
    }

    // 构造请求体，使用传入的 conversationID 参数
    const body = {
      agent_id: this.agentId,
      query: fullQuery,
      stream: true,
      inc_stream: this.incStream,
      custom_querys: ctx?.data,
      conversation_id: conversationID || undefined,
    };

    // 使用 executeWithTokenRefresh 包装 API 调用，支持 token 刷新和重试
    const response = await this.executeWithTokenRefresh(async () => {
      const res = await fetch(
        `${this.baseUrl}/app/${this.agentId}/chat/completion`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'text/event-stream',
            Authorization: `Bearer ${this.token}`,
          },
          body: JSON.stringify(body),
        }
      );

      if (!res.ok) {
        const errText = await res.text();
        const error: any = new Error(`Data Agent API 调用失败: ${res.status} ${errText}`);
        error.status = res.status;
        error.body = errText;
        throw error;
      }

      return res;
    });

    const assistantMessageId = `assistant-${Date.now()}`;
    const initialAssistantMessage: ChatMessage = {
      messageId: assistantMessageId,
      content: '',
      type: ChatMessageType.TEXT,
      role: {
        name: 'AI 助手',
        type: RoleType.ASSISTANT,
        avatar: '',
      },
    };

    this.setState((prevState) => ({
      messages: [...prevState.messages, initialAssistantMessage],
      streamingMessageId: assistantMessageId,
    }));

    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error('无法获取流式响应');
    }

    const finalContent = await this.handleStreamResponse(reader, assistantMessageId);

    return {
      ...initialAssistantMessage,
      content: finalContent,
    };
  }

  /**
   * 解析 Data Agent 的 EventStreamMessage
   * 当 key 仅包含 ["message"] 时，取首词输出；后续仅从包含 ["message", "final_answer"] 的事件取内容
   */
  public reduceEventStreamMessage(
    eventMessage: EventStreamMessage,
    prevBuffer: string
  ): string {
    try {
      const parsed = JSON.parse(eventMessage.data);
      const payload = parsed.data || parsed;
      const action = parsed.action || payload?.action || '';
      const conversationId = parsed.conversation_id || payload?.conversation_id;

      if (conversationId && conversationId !== this.state.conversationID) {
        this.setState({ conversationID: conversationId });
      }

      const key = payload?.key;
      const keyPath = Array.isArray(key)
        ? key.map(String)
        : typeof key === 'string'
          ? [key]
          : [];
      const isMessageOnly = keyPath.length === 1 && keyPath[0] === 'message';
      const hasMessageAndFinal = keyPath.includes('message') && keyPath.includes('final_answer');

      let nextBuffer = prevBuffer;

      // 首词输出
      if (isMessageOnly) {
        const firstChunk = this.extractFirstWord(
          payload?.content?.content?.final_answer?.answer?.text
        );
        if (firstChunk) {
          nextBuffer = prevBuffer + firstChunk;
        }
      } else if (hasMessageAndFinal) {
        // 后续增量输出
        const delta = this.extractContentChunk(payload?.content);

        if (delta) {
          nextBuffer = prevBuffer + delta;
        }
      }

      // action 为 end 代表最后一条输出
      if (action === 'end') {
        return nextBuffer;
      }

      return nextBuffer;
    } catch (e) {
      console.error('解析 Data Agent 事件失败:', e, eventMessage);
      return prevBuffer;
    }
  }

  /**
   * 提取首词，用于第一条文本输出
   */
  private extractFirstWord(raw: any): string {
    if (typeof raw !== 'string') {
      return '';
    }

    const trimmed = raw.trim();
    if (!trimmed) return '';

    const parts = trimmed.split(/\s+/);
    if (parts.length > 1) {
      return parts[0];
    }

    return trimmed.charAt(0);
  }

  /**
   * 提取后续增量输出的内容
   */
  private extractContentChunk(raw: any): string {
    if (!raw) return '';

    if (typeof raw === 'string') {
      return raw;
    }

    if (typeof raw === 'object') {
      if (typeof raw.text === 'string') {
        return raw.text;
      }

      if (typeof raw.value === 'string') {
        return raw.value;
      }

      if (typeof raw.message === 'string') {
        return raw.message;
      }

      if (typeof raw.content === 'string') {
        return raw.content;
      }

      if (typeof raw.answer?.text === 'string') {
        return raw.answer.text;
      }

      if (typeof raw.final_answer?.answer?.text === 'string') {
        return raw.final_answer.answer.text;
      }
    }

    return '';
  }

  /**
   * 检查是否需要刷新 token
   * AISHU Data Agent 平台返回 401 状态码时表示 token 失效
   * @param status HTTP 状态码
   * @param error 错误响应体
   * @returns 返回是否需要刷新 token
   */
  public shouldRefreshToken(status: number, _error: any): boolean {
    // 401 Unauthorized 表示 token 失效
    return status === 401;
  }
}

export default ChatKitDataAgent;
