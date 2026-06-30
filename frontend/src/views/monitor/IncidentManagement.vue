<template>
  <div class="incident-page">
    <header class="incident-header">
      <div class="title-row">
        <h1>告警事件</h1>
        <span>{{ pagination.total }} 条记录</span>
      </div>
      <div class="header-actions">
        <el-button size="small" @click="resetSearch">
          <el-icon><Refresh /></el-icon>
          重置
        </el-button>
        <el-button size="small" type="primary" @click="openDialog()">
          <el-icon><Plus /></el-icon>
          新增
        </el-button>
      </div>
    </header>

    <section class="metric-grid">
      <div v-for="metric in incidentMetrics" :key="metric.key" class="metric-card" :class="`is-${metric.tone}`">
        <div class="metric-label">
          <span>{{ metric.label }}</span>
          <el-icon><component :is="metric.icon" /></el-icon>
        </div>
        <strong>{{ metric.value }}</strong>
        <small>{{ metric.meta }}</small>
      </div>
    </section>

    <section class="incident-layout">
      <div class="panel table-panel">
        <div class="toolbar">
          <div class="toolbar-left">
            <el-select v-model="searchForm.businessLine" size="small" placeholder="业务" clearable class="filter-control">
              <el-option v-for="b in businessLines" :key="b" :label="b" :value="b" />
            </el-select>
            <el-select v-model="searchForm.level" size="small" placeholder="级别" clearable class="filter-control sm">
              <el-option v-for="l in LEVELS" :key="l.value" :label="l.value" :value="l.value" />
            </el-select>
            <el-select v-model="searchForm.status" size="small" placeholder="状态" clearable class="filter-control sm">
              <el-option label="待处理" value="pending" />
              <el-option label="处理中" value="processing" />
              <el-option label="已完成" value="done" />
            </el-select>
            <el-select v-model="searchForm.dept" size="small" placeholder="部门" clearable class="filter-control sm">
              <el-option v-for="d in DEPTS" :key="d" :label="d" :value="d" />
            </el-select>
          </div>
          <div class="toolbar-right">
            <el-button size="small" type="primary" :loading="loading" @click="fetchIncidents">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
          </div>
        </div>

        <el-table :data="tableData" v-loading="loading" class="incident-table" style="width: 100%" height="calc(100vh - 340px)">
          <el-table-column prop="alertTime" label="时间" width="142" show-overflow-tooltip />
          <el-table-column prop="businessLine" label="业务" width="120" show-overflow-tooltip />
          <el-table-column label="级别" width="82">
            <template #default="{ row }"><span class="level-pill" :class="levelTone(row.level)">{{ row.level }}</span></template>
          </el-table-column>
          <el-table-column label="频率" width="92">
            <template #default="{ row }"><span class="status-pill neutral">{{ row.frequency || '偶发' }}</span></template>
          </el-table-column>
          <el-table-column prop="alertDesc" label="告警" min-width="260" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="alert-cell">
                <strong>{{ row.alertDesc }}</strong>
                <small>{{ row.detail || '-' }}</small>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="dept" label="部门" width="100" show-overflow-tooltip />
          <el-table-column prop="handler" label="负责人" width="100" show-overflow-tooltip />
          <el-table-column label="状态" width="112">
            <template #default="{ row }"><span class="status-pill" :class="statusTone(row.status)">{{ statusText(row.status) }}</span></template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right" align="right">
            <template #default="{ row }">
              <el-button-group>
                <el-button size="small" :disabled="row.status !== 'pending'" @click="handleStatusChange(row, 'processing')"><el-icon><VideoPlay /></el-icon></el-button>
                <el-button size="small" :disabled="row.status === 'done'" type="success" plain @click="handleStatusChange(row, 'done')"><el-icon><CircleCheckFilled /></el-icon></el-button>
                <el-button size="small" @click="openDialog(row)"><el-icon><Edit /></el-icon></el-button>
                <el-button size="small" type="danger" plain @click="handleDelete(row)"><el-icon><Delete /></el-icon></el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          class="pagination"
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchIncidents"
          @current-change="fetchIncidents"
        />
      </div>
    </section>

    <!-- ==================== 新增/编辑故障 Dialog ==================== -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="告警时间" prop="alertTime">
              <el-date-picker
                v-model="form.alertTime"
                type="datetime"
                placeholder="选择时间"
                format="YYYY/MM/DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width:100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="业务线" prop="businessLine">
              <el-select v-model="form.businessLine" allow-create filterable placeholder="业务线" style="width:100%">
                <el-option v-for="b in businessLines" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="故障级别" prop="level">
              <el-select v-model="form.level" placeholder="故障级别" style="width:100%">
                <el-option v-for="l in LEVELS" :key="l.value" :label="l.label" :value="l.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="频率" prop="frequency">
              <el-select v-model="form.frequency" placeholder="频率" style="width:100%">
                <el-option label="偶发" value="偶发" />
                <el-option label="频繁" value="频繁" />
                <el-option label="持续" value="持续" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属部门" prop="dept">
              <el-select v-model="form.dept" allow-create filterable placeholder="所属部门" style="width:100%">
                <el-option v-for="d in DEPTS" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="处理人" prop="handler">
              <el-input v-model="form.handler" placeholder="处理人姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="告警描述" prop="alertDesc">
              <el-input v-model="form.alertDesc" type="textarea" :rows="2" placeholder="告警描述" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="故障详情">
              <el-input v-model="form.detail" type="textarea" :rows="4" placeholder="故障详情、影响范围、处理过程..." />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="处理状态">
              <el-select v-model="form.status" style="width:100%">
                <el-option label="未处理" value="pending" />
                <el-option label="处理中" value="processing" />
                <el-option label="已完成" value="done" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="解决时间">
              <el-date-picker
                v-model="form.resolvedAt"
                type="datetime"
                placeholder="解决时间"
                format="YYYY/MM/DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width:100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import {
  Document, Warning, CircleCheck, Search, Refresh, Plus,
  Edit, Delete, VideoPlay, CircleCheckFilled, Loading
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getIncidentStats, getIncidentList, createIncident,
  updateIncident, deleteIncident, updateIncidentStatus, getBusinessLines
} from '@/api/incident.js'

