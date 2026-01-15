const nodeModel = {
  id: 'wangguan',
  label: '网关',
  status: 'normal',
  deviceType: 'PhysicalMachine'
};

const edgeModel = {
  source: 'wangguan',
  target: 'duankou6'
};

export const getNodeAndEdge = (key: any) => {
  const nodes = [];
  const edges = [];

  new Array(10).fill(0).forEach((_, index) => {
    nodes.push({
      ...nodeModel,
      id: `wangguan-${index}`,
      label: `${nodeModel.label}-${index}`,
      parentAssociateId: key
    });

    edges.push({
      target: `${edgeModel.source}-${index}`,
      source: `${key}`,
      parentAssociateId: key
    });
  });

  return {
    nodes,
    edges
  };
};
