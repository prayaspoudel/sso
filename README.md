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
VITE_API_BASE_URL=http://localhost:8080
VITE_SSO_CLIENT_ID=admin-dashboard
```

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
  baseURL: 'http://localhost:8080',
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

**API Documentation:** See `../evero/docs/ACCESS_README.md`

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
│  │  - Audit logging                               │    │
│  └────────────────────────────────────────────────┘    │
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
- **API Reference**: See `../evero/docs/access/QUICK_REFERENCE.md`

### For Backend Developers

- **Access Module**: See `../evero/docs/ACCESS_README.md`
- **Implementation Details**: See `../evero/docs/access/`
- **Database Schema**: See `../evero/database/access/migrations/`

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
go build -o bin/access app/access/main.go
./bin/access
```

### 2. Start the Admin Dashboard

```bash
cd admin-dashboard
npm install
npm run dev
```

### 3. Use the SDK in Your App

```bash
npm install @union-products/sso-sdk
```

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
- **Documentation**: See `../evero/docs/ACCESS_README.md`
- **API Questions**: Check `../evero/docs/access/`
