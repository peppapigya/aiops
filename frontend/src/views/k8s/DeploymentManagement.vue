<template>
  <div class="deployment-management">
    <header class="workload-header">
      <div>
        <div class="eyebrow">Kubernetes 工作负载</div>
        <h1>部署管理</h1>
      </div>
      <div class="header-actions">
        <el-button @click="fetchData" :loading="loading" :icon="RefreshRight">刷新</el-button>
        <el-button type="success" :icon="Monitor">仪表盘</el-button>
        <el-button type="primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:deployment:showcreatedialogtrue')" :icon="Plus">创建</el-button>
      </div>
    </header>

    <section class="workload-layout">
      <main class="workload-panel">
        <div class="panel-toolbar">
          <div>
            <span class="panel-title">部署清单</span>
            <span class="panel-subtitle">{{ filteredDeploymentList.length }} 可见 · {{ deploymentList.length }} 总计</span>
          </div>
          <div class="toolbar-controls">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索工作负载"
              class="filter-input"
              :prefix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            />
            <el-select v-model="selectedNamespace" placeholder="全部命名空间" filterable class="filter-select" @change="fetchData">
              <el-option label="default" value="default" />
              <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
            </el-select>
            <el-button type="primary" @click="handleSearch">搜索</el-button>
            <el-button @click="handleReset">重置</el-button>
          </div>
        </div>

        <el-table
          :data="filteredDeploymentList"
          class="workload-table"
          style="width: 100%"
          v-loading="loading"
          @row-click="handleViewDetail"
        >
          <el-table-column label="工作负载" min-width="200">
            <template #default="{ row }">
              <div class="workload-name-cell">
                <div class="workload-avatar">D</div>
                <div class="workload-identity">
                  <button class="name-link" @click.stop="handleViewDetail(row)">{{ row.name }}</button>
                  <span>{{ row.namespace || selectedNamespace }} · Deployment</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="发布状态" min-width="140">
            <template #default="{ row }">
              <div class="rollout-cell">
                <div class="rollout-top">
                  <span :class="['status-pill', `is-${deploymentStatusTone(row)}`]">{{ deploymentStatusText(row) }}</span>
                  <strong>{{ row.ready || 0 }}/{{ row.replicas || 0 }}</strong>
                </div>
                <div class="progress-track">
                  <span :style="{ width: `${rolloutProgress(row)}%` }" />
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="资源" min-width="140">
            <template #default="{ row }">
              <div class="resource-cell">
                <div><span>CPU</span><strong>{{ formatResourceValue(row.resources?.cpuRequest) }}</strong><em>/ {{ formatResourceValue(row.resources?.cpuLimit) }}</em></div>
                <div><span>MEM</span><strong>{{ formatResourceValue(row.resources?.memoryRequest) }}</strong><em>/ {{ formatResourceValue(row.resources?.memoryLimit) }}</em></div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="镜像" min-width="200">
            <template #default="{ row }">
              <div class="image-cell">
                <el-icon><Box /></el-icon>
                <span :title="primaryImage(row)">{{ primaryImage(row) }}</span>
                <b v-if="row.containers && row.containers.length > 1">+{{ row.containers.length - 1 }}</b>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="标签" min-width="90" align="center">
            <template #default="{ row }">
              <el-popover placement="top" width="auto" trigger="hover">
                <template #reference>
                  <span class="label-badge"><el-icon><PriceTag /></el-icon>{{ Object.keys(row.labels || {}).length }}</span>
                </template>
                <div class="tags-popover">
                  <span v-for="(val, key) in row.labels" :key="key" class="tag-item">{{ key }}: {{ val }}</span>
                  <span v-if="!Object.keys(row.labels || {}).length" class="tag-empty">无标签</span>
                </div>
              </el-popover>
            </template>
          </el-table-column>

          <el-table-column label="创建时间" width="140">
            <template #default="{ row }">
              <span class="time-text">{{ formatDate(row.created) }}</span>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <div class="workload-actions" @click.stop>
                <el-button link type="primary" size="small" @click="handleScale(row)">扩缩</el-button>
                <el-button link type="primary" size="small" @click="handleViewDetail(row)" v-show="permStore.hasPerm('k8s:deployment:handleviewdetail')">详情</el-button>
                <el-button link type="primary" size="small" @click="handleUpdate(row)">更新镜像</el-button>
                <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, row)">
                  <button class="more-action">更多<el-icon><ArrowDown /></el-icon></button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="shell">终端</el-dropdown-item>
                      <el-dropdown-item command="logs">日志</el-dropdown-item>
                      <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
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

    <el-dialog v-model="showCreateDialog" title="创建 Deployment" width="900px" :close-on-click-modal="false" append-to-body>
      <div class="dialog-section">
        <el-tabs v-model="activeTab" type="border-card">
          <el-tab-pane label="表单创建" name="form">
            <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="120px">
              <el-form-item label="名称" prop="name"><el-input v-model="createForm.name" /></el-form-item>
              <el-form-item label="命名空间" prop="namespace">
                <el-select v-model="createForm.namespace" placeholder="请选择命名空间" filterable style="width: 100%">
                  <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
                </el-select>
              </el-form-item>
              <el-form-item label="镜像" prop="image"><el-input v-model="createForm.image" /></el-form-item>
              <el-form-item label="副本数" prop="replicas"><el-input-number v-model="createForm.replicas" /></el-form-item>
            </el-form>
          </el-tab-pane>
          <el-tab-pane label="YAML" name="yaml">
            <div class="mb-2">
              <el-button type="success" size="small" :icon="CircleCheck" @click="handleValidateYAML">YAML 语法校验</el-button>
            </div>
            <el-input type="textarea" v-model="yamlContent" :rows="15" placeholder="在此粘贴 YAML 配置..." class="code-input" />
          </el-tab-pane>
        </el-tabs>
      </div>
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

    <el-dialog v-model="showDetailDialog" :title="detailTitle" width="800px" append-to-body>
      <el-input v-model="detailContent" type="textarea" :rows="20" readonly class="code-input" />
    </el-dialog>

    <el-dialog v-model="showUpdateDialog" title="更新镜像" width="500px" append-to-body>
      <el-form :model="updateForm" label-width="100px">
        <el-form-item label="当前镜像">
          <el-input v-model="updateForm.currentImage" disabled />
        </el-form-item>
        <el-form-item label="新镜像">
          <el-input v-model="updateForm.newImage" placeholder="例如: nginx:1.19.0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUpdateDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmUpdate">更新</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showLogsDialog" title="日志查看" width="80%" top="5vh" append-to-body destroy-on-close>
      <div class="log-toolbar">
        <el-select v-model="selectedLogPod" placeholder="选择 Pod" @change="fetchLogs" style="width: 300px">
          <el-option v-for="pod in logPodList" :key="pod.name" :label="pod.name" :value="pod.name" />
        </el-select>
        <el-button @click="fetchLogs" :loading="logLoading" :icon="RefreshRight" circle />
      </div>
      <div class="log-container">
        <pre>{{ logContent }}</pre>
      </div>
    </el-dialog>

    <el-dialog v-model="showContainerSelectDialog" title="选择容器连接 Shell" width="500px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="Pod">
          <el-select v-model="selectedShellPodName" placeholder="选择 Pod" style="width: 100%" @change="handleShellPodChange">
            <el-option v-for="pod in shellPodList" :key="pod.name" :label="pod.name" :value="pod.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="容器">
          <el-select v-model="selectedContainerForTerminal" placeholder="请选择容器" style="width: 100%;">
            <el-option v-for="c in shellContainerList" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showContainerSelectDialog = false">取消</el-button>
          <el-button type="primary" @click="confirmContainerSelect" :disabled="!selectedShellPodName || !selectedContainerForTerminal">连接</el-button>
        </div>
      </template>
    </el-dialog>

    <PodTerminal
      v-model="showTerminalDialog"
      :namespace="terminalPod.namespace"
      :pod-name="terminalPod.podName"
      :container-name="terminalPod.containerName"
      :instance-id="getSelectedInstanceId()"
      @close="handleTerminalClose"
    />
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search, RefreshRight, Monitor, Plus,
  PriceTag, Platform, Box, Cloudy,
  ArrowDown, CircleCheck
} from '@element-plus/icons-vue'
import { getDeploymentList, scaleDeployment, deleteDeployment, createDeployment, getDeploymentDetail, updateDeployment } from '@/api/k8s/deployment'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getPodList, getPodLogs } from '@/api/k8s/pod'
import dayjs from 'dayjs'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import yaml from 'js-yaml'
import PodTerminal from '@/components/PodTerminal.vue'

