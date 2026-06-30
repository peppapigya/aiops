<template>
  <div class="k8s-page">
    <header class="k8s-header">
      <div class="title-row">
        <h1>Kubernetes 集群</h1>
        <span>{{ clusterList.length }} clusters</span>
      </div>
      <div class="header-actions">
        <el-button size="small" @click="navigateTo('/k8s/event')">事件</el-button>
        <el-button size="small" @click="navigateTo('/k8s/pod')">Pod列表</el-button>
        <el-button size="small" type="primary" :loading="refreshing" @click="refreshData">
          <el-icon><Refresh /></el-icon>
          Refresh
        </el-button>
      </div>
    </header>

    <section class="metric-grid">
      <div v-for="item in metrics" :key="item.key" class="metric-card" :class="`is-${item.tone}`">
        <div class="metric-label">
          <span>{{ item.label }}</span>
          <el-icon><component :is="item.icon" /></el-icon>
        </div>
        <strong>{{ item.value }}</strong>
        <small>{{ item.meta }}</small>
      </div>
    </section>

    <section class="main-grid">
      <div class="panel cluster-panel">
        <div class="panel-header">
          <h2>集群</h2>
          <el-button text size="small" @click="navigateTo('/k8s/cluster')">管理</el-button>
        </div>
        <div class="cluster-list">
          <div v-if="clusterList.length === 0" class="empty-row">No clusters</div>
          <button v-for="cluster in clusterList" :key="cluster.id" class="cluster-row" type="button" @click="viewClusterDetail(cluster)">
            <span class="cluster-dot" :class="cluster.status === 'active' ? 'success' : 'neutral'"></span>
            <div class="row-main">
              <strong>{{ cluster.cluster_name }}</strong>
              <small>{{ cluster.last_sync || '-' }}</small>
            </div>
            <div class="row-actions" @click.stop>
              <el-button link size="small" @click="handleTestConnection(cluster)">测试</el-button>
              <el-button link size="small" type="primary" @click="handleChangeCluster(cluster)">使用</el-button>
            </div>
          </button>
        </div>
      </div>

      <div class="panel health-panel">
        <div class="panel-header">
          <h2>健康状态</h2>
          <span>{{ nodeReadyRate }}%</span>
        </div>
        <div class="health-body">
          <div class="health-bar"><i :style="{ width: `${nodeReadyRate}%` }"></i></div>
          <div class="health-list">
            <div>
              <span>就绪节点</span>
              <strong>{{ resourceStatus.healthyNodes }}/{{ resourceStatus.totalNodes }}</strong>
            </div>
            <div>
              <span>警告Pod</span>
              <strong>{{ resourceStatus.warningPods }}</strong>
            </div>
            <div>
              <span>故障Pod</span>
              <strong>{{ resourceStatus.failedPods }}</strong>
            </div>
            <div>
              <span>等待中Pod</span>
              <strong>{{ resourceStatus.pendingPods }}</strong>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="lower-grid">
      <div class="panel">
        <div class="panel-header">
          <h2>工作负载</h2>
          <span>{{ workloadStats.totalPods }} pods</span>
        </div>
        <div class="resource-grid">
          <button v-for="item in workloadCards" :key="item.path" class="resource-card" type="button" @click="navigateTo(item.path)">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </button>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h2>快速入口</h2>
          <span>resources</span>
        </div>
        <div class="quick-grid">
          <button v-for="action in quickActions" :key="action.path" class="quick-item" type="button" @click="navigateTo(action.path)">
            <el-icon><component :is="action.icon" /></el-icon>
            <span>{{ action.label }}</span>
          </button>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h2>存储</h2>
          <span>{{ storageInfo.usedStorage }} Gi</span>
        </div>
        <div class="storage-grid">
          <div v-for="item in storageCards" :key="item.label" class="storage-card">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Monitor, Box, Files, Connection, Setting, Folder, Refresh, Cpu, Memo,
  Clock, Document, Coin, Check, Warning, Close
} from '@element-plus/icons-vue'
import { getClusterList, changeCluster, testClusterConnection } from '@/api/k8s/cluster'
import { getClusterMetrics, getClusterInfo } from '@/api/k8s/cluster-info'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import eventBus from '@/utils/eventBus'

const router = useRouter()
const refreshing = ref(false)

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

const workloadStats = ref({
  deployments: 0,
  statefulSets: 0,
  daemonSets: 0,
  jobs: 0,
  totalPods: 0,
  runningPods: 0
})

const resourceStatus = ref({
  totalNodes: 0,
  healthyNodes: 0,
  warningPods: 0,
  failedPods: 0,
  pendingPods: 0
})

const storageInfo = ref({
  totalPV: 0,
  totalPVC: 0,
  storageClasses: 0,
  usedStorage: 0
})

const clusterList = ref([])
const clusterInfo = ref({})

