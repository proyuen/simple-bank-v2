-- =====================================================
-- Migration: 000003_add_sessions
-- Description: Create sessions table for JWT refresh tokens
-- Database: MySQL 8.0+
-- =====================================================

-- sessions: 会话表
-- 用于存储 refresh token，支持 token 轮换和会话管理
CREATE TABLE `sessions` (
    `id`            CHAR(36) PRIMARY KEY COMMENT '会话ID,也是refresh token的jti (UUID格式)',
    `username`      VARCHAR(255) NOT NULL COMMENT '关联的用户名',
    `refresh_token` VARCHAR(512) NOT NULL COMMENT 'Refresh Token',
    `user_agent`    VARCHAR(255) NOT NULL DEFAULT '' COMMENT '客户端 User-Agent',
    `client_ip`     VARCHAR(45) NOT NULL DEFAULT '' COMMENT '客户端 IP (支持 IPv6)',
    `is_blocked`    BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否被封禁',
    `expires_at`    TIMESTAMP NOT NULL COMMENT '过期时间',
    `created_at`    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束: 关联到用户表
    CONSTRAINT `fk_sessions_user`
        FOREIGN KEY (`username`)
        REFERENCES `users` (`username`)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表(JWT refresh token)';

-- 索引: 按用户名查询会话
CREATE INDEX `idx_sessions_username` ON `sessions` (`username`);

-- 索引: 清理过期会话
CREATE INDEX `idx_sessions_expires_at` ON `sessions` (`expires_at`);
