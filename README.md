# Union Products SSO - Frontend Components

This repository contains the frontend components for the Union Products SSO system. The backend authentication service has been migrated to the [Evero](../evero) project as the `access` module.

## 📦 What's in This Repository

### 1. Admin Dashboard (`admin-dashboard/`)

A comprehensive admin interface for managing authentication and authorization.

**Features:**
- 👥 User Management
- 🏢 Company/Organization Management  
- 🔑 OAuth2 Client Management
- 📊 Session Monitoring
- 🔍 Audit Log Viewer
- 📈 Analytics Dashboard

**Tech Stack:**
- React 18
- TypeScript
- Vite
- Tailwind CSS
- React Query
- React Router

#### Getting Started

```bash
cd admin-dashboard
npm install
npm run dev
```

The dashboard will be available at `http://localhost:5173`

**Configuration:**

Create `admin-dashboard/.env`:
```env
VITE_API_BASE_URL=http://localhost:3000
VITE_SSO_CLIENT_ID=admin-dashboard
```

**Backend Configuration:**

The Access module uses environment-specific configuration files located in `../evero/config/access/`:
- `local.json` - Local development
- `development.json` - Development environment
- `stage.json` - Staging environment
- `production.json` - Production environment

Example configuration structure:
```json
{
  "app": {
    "name": "Access Service",
    "env": "local"
  },
  "web": {
    "port": 3000,
    "host": "localhost"
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "username": "postgres",
    "password": "password",
    "name": "access_db"
  },
  "jwt": {
    "secret": "your-secret-key",
    "access_expiry": 900,
    "refresh_expiry": 604800
  }
}
```

See `../evero/config/access/` for complete configuration examples.

### 2. TypeScript SDK (`sdk/`)

A ready-to-use TypeScript SDK for integrating SSO authentication into your applications.

**Features:**
- 🔐 Complete authentication flow (register, login, logout)
- 🔄 Automatic token refresh
- ⚛️ React hooks
- 📝 TypeScript definitions
- 🎯 Framework agnostic core
- 💾 Token storage management

#### Installation

```bash
npm install @union-products/sso-sdk
```

#### Quick Start

```typescript
import { SSOClient, useSSOAuth } from '@union-products/sso-sdk';

// Initialize the client
const ssoClient = new SSOClient({
  baseURL: 'http://localhost:3000',  // Updated default port
  clientId: 'your-app-id'
});

// In a React component
function App() {
  const { user, login, logout, isAuthenticated } = useSSOAuth(ssoClient);

  const handleLogin = async () => {
    await login({
      email: 'user@example.com',
      password: 'password123'
    });
  };

  return (
    <div>
      {isAuthenticated ? (
        <>
          <p>Welcome, {user.firstName}!</p>
          <button onClick={logout}>Logout</button>
        </>
      ) : (
        <button onClick={handleLogin}>Login</button>
      )}
    </div>
  );
}
```

**Available Endpoints:**
- POST `/api/v1/auth/register` - User registration
- POST `/api/v1/auth/login` - User login
- POST `/api/v1/auth/refresh` - Refresh access token
- POST `/api/v1/auth/logout` - User logout
- GET `/api/v1/auth/me` - Get current user
- POST `/api/v1/auth/verify-email` - Verify email
- POST `/api/v1/auth/forgot-password` - Request password reset
- POST `/api/v1/auth/reset-password` - Reset password

For complete API documentation, see `../evero/docs/access/QUICK_REFERENCE.md`

## 🔗 Backend Service

The SSO backend service has been integrated into the **Evero** project as the `access` module.

### Running the Backend

From the Evero project:

```bash
cd ../evero

# Build the access module
go build -o bin/access app/access/main.go

# Run it
./bin/access
```

Or using Docker:

```bash
cd ../evero/deployment/access
docker-compose up -d
```

Or using the Makefile:

```bash
cd ../evero
make build-access    # Build the module
make deploy-access   # Deploy with Docker Compose
```

**API Documentation:** 
- Implementation Summary: `../evero/docs/access/IMPLEMENTATION_SUMMARY.md`
- Quick Reference: `../evero/docs/access/QUICK_REFERENCE.md`
- Migration Guide: `../evero/docs/access/MIGRATION_GUIDE.md`

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Frontend Applications                  │
│                                                          │
│  ┌─────────────────┐         ┌──────────────────┐     │
│  │ Admin Dashboard │         │   Your App       │     │
│  │   (React/TS)    │         │  (uses SDK)      │     │
│  └────────┬────────┘         └─────────┬────────┘     │
│           │                             │               │
│           └──────────┬──────────────────┘               │
│                      │                                   │
└──────────────────────┼───────────────────────────────────┘
                       │
                       │ HTTP/REST API
                       │