const loading = ref(false)
const searchKeyword = ref('')
const deploymentList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedNamespace = ref('default')
const namespaceList = ref([])

const showCreateDialog = ref(false)
const showScaleDialog = ref(false)
const activeTab = ref('form')
const createForm = reactive({ name: '', namespace: 'default', image: 'nginx:latest', replicas: 1 })
const yamlContent = ref('')
const scaleForm = reactive({ replicas: 1, row: null })

const showDetailDialog = ref(false)
const detailContent = ref('')
const detailTitle = ref('')

const showUpdateDialog = ref(false)
const updateForm = reactive({ currentImage: '', newImage: '', row: null, originalData: null })

const showLogsDialog = ref(false)
const logContent = ref('')
const logPodList = ref([])
const selectedLogPod = ref('')
const logLoading = ref(false)

const showTerminalDialog = ref(false)
const terminalPod = ref({ namespace: '', podName: '', containerName: '' })
const showContainerSelectDialog = ref(false)
const shellPodList = ref([])
const selectedShellPodName = ref('')
const shellContainerList = ref([])
const selectedContainerForTerminal = ref('')

const filteredDeploymentList = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return deploymentList.value
  return deploymentList.value.filter((item) => {
    return [item.name, item.namespace, item.image, primaryImage(item)].filter(Boolean).some((value) => String(value).toLowerCase().includes(keyword))
  })
})

