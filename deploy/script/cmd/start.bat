@echo off
chcp 65001 >nul

REM IM-Zero 项目启动脚本 (Windows)

echo 🚀 启动 IM-Zero 项目...

REM 检查 Docker 是否运行
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker 未运行，请先启动 Docker
    pause
    exit /b 1
)

REM 创建必要的目录
echo 📁 创建必要的数据目录...
if not exist "data\mysql\data" mkdir "data\mysql\data"
if not exist "data\redis\data" mkdir "data\redis\data"
if not exist "data\etcd" mkdir "data\etcd"
if not exist "data\mongo" mkdir "data\mongo"
if not exist "data\nginx\log" mkdir "data\nginx\log"
if not exist "data\prometheus\data" mkdir "data\prometheus\data"
if not exist "data\server" mkdir "data\server"

REM 停止可能存在的容器
echo 🛑 停止现有容器...
docker-compose down >nul 2>&1
docker-compose -f docker-compose-env.yml down >nul 2>&1

REM 启动基础设施服务
echo 🔧 启动基础设施服务 (MySQL, Redis, Etcd, MongoDB, Prometheus)...
docker-compose -f docker-compose-env.yml up -d

REM 等待数据库启动
echo ⏳ 等待数据库启动...
timeout /t 15 /nobreak >nul

echo 🔍 检查数据库连接...
:check_mysql
docker exec mysql mysqladmin ping -h"localhost" --silent >nul 2>&1
if %errorlevel% neq 0 (
    echo 等待 MySQL 启动...
    timeout /t 2 /nobreak >nul
    goto check_mysql
)

echo ✅ 数据库已启动

REM 创建数据库（如果不存在）
echo 📊 初始化数据库...
docker exec mysql mysql -uroot -pPXDN93VRKUm8TeE7 -e "CREATE DATABASE IF NOT EXISTS im_usercenter;" >nul 2>&1

REM 创建Docker网络（如果不存在）
echo 🌐 创建Docker网络...
docker network create imzero_net >nul 2>&1

REM 启动应用容器 (使用modd进行热重载)
echo 🚀 启动应用容器...
docker-compose up -d

echo.
echo 🎉 IM-Zero 项目启动完成！
echo.
echo 📋 服务信息：
echo    • Nginx 网关: http://localhost:8888
echo    • MySQL 数据库: localhost:3308
echo    • Redis 缓存: localhost:6379
echo    • Etcd 注册中心: localhost:2379
echo    • MongoDB: localhost:27017
echo    • Mongo Express: http://localhost:8081
echo    • Prometheus: http://localhost:9090
echo.
echo 🔧 API 接口：
echo    • 用户中心: http://localhost:8888/usercenter/
echo    • 验证码服务: http://localhost:8888/verifycode/
echo.
echo 📊 直接访问应用服务：
echo    • 用户中心API: http://localhost:8080
echo    • 验证码API: http://localhost:2004
echo.
echo 📝 查看日志: docker-compose logs -f [service-name]
echo 🛑 停止服务: deploy\script\cmd\stop.bat
echo.
pause