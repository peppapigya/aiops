<template>
  <div class="job-management">
    <header class="job-header">
      <div>
        <div class="eyebrow">Kubernetes Batch</div>
        <h1>Jobs</h1>
      </div>
      <div class="header-actions">
        <el-select v-model="selectedNamespace" placeholder="Namespace" filterable class="namespace-select" @change="fetchData">
          <el-option label="All namespaces" value="all" />
          <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
        </el-select>
        <el-button @click="fetchData" :loading="loading" :icon="RefreshRight">Refresh</el-button>
        <el-button type="success" :icon="Monitor">Dashboard</el-button>
        <el-button type="primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:job:showcreatedialogtrue')" :icon="Plus">Create</el-button>
      </div>
    </header>

    <section class="job-layout">
      <main class="job-panel">
        <div class="panel-toolbar">
          <div>
            <span class="panel-title">Job Runs</span>
            <span class="panel-subtitle">{{ filteredJobs.length }} visible · {{ jobList.length }} total</span>
          </div>
          <div class="toolbar-controls">
            <el-input
              v-model="searchKeyword"
              placeholder="Search job"
              class="filter-input"
              :prefix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            />
            <el-button type="primary" @click="handleSearch">Search</el-button>
            <el-button @click="handleReset">Reset</el-button>
          </div>
        </div>

        <el-table
          :data="filteredJobs"
          style="width: 100%"
          v-loading="loading"
          @row-click="handleViewDetail"
        >
          <el-table-column label="Job" min-width="240">
            <template #default="{ row }">
              <div class="job-name-cell">
                <div class="job-avatar"><el-icon><VideoPlay /></el-icon></div>
                <div class="job-identity">
                  <button class="name-link" @click.stop="handleViewDetail(row)">{{ jobName(row) }}</button>
                  <span>{{ row.nameSpace || row.namespace || selectedNamespace }} · Job</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="Execution" min-width="190">
            <template #default="{ row }">
              <div class="execution-cell">
                <div class="execution-top">
                  <span :class="['status-pill', `is-${jobTone(row)}`]">{{ getJobStatus(row) }}</span>
                  <strong>{{ parsePodsStatus(row.podsStatuses).succeeded }}/{{ parsePodsStatus(row.podsStatuses).total }}</strong>
                </div>
                <div class="progress-track">
                  <span :style="{ width: `${jobProgress(row)}%` }" />
                </div>
                <span class="execution-meta">active {{ parsePodsStatus(row.podsStatuses).active }} · failed {{ parsePodsStatus(row.podsStatuses).failed }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="Duration" width="130">
            <template #default="{ row }">
              <span class="time-text">{{ formatDuration(row.startTime, row.endTime) }}</span>
            </template>
          </el-table-column>

          <el-table-column label="Image" min-width="220">
            <template #default="{ row }">
              <div class="image-cell">
                <el-icon><Box /></el-icon>
                <span :title="row.containerImage">{{ row.containerImage || '-' }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="Resources" min-width="180">
            <template #default="{ row }">
              <div class="resource-cell">
                <div><span>CPU</span><strong>{{ formatResourceValue(row.resources?.cpuRequest) }}</strong><em>/ {{ formatResourceValue(row.resources?.cpuLimit) }}</em></div>
                <div><span>MEM</span><strong>{{ formatResourceValue(row.resources?.memoryRequest) }}</strong><em>/ {{ formatResourceValue(row.resources?.memoryLimit) }}</em></div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="Labels" min-width="110" align="center">
            <template #default="{ row }">
              <el-popover placement="top" width="auto" trigger="hover">
                <template #reference>
                  <span class="label-badge"><el-icon><PriceTag /></el-icon>{{ Object.keys(row.labels || {}).length }}</span>
                </template>
                <div class="tags-popover">
                  <span v-for="(val, key) in row.labels" :key="key" class="tag-item">{{ key }}: {{ val }}</span>
                  <span v-if="!Object.keys(row.labels || {}).length" class="tag-empty">No labels</span>
                </div>
              </el-popover>
            </template>
          </el-table-column>

          <el-table-column label="Started" width="170">
            <template #default="{ row }">
              <span class="time-text">{{ formatTime(row.startTime || row.creationTimestamp) }}</span>
            </template>
          </el-table-column>

          <el-table-column label="Actions" width="130">
            <template #default="{ row }">
              <div class="job-actions" @click.stop>
                <el-button link type="primary" size="small" @click="handleViewDetail(row)">Detail</el-button>
                <el-button link type="danger" size="small" @click="handleDelete(row)" v-show="permStore.hasPerm('k8s:job:handledelete')">Delete</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="fetchData"
            @size-change="handleSizeChange"
          />
        </div>
      </main>
    </section>

    <el-dialog
      v-model="showCreateDialog"
      title="创建 Job"
      width="900px"
      :close-on-click-modal="false"
      append-to-body
      class="autoops-dialog"
      @open="handleDialogOpen"
    >
      <el-form label-width="100px">
        <el-form-item label="命名空间" required>
          <el-select v-model="selectedNamespace" placeholder="选择命名空间" style="width: 100%;" filterable>
            <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="YAML配置" required>
          <el-input
            v-model="yamlContent"
            type="textarea"
            :rows="18"
            placeholder="请输入Job的YAML配置"
            class="code-input"
            @input="validateYaml"
          />
          <div class="yaml-actions">
            <el-button size="small" @click="validateYaml" :icon="CircleCheck">验证YAML</el-button>
            <el-tag v-if="yamlContent && isYamlValid" type="success" size="small">语法正确</el-tag>
          </div>
          <el-alert
            v-if="!isYamlValid && yamlError"
            title="YAML语法错误"
            type="error"
            :description="yamlError"
            :closable="false"
            class="yaml-error"
          />
        </el-form-item>
        <el-alert title="YAML示例" type="info" :closable="false">
          <div class="example-block">
            <pre>apiVersion: batch/v1
kind: Job
metadata:
  name: example-job
spec:
  template:
    spec:
      containers:
      - name: example
        image: busybox
        command: ["echo", "Hello World"]
      restartPolicy: Never
  backoffLimit: 4</pre>
            <el-button size="small" type="primary" link @click="loadExample">加载示例</el-button>
          </div>
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :disabled="!isYamlValid || !yamlContent.trim()">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showDetailDialog" title="Job 详情" width="800px" append-to-body class="autoops-dialog">
      <el-input v-model="detailYamlContent" type="textarea" :rows="20" readonly class="code-input" />
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search, RefreshRight, Monitor, Plus,
  VideoPlay, PriceTag, CircleCheck, Box,
  Warning, Timer
} from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import { getJobList, deleteJob, createJob, getJobDetail } from '@/api/k8s/job'
import { getNamespaceList } from '@/api/k8s/namespace'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import relativeTime from 'dayjs/plugin/relativeTime'
import { getSelectedInstanceId } from '@/stores/instanceStore'

dayjs.extend(duration)
dayjs.extend(relativeTime)

const loading = ref(false)
const searchKeyword = ref('')
const jobList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedNamespace = ref('default')

const showCreateDialog = ref(false)
const yamlContent = ref('')
const namespaceList = ref([])
const yamlError = ref('')
const isYamlValid = ref(true)

const showDetailDialog = ref(false)
const detailYamlContent = ref('')

const filteredJobs = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return jobList.value
  return jobList.value.filter((item) => {
    return [jobName(item), item.nameSpace, item.namespace, item.containerImage, getJobStatus(item)]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(keyword))
  })
})

