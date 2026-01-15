import { highColor, initColor } from '.';

/**
 * 自定义集合（失效点）
 */
export const customFailurePointCombo = (cfg:any, group:any): void => {
  const { padding, size } = group.get('comboConfig');
  console.log(cfg, group.get('comboConfig'), 'customFailurePointCombo');
  // 主体背景
  group.addShape('circle', {
    attrs: {
      x: -size[0] / 2 + padding, // 左侧内边距
      y: -size[1] / 2 + padding, // 上侧内边距
      r: 10,
      backgroundType: 'circle',
      fill: initColor,
      stroke: '#ffffff',
      lineWidth: 0
    },
    name: 'combo-circle-shape'
  });

  // 主体icon
  group.addShape('image', {
    attrs: {
      x: -size[0] / 2 + padding - 5, // 左侧内边距
      y: -size[1] / 2 + padding - 5, // 上侧内边距
      img: cfg.imgIcon,
      width: 16,
      height: 16
    },
    name: 'combo-image-shape',
    draggable: true
  });
};
