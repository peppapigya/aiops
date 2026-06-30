<template>
  <div class="credential-page">
    <header class="page-header">
      <div>
        <div class="eyebrow">凭证管理</div>
        <h1>认证凭证</h1>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openDialog()">新增凭证</el-button>
      </div>
    </header>

    <section class="filter-bar">
      <el-input v-model="filter.name" placeholder="凭证名称" clearable class="filter-input" @change="fetchList" />
      <el-select v-model="filter.type" placeholder="类型" clearable class="filter-select" @change="fetchList">
        <el-option label="密码" value="password" />
        <el-option label="密钥" value="private_key" />
        <el-option label="Token" value="token" />
      </el-select>
      <el-input v-model="filter.username" placeholder="用户名" clearable class="filter-input" @change="fetchList" />
      <el-button type="primary" :icon="Search" @click="fetchList">搜索</el-button>
    </section>

    <el-table :data="tableData" v-loading="loading">
      <el-table-column label="凭证名称" min-width="150">
        <template #default="{ row }">
          <div class="cred-name">
            <el-icon><component :is="row.type === 'password' ? Lock : Key" /></el-icon>
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="row.type === 'password' ? '' : 'success'" size="small">
            {{ { password: '密码', private_key: '密钥', token: 'Token' }[row.type] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="username" label="用户名" width="120" />
      <el-table-column label="协议" width="80">
        <template #default="{ row }"> {{ row.protocol?.toUpperCase() || 'SSH' }} </template>
      </el-table-column>
      <el-table-column label="全局" width="70">
        <template #default="{ row }">
          <el-tag :type="row.isGlobal ? 'success' : 'info'" size="small">{{ row.isGlobal ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="priority" label="优先级" width="70" />
      <el-table-column label="备注" min-width="150" show-overflow-tooltip>
        <template #default="{ row }"> {{ row.remark || '-' }} </template>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑凭证' : '新增凭证'" width="560px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="凭证名称" prop="name">
          <el-input v-model="form.name" placeholder="如：生产环境root" />
        </el-form-item>
        <el-form-item label="凭证类型" prop="type">
          <el-select v-model="form.type" style="width:100%">
            <el-option label="密码" value="password" />
            <el-option label="SSH 密钥" value="private_key" />
            <el-option label="Token" value="token" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="如：root" />
        </el-form-item>
        <el-form-item label="协议" prop="protocol">
          <el-select v-model="form.protocol" style="width:100%">
            <el-option label="SSH" value="ssh" />
            <el-option label="RDP" value="rdp" />
            <el-option label="VNC" value="vnc" />
            <el-option label="Telnet" value="telnet" />
          </el-select>
        </el-form-item>
        <template v-if="form.type === 'password'">
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password placeholder="留空则不修改" />
          </el-form-item>
        </template>
        <template v-else-if="form.type === 'private_key'">
          <el-form-item label="私钥内容">
            <el-input v-model="form.privateKey" type="textarea" :rows="5" placeholder="粘贴 SSH 私钥，留空则不修改" />
          </el-form-item>
          <el-form-item label="私钥密码">
            <el-input v-model="form.passphrase" type="password" show-password placeholder="私钥加密密码（如有）" />
          </el-form-item>
        </template>
        <el-form-item label="优先级">
          <el-input-number v-model="form.priority" :min="0" :max="100" />
        </el-form-item>
        <el-form-item label="全局可用">
          <el-switch v-model="form.isGlobal" />
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
import { Search, Refresh, Plus, Lock, Key } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCredentials, createCredential, updateCredential, deleteCredential } from '@/api/jumpserver.js'

const loading = ref(false)
const submitting = ref(false)
const tableData = ref([])
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const filter = reactive({ name: '', type: '', username: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)
const form = reactive({
  id: null, name: '', type: 'password', username: '', password: '',
  privateKey: '', passphrase: '', protocol: 'ssh', priority: 0, isGlobal: false, remark: ''
})
const rules = {
  name: [{ required: true, message: '请输入凭证名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择凭证类型' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  protocol: [{ required: true, message: '请选择协议' }]
}

onMounted(() => fetchList())

async function fetchList() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize }
    if (filter.name) params.name = filter.name
    if (filter.type) params.type = filter.type
    if (filter.username) params.username = filter.username
    const res = await getCredentials(params)
    const data = res?.data?.data || {}
    tableData.value = data.list || []
    pagination.total = data.total || 0
  } catch { tableData.value = [] }
  finally { loading.value = false }
}

function openDialog(row = null) {
  Object.assign(form, {
    id: null, name: '', type: 'password', username: '', password: '',
    privateKey: '', passphrase: '', protocol: 'ssh', priority: 0, isGlobal: false, remark: ''
  })
  if (row) {
    Object.assign(form, {
      id: row.id, name: row.name, type: row.type, username: row.username,
      protocol: row.protocol, priority: row.priority, isGlobal: row.isGlobal, remark: row.remark || ''
    })
    isEdit.value = true
  } else {
    isEdit.value = false
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...form }
    if (form.id) {
      await updateCredential(form.id, payload)
      ElMessage.success('更新成功')
    } else {
      await createCredential(payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除凭证「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await deleteCredential(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch { ElMessage.error('删除失败') }
}

function formatTime(t) {
  return t ? new Date(t).toLocaleString('zh-CN') : '-'
}
</script>

<style scoped>
.credential-page {
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
.cred-name { display: flex; align-items: center; gap: 8px; font-weight: 650; }
.pagination { margin-top: 8px; display: flex; justify-content: flex-end; }
</style>