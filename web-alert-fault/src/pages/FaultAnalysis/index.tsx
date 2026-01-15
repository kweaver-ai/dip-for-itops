import { createContext, useEffect, useState } from 'react';
import { Flex, Breadcrumb, Tabs } from 'antd';
import { useNavigate, useParams } from '@noya/max';
import intl from 'react-intl-universal';
import { getAlertFaultList } from 'Services/alert-fault';
import HeaderInfo from './components/HeaderInfo';
import TraceAnalysis from './TraceAnalysis';
import FaultPoints from './FaultPoints';
import CorrelatedEvents from './CorrelatedEvents';
import Overview from './Overview';
import styles from './index.module.less';
import { transformTimeToMoment } from '@/components/ARTimePicker/TimePickerWithType';

const defaultProblemData = {
  rca_results: {
    rca_context: {
      occurrence: {},
      backtrace: [],
      network: {}
    }
  }
};

export const faultAnalysisContext = createContext<any>({});

function FaultAnalysis() {
  const [problemData, setProblemData] = useState<any>(defaultProblemData);
  const [eventNum, setEventNum] = useState<number>(0);
  const [isGraphFullScreen, setIsGraphFullScreen] = useState<boolean>(false);
  const navigate = useNavigate();
  const params = useParams();

  const tabsData = [
    {
      label: intl.get('detail'),
      key: '1',
      children: <Overview />
    },
    {
      label: intl.get('trace_analysis'),
      key: '2',
      children: <TraceAnalysis />
    },
    {
      label: intl.get('fault_points'),
      key: '3',
      children: <FaultPoints />
    },
    {
      label: intl.get('correlated_events'),
      key: '4',
      children: <CorrelatedEvents onDataChange={setEventNum} />
    }
  ];

  // 初始化获取数据
  const initData = async () => {
    const timeRange = sessionStorage.getItem('arTimeRange') || '[]';
    const dayjsTime = transformTimeToMoment(JSON.parse(timeRange));
    const [start, end] = dayjsTime.map((item) => item.valueOf());

    const formattedParams = {
      filters: {
        field: 'problem_id',
        operation: '==',
        value: params.id,
        value_from: 'const'
      },
      start,
      end,
      offset: 0,
      limit: 1
    };
    const res = await getAlertFaultList(formattedParams);
    const problemData = res?.entries[0] || defaultProblemData;
    let rcaResults = defaultProblemData.rca_results;

    try {
      rcaResults = JSON.parse(problemData.rca_results);
    } catch (error) {
      console.log('error', 'rca_results 解析失败');
    }

    setProblemData({
      ...problemData,
      root_cause_fault_id: problemData.root_cause_fault_id?.toString(),
      rca_results: rcaResults
    });
  };

  const onSetRootCaseNode = (id: any, faultId: any) => {
    setProblemData((prev: any) => ({
      ...prev,
      root_cause_object_id: id,
      root_cause_fault_id: faultId
    }));
  };

  useEffect(() => {
    if (params.id) {
      initData();
    }
    // setProblemData(mockData);
  }, []);

  return (
    <Flex className={styles['layout-wrapper']} vertical>
      <faultAnalysisContext.Provider
        value={{
          problemData,
          isGraphFullScreen,
          setIsGraphFullScreen,
          onSetRootCaseNode
        }}
      >
        <header className={styles.header}>
          <Breadcrumb
            className={styles.breadcrumb}
            items={[
              {
                title: (
                  <span className={styles.breadcrumbLink}>
                    {intl.get('problem')}
                  </span>
                ),
                onClick: () => {
                  navigate(`/fault-analysis`);
                }
              },
              {
                title: problemData?.problem_name
              }
            ]}
          ></Breadcrumb>
          <HeaderInfo />
        </header>
        <div className={styles.content}>
          <Tabs
            defaultActiveKey="1"
            tabBarGutter={24}
            items={tabsData}
            className={styles.tabs}
          />
        </div>
      </faultAnalysisContext.Provider>
    </Flex>
  );
}

export default FaultAnalysis;
