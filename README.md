# Go-CESI

Centralized Supervisor Interface - 多节点 Supervisor 进程管理系统

## 快速开始

```bash
# 构建
./build.sh

# 启动
./start.sh

# 或一步到位
./deploy.sh
```

访问 http://localhost:5000

默认账号：`admin` / 见 `.env` 中的 `ADMIN_PASSWORD`

## 构建选项

```bash
./build.sh backend   # 只构建后端
./build.sh frontend  # 只构建前端
./build.sh all       # 全部构建（默认）
```

## 配置

- `config.toml` - 节点配置
- `.env` - 敏感信息（JWT_SECRET, 密码等）

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
  ├── react-app/         # React 前端源码
  └── static/            # 构建产物
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
