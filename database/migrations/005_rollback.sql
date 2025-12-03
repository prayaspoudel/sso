-- Rollback for Phase 3: External Services

-- Drop indexes first
DROP INDEX IF EXISTS idx_social_login_states_expires_at;
DROP INDEX IF EXISTS idx_social_login_states_state;

DROP INDEX IF EXISTS idx_social_accounts_email;
DROP INDEX IF EXISTS idx_social_accounts_provider_id;
DROP INDEX IF EXISTS idx_social_accounts_provider;
DROP INDEX IF EXISTS idx_social_accounts_user_id;

DROP INDEX IF EXISTS idx_sms_otps_purpose;
DROP INDEX IF EXISTS idx_sms_otps_phone;
DROP INDEX IF EXISTS idx_sms_otps_user_id;

DROP INDEX IF EXISTS idx_sms_logs_message_id;
DROP INDEX IF EXISTS idx_sms_logs_created_at;
DROP INDEX IF EXISTS idx_sms_logs_status;
DROP INDEX IF EXISTS idx_sms_logs_to_phone;
DROP INDEX IF EXISTS idx_sms_logs_user_id;

DROP INDEX IF EXISTS idx_password_resets_email;
DROP INDEX IF EXISTS idx_password_resets_token;
DROP INDEX IF EXISTS idx_password_resets_user_id;

DROP INDEX IF EXISTS idx_email_verifications_email;
DROP INDEX IF EXISTS idx_email_verifications_token;
DROP INDEX IF EXISTS idx_email_verifications_user_id;

DROP INDEX IF EXISTS idx_email_logs_message_id;
DROP INDEX IF EXISTS idx_email_logs_created_at;
DROP INDEX IF EXISTS idx_email_logs_status;
DROP INDEX IF EXISTS idx_email_logs_to_email;
DROP INDEX IF EXISTS idx_email_logs_user_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS social_login_states;
DROP TABLE IF EXISTS social_accounts;
DROP TABLE IF EXISTS sms_otps;
DROP TABLE IF EXISTS sms_logs;
DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS email_logs;

-- Remove metadata column from users if it was added
-- ALTER TABLE users DROP COLUMN IF EXISTS metadata;
