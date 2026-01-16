-- =====================================================
-- Migration: 000003_add_sessions (DOWN)
-- Description: Rollback - drop sessions table
-- =====================================================

DROP TABLE IF EXISTS "sessions";
