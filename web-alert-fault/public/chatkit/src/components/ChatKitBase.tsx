import { Component } from 'react';
import { ChatMessage, RoleType, ApplicationContext, ChatKitInterface, EventStreamMessage, ChatMessageType, OnboardingInfo } from '../types';
import MessageList from './MessageList';
import InputArea from './InputArea';
import Header from './Header';
import Prologue from './Prologue';

/**
 * ChatKitBase 组件的属性接口
 */
export interface ChatKitBaseProps {
  /** 会话 ID，每次新建会话时由后端返回新的会话唯一标识 */
  conversationID?: string;

  /** 当没有指定的 inputContext 时的默认输入上下文 */
  defaultApplicationContext?: ApplicationContext;

  /** 组件标题 */
  title?: string;

  /** 是否显示组件 */
  visible?: boolean;

  /** 关闭组件的回调函数 */
  onClose?: () => void;

  /** 调用接口时携带的令牌，放置到请求头：Authorization:Bearer {token} */
  token?: string;

  /** 刷新 token 的方法，由集成方传入 */
  refreshToken?: () => Promise<string>;
}

/**
 * ChatKitBase 组件的状态接口
 */
export interface ChatKitBaseState {
  /** 会话 ID，每次新建会话时由后端返回新的会话唯一标识 */
  conversationID: string;

  /** 消息列表，这里仅记录渲染到界面上的对话消息 */
  messages: ChatMessage[];

  /** 用户输入的文本 */
  textInput: string;

  /** 和用户输入文本相关的上下文 */
  applicationContext?: ApplicationContext;

  /** 是否正在发送消息 */
  isSending: boolean;

  /** 当前正在流式更新的消息 ID */
  streamingMessageId: string | null;

  /** 开场白信息，包含开场白文案和预置问题 */
  onboardingInfo?: OnboardingInfo;

  /** 是否正在加载开场白信息 */
  isLoadingOnboarding: boolean;
}

/**
 * ChatKitBase 基础组件
 * AI 对话组件的核心类。该类是一个 React 组件，包含标准的交互界面和交互逻辑。
 *
 * 注意：开发者不能够直接挂载 ChatKitBase 到 Web 应用，而是需要创建一个子类继承
 * ChatKitBase 和 ChatKitInterface，并实现 ChatKitInterface 中定义的方法。
 *
 * 该类实现了 ChatKitInterface 接口，子类需要实现 sendMessage 和 reduceEventStreamMessage 方法
 */
export abstract class ChatKitBase<P extends ChatKitBaseProps = ChatKitBaseProps> extends Component<P, ChatKitBaseState> implements ChatKitInterface {
  /**
   * 标记是否正在初始化或已经初始化
   * 用于防止重复初始化（特别是在 React.StrictMode 下）
   */
  private isInitializing = false;
  private hasInitialized = false;

  /**
   * 调用接口时携带的令牌
   */
  protected token: string;

  /**
   * 刷新 token 的方法，由集成方传入
   */
  protected refreshToken?: () => Promise<string>;

  constructor(props: P) {
    super(props);

    // 初始化 token 和 refreshToken
    this.token = props.token || '';
    this.refreshToken = props.refreshToken;

    this.state = {
      conversationID: props.conversationID || '',
      messages: [],
      textInput: '',
      applicationContext: props.defaultApplicationContext,
      isSending: false,
      streamingMessageId: null,
      onboardingInfo: undefined,
      isLoadingOnboarding: false,
    };
  }

  /**
   * 组件挂载后自动创建会话
   * 根据设计文档要求：组件被初始化的时候会自动新建会话
   */
  async componentDidMount(): Promise<void> {
    // 只在组件初次挂载且可见时自动创建会话
    if (this.props.visible && !this.hasInitialized && !this.isInitializing) {
      await this.initializeConversation();
    }
  }

  /**
   * 组件更新时检查是否需要初始化会话
   * 当 visible 从 false 变为 true 时，如果还未初始化，则初始化会话
   */
  async componentDidUpdate(prevProps: P): Promise<void> {
    // 当 visible 从 false 变为 true，且还未初始化时，初始化会话
    if (!prevProps.visible && this.props.visible && !this.hasInitialized && !this.isInitializing) {
      await this.initializeConversation();
    }
  }

