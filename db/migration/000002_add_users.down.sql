-- =====================================================
-- Migration: 000002_add_users (DOWN)
-- Description: Rollback - remove users table and foreign key
-- Database: MySQL 8.0+
-- =====================================================

-- 先删除外键约束
ALTER TABLE `accounts` DROP FOREIGN KEY `fk_accounts_owner`;

-- 再删除用户表
DROP TABLE IF EXISTS `users`;
