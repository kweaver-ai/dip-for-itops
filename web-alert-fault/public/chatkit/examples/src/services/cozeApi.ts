import { COZE_CONFIG } from '../../chatkit_coze/config';
import { Message, Role, InputContext, EventStreamMessage } from '../../../src/types';

/**
 * 扣子 API v3 Chat 响应接口
 */
interface CozeV3ChatResponse {
  /** 响应代码 */
  code: number;
  /** 响应消息 */
  msg: string;
  /** 响应数据 */
  data: {
    /** 会话 ID */
    conversation_id: string;
    /** Chat ID */
    id: string;
    /** 创建时间 */
    created_at: number;
    /** 状态 */
    status: string;
    /** 最后一条错误信息 */
    last_error?: {
      code: number;
      msg: string;
    };
  };
}

/**
 * 扣子 API v3 Chat 检索响应接口
 */
interface CozeV3ChatRetrieveResponse {
  /** 响应代码 */
  code: number;
  /** 响应消息 */
  msg: string;
  /** 响应数据 */
  data: {
    /** Chat ID */
    id: string;
    /** 会话 ID */
    conversation_id: string;
    /** Bot ID */
    bot_id: string;
    /** 创建时间 */
    created_at: number;
    /** 完成时间 */
    completed_at?: number;
    /** 失败时间 */
    failed_at?: number;
    /** 元信息 */
    meta_data?: any;
    /** 最后一条错误 */
    last_error?: {
      code: number;
      msg: string;
    };
    /** 状态: created, in_progress, completed, failed, requires_action */
    status: string;
    /** 必要操作类型 */
    required_action?: any;
    /** 使用情况 */
    usage?: any;
  };
}

/**
 * 扣子 API v3 消息列表响应接口
 */
interface CozeV3MessagesResponse {
  /** 响应代码 */
  code: number;
  /** 响应消息 */
  msg: string;
  /** 响应数据 */
  data: Array<{
    /** 消息 ID */
    id: string;
    /** 会话 ID */
    conversation_id: string;
    /** Bot ID */
    bot_id: string;
    /** Chat ID */
    chat_id: string;
    /** 元信息 */
    meta_data?: any;
    /** 角色 */
    role: string;
    /** 消息类型 */
    type: string;
    /** 消息内容 */
    content: string;
    /** 内容类型 */
    content_type: string;
    /** 创建时间 */
    created_at: number;
    /** 更新时间 */
    updated_at: number;
  }>;
}

/**
 * 扣子 v3 Chat 请求接口
 */
interface CozeV3ChatRequest {
  /** Bot ID */
  bot_id: string;
  /** 用户 ID */
  user_id: string;
  /** 会话 ID (可选,不传则创建新会话) */
  conversation_id?: string;
  /** 是否流式返回 */
  stream: boolean;
  /** 附加消息 */
  additional_messages?: Array<{
    role: 'user' | 'assistant';
    content: string;
    content_type: 'text';
  }>;
}

/**
 * 调用扣子 API v3 发送消息
 * @param message 用户消息
 * @param context 输入上下文
 * @param conversationId 会话 ID
 * @returns 返回助手消息和会话 ID
 */
