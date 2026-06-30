# AIOps Frontend — Vue 3 前端应用

<p align="center">
  <img src="https://img.shields.io/badge/Vue-3.5-brightgreen?logo=vue.js" alt="Vue"/>
  <img src="https://img.shields.io/badge/Element_Plus-2.4-409EFF?logo=element" alt="Element Plus"/>
  <img src="https://img.shields.io/badge/Vite-7.1-646CFF?logo=vite" alt="Vite"/>
  <img src="https://img.shields.io/badge/Pinia-3.0-FFD859?logo=vue.js" alt="Pinia"/>
  <img src="https://img.shields.io/badge/TypeScript-6.0-3178C6?logo=typescript" alt="TypeScript"/>
</p>

## 简介

AIOps Frontend 是整个 DevOps 平台的前端应用，基于 **Vue 3.5 + Element Plus 2.4 + Vite 7.1** 构建，采用 **三栏布局 + 动态路由** 架构，提供丰富的运维管理界面。

### 核心能力

- **三栏布局**：图标导航栏 + 可折叠侧边栏 + 主内容区，支持移动端响应式
- **动态路由**：后端菜单表驱动，`import.meta.glob` 动态匹配组件
- **RBAC 权限**：`permStore.hasPerm()` 控制按钮/菜单显隐，admin 角色自动拥有全部权限
- **暗色/亮色主题**：基于 CSS 变量的设计系统，一键切换
- **全局实例选择器**：顶部栏切换 ES/K8s 实例或 Kafka 集群
- **WebSocket 集成**：Pod 实时日志流、终端交互、任务执行日志流
- **MySQL 独立子系统**：独立的 Axios 实例 + Pinia stores + 预加载优化

---

## 项目结构

