<template>
  <div class="ingress-management">
    <section class="ingress-header">
      <div>
        <div class="eyebrow">Kubernetes 网络</div>
        <h1>路由管理</h1>
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
        <el-button type="primary" class="primary-action" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:ingress:showcreatedialogtrue')">
          <el-icon><Plus /></el-icon>
          创建 Ingress
        </el-button>
      </div>
    </section>

    <section class="ingress-layout">
      <div class="ingress-panel">
        <div class="panel-head">
          <div>
            <h2>路由清单</h2>
            <span>{{ filteredIngressList.length }} 个路由对象</span>
          </div>
          <el-input v-model="searchKeyword" clearable placeholder="搜索主机、地址或类别" class="search-input" />
        </div>

        <div class="table-shell">
          <el-table
            :data="filteredIngressList"
            class="ops-table"
            v-loading="loading"
            element-loading-background="rgba(13, 17, 23, 0.72)"
            height="calc(100vh - 392px)"
            :empty-text="loading ? '加载中...' : '暂无数据'"
          >
            <el-table-column prop="name" label="名称" min-width="170" show-overflow-tooltip>
              <template #default="scope">
                <div class="name-cell">
                  <span>{{ scope.row.name }}</span>
                  <small>{{ scope.row.namespace }}</small>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="className" label="类别" width="130" show-overflow-tooltip>
              <template #default="scope">
                <el-tag class="status-tag" :class="ingressClassTone(scope.row.className)" effect="plain">
                  {{ scope.row.className || 'default' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="hosts" label="主机" min-width="220" show-overflow-tooltip>
              <template #default="scope">
                <div class="tag-row" v-if="scope.row.hosts?.length">
                  <el-tag v-for="host in scope.row.hosts" :key="host" class="host-tag" effect="plain">
                    {{ host }}
                  </el-tag>
                </div>
                <span v-else class="muted">通配符</span>
              </template>
            </el-table-column>
            <el-table-column prop="address" label="地址" min-width="160" show-overflow-tooltip>
              <template #default="scope">
                <span>{{ scope.row.address || '-' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="TLS" width="80" align="center">
              <template #default="scope">
                <el-tag class="status-tag" :class="scope.row.tls?.length ? 'success' : 'neutral'" effect="plain">
                  {{ scope.row.tls?.length ? '开启' : '关闭' }}
                </el-tag>
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
                  <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:ingress:handleviewdetail')">详情</el-button>
                  <el-button link type="info" size="small" @click="handleEdit(scope.row)" v-show="permStore.hasPerm('k8s:ingress:handleedit')">编辑</el-button>
                  <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:ingress:handledelete')">删除</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </section>

    <el-dialog v-model="showCreateDialog" :title="isEditMode ? '编辑 Ingress' : '创建 Ingress'" width="1100px" :close-on-click-modal="false" class="ops-dialog">
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
            <el-form-item label="路由类别" prop="className">
              <el-input v-model="createForm.className" placeholder="nginx、traefik 等" />
            </el-form-item>

            <el-divider content-position="left">路由规则</el-divider>
            <div v-for="(rule, ruleIndex) in createForm.rules" :key="ruleIndex" class="rule-container">
              <div class="rule-header">
                <span class="rule-title">规则 {{ ruleIndex + 1 }}</span>
                <el-button link type="danger" size="small" @click="removeRule(ruleIndex)">删除规则</el-button>
              </div>
              <el-form-item label="主机" label-width="80px">
                <el-input v-model="rule.host" placeholder="example.com" />
              </el-form-item>

              <div v-for="(path, pathIndex) in rule.paths" :key="pathIndex" class="path-item">
                <div class="path-header">
                  <span class="path-label">路径 {{ pathIndex + 1 }}</span>
                  <el-button link type="danger" size="small" @click="removePath(ruleIndex, pathIndex)">删除路径</el-button>
                </div>
                <el-row :gutter="8">
                  <el-col :span="5">
                    <el-form-item label="路径">
                      <el-input v-model="path.path" placeholder="/" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="5">
                    <el-form-item label="路径类型">
                      <el-select v-model="path.pathType" style="width: 100%;">
                        <el-option label="前缀匹配" value="Prefix" />
                        <el-option label="精确匹配" value="Exact" />
                        <el-option label="实现特定" value="ImplementationSpecific" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="服务名称">
                      <el-input v-model="path.serviceName" placeholder="my-service" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="6">
                    <el-form-item label="端口">
                      <el-input-number v-model="path.servicePort" :min="1" :max="65535" style="width: 100%;" controls-position="right" />
                    </el-form-item>
                  </el-col>
                </el-row>
              </div>
              <el-button size="small" type="primary" plain @click="addPath(ruleIndex)">+ 添加 Path</el-button>
            </div>
            <el-button type="primary" plain @click="addRule" style="margin-top: 10px;">+ 添加 Rule</el-button>

            <el-divider content-position="left">TLS</el-divider>
            <div v-for="(tls, tlsIndex) in createForm.tls" :key="tlsIndex" class="tls-item">
              <el-form-item label="主机列表">
                <el-select v-model="tls.hosts" multiple placeholder="选择或输入 hosts" allow-create filterable style="width: calc(100% - 50px);" />
                <el-button link type="danger" size="small" @click="removeTLS(tlsIndex)" style="margin-left: 10px;">删除</el-button>
              </el-form-item>
              <el-form-item label="密钥名称">
                <el-input v-model="tls.secretName" placeholder="tls-secret" />
              </el-form-item>
            </div>
            <el-button @click="addTLS">+ 添加 TLS</el-button>
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

    <el-dialog v-model="showDetailDialog" title="Ingress 详情" width="900px" class="ops-dialog">
      <div v-if="currentIngress" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ currentIngress.metadata.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ currentIngress.metadata.namespace }}</el-descriptions-item>
          <el-descriptions-item label="路由类别">{{ currentIngress.spec.ingressClassName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTimestamp(currentIngress.metadata.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">路由规则</el-divider>
        <div v-for="(rule, index) in currentIngress.spec.rules" :key="index" class="detail-rule">
          <strong>主机: {{ rule.host || '*' }}</strong>
          <el-table :data="rule.http?.paths || []" class="ops-table compact-table">
            <el-table-column prop="path" label="路径" width="150" />
            <el-table-column prop="pathType" label="路径类型" width="150" />
            <el-table-column label="后端服务">
              <template #default="scope">
                {{ scope.row.backend.service?.name }}:{{ scope.row.backend.service?.port.number }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <el-divider content-position="left" v-if="currentIngress.spec.tls && currentIngress.spec.tls.length > 0">TLS</el-divider>
        <el-table v-if="currentIngress.spec.tls && currentIngress.spec.tls.length > 0" :data="currentIngress.spec.tls" class="ops-table compact-table">
          <el-table-column label="主机列表">
            <template #default="scope">
              {{ scope.row.hosts.join(', ') }}
            </template>
          </el-table-column>
          <el-table-column prop="secretName" label="密钥名称" />
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
import { Aim, CircleCheck, Connection, Lock, Plus } from '@element-plus/icons-vue'
import { getIngressList, getIngressDetail, createIngress, updateIngress, deleteIngress } from '@/api/k8s/network'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const isEditMode = ref(false)
const editingIngress = ref(null)
const ingressList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
const currentIngress = ref(null)
const activeTab = ref('form')
const searchKeyword = ref('')

const createForm = ref({
  name: '',
  namespace: 'default',
  className: '',
  rules: [
    {
      host: '',
      paths: [
        {
          path: '/',
          pathType: 'Prefix',
          serviceName: '',
          servicePort: 80
        }
      ]
    }
  ],
  tls: [],
  yaml: ''
})

const filteredIngressList = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return ingressList.value
  return ingressList.value.filter((item) => {
    return [item.name, item.namespace, item.className, item.address, ...(item.hosts || [])]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(keyword))
  })
})

const hostCount = computed(() => ingressList.value.reduce((sum, item) => sum + (item.hosts?.length || 0), 0))
const tlsIngressCount = computed(() => ingressList.value.filter((item) => item.tls?.length).length)
const addressedIngressCount = computed(() => ingressList.value.filter((item) => item.address).length)
const ingressClassCount = computed(() => new Set(ingressList.value.map((item) => item.className || 'default')).size)
const exposureQueue = computed(() => filteredIngressList.value.filter((item) => item.address || (item.hosts?.length || 0) > 0).slice(0, 8))

const namespaceBreakdown = computed(() => {
  const namespaceMap = new Map()
  ingressList.value.forEach((item) => {
    namespaceMap.set(item.namespace, (namespaceMap.get(item.namespace) || 0) + 1)
  })
  return Array.from(namespaceMap.entries())
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 8)
})

const ingressMetrics = computed(() => [
  { key: 'total', label: 'Ingresses', value: ingressList.value.length, meta: `${namespaceList.value.length || 1} namespaces`, tone: 'info', icon: Connection },
  { key: 'hosts', label: 'Hosts', value: hostCount.value, meta: 'public routes', tone: hostCount.value ? 'warning' : 'neutral', icon: Aim },
  { key: 'tls', label: 'TLS Enabled', value: tlsIngressCount.value, meta: 'secured routes', tone: tlsIngressCount.value ? 'success' : 'warning', icon: Lock },
  { key: 'address', label: 'Addressed', value: addressedIngressCount.value, meta: 'lb assigned', tone: addressedIngressCount.value === ingressList.value.length ? 'success' : 'info', icon: CircleCheck }
])

const ingressClassTone = (className) => {
  if (!className) return 'neutral'
  if (className.toLowerCase().includes('nginx')) return 'success'
  return 'info'
}

const formatTimestamp = (timestamp) => {
  if (typeof timestamp === 'number') {
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
  }
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
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
    const res = await getIngressList(selectedNamespace.value, instanceId)
    if (res.status === 200) {
      ingressList.value = res.data.ingressList
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const addRule = () => {
  createForm.value.rules.push({
    host: '',
    paths: [
      {
        path: '/',
        pathType: 'Prefix',
        serviceName: '',
        servicePort: 80
      }
    ]
  })
}

const removeRule = (index) => {
  createForm.value.rules.splice(index, 1)
}

const addPath = (ruleIndex) => {
  createForm.value.rules[ruleIndex].paths.push({
    path: '/',
    pathType: 'Prefix',
    serviceName: '',
    servicePort: 80
  })
}

const removePath = (ruleIndex, pathIndex) => {
  createForm.value.rules[ruleIndex].paths.splice(pathIndex, 1)
}

const addTLS = () => {
  createForm.value.tls.push({
    hosts: [],
    secretName: ''
  })
}

const removeTLS = (index) => {
  createForm.value.tls.splice(index, 1)
}

const resetForm = () => {
  createForm.value = {
    name: '',
    namespace: 'default',
    className: '',
    rules: [
      {
        host: '',
        paths: [
          {
            path: '/',
            pathType: 'Prefix',
            serviceName: '',
            servicePort: 80
          }
        ]
      }
    ],
    tls: [],
    yaml: ''
  }
  isEditMode.value = false
  editingIngress.value = null
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
    const payload = {
      ...createForm.value,
      className: createForm.value.className?.trim() || undefined
    }
    await createIngress(payload, instanceId)
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
    const payload = {
      ...createForm.value,
      className: createForm.value.className?.trim() || undefined
    }
    await updateIngress(
      editingIngress.value.namespace,
      editingIngress.value.name,
      payload,
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
  const res = await getIngressDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    const ingress = res.data.ingressDetail
    editingIngress.value = row
    isEditMode.value = true

    createForm.value = {
      name: ingress.metadata.name,
      namespace: ingress.metadata.namespace,
      className: ingress.spec.ingressClassName || '',
      rules: ingress.spec.rules?.map(rule => ({
        host: rule.host || '',
        paths: rule.http?.paths?.map(p => ({
          path: p.path,
          pathType: p.pathType,
          serviceName: p.backend.service?.name || '',
          servicePort: p.backend.service?.port?.number || 80
        })) || []
      })) || [],
      tls: ingress.spec.tls?.map(t => ({
        hosts: t.hosts || [],
        secretName: t.secretName || ''
      })) || [],
      annotations: ingress.metadata.annotations || {},
      yaml: ''
    }

    showCreateDialog.value = true
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 Ingress ${row.name} 吗?`, '提示', {
    type: 'warning'
  }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteIngress(row.namespace, row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getIngressDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    currentIngress.value = res.data.ingressDetail
    showDetailDialog.value = true
  }
}

onMounted(() => {
  fetchNamespaces()
  fetchData()
})
</script>

<style scoped>
.ingress-management {
  min-height: 100vh;
  padding: 10px;
  color: var(--ds-text-primary, #f8fafc);
  background: var(--ds-bg-page, #0d1117);
}

.ingress-header,
.metric-card,
.ingress-panel,
.side-card {
  border: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  border-radius: 8px;
  background: var(--ds-bg-card, #161b22);
}

.ingress-header {
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

.ingress-header h1,
.panel-head h2,
.side-head h3 {
  margin: 0;
  color: var(--ds-text-primary, #f8fafc);
}

.ingress-header h1 {
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
.queue-item small,
.muted {
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

.metric-card.is-neutral,
.status-tag.neutral {
  color: var(--ds-text-muted, #8b949e);
}

.ingress-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 8px;
}

.ingress-panel,
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

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.host-tag,
.status-tag {
  border-color: currentColor;
  background: transparent;
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

.shape-grid div {
  border: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  border-radius: 8px;
  background: rgba(13, 17, 23, 0.42);
  padding: 12px;
}

.shape-grid strong {
  margin-top: 4px;
  font-size: 20px;
}

.rule-container,
.path-item,
.tls-item,
.detail-rule {
  margin-bottom: 16px;
  border: 1px solid var(--ds-border-default, rgba(148, 163, 184, 0.16));
  border-radius: 8px;
  background: rgba(13, 17, 23, 0.52);
}

.rule-container {
  padding: 18px;
}

.path-item,
.tls-item,
.detail-rule {
  padding: 14px;
}

.rule-header,
.path-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.rule-title,
.path-label,
.detail-rule strong {
  color: var(--ds-text-primary, #f8fafc);
  font-size: 13px;
  font-weight: 600;
}

.detail-content {
  color: var(--ds-text-primary, #f8fafc);
}

.compact-table {
  margin-top: 10px;
}

:deep(.el-card),
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
:deep(.el-textarea__inner),
:deep(.el-input-number .el-input__wrapper) {
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

  .ingress-layout {
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
