# Go-CESI

Centralized Supervisor Interface - 多节点 Supervisor 进程管理系统

## 快速开始

```bash
# 编译
./gesi.sh build

# 启动
./gesi.sh start

# 访问
http://localhost:8081
```

默认账号：`admin` / 见 `config/.env` 中的 `ADMIN_PASSWORD`

## 管理命令

```bash
./gesi.sh build           # 编译前后端
./gesi.sh build-backend   # 仅编译后端
./gesi.sh build-frontend  # 仅编译前端
./gesi.sh start           # 后台启动
./gesi.sh stop            # 停止
./gesi.sh restart         # 重启
./gesi.sh status          # 查看状态
./gesi.sh run             # 前台运行（开发用）
```

## 配置

```
config/
├── config.toml      # 系统配置（端口、日志等）
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
- **告警系统** - 进程异常告警

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
| `/api/alerts/*` | 告警 |
| `/api/activity-logs/*` | 活动日志 |
| `/ws` | WebSocket |

## License

MIT
