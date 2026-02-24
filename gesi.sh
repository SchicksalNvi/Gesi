#!/bin/bash
set -euo pipefail

# Go-CESI 管理脚本
# 用法: ./gesi.sh [build|start|stop|restart|status]

readonly APP_NAME="go-cesi"
readonly PID_FILE="pids/backend.pid"
readonly FRONTEND_DIR="web/react-app"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()    { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }

# 编译后端
build_backend() {
    log_info "Building backend..."
    go mod tidy
    go build -o "$APP_NAME" cmd/main.go
    log_info "Backend built: ./$APP_NAME"
}

# 编译前端
build_frontend() {
    if [ ! -d "$FRONTEND_DIR" ]; then
        log_warn "Frontend directory not found"
        return 1
    fi
    
    log_info "Building frontend..."
    cd "$FRONTEND_DIR"
    
    [ ! -d "node_modules" ] && npm install
    npm run build
    
    cd - > /dev/null
    log_info "Frontend built: $FRONTEND_DIR/dist/"
}

# 编译全部
build_all() {
    build_backend
    build_frontend
}

# 获取 PID
get_pid() {
    if [ -f "$PID_FILE" ]; then
        cat "$PID_FILE"
    fi
}

# 检查是否运行中
is_running() {
    local pid
    pid=$(get_pid)
    [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null
}

# 启动
start() {
    if is_running; then
        log_warn "Already running (PID: $(get_pid))"
        return 0
    fi
    
    if [ ! -f "$APP_NAME" ]; then
        log_error "Binary not found. Run './gesi.sh build' first."
        return 1
    fi
    
    mkdir -p pids logs data
    
    log_info "Starting $APP_NAME..."
    nohup "./$APP_NAME" > logs/app.log 2>&1 &
    echo $! > "$PID_FILE"
    
    sleep 1
    if is_running; then
        log_info "Started (PID: $(get_pid))"
        log_info "Access: http://localhost:8081"
    else
        log_error "Failed to start"
        return 1
    fi
}

# 停止
stop() {
    if ! is_running; then
        log_warn "Not running"
        rm -f "$PID_FILE"
        return 0
    fi
    
    local pid
    pid=$(get_pid)
    log_info "Stopping (PID: $pid)..."
    
    kill "$pid"
    
    # 等待进程退出
    for i in {1..10}; do
        if ! kill -0 "$pid" 2>/dev/null; then
            rm -f "$PID_FILE"
            log_info "Stopped"
            return 0
        fi
        sleep 1
    done
    
    # 强制终止
    kill -9 "$pid" 2>/dev/null || true
    rm -f "$PID_FILE"
    log_info "Force stopped"
}

# 重启
restart() {
    stop
    sleep 1
    start
}

# 状态
status() {
    if is_running; then
        log_info "Running (PID: $(get_pid))"
    else
        log_info "Not running"
    fi
}

# 前台运行（开发用）
run() {
    if [ ! -f "$APP_NAME" ]; then
        log_error "Binary not found. Run './gesi.sh build' first."
        return 1
    fi
    
    mkdir -p pids logs data
    log_info "Running in foreground (Ctrl+C to stop)..."
    exec "./$APP_NAME"
}

# 帮助
show_help() {
    cat << EOF
Go-CESI 管理脚本

用法: $0 <command>

命令:
  build           编译前后端
  build-backend   仅编译后端
  build-frontend  仅编译前端
  start           后台启动
  stop            停止
  restart         重启
  status          查看状态
  run             前台运行（开发用）
  help            显示帮助

示例:
  $0 build        # 编译
  $0 start        # 启动
  $0 restart      # 重启
EOF
}

# 主入口
main() {
    cd "$(dirname "$0")"
    
    case "${1:-help}" in
        build)          build_all ;;
        build-backend)  build_backend ;;
        build-frontend) build_frontend ;;
        start)          start ;;
        stop)           stop ;;
        restart)        restart ;;
        status)         status ;;
        run)            run ;;
        help|-h|--help) show_help ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
