import React from 'react';
import { Flex, Tag, Tooltip } from 'antd';
import dayjs from 'dayjs';
import intl from 'react-intl-universal';
import { getTypeIcon } from '../../TopologyGraph/utils';
import styles from '../index.module.less';
import { Impact_Level_Maps, Status_Maps } from '@/constants/common';
import { ImpactLevel, Status } from '@/constants/commonTypes';
import ARIconfont from '@/components/ARIconfont';

interface FaultCardProps {
  rootCauseFaultId?: string;
  data: {
    fault_id: string;
    fault_level: string;
    fault_name: string;
    fault_occur_time: string;
    entity_object_name: string;
    entity_object_class: string;
    relation_event_ids: string[];
    fault_status: string;
  };
  cardWidth?: number;
  serviceNameWidth?: number;
  active?: boolean;
  onClick?: (faultId: string) => void;
  style?: React.CSSProperties;
}

const DATE_FORMAT = 'YYYY-MM-DD HH:mm:ss';

function FaultCard(props: FaultCardProps) {
  const {
    rootCauseFaultId,
    data,
    active = false,
    cardWidth = 242,
    serviceNameWidth = 70,
    onClick,
    style = {}
  } = props;
  const {
    fault_id: faultId,
    fault_level: faultLevel = '0',
    fault_status: faultStatus = 'expired',
    fault_name: faultName,
    fault_occur_time: faultOccurTime,
    entity_object_name: entityObjectName,
    entity_object_class: entityObjectClass,
    relation_event_ids: relationEventIds
  } = data;

  const seviceIcon = getTypeIcon(entityObjectClass).value;
  const faultLevelInfo =
    Impact_Level_Maps[`${faultLevel}` as unknown as ImpactLevel];
  const faultStatusInfo = Status_Maps[`${faultStatus}` as unknown as Status];

  return (
    <Flex
      className={`${styles['card-wrapper']} ${
        active ? styles['fault-card-active'] : ''
      }`}
      vertical
      gap={4}
      style={{
        borderLeft: `4px solid ${faultLevelInfo?.color}`,
        width: cardWidth,
        ...style
      }}
      onClick={() => onClick?.(faultId)}
      id={faultId}
    >
      <Flex justify="space-between">
        <Flex gap={4} vertical className={styles['fault-info-top']}>
          <div className={styles['fault-time']}>
            {dayjs(faultOccurTime).format(DATE_FORMAT)}
          </div>
          <Tooltip
            title={faultName}
            className={styles['break-all']}
            styles={{ container: { width: 300 } }}
          >
            <div className={styles['fault-title']}>{faultName}</div>
          </Tooltip>
        </Flex>
        <div className={styles['fault-tag']}>
          {rootCauseFaultId?.toString() === faultId?.toString() && (
            <Tag color="#8614FA" variant="filled">
              {intl.get('root_cause')}
            </Tag>
          )}
        </div>
      </Flex>
      <Flex
        className={styles['fault-info']}
        justify="space-between"
        align="center"
        gap={8}
      >
        <Flex gap={8}>
          <Tag color={faultStatusInfo.color} variant="filled">
            {intl.get(faultStatusInfo.name)}
          </Tag>
          <Flex className={styles['fault-node']} gap={4} align="center">
            <div className={styles['fault-node-icon']}>
              <ARIconfont type={seviceIcon} />
            </div>
            <Tooltip title={entityObjectName}>
              <div
                className={styles['fault-node-name']}
                style={{ width: serviceNameWidth }}
              >
                {entityObjectName}
              </div>
            </Tooltip>
          </Flex>
        </Flex>
        <Flex gap={4}>
          <span>{intl.get('event_count')}：</span>
          <span className={styles['fault-events-num']}>
            {relationEventIds?.length || 0}
          </span>
        </Flex>
      </Flex>
    </Flex>
  );
}

export default FaultCard;