```
frontend/
├── src/
│   ├── main.js                       # 应用入口
│   ├── App.vue                       # 根组件 (<router-view/>)
│   ├── copy.ts                       # MongoDB 中英文文案常量
│   ├── style.css                     # 全局基础样式
│   ├── api/                          # API 调用层
│   │   ├── index.js                  # Axios 实例 + 拦截器
│   │   ├── instance.js               # 实例 CRUD + 连接测试
│   │   ├── helm.js                   # Helm 仓库/Chart/Release
│   │   ├── cicd.js                   # CI/CD 流水线
│   │   ├── workflow.js               # 任务调度工作流
│   │   ├── chaos.js                  # Chaos Mesh 实验
│   │   ├── monitor.js                # 自定义监控大盘
│   │   ├── asset.js                  # 资产管理
│   │   ├── domain.js                 # 域名监控 + SSL 证书
│   │   ├── incident.js               # 故障事件
│   │   ├── kafka.js                  # Kafka 集群管理
│   │   ├── system/                   # 系统管理 API
│   │   │   ├── user.js               # 登录/用户CRUD/个人信息
│   │   │   ├── role.js               # 角色CRUD/菜单分配
│   │   │   ├── menu.js               # 菜单树CRUD
│   │   │   ├── dept.js               # 部门树CRUD
│   │   │   └── position.js           # 岗位CRUD
│   │   ├── k8s/                      # K8s API (20+ 文件)
│   │   │   ├── cluster.js            # 集群列表
│   │   │   ├── pod.js                # Pod CRUD/日志/事件
│   │   │   ├── deployment.js         # Deployment CRUD/扩缩容
│   │   │   ├── statefulset.js        # StatefulSet
│   │   │   ├── daemonset.js          # DaemonSet
│   │   │   ├── job.js                # Job
│   │   │   ├── cronjob.js            # CronJob
│   │   │   ├── service.js            # Service
│   │   │   ├── network.js            # Ingress
│   │   │   ├── configmap.js          # ConfigMap
│   │   │   ├── secret.js             # Secret
│   │   │   ├── storage.js            # PV/PVC/StorageClass
│   │   │   ├── hpa.js                # HPA
│   │   │   ├── vpa.js                # VPA
│   │   │   ├── node.js               # Node CRUD/cordon/drain
│   │   │   ├── namespace.js          # Namespace
│   │   │   ├── event.js              # 事件
│   │   │   ├── rbac.js               # RBAC
│   │   │   └── ...                   # 其他资源
│   │   ├── es/                       # ES API (7 文件)
│   │   │   ├── cluster.js            # 集群状态/健康/信息
│   │   │   ├── indices.js            # 索引 CRUD
│   │   │   ├── node.js               # 节点信息
│   │   │   ├── shard.js              # 分片信息
│   │   │   ├── backup.js             # 快照管理
│   │   │   └── data.js               # 通用 ES 请求
│   │   └── mongodb/                  # MongoDB API (TypeScript)
│   │       ├── http.ts               # MongoDB 专用 Axios 实例
│   │       ├── mongo.ts              # 数据库/集合/文档 CRUD
│   │       └── types.ts              # 类型定义
│   ├── components/                   # 公共组件
│   │   ├── CodeEditor.vue            # Monaco Editor 封装
│   │   ├── PodTerminal.vue           # Pod 终端 (xterm.js + WebSocket)
│   │   ├── BrokerConfigDialog.vue    # Kafka Broker 配置对话框
│   │   └── mongodb/                  # MongoDB 专用组件
│   │       ├── SidebarTree.vue       # 数据库树
│   │       └── DocumentEditorDrawer.vue # 文档编辑抽屉
│   ├── layouts/
│   │   └── AppLayout.vue             # 三栏布局主组件
│   ├── router/                       # 路由配置
│   │   ├── index.js                  # 路由入口 (静态路由 + 动态注册)
│   │   └── chaos-routes.js           # 混沌实验路由
│   ├── stores/                       # Pinia 状态管理
│   │   ├── permissionStore.js        # 用户信息/菜单/权限/动态路由
│   │   ├── instanceStore.js          # ES/K8s 实例选择状态
│   │   ├── kafkaStore.js             # Kafka 集群选择状态
│   │   └── mongo.ts                  # MongoDB 连接信息
│   ├── views/                        # 页面组件 (按模块分目录)
│   │   ├── Dashboard.vue             # 总览仪表板
│   │   ├── 404.vue                   # 404 页面
│   │   ├── k8s/                      # K8s 管理页面 (30+ 页面)
│   │   │   ├── K8sDashboard.vue      # K8s 仪表板
│   │   │   ├── ClusterDetail.vue     # 集群详情
│   │   │   ├── DeploymentManagement.vue  # Deployment 管理
│   │   │   ├── PodManagement.vue     # Pod 管理
│   │   │   ├── NodeManagement.vue    # 节点管理
│   │   │   ├── NamespaceManagement.vue   # 命名空间
│   │   │   ├── ServiceManagement.vue     # Service
│   │   │   ├── IngressManagement.vue     # Ingress
│   │   │   ├── ConfigMapManagement.vue   # ConfigMap
│   │   │   ├── SecretManagement.vue      # Secret
│   │   │   ├── PvManagement.vue          # PV
│   │   │   ├── PvcManagement.vue         # PVC
│   │   │   ├── HpaManagement.vue         # HPA
│   │   │   ├── CrdManagement.vue         # CRD
│   │   │   ├── OperatorManagement.vue    # Operator
│   │   │   ├── EventManagement.vue       # 事件
│   │   │   ├── Workload.vue              # 工作负载总览
│   │   │   └── components/               # K8s 专用组件
│   │   │       ├── DynamicPromQLChart.vue    # PromQL 图表
│   │   │       └── NodeMonitorChart.vue      # 节点监控图表
│   │   ├── es/                       # ES 管理页面
│   │   │   ├── EsDashboard.vue       # ES 仪表板
│   │   │   ├── InstanceManagement.vue    # 实例管理
│   │   │   ├── InstanceForm.vue          # 实例表单
│   │   │   ├── InstanceDetail.vue        # 实例详情
│   │   │   ├── IndexManagement.vue       # 索引管理
│   │   │   ├── DataManagement.vue        # 数据管理
│   │   │   ├── NodeManagement.vue        # 节点管理
│   │   │   ├── ShardManagement.vue       # 分片管理
│   │   │   ├── BackupManagement.vue      # 备份管理
│   │   │   └── Settings.vue              # 设置
│   │   ├── kafka/                    # Kafka 管理页面
│   │   │   ├── ClusterManagement.vue     # 集群管理
│   │   │   ├── TopicManagement.vue       # Topic 管理
│   │   │   ├── BrokerManagement.vue      # Broker 管理
│   │   │   ├── ConsumerGroupManagement.vue # 消费者组
│   │   │   ├── MessageBrowser.vue        # 消息浏览
│   │   │   ├── DiscoveryCenter.vue       # 自动发现
│   │   │   └── AuditLog.vue              # 审计日志
│   │   ├── mysql/                    # MySQL 管理页面
│   │   │   ├── ConnectionManagement.vue  # 连接管理/工作台
│   │   │   ├── DatabaseManagement.vue    # 数据库管理
│   │   │   ├── DataManagement.vue        # 数据管理
│   │   │   ├── SqlQuery.vue              # SQL 查询
│   │   │   ├── UserPermissionManagement.vue # 用户权限
│   │   │   └── BackupManagement.vue      # 备份管理
│   │   ├── mongodb/                  # MongoDB 管理页面
│   │   │   ├── ConnectView.vue       # 连接配置
│   │   │   ├── DashboardView.vue     # 仪表板
│   │   │   └── DocsView.vue          # API 文档
│   │   ├── cicd/                     # CI/CD 页面
│   │   │   ├── PipelineList.vue      # 流水线列表
│   │   │   ├── PipelineEditor.vue    # 流水线编辑器
│   │   │   ├── PipelineRunList.vue   # 运行列表
│   │   │   ├── PipelineRunDetail.vue # 运行详情
│   │   │   └── components/
│   │   │       └── PipelineGraph.vue # DAG 图 (@vue-flow)
│   │   ├── chaos/                    # 混沌实验页面
│   │   │   ├── ChaosMeshList.vue     # 实验列表
│   │   │   ├── ChaosMeshCreate.vue   # 创建实验
│   │   │   └── ChaosMeshDetail.vue   # 实验详情
│   │   ├── helm/                     # Helm 页面
│   │   │   ├── RepoManagement.vue    # 仓库管理
│   │   │   ├── AppStore.vue          # 应用商店
│   │   │   └── InstalledApps.vue     # 已安装应用
│   │   ├── task-scheduler/           # 任务调度页面
│   │   │   ├── WorkflowList.vue      # 工作流列表
│   │   │   ├── WorkflowEditor.vue    # 工作流编辑器
│   │   │   ├── ExecutionHistory.vue  # 执行历史
│   │   │   ├── ExecutionLog.vue      # 执行日志
│   │   │   └── routes.js             # 路由定义
│   │   ├── asset/                    # 资产管理页面
│   │   │   └── HostManagement.vue    # 主机管理
│   │   ├── monitor/                  # 监控中心页面
│   │   │   ├── DomainManagement.vue  # 域名监控
│   │   │   └── IncidentManagement.vue # 故障事件
│   │   └── system/                   # 系统管理页面
│   │       ├── login/index.vue       # 登录页
│   │       ├── UserManagement.vue    # 用户管理
│   │       ├── RoleManagement.vue    # 角色管理
│   │       ├── MenuManagement.vue    # 菜单管理
│   │       ├── DeptManagement.vue    # 部门管理
│   │       ├── PositionManagement.vue # 岗位管理
│   │       └── ProfileCenter.vue     # 个人中心
│   ├── mysql/                        # MySQL 可视化子系统
│   │   ├── runtime.ts                # 运行时配置
│   │   ├── components/               # MySQL 专用组件
│   │   │   ├── layout/MainLayout.vue # MySQL 主布局
│   │   │   ├── shared/               # 共享组件
│   │   │   │   ├── MySQLExplorerTree.vue   # 数据库浏览器树
│   │   │   │   ├── MySQLPageHeader.vue     # 页面头部
│   │   │   │   └── MySQLPageSkeleton.vue   # 骨架屏
│   │   │   └── workspace/            # 工作区组件
│   │   │       ├── QueryTab.vue          # SQL 查询标签页
│   │   │       ├── TableDataTab.vue      # 表数据标签页
│   │   │       ├── BackupTab.vue         # 备份标签页
│   │   │       ├── UserManagementTab.vue # 用户管理标签页
│   │   │       └── StructureCompareDialog.vue # 结构对比对话框
│   │   ├── stores/                   # MySQL Pinia stores
│   │   │   ├── connection.ts         # 连接管理 (Token/重连)
│   │   │   ├── workspace.ts          # 活跃数据库/表
│   │   │   └── locale.ts             # 语言设置 (zh-CN/en-US)
│   │   ├── styles/                   # MySQL 样式
│   │   ├── utils/                    # MySQL 工具函数
│   │   │   ├── request.ts            # MySQL 专用 Axios 实例
│   │   │   ├── explorer.ts           # 资源管理器工具
│   │   │   └── security-overview.ts  # 安全概览工具
│   │   └── views/                    # MySQL 页面
│   ├── session/                      # 会话管理
│   │   └── mongoConnection.ts        # MongoDB 连接 (sessionStorage)
│   ├── utils/                        # 工具函数
│   │   └── menuIcons.js              # 菜单图标解析
│   ├── styles/                       # 设计系统样式
│   │   └── design-system.css         # CSS 变量设计系统
│   └── style/                        # 主题样式
├── nginx.conf                        # Nginx 配置
├── vite.config.js                    # Vite 配置
├── tsconfig.json                     # TypeScript 配置
├── Dockerfile                        # Docker 构建
└── package.json
```