// ===================== 常量 =====================
const LEVELS = [
  { value: 'P1', label: 'P1 - 严重' },
  { value: 'P2', label: 'P2 - 高' },
  { value: 'P3', label: 'P3 - 中' },
  { value: 'P4', label: 'P4 - 低' },
]
const DEPTS = ['运维部', '研发部', '产品部', '测试部', '业务部']

// ===================== 状态 =====================
const loading = ref(false)
const submitting = ref(false)
const tableData = ref([])
const businessLines = ref([])
const incidentStats = reactive({ total: 0, pending: 0, processing: 0, done: 0 })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ businessLine: '', level: '', status: '', dept: '' })
const incidentMetrics = computed(() => [
  { key: 'total', label: '总计', value: incidentStats.total, meta: '条记录', tone: 'info', icon: Document },
  { key: 'pending', label: '待处理', value: incidentStats.pending, meta: '待分类', tone: 'error', icon: Warning },
  { key: 'processing', label: '处理中', value: incidentStats.processing, meta: '进行中', tone: 'warning', icon: Loading },
  { key: 'done', label: '已完成', value: incidentStats.done, meta: '已解决', tone: 'success', icon: CircleCheck }
])
const activeIncidents = computed(() => tableData.value.filter((item) => item.status !== 'done').slice(0, 8))

const dialogVisible = ref(false)
const dialogTitle = ref('添加故障记录')
const formRef = ref(null)
const form = reactive({
  id: null, alertTime: '', businessLine: '', level: 'P4', frequency: '偶发',
  alertDesc: '', detail: '', dept: '', handler: 'admin', status: 'pending', resolvedAt: ''
})
const formRules = {
  alertTime: [{ required: true, message: '请选择告警时间', trigger: 'change' }],
  businessLine: [{ required: true, message: '请选择业务线', trigger: 'change' }],
  level: [{ required: true, message: '请选择故障级别' }],
  alertDesc: [{ required: true, message: '请填写告警描述', trigger: 'blur' }],
  dept: [{ required: true, message: '请选择所属部门' }],
}

// ===================== 生命周期 =====================
onMounted(() => {
  fetchStats()
  fetchIncidents()
  fetchBusinessLines()
})

// ===================== 数据获取 =====================
async function fetchStats() {
  try {
    const res = await getIncidentStats()
    Object.assign(incidentStats, res?.data?.data || {})
  } catch {}
}

async function fetchIncidents() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize, ...searchForm }
    Object.keys(params).forEach(k => !params[k] && delete params[k])
    const res = await getIncidentList(params)
    const d = res?.data?.data || {}
    tableData.value = d.list || []
    pagination.total = d.total || 0
  } catch { tableData.value = [] } finally { loading.value = false }
}

async function fetchBusinessLines() {
  try {
    const res = await getBusinessLines()
    businessLines.value = res?.data?.data || []
  } catch { businessLines.value = ['运维部门', '研发团队', '信息中心'] }
}

// ===================== 操作方法 =====================
function resetSearch() {
  Object.assign(searchForm, { businessLine: '', level: '', status: '', dept: '' })
  pagination.page = 1
  fetchIncidents()
}

