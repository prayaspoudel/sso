-- ============================================================================
-- SSO System - Complete Schema Rollback
-- ============================================================================
-- This script drops all tables, indexes, functions, and triggers created by
-- the complete schema migration (000_complete_schema.sql)
-- ============================================================================
-- WARNING: This will delete ALL data in the SSO database!
-- ============================================================================

-- Drop all triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON users CASCADE;
DROP TRIGGER IF EXISTS update_companies_updated_at ON companies CASCADE;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles CASCADE;
DROP TRIGGER IF EXISTS update_two_factor_auth_updated_at ON two_factor_auth CASCADE;
DROP TRIGGER IF EXISTS update_oauth2_clients_updated_at ON oauth2_clients CASCADE;

-- Drop all functions
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS cleanup_old_login_attempts() CASCADE;
DROP FUNCTION IF EXISTS cleanup_expired_oauth2_codes() CASCADE;
DROP FUNCTION IF EXISTS cleanup_expired_oauth2_tokens() CASCADE;
DROP FUNCTION IF EXISTS cleanup_old_backup_codes() CASCADE;

-- Drop all indexes (indexes will be dropped automatically with tables, but explicit for clarity)
-- External Services Indexes
DROP INDEX IF EXISTS idx_social_login_states_expires_at CASCADE;
DROP INDEX IF EXISTS idx_social_login_states_state CASCADE;
DROP INDEX IF EXISTS idx_social_accounts_email CASCADE;
DROP INDEX IF EXISTS idx_social_accounts_provider_id CASCADE;
DROP INDEX IF EXISTS idx_social_accounts_provider CASCADE;
DROP INDEX IF EXISTS idx_social_accounts_user_id CASCADE;
DROP INDEX IF EXISTS idx_sms_otps_purpose CASCADE;
DROP INDEX IF EXISTS idx_sms_otps_phone CASCADE;
DROP INDEX IF EXISTS idx_sms_otps_user_id CASCADE;
DROP INDEX IF EXISTS idx_sms_logs_message_id CASCADE;
DROP INDEX IF EXISTS idx_sms_logs_created_at CASCADE;
DROP INDEX IF EXISTS idx_sms_logs_status CASCADE;
DROP INDEX IF EXISTS idx_sms_logs_to_phone CASCADE;
DROP INDEX IF EXISTS idx_sms_logs_user_id CASCADE;
DROP INDEX IF EXISTS idx_password_resets_email CASCADE;
DROP INDEX IF EXISTS idx_password_resets_token CASCADE;
DROP INDEX IF EXISTS idx_password_resets_user_id CASCADE;
DROP INDEX IF EXISTS idx_email_verifications_email CASCADE;
DROP INDEX IF EXISTS idx_email_verifications_token CASCADE;
DROP INDEX IF EXISTS idx_email_verifications_user_id CASCADE;
DROP INDEX IF EXISTS idx_email_logs_message_id CASCADE;
DROP INDEX IF EXISTS idx_email_logs_created_at CASCADE;
DROP INDEX IF EXISTS idx_email_logs_status CASCADE;
DROP INDEX IF EXISTS idx_email_logs_to_email CASCADE;
DROP INDEX IF EXISTS idx_email_logs_user_id CASCADE;

-- Enhanced Authentication Indexes
DROP INDEX IF EXISTS idx_oauth2_tokens_expires_at CASCADE;
DROP INDEX IF EXISTS idx_oauth2_tokens_user_id CASCADE;
DROP INDEX IF EXISTS idx_oauth2_tokens_client_id CASCADE;
DROP INDEX IF EXISTS idx_oauth2_tokens_refresh_token CASCADE;
DROP INDEX IF EXISTS idx_oauth2_tokens_access_token CASCADE;
DROP INDEX IF EXISTS idx_oauth2_auth_codes_expires_at CASCADE;
DROP INDEX IF EXISTS idx_oauth2_auth_codes_user_id CASCADE;
DROP INDEX IF EXISTS idx_oauth2_auth_codes_client_id CASCADE;
DROP INDEX IF EXISTS idx_oauth2_auth_codes_code CASCADE;
DROP INDEX IF EXISTS idx_oauth2_clients_active CASCADE;
DROP INDEX IF EXISTS idx_oauth2_clients_owner_id CASCADE;
DROP INDEX IF EXISTS idx_oauth2_clients_client_id CASCADE;
DROP INDEX IF EXISTS idx_backup_codes_used_at CASCADE;
DROP INDEX IF EXISTS idx_backup_codes_user_id CASCADE;
DROP INDEX IF EXISTS idx_two_factor_auth_user_id CASCADE;

