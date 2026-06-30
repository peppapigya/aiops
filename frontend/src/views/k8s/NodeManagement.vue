<template>
  <div class="node-management">
    <header class="nodes-header">
      <div>
        <div class="eyebrow">Kubernetes 集群容量</div>
        <h1>节点管理</h1>
      </div>
      <div class="header-actions">
        <el-button @click="refreshData" :loading="refreshing" :icon="Refresh">刷新</el-button>
        <el-button @click="toggleAutoRefresh" :type="autoRefreshEnabled ? 'success' : 'default'" :icon="Timer">
          {{ autoRefreshEnabled ? '自动刷新中' : '手动刷新' }}
        </el-button>
        <el-button type="primary" @click="showAddDialog = true" v-show="permStore.hasPerm('k8s:node:showadddialogtrue')" :icon="Plus">添加</el-button>
      </div>
    </header>

    <section class="nodes-layout">
      <main class="nodes-panel">
        <div class="panel-toolbar">
          <div>
            <span class="panel-title">节点清单</span>
            <span class="panel-subtitle">{{ filteredNodeList.length }} 可见 · {{ nodeList.length }} 总计</span>
          </div>
          <div class="toolbar-controls">
            <el-input v-model="searchText" placeholder="搜索节点" prefix-icon="Search" clearable class="filter-input" @input="handleSearch" />
            <el-select v-model="statusFilter" placeholder="状态" clearable class="filter-select" @change="handleFilter">
              <el-option label="全部" value="" />
              <el-option label="就绪" value="Ready" />
              <el-option label="未就绪" value="NotReady" />
              <el-option label="未知" value="Unknown" />
            </el-select>
          </div>
        </div>

        <div class="autoops-table-wrapper">
          <el-table
            :data="filteredNodeList"
            class="autoops-table"
            v-loading="loading"
            empty-text="暂无节点"
            row-key="name"
            @row-click="handleRowClick"
            height="600"
          >
            <el-table-column type="expand">
              <template #default="props">
                <div class="node-detail-expand">
                  <div class="detail-grid">
                    <div class="detail-section">
                      <h4>系统信息</h4>
                      <el-descriptions :column="1" size="small">
                        <el-descriptions-item label="操作系统">{{ props.row.osImage || '未知' }}</el-descriptions-item>
                        <el-descriptions-item label="内核">{{ props.row.kernelVersion || '未知' }}</el-descriptions-item>
                        <el-descriptions-item label="运行时">{{ props.row.containerRuntime || '未知' }}</el-descriptions-item>
                        <el-descriptions-item label="Kubelet">{{ props.row.kubeletVersion || '未知' }}</el-descriptions-item>
                      </el-descriptions>
                    </div>
                    <div class="detail-section">
                      <h4>可分配资源</h4>
                      <el-descriptions :column="1" size="small">
                        <el-descriptions-item label="CPU">{{ props.row.cpuAllocatable || '未知' }}</el-descriptions-item>
                        <el-descriptions-item label="内存">{{ props.row.memoryAllocatable || '未知' }}</el-descriptions-item>
                        <el-descriptions-item label="Pod数">{{ props.row.podCapacity || '未知' }}</el-descriptions-item>
                        <el-descriptions-item label="存储">{{ props.row.storageCapacity || '未知' }}</el-descriptions-item>
                      </el-descriptions>
                    </div>
                  </div>
                  <div class="detail-section labels-block">
                    <h4>标签</h4>
                    <div class="labels-container">
                      <el-tag v-for="(value, key) in props.row.labels" :key="key" size="small" class="label-tag">{{ key }}: {{ value }}</el-tag>
                      <span v-if="!props.row.labels || Object.keys(props.row.labels).length === 0" class="no-labels">无标签</span>
                    </div>
                  </div>
                </div>
              </template>
            </el-table-column>

            <el-table-column prop="name" label="节点" min-width="220">
              <template #default="scope">
                <div class="node-name-cell">
                  <el-icon class="node-icon" :class="scope.row.isMaster ? 'is-master' : 'is-worker'"><Monitor /></el-icon>
                  <div class="node-title-cell">
                    <span>{{ scope.row.name }}</span>
                    <small>{{ scope.row.internalIP || '-' }}</small>
                  </div>
                  <span v-if="scope.row.isMaster" class="role-pill is-master">主节点</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="120">
              <template #default="scope">
                <span class="status-pill" :class="`is-${nodeStatusTone(scope.row.status)}`"><span class="status-dot" />{{ scope.row.status }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="role" label="角色" width="110">
              <template #default="scope"><span class="role-pill" :class="scope.row.isMaster ? 'is-master' : 'is-worker'">{{ scope.row.isMaster ? '主节点' : '工作节点' }}</span></template>
            </el-table-column>
            <el-table-column prop="externalIP" label="外部IP" width="140"><template #default="scope">{{ scope.row.externalIP || '-' }}</template></el-table-column>
            <el-table-column label="CPU" width="140">
              <template #default="scope"><div class="resource-usage"><el-progress :percentage="scope.row.cpuUsage" :color="getProgressColor(scope.row.cpuUsage)" :show-text="false" :stroke-width="6" /><span class="usage-text">{{ scope.row.cpuUsage }}%</span></div></template>
            </el-table-column>
            <el-table-column label="Memory" width="140">
              <template #default="scope"><div class="resource-usage"><el-progress :percentage="scope.row.memoryUsage" :color="getProgressColor(scope.row.memoryUsage)" :show-text="false" :stroke-width="6" /><span class="usage-text">{{ scope.row.memoryUsage }}%</span></div></template>
            </el-table-column>
            <el-table-column prop="podCount" label="Pods" width="110"><template #default="scope"><span class="pod-count">{{ scope.row.podCount }} / {{ scope.row.podCapacity }}</span></template></el-table-column>
            <el-table-column prop="k8sVersion" label="Kubelet版本" width="140"><template #default="scope">{{ scope.row.kubeletVersion || scope.row.kubeProxyVersion || '-' }}</template></el-table-column>
            <el-table-column prop="createTime" label="创建时间" width="170"><template #default="scope">{{ formatTimestamp(scope.row.createTime) }}</template></el-table-column>
            <el-table-column label="操作" width="220">
              <template #default="scope">
                <div class="action-buttons-inline">
                  <el-button type="primary" link size="small" @click.stop="handleViewMonitor(scope.row)">监控</el-button>
                  <el-button type="primary" link size="small" @click.stop="handleViewDetail(scope.row)">详情</el-button>
                  <el-button type="warning" link size="small" @click.stop="handleCordon(scope.row)" :disabled="scope.row.isMaster">{{ scope.row.cordoned ? '解除隔离' : '隔离' }}</el-button>
                  <el-button type="danger" link size="small" @click.stop="handleDrain(scope.row)" :disabled="scope.row.isMaster">驱逐</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </main>
    </section>

    <!-- 节点详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="节点详情" width="900px" top="5vh">
      <div v-if="currentNode" class="node-detail">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-card class="detail-card">
              <template #header>
                <span>基本信息</span>
              </template>
              <el-descriptions :column="1" border>
                <el-descriptions-item label="节点名称">{{ currentNode.name || '-' }}</el-descriptions-item>
                <el-descriptions-item label="状态">
                  <el-tag :type="getStatusType(currentNode.status)">
                    {{ currentNode.status || '未知' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="角色">
                  <el-tag :type="currentNode.isMaster ? 'danger' : 'primary'">
                    {{ currentNode.isMaster ? '主节点' : '工作节点' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="内部IP">{{ currentNode.internalIP || '-' }}</el-descriptions-item>
                <el-descriptions-item label="外部IP">{{ currentNode.externalIP || '-' }}</el-descriptions-item>
                <el-descriptions-item label="K8s版本">{{ currentNode.kubeletVersion || currentNode.k8sVersion || '-' }}</el-descriptions-item>
                <el-descriptions-item label="创建时间">{{ formatTimestamp(currentNode.createTime) }}</el-descriptions-item>
                <el-descriptions-item label="节点隔离状态">
                  <el-tag :type="currentNode.cordoned ? 'warning' : 'success'">
                    {{ currentNode.cordoned ? '已隔离' : '未隔离' }}
                  </el-tag>
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card class="detail-card">
              <template #header>
                <span>系统信息</span>
              </template>
              <el-descriptions :column="1" border>
                <el-descriptions-item label="操作系统">{{ currentNode.osImage || '未知' }}</el-descriptions-item>
                <el-descriptions-item label="内核版本">{{ currentNode.kernelVersion || '未知' }}</el-descriptions-item>
                <el-descriptions-item label="容器运行时">{{ currentNode.containerRuntime || '未知' }}</el-descriptions-item>
                <el-descriptions-item label="Kubelet版本">{{ currentNode.kubeletVersion || '未知' }}</el-descriptions-item>
                <el-descriptions-item label="kube-proxy版本">{{ currentNode.kubeProxyVersion || '未知' }}</el-descriptions-item>
                <el-descriptions-item label="系统UUID">{{ currentNode.systemUUID || '未知' }}</el-descriptions-item>
                <el-descriptions-item label="操作系统镜像">{{ currentNode.osImage || '未知' }}</el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-col>
        </el-row>
        
        <el-row :gutter="20" style="margin-top: 20px;">
          <el-col :span="24">
            <el-card class="detail-card">
              <template #header>
                <span>资源信息</span>
              </template>
              <el-row :gutter="20">
                <el-col :span="12">
                  <h4>CPU 资源</h4>
                  <el-descriptions :column="1" size="small">
                    <el-descriptions-item label="容量">{{ currentNode.cpuCapacity || '未知' }}</el-descriptions-item>
                    <el-descriptions-item label="可分配">{{ currentNode.cpuAllocatable || '未知' }}</el-descriptions-item>
                    <el-descriptions-item label="使用率">{{ currentNode.cpuUsage || 0 }}%</el-descriptions-item>
                  </el-descriptions>
                </el-col>
                <el-col :span="12">
                  <h4>内存资源</h4>
                  <el-descriptions :column="1" size="small">
                    <el-descriptions-item label="容量">{{ currentNode.memoryCapacity || '未知' }}</el-descriptions-item>
                    <el-descriptions-item label="可分配">{{ currentNode.memoryAllocatable || '未知' }}</el-descriptions-item>
                    <el-descriptions-item label="使用率">{{ currentNode.memoryUsage || 0 }}%</el-descriptions-item>
                  </el-descriptions>
                </el-col>
              </el-row>
              <el-row :gutter="20" style="margin-top: 16px;">
                <el-col :span="8">
                  <h4>Pod 容量</h4>
                  <el-descriptions :column="1" size="small">
                    <el-descriptions-item label="总容量">{{ currentNode.podCapacity || '未知' }}</el-descriptions-item>
                    <el-descriptions-item label="当前数量">{{ currentNode.podCount || 0 }}</el-descriptions-item>
                  </el-descriptions>
                </el-col>
                <el-col :span="8">
                  <h4>存储资源</h4>
                  <el-descriptions :column="1" size="small">
                    <el-descriptions-item label="容量">{{ formatStorage(currentNode.storageCapacity) || '未知' }}</el-descriptions-item>
                    <el-descriptions-item label="可分配">{{ formatStorage(currentNode.storageAllocatable) || '未知' }}</el-descriptions-item>
                    <el-descriptions-item label="Ephemeral存储">{{ formatStorage(currentNode.ephemeralStorageCapacity) || '未知' }}</el-descriptions-item>
                  </el-descriptions>
                </el-col>
                <el-col :span="8">
                  <h4>网络资源</h4>
                  <el-descriptions :column="1" size="small">
                    <el-descriptions-item label="最大Pod数">{{ currentNode.podCapacity || '未知' }}</el-descriptions-item>
                    <el-descriptions-item label="网络策略">{{ currentNode.networkPolicyAvailable ? '已配置' : '未配置' }}</el-descriptions-item>
                    <el-descriptions-item label="PodCIDR">{{ getPodCIDR(currentNode) || '未知' }}</el-descriptions-item>
                  </el-descriptions>
                </el-col>
              </el-row>
            </el-card>
          </el-col>
        </el-row>
        
        <el-row :gutter="20" style="margin-top: 20px;">
          <el-col :span="24">
            <el-card class="detail-card">
              <template #header>
                <span>节点标签和注解</span>
              </template>
              <el-tabs v-model="activeTab">
                <el-tab-pane label="标签" name="labels">
                  <div class="labels-container">
                    <el-tag 
                      v-for="(value, key) in currentNode.labels" 
                      :key="key" 
                      class="label-tag"
                      closable
                      @close="handleRemoveLabel(key)"
                    >
                      {{ key }}: {{ value }}
                    </el-tag>
                    <el-button size="small" @click="showAddLabelDialog = true">
                      <el-icon><Plus /></el-icon>
                      添加标签
                    </el-button>
                  </div>
                </el-tab-pane>
                <el-tab-pane label="注解" name="annotations">
                  <div class="annotations-container">
                    <div v-for="(value, key) in currentNode.annotations" :key="key" class="annotation-item">
                      <span class="annotation-key">{{ key }}:</span>
                      <span class="annotation-value">{{ value }}</span>
                    </div>
                  </div>
                </el-tab-pane>
              </el-tabs>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </el-dialog>

    <!-- 节点监控抽屉 -->
    <el-drawer v-model="showMonitorDrawer" title="节点实时监控" size="70%" destroy-on-close>
      <div v-if="currentNode" class="monitor-container" style="padding: 0 20px;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
          <div>
            <span style="font-weight: bold; margin-right: 15px; font-size: 16px;">
              <el-icon style="vertical-align: middle; margin-right: 5px;"><Monitor /></el-icon>
              {{ currentNode.name }}
            </span>
            <el-select v-model="selectedPrometheusId" placeholder="选择 Prometheus 实例" size="small" style="width: 200px;">
              <el-option v-for="inst in prometheusInstances" :key="inst.id" :label="inst.name" :value="inst.id" />
            </el-select>
          </div>
          <el-button type="primary" size="small" @click="handleOpenCustomMonitorDialog(null)" v-show="permStore.hasPerm('k8s:node:handleopencustommonitordialog')" >
            <el-icon><Plus /></el-icon> 新增自定义监控图表
          </el-button>
        </div>
        
        <div v-if="selectedPrometheusId">
          <!-- 自定义图表区 (Node级别) -->
          <el-row :gutter="20">
            <el-col :span="12" v-for="monitor in customMonitors" :key="monitor.id" style="margin-bottom: 20px;">
              <DynamicPromQLChart 
                :monitor-id="monitor.id"
                :node-name="currentNode.name"
                :prometheus-instance-id="parseInt(selectedPrometheusId)" 
                :title="monitor.title" 
                :chart-type="monitor.chart_type"
                :unit="monitor.unit_suffix"
                :color="monitor.color_theme"
                :promql-template="monitor.promql_template"
                :hours="1" 
                @edit="handleOpenCustomMonitorDialog"
                @delete="handleDeleteCustomMonitor"
              />
            </el-col>
          </el-row>
          <el-empty v-if="customMonitors.length === 0" description="暂无图表，请点击右上角新增" />

        </div>
        <el-empty v-else description="请先选择用于监控的 Prometheus 实例数据源" />
      </div>
    </el-drawer>
    <el-dialog v-model="showMonitorDialog" :title="monitorForm.id ? '编辑自定义图表' : '新增自定义图表'" width="600px" append-to-body destroy-on-close>
      <el-form :model="monitorForm" :rules="monitorRules" ref="monitorFormRef" label-width="120px">
        <el-form-item label="图表标题" prop="title">
          <el-input v-model="monitorForm.title" placeholder="如：系统负载均值" />
        </el-form-item>
        <el-form-item label="PromQL 模板" prop="promql_template">
          <el-input v-model="monitorForm.promql_template" type="textarea" :rows="3" placeholder="例如: node_load1{instance='{{nodeName}}'}" />
          <div style="font-size: 12px; color: #999; margin-top: 5px;">
            可用插值变量: <code>{{nodeName}}</code>
          </div>
        </el-form-item>
        <el-form-item label="图表类型" prop="chart_type">
          <el-radio-group v-model="monitorForm.chart_type">
            <el-radio label="line">折线图</el-radio>
            <el-radio label="bar">柱状图</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Y轴单位">
          <el-input v-model="monitorForm.unit_suffix" placeholder="如: %, MB, 负载 (可选)" />
        </el-form-item>
        <el-form-item label="主题颜色">
          <el-color-picker v-model="monitorForm.color_theme" show-alpha />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showMonitorDialog = false">取消</el-button>
          <el-button type="primary" @click="handleSaveCustomMonitor" :loading="monitorSubmitting">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 添加标签对话框 -->
    <el-dialog v-model="showAddLabelDialog" title="添加标签" width="400px">
      <el-form :model="labelForm" :rules="labelRules" ref="labelFormRef" label-width="80px">
        <el-form-item label="键" prop="key">
          <el-input v-model="labelForm.key" placeholder="请输入标签键" />
        </el-form-item>
        <el-form-item label="值" prop="value">
          <el-input v-model="labelForm.value" placeholder="请输入标签值" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddLabelDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAddLabel">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Monitor, Refresh, Plus, CircleCheck, Warning, Star, Search, View, 
  Lock, Remove, Timer
} from '@element-plus/icons-vue'
import { getNodeList, getNodeDetail, cordonNode, uncordonNode, drainNode, addNodeLabel, removeNodeLabel } from '@/api/k8s/node'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import { getInstanceList } from '@/api/instance'
import { getCustomMonitors, createCustomMonitor, updateCustomMonitor, deleteCustomMonitor } from '@/api/monitor'
import NodeMonitorChart from './components/NodeMonitorChart.vue'
import DynamicPromQLChart from './components/DynamicPromQLChart.vue'

const loading = ref(false)
const refreshing = ref(false)
const showDetailDialog = ref(false)
const showMonitorDrawer = ref(false)
const showAddLabelDialog = ref(false)
const searchText = ref('')
const statusFilter = ref('')
const currentNode = ref(null)
const labelFormRef = ref(null)
const autoRefreshTimer = ref(null)
const autoRefreshEnabled = ref(false)
const activeTab = ref('labels')
const prometheusInstances = ref([])
const selectedPrometheusId = ref('')

// 自定义监控相关状态
const customMonitors = ref([])
const showMonitorDialog = ref(false)
const monitorSubmitting = ref(false)
const monitorFormRef = ref(null)
const monitorForm = ref({
  id: null,
  title: '',
  promql_template: '',
  chart_type: 'line',
  unit_suffix: '',
  color_theme: '#409EFF'
})
const monitorRules = {
  title: [{ required: true, message: '请输入图表标题', trigger: 'blur' }],
  promql_template: [{ required: true, message: '请输入 PromQL 查询模板', trigger: 'blur' }],
  chart_type: [{ required: true, message: '请选择图表类型', trigger: 'change' }]
}

// 节点统计数据
const nodeStats = ref({
  total: 0,
  ready: 0,
  notReady: 0,
  masters: 0
})

// 节点列表
const nodeList = ref([])

// 标签表单
const labelForm = ref({
  key: '',
  value: ''
})

const labelRules = {
  key: [{ required: true, message: '请输入标签键', trigger: 'blur' }],
  value: [{ required: true, message: '请输入标签值', trigger: 'blur' }]
}

// 过滤后的节点列表
const filteredNodeList = computed(() => {
  let filtered = nodeList.value
  
  // 按状态筛选
  if (statusFilter.value) {
    filtered = filtered.filter(node => node.status === statusFilter.value)
  }
  
  // 按名称搜索
  if (searchText.value) {
    const searchLower = searchText.value.toLowerCase()
    filtered = filtered.filter(node => 
      node.name.toLowerCase().includes(searchLower)
    )
  }
  
  return filtered
})

const nodeMetrics = computed(() => [
  { key: 'total', label: '节点', value: nodeStats.value.total, meta: '集群容量', tone: 'info', icon: Monitor },
  { key: 'ready', label: '就绪', value: nodeStats.value.ready, meta: `${readyRate.value}% 就绪`, tone: 'success', icon: CircleCheck },
  { key: 'notReady', label: '未就绪', value: nodeStats.value.notReady, meta: '需处理', tone: 'error', icon: Warning },
  { key: 'masters', label: '主节点', value: nodeStats.value.masters, meta: `${workerCount.value} 工作节点`, tone: 'warning', icon: Star }
])

const readyRate = computed(() => {
  if (!nodeStats.value.total) return 0
  return Math.round((nodeStats.value.ready / nodeStats.value.total) * 100)
})

const workerCount = computed(() => Math.max(nodeStats.value.total - nodeStats.value.masters, 0))

const pressureNodes = computed(() => filteredNodeList.value.filter((node) => {
  return node.status !== 'Ready' || Number(node.cpuUsage) >= 80 || Number(node.memoryUsage) >= 80 || node.cordoned
}).slice(0, 8))

const roleBreakdown = computed(() => [
  { key: 'master', label: '主节点', value: nodeStats.value.masters },
  { key: 'worker', label: '工作节点', value: workerCount.value },
  { key: 'cordoned', label: '已隔离', value: nodeList.value.filter((node) => node.cordoned).length }
])

const nodeStatusTone = (status) => ({ Ready: 'success', NotReady: 'error', Unknown: 'warning' }[status] || 'neutral')

const formatTimestamp = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN')
}

const formatStorage = (storage) => {
  if (!storage) return '未知'
  
  // 如果已经是格式化的值（如 49353520Ki），直接返回
  if (typeof storage === 'string' && /[KMG]i?$/.test(storage)) {
    return storage
  }
  
  // 如果是数字，转换为可读格式
  if (typeof storage === 'number') {
    const units = ['Ki', 'Mi', 'Gi', 'Ti']
    let value = storage
    let unitIndex = 0
    
    while (value >= 1024 && unitIndex < units.length - 1) {
      value /= 1024
      unitIndex++
    }
    
    return `${Math.round(value)}${units[unitIndex]}`
  }
  
  return storage
}

const getPodCIDR = (node) => {
  // 尝试从标签或注解中获取PodCIDR信息
  if (node.podCIDR) return node.podCIDR
  
  // 从标签中查找
  if (node.labels && node.labels['kubernetes.io/pod-cidr']) {
    return node.labels['kubernetes.io/pod-cidr']
  }
  
  // 从注解中查找
  if (node.annotations && node.annotations['kubernetes.io/pod-cidr']) {
    return node.annotations['kubernetes.io/pod-cidr']
  }
  
  return '未知'
}

const checkNetworkPolicy = (nodeData) => {
  // 检查是否配置了网络策略
  if (!nodeData || !nodeData.labels) return false
  
  // 检查常见的网络策略相关标签
  const networkPolicyLabels = [
    'networking.k8s.io/policy-name',
    'policy.beta.kubernetes.io/ingress',
    'policy.beta.kubernetes.io/egress'
  ]
  
  for (const label of networkPolicyLabels) {
    if (nodeData.labels[label]) {
      return true
    }
  }
  
  // 检查注解中是否有网络策略相关信息
  if (nodeData.annotations) {
    const networkPolicyAnnotations = [
      'kubernetes.io/ingress-bandwidth',
      'kubernetes.io/egress-bandwidth'
    ]
    
    for (const annotation of networkPolicyAnnotations) {
      if (nodeData.annotations[annotation]) {
        return true
      }
    }
  }
  
  return false
}

const getNodeRoleColor = (role) => {
  return role === 'master' ? '#f56c6c' : '#409eff'
}

const getStatusType = (status) => {
  switch (status) {
    case 'Ready': return 'success'
    case 'NotReady': return 'danger'
    case 'Unknown': return 'warning'
    default: return 'info'
  }
}

const getProgressColor = (percentage) => {
  if (percentage < 60) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

const refreshData = async () => {
  refreshing.value = true
  try {
    await fetchNodeList()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败: ' + (error.message || error.response?.data?.message || '未知错误'))
  } finally {
    refreshing.value = false
  }
}

const handleSearch = () => {
  // 搜索逻辑已在 computed 中处理
}

const handleFilter = () => {
  // 筛选逻辑已在 computed 中处理
}

const handleRowClick = (row) => {
  handleViewDetail(row)
}

const handleViewDetail = async (node) => {
  try {
    // 显示加载状态
    const loading = ElMessage({
      message: '正在加载节点详情...',
      type: 'info',
      duration: 0,
      showClose: false
    })
    
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getNodeDetail(node.name, instanceId)
    
    // 合并基础信息和详细信息
    currentNode.value = {
      ...node,
      ...response.data,
      labels: response.data?.labels || node.labels || {},
      annotations: response.data?.annotations || node.annotations || {},
      // 添加网络策略和PodCIDR信息
      networkPolicyAvailable: checkNetworkPolicy(response.data),
      podCIDR: getPodCIDR(response.data)
    }

    // 重置标签页到默认选中状态
    activeTab.value = 'labels'
    
    loading.close()
    showDetailDialog.value = true
  } catch (error) {
    ElMessage.error(`获取节点详情失败: ${error.message || error.response?.data?.message || '未知错误'}`)
  }
}

const handleViewMonitor = async (node) => {
  try {
    const loading = ElMessage({
      message: '正在获取监控配置...',
      type: 'info',
      duration: 0,
      showClose: false
    })
    
    currentNode.value = node

    // 自动加载 Prometheus 实例提供下拉选择
    if (prometheusInstances.value.length === 0) {
      const pRes = await getInstanceList({ page: 1, page_size: 100, type_name: 'prometheus' })
      if (pRes.data?.list?.data) {
        prometheusInstances.value = pRes.data.list.data
        if (prometheusInstances.value.length > 0) {
          selectedPrometheusId.value = prometheusInstances.value[0].id
        }
      }
    }
    
    loading.close()
    // 获取当前用户的自定义监控配置 (Node 维度)
    await fetchCustomMonitors()

    showMonitorDrawer.value = true
  } catch (error) {
    ElMessage.error(`获取组件状态失败: ${error.message || error.response?.data?.message || '未知错误'}`)
  }
}

const handleCordon = async (node) => {
  try {
    const action = node.cordoned ? '取消隔离' : '隔离'
    await ElMessageBox.confirm(
      `确定要${action}节点 "${node.name}" 吗？`,
      `确认${action}`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const instanceId = getSelectedInstanceId() || '1'
    
    // 显示加载状态
    const loading = ElMessage({
      message: `正在${action}节点...`,
      type: 'info',
      duration: 0,
      showClose: false
    })
    
    try {
      if (node.cordoned) {
        await uncordonNode(node.name, instanceId)
      } else {
        await cordonNode(node.name, instanceId)
      }
      node.cordoned = !node.cordoned
      loading.close()
      ElMessage.success(`节点${action}成功`)
    } catch (apiError) {
      loading.close()
      throw apiError
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}节点失败: ${error.message || error.response?.data?.message || '未知错误'}`)
    }
  }
}

const handleDrain = async (node) => {
  try {
    await ElMessageBox.confirm(
      `确定要排空节点 "${node.name}" 吗？此操作将驱逐该节点上的所有Pod。`,
      '确认排空',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const instanceId = getSelectedInstanceId() || '1'
    
    // 显示加载状态
    const loading = ElMessage({
      message: '正在排空节点...',
      type: 'info',
      duration: 0,
      showClose: false
    })
    
    try {
      await drainNode(node.name, instanceId)
      loading.close()
      ElMessage.success('节点排空操作已启动，请稍后刷新查看结果')
      // 自动刷新数据
      setTimeout(() => {
        fetchNodeList()
      }, 3000)
    } catch (apiError) {
      loading.close()
      throw apiError
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`排空节点失败: ${error.message || error.response?.data?.message || '未知错误'}`)
    }
  }
}

const handleDeleteLabel = async (key) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除标签 "${key}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const instanceId = getSelectedInstanceId() || '1'
    await removeNodeLabel(currentNode.value.name, key, instanceId)
    delete currentNode.value.labels[key]
    ElMessage.success('标签删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除标签失败: ' + (error.message || '未知错误'))
    }
  }
}

const handleAddLabel = async () => {
  if (!labelFormRef.value) return
  
  await labelFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    try {
      const instanceId = getSelectedInstanceId() || '1'
      await addNodeLabel(currentNode.value.name, labelForm.value.key, labelForm.value.value, instanceId)
      
      if (!currentNode.value.labels) {
        currentNode.value.labels = {}
      }
      
      currentNode.value.labels[labelForm.value.key] = labelForm.value.value
      showAddLabelDialog.value = false
      labelForm.value = { key: '', value: '' }
      ElMessage.success('标签添加成功')
    } catch (error) {
      ElMessage.error('添加标签失败: ' + (error.message || '未知错误'))
    }
  })
}

const fetchNodeList = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getNodeList(instanceId)
    const nodes = response.data?.nodeList || []
    
    // 处理节点数据，计算实际使用率
    nodeList.value = nodes.map(node => ({
      ...node,
      cpuUsage: calculateCpuUsage(node),
      memoryUsage: calculateMemoryUsage(node),
      podCount: parseInt(node.podCount || 0),
      podCapacity: parseInt(node.podCapacity || 110),
      cordoned: node.cordoned || false,
      createTime: node.createTime || node.creationTimestamp || Date.now(),
      role: detectNodeRole(node),
      isMaster: isMasterNode(node)
    }))
    
    // 更新统计数据
    const processedNodes = nodes.map(node => ({
      ...node,
      isMaster: isMasterNode(node)
    }))
    
    nodeStats.value = {
      total: processedNodes.length,
      ready: processedNodes.filter(n => n.status === 'Ready').length,
      notReady: processedNodes.filter(n => n.status === 'NotReady').length,
      masters: processedNodes.filter(n => n.isMaster).length
    }
    
  } catch (error) {
    ElMessage.error('获取节点列表失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// 计算CPU使用率
const calculateCpuUsage = (node) => {
  if (!node.cpuCapacity || !node.cpuAllocatable) return 0
  
  // 将CPU字符串转换为数值（如 "2" -> 2000m）
  const parseCpu = (cpuStr) => {
    if (!cpuStr) return 0
    if (cpuStr.endsWith('m')) {
      return parseInt(cpuStr.slice(0, -1))
    }
    return parseInt(parseFloat(cpuStr) * 1000)
  }
  
  const capacity = parseCpu(node.cpuCapacity)
  const allocatable = parseCpu(node.cpuAllocatable)
  
  if (capacity === 0) return 0
  
  // 计算已分配的CPU（容量 - 可分配）
  const used = capacity - allocatable
  return Math.round((used / capacity) * 100)
}

// 计算内存使用率
const calculateMemoryUsage = (node) => {
  if (!node.memoryCapacity || !node.memoryAllocatable) return 0
  
  // 将内存字符串转换为KB
  const parseMemory = (memStr) => {
    if (!memStr) return 0
    const units = { 'Ki': 1, 'Mi': 1024, 'Gi': 1024 * 1024, 'K': 1, 'M': 1024, 'G': 1024 * 1024 }
    const match = memStr.match(/^(\d+)([KMG]i?)?$/)
    if (!match) return 0
    const value = parseInt(match[1])
    const unit = match[2] || ''
    return value * (units[unit] || 1)
  }
  
  const capacity = parseMemory(node.memoryCapacity)
  const allocatable = parseMemory(node.memoryAllocatable)
  
  if (capacity === 0) return 0
  
  // 计算已分配的内存（容量 - 可分配）
  const used = capacity - allocatable
  return Math.round((used / capacity) * 100)
}

// 检测节点角色
const detectNodeRole = (node) => {
  const labels = node.labels || {}
  
  // 检查各种master/control-plane标签
  if (labels['node-role.kubernetes.io/control-plane'] || 
      labels['node-role.kubernetes.io/master'] ||
      node.name?.toLowerCase().includes('master') ||
      node.name?.toLowerCase().includes('control-plane')) {
    return 'master'
  }
  
  return 'worker'
}

// 判断是否为master节点
const isMasterNode = (node) => {
  return detectNodeRole(node) === 'master'
}

// 切换自动刷新
const toggleAutoRefresh = () => {
  autoRefreshEnabled.value = !autoRefreshEnabled.value
  
  if (autoRefreshEnabled.value) {
    autoRefreshTimer.value = setInterval(() => {
      fetchNodeList()
    }, 30000) // 每30秒刷新一次
    ElMessage.success('已开启自动刷新（每30秒）')
  } else {
    if (autoRefreshTimer.value) {
      clearInterval(autoRefreshTimer.value)
      autoRefreshTimer.value = null
    }
    ElMessage.info('已停止自动刷新')
  }
}

// ============== 自定义监控模块 ==============
const fetchCustomMonitors = async () => {
  try {
    const res = await getCustomMonitors('node')
    if (res && res.data && res.data.list) {
      customMonitors.value = res.data.list
    } else {
      customMonitors.value = []
    }
  } catch (error) {
    console.error('Failed to fetch node custom monitors:', error)
  }
}

const handleOpenCustomMonitorDialog = (monitorId = null) => {
  if (monitorFormRef.value) {
    monitorFormRef.value.resetFields()
  }
  if (monitorId) {
    const monitor = customMonitors.value.find(m => m.id === monitorId)
    if (monitor) {
      monitorForm.value = { ...monitor }
    }
  } else {
    monitorForm.value = {
      id: null,
      title: '',
      promql_template: '',
      chart_type: 'line',
      unit_suffix: '',
      color_theme: '#409EFF'
    }
  }
  showMonitorDialog.value = true
}

const handleSaveCustomMonitor = async () => {
  if (!monitorFormRef.value) return
  const valid = await monitorFormRef.value.validate().catch(() => false)
  if (!valid) return

  monitorSubmitting.value = true
  try {
    const payload = {
      target_type: 'node',
      title: monitorForm.value.title,
      promql_template: monitorForm.value.promql_template,
      chart_type: monitorForm.value.chart_type,
      unit_suffix: monitorForm.value.unit_suffix,
      color_theme: monitorForm.value.color_theme
    }

    if (monitorForm.value.id) {
      await updateCustomMonitor(monitorForm.value.id, payload)
      ElMessage.success('更新自定义图表成功')
    } else {
      await createCustomMonitor(payload)
      ElMessage.success('创建自定义图表成功')
    }
    showMonitorDialog.value = false
    await fetchCustomMonitors()
  } catch (error) {
    ElMessage.error(monitorForm.value.id ? '更新图表失败' : '创建图表失败')
  } finally {
    monitorSubmitting.value = false
  }
}

const handleDeleteCustomMonitor = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个自定义监控图表吗？', '提示', { type: 'warning' })
    await deleteCustomMonitor(id)
    ElMessage.success('图表已删除')
    await fetchCustomMonitors()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除图表失败')
    }
  }
}

onMounted(() => {
  fetchNodeList()
})

// 组件卸载时清理定时器
onUnmounted(() => {
  if (autoRefreshTimer.value) {
    clearInterval(autoRefreshTimer.value)
  }
})
</script>

<style scoped>
.node-management {
  min-height: 100%;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: var(--ds-bg-app);
  color: var(--ds-text-primary);
}

.nodes-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.eyebrow {
  font-size: var(--ds-font-size-11);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--ds-text-muted);
}

.nodes-header h1 {
  margin: 2px 0 0;
  font-size: 18px;
  line-height: 1.1;
}

.header-actions,
.toolbar-controls,
.action-buttons-inline,
.resource-usage,
.node-name-cell {
  display: flex;
  align-items: center;
  gap: var(--ds-space-8);
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--ds-space-12);
}

.metric-card,
.nodes-panel,
.side-panel,
.panel-block {
  background: var(--ds-bg-surface);
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-8);
}

.metric-card {
  padding: var(--ds-space-14);
}

.metric-meta,
.metric-foot,
.panel-subtitle,
.queue-meta,
.no-labels {
  color: var(--ds-text-tertiary);
}

.metric-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--ds-font-size-11);
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.metric-value {
  margin-top: var(--ds-space-10);
  font-size: var(--ds-font-size-26);
  font-weight: 750;
  letter-spacing: -0.03em;
}

.metric-foot {
  margin-top: var(--ds-space-4);
  font-size: var(--ds-font-size-12);
}

.metric-card.is-success .metric-meta,
.metric-card.is-success .metric-value { color: var(--ds-success); }
.metric-card.is-error .metric-meta,
.metric-card.is-error .metric-value { color: var(--ds-error); }
.metric-card.is-warning .metric-meta,
.metric-card.is-warning .metric-value { color: var(--ds-warning); }
.metric-card.is-info .metric-meta,
.metric-card.is-info .metric-value { color: var(--ds-info); }

.nodes-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 8px;
  min-height: 0;
  flex: 1;
}

.nodes-panel {
  min-width: 0;
  padding: 8px;
}

.panel-toolbar,
.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ds-space-12);
  padding-bottom: var(--ds-space-12);
  border-bottom: 1px solid var(--ds-border-subtle);
}

.panel-head.compact {
  padding-bottom: var(--ds-space-10);
}

.panel-title {
  display: block;
  font-size: var(--ds-font-size-13);
  font-weight: 700;
  color: var(--ds-text-primary);
}

.panel-subtitle {
  display: block;
  margin-top: var(--ds-space-2);
  font-size: var(--ds-font-size-12);
}

.filter-input { width: 220px; }
.filter-select { width: 132px; }
.autoops-table-wrapper { margin-top: var(--ds-space-12); }

.node-title-cell {
  display: flex;
  flex-direction: column;
  gap: var(--ds-space-2);
  min-width: 0;
}

.node-title-cell span {
  color: var(--ds-text-primary);
  font-weight: 650;
}

.node-title-cell small,
.usage-text,
.pod-count {
  color: var(--ds-text-tertiary);
  font-size: var(--ds-font-size-12);
}

.node-icon {
  font-size: 18px;
}
.node-icon.is-master { color: var(--ds-error); }
.node-icon.is-worker { color: var(--ds-accent); }

.status-pill,
.role-pill,
.queue-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--ds-space-6);
  min-height: 22px;
  padding: 0 var(--ds-space-8);
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-6);
  font-size: var(--ds-font-size-12);
  font-weight: 650;
  white-space: nowrap;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: currentColor;
}

.status-pill.is-success,
.queue-badge.is-success { color: var(--ds-success); background: var(--ds-bg-success-subtle); }
.status-pill.is-error,
.queue-badge.is-error { color: var(--ds-error); background: var(--ds-bg-danger-subtle); }
.status-pill.is-warning,
.queue-badge.is-warning { color: var(--ds-warning); background: var(--ds-bg-warning-subtle); }
.status-pill.is-neutral,
.queue-badge.is-neutral { color: var(--ds-text-tertiary); background: var(--ds-bg-surface-2); }
.role-pill { color: var(--ds-text-secondary); background: var(--ds-bg-surface-2); }
.role-pill.is-master { color: var(--ds-error); background: var(--ds-bg-danger-subtle); }
.role-pill.is-worker { color: var(--ds-accent); background: var(--ds-bg-info-subtle); }

.node-detail-expand {
  padding: var(--ds-space-16);
  background: var(--ds-bg-surface-2);
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-8);
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--ds-space-12);
}

.detail-section h4,
.detail-card h4 {
  margin: 0 0 var(--ds-space-10) 0;
  font-size: var(--ds-font-size-13);
  font-weight: 700;
  color: var(--ds-text-primary);
}

.labels-block {
  margin-top: var(--ds-space-14);
}

.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ds-space-8);
}

.label-tag { margin: 0; }

.side-panel {
  display: flex;
  flex-direction: column;
  gap: var(--ds-space-14);
  background: transparent;
  border: 0;
}

.panel-block {
  padding: var(--ds-space-14);
}

.queue-list {
  display: flex;
  flex-direction: column;
  gap: var(--ds-space-8);
}

.queue-item {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--ds-space-5);
  padding: var(--ds-space-10);
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-6);
  background: var(--ds-bg-surface-2);
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.queue-item:hover {
  border-color: var(--ds-border-strong);
  background: var(--ds-bg-hover);
}

.queue-main {
  font-weight: 700;
  color: var(--ds-text-primary);
}

.empty-state {
  padding: var(--ds-space-24) var(--ds-space-12);
  border: 1px dashed var(--ds-border-default);
  border-radius: var(--ds-radius-8);
  color: var(--ds-text-muted);
  text-align: center;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  padding: var(--ds-space-8) 0;
  color: var(--ds-text-secondary);
  border-bottom: 1px solid var(--ds-border-subtle);
}

.summary-row:last-child { border-bottom: 0; }

.node-detail { max-height: 70vh; overflow-y: auto; }
.detail-card { margin-bottom: 0; }
.annotations-container { max-height: 200px; overflow-y: auto; }
.annotation-item { padding: var(--ds-space-8) 0; border-bottom: 1px solid var(--ds-border-subtle); }
.annotation-item:last-child { border-bottom: 0; }
.annotation-key { font-weight: 600; color: var(--ds-text-primary); margin-right: var(--ds-space-8); }
.annotation-value { color: var(--ds-text-secondary); }
.monitor-container { min-height: 400px; padding: var(--ds-space-10) 0; }

:deep(.el-card),
:deep(.el-card__body),
:deep(.el-card__header),
:deep(.el-descriptions__body),
:deep(.el-descriptions__cell) {
  background: var(--ds-bg-surface) !important;
  border-color: var(--ds-border-subtle) !important;
  box-shadow: none !important;
  color: var(--ds-text-secondary) !important;
}

@media (max-width: 1280px) {
  .nodes-layout { grid-template-columns: 1fr; }
  .side-panel { display: none; }
}

@media (max-width: 960px) {
  .nodes-header,
  .panel-toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  .metric-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .detail-grid { grid-template-columns: 1fr; }
  .toolbar-controls { flex-direction: column; align-items: stretch; }
  .filter-input,
  .filter-select { width: 100%; }
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
