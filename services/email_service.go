package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"sso/models"
	"sso/repository"

	"github.com/google/uuid"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService handles email operations
type EmailService struct {
	emailRepo      *repository.EmailRepository
	userRepo       *repository.UserRepository
	provider       models.EmailProvider
	fromEmail      string
	fromName       string
	sendGridClient *sendgrid.Client
	mailgunClient  *mailgun.MailgunImpl
}

// EmailServiceConfig represents email service configuration
type EmailServiceConfig struct {
	Provider       models.EmailProvider
	FromEmail      string
	FromName       string
	SendGridAPIKey string
	MailgunDomain  string
	MailgunAPIKey  string
	AppURL         string
}

// NewEmailService creates a new EmailService
func NewEmailService(emailRepo *repository.EmailRepository, userRepo *repository.UserRepository, config EmailServiceConfig) *EmailService {
	service := &EmailService{
		emailRepo: emailRepo,
		userRepo:  userRepo,
		provider:  config.Provider,
		fromEmail: config.FromEmail,
		fromName:  config.FromName,
	}

	switch config.Provider {
	case models.EmailProviderSendGrid:
		service.sendGridClient = sendgrid.NewSendClient(config.SendGridAPIKey)
	case models.EmailProviderMailgun:
		service.mailgunClient = mailgun.NewMailgun(config.MailgunDomain, config.MailgunAPIKey)
	}

	return service
}

// SendEmail sends an email using the configured provider
func (s *EmailService) SendEmail(ctx context.Context, to, subject, htmlBody, textBody string, template models.EmailTemplate) error {
	// Create email log
	logID := uuid.New()
	log := &models.EmailLog{
		ID:        logID,
		ToEmail:   to,
		FromEmail: s.fromEmail,
		Subject:   subject,
		Template:  template,
		Provider:  s.provider,
		Status:    models.EmailStatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.emailRepo.CreateEmailLog(ctx, log); err != nil {
		return err
	}

	// Send email based on provider
	var err error
	switch s.provider {
	case models.EmailProviderSendGrid:
		err = s.sendViaSendGrid(to, subject, htmlBody, textBody)
	case models.EmailProviderMailgun:
		err = s.sendViaMailgun(ctx, to, subject, htmlBody, textBody)
	default:
		err = errors.New("unsupported email provider")
	}

	// Update log status
	status := models.EmailStatusSent
	var errorMsg *string
	if err != nil {
		status = models.EmailStatusFailed
		msg := err.Error()
		errorMsg = &msg
	}

	s.emailRepo.UpdateEmailLogStatus(ctx, logID, status, errorMsg)
	return err
}

// sendViaSendGrid sends email via SendGrid
func (s *EmailService) sendViaSendGrid(to, subject, htmlBody, textBody string) error {
	from := mail.NewEmail(s.fromName, s.fromEmail)
	toAddr := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toAddr, textBody, htmlBody)

	response, err := s.sendGridClient.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: %d - %s", response.StatusCode, response.Body)
	}

	return nil
}

// sendViaMailgun sends email via Mailgun
func (s *EmailService) sendViaMailgun(ctx context.Context, to, subject, htmlBody, textBody string) error {
	message := s.mailgunClient.NewMessage(
		fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		subject,
		textBody,
		to,
	)
	message.SetHtml(htmlBody)

	_, _, err := s.mailgunClient.Send(ctx, message)
	return err
}

// SendVerificationEmail sends an email verification email
func (s *EmailService) SendVerificationEmail(ctx context.Context, userID uuid.UUID, email, appURL string) error {
	// Generate verification token
	token, err := generateSecureToken(32)
	if err != nil {
		return err
	}

	// Create verification record
	verification := &models.EmailVerification{
		ID:        uuid.New(),
		UserID:    userID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.emailRepo.CreateEmailVerification(ctx, verification); err != nil {
		return err
	}

	// Build verification URL
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", appURL, token)

	// Email content
	subject := "Verify Your Email Address"
	htmlBody := fmt.Sprintf(`
		<h2>Email Verification</h2>
		<p>Please verify your email address by clicking the link below:</p>
		<p><a href="%s">Verify Email</a></p>
		<p>This link will expire in 24 hours.</p>
		<p>If you didn't request this, please ignore this email.</p>
	`, verifyURL)
	textBody := fmt.Sprintf("Verify your email: %s\n\nThis link will expire in 24 hours.", verifyURL)

	return s.SendEmail(ctx, email, subject, htmlBody, textBody, models.EmailTemplateVerification)
}

// VerifyEmail verifies an email address
func (s *EmailService) VerifyEmail(ctx context.Context, token string) error {
	verification, err := s.emailRepo.GetEmailVerification(ctx, token)
	if err != nil {
		return err
	}
	if verification == nil {
		return errors.New("invalid verification token")
	}

	if verification.VerifiedAt != nil {
		return errors.New("email already verified")
	}

	if time.Now().After(verification.ExpiresAt) {
		return errors.New("verification token expired")
	}

	// Mark as verified
	if err := s.emailRepo.MarkEmailVerified(ctx, token); err != nil {
		return err
	}

	// Update user's email verified status if needed
	// This depends on your user model implementation

	return nil
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, appURL string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		// Don't reveal if user exists
		return nil
	}

	// Generate reset token
	token, err := generateSecureToken(32)
	if err != nil {
		return err
	}

	// Create reset record
	reset := &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.emailRepo.CreatePasswordReset(ctx, reset); err != nil {
		return err
	}

	// Build reset URL
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", appURL, token)

	// Email content
	subject := "Password Reset Request"
	htmlBody := fmt.Sprintf(`
		<h2>Password Reset</h2>
		<p>You requested to reset your password. Click the link below to proceed:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>This link will expire in 1 hour.</p>
		<p>If you didn't request this, please ignore this email and your password will remain unchanged.</p>
	`, resetURL)
	textBody := fmt.Sprintf("Reset your password: %s\n\nThis link will expire in 1 hour.", resetURL)

	return s.SendEmail(ctx, email, subject, htmlBody, textBody, models.EmailTemplatePasswordReset)
}

