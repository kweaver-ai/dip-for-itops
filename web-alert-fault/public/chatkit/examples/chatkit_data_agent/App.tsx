import React, { useState, useRef } from 'react';
import { ChatKitDataAgent } from '../../src/components/ChatKitDataAgent';
import { ApplicationContext } from '../../src/types';
import { DATA_AGENT_CONFIG } from './config';

/**
 * ChatKitDataAgent Demo
 * 演示 AISHU Data Agent 流式对话
 */
export const ChatKitDataAgentDemo: React.FC = () => {
  const [showChat, setShowChat] = useState(false);
  const chatKitRef = useRef<ChatKitDataAgent>(null);
  const chatOffsetClass = showChat ? 'md:pr-[500px]' : '';

  /**
   * 模拟刷新 token 的方法
   * 实际项目中应该调用真实的 token 刷新接口
   */
  const refreshToken = async (): Promise<string> => {
    console.log('正在刷新 token...');
    // TODO: 在实际项目中，这里应该调用真实的 token 刷新接口
    // 这里仅作演示，返回原 token
    return 'ory_at_SL8UpKSQ6hcQhVylz_Zl6TaUY4hLXJXkzWiKddCJ6rQ.UFX6wqEPTJHGDar_xo1HFKQiH6l-y3apGHbSP50Go0o'
  };

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

  return (
    <div className={`relative flex-1 flex flex-col items-center justify-center p-8 bg-gray-50 min-h-screen ${chatOffsetClass}`}>
      <div className="max-w-2xl w-full bg-white rounded-lg shadow-lg p-8">
        <h1 className="text-3xl font-bold text-gray-800 mb-4">ChatKitDataAgent Demo</h1>
        <p className="text-gray-600 mb-6">
          集成 AISHU Data Agent API 的流式聊天示例。点击下方按钮可以注入上下文并打开聊天窗口。
        </p>

        <div className="space-y-4">
          <button
            onClick={injectExampleContext}
            className="w-full bg-indigo-600 text-white py-3 px-6 rounded-lg hover:bg-indigo-700 transition-colors font-medium"
          >
            添加上下文并打开聊天
          </button>

          <button
            onClick={() => setShowChat(!showChat)}
            className="w-full bg-gray-500 text-white py-3 px-6 rounded-lg hover:bg-gray-600 transition-colors font-medium"
          >
            {showChat ? '关闭聊天窗口' : '打开聊天窗口'}
          </button>
        </div>

        <div className="mt-8 p-4 bg-indigo-50 rounded-lg">
          <h2 className="text-lg font-semibold text-indigo-800 mb-2">功能说明</h2>
          <ul className="text-sm text-indigo-700 space-y-1">
            <li>• key 为 ["message"] 时输出首词；后续从包含 ["message", "final_answer"] 的事件中取 content 增量</li>
            <li>• action 为 end 时结束流式打印</li>
            <li>• 支持对话上下文注入与移除</li>
          </ul>
        </div>

        <div className="mt-6 p-4 bg-green-50 rounded-lg">
          <h2 className="text-lg font-semibold text-green-800 mb-2">Data Agent API 配置</h2>
          <ul className="text-sm text-green-700 space-y-1">
            <li>基础地址: {DATA_AGENT_CONFIG.baseUrl}</li>
            <li>Agent ID: {DATA_AGENT_CONFIG.agentId}</li>
          </ul>
          <p className="text-xs text-green-700 mt-2">
            请确保 bearerToken 正确且包含 Bearer 前缀。
          </p>
        </div>

        <div className="mt-6 p-4 bg-gray-100 rounded-lg">
          <h3 className="text-sm font-semibold text-gray-700 mb-2">示例上下文数据:</h3>
          <pre className="text-xs text-gray-600 bg-white p-3 rounded overflow-x-auto">
{`{
  "title": "故障节点",
  "data": {
    "node_id": "node-uuid-1"
  }
}`}
          </pre>
        </div>
      </div>

      {showChat && (
        <div className="fixed right-4 top-4 bottom-4 w-[480px] max-w-[92vw] z-10">
          <ChatKitDataAgent
            ref={chatKitRef}
            title="Data Agent Copilot"
            visible={showChat}
            onClose={() => setShowChat(false)}
            baseUrl={DATA_AGENT_CONFIG.baseUrl}
            agentId={DATA_AGENT_CONFIG.agentId}
            token={DATA_AGENT_CONFIG.token}
            refreshToken={refreshToken}
            businessDomain={DATA_AGENT_CONFIG.businessDomain}
          />
        </div>
      )}
    </div>
  );
};
