# Database Migration Guide

## Overview

This guide explains how to use the consolidated database migration for the SSO system.

## Migration Files

### Primary Files (RECOMMENDED)

- **`000_complete_schema.sql`** - Single comprehensive migration containing everything
- **`000_complete_rollback.sql`** - Complete rollback script

### Individual Migration Files (Legacy - Still Available)

- `001_initial_schema.sql` + `002_rollback.sql` - Core tables
- `003_security_features.sql` + `003_rollback.sql` - Security features
- `004_enhanced_authentication.sql` + `004_rollback.sql` - 2FA/OAuth2
- `005_external_services.sql` + `005_rollback.sql` - Email/SMS/Social

## Quick Start

### Option 1: Fresh Installation (RECOMMENDED)

For a new installation, use the consolidated migration:

```bash
# 1. Create database
createdb sso_db

# 2. Run complete schema
psql -U postgres -d sso_db -f database/migrations/000_complete_schema.sql

# Expected output:
# CREATE EXTENSION
# CREATE TABLE (x34)
# CREATE INDEX (x70+)
# CREATE FUNCTION (x5)
# CREATE TRIGGER (x5)
# INSERT 0 6 (OAuth clients)
# INSERT 0 4 (Roles)
# INSERT 0 12 (Permissions)
```

### Option 2: Complete Rollback

To completely remove the schema:

```bash
psql -U postgres -d sso_db -f database/migrations/000_complete_rollback.sql
```

### Option 3: Incremental Migration (Legacy)

If you prefer to run migrations incrementally:

```bash
# Run in order
psql -U postgres -d sso_db -f database/migrations/001_initial_schema.sql
psql -U postgres -d sso_db -f database/migrations/003_security_features.sql
psql -U postgres -d sso_db -f database/migrations/004_enhanced_authentication.sql
psql -U postgres -d sso_db -f database/migrations/005_external_services.sql
```

## What Gets Created

### Tables (34 total)

**Core Tables (10)**:
- users
- companies
- user_companies
- oauth_clients
- refresh_tokens
- sessions
- audit_logs
- password_reset_tokens (legacy - superseded by password_resets)
- email_verification_tokens (legacy - superseded by email_verifications)

**Security Tables (5)**:
- login_attempts
- account_lockouts
- roles
- user_roles
- permissions

**Authentication Tables (5)**:
- two_factor_auth
- backup_codes
- oauth2_clients
- oauth2_authorization_codes
- oauth2_tokens

**External Services Tables (7)**:
- email_logs
- email_verifications
- password_resets
- sms_logs
- sms_otps
- social_accounts
- social_login_states

### Indexes (70+)

All tables have appropriate indexes for:
- Primary keys
- Foreign keys
- Lookup fields (email, token, etc.)
- Timestamp fields (for cleanup)
- Status fields (for filtering)

### Functions (5)

1. `update_updated_at_column()` - Auto-update timestamps
2. `cleanup_old_login_attempts()` - Remove old login attempts
3. `cleanup_expired_oauth2_codes()` - Remove expired auth codes
4. `cleanup_expired_oauth2_tokens()` - Remove expired tokens
5. `cleanup_old_backup_codes()` - Remove old used backup codes

### Triggers (5)

1. `update_users_updated_at` - Users table timestamp
2. `update_companies_updated_at` - Companies table timestamp
3. `update_roles_updated_at` - Roles table timestamp
4. `update_two_factor_auth_updated_at` - 2FA table timestamp
5. `update_oauth2_clients_updated_at` - OAuth2 clients table timestamp

### Default Data

**OAuth Clients (6)**:
- host-app
- crm-module
- inventory-module
- hr-module
- finance-module
- task-module

**Roles (4)**:
- super_admin (all permissions)
- admin (management permissions)
- manager (limited management)
- user (basic permissions)

**Permissions (12)**:
- users:read, users:write, users:delete
- roles:read, roles:write, roles:delete
- audit:read
- profile:read, profile:write
- companies:read, companies:write, companies:delete

## Verification

After running the migration, verify it worked:

```bash
# Check tables
psql -U postgres -d sso_db -c "\dt"

# Should show 34 tables

# Check roles
psql -U postgres -d sso_db -c "SELECT name FROM roles ORDER BY name;"

# Should show: admin, manager, super_admin, user

# Check OAuth clients
psql -U postgres -d sso_db -c "SELECT client_id, name FROM oauth_clients;"

# Should show 6 clients
```

## Maintenance

