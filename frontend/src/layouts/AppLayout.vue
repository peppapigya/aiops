<template>
  <div class="ds-app-shell">
    <aside class="ds-icon-rail">
      <button class="ds-rail-logo" type="button" @click="handleLogoClick" title="DevOps Console">
        <el-icon><Monitor /></el-icon>
      </button>

      <el-scrollbar class="ds-rail-scrollbar">
        <nav class="ds-rail-nav">
          <el-tooltip v-for="item in primaryMenus" :key="item.id" :content="item.name" placement="right" effect="dark">
            <button class="ds-rail-item" :class="{ 'is-active': activePrimaryMenu?.id === item.id }" type="button" @click="selectPrimaryMenu(item)">
              <el-icon><component :is="item.icon || Monitor" /></el-icon>
            </button>
          </el-tooltip>
        </nav>
      </el-scrollbar>

      </aside>

    <aside class="ds-section-sidebar" :class="{ 'is-collapsed': isSectionCollapsed }">
      <header class="ds-section-header">
        <div class="ds-section-title-block">
          <span class="ds-section-kicker">工作区</span>
          <strong class="ds-section-title">{{ sectionTitle }}</strong>
        </div>
        <button class="ds-icon-button" type="button" @click="toggleCollapse">
          <el-icon><Expand v-if="isSectionCollapsed" /><Fold v-else /></el-icon>
        </button>
      </header>

      <div v-if="!isSectionCollapsed" class="ds-section-search">
        <el-input v-model="menuSearchQuery" size="small" clearable placeholder="搜索菜单 / 资源">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
      </div>

      <div v-if="!isSectionCollapsed" class="ds-section-instance">
        <el-select
          v-model="selectedInstance"
          :placeholder="isKafkaRoute ? '选择 Kafka 集群' : '选择实例'"
          size="small"
          filterable
          clearable
          popper-class="instance-select-popper"
          @visible-change="handleDropdownVisible"
          @change="handleInstanceSelect"
        >
          <el-option
            v-for="entry in currentSelectionList"
            :key="entry.id"
            :label="entry.name"
            :value="entry.id"
          >
            <div class="instance-option">
              <el-icon class="type-icon" :color="getEntryColor(entry)"><component :is="getEntryIcon(entry)" /></el-icon>
              <span class="instance-option-name">{{ entry.name }}</span>
              <el-tag :type="getEntryStatusType(entry)" size="small" effect="plain">{{ getEntryStatusLabel(entry) }}</el-tag>
            </div>
          </el-option>
        </el-select>
      </div>

      <el-scrollbar v-if="!isSectionCollapsed" class="ds-section-scrollbar">
        <nav class="ds-section-menu">
          <template v-if="sectionMenus.length > 0">
            <div v-for="item in sectionMenus" :key="item.id" class="ds-menu-block">
              <button
                v-if="!item.children || item.children.length === 0"
                class="ds-menu-item"
                :class="{ 'is-active': isMenuActive(item) }"
                type="button"
                @click="navigateMenu(item)"
                @mouseenter="handleMenuHover(item.path)"
              >
                <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
                <span>{{ item.name }}</span>
              </button>

              <template v-else>
                <div class="ds-menu-group-title">{{ item.name }}</div>
                <button
                  v-for="child in filteredChildren(item.children)"
                  :key="child.id"
                  class="ds-menu-item"
                  :class="{ 'is-active': isMenuActive(child), 'is-parent': child.children && child.children.length > 0 }"
                  type="button"
                  @click="navigateMenu(child)"
                  @mouseenter="handleMenuHover(child.path)"
                >
                  <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
                  <span>{{ child.name }}</span>
                </button>

                <template v-for="child in filteredChildren(item.children)" :key="`${child.id}-children`">
                  <div v-if="child.children && child.children.length > 0" class="ds-sub-menu">
                    <button
                      v-for="leaf in filteredChildren(child.children)"
                      :key="leaf.id"
                      class="ds-menu-item is-leaf"
                      :class="{ 'is-active': isMenuActive(leaf) }"
                      type="button"
                      @click="navigateMenu(leaf)"
                      @mouseenter="handleMenuHover(leaf.path)"
                    >
                      <el-icon v-if="leaf.icon"><component :is="leaf.icon" /></el-icon>
                      <span>{{ leaf.name }}</span>
                    </button>
                  </div>
                </template>
              </template>
            </div>
          </template>
          <div v-else class="ds-menu-empty">暂无可用菜单</div>
        </nav>
      </el-scrollbar>
    </aside>

    <main class="ds-main">
      <header class="ds-topbar">
        <div class="ds-topbar-left">
          <div class="ds-breadcrumb">
            <span>{{ activePrimaryMenu?.name || 'DevOps控制台' }}</span>
            <span class="ds-breadcrumb-separator">/</span>
            <strong>{{ currentRouteTitle }}</strong>
          </div>

                  </div>

        <div class="ds-topbar-actions">
          <button class="ds-command-trigger" type="button">
            <el-icon><Search /></el-icon>
            <span>命令</span>
            <kbd>⌘K</kbd>
          </button>

          <el-tooltip :content="isDark ? '切换亮色主题' : '切换暗色主题'" effect="dark" placement="bottom">
            <button class="ds-icon-button" type="button" @click="toggleTheme">
              <el-icon><Sunny v-if="isDark" /><Moon v-else /></el-icon>
            </button>
          </el-tooltip>

          <el-tooltip content="通知" effect="dark" placement="bottom">
            <el-badge :value="3" :max="99" class="ds-badge-action">
              <button class="ds-icon-button" type="button"><el-icon><Bell /></el-icon></button>
            </el-badge>
          </el-tooltip>

          <el-dropdown @command="handleUserCommand" trigger="click">
            <button class="ds-user-menu" type="button">
              <el-avatar :size="24" :src="permStore.userInfo?.avatar || 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png'" />
              <span>{{ permStore.userInfo?.nickname || permStore.userInfo?.username || '管理员' }}</span>
              <el-icon><CaretBottom /></el-icon>
            </button>
            <template #dropdown>
              <el-dropdown-menu class="user-dropdown">
                <div class="user-dropdown-header">
                  <p class="name">{{ permStore.userInfo?.nickname || permStore.userInfo?.username }}</p>
                  <p class="role">{{ permStore.roles?.join(', ') || '' }}</p>
                </div>
                <el-dropdown-item v-if="permStore.hasPerm('sys:user:profile')" command="profile" divided><el-icon><User /></el-icon>个人中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided style="color: #ef4444;"><el-icon><SwitchButton /></el-icon>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <section class="ds-content">
        <router-view v-slot="{ Component }">
          <component :is="Component" :key="route.path" />
        </router-view>
      </section>
    </main>
  </div>
