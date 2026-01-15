import axios from 'axios';
import request from '@/utils/axios-http';

const url = '/api/v1/manager/auth/jwt-token';

export const jwtTokenValidation = async (jwtToken: string): Promise<any> => {
  const res = await new Promise((resolve): void => {
    axios
      .post(
        url,
        {},
        {
          headers: {
            'Jwt-Token': jwtToken
          }
        }
      )
      .then((res) => {
        resolve(res);
      })
      .catch((err) => {
        resolve(err);
      });
  });

  // 202表示jwt-token过期，则status不等于202，则校验通过
  return (res as any)?.status !== 202;
};

export const getArToken = (isfToken: string): Promise<any> => {
  return new Promise((resolve, reject) => {
    axios
      .get(url, {
        headers: {
          'isf-token': isfToken
        }
      })
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err);
      });
  });
};

/**
 * 获取用户信息
 * @param {Object} userId
 */
export const getUser = async (userId: string, jwtToken: string): Promise<any> =>
  await request.axiosGet(
    `/manager/user/${userId}?isAuth=1`,
    {},
    {
      headers: { 'jwt-token': jwtToken }
    }
  );