const readyDeployments = computed(() => deploymentList.value.filter((item) => Number(item.ready) >= Number(item.replicas || 0) && Number(item.replicas || 0) > 0).length)
const totalReplicas = computed(() => deploymentList.value.reduce((sum, item) => sum + Number(item.replicas || 0), 0))
const readyReplicas = computed(() => deploymentList.value.reduce((sum, item) => sum + Number(item.ready || 0), 0))
const rolloutQueue = computed(() => filteredDeploymentList.value.filter((item) => deploymentStatusTone(item) !== 'success').slice(0, 8))
const imageCount = computed(() => new Set(deploymentList.value.map((item) => primaryImage(item)).filter(Boolean)).size)

const deploymentMetrics = computed(() => [
  { key: 'total', label: '部署', value: deploymentList.value.length, meta: `${namespaceList.value.length || 1} 命名空间`, tone: 'info', icon: Cloudy },
  { key: 'ready', label: '就绪', value: readyDeployments.value, meta: '发布健康', tone: 'success', icon: CircleCheck },
  { key: 'replicas', label: '副本', value: `${readyReplicas.value}/${totalReplicas.value}`, meta: '就绪 / 期望', tone: readyReplicas.value === totalReplicas.value ? 'success' : 'warning', icon: Platform },
  { key: 'images', label: '镜像', value: imageCount.value, meta: '独立镜像', tone: 'info', icon: Box }
])

const namespaceBreakdown = computed(() => {
  const counts = new Map()
  deploymentList.value.forEach((item) => {
    const namespace = item.namespace || selectedNamespace.value || 'default'
    counts.set(namespace, (counts.get(namespace) || 0) + 1)
  })
  return Array.from(counts.entries()).map(([name, count]) => ({ name, count })).sort((a, b) => b.count - a.count).slice(0, 8)
})

const primaryImage = (row) => row.image || row.containers?.[0]?.image || '-'
const rolloutProgress = (row) => {
  const replicas = Number(row.replicas || 0)
  if (!replicas) return 0
  return Math.min(100, Math.round((Number(row.ready || 0) / replicas) * 100))
}
const deploymentStatusTone = (row) => {
  const replicas = Number(row.replicas || 0)
  const ready = Number(row.ready || 0)
  if (!replicas) return 'neutral'
  if (ready >= replicas) return 'success'
  if (ready > 0) return 'warning'
  return 'error'
}
const deploymentStatusText = (row) => {
  const tone = deploymentStatusTone(row)
  if (tone === 'success') return '就绪'
  if (tone === 'warning') return '发布中'
  if (tone === 'error') return '不可用'
  return '已缩0'
}

