# Phase 2 Implementation Guide - Enhanced Authentication

## Overview

Phase 2 adds advanced authentication capabilities to the SSO service:
- **Two-Factor Authentication (2FA)** with TOTP (Time-based One-Time Password)
- **OAuth2 Authorization Code Flow** for third-party application integration
- **Backup codes** for 2FA recovery
- **QR code generation** for authenticator apps

---

## Files Created

### Models
- `models/two_factor.go` - 2FA data models and DTOs
- `models/oauth2.go` - OAuth2 data models and DTOs

### Repository Layer
- `repository/two_factor_repository.go` - 2FA and OAuth2 database operations

### Service Layer
- `services/two_factor_service.go` - 2FA business logic (TOTP, backup codes)
- `services/oauth2_service.go` - OAuth2 flows (authorization, token exchange)

### Handler Layer
- `handlers/two_factor_handler.go` - 2FA HTTP endpoints
- `handlers/oauth2_handler.go` - OAuth2 HTTP endpoints

### Database
- `database/migrations/004_enhanced_authentication.sql` - Database schema
- `database/migrations/004_rollback.sql` - Rollback script

---

## Database Schema

### Two-Factor Authentication Tables

#### `user_two_factor`
Stores 2FA configuration for each user.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | Foreign key to users table |
| method | VARCHAR(10) | 'totp' or 'sms' |
| secret | TEXT | TOTP secret (encrypted) |
| phone_number | VARCHAR(20) | Phone for SMS 2FA |
| status | VARCHAR(20) | 'disabled', 'pending', or 'enabled' |
| backup_codes_count | INTEGER | Number of unused backup codes |
| verified_at | TIMESTAMP | When 2FA was verified |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |

#### `backup_codes`
Stores hashed backup codes for 2FA recovery.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | Foreign key to users table |
| code | TEXT | Hashed backup code (bcrypt) |
| used_at | TIMESTAMP | When code was used (NULL if unused) |
| created_at | TIMESTAMP | Creation timestamp |

### OAuth2 Tables

#### `oauth2_clients`
OAuth2 client applications.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| client_id | VARCHAR(255) | Unique client identifier |
| client_secret | TEXT | Hashed client secret |
| name | VARCHAR(255) | Client application name |
| description | TEXT | Client description |
| redirect_uris | TEXT[] | Allowed redirect URIs |
| grant_types | TEXT[] | Allowed grant types |
| scopes | TEXT[] | Allowed scopes |
| owner_id | UUID | User who owns this client |
| logo_url | TEXT | Client logo URL |
| active | BOOLEAN | Whether client is active |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |

#### `oauth2_authorization_codes`
Short-lived authorization codes.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| code | VARCHAR(255) | Authorization code |
| client_id | VARCHAR(255) | Client that requested code |
| user_id | UUID | User who authorized |
| redirect_uri | TEXT | Redirect URI for this flow |
| scopes | TEXT[] | Granted scopes |
| expires_at | TIMESTAMP | Expiration time (10 minutes) |
| used_at | TIMESTAMP | When code was exchanged |
| created_at | TIMESTAMP | Creation timestamp |

#### `oauth2_tokens`
Access and refresh tokens.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| access_token | TEXT | JWT access token |
| refresh_token | TEXT | Refresh token |
| client_id | VARCHAR(255) | Client this token belongs to |
| user_id | UUID | User this token represents |
| scopes | TEXT[] | Granted scopes |
| expires_at | TIMESTAMP | Token expiration |
| revoked_at | TIMESTAMP | When token was revoked |
| created_at | TIMESTAMP | Creation timestamp |

---

## API Endpoints

### Two-Factor Authentication Endpoints

#### Setup 2FA
```
POST /api/v1/auth/2fa/setup
Authorization: Bearer {access_token}
```

