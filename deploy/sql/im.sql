-- 用户会话表
CREATE TABLE IF NOT EXISTS `im_conversation` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `friend_id` bigint(20) NOT NULL COMMENT '好友ID',
  `last_msg_id` bigint(20) DEFAULT NULL COMMENT '最后一条消息ID',
  `unread_count` int(11) DEFAULT 0 COMMENT '未读消息数',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_friend` (`user_id`,`friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户会话表';

-- 消息表
CREATE TABLE IF NOT EXISTS `im_message` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `conversation_id` bigint(20) NOT NULL COMMENT '会话ID',
  `sender_id` bigint(20) NOT NULL COMMENT '发送者ID',
  `receiver_id` bigint(20) NOT NULL COMMENT '接收者ID',
  `content` text COMMENT '消息内容',
  `msg_type` tinyint(4) NOT NULL COMMENT '消息类型：1-文本，2-图片，3-语音，4-视频',
  `status` tinyint(4) DEFAULT 1 COMMENT '状态：1-发送中，2-已发送，3-已接收，4-已读，5-已撤回',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_conversation_id` (`conversation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';

-- 好友关系表
CREATE TABLE IF NOT EXISTS `im_friend` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `friend_id` bigint(20) NOT NULL COMMENT '好友ID',
  `remark` varchar(50) DEFAULT NULL COMMENT '备注',
  `status` tinyint(4) DEFAULT 1 COMMENT '状态：0-待验证，1-已添加，2-已拉黑',
  `del_state` tinyint(4) DEFAULT 0 COMMENT '删除状态：0-未删除，1-已删除',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_friend` (`user_id`,`friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友关系表';