const fetchNamespaces = async () => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getNamespaceList(instanceId)
    const list = res.data?.items || res.data?.namespaceList || res.data || []
    namespaceList.value = list.map(item => ({ name: typeof item === 'string' ? item : item.name }))
  } catch (e) {
    console.error('Fetch namespaces failed', e)
  }
}

const handleValidateYAML = () => {
  if (!yamlContent.value) {
    ElMessage.warning('请输入 YAML 内容')
    return
  }
  try {
    yaml.loadAll(yamlContent.value)
    ElMessage.success('YAML 语法校验通过 ✅')
  } catch (e) {
    console.error(e)
    ElMessageBox.alert(
      `<pre style="color: #EF4444; max-height: 300px; overflow: auto;">${e.message}</pre>`,
      '语法错误',
      { dangerouslyUseHTMLString: true, type: 'error' }
    )
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getDeploymentList(selectedNamespace.value, instanceId)
    deploymentList.value = res.data?.deploymentList || []
    total.value = res.data?.total || deploymentList.value.length
  } catch (e) {
    ElMessage.error('获取 Deployment 列表失败')
    deploymentList.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => fetchData()
const handleReset = () => { searchKeyword.value = ''; fetchData() }
const handleSizeChange = () => fetchData()
const selectNamespace = (namespace) => { selectedNamespace.value = namespace; fetchData() }
const formatDate = (ts) => {
  if (!ts) return '-'
  return dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const formatResourceValue = (val) => {
  if (!val || val === '0' || val === '0m' || val === '0Mi') return '-'
  return val
}

const handleViewDetail = async (row) => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getDeploymentDetail(row.namespace, row.name, instanceId)
    if (res.data) {
      const detail = res.data.deploymentDetail || res.data
      detailContent.value = yaml.dump(detail)
      detailTitle.value = `详情: ${row.name}`
      showDetailDialog.value = true
    }
  } catch(e) { ElMessage.error('获取详情失败') }
}

const handleScale = (row) => {
  scaleForm.row = row
  scaleForm.replicas = row.replicas
  showScaleDialog.value = true
}

const handleScaleConfirm = async () => {
  try {
    const instanceId = getSelectedInstanceId()
    await scaleDeployment(scaleForm.row.namespace, scaleForm.row.name, scaleForm.replicas, instanceId)
    ElMessage.success('扩缩容指令已发送')
    showScaleDialog.value = false
    fetchData()
    setTimeout(() => fetchData(), 2000)
  } catch(e) {
    console.error(e)
    ElMessage.error('操作失败: ' + (e.response?.data?.message || e.message || '未知错误'))
  }
}

const handleUpdate = async (row) => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getDeploymentDetail(row.namespace, row.name, instanceId)
    if (res.data) {
      const detail = res.data.deploymentDetail || res.data
      updateForm.originalData = detail
      updateForm.row = row
      const containers = detail.containers || []
      updateForm.currentImage = containers.length > 0 ? containers[0].image : ''
      updateForm.newImage = updateForm.currentImage
      showUpdateDialog.value = true
    }
  } catch(e) { ElMessage.error('获取详情失败') }
}

const confirmUpdate = async () => {
  try {
    const instanceId = getSelectedInstanceId()
    const data = JSON.parse(JSON.stringify(updateForm.originalData))
    if (data.spec?.template?.spec?.containers?.length > 0) {
      data.spec.template.spec.containers[0].image = updateForm.newImage
      await updateDeployment(updateForm.row.namespace, updateForm.row.name, data, instanceId)
      ElMessage.success('更新镜像指令已发送')
      showUpdateDialog.value = false
      fetchData()
    } else {
      ElMessage.error('无法解析容器信息，请检查后端返回数据')
    }
  } catch(e) { ElMessage.error('更新失败') }
}

