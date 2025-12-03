package models

import (
	"time"

	"github.com/google/uuid"
)

// SMSProvider represents the SMS service provider
type SMSProvider string

const (
	SMSProviderTwilio SMSProvider = "twilio"
	SMSProviderAWSSNS SMSProvider = "aws_sns"
)

// SMSTemplate represents SMS template types
type SMSTemplate string

const (
	SMSTemplateOTP           SMSTemplate = "otp"
	SMSTemplateVerification  SMSTemplate = "verification"
	SMSTemplatePasswordReset SMSTemplate = "password_reset"
	SMSTemplateLoginAlert    SMSTemplate = "login_alert"
)

// SMSStatus represents the status of an SMS
type SMSStatus string

const (
	SMSStatusPending   SMSStatus = "pending"
	SMSStatusSent      SMSStatus = "sent"
	SMSStatusDelivered SMSStatus = "delivered"
	SMSStatusFailed    SMSStatus = "failed"
)

// SMSLog represents an SMS sending log
type SMSLog struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	ToPhone     string      `json:"toPhone" db:"to_phone"`
	Message     string      `json:"message" db:"message"`
	Template    SMSTemplate `json:"template" db:"template"`
	Provider    SMSProvider `json:"provider" db:"provider"`
	Status      SMSStatus   `json:"status" db:"status"`
	ErrorMsg    *string     `json:"errorMsg,omitempty" db:"error_msg"`
	ProviderID  *string     `json:"providerId,omitempty" db:"provider_id"`
	SentAt      *time.Time  `json:"sentAt,omitempty" db:"sent_at"`
	DeliveredAt *time.Time  `json:"deliveredAt,omitempty" db:"delivered_at"`
	CreatedAt   time.Time   `json:"createdAt" db:"created_at"`
}

// SMSOTP represents an SMS OTP token
type SMSOTP struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"userId" db:"user_id"`
	Phone      string     `json:"phone" db:"phone"`
	Code       string     `json:"-" db:"code"` // Hashed
	ExpiresAt  time.Time  `json:"expiresAt" db:"expires_at"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty" db:"verified_at"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
}

// SMS DTOs

// SendSMSRequest represents a request to send an SMS
type SendSMSRequest struct {
	To       string                 `json:"to" binding:"required"`
	Template SMSTemplate            `json:"template" binding:"required"`
	Data     map[string]interface{} `json:"data"`
}

// SendOTPRequest represents a request to send an OTP via SMS
type SendOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// VerifyOTPRequest represents an OTP verification request
type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}
