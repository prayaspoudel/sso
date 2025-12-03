-- Rollback Migration for Security Features
-- Version: 003
-- Description: Removes security features tables

-- Drop triggers
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS cleanup_old_login_attempts();

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS account_lockouts;
DROP TABLE IF EXISTS login_attempts;

-- Remove columns from users table
ALTER TABLE users DROP COLUMN IF EXISTS failed_login_attempts;
ALTER TABLE users DROP COLUMN IF EXISTS last_failed_login;
