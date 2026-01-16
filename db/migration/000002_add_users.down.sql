-- =====================================================
-- Migration: 000002_add_users (DOWN)
-- Description: Rollback - remove users table and foreign key
-- =====================================================

-- 先删除外键约束
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "fk_accounts_owner";

-- 再删除用户表
DROP TABLE IF EXISTS "users";
