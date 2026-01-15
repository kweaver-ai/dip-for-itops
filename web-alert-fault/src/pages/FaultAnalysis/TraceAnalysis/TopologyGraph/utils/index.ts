import intl from 'react-intl-universal';
import iconWangluoshebei from 'Assets/images/icon-wangluoshebei.svg';
import iconFuwu from 'Assets/images/icon-fuwu.svg';
import iconFuwuqi from 'Assets/images/icon-fuwuqi.svg';
import iconzhuji from 'Assets/images/icon-zhuji1.svg';
import iconPOD from 'Assets/images/icon-POD.svg';
import zhujianjianIcon from 'Assets/images/icon-zhongjianjian3.svg';
import shujukuIcon from 'Assets/images/icon-shujuku3.svg';
import { TIcon } from '../type';

// g6 对象类型图标列表
export const iconList: TIcon[] = [
  {
    name: intl.get('physical_machine'),
    icon: iconFuwuqi,
    value: 'icon-fuwuqi',
    type: 'physical_machine'
  },
  {
    name: intl.get('service'),
    icon: iconFuwu,
    value: 'icon-fuwu',
    type: 'service'
  },
  { name: intl.get('pod'), icon: iconPOD, value: 'icon-POD', type: 'pod' },
  {
    name: intl.get('network_device'),
    icon: iconWangluoshebei,
    value: 'icon-wangluoshebei',
    type: 'network_device'
  },
  {
    name: intl.get('middleware'),
    icon: zhujianjianIcon,
    value: 'icon-zhongjianjian1',
    type: 'middleware'
  },
  {
    name: intl.get('host'),
    icon: iconzhuji,
    value: 'icon-zhuji1',
    type: 'host'
  },
  {
    name: intl.get('database'),
    icon: shujukuIcon,
    value: 'icon-shujuku2',
    type: 'database'
  }
];

// 高亮颜色
export const highColor = '#126EE3';

// 初始颜色
export const initColor = '#BFBFBF';

export const getMApplicationSystem = (val: string): string => {
  if (val && val.includes('[') && val.includes(']')) {
    return JSON.parse(val).join(',');
  }

  return val;
};

// 获取设备图标
export const getTypeIcon = (type: string): TIcon => {
  const curIcon = iconList.find((val) => val.type === type) || iconList[1];

  curIcon.name = intl.get(curIcon.type);

  return curIcon;
};