const completedJobs = computed(() => jobList.value.filter((item) => getJobStatus(item) === 'Completed').length)
const runningJobs = computed(() => jobList.value.filter((item) => getJobStatus(item) === 'Running').length)
const failedJobs = computed(() => jobList.value.filter((item) => getJobStatus(item) === 'Failed').length)
const succeededPods = computed(() => jobList.value.reduce((sum, item) => sum + parsePodsStatus(item.podsStatuses).succeeded, 0))
const failedPods = computed(() => jobList.value.reduce((sum, item) => sum + parsePodsStatus(item.podsStatuses).failed, 0))
const failureQueue = computed(() => filteredJobs.value.filter((item) => getJobStatus(item) === 'Failed').slice(0, 8))

const jobMetrics = computed(() => [
  { key: 'total', label: 'Jobs', value: jobList.value.length, meta: `${namespaceList.value.length || 1} namespaces`, tone: 'info', icon: VideoPlay },
  { key: 'completed', label: 'Completed', value: completedJobs.value, meta: 'successful runs', tone: 'success', icon: CircleCheck },
  { key: 'running', label: 'Running', value: runningJobs.value, meta: 'active workload', tone: runningJobs.value ? 'warning' : 'info', icon: Timer },
  { key: 'failed', label: 'Failed', value: failedJobs.value, meta: 'needs triage', tone: failedJobs.value ? 'error' : 'success', icon: Warning }
])

