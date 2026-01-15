import G6 from '@antv/g6';
import { comboForceDefaultConfig } from '../SimpleLayout/utils';

const defaultOptions = {
  layout: comboForceDefaultConfig,
  enabledStack: true,
  fitView: true,
  fitViewPadding: 80,
  fitCenter: true,
  maxZoom: 1.6,
  groupByTypes: false,
  animate: false,
  defaultNode: {
    type: 'overview-node'
  },
  defaultEdge: {
    type: 'custom-edge',
    // type: 'line',
    sourceOffset: [0, 10],
    style: {
      stroke: '#BFBFBF',
      fill: '#BFBFBF',
      lineWidth: 2,
      endArrow: {
        path: G6.Arrow.triangle(6, 12, 4),
        d: 4,
        fill: '#BFBFBF'
      }
    }
  },
  edgeStateStyles: {
    selected: {
      sourceOffset: [0, 10],
      stroke: '#126EE3',
      fill: '#126EE3',
      shadowOffsetX: 0,
      shadowOffsetY: 0,
      shadowBlur: 0,
      endArrow: {
        path: G6.Arrow.triangle(6, 12, 4),
        d: 4,
        stroke: '#126EE3',
        fill: '#126EE3',
        shadowOffsetX: 0,
        shadowOffsetY: 0,
        shadowBlur: 0
      }
    },
    hover: {
      stroke: '#126EE3'
    }
  },
  defaultCombo: {
    type: 'circle',
    fill: '#f8f8f8',
    labelCfg: {
      refY: 4,
      style: {
        fill: '#333'
      }
    }
  },
  modes: {
    default: [
      'drag-canvas',
      {
        type: 'drag-node',
        onlyChangeComboSize: true
      },
      {
        type: 'drag-combo',
        onlyChangeComboSize: true
      },
      'zoom-canvas'
    ],
    edit: ['click-select']
  }
};

export default defaultOptions;
