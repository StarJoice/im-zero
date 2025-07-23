@echo off
chcp 65001 >nul

REM IM-Zero 项目停止脚本 (Windows)

echo 🛑 停止 IM-Zero 项目...

REM 停止应用容器
echo ⏹️ 停止应用容器...
docker-compose down

REM 停止基础设施服务
echo ⏹️ 停止基础设施服务...
docker-compose -f docker-compose-env.yml down

echo ✅ 所有服务已停止

REM 可选：清理资源
echo.
set /p cleanup="🗑️ 是否需要清理数据卷？(y/N): "
if /i "%cleanup%"=="y" (
    echo 🧹 清理数据卷...
    docker system prune -f
    echo ✅ 清理完成
)

echo.
pause