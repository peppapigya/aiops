<template>
  <div class="web-terminal-page">
    <!-- 左侧资产树 -->
    <aside class="terminal-sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <span v-if="!sidebarCollapsed" class="sidebar-title">资产列表</span>
        <el-button link :icon="sidebarCollapsed ? Expand : Fold" @click="sidebarCollapsed = !sidebarCollapsed" />
      </div>
      <div v-if="!sidebarCollapsed" class="sidebar-body">
        <el-input v-model="assetSearch" placeholder="搜索主机..." clearable size="small" class="sidebar-search">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-scrollbar class="asset-tree-scroll">
          <div v-for="group in filteredAssetTree" :key="'g_' + group.id" class="asset-group">
            <div class="group-header" @click="toggleGroup(group.id)">
              <el-icon><component :is="expandedGroups[group.id] ? ArrowDown : ArrowRight" /></el-icon>
              <el-icon><Folder /></el-icon>
              <span class="group-name">{{ group.name }}</span>
              <span class="group-count">{{ group.hostCount || 0 }}</span>
            </div>
            <div v-if="expandedGroups[group.id] && group.children" class="group-children">
              <div v-for="child in group.children" :key="'cg_' + child.id" class="asset-group">
                <div class="group-header sub" @click="toggleGroup(child.id)">
                  <el-icon><component :is="expandedGroups[child.id] ? ArrowDown : ArrowRight" /></el-icon>
                  <el-icon><Folder /></el-icon>
                  <span class="group-name">{{ child.name }}</span>
                  <span class="group-count">{{ child.hostCount || 0 }}</span>
                </div>
                <div v-if="expandedGroups[child.id]" class="group-children">
                  <div v-for="host in getHostsByGroup(child.id)" :key="'h_' + host.id" class="host-item"
                    @contextmenu.prevent="showHostContextMenu($event, host)">
                    <div class="host-info" @dblclick="connectHost(host)">
                      <el-icon><Monitor /></el-icon>
                      <div class="host-detail">
                        <span class="host-name">{{ host.name }}</span>
                        <span class="host-ip">{{ host.ip }}:{{ host.port || 22 }}</span>
                      </div>
                      <span class="host-status" :class="host.status"></span>
                    </div>
                    <div class="host-actions">
                      <el-button link size="small" @click="connectHost(host)">SSH</el-button>
                      <el-button link size="small" @click="connectSFTP(host)">SFTP</el-button>
                    </div>
                  </div>
                </div>
              </div>
              <div v-for="host in getHostsByGroup(group.id)" :key="'h_' + host.id" class="host-item"
                @contextmenu.prevent="showHostContextMenu($event, host)">
                <div class="host-info" @dblclick="connectHost(host)">
                  <el-icon><Monitor /></el-icon>
                  <div class="host-detail">
                    <span class="host-name">{{ host.name }}</span>
                    <span class="host-ip">{{ host.ip }}:{{ host.port || 22 }}</span>
                  </div>
                  <span class="host-status" :class="host.status"></span>
                </div>
                <div class="host-actions">
                  <el-button link size="small" @click="connectHost(host)">SSH</el-button>
                  <el-button link size="small" @click="connectSFTP(host)">SFTP</el-button>
                </div>
              </div>
            </div>
          </div>
        </el-scrollbar>
      </div>
    </aside>

    <!-- 主终端区域 -->
    <main class="terminal-main">
      <!-- Tab 栏 -->
      <div class="tab-bar">
        <div class="tab-list">
          <div v-for="tab in store.terminalTabs" :key="tab.id" class="tab-item"
            :class="{ active: store.activeTabId === tab.id, [tab.status]: true }"
            @click="store.setActiveTab(tab.id)" @contextmenu.prevent="showTabContextMenu($event, tab)">
            <span class="tab-status-dot" :class="tab.status"></span>
            <span class="tab-label">{{ tab.hostName || tab.hostIP }}</span>
            <el-button link class="tab-close" @click.stop="closeTab(tab.id)"><el-icon><Close /></el-icon></el-button>
          </div>
        </div>
        <div class="tab-actions">
          <el-button link :icon="Plus" @click="showNewConnectionDialog">新建连接</el-button>
        </div>
      </div>

      <!-- 终端工具栏 -->
      <div class="terminal-toolbar" v-if="store.activeTab">
        <div class="toolbar-left">
          <el-tooltip content="复制 (Ctrl+Shift+C)" placement="bottom">
            <el-button link size="small" @click="handleCopy"><el-icon><CopyDocument /></el-icon></el-button>
          </el-tooltip>
          <el-tooltip content="粘贴 (Ctrl+Shift+V)" placement="bottom">
            <el-button link size="small" @click="handlePaste"><el-icon><Document /></el-icon></el-button>
          </el-tooltip>
          <el-divider direction="vertical" />
          <el-tooltip content="放大字体" placement="bottom">
            <el-button link size="small" @click="zoomIn"><el-icon><ZoomIn /></el-icon></el-button>
          </el-tooltip>
          <el-tooltip content="缩小字体" placement="bottom">
            <el-button link size="small" @click="zoomOut"><el-icon><ZoomOut /></el-icon></el-button>
          </el-tooltip>
          <el-tooltip content="重置字体" placement="bottom">
            <el-button link size="small" @click="resetZoom"><el-icon><ScaleToOriginal /></el-icon></el-button>
          </el-tooltip>
          <span class="font-size-badge">{{ fontSize }}px</span>
          <el-divider direction="vertical" />
          <el-tooltip content="查找 (Ctrl+Shift+F)" placement="bottom">
            <el-button link size="small" @click="toggleSearch"><el-icon><Search /></el-icon></el-button>
          </el-tooltip>
        </div>
        <div class="toolbar-right">
          <el-tooltip content="清屏" placement="bottom">
            <el-button link size="small" @click="handleClear"><el-icon><Delete /></el-icon></el-button>
          </el-tooltip>
          <el-tooltip content="重连" placement="bottom">
            <el-button link size="small" @click="reconnectTab(store.activeTab)"><el-icon><Refresh /></el-icon></el-button>
          </el-tooltip>
          <el-tooltip content="断开" placement="bottom">
            <el-button link size="small" @click="closeTab(store.activeTabId)"><el-icon><SwitchButton /></el-icon></el-button>
          </el-tooltip>
        </div>
      </div>

      <!-- 搜索栏 -->
      <div class="search-bar" v-if="searchVisible">
        <el-input v-model="searchKeyword" placeholder="搜索终端内容..." size="small" class="search-input"
          @keyup.enter="searchNext" @keyup.esc="searchVisible = false" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
          <template #suffix>
            <span class="search-count" v-if="searchKeyword">{{ searchIndex + 1 }}/{{ searchTotal }}</span>
            <el-button link size="small" @click="searchPrev" :disabled="searchTotal === 0"><el-icon><ArrowUp /></el-icon></el-button>
            <el-button link size="small" @click="searchNext" :disabled="searchTotal === 0"><el-icon><ArrowDown /></el-icon></el-button>
            <el-button link size="small" @click="searchVisible = false"><el-icon><Close /></el-icon></el-button>
          </template>
        </el-input>
      </div>

      <!-- 终端画布 -->
      <div class="terminal-canvas" v-if="store.activeTab">
        <div v-for="tab in store.terminalTabs" :key="'term_' + tab.id" class="terminal-pane"
          :class="{ active: store.activeTabId === tab.id }">
          <div :ref="el => setTerminalRef(tab.id, el)" class="xterm-container" :data-tab-id="tab.id"></div>
          <div v-if="tab.status === 'connecting'" class="terminal-overlay">
            <div class="connecting-spinner">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>正在连接 {{ tab.hostIP }}...</span>
            </div>
          </div>
          <div v-if="tab.status === 'disconnected'" class="terminal-overlay">
            <div class="disconnected-msg">
              <span>连接已断开</span>
              <el-button type="primary" size="small" @click="reconnectTab(tab)">重新连接</el-button>
            </div>
          </div>
        </div>

        <!-- 状态栏 -->
        <div class="terminal-statusbar">
          <span>{{ store.activeTab?.hostName }}@{{ store.activeTab?.hostIP }}</span>
          <span class="statusbar-right">
            <span :class="`status-${store.activeTab?.status}`">{{ statusText[store.activeTab?.status] || '' }}</span>
            <span v-if="store.activeTab?.startTime">{{ formatDuration(store.activeTab?.startTime) }}</span>
          </span>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else class="terminal-empty">
        <div class="empty-content">
          <el-icon :size="48"><Monitor /></el-icon>
          <h3>欢迎使用 Web 终端</h3>
          <p>从左侧资产列表双击主机或点击 SSH 按钮开始连接</p>
          <div class="keyboard-shortcuts">
            <p class="shortcut-title">快捷键</p>
            <div class="shortcut-row"><kbd>Ctrl+Shift+C</kbd> 复制</div>
            <div class="shortcut-row"><kbd>Ctrl+Shift+V</kbd> 粘贴</div>
            <div class="shortcut-row"><kbd>Ctrl+Shift+F</kbd> 搜索</div>
            <div class="shortcut-row"><kbd>Ctrl+滚轮</kbd> 缩放字体</div>
          </div>
          <div class="quick-actions">
            <el-button type="primary" @click="showNewConnectionDialog">新建连接</el-button>
            <el-button @click="goToHostManagement">主机管理</el-button>
          </div>
        </div>
      </div>
    </main>

    <!-- 右键菜单 - 主机 -->
    <div v-show="hostCtxMenu.visible" class="ctx-menu"
      :style="{ left: hostCtxMenu.x + 'px', top: hostCtxMenu.y + 'px' }" @mouseleave="hostCtxMenu.visible = false">
      <div class="ctx-item" @click="connectHost(hostCtxMenu.host); hostCtxMenu.visible = false">
        <el-icon><Connection /></el-icon> SSH 连接
      </div>
      <div class="ctx-item" @click="connectSFTP(hostCtxMenu.host); hostCtxMenu.visible = false">
        <el-icon><FolderOpened /></el-icon> SFTP 文件管理
      </div>
      <div class="ctx-divider" />
      <div class="ctx-item" @click="hostCtxMenu.visible = false">
        <el-icon><InfoFilled /></el-icon> 主机详情
      </div>
    </div>

    <!-- 右键菜单 - Tab -->
    <div v-show="tabCtxMenu.visible" class="ctx-menu"
      :style="{ left: tabCtxMenu.x + 'px', top: tabCtxMenu.y + 'px' }" @mouseleave="tabCtxMenu.visible = false">
      <div class="ctx-item" @click="closeTab(tabCtxMenu.tab?.id); tabCtxMenu.visible = false">
        <el-icon><Close /></el-icon> 关闭
      </div>
      <div class="ctx-item" @click="closeOtherTabs(tabCtxMenu.tab?.id); tabCtxMenu.visible = false">
        <el-icon><SemiSelect /></el-icon> 关闭其他
      </div>
      <div class="ctx-item" @click="closeAllTabs(); tabCtxMenu.visible = false">
        <el-icon><CircleClose /></el-icon> 关闭全部
      </div>
      <div class="ctx-divider" />
      <div class="ctx-item" @click="reconnectTab(tabCtxMenu.tab); tabCtxMenu.visible = false">
        <el-icon><Refresh /></el-icon> 重新连接
      </div>
      <div class="ctx-item" @click="handleCopy(); tabCtxMenu.visible = false">
        <el-icon><CopyDocument /></el-icon> 复制选中内容
      </div>
      <div class="ctx-item" @click="handlePaste(); tabCtxMenu.visible = false">
        <el-icon><Document /></el-icon> 粘贴
      </div>
    </div>

    <!-- 新建连接弹窗 -->
    <el-dialog v-model="newConnDialogVisible" title="新建 SSH 连接" width="520px" destroy-on-close>
      <el-form :model="newConnForm" label-width="90px">
        <el-form-item label="目标主机">
          <el-select v-model="newConnForm.hostId" filterable placeholder="选择主机" style="width:100%">
            <el-option v-for="host in allHosts" :key="host.id"
              :label="`${host.name} (${host.ip}:${host.port || 22})`" :value="host.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="认证凭证">
          <el-select v-model="newConnForm.credentialId" placeholder="选择凭证" style="width:100%">
            <el-option v-for="cred in availableCredentials" :key="cred.id"
              :label="`${cred.name} (${cred.username})`" :value="cred.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="连接方式">
          <el-radio-group v-model="newConnForm.connType">
            <el-radio label="ssh">SSH 终端</el-radio>
            <el-radio label="sftp">SFTP 文件</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="newConnDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleNewConnection">连接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search, Plus, Close, Folder, Monitor, Expand, Fold, ArrowDown, ArrowRight, ArrowUp,
  Connection, FolderOpened, InfoFilled, SemiSelect, CircleClose, Refresh, Loading,
  CopyDocument, Document, ZoomIn, ZoomOut, ScaleToOriginal, Delete, SwitchButton
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { SearchAddon } from 'xterm-addon-search'
import 'xterm/css/xterm.css'
import { useJumpserverStore } from '@/stores/jumpserverStore'
import { getHostGroups, getHostList } from '@/api/asset.js'
import { getAllCredentials, connectHost as apiConnectHost, disconnectSession } from '@/api/jumpserver.js'

