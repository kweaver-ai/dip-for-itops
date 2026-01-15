import React from 'react';

/**
 * Prologue 组件的属性接口
 */
interface PrologueProps {
  /** 点击推荐问题时的回调函数 */
  onQuestionClick?: (question: string) => void;

  /** 开场白文案 */
  prologue?: string;

  /** 预置问题列表 */
  predefinedQuestions?: string[];
}

/**
 * Prologue 组件
 * ChatKit 的欢迎界面，显示欢迎语和推荐问题
 */
const Prologue: React.FC<PrologueProps> = ({
  onQuestionClick,
  prologue,
  predefinedQuestions
}) => {
  // 默认欢迎语
  const defaultPrologue = '你好！我是 Copilot，你的智能浏览助手。我可以帮你理解和分析当前页面的内容。';

  // 默认推荐问题列表
  const defaultQuestions = [
    '🔍 我现在有哪些需要优先处理的高风险工单？',
    '🛠️ 处理 [问题类型] 工单有什么推荐方案？',
    '💡 如何避免处理 [技术领域] 问题的常见错误？',
  ];

  // 使用传入的开场白或默认开场白
  const displayPrologue = prologue || defaultPrologue;

  // 使用传入的预置问题或默认问题
  const questions = predefinedQuestions && predefinedQuestions.length > 0
    ? predefinedQuestions
    : defaultQuestions;

  /**
   * 处理问题点击
   */
  const handleQuestionClick = (question: string) => {
    if (onQuestionClick) {
      onQuestionClick(question);
    }
  };

  return (
    <div className="flex flex-col gap-4 px-6 pt-28 pb-4">
      {/* 欢迎语 */}
      <p
        className="text-[14px] leading-[17px] text-[rgba(0,0,0,0.65)]"
        style={{ fontFamily: 'Noto Sans SC' }}
      >
        {displayPrologue}
      </p>

      {/* 推荐问题列表 */}
      <div className="flex flex-col gap-3">
        {questions.map((question, index) => (
          <button
            key={index}
            onClick={() => handleQuestionClick(question)}
            className="w-full bg-white border border-[rgba(0,0,0,0.1)] rounded-[6px] px-3 py-2 text-left text-[14px] leading-6 text-black hover:border-[#3b9be0] hover:bg-[rgba(18,110,227,0.04)] transition-all"
            style={{ fontFamily: 'Noto Sans SC' }}
          >
            {question}
          </button>
        ))}
      </div>
    </div>
  );
};

export default Prologue;
