import arTimePickerZhCN from 'Components/ARTimePicker/locale/zh-CN.json';
import arTimePickerEnUS from 'Components/ARTimePicker/locale/en-US.json';
import commonZhCN from './common/zh-CN.json';
import commonEnUS from './common/en-US.json';
import alertFaultZhCN from './alert-fault/zh-CN.json';
import alertFaultEnUS from './alert-fault/en-US.json';
import faultAnalysisZhCN from './fault-analysis/zh-CN.json';
import faultAnalysisEnUS from './fault-analysis/en-US.json';
import configureZhCN from './configure/zh-CN.json';
import configureEnUS from './configure/en-US.json';

const zh_CN = {
  ...commonZhCN,
  ...alertFaultZhCN,
  ...faultAnalysisZhCN,
  ...arTimePickerZhCN,
  ...configureZhCN
};
const en_US = {
  ...commonEnUS,
  ...alertFaultEnUS,
  ...faultAnalysisEnUS,
  ...arTimePickerEnUS,
  ...configureEnUS
};

const locales = {
  'zh-CN': zh_CN,
  'en-US': en_US
};

export default locales;