</template>

<script setup>
import {computed, onMounted, ref, watch} from 'vue'
import {storeToRefs} from 'pinia'
import {useRoute, useRouter} from 'vue-router'
import {ElMessage} from 'element-plus'
import {getInstanceList, getInstanceTypes} from '@/api/instance.js'
import {getSelectedInstanceType, setSelectedInstance} from '@/stores/instanceStore.js'
import {useKafkaStore} from '@/stores/kafkaStore.js'
import {usePermissionStore} from '@/stores/permissionStore.js'
import { useConnectionStore } from '@/mysql/stores/connection'
import { preloadExplorerTree } from '@/mysql/utils/explorer'
import { preloadSecurityOverview } from '@/mysql/utils/security-overview'
import { resolveMenuIcon } from '@/utils/menuIcons.js'
import {
  Bell,
  Box,
  CaretBottom,
  DataLine,
  DocumentCopy,
  Expand,
  Fold,
  List,
  Monitor,
  Moon,
  Search,
  Sunny,
  SwitchButton,
  User,
  WarningFilled
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const permStore = usePermissionStore()
const kafkaStore = useKafkaStore()
const connectionStore = useConnectionStore()
const { clusterOptions, selectedClusterId } = storeToRefs(kafkaStore)

// 状态
const currentTheme = ref('light') // 'dark' | 'light'
const isDark = computed(() => currentTheme.value === 'dark')
const isSectionCollapsed = ref(false)
const selectedInstance = ref('')
const instanceList = ref([])
const instanceTypes = ref([])
const searchQuery = ref('')
const menuSearchQuery = ref('')
const typeFilter = ref('')
const statusFilter = ref('')
const isKafkaRoute = computed(() => route.path === '/kafka' || route.path.startsWith('/kafka/'))


// ======================================================
// 动态菜单：直接使用 permissionStore 中的 menuTree
// menuTree 已由路由守卫加载，这里只需读取
// ======================================================
const mysqlMenuOrder = new Map([
  ['/mysql/workbench', 1],
  ['/mysql/databases', 2],
  ['/mysql/data', 3],
  ['/mysql/query', 4],
  ['/mysql/security', 5],
  ['/mysql/backup', 6]
])

const topLevelMenuOrder = new Map([
  ['首页', 1],
  ['实例管理', 2],
  ['Kubernetes', 3],
  ['混沌实验', 4],
  ['Kafka', 5],
  ['Elasticsearch', 6],
  ['MySQL', 7],
  ['MongoDB', 8],
  ['CI/CD 流水线', 9],
  ['任务调度', 10],
  ['资源管理', 12],
  ['监控中心', 13],
  ['系统管理', 14]
])

function resolveMysqlMenuOrder(item) {
  if (!item) {
    return Number.MAX_SAFE_INTEGER
  }

  if (item.path && mysqlMenuOrder.has(item.path)) {
    return mysqlMenuOrder.get(item.path)
  }

  if (!item.children || item.children.length === 0) {
    return Number.MAX_SAFE_INTEGER
  }

  return item.children.reduce((lowest, child) => Math.min(lowest, resolveMysqlMenuOrder(child)), Number.MAX_SAFE_INTEGER)
}

function sortMysqlMenus(items) {
  return [...(items || [])]
    .map((item) => ({
      ...item,
      children: item.children && item.children.length > 0 ? sortMysqlMenus(item.children) : item.children
    }))
    .sort((left, right) => {
      const leftOrder = resolveMysqlMenuOrder(left)
      const rightOrder = resolveMysqlMenuOrder(right)
      if (leftOrder !== rightOrder) {
        return leftOrder - rightOrder
      }
      return 0
    })
}

function resolveTopLevelMenuOrder(item) {
  if (!item) {
    return Number.MAX_SAFE_INTEGER
  }

  if (item.name && topLevelMenuOrder.has(item.name)) {
    return topLevelMenuOrder.get(item.name)
  }

  return Number.MAX_SAFE_INTEGER
}

function sortTopLevelMenus(items) {
  return [...(items || [])].sort((left, right) => {
    const leftOrder = resolveTopLevelMenuOrder(left)
    const rightOrder = resolveTopLevelMenuOrder(right)

    if (leftOrder !== rightOrder) {
      return leftOrder - rightOrder
    }

    return 0
  })
}


function normalizeMenuIcon(icon) {
  if (!icon) {
    return null
  }

  if (typeof icon !== 'string') {
    return icon
  }

  return resolveMenuIcon(icon) || icon
}

function normalizeMenuTree(items) {
  return sortTopLevelMenus((items || []).map((item) => {
    const nextItem = {
      ...item,
      icon: normalizeMenuIcon(item.icon),
      children: item.children && item.children.length > 0 ? normalizeMenuTree(item.children) : item.children
    }

    if (item.path === '/mysql' || item.name === 'MySQL') {
      nextItem.children = sortMysqlMenus(nextItem.children)
    }

    return nextItem
  }))
}

const sidebarMenus = computed(() => normalizeMenuTree(permStore.menuTree))
const primaryMenus = computed(() => sidebarMenus.value)

function hasPathInMenu(item, currentPath) {
  if (!item) return false
  if (item.path && (currentPath === item.path || (item.path !== '/' && currentPath.startsWith(`${item.path}/`)))) {
    return true
  }
  return (item.children || []).some((child) => hasPathInMenu(child, currentPath))
}

const activePrimaryMenu = computed(() => {
  const menus = primaryMenus.value
  return menus.find((item) => hasPathInMenu(item, route.path)) || menus[0] || null
})

const sectionTitle = computed(() => activePrimaryMenu.value?.name || 'DevOps Console')
const currentRouteTitle = computed(() => route.meta?.title || activePrimaryMenu.value?.name || 'Overview')
const sectionMenus = computed(() => {
  const menu = activePrimaryMenu.value
  if (!menu) return []
  const children = menu.children || []
  return children.length > 0 ? filterMenusByQuery(children) : [menu]
})

function menuMatchesQuery(item, query) {
  if (!query) return true
  const lower = query.toLowerCase()
  return [item.name, item.path].filter(Boolean).some((value) => String(value).toLowerCase().includes(lower))
}

function filterMenusByQuery(items) {
  const query = menuSearchQuery.value.trim()
  if (!query) return items || []
  return (items || [])
    .map((item) => ({
      ...item,
      children: item.children ? filterMenusByQuery(item.children) : item.children
    }))
    .filter((item) => menuMatchesQuery(item, query) || (item.children && item.children.length > 0))
}

function filteredChildren(children) {
  return filterMenusByQuery(children)
}

function findFirstNavigable(item) {
  if (!item) return null
  // 如果有子菜单，优先找第一个可导航的子菜单
  if (item.children && item.children.length > 0) {
    for (const child of item.children) {
      const found = findFirstNavigable(child)
      if (found) return found
    }
    return null
  }
  // 没有子菜单且自己有 path，直接返回
  if (item.path) return item
  return null
}

function selectPrimaryMenu(item) {
  const target = findFirstNavigable(item)
  if (target?.path) {
    navigateMenu(target)
  }
}

function navigateMenu(item) {
  const target = findFirstNavigable(item)
  if (target?.path) {
    router.push(target.path)
  }
}

function isMenuActive(item) {
  return hasPathInMenu(item, route.path)
}

// 保留，兼容其他地方的引用
const allMenuRoutes = computed(() => [])


// 计算选中实例名称
const selectedInstanceName = computed(() => {
  if (isKafkaRoute.value) {
    return clusterOptions.value.find((item) => item.id === selectedClusterId.value)?.name || ''
  }
  if (!selectedInstance.value) return ''
  const instance = instanceList.value.find(item => item.id === selectedInstance.value)
  return instance ? instance.name : ''
})

const selectorLabel = computed(() => (isKafkaRoute.value ? '当前 Kafka 集群' : '当前实例'))
const selectorPlaceholder = computed(() => (isKafkaRoute.value ? '选择 Kafka 集群' : '选择实例'))
const selectorPanelTitle = computed(() => (isKafkaRoute.value ? 'Kafka 集群' : '实例管理'))
const currentSelectorType = computed(() => (isKafkaRoute.value ? 'kafka' : getSelectedInstanceType()))
const currentSelectionId = computed(() => (isKafkaRoute.value ? selectedClusterId.value : selectedInstance.value))

const filteredKafkaClusterList = computed(() => {
  let result = [...clusterOptions.value]
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter((cluster) => cluster.name.toLowerCase().includes(query))
  }
  if (statusFilter.value) {
    result = result.filter((cluster) => cluster.status === statusFilter.value)
  }
  return result
})

