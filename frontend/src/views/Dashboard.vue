<template>
  <div class="ds-dashboard">
    <header class="dash-header">
      <div class="dash-title">
        <h1>总览</h1>
        <span>{{ lastSyncedAt }}</span>
      </div>
      <div class="dash-actions">
        <el-button size="small" @click="router.push('/monitor/incident')">告警事件</el-button>
        <el-button size="small" @click="router.push('/cicd/pipelines')">流水线</el-button>
        <el-button size="small" type="primary" :loading="refreshing" @click="refreshData">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </header>

    <section class="metric-grid">
      <div v-for="metric in metrics" :key="metric.key" class="metric-card" :class="`is-${metric.tone}`">
        <div class="metric-top">
          <span>{{ metric.label }}</span>
          <el-icon><component :is="metric.icon" /></el-icon>
        </div>
        <strong>{{ metric.value }}</strong>
        <small>{{ metric.meta }}</small>
      </div>
    </section>

    <section class="dash-grid">
      <div class="panel panel-map">
        <div class="panel-header">
          <h2>工作区</h2>
          <span>{{ modules.length }}</span>
        </div>
        <div class="module-grid">
          <button v-for="module in modules" :key="module.name" class="module-card" type="button" @click="router.push(module.path)">
            <span class="module-icon" :class="`tone-${module.tone}`"><el-icon><component :is="module.icon" /></el-icon></span>
            <span class="module-main">
              <strong>{{ module.name }}</strong>
              <small>{{ module.meta }}</small>
            </span>
            <span class="module-status" :class="module.status"></span>
          </button>
        </div>
      </div>

      <div class="panel panel-health">
        <div class="panel-header">
          <h2>健康状态</h2>
          <span>{{ healthScore }}%</span>
        </div>
        <div class="health-block">
          <div class="health-bar">
            <i :style="{ width: `${healthScore}%` }"></i>
          </div>
          <div class="health-rows">
            <div>
              <span>健康</span>
              <strong>{{ stats.activeInstances }}</strong>
            </div>
            <div>
              <span>故障</span>
              <strong>{{ stats.errorInstances }}</strong>
            </div>
            <div>
              <span>检测</span>
              <strong>{{ stats.todayTests }}</strong>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="dash-grid lower">
      <div class="panel chart-panel">
        <div class="panel-header">
          <h2>实例类型</h2>
          <el-button text size="small" @click="router.push('/es/instances')">查看</el-button>
        </div>
        <div ref="typeChartRef" class="chart"></div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h2>实例列表</h2>
          <span>{{ instances.length }}</span>
        </div>
        <div class="dense-list">
          <div v-if="instances.length === 0" class="empty-row">暂无实例</div>
          <div v-for="instance in instances" :key="instance.id" class="dense-row">
            <span class="avatar">{{ getInitial(instance.name) }}</span>
            <div class="row-main">
              <strong>{{ instance.name || '-' }}</strong>
              <small>{{ instance.instance_type || '未知' }} · {{ instance.address || '-' }}</small>
            </div>
            <span class="status-pill" :class="getStatusClass(instance.status)">{{ getInstanceStatusLabel(instance.status) }}</span>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h2>检测记录</h2>
          <span>{{ recentTests.length }}</span>
        </div>
        <div class="dense-list">
          <div v-if="recentTests.length === 0" class="empty-row">暂无检测记录</div>
          <div v-for="test in recentTests" :key="test.id" class="dense-row">
            <span class="dot" :class="test.result"></span>
            <div class="row-main">
              <strong>{{ test.instanceName || '-' }}</strong>
              <small>{{ formatTime(test.testedAt) }}</small>
            </div>
            <span class="mono" :class="test.result">{{ test.responseTime ? `${test.responseTime}ms` : (test.errorMessage || '失败') }}</span>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Box,
  CircleCheck,
  CircleClose,
  Cloudy,
  Coin,
  Cpu,
  DataAnalysis,
  DataLine,
  Files,
  Monitor,
  Promotion,
  Refresh,
  SetUp,
  Timer,
  Warning
} from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { getInstanceList, getTestHistory, getTodayTestStats } from '@/api/instance.js'

