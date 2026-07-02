# AiOps二进制部署文档

> 直接复制粘贴即可完成部署，无需阅读源码。

---

## 一、环境要求

| 组件 | 最低版本 | 说明 |
|------|---------|------|
| Go    | 1.25.0+  | 编译所需的go环境        |
| MySQL | 8.0+ | 数据库 |
| Redis | 6.0+ | 缓存 & Session |
| Nginx | 1.20+ | 前端静态资源 & 反向代理 |

---

## 二、部署前准备

### 2.1 解压代码或者从github拉取

```bash
mkdir -p /opt/aiops/{backend,frontend}
# 先把文件上传到/tmp目录

# 先解压文件，解压之后会多出来feontend、和backend目录
cd /tmp && tar xf aiops.tar
```

### 2.2 创建数据库

```bash
mysql -u root -p << 'EOF'
CREATE DATABASE IF NOT EXISTS aiops
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;
EOF

# 检查数据库是否创建成功
mysql -u root -p -e "show databases"
```

### 2.3 导入表结构

按顺序执行以下 SQL 文件（文件位于项目 `backend/sql/` 目录）：

```bash
mysql -u root -p aiops < ./backend/sql/aiops.sql

# 查看表结构是否导入成功
mysql -u root -p -e "use aiops; show tables;"
```

---

## 三、后端部署

### 3.1 编译后端

在开发机上执行（或直接在目标服务器上编译）：

```bash
cd /tmp/backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o devops-server ./cmd/server

# 将构建产物移动到aiops
mv devops-server /opt/aiops/backend
```

### 3.3 创建配置文件

在服务器上创建 `/opt/devops-console/backend/config.yaml`：

需要修改数据库和redis的密码

```bash

cat > /opt/aiops/backend/config.yaml << 'EOF'
server:
  port: ":8081"
  log_level: "info"

database:
  type: "mysql"
  mysql:
    host: "127.0.0.1"        # 改为你的 MySQL 地址
    port: 3306                # 改为你的 MySQL 端口
    username: "root"          # 改为你的 MySQL 用户名
    password: "your_password" # 改为你的 MySQL 密码
    database: "aiops"
    charset: "utf8mb4"
    parse_time: true
    max_open_conns: 20
    max_idle_conns: 10

logging:
  format: "json"
  time_format: "2006-01-02 15:04:05"
  report_caller: true

app:
  name: "aiops"
  version: "1.0.0"
  environment: "production"

elasticsearch:
  timeout: 30
  retry: 3
  health_check_interval: 60

kubernetes:
  config_path: ""
  timeout: 30
  retry: 3

swagger:
  enabled: false
  host: "localhost:8081"
  base_path: "/"

health:
  enabled: true
  endpoint: "/health"
  interval: 30

redis:
  host: 127.0.0.1            # 改为你的 Redis 地址
  port: 6379                  # 改为你的 Redis 端口
  password: ""                # 改为你的 Redis 密码
  db: 0

ai:
  mcp:
    enabled: false
    url: "http://127.0.0.1:8080"
    token: ""
    max_retries: 3

jwt:
  secret: "n02y2Zqf4eL0hZ4xjQH9w1zDk1w5FqMnc9R+N8T1v2E="
  expire-time: 3600
  refresh-expire-time: 604800
  exclude-paths:
    - /api/v1/system/login
    - /api/v1/sysUser/refresh
    - /api/v1/sysUser/captcha
    - /swagger/*
    - /jobs/script/
    - /metrics
    - /ws/*
    - /health

encryption:
  key: "So722DGlBzRASCG9so/Knfuy81cdhQ7gArgZjcJHvpQ="

feishu:
  enabled: false
  webhook: ""
  secret: ""
  appId: ""
  appSecret: ""
EOF
```

> 生产环境务必修改 `jwt.secret` 和 `encryption.key`，用 `openssl rand -base64 32` 生成。

### 3.4 创建 Systemd 服务

```bash
cat > /etc/systemd/system/aiops-backend.service << 'EOF'
[Unit]
Description=DevOps Console Backend
After=network.target mysql.service redis.service
Wants=mysql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/aiops/backend
ExecStart=/opt/aiops/backend/devops-server
Restart=always
RestartSec=5
StandardOutput=append:/opt/aiops/backend/server.log
StandardError=append:/opt/aiops/backend/server.log

[Install]
WantedBy=multi-user.target
EOF
```

### 3.5 启动后端

```bash
systemctl daemon-reload
systemctl start aiops-backend
systemctl enable aiops-backend
systemctl status aiops-backend
```

### 3.6 验证后端

```bash
curl http://127.0.0.1:8081/health
# 预期返回: {"status":"ok","timestamp":"..."}
```

---

## 四、前端部署

### 4.1 编译前端

```bash
cd /tmp/frontend && rm -rf ./node_modules
npm install
npm run build
```

### 4.2 上传到服务器

```bash
cp -r dist/* /opt/aiops/frontend/
```

### 4.3 创建 Nginx 配置

```bash
cat > /etc/nginx/conf.d/aiops.conf << 'EOF'
server {
    listen 8084;
    server_name _;
    root /opt/aiops/frontend;
    index index.html;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/v1/ {
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    location /ws/ {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF
```

### 4.4 启动 Nginx

```bash
nginx -t && nginx -s reload
```

---

## 五、验证部署

浏览器访问 `http://your-server-ip`，使用默认账号登录：

| 项目 | 值 |
|------|------|
| 用户名 | `admin` |
| 密码 | `admin` |

> 登录后请立即修改密码。








