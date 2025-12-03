package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"sso/models"
	"sso/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// OAuth2Service handles OAuth2 operations
type OAuth2Service struct {
	oauth2Repo *repository.OAuth2Repository
	userRepo   *repository.UserRepository
	jwtSecret  string
}

// NewOAuth2Service creates a new OAuth2Service
func NewOAuth2Service(oauth2Repo *repository.OAuth2Repository, userRepo *repository.UserRepository, jwtSecret string) *OAuth2Service {
	return &OAuth2Service{
		oauth2Repo: oauth2Repo,
		userRepo:   userRepo,
		jwtSecret:  jwtSecret,
	}
}

// CreateClient creates a new OAuth2 client
func (s *OAuth2Service) CreateClient(ctx context.Context, req *models.CreateOAuth2ClientRequest, ownerID uuid.UUID) (*models.OAuth2ClientResponse, error) {
	// Generate client ID and secret
	clientID, err := generateClientID()
	if err != nil {
		return nil, err
	}

	clientSecret, err := generateClientSecret()
	if err != nil {
		return nil, err
	}

	// Hash the client secret
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Validate grant types
	for _, grantType := range req.GrantTypes {
		if grantType != models.GrantTypeAuthorizationCode &&
			grantType != models.GrantTypeRefreshToken &&
			grantType != models.GrantTypeClientCredentials {
			return nil, fmt.Errorf("invalid grant type: %s", grantType)
		}
	}

	// Validate scopes
	validScopes := map[string]bool{
		"openid": true, "profile": true, "email": true, "offline_access": true,
	}
	for _, scope := range req.Scopes {
		if !validScopes[scope] {
			return nil, fmt.Errorf("invalid scope: %s", scope)
		}
	}

	now := time.Now()
	client := &models.OAuth2Client{
		ID:           uuid.New(),
		ClientID:     clientID,
		ClientSecret: string(hashedSecret),
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		GrantTypes:   req.GrantTypes,
		Scopes:       req.Scopes,
		OwnerID:      ownerID,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if req.Description != "" {
		client.Description = &req.Description
	}
	if req.LogoURL != "" {
		client.LogoURL = &req.LogoURL
	}

	if err := s.oauth2Repo.CreateClient(ctx, client); err != nil {
		return nil, err
	}

	// Return client with unhashed secret (only time it's visible)
	return &models.OAuth2ClientResponse{
		Client:       client,
		ClientSecret: clientSecret,
	}, nil
}

// Authorize handles OAuth2 authorization request
func (s *OAuth2Service) Authorize(ctx context.Context, req *models.AuthorizeRequest, userID uuid.UUID) (*models.AuthorizeResponse, error) {
	// Validate response type
	if req.ResponseType != models.ResponseTypeCode {
		return nil, errors.New("unsupported response type")
	}

	// Get client
	client, err := s.oauth2Repo.GetClientByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("invalid client")
	}
	if !client.Active {
		return nil, errors.New("client is inactive")
	}

	// Validate redirect URI
	validRedirectURI := false
	for _, uri := range client.RedirectURIs {
		if uri == req.RedirectURI {
			validRedirectURI = true
			break
		}
	}
	if !validRedirectURI {
		return nil, errors.New("invalid redirect URI")
	}

	// Validate scopes
	requestedScopes := strings.Split(req.Scope, " ")
	for _, scope := range requestedScopes {
		if scope == "" {
			continue
		}
		validScope := false
		for _, clientScope := range client.Scopes {
			if clientScope == scope {
				validScope = true
				break
			}
		}
		if !validScope {
			return nil, fmt.Errorf("invalid scope: %s", scope)
		}
	}

	// Generate authorization code
	code, err := generateAuthorizationCode()
	if err != nil {
		return nil, err
	}

	// Store authorization code
	authCode := &models.OAuth2AuthorizationCode{
		ID:          uuid.New(),
		Code:        code,
		ClientID:    req.ClientID,
		UserID:      userID,
		RedirectURI: req.RedirectURI,
		Scopes:      requestedScopes,
		ExpiresAt:   time.Now().Add(10 * time.Minute), // 10 minutes expiry
		CreatedAt:   time.Now(),
	}

	if err := s.oauth2Repo.CreateAuthorizationCode(ctx, authCode); err != nil {
		return nil, err
	}

	return &models.AuthorizeResponse{
		Code:  code,
		State: req.State,
	}, nil
}