const nodeReadyRate = computed(() => {
  const total = Number(resourceStatus.value.totalNodes) || 0
  if (!total) return 100
  return Math.round((Number(resourceStatus.value.healthyNodes || 0) / total) * 100)
})

const metrics = computed(() => [
  { key: 'nodes', label: 'Nodes', value: clusterMetrics.value.totalNodes, meta: `${clusterMetrics.value.readyNodes} ready`, tone: 'info', icon: Monitor },
  { key: 'pods', label: 'Pods', value: clusterMetrics.value.totalPods, meta: `${workloadStats.value.runningPods || 0} running`, tone: 'success', icon: Box },
  { key: 'cpu', label: 'CPU', value: `${clusterMetrics.value.cpuUsage}%`, meta: `${clusterMetrics.value.cpuAvailable}/${clusterMetrics.value.cpuTotal} core`, tone: 'warning', icon: Cpu },
  { key: 'memory', label: 'Memory', value: `${clusterMetrics.value.memoryUsage}%`, meta: `${clusterMetrics.value.memoryAvailable}/${clusterMetrics.value.memoryTotal} Mi`, tone: 'info', icon: Memo }
])

const workloadCards = computed(() => [
  { label: 'Deployments', value: workloadStats.value.deployments, path: '/k8s/deployment' },
  { label: 'StatefulSets', value: workloadStats.value.statefulSets, path: '/k8s/statefulset' },
  { label: 'DaemonSets', value: workloadStats.value.daemonSets, path: '/k8s/daemonset' },
  { label: 'Jobs', value: workloadStats.value.jobs, path: '/k8s/job' },
  { label: 'Total Pods', value: workloadStats.value.totalPods, path: '/k8s/pod' },
  { label: 'Running', value: workloadStats.value.runningPods, path: '/k8s/pod' }
])

const quickActions = [
  { label: 'Cluster', path: '/k8s/cluster', icon: Setting },
  { label: 'Nodes', path: '/k8s/node', icon: Monitor },
  { label: 'Pods', path: '/k8s/pod', icon: Box },
  { label: 'Deploy', path: '/k8s/deployment', icon: Files },
  { label: 'Namespace', path: '/k8s/namespace', icon: Folder },
  { label: 'Service', path: '/k8s/service', icon: Connection },
  { label: 'ConfigMap', path: '/k8s/configmap', icon: Document },
  { label: 'CronJob', path: '/k8s/cronjob', icon: Clock },
  { label: 'RBAC', path: '/k8s/role', icon: Check }
]

const storageCards = computed(() => [
  { label: 'PV', value: storageInfo.value.totalPV },
  { label: 'PVC', value: storageInfo.value.totalPVC },
  { label: 'SC', value: storageInfo.value.storageClasses },
  { label: 'Used', value: `${storageInfo.value.usedStorage} Gi` }
])

const navigateTo = (path) => {
  router.push(path)
}

const viewClusterDetail = (cluster) => {
  router.push(`/k8s/cluster/${cluster.cluster_name}`)
}

const refreshData = async () => {
  refreshing.value = true
  try {
    await fetchDashboardData()
    ElMessage.success('refreshed')
  } catch (error) {
    ElMessage.error(error.message || 'refresh failed')
  } finally {
    refreshing.value = false
  }
}

const handleChangeCluster = async (cluster) => {
  try {
    await changeCluster(cluster.cluster_name)
    ElMessage.success(`cluster: ${cluster.cluster_name}`)
    await fetchDashboardData()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || error.message || 'switch failed')
  }
}

const handleTestConnection = async (cluster) => {
  try {
    await testClusterConnection(cluster.cluster_name)
    ElMessage.success('connected')
  } catch (error) {
    ElMessage.error(error.response?.data?.message || error.message || 'test failed')
  }
}

const fetchDashboardData = async () => {
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getClusterList(instanceId)
    clusterList.value = response.data?.clasters || []

    const infoResponse = await getClusterInfo(instanceId)
    clusterInfo.value = infoResponse.data?.clusterInfo || {}

    const metricsResponse = await getClusterMetrics(instanceId)
    const metricsData = metricsResponse.data?.metrics

    if (metricsData) {
      clusterMetrics.value = {
        totalNodes: metricsData.totalNodes,
        readyNodes: metricsData.readyNodes,
        totalPods: metricsData.totalPods,
        cpuUsage: metricsData.cpuUsage,
        cpuAvailable: metricsData.cpuAvailable,
        cpuTotal: metricsData.cpuTotal,
        memoryUsage: metricsData.memoryUsage,
        memoryAvailable: metricsData.memoryAvailable,
        memoryTotal: metricsData.memoryTotal
      }

      workloadStats.value = metricsData.workloadStats
      storageInfo.value = metricsData.storageInfo

      resourceStatus.value = {
        totalNodes: metricsData.totalNodes,
        healthyNodes: metricsData.readyNodes,
        warningPods: metricsData.workloadStats.unknownPods || 0,
        failedPods: metricsData.workloadStats.failedPods || 0,
        pendingPods: metricsData.workloadStats.pendingPods || 0
      }
    }
  } catch (error) {
    console.error('获取仪表板数据失败:', error)
    ElMessage.error(error.message || 'fetch failed')
  }
}

