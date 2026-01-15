import axios from 'axios';
import Cookie from 'js-cookie';
import { merge } from '@noya/max';
import { notification } from 'Components/AIChatButton/context';
import { config } from './config';
import resolveError from './resolveError';

type Method = 'get' | 'post' | 'put' | 'delete';

const language = Cookie.get('language');
const copyLanguage =
  language === 'zh-CN' ? 'zh_CN' : language === 'en-US' ? 'en_US' : '';

const CURRENT_USER_APP = 'currentUserApp';
const JWTTOKEN = 'jwtToken';
/*
 * currentUserApp 和 jwtToken 的赋值不要写在这里，否则会导致请求时无法拿到最新的值
 * const currentUserApp = Cookie.getJSON(CURRENT_USER_APP);
 * const jwtToken = store.get(JWTTOKEN);
 */

// 是否正在刷新token的标记
let isRefreshing = false;
let requests: Array<(token: string) => void> = [];

// 请求列表
const requestList: string[] = [];

// 取消列表
const { CancelToken } = axios;

let sources: Record<string, any> = {};

const service = axios.create({
  baseURL: '/', // 挂载在 process 下的环境变量
  timeout: 300000 // 超时取消请求
});

service.interceptors.request.use(
  (config) => {
    // 由于对kibana的请求，为了防止xsrf攻击需要在请求头部增加kbn-xsrf字段才能正常请求 @zheng.guolei 2019/3/28 3.0.8
    config.headers['kbn-xsrf'] = 'anything';
    config.headers.language = copyLanguage; // 在请求头中添加language字段
    config.headers['Accept-Language'] = language; // 重置Accept-Language 字段
    config.headers['x-language'] = language;
    config.headers['X-Data-Connection-ID'] =
      localStorage.getItem('dataConnectionId');
    config.headers.Authorization = Cookie.get('client.oauth2_token');

    const currentUserApp = Cookie.get(CURRENT_USER_APP);
    const jwtToken = localStorage.getItem(JWTTOKEN);

    if (currentUserApp) {
      const { userId, token } = JSON.parse(currentUserApp);

      config.headers.user = userId;
      config.headers.token = token;
      config.headers['jwt-token'] = jwtToken;
      config.headers.common = { userId, token };
    }

    const request = JSON.stringify(config.url) + JSON.stringify(config.data);

    // 请求处理
    requestList.push(request);

    config.cancelToken = new CancelToken((cancel) => {
      sources[request] = cancel;
    });

    return config;
  },
  (error) => {
    // 异常处理
    return Promise.reject(error);
  }
);

// 响应拦截处理
service.interceptors.response.use(
  (response) => {
    // eslint-disable-next-line prefer-destructuring
    // const needError = response.config.needError ?? true;
    const request =
      JSON.stringify(response.config.url) +
      JSON.stringify(response.config.data);

    // 获取响应后，请求列表里面去除这个值
    requestList.splice(
      requestList.findIndex((item) => item === request),
      1
    );

    // 错误 202
    const errorStatus202 =
      response.status === 202 && response.data && response.data.code;

    // 错误 success 为 0
    const success0 =
      response.data && response.data.success === 0 && response.data.code;

    // 错误码处理
    if (errorStatus202 || success0) {
      const resCode = response.data.code; // 错误码
      const errMsg = response.data.message; // message
      const needMsg = response.data.needMsg || false; // 异常判断字段，类型为Boolean，为false时按原方式处理,为true时抛出message

      // 不需要显示message的错误码
      notification.error({ message: errMsg });
    }

    return response?.data?.error_code ? resolveError(response) : response;
  },
  async (error) => {
    // 没有权限返回登录也
    if (
      config.refreshToken &&
      error.response &&
      error.response.status === 401
    ) {
      try {
        if (!isRefreshing) {
          isRefreshing = true;

          const token = await config.refreshToken();
          const newToken = token ? token.accessToken : config.getAccessToken();

          const jwt = await config.arLogin(newToken);

          if (newToken && jwt) {
            requests.forEach((cb) => cb(newToken));
            requests = [];

            return service.request({
              ...error.config,
              headers: {
                ...(error.config.headers || {}),
                Authorization: `Bearer ${newToken}`
              }
            });
          }

          throw error;
        }

        return new Promise((resolve) => {
          // 将resolve放进队列，用一个函数形式来保存，等token刷新后直接执行
          requests = [
            ...requests,
            (token) =>
              resolve(
                service.request({
                  ...error.config,
                  headers: {
                    ...(error.config.headers || {}),
                    Authorization: `Bearer ${token}`
                  }
                })
              )
          ];
        });
      } catch (e) {
        isRefreshing = false;
        throw error;
      } finally {
        if (!requests.length) {
          isRefreshing = false;
        }
      }

      // const token = error.config.headers['jwt-token'];

      // // jwt-token过期,则退出登录；否则，提示报错信息
      // jwtTokenValidation(token).then((isValide) => {
      //   isValide
      //     ? notification.error({
      //         message: intl.get('AuthenticationFailed')
      //       })
      //     : logout();
      // });

      return;
    }

    // 取消请求
    if (axios.isCancel(error)) {
      requestList.length = 0;
    }
    const request =
      JSON.stringify(error.response.config.url) +
      JSON.stringify(error.response.config.data);

    // 获取响应后，请求列表里面去除这个值
    requestList.splice(
      requestList.findIndex((item) => item === request),
      1
    );

    // 数据流格式报错， 转换JSON返回错误提示
    if (error?.response?.config?.responseType === 'blob') {
      const reader = new FileReader();

      reader.readAsText(error?.response?.data, 'utf-8'); // 读取blob数据为文本
      reader.onload = function (e) {
        try {
          /*
           * 将读取到的文本解析为JSON对象
           * @ts-ignore
           */
          const jsonData = JSON.parse(e.target?.result || '{}');
          // 在这里使用解析后的JSON数据
          const newResponse = {
            ...error.response,
            data: jsonData
          };

          return resolveError(newResponse);
        } catch (error) {
          // 处理解析JSON时可能出现的错误
          console.error('Error parsing JSON:', error);

          if (error instanceof Error) {
            return notification.error({ message: error.message });
          }
        }
      };

      return Promise.reject(error);
    }

    if (error?.response?.data?.error_code) {
      return resolveError(error.response);
    }

    // 取消重复请求，不提示错误信息
    if (
      error?.message !== 'cancleRepeatRequest' &&
      error?.message !== '取消前页面请求'
    ) {
      notification.error({ message: error.message });
    }

    return Promise.reject(error);
  }
);