export async function sendMessageToCoze(
  message: string,
  context: InputContext,
  conversationId?: string
): Promise<{ message: Message; conversationId: string }> {
  // 检查配置
  if (COZE_CONFIG.botId === 'YOUR_BOT_ID' || COZE_CONFIG.apiToken === 'YOUR_API_TOKEN') {
    throw new Error('请先在 examples/chatkit_coze/config.ts 中配置你的扣子 API 信息');
  }

  // 构造上下文信息
  let fullMessage = message;
  if (context && context.title) {
    fullMessage = `【上下文: ${context.title}】\n${JSON.stringify(context.data, null, 2)}\n\n${message}`;
  }

  // 构造请求体
  const requestBody: CozeV3ChatRequest = {
    bot_id: COZE_CONFIG.botId,
    user_id: COZE_CONFIG.userId,
    stream: false,
    additional_messages: [
      {
        role: 'user',
        content: fullMessage,
        content_type: 'text',
      },
    ],
  };

  // 如果有会话 ID,则继续之前的对话
  if (conversationId) {
    requestBody.conversation_id = conversationId;
  }

  try {
    // 第一步: 发起 Chat 请求
    console.log('发起 Chat 请求:', requestBody);
    const chatResponse = await fetch(`${COZE_CONFIG.baseUrl}/v3/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${COZE_CONFIG.apiToken}`,
      },
      body: JSON.stringify(requestBody),
    });

    if (!chatResponse.ok) {
      const errorText = await chatResponse.text();
      throw new Error(`扣子 Chat API 调用失败: ${chatResponse.status} - ${errorText}`);
    }

    const chatData: CozeV3ChatResponse = await chatResponse.json();
    console.log('Chat 响应:', chatData);

    // 检查响应状态
    if (chatData.code !== 0) {
      throw new Error(`扣子 API 返回错误: ${chatData.msg}`);
    }

    const { conversation_id, id: chat_id } = chatData.data;

    // 第二步: 轮询检查 Chat 状态,等待完成
    console.log('开始轮询 Chat 状态...');
    let retries = 60; // 最多等待 60 秒
    let chatStatus = chatData.data.status;

    while (retries > 0 && chatStatus !== 'completed' && chatStatus !== 'failed') {
      await new Promise((resolve) => setTimeout(resolve, 1000)); // 等待 1 秒

      // 检索 Chat 状态
      const retrieveResponse = await fetch(
        `${COZE_CONFIG.baseUrl}/v3/chat/retrieve?conversation_id=${conversation_id}&chat_id=${chat_id}`,
        {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${COZE_CONFIG.apiToken}`,
          },
        }
      );

      if (!retrieveResponse.ok) {
        const errorText = await retrieveResponse.text();
        throw new Error(`检索 Chat 状态失败: ${retrieveResponse.status} - ${errorText}`);
      }

      const retrieveData: CozeV3ChatRetrieveResponse = await retrieveResponse.json();
      console.log('Chat 状态:', retrieveData.data.status);

      if (retrieveData.code !== 0) {
        throw new Error(`检索 Chat 状态失败: ${retrieveData.msg}`);
      }

      chatStatus = retrieveData.data.status;

      // 如果失败,抛出错误
      if (chatStatus === 'failed') {
        const errorMsg = retrieveData.data.last_error?.msg || '未知错误';
        throw new Error(`Chat 执行失败: ${errorMsg}`);
      }

      retries--;
    }

    // 检查是否超时
    if (chatStatus !== 'completed') {
      throw new Error('等待 Chat 完成超时');
    }

    console.log('Chat 已完成,获取消息列表...');

    // 第三步: 获取消息列表
    const messagesResponse = await fetch(
      `${COZE_CONFIG.baseUrl}/v3/chat/message/list?conversation_id=${conversation_id}&chat_id=${chat_id}`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${COZE_CONFIG.apiToken}`,
        },
      }
    );

    if (!messagesResponse.ok) {
      const errorText = await messagesResponse.text();
      throw new Error(`获取消息列表失败: ${messagesResponse.status} - ${errorText}`);
    }

    const messagesData: CozeV3MessagesResponse = await messagesResponse.json();
    console.log('消息列表:', messagesData);

    if (messagesData.code !== 0) {
      throw new Error(`获取消息列表失败: ${messagesData.msg}`);
    }

    // 查找助手回复
    const assistantMessages = messagesData.data.filter(
      (msg) => msg.role === 'assistant' && msg.type === 'answer'
    );

    if (assistantMessages.length === 0) {
      throw new Error('未收到助手回复');
    }

    // 取最后一条助手消息
    const lastAssistantMessage = assistantMessages[assistantMessages.length - 1];

    const assistantMessage: Message = {
      messageId: lastAssistantMessage.id,
      content: lastAssistantMessage.content,
      type: 'text',
      role: Role.ASSISTANT,
    };

    console.log('助手回复:', assistantMessage);

    return {
      message: assistantMessage,
      conversationId: conversation_id,
    };
  } catch (error) {
    console.error('调用扣子 API 失败:', error);
    throw error;
  }
}

