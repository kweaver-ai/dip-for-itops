import React, { useMemo } from 'react';
import intl from 'react-intl-universal';
import { groupBy, omit } from 'lodash';
import { Descriptions, Tag } from 'antd';
import formatTimeDifference from 'Utils/format_date_duration';
import { dateFormat } from './utils';
import { ImpactLevel, Status } from '@/constants/commonTypes';
import { Impact_Level_Maps, Status_Maps } from '@/constants/common';
import getFaultDurationTime from '@/utils/get_fault_duration_time';

const DetialInfo: React.FC<{ dataInfo: any; isFailurePoint: boolean }> = ({
  dataInfo,
  isFailurePoint = false
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

  const getObjectInfo = () => {
    const { object_type = {}, datas = [] } = dataInfo || {};
    const { data_properties = [] } = object_type;
    const curData = datas?.[0] || {};
    const fieldsInfo = groupBy(data_properties, 'name');
    // const displayData = omit(curData, Object.keys(fieldsInfo));

    if (datas.length > 0) {
      const keys = curData?.s_id
        ? ['s_id', ...Object.keys(omit(curData, ['s_id']))]
        : Object.keys(curData);
      const newData = keys.map((key) => ({
        key,
        label: fieldsInfo?.[key]?.[0]?.display_name
          ? fieldsInfo[key][0].display_name
          : key,
        children: curData[key]
      }));

      return newData;
    }

    return [];
  };

  const getFailurePointInfo = () => {
    const newData = filterFields.map((item) => ({
      ...item,
      children: item.valueFormat?.(dataInfo?.[item.key]) || dataInfo?.[item.key]
    }));

    return newData;
  };

  const displayItems = useMemo(
    () => (isFailurePoint ? getFailurePointInfo() : getObjectInfo()),
    [dataInfo, isFailurePoint]
  );

  return <Descriptions items={displayItems} column={3} />;
};

export default DetialInfo;
