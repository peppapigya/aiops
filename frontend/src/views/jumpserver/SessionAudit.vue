<template>
  <div class="session-audit-page">
    <header class="page-header">
      <div>
        <div class="eyebrow">会话审计</div>
        <h1>会话记录</h1>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="fetchSessions">刷新</el-button>
      </div>
    </header>

    <section class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">{{ sessionStats.total || 0 }}</div>
        <div class="stat-label">今日会话</div>
      </div>
      <div class="stat-card is-success">
        <div class="stat-value">{{ sessionStats.online || 0 }}</div>
        <div class="stat-label">在线会话</div>
      </div>
      <div class="stat-card is-info">
        <div class="stat-value">{{ formatDuration(sessionStats.avgDuration || 0) }}</div>
        <div class="stat-label">平均时长</div>
      </div>
    </section>

    <section class="filter-bar">
      <el-input v-model="filter.keyword" placeholder="搜索主机/用户/IP" clearable class="filter-input" @change="fetchSessions" />
      <el-select v-model="filter.status" placeholder="状态" clearable class="filter-select" @change="fetchSessions">
        <el-option label="活跃" value="active" />
        <el-option label="已关闭" value="closed" />
        <el-option label="超时" value="timeout" />
      </el-select>
      <el-select v-model="filter.riskLevel" placeholder="风险等级" clearable class="filter-select" @change="fetchSessions">
        <el-option label="低风险" value="low" />
        <el-option label="中风险" value="medium" />
        <el-option label="高风险" value="high" />
        <el-option label="严重" value="critical" />
      </el-select>
      <el-date-picker
        v-model="filter.dateRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        class="filter-date"
        @change="fetchSessions"
      />
      <el-button type="primary" :icon="Search" @click="fetchSessions">搜索</el-button>
    </section>

    <el-table :data="tableData" v-loading="loading" class="audit-table">
      <el-table-column label="会话ID" width="140" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="session-id">{{ row.sessionId?.slice(0, 8) }}...</span>
        </template>
      </el-table-column>
      <el-table-column label="用户" width="100">
        <template #default="{ row }"> {{ row.username }} </template>
      </el-table-column>
      <el-table-column label="目标主机" min-width="180">
        <template #default="{ row }">
          <div class="host-cell">
            <span class="host-name">{{ row.hostName }}</span>
            <span class="host-ip">{{ row.hostIp }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="协议" width="70">
        <template #default="{ row }"> {{ row.protocol || 'ssh' }} </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)" size="small">{{ statusText[row.status] || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="风险等级" width="90">
        <template #default="{ row }">
          <el-tag :type="riskTag(row.riskLevel)" size="small" v-if="row.riskLevel !== 'low'">
            {{ riskText[row.riskLevel] }}
          </el-tag>
          <span v-else class="muted">-</span>
        </template>
      </el-table-column>
      <el-table-column label="命令数" width="80">
        <template #default="{ row }"> {{ row.commandCount || 0 }} </template>
      </el-table-column>
      <el-table-column label="开始时间" width="160">
        <template #default="{ row }"> {{ formatTime(row.startedAt) }} </template>
      </el-table-column>
      <el-table-column label="时长" width="100">
        <template #default="{ row }">
          <span v-if="row.status === 'active'"><el-icon class="is-loading"><Loading /></el-icon> 进行中</span>
          <span v-else>{{ formatDuration(row.duration) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <div class="row-actions">
            <el-button link type="primary" size="small" @click="viewCommands(row)">命令</el-button>
            <el-button link type="success" size="small" @click="playReplay(row)" :disabled="!row.hasReplay">回放</el-button>
            <el-button link size="small" @click="downloadRecording(row)" :disabled="!row.hasReplay">下载</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
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
      @size-change="fetchSessions"
      @current-change="fetchSessions"
    />

    <!-- 命令列表弹窗 -->
    <el-dialog v-model="cmdDialogVisible" title="命令记录" width="800px" destroy-on-close>
      <el-table :data="commandList" v-loading="cmdLoading" max-height="500">
        <el-table-column label="时间偏移" width="120">
          <template #default="{ row }"> {{ formatTimestamp(row.timestamp) }} </template>
        </el-table-column>
        <el-table-column label="命令" min-width="300">
          <template #default="{ row }">
            <code class="command-text">{{ row.command }}</code>
          </template>
        </el-table-column>
        <el-table-column label="风险" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.isRisky" :type="riskTag(row.riskLevel)" size="small">
              {{ riskText[row.riskLevel] }}
            </el-tag>
            <span v-else class="muted">正常</span>
          </template>
        </el-table-column>
        <el-table-column label="规则" width="120">
          <template #default="{ row }"> {{ row.riskRule || '-' }} </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 录像回放弹窗 -->
    <el-dialog v-model="replayDialogVisible" title="会话回放" width="80%" destroy-on-close @closed="stopReplay">
      <div v-if="replaySessionId" class="replay-container">
        <div class="replay-toolbar">
          <el-button size="small" @click="toggleReplayPlay">
            <el-icon><component :is="replayPlaying ? VideoPause : VideoPlay" /></el-icon>
            {{ replayPlaying ? '暂停' : '播放' }}
          </el-button>
          <el-slider v-model="replaySpeed" :min="0.5" :max="5" :step="0.5" show-input style="width: 200px; margin: 0 16px;" />
          <span class="speed-label">速度: {{ replaySpeed }}x</span>
        </div>
        <div ref="replayContainer" class="replay-terminal"></div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { Search, Refresh, Loading, VideoPlay, VideoPause } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import {
  getSessions, getSessionStats, getSessionCommands,
  deleteSession, getRecordingUrl, getRecordingDownloadUrl
} from '@/api/jumpserver.js'

const loading = ref(false)
const tableData = ref([])
const sessionStats = ref({})
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const filter = reactive({ keyword: '', status: '', riskLevel: '', dateRange: null })

const cmdDialogVisible = ref(false)
const cmdLoading = ref(false)
const commandList = ref([])

const replayDialogVisible = ref(false)
const replaySessionId = ref(null)
const replayContainer = ref(null)
const replayPlaying = ref(false)
const replaySpeed = ref(1)
let replayTerminal = null
let replayFitAddon = null
let replayTimer = null
let replayEvents = []
let replayStartTime = 0
let replayCurrentIndex = 0

const statusText = { active: '活跃', closed: '已关闭', timeout: '超时', disconnected: '已断开' }
const riskText = { low: '低', medium: '中', high: '高', critical: '严重' }

function statusTag(s) { return { active: 'success', closed: 'info', timeout: 'warning', disconnected: 'danger' }[s] || 'info' }
function riskTag(r) { return { low: 'info', medium: 'warning', high: 'danger', critical: 'danger' }[r] || 'info' }

onMounted(() => { fetchSessions(); fetchStats() })
onBeforeUnmount(() => { stopReplay() })

async function fetchSessions() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize }
    if (filter.keyword) params.keyword = filter.keyword
    if (filter.status) params.status = filter.status
    if (filter.riskLevel) params.riskLevel = filter.riskLevel
    if (filter.dateRange) {
      params.dateFrom = new Date(filter.dateRange[0]).toISOString()
      params.dateTo = new Date(filter.dateRange[1]).toISOString()
    }
    const res = await getSessions(params)
    const data = res?.data?.data || {}
    tableData.value = data.list || []
    pagination.total = data.total || 0
  } catch { tableData.value = [] }
  finally { loading.value = false }
}

