#!/bin/bash

# IM-Zero 项目启动脚本

echo "🚀 启动 IM-Zero 项目..."

# 检查 Docker 是否运行
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker 未运行，请先启动 Docker"
    exit 1
fi

# 创建必要的目录
echo "📁 创建必要的数据目录..."
mkdir -p data/mysql/data
mkdir -p data/redis/data
mkdir -p data/etcd
mkdir -p data/mongo
mkdir -p data/nginx/log
mkdir -p data/prometheus/data
mkdir -p data/server

# 停止可能存在的容器
echo "🛑 停止现有容器..."
docker-compose down 2>/dev/null || true
docker-compose -f docker-compose-env.yml down 2>/dev/null || true

# 启动基础设施服务
echo "🔧 启动基础设施服务 (MySQL, Redis, Etcd, MongoDB, Prometheus)..."
docker-compose -f docker-compose-env.yml up -d

# 等待数据库启动
echo "⏳ 等待数据库启动..."
sleep 15

# 检查数据库连接
echo "🔍 检查数据库连接..."
until docker exec mysql mysqladmin ping -h"localhost" --silent; do
    echo "等待 MySQL 启动..."
    sleep 2
done

echo "✅ 数据库已启动"

# 创建数据库（如果不存在）
echo "📊 初始化数据库..."
docker exec mysql mysql -uroot -pPXDN93VRKUm8TeE7 -e "CREATE DATABASE IF NOT EXISTS im_usercenter;" 2>/dev/null || true

# 创建Docker网络（如果不存在）
echo "🌐 创建Docker网络..."
docker network create imzero_net 2>/dev/null || true

# 启动应用容器 (使用modd进行热重载)
echo "🚀 启动应用容器..."
docker-compose up -d

echo "🎉 IM-Zero 项目启动完成！"
echo ""
echo "📋 服务信息："
echo "   • Nginx 网关: http://localhost:8888"
echo "   • MySQL 数据库: localhost:3308"
echo "   • Redis 缓存: localhost:6379"
echo "   • Etcd 注册中心: localhost:2379"
echo "   • MongoDB: localhost:27017"
echo "   • Mongo Express: http://localhost:8081"
echo "   • Prometheus: http://localhost:9090"
echo ""
echo "🔧 API 接口："
echo "   • 用户中心: http://localhost:8888/usercenter/"
echo "   • 验证码服务: http://localhost:8888/verifycode/"
echo ""
echo "📊 直接访问应用服务："
echo "   • 用户中心API: http://localhost:8080"
echo "   • 验证码API: http://localhost:2004"
echo ""
echo "📝 查看日志: docker-compose logs -f [service-name]"
echo "🛑 停止服务: ./deploy/script/cmd/stop.sh"