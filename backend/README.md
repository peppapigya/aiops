# AIOps Backend — Go 后端服务

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-blue?logo=go" alt="Go"/>
  <img src="https://img.shields.io/badge/Gin-1.11-00ADD8?logo=go" alt="Gin"/>
  <img src="https://img.shields.io/badge/GORM-Gen-00ADD8?logo=go" alt="GORM"/>
  <img src="https://img.shields.io/badge/MySQL-8.0-orange?logo=mysql" alt="MySQL"/>
  <img src="https://img.shields.io/badge/Redis-7.x-red?logo=redis" alt="Redis"/>
</p>

## 简介

AIOps Backend 是整个 DevOps 平台的后端服务，基于 **Go 1.25 + Gin 1.11** 构建，使用 **GORM + GORM Gen** 进行数据库操作。

### 核心能力

- **RESTful API**：覆盖 Kubernetes、Elasticsearch、Kafka、MySQL、MongoDB、Helm、CI/CD、任务调度、监控、资产等模块
- **多集群管理**：通过 `map[uint]*Clientset` 管理多个 K8s 集群，支持动态增删
- **WebSocket**：Pod 实时日志流、终端交互、任务执行日志流
- **安全防护**：JWT 双 Token 认证、Redis 黑名单单点登录、IP 限流、AES-256 数据加密
- **AI 集成**：MCP 协议对接 AI Agent 进行故障诊断

---

## 项目结构

```
backend/
├── cmd/
│   ├── server/main.go                # 主服务入口
│   └── generate/                     # 代码生成
│       ├── generate.go               # GORM Gen 代码生成器
│       └── wireInfo/                 # Wire 依赖注入
├── config/config.yaml                # 主配置文件
├── deploy/devops-deploy.yaml         # K8s 部署清单
├── docs/                             # Swagger 文档 + Chaos Mesh 示例
├── internal/                         # 核心业务代码
│   ├── common/                       # 公共配置、常量、错误码
│   ├── controllers/                  # HTTP 控制器层
│   │   ├── asset/                    # 资产管理
│   │   ├── cicd/                     # CI/CD (Argo/Pipeline)
│   │   ├── common/                   # 通用控制器
│   │   ├── es/                       # Elasticsearch (backup/indices/instance/node/shard)
│   │   ├── helm/                     # Helm (repo/chart/release)
│   │   ├── k8s/                      # K8s (20+ 子模块)
│   │   ├── kafka/                    # Kafka (cluster/topic/broker/consumer/message/discovery)
│   │   ├── mcp/                      # AI MCP
│   │   ├── monitor/                  # 监控 (prometheus/custom/domain/ssl/incident)
│   │   ├── system/                   # 系统管理 (login/user/role/menu/dept/position)
│   │   └── task_scheduler/           # 任务调度 (workflow/execution)
│   ├── dal/                          # 数据访问层
│   │   ├── model/                    # 数据模型 (GORM Gen 生成 + 手写)
│   │   ├── mapper/                   # Mapper 层 (CRUD)
│   │   ├── query/                    # GORM Gen 类型安全查询
│   │   ├── redis/                    # Redis 客户端封装
│   │   ├── request/                  # 请求 DTO
│   │   └── response/                 # 响应 DTO
│   ├── middlewares/                  # 中间件 (JWT/指标/IP限流/实例认证)
│   ├── mcp/                          # AI MCP 客户端
│   ├── mongodb/                      # MongoDB 客户端
│   ├── mysql/                        # MySQL 管理模块 (独立子系统)
│   ├── routes/                       # 路由注册
│   ├── services/                     # 业务服务层
│   │   ├── custom_monitor/           # 自定义监控
│   │   ├── helm/                     # Helm 服务 (action/chart/release/repo)
│   │   ├── k8s/chaos/               # Chaos Mesh 混沌策略
│   │   │   └── strategies/           # 网络/IO/Pod/Stress 策略
│   │   ├── kafka/                    # Kafka 服务
│   │   ├── probe/                    # 实例健康探测
│   │   ├── scheduler/               # Cron 定时调度器
│   │   └── task_scheduler/           # 任务调度 (DAG执行器+执行器工厂)
│   ├── watcher/                      # 工作流 Watcher
│   └── websocket/                    # WebSocket (Pod日志/终端/执行日志)
├── pkg/                              # 可复用公共库
│   ├── certprovider/                 # SSL 证书提供商 (阿里云/腾讯云)
│   ├── configs/                      # 配置加载、DB/K8s/ES/Kafka 客户端
│   ├── database/                     # Redis 配置 + IP 限流 Lua 脚本
│   ├── feishu/                       # 飞书通知
│   ├── mysqlresponse/                # MySQL 响应格式化
│   ├── ssh/                          # SSH 执行器
│   └── utils/                        # 工具函数 (AES/JWT/日志/校验/SQL)
├── sql/devops_console-struct.sql     # 数据库 DDL
└── Dockerfile
```

