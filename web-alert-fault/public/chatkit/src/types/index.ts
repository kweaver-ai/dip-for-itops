/**
 * 角色类型枚举
 * 发送该消息的角色
 */
export enum RoleType {
  /** 用户 */
  USER = 'User',
  /** AI 助手 */
  ASSISTANT = 'Assistant'
}

/**
 * 消息类型枚举
 * 消息的类型
 */
export enum ChatMessageType {
  /** Markdown 文本类型 */
  TEXT = 'Text',
  /** JSON 类型 */
  JSON = 'JSON',
  /** Widget 组件 */
  WIDGET = 'Widget'
}

/**
 * 角色接口
 * 定义消息发送者的角色信息
 */
export interface Role {
  /** 角色的名称：
   * - 如果 type 是 Assistant，则名称为 "AI 助手"
   * - 如果 type 是 User，则名称为用户的昵称/显示名
   */
  name: string;

  /** 发送该消息的角色 */
  type: RoleType;

  /** 角色的头像，可以是 URL、Base64 或 SVG */
  avatar: string;
}

/**
 * 消息接口
 * 展示在消息区消息列表中的一条消息
 */
export interface ChatMessage {
  /** 一条消息的 ID */
  messageId: string;

  /** 发送该消息的角色 */
  role: Role;

  /** 该条消息的类型 */
  type: ChatMessageType;

  /** 该条消息的内容 */
  content: string;

  /** 与该消息关联的应用上下文（可选），仅用户消息可能包含此字段 */
  applicationContext?: ApplicationContext;
}

/**
 * 应用上下文接口
 * 与用户输入的文本相关的应用上下文
 */
export interface ApplicationContext {
  /** 显示在输入框上方的应用上下文标题 */
  title: string;

  /** 该应用上下文实际包含的数据 */
  data: any;
}

/**
 * EventStream 消息接口
 * 表示从 SSE 接收到的一条流式消息
 */
export interface EventStreamMessage {
  /** 事件类型 */
  event: string;
  /** 消息数据,通常是 JSON 字符串 */
  data: string;
}

/**
 * 开场白信息接口
 * 包含开场白文案和预置问题
 */
export interface OnboardingInfo {
  /** 开场白文案 */
  prologue: string;
  /** 预置问题列表 */
  predefinedQuestions: Array<string>;
}

/**
 * ChatKit 接口
 * 定义了 ChatKit 的一些抽象方法
 */
export interface ChatKitInterface {
  /**
   * 获取开场白和预置问题
   * 该方法需要由子类继承并重写，以适配扣子、Dify 等 LLMOps 平台的接口
   * 返回开场白信息结构体
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @returns 返回开场白信息，包含开场白文案和预置问题
   */
  getOnboardingInfo(): Promise<OnboardingInfo>;

  /**
   * 新建会话
   * 该方法需要由子类继承并重写，以适配扣子、Dify 等 LLMOps 平台的接口
   * 成功返回会话 ID
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @returns 返回新创建的会话 ID
   */
  generateConversation(): Promise<string>;

  /**
   * 向后端发送消息
   * 该方法需要由开发者实现，以适配扣子、Dify等 LLMOps 平台的接口
   * 发送成功后，返回发送的消息结构
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @param text 发送给后端的用户输入的文本
   * @param ctx 随用户输入文本一起发送的应用上下文
   * @param conversationID 发送的对话消息所属的会话 ID
   * @returns 返回发送的消息结构
   */
  sendMessage(
    text: string,
    ctx: ApplicationContext,
    conversationID?: string
  ): Promise<ChatMessage>;

  /**
   * 解析 EventStreamMessage 并累积文本
   * 当接收到 SSE 消息时触发，该方法需要由开发者实现
   * 将不同的 API 接口返回的 SSE 进行解析成 ChatKit 组件能够处理的标准数据格式后返回
   * 返回解析并积累起来后的 buffer，该 buffer 可以被直接打印到界面上
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state
   * @param eventMessage 接收到的一条 EventStreamMessage
   * @param prevBuffer 之前已经堆积起来的文本
   * @returns 返回解析并积累起来后的 buffer
   */
  reduceEventStreamMessage(
    eventMessage: EventStreamMessage,
    prevBuffer: string
  ): string;

  /**
   * 检查是否需要刷新 token
   * 当发生异常时检查是否需要刷新 token。返回 true 表示需要刷新 token，返回 false 表示无需刷新 token。
   * 该方法需要由子类继承并重写，以适配扣子、Dify 等 LLMOps 平台的接口。
   * 注意：该方法是一个无状态无副作用的函数，不允许修改 state。
   * @param status HTTP 状态码
   * @param error 错误响应体
   * @returns 返回是否需要刷新 token
   */
  shouldRefreshToken(status: number, error: any): boolean;
}
