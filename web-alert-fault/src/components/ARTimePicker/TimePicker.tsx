import { DatePicker } from 'antd';
import dayjs from 'dayjs';
import './index.less';

const { RangePicker } = DatePicker;

const TimePicker = (props: any): JSX.Element => {
  const { type, value, onChange, ...others } = props;

  const handleChange = (date: any): void => {
    onChange(date);
  };

  return (
    <RangePicker
      picker={type === 'date' || type === 'time' ? undefined : type}
      showTime={
        type === 'time'
          ? ({
              defaultValue: [dayjs('00:00:00', 'HH:mm:ss'), dayjs('00:00:00', 'HH:mm:ss')]
            } as any)
          : false
      }
      value={value}
      onChange={(date): void => handleChange(date)}
      // disabledDate={disabledDate} // 限制日期不可选
      // disabledTime={disabledDateTime} // 限制时间不可选
      {...others}
    />
  );
};

export default TimePicker;