---

## 核心实现详解

### 动态路由机制

```
后端 sys_menu 表
    component 字段: "k8s/DeploymentManagement"
    path 字段: "/k8s/deployments"
    perm 字段: "k8s:deployment:list"
         │
         ▼
GET /api/v1/system/auth/info → 返回 menus + perms + roles
         │
         ▼
前端 permissionStore.loadUserAndRoutes(router)
    │
    ├── 1. 存储用户信息、菜单树、权限列表
    │
    ├── 2. flattenMenuToRoutes(menuTree)
    │       遍历菜单树，只处理 type=2 (菜单) 且有 path 和 component 的节点
    │
    ├── 3. import.meta.glob('../views/**/*.vue')
    │       动态匹配组件路径
    │       component: "k8s/DeploymentManagement" → "../views/k8s/DeploymentManagement.vue"
    │
    └── 4. router.addRoute('AppLayout', route)
            将动态路由注册为 AppLayout 的子路由
```

```javascript
// stores/permissionStore.js

const viewModules = import.meta.glob('../views/**/*.vue')

function resolveComponent(component) {
    if (!component) return null
    const fullPath = `../views/${component}.vue`
    return viewModules[fullPath] || null
}

function flattenMenuToRoutes(menus) {
    let routes = []
    for (const menu of menus) {
        if (menu.type === 2 && menu.path && menu.component) {
            const compLoader = resolveComponent(menu.component)
            routes.push({
                path: menu.path,
                component: compLoader || (() => import('../views/404.vue')),
                meta: {
                    title: menu.name,
                    icon: menu.icon || '',
                    menuId: menu.id,
                    hidden: menu.visible === 0,
                },
            })
        }
        if (menu.children && menu.children.length > 0) {
            routes.push(...flattenMenuToRoutes(menu.children))
        }
    }
    return routes
}

async function loadUserAndRoutes(router) {
    const res = await getAuthInfo()
    const data = res?.data?.data || {}

    userInfo.value = { userId, username, nickname, email, phone, avatar }
    menuTree.value = data.menus || []
    perms.value = data.perms || []
    roles.value = data.roles || []

    const dynamicRoutes = flattenMenuToRoutes(menuTree.value)
    for (const route of dynamicRoutes) {
        router.addRoute('AppLayout', route)
    }
    isLoaded.value = true
}

function hasPerm(perm) {
    return perms.value.includes(perm) || roles.value.includes('admin')
}
```

