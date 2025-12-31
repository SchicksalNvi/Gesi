#!/bin/bash
set -euo pipefail

# Go-CESI 统一部署脚本
# 功能：环境检查、编译、配置、启动、重置

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly APP_NAME="go-cesi"
readonly DB_PATH="data/cesi.db"
readonly CONFIG_DIR="config"
readonly FRONTEND_DIR="web/react-app"

# 颜色输出
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "$1 is not installed"
        return 1
    fi
    log_success "$1 is available"
}

# 检查环境依赖
check_environment() {
    log_info "Checking environment dependencies..."
    
    local missing_deps=0
    
    # 检查 Go
    if check_command "go"; then
        local go_version
        go_version=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Go version: $go_version"
    else
        missing_deps=$((missing_deps + 1))
    fi
    
    # 检查 Node.js (可选)
    if [ -d "$FRONTEND_DIR" ]; then
        if check_command "node"; then
            local node_version
            node_version=$(node --version)
            log_info "Node.js version: $node_version"
        else
            log_warn "Node.js not found, frontend build will be skipped"
        fi
        
        if check_command "npm"; then
            local npm_version
            npm_version=$(npm --version)
            log_info "npm version: $npm_version"
        else
            log_warn "npm not found, frontend build will be skipped"
        fi
    fi
    
    if [ $missing_deps -gt 0 ]; then
        log_error "Missing required dependencies. Please install them first."
        exit 1
    fi
    
    log_success "Environment check passed"
}

# 编译后端
build_backend() {
    log_info "Building backend..."
    
    # 下载依赖
    go mod download
    go mod tidy
    
    # 编译
    go build -o "$APP_NAME" cmd/main.go
    
    log_success "Backend built successfully"
}

# 编译前端
build_frontend() {
    if [ ! -d "$FRONTEND_DIR" ]; then
        log_warn "Frontend directory not found, skipping frontend build"
        return 0
    fi
    
    if ! command -v npm &> /dev/null; then
        log_warn "npm not found, skipping frontend build"
        return 0
    fi
    
    log_info "Building frontend..."
    
    cd "$FRONTEND_DIR"
    
    # 安装依赖
    if [ ! -d "node_modules" ]; then
        log_info "Installing frontend dependencies..."
        npm install
    fi
    
    # 构建
    npm run build
    
    cd "$SCRIPT_DIR"
    
    log_success "Frontend built successfully"
}

# 初始化配置
init_config() {
    log_info "Initializing configuration..."
    
    # 创建必要的目录
    mkdir -p data logs pids
    
    # 创建配置目录和示例文件
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        log_info "Created config directory"
    fi
    
    # 检查配置文件
    local config_needed=false
    
    if [ ! -f "$CONFIG_DIR/config.toml" ] && [ ! -f "config.toml" ]; then
        config_needed=true
    fi
    
    if [ "$config_needed" = true ]; then
        log_warn "No configuration file found"
        log_info "Available options:"
        echo "  1. Use config/ directory structure (recommended)"
        echo "  2. Use legacy config.toml format"
        echo "  3. Skip configuration (use defaults)"
        
        read -p "Choose option [1-3]: " -r choice
        
        case "$choice" in
            1)
                if [ -f "$CONFIG_DIR/config.example.toml" ]; then
                    cp "$CONFIG_DIR/config.example.toml" "$CONFIG_DIR/config.toml"
                    log_success "Created $CONFIG_DIR/config.toml from example"
                fi
                if [ -f "$CONFIG_DIR/nodelist.example.toml" ]; then
                    cp "$CONFIG_DIR/nodelist.example.toml" "$CONFIG_DIR/nodelist.toml"
                    log_success "Created $CONFIG_DIR/nodelist.toml from example"
                fi
                ;;
            2)
                if [ -f "config.example.toml" ]; then
                    cp "config.example.toml" "config.toml"
                    log_success "Created config.toml from example"
                fi
                ;;
            3)
                log_info "Using default configuration"
                ;;
            *)
                log_warn "Invalid choice, using defaults"
                ;;
        esac
    fi
    
    # 检查环境变量文件
    local env_created=false
    
    # 优先检查 config/.env
    if [ ! -f "config/.env" ]; then
        if [ -f "config/.env.example" ]; then
            log_warn "No config/.env file found"
            read -p "Create config/.env from config/.env.example? [y/N]: " -r create_env
            if [[ $create_env =~ ^[Yy]$ ]]; then
                cp "config/.env.example" "config/.env"
                log_success "Created config/.env from example"
                log_warn "Please edit config/.env file to set your secrets"
                env_created=true
            fi
        fi
    fi
    
    # 如果 config/.env 不存在，检查根目录的 .env（向后兼容）
    if [ "$env_created" = false ] && [ ! -f ".env" ] && [ -f ".env.example" ]; then
        log_warn "No .env file found"
        read -p "Create .env from .env.example? [y/N]: " -r create_env
        if [[ $create_env =~ ^[Yy]$ ]]; then
            cp ".env.example" ".env"
            log_success "Created .env from example"
            log_warn "Please edit .env file to set your secrets"
        fi
    fi
    
    log_success "Configuration initialized"
}

