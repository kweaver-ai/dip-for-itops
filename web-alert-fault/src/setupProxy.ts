import { Agent } from 'https';
import axios from 'axios';

const clientOauth2Prefix = 'client.oauth2_';
const clientOauth2AcessToken = `${clientOauth2Prefix}token`;
const clientOauth2RefreshToken = `${clientOauth2Prefix}refresh_token`;
const OAuthIsSkipLabel = 'oauth2.isSkip';
const refreshTokenMaxAge = 60 * 60 * 1000 * 24 * 7;

export default function setupProxy({
  app,
  DEBUG_ORIGIN,
  ORIGIN
}: {
  app: any;
  DEBUG_ORIGIN: string;
  ORIGIN: string;
}) {
  const webServiceName = '/ar/webisf';
  const REDIRECT_URI = `${ORIGIN}${webServiceName}/oauth2/login/callback`;
  const POST_LOGOUT_REDIRECT_URI = `${ORIGIN}${webServiceName}/oauth2/logout/callback`;

  const registerClientPromise = axios
    .post(
      '/oauth2/clients',
      {
        grant_types: ['authorization_code', 'refresh_token', 'implicit'],
        scope: 'offline openid all',
        redirect_uris: [REDIRECT_URI],
        post_logout_redirect_uris: [POST_LOGOUT_REDIRECT_URI],
        client_name: 'WebDebugClient',
        metadata: {
          device: {
            name: 'WebDebugClient',
            client_type: 'unknown',
            description: 'WebDebugClient'
          }
        },
        response_types: ['token id_token', 'code', 'token']
      },
      {
        baseURL: DEBUG_ORIGIN,
        httpsAgent: new Agent({
          rejectUnauthorized: false
        })
      }
    )
    .then(({ data }) => data);

  app.get(`${webServiceName}/oauth2/login`, async (req, res) => {
    const { client_id } = await registerClientPromise;
    const { redirect, lang } = req.query;
    const state = Buffer.from(decodeURIComponent(redirect)).toString('base64');

    res.cookie('state', state, { httpOnly: true });
    const url = `${DEBUG_ORIGIN}/oauth2/auth?client_id=${client_id}&response_type=code&scope=offline+openid+all&redirect_uri=${encodeURIComponent(
      REDIRECT_URI
    )}&state=${encodeURIComponent(state)}&lang=${lang}`;

    res.redirect(url);
  });

  app.get(`${webServiceName}/oauth2/login/callback`, async (req, res) => {
    const { client_secret, client_id } = await registerClientPromise;
    const { code, state } = req.query;
    const decodeState = Buffer.from(
      decodeURIComponent(state),
      'base64'
    ).toString();
    const params = new URLSearchParams();

    params.append('grant_type', 'authorization_code');
    params.append('code', code);
    params.append('redirect_uri', REDIRECT_URI);

    try {
      const {
        data: { access_token, id_token, refresh_token }
      } = await axios.post(`${DEBUG_ORIGIN}/oauth2/token`, params, {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          Authorization: `Basic ${Buffer.from(
            `${encodeURIComponent(client_id)}:${encodeURIComponent(
              client_secret
            )}`
          ).toString('base64')}`
        },
        httpsAgent: new Agent({
          rejectUnauthorized: false
        })
      });

      res.cookie('client.oauth2_token', access_token, { httpOnly: false });
      res.cookie('id_token', id_token, { httpOnly: false });
      res.cookie('client.oauth2_refresh_token', refresh_token, {
        httpOnly: false
      });
      res.clearCookie('state');
      res.redirect('/fault-analysis');
    } catch (e) {
      res.status(200).json(e.response.data);
    }
  });

  app.get(`${webServiceName}/oauth2/logout`, async (req, res) => {
    const { redirect } = req.query;

    const state = Buffer.from(decodeURIComponent(redirect)).toString('base64');

    res.cookie('state', state, { httpOnly: true });
    res.clearCookie('client.oauth2_refresh_token');
    res.clearCookie('client.oauth2_token');
    res.clearCookie('id_token');
    res.redirect(`${webServiceName}/oauth2/logout/callback?state=${state}`);
  });

  app.get(`${webServiceName}/oauth2/logout/callback`, async (req, res) => {
    const { state } = req.query;

    res.clearCookie('state');
    res.redirect(Buffer.from(decodeURIComponent(state), 'base64').toString());
  });

  app.get(`${webServiceName}/oauth2/login/refreshToken`, async (req, res) => {
    const cookies: Record<string, string> = {};

    if (req.headers?.cookie) {
      // eslint-disable-next-line no-unused-expressions
      req.headers.cookie.split('; ')?.forEach((item) => {
        const [key, value] = item.split('=');

        cookies[key] = value;
      });
    }
    const refreshToken = cookies?.[clientOauth2RefreshToken];
    const isSkip = cookies?.[OAuthIsSkipLabel] === 'true';
    const query = req.query || {};
    const isforced = query.isforced === 'true';

    // 判断是否由控制台免登录进入的客户端
    if (refreshToken) {
      try {
        const data = {
          grant_type: 'refresh_token',
          refresh_token: refreshToken
        };
        const params = new URLSearchParams();

        Object.keys(data).map((key) => {
          params.append(key, data[key]);
        });
        const { client_secret, client_id } = await registerClientPromise;

        const clientID = client_id;
        const clientSecret = client_secret;
        const {
          data: { access_token, id_token, expires_in, refresh_token }
        } = await axios.post('oauth2/token', params, {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            Authorization: `Basic ${Buffer.from(
              `${encodeURIComponent(clientID)}:${encodeURIComponent(
                clientSecret
              )}`
            ).toString('base64')}`
          },

          baseURL: DEBUG_ORIGIN,
          httpsAgent: new Agent({
            rejectUnauthorized: false
          })
        });

        const isMaxAge =
          Number(expires_in) ||
          Number(expires_in) === 0 ||
          Number(expires_in) === -0;
        const isSetSessionExpires = isMaxAge && isSkip;
        const maxAge = isSetSessionExpires ? { maxAge: expires_in * 1000 } : {};
        const refreshMaxAge = isSetSessionExpires
          ? { maxAge: refreshTokenMaxAge }
          : {};

        res.cookie(clientOauth2AcessToken, access_token, {
          httpOnly: false,
          secure: false,
          ...maxAge
        });
        res.cookie('id_token', id_token, {
          httpOnly: false,
          secure: false,
          ...maxAge
        });
        res.cookie(clientOauth2RefreshToken, refresh_token, {
          httpOnly: false,
          secure: false,
          ...refreshMaxAge
        });
        res.clearCookie('state');

        return res
          .header({
            'Cache-Control': 'no-store',
            Pragma: 'no-cache'
          })
          .json({
            code: 200,
            message: 'ok'
          });
      } catch (e) {
        console.log('refresh_token error: ', e);
        if (e && e.response && e.response.status !== 503) {
          const { status, data } = e.response;

          res.statusCode = status;

          return res.json({
            code: status,
            message: data
          });
        }
        const status = 500;

        res.statusCode = status;

        return res.json({
          code: status,
          message: '内部错误，连接hydra服务失败'
        });
      }
    } else {
      const status = 400;

      res.statusCode = status;

      return res.json({
        code: status,
        message: '参数不合法，缺少refreshToken'
      });
    }
  });
}
