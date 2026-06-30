<template>
  <div class="approval-page">
    <header class="page-header">
      <div>
        <div class="eyebrow">审批管理</div>
        <h1>访问审批</h1>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openApplyDialog()">发起审批</el-button>
      </div>
    </header>

    <section class="tab-bar">
      <el-radio-group v-model="tabFilter" @change="onTabChange">
        <el-radio-button value="pending">待审批</el-radio-button>
        <el-radio-button value="my">我的申请</el-radio-button>
        <el-radio-button value="all">全部</el-radio-button>
      </el-radio-group>
    </section>

    <el-table :data="tableData" v-loading="loading">
      <el-table-column prop="applicantName" label="申请人" width="120" />
      <el-table-column label="目标主机" min-width="180">
        <template #default="{ row }">
          <div>{{ row.hostName || '-' }}</div>
          <div class="text-muted">{{ row.hostIp || '-' }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="reason" label="申请原因" min-width="150" show-overflow-tooltip />
      <el-table-column label="时长" width="100">
        <template #default="{ row }"> {{ formatDuration(row.duration) }} </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="审批人" width="120">
        <template #default="{ row }"> {{ row.approverName || '-' }} </template>
      </el-table-column>
      <el-table-column label="审批时间" width="160">
        <template #default="{ row }"> {{ formatTime(row.approvedAt) }} </template>
      </el-table-column>
      <el-table-column label="过期时间" width="160">
        <template #default="{ row }"> {{ formatTime(row.expiredAt) }} </template>
      </el-table-column>
      <el-table-column label="申请时间" width="160">
        <template #default="{ row }"> {{ formatTime(row.createdAt) }} </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'pending'">
            <el-button link type="success" size="small" @click="handleApprove(row)">通过</el-button>
            <el-button link type="danger" size="small" @click="handleReject(row)">拒绝</el-button>
          </template>
          <span v-else class="text-muted">-</span>
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
      @size-change="fetchList"
      @current-change="fetchList"
    />

    <!-- 发起审批弹窗 -->
    <el-dialog v-model="applyDialogVisible" title="发起审批申请" width="550px" destroy-on-close>
      <el-form ref="applyFormRef" :model="applyForm" :rules="applyRules" label-width="100px">
        <el-form-item label="目标主机" prop="hostId">
          <el-select v-model="applyForm.hostId" filterable placeholder="选择要访问的主机" style="width:100%">
            <el-option v-for="h in hostOptions" :key="h.id" :label="`${h.name} (${h.ip})`" :value="h.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="认证凭证" prop="credentialId">
          <el-select v-model="applyForm.credentialId" filterable placeholder="选择认证凭证" style="width:100%">
            <el-option v-for="c in credentialOptions" :key="c.id" :label="`${c.name} (${c.username})`" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="访问时长" prop="duration">
          <el-select v-model="applyForm.duration" style="width:100%">
            <el-option label="1 小时" :value="3600" />
            <el-option label="2 小时" :value="7200" />
            <el-option label="4 小时" :value="14400" />
            <el-option label="8 小时" :value="28800" />
            <el-option label="24 小时" :value="86400" />
          </el-select>
        </el-form-item>
        <el-form-item label="申请原因" prop="reason">
          <el-input v-model="applyForm.reason" type="textarea" :rows="3" placeholder="请说明申请原因，如：需要排查线上故障" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleApply">提交申请</el-button>
      </template>
    </el-dialog>

    <!-- 审批意见弹窗 -->
    <el-dialog v-model="remarkDialogVisible" :title="remarkTitle" width="450px" destroy-on-close>
      <el-input v-model="remarkText" type="textarea" :rows="3" placeholder="审批意见（可选）" />
      <template #footer>
        <el-button @click="remarkDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="confirmRemark">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getApprovals, createApproval, approveApproval, rejectApproval } from '@/api/jumpserver.js'
import { getAllCredentials } from '@/api/jumpserver.js'
import { getHostList } from '@/api/asset.js'

const loading = ref(false)
const submitting = ref(false)
const tableData = ref([])
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const tabFilter = ref('pending')
const hostOptions = ref([])
const credentialOptions = ref([])

// 发起审批
const applyDialogVisible = ref(false)
const applyFormRef = ref(null)
const applyForm = reactive({ hostId: null, credentialId: null, duration: 3600, reason: '' })
const applyRules = {
  hostId: [{ required: true, message: '请选择目标主机', trigger: 'change' }],
  credentialId: [{ required: true, message: '请选择认证凭证', trigger: 'change' }],
  duration: [{ required: true, message: '请选择访问时长', trigger: 'change' }]
}

// 审批意见
const remarkDialogVisible = ref(false)
const remarkTitle = ref('')
const remarkText = ref('')
let pendingApprovalId = null
let pendingAction = null // 'approve' | 'reject'

onMounted(async () => {
  fetchList()
  // 加载主机列表
  try {
    const hostRes = await getHostList({ page: 1, pageSize: 200 })
    hostOptions.value = hostRes?.data?.data?.list || []
  } catch {}
  // 加载凭证列表
  try {
    const credRes = await getAllCredentials()
    credentialOptions.value = credRes?.data?.data || []
  } catch {}
})

function onTabChange() {
  pagination.page = 1
  fetchList()
}

async function fetchList() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize }
    if (tabFilter.value === 'my') {
      params.myApplies = true
    } else if (tabFilter.value === 'pending') {
      params.status = 'pending'
    }
    const res = await getApprovals(params)
    const data = res?.data?.data || {}
    tableData.value = data.list || []
    pagination.total = data.total || 0
  } catch { tableData.value = [] }
  finally { loading.value = false }
}