**Response:**
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qrCodeUrl": "otpauth://totp/SSO%20Service:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=SSO%20Service",
  "backupCodes": [
    "ABCD1234",
    "EFGH5678",
    "IJKL9012",
    "MNOP3456",
    "QRST7890",
    "UVWX1234",
    "YZAB5678",
    "CDEF9012"
  ]
}
```

#### Enable 2FA
```
POST /api/v1/auth/2fa/enable
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "method": "totp",
  "code": "123456"
}
```

#### Disable 2FA
```
POST /api/v1/auth/2fa/disable
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "code": "123456"
}
```

#### Get 2FA Status
```
GET /api/v1/auth/2fa/status
Authorization: Bearer {access_token}
```

**Response:**
```json
{
  "enabled": true,
  "method": "totp",
  "backupCodesCount": 5,
  "verifiedAt": "2025-10-25T10:30:00Z"
}
```

#### Verify TOTP Code
```
POST /api/v1/auth/2fa/verify
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "code": "123456"
}
```

#### Regenerate Backup Codes
```
POST /api/v1/auth/2fa/backup-codes/regenerate
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "code": "123456"
}
```

**Response:**
```json
{
  "backupCodes": ["NEWCODE1", "NEWCODE2", ...],
  "message": "Backup codes regenerated successfully"
}
```

#### Get QR Code Image
```
GET /api/v1/auth/2fa/qr
Authorization: Bearer {access_token}
```

Returns PNG image of QR code.

---

### OAuth2 Endpoints

#### Create OAuth2 Client
```
POST /api/v1/oauth2/clients
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "My Application",
  "description": "A sample OAuth2 client",
  "redirectUris": ["https://myapp.com/callback"],
  "grantTypes": ["authorization_code", "refresh_token"],
  "scopes": ["openid", "profile", "email"],
  "logoUrl": "https://myapp.com/logo.png"
}
```

**Response:**
```json
{
  "client": {
    "id": "uuid",
    "clientId": "client_abc123",
    "name": "My Application",
    "redirectUris": ["https://myapp.com/callback"],
    ...
  },
  "clientSecret": "secret_xyz789"
}
```

**⚠️ Important:** The `clientSecret` is only returned once during creation!

#### List OAuth2 Clients
```
GET /api/v1/oauth2/clients
Authorization: Bearer {access_token}
```

#### OAuth2 Authorization
```
GET /api/v1/oauth2/authorize?response_type=code&client_id=CLIENT_ID&redirect_uri=REDIRECT_URI&scope=openid profile&state=STATE
```

Redirects to: `REDIRECT_URI?code=AUTH_CODE&state=STATE`

#### Exchange Code for Token
```
POST /api/v1/oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=AUTH_CODE&
redirect_uri=REDIRECT_URI&
client_id=CLIENT_ID&
client_secret=CLIENT_SECRET
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_xyz...",
  "scope": "openid profile email"
}
```

#### Refresh Access Token
```
POST /api/v1/oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&
refresh_token=REFRESH_TOKEN&
client_id=CLIENT_ID&
client_secret=CLIENT_SECRET
```

#### Token Introspection
```
POST /api/v1/oauth2/introspect
Content-Type: application/x-www-form-urlencoded

token=ACCESS_TOKEN
```

**Response:**
```json
{
  "active": true,
  "client_id": "client_abc123",
  "user_id": "user_uuid",
  "scopes": ["openid", "profile"],
  "exp": 1730000000
}
```

#### Revoke Token
```
POST /api/v1/oauth2/revoke
Content-Type: application/x-www-form-urlencoded

token=ACCESS_TOKEN
```

#### Get User Info (OpenID Connect)
```
GET /api/v1/oauth2/userinfo
Authorization: Bearer {access_token}
```

**Response:**
```json
{
  "sub": "user_uuid",
  "name": "John Doe",
  "email": "john@example.com"
}
```

---

## Integration Guide

### 1. Apply Database Migration

```bash
cd /Users/leapfrog/prayas_personal/union-products/sso
psql -d sso_db -f database/migrations/004_enhanced_authentication.sql
```

### 2. Update main.go

Add the new services and handlers to your application:

```go
package main

import (
    "database/sql"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
    
    "sso/handlers"
    "sso/middleware"
    "sso/repository"
    "sso/services"
)