const handlePodEvent = () => {
  fetchDashboardData()
}

onMounted(() => {
  fetchDashboardData()
  eventBus.on('pod:created', handlePodEvent)
  eventBus.on('pod:deleted', handlePodEvent)
})

onUnmounted(() => {
  eventBus.off('pod:created', handlePodEvent)
  eventBus.off('pod:deleted', handlePodEvent)
})
</script>

<style scoped>
.k8s-page {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 8px;
  color: var(--ds-text-primary);
}

.k8s-header,
.panel,
.metric-card {
  border: 1px solid var(--ds-border-default);
  border-radius: 6px;
  background: var(--ds-bg-surface);
}

.k8s-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 40px;
  padding: 0 10px;
}

.title-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.title-row h1 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.title-row span,
.panel-header span,
.metric-card small,
.row-main small,
.resource-card span,
.storage-card span,
.health-list span {
  color: var(--ds-text-muted);
  font-size: 11px;
}

.header-actions {
  display: flex;
  gap: 4px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.metric-card {
  min-height: 72px;
  padding: 8px 10px;
}

.metric-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--ds-text-tertiary);
  font-size: 12px;
}

.metric-card strong {
  display: block;
  margin-top: 8px;
  font-size: 22px;
  font-weight: 600;
  line-height: 1;
}

.metric-card.is-success { border-color: rgba(34, 197, 94, .28); }
.metric-card.is-warning { border-color: rgba(245, 158, 11, .28); }
.metric-card.is-info { border-color: rgba(59, 130, 246, .28); }

.main-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(340px, .8fr);
  gap: 8px;
}

.lower-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(360px, .9fr) minmax(280px, .7fr);
  gap: 8px;
}

.panel {
  min-width: 0;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  padding: 0 10px;
  border-bottom: 1px solid var(--ds-border-subtle);
}

.panel-header h2 {
  margin: 0;
  color: var(--ds-text-secondary);
  font-size: 12px;
  font-weight: 600;
}

.cluster-list,
.health-body {
  padding: 8px 10px;
}

.cluster-row {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: 36px;
  gap: 8px;
  padding: 0 8px;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 6px;
  color: var(--ds-text-primary);
  background: var(--ds-bg-surface-2);
  cursor: pointer;
  text-align: left;
  transition: var(--ds-transition-fast);
}

.cluster-row + .cluster-row {
  margin-top: 6px;
}

.cluster-row:hover,
.resource-card:hover,
.quick-item:hover {
  border-color: var(--ds-border-strong);
  background: var(--ds-bg-hover);
}

.cluster-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--ds-text-muted);
}

.cluster-dot.success { background: var(--ds-success); }
.cluster-dot.neutral { background: var(--ds-text-muted); }

.row-main {
  display: grid;
  min-width: 0;
  flex: 1;
  gap: 2px;
}

.row-main strong {
  overflow: hidden;
  color: var(--ds-text-secondary);
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row-actions {
  display: flex;
  gap: 2px;
}

.health-bar {
  height: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--ds-bg-surface-3);
}

.health-bar i {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--ds-accent), var(--ds-success));
}

.health-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px;
  margin-top: 8px;
}

.health-list div,
.resource-card,
.quick-item,
.storage-card {
  border: 1px solid var(--ds-border-subtle);
  border-radius: 6px;
  background: var(--ds-bg-surface-2);
}

.health-list div {
  padding: 6px 8px;
}

.health-list strong {
  display: block;
  margin-top: 3px;
  font-size: 14px;
}

.resource-grid,
.quick-grid,
.storage-grid {
  display: grid;
  gap: 6px;
  padding: 8px 10px;
}

.resource-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.quick-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.storage-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.resource-card,
.quick-item {
  min-height: 48px;
  color: var(--ds-text-primary);
  cursor: pointer;
  transition: var(--ds-transition-fast);
}

.resource-card {
  display: grid;
  gap: 4px;
  padding: 6px 8px;
  text-align: left;
}

.resource-card strong,
.storage-card strong {
  font-size: 16px;
  font-weight: 600;
}

.quick-item {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 4px;
  color: var(--ds-text-tertiary);
  font-size: 11px;
}

.quick-item .el-icon {
  color: var(--ds-accent);
  font-size: 14px;
}

.storage-card {
  display: grid;
  gap: 4px;
  padding: 6px 8px;
}

.empty-row {
  padding: 16px 8px;
  color: var(--ds-text-muted);
  font-size: 12px;
  text-align: center;
}

@media (max-width: 1280px) {
  .metric-grid,
  .resource-grid,
  .quick-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .main-grid,
  .lower-grid {
    grid-template-columns: 1fr;
  }
}
</style>
