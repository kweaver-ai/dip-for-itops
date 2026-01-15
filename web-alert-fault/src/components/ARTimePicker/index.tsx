import { useState, useRef, useEffect } from 'react';
import { useBoolean } from '@noya/max';
import { Tabs, Row, Col, Button, Dropdown } from 'antd';
import intl from 'react-intl-universal';
import ARIconfont from '@/components/ARIconfont';
import TimePickerWithType, {
  getTypeByTimeRange,
  transformTimeToMoment,
  tranformMomentToDateString
} from './TimePickerWithType';
import { timeArr, timeLabel } from './constants';
import './index.less';

const { TabPane } = Tabs;

const getTimeType = (timeRangeType: string): string => {
  return timeRangeType !== 'quick' && timeRangeType !== 'now'
    ? 'range'
    : timeRangeType;
};

const getNowTimeText = (showTime: boolean): JSX.Element => {
  if (showTime) {
    return <span style={{ lineHeight: '30px' }}>{intl.get('nowTime')}</span>;
  }

  return <></>;
};

interface CustomTimePicker {
  value: [string, string];
  onChange: (value: [string, string]) => void;
  showTime?: boolean;
  getPopupContainer?: (triggerNode: HTMLElement) => HTMLElement;
  allowClear?: boolean;
  hasNow?: boolean; // 是否显示当前时间
}

