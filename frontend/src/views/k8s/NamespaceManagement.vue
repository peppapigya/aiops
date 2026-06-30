<template>
  <div class="namespace-management">
    <header class="namespace-header">
      <div>
        <div class="eyebrow">Kubernetes 资源范围</div>
        <h1>命名空间管理</h1>
      </div>
      <div class="header-actions">
        <el-button @click="fetchNamespaceList" :loading="loading" :icon="Refresh">刷新</el-button>
        <el-button type="primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:namespace:showcreatedialogtrue')" :icon="Plus">创建</el-button>
      </div>
    </header>

    <section class="namespace-layout">
      <main class="namespace-panel">
        <div class="panel-toolbar">
          <div>
            <span class="panel-title">命名空间清单</span>
            <span class="panel-subtitle">{{ namespaceList.length }} 个作用域 · {{ systemNamespaceCount }} 受保护</span>
          </div>
        </div>

        <el-table
          :data="namespaceList"
          style="width: 100%"
          v-loading="loading"
          height="calc(100vh - 360px)"
          :row-class-name="tableRowClassName"
          :empty-text="emptyText"
        >
          <el-table-column label="命名空间" min-width="240">
            <template #default="{ row }">
              <div class="namespace-name">
                <div class="namespace-avatar"><el-icon><Folder /></el-icon></div>
                <div class="namespace-identity">
                  <span>{{ row.name }}</span>
                  <em>{{ isSystemNamespace(row.name) ? '系统保护' : '租户作用域' }}</em>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="110" align="center">
            <template #default="{ row }">
              <span class="status-pill" :class="`is-${namespaceStatusTone(row.status)}`">{{ row.status || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="标签" min-width="200">
            <template #default="{ row }">
              <div v-if="row.labels && Object.keys(row.labels).length" class="labels-container">
                <span v-for="(value, key) in row.labels" :key="key" class="label-chip">{{ key }}={{ value }}</span>
              </div>
              <span v-else class="muted-text">无标签</span>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="150" align="right">
            <template #default="{ row }">
              <div class="time-info">
                <span>{{ formatDate(row.creationTimestamp) }}</span>
                <em>{{ calculateAge(row.age) }}</em>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="90" align="right" fixed="right">
            <template #default="{ row }">
              <el-button
                link
                type="danger"
                size="small"
                @click="handleDelete(row)"
                :disabled="isSystemNamespace(row.name)"
                v-show="permStore.hasPerm('k8s:namespace:handledelete')"
              >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </main>
    </section>

    <el-dialog v-model="showCreateDialog" title="创建命名空间" width="500px">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入命名空间名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="submitting">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Folder, Refresh, Grid, Lock, PriceTag, CircleCheck } from '@element-plus/icons-vue'
import { getNamespaceList, createNamespace, deleteNamespace } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const namespaceList = ref([])
const createFormRef = ref(null)

const createForm = ref({
  name: ''
})

const createRules = {
  name: [
    { required: true, message: '请输入命名空间名称', trigger: 'blur' },
    { pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, message: '名称只能包含小写字母、数字和连字符', trigger: 'blur' }
  ]
}

const systemNamespaces = ['default', 'kube-system', 'kube-public', 'kube-node-lease']

const emptyText = computed(() => loading.value ? '加载中...' : '暂无命名空间数据')
const systemNamespaceCount = computed(() => namespaceList.value.filter((item) => isSystemNamespace(item.name)).length)
const tenantNamespaceCount = computed(() => Math.max(namespaceList.value.length - systemNamespaceCount.value, 0))
const activeNamespaceCount = computed(() => namespaceList.value.filter((item) => item.status === 'Active').length)
const labelTotal = computed(() => namespaceList.value.reduce((sum, item) => sum + labelCount(item), 0))
const protectedNamespaces = computed(() => namespaceList.value.filter((item) => isSystemNamespace(item.name)))
const labelDensity = computed(() => namespaceList.value
  .map((item) => ({ name: item.name, count: labelCount(item) }))
  .filter((item) => item.count > 0)
  .sort((a, b) => b.count - a.count)
  .slice(0, 8))

const namespaceMetrics = computed(() => [
  { key: 'total', label: '命名空间', value: namespaceList.value.length, meta: '资源作用域', tone: 'info', icon: Grid },
  { key: 'active', label: '活跃', value: activeNamespaceCount.value, meta: '可承载工作负载', tone: 'success', icon: CircleCheck },
  { key: 'system', label: '受保护', value: systemNamespaceCount.value, meta: '系统命名空间', tone: 'warning', icon: Lock },
  { key: 'labels', label: '标签', value: labelTotal.value, meta: '元数据键', tone: 'info', icon: PriceTag }
])

const isSystemNamespace = (name) => systemNamespaces.includes(name)
const labelCount = (namespace) => Object.keys(namespace.labels || {}).length

const namespaceStatusTone = (status) => {
  const toneMap = {
    Active: 'success',
    Terminating: 'warning'
  }
  return toneMap[status] || 'neutral'
}

const getStatusType = (status) => {
  switch (status) {
    case 'Active':
      return 'success'
    case 'Terminating':
      return 'warning'
    default:
      return 'info'
  }
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const calculateAge = (age) => {
  if (!age) return '-'
  const now = Date.now()
  const created = age * 1000
  const diff = now - created

  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))

  if (days > 0) return `${days}天前`
  if (hours > 0) return `${hours}小时前`
  return '刚刚创建'
}

