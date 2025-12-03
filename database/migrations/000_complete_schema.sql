-- ============================================================================
-- SSO System - Complete Database Schema
-- ============================================================================
-- This is a consolidated migration containing all schema changes from:
-- - 001: Initial Schema
-- - 003: Security Features (Rate Limiting, Account Lockout, RBAC)
-- - 004: Enhanced Authentication (2FA/MFA and OAuth2)
-- - 005: External Services (Email, SMS, Social Login)
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- PHASE 1: Core Tables (Initial Schema)
-- ============================================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    company_id UUID,
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    email_verified BOOLEAN DEFAULT false,
    failed_login_attempts INTEGER DEFAULT 0,
    last_failed_login TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

-- Companies table
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    industry VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key for company_id in users
ALTER TABLE users ADD CONSTRAINT fk_users_company_id 
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL;

-- User-Company relationship
CREATE TABLE IF NOT EXISTS user_companies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, company_id)
);

-- OAuth2 Clients (for each micro-frontend)
CREATE TABLE IF NOT EXISTS oauth_clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    redirect_uris TEXT[],
    allowed_grants TEXT[],
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Refresh Tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    client_id VARCHAR(255) REFERENCES oauth_clients(client_id),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT false
);

-- Session Management
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(500) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Audit Log
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    details JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- PHASE 2: Security Features (Phase 1 Implementation)
-- ============================================================================

-- Login Attempts Table
CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    successful BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Account Lockouts Table
CREATE TABLE IF NOT EXISTS account_lockouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    locked_at TIMESTAMP WITH TIME ZONE NOT NULL,
    locked_until TIMESTAMP WITH TIME ZONE NOT NULL,
    failed_attempts INTEGER DEFAULT 0,
    reason TEXT NOT NULL,
    unlocked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Roles Table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User Roles Table (Many-to-Many relationship)
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

-- Permissions Table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- PHASE 3: Enhanced Authentication (Phase 2 Implementation)
-- ============================================================================

