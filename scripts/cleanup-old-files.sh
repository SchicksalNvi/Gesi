#!/bin/bash
# 清理旧的静态文件和模板
# 验证需求：12.1, 12.2, 12.3

set -e

echo "🧹 Cleaning up old files..."

# 备份目录
BACKUP_DIR="backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# 1. 备份并删除旧的模板文件
if [ -d "web/templates" ]; then
    echo "📦 Backing up web/templates to $BACKUP_DIR/templates"
    cp -r web/templates "$BACKUP_DIR/"
    echo "🗑️  Removing web/templates"
    rm -rf web/templates
fi

# 2. 备份并删除旧的静态 JS 文件
if [ -d "web/static/js" ]; then
    echo "📦 Backing up web/static/js to $BACKUP_DIR/static_js"
    cp -r web/static/js "$BACKUP_DIR/static_js"
    echo "🗑️  Removing web/static/js"
    rm -rf web/static/js
fi

# 3. 检查并删除未使用的 CSS 文件
if [ -f "web/static/css/style.css" ]; then
    # 检查是否在代码中被引用
    if ! grep -r "style.css" web/react-app/src/ > /dev/null 2>&1; then
        echo "📦 Backing up web/static/css/style.css to $BACKUP_DIR/"
        cp web/static/css/style.css "$BACKUP_DIR/"
        echo "🗑️  Removing unused web/static/css/style.css"
        rm -f web/static/css/style.css
    else
        echo "ℹ️  web/static/css/style.css is still in use, keeping it"
    fi
fi

echo "✅ Cleanup complete!"
echo "📁 Backup saved to: $BACKUP_DIR"
echo ""
echo "⚠️  Note: You may need to update cmd/main.go to remove references to old templates"
echo "   Look for routes like '/legacy' and template loading code"