const getPodsBySelector = async (row) => {
  const instanceId = getSelectedInstanceId()
  const detailRes = await getDeploymentDetail(row.namespace, row.name, instanceId)
  const detail = detailRes.data?.deploymentDetail || detailRes.data
  const selector = detail?.selector
  if (!selector) { throw new Error('无法获取 Pod 选择器') }

  const podsRes = await getPodList(row.namespace, instanceId)
  const allPods = podsRes.data?.items || podsRes.data?.podList || []

  return allPods.filter(pod => {
    const podLabels = pod.labels || {}
    return Object.entries(selector).every(([k, v]) => podLabels[k] === v)
  })
}

const handleShell = async (row) => {
  try {
    const pods = await getPodsBySelector(row)
    if (pods.length === 0) {
      ElMessageBox.alert('未找到属于该 Deployment 的 Pod。可能副本数为 0，或者 Pod 尚未创建。', '提示', {
        confirmButtonText: '确定',
        type: 'warning'
      })
      return
    }

    shellPodList.value = pods

    if (pods.length === 1 && pods[0].containers && pods[0].containers.length === 1) {
      terminalPod.value = {
        namespace: row.namespace,
        podName: pods[0].name,
        containerName: pods[0].containers[0].name
      }
      showTerminalDialog.value = true
      return
    }

    selectedShellPodName.value = pods[0].name
    handleShellPodChange(pods[0].name)
    showContainerSelectDialog.value = true
  } catch(e) {
    console.error(e)
    ElMessage.error('准备 Shell 环境失败: ' + e.message)
  }
}

const handleShellPodChange = (podName) => {
  const pod = shellPodList.value.find(p => p.name === podName)
  if (pod && pod.containers) {
    shellContainerList.value = pod.containers.map(c => c.name)
    selectedContainerForTerminal.value = shellContainerList.value[0] || ''
  } else {
    shellContainerList.value = []
    selectedContainerForTerminal.value = ''
  }
}

const confirmContainerSelect = () => {
  const pod = shellPodList.value.find(p => p.name === selectedShellPodName.value)
  terminalPod.value = {
    namespace: pod.namespace,
    podName: selectedShellPodName.value,
    containerName: selectedContainerForTerminal.value
  }
  showContainerSelectDialog.value = false
  showTerminalDialog.value = true
}

const handleTerminalClose = () => {
  showTerminalDialog.value = false
}

const handleCommand = (cmd, row) => {
  if (cmd === 'shell') handleShell(row)
  if (cmd === 'logs') handleLogs(row)
  if (cmd === 'delete') handleDelete(row)
}

const handleLogs = async (row) => {
  try {
    logLoading.value = true
    const pods = await getPodsBySelector(row)
    logPodList.value = pods

    if (logPodList.value.length === 0) {
      ElMessageBox.alert('未找到属于该 Deployment 的 Pod。可能副本数为 0，或者 Pod 尚未创建。', '提示', {
        confirmButtonText: '确定',
        type: 'warning'
      })
      return
    }

    selectedLogPod.value = logPodList.value[0].name
    showLogsDialog.value = true
    fetchLogs()
  } catch(e) {
    console.error(e)
    ElMessage.error('获取 Pod 日志失败')
  } finally {
    logLoading.value = false
  }
}

