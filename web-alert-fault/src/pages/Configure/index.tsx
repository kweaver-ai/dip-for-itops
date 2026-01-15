import { FC, useEffect, useRef, useState } from 'react';
import { App, Button, Form, Input, Select } from 'antd';
import { DefaultOptionType } from 'antd/lib/select';
import intl from 'react-intl-universal';
import styles from './index.module.less';
import { DEFAULT_EXPIRATION, FORM_LAYOUT } from './constants';
import InputComposition from './InputComposition';
import {
  getConfigure,
  getKnowledgeNetworks,
  updateConfigure
} from '@/services/configure';

const Configure: FC = () => {
  const [form] = Form.useForm();
  const { notification } = App.useApp();
  const [knowledgeNetworks, setKnowledgeNetworks] = useState<
    DefaultOptionType[]
  >([]);
  const [isDisable, setIsDisable] = useState(true);
  const initTokenRef = useRef('');
  const options = [
    // { label: intl.get('Day'), value: 'd' },
    { label: intl.get('Hours'), value: 'h' }
    // { label: intl.get('Minutes'), value: 'm' }
  ];

  const initFormData = async () => {
    const configure = await getConfigure();

    if (configure.platform.auth_token) {
      initTokenRef.current = configure.platform.auth_token;
      form.setFieldsValue(configure);
    }
  };

  /**
   * 获取业务知识网络列表
   */
  const fetchKnowledgeNetworks = async () => {
    const res = await getKnowledgeNetworks();

    if (res.entries) {
      const networks: DefaultOptionType[] = res.entries.map((item: any) => ({
        label: item.name,
        value: item.id
      }));

      setKnowledgeNetworks(networks);
    }
  };

  const onFinish = async (values: any) => {
    console.log(values);
    const isChanged = values.platform.auth_token !== initTokenRef.current;

    const res = await updateConfigure(initTokenRef.current === '', {
      ...values,
      platform: {
        ...(isChanged ? { auth_token: values.platform.auth_token } : {})
      }
    });

    if (res.success === 1) {
      notification.success({
        message: intl.get('SaveSuccess')
      });
      setIsDisable(true);
    }
  };

  useEffect(() => {
    Promise.all([initFormData(), fetchKnowledgeNetworks()]);
  }, []);

  return (
    <div className="bg-white h-full px-[24px] pt-[12px]">
      <Form
        form={form}
        {...FORM_LAYOUT}
        labelAlign="left"
        colon={false}
        onFinish={onFinish}
        onValuesChange={() => {
          setIsDisable(false);
        }}
      >
        <div className={styles['form-row']}>
          <div className={styles['form-row-title']}>
            {intl.get('connection')}
          </div>
          <Form.Item
            label={intl.get('authToken')}
            name={['platform', 'auth_token']}
            extra={intl.get('authTokenHelper')}
            rules={[{ required: true, message: intl.get('authTokenRequired') }]}
          >
            <Input placeholder={intl.get('Input')} />
          </Form.Item>
        </div>
        <div className={styles['form-row']}>
          <div className={styles['form-row-title']}>
            {intl.get('knowledgeNetwork')}
          </div>
          <Form.Item
            label={intl.get('knowledgeNetwork')}
            name={['knowledge_network', 'knowledge_id']}
            rules={[
              { required: true, message: intl.get('knowledgeNetworkRequired') }
            ]}
          >
            <Select
              placeholder={intl.get('Select')}
              options={knowledgeNetworks}
            />
          </Form.Item>
        </div>
        <div className={styles['form-row']}>
          <div className={styles['form-row-title']}>
            {intl.get('faultPointConvergenceStrategy')}
          </div>
          <Form.Item
            label={intl.get('timeCorrelation')}
            name={['fault_point_policy', 'expiration']}
            initialValue={DEFAULT_EXPIRATION}
            extra={intl.get('convergenceStrategyHelper')}
          >
            <InputComposition
              placeholder={intl.get('Input')}
              min={1}
              precision={0}
              options={options}
            />
          </Form.Item>
        </div>
        <div className={styles['form-row']}>
          <div className={styles['form-row-title']}>
            {intl.get('faultPointAssociationStrategy')}
          </div>
          <Form.Item
            label={intl.get('timeCorrelation')}
            name={['problem_policy', 'expiration']}
            initialValue={DEFAULT_EXPIRATION}
            extra={intl.get('associationStrategyHelper')}
          >
            <InputComposition
              placeholder={intl.get('Input')}
              min={1}
              precision={0}
              options={options}
            />
          </Form.Item>
        </div>
        <Form.Item label=" " className={styles.operation}>
          <Button type="primary" htmlType="submit" disabled={isDisable}>
            {intl.get('Save')}
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
};

export default Configure;
