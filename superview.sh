#!/bin/bash
set -euo pipefail

# Superview 管理脚本
# 用法: ./superview.sh [build|start|stop|restart|status]

readonly APP_NAME="superview"
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

    # Vite 5 requires Node >= 18
    local node_major
    node_major=$(node -v 2>/dev/null | sed 's/v\([0-9]*\).*/\1/')
    if [ -z "$node_major" ]; then
        log_error "Node.js not found. Install Node.js >= 18."
        return 1
    fi
    if [ "$node_major" -lt 18 ]; then
        log_error "Node.js >= 18 required (current: $(node -v)). Upgrade Node.js first."
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
        log_error "Binary not found. Run './superview.sh build' first."
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
        log_error "Binary not found. Run './superview.sh build' first."
        return 1
    fi
    
    mkdir -p pids logs data
    log_info "Running in foreground (Ctrl+C to stop)..."
    exec "./$APP_NAME"
}

# 构建发布包
release() {
    local version="${1:-$(date +%Y%m%d)}"
    local release_dir="release"
    local targets="${RELEASE_TARGETS:-linux-amd64}"

    # 先构建前端
    build_frontend

    rm -rf "$release_dir"
    mkdir -p "$release_dir"

    for target in $targets; do
        local goos="${target%-*}"
        local goarch="${target#*-}"
        local bin_name="$APP_NAME"
        [ "$goos" = "windows" ] && bin_name="${APP_NAME}.exe"

        local pkg_name="${APP_NAME}-${version}-${goos}-${goarch}"
        local pkg_dir="${release_dir}/${pkg_name}"

        log_info "Building ${goos}/${goarch}..."
        mkdir -p "$pkg_dir"

        CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
            go build -ldflags="-s -w" -o "${pkg_dir}/${bin_name}" cmd/main.go

        # 打包前端资源和配置模板
        cp -r "$FRONTEND_DIR/dist" "${pkg_dir}/web/react-app/dist"
        mkdir -p "${pkg_dir}/config" "${pkg_dir}/data" "${pkg_dir}/logs" "${pkg_dir}/pids"
        cp config/config.toml "${pkg_dir}/config/config.toml.example"
        cp config/nodelist.toml "${pkg_dir}/config/nodelist.toml.example"
        cp config/.env "${pkg_dir}/config/.env.example"
        cp superview.sh "${pkg_dir}/"
        cp README.md "${pkg_dir}/" 2>/dev/null || true

        # 打包
        if [ "$goos" = "windows" ]; then
            (cd "$release_dir" && zip -rq "${pkg_name}.zip" "$pkg_name")
        else
            tar -czf "${release_dir}/${pkg_name}.tar.gz" -C "$release_dir" "$pkg_name"
        fi
        rm -rf "$pkg_dir"
        log_info "Created: ${release_dir}/${pkg_name}.tar.gz"
    done

    log_info "Release packages in ${release_dir}/"
    ls -lh "$release_dir/"
}

# 帮助
show_help() {
    cat << EOF
Superview 管理脚本

用法: $0 <command>

命令:
  build           编译前后端
  build-backend   仅编译后端
  build-frontend  仅编译前端
  release [ver]   构建发布包（默认当前日期为版本号）
  start           后台启动
  stop            停止
  restart         重启
  status          查看状态
  run             前台运行（开发用）
  help            显示帮助

环境变量:
  RELEASE_TARGETS   构建目标平台（空格分隔，默认 linux-amd64）
                    可选: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

示例:
  $0 build                          # 编译
  $0 release v1.0.0                 # 构建 linux-amd64 发布包
  RELEASE_TARGETS="linux-amd64 linux-arm64" $0 release v1.0.0  # 多平台
  $0 start                          # 启动
  $0 restart                        # 重启
EOF
}

# 主入口
main() {
    cd "$(dirname "$0")"
    
    case "${1:-help}" in
        build)          build_all ;;
        build-backend)  build_backend ;;
        build-frontend) build_frontend ;;
        release)        release "${2:-}" ;;
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
