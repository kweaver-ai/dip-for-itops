import { ChatKitBase, ChatKitBaseProps } from './ChatKitBase';
import { ChatMessage, ChatMessageType, RoleType, ApplicationContext, EventStreamMessage, OnboardingInfo } from '../types';

/**
 * ChatKitCoze 组件的属性接口
 */
export interface ChatKitCozeProps extends ChatKitBaseProps {
  /** 扣子 Bot ID */
  botId: string;

  /** 扣子 API Token */
  apiToken: string;

  /** 扣子 API 基础 URL */
  baseUrl?: string;

  /** 用户 ID */
  userId?: string;
}

/**
 * ChatKitCoze 组件
 * 专门适配扣子(Coze) API 的智能体对话组件
 * 继承自 ChatKitBase,实现了 generateConversation、sendMessage 和 reduceEventStreamMessage 方法
 */
export class ChatKitCoze extends ChatKitBase<ChatKitCozeProps> {
  /** 扣子 Bot ID */
  private botId: string;

  /** 扣子 API Token */
  private apiToken: string;

  /** 扣子 API 基础 URL */
  private baseUrl: string;

  /** 用户 ID */
  private userId: string;

  constructor(props: ChatKitCozeProps) {
    super(props);

    this.botId = props.botId;
    this.apiToken = props.apiToken;
    this.baseUrl = props.baseUrl || 'https://api.coze.cn';
    this.userId = props.userId || 'chatkit-user';
  }