const tableRowClassName = ({ row }) => {
  if (isSystemNamespace(row.name)) return 'system-namespace-row'
  return ''
}

const fetchNamespaceList = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getNamespaceList(instanceId)
    namespaceList.value = response.data?.namespaceList || []
  } catch (error) {
    ElMessage.error('获取命名空间列表失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  if (!createFormRef.value) return

  await createFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const instanceId = getSelectedInstanceId() || '1'
      await createNamespace(createForm.value.name, createForm.value, instanceId)
      ElMessage.success('命名空间创建成功')
      showCreateDialog.value = false
      createForm.value = { name: '' }
      await fetchNamespaceList()
    } catch (error) {
      ElMessage.error('创建命名空间失败: ' + (error.response?.data?.message || error.message || '未知错误'))
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (namespace) => {
  if (isSystemNamespace(namespace.name)) {
    ElMessage.warning('系统命名空间不能删除')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除命名空间 "${namespace.name}" 吗？此操作将删除该命名空间下的所有资源。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const instanceId = getSelectedInstanceId() || '1'
    await deleteNamespace(namespace.name, instanceId)
    ElMessage.success('命名空间删除成功')
    await fetchNamespaceList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除命名空间失败: ' + (error.response?.data?.message || error.message || '未知错误'))
    }
  }
}

onMounted(() => {
  fetchNamespaceList()
})
</script>

<style scoped>
.namespace-management {
  min-height: 100vh;
  padding: 10px;
  background: var(--ds-bg-app);
  color: var(--ds-text-primary);
}

.namespace-header,
.panel-toolbar,
.panel-head,
.header-actions,
.namespace-name {
  display: flex;
  align-items: center;
}

.namespace-header {
  justify-content: space-between;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-lg, 6px);
  background: var(--ds-bg-surface);
}

.eyebrow {
  color: var(--ds-accent);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: .08em;
  text-transform: uppercase;
}

.namespace-header h1 {
  margin: 2px 0 0;
  color: var(--ds-text-primary);
  font-size: 18px;
  font-weight: 700;
}