async function fetchStats() {
  try {
    const res = await getSessionStats()
    sessionStats.value = res?.data?.data || {}
  } catch {}
}

async function viewCommands(row) {
  cmdDialogVisible.value = true
  cmdLoading.value = true
  try {
    const res = await getSessionCommands(row.sessionId, { page: 1, pageSize: 200 })
    commandList.value = res?.data?.data?.list || []
  } catch { commandList.value = [] }
  finally { cmdLoading.value = false }
}

// ========== 回放功能（使用 xterm.js 自己实现） ==========

function parseCastFile(text) {
  const lines = text.trim().split('\n')
  if (lines.length === 0) throw new Error('录像文件为空')

  const header = JSON.parse(lines[0])
  const events = []

  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim()
    if (!line || line[0] !== '[') continue
    try {
      const [time, type, data] = JSON.parse(line)
      // 只处理 "o" (output) 事件，忽略 "i" (input) 事件
      // 因为 SSH echo 开启，输入已包含在 output 中
      if (type === 'o') {
        // base64 解码
        const decoded = atob(data)
        events.push({ time: time * 1000, data: decoded })
      }
    } catch { /* 跳过解析失败的行 */ }
  }

  return { header, events }
}

async function playReplay(row) {
  replaySessionId.value = row.sessionId
  replayDialogVisible.value = true
  stopReplay()

  await nextTick()
  if (!replayContainer.value) return

  try {
    // 获取 cast 文件内容
    const token = localStorage.getItem('access_token')
    const url = getRecordingUrl(row.sessionId)
    const response = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!response.ok) throw new Error('获取录像失败')
    const castText = await response.text()

    // 解析 cast 文件
    const { header, events } = parseCastFile(castText)
    if (events.length === 0) {
      ElMessage.warning('录像中没有操作记录')
      replayDialogVisible.value = false
      return
    }
    replayEvents = events
    replayCurrentIndex = 0

    // 创建 xterm 终端
    replayContainer.value.innerHTML = ''
    replayTerminal = new Terminal({
      cursorBlink: false,
      fontSize: 14,
      fontFamily: 'Consolas, "Courier New", monospace',
      theme: {
        background: '#1e1e2d',
        foreground: '#f8f8f2',
        cursor: '#00ff00'
      },
      cols: header.width || 120,
      rows: header.height || 30,
      disableStdin: true,
      scrollback: 10000
    })
    replayFitAddon = new FitAddon()
    replayTerminal.loadAddon(replayFitAddon)
    replayTerminal.open(replayContainer.value)
    await nextTick()
    replayFitAddon.fit()

    replayStartTime = Date.now()
    replayPlaying.value = false
  } catch (e) {
    ElMessage.error('加载录像失败: ' + (e.message || '未知错误'))
    replayDialogVisible.value = false
  }
}