const router = useRouter()
const store = useJumpserverStore()

// 资产数据
const assetTree = ref([])
const allHosts = ref([])
const groupHostMap = ref({})
const expandedGroups = reactive({})
const assetSearch = ref('')
const sidebarCollapsed = ref(false)
const availableCredentials = ref([])

// 右键菜单
const hostCtxMenu = reactive({ visible: false, x: 0, y: 0, host: null })
const tabCtxMenu = reactive({ visible: false, x: 0, y: 0, tab: null })

// 新建连接
const newConnDialogVisible = ref(false)
const newConnForm = reactive({ hostId: null, credentialId: null, connType: 'ssh' })

// 终端实例
const terminalInstances = ref({}) // tabId -> { terminal, ws, fitAddon, searchAddon }
const terminalRefs = ref({})
const fontSize = ref(14)

// 搜索
const searchVisible = ref(false)
const searchKeyword = ref('')
const searchIndex = ref(0)
const searchTotal = ref(0)

const statusText = { connecting: '连接中...', connected: '已连接', disconnected: '已断开' }

// 终端主题
const terminalTheme = {
  background: '#1e1e2d',
  foreground: '#f8f8f2',
  cursor: '#00ff00',
  selection: '#44475a',
  black: '#000000', red: '#ff5555', green: '#50fa7b', yellow: '#f1fa8c',
  blue: '#bd93f9', magenta: '#ff79c6', cyan: '#8be9fd', white: '#bbbbbb',
  brightBlack: '#6272a4', brightRed: '#ff6e6e', brightGreen: '#69ff94',
  brightYellow: '#ffffa5', brightBlue: '#d6acff', brightMagenta: '#ff92df',
  brightCyan: '#a4ffff', brightWhite: '#ffffff'
}

