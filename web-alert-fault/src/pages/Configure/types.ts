export interface ITOpsConfigure {
  platform: { auth_token: string };
  knowledge_network: { knowledge_id: string };
  fault_point_policy: { expiration: TimeUnit };
  problem_policy: { expiration: TimeUnit };
}
