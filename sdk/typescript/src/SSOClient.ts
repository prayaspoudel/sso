export interface SSOConfig {
  baseURL: string;
  clientId: string;
  redirectUri: string;
  storageKey?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  tokenType: string;
}

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  isActive: boolean;
  isVerified: boolean;
  createdAt: string;
  updatedAt: string;
  lastLogin?: string;
}

export interface Company {
  id: string;
  name: string;
  email: string;
  industry: string;
  status: string;
}

export interface LoginResponse extends AuthTokens {
  user: User;
  companies?: Company[];
}

export interface ChangePasswordData {
  oldPassword: string;
  newPassword: string;
}

export interface TokenClaims {
  user_id: string;
  email: string;
  companies: string[];
  exp: number;
  iat: number;
  iss: string;
}

export class SSOClient {
  private config: SSOConfig;
  private storageKey: string;

  constructor(config: SSOConfig) {
    this.config = config;
    this.storageKey = config.storageKey || 'sso_auth_tokens';
  }

  /**
   * Register a new user
   */
  async register(data: RegisterData): Promise<{ user: User; message: string }> {
    const response = await fetch(`${this.config.baseURL}/api/v1/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }

    return response.json();
  }

  /**
   * Login with email and password
   */
  async login(credentials: LoginCredentials): Promise<LoginResponse> {
    const response = await fetch(`${this.config.baseURL}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...credentials,
        clientId: this.config.clientId,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    const data: LoginResponse = await response.json();
    this.setTokens(data);
    return data;
  }

  /**
   * Refresh access token
   */
  async refreshToken(): Promise<LoginResponse> {
    const tokens = this.getTokens();
    if (!tokens || !tokens.refreshToken) {
      throw new Error('No refresh token available');
    }

    const response = await fetch(`${this.config.baseURL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refreshToken: tokens.refreshToken,
      }),
    });

    if (!response.ok) {
      this.clearTokens();
      throw new Error('Token refresh failed');
    }

    const data: LoginResponse = await response.json();
    this.setTokens(data);
    return data;
  }

  /**
   * Logout user
   */
  async logout(): Promise<void> {
    const tokens = this.getTokens();
    if (tokens && tokens.refreshToken) {
      try {
        await fetch(`${this.config.baseURL}/api/v1/auth/logout`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${tokens.accessToken}`,
          },
          body: JSON.stringify({
            refreshToken: tokens.refreshToken,
          }),
        });
      } catch (error) {
        console.error('Logout request failed:', error);
      }
    }
    this.clearTokens();
  }

  /**
   * Logout from all devices
   */
  async logoutAll(): Promise<void> {
    const tokens = this.getTokens();
    if (!tokens || !tokens.accessToken) {
      throw new Error('Not authenticated');
    }

    const response = await fetch(`${this.config.baseURL}/api/v1/auth/logout-all`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${tokens.accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to logout from all devices');
    }

    this.clearTokens();
  }

  /**
   * Validate current token
   */
  async validateToken(): Promise<boolean> {
    const tokens = this.getTokens();
    if (!tokens || !tokens.accessToken) {
      return false;
    }

    try {
      const response = await fetch(`${this.config.baseURL}/api/v1/auth/validate`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${tokens.accessToken}`,
        },
      });

      return response.ok;
    } catch (error) {
      return false;
    }
  }

  /**
   * Get current user info
   */
  async getCurrentUser(): Promise<{ userId: string; email: string }> {
    const tokens = this.getTokens();
    if (!tokens || !tokens.accessToken) {
      throw new Error('Not authenticated');
    }

    const response = await fetch(`${this.config.baseURL}/api/v1/auth/me`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${tokens.accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to get user info');
    }

    return response.json();
  }

  /**
   * Change password
   */
  async changePassword(data: ChangePasswordData): Promise<void> {
    const tokens = this.getTokens();
    if (!tokens || !tokens.accessToken) {
      throw new Error('Not authenticated');
    }

    const response = await fetch(`${this.config.baseURL}/api/v1/auth/change-password`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${tokens.accessToken}`,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to change password');
    }

    // Clear tokens after password change
    this.clearTokens();
  }

  /**
   * Get stored tokens
   */
  getTokens(): AuthTokens | null {
    if (typeof window === 'undefined') {
      return null;
    }

    const stored = localStorage.getItem(this.storageKey);
    if (!stored) return null;

    try {
      return JSON.parse(stored);
    } catch {
      return null;
    }
  }

  /**
   * Store tokens
   */
  private setTokens(tokens: AuthTokens): void {
    if (typeof window !== 'undefined') {
      localStorage.setItem(this.storageKey, JSON.stringify(tokens));
    }
  }

  /**
   * Clear stored tokens
   */
  private clearTokens(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem(this.storageKey);
    }
  }

  /**
   * Get access token for API calls
   */
  getAccessToken(): string | null {
    const tokens = this.getTokens();
    return tokens?.accessToken || null;
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    const tokens = this.getTokens();
    return !!tokens && !!tokens.accessToken;
  }

  /**
   * Decode JWT token (without verification)
   */
  decodeToken(token: string): TokenClaims | null {
    try {
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split('')
          .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
      return JSON.parse(jsonPayload);
    } catch {
      return null;
    }
  }

  /**
   * Check if token is expired
   */
  isTokenExpired(token: string): boolean {
    const claims = this.decodeToken(token);
    if (!claims) return true;

    const now = Math.floor(Date.now() / 1000);
    return claims.exp < now;
  }
}

// Singleton instance
let ssoClient: SSOClient | null = null;

export function initializeSSO(config: SSOConfig): SSOClient {
  ssoClient = new SSOClient(config);
  return ssoClient;
}

export function getSSO(): SSOClient {
  if (!ssoClient) {
    throw new Error('SSO not initialized. Call initializeSSO first.');
  }
  return ssoClient;
}
