#!/bin/bash

# Go-CESI 快速启动脚本
# 这是一个简化版本的启动脚本，用于快速启动前后端服务

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 项目路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANAGE_SCRIPT="$SCRIPT_DIR/manage.sh"

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查manage.sh是否存在
if [ ! -f "$MANAGE_SCRIPT" ]; then
    print_error "管理脚本不存在: $MANAGE_SCRIPT"
    exit 1
fi

# 检查manage.sh是否可执行
if [ ! -x "$MANAGE_SCRIPT" ]; then
    print_warning "管理脚本不可执行，正在添加执行权限..."
    chmod +x "$MANAGE_SCRIPT"
fi

print_info "Go-CESI 快速启动脚本"
print_info "==================="

# 显示当前状态
print_info "检查当前服务状态..."
"$MANAGE_SCRIPT" status

echo
print_info "启动所有服务..."
"$MANAGE_SCRIPT" start all

echo
print_info "最终状态检查..."
"$MANAGE_SCRIPT" status

echo
print_info "启动完成！"
print_info "前端访问地址: http://localhost:3002"
print_info "后端API地址: http://localhost:8081"
print_info ""
print_info "使用以下命令管理服务:"
print_info "  ./manage.sh status    # 查看状态"
print_info "  ./manage.sh stop all  # 停止所有服务"
print_info "  ./manage.sh logs all  # 查看日志"
print_info "  ./manage.sh help      # 查看完整帮助"