const CustomTimePicker = (props: any): JSX.Element => {
  const {
    value = [],
    onChange,
    showTime = false,
    getPopupContainer = () =>
      document.getElementById('aiAlertFaultRoot') || document.body,
    allowClear = false,
    onlyShow = false,
    hasNow
  } = props;
  const [timeFilterVis, { setTrue: setFilterOpen, setFalse: setFilterClose }] =
    useBoolean(false); // 是否弹出tab页
  const timeRangeType = getTypeByTimeRange(value);
  const [mannualShowTime, setMannualShowTime] = useState(showTime);
  const [timeRange, setTimeRange] = useState<any>([]);
  const [dateType, setDateType] = useState('time');

  const dropRef = useRef() as any;
  const buttonRef = useRef() as any;
  const rootNode = document.getElementById('aiAlertFaultRoot') || document.body;

  const [activeKey, setActiveKey] = useState('quick'); // tab当前选中项

  useEffect(() => {
    if (getTimeType(timeRangeType) === 'range') {
      setTimeRange(transformTimeToMoment(value));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  useEffect(() => {
    if (getTimeType(timeRangeType) !== 'range') {
      setTimeRange([]);
      setDateType('time');
    } else {
      setDateType(
        getTimeType(timeRangeType) === 'range' ? timeRangeType : 'time'
      );
    }
    setActiveKey(getTimeType(timeRangeType));
  }, [timeRangeType]);

  useEffect(() => {
    setMannualShowTime(showTime);
  }, [showTime]);

  useEffect(() => {
    if (timeFilterVis && getTimeType(timeRangeType) === 'range') {
      setDateType(timeRangeType);
      setTimeRange(transformTimeToMoment(value));
    }

    if (timeFilterVis) {
      const rootElement = document.createElement('div');

      rootElement.setAttribute('id', 'ar-timefilter-mask');
      rootElement.classList.add('ar-timefilter-mask');

      rootElement.addEventListener('click', () => {
        setFilterClose();
      });

      rootNode.appendChild(rootElement);
    } else {
      document.getElementById('ar-timefilter-mask') &&
        rootNode.removeChild(
          document.getElementById('ar-timefilter-mask') as any
        );
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [timeFilterVis]);

  const clearValue = (e: React.MouseEvent): void => {
    e.stopPropagation();
    onChange && onChange([]);
  };

  const overlay = () => (
    <div className={'time-filter disable-drag'} key={'time'}>
      <Tabs
        animated={false}
        type="card"
        className={'drop-tabs'}
        activeKey={activeKey}
        onChange={(key): void => {
          setActiveKey(key);
        }}
      >
        {/* 快速选择 */}
        <TabPane tab={intl.get('quickTime')} key="quick">
          <Row>
            {timeArr.map((item, index) => (
              <Col
                className={`section${index} section-wrap`}
                key={item[index][0]}
                span={8}
              >
                <ul>
                  {item.map((unit) => (
                    <li key={unit[0]}>
                      <span
                        className={value[0] === unit[0] ? 'tag-active' : ''}
                        onClick={(e): void => {
                          e.stopPropagation();
                          onChange && onChange(unit);
                          setFilterClose();
                        }}
                      >
                        {intl.get(timeLabel[unit[0]])}
                      </span>
                    </li>
                  ))}
                </ul>
              </Col>
            ))}
          </Row>
        </TabPane>
        {/* 时间段选择 */}
        <TabPane tab={intl.get('timeRange')} key="range">
          <div className="ar-time-range-wrapper">
            <TimePickerWithType
              value={timeRange}
              showStyle="edit"
              onChange={(val): void => {
                setTimeRange(val);
                if (val && val.length) {
                  onChange &&
                    onChange(tranformMomentToDateString(val, dateType));
                  setFilterClose();
                }
              }}
              type={dateType}
              onChangeType={(val): void => {
                setDateType(val);
                setFilterOpen();
              }}
              isClearOnChange
              getPopupContainer={getPopupContainer}
              onOpenChange={(v): void => {
                if (v) {
                  setFilterOpen();
                }
              }}
            />
          </div>
        </TabPane>
        {/* {allowClear && !hasNow ? (
          <></>
        ) : (
          <TabPane tab={intl.get('nowTime')} key="now">
            <Col className={'ar-now-time-wrapper'} span={8}>
              <ul>
                <li>
                  <span
                    className={timeRangeType === 'now' ? 'tag-active' : ''}
                    onClick={(e): void => {
                      e.stopPropagation();
                      onChange && onChange(['', 'now']);
                      setMannualShowTime(true);
                      setFilterClose();
                    }}
                  >
                    {intl.get('nowTime')}
                  </span>
                </li>
              </ul>
            </Col>
          </TabPane>
        )} */}
      </Tabs>
    </div>
  );

  const panelText = !value[0]
    ? ''
    : timeRangeType === 'quick'
    ? `${intl.get(timeLabel[value[0]])}`
    : timeRangeType === 'week'
    ? `${value[0].replace('week', intl.get('theWeek'))} - ${value[1].replace(
        'week',
        intl.get('theWeek')
      )}`
    : `${value[0]} - ${value[1]}`;

  return (
    <>
      <div
        ref={dropRef}
        id="download-timeRange"
        onKeyDown={(): void => {
          buttonRef.current && buttonRef.current.blur();
        }}
      >
        <Dropdown
          open={timeFilterVis}
          popupRender={overlay}
          getPopupContainer={getPopupContainer}
          destroyOnHidden
          placement="bottom"
        >
          {timeRangeType === 'now' && !allowClear ? (
            <span
              onClick={(): void => {
                timeFilterVis ? setFilterClose() : setFilterOpen();
              }}
              className="ar-now-time"
            >
              {getNowTimeText(mannualShowTime)}
            </span>
          ) : onlyShow ? (
            <span title={panelText}>{panelText}</span>
          ) : (
            <Button
              onClick={(e): void => {
                e.stopPropagation();
                e.preventDefault();
                timeFilterVis ? setFilterClose() : setFilterOpen();
              }}
              className={`ar-dashboard-time-picker-quick disable-drag ${
                allowClear ? 'ar-allow-clear' : ''
              }`}
              ref={buttonRef}
            >
              <span className="ar-panel-text" title={panelText}>
                {timeRangeType === 'now' ? intl.get('nowTime') : panelText}
              </span>
              <span className="ar-icon-container">
                <ARIconfont type="icon-a-05" className="disable-drag" />
                {allowClear && (
                  <span className="clear-icon" onClick={clearValue}>
                    <svg
                      viewBox="64 64 896 896"
                      focusable="false"
                      data-icon="close-circle"
                      width="1em"
                      height="1em"
                      fill="currentColor"
                      aria-hidden="true"
                    >
                      <path d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm165.4 618.2l-66-.3L512 563.4l-99.3 118.4-66.1.3c-4.4 0-8-3.5-8-8 0-1.9.7-3.7 1.9-5.2l130.1-155L340.5 359a8.32 8.32 0 01-1.9-5.2c0-4.4 3.6-8 8-8l66.1.3L512 464.6l99.3-118.4 66-.3c4.4 0 8 3.5 8 8 0 1.9-.7 3.7-1.9 5.2L553.5 514l130 155c1.2 1.5 1.9 3.3 1.9 5.2 0 4.4-3.6 8-8 8z"></path>
                    </svg>
                  </span>
                )}
              </span>
            </Button>
          )}
        </Dropdown>
      </div>
    </>
  );
};

export default CustomTimePicker;