/**
 * 使用流式方式调用扣子 API v3
 * @param message 用户消息
 * @param context 输入上下文
 * @param conversationId 会话 ID
 * @param onReceiveEventMessage 接收 EventStream 消息的回调函数
 * @returns 返回助手消息和会话 ID
 */
export async function sendMessageToCozeStream(
  message: string,
  context: InputContext,
  conversationId: string | undefined,
  onReceiveEventMessage: (eventMessage: EventStreamMessage, prevBuffer: string) => string
): Promise<{ message: Message; conversationId: string }> {
  // 检查配置
  if (COZE_CONFIG.botId === 'YOUR_BOT_ID' || COZE_CONFIG.apiToken === 'YOUR_API_TOKEN') {
    throw new Error('请先在 examples/src/config.ts 中配置你的扣子 API 信息');
  }

  // 构造上下文信息
  let fullMessage = message;
  if (context && context.title) {
    fullMessage = `【上下文: ${context.title}】\n${JSON.stringify(context.data, null, 2)}\n\n${message}`;
  }

  // 构造请求体
  const requestBody: CozeV3ChatRequest = {
    bot_id: COZE_CONFIG.botId,
    user_id: COZE_CONFIG.userId,
    stream: true,
    additional_messages: [
      {
        role: 'user',
        content: fullMessage,
        content_type: 'text',
      },
    ],
  };

  if (conversationId) {
    requestBody.conversation_id = conversationId;
  }

  try {
    console.log('发起流式 Chat 请求:', requestBody);

    const response = await fetch(`${COZE_CONFIG.baseUrl}/v3/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${COZE_CONFIG.apiToken}`,
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`扣子流式 API 调用失败: ${response.status} - ${errorText}`);
    }

    // 处理流式响应
    const reader = response.body?.getReader();
    const decoder = new TextDecoder();
    let newConversationId = conversationId || '';
    let buffer = ''; // 累积的文本内容
    let assistantMessageId = '';

    if (!reader) {
      throw new Error('无法获取响应流');
    }

    // 在闭包中处理流式数据
    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          console.log('流式响应完成');
          break;
        }

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split('\n').filter((line) => line.trim());

        for (const line of lines) {
          if (line.startsWith('data:')) {
            const dataStr = line.slice(5).trim();

            if (dataStr === '[DONE]') {
              console.log('收到 DONE 标记');
              continue;
            }

            try {
              const data = JSON.parse(dataStr);
              console.log('流式数据:', data);

              // 记录会话 ID
              if (data.conversation_id) {
                newConversationId = data.conversation_id;
              }

              // 记录 chat_id
              if (data.event === 'conversation.chat.created' && data.id) {
                assistantMessageId = data.id;
              }

              // 构造 EventStreamMessage
              const eventMessage: EventStreamMessage = {
                event: data.event || '',
                data: dataStr,
              };

              // 调用 ChatKit 的标准处理方法,将扣子的事件转换为可显示的文本
              buffer = onReceiveEventMessage(eventMessage, buffer);

            } catch (e) {
              console.error('解析流式响应失败:', e, '原始数据:', dataStr);
            }
          } else if (line.startsWith('event:')) {
            // SSE 事件类型行,可以忽略或记录
            console.log('事件类型:', line.slice(6).trim());
          }
        }
      }
    } finally {
      // 流式传输完成后,闭包会被丢弃
      reader.releaseLock();
    }

    // 返回最终的消息对象
    const assistantMessage: Message = {
      messageId: assistantMessageId || `assistant-${Date.now()}`,
      content: buffer,
      type: 'text',
      role: Role.ASSISTANT,
    };

    console.log('流式响应完成,最终消息:', assistantMessage);

    return {
      message: assistantMessage,
      conversationId: newConversationId,
    };
  } catch (error) {
    console.error('调用扣子流式 API 失败:', error);
    throw error;
  }
}
