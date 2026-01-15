# ChatKit

ChatKit 是一个 AI 对话组件。Web 应用开发者可以将 ChatKit 集成到自己的前端代码中，并通过传入与用户输入有关的上下文对象，实现对 Web 应用中有关的信息发起 AI 对话。

## 特性

- **多轮对话**: 支持基于会话 ID 的多轮对话，保持上下文连贯性
- **会话管理**: 支持创建新会话、清除会话等操作
- **应用上下文**: 支持注入应用上下文，让 AI 理解用户操作背景
- **流式响应**: 支持 SSE 流式响应，实现打字机效果
- **Markdown 渲染**: AI 助手消息支持 Markdown 格式渲染
- **多平台适配**: 支持扣子(Coze)、AISHU Data Agent 等平台

## 项目结构

```
chatkit/
├── src/                          # ChatKit 组件源码
│   ├── components/               # React 组件
│   │   ├── ChatKitBase.tsx       # 核心基类组件
│   │   ├── ChatKitCoze.tsx       # 扣子平台适配组件
│   │   ├── ChatKitDataAgent.tsx  # AISHU Data Agent 适配组件
│   │   ├── MessageList.tsx       # 消息列表组件
│   │   ├── MessageItem.tsx       # 消息项组件
│   │   └── InputArea.tsx         # 输入区域组件
│   ├── types/                    # TypeScript 类型定义
│   │   └── index.ts              # 类型定义文件
│   ├── styles/                   # 样式文件
│   │   └── index.css             # 全局样式
│   └── index.ts                  # 导出入口
├── examples/                     # Demo 示例应用
│   ├── chatkit_coze/             # 扣子 Demo
│   └── chatkit_data_agent/       # Data Agent Demo
├── openapi/                      # OpenAPI 规范（Coze 与 Data Agent）
│   ├── coze.openapi.yaml         # Coze API 入口
│   ├── data-agent.openapi.yaml   # Data Agent API 入口
│   ├── bots/                     # Coze Bot schemas 与 paths
│   ├── chat/                     # Coze Chat schemas 与 paths
│   └── data-agent/               # Data Agent 请求/响应 schemas 与 paths
├── api/                          # 预留 API 目录
├── design/                       # 设计文档
├── public/                       # 静态资源
└── package.json
```

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:5173/ 查看 demo 应用。

### 3. 构建项目

```bash
npm run build
```

### 4. 配置 Coze 接口

- 编辑 `examples/chatkit_coze/config.ts`，填入你的 `baseUrl`、`botId`、`apiToken` 和 `userId`。
- 需要使用 https://www.coze.cn/ 的 Personal Access Token，并确保 Bot ID 与 API Token 均可用。

### 5. 配置 Data Agent 接口

- 编辑 `examples/chatkit_data_agent/config.ts`，填入你的 `baseUrl`、`agentId` 和包含 Bearer 前缀的 `bearerToken`。
- 若使用代理前缀（如 `/data-agent`），需在本地代理配置中转发到真实的 Data Agent 网关。

## 使用方法

### 使用 ChatKitCoze (扣子平台)

```tsx
import React, { useRef } from 'react';
import { ChatKitCoze } from 'chatkit';

function App() {
  const chatKitRef = useRef<ChatKitCoze>(null);

  // 注入应用上下文
  const injectContext = () => {
    chatKitRef.current?.injectApplicationContext({
      title: '故障节点',
      data: { node_id: 'node-uuid-1' },
    });
  };

  // 创建新会话
  const createNewConversation = () => {
    chatKitRef.current?.createConversation();
  };

  // 发送消息
  const sendMessage = () => {
    chatKitRef.current?.send('帮我分析这个故障', {
      title: '中心节点',
      data: { node_id: 'node-uuid-1' },
    });
  };

  return (
    <div>
      <button onClick={injectContext}>添加上下文</button>
      <button onClick={createNewConversation}>新建会话</button>
      <button onClick={sendMessage}>发送消息</button>
      <ChatKitCoze
        ref={chatKitRef}
        botId="你的Bot ID"
        apiToken="你的API Token"
        title="Copilot"
        visible={true}
      />
    </div>
  );
}
```

### 使用 ChatKitDataAgent (AISHU Data Agent)