const filteredAssetTree = computed(() => {
  if (!assetSearch.value) return assetTree.value
  const keyword = assetSearch.value.toLowerCase()
  const filter = (list) => list.reduce((acc, g) => {
    const children = filter(g.children || [])
    if (g.name.toLowerCase().includes(keyword) || children.length) acc.push({ ...g, children })
    return acc
  }, [])
  return filter(assetTree.value)
})

onMounted(async () => {
  await fetchAssets()
  await fetchCredentials()
  document.addEventListener('keydown', handleGlobalKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleGlobalKeydown)
  Object.values(terminalInstances.value).forEach(inst => {
    if (inst.ws) inst.ws.close()
    if (inst.terminal) inst.terminal.dispose()
  })
})

async function fetchAssets() {
  try {
    const res = await getHostGroups()
    const groups = res?.data?.data || []
    assetTree.value = groups
    expandAllGroups(groups)
    const hostRes = await getHostList({ page: 1, pageSize: 200 })
    const hosts = hostRes?.data?.data?.list || []
    allHosts.value = hosts
    const map = {}
    hosts.forEach(h => { const gid = h.groupId || 0; if (!map[gid]) map[gid] = []; map[gid].push(h) })
    groupHostMap.value = map
  } catch (e) { console.error('获取资产失败:', e) }
}

