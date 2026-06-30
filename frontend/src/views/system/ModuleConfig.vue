<template>
  <div class="module-config-page">
    <header class="page-header">
      <div>
        <div class="eyebrow">系统管理</div>
        <h1>模块配置</h1>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
      </div>
    </header>

    <el-alert type="info" :closable="false" show-icon class="tip-alert">
      关闭模块后，该模块的菜单和 API 将不可用，用户需重新登录生效。
    </el-alert>

    <el-table :data="tableData" v-loading="loading">
      <el-table-column prop="moduleKey" label="模块标识" width="150" />
      <el-table-column prop="moduleName" label="模块名称" width="150" />
      <el-table-column prop="description" label="功能描述" min-width="300" show-overflow-tooltip />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.isEnabled ? 'success' : 'danger'" size="default">
            {{ row.isEnabled ? '已启用' : '已停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-switch
            :model-value="row.isEnabled"
            :active-text="row.isEnabled ? '开' : '关'"
            inline-prompt
            :loading="togglingId === row.id"
            @change="handleToggle(row)"
          />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getModuleConfigs, toggleModule } from '@/api/system/module.js'

const loading = ref(false)
const togglingId = ref(null)
const tableData = ref([])

onMounted(() => { fetchList() })

async function fetchList() {
  loading.value = true
  try {
    const res = await getModuleConfigs()
    tableData.value = res?.data?.data || []
  } catch { tableData.value = [] }
  finally { loading.value = false }
}

async function handleToggle(row) {
  const action = row.isEnabled ? '关闭' : '开启'
  try {
    await ElMessageBox.confirm(
      `确认${action}「${row.moduleName}」模块？${row.isEnabled ? '关闭后菜单和功能将不可用。' : ''}`,
      '提示',
      { type: 'warning' }
    )
  } catch {
    return // 取消
  }

  togglingId.value = row.id
  try {
    const res = await toggleModule(row.id)
    const data = res?.data?.data
    ElMessage.success(data?.message || '操作成功')
    row.isEnabled = data?.isEnabled ?? !row.isEnabled
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  }
  finally { togglingId.value = null }
}
</script>

<style scoped>
.module-config-page {
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
.tip-alert { margin: 0; }
</style>