const currentSelectionList = computed(() => (isKafkaRoute.value ? filteredKafkaClusterList.value : filteredInstanceList.value))
const prefetchedRouteComponents = new Set()
const prefetchedRouteData = new Set()
const getInstanceType = (instance) => instance.instance_type || instance.instanceType || instance.type_name || instance.type || ''

// 过滤实例列表
const filteredInstanceList = computed(() => {
  let result = [...instanceList.value]
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(instance =>
      instance.name.toLowerCase().includes(query) ||
      instance.address.toLowerCase().includes(query)
    )
  }
  if (typeFilter.value) {
    result = result.filter(instance => getInstanceType(instance) === typeFilter.value)
  }
  if (statusFilter.value) {
    result = result.filter(instance => instance.status === statusFilter.value)
  }
  return result
})

// === 方法 ===

const toggleCollapse = () => {
  isSectionCollapsed.value = !isSectionCollapsed.value
}

const toggleTheme = () => {
  const next = currentTheme.value === 'dark' ? 'light' : 'dark'
  applyTheme(next)
}

function applyTheme(theme) {
  currentTheme.value = theme
  const html = document.documentElement
  html.classList.toggle('dark', theme === 'dark')
  html.classList.toggle('light', theme === 'light')
  localStorage.setItem('theme', theme)
}

