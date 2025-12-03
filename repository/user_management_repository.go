package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"sso/models"

	"github.com/google/uuid"
)

// UserManagementRepository handles user management operations
type UserManagementRepository struct {
	db *sql.DB
}

// NewUserManagementRepository creates a new UserManagementRepository
func NewUserManagementRepository(db *sql.DB) *UserManagementRepository {
	return &UserManagementRepository{db: db}
}

// ListUsers retrieves paginated list of users with filters
func (r *UserManagementRepository) ListUsers(ctx context.Context, filter models.UserListFilter) ([]models.UserDetail, int, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.CompanyID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.company_id = $%d", argIndex))
		args = append(args, *filter.CompanyID)
		argIndex++
	}

	if filter.Role != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.role = $%d", argIndex))
		args = append(args, *filter.Role)
		argIndex++
	}

	if filter.IsVerified != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.email_verified = $%d", argIndex))
		args = append(args, *filter.IsVerified)
		argIndex++
	}

	if filter.CreatedFrom != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.created_at >= $%d", argIndex))
		args = append(args, *filter.CreatedFrom)
		argIndex++
	}

	if filter.CreatedTo != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.created_at <= $%d", argIndex))
		args = append(args, *filter.CreatedTo)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM users u
		%s
	`, whereClause)

	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	sortBy := "u.created_at"
	if filter.SortBy != "" {
		sortBy = "u." + filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Calculate pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Query users
	query := fmt.Sprintf(`
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.company_id, 
			c.name as company_name, u.role, u.is_active, u.email_verified,
			COALESCE(tf.two_factor_enabled, false) as two_factor_enabled,
			u.last_login_at, u.last_login_ip, 
			COALESCE(l.failed_attempts, 0) as failed_login_count,
			l.locked_until as account_locked_at,
			COALESCE(sa_count.count, 0) as social_account_count,
			u.created_at, u.updated_at
		FROM users u
		LEFT JOIN companies c ON u.company_id = c.id
		LEFT JOIN two_factor_auth tf ON u.id = tf.user_id
		LEFT JOIN account_lockouts l ON u.id = l.user_id AND (l.locked_until IS NULL OR l.locked_until > NOW())
		LEFT JOIN (
			SELECT user_id, COUNT(*) as count
			FROM social_accounts
			GROUP BY user_id
		) sa_count ON u.id = sa_count.user_id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []models.UserDetail{}
	for rows.Next() {
		var user models.UserDetail
		var companyName sql.NullString
		var lastLoginIP sql.NullString
		var accountLockedAt sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CompanyID,
			&companyName, &user.Role, &user.IsActive, &user.EmailVerified,
			&user.TwoFactorEnabled, &user.LastLoginAt, &lastLoginIP,
			&user.FailedLoginCount, &accountLockedAt, &user.SocialAccountCount,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if companyName.Valid {
			user.CompanyName = companyName.String
		}
		if lastLoginIP.Valid {
			user.LastLoginIP = lastLoginIP.String
		}
		if accountLockedAt.Valid {
			user.AccountLockedAt = &accountLockedAt.Time
		}

		users = append(users, user)
	}

	return users, totalCount, nil
}

// GetUserByID retrieves a user by ID with details
func (r *UserManagementRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.UserDetail, error) {
	query := `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.company_id,
			c.name as company_name, u.role, u.is_active, u.email_verified,
			COALESCE(tf.two_factor_enabled, false) as two_factor_enabled,
			u.last_login_at, u.last_login_ip,
			COALESCE(l.failed_attempts, 0) as failed_login_count,
			l.locked_until as account_locked_at,
			COALESCE(sa_count.count, 0) as social_account_count,
			u.created_at, u.updated_at
		FROM users u
		LEFT JOIN companies c ON u.company_id = c.id
		LEFT JOIN two_factor_auth tf ON u.id = tf.user_id
		LEFT JOIN account_lockouts l ON u.id = l.user_id AND (l.locked_until IS NULL OR l.locked_until > NOW())
		LEFT JOIN (
			SELECT user_id, COUNT(*) as count
			FROM social_accounts
			GROUP BY user_id
		) sa_count ON u.id = sa_count.user_id
		WHERE u.id = $1
	`

	var user models.UserDetail
	var companyName sql.NullString
	var lastLoginIP sql.NullString
	var accountLockedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CompanyID,
		&companyName, &user.Role, &user.IsActive, &user.EmailVerified,
		&user.TwoFactorEnabled, &user.LastLoginAt, &lastLoginIP,
		&user.FailedLoginCount, &accountLockedAt, &user.SocialAccountCount,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if companyName.Valid {
		user.CompanyName = companyName.String
	}
	if lastLoginIP.Valid {
		user.LastLoginIP = lastLoginIP.String
	}
	if accountLockedAt.Valid {
		user.AccountLockedAt = &accountLockedAt.Time
	}

	return &user, nil
}

