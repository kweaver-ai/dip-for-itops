import { Select, Space } from 'antd';
import { useUpdateEffect } from '@noya/max';
import dayjs from 'dayjs';
import weekOfYear from 'dayjs/plugin/weekOfYear';
import quarterOfYear from 'dayjs/plugin/quarterOfYear';
import advancedFormat from 'dayjs/plugin/advancedFormat';
import intl from 'react-intl-universal';
import TimePicker from './TimePicker';

// 扩展 dayjs 插件
dayjs.extend(weekOfYear);
dayjs.extend(quarterOfYear);
dayjs.extend(advancedFormat);

export const getTypeByTimeRange = (timeRange: string[]): string => {
  if (!timeRange) {
    return 'time';
  }
  // 时间
  if (timeRange[0]?.includes(':')) {
    return 'time';
  }

  // 日期
  if (/^\d{4}-\d{1,2}-\d{1,2}$/.test(timeRange[0])) {
    return 'date';
  }

  // 周
  if (timeRange[0]?.includes('week')) {
    return 'week';
  }

  // 月
  if (/^\d{4}-\d{1,2}$/.test(timeRange[0])) {
    return 'month';
  }

  // 季度
  if (timeRange[0]?.includes('Q')) {
    return 'quarter';
  }

  // 年
  if (/^\d{4}$/.test(timeRange[0])) {
    return 'year';
  }

  if (timeRange[0] === '' && timeRange[1] === 'now') {
    return 'now';
  }

  return 'quick';
};

export const getDate = (text: string): number[] => {
  const regex = /(\d{4})-(\d{1,2})/;
  const match = text.match(regex);

  if (match) {
    return [Number(match[1]), Number(match[2])];
  }

  return [];
};

export const transformWeekToMoment = (dateString: string[]): dayjs.Dayjs[] => {
  const start = getDate(dateString[0]);
  const end = getDate(dateString[1]);

  // 确保 start 和 end 数组有足够的元素
  const startYear = start[0] || dayjs().year();
  const startWeek = start[1] || dayjs().week();
  const endYear = end[0] || dayjs().year();
  const endWeek = end[1] || dayjs().week();

  return [
    dayjs()
      .year(startYear)
      .week(startWeek)
      .startOf('week'),
    dayjs()
      .year(endYear)
      .week(endWeek)
      .endOf('week')
  ];
};

export const formatDateString = (dateString: string[]): string[] => {
  if (dateString[0].includes('th')) {
    return [dateString[0].replace('th', 'week'), dateString[1].replace('th', 'week')];
  }

  if (dateString[0].includes('周')) {
    return [dateString[0].replace('周', 'week'), dateString[1].replace('周', 'week')];
  }

  return dateString;
};

export const transformTimeToMoment = (dateString: string[]): dayjs.Dayjs[] => {
  if (!dateString) {
    return [];
  }
  let res: dayjs.Dayjs[] = [];
  const type = getTypeByTimeRange(dateString);

  if (dateString[0] === '' && dateString[1] === 'now') {
    return [dayjs('2006-01-01').startOf('date'), dayjs()];
  }

  if (!dateString[0]) {
    return [];
  }

  if (type === 'quick') {
    const unit = dateString[0].charAt(dateString[0].length - 1) as dayjs.ManipulateType;

    // 如果有具体天数，月数
    if (dateString[0].includes('-') && !dateString[0].includes('/')) {
      const num = dateString[0].replace(/[^\d]/g, '');
      const numInt = parseInt(num, 10);

      res[0] = dayjs().subtract(numInt, unit);
      res[1] = dayjs();
    }
    // 如果是过去时，如上周，去年
    if (dateString[0].includes('-') && dateString[0].includes('/')) {
      res[0] = dayjs()
        .subtract(1, unit)
        .startOf(unit);
      res[1] = dayjs()
        .subtract(1, unit)
        .endOf(unit);
    }
    // 如果是现在时，如今天，今年
    if (!dateString[0].includes('-') && dateString[0].includes('/')) {
      res[0] = dayjs().startOf(unit);
      res[1] = dayjs().endOf(unit);
    }

    return res;
  }

  switch (type) {
    case 'time':
      res = [dayjs(dateString[0]), dayjs(dateString[1])];
      break;
    case 'date':
      res = [dayjs(dateString[0]).startOf('date'), dayjs(dateString[1]).endOf('date')];
      break;
    case 'month':
      res = [dayjs(dateString[0]).startOf('month'), dayjs(dateString[1]).endOf('month')];
      break;
    case 'year':
      res = [dayjs(dateString[0], 'YYYY').startOf('year'), dayjs(dateString[1], 'YYYY').endOf('year')];
      break;
    case 'week':
      res = transformWeekToMoment(dateString);
      break;
    case 'quarter':
      res = [
        dayjs()
          .year(Number(dateString[0].split('-')[0]))
          .quarter(Number(dateString[0].split('-')[1][1]))
          .startOf('quarter'),
        dayjs()
          .year(Number(dateString[1].split('-')[0]))
          .quarter(Number(dateString[1].split('-')[1][1]))
          .endOf('quarter')
      ];
      break;
    default:
      res = [dayjs(dateString[0]), dayjs(dateString[1])];
  }

  // if (res[0].valueOf() > dayjs().valueOf() || res[1].valueOf() > dayjs().valueOf()) {
  if (res[0].valueOf() > dayjs().valueOf()) {
    res[1] = dayjs();
  }

  return res;
};

