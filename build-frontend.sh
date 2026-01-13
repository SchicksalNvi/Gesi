#!/bin/bash
set -euo pipefail

# 构建前端脚本
echo "🔨 Building React frontend..."

# 进入前端目录
cd web/react-app

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

# 构建前端
echo "🏗️  Building production bundle..."
npm run build

# 确保目标目录存在
mkdir -p dist

# 复制构建文件到正确位置（如果不是同一目录）
if [ -d "dist" ] && [ "$(pwd)" != "$(realpath ../../react-app)" ]; then
    echo "📁 Copying build files..."
    cp -r dist/* ./
fi

echo "✅ Frontend build completed!"
echo "💡 Restart the Go server to serve the new frontend files."