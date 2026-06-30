<template>
  <div class="host-management">
    <header class="asset-header">
      <div>
        <div class="eyebrow">资产清单</div>
        <h1>主机管理</h1>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="fetchHosts">刷新</el-button>
        <el-button :icon="AlarmClock" @click="handleAlarm">告警</el-button>
        <el-button type="primary" :icon="Plus" @click="openHostDialog()">新增</el-button>
      </div>
    </header>

    <section class="metric-grid">
      <div v-for="metric in assetMetrics" :key="metric.key" class="metric-card" :class="`is-${metric.tone}`">
        <div class="metric-meta">
          <span>{{ metric.label }}</span>
          <el-icon><component :is="metric.icon" /></el-icon>
        </div>
        <div class="metric-value">{{ metric.value }}</div>
        <div class="metric-foot">{{ metric.meta }}</div>
      </div>
    </section>

    <section class="asset-layout">
      <aside class="group-panel">
        <div class="panel-head">
          <div>
            <span class="panel-title">分组</span>
            <span class="panel-subtitle">共 {{ totalGroupCount }} 个</span>
          </div>
          <div class="icon-actions">
            <el-button link :icon="Refresh" @click="fetchGroups" />
            <el-button link :icon="Plus" @click="openGroupDialog(null, 'createRoot')" />
          </div>
        </div>
        <el-input v-model="groupSearch" placeholder="筛选组" clearable size="small" class="group-search">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-scrollbar class="group-tree-scroll">
          <el-tree
            ref="groupTreeRef"
            :data="filteredGroups"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            :default-expanded-keys="expandedKeys"
            :highlight-current="true"
            @node-click="handleGroupClick"
            @node-contextmenu="handleContextMenu"
          >
            <template #default="{ data }">
              <div class="tree-node">
                <el-icon class="node-folder"><Folder /></el-icon>
                <span class="node-label">{{ data.name }}</span>
                <span class="node-count">{{ data.hostCount || 0 }}</span>
              </div>
            </template>
          </el-tree>
        </el-scrollbar>
      </aside>

      <main class="inventory-panel">
        <div class="panel-toolbar">
          <div>
            <span class="panel-title">主机列表</span>
            <span class="panel-subtitle">{{ selectedGroupName }}</span>
          </div>
          <div class="toolbar-controls">
            <el-input v-model="searchForm.name" placeholder="主机" clearable class="filter-input" />
            <el-input v-model="searchForm.ip" placeholder="IP" clearable class="filter-input" />
            <el-select v-model="searchForm.status" placeholder="状态" clearable class="filter-select">
              <el-option label="在线" value="online" />
              <el-option label="离线" value="offline" />
              <el-option label="未知" value="unknown" />
            </el-select>
            <el-button type="primary" :icon="Search" @click="fetchHosts">搜索</el-button>
            <el-button type="danger" :icon="Monitor" @click="handleTerminal">终端</el-button>
            <el-button :disabled="!selectedHosts.length" @click="handleBatchDelete">批量</el-button>
          </div>
        </div>

        <el-table
          :data="tableData"
          v-loading="loading"
          class="asset-table"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="42" />
          <el-table-column label="主机" min-width="180" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="host-cell">
                <span class="host-name">{{ row.name }}</span>
                <span class="host-ip">{{ row.ip }}:{{ row.port || 22 }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="110">
            <template #default="{ row }">
              <span class="status-pill" :class="`is-${statusTone(row.status)}`">
                <span class="status-dot" />{{ statusText(row.status) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="CPU" width="96">
            <template #default="{ row }">
              <span :class="getCpuClass(row.cpuUsage)">{{ row.cpuUsage != null ? row.cpuUsage + '%' : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="Memory" width="104">
            <template #default="{ row }">
              <span :class="getMemClass(row.memUsage)">{{ row.memUsage != null ? row.memUsage + '%' : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="Disk" width="96">
            <template #default="{ row }">
              <span :class="getDiskClass(row.diskUsage)">{{ row.diskUsage != null ? row.diskUsage + '%' : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="云平台" width="110">
            <template #default="{ row }">
              <span class="cloud-pill" :class="`is-${row.cloudProvider || 'self'}`">{{ cloudText(row.cloudProvider) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="信号" width="132">
            <template #default="{ row }">
              <div class="signal-row">
                <span>P {{ row.processCount ?? '-' }}</span>
                <span>Port {{ row.portCount ?? '-' }}</span>
                <span>T {{ row.tunnelCount ?? '-' }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="标签" min-width="140">
            <template #default="{ row }">
              <div class="tag-row">
                <span v-for="tag in (row.tags || [])" :key="tag" class="tag-chip">{{ tag }}</span>
                <span v-if="!(row.tags || []).length" class="muted">-</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <div class="row-actions">
                <el-button link type="primary" size="small" @click="openHostDialog(row)">编辑</el-button>
                <el-button link type="success" size="small" @click="handleTerminalRow(row)">SSH</el-button>
                <el-button link type="danger" size="small" @click="handleDeleteHost(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          class="pagination"
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchHosts"
          @current-change="fetchHosts"
        />
      </main>

      <aside class="side-panel">
        <div class="panel-head compact">
          <div>
            <span class="panel-title">风险</span>
            <span class="panel-subtitle">{{ riskHosts.length }} 台</span>
          </div>
        </div>
        <div class="risk-list">
          <button v-for="host in riskHosts" :key="host.id" class="risk-item" @click="openHostDialog(host)">
            <span class="risk-main">{{ host.name }}</span>
            <span class="risk-meta">{{ host.ip }}</span>
            <span class="risk-load">
              CPU {{ host.cpuUsage ?? '-' }} · MEM {{ host.memUsage ?? '-' }} · DISK {{ host.diskUsage ?? '-' }}
            </span>
          </button>
          <div v-if="!riskHosts.length" class="empty-state">无风险主机</div>
        </div>
        <div class="cloud-panel">
          <span class="panel-title">云平台</span>
          <div v-for="item in cloudBreakdown" :key="item.key" class="cloud-row">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
        </div>
      </aside>
    </section>

    <div
      v-show="ctxMenu.visible"
      class="ctx-menu"
      :style="{ left: ctxMenu.x + 'px', top: ctxMenu.y + 'px' }"
      @mouseleave="hideCtxMenu"
    >
      <div class="ctx-item" @click="openGroupDialog(ctxMenu.node, 'createRoot')">
        <el-icon><Plus /></el-icon> 根分组
      </div>
      <div class="ctx-item" @click="openGroupDialog(ctxMenu.node, 'createChild')">
        <el-icon><FolderAdd /></el-icon> 子分组
      </div>
      <div class="ctx-divider" />
      <div class="ctx-item" @click="openGroupDialog(ctxMenu.node, 'rename')">
        <el-icon><Edit /></el-icon> 重命名
      </div>
      <div class="ctx-item" @click="openGroupDialog(ctxMenu.node, 'edit')">
        <el-icon><EditPen /></el-icon> 编辑分组
      </div>
      <div class="ctx-divider" />
      <div class="ctx-item danger" @click="handleDeleteGroup(ctxMenu.node)">
        <el-icon><Delete /></el-icon> 删除分组
      </div>
    </div>

    <el-dialog v-model="hostDialogVisible" :title="hostDialogTitle" width="600px" destroy-on-close>
      <el-form ref="hostFormRef" :model="hostForm" :rules="hostFormRules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="主机名称" prop="name">
              <el-input v-model="hostForm.name" placeholder="请输入主机名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="IP地址" prop="ip">
              <el-input v-model="hostForm.ip" placeholder="192.168.1.1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="SSH端口" prop="port">
              <el-input-number v-model="hostForm.port" :min="1" :max="65535" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="操作系统">
              <el-select v-model="hostForm.osType" placeholder="操作系统" style="width:100%">
                <el-option label="Linux" value="linux" />
                <el-option label="Windows" value="windows" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="云厂商">
              <el-select v-model="hostForm.cloudProvider" placeholder="云厂商" clearable style="width:100%">
                <el-option label="阿里云" value="aliyun" />
                <el-option label="腾讯云" value="tencent" />
                <el-option label="华为云" value="huawei" />
                <el-option label="AWS" value="aws" />
                <el-option label="Azure" value="azure" />
                <el-option label="自建" value="self" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属分组">
              <el-tree-select
                v-model="hostForm.groupId"
                :data="groupTree"
                :props="{ label: 'name', value: 'id', children: 'children' }"
                placeholder="选择分组"
                clearable
                style="width:100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="登录用户" prop="username">
              <el-input v-model="hostForm.username" placeholder="root" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="认证方式">
              <el-select v-model="hostForm.authType" style="width:100%">
                <el-option label="密码" value="password" />
                <el-option label="密钥" value="key" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24" v-if="hostForm.authType === 'password'">
            <el-form-item label="登录密码">
              <el-input v-model="hostForm.password" type="password" show-password placeholder="登录密码" />
            </el-form-item>
          </el-col>
          <el-col :span="24" v-else>
            <el-form-item label="私钥内容">
              <el-input v-model="hostForm.privateKey" type="textarea" :rows="4" placeholder="粘贴 SSH 私钥内容" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="标签">
              <el-select v-model="hostForm.tags" multiple allow-create filterable placeholder="输入后回车添加标签" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="hostForm.remark" type="textarea" :rows="2" placeholder="备注" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="hostDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleHostSubmit">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="groupDialogVisible" :title="groupDialogTitle" width="420px" destroy-on-close>
      <el-form ref="groupFormRef" :model="groupForm" :rules="groupFormRules" label-width="90px">
        <el-form-item label="分组名称" prop="name">
          <el-input v-model="groupForm.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="上级分组" v-if="groupForm.action !== 'createRoot'">
          <el-tree-select
            v-model="groupForm.parentId"
            :data="groupTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="顶级分组（不选则为根）"
            clearable
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="groupForm.remark" type="textarea" :rows="2" placeholder="备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="groupDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleGroupSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search, Plus, Refresh, Folder, Monitor, CircleCheck, CircleClose,
  AlarmClock, FolderAdd, Edit, EditPen, Delete
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getHostGroups, getHostList, createHost, updateHost,
  deleteHost, batchDeleteHosts, getHostStats,
  createHostGroup, updateHostGroup, deleteHostGroup
} from '@/api/asset.js'
import { encryptPassword } from '@/utils/crypto.js'

const router = useRouter()

const loading = ref(false)
const submitting = ref(false)
const groupSearch = ref('')
const groupTree = ref([])
const expandedKeys = ref([])
const selectedGroupId = ref(null)
const selectedHosts = ref([])
const tableData = ref([])
const groupTreeRef = ref(null)

const stats = reactive({ total: 0, linux: 0, windows: 0, offline: 0, huawei: 0, aliyun: 0, aws: 0 })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ name: '', ip: '', status: '' })
const ctxMenu = reactive({ visible: false, x: 0, y: 0, node: null })

const hostDialogVisible = ref(false)
const hostDialogTitle = ref('新建主机')
const hostFormRef = ref(null)
const hostForm = reactive({
  id: null, name: '', ip: '', port: 22, osType: 'linux', cloudProvider: '',
  groupId: null, username: 'root', authType: 'password', password: '', privateKey: '', tags: [], remark: ''
})
const hostFormRules = {
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  ip: [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入SSH端口' }],
  username: [{ required: true, message: '请输入登录用户名', trigger: 'blur' }]
}

const groupDialogVisible = ref(false)
const groupDialogTitle = ref('新建分组')
const groupFormRef = ref(null)
const groupForm = reactive({ id: null, name: '', parentId: null, remark: '', action: 'createRoot' })
const groupFormRules = { name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }] }

const totalGroupCount = computed(() => {
  const countAll = (list) => list.reduce((sum, group) => sum + 1 + countAll(group.children || []), 0)
  return countAll(groupTree.value)
})

const filteredGroups = computed(() => {
  if (!groupSearch.value) return groupTree.value
  const keyword = groupSearch.value.toLowerCase()
  const filter = (list) => list.reduce((acc, group) => {
    const children = filter(group.children || [])
    if (group.name.toLowerCase().includes(keyword) || children.length) {
      acc.push({ ...group, children })
    }
    return acc
  }, [])
  return filter(groupTree.value)
})

const selectedGroupName = computed(() => {
  if (!selectedGroupId.value) return '全部分组'
  const findGroup = (items) => {
    for (const item of items) {
      if (item.id === selectedGroupId.value) return item.name
      const child = findGroup(item.children || [])
      if (child) return child
    }
    return ''
  }
  return findGroup(groupTree.value) || '已选分组'
})

const assetMetrics = computed(() => [
  { key: 'total', label: '总计', value: stats.total, meta: '已纳管主机', tone: 'info', icon: Monitor },
  { key: 'linux', label: 'Linux', value: stats.linux, meta: `${totalGroupCount.value} 个分组`, tone: 'success', icon: CircleCheck },
  { key: 'windows', label: 'Windows', value: stats.windows, meta: '集群节点', tone: 'info', icon: Monitor },
  { key: 'offline', label: '离线', value: stats.offline, meta: '需处理', tone: 'error', icon: CircleClose }
])

const riskHosts = computed(() => tableData.value.filter((host) => {
  return host.status === 'offline' || Number(host.cpuUsage) > 80 || Number(host.memUsage) > 80 || Number(host.diskUsage) > 80
}).slice(0, 8))

const cloudBreakdown = computed(() => [
  { key: 'huawei', label: '华为云', value: stats.huawei },
  { key: 'aliyun', label: '阿里云', value: stats.aliyun },
  { key: 'aws', label: 'AWS', value: stats.aws }
])

onMounted(() => {
  fetchGroups()
  fetchHosts()
  fetchStats()
  document.addEventListener('click', hideCtxMenu)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', hideCtxMenu)
})

async function fetchGroups() {
  try {
    const res = await getHostGroups()
    groupTree.value = res?.data?.data || []
    if (groupTree.value.length && !expandedKeys.value.length) {
      expandedKeys.value = [groupTree.value[0].id]
    }
  } catch (e) {
    ElMessage.error('获取分组列表失败')
  }
}

function handleGroupClick(data) {
  selectedGroupId.value = data.id
  pagination.page = 1
  fetchHosts()
  fetchStats()
}

function handleContextMenu(event, data) {
  event.preventDefault()
  event.stopPropagation()
  ctxMenu.visible = true
  ctxMenu.x = event.clientX
  ctxMenu.y = event.clientY
  ctxMenu.node = data
}

function hideCtxMenu() {
  ctxMenu.visible = false
}

function openGroupDialog(node, action) {
  hideCtxMenu()
  Object.assign(groupForm, { id: null, name: '', parentId: null, remark: '', action })

  if (action === 'createRoot') {
    groupDialogTitle.value = '创建根分组'
  } else if (action === 'createChild') {
    groupDialogTitle.value = '创建子分组'
    groupForm.parentId = node?.id || null
  } else if (action === 'rename') {
    groupDialogTitle.value = '重命名分组'
    groupForm.id = node.id
    groupForm.name = node.name
    groupForm.parentId = node.parentId || null
    groupForm.remark = node.remark || ''
  } else if (action === 'edit') {
    groupDialogTitle.value = '修改分组'
    groupForm.id = node.id
    groupForm.name = node.name
    groupForm.parentId = node.parentId || null
    groupForm.remark = node.remark || ''
  }
  groupDialogVisible.value = true
}

async function handleGroupSubmit() {
  await groupFormRef.value?.validate()
  submitting.value = true
  try {
    const payload = {
      name: groupForm.name,
      parentId: groupForm.parentId || 0,
      remark: groupForm.remark
    }
    if (groupForm.id) {
      await updateHostGroup(groupForm.id, payload)
      ElMessage.success('分组更新成功')
    } else {
      await createHostGroup(payload)
      ElMessage.success('分组创建成功')
    }
    groupDialogVisible.value = false
    fetchGroups()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDeleteGroup(node) {
  hideCtxMenu()
  if (!node) return
  await ElMessageBox.confirm(`确认删除分组「${node.name}」？删除后分组内主机将移至未分组。`, '提示', { type: 'warning' })
  try {
    await deleteHostGroup(node.id)
    ElMessage.success('删除成功')
    if (selectedGroupId.value === node.id) { selectedGroupId.value = null; fetchHosts() }
    fetchGroups()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || e?.message || '删除失败')
  }
}

async function fetchHosts() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      groupId: selectedGroupId.value,
      ...searchForm
    }
    Object.keys(params).forEach(key => (params[key] === '' || params[key] == null) && delete params[key])
    const res = await getHostList(params)
    const data = res?.data?.data || res?.data || {}
    tableData.value = data.list || []
    pagination.total = data.total || 0
  } catch (e) {
    tableData.value = []
  } finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await getHostStats(selectedGroupId.value)
    const data = res?.data?.data || {}
    Object.assign(stats, { total: 0, linux: 0, windows: 0, offline: 0, huawei: 0, aliyun: 0, aws: 0, ...data })
  } catch {}
}

function handleSelectionChange(rows) { selectedHosts.value = rows }

function openHostDialog(row = null) {
  Object.assign(hostForm, {
    id: null, name: '', ip: '', port: 22, osType: 'linux', cloudProvider: '',
    groupId: selectedGroupId.value, username: 'root', authType: 'password', password: '', privateKey: '', tags: [], remark: ''
  })
  if (row) {
    Object.assign(hostForm, {
      id: row.id, name: row.name, ip: row.ip, port: row.port || 22,
      osType: row.osType || 'linux', cloudProvider: row.cloudProvider || '',
      groupId: row.groupId || null, username: row.username || 'root',
      authType: row.authType || 'password', tags: row.tags || [], remark: row.remark || ''
    })
    hostDialogTitle.value = '编辑主机'
  } else {
    hostDialogTitle.value = '新建主机'
  }
  hostDialogVisible.value = true
}

async function handleHostSubmit() {
  await hostFormRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...hostForm }
    if (payload.authType === 'password') {
      if (payload.password) {
        payload.password = encryptPassword(payload.password)
      } else {
        delete payload.password
      }
    }
    if (payload.id) { await updateHost(payload.id, payload) } else { await createHost(payload) }
    ElMessage.success('操作成功')
    hostDialogVisible.value = false
    fetchHosts()
    fetchStats()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDeleteHost(row) {
  await ElMessageBox.confirm(`确认删除主机「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await deleteHost(row.id)
    ElMessage.success('删除成功')
    fetchHosts()
    fetchStats()
  } catch (e) { ElMessage.error(e?.response?.data?.message || e?.message || '删除失败') }
}

async function handleBatchDelete() {
  if (!selectedHosts.value.length) return
  await ElMessageBox.confirm(`确认删除选中的 ${selectedHosts.value.length} 台主机？`, '提示', { type: 'warning' })
  try {
    await batchDeleteHosts(selectedHosts.value.map(host => host.id))
    ElMessage.success('批量删除成功')
    fetchHosts()
    fetchStats()
  } catch (e) { ElMessage.error(e?.response?.data?.message || e?.message || '操作失败') }
}

function handleAlarm() { ElMessage.info('告警功能待接入监控系统') }
function handleTerminal() {
  if (!selectedHosts.value.length) { ElMessage.warning('请先选择主机'); return }
  const host = selectedHosts.value[0]
  router.push({ path: '/jumpserver/terminal', query: { hostId: host.id } })
}
function handleTerminalRow(row) { router.push({ path: '/jumpserver/terminal', query: { hostId: row.id } }) }

const getCpuClass = (value) => value > 90 ? 'usage-danger' : value > 70 ? 'usage-warning' : 'usage-ok'
const getMemClass = (value) => value > 90 ? 'usage-danger' : value > 80 ? 'usage-warning' : 'usage-ok'
const getDiskClass = (value) => value > 90 ? 'usage-danger' : value > 80 ? 'usage-warning' : 'usage-ok'
const statusTone = (status) => ({ online: 'success', offline: 'error', unknown: 'neutral' }[status] || 'neutral')
const statusText = (status) => ({ online: '在线', offline: '离线', unknown: '未知' }[status] || status || '未知')
const cloudText = (cloud) => ({ aliyun: '阿里云', tencent: '腾讯云', huawei: '华为云', aws: 'AWS', azure: 'Azure', self: '自建' }[cloud] || '自建')
</script>

<style scoped>
.host-management {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--ds-space-16);
  color: var(--ds-text-primary);
}

.asset-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ds-space-16);
}

.eyebrow {
  font-size: var(--ds-font-size-11);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--ds-text-muted);
}

.asset-header h1 {
  margin: var(--ds-space-2) 0 0;
  font-size: var(--ds-font-size-20);
  line-height: 1.1;
}

.header-actions,
.toolbar-controls,
.icon-actions,
.row-actions,
.signal-row,
.tag-row {
  display: flex;
  align-items: center;
  gap: var(--ds-space-8);
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--ds-space-8);
}

.metric-card,
.group-panel,
.inventory-panel,
.side-panel,
.cloud-panel {
  background: var(--ds-bg-surface);
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-8);
}

.metric-card {
  padding: var(--ds-space-10);
}

.metric-meta,
.metric-foot,
.panel-subtitle,
.host-ip,
.risk-meta,
.risk-load,
.muted {
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
  margin-top: var(--ds-space-6);
  font-size: var(--ds-font-size-22);
  font-weight: 750;
  letter-spacing: -0.03em;
}

.metric-foot {
  margin-top: var(--ds-space-2);
  font-size: var(--ds-font-size-11);
}

.metric-card.is-success .metric-meta,
.metric-card.is-success .metric-value { color: var(--ds-success); }
.metric-card.is-error .metric-meta,
.metric-card.is-error .metric-value { color: var(--ds-error); }
.metric-card.is-info .metric-meta,
.metric-card.is-info .metric-value { color: var(--ds-info); }

.asset-layout {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr) 260px;
  gap: var(--ds-space-12);
  min-height: 0;
  flex: 1;
}

.group-panel,
.inventory-panel,
.side-panel {
  min-height: 0;
  overflow: hidden;
}

.group-panel,
.side-panel {
  display: flex;
  flex-direction: column;
}

.inventory-panel {
  display: flex;
  flex-direction: column;
  padding: var(--ds-space-10);
}

.panel-head,
.panel-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ds-space-8);
  padding: var(--ds-space-10);
  border-bottom: 1px solid var(--ds-border-subtle);
}

.panel-toolbar {
  padding: 0 0 var(--ds-space-8);
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

.group-search {
  padding: var(--ds-space-8);
}

.group-tree-scroll {
  flex: 1;
  padding: 0 var(--ds-space-8) var(--ds-space-8);
}

.tree-node {
  width: 100%;
  display: flex;
  align-items: center;
  gap: var(--ds-space-8);
  min-width: 0;
}

.node-folder {
  color: var(--ds-warning);
  font-size: var(--ds-font-size-14);
}

.node-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--ds-text-secondary);
}

.node-count {
  min-width: 22px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--ds-radius-4);
  background: var(--ds-bg-surface-2);
  color: var(--ds-text-tertiary);
  font-size: var(--ds-font-size-11);
}

.filter-input { width: 150px; }
.filter-select { width: 132px; }

.asset-table {
  flex: 1;
  margin-top: var(--ds-space-8);
}

.host-cell {
  display: flex;
  flex-direction: column;
  gap: var(--ds-space-2);
}

.host-name {
  font-weight: 650;
  color: var(--ds-text-primary);
}

.host-ip,
.risk-meta,
.risk-load {
  font-size: var(--ds-font-size-12);
}

.status-pill,
.cloud-pill,
.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--ds-space-6);
  min-height: 22px;
  padding: 0 var(--ds-space-8);
  border-radius: var(--ds-radius-6);
  border: 1px solid var(--ds-border-subtle);
  font-size: var(--ds-font-size-12);
  font-weight: 650;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: currentColor;
}

.status-pill.is-success { color: var(--ds-success); background: var(--ds-bg-success-subtle); }
.status-pill.is-error { color: var(--ds-error); background: var(--ds-bg-danger-subtle); }
.status-pill.is-neutral { color: var(--ds-text-tertiary); background: var(--ds-bg-surface-2); }
.cloud-pill { color: var(--ds-text-secondary); background: var(--ds-bg-surface-2); }
.cloud-pill.is-aliyun { color: var(--ds-warning); }
.cloud-pill.is-huawei { color: var(--ds-error); }
.cloud-pill.is-aws { color: var(--ds-warning); }
.cloud-pill.is-azure { color: var(--ds-info); }

.signal-row {
  gap: var(--ds-space-6);
  color: var(--ds-text-tertiary);
  font-size: var(--ds-font-size-11);
  white-space: nowrap;
}

.tag-row {
  flex-wrap: wrap;
  gap: var(--ds-space-4);
}

.tag-chip {
  min-height: 20px;
  padding: 0 var(--ds-space-6);
  color: var(--ds-text-secondary);
  background: var(--ds-bg-surface-2);
}

.usage-ok { color: var(--ds-success); }
.usage-warning { color: var(--ds-warning); font-weight: 700; }
.usage-danger { color: var(--ds-error); font-weight: 700; }

.pagination {
  margin-top: var(--ds-space-8);
  display: flex;
  justify-content: flex-end;
}

.side-panel {
  padding: var(--ds-space-10);
  gap: var(--ds-space-10);
}

.panel-head.compact {
  padding: 0 0 var(--ds-space-6);
}

.risk-list {
  display: flex;
  flex-direction: column;
  gap: var(--ds-space-8);
}

.risk-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--ds-space-2);
  width: 100%;
  padding: var(--ds-space-8);
  border: 1px solid var(--ds-border-subtle);
  border-radius: var(--ds-radius-6);
  background: var(--ds-bg-surface-2);
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.risk-item:hover {
  border-color: var(--ds-border-strong);
  background: var(--ds-bg-hover);
}

.risk-main {
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

.cloud-panel {
  padding: var(--ds-space-8);
}

.cloud-row {
  display: flex;
  justify-content: space-between;
  padding: var(--ds-space-6) 0;
  color: var(--ds-text-secondary);
  border-bottom: 1px solid var(--ds-border-subtle);
}

.cloud-row:last-child {
  border-bottom: 0;
}

.ctx-menu {
  position: fixed;
  z-index: 9999;
  min-width: 168px;
  padding: var(--ds-space-4) 0;
  background: var(--ds-bg-elevated);
  border: 1px solid var(--ds-border-default);
  border-radius: var(--ds-radius-8);
}

.ctx-item {
  display: flex;
  align-items: center;
  gap: var(--ds-space-8);
  padding: var(--ds-space-8) var(--ds-space-12);
  font-size: var(--ds-font-size-12);
  color: var(--ds-text-secondary);
  cursor: pointer;
}

.ctx-item:hover {
  background: var(--ds-bg-hover);
  color: var(--ds-text-primary);
}

.ctx-item.danger {
  color: var(--ds-error);
}

.ctx-divider {
  height: 1px;
  margin: var(--ds-space-4) 0;
  background: var(--ds-border-subtle);
}

:deep(.el-tree) {
  --el-tree-node-hover-bg-color: var(--ds-bg-hover);
  background: transparent;
  color: var(--ds-text-secondary);
}

:deep(.el-tree-node__content) {
  height: 32px;
  border-radius: var(--ds-radius-6);
}

:deep(.el-tree--highlight-current .el-tree-node.is-current > .el-tree-node__content) {
  background: var(--ds-bg-active);
}

@media (max-width: 1440px) {
  .asset-layout {
    grid-template-columns: 240px minmax(0, 1fr);
  }

  .side-panel {
    display: none;
  }
}

@media (max-width: 960px) {
  .asset-header,
  .panel-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .asset-layout {
    grid-template-columns: 1fr;
  }

  .group-panel {
    min-height: 280px;
  }
}
</style>
