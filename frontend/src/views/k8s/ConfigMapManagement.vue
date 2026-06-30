<template>
  <div class="configmap-management">
    <section class="config-header">
      <div>
        <div class="eyebrow">Kubernetes 配置管理</div>
        <h1>配置管理</h1>
      </div>
      <div class="header-actions">
        <el-select v-model="selectedNamespace" placeholder="命名空间" @change="fetchData" class="namespace-select">
          <el-option label="全部命名空间" value="all" />
          <el-option
            v-for="ns in namespaceList"
            :key="ns.name"
            :label="ns.name"
            :value="ns.name"
          />
        </el-select>
        <el-button type="primary" class="primary-action" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:configmap:showcreatedialogtrue')">
          <el-icon><Plus /></el-icon>
          创建 ConfigMap
        </el-button>
      </div>
    </section>

    <section class="config-layout">
      <div class="config-panel">
        <div class="panel-head">
          <div>
            <h2>配置清单</h2>
            <span>{{ filteredConfigMaps.length }} 个配置对象</span>
          </div>
          <el-input v-model="searchKeyword" clearable placeholder="搜索名称或命名空间" class="search-input" />
        </div>

        <div class="table-shell">
          <el-table
            :data="filteredConfigMaps"
            class="ops-table"
            v-loading="loading"
            element-loading-background="rgba(13, 17, 23, 0.72)"
            height="calc(100vh - 392px)"
            :empty-text="loading ? '加载中...' : '暂无数据'"
          >
            <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
              <template #default="scope">
                <div class="name-cell">
                  <span>{{ scope.row.name }}</span>
                  <small>{{ scope.row.namespace }}</small>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="dataCount" label="键数" width="100" align="center">
              <template #default="scope">
                <el-tag class="status-tag info" effect="plain">{{ scope.row.dataCount || 0 }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="数据量" min-width="160">
              <template #default="scope">
                <div class="density-cell">
                  <div class="density-track">
                    <span :style="{ width: `${dataDensity(scope.row)}%` }" />
                  </div>
                  <small>{{ dataDensity(scope.row) }}%</small>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="150">
              <template #default="scope">
                {{ formatTimestamp(scope.row.age) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="scope">
                <div class="action-group">
                  <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:configmap:handleviewdetail')">详情</el-button>
                  <el-button link type="info" size="small" @click="handleEdit(scope.row)" v-show="permStore.hasPerm('k8s:configmap:handleedit')">编辑</el-button>
                  <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:configmap:handledelete')">删除</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </section>

    <el-dialog v-model="showCreateDialog" :title="isEditMode ? '编辑 ConfigMap' : '创建 ConfigMap'" width="900px" :close-on-click-modal="false" class="ops-dialog">
      <el-form :model="createForm" ref="createFormRef" label-width="120px">
        <el-tabs v-model="activeTab" type="border-card" class="ops-tabs">
          <el-tab-pane label="表单创建" name="form">
            <el-form-item label="命名空间" prop="namespace" required>
              <el-select v-model="createForm.namespace" style="width: 100%;">
                <el-option
                  v-for="ns in namespaceList"
                  :key="ns.name"
                  :label="ns.name"
                  :value="ns.name"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" prop="name" required>
              <el-input v-model="createForm.name" :disabled="isEditMode" />
            </el-form-item>

            <el-divider content-position="left">数据项</el-divider>
            <div class="data-editor">
              <div v-for="(keyName, index) in dataKeys" :key="index" class="data-item">
                <el-row :gutter="10">
                  <el-col :span="7">
                    <el-input v-model="dataKeys[index]" placeholder="键" @blur="updateDataKey(index, dataKeys[index])" />
                  </el-col>
                  <el-col :span="15">
                    <el-input v-model="createForm.data[keyName]" type="textarea" :rows="2" placeholder="值" />
                  </el-col>
                  <el-col :span="2">
                    <el-button link type="danger" @click="removeDataItem(keyName)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </el-col>
                </el-row>
              </div>
              <el-button type="primary" plain @click="addDataItem" style="margin-top: 10px;">
                <el-icon><Plus /></el-icon>
                添加数据项
              </el-button>
            </div>
          </el-tab-pane>

          <el-tab-pane label="YAML创建" name="yaml">
            <el-input
              v-model="createForm.yaml"
              type="textarea"
              :rows="20"
              placeholder="请输入 YAML 配置"
            />
          </el-tab-pane>
        </el-tabs>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleCloseDialog">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ isEditMode ? '更新' : '创建' }}</el-button>
        </span>
      </template>
    </el-dialog>

    <el-dialog v-model="showDetailDialog" title="ConfigMap 详情" width="800px" class="ops-dialog">
      <div v-if="currentConfigMap" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ currentConfigMap.metadata.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ currentConfigMap.metadata.namespace }}</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatTimestamp(currentConfigMap.metadata.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">数据项</el-divider>
        <el-table :data="formatDataToTable(currentConfigMap.data)" class="ops-table compact-table">
          <el-table-column prop="key" label="键" width="250" />
          <el-table-column prop="value" label="值" show-overflow-tooltip />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { computed, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Coin, Delete, Document, Files, Plus } from '@element-plus/icons-vue'
import { getConfigMapList, getConfigMapDetail, createConfigMap, updateConfigMap, deleteConfigMap } from '@/api/k8s/configmap'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const isEditMode = ref(false)
const editingConfigMap = ref(null)
const configMapList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
const currentConfigMap = ref(null)
const activeTab = ref('form')
const dataKeys = ref([])
const searchKeyword = ref('')

const createForm = ref({
  name: '',
  namespace: 'default',
  data: {},
  yaml: ''
})

const filteredConfigMaps = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return configMapList.value
  return configMapList.value.filter((item) => {
    return [item.name, item.namespace]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(keyword))
  })
})

