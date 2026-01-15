/* eslint-disable max-lines */
const mockData: any = {
  problem_id: 'P-20250718-001',
  problem_name: '主机(10.20.84.93)磁盘故障导致pod离线',
  problem_create_timestamp: '2025-12-16T14:00:00Z',
  problem_start_time: '2025-12-16T14:02:10Z',
  problem_end_time: '2025-12-16T14:27:10Z',
  problem_closed_time: '2025-12-16T14:27:10Z',
  problem_duration: 1500,
  problem_level: '0',
  problem_status: '0',
  root_cause_object_id: 'svc-order-api',
  root_cause_fault_id: 'fault-pod-cpu-throttle-001',
  rca_results: {
    rca_id: 'A-20251216-001',
    itops_app_id: 'itops-03',
    itops_app_name: 'ITops—故障根因分析',
    adp_kn_name: '运维知识网络',
    adp_kn_id: 'systemkn',
    rca_context: {
      occurrence: {
        description:
          '用户反馈订单服务在14:00左右出现大量超时错误，API响应延迟超过10秒。用户反馈订单服务在14:00左右出现大量超时错误，API响应延迟超过10秒。用户反馈订单服务在14:00左右出现大量超时错误，API响应延迟超过10秒。用户反馈订单服务在14:00左右出现大量超时错误，API响应延迟超过10秒。',
        impact: '影响核心交易链路，导致订单创建失败率上升至15%，持续约25分钟。'
      },
      backtrace: [
        {
          fault_id: 'fault-pod-cpu-throttle-001',
          fault_mode: 'CPU资源争用',
          fault_level: '1',
          fault_status: '1',
          fault_description:
            "POD 'order-svc-7d5b9c6f8-xk2l9' 因CPU限制过低导致频繁节流（throttling）",
          fault_occur_time: '2025-12-16T14:02:10Z',
          fault_latest_time: '2025-12-16T14:02:10Z',
          fault_duration_time: 1500,
          fault_recovery_time: '2025-12-16T14:27:10Z',
          entity_object_class: 'pod',
          entity_object_name: 'order-svc-7d5b9c6f8-xk2l9',
          entity_object_id: 'pod-ord-7d5b9c6f8-xk2l9',
          relation_event_ids: [
            'evt-cpu-throttle-001',
            'evt-k8s-pod-latency-002',
            'evt-prom-alert-003'
          ]
        },
        {
          fault_id: 'fault-service-cpu-throttle-002',
          fault_mode: 'CPU资源争用2',
          fault_level: '2',
          fault_status: '2',
          fault_description:
            "POD 'order-svc-7d5b9c6f8-xk2l9' 因CPU限制过低导致频繁节流（throttling）",
          fault_occur_time: '2025-12-16T14:03:10Z',
          fault_latest_time: '2025-12-16T14:03:10Z',
          fault_duration_time: 1500,
          fault_recovery_time: '2025-12-16T14:27:10Z',
          entity_object_class: 'service',
          entity_object_name: 'fault-service-cpu-throttle-002',
          entity_object_id: 'svc-order-api',
          relation_event_ids: [
            'evt-cpu-throttle-001',
            'evt-k8s-pod-latency-002',
            'evt-prom-alert-003'
          ]
        },
        {
          fault_id: 'fault-host-cpu-throttle-003',
          fault_mode: '主机网络拥塞',
          fault_level: '2',
          fault_status: '1',
          fault_description:
            "POD 'order-svc-7d5b9c6f8-xk2l9' 因CPU限制过低导致频繁节流（throttling）",
          fault_occur_time: '2025-12-16T14:05:10Z',
          fault_latest_time: '2025-12-16T14:05:10Z',
          fault_duration_time: 1500,
          fault_recovery_time: '2025-12-16T14:27:10Z',
          entity_object_class: 'host',
          entity_object_name: 'fault-host-cpu-throttle-003',
          entity_object_id: 'host-08-prod',
          relation_event_ids: [
            'evt-cpu-throttle-001',
            'evt-k8s-pod-latency-002',
            'evt-prom-alert-003'
          ]
        },
        {
          fault_id: 'fault-middleware-cpu-throttle-004',
          fault_mode: 'CPU资源争用4',
          fault_level: 1,
          fault_status: '1',
          fault_description:
            "POD 'order-svc-7d5b9c6f8-xk2l9' 因CPU限制过低导致频繁节流（throttling）",
          fault_occur_time: '2025-12-16T14:05:10Z',
          fault_latest_time: '2025-12-16T14:05:10Z',
          fault_duration_time: 1500,
          fault_recovery_time: '2025-12-16T14:27:10Z',
          entity_object_class: 'middleware',
          entity_object_name: 'fault-middleware-cpu-throttle-004',
          entity_object_id: 'mw-redis-ord-01',
          relation_event_ids: [
            'evt-cpu-throttle-001',
            'evt-k8s-pod-latency-002',
            'evt-prom-alert-003'
          ]
        },
        {
          fault_id: 'fault-database-cpu-throttle-005',
          fault_mode: 'CPU资源争用5',
          fault_level: '3',
          fault_status: '2',
          fault_description:
            "POD 'order-svc-7d5b9c6f8-xk2l9' 因CPU限制过低导致频繁节流（throttling）",
          fault_occur_time: '2025-12-16T14:05:30Z',
          fault_latest_time: '2025-12-16T14:05:30Z',
          fault_duration_time: 1500,
          fault_recovery_time: '2025-12-16T14:27:10Z',
          entity_object_class: 'database',
          entity_object_name: 'fault-database-cpu-throttle-005',
          entity_object_id: 'db-mysql-ord-01',
          relation_event_ids: [
            'evt-cpu-throttle-001',
            'evt-k8s-pod-latency-002',
            'evt-prom-alert-003'
          ]
        },
        {
          fault_id: 'fault-physical_machine-cpu-throttle-006',
          fault_mode: 'CPU资源争用6',
          fault_level: '2',
          fault_status: '2',
          fault_description:
            "POD 'order-svc-7d5b9c6f8-xk2l9' 因CPU限制过低导致频繁节流（throttling）",
          fault_occur_time: '2025-12-16T14:06:30Z',
          fault_latest_time: '2025-12-16T14:06:30Z',
          fault_duration_time: 1500,
          fault_recovery_time: '2025-12-16T14:27:10Z',
          entity_object_class: 'physical_machine',
          entity_object_name: 'fault-physical_machine-cpu-throttle-006',
          entity_object_id: 'pm-rack3-u12',
          relation_event_ids: [
            'evt-cpu-throttle-001',
            'evt-k8s-pod-latency-002',
            'evt-prom-alert-003'
          ]
        },
        {
          fault_id: 'fault-network-switch-cpu-throttle-007',
          fault_mode: 'CPU资源争用7',
          fault_level: '2',
          fault_status: '3',
          fault_description:
            "POD 'order-svc-7d5b9c6f8-xk2l9' 因CPU限制过低导致频繁节流（throttling）",
          fault_occur_time: '2025-12-16T14:06:32Z',
          fault_latest_time: '2025-12-16T14:06:32Z',
          fault_duration_time: 1500,
          fault_recovery_time: '2025-12-16T14:27:10Z',
          entity_object_class: 'network_switch',
          entity_object_name: 'fault-network-switch-cpu-throttle-007',
          entity_object_id: 'sw-core-dc1',
          relation_event_ids: [
            'evt-cpu-throttle-001',
            'evt-k8s-pod-latency-002',
            'evt-prom-alert-003'
          ]
        }
      ],
      network: {
        nodes: [
          {
            s_id: 'svc-order-api',
            s_create_time: '2025-11-01T10:00:00Z',
            s_update_time: '2025-12-15T09:30:00Z',
            ip_address: ['10.100.45.12'],
            object_ports: '8080,9090',
            name: 'order-api',
            object_type: 'order-service',
            object_namespace: 'prod-order',
            object_impact_level: '1',
            object_class: 'service',
            relation_event_ids: ['evt-api-timeout-001', 'evt-svc-latency-002'],
            relation_object_ids: ['pod-ord-7d5b9c6f8-xk2l9'],
            relation_fault_point_ids: ['fault-service-cpu-throttle-002'],
            relation_resource: []
          },
          {
            s_id: 'pod-ord-7d5b9c6f8-xk2l9',
            s_create_time: '2025-12-16T10:15:00Z',
            s_update_time: '2025-12-16T14:03:00Z',
            ip_address: ['10.244.8.23'],
            name: 'order-svc-7d5b9c6f8-xk2l9',
            object_impact_level: '1',
            object_class: 'pod',
            relation_event_ids: [
              'evt-cpu-throttle-001',
              'evt-k8s-pod-latency-002',
              'evt-oomkilled-008'
            ],
            relation_object_ids: [
              'svc-order-api',
              'host-08-prod',
              'mw-redis-ord-01',
              'db-mysql-ord-01'
            ],
            relation_fault_point_ids: ['fault-pod-cpu-throttle-001'],
            relation_resource: []
          },
          {
            s_id: 'host-08-prod',
            s_create_time: '2024-05-20T08:00:00Z',
            s_update_time: '2025-12-10T16:45:00Z',
            ip_address: ['192.168.10.108'],
            object_ports: '',
            name: 'host-node-08',
            object_type: 'compute-host',
            object_namespace: '',
            object_impact_level: '2',
            object_class: 'host',
            relation_event_ids: ['evt-net-drop-004', 'evt-disk-io-009'],
            relation_object_ids: ['pod-ord-7d5b9c6f8-xk2l9', 'pm-rack3-u12'],
            relation_fault_point_ids: ['fault-host-cpu-throttle-003'],
            relation_resource: []
          },
          {
            s_id: 'mw-redis-ord-01',
            s_create_time: '2025-10-01T08:00:00Z',
            s_update_time: '2025-12-10T11:20:00Z',
            ip_address: ['10.100.50.20'],
            object_ports: '6379',
            name: 'redis-order-cache',
            object_type: 'redis-cluster',
            object_namespace: 'prod-cache',
            object_impact_level: '3',
            object_class: 'middleware',
            relation_event_ids: ['evt-redis-conn-full-006'],
            relation_object_ids: ['pod-ord-7d5b9c6f8-xk2l9'],
            relation_fault_point_ids: ['fault-middleware-cpu-throttle-004'],
            relation_resource: []
          },
          {
            s_id: 'db-mysql-ord-01',
            s_create_time: '2025-09-15T09:00:00Z',
            s_update_time: '2025-12-12T14:10:00Z',
            ip_address: ['10.100.60.10'],
            object_ports: '3306',
            name: 'order-db-master',
            object_type: 'mysql-8.0',
            object_namespace: 'prod-db',
            object_impact_level: '4',
            object_class: 'database',
            relation_event_ids: ['evt-db-slowlog-007'],
            relation_object_ids: ['pod-ord-7d5b9c6f8-xk2l9'],
            relation_fault_point_ids: ['fault-database-cpu-throttle-005'],
            relation_resource: []
          },
          {
            s_id: 'pm-rack3-u12',
            s_create_time: '2023-06-10T00:00:00Z',
            s_update_time: '2025-11-30T17:00:00Z',
            ip_address: ['192.168.1.12'],
            object_ports: '',
            name: 'rack3-server12',
            object_type: 'x86-server',
            object_namespace: '',
            object_impact_level: '1',
            object_class: 'physical_machine',
            relation_event_ids: ['evt-temp-alert-010'],
            relation_object_ids: ['host-08-prod', 'sw-core-dc1'],
            relation_fault_point_ids: [
              'fault-physical_machine-cpu-throttle-006'
            ],
            relation_resource: []
          },
          {
            s_id: 'sw-core-dc1',
            s_create_time: '2023-05-01T00:00:00Z',
            s_update_time: '2025-10-25T09:15:00Z',
            ip_address: ['192.168.0.1'],
            object_ports: '',
            name: 'core-switch-dc1',
            object_type: 'network-switch',
            object_namespace: '',
            object_impact_level: '2',
            object_class: 'network_machine',
            relation_object_ids: ['pm-rack3-u12'],
            relation_fault_point_ids: ['fault-network-switch-cpu-throttle-007'],
            relation_resource: []
          }
        ],
        edges: [
          {
            relation_id: 'rel-svc-pod-001',
            relation_class: 'service_include_pod',
            relation_create_time: '2025-12-16T10:15:05Z',
            source_object_id: 'svc-order-api',
            source_object_class: 'Service',
            target_object_id: 'pod-ord-7d5b9c6f8-xk2l9',
            target_object_class: 'pod'
          },
          {
            relation_id: 'rel-pod-host-002',
            relation_class: 'pod_runs_on_host',
            relation_create_time: '2025-12-16T10:15:10Z',
            source_object_id: 'pod-ord-7d5b9c6f8-xk2l9',
            source_object_class: 'Pod',
            target_object_id: 'host-08-prod',
            target_object_class: 'host'
          },
          {
            relation_id: 'rel-host-physical-003',
            relation_class: 'host_runs_on_physical',
            relation_create_time: '2024-05-20T08:05:00Z',
            source_object_id: 'host-08-prod',
            source_object_class: 'host',
            target_object_id: 'pm-rack3-u12',
            target_object_class: 'physical_machine'
          },
          {
            relation_id: 'rel-physical-network-004',
            relation_class: 'physical_connects_to_network',
            relation_create_time: '2023-06-10T01:00:00Z',
            source_object_id: 'pm-rack3-u12',
            source_object_class: 'physical_machine',
            target_object_id: 'sw-core-dc1',
            target_object_class: 'network_machine'
          },
          {
            relation_id: 'rel-pod-mw-005',
            relation_class: 'pod_calls_middleware',
            relation_create_time: '2025-12-16T10:15:20Z',
            source_object_id: 'pod-ord-7d5b9c6f8-xk2l9',
            source_object_class: 'pod',
            target_object_id: 'mw-redis-ord-01',
            target_object_class: 'middleware'
          },
          {
            relation_id: 'rel-pod-db-006',
            relation_class: 'pod_writes_to_database',
            relation_create_time: '2025-12-16T10:15:25Z',
            source_object_id: 'pod-ord-7d5b9c6f8-xk2l9',
            source_object_class: 'pod',
            target_object_id: 'db-mysql-ord-01',
            target_object_class: 'database'
          }
        ]
      }
    }
  }
};

export default mockData;