┌──────────────────────▼───────────────────────────────────┐
│              Evero Backend - Access Module               │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │  Authentication & Authorization Service        │    │
│  │  - JWT token management                        │    │
│  │  - User/Company management                     │    │
│  │  - OAuth2 flows                                │    │
│  │  - Session management                          │    │
│  │  - Two-factor authentication                   │    │
│  │  - Audit logging                               │    │
│  └────────────────────────────────────────────────┘    │
│                                                          │
│  Infrastructure Modules:                                │
│  - Config (Viper)                                       │
│  - Database (PostgreSQL/GORM)                           │
│  - Logger (Logrus/Zap)                                  │
│  - Router (Fiber)                                       │
│  - Validator (go-playground)                            │
│  - Message Broker (Kafka)                               │
│                                                          │
└──────────────────────────────────────────────────────────┘
                       │
                       ▼
                  PostgreSQL
```

## 📚 Documentation

### For Frontend Developers

- **SDK Documentation**: See `sdk/README.md`
- **Admin Dashboard**: See `admin-dashboard/README.md`
- **API Quick Reference**: See `../evero/docs/access/QUICK_REFERENCE.md`
- **Integration Examples**: See `../evero/docs/access/MIGRATION_GUIDE.md`

### For Backend Developers

- **Access Module Overview**: See `../evero/docs/access/IMPLEMENTATION_SUMMARY.md`
- **Implementation Details**: See `../evero/docs/access/`
- **Database Schema**: See `../evero/database/access/migrations/`
- **Platform Infrastructure**: See `../evero/docs/evero/INFRASTRUCTURE.md`
- **Deployment Guide**: See `../evero/docs/evero/DEPLOYMENT.md`

## 🔄 Migration from Standalone SSO

The backend code has been fully migrated to the Evero project. This repository now focuses solely on:

1. **Admin Dashboard** - UI for managing the authentication system
2. **SDK** - Client libraries for consuming the authentication APIs

### What Changed?

**Before:**
```
sso/
├── cmd/              # Go backend (removed)
├── handlers/         # Go handlers (removed)  
├── models/           # Go models (removed)
├── services/         # Go services (removed)
├── admin-dashboard/  # ✅ Kept
└── sdk/              # ✅ Kept
```

**After:**
```
sso/
├── admin-dashboard/  # Admin UI
└── sdk/              # TypeScript SDK

evero/
└── modules/access/   # Backend service (migrated here)
```

## 🚀 Quick Start Guide

### 1. Start the Backend (from Evero)

```bash
cd ../evero

# Using Go directly
go build -o bin/access app/access/main.go
./bin/access

# Or using Makefile
make build-access
make run-access

# Or using Docker
cd deployment/access
docker-compose up -d
```

The API will be available at `http://localhost:3000` (default port)

### 2. Start the Admin Dashboard

```bash
cd admin-dashboard
npm install
npm run dev
```

The dashboard will be available at `http://localhost:5173`

### 3. Use the SDK in Your App

```bash
npm install @union-products/sso-sdk
```

See the [SDK examples](#quick-start) above for integration details.

## 🛠️ Development

### Admin Dashboard Development

```bash
cd admin-dashboard
npm run dev          # Start dev server
npm run build        # Build for production
npm run preview      # Preview production build
npm run lint         # Run ESLint
```

### SDK Development

```bash
cd sdk
npm install
npm run build        # Build the SDK
npm run test         # Run tests
npm run type-check   # TypeScript validation
```

## 📝 License

See the main Evero project for license information.

## 🤝 Contributing

This is part of the Union Products platform. For contribution guidelines, see the main Evero repository.

## 📞 Support

- **Issues**: Report in the Evero repository
- **Documentation**: 
  - Access Module: `../evero/docs/access/IMPLEMENTATION_SUMMARY.md`
  - Infrastructure: `../evero/docs/evero/INFRASTRUCTURE.md`
  - Deployment: `../evero/docs/evero/DEPLOYMENT.md`
- **API Questions**: Check `../evero/docs/access/QUICK_REFERENCE.md`
- **Architecture Overview**: Check `../evero/docs/evero/ARCHITECTURE.md`

## 🔧 Recent Updates

### Infrastructure Improvements
- Modular infrastructure organization (config, database, logger, router, validator, message-broker)
- Backwards-compatible wrapper functions for easy migration
- Structured factory pattern for advanced use cases
- Comprehensive deployment configurations for all modules
- Enhanced documentation in `docs/evero/`

### Access Module Features
- Complete authentication and authorization system
- Multi-tenant (company) support
- OAuth 2.0 authorization code flow
- Two-factor authentication (TOTP)
- Session management with device tracking
- Comprehensive audit logging
- Email verification workflow
- Password reset functionality

For detailed platform architecture and infrastructure setup, see the [Evero Platform Documentation](../evero/docs/evero/).
