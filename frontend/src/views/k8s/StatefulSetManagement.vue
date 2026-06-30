<template>
  <div class="statefulset-management">
    <header class="workload-header">
      <div>
        <div class="eyebrow">Kubernetes Stateful Workloads</div>
        <h1>StatefulSets</h1>
      </div>
      <div class="header-actions">
        <el-select v-model="selectedNamespace" placeholder="Namespace" filterable class="namespace-select">
          <el-option label="All namespaces" value="all" />
          <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
        </el-select>
        <el-button @click="fetchData" :loading="loading" :icon="RefreshRight">Refresh</el-button>
        <el-button type="success" :icon="Monitor">Dashboard</el-button>
        <el-button type="primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:statefulset:showcreatedialogtrue')" :icon="Plus">Create</el-button>
      </div>
    </header>

    <section class="workload-layout">
      <main class="workload-panel">
        <div class="panel-toolbar">
          <div>
            <span class="panel-title">Stateful Inventory</span>
            <span class="panel-subtitle">{{ filteredStatefulSets.length }} visible · {{ statefulSetList.length }} total</span>
          </div>
          <div class="toolbar-controls">
            <el-input
              v-model="searchKeyword"
              placeholder="Search statefulset"
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
          :data="filteredStatefulSets"
          style="width: 100%"
          v-loading="loading"
          @row-click="handleViewDetail"
        >
          <el-table-column label="Workload" min-width="240">
            <template #default="{ row }">
              <div class="workload-name-cell">
                <div class="workload-avatar"><el-icon><Coin /></el-icon></div>
                <div class="workload-identity">
                  <button class="name-link" @click.stop="handleViewDetail(row)">{{ row.name }}</button>
                  <span>{{ row.namespace || selectedNamespace }} · StatefulSet</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="Stability" min-width="180">
            <template #default="{ row }">
              <div class="rollout-cell">
                <div class="rollout-top">
                  <span :class="['status-pill', `is-${statefulSetTone(row)}`]">{{ statefulSetStatus(row) }}</span>
                  <strong>{{ row.ready || 0 }}/{{ row.replicas || 0 }}</strong>
                </div>
                <div class="progress-track">
                  <span :style="{ width: `${readyProgress(row)}%` }" />
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="Resources" min-width="170">
            <template #default>
              <div class="resource-cell">
                <div><span>CPU</span><strong>200m</strong><em>/ 500m</em></div>
                <div><span>MEM</span><strong>300Mi</strong><em>/ 500Mi</em></div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="Image" min-width="240">
            <template #default="{ row }">
              <div class="image-cell">
                <el-icon><Box /></el-icon>
                <span :title="primaryImage(row)">{{ primaryImage(row) }}</span>
                <b v-if="row.containers && row.containers.length > 1">+{{ row.containers.length - 1 }}</b>
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

          <el-table-column label="Created" width="170">
            <template #default="{ row }">
              <span class="time-text">{{ formatDate(row.created) }}</span>
            </template>
          </el-table-column>

          <el-table-column label="Actions" width="240">
            <template #default="{ row }">
              <div class="workload-actions" @click.stop>
                <el-button link type="primary" size="small" @click="handleScale(row)">Scale</el-button>
                <el-button link type="primary" size="small" @click="handleViewDetail(row)" v-show="permStore.hasPerm('k8s:statefulset:handleviewdetail')">Detail</el-button>
                <el-button link type="primary" size="small" @click="handleUpdate(row)">Image</el-button>
                <el-button link type="primary" size="small" @click="handleLogs(row)" v-show="permStore.hasPerm('k8s:statefulset:handlelogs')">Logs</el-button>
                <el-button link type="danger" size="small" @click="handleDelete(row)" v-show="permStore.hasPerm('k8s:statefulset:handledelete')">Delete</el-button>
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

    <el-dialog v-model="showCreateDialog" title="创建 StatefulSet" width="900px" :close-on-click-modal="false" append-to-body class="autoops-dialog">
      <el-form label-width="100px">
        <el-form-item label="命名空间">
          <el-select v-model="selectedNamespace" placeholder="选择命名空间" style="width: 100%;">
            <el-option label="default" value="default" />
            <el-option label="kube-system" value="kube-system" />
          </el-select>
        </el-form-item>
        <el-form-item label="YAML配置">
          <el-input
            v-model="yamlContent"
            type="textarea"
            :rows="14"
            placeholder="请输入StatefulSet的YAML配置"
            class="code-input"
          />
        </el-form-item>
        <el-alert title="YAML示例" type="info" :closable="false">
          <pre class="yaml-example">apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: example-statefulset