-- Security Features Indexes
DROP INDEX IF EXISTS idx_permissions_action CASCADE;
DROP INDEX IF EXISTS idx_permissions_resource CASCADE;
DROP INDEX IF EXISTS idx_user_roles_role_id CASCADE;
DROP INDEX IF EXISTS idx_user_roles_user_id CASCADE;
DROP INDEX IF EXISTS idx_account_lockouts_locked_until CASCADE;
DROP INDEX IF EXISTS idx_account_lockouts_user_id CASCADE;
DROP INDEX IF EXISTS idx_login_attempts_created_at CASCADE;
DROP INDEX IF EXISTS idx_login_attempts_email CASCADE;

-- Core Tables Indexes
DROP INDEX IF EXISTS idx_users_created_at CASCADE;
DROP INDEX IF EXISTS idx_users_email_verified CASCADE;
DROP INDEX IF EXISTS idx_users_is_active CASCADE;
DROP INDEX IF EXISTS idx_users_role CASCADE;
DROP INDEX IF EXISTS idx_users_company_id CASCADE;
DROP INDEX IF EXISTS idx_users_email CASCADE;
DROP INDEX IF EXISTS idx_companies_status CASCADE;
DROP INDEX IF EXISTS idx_companies_name CASCADE;
DROP INDEX IF EXISTS idx_audit_logs_created_at CASCADE;
DROP INDEX IF EXISTS idx_audit_logs_action CASCADE;
DROP INDEX IF EXISTS idx_audit_logs_user_id CASCADE;
DROP INDEX IF EXISTS idx_user_companies_company_id CASCADE;
DROP INDEX IF EXISTS idx_user_companies_user_id CASCADE;
DROP INDEX IF EXISTS idx_sessions_expires_at CASCADE;
DROP INDEX IF EXISTS idx_sessions_token CASCADE;
DROP INDEX IF EXISTS idx_sessions_user_id CASCADE;
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at CASCADE;
DROP INDEX IF EXISTS idx_refresh_tokens_token CASCADE;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id CASCADE;

-- Drop all tables in reverse order of dependencies
-- Phase 4: External Services Tables
DROP TABLE IF EXISTS social_login_states CASCADE;
DROP TABLE IF EXISTS social_accounts CASCADE;
DROP TABLE IF EXISTS sms_otps CASCADE;
DROP TABLE IF EXISTS sms_logs CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS email_verifications CASCADE;
DROP TABLE IF EXISTS email_logs CASCADE;

-- Phase 3: Enhanced Authentication Tables
DROP TABLE IF EXISTS oauth2_tokens CASCADE;
DROP TABLE IF EXISTS oauth2_authorization_codes CASCADE;
DROP TABLE IF EXISTS oauth2_clients CASCADE;
DROP TABLE IF EXISTS backup_codes CASCADE;
DROP TABLE IF EXISTS two_factor_auth CASCADE;

-- Phase 2: Security Features Tables
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS account_lockouts CASCADE;
DROP TABLE IF EXISTS login_attempts CASCADE;

-- Phase 1: Core Tables
DROP TABLE IF EXISTS email_verification_tokens CASCADE;
DROP TABLE IF EXISTS password_reset_tokens CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS oauth_clients CASCADE;
DROP TABLE IF EXISTS user_companies CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS companies CASCADE;

-- Drop extensions (optional - only if no other databases use them)
-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- ============================================================================
-- ROLLBACK COMPLETE
-- ============================================================================
-- All tables, indexes, functions, and triggers have been removed
-- ============================================================================
