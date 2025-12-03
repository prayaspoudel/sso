# Phase 3: External Services Implementation

This document provides a complete guide to the External Services features implemented in Phase 3, including Email, SMS, and Social Login capabilities.

## Overview

Phase 3 adds comprehensive external service integrations to the SSO system:

- **Email Services**: Multi-provider email delivery with SendGrid and Mailgun support
- **SMS Services**: Twilio-powered SMS with OTP functionality
- **Social Login**: OAuth2 integration with Google, GitHub, and LinkedIn

## Table of Contents

1. [Architecture](#architecture)
2. [Database Schema](#database-schema)
3. [Email Services](#email-services)
4. [SMS Services](#sms-services)
5. [Social Login](#social-login)
6. [API Endpoints](#api-endpoints)
7. [Configuration](#configuration)
8. [Testing](#testing)

## Architecture

### Component Structure

```
Phase 3 Components:
├── models/
│   ├── email.go          # Email models and templates
│   ├── sms.go            # SMS models and templates
│   └── social.go         # Social login models
├── repository/
│   └── external_services_repository.go  # Combined repository
├── services/
│   ├── email_service.go   # Email provider abstraction
│   ├── sms_service.go     # SMS provider abstraction
│   └── social_service.go  # OAuth2 social login
├── handlers/
│   └── external_services_handler.go  # HTTP endpoints
└── database/migrations/
    ├── 005_external_services.sql  # Schema
    └── 005_rollback.sql          # Rollback
```

### Design Patterns

1. **Provider Pattern**: Abstract email/SMS providers for easy switching
2. **Template System**: Predefined templates for consistent messaging
3. **OAuth2 Flow**: Standard authorization code flow for social login
4. **Token Management**: Secure token generation and verification

## Database Schema

### Email Tables

#### email_logs
Tracks all email delivery attempts and statuses.

```sql
CREATE TABLE email_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    to_email VARCHAR(255) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    message_id VARCHAR(255),
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    bounced_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

**Status Values**: `pending`, `sent`, `failed`, `bounced`

#### email_verifications
Manages email verification tokens.

```sql
CREATE TABLE email_verifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

#### password_resets
Stores password reset tokens.

```sql
CREATE TABLE password_resets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL
);
```

### SMS Tables

#### sms_logs
Tracks all SMS delivery attempts.

```sql
CREATE TABLE sms_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    to_phone VARCHAR(20) NOT NULL,
    from_phone VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    message_id VARCHAR(255),
    segments INTEGER DEFAULT 1,
    cost DECIMAL(10, 4),
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    failed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

**Status Values**: `pending`, `sent`, `failed`, `delivered`

#### sms_otps
Manages SMS OTP codes.

```sql
CREATE TABLE sms_otps (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    phone VARCHAR(20) NOT NULL,
    code_hash VARCHAR(255) NOT NULL,
    purpose VARCHAR(100) NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

### Social Login Tables

#### social_accounts
Links users with social media accounts.

```sql
CREATE TABLE social_accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    name VARCHAR(255),
    avatar TEXT,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    linked_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(provider, provider_id)
);
```

**Providers**: `google`, `github`, `linkedin`, `facebook`, `twitter`

#### social_login_states
Temporary OAuth state storage.

```sql
CREATE TABLE social_login_states (
    id UUID PRIMARY KEY,
    state VARCHAR(255) NOT NULL UNIQUE,
    provider VARCHAR(50) NOT NULL,
    redirect_uri TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

## Email Services

### Features

1. **Multi-Provider Support**: SendGrid and Mailgun
2. **Email Templates**: Predefined templates for common scenarios
3. **Delivery Tracking**: Track sent, delivered, opened, clicked, bounced
4. **Email Verification**: Secure token-based verification
5. **Password Reset**: Secure password reset flow

### Email Templates

#### 1. Verification Email
```go
templateVars := map[string]string{
    "verification_url": appURL + "/verify-email?token=" + token,
    "user_name": firstName,
}
service.SendVerificationEmail(ctx, email, firstName, appURL+"/verify-email?token="+token)
```

#### 2. Password Reset
```go
service.SendPasswordResetEmail(ctx, email, appURL)
```

#### 3. Welcome Email
```go
service.SendWelcomeEmail(ctx, email, firstName)
```

#### 4. 2FA Notifications
```go
service.Send2FAEnabledEmail(ctx, email, firstName)
service.Send2FADisabledEmail(ctx, email, firstName)
```

#### 5. Security Alerts
```go
service.SendPasswordChangedEmail(ctx, email, firstName)
service.SendLoginNotification(ctx, email, firstName, ipAddress, userAgent)
```

### Configuration

```go
config := services.EmailServiceConfig{
    Provider:            "sendgrid", // or "mailgun"
    FromEmail:           "noreply@example.com",
    FromName:            "Your App Name",
    SendGridAPIKey:      os.Getenv("SENDGRID_API_KEY"),
    MailgunDomain:       os.Getenv("MAILGUN_DOMAIN"),
    MailgunAPIKey:       os.Getenv("MAILGUN_API_KEY"),
    AppURL:              "https://yourapp.com",
}

emailService := services.NewEmailService(emailRepo, userRepo, config)
```

### Usage Examples

#### Send Verification Email
```go
err := emailService.SendVerificationEmail(
    ctx,
    "user@example.com",
    "John",
    "https://app.com/verify?token=abc123",
)
```

#### Verify Email
```go
err := emailService.VerifyEmail(ctx, "verification-token")
```

## SMS Services

### Features

1. **Twilio Integration**: Reliable SMS delivery
2. **OTP Generation**: Secure 6-digit OTP codes
3. **SMS Templates**: Predefined message templates
4. **Rate Limiting**: Prevent SMS spam
5. **Delivery Tracking**: Monitor SMS status

### SMS Templates

#### 1. OTP
```
Your verification code is: {{code}}. It expires in 10 minutes.
```

#### 2. Verification
```
Your verification code for {{app_name}} is: {{code}}
```

#### 3. Password Reset
```
Reset your password using code: {{code}}. Valid for 10 minutes.
```

#### 4. Login Alert
```
New login detected at {{timestamp}}. If this wasn't you, secure your account immediately.
```

### Configuration

```go
config := services.SMSServiceConfig{
    Provider:           "twilio",
    TwilioAccountSID:   os.Getenv("TWILIO_ACCOUNT_SID"),
    TwilioAuthToken:    os.Getenv("TWILIO_AUTH_TOKEN"),
    TwilioFromNumber:   os.Getenv("TWILIO_FROM_NUMBER"),
}

smsService := services.NewSMSService(smsRepo, userRepo, config)
```

### Usage Examples

#### Send OTP
```go
otpCode, err := smsService.SendOTP(ctx, userID, "+1234567890")
// OTP is also stored as bcrypt hash in database
```

#### Verify OTP
```go
verified, err := smsService.VerifyOTP(ctx, "+1234567890", "123456")
```

#### Send Verification Code
```go
err := smsService.SendVerificationCode(ctx, userID, "+1234567890", "123456")
```

## Social Login

### Supported Providers

1. **Google** - OAuth2 with OpenID Connect
2. **GitHub** - OAuth2 with user profile
3. **LinkedIn** - OAuth2 with basic profile
4. **Facebook** - (Ready to implement)
5. **Twitter** - (Ready to implement)

### OAuth2 Flow

```
1. Client → GET /auth/social/{provider}?redirect_uri={uri}
2. Server → Generate state, return auth URL
3. Client → Redirect to provider OAuth URL
4. User → Authorizes application
5. Provider → Callback to /auth/social/{provider}/callback?code={code}&state={state}
6. Server → Exchange code for token
7. Server → Fetch user info from provider
8. Server → Create/link account
9. Server → Return JWT tokens
```

### Configuration

```go
config := services.SocialServiceConfig{
    AppURL:                  "https://yourapp.com",
    GoogleClientID:          os.Getenv("GOOGLE_CLIENT_ID"),
    GoogleClientSecret:      os.Getenv("GOOGLE_CLIENT_SECRET"),
    GitHubClientID:          os.Getenv("GITHUB_CLIENT_ID"),
    GitHubClientSecret:      os.Getenv("GITHUB_CLIENT_SECRET"),
    LinkedInClientID:        os.Getenv("LINKEDIN_CLIENT_ID"),
    LinkedInClientSecret:    os.Getenv("LINKEDIN_CLIENT_SECRET"),
}

socialService := services.NewSocialService(socialRepo, userRepo, config)
```

### Usage Examples

#### Get Authorization URL
```go
authURL, state, err := socialService.GetAuthURL(
    ctx,
    models.SocialProviderGoogle,
    "https://app.com/callback",
)
// Redirect user to authURL
```

#### Handle Callback
```go
userInfo, err := socialService.HandleCallback(
    ctx,
    models.SocialProviderGoogle,
    "authorization-code",
    "state-token",
)

user, isNewUser, err := socialService.GetOrCreateUserFromSocial(ctx, userInfo)
```

#### Link Account to Existing User
```go
err := socialService.LinkAccount(ctx, userID, userInfo)
```

#### Unlink Account
```go
err := socialService.UnlinkAccount(ctx, userID, models.SocialProviderGoogle)
```

#### Get Linked Accounts
```go
accounts, err := socialService.GetLinkedAccounts(ctx, userID)
```

## API Endpoints

### Email Endpoints

#### Verify Email
```http
POST /api/v1/email/verify
Content-Type: application/json

{
  "token": "verification-token"
}

Response:
{
  "message": "Email verified successfully"
}
```

#### Request Password Reset
```http
POST /api/v1/email/password-reset
Content-Type: application/json

{
  "email": "user@example.com"
}

Response:
{
  "message": "If an account exists with this email, a password reset link has been sent"
}
```

#### Reset Password
```http
POST /api/v1/email/password-reset/confirm
Content-Type: application/json

{
  "token": "reset-token",
  "new_password": "newSecurePass123!"
}

Response:
{
  "message": "Password reset successfully"
}
```

### SMS Endpoints

#### Send OTP
```http
POST /api/v1/sms/otp/send
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "phone_number": "+1234567890",
  "purpose": "verification"
}

Response:
{
  "message": "OTP sent successfully",
  "expires_at": "2024-01-15T10:35:00Z"
}
```

#### Verify OTP
```http
POST /api/v1/sms/otp/verify
Content-Type: application/json

{
  "phone_number": "+1234567890",
  "otp": "123456"
}

Response:
{
  "message": "OTP verified successfully"
}
```

### Social Login Endpoints

#### Get OAuth URL
```http
GET /api/v1/auth/social/{provider}?redirect_uri=https://app.com/callback

Response:
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?...",
  "state": "random-state-token"
}
```

#### OAuth Callback
```http
GET /api/v1/auth/social/{provider}/callback?code={code}&state={state}

Response:
{
  "message": "Social login successful",
  "is_new_user": false,
  "user": { ... },
  "access_token": "jwt-token",
  "refresh_token": "refresh-token"
}
```

#### Link Social Account
```http
POST /api/v1/user/social/link
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "provider": "google"
}

Response:
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?...",
  "state": "random-state-token"
}
```

#### Unlink Social Account
```http
DELETE /api/v1/user/social/{provider}
Authorization: Bearer {access_token}

Response:
{
  "message": "Social account unlinked successfully"
}
```

#### Get Linked Accounts
```http
GET /api/v1/user/social/accounts
Authorization: Bearer {access_token}

Response:
{
  "accounts": [
    {
      "id": "uuid",
      "provider": "google",
      "email": "user@gmail.com",
      "name": "John Doe",
      "avatar": "https://...",
      "linked_at": "2024-01-15T10:00:00Z",
      "last_used_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

## Configuration

### Environment Variables

```bash
# Email Configuration
EMAIL_PROVIDER=sendgrid              # or "mailgun"
EMAIL_FROM=noreply@example.com
EMAIL_FROM_NAME="Your App Name"
SENDGRID_API_KEY=your-sendgrid-key
MAILGUN_DOMAIN=yourdomain.com
MAILGUN_API_KEY=your-mailgun-key

# SMS Configuration
SMS_PROVIDER=twilio
TWILIO_ACCOUNT_SID=your-account-sid
TWILIO_AUTH_TOKEN=your-auth-token
TWILIO_FROM_NUMBER=+1234567890

# Social Login - Google
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-secret

# Social Login - GitHub
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-secret

# Social Login - LinkedIn
LINKEDIN_CLIENT_ID=your-linkedin-client-id
LINKEDIN_CLIENT_SECRET=your-linkedin-secret

# Application URL
APP_URL=https://yourapp.com
```

### Provider Setup Guides

#### SendGrid Setup
1. Create account at sendgrid.com
2. Verify sender email
3. Create API key with "Mail Send" permission
4. Add to environment variables

#### Mailgun Setup
1. Create account at mailgun.com
2. Verify domain
3. Get API key from dashboard
4. Add to environment variables

#### Twilio Setup
1. Create account at twilio.com
2. Purchase phone number
3. Get Account SID and Auth Token
4. Add to environment variables

#### Google OAuth Setup
1. Go to Google Cloud Console
2. Create project and enable Google+ API
3. Create OAuth2 credentials
4. Add authorized redirect URI: `{APP_URL}/auth/google/callback`
5. Add to environment variables

#### GitHub OAuth Setup
1. Go to GitHub Settings → Developer settings
2. Create OAuth App
3. Add callback URL: `{APP_URL}/auth/github/callback`
4. Get Client ID and Secret
5. Add to environment variables

#### LinkedIn OAuth Setup
1. Go to LinkedIn Developers
2. Create application
3. Add redirect URL: `{APP_URL}/auth/linkedin/callback`
4. Get Client ID and Secret
5. Add to environment variables

## Testing

### Email Testing

```go
// Test email verification
token := "test-verification-token"
err := emailService.VerifyEmail(ctx, token)
assert.NoError(t, err)

// Test password reset
err = emailService.SendPasswordResetEmail(ctx, "user@example.com", "http://localhost:8080")
assert.NoError(t, err)
```

### SMS Testing

```go
// Test OTP generation and verification
otpCode, err := smsService.SendOTP(ctx, userID, "+1234567890")
assert.NoError(t, err)
assert.Len(t, otpCode, 6)

verified, err := smsService.VerifyOTP(ctx, "+1234567890", otpCode)
assert.NoError(t, err)
assert.True(t, verified)
```

### Social Login Testing

```go
// Test OAuth URL generation
authURL, state, err := socialService.GetAuthURL(
    ctx,
    models.SocialProviderGoogle,
    "http://localhost:8080/callback",
)
assert.NoError(t, err)
assert.Contains(t, authURL, "accounts.google.com")
assert.NotEmpty(t, state)
```

### Manual Testing with Postman

1. **Email Verification**
   - Send verification email
   - Check email for token
   - Call verify endpoint with token

2. **SMS OTP**
   - Request OTP
   - Check phone for code
   - Verify OTP

3. **Social Login**
   - Get OAuth URL
   - Open URL in browser
   - Authorize application
   - Capture callback code
   - Exchange for tokens

## Security Considerations

1. **Token Expiration**: All tokens have expiration times
2. **Rate Limiting**: Prevent abuse of email/SMS endpoints
3. **HTTPS Only**: All OAuth callbacks must use HTTPS
4. **State Validation**: Prevent CSRF attacks in OAuth flow
5. **Secure Storage**: Access tokens encrypted at rest
6. **OTP Hashing**: OTP codes stored as bcrypt hashes
7. **Attempt Limiting**: Maximum 3 OTP verification attempts

## Performance Optimization

1. **Async Email Sending**: Queue emails for background processing
2. **Template Caching**: Cache rendered email templates
3. **Connection Pooling**: Reuse HTTP connections to providers
4. **Batch Operations**: Send multiple SMSs in batch when supported
5. **Token Cleanup**: Periodically clean expired tokens

## Monitoring and Logging

### Metrics to Track

1. **Email Metrics**
   - Sent count
   - Delivered count
   - Open rate
   - Click rate
   - Bounce rate

2. **SMS Metrics**
   - Sent count
   - Delivered count
   - Failed count
   - Cost per message

3. **Social Login Metrics**
   - Login attempts per provider
   - Success rate
   - New user conversion rate

### Log Examples

```go
log.Info("Email sent", 
    "to", email,
    "template", templateName,
    "provider", provider,
    "message_id", messageID,
)

log.Info("SMS delivered",
    "to", phone,
    "template", templateName,
    "segments", segments,
    "cost", cost,
)

log.Info("Social login successful",
    "provider", provider,
    "user_id", userID,
    "is_new_user", isNewUser,
)
```

## Troubleshooting

### Common Issues

1. **Email not received**
   - Check spam folder
   - Verify sender email is authenticated
   - Check email logs for delivery status

2. **SMS not received**
   - Verify phone number format
   - Check SMS logs for error messages
   - Verify Twilio account balance

3. **OAuth callback fails**
   - Verify redirect URI matches exactly
   - Check state token validity
   - Ensure HTTPS in production

4. **Token expired errors**
   - Increase token expiration time
   - Implement token refresh logic
   - Clear expired tokens from database

## Next Steps (Phase 4)

- User management CRUD API
- Company management API
- Audit log search and filtering
- Advanced reporting dashboard

## Additional Resources

- [SendGrid Documentation](https://docs.sendgrid.com/)
- [Mailgun Documentation](https://documentation.mailgun.com/)
- [Twilio Documentation](https://www.twilio.com/docs)
- [Google OAuth2 Guide](https://developers.google.com/identity/protocols/oauth2)
- [GitHub OAuth Guide](https://docs.github.com/en/developers/apps/building-oauth-apps)
- [LinkedIn OAuth Guide](https://docs.microsoft.com/en-us/linkedin/shared/authentication/authentication)
