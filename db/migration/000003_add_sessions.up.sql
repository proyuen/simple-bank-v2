-- =====================================================
-- Migration: 000003_add_sessions
-- Description: Create sessions table for JWT refresh tokens
-- =====================================================

-- sessions: 会话表
-- 用于存储 refresh token，支持 token 轮换和会话管理
CREATE TABLE "sessions" (
    "id"            UUID PRIMARY KEY,                    -- 会话ID (也是refresh token的payload)
    "username"      VARCHAR(255) NOT NULL,               -- 关联的用户名
    "refresh_token" VARCHAR(512) NOT NULL,               -- Refresh Token
    "user_agent"    VARCHAR(255) NOT NULL DEFAULT '',    -- 客户端 User-Agent
    "client_ip"     VARCHAR(45) NOT NULL DEFAULT '',     -- 客户端 IP (支持 IPv6)
    "is_blocked"    BOOLEAN NOT NULL DEFAULT FALSE,      -- 是否被封禁
    "expires_at"    TIMESTAMPTZ NOT NULL,                -- 过期时间
    "created_at"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 外键约束: 关联到用户表
    CONSTRAINT "fk_sessions_user"
        FOREIGN KEY ("username")
        REFERENCES "users" ("username")
        ON DELETE CASCADE  -- 删除用户时级联删除会话
);

-- 索引: 按用户名查询会话
CREATE INDEX "idx_sessions_username" ON "sessions" ("username");

-- 索引: 清理过期会话
CREATE INDEX "idx_sessions_expires_at" ON "sessions" ("expires_at");

-- =====================================================
-- 添加表注释
COMMENT ON TABLE "sessions" IS '用户会话表(JWT refresh token)';
COMMENT ON COLUMN "sessions"."id" IS '会话ID,也是refresh token的jti';
COMMENT ON COLUMN "sessions"."is_blocked" IS '会话是否被手动封禁';
COMMENT ON COLUMN "sessions"."user_agent" IS '创建会话时的客户端标识';
COMMENT ON COLUMN "sessions"."client_ip" IS '创建会话时的客户端IP';