func main() {
    // Database connection
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Initialize repositories
    userRepo := repository.NewUserRepository(db)
    securityRepo := repository.NewSecurityRepository(db)
    twoFactorRepo := repository.NewTwoFactorRepository(db)
    oauth2Repo := repository.NewOAuth2Repository(db)

    // Initialize services
    jwtSecret := os.Getenv("JWT_SECRET")
    authService := services.NewAuthService(userRepo, jwtSecret)
    securityService := services.NewSecurityService(securityRepo, userRepo)
    twoFactorService := services.NewTwoFactorService(twoFactorRepo, userRepo)
    oauth2Service := services.NewOAuth2Service(oauth2Repo, userRepo, jwtSecret)

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(authService)
    securityHandler := handlers.NewSecurityHandler(securityService)
    twoFactorHandler := handlers.NewTwoFactorHandler(twoFactorService)
    oauth2Handler := handlers.NewOAuth2Handler(oauth2Service)

    // Setup router
    router := gin.Default()

    // Public routes
    public := router.Group("/api/v1/auth")
    {
        public.POST("/register", authHandler.Register)
        public.POST("/login", authHandler.Login)
        public.POST("/refresh", authHandler.RefreshToken)
    }

    // Protected routes
    protected := router.Group("/api/v1/auth")
    protected.Use(middleware.AuthMiddleware(authService))
    {
        protected.GET("/profile", authHandler.GetProfile)
        protected.POST("/logout", authHandler.Logout)
        
        // 2FA routes
        protected.POST("/2fa/setup", twoFactorHandler.SetupTOTP)
        protected.POST("/2fa/enable", twoFactorHandler.EnableTwoFactor)
        protected.POST("/2fa/disable", twoFactorHandler.DisableTwoFactor)
        protected.GET("/2fa/status", twoFactorHandler.GetTwoFactorStatus)
        protected.POST("/2fa/verify", twoFactorHandler.VerifyTOTP)
        protected.POST("/2fa/backup-codes/regenerate", twoFactorHandler.RegenerateBackupCodes)
        protected.GET("/2fa/qr", twoFactorHandler.GetQRCode)
    }

    // OAuth2 routes
    oauth2Group := router.Group("/api/v1/oauth2")
    {
        // Public OAuth2 endpoints
        oauth2Group.GET("/authorize", oauth2Handler.Authorize)
        oauth2Group.POST("/token", oauth2Handler.Token)
        oauth2Group.POST("/introspect", oauth2Handler.Introspect)
        oauth2Group.POST("/revoke", oauth2Handler.Revoke)
        oauth2Group.GET("/userinfo", oauth2Handler.GetUserInfo)
        
        // Protected OAuth2 client management
        oauth2Protected := oauth2Group.Group("")
        oauth2Protected.Use(middleware.AuthMiddleware(authService))
        {
            oauth2Protected.POST("/clients", oauth2Handler.CreateClient)
            oauth2Protected.GET("/clients", oauth2Handler.ListClients)
        }
    }

    // Admin routes
    admin := router.Group("/api/v1/admin")
    admin.Use(middleware.AuthMiddleware(authService))
    admin.Use(middleware.RequireRole(securityService, "admin"))
    {
        admin.POST("/unlock-account", securityHandler.UnlockAccount)
        admin.POST("/assign-role", securityHandler.AssignRole)
        admin.POST("/remove-role", securityHandler.RemoveRole)
    }

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatal(err)
    }
}
```

### 3. Update Login Flow for 2FA

Modify your login handler to check for 2FA:

```go
func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate credentials
    user, err := h.authService.ValidateCredentials(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Check if 2FA is enabled
    twoFactor, _ := h.twoFactorService.GetTwoFactorStatus(c.Request.Context(), user.ID)
    if twoFactor != nil && twoFactor.Status == models.TwoFactorStatusEnabled {
        // 2FA is enabled - require verification
        if req.TOTPCode == "" {
            c.JSON(http.StatusOK, gin.H{
                "requiresTwoFactor": true,
                "userId": user.ID,
            })
            return
        }

        // Verify TOTP code
        valid, err := h.twoFactorService.VerifyTOTP(c.Request.Context(), user.ID, req.TOTPCode)
        if err != nil || !valid {
            // Try backup code
            valid, err = h.twoFactorService.VerifyBackupCode(c.Request.Context(), user.ID, req.TOTPCode)
            if err != nil || !valid {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid 2FA code"})
                return
            }
        }
    }

    // Generate tokens
    tokens, err := h.authService.GenerateTokens(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, tokens)
}
```

---

## Testing Guide

### Testing 2FA

#### 1. Setup 2FA
```bash
# Login first
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' | jq -r '.accessToken')