---

## 应用启动流程

```
main.go
    │
    ├── 1. 创建 Gin Engine
    │       r := gin.Default()
    │
    ├── 2. 加载配置
    │       configs.LoadConfig()
    │       ├── Viper 读取 config/config.yaml
    │       ├── 环境变量覆盖 (DEVOPS_ 前缀 + DB_ 前缀)
    │       └── 解析到 AppConfig 结构体
    │
    ├── 3. 挂载中间件
    │       setMiddleware(r, globalConfig)
    │       ├── middlewares.Authenticate()     JWT 认证
    │       ├── middlewares.Metrics()          Prometheus 指标
    │       └── middlewares.IPRateLimit()      IP 限流
    │
    ├── 4. 初始化基础设施
    │       ├── database.InitRedis()            Redis 连接
    │       ├── configs.NewDB()                 MySQL + GORM
    │       ├── configs.InitConfig()            ES/K8s 客户端
    │       │   ├── InitEsClients()             ES 多实例客户端
    │       │   └── InitK8sClients()            K8s 多集群客户端
    │       └── probe.StartInstanceStatusProbe() 实例健康探测
    │
    ├── 5. 初始化 Swagger
    │       configs.InitSwagger(r)
    │
    ├── 6. 注册健康检查
    │       r.GET("/health", ...)
    │
    ├── 7. 初始化执行器
    │       executor.InitExecutors()
    │       ├── Register(HTTPExecutor)
    │       ├── Register(ScriptExecutor)
    │       ├── Register(SQLExecutor)
    │       └── Register(K8sExecutor)
    │
    ├── 8. 注册路由
    │       routers.RegisterRouters(r, db)
    │       websocket.RegisterWebSocketRoutes(r)
    │
    ├── 9. 异步启动定时任务
    │       go loadCronSchedules(db)
    │
    ├── 10. 异步启动 Prometheus 指标
    │       go http.ListenAndServe(":9090", promhttp.Handler())
    │
    └── 11. 启动 HTTP 服务
            r.Run(":8081")
```

---

## 核心实现详解

### 多集群 K8s 客户端管理

```go
// pkg/configs/k8s_client.go

// 全局变量 — 以 instanceID 为 key 管理多个集群客户端
var (
    k8sClients        map[uint]*kubernetes.Clientset       // 标准 K8s 客户端
    k8sDynamicClients map[uint]dynamic.Interface           // 动态资源客户端 (CRD)
    configMap         map[uint]*rest.Config                // REST 配置
)

// 初始化 — 启动时从数据库查询所有 K8s 实例
func InitK8sClients() error {
    // 1. 查询 kubernetes 类型的所有实例
    instances := instanceRepo.GetByTypeID(k8sType.ID)

    // 2. 为每个实例创建客户端
    for _, instance := range instances {
        // 获取认证配置 (kubeconfig)
        authConfigs := authConfigRepo.GetByInstanceID(instance.ID)

        // 解析 kubeconfig 内容
        restConfig := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))

        // 创建三种客户端
        clientSet := kubernetes.NewForConfig(restConfig)
        dynamicClient := dynamic.NewForConfig(restConfig)

        // 存储到 map
        k8sClients[instance.ID] = clientSet
        k8sDynamicClients[instance.ID] = dynamicClient
        configMap[instance.ID] = restConfig
    }
}

// 获取客户端 — 通过 instanceID 获取
func GetK8sClient(instanceID uint) (*kubernetes.Clientset, bool) {
    client, exists := k8sClients[instanceID]
    return client, exists
}

// 动态添加 — 新增实例时动态创建
func AddK8sClient(instance *dal.Instance, authConfig *dal.AuthConfig) error {
    // 解析 kubeconfig → 创建 clientset → 存入 map
}

// 动态移除 — 删除实例时移除
func RemoveK8sClient(instanceID uint) {
    delete(k8sClients, instanceID)
    delete(k8sDynamicClients, instanceID)
    delete(configMap, instanceID)
}
```

