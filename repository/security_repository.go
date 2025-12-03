package repository

import (
	"context"
	"database/sql"
	"time"

	"sso/models"

	"github.com/google/uuid"
)

type SecurityRepository struct {
	db *sql.DB
}

func NewSecurityRepository(db *sql.DB) *SecurityRepository {
	return &SecurityRepository{db: db}
}

// Login Attempts
func (r *SecurityRepository) RecordLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) error {
	query := `
		INSERT INTO login_attempts (id, email, ip_address, successful, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		attempt.ID, attempt.Email, attempt.IPAddress, attempt.Successful, attempt.CreatedAt,
	)
	return err
}

func (r *SecurityRepository) GetRecentFailedAttempts(ctx context.Context, email string, duration time.Duration) (int, error) {
	query := `
		SELECT COUNT(*) FROM login_attempts
		WHERE email = $1 AND successful = false AND created_at > $2
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, email, time.Now().Add(-duration)).Scan(&count)
	return count, err
}

func (r *SecurityRepository) ClearLoginAttempts(ctx context.Context, email string) error {
	query := `DELETE FROM login_attempts WHERE email = $1`
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}

// Account Lockouts
func (r *SecurityRepository) LockAccount(ctx context.Context, lockout *models.AccountLockout) error {
	query := `
		INSERT INTO account_lockouts (id, user_id, locked_at, locked_until, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		lockout.ID, lockout.UserID, lockout.LockedAt, lockout.LockedUntil, lockout.Reason, lockout.CreatedAt,
	)
	return err
}

func (r *SecurityRepository) IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*) FROM account_lockouts
		WHERE user_id = $1 AND locked_until > $2 AND unlocked_at IS NULL
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, time.Now()).Scan(&count)
	return count > 0, err
}

func (r *SecurityRepository) GetAccountLockout(ctx context.Context, userID uuid.UUID) (*models.AccountLockout, error) {
	query := `
		SELECT id, user_id, locked_at, locked_until, reason, unlocked_at, created_at
		FROM account_lockouts
		WHERE user_id = $1 AND locked_until > $2 AND unlocked_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`
	lockout := &models.AccountLockout{}
	err := r.db.QueryRowContext(ctx, query, userID, time.Now()).Scan(
		&lockout.ID, &lockout.UserID, &lockout.LockedAt, &lockout.LockedUntil,
		&lockout.Reason, &lockout.UnlockedAt, &lockout.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return lockout, err
}

func (r *SecurityRepository) UnlockAccount(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE account_lockouts
		SET unlocked_at = $1
		WHERE user_id = $2 AND unlocked_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// RBAC - Roles
func (r *SecurityRepository) CreateRole(ctx context.Context, role *models.Role) error {
	query := `
		INSERT INTO roles (id, name, description, permissions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		role.ID, role.Name, role.Description, role.Permissions, role.CreatedAt, role.UpdatedAt,
	)
	return err
}

func (r *SecurityRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM roles WHERE id = $1
	`
	role := &models.Role{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.Permissions,
		&role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *SecurityRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM roles WHERE name = $1
	`
	role := &models.Role{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.Permissions,
		&role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *SecurityRepository) ListRoles(ctx context.Context) ([]*models.Role, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM roles ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.Permissions,
			&role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

// User Roles
func (r *SecurityRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `
		INSERT INTO user_roles (id, user_id, role_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New(), userID, roleID, time.Now())
	return err
}

func (r *SecurityRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *SecurityRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.permissions, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.Permissions,
			&role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *SecurityRepository) UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM user_roles ur
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND $2 = ANY(r.permissions)
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, permission).Scan(&count)
	return count > 0, err
}