```tsx
import React, { useRef } from 'react';
import { ChatKitDataAgent } from 'chatkit';

function App() {
  const chatKitRef = useRef<ChatKitDataAgent>(null);

  return (
    <ChatKitDataAgent
      ref={chatKitRef}
      agentId="你的Agent ID"
      bearerToken="Bearer your-token"
      title="Copilot"
      visible={true}
    />
  );
}
```

## API 文档

### ChatKitBase 基类

ChatKitBase 是 AI 对话组件的核心基类。开发者不能直接挂载 ChatKitBase，而是需要使用子类（如 ChatKitCoze、ChatKitDataAgent）或创建自定义子类。

#### 属性 (Props)

| 属性名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| conversationID | `string` | 否 | `''` | 会话 ID |
| defaultApplicationContext | `ApplicationContext` | 否 | - | 默认的应用上下文 |
| title | `string` | 否 | `'Copilot'` | 组件标题 |
| visible | `boolean` | 否 | `true` | 是否显示组件 |
| onClose | `() => void` | 否 | - | 关闭组件的回调函数 |

#### 公开方法

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| createConversation | `()` | `Promise<void>` | 创建新的会话，清除现有消息 |
| send | `(text: string, ctx?: ApplicationContext, conversationID?: string)` | `Promise<ChatMessage>` | 发送消息，支持传入上下文和会话ID |
| injectApplicationContext | `(ctx: ApplicationContext)` | `void` | 向 ChatKit 注入应用上下文 |
| removeApplicationContext | `()` | `void` | 移除注入的应用上下文 |

### ChatKitCoze

扣子(Coze)平台适配组件。

#### 额外属性

| 属性名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| botId | `string` | 是 | - | 扣子 Bot ID |
| apiToken | `string` | 是 | - | 扣子 API Token |
| baseUrl | `string` | 否 | `'https://api.coze.cn'` | 扣子 API 基础 URL |
| userId | `string` | 否 | `'chatkit-user'` | 用户 ID |

### ChatKitDataAgent

AISHU Data Agent 平台适配组件。

#### 额外属性

| 属性名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| agentId | `string` | 是 | - | Agent ID |
| bearerToken | `string` | 是 | - | 访问令牌（需包含 Bearer 前缀）|
| baseUrl | `string` | 否 | `'https://dip.aishu.cn/api/agent-app/v1'` | 服务端基础地址 |
| enableIncrementalStream | `boolean` | 否 | `true` | 是否开启增量流式返回 |

### 类型定义

#### ChatMessage

```typescript
interface ChatMessage {
  /** 消息 ID */
  messageId: string;
  /** 发送该消息的角色 */
  role: Role;
  /** 消息类型 */
  type: ChatMessageType;
  /** 消息内容 */
  content: string;
  /** 与该消息关联的应用上下文（可选）*/
  applicationContext?: ApplicationContext;
}
```

#### ApplicationContext

```typescript
interface ApplicationContext {
  /** 显示在输入框上方的应用上下文标题 */
  title: string;
  /** 该应用上下文实际包含的数据 */
  data: any;
}
```

#### RoleType

```typescript
enum RoleType {
  /** 用户 */
  USER = 'User',
  /** AI 助手 */
  ASSISTANT = 'Assistant'
}
```

#### ChatMessageType

```typescript
enum ChatMessageType {
  /** Markdown 文本类型 */
  TEXT = 'Text',
  /** JSON 类型 */
  JSON = 'JSON',
  /** Widget 组件 */
  WIDGET = 'Widget'
}
```

### ChatKitInterface 接口

如果需要创建自定义的 ChatKit 子类，需要实现 ChatKitInterface 接口中的方法：

```typescript
interface ChatKitInterface {
  /** 新建会话，返回会话 ID */
  generateConversation(): Promise<string>;

  /** 向后端发送消息 */
  sendMessage(
    text: string,
    ctx: ApplicationContext,
    conversationID?: string
  ): Promise<ChatMessage>;

  /** 解析 EventStreamMessage 并累积文本 */
  reduceEventStreamMessage(
    eventMessage: EventStreamMessage,
    prevBuffer: string
  ): string;
}
```

## 技术栈

- **TypeScript** - 类型安全的 JavaScript
- **React** - UI 框架
- **Tailwind CSS** - 样式框架
- **Vite** - 构建工具
- **react-markdown** - Markdown 渲染

## 协议

MIT
