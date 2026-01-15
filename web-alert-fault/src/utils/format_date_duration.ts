import dayjs from 'dayjs';
import intl from 'react-intl-universal';

// duration 格式化时间差 单位毫秒
function formatTimeDifference(duration: number) {
  const durationObj = dayjs.duration(duration);

  const days = durationObj.days();
  const hours = durationObj.hours();
  const minutes = durationObj.minutes();
  const seconds = durationObj.seconds();

  if (days > 0) {
    return `${days}${intl.get('Day')}${hours}${intl.get(
      'Hour'
    )}${minutes}${intl.get('Minute')}${seconds}${intl.get('Second')}`;
  }

  if (hours > 0) {
    return `${hours}${intl.get('Hour')}${minutes}${intl.get(
      'Minute'
    )}${seconds}${intl.get('Second')}`;
  }

  if (minutes > 0) {
    return `${minutes}${intl.get('Minute')}${seconds}${intl.get('Second')}`;
  }

  return `${seconds}${intl.get('Second')}`;
}

export default formatTimeDifference;
