#!/bin/bash

# IM-Zero 项目停止脚本

echo "🛑 停止 IM-Zero 项目..."

# 停止应用容器
echo "⏹️ 停止应用容器..."
docker-compose down

# 停止基础设施服务
echo "⏹️ 停止基础设施服务..."
docker-compose -f docker-compose-env.yml down

echo "✅ 所有服务已停止"

# 可选：清理资源
read -p "🗑️ 是否需要清理数据卷？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 清理数据卷..."
    docker system prune -f
    echo "✅ 清理完成"
fi