const totalDataKeys = computed(() => configMapList.value.reduce((sum, item) => sum + Number(item.dataCount || 0), 0))
const averageDataKeys = computed(() => configMapList.value.length ? Math.round(totalDataKeys.value / configMapList.value.length) : 0)
const emptyConfigMaps = computed(() => configMapList.value.filter((item) => !Number(item.dataCount || 0)).length)
const largeConfigMaps = computed(() => filteredConfigMaps.value.filter((item) => Number(item.dataCount || 0) >= 5).slice(0, 8))

const namespaceBreakdown = computed(() => {
  const namespaceMap = new Map()
  configMapList.value.forEach((item) => {
    namespaceMap.set(item.namespace, (namespaceMap.get(item.namespace) || 0) + 1)
  })
  return Array.from(namespaceMap.entries())
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 8)
})

const configMetrics = computed(() => [
  { key: 'total', label: '配置项', value: configMapList.value.length, meta: `${namespaceList.value.length || 1} 个命名空间`, tone: 'info', icon: Files },
  { key: 'keys', label: '数据键', value: totalDataKeys.value, meta: '配置条目', tone: 'success', icon: Coin },
  { key: 'large', label: '大型', value: largeConfigMaps.value.length, meta: '5+ 键对象', tone: largeConfigMaps.value.length ? 'warning' : 'success', icon: Document },
  { key: 'empty', label: '空', value: emptyConfigMaps.value, meta: '待清理', tone: emptyConfigMaps.value ? 'warning' : 'success', icon: Delete }
])

const dataDensity = (row) => {
  const maxCount = Math.max(...configMapList.value.map((item) => Number(item.dataCount || 0)), 1)
  return Math.round((Number(row.dataCount || 0) / maxCount) * 100)
}