const handleLogoClick = () => {
  router.push('/')
}

const handleUserCommand = (command) => {
  switch (command) {
    case 'profile':
      router.push('/system/profile')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      permStore.reset()
      ElMessage.success('已退出登录')
      // 跳转登录页并强制刷新，彻底清除动态路由和内存状态
      window.location.href = '/login'
      break
  }
}

// 实例管理相关
const selectInstance = (instanceId) => {
  if (isKafkaRoute.value) {
    kafkaStore.setSelectedClusterId(instanceId)
    const cluster = clusterOptions.value.find((item) => item.id === instanceId)
    if (cluster) {
      ElMessage.success(`已切换到 Kafka 集群: ${cluster.name}`)
    }
    return
  }
  selectedInstance.value = instanceId
  const instance = instanceList.value.find(item => item.id === instanceId)
  if (instance) {
    setSelectedInstance(instance)
    ElMessage.success(`已切换到: ${instance.name}`)
    // 切换实例后保持在当前页面，不强制跳转
  }
}

const handleAddInstance = () => {
  if (isKafkaRoute.value) {
    router.push('/kafka/clusters')
    return
  }
  router.push('/es/instances/add')
}

const handleRefreshInstances = async () => {
  if (isKafkaRoute.value) {
    await kafkaStore.loadClusterOptions({ force: true })
    ElMessage.success('Kafka 集群已刷新')
    return
  }
  await Promise.all([fetchInstances(), fetchInstanceTypes()])
  ElMessage.success('数据已刷新')
}