const namespaceBreakdown = computed(() => {
  const counts = new Map()
  jobList.value.forEach((item) => {
    const namespace = item.nameSpace || item.namespace || selectedNamespace.value || 'default'
    counts.set(namespace, (counts.get(namespace) || 0) + 1)
  })
  return Array.from(counts.entries()).map(([name, count]) => ({ name, count })).sort((a, b) => b.count - a.count).slice(0, 8)
})

const jobName = (row) => row.jobName || row.name || '-'
const jobTone = (row) => {
  const status = getJobStatus(row)
  if (status === 'Completed') return 'success'
  if (status === 'Running') return 'warning'
  if (status === 'Failed') return 'error'
  return 'neutral'
}
const jobProgress = (row) => {
  const parsed = parsePodsStatus(row.podsStatuses)
  if (!parsed.total) return 0
  return Math.min(100, Math.round((parsed.succeeded / parsed.total) * 100))
}

const fetchData = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getJobList(selectedNamespace.value, instanceId)
    jobList.value = res.data?.jobList || []
    total.value = res.data?.total || jobList.value.length
  } catch (e) {
    console.error('获取Job列表失败:', e)
    ElMessage.error('获取 Job 列表失败')
    jobList.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => fetchData()
const handleReset = () => { searchKeyword.value = ''; fetchData() }
const handleSizeChange = () => fetchData()
const selectNamespace = (namespace) => { selectedNamespace.value = namespace; fetchData() }
const formatTime = (ts) => dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'
const formatDuration = (startTime, endTime) => {
  if (!startTime || !endTime) return '-'
  const start = dayjs(startTime)
  const end = dayjs(endTime)
  const seconds = end.diff(start, 'second')

  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${hours}h ${minutes}m`
}

const formatResourceValue = (val) => {
  if (!val || val === '0' || val === '0m' || val === '0Mi') return '-'
  return val
}

const parsePodsStatus = (statusStr) => {
  if (!statusStr || typeof statusStr !== 'string') return { active: 0, succeeded: 0, failed: 0, total: 1 }
  const match = statusStr.match(/活跃:\s*(\d+),\s*成功:\s*(\d+),\s*失败:\s*(\d+)/)
  if (match) {
    const active = parseInt(match[1])
    const succeeded = parseInt(match[2])
    const failed = parseInt(match[3])
    return { active, succeeded, failed, total: Math.max(succeeded + failed + active, 1) }
  }
  return { active: 0, succeeded: 0, failed: 0, total: 1 }
}

const getJobStatus = (row) => {
  const parsed = parsePodsStatus(row.podsStatuses)
  if (parsed.active > 0) return 'Running'
  if (parsed.failed > 0) return 'Failed'
  if (parsed.succeeded > 0) return 'Completed'
  return 'Pending'
}

const handleViewDetail = async (row) => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getJobDetail(row.nameSpace || row.namespace, jobName(row), instanceId)
    if (res.data) {
      const detail = res.data.jobDetail || res.data
      detailYamlContent.value = yaml.dump(detail)
      showDetailDialog.value = true
    }
  } catch (e) {
    console.error('获取Job详情失败:', e)
    ElMessage.error('获取详情失败')
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除 Job "${jobName(row)}" 吗？删除后无法恢复！`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
      center: true
    }
  ).then(async () => {
    try {
      const instanceId = getSelectedInstanceId()
      await deleteJob(row.nameSpace || row.namespace, jobName(row), instanceId)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败: ' + (error.message || '未知错误'))
    }
  }).catch(() => {})
}

const handleDialogOpen = async () => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getNamespaceList(instanceId)
    namespaceList.value = res.data?.namespaceList || []
    if (namespaceList.value.length > 0 && !selectedNamespace.value) {
      selectedNamespace.value = namespaceList.value[0].name
    }
  } catch (error) {
    ElMessage.error('获取命名空间列表失败')
    namespaceList.value = []
  }
}

const validateYaml = () => {
  if (!yamlContent.value.trim()) {
    isYamlValid.value = true
    yamlError.value = ''
    return
  }

  try {
    yaml.load(yamlContent.value)
    isYamlValid.value = true
    yamlError.value = ''
  } catch (e) {
    isYamlValid.value = false
    yamlError.value = e.message
  }
}