### 路由守卫流程

```javascript
// router/index.js

router.beforeEach(async (to, from, next) => {
    // 1. 设置页面标题
    document.title = `${to.meta.title} - DevOps Console`

    // 2. 检查 Token
    const token = localStorage.getItem('access_token')
    if (!token) {
        if (to.path === '/login') next()
        else next(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
        return
    }

    // 3. 已登录跳转到首页
    if (to.path === '/login') { next('/'); return }

    // 4. MongoDB 连接检查
    if (to.meta?.requiresMongoConnection) {
        if (!hasMongoSessionConnection()) {
            next({ name: 'MongoDBConnect', query: { redirect: to.fullPath } })
            return
        }
    }

    // 5. 首次加载 → 获取用户信息和动态路由
    const permStore = usePermissionStore()
    if (!permStore.isLoaded) {
        await permStore.loadUserAndRoutes(router)
        next({ path: to.path, query: to.query, replace: true })
    } else {
        next()
    }
})
```

### 三栏布局

```
┌──────────┬──────────────────┬─────────────────────────────────────────┐
│ 图标导航栏 │ 可折叠侧边栏        │ 主内容区                                  │
│ (56px)   │ (240px)          │ (自适应)                                  │
│          │                  │                                         │
│  ┌────┐  │ 工作区             │ ┌─────────────────────────────────────┐ │
│  │ 📊 │  │ ┌───────────────┐ │ │ 面包屑 / 实例选择器 / 用户菜单         │ │
│  └────┘  │ │ 搜索菜单        │ │ ├─────────────────────────────────────┤ │
│  ┌────┐  │ └───────────────┘ │ │                                     │ │
│  │ ☸  │  │                  │ │  <router-view>                       │ │
│  └────┘  │ 菜单组1           │ │                                     │ │
│  ┌────┐  │  ├ 菜单项1        │ │  当前页面内容                          │ │
│  │ 🔥 │  │  ├ 菜单项2        │ │                                     │ │
│  └────┘  │  └ 菜单项3        │ │                                     │ │
│  ┌────┐  │                  │ │                                     │ │
│  │ 📦 │  │ 菜单组2           │ │                                     │ │
│  └────┘  │  ├ 菜单项4        │ │                                     │ │
│  ┌────┐  │  └ 菜单项5        │ └─────────────────────────────────────┘ │
│  │ 🔍 │  │                  │                                         │
│  └────┘  │                  │                                         │
│  ...     │                  │                                         │
└──────────┴──────────────────┴─────────────────────────────────────────┘
```

