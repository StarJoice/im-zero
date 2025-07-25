# Message 模块架构文档

## 概述

Message 模块是 IM-Zero 即时通讯系统的核心组件，负责处理消息的发送、接收、存储、查询等核心业务逻辑。采用微服务架构，包含 API 服务和 RPC 服务，支持实时消息推送和离线消息处理。

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Message 模块架构                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │
│  │   Client    │◄──►│ API Service │◄──►│ RPC Service │      │
│  │  (WebSocket)│    │             │    │             │      │
│  └─────────────┘    └─────────────┘    └─────────────┘      │
│         │                   │                   │           │
│         │            ┌─────────────┐    ┌─────────────┐      │
│         └───────────►│ WebSocket   │    │   MySQL     │      │
│                      │  Manager    │    │ Database    │      │
│                      └─────────────┘    └─────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### 模块结构

```
app/message/
├── cmd/
│   ├── api/                    # API 服务
│   │   ├── internal/
│   │   │   ├── config/         # 配置
│   │   │   ├── handler/        # HTTP 处理器
│   │   │   │   ├── message/    # 消息相关处理器
│   │   │   │   └── ws/         # WebSocket 处理器
│   │   │   ├── logic/          # 业务逻辑
│   │   │   │   ├── message/    # 消息业务逻辑
│   │   │   │   └── ws/         # WebSocket业务逻辑
│   │   │   ├── svc/            # 服务上下文
│   │   │   └── types/          # 类型定义
│   │   └── message.go          # API 服务入口
│   └── rpc/                    # RPC 服务
│       ├── internal/
│       │   ├── config/         # 配置
│       │   ├── logic/          # RPC 业务逻辑
│       │   ├── server/         # gRPC 服务器
│       │   └── svc/            # 服务上下文
│       ├── message/            # gRPC 生成代码
│       ├── messageClient/      # RPC 客户端
│       └── message.go          # RPC 服务入口
├── model/                      # 数据模型
└── README.md                   # 本文档
```

## 核心功能

### 1. 消息发送 (SendMessage)

**流程图:**
```
客户端发送请求 → API验证 → RPC处理 → 数据库事务 → WebSocket推送 → 返回结果
      ↓              ↓         ↓         ↓           ↓         ↓
   HTTP POST → JWT验证 → 好友验证 → 消息入库 → 实时推送 → 响应客户端
                                  ↓
                              会话更新
```

**关键特性:**
- 好友关系验证
- 数据库事务保证一致性
- 实时 WebSocket 推送
- 会话信息自动维护
- 未读计数管理

### 2. 消息接收与推送

**WebSocket 连接管理:**
```go
type Hub struct {
    connections map[int64][]*Connection  // 用户ID -> 连接列表
    register    chan *Connection         // 注册连接
    unregister  chan *Connection         // 注销连接  
    broadcast   chan *BroadcastMessage   // 广播消息
    mutex       sync.RWMutex            // 并发安全
}
```

**推送流程:**
1. 检查用户在线状态
2. 在线用户直接推送，更新消息状态为"已送达"
3. 离线用户消息保持"已发送"状态，待上线后获取

### 3. 聊天记录查询 (GetChatHistory)

**分页查询:**
- 支持基于消息ID的分页
- 按时间倒序查询，返回时正序排列
- 自动过滤已删除消息
- 撤回消息显示特殊内容

### 4. 消息状态管理

**状态流转:**
```
发送中(0) → 已发送(1) → 已送达(2) → 已读(3)
              ↓
          撤回(4) / 删除(5)
```

**已读标记:**
- 支持单条消息标记
- 支持批量标记到指定消息
- 自动清零会话未读计数

### 5. 消息撤回 (RecallMessage)

**限制条件:**
- 只有发送者可以撤回
- 2分钟内可撤回
- 已撤回/已删除消息不能再次操作

**撤回后处理:**
- 更新消息状态为"撤回"
- 如果是最后一条消息，更新会话信息

### 6. 消息删除 (DeleteMessage)

**删除特性:**
- 软删除机制，不真正删除数据
- 发送者和接收者都可删除
- 删除后自动更新会话最后消息

### 7. 会话管理 (Conversations)

**会话功能:**
- 自动创建用户对话会话
- 维护最后消息信息
- 未读消息计数
- 会话列表分页查询

## 数据库设计

### 消息表 (im_message)

```sql
CREATE TABLE `im_message` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `from_user_id` bigint(20) NOT NULL COMMENT '发送者ID',
  `to_user_id` bigint(20) NOT NULL COMMENT '接收者ID', 
  `conversation_id` varchar(64) NOT NULL COMMENT '会话ID',
  `message_type` tinyint(4) NOT NULL COMMENT '消息类型：1-文本，2-图片，3-语音，4-视频，5-文件',
  `content` text NOT NULL COMMENT '消息内容',
  `extra` text DEFAULT NULL COMMENT '额外信息(JSON格式)',
  `status` tinyint(4) DEFAULT 0 COMMENT '消息状态：0-发送中，1-已发送，2-已送达，3-已读，4-撤回，5-删除',
  `seq` bigint(20) NOT NULL COMMENT '消息序号，用于排序',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  KEY `idx_conversation_id` (`conversation_id`),
  KEY `idx_from_user_id` (`from_user_id`),
  KEY `idx_to_user_id` (`to_user_id`),
  KEY `idx_seq` (`seq`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';
```

