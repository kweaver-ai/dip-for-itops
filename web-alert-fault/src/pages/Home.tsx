import { Spin, Flex } from 'antd';
import { useEffect } from 'react';
import { useNavigate } from '@noya/max';

const Home = () => {
  const navigate = useNavigate();

  useEffect(() => {
    // window.location.href = `/dip-hub/application/${urlParams.applicationId}/fault-analysis`;
    navigate('/fault-analysis');
  }, []);

  return (
    <Flex justify="center" align="center" style={{ height: '100%' }}>
      <Spin size="large" />
    </Flex>
  );
};

export default Home;