async function fetchCredentials() {
  try {
    const res = await getAllCredentials()
    availableCredentials.value = res?.data?.data || []
  } catch { }
}

function expandAllGroups(groups) {
  groups.forEach(g => {
    expandedGroups[g.id] = true
    if (g.children) expandAllGroups(g.children)
  })
}

function toggleGroup(id) { expandedGroups[id] = !expandedGroups[id] }
function getHostsByGroup(groupId) { return groupHostMap.value[groupId] || [] }

function showHostContextMenu(e, host) {
  Object.assign(hostCtxMenu, { visible: true, x: e.clientX, y: e.clientY, host })
}

function showTabContextMenu(e, tab) {
  Object.assign(tabCtxMenu, { visible: true, x: e.clientX, y: e.clientY, tab })
}

async function connectHost(host) {
  const creds = availableCredentials.value
  if (!creds.length) { ElMessage.warning('没有可用的认证凭证，请先在凭证管理中创建'); return }
  const credId = host.credentialId || creds[0].id
  const tab = store.addTab({ ...host, credentialId: credId })
  await nextTick()
  initTerminal(tab)
}

function connectSFTP(host) { ElMessage.info('SFTP 文件管理功能即将上线') }

// 创建终端实例
function createTerminal() {
  const terminal = new Terminal({
    cursorBlink: true,
    fontSize: fontSize.value,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: terminalTheme,
    allowTransparency: true,
    scrollback: 10000,
    tabStopWidth: 4,
    rightClickSelectsWord: true,
    convertEol: true,
    disableStdin: false
  })

  const fitAddon = new FitAddon()
  const searchAddon = new SearchAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())
  terminal.loadAddon(searchAddon)

  return { terminal, fitAddon, searchAddon }
}