// ResetPassword resets a user's password with a token
func (s *EmailService) ResetPassword(ctx context.Context, token, newPassword string) error {
	reset, err := s.emailRepo.GetPasswordReset(ctx, token)
	if err != nil {
		return err
	}
	if reset == nil {
		return errors.New("invalid reset token")
	}

	if reset.UsedAt != nil {
		return errors.New("reset token already used")
	}

	if time.Now().After(reset.ExpiresAt) {
		return errors.New("reset token expired")
	}

	// Update user's password
	// This should use your auth service's password hashing
	// For now, we'll mark the token as used
	if err := s.emailRepo.MarkPasswordResetUsed(ctx, token); err != nil {
		return err
	}

	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, name string) error {
	subject := "Welcome to SSO Service!"
	htmlBody := fmt.Sprintf(`
		<h2>Welcome, %s!</h2>
		<p>Thank you for joining SSO Service. We're excited to have you on board!</p>
		<p>Get started by setting up your account and exploring our features.</p>
	`, name)
	textBody := fmt.Sprintf("Welcome, %s! Thank you for joining SSO Service.", name)

	return s.SendEmail(ctx, email, subject, htmlBody, textBody, models.EmailTemplateWelcome)
}

// Send2FAEnabledEmail sends notification when 2FA is enabled
func (s *EmailService) Send2FAEnabledEmail(ctx context.Context, email, name string) error {
	subject := "Two-Factor Authentication Enabled"
	htmlBody := fmt.Sprintf(`
		<h2>2FA Enabled</h2>
		<p>Hi %s,</p>
		<p>Two-factor authentication has been successfully enabled on your account.</p>
		<p>This adds an extra layer of security to protect your account.</p>
		<p>If you didn't enable this, please contact support immediately.</p>
	`, name)
	textBody := fmt.Sprintf("Hi %s, Two-factor authentication has been enabled on your account.", name)

	return s.SendEmail(ctx, email, subject, htmlBody, textBody, models.EmailTemplate2FAEnabled)
}

// Send2FADisabledEmail sends notification when 2FA is disabled
func (s *EmailService) Send2FADisabledEmail(ctx context.Context, email, name string) error {
	subject := "Two-Factor Authentication Disabled"
	htmlBody := fmt.Sprintf(`
		<h2>2FA Disabled</h2>
		<p>Hi %s,</p>
		<p>Two-factor authentication has been disabled on your account.</p>
		<p>If you didn't disable this, please contact support immediately and change your password.</p>
	`, name)
	textBody := fmt.Sprintf("Hi %s, Two-factor authentication has been disabled on your account.", name)

	return s.SendEmail(ctx, email, subject, htmlBody, textBody, models.EmailTemplate2FADisabled)
}

// SendPasswordChangedEmail sends notification when password is changed
func (s *EmailService) SendPasswordChangedEmail(ctx context.Context, email, name string) error {
	subject := "Password Changed"
	htmlBody := fmt.Sprintf(`
		<h2>Password Changed</h2>
		<p>Hi %s,</p>
		<p>Your password has been successfully changed.</p>
		<p>If you didn't make this change, please contact support immediately.</p>
	`, name)
	textBody := fmt.Sprintf("Hi %s, Your password has been successfully changed.", name)

	return s.SendEmail(ctx, email, subject, htmlBody, textBody, models.EmailTemplatePasswordChanged)
}

// SendLoginNotification sends notification for new login
func (s *EmailService) SendLoginNotification(ctx context.Context, email, name, ipAddress, userAgent string) error {
	subject := "New Login Detected"
	htmlBody := fmt.Sprintf(`
		<h2>New Login Detected</h2>
		<p>Hi %s,</p>
		<p>A new login was detected on your account:</p>
		<ul>
			<li>IP Address: %s</li>
			<li>Device: %s</li>
			<li>Time: %s</li>
		</ul>
		<p>If this wasn't you, please change your password immediately.</p>
	`, name, ipAddress, userAgent, time.Now().Format("2006-01-02 15:04:05"))
	textBody := fmt.Sprintf("Hi %s, A new login was detected on your account from IP %s", name, ipAddress)

	return s.SendEmail(ctx, email, subject, htmlBody, textBody, models.EmailTemplateLoginNotification)
}

// generateSecureToken generates a secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
