import {
  forwardRef,
  ForwardedRef,
  useEffect,
  useImperativeHandle,
  useRef,
  useState
} from 'react';
import { FloatButton } from 'antd';
import { Copilot, type ApplicationContext } from '@kweaver-ai/chatkit';
import { useBoolean } from '@noya/max';
import intl from 'react-intl-universal';
import Cookies from 'js-cookie';
import ARIconfont5 from '@/components/ARIconfont5';

export interface AIChatButtonProps {}

export type { ApplicationContext };

const AIChatButton = (
  props: AIChatButtonProps,
  ref: ForwardedRef<{
    send: (message: string, applicationContext: ApplicationContext) => void;
  }>
) => {
  const [chatVisible, { setTrue, setFalse }] = useBoolean(false);
  const chatKitRef = useRef<Copilot>(null);
  const [token, setToken] = useState('');

  useEffect(() => {
    const token = Cookies.get('client.oauth2_token');

    if (!token) {
      return;
    }

    setToken(token);
  }, []);

  const send = (message: string, applicationContext: ApplicationContext) => {
    setTrue();
    setTimeout(async () => {
      try {
        await chatKitRef.current?.send(message, applicationContext as any);
      } catch (error) {
        console.error('AIChatButton send error:', error);
      }
    }, 10);
  };

  useImperativeHandle(ref, () => ({
    send
  }));

  return (
    <>
      {!chatVisible && (
        <FloatButton
          icon={<ARIconfont5 type="icon-AIyunwei" />}
          onClick={() => setTrue()}
        />
      )}
      {chatVisible && (
        <div className="absolute right-4 top-4 bottom-4 w-[480px] max-w-[92vw] z-10">
          <Copilot
            ref={chatKitRef}
            agentKey="01KDKWEAP7HF5BZYXX551Z5ETG"
            bearerToken={`Bearer ${token}`}
            title={intl.get('AIYunwei')}
            visible={chatVisible}
            baseUrl="/api/agent-app/v1"
            onClose={() => setFalse()}
          />
        </div>
      )}
    </>
  );
};

export default forwardRef(AIChatButton);
