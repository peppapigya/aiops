<template>
  <div class="permission-page">
    <header class="page-header">
      <div>
        <div class="eyebrow">权限管理</div>
        <h1>资产授权规则</h1>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openDialog()">新增规则</el-button>
      </div>
    </header>

    <section class="filter-bar">
      <el-input v-model="filter.name" placeholder="规则名称" clearable class="filter-input" @change="fetchList" />
      <el-select v-model="filter.isActive" placeholder="状态" clearable class="filter-select" @change="fetchList">
        <el-option label="启用" :value="true" />
        <el-option label="禁用" :value="false" />
      </el-select>
      <el-button type="primary" :icon="Search" @click="fetchList">搜索</el-button>
    </section>

    <el-table :data="tableData" v-loading="loading">
      <el-table-column prop="name" label="规则名称" min-width="150" />
      <el-table-column label="授权对象" min-width="150">
        <template #default="{ row }">
          <span>{{ row.userIds?.length ? `${row.userIds.length}个用户` : '所有用户' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="授权资产" min-width="150">
        <template #default="{ row }">
          <span>{{ (row.hostIds?.length || 0) + (row.hostGroupIds?.length || 0) ? `${(row.hostIds?.length || 0) + (row.hostGroupIds?.length || 0)}个资产` : '所有资产' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="凭证" width="80">
        <template #default="{ row }"> {{ row.credentialIds?.length || 0 }}个 </template>
      </el-table-column>
      <el-table-column label="协议" width="120">
        <template #default="{ row }"> {{ (row.protocols || ['ssh']).join(', ') }} </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.isActive ? 'success' : 'info'" size="small">{{ row.isActive ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="需审批" width="80">
        <template #default="{ row }">
          <el-tag :type="row.needApproval ? 'warning' : 'info'" size="small">{{ row.needApproval ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="160">
        <template #default="{ row }"> {{ formatTime(row.createdAt) }} </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openDialog(row)">编辑</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑权限规则' : '新增权限规则'" width="700px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="如：运维组-生产环境授权" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="授权用户">
              <el-select v-model="form.userIds" multiple filterable placeholder="留空=所有用户" style="width:100%">
                <el-option v-for="u in userOptions" :key="u.id" :label="u.username" :value="u.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="授权凭证">
              <el-select v-model="form.credentialIds" multiple filterable placeholder="选择凭证" style="width:100%">
                <el-option v-for="c in credentialOptions" :key="c.id" :label="`${c.name} (${c.username})`" :value="c.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="授权协议">
          <el-select v-model="form.protocols" multiple placeholder="协议" style="width:100%">
            <el-option label="SSH" value="ssh" />
            <el-option label="RDP" value="rdp" />
            <el-option label="VNC" value="vnc" />
            <el-option label="Telnet" value="telnet" />
          </el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="启用状态">
              <el-switch v-model="form.isActive" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="需要审批">
              <el-switch v-model="form.needApproval" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item v-if="form.needApproval" label="审批人">
          <el-select v-model="form.approverIds" multiple filterable placeholder="选择审批人" style="width:100%">
            <el-option v-for="u in userOptions" :key="u.id" :label="`${u.username}${u.nickname ? ' (' + u.nickname + ')' : ''}`" :value="u.id" />
          </el-select>
          <span class="unit-hint">至少指定一位审批人</span>
        </el-form-item>
        <el-form-item label="会话时长限制">
          <el-input-number v-model="form.maxSessionDuration" :min="0" :step="3600" />
          <span class="unit-hint">秒（0=不限制）</span>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPermissions, createPermission, updatePermission, deletePermission } from '@/api/jumpserver.js'
import { getAllCredentials } from '@/api/jumpserver.js'
import { getUserList } from '@/api/system/user.js'

const loading = ref(false)
const submitting = ref(false)
const tableData = ref([])
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const filter = reactive({ name: '', isActive: null })
const userOptions = ref([])
const credentialOptions = ref([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)
const form = reactive({
  id: null, name: '', userIds: [], roleIds: [], hostIds: [], hostGroupIds: [],
  credentialIds: [], protocols: ['ssh'], isActive: true, needApproval: false,
  maxSessionDuration: 0, approverIds: [], remark: ''
})
const rules = { name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }] }

onMounted(async () => {
  fetchList()
  // 获取凭证列表
  try {
    const res = await getAllCredentials()
    credentialOptions.value = res?.data?.data || []
  } catch {}
  // 获取用户列表
  try {
    const userRes = await getUserList({ page: 1, pageSize: 100 })
    userOptions.value = userRes?.data?.data?.list || []
  } catch {}
})

async function fetchList() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize }
    if (filter.name) params.name = filter.name
    if (filter.isActive !== null && filter.isActive !== '') params.isActive = filter.isActive
    const res = await getPermissions(params)
    const data = res?.data?.data || {}
    tableData.value = data.list || []
    pagination.total = data.total || 0
  } catch { tableData.value = [] }
  finally { loading.value = false }
}

function openDialog(row = null) {
  Object.assign(form, {
    id: null, name: '', userIds: [], roleIds: [], hostIds: [], hostGroupIds: [],
    credentialIds: [], protocols: ['ssh'], isActive: true, needApproval: false,
    maxSessionDuration: 0, approverIds: [], remark: ''
  })
  if (row) {
    Object.assign(form, {
      id: row.id, name: row.name, userIds: row.userIds || [], roleIds: row.roleIds || [],
      hostIds: row.hostIds || [], hostGroupIds: row.hostGroupIds || [],
      credentialIds: row.credentialIds || [], protocols: row.protocols || ['ssh'],
      isActive: row.isActive, needApproval: row.needApproval,
      maxSessionDuration: row.maxSessionDuration || 0,
      approverIds: row.approverIds || [], remark: row.remark || ''
    })
    isEdit.value = true
  } else { isEdit.value = false }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...form }
    if (form.id) {
      await updatePermission(form.id, payload)
      ElMessage.success('更新成功')
    } else {
      await createPermission(payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { submitting.value = false }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除权限规则「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await deletePermission(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch { ElMessage.error('删除失败') }
}

function formatTime(t) { return t ? new Date(t).toLocaleString('zh-CN') : '-' }
</script>

<style scoped>
.permission-page {
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
.filter-bar { display: flex; gap: 8px; align-items: center; }
.filter-input { width: 200px; }
.filter-select { width: 120px; }
.pagination { margin-top: 8px; display: flex; justify-content: flex-end; }
.unit-hint { margin-left: 8px; font-size: 12px; color: var(--ds-text-tertiary); }
</style>