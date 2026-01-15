import { useState, ReactNode, useEffect } from 'react';
import { Row, Col, Divider, Space, Select, ConfigProvider } from 'antd';
import intl from 'react-intl-universal';
import CustomTimePicker from '@/components/ARTimePicker';
import { ProblemLevel, ProblemStatus } from '@/constants/problemTypes';
import { refreshIntervals } from './constants';
import styles from './index.module.less';
import { transformTimeToMoment } from '@/components/ARTimePicker/TimePickerWithType';
import { Level_Maps, Status_Maps } from '@/constants/common';
import { StatusesText, statusStyle } from '@/pages/AlertFault/constants';

const { Option } = Select;

export interface FaultHeaderProps {
  data: number[];
  dataType?: 'problem' | 'normal';
  extra?: ReactNode[] | ReactNode;
  showLevel?: number[];
  onTimeChange?: (value: number[], originTime?: string[]) => void;
  onRefreshChange?: (value: number) => void;
  onLevelChange?: (value: ProblemLevel[]) => void;
  onStatusChange?: (value: ProblemStatus[]) => void;
}

const FaultHeader: React.FC<FaultHeaderProps> = ({
  data,
  dataType = 'normal',
  showLevel = [1, 2, 3, 4],
  extra,
  onTimeChange,
  onRefreshChange,
  onLevelChange,
  onStatusChange
}) => {
  const [refreshInterval, setRefreshInterval] = useState(0);
  const [timeRange, setTimeRange] = useState<string[]>(['now-7d', 'now']);
  const [selectedLevelKeys, setSelectedLevelKeys] = useState<any[]>([]);
  const [selectedStatusKeys, setSelectedStatusKeys] = useState<any[]>([]);

  const onLevelSelect = (key: any) => {
    if (selectedLevelKeys.includes(key)) {
      setSelectedLevelKeys(selectedLevelKeys.filter((k) => k !== key));
      onLevelChange?.(selectedLevelKeys.filter((k) => k !== key));
    } else {
      setSelectedLevelKeys([...selectedLevelKeys, key]);
      onLevelChange?.([...selectedLevelKeys, key]);
    }
  };

  const onStatusSelect = (key: any) => {
    if (selectedStatusKeys.includes(key)) {
      setSelectedStatusKeys(selectedStatusKeys.filter((k) => k !== key));
      onStatusChange?.(selectedStatusKeys.filter((k) => k !== key));
    } else {
      setSelectedStatusKeys([...selectedStatusKeys, key]);
      onStatusChange?.([...selectedStatusKeys, key]);
    }
  };

  const getColor = (isSelect: boolean, color: string) => {
    // 将hex颜色转为rgb颜色
    const rgbColor = color.replace('#', '');
    const r = parseInt(rgbColor.substring(0, 2), 16);
    const g = parseInt(rgbColor.substring(2, 4), 16);
    const b = parseInt(rgbColor.substring(4, 6), 16);

    return isSelect ? `rgba(${r}, ${g}, ${b}, 0.3)` : 'transparent';
  };

  const levelItems = Object.keys(Level_Maps)
    .map((key) => {
      if (showLevel.includes(Number.parseInt(key))) {
        return {
          key,
          name: intl.get((Level_Maps as any)[key]?.name),
          borderColor: (Level_Maps as any)[key]?.color,
          background: getColor(
            selectedLevelKeys.includes(key),
            (Level_Maps as any)[key]?.color
          )
        };
      }

      return undefined;
    })
    .filter(Boolean);

  const getLevelItems = () => {
    return levelItems.map((item, index) => {
      return (
        <span
          className={`${styles.navItem}`}
          style={{
            borderColor: item!.borderColor,
            backgroundColor: item!.background
          }}
          key={index}
          onClick={() => onLevelSelect(item!.key)}
        >
          {item!.name}：{data[index] ?? 0}
        </span>
      );
    });
  };

  const getProblemStatusItems = () => {
    const items = Object.keys(statusStyle).map((key) => ({
      key,
      name: intl.get(StatusesText[key as keyof typeof StatusesText]),
      borderColor: statusStyle[key as keyof typeof statusStyle].color,
      background: getColor(
        selectedStatusKeys.includes(key),
        statusStyle[key as keyof typeof statusStyle].color ?? ''
      )
    }));

    return items.map((item, index) => {
      return (
        <span
          className={`${styles.navItem}`}
          style={{
            borderColor: item!.borderColor,
            backgroundColor: item!.background
          }}
          key={index}
          onClick={() => onStatusSelect(item!.key)}
        >
          {item!.name}：{data[index + levelItems.length] ?? 0}
        </span>
      );
    });
  };

  const getCommonStatusItems = () => {
    const items = Object.keys(Status_Maps).map((key) => ({
      key,
      name: intl.get(Status_Maps[key as keyof typeof Status_Maps].name),
      borderColor: Status_Maps[key as keyof typeof Status_Maps].color,
      background: getColor(
        selectedStatusKeys.includes(key),
        Status_Maps[key as keyof typeof Status_Maps].color
      )
    }));

    return items.map((item, index) => {
      return (
        <span
          className={`${styles.navItem}`}
          style={{
            borderColor: item!.borderColor,
            backgroundColor: item!.background
          }}
          key={index}
          onClick={() => onStatusSelect(item!.key)}
        >
          {item!.name}：{data[index + levelItems.length] ?? 0}
        </span>
      );
    });
  };

  // 刷新间隔变化时触发
  const onChangeRefreshInterval = (value: number) => {
    setRefreshInterval(value);
  };

  // 时间范围变化时触发
  const onChange = (value: [string, string]) => {
    setTimeRange(value);
  };

  useEffect(() => {
    const dayjsTime = transformTimeToMoment(timeRange);
    const timestampRange = dayjsTime.map((item) => item.valueOf());

    console.log(timestampRange, 'timestampRange');
    onTimeChange?.(timestampRange, timeRange);
    sessionStorage.setItem('arTimeRange', JSON.stringify(timeRange));
  }, [timeRange]);

  useEffect(() => {
    onRefreshChange?.(refreshInterval);
    sessionStorage.setItem(
      'arRefreshInterval',
      JSON.stringify(refreshInterval)
    );
  }, [refreshInterval]);

  return (
    <Row className={styles.topNav} justify="space-between" align="middle">
      <Col>
        <div className={styles.navItems}>
          {getLevelItems()}
          <Divider vertical className={styles.divider} />
          {dataType === 'problem'
            ? getProblemStatusItems()
            : getCommonStatusItems()}
        </div>
      </Col>
      <Col>
        <ConfigProvider
          theme={{
            token: {
              fontSize: 12
            },
            components: {
              Select: {
                optionFontSize: 12
              }
            }
          }}
        >
          <Space className={styles.chartToolbar}>
            <CustomTimePicker
              getPopupContainer={(): any => {
                return (
                  document.getElementById('aiAlertFaultRoot') || document.body
                );
              }}
              value={timeRange}
              onChange={onChange}
            />
            <Select
              classNames={{
                popup: { root: styles['refresh-dropdown'] }
              }}
              onChange={onChangeRefreshInterval}
              value={refreshInterval}
            >
              {refreshIntervals.map(({ label, value }) => (
                <Option value={value} key={value}>
                  {intl.get(label)}
                </Option>
              ))}
            </Select>
            {extra && Array.isArray(extra) ? extra.map((item) => item) : extra}
          </Space>
        </ConfigProvider>
      </Col>
    </Row>
  );
};

export default FaultHeader;
