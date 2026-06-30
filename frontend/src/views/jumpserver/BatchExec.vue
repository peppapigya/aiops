<template>
  <div class="batch-exec-page">
    <header class="page-header">
      <div>
        <div class="eyebrow">批量操作</div>
        <h1>批量命令执行</h1>
      </div>
    </header>

    <div class="batch-layout">
      <div class="batch-form">
        <el-form label-width="90px">
          <el-form-item label="目标主机">
            <el-select v-model="selectedHosts" multiple filterable placeholder="选择主机" style="width:100%">
              <el-option v-for="h in allHosts" :key="h.id" :label="`${h.name} (${h.ip})`" :value="h.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="认证凭证">
            <el-select v-model="selectedCredentials" multiple placeholder="选择凭证" style="width:100%">
              <el-option v-for="c in credentials" :key="c.id" :label="`${c.name} (${c.username})`" :value="c.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="执行命令">
            <el-input
              v-model="command"
              type="textarea"
              :rows="6"
              placeholder="输入要批量执行的命令，如：&#10;uptime&#10;df -h&#10;free -m"
              class="command-input"
            />
          </el-form-item>
          <el-form-item label="超时时间">
            <el-input-number v-model="timeout" :min="5" :max="300" :step="5" />
            <span class="unit-hint">秒</span>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :icon="CaretRight" :loading="executing" @click="handleBatchExec">
              开始执行
            </el-button>
            <el-button @click="clearResults">清空结果</el-button>
          </el-form-item>
        </el-form>
      </div>

      <div class="batch-results" v-if="results.length">
        <div class="results-header">
          <span>执行结果</span>
          <span class="result-summary">
            成功 {{ results.filter(r => r.success).length }} / 失败 {{ results.filter(r => !r.success).length }}
          </span>
        </div>
        <div class="results-list">
          <div
            v-for="(r, idx) in results"
            :key="idx"
            class="result-item"
            :class="{ success: r.success, failure: !r.success }"
          >
            <div class="result-header">
              <div class="result-host">
                <el-icon><component :is="r.success ? CircleCheck : CircleClose" /></el-icon>
                <span class="host-name">{{ r.hostName }}</span>
                <span class="host-ip">{{ r.hostIP }}</span>
              </div>
              <div class="result-meta">
                <span>退出码: {{ r.exitCode }}</span>
                <span>耗时: {{ r.duration }}ms</span>
              </div>
            </div>
            <pre class="result-output" v-if="r.output"><code>{{ r.output }}</code></pre>
            <div class="result-error" v-if="r.error">{{ r.error }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { CaretRight, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getHostList } from '@/api/asset.js'
import { getAllCredentials, batchExec, getBatchTask } from '@/api/jumpserver.js'

const allHosts = ref([])
const credentials = ref([])
const selectedHosts = ref([])
const selectedCredentials = ref([])
const command = ref('')
const timeout = ref(30)
const executing = ref(false)
const results = ref([])

onMounted(async () => {
  try {
    const res = await getHostList({ page: 1, pageSize: 200 })
    allHosts.value = res?.data?.data?.list || []
  } catch {}
  try {
    const res = await getAllCredentials()
    credentials.value = res?.data?.data || []
  } catch {}
})

async function handleBatchExec() {
  if (!selectedHosts.value.length) {
    ElMessage.warning('请选择目标主机')
    return
  }
  if (!command.value.trim()) {
    ElMessage.warning('请输入要执行的命令')
    return
  }
  executing.value = true
  results.value = []
  try {
    const res = await batchExec({
      hostIds: selectedHosts.value,
      credentialIds: selectedCredentials.value.length ? selectedCredentials.value : credentials.value.map(c => c.id),
      command: command.value,
      timeout: timeout.value
    })
    const taskId = res?.data?.data?.taskId
    if (taskId) {
      // 轮询获取结果
      const poll = setInterval(async () => {
        try {
          const taskRes = await getBatchTask(taskId)
          const task = taskRes?.data?.data
          if (task) {
            results.value = task.results || []
            if (task.status === 'completed') {
              clearInterval(poll)
              executing.value = false
              ElMessage.success(`批量执行完成，成功 ${task.results.filter(r => r.success).length} 台`)
            }
          }
        } catch {
          clearInterval(poll)
          executing.value = false
        }
      }, 1000)
    }
  } catch (e) {
    ElMessage.error('执行失败: ' + (e?.response?.data?.message || e?.message))
    executing.value = false
  }
}

function clearResults() {
  results.value = []
}
</script>

<style scoped>
.batch-exec-page {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  color: var(--ds-text-primary);
}
.page-header { display: flex; align-items: center; justify-content: space-between; }
.eyebrow { font-size: 11px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; color: var(--ds-text-muted); }
.page-header h1 { margin: 2px 0 0; font-size: 20px; }
.batch-layout { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
.batch-form { background: var(--ds-bg-surface); border: 1px solid var(--ds-border-subtle); border-radius: 8px; padding: 24px; }
.command-input :deep(textarea) { font-family: monospace; }
.unit-hint { margin-left: 8px; font-size: 12px; color: var(--ds-text-tertiary); }
.batch-results { background: var(--ds-bg-surface); border: 1px solid var(--ds-border-subtle); border-radius: 8px; overflow: hidden; }
.results-header {
  display: flex; justify-content: space-between;
  padding: 12px 16px; border-bottom: 1px solid var(--ds-border-subtle);
  font-weight: 650;
}
.result-summary { font-size: 12px; color: var(--ds-text-secondary); }
.results-list { max-height: 600px; overflow-y: auto; }
.result-item { padding: 12px 16px; border-bottom: 1px solid var(--ds-border-subtle); }
.result-item.success { border-left: 3px solid var(--ds-success); }
.result-item.failure { border-left: 3px solid var(--ds-error); }
.result-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.result-host { display: flex; align-items: center; gap: 8px; }
.result-meta { display: flex; gap: 16px; font-size: 12px; color: var(--ds-text-tertiary); }
.result-output { margin: 0; padding: 8px; background: #1e1e2d; color: #f8f8f2; border-radius: 4px; font-size: 12px; overflow-x: auto; }
.result-error { color: var(--ds-error); font-size: 12px; margin-top: 4px; }
@media (max-width: 1200px) { .batch-layout { grid-template-columns: 1fr; } }
</style>