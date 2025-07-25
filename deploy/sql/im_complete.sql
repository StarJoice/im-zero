-- IM系统完整数据库表结构

-- 消息服务相关表

-- 会话表 - 维护用户间的对话关系
CREATE TABLE IF NOT EXISTS `im_conversation` (
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

-- 消息表 - 存储所有消息记录
CREATE TABLE IF NOT EXISTS `im_message` (
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

-- 好友服务相关表

-- 好友关系表
CREATE TABLE IF NOT EXISTS `im_friend` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `friend_id` bigint(20) NOT NULL COMMENT '好友ID',
  `remark` varchar(50) DEFAULT NULL COMMENT '好友备注',
  `status` tinyint(4) DEFAULT 1 COMMENT '好友状态：1-正常，2-已拉黑',
  `is_top` tinyint(1) DEFAULT 0 COMMENT '是否置顶：0-否，1-是',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_friend` (`user_id`,`friend_id`),
  KEY `idx_friend_id` (`friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友关系表';

-- 好友请求表
CREATE TABLE IF NOT EXISTS `im_friend_request` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `from_user_id` bigint(20) NOT NULL COMMENT '发起者ID',
  `to_user_id` bigint(20) NOT NULL COMMENT '接收者ID',
  `message` varchar(200) DEFAULT NULL COMMENT '请求验证消息',
  `status` tinyint(4) DEFAULT 0 COMMENT '请求状态：0-待处理，1-已同意，2-已拒绝，3-已过期',
  `handle_message` varchar(200) DEFAULT NULL COMMENT '处理消息',
  `handle_time` datetime DEFAULT NULL COMMENT '处理时间',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  KEY `idx_from_user_id` (`from_user_id`),
  KEY `idx_to_user_id` (`to_user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友请求表';

-- 黑名单表
CREATE TABLE IF NOT EXISTS `im_user_blacklist` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `blocked_user_id` bigint(20) NOT NULL COMMENT '被拉黑用户ID',
  `reason` varchar(200) DEFAULT NULL COMMENT '拉黑原因',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_blocked` (`user_id`,`blocked_user_id`),
  KEY `idx_blocked_user_id` (`blocked_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户黑名单表';

-- 群组服务相关表

-- 群组表
CREATE TABLE IF NOT EXISTS `im_group` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '群组ID',
  `name` varchar(100) NOT NULL COMMENT '群名称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '群头像URL',
  `description` varchar(500) DEFAULT NULL COMMENT '群描述',
  `notice` text DEFAULT NULL COMMENT '群公告',
  `owner_id` bigint(20) NOT NULL COMMENT '群主ID',
  `member_count` int(11) DEFAULT 0 COMMENT '当前成员数量',
  `max_members` int(11) DEFAULT 500 COMMENT '最大成员数量',
  `status` tinyint(4) DEFAULT 1 COMMENT '群状态：0-禁用，1-正常，2-解散',
  `is_private` tinyint(1) DEFAULT 0 COMMENT '是否私有群：0-公开，1-私有',
  `join_approval` tinyint(1) DEFAULT 0 COMMENT '入群是否需要审批：0-不需要，1-需要',
  `allow_invite` tinyint(1) DEFAULT 1 COMMENT '是否允许邀请：0-不允许，1-允许',
  `allow_member_modify` tinyint(1) DEFAULT 0 COMMENT '是否允许成员修改群信息：0-不允许，1-允许',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  KEY `idx_owner_id` (`owner_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群组表';

-- 群成员表
CREATE TABLE IF NOT EXISTS `im_group_member` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `group_id` bigint(20) NOT NULL COMMENT '群组ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `nickname` varchar(50) DEFAULT NULL COMMENT '群内昵称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '群内头像',
  `role` tinyint(4) DEFAULT 1 COMMENT '成员角色：1-普通成员，2-管理员，3-群主',
  `status` tinyint(4) DEFAULT 1 COMMENT '成员状态：0-已退出，1-正常，2-被踢出',
  `mute_end_time` datetime DEFAULT NULL COMMENT '禁言结束时间',
  `join_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '入群时间',
  `join_source` tinyint(4) DEFAULT 1 COMMENT '入群方式：1-邀请，2-扫码，3-搜索',
  `inviter_id` bigint(20) DEFAULT NULL COMMENT '邀请人ID',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`,`user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群成员表';

-- 群消息表
CREATE TABLE IF NOT EXISTS `im_group_message` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `group_id` bigint(20) NOT NULL COMMENT '群组ID',
  `from_user_id` bigint(20) NOT NULL COMMENT '发送者ID',
  `message_type` tinyint(4) NOT NULL COMMENT '消息类型：1-文本，2-图片，3-语音，4-视频，5-文件',
  `content` text NOT NULL COMMENT '消息内容',
  `extra` text DEFAULT NULL COMMENT '额外信息(JSON格式)',
  `status` tinyint(4) DEFAULT 1 COMMENT '消息状态：0-发送中，1-已发送，2-已送达，3-已读，4-撤回，5-删除',
  `seq` bigint(20) NOT NULL COMMENT '消息序号，用于排序',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_from_user_id` (`from_user_id`),
  KEY `idx_seq` (`seq`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群消息表';

-- 群申请表
CREATE TABLE IF NOT EXISTS `im_group_request` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `group_id` bigint(20) NOT NULL COMMENT '群组ID',
  `user_id` bigint(20) NOT NULL COMMENT '申请人ID',
  `inviter_id` bigint(20) DEFAULT NULL COMMENT '邀请人ID(如果是邀请入群)',
  `request_type` tinyint(4) NOT NULL DEFAULT 1 COMMENT '请求类型：1-申请入群，2-邀请入群',
  `message` varchar(200) DEFAULT NULL COMMENT '申请消息',
  `status` tinyint(4) DEFAULT 0 COMMENT '申请状态：0-待处理，1-已同意，2-已拒绝，3-已过期',
  `handle_user_id` bigint(20) DEFAULT NULL COMMENT '处理人ID',
  `handle_message` varchar(200) DEFAULT NULL COMMENT '处理消息',
  `handle_time` datetime DEFAULT NULL COMMENT '处理时间',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群申请表';