const handleDropdownVisible = (visible) => {
  if (visible) {
    if (isKafkaRoute.value) {
      kafkaStore.loadClusterOptions({ force: true })
      return
    }
    Promise.all([fetchInstances(), fetchInstanceTypes()])
  }
}

const handleInstanceSelect = (instanceId) => {
  if (!instanceId) return
  if (isKafkaRoute.value) {
    kafkaStore.setSelectedClusterId(instanceId)
    const cluster = clusterOptions.value.find((item) => item.id === instanceId)
    if (cluster) {
      ElMessage.success(`已切换到 Kafka 集群: ${cluster.name}`)
    }
    return
  }
  const instance = instanceList.value.find(item => item.id === instanceId)
  if (instance) {
    setSelectedInstance(instance)
    ElMessage.success(`已切换到: ${instance.name}`)
  }
}

const mysqlRouteComponentPreloaders = {
  '/mysql/workbench': () => import('../views/mysql/ConnectionManagement.vue'),
  '/mysql/databases': () => import('../views/mysql/DatabaseManagement.vue'),
  '/mysql/data': () => import('../views/mysql/DataManagement.vue'),
  '/mysql/query': () => import('../views/mysql/SqlQuery.vue'),
  '/mysql/security': () => import('../views/mysql/UserPermissionManagement.vue'),
  '/mysql/backup': () => import('../views/mysql/BackupManagement.vue')
}

const mysqlRouteDataPreloaders = {
  '/mysql/workbench': () => preloadExplorerTree(),
  '/mysql/databases': () => preloadExplorerTree(),
  '/mysql/security': () => preloadSecurityOverview()
}

function prefetchMenuRoute(path) {
  const preloadComponent = mysqlRouteComponentPreloaders[path]
  if (preloadComponent && !prefetchedRouteComponents.has(path)) {
    prefetchedRouteComponents.add(path)
    void preloadComponent().catch(() => {
      prefetchedRouteComponents.delete(path)
    })
  }

  const preloadData = mysqlRouteDataPreloaders[path]
  if (preloadData && connectionStore.hasConnection && !prefetchedRouteData.has(path)) {
    prefetchedRouteData.add(path)
    void preloadData().catch(() => {
      prefetchedRouteData.delete(path)
    })
  }
}

