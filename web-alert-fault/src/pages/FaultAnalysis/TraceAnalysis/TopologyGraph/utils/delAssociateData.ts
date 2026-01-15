// 去掉关联数据
const delAssociateData = (data: any[]) =>
  data.filter(
    ({ parentAssociateId }: { parentAssociateId?: string }) =>
      !parentAssociateId
  );

export default delAssociateData;
