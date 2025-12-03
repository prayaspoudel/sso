import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { SSOClient, User, LoginCredentials, RegisterData, ChangePasswordData } from './SSOClient';

interface SSOContextValue {
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

const SSOContext = createContext<SSOContextValue | undefined>(undefined);

interface SSOProviderProps {
  client: SSOClient;
  children: ReactNode;
}

export const SSOProvider: React.FC<SSOProviderProps> = ({ client, children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is already authenticated
    const checkAuth = async () => {
      try {
        if (client.isAuthenticated()) {
          const isValid = await client.validateToken();
          if (isValid) {
            const currentUser = await client.getCurrentUser();
            // Map the response to User type
            setUser({
              id: currentUser.userId,
              email: currentUser.email,
              firstName: '',
              lastName: '',
              isActive: true,
              isVerified: true,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString(),
            });
          } else {
            // Try to refresh token
            try {
              const response = await client.refreshToken();
              setUser(response.user);
            } catch {
              client.getTokens(); // Clear invalid tokens
            }
          }
        }
      } catch (error) {
        console.error('Auth check failed:', error);
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, [client]);

  const login = async (credentials: LoginCredentials) => {
    setIsLoading(true);
    try {
      const response = await client.login(credentials);
      setUser(response.user);
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (data: RegisterData) => {
    setIsLoading(true);
    try {
      await client.register(data);
      // After registration, you might want to auto-login
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    setIsLoading(true);
    try {
      await client.logout();
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  };

  const logoutAll = async () => {
    setIsLoading(true);
    try {
      await client.logoutAll();
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  };

  const refreshToken = async () => {
    try {
      const response = await client.refreshToken();
      setUser(response.user);
    } catch (error) {
      console.error('Token refresh failed:', error);
      setUser(null);
    }
  };

  const changePassword = async (data: ChangePasswordData) => {
    await client.changePassword(data);
    setUser(null);
  };

  const value: SSOContextValue = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    register,
    logout,
    logoutAll,
    refreshToken,
    changePassword,
  };

  return <SSOContext.Provider value={value}>{children}</SSOContext.Provider>;
};

export const useSSO = (): SSOContextValue => {
  const context = useContext(SSOContext);
  if (!context) {
    throw new Error('useSSO must be used within SSOProvider');
  }
  return context;
};
