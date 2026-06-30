import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useJumpserverStore = defineStore('jumpserver', () => {
  // 终端 Tab 列表
  const terminalTabs = ref([])
  const activeTabId = ref(null)

  // 活跃连接会话ID映射
  const activeSessions = ref({}) // tabId -> sessionId

  const activeTab = computed(() => terminalTabs.value.find(t => t.id === activeTabId.value))

  // 添加终端 Tab
  function addTab(host) {
    const id = `tab_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
    const tab = {
      id,
      hostId: host.id,
      hostName: host.name,
      hostIP: host.ip,
      username: host.username || 'root',
      sessionId: null,
      status: 'connecting', // connecting/connected/disconnected
      credentialId: host.credentialId || null,
      startTime: new Date()
    }
    terminalTabs.value.push(tab)
    activeTabId.value = id
    return tab
  }

  // 移除终端 Tab
  function removeTab(tabId) {
    const idx = terminalTabs.value.findIndex(t => t.id === tabId)
    if (idx === -1) return
    terminalTabs.value.splice(idx, 1)
    delete activeSessions.value[tabId]
    if (activeTabId.value === tabId) {
      activeTabId.value = terminalTabs.value.length > 0 ? terminalTabs.value[terminalTabs.value.length - 1].id : null
    }
  }

  // 设置活跃 Tab
  function setActiveTab(tabId) {
    activeTabId.value = tabId
  }

  // 更新 Tab 会话状态
  function updateTabSession(tabId, sessionId, status) {
    const tab = terminalTabs.value.find(t => t.id === tabId)
    if (tab) {
      tab.sessionId = sessionId
      tab.status = status
    }
    if (sessionId) {
      activeSessions.value[tabId] = sessionId
    }
  }

  // 获取 Tab 的会话 ID
  function getSessionId(tabId) {
    return activeSessions.value[tabId] || null
  }

  return {
    terminalTabs,
    activeTabId,
    activeSessions,
    activeTab,
    addTab,
    removeTab,
    setActiveTab,
    updateTabSession,
    getSessionId
  }
})