// CreateUser creates a new user
func (r *UserManagementRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, company_id, 
			role, is_active, email_verified, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.CompanyID, user.Role, user.IsActive, user.EmailVerified,
		user.CreatedAt, user.UpdatedAt,
	)

	return err
}

// UpdateUser updates user information
func (r *UserManagementRepository) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build UPDATE clause
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Always update updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add user ID for WHERE clause
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteUser soft deletes a user (or hard delete if needed)
func (r *UserManagementRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Soft delete by marking inactive
	query := `
		UPDATE users
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// HardDeleteUser permanently deletes a user
func (r *UserManagementRepository) HardDeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateUserStatus updates user active status
func (r *UserManagementRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, isActive bool) error {
	query := `
		UPDATE users
		SET is_active = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, isActive, time.Now(), userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UnlockUserAccount unlocks a locked user account
func (r *UserManagementRepository) UnlockUserAccount(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE account_lockouts
		SET locked_until = NULL, failed_attempts = 0, updated_at = $1
		WHERE user_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// BulkUpdateUsers performs bulk update on multiple users
func (r *UserManagementRepository) BulkUpdateUsers(ctx context.Context, userIDs []uuid.UUID, updates map[string]interface{}) error {
	if len(userIDs) == 0 || len(updates) == 0 {
		return nil
	}

	// Build UPDATE clause
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Always update updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Build IN clause for user IDs
	placeholders := []string{}
	for _, userID := range userIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, userID)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id IN (%s)
	`, strings.Join(setClauses, ", "), strings.Join(placeholders, ", "))

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetUserStats retrieves user statistics
func (r *UserManagementRepository) GetUserStats(ctx context.Context) (*models.UserStats, error) {
	stats := &models.UserStats{
		UsersByRole:    make(map[string]int),
		UsersByCompany: make(map[string]int),
	}

	// Total users
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	// Active users
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, err
	}

	// Inactive users
	stats.InactiveUsers = stats.TotalUsers - stats.ActiveUsers

	// Verified users
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email_verified = true").Scan(&stats.VerifiedUsers)
	if err != nil {
		return nil, err
	}

	// Users with 2FA
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT user_id) FROM two_factor_auth WHERE two_factor_enabled = true").Scan(&stats.Users2FA)
	if err != nil {
		return nil, err
	}

	// New users today
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURRENT_DATE").Scan(&stats.NewUsersToday)
	if err != nil {
		return nil, err
	}

	// New users this week
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= DATE_TRUNC('week', CURRENT_DATE)").Scan(&stats.NewUsersThisWeek)
	if err != nil {
		return nil, err
	}

	// New users this month
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= DATE_TRUNC('month', CURRENT_DATE)").Scan(&stats.NewUsersThisMonth)
	if err != nil {
		return nil, err
	}

	// Logins today
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sessions WHERE DATE(created_at) = CURRENT_DATE").Scan(&stats.LoginsTodayCount)
	if err != nil {
		return nil, err
	}

	// Users by role
	rows, err := r.db.QueryContext(ctx, "SELECT role, COUNT(*) FROM users GROUP BY role")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			return nil, err
		}
		stats.UsersByRole[role] = count
	}

	// Users by company
	rows, err = r.db.QueryContext(ctx, `
		SELECT c.name, COUNT(u.id)
		FROM users u
		JOIN companies c ON u.company_id = c.id
		GROUP BY c.name
		ORDER BY COUNT(u.id) DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var companyName string
		var count int
		if err := rows.Scan(&companyName, &count); err != nil {
			return nil, err
		}
		stats.UsersByCompany[companyName] = count
	}

	return stats, nil
}
