// API Configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
export const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8080/api/v1';

// API Endpoints
export const API_ENDPOINTS = {
  // Auth
  AUTH: {
    LOGIN: '/auth/login',
    REGISTER: '/auth/register',
    LOGOUT: '/auth/logout',
    REFRESH: '/auth/refresh',
    ME: '/auth/me',
    CHANGE_PASSWORD: '/auth/change-password',
  },
  // Users
  USERS: {
    LIST: '/users',
    GET: (id: string) => `/users/${id}`,
    CREATE: '/users',
    UPDATE: (id: string) => `/users/${id}`,
    DELETE: (id: string) => `/users/${id}`,
    STATS: '/users/stats',
  },
  // Companies
  COMPANIES: {
    LIST: '/companies',
    GET: (id: string) => `/companies/${id}`,
    CREATE: '/companies',
    UPDATE: (id: string) => `/companies/${id}`,
    DELETE: (id: string) => `/companies/${id}`,
    STATS: '/companies/stats',
    USERS: (id: string) => `/companies/${id}/users`,
    ADD_USER: (id: string) => `/companies/${id}/users`,
    REMOVE_USER: (id: string, userId: string) => `/companies/${id}/users/${userId}`,
  },
  // Audit Logs
  AUDIT_LOGS: {
    LIST: '/audit-logs',
    GET: (id: string) => `/audit-logs/${id}`,
    STATS: '/audit-logs/stats',
    TIMELINE: '/audit-logs/timeline',
    EXPORT: '/audit-logs/export',
    ACTIONS: '/audit-logs/actions',
    RESOURCES: '/audit-logs/resources',
    COMPARE: '/audit-logs/compare',
  },
  // Notifications
  NOTIFICATIONS: {
    LIST: '/notifications',
    GET: (id: string) => `/notifications/${id}`,
    UNREAD_COUNT: '/notifications/unread-count',
    STATS: '/notifications/stats',
    CREATE: '/notifications',
    BROADCAST: '/notifications/broadcast',
    MARK_READ: (id: string) => `/notifications/${id}/read`,
    MARK_MULTIPLE_READ: '/notifications/read',
    MARK_ALL_READ: '/notifications/read-all',
    DELETE: (id: string) => `/notifications/${id}`,
    PREFERENCES: '/notifications/preferences',
    CONNECTIONS: '/notifications/connections',
    TEST: '/notifications/test',
  },
  // WebSocket
  WS: '/ws',
};

// Storage keys
export const STORAGE_KEYS = {
  TOKEN: 'auth_token',
  REFRESH_TOKEN: 'refresh_token',
  USER: 'user',
};
