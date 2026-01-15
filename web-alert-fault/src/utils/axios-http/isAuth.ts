/**
 * 第三方认证
 * 检测URL地址是否为登录前（进入/home）发送的请求
 */
export default url => {
  let isAuth = false; // 第三方登录，进去/home前发送的请求
  const fieldArr = url.split('?')[1] && url.split('?')[1].split('&');

  Array.isArray(fieldArr) &&
    fieldArr.forEach(value => {
      // 取第一个等号出现的下标
      const flag = value.indexOf('=');
      // 设置key
      const field = value.slice(0, flag);
      // 设置value
      const fieldValue = value.slice(flag + 1);

      if (field === 'isAuth' && fieldValue === '1') {
        isAuth = true;
      }
    });

  return isAuth;
};