const router = useRouter()
const refreshing = ref(false)
const stats = ref({ totalInstances: 0, activeInstances: 0, errorInstances: 0, todayTests: 0 })
const recentTests = ref([])
const instances = ref([])
const lastSyncedAt = ref('--:--')
const typeChartRef = ref(null)
let typeChart = null

const healthScore = computed(() => {
  const total = stats.value.totalInstances || 0
  if (!total) return 100
  return Math.max(0, Math.min(100, Math.round(((total - stats.value.errorInstances) / total) * 100)))
})

const metrics = computed(() => [
  { key: 'instances', label: '实例总数', value: stats.value.totalInstances, meta: '已托管', tone: 'info', icon: Monitor },
  { key: 'healthy', label: '健康', value: stats.value.activeInstances, meta: `${healthScore.value}%`, tone: 'success', icon: CircleCheck },
  { key: 'failed', label: '故障', value: stats.value.errorInstances, meta: '待处理', tone: 'error', icon: CircleClose },
  { key: 'checks', label: '检测', value: stats.value.todayTests, meta: '今日', tone: 'warning', icon: Timer }
])

const modules = [
  { name: 'Kubernetes', meta: '容器编排', path: '/k8s', icon: Cloudy, tone: 'cyan', status: 'ok' },
  { name: 'CI/CD', meta: '流水线', path: '/cicd/pipelines', icon: Promotion, tone: 'violet', status: 'ok' },
  { name: 'Kafka', meta: '消息队列', path: '/kafka/clusters', icon: DataLine, tone: 'orange', status: 'ok' },
  { name: 'Elasticsearch', meta: '搜索引擎', path: '/es', icon: DataAnalysis, tone: 'blue', status: 'ok' },
  { name: 'MySQL', meta: '关系数据库', path: '/mysql/workbench', icon: Coin, tone: 'green', status: 'ok' },
  { name: 'MongoDB', meta: '文档数据库', path: '/mongodb', icon: Files, tone: 'green', status: 'ok' },
  { name: '监控中心', meta: '告警', path: '/monitor/incident', icon: Warning, tone: 'red', status: 'warn' },
  { name: '任务调度', meta: '工作流', path: '/task-scheduler/workflows', icon: SetUp, tone: 'slate', status: 'ok' },
  { name: '资产管理', meta: '主机', path: '/asset/hosts', icon: Cpu, tone: 'indigo', status: 'ok' }
]

const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

const getInitial = (name = '') => name.trim().slice(0, 1).toUpperCase() || 'D'
const getStatusClass = (status) => ({ active: 'success', online: 'success', warning: 'warning', offline: 'error', error: 'error', inactive: 'neutral' }[status] || 'neutral')
const getInstanceStatusLabel = (status) => ({ active: '在线', online: '在线', warning: '警告', offline: '离线', error: '异常', inactive: '空闲' }[status] || status || '未知')

function initTypeChart() {
  if (!typeChartRef.value) return
  typeChart = echarts.init(typeChartRef.value)
  updateTypeChart([])
}

function updateTypeChart(allInstances) {
  if (!typeChart) return
  const typeCount = {}
  allInstances.forEach((instance) => {
    const type = instance.instance_type || 'other'
    typeCount[type] = (typeCount[type] || 0) + 1
  })
  const palette = ['#3b82f6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#6b7280']
  const chartData = Object.keys(typeCount).map((type, index) => ({ value: typeCount[type], name: type, itemStyle: { color: palette[index % palette.length] } }))
  typeChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'item', backgroundColor: '#161b22', borderColor: 'rgba(255,255,255,.1)', textStyle: { color: '#f0f6fc' } },
    legend: { bottom: 0, icon: 'circle', textStyle: { color: '#8b949e', fontSize: 11 } },
    series: [{
      name: 'type',
      type: 'pie',
      radius: ['55%', '76%'],
      center: ['50%', '43%'],
      label: { show: false },
      labelLine: { show: false },
      data: chartData.length ? chartData : [{ value: 1, name: 'none', itemStyle: { color: '#30363d' } }]
    }]
  })
}