function queueMysqlRoutePrefetch() {
  const task = () => {
    Object.keys(mysqlRouteComponentPreloaders).forEach((path) => {
      prefetchMenuRoute(path)
    })
  }

  if (typeof window !== 'undefined' && 'requestIdleCallback' in window) {
    window.requestIdleCallback(task, { timeout: 1200 })
    return
  }

  window.setTimeout(task, 200)
}

function handleMenuHover(path) {
  if (typeof path === 'string' && path.startsWith('/mysql/')) {
    prefetchMenuRoute(path)
  }
}

// Data fetching
const fetchInstances = async () => {
  try {
    const response = await getInstanceList({ page: 1, page_size: 100 })
    const listData = response.data?.list || {}
    instanceList.value = (listData.data || []).map(instance => ({
      ...instance,
      instance_type: getInstanceType(instance)
    }))

    if (instanceList.value.length > 0 && !selectedInstance.value) {
      // 优先选择 elasticsearch 或 kubernetes，避免默认选中 prometheus
      const preferType = route.path.startsWith('/es') ? 'elasticsearch' : (route.path.startsWith('/k8s') ? 'kubernetes' : null)
      let defaultInstance = instanceList.value.find(inst => getInstanceType(inst) === preferType)
      if (!defaultInstance) {
          defaultInstance = instanceList.value.find(inst => ['elasticsearch', 'kubernetes'].includes(getInstanceType(inst)))
      }
      if (!defaultInstance) {
          defaultInstance = instanceList.value[0]
      }

      selectedInstance.value = defaultInstance.id
      setSelectedInstance(defaultInstance)
    }
  } catch (error) {
    console.error('获取实例列表失败:', error)
  }
}

const fetchInstanceTypes = async () => {
  try {
    const response = await getInstanceTypes()
    instanceTypes.value = response.data?.instance_types || []
  } catch (error) {
    console.error('获取类型失败:', error)
  }
}

// Utility
const getTypeIcon = (type) => {
  const icons = {
    elasticsearch: Monitor, kubernetes: Box, kibana: DataLine,
    logstash: DocumentCopy, filebeat: DocumentCopy, metricbeat: DocumentCopy, apm: WarningFilled, kafka: DataLine
  }
  return icons[type] || Monitor
}

const getTypeColor = (type) => {
  const colors = {
    elasticsearch: '#005FD4', kubernetes: '#326CE5', kibana: '#00BFB3',
    logstash: '#FEC514', filebeat: '#00BFB3', metricbeat: '#00BFB3', apm: '#8A0A4A', kafka: '#e67e22'
  }
  return colors[type] || '#666'
}

const getStatusType = (status) => ({ active: 'success', online: 'success', offline: 'danger', inactive: 'info', error: 'danger' }[status] || 'info')
const getStatusLabel = (status) => ({ active: '在线', online: '在线', offline: '离线', inactive: '非活跃', error: '异常' }[status] || status)
const getKafkaStatusType = (status) => ({ active: 'success', error: 'danger', unknown: 'info' }[status] || 'info')
const getKafkaStatusLabel = (status) => ({ active: '正常', error: '异常', unknown: '未知' }[status] || status || '未知')
const getEntryIcon = (entry) => getTypeIcon(isKafkaRoute.value ? 'kafka' : getInstanceType(entry))
const getEntryColor = (entry) => getTypeColor(isKafkaRoute.value ? 'kafka' : getInstanceType(entry))
const getEntrySubtitle = (entry) => (isKafkaRoute.value ? 'Kafka 集群' : entry.address)
const getEntryStatusType = (entry) => (isKafkaRoute.value ? getKafkaStatusType(entry.status) : getStatusType(entry.status))
const getEntryStatusLabel = (entry) => (isKafkaRoute.value ? getKafkaStatusLabel(entry.status) : getStatusLabel(entry.status))

onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  applyTheme(savedTheme === 'dark' ? 'dark' : 'light')
  if (route.path === '/kafka') {
    router.replace('/kafka/clusters')
  }
  fetchInstances()
  fetchInstanceTypes()
  if (isKafkaRoute.value) {
    kafkaStore.loadClusterOptions().catch(() => {})
  }
  queueMysqlRoutePrefetch()
})