### Scheduled Cleanup

Run these periodically (cron job recommended):

```sql
-- Clean old login attempts (30+ days)
SELECT cleanup_old_login_attempts();

-- Clean expired OAuth2 codes (1+ day)
SELECT cleanup_expired_oauth2_codes();

-- Clean expired OAuth2 tokens (7+ days)
SELECT cleanup_expired_oauth2_tokens();

-- Clean old backup codes (90+ days after use)
SELECT cleanup_old_backup_codes();
```

### Manual Cleanup

```sql
-- Clean expired sessions
DELETE FROM sessions WHERE expires_at < NOW();

-- Clean expired refresh tokens
DELETE FROM refresh_tokens WHERE expires_at < NOW();

-- Clean expired email verifications
DELETE FROM email_verifications WHERE expires_at < NOW() AND verified = false;

-- Clean expired password resets
DELETE FROM password_resets WHERE expires_at < NOW() AND used = false;

-- Clean expired SMS OTPs
DELETE FROM sms_otps WHERE expires_at < NOW() AND verified = false;

-- Clean expired social login states
DELETE FROM social_login_states WHERE expires_at < NOW();
```

## Troubleshooting

### Issue: "relation already exists"

If you see errors about tables already existing:

```bash
# Option 1: Drop specific table
psql -U postgres -d sso_db -c "DROP TABLE IF EXISTS table_name CASCADE;"

# Option 2: Complete rollback and retry
psql -U postgres -d sso_db -f database/migrations/000_complete_rollback.sql
psql -U postgres -d sso_db -f database/migrations/000_complete_schema.sql
```

### Issue: "extension uuid-ossp does not exist"

```bash
# Install PostgreSQL contrib package
# Ubuntu/Debian:
sudo apt-get install postgresql-contrib

# macOS (Homebrew):
# Usually included with PostgreSQL

# Then create extension
psql -U postgres -d sso_db -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
```

### Issue: Foreign key violations

```bash
# Check for orphaned records
psql -U postgres -d sso_db -c "
  SELECT * FROM users WHERE company_id IS NOT NULL 
  AND company_id NOT IN (SELECT id FROM companies);
"

# Fix by creating missing company or setting to NULL
```

## Best Practices

### 1. Always Backup Before Migration

```bash
# Backup before migration
pg_dump -U postgres sso_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore if needed
psql -U postgres -d sso_db < backup_20241025_120000.sql
```

### 2. Test in Development First

Always test migrations in development before production:

```bash
# Development
psql -U postgres -d sso_db_dev -f database/migrations/000_complete_schema.sql

# If successful, then production
psql -U postgres -d sso_db_prod -f database/migrations/000_complete_schema.sql
```

### 3. Use Transactions for Safety

```bash
# Run in transaction (can rollback on error)
psql -U postgres -d sso_db << EOF
BEGIN;
\i database/migrations/000_complete_schema.sql
-- Check results
\dt
-- If looks good: COMMIT;
-- If problems: ROLLBACK;
COMMIT;
EOF
```

### 4. Monitor Performance

After migration, check index usage:

```sql
-- Check index usage
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

## Production Checklist

Before running in production:

- [ ] Backup database
- [ ] Test in staging environment
- [ ] Review migration script
- [ ] Schedule maintenance window
- [ ] Notify users of downtime
- [ ] Run migration
- [ ] Verify tables created
- [ ] Verify default data inserted
- [ ] Run smoke tests
- [ ] Monitor application logs
- [ ] Monitor database performance

## Additional Resources

- **Schema Diagram**: See `docs/schema_diagram.png` (TODO)
- **API Documentation**: See `docs/API.md` and `docs/PHASE3_IMPLEMENTATION.md`
- **Implementation Guide**: See `docs/IMPLEMENTATION.md`
- **Progress Summary**: See `docs/PROGRESS_SUMMARY.md`

## Support

For issues or questions:
1. Check error messages carefully
2. Review this guide
3. Check PostgreSQL logs: `tail -f /var/log/postgresql/postgresql-*.log`
4. Verify PostgreSQL version (requires 12+)

## Migration Timeline

- **001**: Initial schema (Core tables)
- **003**: Security features (RBAC, lockouts)
- **004**: Enhanced authentication (2FA, OAuth2)
- **005**: External services (Email, SMS, Social)
- **000**: Consolidated (All of the above)

---

**Recommendation**: Use `000_complete_schema.sql` for all new installations. Legacy individual migration files are kept for reference and incremental updates.
