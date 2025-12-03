# Union Products SSO - API Documentation

## Base URL

```
Development: http://localhost:8080
Production: https://sso.yourcompany.com
```

## Authentication

Most endpoints require authentication using JWT tokens. Include the access token in the Authorization header:

```
Authorization: Bearer <access_token>
```

## Response Format

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

## HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required or failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Endpoints

### Authentication

#### Register User

Creates a new user account.

**Endpoint:** `POST /api/v1/auth/register`

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
    "createdAt": "2025-10-20T10:00:00Z",
    "updatedAt": "2025-10-20T10:00:00Z"
  },
  "message": "Registration successful. Please verify your email."
}
```

**Errors:**
- `400`: User already exists
- `400`: Invalid request parameters

---

#### Login

Authenticates a user and returns access and refresh tokens.

**Endpoint:** `POST /api/v1/auth/login`

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
    "lastLogin": "2025-10-20T10:00:00Z"
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

---

#### Refresh Token

Exchanges a refresh token for a new access token and refresh token.

**Endpoint:** `POST /api/v1/auth/refresh`

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

#### Logout

Revokes the refresh token and ends the session.

**Endpoint:** `POST /api/v1/auth/logout`

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

#### Validate Token

Validates an access token and returns the claims.

**Endpoint:** `GET /api/v1/auth/validate`

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

---

### Protected Endpoints

These endpoints require authentication.

#### Get Current User

Returns information about the authenticated user.

**Endpoint:** `GET /api/v1/auth/me`

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

#### Change Password

Changes the user's password.

**Endpoint:** `POST /api/v1/auth/change-password`

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

#### Logout All Devices

Revokes all refresh tokens and sessions for the user.

**Endpoint:** `POST /api/v1/auth/logout-all`

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

Returns the health status of the service.

**Endpoint:** `GET /health`

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

Returns information about the SSO service and available endpoints.

**Endpoint:** `GET /`

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

## JWT Token Structure

### Access Token Claims

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

### Token Expiry

- **Access Token**: 15 minutes
- **Refresh Token**: 7 days

## Rate Limiting

(To be implemented)

- Login: 5 attempts per minute per IP
- Register: 3 attempts per minute per IP
- Refresh: 10 attempts per minute per user

## CORS

The service supports CORS for the following origins (configurable):

- `http://localhost:3000` - Host Application
- `http://localhost:3001` - CRM Module
- `http://localhost:3002` - Inventory Module
- `http://localhost:3003` - HR Module
- `http://localhost:3004` - Finance Module
- `http://localhost:3005` - Task Module

## Error Codes

| Code | Description |
|------|-------------|
| `AUTH_001` | Invalid credentials |
| `AUTH_002` | Account inactive |
| `AUTH_003` | Account not verified |
| `AUTH_004` | Token expired |
| `AUTH_005` | Token invalid |
| `AUTH_006` | Token revoked |
| `USER_001` | User already exists |
| `USER_002` | User not found |
| `PASS_001` | Invalid old password |
| `PASS_002` | Password too weak |

## Examples

### cURL Examples

**Register:**
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

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "clientId": "crm-module"
  }'
```

**Validate Token:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer <your-access-token>"
```

### JavaScript Examples

**Using Fetch:**
```javascript
// Login
const response = await fetch('http://localhost:8080/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'SecurePass123!',
    clientId: 'crm-module',
  }),
});

const data = await response.json();
console.log(data.accessToken);
```

**Using Axios:**
```javascript
import axios from 'axios';

// Login
const { data } = await axios.post('http://localhost:8080/api/v1/auth/login', {
  email: 'user@example.com',
  password: 'SecurePass123!',
  clientId: 'crm-module',
});

// Store tokens
localStorage.setItem('accessToken', data.accessToken);
localStorage.setItem('refreshToken', data.refreshToken);

// Make authenticated request
const userInfo = await axios.get('http://localhost:8080/api/v1/auth/me', {
  headers: {
    Authorization: `Bearer ${data.accessToken}`,
  },
});
```

## Best Practices

1. **Always use HTTPS in production**
2. **Store tokens securely** (httpOnly cookies or secure storage)
3. **Implement token refresh** before access token expires
4. **Handle 401 errors** by refreshing token or redirecting to login
5. **Clear tokens on logout**
6. **Validate tokens** on protected routes
7. **Use environment-specific URLs**
8. **Implement proper error handling**
9. **Log security events**
10. **Rotate secrets regularly**

## Support

For API support, please contact the development team or open an issue on GitHub.
