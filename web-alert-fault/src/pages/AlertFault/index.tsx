import { Tabs } from 'antd';
import intl from 'react-intl-universal';
import Configure from '../Configure';
import Problem from './Problem';
import AlertEvents from './Alert';
import styles from './index.module.less';

function App() {
  const items = [
    {
      label: intl.get('problem_overview'),
      key: '1',
      children: <Problem />
    },
    {
      label: intl.get('event'),
      key: '2',
      children: <AlertEvents />
    },
    {
      label: intl.get('Configure'),
      key: '3',
      children: <Configure />
    }
  ];

  return <Tabs type="card" className={styles.tabs} items={items} />;
}

export default App;
