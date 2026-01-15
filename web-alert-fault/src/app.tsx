/* eslint-disable import/no-unresolved */
/* eslint-disable @typescript-eslint/ban-ts-comment */
import { useState, useEffect, useRef } from 'react';
import cookie from 'js-cookie';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration';
import { useAppData } from '@noya/max';
import { ConfigProvider, App } from 'antd';
import { StyleProvider } from '@ant-design/cssinjs';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import { setConfig, config } from '@/utils/axios-http/config';
import { initializeI18n } from '@/i18n';
import AIChatProvider from '@/components/AIChatButton/context';
import 'Assets/styles/reset.less';
import './app.css';

const defaultTheme = {
  token: {
    colorPrimary: '#1677ff'
  }
};

function getLocale(lang: string) {
  if (lang === 'en-US') {
    return enUS;
  }

  return zhCN;
}

let parentProps: any = {};

const handleRefreshToken = async () => {
  if (!parentProps.refreshToken) return;

  await parentProps.refreshToken();

  const isfToken = cookie.get('client.oauth2_token');

  if (!isfToken) return;

  config.arLogin(isfToken);
};

export const qiankun = {
  mount(props: any) {
    console.log('mount----', props);
    const { token } = props;
    const { refreshToken, accessToken } = token;

    setConfig({
      refreshToken
    });

    parentProps.refreshToken = refreshToken;
    parentProps.container = props.container;

    if (accessToken) {
      config.arLogin(accessToken);
    }

    console.log('mount-config', config);
  },
  configMap: {
    base: 'route.basename'
  }
};

function AntdProvider({ children }: { children: JSX.Element }) {
  let container = children;

  const [antdLocal, setAntdLocal] = useState(() => getLocale('zh-CN'));
  const refreshTokenRef = useRef<NodeJS.Timeout | null>(null);
  const appData = useAppData();
  const mountNode =
    (appData as any)?.rootElement ??
    document.getElementById('aiAlertFaultRoot')!;

  useEffect(() => {
    const isfToken = cookie.get('client.oauth2_token');

    if (isfToken) {
      refreshTokenRef.current = setInterval(() => {
        handleRefreshToken();
      }, 1000 * 60 * 30);
    } else if (process.env.NODE_ENV === 'development') {
      if (refreshTokenRef.current) {
        clearInterval(refreshTokenRef.current);
        refreshTokenRef.current = null;
      }
      location.href = `${location.protocol}//${location.host}/ar/webisf/oauth2/login?lang=zh-cn&x-forwarded-prefix=&integrated=false`;
    }

    const locale = initializeI18n();

    dayjs.locale(locale === 'zh-CN' ? 'zh-cn' : 'en');
    dayjs.extend(duration);

    const antdLocal: any = getLocale(locale);

    setAntdLocal(antdLocal);

    return () => {
      if (refreshTokenRef.current) {
        clearInterval(refreshTokenRef.current);
        refreshTokenRef.current = null;
      }
    };
  }, []);

  container = (
    <StyleProvider hashPriority="high">
      <ConfigProvider
        locale={antdLocal}
        prefixCls="ar-ant"
        getPopupContainer={() => mountNode}
      >
        <App
          notification={{
            getContainer() {
              return mountNode;
            }
          }}
        >
          {container}
        </App>
      </ConfigProvider>
    </StyleProvider>
  );

  return container;
}

export function rootContainer(container: JSX.Element) {
  return (
    <AntdProvider>
      <AIChatProvider>{container}</AIChatProvider>
    </AntdProvider>
  );
}