### 全局实例选择器

```javascript
// AppLayout.vue

// 根据当前路由自动切换选择器类型
const isKafkaRoute = computed(() => route.path.startsWith('/kafka/'))

// 非 Kafka 路由 → 显示 ES/K8s 实例选择器
// Kafka 路由 → 显示 Kafka 集群选择器

// 实例列表下拉
const currentSelectionList = computed(() =>
    isKafkaRoute.value ? filteredKafkaClusterList.value : filteredInstanceList.value
)

// 选择实例
function selectInstance(instanceId) {
    if (isKafkaRoute.value) {
        kafkaStore.setSelectedClusterId(instanceId)
    } else {
        selectedInstance.value = instanceId
        setSelectedInstance(instance)
    }
}
```

### 暗色/亮色主题

```javascript
// AppLayout.vue

function applyTheme(theme) {
    currentTheme.value = theme
    const html = document.documentElement
    html.classList.toggle('dark', theme === 'dark')
    html.classList.toggle('light', theme === 'light')
    localStorage.setItem('theme', theme)
}

onMounted(() => {
    const savedTheme = localStorage.getItem('theme')
    applyTheme(savedTheme === 'light' ? 'light' : 'dark') // 默认暗色
})
```

### MySQL 子系统预加载优化

```javascript
// AppLayout.vue

// 定义 MySQL 页面组件预加载映射
const mysqlRouteComponentPreloaders = {
    '/mysql/workbench': () => import('../views/mysql/ConnectionManagement.vue'),
    '/mysql/databases': () => import('../views/mysql/DatabaseManagement.vue'),
    '/mysql/data': () => import('../views/mysql/DataManagement.vue'),
    '/mysql/query': () => import('../views/mysql/SqlQuery.vue'),
    '/mysql/security': () => import('../views/mysql/UserPermissionManagement.vue'),
    '/mysql/backup': () => import('../views/mysql/BackupManagement.vue')
}

// 空闲时预加载
function queueMysqlRoutePrefetch() {
    if ('requestIdleCallback' in window) {
        window.requestIdleCallback(() => {
            Object.keys(mysqlRouteComponentPreloaders).forEach(prefetchMenuRoute)
        }, { timeout: 1200 })
    } else {
        window.setTimeout(() => { /* ... */ }, 200)
    }
}

// 鼠标悬停 MySQL 菜单时预加载
function handleMenuHover(path) {
    if (typeof path === 'string' && path.startsWith('/mysql/')) {
        prefetchMenuRoute(path)
    }
}
```

