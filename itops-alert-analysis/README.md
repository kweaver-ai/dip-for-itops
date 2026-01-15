# ITOps Alert Analysis

ITOPS 告警分析服务 - 提供告警事件收敛、故障点关联、问题聚合和根因分析能力。

## 目录

- [前置条件](#前置条件)
- [Helm 3 部署](#helm-3-部署)
- [配置管理](#配置管理)
- [升级与回滚](#升级与回滚)
- [卸载](#卸载)
- [故障排查](#故障排查)
- [常用运维命令](#常用运维命令)
- [Helm Chart 开发](#helm-chart-开发)

---

## 前置条件

- Kubernetes 1.19+
- **Helm 3.0+**（不支持 Helm 2）
- 依赖服务：
    - Kafka (消息队列)
    - OpenSearch (数据存储)
    - Redis (缓存)

---

## Helm 3 部署

### 1. 准备配置文件

编辑 `values.yaml` 配置：

```bash
vim helm/itops-alert-analysis/values.yaml
```

关键配置项：

```yaml
# 命名空间和集群
namespace: dip
cluster: dip

# 副本数
replicaCount: 1

# 镜像配置
image:
  registry: acr.aishu.cn
  repository: ar/anyrobot-itops-alert-analysis
  tag: latest  # 修改为实际镜像版本
  pullPolicy: Always

# 服务端口
service:
  alertAnalysis:
    port: 13047

# 依赖服务配置
config:
  depServices:
    mq:
      mqHost: kafka-headless.resource.svc.cluster.local.
      mqPort: 9097
      auth:
        username: anyrobot
        password: eisoo.com123
    opensearch:
      host: opensearch-master.resource.svc.cluster.local.
      port: 9200
      user: admin
      password: eisoo.com123
    redis:
      connectInfo:
        masterGroupName: mymaster
        password: eisoo.com123
        sentinelHost: proton-redis-proton-redis-sentinel.resource
        sentinelPort: 26379
```

### 2. 使用 Helm 3 安装

#### 方式1：从项目目录安装（开发/测试）

```bash
# 进入项目目录
cd /Users/xiaoxingxing/GoglandProjects/AlertAnalysis

# 使用 Helm 3 安装
helm3 install itops-alert-analysis \
  ./helm/itops-alert-analysis \
  --namespace dip \
  --create-namespace

# 使用自定义 values 文件
helm3 install itops-alert-analysis \
  ./helm/itops-alert-analysis \
  --namespace dip \
  --create-namespace \
  --values ./custom-values.yaml
```

#### 方式2：从 Helm Repository 安装

```bash
# 添加 Helm Repository
helm3 repo add ar https://acr.aishu.cn/chartrepo/ar
helm3 repo update

# 搜索可用版本
helm3 search repo itops-alert-analysis --versions

# 安装最新版本
helm3 install itops-alert-analysis \
  ar/itops-alert-analysis \
  --namespace dip \
  --create-namespace

# 安装指定版本
helm3 install itops-alert-analysis \
  --version 5.1-0-feature-arp-797669.1704491 \
  ar/itops-alert-analysis \
  --namespace dip \
  --create-namespace

# 安装时覆盖配置
helm3 install itops-alert-analysis \
  --version 5.1-0-feature-arp-797669.1704491 \
  ar/itops-alert-analysis \
  --namespace dip \
  --create-namespace \
  --set config.log.level=debug \
  --set replicaCount=2
```

### 3. 验证部署

```bash
# 查看 Helm Release 状态
helm3 list -n dip

# 查看 Pod 状态
kubectl get pods -n dip -l anyrobot-module=itops-alert-analysis

# 查看 Pod 日志
kubectl logs -f -n dip deployment/itops-alert-analysis

# 查看 ConfigMap
kubectl get configmap itops-alert-analysis-configmap -n dip -o yaml

# 查看 Service
kubectl get svc -n dip itops-alert-analysis-dip

# 查看 Ingress
kubectl get ingress -n dip

# 查看所有相关资源
kubectl get all,cm,ingress -n dip -l anyrobot-module=itops-alert-analysis
```

---

## 配置管理

### 查看当前配置

```bash
# 查看 Helm Release 的配置值
helm3 get values itops-alert-analysis -n dip

# 查看完整配置（包括默认值）
helm3 get values itops-alert-analysis -n dip --all

# 查看 ConfigMap 内容
kubectl get configmap itops-alert-analysis-configmap -n dip -o yaml
```

### 修改配置

**⚠️ 重要**：修改配置必须通过 `helm3 upgrade`，直接修改 ConfigMap 不会触发 Pod 重启！

#### 方式1：使用 --set 参数

```bash
helm3 upgrade itops-alert-analysis \
  ./helm/itops-alert-analysis \
  --namespace dip \
  --set config.log.level=debug \
  --set replicaCount=2
```

#### 方式2：修改 values.yaml（推荐）

```bash
# 1. 修改 values.yaml
vim helm/itops-alert-analysis/values.yaml

# 2. 使用 helm3 升级
helm3 upgrade itops-alert-analysis \
  ./helm/itops-alert-analysis \
  --namespace dip

# 3. 从 Repository 升级
helm3 upgrade itops-alert-analysis \
  --version 5.1-0-feature-arp-797669.1704500 \
  ar/itops-alert-analysis \
  --namespace dip
```

### ConfigMap 自动重启机制

本项目的 Deployment 配置了 ConfigMap checksum 注解：

```yaml
annotations:
  checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
```

**工作原理**：
```
修改 values.yaml
    ↓
helm3 upgrade
    ↓
ConfigMap 内容变化 → checksum 变化
    ↓
Deployment 检测到 Pod Template 变化
    ↓
自动触发滚动更新重启 Pod
```

**❌ 错误做法**：
```bash
# 直接修改 ConfigMap 不会触发 Pod 重启！
kubectl edit configmap itops-alert-analysis-configmap -n dip
```

**✅ 正确做法**：
```bash
# 1. 修改 values.yaml
# 2. 使用 helm3 升级
helm3 upgrade itops-alert-analysis ./helm/itops-alert-analysis -n dip
```

---

## 升级与回滚

### 升级到新版本

```bash
# 从项目目录升级
helm3 upgrade itops-alert-analysis \
  ./helm/itops-alert-analysis \
  --namespace dip

# 从 Repository 升级到指定版本
helm3 upgrade itops-alert-analysis \
  --version 5.1-0-feature-arp-797669.1704500 \
  ar/itops-alert-analysis \
  --namespace dip

# 升级并修改配置
helm3 upgrade itops-alert-analysis \
  ./helm/itops-alert-analysis \
  --namespace dip \
  --set config.log.level=debug \
  --reuse-values
```

### 查看升级历史

```bash
# 查看所有版本历史
helm3 history itops-alert-analysis -n dip

# 输出示例：
# REVISION  UPDATED                   STATUS      CHART                           DESCRIPTION
# 1         Mon Jan 06 10:00:00 2026  superseded  itops-alert-analysis-1.0.0     Install complete
# 2         Mon Jan 06 11:00:00 2026  deployed    itops-alert-analysis-1.0.1     Upgrade complete
```

### 回滚版本

```bash
# 回滚到上一个版本
helm3 rollback itops-alert-analysis -n dip

# 回滚到指定 REVISION
helm3 rollback itops-alert-analysis 1 -n dip

# 查看回滚后的状态
helm3 history itops-alert-analysis -n dip
```

### 验证升级/回滚状态

```bash
# 查看滚动更新状态
kubectl rollout status deployment/itops-alert-analysis -n dip

# 查看 Pod 重启时间
kubectl get pods -n dip -l anyrobot-module=itops-alert-analysis -o wide

# 查看最新日志
kubectl logs -f -n dip deployment/itops-alert-analysis --tail=100
```

---

## 卸载

### 使用 Helm 3 卸载（推荐）

```bash
# 卸载 Release
helm3 uninstall itops-alert-analysis -n dip

# 卸载并等待所有资源删除完成
helm3 uninstall itops-alert-analysis -n dip --wait

# 确认已卸载
helm3 list -n dip
```

Helm 3 会自动删除以下资源：
- Deployment
- Service
- ConfigMap
- Ingress

### 手动清理资源（仅在 Helm 卸载失败时使用）

```bash
# 删除所有相关资源
kubectl delete deployment itops-alert-analysis -n dip
kubectl delete svc itops-alert-analysis-dip -n dip
kubectl delete configmap itops-alert-analysis-configmap -n dip
kubectl delete ingress itops-alert-analysis-dip-ingress -n dip
```

### 验证卸载完成

```bash
# 确认 Helm Release 已删除
helm3 list -n dip

# 确认 Pod 已删除
kubectl get pods -n dip -l anyrobot-module=itops-alert-analysis

# 确认所有资源已清理
kubectl get all,cm,ingress -n dip -l anyrobot-module=itops-alert-analysis
```

---

## 故障排查

### 1. Pod 一直处于 Pending 状态

```bash
# 查看 Pod 事件
kubectl describe pod <pod-name> -n dip

# 常见原因：
# - 节点资源不足：调整 resources.requests
# - 镜像拉取失败：检查 image.registry 和 imagePullSecrets
# - PVC 绑定失败：检查存储类配置
```

### 2. Pod 一直 CrashLoopBackOff

```bash
# 查看容器日志（包括上一次崩溃的日志）
kubectl logs <pod-name> -n dip --previous

# 查看容器详情
kubectl describe pod <pod-name> -n dip

# 常见原因：
# - 配置错误：检查 ConfigMap 内容
# - 依赖服务不可达：检查 Kafka/OpenSearch/Redis 连接
# - 端口冲突：检查 service.alertAnalysis.port
# - 启动命令错误：检查镜像 ENTRYPOINT
```

### 3. 修改 ConfigMap 后 Pod 未重启

```bash
# 确认是否通过 helm3 upgrade 修改
helm3 get values itops-alert-analysis -n dip

# 查看 Deployment 的 checksum 注解
kubectl get deployment itops-alert-analysis -n dip -o yaml | grep checksum

# ❌ 如果是直接 kubectl edit configmap 修改的，需要手动重启：
kubectl rollout restart deployment/itops-alert-analysis -n dip

# ✅ 正确做法：修改 values.yaml 后执行 helm3 upgrade
helm3 upgrade itops-alert-analysis ./helm/itops-alert-analysis -n dip
```

### 4. 服务无法访问

```bash
# 检查 Service
kubectl get svc itops-alert-analysis-dip -n dip
kubectl describe svc itops-alert-analysis-dip -n dip

# 检查 Ingress
kubectl get ingress -n dip
kubectl describe ingress itops-alert-analysis-dip-ingress -n dip

# 测试 Pod 内部访问
kubectl exec -it <pod-name> -n dip -- curl http://localhost:13047/api/itops-alert-analysis/healthz

# 从集群内其他 Pod 测试
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://itops-alert-analysis-dip.dip.svc.cluster.local:13047/api/itops-alert-analysis/healthz
```

### 5. 查看详细日志

```bash
# 实时查看日志
kubectl logs -f deployment/itops-alert-analysis -n dip

# 查看最近 100 行日志
kubectl logs deployment/itops-alert-analysis -n dip --tail=100

# 查看所有容器的日志
kubectl logs -l anyrobot-module=itops-alert-analysis -n dip --all-containers=true

# 导出日志到文件
kubectl logs deployment/itops-alert-analysis -n dip > app.log

# 查看指定时间范围的日志
kubectl logs deployment/itops-alert-analysis -n dip --since=1h
```

### 6. 临时开启 Debug 日志

```bash
# 使用 helm3 临时开启 Debug
helm3 upgrade itops-alert-analysis \
  ./helm/itops-alert-analysis \
  --namespace dip \
  --set config.log.level=debug \
  --reuse-values

# 查看日志验证
kubectl logs -f deployment/itops-alert-analysis -n dip
```

### 7. 临时关闭语义相关性
```bash
helm3 upgrade itops-alert-analysis --version 0.1-0-feature-arp-797669 ar/itops-alert-analysis --set config.log.level=debug --set config.itops_alert_analysis.problem.semantic_correlation.enabled=false

# 查看日志验证
kubectl logs -f deployment/itops-alert-analysis -n dip
```

---

## 常用运维命令

```bash
# 查看所有相关资源
kubectl get all,cm,ingress -n dip -l anyrobot-module=itops-alert-analysis

# 进入容器 Shell
kubectl exec -it deployment/itops-alert-analysis -n dip -- /bin/sh

# 查看容器环境变量
kubectl exec deployment/itops-alert-analysis -n dip -- env | sort

# 查看容器内配置文件
kubectl exec deployment/itops-alert-analysis -n dip -- cat /opt/itops-alert-analysis/config/config.yaml

# 扩缩容
kubectl scale deployment/itops-alert-analysis -n dip --replicas=3

# 查看资源使用情况
kubectl top pod -n dip -l anyrobot-module=itops-alert-analysis

# 手动触发滚动重启
kubectl rollout restart deployment/itops-alert-analysis -n dip

# 查看滚动更新历史
kubectl rollout history deployment/itops-alert-analysis -n dip

# 暂停/恢复滚动更新
kubectl rollout pause deployment/itops-alert-analysis -n dip
kubectl rollout resume deployment/itops-alert-analysis -n dip
```

---

## Helm Chart 开发

### 本地测试模板

```bash
# 渲染模板查看最终 YAML
helm3 template test ./helm/itops-alert-analysis --debug

# 只渲染 ConfigMap
helm3 template test ./helm/itops-alert-analysis --show-only templates/configmap.yaml

# 验证 Chart 语法
helm3 lint ./helm/itops-alert-analysis

# 模拟安装（不实际部署到 Kubernetes）
helm3 install --dry-run --debug itops-alert-analysis ./helm/itops-alert-analysis -n dip
```

### 打包与发布

```bash
# 打包 Chart
helm3 package ./helm/itops-alert-analysis

# 查看打包结果
ls -lh itops-alert-analysis-*.tgz

# 上传到 Repository（需要配置凭证）
# helm3 push itops-alert-analysis-*.tgz ar/
```

---

## 更多信息



### 相关链接

- [Helm 3 官方文档](https://helm.sh/docs/)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)

---

**最后更新**: 2026-01-07
