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