### 中间件链

```go
// internal/middlewares/middlewares.go

// 1. JWT 认证中间件
func Authenticate(excludePaths ...string) gin.HandlerFunc {
    // 支持通配符路径排除: /swagger/*, /ws/*
    // 解析 Bearer Token
    // 验证 JWT 签名
    // Redis 黑名单检测 (单点登录踢下线)
    // 将用户信息注入 gin.Context
}

// 2. Prometheus 指标中间件
func Metrics() gin.HandlerFunc {
    // 记录请求总数: HttpRequestsTotal (method, path, status)
    // 记录响应时间: HttpDuration (path)
}

// 3. IP 限流中间件
func IPRateLimit() gin.HandlerFunc {
    // Redis + Lua 脚本实现滑动窗口限流
    // 每分钟 100 次 / 每 IP
    // 超限返回 429
}

// 4. 实例认证中间件
func InstanceAuth() gin.HandlerFunc {
    // 从 URL 参数或 Header 获取 instance_id
    // 注入到 gin.Context
}
```

### DAG 任务执行器

```go
// internal/services/task_scheduler/dag_executor.go

type DAGExecutor struct {
    executionMapper *mapper.TaskExecutionMapper
    nodeExecMapper  *mapper.TaskNodeExecutionMapper
}

func (d *DAGExecutor) ExecuteWorkflow(workflowID, nodes, edges, triggeredBy) {
    // 1. 创建执行记录 (status: running)
    execution := &model.TaskExecution{...}
    executionMapper.Create(execution)

    // 2. 异步执行 DAG
    go d.runDAG(ctx, executionID, nodes, edges)
}

func (d *DAGExecutor) runDAG(nodes, edges) {
    // 构建依赖图
    dependencies := make(map[uint64][]uint64)
    inDegree := make(map[uint64]int)

    // 拓扑排序执行
    for len(inDegree) > 0 {
        readyNodes := [] // 入度为 0 的节点
        if len(readyNodes) == 0 {
            return "循环依赖" // 检测到环
        }
        for _, nodeID := range readyNodes {
            d.executeNode(ctx, executionID, node)
            delete(inDegree, nodeID)
            for _, nextID := range dependencies[nodeID] {
                inDegree[nextID]--
            }
        }
    }
}

func (d *DAGExecutor) executeNode(ctx, executionID, node) {
    // 1. 创建节点执行记录
    // 2. 通过执行器工厂获取对应执行器
    factory := executor.GetExecutorFactory()
    exec := factory.GetExecutor(node.NodeType) // "http"/"script"/"sql"/"k8s"
    // 3. 执行
    result := exec.Execute(ctx, execCtx)
    // 4. 更新节点执行结果
    // 5. WebSocket 广播日志
    websocket.BroadcastLog(executionID, logEntry)
}
```

### 执行器工厂模式

```go
// internal/services/task_scheduler/executor/executor_factory.go

type ExecutorFactory struct {
    executors map[string]TaskExecutor  // "http"→HTTPExecutor, "script"→ScriptExecutor
    mu        sync.RWMutex             // "sql"→SQLExecutor, "k8s"→K8sExecutor
}

func (f *ExecutorFactory) Register(executor TaskExecutor) {
    f.executors[executor.GetType()] = executor
}

func (f *ExecutorFactory) GetExecutor(taskType string) (TaskExecutor, bool) {
    return f.executors[taskType], exists
}

// internal/services/task_scheduler/executor/init.go
func InitExecutors() {
    factory := GetExecutorFactory()
    factory.Register(NewHTTPExecutor())    // HTTP 请求调用
    factory.Register(NewScriptExecutor())  // Shell 脚本执行
    factory.Register(NewSQLExecutor())     // SQL 语句执行
    factory.Register(NewK8sExecutor())     // K8s 资源操作
}
```