# Setup 2FA
curl -X POST http://localhost:8080/api/v1/auth/2fa/setup \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 2. Scan QR Code
Use Google Authenticator, Authy, or similar app to scan the QR code URL.

#### 3. Enable 2FA
```bash
# Get code from authenticator app
CODE="123456"

curl -X POST http://localhost:8080/api/v1/auth/2fa/enable \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"method\":\"totp\",\"code\":\"$CODE\"}"
```

#### 4. Test Login with 2FA
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password","totpCode":"123456"}'
```

### Testing OAuth2

#### 1. Create OAuth2 Client
```bash
curl -X POST http://localhost:8080/api/v1/oauth2/clients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test App",
    "redirectUris": ["http://localhost:3000/callback"],
    "grantTypes": ["authorization_code", "refresh_token"],
    "scopes": ["openid", "profile", "email"]
  }' | jq

# Save CLIENT_ID and CLIENT_SECRET from response
```

#### 2. Authorization Flow
```bash
# Open in browser:
http://localhost:8080/api/v1/oauth2/authorize?response_type=code&client_id=CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=openid+profile&state=xyz

# You'll be redirected to:
http://localhost:3000/callback?code=AUTH_CODE&state=xyz
```

#### 3. Exchange Code for Token
```bash
curl -X POST http://localhost:8080/api/v1/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=AUTH_CODE&redirect_uri=http://localhost:3000/callback&client_id=CLIENT_ID&client_secret=CLIENT_SECRET" | jq
```

#### 4. Use Access Token
```bash
ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIs..."

curl -X GET http://localhost:8080/api/v1/oauth2/userinfo \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
```

---

## Security Considerations

### 2FA Security
1. **TOTP Secrets:** Store encrypted at rest
2. **Backup Codes:** Always hashed with bcrypt
3. **Rate Limiting:** Apply to 2FA endpoints
4. **Backup Code Usage:** Track and alert on usage
5. **Recovery Process:** Require admin intervention if all codes lost

### OAuth2 Security
1. **Client Secrets:** Always hashed with bcrypt
2. **Authorization Codes:** Short-lived (10 minutes)
3. **PKCE:** Consider implementing for public clients
4. **State Parameter:** Always validate to prevent CSRF
5. **Redirect URIs:** Strictly validate against whitelist
6. **Token Rotation:** Implement refresh token rotation
7. **Scope Validation:** Strictly enforce scope permissions

---

## Dependencies

### New Go Packages Added
```bash
go get github.com/pquerna/otp@latest
```

This adds:
- `github.com/pquerna/otp` - TOTP implementation
- `github.com/boombuler/barcode` - QR code generation (dependency)

---

## Performance Optimization

### Database Cleanup
Run periodic cleanup to remove expired data:

```sql
-- Run daily
SELECT cleanup_expired_oauth2_codes();
SELECT cleanup_expired_oauth2_tokens();
SELECT cleanup_old_backup_codes();
```

Add to cron or scheduled job:
```go
func setupCleanupJobs(db *sql.DB) {
    ticker := time.NewTicker(24 * time.Hour)
    go func() {
        for range ticker.C {
            db.Exec("SELECT cleanup_expired_oauth2_codes()")
            db.Exec("SELECT cleanup_expired_oauth2_tokens()")
            db.Exec("SELECT cleanup_old_backup_codes()")
        }
    }()
}
```

---

## Next Steps

✅ **Phase 2 Complete!**

Ready for **Phase 3: External Services**
- Email integration (SendGrid/Mailgun)
- SMS integration (Twilio)
- Social login (Google, GitHub, LinkedIn)

---

*Last Updated: October 25, 2025*  
*Version: 2.0.0*  
*Status: Phase 2 Complete ✅*
