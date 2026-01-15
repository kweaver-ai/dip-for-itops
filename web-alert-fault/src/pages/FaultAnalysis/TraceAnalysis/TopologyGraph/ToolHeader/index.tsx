import { useContext } from 'react';
import { Flex, Switch } from 'antd';
import type { IGraph } from '@antv/g6';
import { useFullscreen } from 'ahooks';
import intl from 'react-intl-universal';
import ARIconfont5 from 'Components/ARIconfont5';
import styles from './index.module.less';
import { faultAnalysisContext } from '@/pages/FaultAnalysis';

interface TProps {
  graphInstance: IGraph;
  fullDom: any;
  originalWidth: number;
  originalHeight: number;
  isShowLegend: boolean;
  setIsShowLegend: (val: boolean) => void;
  setIsShowFailurePoint: (val: boolean) => void;
  isShowFailurePoint: boolean;
  isPlaying: boolean;
}

const ToolHeader = (props: TProps) => {
  const {
    graphInstance,
    fullDom,
    originalWidth,
    originalHeight,
    isShowLegend,
    setIsShowLegend,
    setIsShowFailurePoint,
    isShowFailurePoint,
    isPlaying
  } = props;
  const [isFullscreen, { toggleFullscreen }] = useFullscreen(fullDom);

  const {
      setIsGraphFullScreen
    } = useContext(faultAnalysisContext);

  const handleGraphInstance = (
    val: 'enlarge' | 'lower' | 'location' | 'adaptive' | 'full' | 'cancelFull'
  ): void => {
    if (!graphInstance) return;
    if (isPlaying) return;

    const graphSnapInstanceContainer: any =
      document.getElementById('graph-content');
    const width = graphSnapInstanceContainer?.offsetWidth || 0;
    const height = graphSnapInstanceContainer?.offsetHeight || 0;

    if (val === 'enlarge') {
      graphInstance.zoom(1.1, { x: width / 2, y: height / 2 });
    }
    if (val === 'lower') {
      graphInstance.zoom(0.9, { x: width / 2, y: height / 2 });
    }
    if (val === 'location') {
      graphInstance.changeSize(width, height);
      graphInstance.fitView(20);
    }
    if (val === 'adaptive') {
      graphInstance.zoomTo(1, { x: width / 2, y: height / 2 });
    }
    if (val === 'full') {
      setIsGraphFullScreen(true);
      toggleFullscreen();
      setTimeout(() => {
        const graphSnapInstanceContainer: any =
          document.getElementById('graph-content');
        const width = graphSnapInstanceContainer?.offsetWidth || 0;
        const height = graphSnapInstanceContainer?.offsetHeight || 0;

        graphInstance.changeSize(width, height);
      }, 50);
    }

    if (val === 'cancelFull') {
      setIsGraphFullScreen(false);
      toggleFullscreen();
      graphInstance.changeSize(originalWidth, originalHeight);
    }
  };

  return (
    <Flex
      className={styles['layout-wrapper']}
      vertical={false}
      justify="space-between"
    >
      <Flex className={styles['left-content']} gap={12} align="center">
        <div
          onClick={(): void => handleGraphInstance('enlarge')}
          className={styles['icon-wrapper']}
        >
          <ARIconfont5 type="icon-quanjingtu-fangda" />
        </div>
        <div
          onClick={(): void => handleGraphInstance('lower')}
          className={styles['icon-wrapper']}
        >
          <ARIconfont5 type="icon-quanjingtu-suoxiao" />
        </div>
        <div
          onClick={(): void => handleGraphInstance('location')}
          className={styles['icon-wrapper']}
        >
          <ARIconfont5 type="icon-dingwei" />
        </div>
        <div
          onClick={(): void => handleGraphInstance('adaptive')}
          className={styles['icon-wrapper']}
        >
          <ARIconfont5 type="icon-zishiying" />
        </div>
        {!isFullscreen ? (
          <div
            onClick={(): void => handleGraphInstance('full')}
            className={styles['icon-wrapper']}
          >
            <ARIconfont5 type="icon-quanping" />
          </div>
        ) : (
          <div
            onClick={(): void => handleGraphInstance('cancelFull')}
            className={styles['icon-wrapper']}
          >
            <ARIconfont5 type="icon-quxiaoquanping" />
          </div>
        )}
      </Flex>
      <Flex
        className={styles['right-content']}
        gap={12}
        align="center"
        justify="flex-end"
      >
        <Flex align="center" gap={4}>
          <Switch
            size="small"
            checked={isShowFailurePoint}
            onChange={setIsShowFailurePoint}
            disabled={isPlaying}
          />
          {intl.get('fault_points')}
        </Flex>
        <Flex align="center" gap={4}>
          <Switch
            size="small"
            checked={isShowLegend}
            onChange={setIsShowLegend}
            disabled={isPlaying}
          />
          {intl.get('graph_legend')}
        </Flex>
      </Flex>
    </Flex>
  );
};

export default ToolHeader;