### 会话表 (im_conversation)

```sql
CREATE TABLE `im_conversation` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '会话ID',  
  `conversation_id` varchar(64) NOT NULL COMMENT '会话唯一标识',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `friend_id` bigint(20) NOT NULL COMMENT '对方用户ID',
  `conversation_type` tinyint(4) NOT NULL DEFAULT 1 COMMENT '会话类型：1-单聊，2-群聊',
  `last_message_id` bigint(20) DEFAULT NULL COMMENT '最后一条消息ID',
  `last_message_content` text DEFAULT NULL COMMENT '最后一条消息内容',
  `last_message_time` datetime DEFAULT NULL COMMENT '最后一条消息时间',
  `unread_count` int(11) DEFAULT 0 COMMENT '未读消息数',
  `is_top` tinyint(1) DEFAULT 0 COMMENT '是否置顶：0-否，1-是',
  `is_mute` tinyint(1) DEFAULT 0 COMMENT '是否免打扰：0-否，1-是',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_conversation_id` (`conversation_id`),
  UNIQUE KEY `uk_user_friend` (`user_id`,`friend_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话表';
```

## API 接口

### HTTP API (端口: 8080)

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 发送消息 | POST | /message/v1/send | 发送消息 |
| 获取聊天记录 | POST | /message/v1/history | 获取聊天记录 |
| 获取会话列表 | POST | /message/v1/conversations | 获取会话列表 |
| 标记已读 | POST | /message/v1/read | 标记消息已读 |
| 撤回消息 | POST | /message/v1/recall | 撤回消息 |
| 删除消息 | POST | /message/v1/delete | 删除消息 |
| WebSocket连接 | GET | /message/v1/ws | WebSocket连接 |

### RPC 接口 (端口: 8081)

| 方法 | 说明 |
|------|------|
| SendMessage | 发送消息 |
| GetChatHistory | 获取聊天记录 |
| GetConversations | 获取会话列表 |
| MarkAsRead | 标记已读 |
| RecallMessage | 撤回消息 |
| DeleteMessage | 删除消息 |
| PushMessage | 推送消息 |
| GetUnreadCount | 获取未读计数 |

## WebSocket 实时通信

### 连接建立
```
GET ws://localhost:8080/message/v1/ws?user_id=123
Upgrade: websocket
```

### 消息格式
```json
{
  "type": "new_message",
  "content": {
    "id": 12345,
    "from_user_id": 1,
    "to_user_id": 2,
    "conversation_id": "1_2",
    "message_type": 1,
    "content": "Hello World",
    "create_time": 1640995200
  }
}
```

### 支持的消息类型
- `new_message`: 新消息推送
- `ping/pong`: 心跳检测  
- `typing`: 正在输入状态
- `read_receipt`: 已读回执

## 技术特性

### 1. 高并发支持
- WebSocket 连接池管理
- 读写锁保证并发安全
- 连接自动清理机制

### 2. 数据一致性
- 数据库事务保证
- 乐观锁版本控制
- 消息状态严格管理

### 3. 容错处理
- 推送失败不影响消息发送
- 离线消息延迟推送
- 连接断开自动重连

### 4. 性能优化
- 分页查询避免大量数据加载
- 索引优化提升查询性能
- 缓存机制减少数据库访问

## 部署说明

### 环境要求
- Go 1.19+
- MySQL 8.0+
- Redis (可选，用于缓存)

### 配置文件

**API 服务配置 (etc/message.yaml):**
```yaml
Name: message-api
Host: 0.0.0.0
Port: 8080

JwtAuth:
  AccessSecret: your-access-secret
  AccessExpire: 86400

MessageRpc:
  Endpoints:
    - 127.0.0.1:8081
```

**RPC 服务配置 (etc/message.yaml):**
```yaml
Name: message-rpc
ListenOn: 0.0.0.0:8081

DB:
  DataSource: root:password@tcp(localhost:3306)/im_zero?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai

Cache:
  - Host: localhost:6379

UsercenterRpc:
  Endpoints:
    - 127.0.0.1:8082

FriendRpc:
  Endpoints:
    - 127.0.0.1:8083
```

### 启动服务

```bash
# 启动 RPC 服务
cd app/message/cmd/rpc
go run message.go -f etc/message.yaml

# 启动 API 服务  
cd app/message/cmd/api
go run message.go -f etc/message.yaml
```

## 监控与日志

### 关键指标
- 消息发送成功率
- WebSocket 连接数
- 消息推送延迟
- 数据库查询性能

### 日志说明
- 消息发送记录
- WebSocket 连接状态
- 错误异常日志
- 性能监控日志

## 扩展说明

### 支持群聊
- 扩展会话类型为群聊 (conversation_type=2)
- 消息推送支持多人广播
- 群成员管理集成

### 消息类型扩展
- 图片消息 (message_type=2)
- 语音消息 (message_type=3)  
- 视频消息 (message_type=4)
- 文件消息 (message_type=5)

### 高可用部署
- 多实例负载均衡
- 数据库主从复制
- WebSocket 连接迁移

---

*本文档描述了 Message 模块的完整架构设计与实现细节，为系统维护和功能扩展提供参考。*