// 取消全部等待中请求
const clearAllPendingRequest = () => {
  Object.keys(sources).forEach((item) => {
    sources[item]('取消前页面请求');
  });
  sources = {};
};

// axios 对请求的处理
export const request = (
  url: string,
  method: Method,
  params?: any,
  config?: any,
  type?: string
): Promise<any> => {
  return new Promise((resolve, reject) => {
    // get delete合并param和config
    const paramsObj = ['get', 'delete'].includes(method)
      ? { ...params, ...config }
      : params;

    service[method](url, paramsObj, Object.assign({}, config))
      .then(
        (response) => {
          type === 'downLoad' ? resolve(response) : resolve(response?.data);
        },
        (err) => {
          if (err.Cancel) {
            // message.error(err);
          } else {
            // 502
            requestList.length = 0;

            // 需要抛出来，不然promise.all捕捉不了错误
            reject();
            // message.error(`responseErr:${JSON.stringify(err)}`);
          }
        }
      )
      .catch((err) => {
        reject(err);
      });
  });
};

// get方法
export const axiosGet = (url: string, params?: any, config = {}, type = '') => {
  return request(url, 'get', params, config, type);
};

// get body转义
export const axiosGetEncode = (
  url: string,
  body: any,
  params?: any,
  config = {}
) => {
  const key = Object.keys(body);
  let str = '';

  key.map((value) => {
    if (!body[value] && body[value] !== 0) {
      return;
    }
    str += `${value}=${encodeURIComponent(body[value])}&`;

    return value;
  });

  return request(
    `${url}?${str.slice(0, str.length - 1)}`,
    'get',
    params,
    config
  );
};

// delete 方法
export const axiosDelete = (url: string, params?: any, config = {}) => {
  return request(url, 'delete', params, config);
};

// post方法
export const axiosPost = (
  url: string,
  params?: any,
  config = {},
  type = ''
) => {
  return request(url, 'post', params, config, type);
};

export const axiosPostOverridePut = (
  url: string,
  params?: any,
  config = {},
  type = ''
) => {
  return request(
    url,
    'post',
    params,
    merge(config, { headers: { 'x-http-method-override': 'PUT' } }),
    type
  );
};

export const axiosPostOverrideGet = (
  url: string,
  params?: any,
  config = {},
  type = ''
) => {
  return request(
    url,
    'post',
    params,
    merge(config, { headers: { 'x-http-method-override': 'GET' } }),
    type
  );
};

export const axiosPostOverrideDelete = (
  url: string,
  params?: any,
  config = {},
  type = ''
) => {
  return request(
    url,
    'post',
    params,
    merge(config, { headers: { 'x-http-method-override': 'DELETE' } }),
    type
  );
};

export const axiosPostOverridePost = (
  url: string,
  params?: any,
  config = {},
  type = ''
) => {
  return request(
    url,
    'post',
    params,
    merge(config, { headers: { 'x-http-method-override': 'POST' } }),
    type
  );
};

// put方法
export const axiosPut = (url: string, params?: any, config = {}) => {
  return request(url, 'put', params, config);
};

// 通用请求方法
export const axiosRequest = async (params: any) => {
  const { type = 'get', url, data = {} } = params;

  if (!url) return;

  const urlType = type.toLocaleLowerCase();

  if (urlType === 'post') return await axiosPost(url, data);

  if (urlType === 'get') return await axiosGetEncode(url, data);
};

// fetch请求方法
export const fetchRequest = (url: string, config: any) => {
  const { headers = {}, ...others } = config;
  const currentUserApp = Cookie.get(CURRENT_USER_APP);
  const jwtToken = localStorage.getItem(JWTTOKEN);

  headers.language = copyLanguage; // 在请求头中添加language字段
  headers['Accept-Language'] = language; // 重置Accept-Language 字段
  headers['x-language'] = language;

  const newUrl = url;

  if (currentUserApp) {
    const { userId, token } = JSON.parse(currentUserApp);

    headers.user = userId;
    headers.token = token;
    config.headers['jwt-token'] = jwtToken;
    headers.common = { userId, token };
  }

  return fetch(newUrl, { headers, ...others });
};

export default {
  sources,
  clearAllPendingRequest,
  axiosGet,
  axiosGetEncode,
  axiosDelete,
  axiosPost,
  axiosPut,
  axiosRequest,
  axiosPostOverridePut,
  axiosPostOverrideGet,
  axiosPostOverrideDelete,
  axiosPostOverridePost,
  fetchRequest
};
