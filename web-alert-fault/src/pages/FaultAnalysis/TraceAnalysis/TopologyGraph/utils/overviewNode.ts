import genyingaojing from 'Assets/images/genyingaojing.svg';
import intl from 'react-intl-universal';
import fitTextToWidth from './fitTextToWidth';
import { highColor, getTypeIcon } from '.';
import { ImpactLevel } from '@/constants/commonTypes';
import { Impact_Level_Maps } from '@/constants/common';

/**
 * 自定义节点
 * 溯源拓扑节点
 */
export const customNodeDraw = (cfg: any, group: any): any => {
  const {
    label,
    style,
    object_impact_level: impactLevel = '0',
    object_class: objectClass,
    ip_address: ipAddress = [],
    name: objectName,
    relation_event_ids: relationEventIds = [],
    checked = false,
    isPlaying = false,
    active = false,
    parentAssociateId
  } = cfg;
  const { width = 44, height = 44, fill = '#bfbfbf' } = style;
  const deviceIcon = getTypeIcon(objectClass).icon;
  const failureLevel =
    Impact_Level_Maps[`${impactLevel}` as unknown as ImpactLevel];
  let shape: any;

  // 激活背景圆圈
  if (active) {
    shape = group.addShape('circle', {
      attrs: {
        x: 0,
        y: 0,
        r: 22,
        backgroundType: 'circle',
        fill: '#fff',
        stroke: highColor,
        lineWidth: 2
      },
      name: 'circle-active',
      draggable: true
    });
  }

  // 主体背景色
  const mainBodyColor = isPlaying && !active ? fill : failureLevel?.color;

  // 主体背景
  shape = group.addShape('circle', {
    attrs: {
      x: 0,
      y: 0,
      r: 18,
      backgroundType: 'circle',
      fill: mainBodyColor,
      ...style
    },
    name: 'main-body-shape',
    draggable: true
  });

  // 主体icon
  group.addShape('image', {
    attrs: {
      x: -10,
      y: -10,
      img: deviceIcon,
      width: 20,
      height: 20
    },
    name: 'image-shape',
    draggable: true
  });

  // 选中时，矩形背景
  const rectWidth = 210;
  const X = 26;

  // 矩形背景框
  if (checked && !parentAssociateId) {
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
        fill: failureLevel?.color || fill,
        ...style
      },
      name: 'copy-circle-shape',
      draggable: true
    });

    group.addShape('image', {
      attrs: {
        x: -10,
        y: -10,
        img: deviceIcon,
        width: 20,
        height: 20
      },
      name: 'copy-image-shape',
      draggable: true
    });

    // 上级对象
    group.addShape('text', {
      attrs: {
        x: X,
        y: -6,
        fontSize: 10,
        textAlign: 'left',
        text: fitTextToWidth(ipAddress[0] || '-', 130),
        fill: 'rgba(0,0,0,0.85)'
      },
      name: 'devicetName',
      draggable: true
    });
    // 对象名称
    group.addShape('text', {
      attrs: {
        x: X,
        y: 7,
        fontSize: 10,
        textAlign: 'left',
        text: fitTextToWidth(objectName, 130),
        fill: 'rgba(0,0,0,0.65)'
      },
      name: 'eventName',
      draggable: true
    });

    group.addShape('circle', {
      attrs: {
        x: X + 4,
        y: 14,
        r: 4,
        backgroundType: 'circle',
        width: 8,
        height: 8,
        fill: failureLevel?.color
      },
      name: 'failureLevelCircle',
      draggable: true
    });

    // 等级状态文本
    group.addShape('text', {
      attrs: {
        x: X + 12,
        y: 20,
        fontSize: 10,
        textAlign: 'left',
        text: intl.get(failureLevel?.name),
        fill: 'rgba(0,0,0,0.65)'
      },
      name: 'failureLevelName',
      draggable: true
    });

    // 事件数标题
    group.addShape('text', {
      attrs: {
        x: X + 100,
        y: 20,
        fontSize: 10,
        textAlign: 'left',
        text: `${intl.get('event_count')}：`,
        fill: 'rgba(0,0,0,0.45)'
      },
      name: 'eventNumTitle',
      draggable: true
    });

    // 事件数量
    group.addShape('text', {
      attrs: {
        x: X + 140,
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

  if (!checked) {
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

  return shape;
};

// 根因告警动画
export const customNodeAfterDraw = (cfg: any, group: any): any => {
  const { rootCause = false } = cfg;

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
};
