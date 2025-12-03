# Union Products SSO Service

A complete Single Sign-On (SSO) authentication service built with Go and PostgreSQL, designed to work seamlessly with micro-frontend architectures.

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Frontend Integration](#frontend-integration)
- [TypeScript SDK](#typescript-sdk)
- [Database Schema](#database-schema)
- [Testing Guide](#testing-guide)
- [Security Best Practices](#security-best-practices)
- [Development](#development)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

## Features

- 🔐 **Secure Authentication**: JWT-based authentication with access and refresh tokens
- 🔄 **Token Rotation**: Automatic refresh token rotation for enhanced security
- 👥 **Multi-Company Support**: Users can belong to multiple companies/organizations
- 🌐 **CORS Support**: Pre-configured for multiple micro-frontend applications
- 📊 **Session Management**: Track active user sessions across devices
- 🔍 **Audit Logging**: Comprehensive audit trail for security compliance
- 🔑 **Password Management**: Secure password reset and change functionality
- ✉️ **Email Verification**: Built-in email verification system
- 🐳 **Docker Ready**: Complete Docker and Docker Compose setup
- 📦 **TypeScript SDK**: Ready-to-use client library with React hooks

## Architecture

This SSO service is designed to work with the following micro-frontend modules:
- Host Application (port 3000)
- CRM Module (port 3001)
- Inventory Module (port 3002)
- HR Module (port 3003)
- Finance Module (port 3004)
- Task Module (port 3005)

```
┌─────────────────────────────────────────────────────────────┐
│                     Micro-Frontends                         │
│  ┌────────┐  ┌──────┐  ┌──────────┐  ┌────┐  ┌────────┐  │
│  │  Host  │  │ CRM  │  │ Inventory │  │ HR │  │Finance │  │
│  └────┬───┘  └───┬──┘  └─────┬────┘  └──┬─┘  └───┬────┘  │
│       │          │            │           │        │        │
│       └──────────┴────────────┴───────────┴────────┘        │
│                          │                                   │
│                     SSO SDK                                  │
└──────────────────────────┼──────────────────────────────────┘
                           │
                           ▼
         ┌─────────────────────────────────────┐
         │      SSO Service (Port 8080)        │
         │  ┌──────────────────────────────┐  │
         │  │    Authentication API         │  │
         │  │  - Register, Login, Logout    │  │
         │  │  - Token Management           │  │
         │  │  - Session Management         │  │
         │  └──────────────────────────────┘  │
         │  ┌──────────────────────────────┐  │
         │  │   Authorization & Validation  │  │
         │  │  - JWT Generation             │  │
         │  │  - Token Validation           │  │
         │  │  - User Authorization         │  │
         │  └──────────────────────────────┘  │
         └─────────────────┬───────────────────┘
                           │
                           ▼
         ┌─────────────────────────────────────┐
         │   PostgreSQL Database (Port 5432)    │
         │  - Users & Companies                 │
         │  - Sessions & Tokens                 │
         │  - OAuth Clients                     │
         │  - Audit Logs                        │
         └──────────────────────────────────────┘
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Docker & Docker Compose (optional)
- Node.js 18+ (for TypeScript SDK)

---

## Quick Start

### 🚀 Get Started in 3 Minutes

**Step 1: Choose Your Setup Method**

#### Option A: Docker (Easiest)

#### Option A: Docker (Easiest)
```bash
docker-compose up -d
```

**Option B: Local**
```bash
./setup.sh
```

**Option C: Manual**
```bash
# Create and migrate database
createdb sso_db
psql -d sso_db -f database/migrations/001_initial_schema.sql

# Run server
./bin/sso-server
```

**Step 2: Test It**

```bash
# Health check
curl http://localhost:8080/health

# Should return: {"status":"healthy","service":"sso","version":"1.0.0"}
```

**Step 3: Create Your First User**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "firstName": "Admin",
    "lastName": "User"
  }'
```

**Step 4: Login**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }'
```

**Step 5: Use in Your Frontend** - See [Frontend Integration](#frontend-integration) section below.

---

### Using Docker Compose (Recommended)

1. Clone the repository and navigate to the SSO directory:
```bash
cd sso
```

2. Start the services:
```bash
docker-compose up -d
```

3. The SSO service will be available at `http://localhost:8080`

4. View logs:
```bash
docker-compose logs -f sso-server
```

5. Stop services:
```bash
docker-compose down
```

### Manual Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database:
```bash
createdb sso_db
```

3. Run database migrations:
```bash
make migrate-up
# or
psql -d sso_db -f database/migrations/001_initial_schema.sql
```

4. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

5. Run the server:
```bash
make run
# or
go run cmd/server/main.go
```

---

---

## Configuration

Edit the `.env` file to configure the service:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=sso_db
DB_SSLMODE=disable

# JWT Secrets (CHANGE IN PRODUCTION!)
JWT_ACCESS_SECRET=your-super-secret-access-key
JWT_REFRESH_SECRET=your-super-secret-refresh-key
JWT_ISSUER=union-products-sso

# Server
SERVER_PORT=8080
ENV=development

# CORS - Add your frontend URLs
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,...
```

### Service URLs

- **API**: http://localhost:8080
- **Health**: http://localhost:8080/health
- **Service Info**: http://localhost:8080/

### Pre-configured Modules

Your micro-frontends are pre-configured:
- `host-app` → localhost:3000
- `crm-module` → localhost:3001
- `inventory-module` → localhost:3002
- `hr-module` → localhost:3003
- `finance-module` → localhost:3004
- `task-module` → localhost:3005

---

## API Documentation

### Base URL

```
Development: http://localhost:8080
Production: https://sso.yourcompany.com
```

### Authentication

Most endpoints require authentication using JWT tokens. Include the access token in the Authorization header:

```
Authorization: Bearer <access_token>
```

### Response Format

All API responses follow this structure:

**Success Response:**
```json
{
  "data": {...},
  "message": "Success message (optional)"
}
```

**Error Response:**
```json
{
  "error": "Error message"
}
```

### HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required or failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

### Public API Endpoints

## API Endpoints

### Public Endpoints

### Public API Endpoints

#### POST /api/v1/auth/register
Register a new user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "firstName": "John",
  "lastName": "Doe"
}
```

**Validation:**
- `email`: Required, valid email format
- `password`: Required, minimum 8 characters
- `firstName`: Required
- `lastName`: Required

**Response:** `201 Created`
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "isActive": true,
    "isVerified": false
  },
  "message": "Registration successful. Please verify your email."
}
```

**Errors:**
- `400`: User already exists
- `400`: Invalid request parameters

---

#### POST /api/v1/auth/login
Login with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "clientId": "crm-module"
}
```

**Parameters:**
- `email`: Required, user's email
- `password`: Required, user's password
- `clientId`: Optional, OAuth client identifier

**Response:** `200 OK`
```json
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "random-token",
  "expiresIn": 900,
  "tokenType": "Bearer",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe"
  },
  "companies": []
}
```

**Errors:**
- `401`: Invalid credentials
- `401`: Account is inactive
- `400`: Invalid request parameters

---
Refresh access token.

**Request:**
```json
{
  "refreshToken": "your-refresh-token"
}
```

**Response:**
```json
{
  "accessToken": "new-access-token",
  "refreshToken": "new-refresh-token",
  "expiresIn": 900,
  "tokenType": "Bearer",
  "user": {...}
}
```

#### POST /api/v1/auth/logout
Logout and revoke refresh token.

**Request:**
```json
{
  "refreshToken": "your-refresh-token"
}
```

#### GET /api/v1/auth/validate
Validate access token.

**Headers:**
```
Authorization: Bearer <access-token>
```

**Response:**
```json
{
  "valid": true,
  "claims": {
    "user_id": "uuid",
    "email": "user@example.com"
  }
}
```

### Protected Endpoints (Require Authentication)

#### GET /api/v1/auth/me
Get current user information.

**Headers:**
```
Authorization: Bearer <access-token>
```

#### POST /api/v1/auth/change-password
Change user password.

**Headers:**
```
Authorization: Bearer <access-token>
```

**Request:**
```json
{
  "oldPassword": "currentpassword",
  "newPassword": "newsecurepassword"
}
```

#### POST /api/v1/auth/logout-all
Logout from all devices.

**Headers:**
```
Authorization: Bearer <access-token>
```

## Frontend Integration

### TypeScript/JavaScript SDK

Install the SDK (once published to npm):
```bash
npm install @union-products/sso-client
```

Or copy the SDK from `sdk/typescript/src/` to your project.

### React Integration

```typescript
// main.tsx or App.tsx
import { initializeSSO, SSOProvider } from './sso-sdk';

const ssoClient = initializeSSO({
  baseURL: 'http://localhost:8080',
  clientId: 'crm-module',
  redirectUri: 'http://localhost:3001/callback',
});

function App() {
  return (
    <SSOProvider client={ssoClient}>
      <YourApp />
    </SSOProvider>
  );
}
```

### Using the SSO Hook

```typescript
import { useSSO } from './sso-sdk';

function LoginPage() {
  const { login, isLoading, isAuthenticated } = useSSO();

  const handleLogin = async () => {
    try {
      await login({
        email: 'user@example.com',
        password: 'password',
      });
      // Redirect to dashboard
    } catch (error) {
      console.error('Login failed:', error);
    }
  };

  return (
    <div>
      {isAuthenticated ? (
        <p>Already logged in!</p>
      ) : (
        <button onClick={handleLogin} disabled={isLoading}>
          Login
        </button>
      )}
    </div>
  );
}
```

### Protected Routes

```typescript
import { useSSO } from './sso-sdk';
import { Navigate } from 'react-router-dom';

function ProtectedRoute({ children }) {
  const { isAuthenticated, isLoading } = useSSO();

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  return children;
}
```

### Making Authenticated API Calls

```typescript
import { getSSO } from './sso-sdk';

async function fetchData() {
  const sso = getSSO();
  const token = sso.getAccessToken();

  const response = await fetch('https://api.example.com/data', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  return response.json();
}
```

## Database Schema

The SSO service uses the following main tables:

- **users**: User accounts and profiles
- **companies**: Company/organization information
- **user_companies**: User-company relationships with roles
- **oauth_clients**: Registered OAuth2 clients (micro-frontends)
- **refresh_tokens**: Refresh tokens for token rotation
- **sessions**: Active user sessions
- **audit_logs**: Security audit trail
- **password_reset_tokens**: Password reset functionality
- **email_verification_tokens**: Email verification

## Security Best Practices

1. **Change Default Secrets**: Update `JWT_ACCESS_SECRET` and `JWT_REFRESH_SECRET` in production
2. **Use HTTPS**: Always use HTTPS in production
3. **Secure Database**: Use strong database passwords and restrict access
4. **Token Expiry**: Access tokens expire in 15 minutes, refresh tokens in 7 days
5. **CORS**: Only allow trusted origins in `ALLOWED_ORIGINS`
6. **Rate Limiting**: Implement rate limiting for login attempts (TODO)
7. **Password Requirements**: Minimum 8 characters (enforce stronger policies)

## Development

### Available Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make docker-build  # Build Docker image
make docker-up     # Start Docker containers
make docker-down   # Stop Docker containers
make migrate-up    # Run database migrations
make migrate-down  # Rollback migrations
make clean         # Clean build files
```

### Project Structure

```
sso/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── config/
│   └── config.go             # Configuration management
├── database/
│   └── migrations/           # SQL migrations
├── handlers/
│   └── auth_handler.go       # HTTP handlers
├── middleware/
│   └── middleware.go         # HTTP middleware
├── models/
│   └── models.go             # Data models
├── repository/
│   ├── user_repository.go    # User data access
│   ├── session_repository.go # Session data access
│   └── token_repository.go   # Token data access
├── services/
│   └── auth_service.go       # Business logic
├── sdk/
│   └── typescript/           # TypeScript/React SDK
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## Testing

Run tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

## Deployment

### Using Docker

1. Build the image:
```bash
docker build -t sso-service:latest .
```

2. Run the container:
```bash
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e JWT_ACCESS_SECRET=your-secret \
  --name sso-service \
  sso-service:latest
```

### Production Considerations

1. Use a reverse proxy (nginx/Caddy)
2. Set up SSL/TLS certificates
3. Configure proper logging and monitoring
4. Set up database backups
5. Use a secrets management system
6. Implement rate limiting
7. Set up health checks and alerts

## Troubleshooting

### Cannot connect to database
- Check database credentials in `.env`
- Ensure PostgreSQL is running
- Verify network connectivity

### CORS errors
- Add your frontend URL to `ALLOWED_ORIGINS` in `.env`
- Restart the server after changing environment variables

### Token validation fails
- Ensure JWT secrets match between services
- Check token expiration
- Verify Authorization header format

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions, please open an issue on GitHub or contact the development team.

## Roadmap

- [ ] OAuth2 Authorization Code flow
- [ ] Social login (Google, GitHub)
- [ ] Two-factor authentication (2FA)
- [ ] Rate limiting
- [ ] Email service integration
- [ ] Password strength requirements
- [ ] Account lockout after failed attempts
- [ ] API documentation with Swagger
- [ ] Metrics and monitoring
- [ ] Admin dashboard
