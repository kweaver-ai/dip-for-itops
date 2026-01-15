import dayjs from 'dayjs';
import intl from 'react-intl-universal';

const DATE_FORMAT = 'YYYY-MM-DD HH:mm:ss';

export const dateFormat = (v: string): string => dayjs(v).format(DATE_FORMAT);
const dateFields = [
  's_create_time',
  's_update_time',
  'fault_create_time',
  'fault_update_time',
  'fault_occur_time',
  'fault_latest_time',
  'fault_duration_time',
  'fault_close_time'
];

export const formatValue = (key: string, v: any): string => {
  if (dateFields.includes(key)) {
    return dateFormat(v);
  }

  return v;
};

export const getLabel = (key: string): string => {
  return intl.get(key) || key;
};
