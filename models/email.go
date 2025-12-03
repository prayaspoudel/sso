package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailProvider represents the email service provider
type EmailProvider string

const (
	EmailProviderSendGrid EmailProvider = "sendgrid"
	EmailProviderMailgun  EmailProvider = "mailgun"
	EmailProviderSMTP     EmailProvider = "smtp"
)

// EmailTemplate represents email template types
type EmailTemplate string

const (
	EmailTemplateVerification      EmailTemplate = "verification"
	EmailTemplatePasswordReset     EmailTemplate = "password_reset"
	EmailTemplateWelcome           EmailTemplate = "welcome"
	EmailTemplate2FAEnabled        EmailTemplate = "2fa_enabled"
	EmailTemplate2FADisabled       EmailTemplate = "2fa_disabled"
	EmailTemplatePasswordChanged   EmailTemplate = "password_changed"
	EmailTemplateLoginNotification EmailTemplate = "login_notification"
)

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending EmailStatus = "pending"
	EmailStatusSent    EmailStatus = "sent"
	EmailStatusFailed  EmailStatus = "failed"
)

// EmailLog represents an email sending log
type EmailLog struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	ToEmail   string        `json:"toEmail" db:"to_email"`
	FromEmail string        `json:"fromEmail" db:"from_email"`
	Subject   string        `json:"subject" db:"subject"`
	Template  EmailTemplate `json:"template" db:"template"`
	Provider  EmailProvider `json:"provider" db:"provider"`
	Status    EmailStatus   `json:"status" db:"status"`
	ErrorMsg  *string       `json:"errorMsg,omitempty" db:"error_msg"`
	SentAt    *time.Time    `json:"sentAt,omitempty" db:"sent_at"`
	CreatedAt time.Time     `json:"createdAt" db:"created_at"`
}

// EmailVerification represents an email verification token
type EmailVerification struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"userId" db:"user_id"`
	Email      string     `json:"email" db:"email"`
	Token      string     `json:"token" db:"token"`
	ExpiresAt  time.Time  `json:"expiresAt" db:"expires_at"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty" db:"verified_at"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
}

// PasswordReset represents a password reset token
type PasswordReset struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"userId" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expiresAt" db:"expires_at"`
	UsedAt    *time.Time `json:"usedAt,omitempty" db:"used_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
}

// Email DTOs

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To       string                 `json:"to" binding:"required,email"`
	Subject  string                 `json:"subject" binding:"required"`
	Template EmailTemplate          `json:"template" binding:"required"`
	Data     map[string]interface{} `json:"data"`
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// RequestPasswordResetRequest represents a password reset request
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a password reset with token
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}
