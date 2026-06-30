<template>
  <div class="pipeline-page">
    <header class="pipeline-header">
      <div class="title-row">
        <h1>流水线</h1>
        <span>共 {{ totalPipelines }} 条</span>
      </div>
      <div class="header-actions">
        <el-input
          v-model="searchQuery"
          class="search-input"
          size="small"
          placeholder="搜索"
          clearable
          @keyup.enter="refreshData"
          @change="refreshData"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button size="small" :loading="loading" @click="refreshData">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button size="small" type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新建
        </el-button>
      </div>
    </header>

    <section class="pipeline-layout">
      <div class="panel list-panel">
        <div class="panel-header">
          <h2>部署记录</h2>
          <div class="view-switch">
            <button :class="{ active: viewMode === 'list' }" type="button" @click="viewMode = 'list'"><el-icon><List /></el-icon></button>
            <button :class="{ active: viewMode === 'grid' }" type="button" @click="viewMode = 'grid'"><el-icon><Grid /></el-icon></button>
          </div>
        </div>

        <div v-if="viewMode === 'grid'" class="pipeline-grid" v-loading="loading">
          <div v-if="filteredPipelines.length === 0" class="empty-row">No pipelines</div>
          <article v-for="item in filteredPipelines" :key="item.id" class="pipeline-card" @click="handleHistory(item)">
            <div class="card-top">
              <span class="status-dot" :class="getStatusClass(item.last_run_status)"></span>
              <strong>{{ item.name }}</strong>
              <el-dropdown trigger="click" @click.stop>
                <button class="icon-action" type="button"><el-icon><MoreFilled /></el-icon></button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleEdit(item)">编辑</el-dropdown-item>
                    <el-dropdown-item @click="handleDelete(item)" class="text-danger">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
            <div class="card-meta">
              <span>{{ item.branch || 'main' }}</span>
              <span>{{ formatTime(item.last_run_time || item.updated_at) }}</span>
            </div>
            <div class="card-repo">
              <el-icon><Connection /></el-icon>
              <span>{{ item.gitUrl || '-' }}</span>
            </div>
            <div class="card-footer">
              <span class="status-pill" :class="statusTone(item.last_run_status)">{{ item.last_run_status || 'Pending' }}</span>
              <el-button link size="small" type="primary" @click.stop="handleRun(item)">运行</el-button>
            </div>
          </article>
        </div>

        <div v-else class="pipeline-table" v-loading="loading">
          <el-table :data="filteredPipelines" style="width: 100%">
            <el-table-column prop="name" label="名称" min-width="220">
              <template #default="{ row }">
                <div class="name-cell">
                  <span class="status-dot" :class="getStatusClass(row.last_run_status)"></span>
                  <div>
                    <strong>{{ row.name }}</strong>
                    <small>{{ row.description || '-' }}</small>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="gitUrl" label="仓库" min-width="240" show-overflow-tooltip>
              <template #default="{ row }">
                <div class="repo-cell"><el-icon><Connection /></el-icon><span>{{ row.gitUrl || '-' }}</span></div>
              </template>
            </el-table-column>
            <el-table-column prop="last_run_status" label="状态" width="120">
              <template #default="{ row }">
                <span class="status-pill" :class="statusTone(row.last_run_status)">{{ row.last_run_status || 'Pending' }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="updated_at" label="更新时间" width="150">
              <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="240" align="center">
              <template #default="{ row }">
                <el-button-group>
                  <el-button size="small" @click="handleRun(row)"><el-icon><VideoPlay /></el-icon></el-button>
                  <el-button size="small" @click="handleHistory(row)"><el-icon><Timer /></el-icon></el-button>
                  <el-button size="small" @click="handleEdit(row)"><el-icon><Edit /></el-icon></el-button>
                  <el-button size="small" type="danger" plain @click="handleDelete(row)"><el-icon><Delete /></el-icon></el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getPipelines, triggerPipeline, deletePipeline } from '@/api/cicd.js'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import {
  Monitor,
  VideoPlay,
  TrendCharts,
  Search,
  List,
  Grid,
  Plus,
  Refresh,
  MoreFilled,
  Connection,
  Timer,
  Edit,
  Delete,
  CircleClose
} from '@element-plus/icons-vue'
import dayjs from 'dayjs'

const router = useRouter()
const viewMode = ref('list')
const loading = ref(false)
const searchQuery = ref('')
const pipelines = ref([])
const totalPipelines = ref(0)
const page = ref(1)
const pageSize = ref(20)