spec:
  serviceName: example
  replicas: 3
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: example
        image: nginx
        ports:
        - containerPort: 80</pre>
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showScaleDialog" title="扩缩容" width="400px" append-to-body>
      <div class="dialog-section">
        <p class="dialog-label">目标副本数</p>
        <el-input-number v-model="scaleForm.replicas" :min="0" :max="50" class="w-full" />
      </div>
      <template #footer>
        <el-button @click="showScaleDialog = false">取消</el-button>
        <el-button type="primary" @click="handleScaleConfirm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
import { ref, onMounted, watch, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshRight, Monitor, Plus, PriceTag, Platform, Box, Coin, CircleCheck, Warning } from '@element-plus/icons-vue'
import { getStatefulSetList, deleteStatefulSet, scaleStatefulSet } from '@/api/k8s/statefulset'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const permStore = usePermissionStore()

const loading = ref(false)
const statefulSetList = ref([])
const searchKeyword = ref('')
const selectedNamespace = ref('default')
const namespaceList = ref([])

const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const showCreateDialog = ref(false)
const yamlContent = ref('')

const showScaleDialog = ref(false)
const scaleForm = ref({
  row: null,
  replicas: 0
})

const filteredStatefulSets = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return statefulSetList.value
  return statefulSetList.value.filter((item) => {
    return [item.name, item.namespace, item.image, primaryImage(item)].filter(Boolean).some((value) => String(value).toLowerCase().includes(keyword))
  })
})

const readySets = computed(() => statefulSetList.value.filter((item) => Number(item.ready || 0) >= Number(item.replicas || 0) && Number(item.replicas || 0) > 0).length)
const totalReplicas = computed(() => statefulSetList.value.reduce((sum, item) => sum + Number(item.replicas || 0), 0))
const readyReplicas = computed(() => statefulSetList.value.reduce((sum, item) => sum + Number(item.ready || 0), 0))
const stabilityQueue = computed(() => filteredStatefulSets.value.filter((item) => statefulSetTone(item) !== 'success').slice(0, 8))
const imageCount = computed(() => new Set(statefulSetList.value.map((item) => primaryImage(item)).filter(Boolean)).size)

const statefulSetMetrics = computed(() => [
  { key: 'total', label: 'StatefulSets', value: statefulSetList.value.length, meta: `${namespaceList.value.length || 1} namespaces`, tone: 'info', icon: Coin },
  { key: 'ready', label: 'Stable', value: readySets.value, meta: 'ordered pods ready', tone: 'success', icon: CircleCheck },
  { key: 'replicas', label: 'Replicas', value: `${readyReplicas.value}/${totalReplicas.value}`, meta: 'ready / desired', tone: readyReplicas.value === totalReplicas.value ? 'success' : 'warning', icon: Platform },
  { key: 'images', label: 'Images', value: imageCount.value, meta: 'unique artifacts', tone: 'info', icon: Box }
])

const namespaceBreakdown = computed(() => {
  const counts = new Map()
  statefulSetList.value.forEach((item) => {
    const namespace = item.namespace || selectedNamespace.value || 'default'
    counts.set(namespace, (counts.get(namespace) || 0) + 1)
  })
  return Array.from(counts.entries()).map(([name, count]) => ({ name, count })).sort((a, b) => b.count - a.count).slice(0, 8)
})

watch(selectedNamespace, () => {
  currentPage.value = 1
  fetchData()
})

const primaryImage = (row) => row.image || row.containers?.[0]?.image || '-'
const readyProgress = (row) => {
  const replicas = Number(row.replicas || 0)
  if (!replicas) return 0
  return Math.min(100, Math.round((Number(row.ready || 0) / replicas) * 100))
}
const statefulSetTone = (row) => {
  const replicas = Number(row.replicas || 0)
  const ready = Number(row.ready || 0)
  if (!replicas) return 'neutral'
  if (ready >= replicas) return 'success'
  if (ready > 0) return 'warning'
  return 'error'
}
const statefulSetStatus = (row) => {
  const tone = statefulSetTone(row)
  if (tone === 'success') return 'Stable'
  if (tone === 'warning') return 'Reconciling'
  if (tone === 'error') return 'Unavailable'
  return 'Scaled 0'
}

const fetchData = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getStatefulSetList(selectedNamespace.value, instanceId)
    statefulSetList.value = res.data?.statefulSetList || []
    total.value = res.data?.total || statefulSetList.value.length
  } catch (e) {
    ElMessage.error('获取 StatefulSet 列表失败')
    statefulSetList.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const fetchNamespaces = async () => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getNamespaceList(instanceId)
    namespaceList.value = res.data?.namespaceList || []
    if (!namespaceList.value.some(ns => ns.name === selectedNamespace.value)) {
      selectedNamespace.value = namespaceList.value[0]?.name || 'default'
    }
  } catch (e) {
    ElMessage.error('获取命名空间列表失败')
  }
}

