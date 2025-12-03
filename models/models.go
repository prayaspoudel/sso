package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	FirstName     string     `json:"firstName" db:"first_name"`
	LastName      string     `json:"lastName" db:"last_name"`
	CompanyID     uuid.UUID  `json:"companyId" db:"company_id"`
	Role          string     `json:"role" db:"role"`
	IsActive      bool       `json:"isActive" db:"is_active"`
	IsVerified    bool       `json:"isVerified" db:"is_verified"`
	EmailVerified bool       `json:"emailVerified" db:"email_verified"`
	LastLoginAt   *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
	LastLoginIP   string     `json:"lastLoginIp,omitempty" db:"last_login_ip"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" db:"updated_at"`
	LastLogin     *time.Time `json:"lastLogin,omitempty" db:"last_login"` // Keep for backward compatibility
}

type Company struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Industry  string    `json:"industry" db:"industry"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type UserCompany struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	CompanyID uuid.UUID `json:"companyId" db:"company_id"`
	Role      string    `json:"role" db:"role"`
	IsPrimary bool      `json:"isPrimary" db:"is_primary"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type Session struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"userId" db:"user_id"`
	SessionToken string    `json:"sessionToken" db:"session_token"`
	IPAddress    string    `json:"ipAddress" db:"ip_address"`
	UserAgent    string    `json:"userAgent" db:"user_agent"`
	ExpiresAt    time.Time `json:"expiresAt" db:"expires_at"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ClientID  string    `json:"clientId" db:"client_id"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	Revoked   bool      `json:"revoked" db:"revoked"`
}

type OAuthClient struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ClientID      string    `json:"clientId" db:"client_id"`
	ClientSecret  string    `json:"-" db:"client_secret"`
	Name          string    `json:"name" db:"name"`
	RedirectURIs  []string  `json:"redirectUris"`
	AllowedGrants []string  `json:"allowedGrants"`
	IsActive      bool      `json:"isActive" db:"is_active"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

type AuditLog struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	UserID    *uuid.UUID             `json:"userId,omitempty" db:"user_id"`
	Action    string                 `json:"action" db:"action"`
	Resource  string                 `json:"resource" db:"resource"`
	Details   map[string]interface{} `json:"details" db:"details"`
	IPAddress string                 `json:"ipAddress" db:"ip_address"`
	CreatedAt time.Time              `json:"createdAt" db:"created_at"`
}

type PasswordResetToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type EmailVerificationToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	Verified  bool      `json:"verified" db:"verified"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
