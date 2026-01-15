import { useRef, useState, useContext, useEffect } from 'react';
import { Flex } from 'antd';
import dayjs from 'dayjs';
import { faultAnalysisContext } from '../index';
import FaultTimeLine from './FaultTimeLine';
import FaultAnalysisTimeGraph from './TimeGraph';
import TopologyGraph from './TopologyGraph';
import styles from './index.module.less';
import { set } from 'node_modules/@types/js-cookie';

let timerId: any = null;
let intervalId: any = null;
let currentIndex = 0;
// 最小时间范围 3 分钟
const MixTimeDuration = 1000 * 60 * 3;

function FaultAnalysis() {
  const topologGraphRef = useRef<any>(null);
  const isPlayingRef = useRef<boolean>(false);
  const [isPlaying, setIsPlaying] = useState<boolean>(false);
  const [activeFaultIds, setActiveFaultIds] = useState<string[]>([]);
  const [sideValue, setSideValue] = useState<number>(0);
  const [sideStartTime, setSideStartTime] = useState<number>(0);
  const [sideEndTime, setSideEndTime] = useState<number>(0);
  const faultTimeLineRef = useRef<any>(null);

  const {
    problemData: {
      rca_results: {
        rca_context: { backtrace: faultBacktrace = [] }
      }
    }
  } = useContext(faultAnalysisContext);

  const allFaultIds = faultBacktrace.map((item) => item.fault_id);

  const initSidle = () => {
    const startTime = dayjs(faultBacktrace[0]?.fault_occur_time).unix() * 1000;
    let endTime =
      dayjs(
        faultBacktrace[faultBacktrace.length - 1]?.fault_latest_time
      ).unix() * 1000;

    // 时间范围不能小于 3 分钟
    if (endTime - startTime < MixTimeDuration) {
      endTime = startTime + MixTimeDuration;
    }

    setSideStartTime(startTime);
    setSideEndTime(endTime);
  };

  const onPlay = () => {
    if (isPlayingRef.current) {
      return;
    }
    setIsPlaying(true);
    isPlayingRef.current = true;
    timerId && clearTimeout(timerId);

    topologGraphRef.current?.onBeforePlay();
    setActiveFaultIds([]);
    currentIndex = 0;

    const intervalTimeNum = allFaultIds.length > 10 ? 900 : 1300;

    setTimeout(() => {
      faultTimeLineRef.current?.scrollTimeLine();
    }, 100);

    intervalId = setInterval((): void => {
      if (currentIndex >= allFaultIds.length || !isPlayingRef.current) {
        clearTimeout(intervalId as any);
        setIsPlaying(false);
        isPlayingRef.current = false;
        setActiveFaultIds([]);
        setSideValue(0);
        topologGraphRef.current?.onRestGraph();

        return;
      }

      setActiveFaultIds((prev) => {
        return [...prev, allFaultIds[currentIndex]?.toString()];
      });
      topologGraphRef.current?.onPlayActive(allFaultIds[currentIndex]);

      const newSideValue =
        dayjs(faultBacktrace[currentIndex]?.fault_occur_time).unix() * 1000;

      setSideValue(newSideValue);

      if (
        allFaultIds.length - 1 === currentIndex &&
        sideEndTime > newSideValue
      ) {
        timerId = setTimeout(() => {
          setSideValue(sideEndTime);
        }, 500);
      }

      currentIndex++;
    }, intervalTimeNum);
  };

  const onPause = () => {
    clearInterval(intervalId);
    isPlayingRef.current = false;
    setIsPlaying(false);
    topologGraphRef.current?.onRestGraph();
  };

  // 处理选则故障点，切换选中状态的逻辑
  const handleSelChangeGraph = (ids: string[]): void => {
    intervalId && clearInterval(intervalId);
    if (isPlayingRef.current) {
      setIsPlaying(false);
      isPlayingRef.current = false;
      topologGraphRef.current?.onRestGraph();
    }

    topologGraphRef.current?.onSelectPatchActive(ids);
    setActiveFaultIds(ids);
  };

  // 点击故障点复选框，切换选中状态
  const onSelectFaultPoint = (ids: string[]): void => {
    handleSelChangeGraph(ids);
    setSideValue(0);
  };

  // 点击故障卡片，切换选中状态
  const onFaultCardClick = (id: string): void => {
    intervalId && clearInterval(intervalId);
    if (isPlayingRef.current) {
      setIsPlaying(false);
      isPlayingRef.current = false;
      topologGraphRef.current?.onRestGraph();
    }

    setActiveFaultIds((prev) => {
      const newVal = prev.includes(id)
        ? prev.filter((v) => v !== id)
        : [...prev, id];

      handleSelChangeGraph(newVal);

      return newVal;
    });

    setSideValue(0);
  };

  const onSliderChange = (value: number): void => {
    topologGraphRef.current?.onBeforePlay();

    const selItems = faultBacktrace.filter(
      (item) => dayjs(item.fault_occur_time).unix() * 1000 <= value
    );
    const selIds = selItems.map((item) => item.fault_id);

    topologGraphRef.current?.onSelectPatchActive(selIds);

    setSideValue(value);
    setActiveFaultIds(selIds);
    setTimeout(() => {
      faultTimeLineRef.current?.scrollTimeLine();
    }, 100);
  };

  useEffect(() => {
    initSidle();
  }, [faultBacktrace]);

  return (
    <Flex className={styles['layout-wrapper']} vertical={false}>
      <aside className={styles['fault-card-wrap']}>
        <FaultTimeLine
          activeFaultIds={activeFaultIds}
          onPlay={onPlay}
          onPause={onPause}
          isPlaying={isPlaying}
          onSelectFaultPoint={onSelectFaultPoint}
          onFaultCardClick={onFaultCardClick}
          ref={faultTimeLineRef}
        />
      </aside>
      <Flex className={styles['content-wrap']} vertical>
        <div className={styles['graph-content']}>
          <TopologyGraph ref={topologGraphRef} isPlaying={isPlaying} />
        </div>
        <FaultAnalysisTimeGraph
          startTime={sideStartTime}
          endTime={sideEndTime}
          value={sideValue}
          onPlay={onPlay}
          onPause={onPause}
          onSliderChange={onSliderChange}
        />
      </Flex>
    </Flex>
  );
}

export default FaultAnalysis;