// ExchangeToken exchanges authorization code for access token
func (s *OAuth2Service) ExchangeToken(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	switch req.GrantType {
	case models.GrantTypeAuthorizationCode:
		return s.exchangeAuthorizationCode(ctx, req)
	case models.GrantTypeRefreshToken:
		return s.refreshToken(ctx, req)
	default:
		return nil, errors.New("unsupported grant type")
	}
}

// exchangeAuthorizationCode exchanges authorization code for tokens
func (s *OAuth2Service) exchangeAuthorizationCode(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	// Get authorization code
	authCode, err := s.oauth2Repo.GetAuthorizationCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if authCode == nil {
		return nil, errors.New("invalid authorization code")
	}

	// Check if already used
	if authCode.UsedAt != nil {
		return nil, errors.New("authorization code already used")
	}

	// Check expiration
	if time.Now().After(authCode.ExpiresAt) {
		return nil, errors.New("authorization code expired")
	}

	// Verify client
	client, err := s.oauth2Repo.GetClientByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("invalid client")
	}

	// Verify client secret
	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecret), []byte(req.ClientSecret)); err != nil {
		return nil, errors.New("invalid client credentials")
	}

	// Verify redirect URI
	if authCode.RedirectURI != req.RedirectURI {
		return nil, errors.New("redirect URI mismatch")
	}

	// Mark code as used
	if err := s.oauth2Repo.MarkAuthorizationCodeUsed(ctx, authCode.ID); err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(authCode.UserID, authCode.ClientID, authCode.Scopes)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store token
	token := &models.OAuth2Token{
		ID:           uuid.New(),
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		ClientID:     req.ClientID,
		UserID:       authCode.UserID,
		Scopes:       authCode.Scopes,
		ExpiresAt:    time.Now().Add(1 * time.Hour), // 1 hour expiry
		CreatedAt:    time.Now(),
	}

	if err := s.oauth2Repo.CreateToken(ctx, token); err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		RefreshToken: refreshToken,
		Scope:        strings.Join(authCode.Scopes, " "),
	}, nil
}

// refreshToken refreshes an access token using refresh token
func (s *OAuth2Service) refreshToken(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	// Get token by refresh token
	token, err := s.oauth2Repo.GetTokenByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if revoked
	if token.RevokedAt != nil {
		return nil, errors.New("token has been revoked")
	}

	// Verify client
	client, err := s.oauth2Repo.GetClientByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("invalid client")
	}

	// Verify client secret
	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecret), []byte(req.ClientSecret)); err != nil {
		return nil, errors.New("invalid client credentials")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(token.UserID, token.ClientID, token.Scopes)
	if err != nil {
		return nil, err
	}

	// Create new token record
	newToken := &models.OAuth2Token{
		ID:           uuid.New(),
		AccessToken:  accessToken,
		RefreshToken: token.RefreshToken,
		ClientID:     token.ClientID,
		UserID:       token.UserID,
		Scopes:       token.Scopes,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := s.oauth2Repo.CreateToken(ctx, newToken); err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: *token.RefreshToken,
		Scope:        strings.Join(token.Scopes, " "),
	}, nil
}

// ValidateAccessToken validates an OAuth2 access token
func (s *OAuth2Service) ValidateAccessToken(ctx context.Context, accessToken string) (*models.OAuth2Token, error) {
	token, err := s.oauth2Repo.GetTokenByAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("invalid access token")
	}

	// Check expiration
	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("access token expired")
	}

	// Check if revoked
	if token.RevokedAt != nil {
		return nil, errors.New("token has been revoked")
	}

	return token, nil
}

// RevokeToken revokes an OAuth2 token
func (s *OAuth2Service) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	return s.oauth2Repo.RevokeToken(ctx, tokenID)
}

// ListClientsByOwner lists all OAuth2 clients owned by a user
func (s *OAuth2Service) ListClientsByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.OAuth2Client, error) {
	clients, err := s.oauth2Repo.ListClientsByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	// Remove sensitive data
	for i := range clients {
		clients[i].ClientSecret = ""
	}

	return clients, nil
}

// generateAccessToken generates a JWT access token
func (s *OAuth2Service) generateAccessToken(userID uuid.UUID, clientID string, scopes []string) (string, error) {
	claims := jwt.MapClaims{
		"sub":       userID.String(),
		"client_id": clientID,
		"scopes":    scopes,
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateClientID generates a random client ID
func generateClientID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateClientSecret generates a random client secret
func generateClientSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateAuthorizationCode generates a random authorization code
func generateAuthorizationCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	hash := sha256.Sum256(b)
	return base64.URLEncoding.EncodeToString(hash[:]), nil
}

// generateRefreshToken generates a random refresh token
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
