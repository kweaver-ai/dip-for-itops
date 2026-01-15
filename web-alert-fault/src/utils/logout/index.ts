import Cookies from "js-cookie";

export const logout = async () => {
  // 正常登出
  localStorage.clear();

  // 清除cookies
  Cookies.remove('jwt-token');
};
