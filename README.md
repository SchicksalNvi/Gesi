# Go-CESI

Centralized Supervisor Interface - 多节点 Supervisor 进程管理系统

## 快速开始

### 一键部署（推荐）

```bash
# 完整部署：环境检查 + 编译 + 配置 + 启动
./deploy.sh deploy

# 部署并重置数据库
./deploy.sh deploy --reset-db

# 仅编译后端（跳过前端）
./deploy.sh deploy --skip-frontend
```

### 分步操作

```bash
# 检查环境
./deploy.sh check-env

# 编译
./deploy.sh build              # 编译前后端
./deploy.sh build-backend      # 仅编译后端
./deploy.sh build-frontend     # 仅编译前端

# 启动/停止
./deploy.sh start              # 启动应用
./deploy.sh stop               # 停止应用
./deploy.sh restart            # 重启应用

# 配置和数据库
./deploy.sh init-config        # 初始化配置
./deploy.sh reset-db           # 重置数据库
```

访问 http://localhost:8081

默认账号：`admin` / 见 `.env` 中的 `ADMIN_PASSWORD`

## 配置

### 新配置结构（推荐）

```
config/
├── config.toml          # 系统配置
├── nodelist.toml        # 节点配置
├── .env                 # 环境变量
└── .env.example         # 环境变量示例
```

### 传统配置（向后兼容）

```
config.toml              # 包含所有配置
.env                     # 环境变量（根目录）
```

### 环境变量

推荐将环境变量文件放在 `config/.env`，应用会自动检测：
1. 优先加载 `config/.env`
2. 回退到根目录 `.env`（向后兼容）

部署脚本会自动检测配置格式并提供迁移建议。

## 架构

```
cmd/main.go              # 入口
internal/
  ├── api/               # HTTP handlers
  ├── services/          # 业务逻辑
  ├── repository/        # 数据访问
  ├── models/            # 数据模型
  └── supervisor/        # Supervisor 集成
web/
  └── react-app/         # React 前端源码
      └── dist/          # 构建产物
```

## 功能

- 多节点管理
- 进程控制（启动/停止/重启）
- 实时状态监控（WebSocket）
- 用户管理（RBAC）
- 活动日志
- 告警系统

## 技术栈

- 后端：Go + Gin + GORM + SQLite
- 前端：React + TypeScript + Ant Design
- 实时：WebSocket
- 认证：JWT

## 开发

```bash
# 后端
go run cmd/main.go

# 前端（开发模式）
cd web/react-app
npm run dev
```

## 数据库

```bash
# 重置数据库
./reset-db.sh
```

## License

MIT