function openApplyDialog() {
  Object.assign(applyForm, { hostId: null, credentialId: null, duration: 3600, reason: '' })
  applyDialogVisible.value = true
}

async function handleApply() {
  await applyFormRef.value?.validate()
  submitting.value = true
  try {
    await createApproval({ ...applyForm })
    ElMessage.success('审批申请已提交，请等待审批人处理')
    applyDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || '提交失败')
  }
  finally { submitting.value = false }
}

function handleApprove(row) {
  pendingApprovalId = row.id
  pendingAction = 'approve'
  remarkTitle.value = '审批通过'
  remarkText.value = ''
  remarkDialogVisible.value = true
}

function handleReject(row) {
  pendingApprovalId = row.id
  pendingAction = 'reject'
  remarkTitle.value = '审批拒绝'
  remarkText.value = ''
  remarkDialogVisible.value = true
}

async function confirmRemark() {
  submitting.value = true
  try {
    if (pendingAction === 'approve') {
      await approveApproval(pendingApprovalId, { remark: remarkText.value })
      ElMessage.success('审批已通过')
    } else {
      await rejectApproval(pendingApprovalId, { remark: remarkText.value })
      ElMessage.success('审批已拒绝')
    }
    remarkDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  }
  finally { submitting.value = false }
}

function statusTag(status) {
  return { pending: 'warning', approved: 'success', rejected: 'danger' }[status] || 'info'
}

function statusLabel(status) {
  return { pending: '待审批', approved: '已通过', rejected: '已拒绝' }[status] || status
}

function formatDuration(seconds) {
  if (!seconds) return '-'
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟`
  return `${Math.floor(seconds / 3600)} 小时`
}

function formatTime(t) { return t ? new Date(t).toLocaleString('zh-CN') : '-' }
</script>

<style scoped>
.approval-page {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  color: var(--ds-text-primary);
}
.page-header { display: flex; align-items: center; justify-content: space-between; }
.eyebrow { font-size: 11px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; color: var(--ds-text-muted); }
.page-header h1 { margin: 2px 0 0; font-size: 20px; }
.header-actions { display: flex; gap: 8px; }
.tab-bar { display: flex; align-items: center; }
.pagination { margin-top: 8px; display: flex; justify-content: flex-end; }
.text-muted { font-size: 12px; color: var(--ds-text-tertiary); }
</style>