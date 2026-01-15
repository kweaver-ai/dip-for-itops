import React from 'react';
import intl from 'react-intl-universal';
import { Descriptions, Tag } from 'antd';
import { dateFormat } from './utils';
import { ImpactLevel, Status } from '@/constants/commonTypes';
import { Impact_Level_Maps, Status_Maps } from '@/constants/common';
import getFaultDurationTime from '@/utils/get_fault_duration_time';

const DetialInfo: React.FC<{ dataInfo: any; }> = ({
  dataInfo,
}) => {
  const filterFields = [
    { key: 'fault_id', label: intl.get('fault_id') },
    { key: 'fault_name', label: intl.get('fault_name') },
    {
      key: 'fault_occur_time',
      label: intl.get('fault_point_start_time'),
      valueFormat: dateFormat
    },
    {
      key: 'fault_recovery_time',
      label: intl.get('fault_point_end_time'),
      valueFormat: dateFormat
    },
    {
      key: 'fault_duration_time',
      label: intl.get('fault_point_duration'),
      valueFormat: () => getFaultDurationTime(dataInfo)
    },
    {
      key: 'fault_level',
      label: intl.get('level'),
      valueFormat: (value: string) =>
        value !== undefined ? (
          <Tag
            color={Impact_Level_Maps[value as unknown as ImpactLevel]?.color}
            variant="filled"
          >
            {intl.get(Impact_Level_Maps[value as unknown as ImpactLevel]?.name)}
          </Tag>
        ) : (
          ''
        )
    },
    {
      key: 'fault_status',
      label: intl.get('status'),
      valueFormat: (value: string) =>
        value !== undefined ? (
          <Tag
            color={Status_Maps[value as unknown as Status].color}
            variant="filled"
          >
            {intl.get(Status_Maps[value as unknown as Status].name)}
          </Tag>
        ) : (
          ''
        )
    },
    { key: 'entity_object_id', label: intl.get('entity_object_id') },
    { key: 'entity_object_name', label: intl.get('entity_object_name') },
    { key: 'entity_object_class', label: intl.get('entity_object_class') },
    { key: 'fault_description', label: intl.get('fault_description') }
  ];

  const getFailurePointInfo = () => {
    const newData = filterFields.map((item) => ({
      ...item,
      children: item.valueFormat?.(dataInfo?.[item.key]) || dataInfo?.[item.key]
    }));

    return newData;
  };

  const displayItems = getFailurePointInfo();

  return <Descriptions items={displayItems} column={3} />;
};

export default DetialInfo;
