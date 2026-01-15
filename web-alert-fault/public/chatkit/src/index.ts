/**
 * ChatKit - AI 对话组件
 *
 * @module chatkit
 */

// 导入样式文件
import './styles/index.css';

export { ChatKitBase } from './components/ChatKitBase';
export type { ChatKitBaseProps, ChatKitBaseState } from './components/ChatKitBase';

export { ChatKitCoze } from './components/ChatKitCoze';
export type { ChatKitCozeProps } from './components/ChatKitCoze';

export { ChatKitDataAgent } from './components/ChatKitDataAgent';
export type { ChatKitDataAgentProps } from './components/ChatKitDataAgent';

export { RoleType, ChatMessageType } from './types';
export type {
  Role,
  ChatMessage,
  ApplicationContext,
  ChatKitInterface,
  EventStreamMessage,
} from './types';
