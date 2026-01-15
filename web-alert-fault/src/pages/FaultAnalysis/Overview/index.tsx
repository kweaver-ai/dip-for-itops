import {
  Flex,
  Row,
  Col,
  Collapse,
  theme,
  Divider,
  Typography,
  Tooltip,
  Tag
} from 'antd';
import { useContext, useEffect, useMemo, useState } from 'react';
import classnames from 'classnames';
import { CaretRightOutlined } from '@ant-design/icons';
import intl from 'react-intl-universal';
import dayjs from 'dayjs';
import { groupBy } from 'lodash';
import { useBoolean } from 'ahooks';
import { faultAnalysisContext } from '../index';
import styles from './index.module.less';
import ImpactGraph from './ImpactGraph';
import FaultDetialDrawer from './FaultDetialDrawer';
import { Level_Maps } from '@/constants/common';
import { Level } from '@/constants/commonTypes';
import { ProblemStatus } from '@/constants/problemTypes';
import { StatusesText, statusStyle } from '@/pages/AlertFault/constants';
import ARIconfont from '@/components/ARIconfont';
import FaultCard from '@/pages/FaultAnalysis/TraceAnalysis/FaultTimeLine/FaultCard';
import { getTypeIcon } from '@/pages/FaultAnalysis/TraceAnalysis/TopologyGraph/utils/index';
import getProblemDurationTime from '@/utils/get_problem_duration_time';

const { Text, Paragraph } = Typography;
const DateFormat = 'YYYY-MM-DD HH:mm:ss';

