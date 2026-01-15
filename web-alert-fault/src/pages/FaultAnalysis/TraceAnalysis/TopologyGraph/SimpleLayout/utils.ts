/* eslint-disable new-cap */
import G6 from '@antv/g6';

// 布局方向
const TB = 'TB'; // 从上到下, 根节点在上，往下布局
const BT = 'BT'; // 从下到上, 根节点在下，往上布局
const LR = 'LR'; // 从左往右, 根节点在左，往右布局
const RL = 'RL'; // 从右往左, 根节点在右，往左布局

export const layoutDirection = [
  { d: TB, label: '从上到下', transform: '', icon: 'icon-shuzhuangbuju' },
  {
    d: BT,
    label: '从下到上',
    transform: 'rotate(180deg)',
    icon: 'icon-shuzhuangbuju'
  },
  {
    d: LR,
    label: '从左往右',
    transform: 'rotate(270deg)',
    icon: 'icon-shuzhuangbuju'
  },
  {
    d: RL,
    label: '从右往左',
    transform: 'rotate(90deg)',
    icon: 'icon-shuzhuangbuju'
  }
];

// 对齐方式
// const MM = 'MM'; // 中间对齐
// const UL = 'UL'; // 对齐到左上角
// const UR = 'UR'; // 对齐到右上角
// const DL = 'DL'; // 对齐到左下角
// const DR = 'DR'; // 对齐到右下角

// 力导布局默认值
export const forceDefaultConfig = {
  type: 'force',
  linkDistance: 150,
  nodesep: 50,
  preventOverlap: true,
  edgeStrength: 0.05,
  nodeStrength: 40,
  nodeSpacing: 50
};

// combo力导布局默认值
export const comboForceDefaultConfig = {
  type: 'comboForce',
  linkDistance: 250,
  preventOverlap: true,
  edgeStrength: 0.5,
  nodeStrength: 10,
  nodeSpacing: 30,
  comboSpacing: 250
};

// 组合布局默认值
export const comboCombinedDefaultConfig = {
  type: 'comboCombined',
  nodeSize: [50, 50],
  spacing: 30,
  outerLayout: new G6.Layout.dagre({
    rankdir: LR,
    sortByCombo: true,
    nodesep: 50,
    ranksep: 200
  }),
  innerLayout: new G6.Layout.force2({
    linkDistance: 100,
    nodeStrength: 25,
    edgeStrength: 0,
    nodeSize: 50,
    nodeSpacing: 50,
    preventOverlap: true
  })
};

// 组合布局配置-层次布局
export const getComboCombinedDagreConfig = (rankdir: string) => ({
  ...comboCombinedDefaultConfig,
  outerLayout: new G6.Layout.dagre({
    rankdir,
    sortByCombo: true,
    nodesep: 50,
    ranksep: 120
  })
});

// 组合布局配置-随机布局
export const getComboCombinedRandomConfig = () => ({
  ...comboCombinedDefaultConfig,
  outerLayout: new G6.Layout.random({
	width: 1500,
	height: 1800
  })
});

// 层次布局默认值
export const dagreDefaultConfig = {
  type: 'dagre',
  rankdir: LR,
  nodesep: 50,
  ranksep: 80
};

// 自由布局默认值
export const randomDefaultConfig = {
  type: 'dagre',
  rankdir: LR,
  nodesep: 50,
  ranksep: 80
};

// 紧凑树布局默认值
export const treeDefaultConfig = {
  rankdir: LR,
  hGap: 80,
  vGap: 80,
  limit: 15,
  isGroup: true
};

// 紧凑树布局
const H = 'H'; // 水平对称, 根节点在中间，水平对称布局
const V = 'V'; // 垂直对称, 根节点在中间，垂直对称布局

export const treeListLayout = [
  {
    d: LR,
    label: '从左往右',
    transform: 'rotate(270deg)',
    icon: 'icon-shuzhuangbuju'
  },
  {
    d: RL,
    label: '从右往左',
    transform: 'rotate(90deg)',
    icon: 'icon-shuzhuangbuju'
  },
  { d: TB, label: '从上到下', transform: '', icon: 'icon-shuzhuangbuju' },
  {
    d: BT,
    label: '从下到上',
    transform: 'rotate(180deg)',
    icon: 'icon-shuzhuangbuju'
  },
  { d: H, label: '水平对称', icon: 'icon-shuipingduicheng' },
  { d: V, label: '垂直对称', icon: 'icon-chuizhiduicheng' }
];

const FREE = 'random'; // 自由布局
const TREE = 'tree'; // 层次布局
const DAGRE = 'dagre'; // 层次布局
const FORCE = 'force'; // 力导布局

export const listLayout = [
  { v: FREE, label: '自由布局', icon: 'icon-guanxibuju' },
  { v: FORCE, label: '力导布局', icon: 'icon-lidaobuju' },
  { v: DAGRE, label: '层次布局', icon: 'icon-cengcibuju' }
  // { v: TREE, icon: 'icon-shuzhuangbuju' }
];