const loadExample = () => {
  yamlContent.value = `apiVersion: batch/v1
kind: Job
metadata:
  name: example-job
spec:
  template:
    spec:
      containers:
      - name: example
        image: busybox
        command: ["echo", "Hello World"]
      restartPolicy: Never
  backoffLimit: 4`
  validateYaml()
}

const handleCreate = async () => {
  if (!yamlContent.value.trim()) {
    ElMessage.warning('请输入YAML配置')
    return
  }
  try {
    const instanceId = getSelectedInstanceId()
    await createJob(selectedNamespace.value, yamlContent.value, instanceId)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    yamlContent.value = ''
    fetchData()
  } catch (error) {
    ElMessage.error(error.message || '创建失败')
  }
}

onMounted(async () => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getNamespaceList(instanceId)
    namespaceList.value = res.data?.namespaceList || []
    if (namespaceList.value.length > 0 && !selectedNamespace.value) {
      selectedNamespace.value = namespaceList.value[0].name
    }
  } catch (error) {
    console.error('获取命名空间列表失败:', error)
  }
  fetchData()
})
</script>

<style scoped>
.job-management {
  min-height: 100vh;
  padding: var(--ds-space-16, 16px);
  background: var(--ds-bg-app);
  color: var(--ds-text-primary);
}

.job-header,
.panel-toolbar,
.panel-head,
.header-actions,
.toolbar-controls,
.job-actions {
  display: flex;
  align-items: center;
}

.job-header {
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-lg, 8px);
  background: var(--ds-bg-surface);
}

.eyebrow {
  color: var(--ds-accent);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: .08em;
  text-transform: uppercase;
}

.job-header h1 {
  margin: 4px 0 0;
  color: var(--ds-text-primary);
  font-size: 22px;
  font-weight: 700;
}

.header-actions,
.toolbar-controls,
.job-actions {
  gap: 8px;
}

.namespace-select,
.filter-input {
  width: 220px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.metric-card,
.job-panel,
.panel-block {
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-lg, 8px);
  background: var(--ds-bg-surface);
  box-shadow: none;
}

.metric-card {
  padding: 14px;
}

.metric-meta,
.metric-foot,
.panel-subtitle,
.time-text,
.queue-meta,
.resource-cell em,
.job-identity span,
.execution-meta {
  color: var(--ds-text-tertiary);
}

.metric-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.metric-value {
  margin-top: 10px;
  color: var(--ds-text-primary);
  font-size: 26px;
  font-weight: 750;
}

.metric-foot {
  margin-top: 4px;
  font-size: 12px;
}

.metric-card.is-success { border-color: rgba(34, 197, 94, .28); }
.metric-card.is-warning { border-color: rgba(245, 158, 11, .28); }
.metric-card.is-error { border-color: rgba(239, 68, 68, .28); }
.metric-card.is-info { border-color: rgba(59, 130, 246, .28); }

.job-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 12px;
  margin-top: 12px;
}

.job-panel,
.panel-block {
  overflow: hidden;
}

.panel-toolbar {
  justify-content: space-between;
  gap: 12px;
  padding: 14px;
  border-bottom: 1px solid var(--ds-border-subtle);
}

.panel-title {
  display: block;
  color: var(--ds-text-primary);
  font-size: 14px;
  font-weight: 700;
}

.panel-subtitle {
  display: block;
  margin-top: 2px;
  font-size: 12px;
}

.job-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.job-avatar {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid rgba(34, 197, 94, .28);
  border-radius: 7px;
  background: var(--ds-bg-success-subtle);
  color: var(--ds-success);
}

.name-link,
.queue-item,
.summary-row {
  border: 0;
  background: transparent;
  cursor: pointer;
  font: inherit;
}