### CI/CD 流水线执行

```go
// internal/controllers/cicd/argo_controller.go

func (c *ArgoController) ExecutePipeline(ctx *gin.Context) {
    // 1. 获取流水线步骤
    pipelineInfo := pipelineMapper.GetPipelineById(pipelineId)
    steps := pipelineStepMapper.GetPipelineStepByPipelineId(pipelineId)

    // 2. 组装 Argo Workflow
    var tasks []wfv1.DAGTask
    var templates []wfv1.Template

    // 2.1 自动插入 git-clone
    gitCloneTemplate := createGitCloneTemplate(gitURL, branch, gitToken)
    templates = append(templates, gitCloneTemplate)

    // 2.2 用户步骤转换
    for _, step := range steps {
        template := createArgoWorkflowTemplateNamed(step, templateName)
        task := wfv1.DAGTask{
            Name:     argoTaskName,
            Template: templateName,
            Depends:  "git-clone && (依赖步骤名)",
        }
        tasks = append(tasks, task)
    }

    // 3. 创建 Workflow 对象
    wf := createArgoWorkflow(pipelineInfo, tasks, templates)

    // 4. 提交到 K8s 集群
    createWorkflow := argoClient.ArgoprojV1alpha1().Workflows("argo").Create(ctx, wf, ...)

    // 5. 记录运行记录
    pipelineRun := &model.PipelineRun{
        PipelineID:   pipelineId,
        WorkflowName: createWorkflow.Name,
        Status:       status,
        Operator:     operator,
        Branch:       branch,
        CommitID:     commitId,
    }
    pipelineRunMapper.CreatePipelineRun(pipelineRun)
}
```

### Chaos Mesh 策略模式

```go
// internal/services/k8s/chaos/strategies/strategy.go
type FaultStrategy interface {
    CreateSpec(request interface{}) (*unstructured.Unstructured, error)
    GetGVK() schema.GroupVersionKind
}

// 四种策略实现
// strategies/network_chaos.go  → NetworkChaosStrategy  (延迟/丢包/分区/带宽)
// strategies/io_chaos.go       → IOChaosStrategy       (文件IO延迟/错误)
// strategies/pod_chaos.go      → PodChaosStrategy      (Pod Kill/Failure)
// strategies/stress_chaos.go   → StressChaosStrategy   (CPU/Memory压力)

// internal/services/k8s/chaos/chaos_service.go
func (s *ChaosService) CreateExperiment(experimentType, request) {
    // 1. 根据类型选择策略
    strategy := s.getStrategy(experimentType)
    // 2. 策略生成 CRD 对象
    unstructuredObj := strategy.CreateSpec(request)
    // 3. 通过 Dynamic Client 提交到 K8s
    dynamicClient.Resource(strategy.GetGVK()).Namespace(ns).Create(ctx, unstructuredObj, ...)
}
```

### WebSocket 实现

```go
// internal/websocket/pod_logs.go

type PodLogHandler struct {
    clients map[*websocket.Conn]bool
}

func (h *PodLogHandler) HandleWebSocket(c *gin.Context) {
    // 1. 升级 HTTP 为 WebSocket
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)

    // 2. 获取参数
    namespace := c.Query("namespace")
    podName := c.Param("podname")
    instanceID := c.Query("instance_id")

    // 3. 获取 K8s 客户端
    client, exists := configs.GetK8sClient(instanceID)

    // 4. 获取 Pod 日志流 (Follow=true 持续监听)
    req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
        Container: container,
        Follow:    true,
        TailLines: tailLinesPtr,
    })
    logStream, _ := req.Stream(ctx)

    // 5. 循环读取日志并推送到 WebSocket
    for {
        n, _ := logStream.Read(buf)
        if n > 0 {
            conn.WriteJSON(map[string]interface{}{
                "type":    "log",
                "content": string(buf[:n]),
                "time":    time.Now().Unix(),
            })
        }
    }
}
```

---

## API 接口

### 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 路由注册

