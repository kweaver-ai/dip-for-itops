import { createContext, ReactNode, useRef, useContext } from 'react';
import { App } from 'antd';
import { NotificationInstance } from 'antd/es/notification/interface';
import AIChatButton, { ApplicationContext } from '.';

// eslint-disable-next-line import/no-mutable-exports
let notification: NotificationInstance | null = null;

export interface AIChatContextValue {
  send: (message: string, applicationContext: ApplicationContext) => void;
}

export const AIChatContext = createContext<AIChatContextValue>({
  send: () => {}
});

export const useAIChatContext = () => useContext(AIChatContext);

const AIChatProvider = ({ children }: { children: ReactNode }) => {
  const staticFunction = App.useApp();
  const chatRef = useRef<AIChatContextValue>(null);
  const send = (message: string, context: ApplicationContext) => {
    if (chatRef.current) {
      chatRef.current.send(message, context);
    }
  };

  // eslint-disable-next-line prefer-destructuring
  notification = staticFunction.notification;

  return (
    <AIChatContext.Provider value={{ send }}>
      {children}
      <AIChatButton ref={chatRef} />
    </AIChatContext.Provider>
  );
};

export { notification };

export default AIChatProvider;