function toggleReplayPlay() {
  if (!replayTerminal || replayEvents.length === 0) return

  replayPlaying.value = !replayPlaying.value
  if (replayPlaying.value) {
    startReplayPlayback()
  } else {
    stopReplayPlayback()
  }
}

function startReplayPlayback() {
  if (replayTimer) return
  // 从当前索引继续播放
  const elapsed = replayEvents[replayCurrentIndex]?.time || 0
  replayStartTime = Date.now() - elapsed / replaySpeed.value

  function tick() {
    if (!replayPlaying.value || !replayTerminal) {
      replayTimer = null
      return
    }

    const now = Date.now()
    const virtualTime = (now - replayStartTime) * replaySpeed.value

    // 播放所有时间 <= virtualTime 的事件
    while (replayCurrentIndex < replayEvents.length && replayEvents[replayCurrentIndex].time <= virtualTime) {
      replayTerminal.write(replayEvents[replayCurrentIndex].data)
      replayCurrentIndex++
    }

    // 播放完毕
    if (replayCurrentIndex >= replayEvents.length) {
      replayPlaying.value = false
      replayTimer = null
      return
    }

    replayTimer = setTimeout(tick, 50)
  }

  tick()
}

function stopReplayPlayback() {
  if (replayTimer) {
    clearTimeout(replayTimer)
    replayTimer = null
  }
}

function stopReplay() {
  stopReplayPlayback()
  replayPlaying.value = false
  if (replayTerminal) {
    replayTerminal.dispose()
    replayTerminal = null
  }
  replayFitAddon = null
  replayEvents = []
  replayCurrentIndex = 0
}

// 速度变化时重置播放基准时间
watch(replaySpeed, () => {
  if (replayPlaying.value && replayTerminal && replayCurrentIndex < replayEvents.length) {
    const elapsed = replayEvents[replayCurrentIndex]?.time || 0
    replayStartTime = Date.now() - elapsed / replaySpeed.value
  }
})

async function downloadRecording(row) {
  try {
    const token = localStorage.getItem('access_token')
    const url = getRecordingDownloadUrl(row.sessionId)
    const response = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!response.ok) throw new Error('下载失败')
    const blob = await response.blob()
    const blobUrl = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = `${row.hostName || 'session'}_${row.sessionId?.slice(0, 8) || 'unknown'}.cast`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(blobUrl)
  } catch (e) {
    ElMessage.error('下载录像失败')
  }
}

async function handleDelete(row) {
  await ElMessageBox.confirm('确认删除该会话记录？', '提示', { type: 'warning' })
  try {
    await deleteSession(row.id)
    ElMessage.success('删除成功')
    fetchSessions()
  } catch { ElMessage.error('删除失败') }
}

function formatTime(t) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

function formatDuration(d) {
  if (!d) return '-'
  const h = Math.floor(d / 3600)
  const m = Math.floor((d % 3600) / 60)
  const s = d % 60
  return `${h > 0 ? h + '时' : ''}${m}分${s}秒`
}

function formatTimestamp(ts) {
  if (!ts) return '-'
  const m = Math.floor(ts / 60)
  const s = Math.floor(ts % 60)
  return `${m > 0 ? m + '分' : ''}${s}秒`
}
</script>

<style scoped>
.session-audit-page {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  color: var(--ds-text-primary);
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.eyebrow {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--ds-text-muted);
}
.page-header h1 { margin: 2px 0 0; font-size: 20px; }
.header-actions { display: flex; gap: 8px; }
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}
.stat-card {
  background: var(--ds-bg-surface);
  border: 1px solid var(--ds-border-subtle);
  border-radius: 8px;
  padding: 16px;
}
.stat-value { font-size: 28px; font-weight: 750; }
.stat-label { font-size: 11px; color: var(--ds-text-tertiary); margin-top: 4px; }
.stat-card.is-success .stat-value { color: var(--ds-success); }
.stat-card.is-info .stat-value { color: var(--ds-info); }
.filter-bar { display: flex; gap: 8px; align-items: center; }
.filter-input { width: 200px; }
.filter-select { width: 120px; }
.filter-date { width: 360px; }
.audit-table { flex: 1; }
.host-cell { display: flex; flex-direction: column; gap: 2px; }
.host-name { font-weight: 650; }
.host-ip { font-size: 12px; color: var(--ds-text-tertiary); }
.session-id { font-family: monospace; font-size: 12px; }
.command-text { font-family: monospace; font-size: 13px; }
.muted { color: var(--ds-text-tertiary); }
.row-actions { display: flex; gap: 4px; }
.pagination { margin-top: 8px; display: flex; justify-content: flex-end; }
.replay-container { background: #1e1e2d; border-radius: 8px; overflow: hidden; }
.replay-toolbar { display: flex; align-items: center; padding: 8px 12px; background: #2a2a3e; }
.speed-label { font-size: 12px; color: #a0a0b0; }
.replay-terminal { height: 500px; padding: 0; }
</style>