# 重置数据库
reset_database() {
    log_info "Resetting database..."
    
    if [ -f "$DB_PATH" ]; then
        local timestamp
        timestamp=$(date +%Y%m%d_%H%M%S)
        local backup_path="${DB_PATH}.backup_${timestamp}"
        
        log_info "Backing up database to $backup_path"
        cp "$DB_PATH" "$backup_path"
        
        rm -f "${DB_PATH}"*
        log_success "Database reset complete"
    else
        log_info "No existing database found"
    fi
}

# 启动应用
start_app() {
    log_info "Starting $APP_NAME..."
    
    if [ ! -f "$APP_NAME" ]; then
        log_error "Binary not found. Please build first."
        exit 1
    fi
    
    # 检查是否已经在运行
    if [ -f "pids/backend.pid" ]; then
        local pid
        pid=$(cat "pids/backend.pid")
        if kill -0 "$pid" 2>/dev/null; then
            log_warn "Application is already running (PID: $pid)"
            read -p "Stop and restart? [y/N]: " -r restart
            if [[ $restart =~ ^[Yy]$ ]]; then
                kill "$pid"
                sleep 2
            else
                return 0
            fi
        fi
    fi
    
    log_success "Starting application..."
    exec "./$APP_NAME"
}

# 停止应用
stop_app() {
    log_info "Stopping $APP_NAME..."
    
    if [ -f "pids/backend.pid" ]; then
        local pid
        pid=$(cat "pids/backend.pid")
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid"
            log_success "Application stopped"
        else
            log_warn "Application was not running"
        fi
        rm -f "pids/backend.pid"
    else
        log_warn "No PID file found"
    fi
}

# 显示帮助
show_help() {
    cat << EOF
Go-CESI 部署脚本

用法: $0 [COMMAND] [OPTIONS]

命令:
  deploy          完整部署 (检查环境 + 编译 + 配置 + 启动)
  build           仅编译 (后端 + 前端)
  build-backend   仅编译后端
  build-frontend  仅编译前端
  start           启动应用
  stop            停止应用
  restart         重启应用
  reset-db        重置数据库
  init-config     初始化配置
  check-env       检查环境
  help            显示此帮助

选项:
  --skip-frontend 跳过前端编译
  --reset-db      部署前重置数据库
  --force         强制执行，不询问确认

示例:
  $0 deploy                    # 完整部署
  $0 deploy --reset-db         # 部署并重置数据库
  $0 build --skip-frontend     # 仅编译后端
  $0 restart                   # 重启应用

EOF
}

# 主函数
main() {
    local command="${1:-deploy}"
    local skip_frontend=false
    local reset_db=false
    local force=false
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-frontend)
                skip_frontend=true
                shift
                ;;
            --reset-db)
                reset_db=true
                shift
                ;;
            --force)
                force=true
                shift
                ;;
            -h|--help|help)
                show_help
                exit 0
                ;;
            *)
                if [ -z "${command_set:-}" ]; then
                    command="$1"
                    command_set=true
                fi
                shift
                ;;
        esac
    done
    
    cd "$SCRIPT_DIR"
    
    case "$command" in
        deploy)
            log_info "Starting full deployment..."
            check_environment
            
            if [ "$reset_db" = true ]; then
                reset_database
            fi
            
            build_backend
            
            if [ "$skip_frontend" = false ]; then
                build_frontend
            fi
            
            init_config
            start_app
            ;;
        build)
            build_backend
            if [ "$skip_frontend" = false ]; then
                build_frontend
            fi
            ;;
        build-backend)
            build_backend
            ;;
        build-frontend)
            build_frontend
            ;;
        start)
            start_app
            ;;
        stop)
            stop_app
            ;;
        restart)
            stop_app
            sleep 1
            start_app
            ;;
        reset-db)
            if [ "$force" = false ]; then
                read -p "Are you sure you want to reset the database? [y/N]: " -r confirm
                if [[ ! $confirm =~ ^[Yy]$ ]]; then
                    log_info "Database reset cancelled"
                    exit 0
                fi
            fi
            reset_database
            ;;
        init-config)
            init_config
            ;;
        check-env)
            check_environment
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"