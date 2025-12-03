package models

import (
	"time"

	"github.com/google/uuid"
)

// OAuth2 Grant Types
const (
	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeRefreshToken      = "refresh_token"
	GrantTypeClientCredentials = "client_credentials"
)

// OAuth2 Response Types
const (
	ResponseTypeCode = "code"
)

// OAuth2Client represents an OAuth2 application
type OAuth2Client struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ClientID     string    `json:"clientId" db:"client_id"`
	ClientSecret string    `json:"-" db:"client_secret"` // Hashed, never expose
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description,omitempty" db:"description"`
	RedirectURIs []string  `json:"redirectUris" db:"redirect_uris"` // PostgreSQL array
	GrantTypes   []string  `json:"grantTypes" db:"grant_types"`     // PostgreSQL array
	Scopes       []string  `json:"scopes" db:"scopes"`              // PostgreSQL array
	OwnerID      uuid.UUID `json:"ownerId" db:"owner_id"`
	LogoURL      *string   `json:"logoUrl,omitempty" db:"logo_url"`
	Active       bool      `json:"active" db:"active"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// OAuth2AuthorizationCode represents an authorization code
type OAuth2AuthorizationCode struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Code        string     `json:"code" db:"code"`
	ClientID    string     `json:"clientId" db:"client_id"`
	UserID      uuid.UUID  `json:"userId" db:"user_id"`
	RedirectURI string     `json:"redirectUri" db:"redirect_uri"`
	Scopes      []string   `json:"scopes" db:"scopes"`
	ExpiresAt   time.Time  `json:"expiresAt" db:"expires_at"`
	UsedAt      *time.Time `json:"usedAt,omitempty" db:"used_at"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
}

// OAuth2Token represents an OAuth2 access token
type OAuth2Token struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	AccessToken  string     `json:"accessToken" db:"access_token"`
	RefreshToken *string    `json:"refreshToken,omitempty" db:"refresh_token"`
	ClientID     string     `json:"clientId" db:"client_id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	Scopes       []string   `json:"scopes" db:"scopes"`
	ExpiresAt    time.Time  `json:"expiresAt" db:"expires_at"`
	RevokedAt    *time.Time `json:"revokedAt,omitempty" db:"revoked_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

// OAuth2Scope represents available OAuth2 scopes
type OAuth2Scope struct {
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

// Standard OAuth2 scopes
var StandardScopes = []OAuth2Scope{
	{Name: "openid", Description: "OpenID Connect authentication"},
	{Name: "profile", Description: "Access to user profile information"},
	{Name: "email", Description: "Access to user email address"},
	{Name: "offline_access", Description: "Request refresh token"},
}

// OAuth2 Request/Response DTOs

// AuthorizeRequest represents an OAuth2 authorization request
type AuthorizeRequest struct {
	ResponseType string `form:"response_type" binding:"required"`
	ClientID     string `form:"client_id" binding:"required"`
	RedirectURI  string `form:"redirect_uri" binding:"required"`
	Scope        string `form:"scope"`
	State        string `form:"state"`
}

// AuthorizeResponse represents an OAuth2 authorization response
type AuthorizeResponse struct {
	Code  string `json:"code"`
	State string `json:"state,omitempty"`
}

// TokenRequest represents an OAuth2 token request
type TokenRequest struct {
	GrantType    string `form:"grant_type" binding:"required"`
	Code         string `form:"code"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" binding:"required"`
	RefreshToken string `form:"refresh_token"`
	Scope        string `form:"scope"`
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// CreateOAuth2ClientRequest represents a request to create an OAuth2 client
type CreateOAuth2ClientRequest struct {
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description"`
	RedirectURIs []string `json:"redirectUris" binding:"required,min=1"`
	GrantTypes   []string `json:"grantTypes" binding:"required,min=1"`
	Scopes       []string `json:"scopes" binding:"required,min=1"`
	LogoURL      string   `json:"logoUrl"`
}

// OAuth2ClientResponse represents an OAuth2 client response with secret
type OAuth2ClientResponse struct {
	Client       *OAuth2Client `json:"client"`
	ClientSecret string        `json:"clientSecret,omitempty"` // Only returned on creation
}
