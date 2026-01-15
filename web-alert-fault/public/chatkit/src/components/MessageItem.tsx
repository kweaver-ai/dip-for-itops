import React from 'react';
import ReactMarkdown from 'react-markdown';
import type { Components } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import { ChatMessage, ChatMessageType, RoleType } from '../types';

/**
 * MessageItem 组件的属性接口
 */
interface MessageItemProps {
  /** 消息对象 */
  message: ChatMessage;
}

/**
 * MessageItem 组件
 * 显示单条消息
 */
const MessageItem: React.FC<MessageItemProps> = ({ message }) => {
  const isUser = message.role.type === RoleType.USER;
  const isAssistantMarkdown =
    message.role.type === RoleType.ASSISTANT &&
    message.type === ChatMessageType.TEXT;

  const renderContent = () => {
    if (isAssistantMarkdown) {
      const markdownComponents: Components = {
        code: ({ className, children, ...props }: any) => {
          const isInline = !className;

          if (isInline) {
            return (
              <code
                className="bg-gray-200 rounded px-1 py-0.5 text-xs font-mono"
                {...props}
              >
                {children}
              </code>
            );
          }

          return (
            <pre
              className="bg-gray-900 text-gray-100 rounded-lg p-3 overflow-auto text-xs"
              {...props}
            >
              <code className={className}>{children}</code>
            </pre>
          );
        },
        ul: ({ children }) => (
          <ul className="list-disc pl-5 space-y-1">{children}</ul>
        ),
        ol: ({ children }) => (
          <ol className="list-decimal pl-5 space-y-1">{children}</ol>
        ),
      };

      return (
        <div className="text-sm leading-relaxed space-y-2 markdown-body">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeHighlight]}
            components={markdownComponents}
          >
            {message.content}
          </ReactMarkdown>
        </div>
      );
    }

    return (
      <p className="text-[14px] whitespace-pre-wrap break-words leading-[24px]">
        {message.content}
      </p>
    );
  };

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}>
      {/* AI 头像 (仅助手消息) */}
      {!isUser && (
        <div className="w-[21px] h-[21px] mr-2 flex-shrink-0 mt-1">
          <img src="/icons/assistant.svg" alt="AI Assistant" className="w-[21px] h-[21px]" />
        </div>
      )}

      <div className="flex flex-col gap-2 max-w-[calc(100%-40px)]">
        {/* 如果用户消息包含应用上下文，在消息内容中显示 */}
        {isUser && message.applicationContext && message.applicationContext.title && (
          <div className="self-end">
            <span className="inline-block bg-[rgba(18,110,227,0.04)] rounded-[8px] px-[14px] py-[5px] text-[12px] leading-[24px] text-[rgba(0,0,0,0.85)]" style={{ fontFamily: 'Noto Sans SC' }}>
              {message.applicationContext.title}
            </span>
          </div>
        )}

        {/* 消息内容 */}
        <div
          className={`rounded-[8px] ${
            isUser
              ? 'bg-[rgba(18,110,227,0.1)] text-[rgba(0,0,0,0.85)] px-[14px] py-[5px]'
              : 'bg-white text-black px-4 py-3'
          }`}
          style={{ fontFamily: 'Noto Sans SC' }}
        >
          {renderContent()}
        </div>
      </div>
    </div>
  );
};

export default MessageItem;
