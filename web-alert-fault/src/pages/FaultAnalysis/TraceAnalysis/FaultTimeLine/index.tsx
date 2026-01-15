/* eslint-disable react-hooks/exhaustive-deps */
import React, {
  useState,
  useContext,
  useCallback,
  useRef,
  useEffect,
  forwardRef,
  useImperativeHandle
} from 'react';
import { Flex, Checkbox, Timeline, Tooltip } from 'antd';
import ARIconfont5 from 'Components/ARIconfont5';
import intl from 'react-intl-universal';
import { faultAnalysisContext } from '../../index';
import FaultCard from './FaultCard';
import styles from './index.module.less';

interface FaultTimeLineProps {
  activeFaultIds: string[];
  onPlay: (node: any) => void;
  onPause: (node: any) => void;
  isPlaying: boolean;
  onSelectFaultPoint: (ids: string[]) => void;
  onFaultCardClick: (id: string) => void;
}

const CARD_HEIGHT = 90.5;

const FaultTimeLine = (props: FaultTimeLineProps, ref) => {
  const {
    activeFaultIds,
    onPlay,
    onPause,
    isPlaying,
    onSelectFaultPoint,
    onFaultCardClick
  } = props;

  const {
    problemData: {
      rca_results: {
        rca_context: { backtrace: faultBacktrace = [] }
      },
      root_cause_fault_id: rootCauseFaultId
    }
  } = useContext(faultAnalysisContext);

  const onCheckBoxChange = (v: string[]): void => {
    onSelectFaultPoint(v);
  };

  const containerRef = useRef<HTMLDivElement>(null);

  const scrollTimeLine = () => {
    const container = containerRef.current;
    const activeFaultIdsLength = activeFaultIds.length;

    if (container) {
      container.scrollTop = (activeFaultIdsLength - 3) * CARD_HEIGHT;
    }
  };

  useEffect(() => {
    if (!isPlaying) return;
    if (activeFaultIds.length === 0) return;

    setTimeout(() => {
      const container = containerRef.current;
      const activeFaultIdsLength = activeFaultIds.length;

      if (container && activeFaultIdsLength % 5 === 0) {
        container.scrollTop = (activeFaultIdsLength - 1) * CARD_HEIGHT;
      }
    }, 200);
  }, [activeFaultIds, isPlaying]);

  useImperativeHandle(ref, () => ({
    scrollTimeLine
  }));

  return (
    <Flex className={styles['layout-wrapper']} vertical>
      <Flex className={styles['fault-cards-title']} gap={16} align="center">
        <Tooltip
          placement="right"
          title={isPlaying ? intl.get('pause') : intl.get('play')}
        >
          <span
            className={styles['icon-wrapper']}
            onClick={isPlaying ? onPause : onPlay}
          >
            <ARIconfont5
              type={isPlaying ? 'icon-zanting2' : 'icon-yunhangzhong1'}
            />
          </span>
        </Tooltip>

        <span>{intl.get('fault_backtrace')}</span>
      </Flex>
      <div className={styles['fault-cards-container']} ref={containerRef}>
        <Checkbox.Group value={activeFaultIds} onChange={onCheckBoxChange}>
          <Timeline
            items={faultBacktrace.map((item) => ({
              content: (
                <FaultCard
                  data={item}
                  rootCauseFaultId={rootCauseFaultId?.toString()}
                  key={item.fault_id}
                  active={activeFaultIds.includes(item.fault_id)}
                  onClick={onFaultCardClick}
                />
              ),
              icon: <Checkbox key={item.fault_id} value={item.fault_id} />
            }))}
          />
        </Checkbox.Group>
      </div>
    </Flex>
  );
};

export default forwardRef(FaultTimeLine);