async function refreshData() {
  refreshing.value = true
  try {
    await fetchDashboardData()
    ElMessage.success('已刷新')
  } catch (error) {
    ElMessage.error(error.message || '刷新失败')
  } finally {
    refreshing.value = false
  }
}

async function fetchDashboardData() {
  const response = await getInstanceList({ page: 1, page_size: 100 })
  const allInstances = response.data?.list?.data || []
  stats.value.totalInstances = allInstances.length
  stats.value.activeInstances = allInstances.filter((item) => ['active', 'online'].includes(item.status)).length
  stats.value.errorInstances = allInstances.filter((item) => ['error', 'offline'].includes(item.status)).length
  instances.value = allInstances.slice(0, 8)
  await fetchTodayTestsData()
  await fetchRecentTestsData()
  updateTypeChart(allInstances)
  lastSyncedAt.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

async function fetchTodayTestsData() {
  try {
    const response = await getTodayTestStats()
    stats.value.todayTests = response.data?.total_tests || 0
  } catch {
    stats.value.todayTests = 0
  }
}

async function fetchRecentTestsData() {
  try {
    const testPromises = instances.value.map((instance) => getTestHistory(instance.id, { page_size: 8 }).catch(() => ({ data: { test_list: [] } })))
    const testResults = await Promise.all(testPromises)
    const allTests = []
    testResults.forEach((result, index) => {
      const testList = result.data?.tests || result.data?.test_list || []
      testList.forEach((test) => {
        allTests.push({
          id: test.id || `${index}-${test.tested_at}`,
          instanceName: test.instance?.name || instances.value[index]?.name || '-',
          result: test.test_result === 'success' ? 'success' : test.test_result === 'timeout' ? 'warning' : 'error',
          responseTime: test.response_time || 0,
          errorMessage: test.error_message || '',
          testedAt: test.tested_at
        })
      })
    })
    allTests.sort((a, b) => new Date(b.testedAt) - new Date(a.testedAt))
    recentTests.value = allTests.slice(0, 8)
  } catch {
    recentTests.value = []
  }
}

onMounted(async () => {
  await nextTick()
  initTypeChart()
  await fetchDashboardData()
  window.addEventListener('resize', () => typeChart?.resize())
})
</script>

<style scoped>
.ds-dashboard {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 8px;
  color: var(--ds-text-primary);
}

.dash-header,
.panel,
.metric-card {
  border: 1px solid var(--ds-border-default);
  border-radius: 6px;
  background: var(--ds-bg-surface);
}

.dash-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 38px;
  padding: 0 10px;
}

.dash-title {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.dash-title h1 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: -0.02em;
}

.dash-title span,
.panel-header span,
.metric-card small,
.module-main small,
.row-main small {
  color: var(--ds-text-muted);
  font-size: 10px;
}

.dash-actions {
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
  padding: 10px;
}

.metric-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--ds-text-tertiary);
  font-size: 11px;
}

.metric-card strong {
  display: block;
  margin-top: 8px;
  font-size: 24px;
  font-weight: 600;
  line-height: 1;
}

.metric-card.is-success { border-color: rgba(34, 197, 94, .28); }
.metric-card.is-warning { border-color: rgba(245, 158, 11, .28); }
.metric-card.is-error { border-color: rgba(239, 68, 68, .28); }
.metric-card.is-info { border-color: rgba(59, 130, 246, .28); }

.dash-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(260px, .7fr);
  gap: 8px;
}

.dash-grid.lower {
  grid-template-columns: minmax(0, 1fr) minmax(280px, .8fr) minmax(280px, .8fr);
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
  font-size: 11px;
  font-weight: 600;
}

