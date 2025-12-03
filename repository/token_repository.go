package repository

import (
	"context"
	"database/sql"

	"sso/models"

	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, client_id, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.Token, token.ClientID,
		token.ExpiresAt, token.CreatedAt, token.Revoked,
	)
	return err
}

func (r *TokenRepository) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, client_id, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token = $1
	`
	rt := &models.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ClientID,
		&rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked,
	)
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *TokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = true`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *TokenRepository) GetUserTokens(ctx context.Context, userID uuid.UUID) ([]models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, client_id, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE user_id = $1 AND revoked = false AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.RefreshToken
	for rows.Next() {
		var token models.RefreshToken
		if err := rows.Scan(
			&token.ID, &token.UserID, &token.Token, &token.ClientID,
			&token.ExpiresAt, &token.CreatedAt, &token.Revoked,
		); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}