function openDialog(row = null) {
  Object.assign(form, {
    id: null, alertTime: '', businessLine: '', level: 'P4', frequency: '偶发',
    alertDesc: '', detail: '', dept: '', handler: 'admin', status: 'pending', resolvedAt: ''
  })
  if (row) {
    Object.assign(form, {
      id: row.id, alertTime: row.alertTime, businessLine: row.businessLine,
      level: row.level, frequency: row.frequency || '偶发', alertDesc: row.alertDesc,
      detail: row.detail || '', dept: row.dept, handler: row.handler, status: row.status, resolvedAt: row.resolvedAt || ''
    })
    dialogTitle.value = '编辑故障记录'
  } else { dialogTitle.value = '添加故障记录' }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (form.id) { await updateIncident(form.id, form) } else { await createIncident(form) }
    ElMessage.success('操作成功'); dialogVisible.value = false; fetchIncidents(); fetchStats()
  } catch (e) { ElMessage.error(e?.message || '操作失败') } finally { submitting.value = false }
}

async function handleStatusChange(row, newStatus) {
  try {
    await updateIncidentStatus(row.id, { status: newStatus })
    row.status = newStatus
    ElMessage.success(`状态已更新为：${getStatusLabel(newStatus)}`)
    fetchStats()
  } catch (e) { ElMessage.error(e?.message || '操作失败') }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除故障记录「${row.alertDesc}」？`, '提示', { type: 'warning' })
  try { await deleteIncident(row.id); ElMessage.success('删除成功'); fetchIncidents(); fetchStats() }
  catch (e) { ElMessage.error(e?.message || '删除失败') }
}

// ===================== 工具函数 =====================
const getLevelTagType = (level) => {
  const types = { P1: 'danger', P2: 'warning', P3: '', P4: 'info' }
  return types[level] || 'info'
}
const getStatusLabel = (s) => ({ pending: '未处理', processing: '处理中', done: '已完成' }[s] || s)
const getStatusTagType = (s) => ({ pending: 'danger', processing: 'warning', done: 'success' }[s] || 'info')
const levelTone = (level) => ({ P1: 'error', P2: 'warning', P3: 'info', P4: 'neutral' }[level] || 'neutral')
const statusTone = (status) => ({ pending: 'error', processing: 'warning', done: 'success' }[status] || 'neutral')
const statusText = (status) => ({ pending: '待处理', processing: '处理中', done: '已完成' }[status] || status || '未知')
</script>

<style scoped>
.incident-page {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 8px;
  color: var(--ds-text-primary);
}

.incident-header,
.metric-card,
.panel {
  border: 1px solid var(--ds-border-default);
  border-radius: 8px;
  background: var(--ds-bg-surface);
}

.incident-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 42px;
  padding: 0 10px;
}

.title-row {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.title-row h1 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.title-row span,
.metric-card small,
.alert-cell small {
  color: var(--ds-text-muted);
  font-size: 11px;
}

.header-actions {
  display: flex;
  gap: 6px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.metric-card {
  min-height: 78px;
  padding: 10px;
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
  margin-top: 8px;
  font-size: 22px;
  font-weight: 600;
  line-height: 1;
}

.metric-card.is-success { border-color: rgba(34, 197, 94, .28); }
.metric-card.is-warning { border-color: rgba(245, 158, 11, .28); }
.metric-card.is-error { border-color: rgba(239, 68, 68, .28); }
.metric-card.is-info { border-color: rgba(59, 130, 246, .28); }

.incident-layout {
  display: flex;
  flex-direction: column;
}

.panel {
  min-width: 0;
  overflow: hidden;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 38px;
  padding: 0 12px;
  border-bottom: 1px solid var(--ds-border-subtle);
}

.panel-header h2 {
  margin: 0;
  color: var(--ds-text-secondary);
  font-size: 12px;
  font-weight: 600;
}

.toolbar {
  min-height: 42px;
  gap: 6px;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.filter-control {
  width: 150px;
}

.filter-control.sm {
  width: 116px;
}

.incident-table :deep(.el-table__cell) {
  font-size: 12px;
}

.alert-cell,
.queue-main {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.alert-cell strong,
.queue-main strong {
  overflow: hidden;
  color: var(--ds-text-secondary);
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.level-pill,
.status-pill {
  display: inline-flex;
  align-items: center;
  height: 20px;
  padding: 0 8px;
  border-radius: 999px;
  font-size: 11px;
  line-height: 20px;
}

.level-pill.error,
.status-pill.error { color: var(--ds-error); background: var(--ds-bg-danger-subtle); }
.level-pill.warning,
.status-pill.warning { color: var(--ds-warning); background: var(--ds-bg-warning-subtle); }
.level-pill.info { color: var(--ds-info); background: var(--ds-bg-info-subtle); }
.level-pill.neutral,
.status-pill.neutral { color: var(--ds-text-tertiary); background: var(--ds-bg-surface-3); }
.status-pill.success { color: var(--ds-success); background: var(--ds-bg-success-subtle); }

.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 8px;
  border-top: 1px solid var(--ds-border-subtle);
}

@media (max-width: 1280px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