// 初始化终端连接
async function initTerminal(tab) {
  await nextTick()
  const container = terminalRefs.value[tab.id]
  if (!container) return

  const { terminal, fitAddon, searchAddon } = createTerminal()
  terminal.open(container)

  // 自定义右键：自动复制选中内容
  terminal.element.addEventListener('contextmenu', (e) => {
    e.preventDefault()
    const selection = terminal.getSelection()
    if (selection) {
      navigator.clipboard.writeText(selection).then(() => {
        ElMessage({ message: '已复制到剪贴板', type: 'success', duration: 1500, showClose: false })
      }).catch(() => {})
      terminal.clearSelection()
    }
  })

  // 等待容器有实际尺寸后再 fit（使用 ResizeObserver 确保容器已渲染完毕）
  await new Promise(resolve => {
    if (container.offsetHeight > 0) {
      resolve()
      return
    }
    const observer = new ResizeObserver(() => {
      if (container.offsetHeight > 0) {
        observer.disconnect()
        resolve()
      }
    })
    observer.observe(container)
    // 超时兜底
    setTimeout(() => { observer.disconnect(); resolve() }, 2000)
  })
  await nextTick()
  fitAddon.fit()

  try {
    const connectRes = await apiConnectHost({
      hostId: tab.hostId,
      credentialId: tab.credentialId,
      width: terminal.cols,
      height: terminal.rows
    })
    const data = connectRes?.data?.data
    if (!data?.sessionId) throw new Error(data?.message || '连接失败')

    const sessionId = data.sessionId
    store.updateTabSession(tab.id, sessionId, 'connected')

    // 使用相对路径，通过 Vite 代理（开发模式）或 Nginx 代理（生产模式）转发到后端
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${location.host}/ws/jumpserver/terminal?sessionId=${sessionId}`
    const ws = new WebSocket(wsUrl)
    terminalInstances.value[tab.id] = { terminal, ws, fitAddon, searchAddon }

    ws.onopen = () => {
      store.updateTabSession(tab.id, sessionId, 'connected')
      terminal.focus()
      // 累积用户输入，检测到回车时发送完整命令
      let cmdBuffer = ''
      terminal.onData(data => {
        if (ws.readyState === WebSocket.OPEN) {
          // 只累积可打印 ASCII 字符和退格
          for (const ch of data) {
            if (ch === '\r') {
              // 回车：发送完整命令
              const cmd = cmdBuffer.trim()
              cmdBuffer = ''
              if (cmd && cmd.length <= 200) {
                ws.send(JSON.stringify({ type: 'command', data: cmd }))
              }
            } else if (ch === '\x7f' || ch === '\b') {
              // 退格：删除最后一个字符
              cmdBuffer = cmdBuffer.slice(0, -1)
            } else if (ch >= ' ' && ch <= '~') {
              // 可打印字符
              cmdBuffer += ch
            }
            // 其他控制字符（如 \t, \x1b 等）忽略
          }
          ws.send(JSON.stringify({ type: 'stdin', data }))
        }
      })
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'stdout' || msg.type === 'stderr') {
          terminal.write(msg.data)
        } else if (msg.type === 'error') {
          terminal.writeln(`\r\n\x1b[31m错误: ${msg.error}\x1b[0m`)
        }
      } catch { }
    }

    ws.onerror = () => {
      store.updateTabSession(tab.id, sessionId, 'disconnected')
      terminal.writeln('\r\n\x1b[31m✗ 连接错误\x1b[0m')
    }

    ws.onclose = () => {
      store.updateTabSession(tab.id, sessionId, 'disconnected')
      terminal.writeln('\r\n\x1b[33m⚠ 连接已关闭\x1b[0m')
    }
  } catch (e) {
    store.updateTabSession(tab.id, null, 'disconnected')
    const errMsg = e?.response?.data?.message || e.message || '未知错误'
    terminal.writeln(`\r\n\x1b[31m✗ 连接失败: ${errMsg}\x1b[0m`)
    if (errMsg.includes('审批')) {
      terminal.writeln('\x1b[33m💡 请前往「跳板机 → 审批管理」提交审批申请\x1b[0m')
    }
  }

  // 窗口大小自适应
  const resizeObserver = new ResizeObserver(() => {
    fitAddon.fit()
    const inst = terminalInstances.value[tab.id]
    if (inst?.ws && inst.ws.readyState === WebSocket.OPEN) {
      inst.ws.send(JSON.stringify({ type: 'resize', rows: terminal.rows, cols: terminal.cols }))
    }
  })
  resizeObserver.observe(container)
}

function setTerminalRef(tabId, el) { if (el) terminalRefs.value[tabId] = el }

// 获取当前活跃终端
function getActiveTerminal() {
  const tabId = store.activeTabId
  if (!tabId) return null
  return terminalInstances.value[tabId]
}

// 工具栏操作
function handleCopy() {
  const inst = getActiveTerminal()
  if (!inst) return
  const selection = inst.terminal.getSelection()
  if (selection) {
    navigator.clipboard.writeText(selection).then(() => {
      ElMessage({ message: '已复制', type: 'success', duration: 1500, showClose: false })
    }).catch(() => {})
  }
}

function handlePaste() {
  const inst = getActiveTerminal()
  if (!inst) return
  navigator.clipboard.readText().then(text => {
    if (inst.ws && inst.ws.readyState === WebSocket.OPEN) {
      inst.ws.send(JSON.stringify({ type: 'stdin', data: text }))
    } else {
      inst.terminal.write(text)
    }
  }).catch(() => { ElMessage.warning('粘贴失败，请使用 Ctrl+Shift+V') })
}

function zoomIn() {
  const inst = getActiveTerminal()
  if (inst) {
    fontSize.value = Math.min(fontSize.value + 1, 36)
    inst.terminal.options.fontSize = fontSize.value
    inst.fitAddon.fit()
  }
}

function zoomOut() {
  const inst = getActiveTerminal()
  if (inst) {
    fontSize.value = Math.max(fontSize.value - 1, 8)
    inst.terminal.options.fontSize = fontSize.value
    inst.fitAddon.fit()
  }
}

function resetZoom() {
  const inst = getActiveTerminal()
  fontSize.value = 14
  if (inst) {
    inst.terminal.options.fontSize = 14
    inst.fitAddon.fit()
  }
}

function handleClear() {
  const inst = getActiveTerminal()
  if (inst) inst.terminal.clear()
}

function toggleSearch() {
  searchVisible.value = !searchVisible.value
  if (!searchVisible.value) {
    searchKeyword.value = ''
    searchIndex.value = 0
    searchTotal.value = 0
  }
}

function doSearch(direction = 0) {
  const inst = getActiveTerminal()
  if (!inst || !searchKeyword.value) return
  const addon = inst.searchAddon
  if (direction === 0) {
    addon.findNext(searchKeyword.value)
  } else {
    addon.findPrevious(searchKeyword.value)
  }
  // 粗略计数
  searchTotal.value = addon.findNext(searchKeyword.value) ? 1 : 0
}

function searchNext() { doSearch(0) }
function searchPrev() { doSearch(1) }

// 全局快捷键
function handleGlobalKeydown(e) {
  if (!e.ctrlKey && !e.metaKey) return
  const inst = getActiveTerminal()
  if (!inst) return

  if (e.shiftKey) {
    switch (e.key.toUpperCase()) {
      case 'C':
        e.preventDefault()
        handleCopy()
        break
      case 'V':
        e.preventDefault()
        handlePaste()
        break
      case 'F':
        e.preventDefault()
        toggleSearch()
        break
    }
    return
  }

  if (e.key === '=' || e.key === '+') {
    e.preventDefault()
    zoomIn()
  } else if (e.key === '-') {
    e.preventDefault()
    zoomOut()
  } else if (e.key === '0') {
    e.preventDefault()
    resetZoom()
  }
}

// Tab 操作
function closeTab(tabId) {
  const inst = terminalInstances.value[tabId]
  if (inst) {
    if (inst.ws) {
      inst.ws.send(JSON.stringify({ type: 'disconnect' }))
      inst.ws.close()
    }
    inst.terminal?.dispose()
    delete terminalInstances.value[tabId]
  }
  const sessionId = store.getSessionId(tabId)
  if (sessionId) disconnectSession(sessionId).catch(() => { })
  store.removeTab(tabId)
}

function closeOtherTabs(tabId) {
  store.terminalTabs.filter(t => t.id !== tabId).forEach(t => closeTab(t.id))
}

function closeAllTabs() { store.terminalTabs.forEach(t => closeTab(t.id)) }

function reconnectTab(tab) {
  closeTab(tab.id)
  nextTick(() => {
    // 从 allHosts 中找到原始主机对象，确保 id 是主机 ID 而非 tab ID
    const host = allHosts.value.find(h => h.id === tab.hostId)
    if (host) {
      connectHost({ ...host, credentialId: tab.credentialId })
    }
  })
}

function showNewConnectionDialog() {
  newConnForm.hostId = null
  newConnForm.credentialId = availableCredentials.value[0]?.id || null
  newConnForm.connType = 'ssh'
  newConnDialogVisible.value = true
}

function handleNewConnection() {
  if (!newConnForm.hostId) { ElMessage.warning('请选择目标主机'); return }
  const host = allHosts.value.find(h => h.id === newConnForm.hostId)
  if (!host) { ElMessage.error('主机不存在'); return }
  newConnDialogVisible.value = false
  if (newConnForm.connType === 'ssh') {
    connectHost({ ...host, credentialId: newConnForm.credentialId })
  } else {
    connectSFTP(host)
  }
}

function goToHostManagement() { router.push('/asset/hosts') }

function formatDuration(startTime) {
  if (!startTime) return ''
  const diff = Math.floor((Date.now() - new Date(startTime).getTime()) / 1000)
  const h = Math.floor(diff / 3600)
  const m = Math.floor((diff % 3600) / 60)
  const s = diff % 60
  return `${h > 0 ? h + ':' : ''}${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}
</script>

<style scoped>
.web-terminal-page {
  display: flex;
  height: calc(100vh - 120px);
  background: var(--ds-bg-app);
  color: var(--ds-text-primary);
}

/* 侧边栏 */
.terminal-sidebar {
  width: 260px;
  min-width: 260px;
  border-right: 1px solid var(--ds-border-subtle);
  background: var(--ds-bg-surface);
  display: flex;
  flex-direction: column;
  transition: all 0.2s;
}
.terminal-sidebar.collapsed { width: 48px; min-width: 48px; }
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-bottom: 1px solid var(--ds-border-subtle);
}
.sidebar-title { font-size: 13px; font-weight: 700; }
.sidebar-body { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.sidebar-search { padding: 8px; }
.asset-tree-scroll { flex: 1; padding: 0 4px; }
.asset-group { user-select: none; }
.group-header {
  display: flex; align-items: center; gap: 4px; padding: 4px 8px;
  cursor: pointer; border-radius: 4px; font-size: 12px; color: var(--ds-text-secondary);
}
.group-header:hover { background: var(--ds-bg-hover); }
.group-header.sub { padding-left: 24px; }
.group-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.group-count {
  font-size: 11px; color: var(--ds-text-tertiary);
  background: var(--ds-bg-surface-2); padding: 0 6px; border-radius: 10px;
  min-width: 20px; text-align: center;
}
.host-item { padding: 4px 8px 4px 40px; cursor: pointer; border-radius: 4px; }
.host-item:hover { background: var(--ds-bg-hover); }
.host-info { display: flex; align-items: center; gap: 8px; }
.host-detail { flex: 1; display: flex; flex-direction: column; gap: 1px; min-width: 0; }
.host-name { font-size: 12px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.host-ip { font-size: 11px; color: var(--ds-text-tertiary); }
.host-status { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.host-status.online { background: var(--ds-success); }
.host-status.offline { background: var(--ds-error); }
.host-status.unknown { background: var(--ds-text-tertiary); }
.host-actions { display: flex; gap: 4px; margin-top: 2px; opacity: 0; transition: opacity 0.15s; }
.host-item:hover .host-actions { opacity: 1; }

/* 主区域 */
.terminal-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-width: 0; }

/* Tab 栏 */
.tab-bar {
  display: flex; align-items: center; background: var(--ds-bg-surface);
  border-bottom: 1px solid var(--ds-border-subtle); min-height: 36px;
}
.tab-list { display: flex; flex: 1; overflow-x: auto; scrollbar-width: none; }
.tab-list::-webkit-scrollbar { display: none; }
.tab-item {
  display: flex; align-items: center; gap: 6px; padding: 6px 12px;
  cursor: pointer; border-right: 1px solid var(--ds-border-subtle);
  font-size: 12px; white-space: nowrap; min-width: 0; position: relative;
}
.tab-item:hover { background: var(--ds-bg-hover); }
.tab-item.active { background: var(--ds-bg-app); border-bottom: 2px solid var(--ds-primary); }
.tab-status-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.tab-status-dot.connected { background: var(--ds-success); }
.tab-status-dot.connecting { background: var(--ds-warning); animation: pulse 1.5s infinite; }
.tab-status-dot.disconnected { background: var(--ds-text-tertiary); }
.tab-label { max-width: 120px; overflow: hidden; text-overflow: ellipsis; }
.tab-close { opacity: 0; transition: opacity 0.15s; }
.tab-item:hover .tab-close { opacity: 1; }
.tab-actions { padding: 0 8px; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }

/* 工具栏 */
.terminal-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 2px 8px; background: var(--ds-bg-surface);
  border-bottom: 1px solid var(--ds-border-subtle); min-height: 32px;
}
.toolbar-left, .toolbar-right { display: flex; align-items: center; gap: 2px; }
.font-size-badge { font-size: 11px; color: var(--ds-text-tertiary); min-width: 28px; text-align: center; }

/* 搜索栏 */
.search-bar {
  padding: 4px 8px; background: var(--ds-bg-surface);
  border-bottom: 1px solid var(--ds-border-subtle);
}
.search-input { max-width: 360px; }
.search-count { font-size: 11px; color: var(--ds-text-tertiary); margin-right: 4px; }

/* 终端画布 */
.terminal-canvas {
  flex: 1;
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}
/* 终端 pane 填满整个画布，状态栏覆盖在上面 */
.terminal-pane {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  opacity: 0;
  pointer-events: none;
}
.terminal-pane.active {
  opacity: 1;
  pointer-events: all;
}
.xterm-container {
  width: 100%;
  height: 100%;
  background: #1e1e2d;
}
.terminal-overlay {
  position: absolute; inset: 0; display: flex; align-items: center;
  justify-content: center; background: rgba(0, 0, 0, 0.7); z-index: 10;
}
.connecting-spinner { display: flex; flex-direction: column; align-items: center; gap: 12px; color: #fff; font-size: 14px; }
.disconnected-msg { display: flex; flex-direction: column; align-items: center; gap: 12px; color: #fff; font-size: 14px; }

/* 状态栏 - 覆盖在终端底部 */
.terminal-statusbar {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 2px 12px;
  background: var(--ds-bg-surface);
  border-top: 1px solid var(--ds-border-subtle);
  font-size: 11px;
  color: var(--ds-text-tertiary);
  min-height: 24px;
  z-index: 5;
}
.statusbar-right { display: flex; gap: 12px; }
.status-connected { color: var(--ds-success); }
.status-connecting { color: var(--ds-warning); }
.status-disconnected { color: var(--ds-text-tertiary); }

/* 空状态 */
.terminal-empty {
  flex: 1; display: flex; align-items: center; justify-content: center;
}
.empty-content { text-align: center; color: var(--ds-text-secondary); }
.empty-content h3 { margin: 12px 0 4px; }
.empty-content p { font-size: 13px; color: var(--ds-text-tertiary); margin-bottom: 4px; }
.keyboard-shortcuts {
  text-align: left; max-width: 260px; margin: 16px auto; padding: 12px;
  background: var(--ds-bg-surface); border-radius: 8px; border: 1px solid var(--ds-border-subtle);
}
.shortcut-title { font-size: 12px; font-weight: 700; margin-bottom: 8px; color: var(--ds-text-primary); }
.shortcut-row { font-size: 12px; padding: 2px 0; color: var(--ds-text-secondary); }
kbd {
  display: inline-block; padding: 1px 6px; font-size: 11px; font-family: monospace;
  background: var(--ds-bg-surface-2); border: 1px solid var(--ds-border-subtle);
  border-radius: 3px; margin-right: 6px;
}
.quick-actions { margin-top: 16px; display: flex; gap: 8px; justify-content: center; }

/* 右键菜单 */
.ctx-menu {
  position: fixed; z-index: 9999; min-width: 180px;
  background: var(--ds-bg-surface); border: 1px solid var(--ds-border-subtle);
  border-radius: 8px; box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3); padding: 4px;
}
.ctx-item {
  display: flex; align-items: center; gap: 8px; padding: 6px 12px;
  cursor: pointer; border-radius: 4px; font-size: 13px;
}
.ctx-item:hover { background: var(--ds-bg-hover); }
.ctx-divider { height: 1px; background: var(--ds-border-subtle); margin: 4px 0; }
</style>