const handleSearch = () => fetchData()
const handleReset = () => { searchKeyword.value = ''; fetchData() }
const handleSizeChange = () => fetchData()
const selectNamespace = (namespace) => { selectedNamespace.value = namespace; fetchData() }
const formatDate = (ts) => dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'

const handleViewDetail = (row) => ElMessage.info(`查看详情: ${row.name}`)
const handleScale = (row) => {
  scaleForm.value.row = row
  scaleForm.value.replicas = row.replicas
  showScaleDialog.value = true
}
const handleScaleConfirm = async () => {
  try {
    await scaleStatefulSet(scaleForm.value.row.namespace, scaleForm.value.row.name, scaleForm.value.replicas)
    ElMessage.success('扩缩容指令已发送')
    showScaleDialog.value = false
    fetchData()
  } catch(e) { ElMessage.error('操作失败') }
}

const handleUpdate = (row) => ElMessage.info('更新功能待完善')
const handleLogs = (row) => ElMessage.info('日志功能待完善')
const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 ${row.name}?`, '警告', { type: 'warning' })
    .then(async () => {
      await deleteStatefulSet(row.namespace, row.name)
      ElMessage.success('删除成功')
      fetchData()
    })
}

const handleCreate = () => {
  ElMessage.info('功能待完善')
}

onMounted(() => {
  fetchData()
  fetchNamespaces()
})
</script>

<style scoped>
.statefulset-management {
  min-height: 100vh;
  padding: var(--ds-space-16, 16px);
  background: var(--ds-bg-app);
  color: var(--ds-text-primary);
}

.workload-header,
.panel-toolbar,
.panel-head,
.header-actions,
.toolbar-controls,
.workload-actions {
  display: flex;
  align-items: center;
}

.workload-header {
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

.workload-header h1 {
  margin: 4px 0 0;
  color: var(--ds-text-primary);
  font-size: 22px;
  font-weight: 700;
}

.header-actions,
.toolbar-controls,
.workload-actions {
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
.workload-panel,
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
.workload-identity span {
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

.workload-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 12px;
  margin-top: 12px;
}

.workload-panel,
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

.workload-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.workload-avatar {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid rgba(59, 130, 246, .28);
  border-radius: 7px;
  background: var(--ds-bg-info-subtle);
  color: var(--ds-accent);
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

.rollout-cell {
  display: grid;
  gap: 8px;
}

.rollout-top {
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

.image-cell b,
.label-badge {
  flex: none;
  color: var(--ds-accent);
  font-size: 12px;
}

.label-badge {
  gap: 5px;
  border-color: rgba(59, 130, 246, .28);
  background: var(--ds-bg-info-subtle);
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

.dialog-section {
  padding: 12px;
}

.dialog-label {
  margin: 0 0 8px;
  color: var(--ds-text-secondary);
}

.yaml-example {
  margin: 8px 0;
  color: var(--ds-text-secondary);
  font-size: 12px;
}

.code-input :deep(.el-textarea__inner) {
  font-family: var(--ds-font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
}

.statefulset-management :deep(.el-table),
.statefulset-management :deep(.el-table__inner-wrapper),
.statefulset-management :deep(.el-table__header-wrapper),
.statefulset-management :deep(.el-table__body-wrapper),
.statefulset-management :deep(.el-table tr),
.statefulset-management :deep(.el-table td.el-table__cell),
.statefulset-management :deep(.el-table th.el-table__cell),
.statefulset-management :deep(.el-table__empty-block) {
  background: var(--ds-bg-surface) !important;
  color: var(--ds-text-secondary) !important;
  border-color: var(--ds-border-subtle) !important;
}

.statefulset-management :deep(.el-table th.el-table__cell) {
  background: var(--ds-bg-surface-2) !important;
  color: var(--ds-text-tertiary) !important;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

.statefulset-management :deep(.el-table__row:hover > td.el-table__cell) {
  background: var(--ds-bg-surface-2) !important;
}

.statefulset-management :deep(.el-dialog),
.statefulset-management :deep(.el-dialog__header),
.statefulset-management :deep(.el-dialog__body),
.statefulset-management :deep(.el-dialog__footer),
.statefulset-management :deep(.el-alert) {
  background: var(--ds-bg-surface) !important;
  color: var(--ds-text-secondary) !important;
  border-color: var(--ds-border-subtle) !important;
}

.statefulset-management :deep(.el-input__wrapper),
.statefulset-management :deep(.el-select__wrapper),
.statefulset-management :deep(.el-textarea__inner),
.statefulset-management :deep(.el-input-number__decrease),
.statefulset-management :deep(.el-input-number__increase) {
  background: var(--ds-bg-surface-2) !important;
  box-shadow: none;
  border-color: var(--ds-border-default) !important;
  color: var(--ds-text-secondary) !important;
}

@media (max-width: 1280px) {
  .workload-layout {
    grid-template-columns: 1fr;
  }

  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .workload-header,
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
