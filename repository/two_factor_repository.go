package repository

import (
	"context"
	"database/sql"
	"time"

	"sso/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TwoFactorRepository handles database operations for 2FA
type TwoFactorRepository struct {
	db *sql.DB
}

// NewTwoFactorRepository creates a new TwoFactorRepository
func NewTwoFactorRepository(db *sql.DB) *TwoFactorRepository {
	return &TwoFactorRepository{db: db}
}

// CreateTwoFactor creates a new 2FA configuration for a user
func (r *TwoFactorRepository) CreateTwoFactor(ctx context.Context, twoFactor *models.UserTwoFactor) error {
	query := `
		INSERT INTO user_two_factor (id, user_id, method, secret, phone_number, status, backup_codes_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		twoFactor.ID,
		twoFactor.UserID,
		twoFactor.Method,
		twoFactor.Secret,
		twoFactor.PhoneNumber,
		twoFactor.Status,
		twoFactor.BackupCodesCount,
		twoFactor.CreatedAt,
		twoFactor.UpdatedAt,
	)
	return err
}

// GetTwoFactorByUserID retrieves 2FA configuration by user ID
func (r *TwoFactorRepository) GetTwoFactorByUserID(ctx context.Context, userID uuid.UUID) (*models.UserTwoFactor, error) {
	query := `
		SELECT id, user_id, method, secret, phone_number, status, backup_codes_count, verified_at, created_at, updated_at
		FROM user_two_factor
		WHERE user_id = $1
	`
	var twoFactor models.UserTwoFactor
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&twoFactor.ID,
		&twoFactor.UserID,
		&twoFactor.Method,
		&twoFactor.Secret,
		&twoFactor.PhoneNumber,
		&twoFactor.Status,
		&twoFactor.BackupCodesCount,
		&twoFactor.VerifiedAt,
		&twoFactor.CreatedAt,
		&twoFactor.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &twoFactor, nil
}

// UpdateTwoFactorStatus updates the status of 2FA
func (r *TwoFactorRepository) UpdateTwoFactorStatus(ctx context.Context, userID uuid.UUID, status models.TwoFactorStatus) error {
	query := `
		UPDATE user_two_factor
		SET status = $1, verified_at = $2, updated_at = $3
		WHERE user_id = $4
	`
	now := time.Now()
	var verifiedAt *time.Time
	if status == models.TwoFactorStatusEnabled {
		verifiedAt = &now
	}
	_, err := r.db.ExecContext(ctx, query, status, verifiedAt, now, userID)
	return err
}

// DeleteTwoFactor deletes 2FA configuration for a user
func (r *TwoFactorRepository) DeleteTwoFactor(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_two_factor WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// CreateBackupCodes creates backup codes for a user
func (r *TwoFactorRepository) CreateBackupCodes(ctx context.Context, codes []models.BackupCode) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing unused backup codes
	_, err = tx.ExecContext(ctx, `DELETE FROM backup_codes WHERE user_id = $1 AND used_at IS NULL`, codes[0].UserID)
	if err != nil {
		return err
	}

	// Insert new backup codes
	query := `
		INSERT INTO backup_codes (id, user_id, code, created_at)
		VALUES ($1, $2, $3, $4)
	`
	for _, code := range codes {
		_, err = tx.ExecContext(ctx, query, code.ID, code.UserID, code.Code, code.CreatedAt)
		if err != nil {
			return err
		}
	}

	// Update backup codes count
	_, err = tx.ExecContext(ctx, `
		UPDATE user_two_factor
		SET backup_codes_count = $1, updated_at = $2
		WHERE user_id = $3
	`, len(codes), time.Now(), codes[0].UserID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetBackupCode retrieves a backup code
func (r *TwoFactorRepository) GetBackupCode(ctx context.Context, userID uuid.UUID, codeHash string) (*models.BackupCode, error) {
	query := `
		SELECT id, user_id, code, used_at, created_at
		FROM backup_codes
		WHERE user_id = $1 AND code = $2
	`
	var code models.BackupCode
	err := r.db.QueryRowContext(ctx, query, userID, codeHash).Scan(
		&code.ID,
		&code.UserID,
		&code.Code,
		&code.UsedAt,
		&code.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// MarkBackupCodeUsed marks a backup code as used
func (r *TwoFactorRepository) MarkBackupCodeUsed(ctx context.Context, codeID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	// Mark code as used
	_, err = tx.ExecContext(ctx, `
		UPDATE backup_codes
		SET used_at = $1
		WHERE id = $2
	`, now, codeID)
	if err != nil {
		return err
	}

	// Decrement backup codes count
	_, err = tx.ExecContext(ctx, `
		UPDATE user_two_factor
		SET backup_codes_count = backup_codes_count - 1, updated_at = $1
		WHERE user_id = (SELECT user_id FROM backup_codes WHERE id = $2)
	`, now, codeID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetUnusedBackupCodesCount returns the count of unused backup codes
func (r *TwoFactorRepository) GetUnusedBackupCodesCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM backup_codes
		WHERE user_id = $1 AND used_at IS NULL
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// ListBackupCodes lists all backup codes for a user (for admin purposes)
func (r *TwoFactorRepository) ListBackupCodes(ctx context.Context, userID uuid.UUID) ([]models.BackupCode, error) {
	query := `
		SELECT id, user_id, code, used_at, created_at
		FROM backup_codes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []models.BackupCode
	for rows.Next() {
		var code models.BackupCode
		err := rows.Scan(
			&code.ID,
			&code.UserID,
			&code.Code,
			&code.UsedAt,
			&code.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

// OAuth2Repository handles database operations for OAuth2
type OAuth2Repository struct {
	db *sql.DB
}

// NewOAuth2Repository creates a new OAuth2Repository
func NewOAuth2Repository(db *sql.DB) *OAuth2Repository {
	return &OAuth2Repository{db: db}
}

// CreateClient creates a new OAuth2 client
func (r *OAuth2Repository) CreateClient(ctx context.Context, client *models.OAuth2Client) error {
	query := `
		INSERT INTO oauth2_clients (id, client_id, client_secret, name, description, redirect_uris, grant_types, scopes, owner_id, logo_url, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		client.ID,
		client.ClientID,
		client.ClientSecret,
		client.Name,
		client.Description,
		pq.Array(client.RedirectURIs),
		pq.Array(client.GrantTypes),
		pq.Array(client.Scopes),
		client.OwnerID,
		client.LogoURL,
		client.Active,
		client.CreatedAt,
		client.UpdatedAt,
	)
	return err
}

// GetClientByClientID retrieves an OAuth2 client by client ID
func (r *OAuth2Repository) GetClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	query := `
		SELECT id, client_id, client_secret, name, description, redirect_uris, grant_types, scopes, owner_id, logo_url, active, created_at, updated_at
		FROM oauth2_clients
		WHERE client_id = $1
	`
	var client models.OAuth2Client
	err := r.db.QueryRowContext(ctx, query, clientID).Scan(
		&client.ID,
		&client.ClientID,
		&client.ClientSecret,
		&client.Name,
		&client.Description,
		pq.Array(&client.RedirectURIs),
		pq.Array(&client.GrantTypes),
		pq.Array(&client.Scopes),
		&client.OwnerID,
		&client.LogoURL,
		&client.Active,
		&client.CreatedAt,
		&client.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// ListClientsByOwner lists all OAuth2 clients owned by a user
func (r *OAuth2Repository) ListClientsByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.OAuth2Client, error) {
	query := `
		SELECT id, client_id, client_secret, name, description, redirect_uris, grant_types, scopes, owner_id, logo_url, active, created_at, updated_at
		FROM oauth2_clients
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.OAuth2Client
	for rows.Next() {
		var client models.OAuth2Client
		err := rows.Scan(
			&client.ID,
			&client.ClientID,
			&client.ClientSecret,
			&client.Name,
			&client.Description,
			pq.Array(&client.RedirectURIs),
			pq.Array(&client.GrantTypes),
			pq.Array(&client.Scopes),
			&client.OwnerID,
			&client.LogoURL,
			&client.Active,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, rows.Err()
}

// CreateAuthorizationCode creates a new authorization code
func (r *OAuth2Repository) CreateAuthorizationCode(ctx context.Context, code *models.OAuth2AuthorizationCode) error {
	query := `
		INSERT INTO oauth2_authorization_codes (id, code, client_id, user_id, redirect_uri, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		code.ID,
		code.Code,
		code.ClientID,
		code.UserID,
		code.RedirectURI,
		pq.Array(code.Scopes),
		code.ExpiresAt,
		code.CreatedAt,
	)
	return err
}

// GetAuthorizationCode retrieves an authorization code
func (r *OAuth2Repository) GetAuthorizationCode(ctx context.Context, code string) (*models.OAuth2AuthorizationCode, error) {
	query := `
		SELECT id, code, client_id, user_id, redirect_uri, scopes, expires_at, used_at, created_at
		FROM oauth2_authorization_codes
		WHERE code = $1
	`
	var authCode models.OAuth2AuthorizationCode
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&authCode.ID,
		&authCode.Code,
		&authCode.ClientID,
		&authCode.UserID,
		&authCode.RedirectURI,
		pq.Array(&authCode.Scopes),
		&authCode.ExpiresAt,
		&authCode.UsedAt,
		&authCode.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &authCode, nil
}

// MarkAuthorizationCodeUsed marks an authorization code as used
func (r *OAuth2Repository) MarkAuthorizationCodeUsed(ctx context.Context, codeID uuid.UUID) error {
	query := `
		UPDATE oauth2_authorization_codes
		SET used_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), codeID)
	return err
}

// CreateToken creates a new OAuth2 token
func (r *OAuth2Repository) CreateToken(ctx context.Context, token *models.OAuth2Token) error {
	query := `
		INSERT INTO oauth2_tokens (id, access_token, refresh_token, client_id, user_id, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.AccessToken,
		token.RefreshToken,
		token.ClientID,
		token.UserID,
		pq.Array(token.Scopes),
		token.ExpiresAt,
		token.CreatedAt,
	)
	return err
}

// GetTokenByAccessToken retrieves a token by access token
func (r *OAuth2Repository) GetTokenByAccessToken(ctx context.Context, accessToken string) (*models.OAuth2Token, error) {
	query := `
		SELECT id, access_token, refresh_token, client_id, user_id, scopes, expires_at, revoked_at, created_at
		FROM oauth2_tokens
		WHERE access_token = $1
	`
	var token models.OAuth2Token
	err := r.db.QueryRowContext(ctx, query, accessToken).Scan(
		&token.ID,
		&token.AccessToken,
		&token.RefreshToken,
		&token.ClientID,
		&token.UserID,
		pq.Array(&token.Scopes),
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetTokenByRefreshToken retrieves a token by refresh token
func (r *OAuth2Repository) GetTokenByRefreshToken(ctx context.Context, refreshToken string) (*models.OAuth2Token, error) {
	query := `
		SELECT id, access_token, refresh_token, client_id, user_id, scopes, expires_at, revoked_at, created_at
		FROM oauth2_tokens
		WHERE refresh_token = $1
	`
	var token models.OAuth2Token
	err := r.db.QueryRowContext(ctx, query, refreshToken).Scan(
		&token.ID,
		&token.AccessToken,
		&token.RefreshToken,
		&token.ClientID,
		&token.UserID,
		pq.Array(&token.Scopes),
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// RevokeToken revokes an OAuth2 token
func (r *OAuth2Repository) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE oauth2_tokens
		SET revoked_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), tokenID)
	return err
}

// DeleteExpiredAuthorizationCodes deletes expired authorization codes
func (r *OAuth2Repository) DeleteExpiredAuthorizationCodes(ctx context.Context) error {
	query := `DELETE FROM oauth2_authorization_codes WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

// DeleteExpiredTokens deletes expired tokens
func (r *OAuth2Repository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM oauth2_tokens WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
