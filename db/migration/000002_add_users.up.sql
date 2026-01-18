-- =====================================================
-- Migration: 000002_add_users
-- Description: Create users table and link to accounts
-- Database: MySQL 8.0+
-- =====================================================

-- users: 用户表
CREATE TABLE `users` (
    `id`                  BIGINT AUTO_INCREMENT PRIMARY KEY,
    `username`            VARCHAR(255) NOT NULL UNIQUE COMMENT '用户名(唯一)',
    `hashed_password`     VARCHAR(255) NOT NULL COMMENT 'bcrypt加密的密码',
    `full_name`           VARCHAR(255) NOT NULL COMMENT '真实姓名',
    `email`               VARCHAR(255) NOT NULL UNIQUE COMMENT '邮箱(唯一)',
    `password_changed_at` TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01' COMMENT '密码最后修改时间',
    `created_at`          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- =====================================================
-- 修改 accounts 表: 添加外键关联到 users
-- =====================================================

-- 添加外键约束: accounts.owner 必须是已存在的用户名
-- 注意: 这要求 accounts.owner 的值必须在 users.username 中存在
ALTER TABLE `accounts`
    ADD CONSTRAINT `fk_accounts_owner`
    FOREIGN KEY (`owner`)
    REFERENCES `users` (`username`);
