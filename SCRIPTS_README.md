# Go-CESI 启停脚本使用说明

本文档介绍如何使用 Go-CESI 项目的启停管理脚本。

## 脚本文件

### 1. `manage.sh` - 主管理脚本

功能完整的服务管理脚本，支持启动、停止、重启、状态查看、日志查看等功能。

### 2. `start.sh` - 快速启动脚本

简化版启动脚本，用于快速启动所有服务。

## 服务配置

- **后端服务**: Go-CESI 后端 API 服务
  - 端口: 8081
  - 二进制文件: `go-cesi`
  - 配置文件: `config.toml`

- **前端服务**: React 前端应用
  - 端口: 3002
  - 构建目录: `web/react-app/build`
  - 使用 Python HTTP 服务器提供静态文件

## 使用方法

### 快速启动

```bash
# 快速启动所有服务
./start.sh
```

### 详细管理

#### 启动服务

```bash
# 启动所有服务
./manage.sh start
./manage.sh start all

# 仅启动后端服务
./manage.sh start backend

# 仅启动前端服务
./manage.sh start frontend
```

#### 停止服务

```bash
# 停止所有服务
./manage.sh stop
./manage.sh stop all

# 仅停止后端服务
./manage.sh stop backend

# 仅停止前端服务
./manage.sh stop frontend
```

#### 重启服务

```bash
# 重启所有服务
./manage.sh restart
./manage.sh restart all

# 仅重启后端服务
./manage.sh restart backend

# 仅重启前端服务
./manage.sh restart frontend
```

#### 查看状态

```bash
# 查看所有服务状态
./manage.sh status
```

#### 查看日志

```bash
# 查看所有服务日志（最近50行）
./manage.sh logs all

# 查看后端日志（最近50行）
./manage.sh logs backend

# 查看前端日志（最近50行）
./manage.sh logs frontend

# 查看指定行数的日志
./manage.sh logs backend 100
./manage.sh logs frontend 200
```

#### 清理

```bash
# 清理PID文件
./manage.sh cleanup

# 清理PID文件和日志文件
./manage.sh cleanup --logs
```

#### 帮助信息

```bash
# 显示帮助信息
./manage.sh help
```

## 脚本特性

### 1. 进程管理

- **PID文件管理**: 使用PID文件跟踪进程状态
- **优雅停止**: 先尝试TERM信号，再使用KILL信号
- **进程检查**: 启动前检查是否已有进程运行

### 2. 端口管理

- **端口占用检查**: 启动前检查端口是否被占用
- **强制清理**: 自动杀死占用端口的进程
- **端口状态显示**: 显示当前端口占用情况

### 3. 自动化功能

- **依赖检查**: 自动检查并安装前端依赖
- **构建检查**: 自动构建前端应用（如果需要）
- **编译检查**: 自动编译Go后端（如果需要）

### 4. 日志管理

- **分离日志**: 前后端日志分别存储
- **实时日志**: 支持查看指定行数的日志
- **日志清理**: 支持清理历史日志文件

### 5. 错误处理

- **启动检查**: 启动后验证服务是否正常运行
- **错误提示**: 详细的错误信息和解决建议
- **回滚机制**: 启动失败时自动清理

## 文件结构

```
go-cesi/
├── manage.sh           # 主管理脚本
├── start.sh            # 快速启动脚本
├── pids/               # PID文件目录
│   ├── backend.pid     # 后端进程PID
│   └── frontend.pid    # 前端进程PID
├── logs/               # 日志文件目录
│   ├── backend.log     # 后端日志
│   └── frontend.log    # 前端日志
├── go-cesi             # Go后端二进制文件
├── config.toml         # 后端配置文件
└── web/react-app/      # 前端应用目录
    ├── build/          # 前端构建目录
    └── ...
```

## 常见问题

### 1. 端口被占用

**问题**: 启动时提示端口已被占用

**解决**: 脚本会自动杀死占用端口的进程，或手动执行：
```bash
# 查看端口占用
lsof -i :8081
lsof -i :3002

# 杀死进程
kill -9 <PID>
```

### 2. 前端构建失败

**问题**: 前端依赖安装或构建失败

**解决**: 
```bash
# 手动安装依赖
cd web/react-app
npm install

# 手动构建
npm run build
```

### 3. 后端编译失败

**问题**: Go后端编译失败

**解决**:
```bash
# 检查Go环境
go version

# 手动编译
go build -o go-cesi cmd/main.go
```

### 4. 服务启动失败

**问题**: 服务启动后立即退出

**解决**:
```bash
# 查看详细日志
./manage.sh logs backend 100
./manage.sh logs frontend 100

# 检查配置文件
cat config.toml
```

## 访问地址

启动成功后，可以通过以下地址访问：

- **前端应用**: http://localhost:3002
- **后端API**: http://localhost:8081
- **API文档**: http://localhost:8081/api/docs（如果有）

## 注意事项

1. **权限要求**: 脚本需要执行权限，首次使用时会自动添加
2. **依赖要求**: 需要安装 Node.js、npm、Go、Python3
3. **网络要求**: 前端构建可能需要网络连接下载依赖
4. **磁盘空间**: 确保有足够空间存储日志和构建文件
5. **防火墙**: 确保端口 8081 和 3002 未被防火墙阻止

## 开发建议

1. **开发模式**: 开发时建议分别启动前后端，便于调试
2. **生产模式**: 生产环境建议使用 `start.sh` 快速启动
3. **监控**: 定期检查服务状态和日志
4. **备份**: 重要配置文件建议定期备份

## 脚本维护

如需修改脚本配置，可以编辑以下变量：

```bash
# 端口配置
BACKEND_PORT=8081
FRONTEND_PORT=3002

# 路径配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REACT_APP_PATH="$PROJECT_ROOT/web/react-app"
GO_BACKEND_PATH="$PROJECT_ROOT"
```

更多问题请参考项目文档或联系开发团队。