const formatTimestamp = (timestamp) => {
  if (typeof timestamp === 'number') {
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
  }
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const formatDataToTable = (data) => {
  if (!data) return []
  return Object.keys(data).map(key => ({
    key,
    value: data[key]
  }))
}

const fetchNamespaces = async () => {
  const instanceId = getSelectedInstanceId()
  const res = await getNamespaceList(instanceId)
  if (res.status === 200) {
    namespaceList.value = res.data.namespaceList
  }
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return

  loading.value = true
  try {
    const res = await getConfigMapList(selectedNamespace.value, instanceId)
    if (res.status === 200) {
      configMapList.value = res.data.configMapList
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const addDataItem = () => {
  const newKey = `key${Object.keys(createForm.value.data).length + 1}`
  createForm.value.data[newKey] = ''
  dataKeys.value.push(newKey)
}

const removeDataItem = (key) => {
  delete createForm.value.data[key]
  const index = dataKeys.value.indexOf(key)
  if (index > -1) {
    dataKeys.value.splice(index, 1)
  }
}

const updateDataKey = (index, newKey) => {
  const oldKey = dataKeys.value[index]
  if (oldKey !== newKey && createForm.value.data.hasOwnProperty(oldKey)) {
    createForm.value.data[newKey] = createForm.value.data[oldKey]
    delete createForm.value.data[oldKey]
    dataKeys.value[index] = newKey
  }
}

const resetForm = () => {
  createForm.value = {
    name: '',
    namespace: 'default',
    data: {},
    yaml: ''
  }
  dataKeys.value = []
  isEditMode.value = false
  editingConfigMap.value = null
  activeTab.value = 'form'
}

const handleCloseDialog = () => {
  showCreateDialog.value = false
  resetForm()
}

const handleCreate = async () => {
  const instanceId = getSelectedInstanceId()
  submitting.value = true
  try {
    await createConfigMap(createForm.value, instanceId)
    ElMessage.success('创建成功')
    handleCloseDialog()
    fetchData()
  } catch (error) {
    console.error(error)
    ElMessage.error('创建失败')
  } finally {
    submitting.value = false
  }
}

const handleUpdate = async () => {
  const instanceId = getSelectedInstanceId()
  submitting.value = true
  try {
    await updateConfigMap(
      editingConfigMap.value.namespace,
      editingConfigMap.value.name,
      createForm.value,
      instanceId
    )
    ElMessage.success('更新成功')
    handleCloseDialog()
    fetchData()
  } catch (error) {
    console.error(error)
    ElMessage.error('更新失败')
  } finally {
    submitting.value = false
  }
}

const handleSubmit = () => {
  if (isEditMode.value) {
    handleUpdate()
  } else {
    handleCreate()
  }
}

const handleEdit = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getConfigMapDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    const configMap = res.data.configMapDetail
    editingConfigMap.value = row
    isEditMode.value = true

    createForm.value = {
      name: configMap.metadata.name,
      namespace: configMap.metadata.namespace,
      data: configMap.data || {},
      yaml: ''
    }

    dataKeys.value = Object.keys(configMap.data || {})

    showCreateDialog.value = true
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 ConfigMap ${row.name} 吗?`, '提示', {
    type: 'warning'
  }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteConfigMap(row.namespace, row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getConfigMapDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    currentConfigMap.value = res.data.configMapDetail
    showDetailDialog.value = true
  }
}

onMounted(() => {
  fetchNamespaces()
  fetchData()
})
</script>

<style scoped>
.configmap-management {
  min-height: 100vh;
  padding: 10px;
  color: var(--ds-text-primary, #f8fafc);
  background: var(--ds-bg-page, #0d1117);
}

.config-header,
.metric-card,
.config-panel,
.side-card {
  border: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  border-radius: 8px;
  background: var(--ds-bg-card, #161b22);
}

.config-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
  padding: 12px;
}

.eyebrow {
  margin-bottom: 6px;
  color: var(--ds-color-info, #3b82f6);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.config-header h1,
.panel-head h2,
.side-head h3 {
  margin: 0;
  color: var(--ds-text-primary, #f8fafc);
}

.config-header h1 {
  font-size: 18px;
  line-height: 1.2;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.namespace-select {
  width: 220px;
}

.primary-action {
  border: 1px solid rgba(59, 130, 246, 0.65);
  background: var(--ds-color-info, #3b82f6);
  box-shadow: none;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.metric-card {
  display: flex;
  gap: 12px;
  padding: 16px;
}

.metric-icon {
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border: 1px solid currentColor;
  border-radius: 8px;
  background: rgba(59, 130, 246, 0.08);
}

.metric-card span,
.metric-card small,
.panel-head span,
.side-head span,
.namespace-row span,
.shape-grid span,
.name-cell small,
.queue-item small {
  color: var(--ds-text-muted, #8b949e);
}

.metric-card strong {
  display: block;
  margin: 4px 0;
  color: var(--ds-text-primary, #f8fafc);
  font-size: 24px;
  line-height: 1;
}

.metric-card.is-success,
.status-tag.success {
  color: var(--ds-color-success, #22c55e);
}

.metric-card.is-warning,
.status-tag.warning {
  color: var(--ds-color-warning, #f59e0b);
}

.metric-card.is-info,
.status-tag.info {
  color: var(--ds-color-info, #3b82f6);
}

.config-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 8px;
}

.config-panel,
.side-card {
  min-width: 0;
}

.panel-head,
.side-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px;
  border-bottom: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
}

.panel-head h2,
.side-head h3 {
  font-size: 14px;
}

.search-input {
  width: 260px;
}

.table-shell {
  padding: 0 8px 8px;
}

.name-cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
}

.name-cell span {
  overflow: hidden;
  color: var(--ds-text-primary, #f8fafc);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  border-color: currentColor;
  background: transparent;
}

.density-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.density-track {
  position: relative;
  width: 110px;
  height: 6px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.14);
}

.density-track span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: var(--ds-color-info, #3b82f6);
}

.density-cell small {
  color: var(--ds-text-muted, #8b949e);
}

.action-group {
  display: flex;
  gap: 4px;
  white-space: nowrap;
}

.side-panel {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 16px;
}

.queue-list,
.namespace-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
}

.queue-item,
.namespace-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  border-radius: 8px;
  background: rgba(13, 17, 23, 0.42);
}

.queue-item {
  padding: 12px;
}

.queue-item div {
  min-width: 0;
}

.queue-item strong,
.namespace-row strong,
.shape-grid strong {
  display: block;
  color: var(--ds-text-primary, #f8fafc);
}

.queue-item small {
  display: block;
  overflow: hidden;
  margin-top: 4px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.namespace-row {
  padding: 10px 12px;
}

.shape-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  padding: 12px;
}

.shape-grid div,
.data-item {
  border: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  border-radius: 8px;
  background: rgba(13, 17, 23, 0.42);
}

.shape-grid div {
  padding: 12px;
}

.shape-grid strong {
  margin-top: 4px;
  font-size: 20px;
}

.data-editor {
  max-height: 400px;
  overflow-y: auto;
}

.data-item {
  margin-bottom: 10px;
  padding: 10px;
}

.detail-content {
  color: var(--ds-text-primary, #f8fafc);
}

.compact-table {
  margin-top: 10px;
}

:deep(.el-dialog),
:deep(.el-tabs--border-card) {
  border: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  background: var(--ds-bg-card, #161b22);
  box-shadow: none;
}

:deep(.el-dialog__title),
:deep(.el-form-item__label),
:deep(.el-tabs__item),
:deep(.el-descriptions__label),
:deep(.el-descriptions__content),
:deep(.el-divider__text) {
  color: var(--ds-text-primary, #f8fafc);
}

:deep(.el-dialog__header),
:deep(.el-dialog__footer),
:deep(.el-tabs--border-card > .el-tabs__header) {
  border-color: var(--ds-border-default, rgba(148, 163, 184, 0.16));
  background: var(--ds-bg-card, #161b22);
}

:deep(.el-tabs--border-card > .el-tabs__content),
:deep(.el-tabs--border-card > .el-tabs__header .el-tabs__item.is-active),
:deep(.el-divider__text) {
  background: var(--ds-bg-card, #161b22);
}

:deep(.el-input__wrapper),
:deep(.el-select__wrapper),
:deep(.el-textarea__inner) {
  border: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  background: var(--ds-bg-page, #0d1117);
  box-shadow: none;
}

:deep(.el-input__inner),
:deep(.el-select__placeholder),
:deep(.el-textarea__inner) {
  color: var(--ds-text-primary, #f8fafc);
}

:deep(.el-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: transparent;
  --el-table-border-color: var(--ds-border-default, rgba(148, 163, 184, 0.16));
  --el-table-text-color: var(--ds-text-secondary, #c9d1d9);
  --el-table-header-text-color: var(--ds-text-muted, #8b949e);
  background: transparent;
  color: var(--ds-text-secondary, #c9d1d9);
}

:deep(.el-table__inner-wrapper),
:deep(.el-table__header-wrapper),
:deep(.el-table__body-wrapper),
:deep(.el-table tr),
:deep(.el-table th.el-table__cell),
:deep(.el-table td.el-table__cell) {
  background: transparent;
}

:deep(.el-table th.el-table__cell) {
  border-bottom: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  color: var(--ds-text-muted, #8b949e);
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

:deep(.el-table td.el-table__cell) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

:deep(.el-table__body tr:hover > td.el-table__cell) {
  background: rgba(59, 130, 246, 0.08);
}

:deep(.el-table__empty-block),
:deep(.el-empty__description p) {
  color: var(--ds-text-muted, #8b949e);
}

:deep(.el-button) {
  box-shadow: none;
}

:deep(.el-button.is-link) {
  padding: 0;
}

:deep(.el-descriptions__body),
:deep(.el-descriptions__table),
:deep(.el-descriptions__cell) {
  border-color: var(--ds-border-default, rgba(148, 163, 184, 0.16)) !important;
  background: transparent !important;
}

@media (max-width: 1280px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .config-layout {
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
