<template>
  <div class="cluster-detail">
    <el-card class="page-header-card">
      <div class="page-header">
        <div class="header-left">
          <el-button @click="goBack" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <div class="title-section">
            <div class="cluster-icon">
              <el-icon size="24"><Connection /></el-icon>
            </div>
            <div>
              <h2>集群详情</h2>
              <div class="status-tags">
                <el-tag type="info" size="small">未知</el-tag>
                <el-tag type="info" size="small">未知类型</el-tag>
              </div>
            </div>
          </div>
        </div>
        <div class="header-right">
          <el-button @click="refreshData" :loading="refreshing">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 核心指标卡片 -->
    <el-row :gutter="20" class="metrics-row">
      <el-col :span="6">
        <el-card class="metric-card">
          <div class="metric-content">
            <div class="metric-icon nodes">
              <el-icon size="32"><Monitor /></el-icon>
            </div>
            <div class="metric-info">
              <div class="metric-value">{{ clusterMetrics.totalNodes }}</div>
              <div class="metric-label">节点总数</div>
              <div class="metric-sub">就绪: {{ clusterMetrics.readyNodes }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="metric-card">
          <div class="metric-content">
            <div class="metric-icon pods">
              <el-icon size="32"><Box /></el-icon>
            </div>
            <div class="metric-info">
              <div class="metric-value">{{ clusterMetrics.totalPods }}</div>
              <div class="metric-label">工作负载</div>
              <div class="metric-sub">Pod: {{ clusterMetrics.totalPods }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="metric-card">
          <div class="metric-content">
            <div class="metric-icon cpu">
              <el-icon size="32"><Cpu /></el-icon>
            </div>
            <div class="metric-info">
              <div class="metric-value">{{ clusterMetrics.cpuUsage }}%</div>
              <div class="metric-label">CPU使用率</div>
              <div class="metric-sub">{{ clusterMetrics.cpuAvailable }}核可用 / {{ clusterMetrics.cpuTotal }}核总量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="metric-card">
          <div class="metric-content">
            <div class="metric-icon memory">
              <el-icon size="32"><Memo /></el-icon>
            </div>
            <div class="metric-info">
              <div class="metric-value">{{ clusterMetrics.memoryUsage }}%</div>
              <div class="metric-label">内存使用率</div>
              <div class="metric-sub">{{ clusterMetrics.memoryAvailable }}Mi可用 / {{ clusterMetrics.memoryTotal }}Mi总量</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 详细信息区域 -->
    <el-row :gutter="20" class="detail-row">
      <!-- 左栏 -->
      <el-col :span="12">
        <!-- 基础信息 -->
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon><InfoFilled /></el-icon>
              <span>基础信息</span>
            </div>
          </template>
          <div class="info-content">
            <div class="info-row">
              <span class="info-label">集群名称</span>
              <span class="info-value">{{ clusterInfo.name || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Kubernetes版本</span>
              <span class="info-value">{{ clusterInfo.version || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">集群类型</span>
              <span class="info-value">{{ clusterInfo.type || '未知类型' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ clusterInfo.createTime || '未知' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">最后同步</span>
              <span class="info-value">{{ clusterInfo.lastSync || '未知' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">运行时间</span>
              <span class="info-value">{{ clusterInfo.uptime || '未知' }}</span>
            </div>
          </div>
        </el-card>

        <!-- 网络配置 -->
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon><Link /></el-icon>
              <span>网络配置</span>
            </div>
          </template>
          <div class="info-content">
            <div class="info-row">
              <span class="info-label">Service CIDR</span>
              <span class="info-value">{{ networkConfig.serviceCidr || '未配置' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Pod CIDR</span>
              <span class="info-value">{{ networkConfig.podCidr || '未配置' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">API Server</span>
              <span class="info-value">{{ networkConfig.apiServer || '未配置' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">网络插件</span>
              <span class="info-value">{{ networkConfig.networkPlugin || '未知' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">服务转发</span>
              <span class="info-value">{{ networkConfig.serviceForward || '未知' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">DNS服务</span>
              <span class="info-value">{{ networkConfig.dnsService || '未知' }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右栏 -->
      <el-col :span="12">
        <!-- 工作负载统计 -->
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon><DataAnalysis /></el-icon>
              <span>工作负载统计</span>
            </div>
          </template>
          <div class="workload-stats">
            <div class="stat-grid">
              <div class="stat-item">
                <div class="stat-icon deployment">
                  <el-icon><Files /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-number">{{ workloadStats.deployments }}</div>
                  <div class="stat-name">Deployments</div>
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-icon statefulset">
                  <el-icon><Collection /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-number">{{ workloadStats.statefulSets }}</div>
                  <div class="stat-name">StatefulSets</div>
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-icon daemonset">
                  <el-icon><Operation /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-number">{{ workloadStats.daemonSets }}</div>
                  <div class="stat-name">DaemonSets</div>
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-icon job">
                  <el-icon><Timer /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-number">{{ workloadStats.jobs }}</div>
                  <div class="stat-name">Jobs</div>
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-icon total-pods">
                  <el-icon><Box /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-number">{{ workloadStats.totalPods }}</div>
                  <div class="stat-name">总Pod数</div>
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-icon running-pods">
                  <el-icon><VideoPlay /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-number">{{ workloadStats.runningPods }}</div>
                  <div class="stat-name">运行中Pod</div>
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 安装的组件 -->
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon><Grid /></el-icon>
              <span>安装的组件</span>
            </div>
          </template>
          <div class="components-list">
            <el-empty v-if="components.length === 0" description="暂无组件信息" />
            <div v-else class="component-grid">
              <div v-for="component in components" :key="component.name" class="component-item">
                <el-icon><SetUp /></el-icon>
                <span>{{ component.name }}</span>
                <el-tag size="small" :type="component.status === 'active' ? 'success' : 'info'">
                  {{ component.status }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 运行时信息 -->
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon><Setting /></el-icon>
              <span>运行时信息</span>
            </div>
          </template>
          <div class="runtime-info">
            <div class="info-row">
              <span class="info-label">容器运行时</span>
              <span class="info-value">{{ runtimeInfo.containerRuntime || '未知' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">API Server版本</span>
              <span class="info-value">{{ runtimeInfo.apiServerVersion || '未知' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">ETCD版本</span>
              <span class="info-value">{{ runtimeInfo.etcdVersion || '未知' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">CoreDNS版本</span>
              <span class="info-value">{{ runtimeInfo.coreDnsVersion || '未知' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Kube-proxy版本</span>
              <span class="info-value">{{ runtimeInfo.kubeProxyVersion || '未知' }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 节点信息表格 -->
    <el-card class="nodes-card">
      <template #header>
        <div class="card-header">
          <el-icon><Monitor /></el-icon>
          <span>节点信息</span>
          <el-badge :value="nodes.length" type="info" class="node-count" />
        </div>
      </template>
      <el-table :data="nodes" v-loading="loading" empty-text="暂无节点数据">
        <el-table-column prop="name" label="节点名称" min-width="150" />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.role === 'master' ? 'danger' : 'primary'">
              {{ scope.row.role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.status === 'Ready' ? 'success' : 'warning'">
              {{ scope.row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="internalIP" label="内部IP" width="130" />
        <el-table-column prop="externalIP" label="外部IP" width="130" />
        <el-table-column prop="k8sVersion" label="K8s版本" width="120" />
        <el-table-column prop="os" label="操作系统" min-width="150" />
        <el-table-column prop="createTime" label="创建时间" width="160" />
      </el-table>
    </el-card>

    <!-- 存储信息 -->
    <el-card class="storage-card">
      <template #header>
        <div class="card-header">
          <el-icon><FolderOpened /></el-icon>
          <span>存储信息</span>
        </div>
      </template>
      <div class="storage-stats">
        <div class="storage-item">
          <div class="storage-icon">
            <el-icon><Document /></el-icon>
          </div>
          <div class="storage-info">
            <div class="storage-value">{{ storageInfo.totalPV }}</div>
            <div class="storage-label">PV总数</div>
          </div>
        </div>
        <div class="storage-item">
          <div class="storage-icon">
            <el-icon><Folder /></el-icon>
          </div>
          <div class="storage-info">
            <div class="storage-value">{{ storageInfo.totalPVC }}</div>
            <div class="storage-label">PVC总数</div>
          </div>
        </div>
        <div class="storage-item">
          <div class="storage-icon">
            <el-icon><Coin /></el-icon>
          </div>
          <div class="storage-info">
            <div class="storage-value">{{ storageInfo.storageClasses }}</div>
            <div class="storage-label">存储类</div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft, Refresh, Connection, Monitor, Box, Cpu, Memo,
  InfoFilled, Link, DataAnalysis, Files, Collection, Operation,
  Timer, VideoPlay, Grid, SetUp, Setting, FolderOpened,
  Document, Folder, Coin
} from '@element-plus/icons-vue'
import { getClusterInfo, getClusterMetrics, getNodeList } from '@/api/k8s/cluster-info'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const refreshing = ref(false)

// 集群指标数据
const clusterMetrics = ref({
  totalNodes: 0,
  readyNodes: 0,
  totalPods: 0,
  cpuUsage: 0,
  cpuAvailable: 0,
  cpuTotal: 0,
  memoryUsage: 0,
  memoryAvailable: 0,
  memoryTotal: 0
})

// 集群基础信息
const clusterInfo = ref({
  name: '',
  version: '',
  type: '',
  createTime: '',
  lastSync: '',
  uptime: ''
})

// 网络配置
const networkConfig = ref({
  serviceCidr: '',
  podCidr: '',
  apiServer: '',
  networkPlugin: '',
  serviceForward: '',
  dnsService: ''
})

// 工作负载统计
const workloadStats = ref({
  deployments: 0,
  statefulSets: 0,
  daemonSets: 0,
  jobs: 0,
  totalPods: 0,
  runningPods: 0
})

// 组件列表
const components = ref([])

// 运行时信息
const runtimeInfo = ref({
  containerRuntime: '',
  apiServerVersion: '',
  etcdVersion: '',
  coreDnsVersion: '',
  kubeProxyVersion: ''
})

// 节点列表
const nodes = ref([])

// 存储信息
const storageInfo = ref({
  totalPV: 0,
  totalPVC: 0,
  storageClasses: 0
})

const goBack = () => {
  router.go(-1)
}

const refreshData = async () => {
  refreshing.value = true
  try {
    await fetchClusterDetail()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败: ' + (error.message || '未知错误'))
  } finally {
    refreshing.value = false
  }
}

const fetchClusterDetail = async () => {
  loading.value = true
  try {
    const clusterName = route.params.clusterName || 'default'
    
    // 获取集群基本信息
    const infoResponse = await getClusterInfo(instanceId)
    const data = infoResponse.data?.data
    
    if (data) {
      clusterInfo.value = data.clusterInfo
      networkConfig.value = data.networkConfig
      runtimeInfo.value = data.runtimeInfo
    }

    // 获取集群指标数据
    const metricsResponse = await getClusterMetrics(instanceId)
    const metrics = metricsResponse.data?.metrics
    
    if (metrics) {
      clusterMetrics.value = {
        totalNodes: metrics.totalNodes,
        readyNodes: metrics.readyNodes,
        totalPods: metrics.totalPods,
        cpuUsage: metrics.cpuUsage,
        cpuAvailable: metrics.cpuAvailable,
        cpuTotal: metrics.cpuTotal,
        memoryUsage: metrics.memoryUsage,
        memoryAvailable: metrics.memoryAvailable,
        memoryTotal: metrics.memoryTotal
      }

      workloadStats.value = metrics.workloadStats
      storageInfo.value = metrics.storageInfo
    }

    // 获取节点列表
    const nodesResponse = await getNodeList(instanceId)
    const nodeList = nodesResponse.data?.nodeList || []
    
    nodes.value = nodeList.map(node => ({
      name: node.name,
      role: node.role,
      status: node.status,
      internalIP: node.internalIP,
      externalIP: node.externalIP,
      k8sVersion: node.k8sVersion,
      os: node.osImage,
      createTime: new Date(node.createTime * 1000).toLocaleString('zh-CN')
    }))

    // 模拟组件列表（实际应该从集群中获取）
    components.value = [
      { name: 'CoreDNS', status: 'active' },
      { name: 'Calico', status: 'active' },
      { name: 'Metrics Server', status: 'active' }
    ]

  } catch (error) {
    ElMessage.error('获取集群详情失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchClusterDetail()
})
</script>

<style scoped>
.cluster-detail {
  min-height: 100vh;
  padding: 20px;
  background: var(--ds-bg-app, #0d1117);
  color: var(--ds-text-primary, #f0f6fc);
}

.page-header-card,
.metric-card,
.info-card,
.nodes-card,
.storage-card {
  margin-bottom: 20px;
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 8px;
  background: var(--ds-bg-surface, #161b22) !important;
  box-shadow: none !important;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-btn {
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1));
  border-radius: 6px;
  color: var(--ds-accent, #3b82f6);
  background: var(--ds-bg-surface-2, #1c212b);
}

.title-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cluster-icon,
.metric-icon,
.stat-icon,
.storage-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid currentColor;
  background: transparent;
  box-shadow: none;
}

.cluster-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  color: var(--ds-accent, #3b82f6);
}

.title-section h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--ds-text-primary, #f0f6fc);
}

.status-tags {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.metrics-row,
.detail-row {
  margin-bottom: 20px;
}

.metric-content {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 100%;
}

.metric-icon {
  width: 64px;
  height: 64px;
  border-radius: 8px;
}

.metric-icon.nodes { color: #a78bfa; }
.metric-icon.pods { color: #f472b6; }
.metric-icon.cpu { color: var(--ds-accent, #3b82f6); }
.metric-icon.memory { color: var(--ds-success, #22c55e); }

.metric-info {
  flex: 1;
}

.metric-value,
.stat-number,
.storage-value {
  color: var(--ds-text-primary, #f0f6fc);
  font-weight: 700;
  line-height: 1.2;
}

.metric-value,
.storage-value {
  font-size: 24px;
}

.metric-label,
.info-label,
.stat-name,
.storage-label {
  color: var(--ds-text-tertiary, #8b949e);
}

.metric-label,
.storage-label {
  margin: 4px 0;
  font-size: 14px;
}

.metric-sub {
  font-size: 12px;
  color: var(--ds-text-muted, #6e7681);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--ds-text-primary, #f0f6fc);
  font-size: 16px;
}

.info-content,
.workload-stats,
.components-list,
.runtime-info {
  padding: 0;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--ds-border-subtle, rgba(255,255,255,0.06));
}

.info-row:last-child {
  border-bottom: none;
}

.info-label,
.info-value {
  font-size: 14px;
}

.info-value {
  color: var(--ds-text-secondary, #c9d1d9);
  font-weight: 500;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.stat-item,
.component-item {
  display: flex;
  align-items: center;
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1));
  border-radius: 8px;
  background: var(--ds-bg-surface-2, #1c212b);
}

.stat-item {
  gap: 12px;
  padding: 12px;
}

.stat-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
}

.stat-icon.deployment,
.stat-icon.total-pods { color: var(--ds-accent, #3b82f6); }
.stat-icon.statefulset,
.stat-icon.running-pods { color: var(--ds-success, #22c55e); }
.stat-icon.daemonset { color: var(--ds-warning, #f59e0b); }
.stat-icon.job { color: var(--ds-error, #ef4444); }

.stat-info {
  flex: 1;
}

.stat-number {
  font-size: 18px;
}

.stat-name {
  font-size: 12px;
}

.component-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.component-item {
  gap: 8px;
  padding: 8px;
}

.component-item span {
  flex: 1;
  font-size: 14px;
  color: var(--ds-text-secondary, #c9d1d9);
}

.node-count {
  margin-left: auto;
}

.storage-stats {
  display: flex;
  gap: 24px;
}

.storage-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.storage-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  color: var(--ds-accent, #3b82f6);
}

.storage-info {
  display: flex;
  flex-direction: column;
}

:deep(.el-card__header),
:deep(.el-card__body) {
  border-color: var(--ds-border-subtle, rgba(255,255,255,0.06));
  background: transparent;
}

:deep(.el-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: var(--ds-bg-surface-2, #1c212b);
  --el-table-row-hover-bg-color: var(--ds-bg-hover, rgba(255,255,255,0.04));
  --el-table-border-color: var(--ds-border-subtle, rgba(255,255,255,0.06));
  --el-table-text-color: var(--ds-text-secondary, #c9d1d9);
  --el-table-header-text-color: var(--ds-text-tertiary, #8b949e);
  background: transparent;
  color: var(--ds-text-secondary, #c9d1d9);
}

:deep(.el-table__inner-wrapper),
:deep(.el-table__header-wrapper),
:deep(.el-table__body-wrapper),
:deep(.el-table tr),
:deep(.el-table td.el-table__cell),
:deep(.el-table th.el-table__cell) {
  background: transparent;
}

:deep(.el-table th.el-table__cell) {
  background: var(--ds-bg-surface-2, #1c212b);
  color: var(--ds-text-tertiary, #8b949e);
}

@media (max-width: 1200px) {
  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .component-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .metrics-row .el-col,
  .detail-row .el-col {
    margin-bottom: 16px;
  }

  .stat-grid {
    grid-template-columns: 1fr;
  }

  .storage-stats {
    flex-direction: column;
    gap: 16px;
  }
}
</style>