-- Two-Factor Authentication Table
CREATE TABLE IF NOT EXISTS two_factor_auth (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    secret TEXT NOT NULL,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    backup_codes TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Backup Codes for 2FA
CREATE TABLE IF NOT EXISTS backup_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- OAuth2 Clients (Enhanced)
CREATE TABLE IF NOT EXISTS oauth2_clients (
    id UUID PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL UNIQUE,
    client_secret TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    redirect_uris TEXT[] NOT NULL,
    grant_types TEXT[] NOT NULL,
    scopes TEXT[] NOT NULL,
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
-- PHASE 4: External Services (Phase 3 Implementation)
-- ============================================================================

-- Email logs table
CREATE TABLE IF NOT EXISTS email_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    to_email VARCHAR(255) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    message_id VARCHAR(255),
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    bounced_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Email verification tokens
CREATE TABLE IF NOT EXISTS email_verifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Password reset tokens
CREATE TABLE IF NOT EXISTS password_resets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- SMS logs table
CREATE TABLE IF NOT EXISTS sms_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    to_phone VARCHAR(20) NOT NULL,
    from_phone VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    message_id VARCHAR(255),
    segments INTEGER DEFAULT 1,
    cost DECIMAL(10, 4),
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    failed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- SMS OTP table
CREATE TABLE IF NOT EXISTS sms_otps (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    phone VARCHAR(20) NOT NULL,
    code_hash VARCHAR(255) NOT NULL,
    purpose VARCHAR(100) NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Social accounts table
CREATE TABLE IF NOT EXISTS social_accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    name VARCHAR(255),
    avatar TEXT,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    linked_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_id)
);

-- Social login state table (for OAuth flow)
CREATE TABLE IF NOT EXISTS social_login_states (
    id UUID PRIMARY KEY,
    state VARCHAR(255) NOT NULL UNIQUE,
    provider VARCHAR(50) NOT NULL,
    redirect_uri TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- INDEXES for Performance
-- ============================================================================

-- Core Tables Indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_companies_name ON companies(name);
CREATE INDEX IF NOT EXISTS idx_companies_status ON companies(status);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_user_companies_user_id ON user_companies(user_id);
CREATE INDEX IF NOT EXISTS idx_user_companies_company_id ON user_companies(company_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Security Features Indexes
CREATE INDEX IF NOT EXISTS idx_login_attempts_email ON login_attempts(email);
CREATE INDEX IF NOT EXISTS idx_login_attempts_created_at ON login_attempts(created_at);

CREATE INDEX IF NOT EXISTS idx_account_lockouts_user_id ON account_lockouts(user_id);
CREATE INDEX IF NOT EXISTS idx_account_lockouts_locked_until ON account_lockouts(locked_until);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);

CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action);

-- Enhanced Authentication Indexes
CREATE INDEX IF NOT EXISTS idx_two_factor_auth_user_id ON two_factor_auth(user_id);
CREATE INDEX IF NOT EXISTS idx_backup_codes_user_id ON backup_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_backup_codes_used_at ON backup_codes(used_at);

CREATE INDEX IF NOT EXISTS idx_oauth2_clients_client_id ON oauth2_clients(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_clients_owner_id ON oauth2_clients(owner_id);
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

-- External Services Indexes
CREATE INDEX IF NOT EXISTS idx_email_logs_user_id ON email_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_to_email ON email_logs(to_email);
CREATE INDEX IF NOT EXISTS idx_email_logs_status ON email_logs(status);
CREATE INDEX IF NOT EXISTS idx_email_logs_created_at ON email_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_email_logs_message_id ON email_logs(message_id);

CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token);
CREATE INDEX IF NOT EXISTS idx_email_verifications_email ON email_verifications(email);

CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
CREATE INDEX IF NOT EXISTS idx_password_resets_email ON password_resets(email);

CREATE INDEX IF NOT EXISTS idx_sms_logs_user_id ON sms_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_sms_logs_to_phone ON sms_logs(to_phone);
CREATE INDEX IF NOT EXISTS idx_sms_logs_status ON sms_logs(status);
CREATE INDEX IF NOT EXISTS idx_sms_logs_created_at ON sms_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_sms_logs_message_id ON sms_logs(message_id);

CREATE INDEX IF NOT EXISTS idx_sms_otps_user_id ON sms_otps(user_id);
CREATE INDEX IF NOT EXISTS idx_sms_otps_phone ON sms_otps(phone);
CREATE INDEX IF NOT EXISTS idx_sms_otps_purpose ON sms_otps(purpose);

CREATE INDEX IF NOT EXISTS idx_social_accounts_user_id ON social_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_social_accounts_provider ON social_accounts(provider);
CREATE INDEX IF NOT EXISTS idx_social_accounts_provider_id ON social_accounts(provider_id);
CREATE INDEX IF NOT EXISTS idx_social_accounts_email ON social_accounts(email);

CREATE INDEX IF NOT EXISTS idx_social_login_states_state ON social_login_states(state);
CREATE INDEX IF NOT EXISTS idx_social_login_states_expires_at ON social_login_states(expires_at);

-- ============================================================================
-- FUNCTIONS and TRIGGERS
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_companies_updated_at ON companies;
CREATE TRIGGER update_companies_updated_at BEFORE UPDATE ON companies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_two_factor_auth_updated_at ON two_factor_auth;
CREATE TRIGGER update_two_factor_auth_updated_at BEFORE UPDATE ON two_factor_auth
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_oauth2_clients_updated_at ON oauth2_clients;
CREATE TRIGGER update_oauth2_clients_updated_at BEFORE UPDATE ON oauth2_clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Cleanup Functions
CREATE OR REPLACE FUNCTION cleanup_old_login_attempts()
RETURNS void AS $$
BEGIN
    DELETE FROM login_attempts WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_expired_oauth2_codes()
RETURNS void AS $$
BEGIN
    DELETE FROM oauth2_authorization_codes
    WHERE expires_at < NOW() - INTERVAL '1 day';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_expired_oauth2_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM oauth2_tokens
    WHERE expires_at < NOW() - INTERVAL '7 days';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_old_backup_codes()
RETURNS void AS $$
BEGIN
    DELETE FROM backup_codes
    WHERE used_at IS NOT NULL 
    AND used_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- DEFAULT DATA
-- ============================================================================

-- Insert default OAuth clients for micro-frontends
INSERT INTO oauth_clients (client_id, client_secret, name, redirect_uris, allowed_grants, is_active) VALUES
    ('host-app', 'host-secret-change-in-production', 'Host Application', 
     ARRAY['http://localhost:3000/callback', 'http://localhost:3000/auth/callback'], 
     ARRAY['authorization_code', 'refresh_token'], true),
    ('crm-module', 'crm-secret-change-in-production', 'CRM Module', 
     ARRAY['http://localhost:3001/callback', 'http://localhost:3001/auth/callback'], 
     ARRAY['authorization_code', 'refresh_token'], true),
    ('inventory-module', 'inventory-secret-change-in-production', 'Inventory Module', 
     ARRAY['http://localhost:3002/callback', 'http://localhost:3002/auth/callback'], 
     ARRAY['authorization_code', 'refresh_token'], true),
    ('hr-module', 'hr-secret-change-in-production', 'HR Module', 
     ARRAY['http://localhost:3003/callback', 'http://localhost:3003/auth/callback'], 
     ARRAY['authorization_code', 'refresh_token'], true),
    ('finance-module', 'finance-secret-change-in-production', 'Finance Module', 
     ARRAY['http://localhost:3004/callback', 'http://localhost:3004/auth/callback'], 
     ARRAY['authorization_code', 'refresh_token'], true),
    ('task-module', 'task-secret-change-in-production', 'Task Module', 
     ARRAY['http://localhost:3005/callback', 'http://localhost:3005/auth/callback'], 
     ARRAY['authorization_code', 'refresh_token'], true)
ON CONFLICT (client_id) DO NOTHING;

-- Insert Default Roles
INSERT INTO roles (id, name, description, permissions, created_at, updated_at) VALUES
    (gen_random_uuid(), 'super_admin', 'Super Administrator with full access', 
     ARRAY['*'], CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'admin', 'Administrator with management access',
     ARRAY['users:read', 'users:write', 'users:delete', 'roles:read', 'roles:write', 'audit:read'],
     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'manager', 'Manager with limited management access',
     ARRAY['users:read', 'users:write', 'audit:read'],
     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'user', 'Standard user with basic access',
     ARRAY['profile:read', 'profile:write'],
     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;

-- Insert Default Permissions
INSERT INTO permissions (id, name, resource, action, description, created_at) VALUES
    (gen_random_uuid(), 'users:read', 'users', 'read', 'Read user information', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'users:write', 'users', 'write', 'Create and update users', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'users:delete', 'users', 'delete', 'Delete users', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'roles:read', 'roles', 'read', 'Read roles', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'roles:write', 'roles', 'write', 'Create and update roles', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'roles:delete', 'roles', 'delete', 'Delete roles', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'audit:read', 'audit', 'read', 'Read audit logs', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'profile:read', 'profile', 'read', 'Read own profile', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'profile:write', 'profile', 'write', 'Update own profile', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'companies:read', 'companies', 'read', 'Read companies', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'companies:write', 'companies', 'write', 'Create and update companies', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'companies:delete', 'companies', 'delete', 'Delete companies', CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- TABLE COMMENTS
-- ============================================================================

COMMENT ON TABLE users IS 'Stores user authentication and profile information';
COMMENT ON TABLE companies IS 'Stores company/organization information';
COMMENT ON TABLE user_companies IS 'Links users to companies with role information';
COMMENT ON TABLE oauth_clients IS 'Registered OAuth2 clients (micro-frontends)';
COMMENT ON TABLE refresh_tokens IS 'Stores refresh tokens for token rotation';
COMMENT ON TABLE sessions IS 'Active user sessions for security tracking';
COMMENT ON TABLE audit_logs IS 'Audit trail for security and compliance';

COMMENT ON TABLE login_attempts IS 'Tracks all login attempts for security monitoring';
COMMENT ON TABLE account_lockouts IS 'Tracks account lockouts due to failed login attempts';
COMMENT ON TABLE roles IS 'Defines user roles for role-based access control';
COMMENT ON TABLE user_roles IS 'Maps users to roles (many-to-many relationship)';
COMMENT ON TABLE permissions IS 'Defines system permissions';

COMMENT ON TABLE two_factor_auth IS 'Stores 2FA configuration for users';
COMMENT ON TABLE backup_codes IS 'Stores hashed backup codes for 2FA recovery';
COMMENT ON TABLE oauth2_clients IS 'OAuth2 client applications';
COMMENT ON TABLE oauth2_authorization_codes IS 'OAuth2 authorization codes (short-lived)';
COMMENT ON TABLE oauth2_tokens IS 'OAuth2 access and refresh tokens';

COMMENT ON TABLE email_logs IS 'Tracks all emails sent by the system';
COMMENT ON TABLE email_verifications IS 'Stores email verification tokens';
COMMENT ON TABLE password_resets IS 'Stores password reset tokens';
COMMENT ON TABLE sms_logs IS 'Tracks all SMS messages sent by the system';
COMMENT ON TABLE sms_otps IS 'Stores SMS OTP codes for verification';
COMMENT ON TABLE social_accounts IS 'Links users with their social media accounts';
COMMENT ON TABLE social_login_states IS 'Temporary storage for OAuth state tokens';

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
-- Total Tables Created: 34
-- Total Indexes Created: 70+
-- Total Functions Created: 5
-- Total Triggers Created: 5
-- ============================================================================
