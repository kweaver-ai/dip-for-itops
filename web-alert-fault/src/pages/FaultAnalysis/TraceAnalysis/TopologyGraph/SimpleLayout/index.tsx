import React, { useEffect, useState } from 'react';
import classnames from 'classnames';
import { Tooltip, Button, Dropdown } from 'antd';
import './style.less';
import IconFont from 'Components/ARIconfont';
import {
  comboForceDefaultConfig,
  dagreDefaultConfig,
  forceDefaultConfig,
  layoutDirection,
  listLayout,
  treeListLayout,
  getComboCombinedDagreConfig,
  getComboCombinedRandomConfig
} from './utils';

// 紧凑树布局的选项
const treeList = treeListLayout.map((val) => ({
  key: val.d,
  label: val.label,
  icon: (
    <IconFont
      type={val.icon}
      style={{ fontSize: 16, transform: val.transform }}
    />
  )
}));

// 层次布局选项
const dagreList = layoutDirection.map((val) => ({
  key: val.d,
  label: val.label,
  icon: (
    <IconFont
      type={val.icon}
      style={{ fontSize: 16, transform: val.transform }}
    />
  )
}));

// 快捷布局按钮
const list = listLayout.map((val) => {
  const menus =
    val.v === 'tree' ? treeList : val.v === 'dagre' ? dagreList : undefined;

  return {
    key: val.v,
    label: val.label,
    icon: <IconFont type={val.icon} style={{ fontSize: 14 }} />,
    menus
  };
});

type SimpleLayoutType = {
  type: string;
  graphInstance: any;
  isCombo?: boolean;
  isUpdateLayout?: boolean;
  onChange?: (value: any) => void;
};

let prevType = '';

const SimpleLayout = (props: SimpleLayoutType): JSX.Element => {
  const {
    type: propType,
    graphInstance,
    isCombo = false,
    isUpdateLayout = false,
    onChange
  } = props;
  const [type, setType] = useState<string>(propType);
  const [direction, setDirection] = useState<string>();

  /** 切换选中 */
  const onSelectDropdown = (key: string) => (): void => {
    prevType = key;
  };

  const toLayout = (key: string, d?: string): void => {
    (graphInstance as any).destroyLayout();
    if (key === 'force' && isCombo) {
      graphInstance.updateLayout(comboForceDefaultConfig);
      onChange?.(comboForceDefaultConfig);
    } else if (key === 'force') {
      graphInstance.updateLayout(forceDefaultConfig);
      onChange?.(forceDefaultConfig);
    }

    if (key === 'dagre' && d) {
      const config = isCombo
        ? getComboCombinedDagreConfig(d)
        : { ...dagreDefaultConfig, rankdir: d };

      graphInstance.updateLayout(config);
      onChange?.(config);
    }
    if (key === 'random') {
      const config = isCombo
        ? getComboCombinedRandomConfig()
        : { type: 'random' };

      graphInstance.updateLayout(config);
      onChange?.(config);
    }
    setType(key);
    setTimeout(() => {
      graphInstance.fitView(20);
    });
  };

  const onClickMenu = (item: any): void => {
    setDirection(item.key);
    toLayout(prevType, item.key);
  };

  useEffect(() => {
    setType('');
    setDirection(undefined);
  }, [isUpdateLayout]);

  return (
    <div className="simpleLayoutChangeRoot">
      {list.map((item) => {
        const { key, icon, label, menus } = item;
        const iconBoxSelected = key === type;
        const borderStyle = {};

        if (menus) {
          return (
            <Dropdown
              key={key}
              className="ad-border ad-space-between"
              trigger={['click']}
              placement="bottomLeft"
              menu={{
                items: menus,
                onClick: onClickMenu,
                activeKey: direction
              }}
            >
              <div
                key={key}
                className={classnames('layoutIcon triangularSubscript', {
                  selected: type === key
                })}
                style={borderStyle}
                onClick={onSelectDropdown(key)}
              >
                <Tooltip title={label} placement="left">
                  <Button
                    className={classnames('iconBox', { iconBoxSelected })}
                  >
                    {icon}
                  </Button>
                </Tooltip>
              </div>
            </Dropdown>
          );
        }

        return (
          <div
            key={key}
            className="layoutIcon"
            style={borderStyle}
            onClick={(): void => {
              toLayout(key);
            }}
          >
            <Tooltip title={label} placement="left">
              <Button className={classnames('iconBox', { iconBoxSelected })}>
                {icon}
              </Button>
            </Tooltip>
          </div>
        );
      })}
    </div>
  );
};

export default SimpleLayout;
