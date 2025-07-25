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