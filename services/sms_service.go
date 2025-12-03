package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"sso/models"
	"sso/repository"

	"github.com/google/uuid"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"golang.org/x/crypto/bcrypt"
)

// SMSService handles SMS operations
type SMSService struct {
	smsRepo      *repository.SMSRepository
	provider     models.SMSProvider
	twilioClient *twilio.RestClient
	twilioFrom   string
}

// SMSServiceConfig represents SMS service configuration
type SMSServiceConfig struct {
	Provider    models.SMSProvider
	TwilioSID   string
	TwilioToken string
	TwilioFrom  string
}

// NewSMSService creates a new SMSService
func NewSMSService(smsRepo *repository.SMSRepository, config SMSServiceConfig) *SMSService {
	service := &SMSService{
		smsRepo:  smsRepo,
		provider: config.Provider,
	}

	if config.Provider == models.SMSProviderTwilio {
		service.twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: config.TwilioSID,
			Password: config.TwilioToken,
		})
		service.twilioFrom = config.TwilioFrom
	}

	return service
}

// SendSMS sends an SMS using the configured provider
func (s *SMSService) SendSMS(ctx context.Context, to, message string, template models.SMSTemplate) error {
	// Create SMS log
	logID := uuid.New()
	log := &models.SMSLog{
		ID:        logID,
		ToPhone:   to,
		Message:   message,
		Template:  template,
		Provider:  s.provider,
		Status:    models.SMSStatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.smsRepo.CreateSMSLog(ctx, log); err != nil {
		return err
	}

	// Send SMS based on provider
	var err error
	var providerID *string
	switch s.provider {
	case models.SMSProviderTwilio:
		pid, sendErr := s.sendViaTwilio(to, message)
		err = sendErr
		providerID = pid
	default:
		err = errors.New("unsupported SMS provider")
	}

	// Update log status
	status := models.SMSStatusSent
	var errorMsg *string
	if err != nil {
		status = models.SMSStatusFailed
		msg := err.Error()
		errorMsg = &msg
	}

	s.smsRepo.UpdateSMSLogStatus(ctx, logID, status, errorMsg, providerID)
	return err
}

// sendViaTwilio sends SMS via Twilio
func (s *SMSService) sendViaTwilio(to, message string) (*string, error) {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(s.twilioFrom)
	params.SetBody(message)

	resp, err := s.twilioClient.Api.CreateMessage(params)
	if err != nil {
		return nil, err
	}

	sid := *resp.Sid
	return &sid, nil
}

// SendOTP generates and sends an OTP via SMS
func (s *SMSService) SendOTP(ctx context.Context, userID uuid.UUID, phone string) (string, error) {
	// Generate 6-digit OTP
	code := generateOTP(6)

	// Hash the OTP
	hashedCode, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Create OTP record
	otp := &models.SMSOTP{
		ID:        uuid.New(),
		UserID:    userID,
		Phone:     phone,
		Code:      string(hashedCode),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.smsRepo.CreateSMSOTP(ctx, otp); err != nil {
		return "", err
	}

	// Send SMS
	message := fmt.Sprintf("Your verification code is: %s. This code will expire in 5 minutes.", code)
	if err := s.SendSMS(ctx, phone, message, models.SMSTemplateOTP); err != nil {
		return "", err
	}

	return code, nil
}

// VerifyOTP verifies an OTP code
func (s *SMSService) VerifyOTP(ctx context.Context, phone, code string) (bool, error) {
	// Get OTP record
	otp, err := s.smsRepo.GetSMSOTP(ctx, phone)
	if err != nil {
		return false, err
	}
	if otp == nil {
		return false, errors.New("no OTP found for this phone number")
	}

	// Check expiration
	if time.Now().After(otp.ExpiresAt) {
		return false, errors.New("OTP expired")
	}

	// Verify code
	if err := bcrypt.CompareHashAndPassword([]byte(otp.Code), []byte(code)); err != nil {
		return false, nil
	}

	// Mark as verified
	if err := s.smsRepo.MarkSMSOTPVerified(ctx, otp.ID); err != nil {
		return false, err
	}

	return true, nil
}

// SendVerificationCode sends a verification code via SMS
func (s *SMSService) SendVerificationCode(ctx context.Context, phone string) error {
	code := generateOTP(6)
	message := fmt.Sprintf("Your verification code is: %s", code)
	return s.SendSMS(ctx, phone, message, models.SMSTemplateVerification)
}

// SendPasswordResetCode sends a password reset code via SMS
func (s *SMSService) SendPasswordResetCode(ctx context.Context, phone string) error {
	code := generateOTP(6)
	message := fmt.Sprintf("Your password reset code is: %s. Do not share this code.", code)
	return s.SendSMS(ctx, phone, message, models.SMSTemplatePasswordReset)
}

// SendLoginAlert sends a login alert via SMS
func (s *SMSService) SendLoginAlert(ctx context.Context, phone, ipAddress string) error {
	message := fmt.Sprintf("New login detected from IP: %s. If this wasn't you, please secure your account immediately.", ipAddress)
	return s.SendSMS(ctx, phone, message, models.SMSTemplateLoginAlert)
}

// generateOTP generates a random numeric OTP
func generateOTP(length int) string {
	digits := "0123456789"
	otp := make([]byte, length)
	for i := range otp {
		otp[i] = digits[rand.Intn(len(digits))]
	}
	return string(otp)
}
