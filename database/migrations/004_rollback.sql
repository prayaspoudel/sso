-- Rollback Migration 004: Phase 2 - Enhanced Authentication
-- This script rolls back all changes from migration 004

-- Drop cleanup functions
DROP FUNCTION IF EXISTS cleanup_old_backup_codes();
DROP FUNCTION IF EXISTS cleanup_expired_oauth2_tokens();
DROP FUNCTION IF EXISTS cleanup_expired_oauth2_codes();

-- Drop triggers
DROP TRIGGER IF EXISTS update_oauth2_clients_updated_at ON oauth2_clients;
DROP TRIGGER IF EXISTS update_user_two_factor_updated_at ON user_two_factor;

-- Drop OAuth2 tables (in order due to foreign keys)
DROP TABLE IF EXISTS oauth2_tokens CASCADE;
DROP TABLE IF EXISTS oauth2_authorization_codes CASCADE;
DROP TABLE IF EXISTS oauth2_clients CASCADE;

-- Drop Two-Factor Authentication tables
DROP TABLE IF EXISTS backup_codes CASCADE;
DROP TABLE IF EXISTS user_two_factor CASCADE;

-- Remove migration record (if you have a migrations table)
-- DELETE FROM schema_migrations WHERE version = '004';

COMMENT ON SCHEMA public IS 'Phase 2 migration rolled back: Enhanced Authentication removed';
