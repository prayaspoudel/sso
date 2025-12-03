# Union Products SSO Service - Complete Documentation

A comprehensive Single Sign-On (SSO) authentication service built with Go and PostgreSQL, designed to work seamlessly with micro-frontend architectures.

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start (3 Minutes)](#quick-start-3-minutes)
- [Configuration](#configuration)
- [Complete API Reference](#complete-api-reference)
- [Frontend Integration & SDK](#frontend-integration--sdk)
- [Database Schema](#database-schema)
- [Testing Guide](#testing-guide)
- [Security Best Practices](#security-best-practices)
- [Development Guide](#development-guide)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)
- [Project Structure](#project-structure)

---

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

---

## Architecture

This SSO service is designed to work with the following micro-frontend modules:

| Module | Port | Client ID |
|--------|------|-----------|
| Host Application | 3000 | `host-app` |
| CRM Module | 3001 | `crm-module` |
| Inventory Module | 3002 | `inventory-module` |
| HR Module | 3003 | `hr-module` |
| Finance Module | 3004 | `finance-module` |
| Task Module | 3005 | `task-module` |

### Architecture Diagram

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

---

## Prerequisites

- **Go** 1.21 or higher
- **PostgreSQL** 15 or higher
- **Docker & Docker Compose** (optional, but recommended)
- **Node.js** 18+ (for TypeScript SDK)

---

## Quick Start (3 Minutes)

### Step 1: Choose Your Setup Method

#### Option A: Docker (Easiest) 🐳
```bash
cd sso
docker-compose up -d
```

#### Option B: Automated Setup Script
```bash
cd sso
./setup.sh
```

#### Option C: Manual Setup
```bash
# Create database
createdb sso_db

# Run migrations
psql -d sso_db -f database/migrations/001_initial_schema.sql

# Build and run
make build
./bin/sso-server
```

### Step 2: Test the Service

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","service":"sso","version":"1.0.0"}
```

### Step 3: Create Your First User

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

### Step 4: Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "clientId": "host-app"
  }'
```

You'll receive:
```json
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "ZGVm...",
  "expiresIn": 900,
  "tokenType": "Bearer",
  "user": {...}
}
```

### Step 5: Use in Your Frontend

See the [Frontend Integration](#frontend-integration--sdk) section below.

---

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=sso_db
DB_SSLMODE=disable  # Use 'require' in production

# JWT Configuration (⚠️ CHANGE IN PRODUCTION!)
JWT_ACCESS_SECRET=your-super-secret-access-key-min-32-chars
JWT_REFRESH_SECRET=your-super-secret-refresh-key-min-32-chars
JWT_ISSUER=union-products-sso

# Server Configuration
SERVER_PORT=8080
ENV=development  # development, staging, or production

# CORS Configuration
# Add all your frontend URLs (comma-separated)
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:3002,http://localhost:3003,http://localhost:3004,http://localhost:3005

# Token Expiry (in minutes)
ACCESS_TOKEN_EXPIRY=15
REFRESH_TOKEN_EXPIRY=10080  # 7 days

# Email Configuration (Optional, for email verification)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### Service URLs

- **API Base**: `http://localhost:8080`
- **Health Check**: `http://localhost:8080/health`
- **Service Info**: `http://localhost:8080/`

---

## Complete API Reference

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

**Success Response:**
```json
{
  "data": {...},
  "message": "Success message"
}
```

**Error Response:**
```json
{
  "error": "Error message"
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| `200` | Request successful |
| `201` | Resource created successfully |
| `400` | Invalid request parameters |
| `401` | Authentication required or failed |
| `403` | Insufficient permissions |
| `404` | Resource not found |
| `500` | Internal server error |

---

### Public Endpoints

#### 1. Register User

**Endpoint:** `POST /api/v1/auth/register`

**Description:** Creates a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
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
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "isActive": true,
    "isVerified": false,
    "createdAt": "2025-10-25T10:00:00Z",
    "updatedAt": "2025-10-25T10:00:00Z"
  },
  "message": "Registration successful. Please verify your email."
}
```

**Errors:**
- `400`: User already exists
- `400`: Invalid request parameters

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

---

#### 2. Login

**Endpoint:** `POST /api/v1/auth/login`

**Description:** Authenticates a user and returns access and refresh tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
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
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "ZGVmNTAyMjIzMzQ0NTU2Njc3ODg5OTAw",
  "expiresIn": 900,
  "tokenType": "Bearer",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "isActive": true,
    "isVerified": true,
    "lastLogin": "2025-10-25T10:00:00Z"
  },
  "companies": [
    {
      "id": "company-uuid",
      "name": "Acme Corp",
      "email": "contact@acme.com",
      "industry": "Technology",
      "status": "active"
    }
  ]
}
```

**Errors:**
- `401`: Invalid credentials
- `401`: Account is inactive
- `400`: Invalid request parameters

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "clientId": "crm-module"
  }'
```

---

#### 3. Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Description:** Exchanges a refresh token for a new access token and refresh token.

**Request Body:**
```json
{
  "refreshToken": "ZGVmNTAyMjIzMzQ0NTU2Njc3ODg5OTAw"
}
```

**Response:** `200 OK`
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "bmV3LXJlZnJlc2gtdG9rZW4tMTIzNDU2Nzg5",
  "expiresIn": 900,
  "tokenType": "Bearer",
  "user": {...},
  "companies": [...]
}
```

**Errors:**
- `401`: Invalid or expired refresh token
- `401`: Token has been revoked

---

#### 4. Logout

**Endpoint:** `POST /api/v1/auth/logout`

**Description:** Revokes the refresh token and ends the session.

**Request Body:**
```json
{
  "refreshToken": "ZGVmNTAyMjIzMzQ0NTU2Njc3ODg5OTAw"
}
```

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

---

#### 5. Validate Token

**Endpoint:** `GET /api/v1/auth/validate`

**Description:** Validates an access token and returns the claims.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`
```json
{
  "valid": true,
  "claims": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "companies": [],
    "exp": 1729512000,
    "iat": 1729511100,
    "iss": "union-products-sso"
  }
}
```

**Errors:**
- `401`: Invalid or expired token
- `401`: Authorization header required

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer <your-access-token>"
```

---

### Protected Endpoints

These endpoints require authentication (Bearer token in Authorization header).

#### 6. Get Current User

**Endpoint:** `GET /api/v1/auth/me`

**Description:** Returns information about the authenticated user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`
```json
{
  "userId": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com"
}
```

---

#### 7. Change Password

**Endpoint:** `POST /api/v1/auth/change-password`

**Description:** Changes the user's password.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "oldPassword": "OldSecurePass123!",
  "newPassword": "NewSecurePass456!"
}
```

**Validation:**
- `oldPassword`: Required
- `newPassword`: Required, minimum 8 characters

**Response:** `200 OK`
```json
{
  "message": "Password changed successfully. Please login again."
}
```

**Note:** All existing tokens are revoked after password change.

**Errors:**
- `400`: Invalid old password
- `400`: Invalid request parameters
- `401`: Not authenticated

---

#### 8. Logout All Devices

**Endpoint:** `POST /api/v1/auth/logout-all`

**Description:** Revokes all refresh tokens and sessions for the user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`
```json
{
  "message": "Logged out from all devices successfully"
}
```

---

### System Endpoints

#### Health Check

**Endpoint:** `GET /health`

**Description:** Returns the health status of the service.

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "service": "sso",
  "version": "1.0.0"
}
```

---

#### Service Info

**Endpoint:** `GET /`

**Description:** Returns information about the SSO service and available endpoints.

**Response:** `200 OK`
```json
{
  "service": "Union Products SSO",
  "version": "1.0.0",
  "endpoints": {
    "health": "/health",
    "register": "/api/v1/auth/register",
    "login": "/api/v1/auth/login",
    "refresh": "/api/v1/auth/refresh",
    "logout": "/api/v1/auth/logout",
    "validate": "/api/v1/auth/validate",
    "me": "/api/v1/auth/me (protected)"
  }
}
```

---

### JWT Token Structure

#### Access Token Claims

```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "companies": ["company-id-1", "company-id-2"],
  "exp": 1729512000,
  "iat": 1729511100,
  "iss": "union-products-sso"
}
```

#### Token Expiry

- **Access Token**: 15 minutes
- **Refresh Token**: 7 days

---

## Frontend Integration & SDK

### TypeScript SDK Installation

**Option 1: Copy to your project**
```bash
cp -r sso/sdk/typescript/src ./src/sso-sdk
```

**Option 2: Install from npm** (once published)
```bash
npm install @union-products/sso-client
```

### React Integration

#### 1. Initialize the SSO Client

```typescript
// main.tsx or App.tsx
import { initializeSSO, SSOProvider } from './sso-sdk';

const ssoClient = initializeSSO({
  baseURL: 'http://localhost:8080',
  clientId: 'crm-module', // Change per module
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

#### 2. Use the SSO Hook

```typescript
import { useSSO } from './sso-sdk';

function LoginPage() {
  const { login, isLoading, isAuthenticated, error } = useSSO();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login({ email, password });
      // Redirect to dashboard
    } catch (err) {
      console.error('Login failed:', err);
    }
  };

  if (isAuthenticated) {
    return <Navigate to="/dashboard" />;
  }

  return (
    <form onSubmit={handleLogin}>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
      />
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
      />
      <button type="submit" disabled={isLoading}>
        {isLoading ? 'Logging in...' : 'Login'}
      </button>
      {error && <div className="error">{error}</div>}
    </form>
  );
}
```

#### 3. Protected Routes

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

// Usage
<Route
  path="/dashboard"
  element={
    <ProtectedRoute>
      <Dashboard />
    </ProtectedRoute>
  }
/>
```

#### 4. User Profile Component

```typescript
import { useSSO } from './sso-sdk';

function UserProfile() {
  const { user, logout, isLoading } = useSSO();

  if (isLoading || !user) return <div>Loading...</div>;

  return (
    <div>
      <h2>Welcome, {user.firstName}!</h2>
      <p>Email: {user.email}</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

#### 5. Making Authenticated API Calls

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

### SDK API Reference

#### useSSO Hook

Returns:
```typescript
{
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  refreshToken: () => Promise<void>;
  changePassword: (data: ChangePasswordData) => Promise<void>;
}
```

For complete SDK documentation, see `sdk/typescript/README.md`.

---

## Database Schema

The SSO service uses PostgreSQL with the following tables:

### Tables

1. **users** - User accounts and profiles
   - id (UUID, Primary Key)
   - email (Unique)
   - password_hash
   - first_name, last_name
   - is_active, is_verified
   - created_at, updated_at, last_login

2. **companies** - Company/organization information
   - id (UUID, Primary Key)
   - name, email, phone
   - industry, size, status
   - created_at, updated_at

3. **user_companies** - User-company relationships with roles
   - user_id (Foreign Key → users)
   - company_id (Foreign Key → companies)
   - role (e.g., 'admin', 'member')
   - joined_at

4. **oauth_clients** - Registered OAuth2 clients (micro-frontends)
   - client_id (Primary Key)
   - client_secret
   - name, redirect_uri
   - allowed_grants, allowed_scopes

5. **refresh_tokens** - Refresh tokens for token rotation
   - token (Primary Key)
   - user_id (Foreign Key → users)
   - client_id (Foreign Key → oauth_clients)
   - expires_at, revoked_at

6. **sessions** - Active user sessions
   - id (UUID, Primary Key)
   - user_id (Foreign Key → users)
   - token, ip_address, user_agent
   - created_at, expires_at

7. **audit_logs** - Security audit trail
   - id (Serial, Primary Key)
   - user_id, action, resource
   - ip_address, user_agent
   - created_at

8. **password_reset_tokens** - Password reset functionality
   - token (Primary Key)
   - user_id (Foreign Key → users)
   - expires_at, used_at

9. **email_verification_tokens** - Email verification
   - token (Primary Key)
   - user_id (Foreign Key → users)
   - expires_at, verified_at

### Migrations

Run migrations:
```bash
psql -d sso_db -f database/migrations/001_initial_schema.sql
```

Rollback:
```bash
psql -d sso_db -f database/migrations/002_rollback.sql
```

---

## Testing Guide

### Server Status

Check if the server is running:
```bash
# Health check
curl http://localhost:8080/health

# View logs (Docker)
docker-compose logs -f sso-server

# Check process
ps aux | grep sso-server
```

### Test All Endpoints

```bash
# 1. Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!",
    "firstName": "Test",
    "lastName": "User"
  }'

# 2. Login
TOKEN_RESPONSE=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!",
    "clientId": "host-app"
  }')

echo $TOKEN_RESPONSE

# Extract access token (using jq)
ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.accessToken')

# 3. Get current user
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 4. Validate token
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 5. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$(echo $TOKEN_RESPONSE | jq -r '.refreshToken')\"}"
```

### Database Verification

```bash
# Connect to database
docker exec -it postgres-container psql -U postgres -d sso_db

# Check tables
\dt

# View users
SELECT id, email, first_name, is_active FROM users;

# View OAuth clients
SELECT client_id, name FROM oauth_clients;

# Exit
\q
```

---

## Security Best Practices

### 🔒 Essential Security Measures

1. **Change Default Secrets**
   - Update `JWT_ACCESS_SECRET` and `JWT_REFRESH_SECRET` with strong random values (32+ characters)
   - Use different secrets for access and refresh tokens
   ```bash
   # Generate secure secrets
   openssl rand -base64 32
   ```

2. **Use HTTPS in Production**
   - Always use HTTPS/TLS in production
   - Enable `DB_SSLMODE=require` for database connections
   - Configure SSL certificates with Let's Encrypt or your provider

3. **Secure Database**
   - Use strong database passwords
   - Restrict database access to specific IP addresses
   - Enable PostgreSQL SSL connections
   - Regular backups

4. **Token Management**
   - Access tokens expire in 15 minutes (configurable)
   - Refresh tokens expire in 7 days (configurable)
   - Tokens are automatically rotated on refresh
   - All tokens revoked on password change

5. **CORS Configuration**
   - Only allow trusted origins in `ALLOWED_ORIGINS`
   - Never use `*` in production
   - Specify exact URLs including protocol and port

6. **Rate Limiting** (TODO: Implement)
   - Login attempts: 5 per minute per IP
   - Registration: 3 per minute per IP
   - API calls: 100 per minute per user

7. **Password Requirements**
   - Minimum 8 characters (enforce stronger policies)
   - Consider implementing: uppercase, lowercase, number, special char
   - Implement password strength meter on frontend

8. **Monitoring & Logging**
   - All authentication events logged to audit_logs table
   - Monitor failed login attempts
   - Set up alerts for suspicious activity

9. **Session Management**
   - Track active sessions per user
   - Allow users to view and revoke sessions
   - Automatic session cleanup for expired tokens

10. **Security Headers**
    - Implement security headers (HSTS, CSP, X-Frame-Options)
    - Use secure cookies for sensitive data
    - Implement CSRF protection

---

## Development Guide

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
│       ├── 001_initial_schema.sql
│       └── 002_rollback.sql
├── handlers/
│   └── auth_handler.go       # HTTP handlers
├── middleware/
│   └── middleware.go         # HTTP middleware (CORS, Auth, Logging)
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
│       ├── src/
│       │   ├── SSOClient.ts
│       │   ├── SSOContext.tsx
│       │   └── index.ts
│       ├── package.json
│       └── README.md
├── bin/
│   └── sso-server           # Compiled binary
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── setup.sh
├── .env
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

### Available Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make test-coverage # Run tests with coverage
make docker-build  # Build Docker image
make docker-up     # Start Docker containers
make docker-down   # Stop Docker containers
make migrate-up    # Run database migrations
make migrate-down  # Rollback migrations
make clean         # Clean build files
make lint          # Run linter
```

### Development Workflow

1. **Make Changes**
   ```bash
   # Edit code
   vim handlers/auth_handler.go
   ```

2. **Build**
   ```bash
   make build
   ```

3. **Test**
   ```bash
   make test
   ```

4. **Run Locally**
   ```bash
   make run
   ```

5. **Check for Errors**
   ```bash
   # View logs
   tail -f sso.log
   ```

### Adding New Endpoints

1. Add handler function in `handlers/auth_handler.go`
2. Register route in `cmd/server/main.go`
3. Add service logic in `services/auth_service.go`
4. Add repository methods if needed
5. Update API documentation
6. Add tests

---

## Production Deployment

### Pre-Deployment Checklist

- [ ] Change JWT secrets to secure random values
- [ ] Update `.env` with production values
- [ ] Set `ENV=production`
- [ ] Enable database SSL (`DB_SSLMODE=require`)
- [ ] Configure proper `ALLOWED_ORIGINS`
- [ ] Set up SSL/TLS certificates
- [ ] Configure reverse proxy (nginx/Caddy)
- [ ] Set up logging and monitoring
- [ ] Configure database backups
- [ ] Implement rate limiting
- [ ] Set up health checks and alerts
- [ ] Review and adjust token expiry times
- [ ] Configure firewall rules
- [ ] Set up email service (SMTP)

### Docker Deployment

1. **Build the Docker image**
   ```bash
   docker build -t sso-service:latest .
   ```

2. **Run the container**
   ```bash
   docker run -d \
     -p 8080:8080 \
     -e DB_HOST=your-db-host \
     -e DB_PASSWORD=secure-password \
     -e JWT_ACCESS_SECRET=your-secret \
     -e JWT_REFRESH_SECRET=your-secret \
     -e ENV=production \
     --name sso-service \
     sso-service:latest
   ```

3. **Use Docker Compose**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name sso.yourcompany.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name sso.yourcompany.com;

    ssl_certificate /etc/letsencrypt/live/sso.yourcompany.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sso.yourcompany.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Health Checks

```bash
# Docker health check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080/health || exit 1
```

### Monitoring

Set up monitoring for:
- Server uptime
- Response times
- Error rates
- Database connections
- Failed login attempts
- Token generation/validation rates

---

## Troubleshooting

### Common Issues

#### 1. Cannot connect to database

**Symptoms:**
```
Failed to ping database: pq: password authentication failed
```

**Solutions:**
- Check database credentials in `.env`
- Ensure PostgreSQL is running
  ```bash
  docker ps | grep postgres
  # or
  pg_isready
  ```
- Verify network connectivity
- Check if `.env` file is being loaded

**Fix that worked:**
Added `godotenv` package to load `.env` file:
```go
import "github.com/joho/godotenv"

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found")
    }
    // ... rest of config
}
```

---

#### 2. Port already in use

**Symptoms:**
```
listen tcp :8080: bind: address already in use
```

**Solutions:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill <PID>

# Or change port in .env
SERVER_PORT=8081
```

---

#### 3. CORS errors

**Symptoms:**
```
Access to fetch at 'http://localhost:8080' from origin 'http://localhost:3001' 
has been blocked by CORS policy
```

**Solutions:**
- Add your frontend URL to `ALLOWED_ORIGINS` in `.env`
  ```env
  ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
  ```
- Restart the server after changing environment variables
- Check that the Origin header matches exactly (including port)

---

#### 4. Token validation fails

**Symptoms:**
```
401 Unauthorized: Invalid token
```

**Solutions:**
- Ensure JWT secrets match between services
- Check token expiration (`exp` claim)
- Verify Authorization header format: `Bearer <token>`
- Check if token was revoked (password change, logout)

---

#### 5. Docker build fails

**Solutions:**
```bash
# Clean Docker cache
docker system prune -a

# Rebuild from scratch
docker-compose build --no-cache

# Check logs
docker-compose logs sso-server
```

---

#### 6. Migration fails

**Solutions:**
```bash
# Check if database exists
psql -l | grep sso_db

# Create database
createdb sso_db

# Try migration again
psql -d sso_db -f database/migrations/001_initial_schema.sql

# Check for errors
psql -d sso_db -c "\dt"
```

---

### Debug Mode

Enable debug logging:
```env
ENV=development
LOG_LEVEL=debug
```

View detailed logs:
```bash
# Follow logs
tail -f sso.log

# Docker logs
docker-compose logs -f sso-server

# Search logs
grep "ERROR" sso.log
```

---

## FAQ

**Q: How do I change token expiry times?**
A: Update the constants in `services/auth_service.go`:
```go
const (
    AccessTokenExpiry  = 15 * time.Minute  // Change this
    RefreshTokenExpiry = 7 * 24 * time.Hour  // Change this
)
```

**Q: Can I use this with non-React frontends?**
A: Yes! The SDK is framework-agnostic. You can use the `SSOClient` class directly without React hooks.

**Q: How do I add a new OAuth client?**
A: Insert into the `oauth_clients` table:
```sql
INSERT INTO oauth_clients (client_id, client_secret, name, redirect_uri)
VALUES ('new-module', 'secret', 'New Module', 'http://localhost:3006/callback');
```

**Q: How do I reset a user's password from the database?**
A: Generate a new bcrypt hash and update:
```bash
# Generate hash (using htpasswd or online tool)
# Then update:
psql -d sso_db -c "UPDATE users SET password_hash = 'new-hash' WHERE email = 'user@example.com';"
```

**Q: Can users belong to multiple companies?**
A: Yes! The `user_companies` table supports many-to-many relationships with role support.

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT License - see LICENSE file for details

---

## Support

For issues and questions:
- Open an issue on GitHub
- Check the [Troubleshooting](#troubleshooting) section
- Review the API documentation
- Contact the development team

---

## Roadmap

### Planned Features

- [ ] OAuth2 Authorization Code flow
- [ ] Social login (Google, GitHub, LinkedIn)
- [ ] Two-factor authentication (2FA/MFA)
- [ ] Rate limiting implementation
- [ ] Email service integration (SendGrid/Mailgun)
- [ ] SMS verification (Twilio)
- [ ] Password strength requirements enforcement
- [ ] Account lockout after failed attempts
- [ ] API documentation with Swagger/OpenAPI
- [ ] Prometheus metrics
- [ ] Grafana dashboards
- [ ] Admin dashboard UI
- [ ] User management API
- [ ] Company/organization management
- [ ] Role-based access control (RBAC)
- [ ] Audit log search and filtering
- [ ] WebSocket support for real-time notifications
- [ ] Mobile SDK (React Native)

---

## Changelog

### v1.0.0 (2025-10-25)
- ✅ Initial release
- ✅ Complete authentication API
- ✅ JWT token management
- ✅ TypeScript/React SDK
- ✅ Docker support
- ✅ PostgreSQL schema
- ✅ Session management
- ✅ Audit logging
- ✅ Multi-company support
- ✅ Comprehensive documentation

---

## Acknowledgments

Built with:
- [Go](https://golang.org/) - Backend language
- [Gin](https://gin-gonic.com/) - HTTP framework
- [PostgreSQL](https://www.postgresql.org/) - Database
- [JWT](https://jwt.io/) - Token authentication
- [Docker](https://www.docker.com/) - Containerization
- [React](https://reactjs.org/) - Frontend framework (SDK)
- [TypeScript](https://www.typescriptlang.org/) - Type safety (SDK)

---

## Summary

🎉 **You now have a complete, production-ready SSO service!**

### What You Get:
- ✅ Secure JWT-based authentication
- ✅ Complete REST API with 8+ endpoints
- ✅ TypeScript/React SDK with hooks
- ✅ Multi-company support
- ✅ Session management
- ✅ Audit logging
- ✅ Docker deployment
- ✅ Comprehensive documentation

### Quick Links:
- **API**: http://localhost:8080
- **Health**: http://localhost:8080/health
- **SDK Docs**: `sdk/typescript/README.md`

### Next Steps:
1. ✅ Start the service
2. ✅ Test the API
3. ✅ Integrate with frontends
4. ✅ Deploy to production

**Happy coding! 🚀**

---

*Last updated: October 25, 2025*
*Version: 1.0.0*
