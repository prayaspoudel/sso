package repository

import (
	"context"
	"database/sql"
	"time"

	"sso/models"

	"github.com/google/uuid"
)

// EmailRepository handles email-related database operations
type EmailRepository struct {
	db *sql.DB
}

// NewEmailRepository creates a new EmailRepository
func NewEmailRepository(db *sql.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// CreateEmailLog creates a new email log entry
func (r *EmailRepository) CreateEmailLog(ctx context.Context, log *models.EmailLog) error {
	query := `
		INSERT INTO email_logs (id, to_email, from_email, subject, template, provider, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.ToEmail,
		log.FromEmail,
		log.Subject,
		log.Template,
		log.Provider,
		log.Status,
		log.CreatedAt,
	)
	return err
}

// UpdateEmailLogStatus updates the status of an email log
func (r *EmailRepository) UpdateEmailLogStatus(ctx context.Context, logID uuid.UUID, status models.EmailStatus, errorMsg *string) error {
	query := `
		UPDATE email_logs
		SET status = $1, error_msg = $2, sent_at = $3
		WHERE id = $4
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, status, errorMsg, now, logID)
	return err
}

// CreateEmailVerification creates a new email verification token
func (r *EmailRepository) CreateEmailVerification(ctx context.Context, verification *models.EmailVerification) error {
	query := `
		INSERT INTO email_verifications (id, user_id, email, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		verification.ID,
		verification.UserID,
		verification.Email,
		verification.Token,
		verification.ExpiresAt,
		verification.CreatedAt,
	)
	return err
}

// GetEmailVerification retrieves an email verification by token
func (r *EmailRepository) GetEmailVerification(ctx context.Context, token string) (*models.EmailVerification, error) {
	query := `
		SELECT id, user_id, email, token, expires_at, verified_at, created_at
		FROM email_verifications
		WHERE token = $1
	`
	var verification models.EmailVerification
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.Email,
		&verification.Token,
		&verification.ExpiresAt,
		&verification.VerifiedAt,
		&verification.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// MarkEmailVerified marks an email as verified
func (r *EmailRepository) MarkEmailVerified(ctx context.Context, token string) error {
	query := `
		UPDATE email_verifications
		SET verified_at = $1
		WHERE token = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), token)
	return err
}

// CreatePasswordReset creates a new password reset token
func (r *EmailRepository) CreatePasswordReset(ctx context.Context, reset *models.PasswordReset) error {
	query := `
		INSERT INTO password_resets (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		reset.ID,
		reset.UserID,
		reset.Token,
		reset.ExpiresAt,
		reset.CreatedAt,
	)
	return err
}

// GetPasswordReset retrieves a password reset by token
func (r *EmailRepository) GetPasswordReset(ctx context.Context, token string) (*models.PasswordReset, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_resets
		WHERE token = $1
	`
	var reset models.PasswordReset
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&reset.ID,
		&reset.UserID,
		&reset.Token,
		&reset.ExpiresAt,
		&reset.UsedAt,
		&reset.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &reset, nil
}

// MarkPasswordResetUsed marks a password reset token as used
func (r *EmailRepository) MarkPasswordResetUsed(ctx context.Context, token string) error {
	query := `
		UPDATE password_resets
		SET used_at = $1
		WHERE token = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), token)
	return err
}

// DeleteExpiredVerifications deletes expired email verifications
func (r *EmailRepository) DeleteExpiredVerifications(ctx context.Context) error {
	query := `DELETE FROM email_verifications WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

// DeleteExpiredPasswordResets deletes expired password resets
func (r *EmailRepository) DeleteExpiredPasswordResets(ctx context.Context) error {
	query := `DELETE FROM password_resets WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

// SMSRepository handles SMS-related database operations
type SMSRepository struct {
	db *sql.DB
}

// NewSMSRepository creates a new SMSRepository
func NewSMSRepository(db *sql.DB) *SMSRepository {
	return &SMSRepository{db: db}
}

// CreateSMSLog creates a new SMS log entry
func (r *SMSRepository) CreateSMSLog(ctx context.Context, log *models.SMSLog) error {
	query := `
		INSERT INTO sms_logs (id, to_phone, message, template, provider, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.ToPhone,
		log.Message,
		log.Template,
		log.Provider,
		log.Status,
		log.CreatedAt,
	)
	return err
}

// UpdateSMSLogStatus updates the status of an SMS log
func (r *SMSRepository) UpdateSMSLogStatus(ctx context.Context, logID uuid.UUID, status models.SMSStatus, errorMsg *string, providerID *string) error {
	query := `
		UPDATE sms_logs
		SET status = $1, error_msg = $2, provider_id = $3, sent_at = $4
		WHERE id = $5
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, status, errorMsg, providerID, now, logID)
	return err
}

// CreateSMSOTP creates a new SMS OTP
func (r *SMSRepository) CreateSMSOTP(ctx context.Context, otp *models.SMSOTP) error {
	query := `
		INSERT INTO sms_otps (id, user_id, phone, code, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		otp.ID,
		otp.UserID,
		otp.Phone,
		otp.Code,
		otp.ExpiresAt,
		otp.CreatedAt,
	)
	return err
}

// GetSMSOTP retrieves an SMS OTP by phone
func (r *SMSRepository) GetSMSOTP(ctx context.Context, phone string) (*models.SMSOTP, error) {
	query := `
		SELECT id, user_id, phone, code, expires_at, verified_at, created_at
		FROM sms_otps
		WHERE phone = $1 AND verified_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var otp models.SMSOTP
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&otp.ID,
		&otp.UserID,
		&otp.Phone,
		&otp.Code,
		&otp.ExpiresAt,
		&otp.VerifiedAt,
		&otp.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

// MarkSMSOTPVerified marks an SMS OTP as verified
func (r *SMSRepository) MarkSMSOTPVerified(ctx context.Context, otpID uuid.UUID) error {
	query := `
		UPDATE sms_otps
		SET verified_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), otpID)
	return err
}