.name-link {
  display: block;
  max-width: 190px;
  padding: 0;
  overflow: hidden;
  color: var(--ds-text-primary);
  font-weight: 650;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.name-link:hover {
  color: var(--ds-accent);
}

.execution-cell {
  display: grid;
  gap: 6px;
}

.execution-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.status-pill,
.queue-badge,
.label-badge {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 700;
}

.status-pill.is-success,
.queue-badge.is-success {
  border-color: rgba(34, 197, 94, .32);
  background: var(--ds-bg-success-subtle);
  color: var(--ds-success);
}

.status-pill.is-warning,
.queue-badge.is-warning {
  border-color: rgba(245, 158, 11, .32);
  background: var(--ds-bg-warning-subtle);
  color: var(--ds-warning);
}

.status-pill.is-error,
.queue-badge.is-error {
  border-color: rgba(239, 68, 68, .32);
  background: var(--ds-bg-danger-subtle);
  color: var(--ds-error);
}

.status-pill.is-neutral,
.queue-badge.is-neutral {
  background: var(--ds-bg-surface-2);
  color: var(--ds-text-tertiary);
}

.progress-track {
  height: 5px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--ds-bg-surface-2);
}

.progress-track span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: var(--ds-accent);
}

.execution-meta {
  font-size: 12px;
}

.image-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  color: var(--ds-text-secondary);
}

.image-cell span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.resource-cell {
  display: grid;
  gap: 4px;
  font-family: var(--ds-font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
  font-size: 12px;
}

.resource-cell div {
  display: flex;
  gap: 6px;
}

.resource-cell span {
  width: 34px;
  color: var(--ds-text-tertiary);
}

.resource-cell strong {
  color: var(--ds-success);
  font-weight: 650;
}

.label-badge {
  gap: 5px;
  border-color: rgba(59, 130, 246, .28);
  background: var(--ds-bg-info-subtle);
  color: var(--ds-accent);
}

.tags-popover {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-width: 360px;
}

.tag-item,
.tag-empty {
  color: var(--ds-text-secondary);
  font-size: 12px;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  padding: 14px;
  border-top: 1px solid var(--ds-border-subtle);
}

.side-panel {
  display: grid;
  align-content: start;
  gap: 12px;
}

.panel-block {
  padding: 14px;
}

.panel-head {
  justify-content: space-between;
  margin-bottom: 10px;
}

.queue-list,
.summary-list {
  display: grid;
  gap: 8px;
  margin-top: 12px;
}

.queue-item,
.summary-row {
  width: 100%;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 8px;
  background: var(--ds-bg-surface-2);
  color: var(--ds-text-secondary);
  text-align: left;
}

.queue-item {
  display: grid;
  gap: 5px;
  padding: 10px;
}

.queue-item:hover,
.summary-row:hover {
  border-color: var(--ds-border-default);
  background: var(--ds-bg-surface-3);
}

.queue-main {
  color: var(--ds-text-primary);
  font-weight: 650;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 9px 10px;
}

.summary-row strong {
  color: var(--ds-text-primary);
}

.runtime-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-top: 12px;
}

.runtime-grid div {
  border: 1px solid var(--ds-border-subtle);
  border-radius: 8px;
  padding: 10px;
  background: var(--ds-bg-surface-2);
}

.runtime-grid strong {
  display: block;
  margin-top: 6px;
  color: var(--ds-text-primary);
  font-size: 20px;
}

.empty-state {
  padding: 16px;
  border: 1px dashed var(--ds-border-default);
  border-radius: 8px;
  color: var(--ds-text-tertiary);
  text-align: center;
}

.yaml-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.yaml-error {
  margin-top: 8px;
}