const Overview = () => {
  const {
    problemData: {
      problem_id: problemId,
      problem_name: problemName,
      problem_level: problemLevel = '0',
      problem_status: problemStatus = '0',
      problem_occur_time: problemOccurTime,
      problem_close_time: problemCloseTime,
      root_cause_object_id: rootCauseObjectId,
      root_cause_fault_id: rootCauseFaultId,
      rca_results: {
        rca_context: {
          occurrence: { description, impact },
          backtrace = [],
          network = {
            nodes: []
          }
        }
      }
    }
  } = useContext(faultAnalysisContext);

  const [rootCauseFault, setRootCauseFault] = useState<any>({});
  const [rootCauseObj, setRootCauseObj] = useState<any>({});
  const [drawerDetialId, setDrawerDetialId] = useState<string>('');
  const [
    drawerVisible,
    { setTrue: openDetailDrawer, setFalse: closeDetailDrawer }
  ] = useBoolean(false);

  const { token } = theme.useToken();

  const getServiceTree = (nodes: Node[]) => {
    const data = groupBy(nodes, 'object_class');
    const serviceTree = Object.keys(data).map((key) => {
      const iconType = getTypeIcon(key);

      return {
        key,
        label: `${iconType.name}（${data[key].length}）`,
        children: data[key].map((item) => (
          <div
            className={styles.equipmentItem}
            title={item.name}
            key={item.s_id}
          >
            <span>
              <ARIconfont
                type={iconType.value}
                className={styles.monitorIcon}
              />
              {item.name}
            </span>
          </div>
        ))
      };
    });

    return serviceTree;
  };

  const serviceTreeData = useMemo(
    () => getServiceTree(network.nodes || []),
    [network.nodes]
  );

  const onFaultDetailClick = (id: string) => {
    setDrawerDetialId(id);
    openDetailDrawer();
  };

  useEffect(() => {
    const rootCauseFault =
      backtrace?.find(
        (item) => item.fault_id?.toString() === rootCauseFaultId?.toString()
      ) || {};

    const rootCauseObj =
      network.nodes?.find((item) => item.s_id === rootCauseObjectId) || {};

    setRootCauseFault(rootCauseFault);
    setRootCauseObj(rootCauseObj);
  }, [backtrace, network.nodes, rootCauseFaultId, rootCauseObjectId]);

  return (
    <Flex className={styles.overviewContainer} gap={16}>
      <Flex vertical flex={1} gap={16} style={{ overflow: 'hidden' }}>
        <div className={classnames(styles.card, styles.cardHeight)}>
          <div className={styles.cardHeader}>
            {intl.get('basic_information')}
          </div>
          <div className={styles.content}>
            <Row className={styles.row}>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('problem_name')}：
                  </div>
                  <div className={styles.infoItemWrap}>
                    <Tooltip title={problemName} placement="topRight">
                      <span className={styles.textWrap}>
                        <span className={styles.infoContent}>
                          {problemName}
                        </span>
                      </span>
                    </Tooltip>
                  </div>
                </Flex>
              </Col>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('problem_id')}：
                  </div>
                  <div>{problemId}</div>
                </Flex>
              </Col>
            </Row>
            <Row className={styles.row}>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('problem')}
                    {intl.get('level')}：
                  </div>
                  <div>
                    <span>
                      <span
                        className={styles.levelIcon}
                        style={{
                          backgroundColor:
                            Level_Maps[problemLevel as unknown as Level]?.color
                        }}
                      />
                      {intl.get(
                        Level_Maps[problemLevel as unknown as Level]?.name
                      )}
                    </span>
                  </div>
                </Flex>
              </Col>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('problem')}
                    {intl.get('status')}：
                  </div>
                  <div>
                    <Tag
                      color={
                        statusStyle[problemStatus as unknown as ProblemStatus]
                          .color
                      }
                      variant="filled"
                    >
                      {intl.get(
                        StatusesText[problemStatus as unknown as ProblemStatus]
                      )}
                    </Tag>
                  </div>
                </Flex>
              </Col>
            </Row>
            <Row className={styles.row}>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('occur_time')}：
                  </div>
                  <div>{dayjs(problemOccurTime).format(DateFormat)}</div>
                </Flex>
              </Col>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('close_time')}：
                  </div>
                  <div>
                    {problemStatus === ProblemStatus.Closed
                      ? dayjs(problemCloseTime).format('YYYY-MM-DD HH:mm:ss')
                      : '--'}
                  </div>
                </Flex>
              </Col>
            </Row>
            <Row className={styles.row}>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('duration')}：
                  </div>
                  <div>
                    <Tag color="#1677ff" variant="filled">
                      {getProblemDurationTime({
                        problem_status: problemStatus,
                        problem_occur_time: problemOccurTime,
                        problem_close_time: problemCloseTime
                      })}
                    </Tag>
                  </div>
                </Flex>
              </Col>
              <Col span={12}>
                {/* <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('close_time')}：
                  </div>
                  <div>
                    {problemStatus === ProblemStatus.Closed
                      ? dayjs(problemCloseTime).format(DateFormat)
                      : '--'}
                  </div>
                </Flex> */}
              </Col>
            </Row>
            <Row className={styles.row}>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('root_cause_objects')}：
                  </div>
                  <div>
                    <Tag color="#126ee3" variant="filled">
                      {rootCauseObj?.name}
                    </Tag>
                  </div>
                </Flex>
              </Col>
              <Col span={12}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('root_cause_fault')}：
                  </div>
                  <div className={styles.infoItemWrap}>
                    <Tooltip
                      title={rootCauseFault?.fault_name}
                      placement="topRight"
                    >
                      <span className={styles.textWrap}>
                        <span className={styles.infoContent}>
                          {rootCauseFault?.fault_name}
                        </span>
                      </span>
                    </Tooltip>
                  </div>
                </Flex>
              </Col>
            </Row>
            <Row className={styles.row}>
              <Col span={24}>
                <Flex className={styles.itemWrap}>
                  <div className={styles.itemTitle}>
                    {intl.get('root_cause_fault_points')}：
                  </div>
                  <div
                    className={classnames(
                      styles.infoItemWrap,
                      styles.linkColor
                    )}
                  >
                    <Tooltip
                      title={rootCauseFault?.fault_description}
                      placement="topRight"
                    >
                      <span className={styles.textWrap}>
                        <span className={styles.infoContent}>
                          {rootCauseFault?.fault_description}
                        </span>
                      </span>
                    </Tooltip>
                  </div>
                </Flex>
              </Col>
            </Row>
          </div>
        </div>
        <Flex className={classnames(styles.card, styles.impactCard)} vertical>
          <div className={styles.cardHeader}>
            {intl.get('impact_range_evaluation')}
          </div>
          <Flex className={classnames(styles.content, styles.impactContent)}>
            <div className={styles.rcoListContainer}>
              <div className={styles.rcoList}>
                <Collapse
                  bordered={false}
                  defaultActiveKey={['1']}
                  expandIcon={({ isActive }) => (
                    <CaretRightOutlined rotate={isActive ? 90 : 0} />
                  )}
                  style={{ background: token.colorBgContainer }}
                  items={serviceTreeData}
                />
              </div>
            </div>
            {/* 拓扑图表 */}
            <div className={styles.graphContainer}>
              <ImpactGraph />
            </div>
          </Flex>
        </Flex>
      </Flex>
      <Flex className={classnames(styles.card, styles.faultCardWrap)} vertical>
        <div>
          <div className={styles.cardHeader}>
            {intl.get('occurrence_process')}
          </div>
          <div className={styles.faultProcessWrap}>
            <div>
              <Tooltip
                title={description}
                styles={{ container: { width: 630 } }}
                placement="bottomLeft"
                getTooltipContainer={() => document.body}
                className={styles.processText}
              >
                <Paragraph
                  ellipsis={{ rows: 5 }}
                  className={styles.processText}
                >
                  {intl.get('fault_process')}：{description}
                </Paragraph>
              </Tooltip>
            </div>
            <Text ellipsis={false} className={styles.impactText}>
              {intl.get('impact')}：{impact}
            </Text>
          </div>
          <Divider />
        </div>
        <div className={styles.faultCardContent}>
          {backtrace.map((item) => (
            <FaultCard
              style={{ marginBottom: 8 }}
              data={item}
              rootCauseFaultId={rootCauseFaultId}
              key={item.fault_id}
              cardWidth={616}
              serviceNameWidth={450}
              onClick={onFaultDetailClick}
            />
          ))}
        </div>
      </Flex>
      <FaultDetialDrawer
        visible={drawerVisible}
        onCancel={closeDetailDrawer}
        faultId={drawerDetialId}
      />
    </Flex>
  );
};

export default Overview;