const activeRuns = computed(() => pipelines.value.filter((item) => item.last_run_status === 'Running').length)
const failedRuns = computed(() => pipelines.value.filter((item) => item.last_run_status === 'Failed').length)
const successRate = computed(() => {
  const finished = pipelines.value.filter((item) => ['Succeeded', 'Failed'].includes(item.last_run_status))
  if (finished.length === 0) return 100
  return Math.round((finished.filter((item) => item.last_run_status === 'Succeeded').length / finished.length) * 100)
})
const filteredPipelines = computed(() => pipelines.value)
const recentPipelines = computed(() => [...pipelines.value].slice(0, 8))

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getPipelines({ pageNum: page.value, pageSize: pageSize.value, keyword: searchQuery.value })
    pipelines.value = res.data?.data?.data || []
    totalPipelines.value = res.data?.data?.total || pipelines.value.length
  } catch (e) {
    ElMessage.error('load failed')
  } finally {
    loading.value = false
  }
}

const refreshData = () => {
  fetchData()
}

const formatTime = (time) => {
  if (!time) return '-'
  return dayjs(time).format('MM-DD HH:mm')
}

const getStatusClass = (status) => {
  const map = {
    Succeeded: 'status-success',
    Failed: 'status-failed',
    Running: 'status-running',
    Pending: 'status-pending'
  }
  return map[status] || 'status-pending'
}

const statusTone = (status) => {
  const map = {
    Succeeded: 'success',
    Failed: 'error',
    Running: 'warning',
    Pending: 'neutral'
  }
  return map[status] || 'neutral'
}

const handleCreate = () => {
  router.push('/cicd/pipelines/create')
}

const handleEdit = (row) => {
  router.push(`/cicd/pipelines/${row.id}/edit`)
}

const handleRun = async (row) => {
  try {
    await triggerPipeline(row.id, {})
    ElNotification({ title: 'Pipeline', message: `${row.name} started`, type: 'success', duration: 3000 })
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || 'run failed')
  }
}

const handleHistory = (row) => {
  router.push(`/cicd/pipelines/${row.id}/runs`)
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`Delete ${row.name}?`, 'Delete', { confirmButtonText: 'Delete', cancelButtonText: 'Cancel', type: 'warning' })
    await deletePipeline(row.id)
    ElMessage.success('deleted')
    fetchData()
  } catch (e) {
    // cancelled
  }
}

onMounted(fetchData)
</script>

<style scoped>
.pipeline-page {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 8px;
  color: var(--ds-text-primary);
}

.pipeline-header,
.metric-card,
.panel {
  border: 1px solid var(--ds-border-default);
  border-radius: 8px;
  background: var(--ds-bg-surface);
}

.pipeline-header {
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
.name-cell small,
.card-meta,
.card-repo {
  color: var(--ds-text-muted);
  font-size: 11px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.search-input {
  width: 180px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.metric-card {
  min-height: 92px;
  padding: 12px;
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
  margin-top: 12px;
  font-size: 26px;
  font-weight: 600;
  line-height: 1;
}

.metric-card.is-success { border-color: rgba(34, 197, 94, .28); }
.metric-card.is-warning { border-color: rgba(245, 158, 11, .28); }
.metric-card.is-error { border-color: rgba(239, 68, 68, .28); }
.metric-card.is-info { border-color: rgba(59, 130, 246, .28); }

.pipeline-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
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
  font-size: 11px;
  font-weight: 600;
}

.view-switch {
  display: flex;
  gap: 2px;
}

.view-switch button,
.icon-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 24px;
  border: 1px solid transparent;
  border-radius: 5px;
  color: var(--ds-text-tertiary);
  background: transparent;
  cursor: pointer;
}

.view-switch button.active,
.view-switch button:hover,
.icon-action:hover {
  color: var(--ds-text-primary);
  border-color: var(--ds-border-default);
  background: var(--ds-bg-hover);
}

.pipeline-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 6px;
  padding: 8px;
}

.pipeline-card {
  display: grid;
  gap: 6px;
  min-height: 110px;
  padding: 8px;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 6px;
  background: var(--ds-bg-surface-2);
  cursor: pointer;
  transition: var(--ds-transition-fast);
}

.pipeline-card:hover,
.run-row:hover {
  border-color: var(--ds-border-strong);
  background: var(--ds-bg-hover);
}