  /**
   * 初始化会话的内部方法
   * 仅在组件首次可见时调用，防止重复初始化
   */
  private async initializeConversation(): Promise<void> {
    // 防止并发调用
    if (this.isInitializing || this.hasInitialized) {
      console.log('会话正在初始化或已初始化，跳过');
      return;
    }

    this.isInitializing = true;
    // 设置加载状态，防止显示默认开场白
    this.setState({ isLoadingOnboarding: true });

    try {
      console.log('ChatKit 组件初始化，自动创建会话...');
      await this.createConversation();
      this.hasInitialized = true;
    } catch (error) {
      console.error('自动创建会话失败:', error);
      // 即使创建会话失败，也尝试获取开场白信息
      try {
        const onboardingInfo = await this.getOnboardingInfo();
        this.setState({ onboardingInfo });
        this.hasInitialized = true;
      } catch (e) {
        console.error('获取开场白信息失败:', e);
      }
    } finally {
      this.isInitializing = false;
      // 加载完成，无论成功或失败都取消加载状态
      this.setState({ isLoadingOnboarding: false });
    }
  }

  /**
   * 获取开场白和预置问题 (抽象方法，由子类实现)
   * 该方法需要由子类继承并重写，以适配扣子、Dify 等 LLMOps 平台的接口
   * 返回开场白信息结构体
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @returns 返回开场白信息，包含开场白文案和预置问题
   */
  public abstract getOnboardingInfo(): Promise<OnboardingInfo>;

  /**
   * 新建会话 (抽象方法，由子类实现)
   * 该方法需要由开发者实现，以适配扣子、Dify 等 LLMOps 平台的接口
   * 成功返回会话 ID
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @returns 返回新创建的会话 ID
   */
  public abstract generateConversation(): Promise<string>;

  /**
   * 向后端发送消息 (抽象方法，由子类实现)
   * 该方法需要由开发者实现，以适配扣子、Dify等 LLMOps 平台的接口
   * 发送成功后，返回发送的消息结构
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @param text 发送给后端的用户输入的文本
   * @param ctx 随用户输入文本一起发送的应用上下文
   * @param conversationID 发送的对话消息所属的会话 ID
   * @returns 返回发送的消息结构
   */
  public abstract sendMessage(text: string, ctx: ApplicationContext, conversationID?: string): Promise<ChatMessage>;

  /**
   * 解析 EventStreamMessage 并累积文本 (抽象方法，由子类实现)
   * 当接收到 SSE 消息时触发，该方法需要由开发者实现
   * 将不同的 API 接口返回的 SSE 进行解析成 ChatKit 组件能够处理的标准数据格式后返回
   * 返回解析并积累起来后的 buffer，该 buffer 可以被直接打印到界面上
   * @param eventMessage 接收到的一条 EventStreamMessage
   * @param prevBuffer 之前已经堆积起来的文本
   * @returns 返回解析并积累起来后的 buffer
   */
  public abstract reduceEventStreamMessage(eventMessage: EventStreamMessage, prevBuffer: string): string;

  /**
   * 检查是否需要刷新 token (抽象方法，由子类实现)
   * 当发生异常时检查是否需要刷新 token。返回 true 表示需要刷新 token，返回 false 表示无需刷新 token。
   * 该方法需要由子类继承并重写，以适配扣子、Dify 等 LLMOps 平台的接口。
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state。
   * @param status HTTP 状态码
   * @param error 错误响应体
   * @returns 返回是否需要刷新 token
   */
  public abstract shouldRefreshToken(status: number, error: any): boolean;

  /**
   * 向 ChatKit 注入应用上下文
   * @param ctx 要注入的应用上下文
   */
  public injectApplicationContext = (ctx: ApplicationContext): void => {
    this.setState({ applicationContext: ctx });
  };

  /**
   * 移除注入的应用上下文
   */
  public removeApplicationContext = (): void => {
    this.setState({ applicationContext: this.props.defaultApplicationContext });
  };

  /**
   * 创建新的会话
   * 内部会调用子类实现的 generateConversation() 和 getOnboardingInfo() 方法
   */
  public createConversation = async (): Promise<void> => {
    // 设置加载状态
    this.setState({ isLoadingOnboarding: true });

    try {
      // 先清除现有会话
      this.clearConversation();

      // 调用子类实现的 generateConversation 方法创建新会话
      const newConversationID = await this.generateConversation();

      // 调用子类实现的 getOnboardingInfo 方法获取开场白信息
      const onboardingInfo = await this.getOnboardingInfo();

      // 更新会话 ID 和开场白信息
      this.setState({
        conversationID: newConversationID,
        onboardingInfo: onboardingInfo,
      });

      console.log('新会话已创建, conversationID:', newConversationID);
      console.log('开场白信息已加载:', onboardingInfo);
    } catch (error) {
      console.error('创建新会话失败:', error);
      throw error;
    } finally {
      // 无论成功或失败，都取消加载状态
      this.setState({ isLoadingOnboarding: false });
    }
  };

