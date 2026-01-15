import G6 from '@antv/g6';

// 自定义边 (直线)
export const customEdgeDraw = (cfg, group): any => {
  const { checked, isPlaying = false, active = false } = cfg;
  const color = checked || active ? '#126EE3' : '#bfbfbf';
  const { startPoint, endPoint } = cfg;

  const shape = group.addShape('path', {
    attrs: {
      path: [
        ['M', startPoint.x, startPoint.y],
        ['L', endPoint.x, endPoint.y]
      ],
      fill: color,
      stroke: color,
      // lineDash: [12,2],
      lineWidth: 2,
      curveOffset: cfg.curveOffset,
      endArrow: {
        path: G6.Arrow.triangle(6, 12, 4),
        d: 4,
        fill: color
      }
    }
  });

  if (active) {
    let index = 0;
    const lineDash = [18, 3];
    const animateEle = group.get('children')[0];

    animateEle.animate(
      () => {
        index++;
        if (index > 9) {
          index = 0;
        }
        const res = {
          lineDash,
          lineDashOffset: -index
        };

        return res;
      },
      {
        repeat: true,
        duration: 1000
      }
    );
  }

  return shape;
};

export const customEdge = (cfg, group): any => {
  let index = 0;
  const lineDash = [6, 4, 6];

  if (!cfg.active) {
    return;
  }

  const animateEle = group.get('children')[0];

  animateEle.animate(
    () => {
      index++;
      if (index > 9) {
        index = 0;
      }
      const res = {
        lineDash,
        lineDashOffset: -index
      };

      return res;
    },
    {
      repeat: true,
      duration: 6000000
    }
  );
};