### Pod 终端实现

```vue
<!-- components/PodTerminal.vue -->
<template>
  <div ref="terminalContainer" class="pod-terminal"></div>
</template>

<script setup>
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import 'xterm/css/xterm.css'

const terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
    },
})

const fitAddon = new FitAddon()
terminal.loadAddon(fitAddon)
terminal.loadAddon(new WebLinksAddon())

// WebSocket 连接
const wsUrl = `${import.meta.env.VITE_WS_BASE_URL}/ws/pod/${podName}/exec`
const ws = new WebSocket(`${wsUrl}?namespace=${namespace}&instance_id=${instanceId}`)

// 终端输入 → WebSocket
terminal.onData((data) => {
    ws.send(JSON.stringify({ type: 'stdin', data }))
})

// WebSocket → 终端输出
ws.onmessage = (event) => {
    const msg = JSON.parse(event.data)
    if (msg.type === 'stdout') {
        terminal.write(msg.data)
    }
}

terminal.open(terminalContainer.value)
fitAddon.fit()
</script>
```

### CI/CD 流水线 DAG 图

```vue
<!-- views/cicd/components/PipelineGraph.vue -->
<template>
  <VueFlow :nodes="nodes" :edges="edges" :default-viewport="{ zoom: 1 }">
    <Background />
    <Controls />
  </VueFlow>
</template>

<script setup>
import { VueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'

// 将流水线步骤转换为 DAG 图节点和边
const nodes = steps.map((step, index) => ({
    id: step.name,
    type: 'default',
    position: { x: 250 * index, y: 100 },
    data: { label: step.name },
}))

const edges = steps
    .filter(step => step.depends_on)
    .flatMap(step =>
        step.depends_on.split(',').map(dep => ({
            id: `${dep.trim()}-${step.name}`,
            source: dep.trim(),
            target: step.name,
        }))
    )
</script>
```

### API 请求层