```go
// internal/routes/routers.go
func RegisterRouters(r *gin.Engine, db *gorm.DB) {
    apiGroup := r.Group("/api/v1")
    {
        elasticsearch.RegisterSubRouter(apiGroup)   // ES
        backup.RegisterSubRouter(apiGroup)           // ES 备份
        instance.RegisterSubRouter(apiGroup)         // ES 实例
        node.RegisterSubRouter(apiGroup)             // ES 节点
        shard.RegisterSubRouter(apiGroup)            // ES 分片
        indices.RegisterSubRouter(apiGroup)          // ES 索引

        k8s.RegisterK8sRoutes(apiGroup, db)          // K8s

        helmRoute := helm.NewHelmRoute(db)
        helmRoute.RegisterSubRouter(apiGroup)        // Helm

        system.RegisterSystemRouters(apiGroup)       // 系统管理
        kafka.RegisterKafkaRouters(apiGroup)         // Kafka
        mysql.RegisterMySQLRouters(apiGroup)         // MySQL
        monitor.RegisterMonitorRouters(apiGroup, db) // 监控
        mongodb.RegisterSubRouter(apiGroup)          // MongoDB
        asset.RegisterAssetRouters(apiGroup, db)     // 资产管理
        cicd.RegisterCiCdRouters(apiGroup)           // CI/CD
        task_scheduler.RegisterTaskSchedulerRouters(apiGroup, db) // 任务调度
    }
}
```

### WebSocket 端点

```
/ws/pod/:podname/logs?namespace=xxx&instance_id=xxx&container=xxx&tail=100
/ws/pod/:podname/exec?namespace=xxx&instance_id=xxx&container=xxx
/ws/executions/:id/logs
```

---

## 快速开始

### 环境要求

| 组件 | 版本 |
|------|------|
| Go | >= 1.25 |
| MySQL | >= 8.0 |
| Redis | >= 7.0 |

### 启动步骤

```bash
# 1. 修改配置
vim config/config.yaml

# 2. 初始化数据库
mysql -u root -p devops_console < sql/devops_console-struct.sql

# 3. 安装依赖
go mod download

# 4. 代码生成 (可选)
go run cmd/generate/generate.go

# 5. 启动服务
go run cmd/server/main.go
# → 服务启动在 http://localhost:8081
# → Swagger 文档: http://localhost:8081/swagger/index.html
# → Prometheus 指标: http://localhost:9090/metrics
```

### 服务端口

| 端口 | 说明 |
|------|------|
| 8081 | HTTP API 服务 |
| 9090 | Prometheus 指标 |

---

## 配置说明

```yaml
# config/config.yaml
server:
  port: ":8081"
  log_level: "info"

database:
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "password"
    database: "devops_console"
    charset: "utf8mb4"
    max_open_conns: 10
    max_idle_conns: 5

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-secret"
  expire-time: 3600
  refresh-expire-time: 604800

encryption:
  key: "base64-encoded-32bytes-key"

ai:
  mcp:
    enabled: true
    url: "http://127.0.0.1:8080"
    token: ""

feishu:
  enabled: true
  webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
```

### 环境变量

| 变量 | 对应配置 |
|------|----------|
| `DEVOPS_SERVER_PORT` | `server.port` |
| `DEVOPS_DATABASE_MYSQL_HOST` | `database.mysql.host` |
| `DB_HOST` | `database.mysql.host` (简写) |
| `DB_PORT` | `database.mysql.port` |
| `DB_USER` | `database.mysql.username` |
| `DB_PASSWORD` | `database.mysql.password` |
| `DB_NAME` | `database.mysql.database` |

---

## 技术栈

| 技术 | 用途 |
|------|------|
| Go 1.25 | 主语言 |
| Gin 1.11 | HTTP 框架 |
| GORM + GORM Gen | ORM + 类型安全代码生成 |
| Google Wire | 编译时依赖注入 |
| IBM/Sarama 1.46 | Kafka 客户端 |
| go-elasticsearch 8.19 | Elasticsearch 客户端 |
| client-go | Kubernetes 客户端 |
| argo-workflows 3.7 | Argo Workflows 客户端 |
| golang-jwt | JWT 认证 |
| logrus | 结构化日志 |
| Viper | 配置管理 |
| gorilla/websocket | WebSocket |
| Prometheus client | 指标采集 |
| Swagger | API 文档 |