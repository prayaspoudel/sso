package models

import (
	"time"

	"github.com/google/uuid"
)

// SocialProvider represents social login providers
type SocialProvider string

const (
	SocialProviderGoogle   SocialProvider = "google"
	SocialProviderGitHub   SocialProvider = "github"
	SocialProviderLinkedIn SocialProvider = "linkedin"
	SocialProviderFacebook SocialProvider = "facebook"
	SocialProviderTwitter  SocialProvider = "twitter"
)

// SocialAccount represents a linked social account
type SocialAccount struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	UserID       uuid.UUID      `json:"userId" db:"user_id"`
	Provider     SocialProvider `json:"provider" db:"provider"`
	ProviderID   string         `json:"providerId" db:"provider_id"`
	Email        string         `json:"email" db:"email"`
	Name         string         `json:"name" db:"name"`
	Avatar       *string        `json:"avatar,omitempty" db:"avatar"`
	AccessToken  string         `json:"-" db:"access_token"`  // Encrypted
	RefreshToken *string        `json:"-" db:"refresh_token"` // Encrypted
	ExpiresAt    *time.Time     `json:"expiresAt,omitempty" db:"expires_at"`
	LinkedAt     time.Time      `json:"linkedAt" db:"linked_at"`
	LastUsedAt   *time.Time     `json:"lastUsedAt,omitempty" db:"last_used_at"`
	CreatedAt    time.Time      `json:"createdAt" db:"created_at"`
}

// SocialLoginState represents OAuth state for social login
type SocialLoginState struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	State       string         `json:"state" db:"state"`
	Provider    SocialProvider `json:"provider" db:"provider"`
	RedirectURI string         `json:"redirectUri" db:"redirect_uri"`
	ExpiresAt   time.Time      `json:"expiresAt" db:"expires_at"`
	CreatedAt   time.Time      `json:"createdAt" db:"created_at"`
}

// Social Login DTOs

// SocialLoginRequest represents a social login initiation
type SocialLoginRequest struct {
	Provider    SocialProvider `json:"provider" binding:"required"`
	RedirectURI string         `json:"redirectUri"`
}

// SocialLoginCallbackRequest represents OAuth callback data
type SocialLoginCallbackRequest struct {
	Provider SocialProvider `form:"provider" binding:"required"`
	Code     string         `form:"code" binding:"required"`
	State    string         `form:"state" binding:"required"`
}

// SocialLoginResponse represents social login response
type SocialLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         *User  `json:"user"`
	IsNewUser    bool   `json:"isNewUser"`
}

// LinkSocialAccountRequest represents linking a social account
type LinkSocialAccountRequest struct {
	Provider SocialProvider `json:"provider" binding:"required"`
	Code     string         `json:"code" binding:"required"`
	State    string         `json:"state" binding:"required"`
}

// UnlinkSocialAccountRequest represents unlinking a social account
type UnlinkSocialAccountRequest struct {
	Provider SocialProvider `json:"provider" binding:"required"`
}

// SocialUserInfo represents user info from social provider
type SocialUserInfo struct {
	Provider     SocialProvider `json:"provider"`
	ProviderID   string         `json:"providerId"`
	Email        string         `json:"email"`
	Name         string         `json:"name"`
	FirstName    string         `json:"firstName"`
	LastName     string         `json:"lastName"`
	Avatar       string         `json:"avatar"`
	AccessToken  string         `json:"-"`
	RefreshToken string         `json:"-"`
	ExpiresAt    *time.Time     `json:"expiresAt"`
}
