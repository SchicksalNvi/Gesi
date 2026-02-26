# Superview

Centralized Supervisor Interface - 多节点 Supervisor 进程管理系统

## 快速开始

```bash
# 编译
./superview.sh build

# 启动
./superview.sh start

# 访问
http://localhost:8081
```

默认账号：`admin` / 见 `config/.env` 中的 `ADMIN_PASSWORD`

## 管理命令

```bash
./superview.sh build           # 编译前后端
./superview.sh build-backend   # 仅编译后端
./superview.sh build-frontend  # 仅编译前端
./superview.sh start           # 后台启动
./superview.sh stop            # 停止
./superview.sh restart         # 重启
./superview.sh status          # 查看状态
./superview.sh run             # 前台运行（开发用）
```

## 配置

```
config/
├── config.toml      # 系统配置（端口、日志、metrics等）
├── nodelist.toml    # Supervisor 节点列表
└── .env             # 敏感配置（JWT_SECRET、密码等）
```

### 环境变量 (config/.env)

```bash
JWT_SECRET=your-secret-key
ADMIN_PASSWORD=admin123
NODE_PASSWORD=supervisor-password
```

## 功能

- **节点管理** - 多 Supervisor 节点集中管理
- **进程控制** - 启动/停止/重启进程
- **实时监控** - WebSocket 推送状态变化
- **节点发现** - CIDR 网段扫描自动发现节点
- **用户管理** - 基于角色的访问控制 (RBAC)
- **活动日志** - 操作审计追踪
- **Prometheus 指标** - 暴露监控指标供外部系统采集

## Prometheus 监控指标

Superview 支持暴露 Prometheus 格式的监控指标，可接入现有监控系统。

### 启用配置

在 `config/config.toml` 中添加：

```toml
[metrics]
enabled = true
path = "/metrics"
username = "prometheus"  # 可选，Basic Auth 用户名
password = "secret"      # 可选，Basic Auth 密码
```

### Prometheus 采集配置

```yaml
scrape_configs:
  - job_name: 'superview'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: /metrics
    basic_auth:  # 如果启用了认证
      username: prometheus
      password: secret
```

### 指标列表

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `superview_node_up` | gauge | node, environment, host, port | 节点连接状态 (1=在线, 0=离线) |
| `superview_node_last_ping_timestamp_seconds` | gauge | node | 最后成功连接时间戳 |
| `superview_process_state` | gauge | node, process, group | 进程状态码 |
| `superview_process_up` | gauge | node, process, group | 进程运行状态 (1=运行中, 0=未运行) |
| `superview_process_pid` | gauge | node, process, group | 进程 PID |
| `superview_process_uptime_seconds` | gauge | node, process, group | 进程运行时长（秒） |
| `superview_process_start_timestamp_seconds` | gauge | node, process, group | 进程启动时间戳 |
| `superview_process_exit_status` | gauge | node, process, group | 进程退出状态码 |
| `superview_nodes_total` | gauge | environment (可选) | 配置的节点总数 |
| `superview_nodes_connected` | gauge | environment (可选) | 已连接节点数 |
| `superview_processes_total` | gauge | environment (可选) | 进程总数 |
| `superview_processes_running` | gauge | environment (可选) | 运行中进程数 |
| `superview_processes_stopped` | gauge | - | 已停止进程数 |
| `superview_processes_failed` | gauge | - | 失败进程数 (FATAL/EXITED) |
| `superview_info` | gauge | version | 构建信息 |

### 进程状态码

| 状态码 | 状态名 | 说明 |
|--------|--------|------|
| 0 | STOPPED | 已停止 |
| 10 | STARTING | 启动中 |
| 20 | RUNNING | 运行中 |
| 30 | BACKOFF | 重试中 |
| 40 | STOPPING | 停止中 |
| 100 | EXITED | 已退出 |
| 200 | FATAL | 致命错误 |
| 1000 | UNKNOWN | 未知 |

### 告警规则示例

```yaml
groups:
  - name: superview
    rules:
      - alert: NodeDown
        expr: superview_node_up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Node {{ $labels.node }} is down"
          
      - alert: ProcessDown
        expr: superview_process_up == 0
        for: 30s
        labels:
          severity: warning
        annotations:
          summary: "Process {{ $labels.process }} on {{ $labels.node }} is not running"
          
      - alert: ProcessFailed
        expr: superview_process_state == 200
        for: 0s
        labels:
          severity: critical
        annotations:
          summary: "Process {{ $labels.process }} on {{ $labels.node }} has FATAL status"
```

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.24 + Gin + GORM |
| 前端 | React 18 + TypeScript + Ant Design |
| 数据库 | SQLite |
| 实时 | WebSocket |
| 认证 | JWT |

## 项目结构

```
├── cmd/main.go              # 入口
├── internal/
│   ├── api/                 # HTTP handlers
│   ├── services/            # 业务逻辑
│   ├── repository/          # 数据访问
│   ├── models/              # 数据模型
│   ├── metrics/             # Prometheus 指标
│   ├── supervisor/          # Supervisor XML-RPC 客户端
│   └── websocket/           # WebSocket hub
├── web/react-app/           # React 前端
│   ├── src/
│   └── dist/                # 构建产物
├── config/                  # 配置文件
├── data/                    # SQLite 数据库
└── logs/                    # 日志文件
```

## 开发

```bash
# 后端（热重载需要 air）
go run cmd/main.go

# 前端开发服务器
cd web/react-app && npm run dev

# 运行测试
go test ./...
```

## API

主要端点：

| 路径 | 说明 |
|---|---|
| `/api/auth/*` | 认证 |
| `/api/nodes/*` | 节点管理 |
| `/api/processes/*` | 进程控制 |
| `/api/discovery/*` | 节点发现 |
| `/api/users/*` | 用户管理 |
| `/api/activity-logs/*` | 活动日志 |
| `/metrics` | Prometheus 指标 |
| `/ws` | WebSocket |

## License

MIT