.header-actions {
  gap: 6px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.metric-card,
.namespace-panel,
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
.muted-text,
.queue-meta,
.namespace-identity em,
.time-info em,
.guardrail-list span {
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

.namespace-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 8px;
  margin-top: 8px;
}

.namespace-panel,
.panel-block {
  overflow: hidden;
}

.panel-toolbar {
  justify-content: space-between;
  gap: 8px;
  padding: 10px;
  border-bottom: 1px solid var(--ds-border-subtle);
}

.panel-title {
  display: block;
  color: var(--ds-text-primary);
  font-size: 13px;
  font-weight: 700;
}

.panel-subtitle {
  display: block;
  margin-top: 2px;
  font-size: 12px;
}

.namespace-name {
  gap: 10px;
}

.namespace-avatar {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid rgba(59, 130, 246, .28);
  border-radius: 7px;
  background: var(--ds-bg-info-subtle);
  color: var(--ds-accent);
}

.namespace-identity span {
  display: block;
  color: var(--ds-text-primary);
  font-weight: 650;
}

.namespace-identity em,
.time-info em {
  display: block;
  margin-top: 2px;
  font-size: 12px;
  font-style: normal;
}

.status-pill,
.queue-badge {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 700;
}

.status-pill.is-success {
  border-color: rgba(34, 197, 94, .32);
  background: var(--ds-bg-success-subtle);
  color: var(--ds-success);
}

.status-pill.is-warning {
  border-color: rgba(245, 158, 11, .32);
  background: var(--ds-bg-warning-subtle);
  color: var(--ds-warning);
}

.status-pill.is-neutral,
.queue-badge {
  background: var(--ds-bg-surface-2);
  color: var(--ds-text-tertiary);
}

.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.label-chip {
  max-width: 180px;
  overflow: hidden;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 999px;
  padding: 2px 8px;
  background: var(--ds-bg-surface-2);
  color: var(--ds-text-secondary);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time-info {
  text-align: right;
}

.time-info span {
  color: var(--ds-text-secondary);
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
}

.queue-item {
  display: grid;
  gap: 5px;
  padding: 10px;
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

.guardrail-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-top: 12px;
}

.guardrail-list div {
  border: 1px solid var(--ds-border-subtle);
  border-radius: 8px;
  padding: 10px;
  background: var(--ds-bg-surface-2);
}

.guardrail-list strong {
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

.namespace-management :deep(.el-table),
.namespace-management :deep(.el-table__inner-wrapper),
.namespace-management :deep(.el-table__header-wrapper),
.namespace-management :deep(.el-table__body-wrapper),
.namespace-management :deep(.el-table tr),
.namespace-management :deep(.el-table td.el-table__cell),
.namespace-management :deep(.el-table th.el-table__cell),
.namespace-management :deep(.el-table__empty-block) {
  background: var(--ds-bg-surface) !important;
  color: var(--ds-text-secondary) !important;
  border-color: var(--ds-border-subtle) !important;
}

.namespace-management :deep(.el-table th.el-table__cell) {
  background: var(--ds-bg-surface-2) !important;
  color: var(--ds-text-tertiary) !important;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

.namespace-management :deep(.el-table__row:hover > td.el-table__cell),
.namespace-management :deep(.system-namespace-row > td.el-table__cell) {
  background: var(--ds-bg-surface-2) !important;
}

.namespace-management :deep(.el-dialog),
.namespace-management :deep(.el-dialog__header),
.namespace-management :deep(.el-dialog__body),
.namespace-management :deep(.el-dialog__footer) {
  background: var(--ds-bg-surface) !important;
  color: var(--ds-text-secondary) !important;
}

.namespace-management :deep(.el-input__wrapper) {
  background: var(--ds-bg-surface-2) !important;
  box-shadow: none;
}

.namespace-management :deep(.el-button--default) {
  background: var(--ds-bg-surface-2);
  color: var(--ds-text-secondary);
  border-color: var(--ds-border-subtle);
}

.namespace-management :deep(.el-button--primary) {
  background: var(--ds-accent);
  color: #fff;
  border-color: transparent;
}

.namespace-management :deep(.el-button--danger.is-link) {
  color: var(--ds-error);
}

.namespace-management :deep(.el-button--danger.is-link:hover) {
  background: var(--ds-bg-danger-subtle);
  color: var(--ds-error);
}

@media (max-width: 1280px) {
  .namespace-layout {
    grid-template-columns: 1fr;
  }

  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .namespace-header {
    align-items: stretch;
    flex-direction: column;
  }

  .metric-grid,
  .guardrail-list {
    grid-template-columns: 1fr;
  }
}

/* k8s-clean-list-page: focused table-only management page */
.metric-grid {
  display: none !important;
}

.workload-page,
.node-page,
.page-container,
.k8s-page,
.namespace-page,
.service-page,
.ingress-page,
.configmap-page,
.secret-page {
  background: var(--ds-bg-app, #0d1117) !important;
}

.workload-layout,
.job-layout,
.cronjob-layout,
.nodes-layout,
.namespace-layout,
.service-layout,
.ingress-layout,
.configmap-layout,
.secret-layout,
.content-layout {
  display: grid !important;
  grid-template-columns: minmax(0, 1fr) !important;
  gap: 16px !important;
  width: 100% !important;
  margin-top: 16px !important;
}

.inventory-panel,
.table-panel,
.nodes-main,
.main-panel,
.content-panel,
.list-panel {
  width: 100% !important;
  min-width: 0 !important;
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 8px !important;
  background: var(--ds-bg-surface, #161b22) !important;
  overflow: hidden !important;
}

.panel-toolbar,
.inventory-header,
.table-toolbar,
.list-toolbar {
  display: flex !important;
  align-items: center !important;
  justify-content: space-between !important;
  gap: 16px !important;
  margin: 0 !important;
  padding: 14px 16px !important;
  border: 0 !important;
  border-bottom: 1px solid var(--ds-border-default, rgba(255,255,255,0.1)) !important;
  border-radius: 0 !important;
  background: var(--ds-bg-surface, #161b22) !important;
  background-color: var(--ds-bg-surface, #161b22) !important;
  background-image: none !important;
  box-shadow: none !important;
  color: var(--ds-text-primary, #f0f6fc) !important;
}

.panel-toolbar::before,
.panel-toolbar::after,
.inventory-header::before,
.inventory-header::after,
.table-toolbar::before,
.table-toolbar::after,
.list-toolbar::before,
.list-toolbar::after {
  display: none !important;
}

.panel-title,
.inventory-title,
.table-title,
.list-title,
.panel-toolbar h2,
.inventory-header h2,
.table-toolbar h2,
.list-toolbar h2 {
  color: var(--ds-text-primary, #f0f6fc) !important;
  font-size: 15px !important;
  font-weight: 700 !important;
  letter-spacing: 0 !important;
  text-transform: none !important;
}

.panel-subtitle,
.inventory-subtitle,
.table-subtitle,
.list-subtitle,
.panel-toolbar p,
.inventory-header p,
.table-toolbar p,
.list-toolbar p {
  color: var(--ds-text-muted, #8b949e) !important;
  font-size: 13px !important;
}

.toolbar-controls,
.filter-actions,
.search-actions {
  display: flex !important;
  align-items: center !important;
  justify-content: flex-end !important;
  gap: 10px !important;
  flex-wrap: wrap !important;
}

.filter-input,
.filter-select,
.search-input {
  width: 220px !important;
}

.autoops-table-wrapper,
.table-wrapper {
  border: 0 !important;
  border-radius: 0 !important;
  background: var(--ds-bg-surface, #161b22) !important;
}

.autoops-table,
.autoops-table :deep(.el-table),
.autoops-table :deep(.el-table__inner-wrapper),
.autoops-table :deep(.el-table__body-wrapper),
.autoops-table :deep(.el-table__empty-block),
.autoops-table :deep(.el-table__append-wrapper),
.table-wrapper :deep(.el-table),
.table-wrapper :deep(.el-table__inner-wrapper),
.table-wrapper :deep(.el-table__body-wrapper),
.table-wrapper :deep(.el-table__empty-block) {
  background: var(--ds-bg-surface, #161b22) !important;
  background-color: var(--ds-bg-surface, #161b22) !important;
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

@media (max-width: 960px) {
  .panel-toolbar,
  .inventory-header,
  .table-toolbar,
  .list-toolbar,
  .toolbar-controls,
  .filter-actions,
  .search-actions {
    align-items: stretch !important;
    flex-direction: column !important;
  }

  .filter-input,
  .filter-select,
  .search-input,
  .toolbar-controls .el-button,
  .filter-actions .el-button,
  .search-actions .el-button {
    width: 100% !important;
  }
}

</style>
