# SSO Service Testing Guide

## Server Status

The SSO service is successfully running! ✅

- **Status**: Running
- **Port**: 8080
- **Database**: PostgreSQL (connected)
- **Environment**: Development
- **Health Check**: http://localhost:8080/health

## API Endpoints Tested

### 1. Health Check
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "service": "sso",
  "status": "healthy",
  "version": "1.0.0"
}
```

### 2. User Registration
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

**Response:**
```json
{
  "message": "Registration successful. Please verify your email.",
  "user": {
    "id": "f7fe17cc-8b91-4cc9-9765-bab04d0b0532",
    "email": "test@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "isActive": true,
    "isVerified": false,
    "createdAt": "2025-10-20T01:53:10.993241+05:45",
    "updatedAt": "2025-10-20T01:53:10.993241+05:45"
  }
}
```

### 3. User Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "clientId": "host-app"
  }'
```

**Response:**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "5ccddzzPm10U8E0JQdXCOaall2OMyqR9mcSTvV7Pgl4=",
  "expiresIn": 900,
  "tokenType": "Bearer",
  "user": {
    "id": "f7fe17cc-8b91-4cc9-9765-bab04d0b0532",
    "email": "test@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "isActive": true,
    "isVerified": false,
    "createdAt": "2025-10-20T01:53:10.993241Z",
    "updatedAt": "2025-10-20T01:53:10.993241Z",
    "lastLogin": "2025-10-20T01:53:35.979089+05:45"
  }
}
```

### 4. Get Current User
```bash
TOKEN="<your_access_token>"
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "email": "test@example.com",
  "userId": "f7fe17cc-8b91-4cc9-9765-bab04d0b0532"
}
```

### 5. Validate Token
```bash
TOKEN="<your_access_token>"
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "claims": {
    "user_id": "f7fe17cc-8b91-4cc9-9765-bab04d0b0532",
    "email": "test@example.com",
    "companies": null,
    "iss": "union-products-sso",
    "exp": 1760905415,
    "iat": 1760904515
  },
  "valid": true
}
```

## Available OAuth Clients

The following OAuth clients are pre-configured for your micro-frontends:

1. **host-app** - Main host application
2. **crm-module** - CRM module
3. **inventory-module** - Inventory module  
4. **hr-module** - HR module
5. **finance-module** - Finance module
6. **task-module** - Task module

All clients have redirect URIs configured for `http://localhost:3000/callback` through `http://localhost:3005/callback`.

## Server Management

### Start Server
```bash
# Foreground
./bin/sso-server

# Background (with logs)
nohup ./bin/sso-server > sso.log 2>&1 &
```

### Stop Server
```bash
# Find process
ps aux | grep sso-server

# Kill process
kill <PID>

# Or if running with nohup
pkill -f sso-server
```

### Check Logs
```bash
tail -f sso.log
```

### Rebuild Server
```bash
go build -o bin/sso-server cmd/server/main.go
```

## Integration with Micro-Frontends

To integrate with your React micro-frontends, use the TypeScript SDK located in `sdk/typescript/`:

```typescript
import { SSOClient, SSOProvider } from 'sso-sdk';

// Initialize the client
const ssoClient = new SSOClient({
  ssoUrl: 'http://localhost:8080',
  clientId: 'host-app', // or your module's client ID
  redirectUri: 'http://localhost:3000/callback',
});

// Wrap your app with the provider
function App() {
  return (
    <SSOProvider client={ssoClient}>
      <YourApp />
    </SSOProvider>
  );
}

// Use the hook in your components
function YourComponent() {
  const { user, isAuthenticated, login, logout } = useSSO();
  
  // Your component logic
}
```

## Next Steps

1. ✅ Server is running successfully
2. ✅ All core authentication endpoints working
3. ✅ Database schema applied
4. ✅ OAuth clients configured
5. ⏭️ Install and integrate the TypeScript SDK in your micro-frontends
6. ⏭️ Test complete authentication flow with frontend
7. ⏭️ Configure production environment variables
8. ⏭️ Set up SSL/TLS for production
9. ⏭️ Configure email service for verification emails

## Troubleshooting

### Issue Resolved: Password Authentication Failed

**Problem**: The server couldn't connect to PostgreSQL due to password authentication failure.

**Root Cause**: The `.env` file wasn't being loaded by the Go application.

**Solution**: Added `github.com/joho/godotenv` package to load environment variables from the `.env` file.

```go
import "github.com/joho/godotenv"

func Load() *Config {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found, using environment variables")
    }
    // ... rest of config loading
}
```

### PostgreSQL Connection

The server uses the following connection details:
- Host: 127.0.0.1
- Port: 5432
- User: postgres
- Database: sso_db
- SSL Mode: disable (development only)

Make sure the PostgreSQL Docker container is running:
```bash
docker ps | grep postgres-container
```

If not running:
```bash
docker start postgres-container
```

## Success Metrics

✅ All systems operational:
- Database connection: **WORKING**
- User registration: **WORKING**
- User login: **WORKING**
- Token generation: **WORKING**
- Token validation: **WORKING**
- Protected endpoints: **WORKING**
- Health check: **WORKING**

The SSO service is production-ready for development/staging environments! 🎉
