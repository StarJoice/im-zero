# IM-Zero 即时通讯系统

> 基于 Go-Zero 微服务框架开发的即时通讯系统

## 项目概述

IM-Zero 是一个现代化的即时通讯系统，采用微服务架构设计，基于 Go-Zero 框架开发。系统提供用户中心、短信验证码服务等核心功能，支持用户注册、登录、消息收发等功能。

## 技术栈

- **框架**: [Go-Zero](https://github.com/zeromicro/go-zero) v1.8.3
- **语言**: Go 1.22.1
- **数据库**: MySQL 8.0
- **缓存**: Redis 6.2
- **注册中心**: Etcd
- **网关**: Nginx
- **容器化**: Docker & Docker Compose
- **监控**: Prometheus
- **认证**: JWT
- **通信协议**: gRPC, WebSocket

## 系统架构

```
├── app/                          # 应用服务
│   ├── usercenter/              # 用户中心服务
│   │   ├── cmd/
│   │   │   ├── api/            # HTTP API 服务 (端口: 8080)
│   │   │   └── rpc/            # gRPC 服务 (端口: 8001)
│   │   └── model/              # 数据模型
│   └── verifycode/              # 短信验证码服务
│       ├── cmd/
│       │   ├── api/            # HTTP API 服务 (端口: 2004)
│       │   └── rpc/            # gRPC 服务
├── pkg/                         # 公共包
│   ├── constants/              # 系统常量
│   ├── ctxdata/                # 上下文数据
│   ├── globalkey/              # 全局常量
│   ├── middleware/             # 中间件
│   ├── sms/                    # 短信服务
│   ├── tool/                   # 工具函数
│   └── xerrs/                  # 错误处理
├── deploy/                      # 部署配置
│   ├── nginx/                  # Nginx 配置
│   ├── sql/                    # 数据库脚本
│   └── script/                 # 脚本文件
│       └── cmd/                # 启动停止脚本
└── data/                        # 数据目录
```

## 核心功能

### 用户中心 (UserCenter)
- ✅ 用户注册/登录
- ✅ 用户信息管理
- ✅ JWT 认证
- ✅ 多平台授权支持

### 短信验证码 (VerifyCode)
- ✅ 短信验证码发送
- ✅ 验证码校验
- ✅ 安全防护机制
- ✅ 频率限制

## 🚀 快速开始

### 环境要求

- Docker & Docker Compose
- Go 1.22.1+ (本地开发可选)

### ⚡ 一键启动 (推荐)

**Windows:**
```bash
# 进入项目目录
cd im-zero

# 执行启动脚本
deploy\script\cmd\start.bat
```

**Linux/macOS:**
```bash
# 进入项目目录
cd im-zero

# 给脚本添加执行权限
chmod +x deploy/script/cmd/start.sh

# 执行启动脚本
./deploy/script/cmd/start.sh
```

### 🔧 架构说明

项目采用**双层Docker架构**：

1. **基础设施层** (`docker-compose-env.yml`)
   - MySQL 8.0 (数据库)
   - Redis 6.2 (缓存)
   - Etcd (服务注册中心)
   - MongoDB (消息存储)
   - Prometheus (监控)
   - Filebeat (日志收集)

2. **应用层** (`docker-compose.yml`)
   - IM-Zero 应用容器 (使用modd热重载)
   - Nginx 网关

### 📋 手动启动

**步骤1: 启动基础设施**
```bash
# 启动 MySQL, Redis, Etcd, MongoDB, Prometheus
docker-compose -f docker-compose-env.yml up -d
```

**步骤2: 等待服务就绪**
```bash
# 检查MySQL是否启动完成
docker exec mysql mysqladmin ping -h"localhost" --silent

# 创建数据库
docker exec mysql mysql -uroot -pPXDN93VRKUm8TeE7 -e "CREATE DATABASE IF NOT EXISTS im_usercenter;"
```

**步骤3: 创建Docker网络**
```bash
# 创建应用网络
docker network create imzero_net
```

**步骤4: 启动应用服务**
```bash
# 启动应用容器 (自动使用modd热重载)
docker-compose up -d
```

### 🛠️ 本地开发

如果要在本地直接运行Go服务（不使用Docker容器）：

1. **启动基础设施**
```bash
docker-compose -f docker-compose-env.yml up -d
```

2. **等待服务启动并初始化数据库**
```bash
# 等待MySQL启动
sleep 15

# 创建数据库
docker exec mysql mysql -uroot -pPXDN93VRKUm8TeE7 -e "CREATE DATABASE IF NOT EXISTS im_usercenter;"

# 导入数据库结构
docker exec -i mysql mysql -uroot -pPXDN93VRKUm8TeE7 im_usercenter < deploy/sql/im_usercenter.sql
```

3. **启动Go服务**
```bash
# 使用modd热重载 (推荐)
modd

# 或者手动启动各个服务
go run app/usercenter/cmd/rpc/usercenter.go -f app/usercenter/cmd/rpc/etc/usercenter.yaml &
go run app/usercenter/cmd/api/usercenter.go -f app/usercenter/cmd/api/etc/usercenter.yaml &
go run app/verifycode/cmd/rpc/verifycode.go -f app/verifycode/cmd/rpc/etc/verifycode.yaml &
go run app/verifycode/cmd/api/verifycode.go -f app/verifycode/cmd/api/etc/verifycode.yaml &
```

### 🛑 停止服务

**使用脚本停止:**
```bash
# Windows
deploy\script\cmd\stop.bat

# Linux/macOS
./deploy/script/cmd/stop.sh
```

**手动停止:**
```bash
# 停止应用层
docker-compose down

# 停止基础设施层
docker-compose -f docker-compose-env.yml down
```

## 📋 服务信息

| 服务类型 | 服务名称 | 端口 | 描述 | 访问地址 |
|---------|---------|------|------|---------|
| **网关** | Nginx Gateway | 8888 | HTTP API 网关 | http://localhost:8888 |
| **应用服务** | UserCenter API | 8080 | 用户中心 HTTP 服务 | http://localhost:8080 |
| **应用服务** | UserCenter RPC | 8001 | 用户中心 gRPC 服务 | localhost:8001 |
| **应用服务** | VerifyCode API | 2004 | 验证码 HTTP 服务 | http://localhost:2004 |
| **应用服务** | VerifyCode RPC | 2001 | 验证码 gRPC 服务 | localhost:2001 |
| **基础设施** | MySQL | 3308 | 数据库服务 | localhost:3308 |
| **基础设施** | Redis | 6379 | 缓存服务 | localhost:6379 |
| **基础设施** | Etcd | 2379 | 服务注册中心 | localhost:2379 |
| **基础设施** | MongoDB | 27017 | 消息存储 | localhost:27017 |
| **管理界面** | Mongo Express | 8081 | MongoDB 管理界面 | http://localhost:8081 |
| **监控** | Prometheus | 9090 | 监控数据收集 | http://localhost:9090 |
| **监控** | UserCenter Metrics | 4008 | 用户中心监控指标 | http://localhost:4008/metrics |
| **监控** | VerifyCode Metrics | 4009 | 验证码监控指标 | http://localhost:4009/metrics |


## ⚙️ 配置说明

### 环境变量

- `TZ`: 时区设置 (默认: Asia/Shanghai)
- `GOPROXY`: Go 模块代理 (默认: https://goproxy.cn,direct)

### 数据库配置

- **用户名**: root
- **密码**: PXDN93VRKUm8TeE7
- **数据库**: im_usercenter
- **端口**: 3308

### Redis配置

- **密码**: G62m50oigInC30sf
- **端口**: 6379

## 🛠️ 开发工具

### 代码生成

```bash
# 生成 API 代码
cd deploy/script/gencode && ./gen.sh

# 生成数据模型
cd deploy/script/mysql && ./genModel.sh
```

### 热重载开发

```bash
# 安装 modd
go install github.com/cortesi/modd/cmd/modd@latest

# 启动热重载
modd
```

## 📊 监控与日志

- **Prometheus**: 系统监控和指标收集 (端口: 9090)
- **日志**: 统一日志格式，支持结构化日志
- **链路追踪**: 基于 OpenTelemetry

## 🔒 安全特性

- JWT 认证机制
- 密码 BCrypt 加密
- 短信验证码防刷机制
- IP 频率限制
- 输入参数验证
- 分布式锁防并发攻击

## 🏗️ 项目状态

### ✅ 已完成功能

- [x] 用户注册/登录
- [x] 短信验证码发送/验证
- [x] JWT 认证
- [x] 数据库模型
- [x] Docker 容器化
- [x] Nginx 网关
- [x] 基础监控

### 🚧 开发中功能

- [ ] 即时消息功能
- [ ] 好友系统
- [ ] 群组聊天
- [ ] 文件上传
- [ ] 消息推送

### 📋 待开发功能

- [ ] 管理后台
- [ ] 数据统计
- [ ] 消息存储优化
- [ ] 分布式部署
- [ ] 性能优化

## 🐛 故障排除

### 常见问题

1. **Docker 启动失败**
   - 检查 Docker 是否运行
   - 检查端口是否被占用

2. **数据库连接失败**
   - 等待 MySQL 完全启动 (约10-15秒)
   - 检查数据库密码配置

3. **服务注册失败**
   - 确认 Etcd 服务正常运行
   - 检查网络连接

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f mysql
docker-compose logs -f redis
docker-compose logs -f im-zero
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 👨‍💻 联系方式

- 作者: StarJoice
- 项目地址: [https://github.com/StarJoice/im-zero](https://github.com/StarJoice/im-zero)

---

**注意**: 这是一个积极开发中的项目，欢迎提交 Issue 和 PR 来改进项目！