.card-top,
.card-footer,
.name-cell,
.repo-cell,
.run-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-top strong {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  color: var(--ds-text-secondary);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-meta,
.card-repo {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 6px;
}

.card-repo span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-footer {
  justify-content: space-between;
  border-top: 1px solid var(--ds-border-subtle);
  padding-top: 4px;
}

.pipeline-table {
  padding: 0;
}

.name-cell div,
.row-main {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.name-cell strong,
.row-main strong {
  overflow: hidden;
  color: var(--ds-text-secondary);
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex: none;
  background: var(--ds-text-muted);
}

.status-success { background: var(--ds-success); }
.status-failed { background: var(--ds-error); }
.status-running { background: var(--ds-warning); }
.status-pending { background: var(--ds-text-muted); }

.status-pill {
  height: 20px;
  padding: 0 8px;
  border-radius: 999px;
  font-size: 11px;
  line-height: 20px;
}

.status-pill.success { color: var(--ds-success); background: var(--ds-bg-success-subtle); }
.status-pill.warning { color: var(--ds-warning); background: var(--ds-bg-warning-subtle); }
.status-pill.error { color: var(--ds-error); background: var(--ds-bg-danger-subtle); }
.status-pill.neutral { color: var(--ds-text-tertiary); background: var(--ds-bg-surface-3); }

.side-panel {
  align-self: start;
}

.run-list {
  padding: 8px;
}

.run-row {
  width: 100%;
  min-height: 42px;
  padding: 0 8px;
  border: 1px solid transparent;
  border-radius: 6px;
  color: var(--ds-text-primary);
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.row-main {
  flex: 1;
}

.empty-row {
  padding: 28px 12px;
  color: var(--ds-text-muted);
  font-size: 12px;
  text-align: center;
}

.text-danger {
  color: var(--ds-error);
}

@media (max-width: 1280px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .pipeline-layout {
    grid-template-columns: 1fr;
  }
}

/* cicd-clean-list: full-width table, no side noise */
.metric-grid {
  display: none !important;
}

.pipeline-layout {
  display: grid !important;
  grid-template-columns: minmax(0, 1fr) !important;
  gap: 8px !important;
  width: 100% !important;
}

.pipeline-page {
  background: var(--ds-bg-app, #0d1117) !important;
}

.pipeline-header {
  background: var(--ds-bg-surface, #161b22) !important;
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 8px !important;
  padding: 8px 12px !important;
  margin-bottom: 8px !important;
}

.list-panel {
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 8px !important;
  background: var(--ds-bg-surface, #161b22) !important;
  overflow: hidden !important;
}

.panel-header {
  padding: 8px 12px !important;
  border-bottom: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  background: var(--ds-bg-surface, #161b22) !important;
  color: var(--ds-text-primary, #f0f6fc) !important;
}

.panel-header h3 {
  color: var(--ds-text-primary, #f0f6fc) !important;
  font-size: 13px !important;
  font-weight: 700 !important;
}

.pipeline-table,
.pipeline-grid,
.pipeline-card {
  background: var(--ds-bg-surface, #161b22) !important;
}

.el-table :deep(.el-table__inner-wrapper),
.el-table :deep(.el-table__body-wrapper),
.el-table :deep(.el-table__empty-block) {
  background: var(--ds-bg-surface, #161b22) !important;
  background-color: var(--ds-bg-surface, #161b22) !important;
}

/* actions row: flat buttons inline */
.actions-row {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  gap: 6px !important;
  white-space: nowrap !important;
}

.actions-row .el-button,
.actions-row .el-button--small {
  flex: 0 0 auto !important;
  height: 26px !important;
  padding: 0 8px !important;
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 6px !important;
  background: var(--ds-bg-surface-2, #1c212b) !important;
  background-color: var(--ds-bg-surface-2, #1c212b) !important;
  box-shadow: none !important;
  color: var(--ds-text-secondary, #c9d1d9) !important;
}

.actions-row .el-button:hover,
.actions-row .el-button--small:hover {
  background: var(--ds-bg-hover, rgba(255,255,255,0.04)) !important;
  color: var(--ds-text-primary, #f0f6fc) !important;
}

.actions-row .el-button.is-link,
.actions-row .el-button--small.is-link {
  border-color: transparent !important;
  background: transparent !important;
}

.actions-row .el-button--danger.is-link {
  color: var(--ds-error, #ef4444) !important;
}

@media (max-width: 960px) {
  .pipeline-layout {
    grid-template-columns: 1fr !important;
  }
}

</style>
