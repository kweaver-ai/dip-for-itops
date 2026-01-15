import React, { useContext, useEffect, useState } from 'react';
import { Drawer } from 'antd';
import { faultAnalysisContext } from '../../index';
import DetialInfo from './DetialInfo';
import styles from './index.module.less';

interface DetialDrawerProps {
  visible: boolean;
  onCancel: () => void;
  faultId: string;
}

const DetialDrawer: React.FC<DetialDrawerProps> = ({
  visible,
  onCancel,
  faultId
}) => {
  const [detialData, setDetialData] = useState<any>({});

  const {
    problemData: {
      rca_results: {
        rca_context: { backtrace = [] }
      }
    }
  } = useContext(faultAnalysisContext);

  useEffect(() => {
    if (visible) {
      setDetialData(backtrace.find((item) => item.fault_id === faultId));

      return;
    }

    setTimeout(() => {
      setDetialData({});
    }, 500);
  }, [visible]);

  return (
    <Drawer
      title={detialData?.fault_name}
      placement="bottom"
      onClose={onCancel}
      open={visible}
      size={400}
      className={styles.drawer}
      destroyOnHidden
    >
      <DetialInfo dataInfo={detialData} />
    </Drawer>
  );
};

export default DetialDrawer;
