package models

import (
	"time"

	"github.com/google/uuid"
)

// TwoFactorMethod represents the type of 2FA method
type TwoFactorMethod string

const (
	TwoFactorMethodTOTP TwoFactorMethod = "totp"
	TwoFactorMethodSMS  TwoFactorMethod = "sms"
)

// TwoFactorStatus represents the status of 2FA for a user
type TwoFactorStatus string

const (
	TwoFactorStatusDisabled TwoFactorStatus = "disabled"
	TwoFactorStatusPending  TwoFactorStatus = "pending"
	TwoFactorStatusEnabled  TwoFactorStatus = "enabled"
)

// UserTwoFactor represents a user's 2FA settings
type UserTwoFactor struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	UserID           uuid.UUID       `json:"userId" db:"user_id"`
	Method           TwoFactorMethod `json:"method" db:"method"`
	Secret           string          `json:"-" db:"secret"` // TOTP secret (never expose in JSON)
	PhoneNumber      *string         `json:"phoneNumber,omitempty" db:"phone_number"`
	Status           TwoFactorStatus `json:"status" db:"status"`
	BackupCodesCount int             `json:"backupCodesCount" db:"backup_codes_count"`
	VerifiedAt       *time.Time      `json:"verifiedAt,omitempty" db:"verified_at"`
	CreatedAt        time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time       `json:"updatedAt" db:"updated_at"`
}

// BackupCode represents a 2FA backup code
type BackupCode struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"userId" db:"user_id"`
	Code      string     `json:"code" db:"code"` // Hashed in database
	UsedAt    *time.Time `json:"usedAt,omitempty" db:"used_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
}

// TwoFactorSetupResponse contains QR code and backup codes for 2FA setup
type TwoFactorSetupResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qrCodeUrl"`
	BackupCodes []string `json:"backupCodes"`
}

// VerifyTOTPRequest represents a TOTP verification request
type VerifyTOTPRequest struct {
	Code string `json:"code" binding:"required"`
}

// Enable2FARequest represents a request to enable 2FA
type Enable2FARequest struct {
	Method TwoFactorMethod `json:"method" binding:"required"`
	Code   string          `json:"code" binding:"required"`
}
