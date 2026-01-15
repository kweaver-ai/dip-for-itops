import cookie from 'js-cookie';
import { getArToken } from '@/services/login';

interface Config {
  refreshToken?: () => Promise<{ accessToken: string }>;
  getAccessToken: () => string;
  arLogin: (isfToken: string) => Promise<void>;
}

const ACCESS_TOKEN_KEY = 'client.oauth2_token';

export function getAccessToken(): string {
  return cookie.get(ACCESS_TOKEN_KEY) || '';
}

export const setConfig = (obj: Record<string, any>) => {
  Object.keys(obj).forEach((key: string) => {
    (config as any)[key] = obj[key];
  });
};

export const arLogin = async (isfToken: string): Promise<void> => {
  cookie.set('client.oauth2_token', isfToken);

  // const resLogin = await getArToken(isfToken);

  // if (resLogin && !resLogin.error_code) {
  //   // 获取当前用户的信息
  //   const { 'jwt-token': jwt } = resLogin;
  //   const { userId, jwtToken } = jwt;

  //   cookie.set('userId', userId);
  //   localStorage.setItem('jwtToken', jwtToken);

  //   return jwt;
  // }

  return isfToken;
};

export const config: Config = {
  refreshToken: undefined,
  getAccessToken,
  arLogin
};