.module-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
  padding: 10px;
}

.module-card {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 44px;
  padding: 7px;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 5px;
  color: var(--ds-text-primary);
  background: var(--ds-bg-surface-2);
  cursor: pointer;
  text-align: left;
  transition: var(--ds-transition-fast);
}

.module-card:hover {
  border-color: var(--ds-border-strong);
  background: var(--ds-bg-hover);
}

.module-icon {
  display: grid;
  width: 24px;
  height: 24px;
  place-items: center;
  flex: none;
  border-radius: 5px;
  background: rgba(59, 130, 246, .12);
  color: var(--ds-accent);
  font-size: 13px;
}

.tone-green { color: var(--ds-success); background: rgba(34, 197, 94, .12); }
.tone-red { color: var(--ds-error); background: rgba(239, 68, 68, .12); }
.tone-orange { color: var(--ds-warning); background: rgba(245, 158, 11, .12); }
.tone-cyan { color: #06b6d4; background: rgba(6, 182, 212, .12); }
.tone-violet { color: #8b5cf6; background: rgba(139, 92, 246, .12); }
.tone-slate { color: #94a3b8; background: rgba(148, 163, 184, .12); }
.tone-indigo { color: #818cf8; background: rgba(129, 140, 248, .12); }

.module-main,
.row-main {
  display: grid;
  min-width: 0;
  gap: 1px;
  flex: 1;
}

.module-main strong,
.row-main strong {
  overflow: hidden;
  color: var(--ds-text-secondary);
  font-size: 11px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.module-status,
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--ds-text-muted);
}

.module-status.ok,
.dot.success { background: var(--ds-success); }
.module-status.warn,
.dot.warning { background: var(--ds-warning); }
.dot.error { background: var(--ds-error); }

.health-block {
  padding: 10px;
}

.health-bar {
  height: 6px;
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

.health-rows {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  margin-top: 10px;
}

.health-rows div {
  padding: 8px;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 5px;
  background: var(--ds-bg-surface-2);
}

.health-rows span {
  display: block;
  color: var(--ds-text-muted);
  font-size: 10px;
}

.health-rows strong {
  display: block;
  margin-top: 4px;
  font-size: 16px;
}

.chart {
  height: 240px;
}

.dense-list {
  display: grid;
  gap: 0;
}

.dense-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 0 10px;
  border-bottom: 1px solid var(--ds-border-subtle);
}

.dense-row:last-child {
  border-bottom: 0;
}

.avatar {
  display: grid;
  width: 22px;
  height: 22px;
  place-items: center;
  flex: none;
  border-radius: 5px;
  color: var(--ds-text-primary);
  background: var(--ds-bg-surface-3);
  font-size: 10px;
  font-weight: 700;
}

.status-pill {
  height: 18px;
  padding: 0 6px;
  border-radius: 999px;
  font-size: 10px;
  line-height: 18px;
}

.status-pill.success { color: var(--ds-success); background: var(--ds-bg-success-subtle); }
.status-pill.warning { color: var(--ds-warning); background: var(--ds-bg-warning-subtle); }
.status-pill.error { color: var(--ds-error); background: var(--ds-bg-danger-subtle); }
.status-pill.neutral { color: var(--ds-text-tertiary); background: var(--ds-bg-surface-3); }

.mono {
  font-family: var(--ds-font-mono);
  color: var(--ds-text-tertiary);
  font-size: 10px;
}

.mono.success { color: var(--ds-success); }
.mono.warning { color: var(--ds-warning); }
.mono.error { color: var(--ds-error); }

.empty-row {
  padding: 18px 10px;
  color: var(--ds-text-muted);
  font-size: 11px;
  text-align: center;
}

@media (max-width: 1280px) {
  .metric-grid,
  .module-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .dash-grid,
  .dash-grid.lower {
    grid-template-columns: 1fr;
  }
}
</style>
