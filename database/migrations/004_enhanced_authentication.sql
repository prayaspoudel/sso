-- Migration 004: Phase 2 - Enhanced Authentication (2FA/MFA and OAuth2)
-- This migration adds tables for Two-Factor Authentication and OAuth2 support

-- ============================================================================
-- Two-Factor Authentication Tables
-- ============================================================================

-- User Two-Factor Configuration
CREATE TABLE IF NOT EXISTS user_two_factor (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    method VARCHAR(10) NOT NULL CHECK (method IN ('totp', 'sms')),
    secret TEXT NOT NULL, -- Encrypted TOTP secret or phone number
    phone_number VARCHAR(20),
    status VARCHAR(20) NOT NULL CHECK (status IN ('disabled', 'pending', 'enabled')),
    backup_codes_count INTEGER NOT NULL DEFAULT 0,
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Backup Codes for 2FA
CREATE TABLE IF NOT EXISTS backup_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code TEXT NOT NULL, -- Hashed backup code
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- OAuth2 Tables
-- ============================================================================

-- OAuth2 Clients
CREATE TABLE IF NOT EXISTS oauth2_clients (
    id UUID PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL UNIQUE,
    client_secret TEXT NOT NULL, -- Hashed
    name VARCHAR(255) NOT NULL,
    description TEXT,
    redirect_uris TEXT[] NOT NULL, -- Array of allowed redirect URIs
    grant_types TEXT[] NOT NULL, -- Array of allowed grant types
    scopes TEXT[] NOT NULL, -- Array of allowed scopes
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    logo_url TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- OAuth2 Authorization Codes
CREATE TABLE IF NOT EXISTS oauth2_authorization_codes (
    id UUID PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    scopes TEXT[] NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- OAuth2 Access Tokens
CREATE TABLE IF NOT EXISTS oauth2_tokens (
    id UUID PRIMARY KEY,
    access_token TEXT NOT NULL UNIQUE,
    refresh_token TEXT UNIQUE,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- Indexes for Performance
-- ============================================================================

-- Two-Factor Authentication Indexes
CREATE INDEX IF NOT EXISTS idx_user_two_factor_user_id ON user_two_factor(user_id);
CREATE INDEX IF NOT EXISTS idx_user_two_factor_status ON user_two_factor(status);
CREATE INDEX IF NOT EXISTS idx_backup_codes_user_id ON backup_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_backup_codes_used_at ON backup_codes(used_at);

-- OAuth2 Indexes
CREATE INDEX IF NOT EXISTS idx_oauth2_clients_owner_id ON oauth2_clients(owner_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_clients_client_id ON oauth2_clients(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_clients_active ON oauth2_clients(active);

CREATE INDEX IF NOT EXISTS idx_oauth2_auth_codes_code ON oauth2_authorization_codes(code);
CREATE INDEX IF NOT EXISTS idx_oauth2_auth_codes_client_id ON oauth2_authorization_codes(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_auth_codes_user_id ON oauth2_authorization_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_auth_codes_expires_at ON oauth2_authorization_codes(expires_at);

CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_access_token ON oauth2_tokens(access_token);
CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_refresh_token ON oauth2_tokens(refresh_token);
CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_client_id ON oauth2_tokens(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_user_id ON oauth2_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_expires_at ON oauth2_tokens(expires_at);

-- ============================================================================
-- Triggers for Updated_at
-- ============================================================================

-- Trigger for user_two_factor
CREATE OR REPLACE TRIGGER update_user_two_factor_updated_at
    BEFORE UPDATE ON user_two_factor
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for oauth2_clients
CREATE OR REPLACE TRIGGER update_oauth2_clients_updated_at
    BEFORE UPDATE ON oauth2_clients
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Cleanup Functions
-- ============================================================================

-- Function to cleanup expired authorization codes
CREATE OR REPLACE FUNCTION cleanup_expired_oauth2_codes()
RETURNS void AS $$
BEGIN
    DELETE FROM oauth2_authorization_codes
    WHERE expires_at < NOW() - INTERVAL '1 day';
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup expired tokens
CREATE OR REPLACE FUNCTION cleanup_expired_oauth2_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM oauth2_tokens
    WHERE expires_at < NOW() - INTERVAL '7 days';
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup old used backup codes
CREATE OR REPLACE FUNCTION cleanup_old_backup_codes()
RETURNS void AS $$
BEGIN
    DELETE FROM backup_codes
    WHERE used_at IS NOT NULL 
    AND used_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Default OAuth2 Scopes Documentation
-- ============================================================================

-- Standard OAuth2/OpenID Connect scopes:
-- - openid: OpenID Connect authentication
-- - profile: Access to user profile information (name, etc.)
-- - email: Access to user email address
-- - offline_access: Request refresh token for offline access

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE user_two_factor IS 'Stores 2FA configuration for users';
COMMENT ON TABLE backup_codes IS 'Stores hashed backup codes for 2FA recovery';
COMMENT ON TABLE oauth2_clients IS 'OAuth2 client applications';
COMMENT ON TABLE oauth2_authorization_codes IS 'OAuth2 authorization codes (short-lived)';
COMMENT ON TABLE oauth2_tokens IS 'OAuth2 access and refresh tokens';

COMMENT ON COLUMN user_two_factor.secret IS 'TOTP secret (should be encrypted at rest)';
COMMENT ON COLUMN backup_codes.code IS 'Hashed backup code using bcrypt';
COMMENT ON COLUMN oauth2_clients.client_secret IS 'Hashed client secret using bcrypt';
COMMENT ON COLUMN oauth2_tokens.access_token IS 'JWT access token';
COMMENT ON COLUMN oauth2_tokens.refresh_token IS 'Opaque refresh token';

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Insert migration record (if you have a migrations table)
-- INSERT INTO schema_migrations (version, applied_at) VALUES ('004', CURRENT_TIMESTAMP);

COMMENT ON SCHEMA public IS 'Phase 2 migration applied: Enhanced Authentication (2FA/MFA and OAuth2)';
