-- =====================================================
-- Migration: 000002_add_users
-- Description: Create users table and link to accounts
-- =====================================================

-- users: 用户表
CREATE TABLE "users" (
    "id"                  BIGSERIAL PRIMARY KEY,
    "username"            VARCHAR(255) NOT NULL UNIQUE,  -- 用户名(唯一)
    "hashed_password"     VARCHAR(255) NOT NULL,         -- 加密后的密码(bcrypt)
    "full_name"           VARCHAR(255) NOT NULL,         -- 真实姓名
    "email"               VARCHAR(255) NOT NULL UNIQUE,  -- 邮箱(唯一)
    "password_changed_at" TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00Z', -- 密码修改时间
    "created_at"          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =====================================================
-- 修改 accounts 表: 添加外键关联到 users
-- =====================================================

-- 添加外键约束: accounts.owner 必须是已存在的用户名
-- 注意: 这要求 accounts.owner 的值必须在 users.username 中存在
ALTER TABLE "accounts"
    ADD CONSTRAINT "fk_accounts_owner"
    FOREIGN KEY ("owner")
    REFERENCES "users" ("username");

-- =====================================================
-- 添加表注释
COMMENT ON TABLE "users" IS '用户表';
COMMENT ON COLUMN "users"."hashed_password" IS 'bcrypt加密的密码';
COMMENT ON COLUMN "users"."password_changed_at" IS '密码最后修改时间';