// Watchers
watch(
  () => route.path,
  (path) => {
    if (path === '/kafka') {
      router.replace('/kafka/clusters')
      return
    }
    searchQuery.value = ''
    statusFilter.value = ''
    typeFilter.value = ''
    if (path.startsWith('/kafka/')) {
      kafkaStore.loadClusterOptions().catch(() => {})
    }
  },
)

watch(() => getSelectedInstanceType(), () => {
  // Instance Type change might trigger menu refresh automatically via computed
})
</script>

<style scoped>
.ds-app-shell {
  display: grid;
  grid-template-columns: 56px minmax(0, var(--ds-sidebar-panel-width)) minmax(0, 1fr);
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  color: var(--ds-text-primary);
  background: var(--ds-bg-app);
  font-family: var(--ds-font-sans);
}

.ds-icon-rail {
  display: flex;
  width: 56px;
  min-width: 56px;
  height: 100vh;
  flex-direction: column;
  align-items: center;
  border-right: 1px solid var(--ds-border-default);
  background: var(--ds-bg-sidebar);
  z-index: 30;
}

.ds-rail-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 40px;
  border: 0;
  border-bottom: 1px solid var(--ds-border-subtle);
  color: var(--ds-accent);
  background: transparent;
  cursor: pointer;
  font-size: 20px;
}

.ds-rail-scrollbar {
  width: 100%;
  flex: 1;
}

.ds-rail-nav,
.ds-rail-bottom {
  display: flex;
  width: 100%;
  flex-direction: column;
  gap: 2px;
  padding: 6px 0;
}

.ds-rail-bottom {
  border-top: 1px solid var(--ds-border-subtle);
}

.ds-rail-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 32px;
  margin: 0 auto;
  border: 1px solid transparent;
  border-radius: 6px;
  color: var(--ds-text-tertiary);
  background: transparent;
  cursor: pointer;
  font-size: 17px;
  transition: var(--ds-transition-fast);
}

.ds-rail-item:hover {
  color: var(--ds-text-primary);
  border-color: var(--ds-border-subtle);
  background: var(--ds-bg-hover);
}

.ds-rail-item.is-active {
  color: var(--ds-accent);
  border-color: rgba(59, 130, 246, 0.24);
  background: var(--ds-bg-active);
}

.ds-rail-item.is-active::before {
  content: '';
  position: absolute;
  left: -7px;
  width: 2px;
  height: 18px;
  border-radius: 999px;
  background: var(--ds-accent);
}

.ds-section-sidebar {
  display: flex;
  width: var(--ds-sidebar-panel-width);
  min-width: var(--ds-sidebar-panel-width);
  height: 100vh;
  flex-direction: column;
  border-right: 1px solid var(--ds-border-default);
  background: var(--ds-bg-surface);
  transition: width 160ms cubic-bezier(0.2, 0, 0, 1), min-width 160ms cubic-bezier(0.2, 0, 0, 1);
}

.ds-section-sidebar.is-collapsed {
  width: 0;
  min-width: 0;
  overflow: hidden;
  border-right: 0;
}

.ds-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
  min-height: 40px;
  padding: 0 10px 0 12px;
  border-bottom: 1px solid var(--ds-border-subtle);
}

.ds-section-title-block {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 1px;
}

