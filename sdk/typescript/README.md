# TypeScript SDK for Union Products SSO

This SDK provides a TypeScript/JavaScript client for integrating with the Union Products SSO service.

## Installation

Copy the `sdk/typescript/src` folder to your project or install from npm (once published):

```bash
npm install @union-products/sso-client
```

## Quick Start

### 1. Initialize the SSO Client

```typescript
import { initializeSSO, SSOProvider } from '@union-products/sso-client';

// Initialize the client
const ssoClient = initializeSSO({
  baseURL: 'http://localhost:8080',
  clientId: 'your-module-id', // e.g., 'crm-module'
  redirectUri: 'http://localhost:3001/callback',
  storageKey: 'sso_tokens', // optional, default is 'sso_auth_tokens'
});
```

### 2. Wrap Your App with SSOProvider

```typescript
import React from 'react';
import ReactDOM from 'react-dom/client';
import { SSOProvider } from '@union-products/sso-client';
import App from './App';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <SSOProvider client={ssoClient}>
      <App />
    </SSOProvider>
  </React.StrictMode>
);
```

### 3. Use the SSO Hook in Your Components

```typescript
import { useSSO } from '@union-products/sso-client';

function LoginPage() {
  const { login, isLoading, error } = useSSO();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login({ email, password });
      // Navigate to dashboard or home
    } catch (err) {
      console.error('Login failed:', err);
    }
  };

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
    </form>
  );
}
```

## API Reference

### SSOClient

The main client class for interacting with the SSO service.

#### Constructor

```typescript
new SSOClient(config: SSOConfig)
```

**Config Options:**
- `baseURL` (string): The base URL of the SSO service
- `clientId` (string): Your application's client ID
- `redirectUri` (string): OAuth redirect URI
- `storageKey` (string, optional): LocalStorage key for tokens

#### Methods

##### `register(data: RegisterData): Promise<{ user: User; message: string }>`

Register a new user.

```typescript
const result = await ssoClient.register({
  email: 'user@example.com',
  password: 'securepassword',
  firstName: 'John',
  lastName: 'Doe',
});
```

##### `login(credentials: LoginCredentials): Promise<LoginResponse>`

Login with email and password.

```typescript
const response = await ssoClient.login({
  email: 'user@example.com',
  password: 'securepassword',
});
```

##### `logout(): Promise<void>`

Logout the current user.

```typescript
await ssoClient.logout();
```

##### `logoutAll(): Promise<void>`

Logout from all devices.

```typescript
await ssoClient.logoutAll();
```

##### `refreshToken(): Promise<LoginResponse>`

Refresh the access token.

```typescript
const response = await ssoClient.refreshToken();
```

##### `validateToken(): Promise<boolean>`

Validate the current access token.

```typescript
const isValid = await ssoClient.validateToken();
```

##### `getCurrentUser(): Promise<{ userId: string; email: string }>`

Get current user information.

```typescript
const user = await ssoClient.getCurrentUser();
```

##### `changePassword(data: ChangePasswordData): Promise<void>`

Change the user's password.

```typescript
await ssoClient.changePassword({
  oldPassword: 'oldpassword',
  newPassword: 'newpassword',
});
```

##### `getAccessToken(): string | null`

Get the stored access token.

```typescript
const token = ssoClient.getAccessToken();
```

##### `isAuthenticated(): boolean`

Check if the user is authenticated.

```typescript
if (ssoClient.isAuthenticated()) {
  // User is logged in
}
```

##### `decodeToken(token: string): TokenClaims | null`

Decode a JWT token (without verification).

```typescript
const claims = ssoClient.decodeToken(accessToken);
```

##### `isTokenExpired(token: string): boolean`

Check if a token is expired.

```typescript
const expired = ssoClient.isTokenExpired(accessToken);
```

### useSSO Hook

React hook for accessing SSO functionality.

**Returns:**
```typescript
{
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  refreshToken: () => Promise<void>;
  changePassword: (data: ChangePasswordData) => Promise<void>;
}
```

## Examples

### Protected Route Component

```typescript
import { useSSO } from '@union-products/sso-client';
import { Navigate } from 'react-router-dom';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useSSO();

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
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

### User Profile Component

```typescript
import { useSSO } from '@union-products/sso-client';

function UserProfile() {
  const { user, logout } = useSSO();

  if (!user) return null;

  return (
    <div>
      <h2>Welcome, {user.firstName}!</h2>
      <p>Email: {user.email}</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

### Registration Form

```typescript
import { useSSO } from '@union-products/sso-client';
import { useState } from 'react';

function RegisterForm() {
  const { register, isLoading } = useSSO();
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    firstName: '',
    lastName: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await register(formData);
      alert('Registration successful! Please check your email.');
    } catch (error) {
      console.error('Registration failed:', error);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="First Name"
        value={formData.firstName}
        onChange={(e) => setFormData({ ...formData, firstName: e.target.value })}
      />
      <input
        type="text"
        placeholder="Last Name"
        value={formData.lastName}
        onChange={(e) => setFormData({ ...formData, lastName: e.target.value })}
      />
      <input
        type="email"
        placeholder="Email"
        value={formData.email}
        onChange={(e) => setFormData({ ...formData, email: e.target.value })}
      />
      <input
        type="password"
        placeholder="Password"
        value={formData.password}
        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
      />
      <button type="submit" disabled={isLoading}>
        Register
      </button>
    </form>
  );
}
```

### API Client with Auto-Refresh

```typescript
import { getSSO } from '@union-products/sso-client';

async function apiRequest(url: string, options: RequestInit = {}) {
  const sso = getSSO();
  let token = sso.getAccessToken();

  // Check if token is expired
  if (token && sso.isTokenExpired(token)) {
    try {
      await sso.refreshToken();
      token = sso.getAccessToken();
    } catch (error) {
      // Refresh failed, redirect to login
      window.location.href = '/login';
      throw error;
    }
  }

  const response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      Authorization: `Bearer ${token}`,
    },
  });

  if (response.status === 401) {
    // Try to refresh token once
    try {
      await sso.refreshToken();
      token = sso.getAccessToken();
      
      // Retry request
      return fetch(url, {
        ...options,
        headers: {
          ...options.headers,
          Authorization: `Bearer ${token}`,
        },
      });
    } catch (error) {
      window.location.href = '/login';
      throw error;
    }
  }

  return response;
}

// Usage
const data = await apiRequest('https://api.example.com/data').then(r => r.json());
```

### Change Password Component

```typescript
import { useSSO } from '@union-products/sso-client';
import { useState } from 'react';

function ChangePasswordForm() {
  const { changePassword } = useSSO();
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await changePassword({ oldPassword, newPassword });
      alert('Password changed! Please login again.');
      // Redirect to login
    } catch (error) {
      console.error('Failed to change password:', error);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="password"
        placeholder="Old Password"
        value={oldPassword}
        onChange={(e) => setOldPassword(e.target.value)}
      />
      <input
        type="password"
        placeholder="New Password"
        value={newPassword}
        onChange={(e) => setNewPassword(e.target.value)}
      />
      <button type="submit">Change Password</button>
    </form>
  );
}
```

## TypeScript Types

All types are exported from the main package:

```typescript
import type {
  SSOConfig,
  LoginCredentials,
  RegisterData,
  AuthTokens,
  User,
  Company,
  LoginResponse,
  ChangePasswordData,
  TokenClaims,
} from '@union-products/sso-client';
```

## Building the SDK

```bash
cd sdk/typescript
npm install
npm run build
```

This will generate the compiled JavaScript and TypeScript definitions in the `dist/` folder.

## License

MIT