```javascript
// api/index.js — Axios 实例配置

const service = axios.create({
    baseURL: '/api/v1',
    timeout: 60000,
})

// 请求拦截器 — 注入 Bearer Token
service.interceptors.request.use(config => {
    const token = localStorage.getItem('access_token')
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})

// 响应拦截器 — 401 跳转登录
service.interceptors.response.use(
    response => response,
    error => {
        if (error.response?.status === 401) {
            localStorage.removeItem('access_token')
            localStorage.removeItem('refresh_token')
            window.location.href = '/login'
        }
        return Promise.reject(error)
    }
)
```

### Pinia 状态管理

| Store | 文件 | 功能 |
|-------|------|------|
| `permission` | `stores/permissionStore.js` | 用户信息、菜单树、权限列表、动态路由注册 |
| `instance` | `stores/instanceStore.js` | ES/K8s 实例选择，持久化到 localStorage |
| `kafka` | `stores/kafkaStore.js` | Kafka 集群选项、选中集群，持久化到 localStorage |
| `mongo` | `stores/mongo.ts` | MongoDB 连接信息、数据库列表、集合缓存 |
| `mysqlConnection` | `mysql/stores/connection.ts` | MySQL 连接 token、连接状态、自动重连 |
| `mysqlWorkspace` | `mysql/stores/workspace.ts` | MySQL 当前活跃数据库和表 |
| `mysqlLocale` | `mysql/stores/locale.ts` | MySQL 子系统语言设置 (zh-CN/en-US) |

---

## 快速开始

### 环境要求

| 组件 | 版本 |
|------|------|
| Node.js | >= 20 |
| npm | 最新稳定版 |

### 启动步骤

```bash
# 1. 安装依赖
npm install

# 2. 配置环境变量 (可选)
# .env 文件
VITE_WS_BASE_URL=ws://127.0.0.1:8081

# 3. 启动开发服务器
npm run dev
# → 前端启动在 http://localhost:5173
# → API 请求自动代理到 http://localhost:8081
```

### 构建生产版本

```bash
npm run build
# → 输出到 dist/ 目录
```

### 预览构建结果

```bash
npm run preview
```

---

## 配置说明

### Vite 配置 (`vite.config.js`)

```javascript
export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': '/src',  // @ 别名指向 src/
        },
    },
    server: {
        proxy: {
            '/api': {
                target: 'http://127.0.0.1:8081',  // 后端 API 代理
                changeOrigin: true,
            },
            '/ws': {
                target: 'ws://127.0.0.1:8081',  // WebSocket 代理
                ws: true,
            },
        },
    },
})
```

### Nginx 配置 (`nginx.conf`)

```nginx
server {
    listen 8081;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # Vue Router history 模式
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api/v1/ {
        proxy_pass http://devops-backend-svc:8081;
    }

    # WebSocket 代理
    location /ws/ {
        proxy_pass http://devops-backend-svc:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}

# Prometheus 代理
server {
    listen 9200;
    location / {
        proxy_pass http://devops-backend-svc:9200;
    }
}
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `VITE_WS_BASE_URL` | WebSocket 连接地址 | `ws://127.0.0.1:8081` |

---

## 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue | 3.5 | 核心框架 |
| Element Plus | 2.4 | UI 组件库 |
| Vite | 7.1 | 构建工具 |
| Vue Router | 4.2 | 路由管理 |
| Pinia | 3.0 | 状态管理 |
| Axios | 1.6 | HTTP 客户端 |
| ECharts | 5.4 | 图表可视化 |
| Monaco Editor | 0.55 | 代码编辑器 |
| xterm | 5.3 | 终端模拟器 |
| @vue-flow/core | 1.48 | DAG 流水线编辑器 |
| openai SDK | 6.9 | AI Agent 对话 |
| xlsx | 0.18 | Excel 导入导出 |
| crypto-js | 4.2 | 前端加密 |
| bcryptjs | 3.0 | 前端密码哈希 |
| js-yaml | 4.1 | YAML 解析 |
| dayjs | 1.11 | 日期处理 |
| TypeScript | 6.0 | 类型检查 |