import React from 'react';

/**
 * Header 组件的属性接口
 */
interface HeaderProps {
  /** 组件标题 */
  title?: string;

  /** 关闭组件的回调函数 */
  onClose?: () => void;

  /** 新建会话的回调函数 */
  onNewChat?: () => void;

  /** 展开/收起的回调函数 */
  onExpand?: () => void;

  /** 更多选项的回调函数 */
  onMore?: () => void;
}

/**
 * Header 组件
 * ChatKit 的头部组件，包含 Logo、标题和操作按钮
 */
const Header: React.FC<HeaderProps> = ({
  title = 'Copilot',
  onClose,
  onNewChat,
  onExpand,
  onMore,
}) => {
  return (
    <div className="relative h-14 w-full bg-white">
      <div className="absolute left-[23px] top-3 h-8 w-[440px]">
        {/* Logo 和标题 */}
        <div className="absolute left-0 top-0 flex items-center h-8">
          {/* AI 助手图标 */}
          <div className="w-8 h-8">
            <img src="/icons/assistant.svg" alt="Copilot" className="w-8 h-8" />
          </div>
          {/* 标题 */}
          <div className="absolute left-[40px] top-1/2 -translate-y-1/2">
            <p className="font-medium text-[18px] leading-4 text-black whitespace-nowrap" style={{ fontFamily: 'Noto Sans SC' }}>
              {title}
            </p>
          </div>
        </div>

        {/* 操作按钮组 - 从右到左：Close -> Expand -> New Chat -> More */}
        <div className="absolute right-0 top-0 h-8 flex items-center">
          {/* 更多按钮 */}
          {onMore && (
            <button
              onClick={onMore}
              className="w-5 h-5 flex items-center justify-center hover:opacity-70 transition-opacity"
              title="更多选项"
            >
              <img src="/icons/more.svg" alt="更多" className="w-5 h-5" />
            </button>
          )}

          {/* 新建对话按钮 */}
          {onNewChat && (
            <button
              onClick={onNewChat}
              className="ml-2 w-[14px] h-[14px] flex items-center justify-center hover:opacity-70 transition-opacity"
              title="新建会话"
            >
              <img src="/icons/new.svg" alt="新建会话" className="w-[14px] h-[14px]" />
            </button>
          )}

          {/* 展开按钮 */}
          {onExpand && (
            <button
              onClick={onExpand}
              className="ml-2 w-[14px] h-[13px] flex items-center justify-center hover:opacity-70 transition-opacity"
              title="展开"
            >
              <img src="/icons/expand.svg" alt="展开" className="w-[14px] h-[13px]" />
            </button>
          )}

          {/* 关闭按钮 */}
          {onClose && (
            <button
              onClick={onClose}
              className="ml-2 w-4 h-4 flex items-center justify-center hover:opacity-70 transition-opacity"
              title="关闭"
            >
              <img src="/icons/close.svg" alt="关闭" className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default Header;
