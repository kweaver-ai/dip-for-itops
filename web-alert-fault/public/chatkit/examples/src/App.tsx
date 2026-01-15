import React, { useMemo, useState } from 'react';
import { ChatKitCozeDemo } from '../chatkit_coze/App';
import { ChatKitDataAgentDemo } from '../chatkit_data_agent/App';

/**
 * Demo 入口,提供两个示例:
 * 1. ChatKitCoze
 * 2. ChatKitDataAgent
 */
const App: React.FC = () => {
  const [activeDemo, setActiveDemo] = useState<'coze' | 'dataAgent'>('coze');

  const demoTitle = useMemo(
    () => (activeDemo === 'coze' ? 'ChatKitCoze' : 'ChatKitDataAgent'),
    [activeDemo]
  );

  return (
    <div className="flex h-screen bg-gray-50">
      <div className="w-72 border-r border-gray-200 bg-white p-6 flex flex-col gap-4">
        <h1 className="text-2xl font-bold text-gray-800">ChatKit Demo</h1>
        <p className="text-sm text-gray-600">
          选择要体验的组件。每个 Demo 支持上下文注入与流式响应。
        </p>
        <div className="flex flex-col gap-2">
          <button
            className={`text-left px-3 py-2 rounded-lg border transition-colors ${
              activeDemo === 'coze'
                ? 'border-blue-500 bg-blue-50 text-blue-700'
                : 'border-gray-200 hover:border-blue-200 text-gray-700'
            }`}
            onClick={() => setActiveDemo('coze')}
          >
            ChatKitCoze Demo
          </button>
          <button
            className={`text-left px-3 py-2 rounded-lg border transition-colors ${
              activeDemo === 'dataAgent'
                ? 'border-indigo-500 bg-indigo-50 text-indigo-700'
                : 'border-gray-200 hover:border-indigo-200 text-gray-700'
            }`}
            onClick={() => setActiveDemo('dataAgent')}
          >
            ChatKitDataAgent Demo
          </button>
        </div>
        <div className="text-xs text-gray-500">
          当前示例: <span className="font-semibold text-gray-700">{demoTitle}</span>
        </div>
      </div>

      {activeDemo === 'coze' ? <ChatKitCozeDemo /> : <ChatKitDataAgentDemo />}
    </div>
  );
};

export default App;
