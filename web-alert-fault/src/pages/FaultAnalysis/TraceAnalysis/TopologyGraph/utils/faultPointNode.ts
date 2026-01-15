import intl from 'react-intl-universal';
import dayjs from 'dayjs';
import zhengqueIcon from 'Assets/images/icon-zhengque.svg';
import shixiaoIcon from 'Assets/images/icon-shixiao.svg';
import fashengIcon from 'Assets/images/icon-fasheng.svg';
import genyingaojing from 'Assets/images/genyingaojing.svg';
import { highColor } from '.';
import { ImpactLevel } from '@/constants/commonTypes';
import { Impact_Level_Maps } from '@/constants/common';

// 故障状态图标列表
const faultStatusIcons = {
  1: fashengIcon,
  2: zhengqueIcon,
  3: shixiaoIcon
};

/**
 * 自定义节点
 * 溯源拓扑节点
 */
export const faultPointDraw = (cfg: any, group: any): any => {
  const {
    label,
    style,
    fault_level: level = '0',
    fault_status: faultStatus = '3',
    fault_occur_time: faultOccurTime,
    fault_recovery_time: faultRecoveryTime,
    relation_event_ids: relationEventIds = [],
    checked = false,
    isPlaying = false,
    active = false,
    rootCause
  } = cfg;
  const { width = 44, height = 44, fill = '#BFBFBF' } = style;
  const failureLevel = Impact_Level_Maps[`${level}` as unknown as ImpactLevel];
  let activeStyle = {};

  // 激活背景圆圈
  if (active) {
    activeStyle = { stroke: highColor, lineWidth: 2 };
  }

  // 主体背景色
  const mainBodyColor = isPlaying && !active ? fill : failureLevel?.color;
  // 主体背景
  const shape = group.addShape('circle', {
    attrs: {
      x: 0,
      y: 0,
      r: 18,
      backgroundType: 'circle',
      fill: mainBodyColor,
      ...activeStyle,
      ...style
    },
    name: 'circle-shape'
  });

  // 选中时，矩形背景
  const rectWidth = 210;
  const X = 26;

  // 矩形背景框
  if (checked) {
    group.addShape('rect', {
      attrs: {
        x: -22,
        y: -22,
        width: rectWidth,
        height: 44,
        radius: [22, 4, 4, 22],
        fill: '#f2f2f2',
        stroke: highColor,
        lineWidth: 1
      },
      name: 'rect-shape4',
      draggable: true
    });

    group.addShape('circle', {
      attrs: {
        x: 0,
        y: 0,
        r: 18,
        backgroundType: 'circle',
        width,
        height,
        fill: mainBodyColor,
        ...style
      },
      name: 'copy-circle-shape'
    });

    group.addShape('image', {
      attrs: {
        x: -10,
        y: -10,
        img: faultStatusIcons[faultStatus as keyof typeof faultStatusIcons],
        width: 20,
        height: 20
      },
      name: 'copy-image-shape',
      draggable: true
    });

    // 开始事件
    group.addShape('text', {
      attrs: {
        x: X,
        y: -6,
        fontSize: 10,
        textAlign: 'left',
        text: `${intl.get('start_time')}：${
          faultOccurTime
            ? dayjs(faultOccurTime).format('YYYY-MM-DD HH:mm:ss')
            : '-'
        }`,
        fill: 'rgba(0,0,0,0.65)'
      },
      name: 'startTime',
      draggable: true
    });

    // 结束时间
    group.addShape('text', {
      attrs: {
        x: X,
        y: 7,
        fontSize: 10,
        textAlign: 'left',
        text: `${intl.get('end_time')}：${
          faultRecoveryTime
            ? dayjs(faultRecoveryTime).format('YYYY-MM-DD HH:mm:ss')
            : '-'
        }`,
        fill: 'rgba(0,0,0,0.65)'
      },
      name: 'endTime',
      draggable: true
    });

    // 事件数标题
    group.addShape('text', {
      attrs: {
        x: X,
        y: 20,
        fontSize: 10,
        textAlign: 'left',
        text: `${intl.get('event_count')}：`,
        fill: 'rgba(0,0,0,0.65)'
      },
      name: 'eventNumTitle',
      draggable: true
    });

    // 事件数量
    group.addShape('text', {
      attrs: {
        x: X + 40,
        y: 20,
        fontSize: 10,
        textAlign: 'left',
        text: `${relationEventIds.length}`,
        fill: '#126EE3'
      },
      name: 'eventNum',
      draggable: true
    });
  }

  if (!cfg.checked) {
    group.addShape('text', {
      attrs: {
        x: 0,
        y: 32,
        fontSize: 12,
        textAlign: 'center',
        text: label || '-',
        fill: '#333'
      },
      name: 'label',
      draggable: true
    });
  }

  if (rootCause) {
    const image = group.addShape('image', {
      attrs: {
        x: -15,
        y: -18,
        img: genyingaojing,
        width: 30,
        height: 30
      },
      name: 'genyingaojing',
      draggable: true
    });

    image.animate(
      {
        opacity: 0
      },
      {
        repeat: true,
        duration: 2000,
        easing: 'easeCubic'
      }
    );
  }

  return shape;
};