  /**
   * 获取开场白和预置问题
   * 调用扣子 API 获取智能体配置信息，提取开场白和预置问题
   * API 端点: GET /v1/bots/{bot_id}
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @returns 返回开场白信息，包含开场白文案和预置问题
   */
  public async getOnboardingInfo(): Promise<OnboardingInfo> {
    try {
      console.log('正在获取扣子智能体配置...');

      // 使用正确的 API 端点
      const response = await fetch(`${this.baseUrl}/v1/bots/${encodeURIComponent(this.botId)}?is_published=true`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${this.apiToken}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`获取扣子智能体配置失败: ${response.status} - ${errorText}`);
      }

      const result = await response.json();
      console.log('扣子智能体配置响应:', result);

      // 从响应中提取开场白和预置问题
      // 根据 OpenAPI 文档，data 包含 onboarding_info 对象
      const botInfo = result.data || {};
      const onboardingInfoV2 = botInfo.onboarding_info || {};

      const onboardingInfo: OnboardingInfo = {
        prologue: onboardingInfoV2.prologue || '你好！我是 AI 助手，有什么可以帮你的吗？',
        predefinedQuestions: onboardingInfoV2.suggested_questions || [],
      };

      console.log('开场白信息已提取:', onboardingInfo);
      return onboardingInfo;
    } catch (error) {
      console.error('获取扣子智能体配置失败:', error);
      // 返回默认开场白信息
      return {
        prologue: '你好！我是 AI 助手，有什么可以帮你的吗？',
        predefinedQuestions: [],
      };
    }
  }

  /**
   * 创建新的会话
   * 调用扣子 API 创建新的会话，返回会话 ID
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @returns 返回新创建的会话 ID
   */
  public async generateConversation(): Promise<string> {
    try {
      console.log('正在创建扣子会话...');

      const response = await fetch(`${this.baseUrl}/v1/conversation/create`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiToken}`,
        },
        body: JSON.stringify({}),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`创建扣子会话失败: ${response.status} - ${errorText}`);
      }

      const result = await response.json();
      const conversationId = result.data?.id || result.conversation_id || '';

      console.log('扣子会话创建成功, conversationID:', conversationId);
      return conversationId;
    } catch (error) {
      console.error('创建扣子会话失败:', error);
      // 返回空字符串，允许在没有会话 ID 的情况下继续（扣子 v3 API 支持自动创建会话）
      return '';
    }
  }

  /**
   * 向扣子后端发送消息 (流式响应)
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @param text 发送给后端的用户输入的文本
   * @param ctx 随用户输入文本一起发送的应用上下文
   * @param conversationID 发送的对话消息所属的会话 ID
   * @returns 返回发送的消息结构
   */
  public async sendMessage(text: string, ctx: ApplicationContext, conversationID?: string): Promise<ChatMessage> {
    // 构造上下文信息
    let fullMessage = text;
    if (ctx && ctx.title) {
      fullMessage = `【上下文: ${ctx.title}】\n${JSON.stringify(ctx.data, null, 2)}\n\n${text}`;
    }

    // 构造请求体
    const requestBody = {
      bot_id: this.botId,
      user_id: this.userId,
      stream: true,
      additional_messages: [
        {
          role: 'user',
          content: fullMessage,
          content_type: 'text',
        },
      ],
    };

    // 构造请求 URL，conversation_id 作为 Query Param 传递
    let chatUrl = `${this.baseUrl}/v3/chat`;
    if (conversationID) {
      chatUrl += `?conversation_id=${encodeURIComponent(conversationID)}`;
    }

    try {
      console.log('发起流式 Chat 请求:', { url: chatUrl, body: requestBody });

      const response = await fetch(chatUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiToken}`,
        },
        body: JSON.stringify(requestBody),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`扣子流式 API 调用失败: ${response.status} - ${errorText}`);
      }

      // 创建空的助手消息用于流式更新
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

      // 添加空的助手消息到列表
      this.setState((prevState) => ({
        messages: [...prevState.messages, initialAssistantMessage],
        streamingMessageId: assistantMessageId,
      }));

      // 处理流式响应
      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('无法获取响应流');
      }

      // 调用父类的 handleStreamResponse 方法处理流式数据
      const finalContent = await this.handleStreamResponse(reader, assistantMessageId);

      // 返回最终的消息对象
      return {
        messageId: assistantMessageId,
        content: finalContent,
        type: ChatMessageType.TEXT,
        role: {
          name: 'AI 助手',
          type: RoleType.ASSISTANT,
          avatar: '',
        },
      };
    } catch (error) {
      console.error('调用扣子流式 API 失败:', error);
      throw error;
    }
  }

  /**
   * 解析扣子的 EventStreamMessage 并累积文本
   * @param eventMessage 接收到的一条 EventStreamMessage
   * @param prevBuffer 之前已经堆积起来的文本
   * @returns 返回解析并积累起来后的 buffer
   */
  public reduceEventStreamMessage(eventMessage: EventStreamMessage, prevBuffer: string): string {
    try {
      const data = JSON.parse(eventMessage.data);
      console.log('reduceEventStreamMessage 调用:', {
        event: eventMessage.event,
        prevBuffer,
        data,
      });

      // 记录会话 ID
      if (data.conversation_id && data.conversation_id !== this.state.conversationID) {
        this.setState({ conversationID: data.conversation_id });
      }

      // 扣子v3 API的事件处理
      // 根据 event 字段判断事件类型
      if (eventMessage.event === 'conversation.message.delta') {
        // 增量内容
        if (data.content && data.type === 'answer') {
          const newBuffer = prevBuffer + data.content;
          console.log('增量内容:', data.content, '新buffer:', newBuffer);
          return newBuffer;
        }
      } else if (eventMessage.event === 'conversation.message.completed') {
        // 消息完成,只处理 type 为 answer 的消息
        if (data.content && data.type === 'answer') {
          console.log('消息完成,完整内容:', data.content);
          return data.content;
        }
        // 忽略 verbose 类型的消息
        if (data.type === 'verbose') {
          console.log('忽略 verbose 类型消息');
          return prevBuffer;
        }
      } else if (eventMessage.event === 'conversation.chat.completed') {
        // Chat 完成事件
        console.log('Chat完成');
        return prevBuffer;
      } else if (eventMessage.event === 'conversation.chat.created') {
        // Chat 创建事件
        console.log('Chat创建');
        return prevBuffer;
      } else if (eventMessage.event === 'conversation.chat.in_progress') {
        // Chat 进行中事件
        console.log('Chat进行中');
        return prevBuffer;
      } else if (eventMessage.event === 'conversation.message.started') {
        // 消息开始事件
        console.log('消息开始');
        return prevBuffer;
      } else if (eventMessage.event === 'done') {
        // DONE 事件
        console.log('收到done事件');
        return prevBuffer;
      }

      // 处理扣子的其他消息类型事件
      if (data.msg_type === 'generate_answer_finish') {
        // 答案生成完成事件，保持当前 buffer
        console.log('答案生成完成');
        return prevBuffer;
      }

      // 如果没有 event 字段,尝试从其他字段推断
      // 检查是否是 Chat 完成响应 (status: "completed")
      if (data.status === 'completed') {
        console.log('检测到Chat完成状态');
        return prevBuffer;
      }

      // 其他未知事件类型，保持当前buffer而不是返回原始数据
      console.log('未知事件类型,保持原buffer:', eventMessage.event, data);
      return prevBuffer;
    } catch (e) {
      console.error('解析扣子事件失败:', e);
      return prevBuffer;
    }
  }

  /**
   * 检查是否需要刷新 token
   * Coze 平台返回 401 状态码时表示 token 失效
   * @param status HTTP 状态码
   * @param error 错误响应体
   * @returns 返回是否需要刷新 token
   */
  public shouldRefreshToken(status: number, _error: any): boolean {
    // 401 Unauthorized 表示 token 失效
    return status === 401;
  }
}

export default ChatKitCoze;