const fetchLogs = async () => {
  if (!selectedLogPod.value) return
  try {
    const instanceId = getSelectedInstanceId()
    const row = logPodList.value.find(p => p.name === selectedLogPod.value)
    const container = row?.containers ? row.containers[0].name : ''
    const res = await getPodLogs(row.namespace, selectedLogPod.value, container, instanceId, { tailLines: 500 })
    logContent.value = typeof res.data === 'string' ? res.data : JSON.stringify(res.data, null, 2)
  } catch(e) {
    logContent.value = '获取日志失败。'
  }
}
const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 ${row.name}?`, '警告', { type: 'warning' })
    .then(async () => {
      await deleteDeployment(row.namespace, row.name)
      ElMessage.success('删除成功')
      fetchData()
    })
}

const createFormRef = ref(null)
const createRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  namespace: [{ required: true, message: '请输入命名空间', trigger: 'blur' }],
  image: [{ required: true, message: '请输入镜像', trigger: 'blur' }],
}

const handleCreate = async () => {
  if (!createFormRef.value) return

  await createFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const instanceId = getSelectedInstanceId()
        const data = {
          name: createForm.name,
          namespace: createForm.namespace,
          image: createForm.image,
          replicas: createForm.replicas,
        }

        await createDeployment(createForm.namespace, data, instanceId)

        ElMessage.success('创建成功')
        showCreateDialog.value = false
        fetchData()
      } catch (e) {
        console.error(e)
        ElMessage.error('创建失败: ' + (e.response?.data?.message || e.message || '未知错误'))
      }
    }
  })
}

onMounted(() => {
  fetchData()
  fetchNamespaces()
})
</script>

<style scoped>
.deployment-management {
  min-height: 100%;
  padding: 12px;
  background: var(--ds-bg-app);
  color: var(--ds-text-primary);
}

.workload-header,
.panel-toolbar,
.panel-head,
.header-actions,
.toolbar-controls,
.workload-actions,
.log-toolbar {
  display: flex;
  align-items: center;
}

.workload-header {
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px;
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
  margin: 2px 0 0;
  color: var(--ds-text-primary);
  font-size: 20px;
  font-weight: 700;
}

.header-actions,
.toolbar-controls,
.workload-actions,
.log-toolbar {
  gap: 8px;
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
.workload-identity span,
.runtime-grid span {
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
  gap: 10px;
  padding: 10px 14px;
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

.filter-input {
  width: 220px;
}

.filter-select {
  width: 150px;
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
  font-size: 12px;
  font-weight: 800;
}

.workload-identity {
  min-width: 0;
}

.name-link,
.more-action,
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

.name-link:hover,
.more-action:hover {
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
  display: inline-flex;
  align-items: center;
  gap: 5px;
  border: 1px solid rgba(59, 130, 246, .28);
  border-radius: 999px;
  padding: 2px 8px;
  background: var(--ds-bg-info-subtle);
  font-weight: 700;
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

.more-action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 0;
  color: var(--ds-accent);
  font-size: 12px;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  padding: 8px 14px;
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

.queue-badge {
  margin-top: 2px;
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

.code-input :deep(.el-textarea__inner) {
  font-family: var(--ds-font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
}

.log-toolbar {
  margin-bottom: 12px;
}

.log-container {
  height: 500px;
  overflow-y: auto;
  border: 1px solid var(--ds-border-subtle);
  border-radius: 8px;
  background: #0b1020;
  color: var(--ds-text-secondary);
  padding: 12px;
  font-family: var(--ds-font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
  font-size: 12px;
  white-space: pre-wrap;
}

.deployment-management :deep(.el-table),
.deployment-management :deep(.el-table__inner-wrapper),
.deployment-management :deep(.el-table__header-wrapper),
.deployment-management :deep(.el-table__body-wrapper),
.deployment-management :deep(.el-table tr),
.deployment-management :deep(.el-table td.el-table__cell),
.deployment-management :deep(.el-table th.el-table__cell) {
  background: var(--ds-bg-surface) !important;
  color: var(--ds-text-secondary) !important;
  border-color: var(--ds-border-subtle) !important;
}

.deployment-management :deep(.el-table th.el-table__cell) {
  background: var(--ds-bg-surface-2) !important;
  color: var(--ds-text-tertiary) !important;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

.deployment-management :deep(.el-table__row:hover > td.el-table__cell) {
  background: var(--ds-bg-surface-2) !important;
}

.deployment-management :deep(.el-table td.el-table__cell) {
  padding: 8px 0 !important;
}

.deployment-management :deep(.el-table th.el-table__cell) {
  padding: 10px 0 !important;
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
  .toolbar-controls {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-input,
  .filter-select {
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
  gap: 4px;
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
  min-width: 28px;
  height: 28px;
  padding: 0 6px;
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
  gap: 10px !important;
  padding: 10px 14px !important;
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
  margin-top: 10px !important;
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
  padding: 10px 14px !important;
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
