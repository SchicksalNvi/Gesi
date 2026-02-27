# Superview

Centralized Supervisor Interface - 多节点 Supervisor 进程管理系统

## 快速部署

下载发布包，解压即用，无需安装 Go 或 Node.js：

```bash
tar -xzf superview-v2.0.0-linux-amd64.tar.gz
cd superview-v2.0.0-linux-amd64

# 复制并修改配置
cp config/config.toml.example config/config.toml
cp config/nodelist.toml.example config/nodelist.toml
cp config/.env.example config/.env
# 编辑以上配置文件...

# 启动
./superview.sh start

# 访问
# http://localhost:8081
```

默认账号：`admin` / 见 `config/.env` 中的 `ADMIN_PASSWORD`

## 运维命令

```bash
./superview.sh start     # 后台启动
./superview.sh stop      # 停止
./superview.sh restart   # 重启
./superview.sh status    # 查看状态
./superview.sh run       # 前台运行（调试用）
```

## 从源码构建

### 环境要求

- Go >= 1.22
- Node.js >= 18（见 `web/react-app/.nvmrc`）
- npm >= 8

### 构建命令

```bash
make                # 构建前后端
make frontend       # 仅构建前端
make backend        # 仅构建后端
make clean          # 清理构建产物
```

### 打包发布

```bash
# 默认 linux/amd64
make release VERSION=v2.0.0

# 交叉编译
make release VERSION=v2.0.0 GOOS=linux GOARCH=arm64
make release VERSION=v2.0.0 GOOS=darwin GOARCH=arm64
```

产物在 `release/` 目录，格式为 `superview-{version}-{os}-{arch}.tar.gz`。

## 配置

```
config/
├── config.toml      # 系统配置（端口、日志、metrics 等）
├── nodelist.toml    # Supervisor 节点列表
└── .env             # 敏感配置（JWT_SECRET、密码等）
```

### 环境变量 (config/.env)

```bash
JWT_SECRET=your-secret-key       # JWT 签名密钥
ADMIN_PASSWORD=admin123          # 管理员密码
ADMIN_EMAIL=admin@example.com    # 管理员邮箱
NODE_PASSWORD=supervisor-password # 节点连接密码
```

## 功能

- 多 Supervisor 节点集中管理
- 进程启动/停止/重启控制
- WebSocket 实时状态推送
- CIDR 网段扫描自动发现节点
- 基于角色的访问控制 (RBAC)
- 操作审计日志
- Prometheus 监控指标

## Prometheus 监控

在 `config/config.toml` 中启用：

```toml
[metrics]
enabled = true
path = "/metrics"
username = "prometheus"  # 可选 Basic Auth
password = "secret"
```

Prometheus 采集配置：

```yaml
scrape_configs:
  - job_name: 'superview'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: /metrics
    basic_auth:
      username: prometheus
      password: secret
```

### 指标列表

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `superview_node_up` | gauge | 节点连接状态 (1=在线, 0=离线) |
| `superview_process_up` | gauge | 进程运行状态 (1=运行, 0=停止) |
| `superview_process_state` | gauge | 进程状态码 |
| `superview_process_uptime_seconds` | gauge | 进程运行时长 |
| `superview_nodes_total` | gauge | 节点总数 |
| `superview_nodes_connected` | gauge | 已连接节点数 |
| `superview_processes_total` | gauge | 进程总数 |
| `superview_processes_running` | gauge | 运行中进程数 |

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
      - alert: ProcessFailed
        expr: superview_process_state == 200
        for: 0s
        labels:
          severity: critical
```

## 开发

```bash
# 后端
go run cmd/main.go

# 前端开发服务器
cd web/react-app && npm run dev

# 测试
go test ./...
```

## 项目结构

```
├── cmd/main.go              # 入口
├── internal/
│   ├── api/                 # HTTP handlers
│   ├── services/            # 业务逻辑
│   ├── repository/          # 数据访问
│   ├── models/              # 数据模型
│   ├── metrics/             # Prometheus 指标
│   ├── supervisor/          # Supervisor XML-RPC
│   └── websocket/           # WebSocket hub
├── web/react-app/           # React 前端
├── config/                  # 配置文件
├── Makefile                 # 构建入口
└── superview.sh             # 运维脚本
```

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go + Gin + GORM + SQLite |
| 前端 | React 18 + TypeScript + Ant Design |
| 实时 | WebSocket |
| 认证 | JWT |

## API

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