  /**
   * 清除会话中的对话消息及会话 ID
   */
  private clearConversation = (): void => {
    this.setState({
      conversationID: '',
      messages: [],
    });
    console.log('会话已清除');
  };

  /**
   * 发送消息的核心方法
   * 该方法是暴露给集成方进行调用的接口，内部会调用子类实现的 sendMessage() 方法
   * @param text 用户输入的文本
   * @param ctx 应用上下文
   * @param conversationID 发送的对话消息所属的会话 ID（可选）
   * @returns 返回发送的消息结构
   */
  public send = async (text: string, ctx?: ApplicationContext, conversationID?: string): Promise<ChatMessage> => {
    if (!text.trim()) {
      throw new Error('消息内容不能为空');
    }

    // 如果传入了 ctx，则设置 applicationContext
    if (ctx) {
      this.setState({ applicationContext: ctx });
    }

    // 使用传入的 conversationID，或使用当前 state 中的 conversationID
    let currentConversationID = conversationID || this.state.conversationID;

    // 如果没有会话 ID，则创建新会话
    if (!currentConversationID) {
      try {
        currentConversationID = await this.generateConversation();
        this.setState({ conversationID: currentConversationID });
        console.log('自动创建新会话, conversationID:', currentConversationID);
      } catch (error) {
        console.error('自动创建会话失败:', error);
        // 即使创建会话失败，也继续发送消息（某些平台可能不需要预先创建会话）
      }
    }

    this.setState({ isSending: true });

    // 获取最终使用的上下文
    const finalContext = ctx || this.state.applicationContext || this.props.defaultApplicationContext || { title: '', data: {} };

    // 创建用户消息
    const userMessage: ChatMessage = {
      messageId: `user-${Date.now()}`,
      content: text,
      type: ChatMessageType.TEXT,
      role: {
        name: '用户',
        type: RoleType.USER,
        avatar: '',
      },
      // 如果有应用上下文，则附加到用户消息中
      applicationContext: finalContext.title || finalContext.data ? finalContext : undefined,
    };

    // 将用户消息添加到消息列表
    this.setState((prevState) => ({
      messages: [...prevState.messages, userMessage],
    }));

    try {
      // 调用子类实现的 sendMessage 方法，传入 conversationID
      const assistantMessage = await this.sendMessage(text, finalContext, currentConversationID);

      // 流式响应时,子类已经添加并更新了消息,这里只需要清理状态
      this.setState({
        textInput: '',
        isSending: false,
        streamingMessageId: null,
      });

      return assistantMessage;
    } catch (error) {
      console.error('发送消息失败:', error);
      this.setState({ isSending: false, streamingMessageId: null });
      throw error;
    }
  };