// DeleteExpiredSMSOTPs deletes expired SMS OTPs
func (r *SMSRepository) DeleteExpiredSMSOTPs(ctx context.Context) error {
	query := `DELETE FROM sms_otps WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

// SocialRepository handles social login database operations
type SocialRepository struct {
	db *sql.DB
}

// NewSocialRepository creates a new SocialRepository
func NewSocialRepository(db *sql.DB) *SocialRepository {
	return &SocialRepository{db: db}
}

// CreateSocialAccount creates a new social account link
func (r *SocialRepository) CreateSocialAccount(ctx context.Context, account *models.SocialAccount) error {
	query := `
		INSERT INTO social_accounts (id, user_id, provider, provider_id, email, name, avatar, access_token, refresh_token, expires_at, linked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		account.ID,
		account.UserID,
		account.Provider,
		account.ProviderID,
		account.Email,
		account.Name,
		account.Avatar,
		account.AccessToken,
		account.RefreshToken,
		account.ExpiresAt,
		account.LinkedAt,
		account.CreatedAt,
	)
	return err
}

// GetSocialAccount retrieves a social account by provider and provider ID
func (r *SocialRepository) GetSocialAccount(ctx context.Context, provider models.SocialProvider, providerID string) (*models.SocialAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_id, email, name, avatar, access_token, refresh_token, expires_at, linked_at, last_used_at, created_at
		FROM social_accounts
		WHERE provider = $1 AND provider_id = $2
	`
	var account models.SocialAccount
	err := r.db.QueryRowContext(ctx, query, provider, providerID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderID,
		&account.Email,
		&account.Name,
		&account.Avatar,
		&account.AccessToken,
		&account.RefreshToken,
		&account.ExpiresAt,
		&account.LinkedAt,
		&account.LastUsedAt,
		&account.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetSocialAccountsByUser retrieves all social accounts for a user
func (r *SocialRepository) GetSocialAccountsByUser(ctx context.Context, userID uuid.UUID) ([]models.SocialAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_id, email, name, avatar, access_token, refresh_token, expires_at, linked_at, last_used_at, created_at
		FROM social_accounts
		WHERE user_id = $1
		ORDER BY linked_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.SocialAccount
	for rows.Next() {
		var account models.SocialAccount
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderID,
			&account.Email,
			&account.Name,
			&account.Avatar,
			&account.AccessToken,
			&account.RefreshToken,
			&account.ExpiresAt,
			&account.LinkedAt,
			&account.LastUsedAt,
			&account.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

// UpdateSocialAccountLastUsed updates the last used timestamp
func (r *SocialRepository) UpdateSocialAccountLastUsed(ctx context.Context, accountID uuid.UUID) error {
	query := `
		UPDATE social_accounts
		SET last_used_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), accountID)
	return err
}

// DeleteSocialAccount deletes a social account link
func (r *SocialRepository) DeleteSocialAccount(ctx context.Context, userID uuid.UUID, provider models.SocialProvider) error {
	query := `DELETE FROM social_accounts WHERE user_id = $1 AND provider = $2`
	_, err := r.db.ExecContext(ctx, query, userID, provider)
	return err
}

// CreateSocialLoginState creates a new OAuth state
func (r *SocialRepository) CreateSocialLoginState(ctx context.Context, state *models.SocialLoginState) error {
	query := `
		INSERT INTO social_login_states (id, state, provider, redirect_uri, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		state.ID,
		state.State,
		state.Provider,
		state.RedirectURI,
		state.ExpiresAt,
		state.CreatedAt,
	)
	return err
}

// GetSocialLoginState retrieves an OAuth state
func (r *SocialRepository) GetSocialLoginState(ctx context.Context, state string) (*models.SocialLoginState, error) {
	query := `
		SELECT id, state, provider, redirect_uri, expires_at, created_at
		FROM social_login_states
		WHERE state = $1
	`
	var loginState models.SocialLoginState
	err := r.db.QueryRowContext(ctx, query, state).Scan(
		&loginState.ID,
		&loginState.State,
		&loginState.Provider,
		&loginState.RedirectURI,
		&loginState.ExpiresAt,
		&loginState.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &loginState, nil
}

// DeleteSocialLoginState deletes an OAuth state
func (r *SocialRepository) DeleteSocialLoginState(ctx context.Context, state string) error {
	query := `DELETE FROM social_login_states WHERE state = $1`
	_, err := r.db.ExecContext(ctx, query, state)
	return err
}

// DeleteExpiredSocialLoginStates deletes expired OAuth states
func (r *SocialRepository) DeleteExpiredSocialLoginStates(ctx context.Context) error {
	query := `DELETE FROM social_login_states WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
