-- Migration: Add Security Features (Rate Limiting, Account Lockout, RBAC)
-- Version: 003
-- Description: Adds tables for login attempts, account lockouts, roles, and permissions

-- Login Attempts Table
CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    successful BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_login_attempts_email (email),
    INDEX idx_login_attempts_created_at (created_at)
);

-- Account Lockouts Table
CREATE TABLE IF NOT EXISTS account_lockouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    locked_at TIMESTAMP WITH TIME ZONE NOT NULL,
    locked_until TIMESTAMP WITH TIME ZONE NOT NULL,
    reason TEXT NOT NULL,
    unlocked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_account_lockouts_user_id (user_id),
    INDEX idx_account_lockouts_locked_until (locked_until)
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
    UNIQUE(user_id, role_id),
    INDEX idx_user_roles_user_id (user_id),
    INDEX idx_user_roles_role_id (role_id)
);

-- Permissions Table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_permissions_resource (resource),
    INDEX idx_permissions_action (action)
);

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

-- Add failed_login_attempts counter to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_failed_login TIMESTAMP WITH TIME ZONE;

-- Create function to clean old login attempts
CREATE OR REPLACE FUNCTION cleanup_old_login_attempts()
RETURNS void AS $$
BEGIN
    DELETE FROM login_attempts WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

-- Create trigger to update user's updated_at on role changes
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments
COMMENT ON TABLE login_attempts IS 'Tracks all login attempts for security monitoring';
COMMENT ON TABLE account_lockouts IS 'Tracks account lockouts due to failed login attempts';
COMMENT ON TABLE roles IS 'Defines user roles for role-based access control';
COMMENT ON TABLE user_roles IS 'Maps users to roles (many-to-many relationship)';
COMMENT ON TABLE permissions IS 'Defines system permissions';