export const tranformMomentToDateString = (momentTime: dayjs.Dayjs[], type: string): string[] | null[] => {
  if (type === 'now') {
    return ['', 'now'];
  }

  if (!momentTime[0] || !momentTime[1]) {
    return [null, null];
  }

  switch (type) {
    case 'time':
      return [momentTime[0].format('YYYY-MM-DD HH:mm:ss'), momentTime[1].format('YYYY-MM-DD HH:mm:ss')];
    case 'date':
      return [momentTime[0].format('YYYY-MM-DD'), momentTime[1].format('YYYY-MM-DD')];
    case 'month':
      return [momentTime[0].format('YYYY-MM'), momentTime[1].format('YYYY-MM')];
    case 'year':
      return [momentTime[0].format('YYYY'), momentTime[1].format('YYYY')];
    case 'week':
      return [`${momentTime[0].format('YYYY-ww')}week`, `${momentTime[1].format('YYYY-ww')}week`];
    case 'quarter':
      return [
        `${momentTime[0].format('YYYY')}-Q${momentTime[0].format('Q')}`,
        `${momentTime[1].format('YYYY')}-Q${momentTime[1].format('Q')}`
      ];
    default:
      return [null, null];
  }
};

interface Props {
  value?: dayjs.Dayjs[];
  onChange?: (arg: dayjs.Dayjs[]) => void;
  showStyle?: 'use' | 'edit';
  type?: string;
  onChangeType?: (val: string) => void;
  getPopupContainer?: () => HTMLElement;
  isClearOnChange?: boolean;
  onOpenChange?: (open: boolean) => void;
}

const TimePickWithType = (props: Props): JSX.Element => {
  const { value = [], onChange, showStyle = 'edit', type, onChangeType, isClearOnChange = false, ...others } = props;

  const handleChange = (val: dayjs.Dayjs[]): void => {
    onChange && onChange(val);
  };

  const handleChangeOption = (val: string): void => {
    onChangeType && onChangeType(val);
  };

  useUpdateEffect(() => {
    isClearOnChange && onChange && onChange([]);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [type]);

  return (
    <Space className="ar-dashboard-time-space">
      {showStyle !== 'use' && (
        <Select
          defaultValue={type || 'time'}
          options={[
            {
              value: 'time',
              label: intl.get('theTime')
            },
            {
              value: 'date',
              label: intl.get('date')
            },
            {
              value: 'week',
              label: intl.get('theWeek')
            },
            {
              value: 'month',
              label: intl.get('theMonth')
            },
            {
              value: 'quarter',
              label: intl.get('theQuarter')
            },
            {
              value: 'year',
              label: intl.get('theYear')
            }
          ]}
          onChange={handleChangeOption}
          {...others}
        />
      )}
      <TimePicker
        value={value}
        onChange={handleChange}
        type={type}
        showStyle={showStyle}
        allowClear={showStyle !== 'use'}
        className={
          showStyle === 'edit'
            ? 'ar-dashboard-time-picker-edit'
            : `ar-dashboard-time-picker-${type} ar-dashboard-time-picker-quick`
        }
        {...others}
      />
    </Space>
  );
};

export default TimePickWithType;