.example-block {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.example-block pre {
  flex: 1;
  margin: 0;
  color: var(--ds-text-secondary);
  font-size: 12px;
}

.code-input :deep(.el-textarea__inner) {
  font-family: var(--ds-font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
}

.job-management :deep(.el-table),
.job-management :deep(.el-table__inner-wrapper),
.job-management :deep(.el-table__header-wrapper),
.job-management :deep(.el-table__body-wrapper),
.job-management :deep(.el-table tr),
.job-management :deep(.el-table td.el-table__cell),
.job-management :deep(.el-table th.el-table__cell),
.job-management :deep(.el-table__empty-block) {
  background: var(--ds-bg-surface) !important;
  color: var(--ds-text-secondary) !important;
  border-color: var(--ds-border-subtle) !important;
}

.job-management :deep(.el-table th.el-table__cell) {
  background: var(--ds-bg-surface-2) !important;
  color: var(--ds-text-tertiary) !important;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

.job-management :deep(.el-table__row:hover > td.el-table__cell) {
  background: var(--ds-bg-surface-2) !important;
}

.job-management :deep(.el-dialog),
.job-management :deep(.el-dialog__header),
.job-management :deep(.el-dialog__body),
.job-management :deep(.el-dialog__footer),
.job-management :deep(.el-alert) {
  background: var(--ds-bg-surface) !important;
  color: var(--ds-text-secondary) !important;
  border-color: var(--ds-border-subtle) !important;
}

.job-management :deep(.el-input__wrapper),
.job-management :deep(.el-select__wrapper),
.job-management :deep(.el-textarea__inner) {
  background: var(--ds-bg-surface-2) !important;
  box-shadow: none;
  border-color: var(--ds-border-default) !important;
  color: var(--ds-text-secondary) !important;
}

@media (max-width: 1280px) {
  .job-layout {
    grid-template-columns: 1fr;
  }

  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .job-header,
  .panel-toolbar,
  .toolbar-controls,
  .header-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .namespace-select,
  .filter-input {
    width: 100%;
  }

  .metric-grid {
    grid-template-columns: 1fr;
  }
}

/* k8s-table-cleanup: keep dense workload tables tidy on dark UI */
.autoops-table-wrapper {
  overflow: hidden;
}

.autoops-table :deep(.el-table__body-wrapper),
.autoops-table :deep(.el-scrollbar__wrap) {
  background: var(--ds-bg-surface, #161b22);
}

.autoops-table :deep(.el-table__fixed-right),
.autoops-table :deep(.el-table__fixed-right-patch),
.autoops-table :deep(.el-table-fixed-column--right) {
  background: var(--ds-bg-surface, #161b22) !important;
  box-shadow: none !important;
}

.autoops-table :deep(.el-table__fixed-right::before),
.autoops-table :deep(.el-table__inner-wrapper::before),
.autoops-table :deep(.el-table__inner-wrapper::after) {
  display: none !important;
}

.autoops-table :deep(.el-table__cell) {
  overflow: hidden;
}

.workload-actions,
.node-actions,
.action-buttons,
.actions-cell {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  width: 100%;
  min-width: 0;
  white-space: nowrap;
}

.workload-actions .el-button,
.node-actions .el-button,
.action-buttons .el-button,
.actions-cell .el-button,
.autoops-action-btn {
  flex: 0 0 auto;
  min-width: 30px;
  height: 30px;
  padding: 0 8px;
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 6px !important;
  background: var(--ds-bg-surface-2, #1c212b) !important;
  box-shadow: none !important;
  color: var(--ds-text-secondary, #c9d1d9) !important;
}

.workload-actions .el-button + .el-button,
.node-actions .el-button + .el-button,
.action-buttons .el-button + .el-button,
.actions-cell .el-button + .el-button {
  margin-left: 0;
}

.image-cell,
.resource-cell,
.workload-meta,
.node-meta,
.label-summary,
.labels-cell {
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
}

.image-cell,
.resource-cell,
.workload-meta,
.node-meta {
  white-space: nowrap;
  text-overflow: ellipsis;
}

.label-summary,
.labels-cell {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  max-width: 100%;
}

.label-summary .el-tag,
.labels-cell .el-tag {
  max-width: 96px;
  overflow: hidden;
  text-overflow: ellipsis;
}


/* k8s-table-full-width-fix: avoid fixed right column artifacts */
.autoops-table,
.autoops-table :deep(.el-table__inner-wrapper),
.autoops-table :deep(.el-table__body-wrapper),
.autoops-table :deep(.el-scrollbar),
.autoops-table :deep(.el-scrollbar__wrap),
.autoops-table :deep(.el-scrollbar__view),
.autoops-table :deep(table) {
  width: 100% !important;
}

.autoops-table :deep(.el-table__body),
.autoops-table :deep(.el-table__header) {
  min-width: 100% !important;
}


/* k8s-list-full-row: list owns the full row; insights move below */
.workload-layout,
.job-layout,
.cronjob-layout,
.nodes-layout {
  display: grid !important;
  grid-template-columns: minmax(0, 1fr) !important;
  gap: 16px !important;
  width: 100% !important;
}

.inventory-panel,
.table-panel,
.nodes-main,
.main-panel {
  width: 100% !important;
  min-width: 0 !important;
}

.side-panel {
  display: grid !important;
  grid-template-columns: repeat(3, minmax(0, 1fr)) !important;
  gap: 16px !important;
  width: 100% !important;
  min-width: 0 !important;
}

.side-panel > * {
  min-width: 0 !important;
}

@media (max-width: 1280px) {
  .side-panel {
    grid-template-columns: 1fr !important;
  }
}


/* k8s-minimal-list: remove side noise and keep inventory focused */
.panel-toolbar {
  display: flex !important;
  align-items: center !important;
  justify-content: space-between !important;
  gap: 16px !important;
  padding: 16px 18px !important;
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-bottom: 0 !important;
  border-radius: 8px 8px 0 0 !important;
  background: var(--ds-bg-surface, #161b22) !important;
  color: var(--ds-text-primary, #f0f6fc) !important;
}

.panel-heading {
  min-width: 0 !important;
}

.panel-title {
  color: var(--ds-text-primary, #f0f6fc) !important;
  font-size: 15px !important;
  font-weight: 700 !important;
  letter-spacing: 0 !important;
  text-transform: none !important;
}

.panel-subtitle {
  margin-top: 4px !important;
  color: var(--ds-text-muted, #8b949e) !important;
  font-size: 13px !important;
}

.toolbar-controls {
  display: flex !important;
  align-items: center !important;
  justify-content: flex-end !important;
  gap: 10px !important;
  flex-wrap: wrap !important;
}

.filter-input,
.filter-select {
  width: 220px !important;
}

.autoops-table-wrapper {
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 0 0 8px 8px !important;
  background: var(--ds-bg-surface, #161b22) !important;
}

.autoops-table :deep(.el-table__empty-block) {
  background: var(--ds-bg-surface, #161b22) !important;
}

@media (max-width: 960px) {
  .panel-toolbar,
  .toolbar-controls {
    align-items: stretch !important;
    flex-direction: column !important;
  }

  .filter-input,
  .filter-select,
  .toolbar-controls .el-button {
    width: 100% !important;
  }
}


/* k8s-clean-list-page: focused table-only management page */
.metric-grid {
  display: none !important;
}

.workload-page,
.node-page,
.page-container,
.k8s-page {
  background: var(--ds-bg-app, #0d1117) !important;
}

.workload-layout,
.job-layout,
.cronjob-layout,
.nodes-layout {
  margin-top: 16px !important;
}

.inventory-panel,
.table-panel,
.nodes-main,
.main-panel {
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 8px !important;
  background: var(--ds-bg-surface, #161b22) !important;
  overflow: hidden !important;
}

.panel-toolbar {
  margin: 0 !important;
  padding: 14px 16px !important;
  border: 0 !important;
  border-bottom: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 0 !important;
  background: var(--ds-bg-surface, #161b22) !important;
  background-color: var(--ds-bg-surface, #161b22) !important;
  background-image: none !important;
  box-shadow: none !important;
}

.panel-toolbar::before,
.panel-toolbar::after {
  display: none !important;
}

.panel-title {
  color: var(--ds-text-primary, #f0f6fc) !important;
}

.panel-subtitle {
  color: var(--ds-text-muted, #8b949e) !important;
}

.autoops-table-wrapper {
  border: 0 !important;
  border-radius: 0 !important;
  background: var(--ds-bg-surface, #161b22) !important;
}

.autoops-table,
.autoops-table :deep(.el-table),
.autoops-table :deep(.el-table__inner-wrapper),
.autoops-table :deep(.el-table__body-wrapper),
.autoops-table :deep(.el-table__empty-block),
.autoops-table :deep(.el-table__append-wrapper) {
  background: var(--ds-bg-surface, #161b22) !important;
  background-color: var(--ds-bg-surface, #161b22) !important;
}

.autoops-table :deep(.el-pagination),
.autoops-table-wrapper :deep(.el-pagination),
.pagination-wrapper,
.table-pagination {
  background: var(--ds-bg-surface, #161b22) !important;
  color: var(--ds-text-secondary, #c9d1d9) !important;
}

:deep(.el-pagination button),
:deep(.el-pager li),
:deep(.el-pagination .el-select .el-select__wrapper),
:deep(.el-pagination .el-input__wrapper) {
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  background: var(--ds-bg-surface-2, #1c212b) !important;
  color: var(--ds-text-secondary, #c9d1d9) !important;
  box-shadow: none !important;
}

:deep(.el-pager li.is-active) {
  border-color: var(--ds-accent, #3b82f6) !important;
  background: rgba(59,130,246,0.16) !important;
  color: var(--ds-accent, #3b82f6) !important;
}

</style>
