/*
account: 服务应用账户表
domain_event_publish：领域发布事件表
domain_event_subscribe：领域订阅事件表
*********************************************************************
*/
CREATE TABLE IF NOT EXISTS `account` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `client_id` varchar(36) DEFAULT NULL,
  `client_secret` varchar(12) DEFAULT NULL,
  `name` varchar(36) DEFAULT NULL COMMENT '应用账户名称',
  `perm` bigint(20) DEFAULT NULL COMMENT '0:未配置 1:已配置',
  `created` bigint(20) DEFAULT NULL,
  `updated` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `domain_event_publish` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `topic` varchar(50) NOT NULL COMMENT '主题',
  `content` varchar(2000) NOT NULL COMMENT '内容',
  `status` bigint(20) NOT NULL COMMENT '0:待处理 1:处理失败',
  `created` bigint(20) NOT NULL,
  `updated` bigint(20) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `domain_event_subscribe` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `topic` varchar(50) NOT NULL,
  `status` bigint(20) NOT NULL,
  `content` varchar(2000) NOT NULL,
  `created` bigint(20) NOT NULL,
  `updated` bigint(20) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;