.ds-section-kicker,
.ds-panel-kicker {
  color: var(--ds-text-muted);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.ds-section-title {
  overflow: hidden;
  color: var(--ds-text-primary);
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ds-section-search {
  padding: 6px 10px;
  border-bottom: 1px solid var(--ds-border-subtle);
}

.ds-section-instance {
  padding: 6px 10px;
  border-bottom: 1px solid var(--ds-border-subtle);
}

.ds-section-instance .el-select {
  width: 100%;
}

.instance-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.instance-option .type-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.instance-option-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ds-section-scrollbar {
  flex: 1;
}

.ds-section-menu {
  padding: 6px;
}

.ds-menu-block {
  margin-bottom: 4px;
}

.ds-menu-group-title {
  padding: 6px 6px 4px;
  color: var(--ds-text-muted);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.ds-menu-item {
  display: flex;
  align-items: center;
  width: 100%;
  height: 28px;
  gap: 6px;
  padding: 0 6px;
  border: 1px solid transparent;
  border-radius: 5px;
  color: var(--ds-text-tertiary);
  background: transparent;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  text-align: left;
  transition: var(--ds-transition-fast);
}

.ds-menu-item:hover {
  color: var(--ds-text-primary);
  background: var(--ds-bg-hover);
}

.ds-menu-item.is-active {
  color: var(--ds-text-primary);
  border-color: rgba(59, 130, 246, 0.22);
  background: var(--ds-bg-active);
}

.ds-menu-item.is-parent {
  margin-top: 2px;
  color: var(--ds-text-secondary);
}

.ds-menu-item.is-leaf {
  padding-left: 20px;
  color: var(--ds-text-tertiary);
}

.ds-sub-menu {
  margin: 2px 0 6px;
}

.ds-menu-empty {
  padding: 18px 8px;
  color: var(--ds-text-muted);
  font-size: 12px;
  text-align: center;
}

.ds-main {
  display: flex;
  min-width: 0;
  height: 100vh;
  flex-direction: column;
  overflow: hidden;
  background: var(--ds-bg-app);
}

.ds-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
  min-height: 40px;
  padding: 0 10px 0 14px;
  border-bottom: 1px solid var(--ds-border-default);
  background: var(--ds-bg-app);
  z-index: 20;
}

.ds-topbar-left,
.ds-topbar-actions {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.ds-breadcrumb {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 6px;
  color: var(--ds-text-tertiary);
  font-size: 12px;
  white-space: nowrap;
}

.ds-breadcrumb strong {
  overflow: hidden;
  max-width: 220px;
  color: var(--ds-text-secondary);
  font-weight: 500;
  text-overflow: ellipsis;
}

.ds-breadcrumb-separator {
  color: var(--ds-text-muted);
}

.ds-command-trigger,
.ds-icon-button,
.ds-user-menu {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  gap: 6px;
  border: 1px solid var(--ds-border-default);
  border-radius: 6px;
  color: var(--ds-text-tertiary);
  background: var(--ds-bg-surface);
  cursor: pointer;
  font-size: 12px;
  transition: var(--ds-transition-fast);
}

.ds-command-trigger {
  padding: 0 8px;
}

.ds-command-trigger kbd {
  height: 18px;
  padding: 0 5px;
  border: 1px solid var(--ds-border-default);
  border-radius: 4px;
  color: var(--ds-text-muted);
  background: var(--ds-bg-surface-2);
  font-family: var(--ds-font-mono);
  font-size: 10px;
  line-height: 16px;
}

.ds-icon-button {
  width: 32px;
  padding: 0;
}

.ds-command-trigger:hover,
.ds-icon-button:hover,
.ds-user-menu:hover {
  color: var(--ds-text-primary);
  border-color: var(--ds-border-strong);
  background: var(--ds-bg-hover);
}

.ds-user-menu {
  padding: 0 8px 0 4px;
  color: var(--ds-text-secondary);
}

.ds-user-menu span {
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ds-badge-action :deep(.el-badge__content) {
  border: 1px solid var(--ds-bg-app);
  background: var(--ds-error);
}

.ds-content {
  flex: 1;
  min-width: 0;
  overflow: auto;
  padding: 8px 10px;
  background: var(--ds-bg-app);
}

.user-dropdown-header {
  padding: 12px 14px;
  border-bottom: 1px solid var(--ds-border-subtle);
  margin-bottom: 4px;
}

.user-dropdown-header .name {
  margin: 0 0 4px;
  color: var(--ds-text-primary);
  font-size: 13px;
  font-weight: 600;
}

.user-dropdown-header .role {
  margin: 0;
  color: var(--ds-text-muted);
  font-size: 11px;
}

@media (max-width: 1023px) {
  .ds-app-shell {
    grid-template-columns: 64px minmax(0, 1fr);
  }

  .ds-section-sidebar {
    position: fixed;
    left: 64px;
    top: 0;
    bottom: 0;
    transform: translateX(-100%);
    z-index: 40;
  }

  .ds-section-sidebar:not(.is-collapsed) {
    transform: translateX(0);
  }

  .ds-breadcrumb,
  .ds-command-trigger span,
  .ds-command-trigger kbd {
    display: none;
  }
}
</style>