  /**
   * 处理流式响应
   * 在闭包中处理 EventStream,并在处理完成后丢弃闭包
   * @param reader ReadableStreamDefaultReader
   * @param assistantMessageId 助手消息 ID
   */
  protected async handleStreamResponse(
    reader: ReadableStreamDefaultReader<Uint8Array>,
    assistantMessageId: string
  ): Promise<string> {
    const decoder = new TextDecoder();
    let buffer = '';

    try {
      let currentEvent = ''; // 保存当前的 SSE event 类型

      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          console.log('流式响应完成');
          break;
        }

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split('\n').filter((line) => line.trim());

        for (const line of lines) {
          // 处理 SSE 的 event 行
          if (line.startsWith('event:')) {
            // 保存 SSE 事件类型，用于下一个 data 行
            currentEvent = line.slice(6).trim();
            console.log('收到SSE event:', currentEvent);
            continue;
          }

          if (line.startsWith('data:')) {
            const dataStr = line.slice(5).trim();

            if (dataStr === '[DONE]') {
              console.log('收到 DONE 标记');
              continue;
            }

            try {
              const data = JSON.parse(dataStr);
              console.log('收到SSE数据:', data);

              // 构造 EventStreamMessage
              // 使用 SSE 的 event 行中的事件类型，如果没有则尝试从 data 中获取
              const eventMessage: EventStreamMessage = {
                event: currentEvent || data.event || data.type || '',
                data: dataStr,
              };

              console.log('构造的 EventStreamMessage:', eventMessage);

              // 调用子类的 reduceEventStreamMessage 方法解析事件
              buffer = this.reduceEventStreamMessage(eventMessage, buffer);
              console.log('handleStreamResponse - buffer更新后:', buffer);

              // 更新界面上的消息内容
              this.setState((prevState) => {
                console.log('更新消息列表, assistantMessageId:', assistantMessageId);
                const newMessages = prevState.messages.map((msg) =>
                  msg.messageId === assistantMessageId
                    ? { ...msg, content: buffer }
                    : msg
                );
                console.log('新消息列表:', newMessages);
                return { messages: newMessages };
              });

              // 重置 currentEvent，准备处理下一个事件
              currentEvent = '';
            } catch (e) {
              console.error('解析流式响应失败:', e, '原始数据:', dataStr);
            }
          }
        }
      }
    } finally {
      // 流式传输完成后,闭包会被丢弃
      reader.releaseLock();
    }

    return buffer;
  }

  /**
   * 执行 API 调用，并在需要时自动刷新 token 并重试一次
   * @param apiCall API 调用函数
   * @returns API 调用结果
   */
  protected async executeWithTokenRefresh<T>(
    apiCall: () => Promise<T>
  ): Promise<T> {
    try {
      // 第一次尝试
      return await apiCall();
    } catch (error: any) {
      const status = error.status || error.response?.status || 0;
      const errorBody = error.body || error.response?.data || error;

      // 检查是否需要刷新 token
      const needsRefresh = this.shouldRefreshToken(status, errorBody);

      if (needsRefresh && this.refreshToken) {
        console.log('检测到 token 失效，正在刷新 token...');

        try {
          // 调用 refreshToken 方法获取新 token
          const newToken = await this.refreshToken();

          // 更新 token 属性
          this.token = newToken;

          console.log('Token 刷新成功，正在重试请求...');

          // 重试 API 调用
          try {
            return await apiCall();
          } catch (retryError: any) {
            // 重试后仍然失败，检查是否还是 token 问题
            const retryStatus = retryError.status || retryError.response?.status || 0;
            const retryErrorBody = retryError.body || retryError.response?.data || retryError;

            if (this.shouldRefreshToken(retryStatus, retryErrorBody)) {
              console.error('重试后仍然提示 token 失效，放弃重试');
            }

            // 抛出重试后的错误
            throw retryError;
          }
        } catch (refreshError) {
          console.error('刷新 token 失败:', refreshError);
          // 刷新失败，抛出原始错误
          throw error;
        }
      }

      // 不需要刷新 token，直接抛出错误
      throw error;
    }
  }

  /**
   * 处理发送按钮点击
   */
  private handleSend = async () => {
    if (!this.state.textInput.trim() || this.state.isSending) {
      return;
    }

    const context = this.state.applicationContext || this.props.defaultApplicationContext || { title: '', data: {} };

    console.log('发送消息:', this.state.textInput);
    console.log('选中的上下文:', context);

    try {
      await this.send(this.state.textInput, context);
    } catch (error) {
      console.error('发送消息失败:', error);
    }
  };

  /**
   * 更新用户输入
   */
  private setTextInput = (value: string) => {
    this.setState({ textInput: value });
  };

  /**
   * 处理推荐问题点击
   */
  private handleQuestionClick = (question: string) => {
    this.setState({ textInput: question });
  };

  render() {
    if (!this.props.visible) {
      return null;
    }

    const { title = 'Copilot', onClose } = this.props;
    const { messages, textInput, applicationContext, isSending, onboardingInfo, isLoadingOnboarding } = this.state;
    const showPrologue = messages.length === 0;

    return (
      <div className="flex flex-col h-full w-full bg-white shadow-2xl">
        {/* 头部 */}
        <Header
          title={title}
          onClose={onClose}
          onNewChat={this.createConversation}
        />

        {/* 消息列表区域或欢迎界面 */}
        <div className="flex-1 overflow-y-auto">
          {showPrologue ? (
            isLoadingOnboarding ? (
              // 加载中，显示加载提示
              <div className="flex items-center justify-center h-full">
                <div className="text-center">
                  <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mb-4"></div>
                  <p className="text-sm text-gray-500">正在加载...</p>
                </div>
              </div>
            ) : (
              // 加载完成，显示开场白
              <Prologue
                onQuestionClick={this.handleQuestionClick}
                prologue={onboardingInfo?.prologue}
                predefinedQuestions={onboardingInfo?.predefinedQuestions}
              />
            )
          ) : (
            <MessageList messages={messages} />
          )}
        </div>

        {/* 输入区域 */}
        <InputArea
          value={textInput}
          onChange={this.setTextInput}
          onSend={this.handleSend}
          context={applicationContext}
          onRemoveContext={this.removeApplicationContext}
          disabled={isSending}
        />
      </div>
    );
  }
}

export default ChatKitBase;
