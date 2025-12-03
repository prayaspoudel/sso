# Quick Start Guide

## 🚀 Get Started in 3 Minutes

### Step 1: Start the Service

Choose one option:

**Option A: Docker (Easiest)**
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

### Step 2: Test It

```bash
# Health check
curl http://localhost:8080/health

# Should return: {"status":"healthy","service":"sso","version":"1.0.0"}
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
    "password": "admin123"
  }'
```

You'll get:
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

```typescript
// Install/copy the SDK
import { initializeSSO, SSOProvider, useSSO } from './sso-sdk';

// Initialize
const ssoClient = initializeSSO({
  baseURL: 'http://localhost:8080',
  clientId: 'your-module-id',
  redirectUri: 'http://localhost:3001/callback',
});

// Wrap your app
<SSOProvider client={ssoClient}>
  <App />
</SSOProvider>

// Use in components
function MyComponent() {
  const { login, isAuthenticated, user } = useSSO();
  
  return isAuthenticated ? (
    <div>Welcome {user?.firstName}</div>
  ) : (
    <button onClick={() => login({email, password})}>
      Login
    </button>
  );
}
```

## 📚 Next Steps

- Read [README.md](README.md) for full documentation
- Check [API.md](API.md) for API reference
- See [sdk/typescript/README.md](sdk/typescript/README.md) for SDK docs

## 🔧 Useful Commands

```bash
# View logs (Docker)
docker-compose logs -f sso-server

# Stop service (Docker)
docker-compose down

# Rebuild
make build

# Run locally
make run

# Run migrations
make migrate-up
```

## ⚙️ Configuration

Edit `.env` file:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=sso_db

JWT_ACCESS_SECRET=change-in-production
JWT_REFRESH_SECRET=change-in-production

SERVER_PORT=8080
```

## 🌐 Service URLs

- **API**: http://localhost:8080
- **Health**: http://localhost:8080/health
- **Service Info**: http://localhost:8080/

## 📋 Pre-configured Modules

Your micro-frontends are pre-configured:
- `host-app` → localhost:3000
- `crm-module` → localhost:3001
- `inventory-module` → localhost:3002
- `hr-module` → localhost:3003
- `finance-module` → localhost:3004
- `task-module` → localhost:3005

## ❓ Troubleshooting

**Can't connect to database?**
```bash
docker-compose up -d  # This starts PostgreSQL too
```

**Port already in use?**
```bash
# Change SERVER_PORT in .env
SERVER_PORT=8081
```

**CORS errors?**
Add your frontend URL to `ALLOWED_ORIGINS` in `.env`

## ✅ Verification Checklist

- [ ] Service starts without errors
- [ ] Health check returns 200 OK
- [ ] Can register a user
- [ ] Can login and get tokens
- [ ] Can validate tokens
- [ ] Frontend can connect

---

🎉 **You're all set!** Start building your authenticated micro-frontends.

For help: Check README.md or open an issue on GitHub.
