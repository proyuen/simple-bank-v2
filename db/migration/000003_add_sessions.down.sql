-- =====================================================
-- Migration: 000003_add_sessions (DOWN)
-- Description: Rollback - drop sessions table
-- Database: MySQL 8.0+
-- =====================================================

DROP TABLE IF EXISTS `sessions`;
