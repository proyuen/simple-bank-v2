-- =====================================================
-- Migration: 000001_init_schema (DOWN)
-- Description: Rollback - drop core banking tables
-- Database: MySQL 8.0+
-- =====================================================
-- 注意: 删除顺序与创建顺序相反(因为外键约束)

DROP TABLE IF EXISTS `transfers`;
DROP TABLE IF EXISTS `entries`;
DROP TABLE IF EXISTS `accounts`;
