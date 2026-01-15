import React, { useState, useRef } from 'react';
import { ChatKitCoze } from '../../src/components/ChatKitCoze';
import { ApplicationContext } from '../../src/types';
import { COZE_CONFIG } from './config';

/**
 * ChatKitCoze Demo
 * 演示扣子(Coze) 流式对话
 */
export const ChatKitCozeDemo: React.FC = () => {
  const [showChat, setShowChat] = useState(false);
  const chatKitRef = useRef<ChatKitCoze>(null);
  const chatOffsetClass = showChat ? 'md:pr-[500px]' : '';

  /**
   * 注入上下文示例
   */
  const injectExampleContext = () => {
    const context: ApplicationContext = {
      title: '故障节点',
      data: {
        node_id: 'node-uuid-1',
      },
    };

    chatKitRef.current?.injectApplicationContext(context);

    if (!showChat) {
      setShowChat(true);
    }
  };

  /**
   * 一键发送对话示例
   * 直接调用 sendMessage() 方法发送消息
   */
  const sendExampleMessage = async () => {
    if (!showChat) {
      setShowChat(true);
    }

    // 等待组件渲染完成
    setTimeout(async () => {
      const context: ApplicationContext = {
        title: '中心节点',
        data: {
          node_id: 'node-uuid-1',
        },
      };

      try {
        await chatKitRef.current?.send(
          '节点故障,帮我分析可能的原因并给出解决方案',
          context
        );
      } catch (error) {
        console.error('发送消息失败:', error);
      }
    }, 100);
  };

  return (
    <div className={`relative flex-1 flex flex-col items-center justify-center p-8 bg-gray-50 min-h-screen ${chatOffsetClass}`}>
      <div className="max-w-2xl w-full bg-white rounded-lg shadow-lg p-8">
        <h1 className="text-3xl font-bold text-gray-800 mb-4">ChatKitCoze Demo</h1>
        <p className="text-gray-600 mb-6">
          集成扣子 (Coze) API 的流式聊天示例。点击下方按钮可以注入上下文并打开聊天窗口。
        </p>

        <div className="space-y-4">
          <button
            onClick={injectExampleContext}
            className="w-full bg-blue-500 text-white py-3 px-6 rounded-lg hover:bg-blue-600 transition-colors font-medium"
          >
            【添加应用上下文】
          </button>

          <button
            onClick={sendExampleMessage}
            className="w-full bg-green-500 text-white py-3 px-6 rounded-lg hover:bg-green-600 transition-colors font-medium"
          >
            【一键发送对话】
          </button>

          <button
            onClick={() => setShowChat(!showChat)}
            className="w-full bg-gray-500 text-white py-3 px-6 rounded-lg hover:bg-gray-600 transition-colors font-medium"
          >
            {showChat ? '关闭聊天窗口' : '打开聊天窗口'}
          </button>
        </div>

        <div className="mt-8 p-4 bg-blue-50 rounded-lg">
          <h2 className="text-lg font-semibold text-blue-800 mb-2">功能说明</h2>
          <ul className="text-sm text-blue-700 space-y-1">
            <li>• 【添加应用上下文】按钮会注入"故障节点"上下文到 ChatKit</li>
            <li>• 【一键发送对话】按钮会直接发送消息："节点故障,帮我分析可能的原因并给出解决方案",并携带"中心节点"上下文</li>
            <li>• 上下文会显示在输入框上方,可以通过 × 按钮移除</li>
            <li>• ChatKitCoze 组件自动使用流式响应</li>
            <li>• 使用扣子 API 时会保持会话上下文</li>
          </ul>
        </div>

        <div className="mt-6 p-4 bg-green-50 rounded-lg">
          <h2 className="text-lg font-semibold text-green-800 mb-2">扣子 API 配置</h2>
          <ol className="text-sm text-green-700 space-y-1 list-decimal list-inside">
            <li>访问 <a href="https://www.coze.cn/" target="_blank" rel="noopener noreferrer" className="underline">https://www.coze.cn/</a> 创建你的 Bot</li>
            <li>在 Bot 设置中获取 Bot ID</li>
            <li>在个人设置中创建 Personal Access Token</li>
            <li>修改 examples/chatkit_coze/config.ts 中的配置</li>
            <li>自动使用真实的 AI 对话流式响应</li>
          </ol>
        </div>

        <div className="mt-6 p-4 bg-gray-100 rounded-lg">
          <h3 className="text-sm font-semibold text-gray-700 mb-2">【添加应用上下文】示例数据:</h3>
          <pre className="text-xs text-gray-600 bg-white p-3 rounded overflow-x-auto">
{`{
  "title": "故障节点",
  "data": {
    "node_id": "node-uuid-1"
  }
}`}
          </pre>
        </div>

        <div className="mt-4 p-4 bg-gray-100 rounded-lg">
          <h3 className="text-sm font-semibold text-gray-700 mb-2">【一键发送对话】示例数据:</h3>
          <div className="text-xs text-gray-600 bg-white p-3 rounded overflow-x-auto space-y-2">
            <div>
              <strong>消息文本:</strong>
              <p className="mt-1">节点故障,帮我分析可能的原因并给出解决方案</p>
            </div>
            <div>
              <strong>应用上下文:</strong>
              <pre className="mt-1">
{`{
  "title": "中心节点",
  "data": {
    "node_id": "node-uuid-1"
  }
}`}
              </pre>
            </div>
          </div>
        </div>
      </div>

      {showChat && (
        <div className="fixed right-4 top-4 bottom-4 w-[480px] max-w-[92vw] z-10">
          <ChatKitCoze
            ref={chatKitRef}
            botId={COZE_CONFIG.botId}
            apiToken={COZE_CONFIG.apiToken}
            baseUrl={COZE_CONFIG.baseUrl}
            userId={COZE_CONFIG.userId}
            title="Copilot"
            visible={showChat}
            onClose={() => setShowChat(false)}
          />
        </div>
      )}
    </div>
  );
};
