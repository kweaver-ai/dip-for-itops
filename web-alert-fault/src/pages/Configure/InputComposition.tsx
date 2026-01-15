import { InputNumber, Space, Select } from 'antd';
import { DefaultOptionType } from 'antd/es/select';
import { InputNumberProps } from 'antd/lib/input-number';

const InputComposition = (
  props: Omit<InputNumberProps, 'value' | 'onChange'> & {
    value?: TimeUnit;
    onChange?: (value: TimeUnit) => void;
    options?: DefaultOptionType[];
  }
) => {
  const { value, onChange, options, ...rest } = props;

  const onInputChange = (val: any) => {
    onChange?.({
      time_relativity: val,
      time_type: value?.time_type || 'h'
    });
  };

  const onSelectChange = (val: any) => {
    onChange?.({
      time_relativity: value?.time_relativity || 1,
      time_type: val || ('h' as const)
    });
  };

  return (
    <Space.Compact className="w-[200px]">
      <InputNumber
        {...rest}
        formatter={(val: any) => `${val}`}
        parser={(val: any) => parseInt(val, 10)}
        value={value?.time_relativity}
        onChange={onInputChange}
      />
      <Select
        className="!w-[70px]"
        options={options}
        value={value?.time_type}
        onChange={onSelectChange}
      />
    </Space.Compact>
  );
};

export default InputComposition;
