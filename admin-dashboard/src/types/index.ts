// User Types
export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone_number?: string;
  role: string;
  status: 'active' | 'inactive' | 'suspended' | 'locked';
  company_id?: string;
  created_at: string;
  updated_at: string;
  last_login?: string;
  login_count: number;
  failed_login_attempts: number;
}

export interface UserStats {
  total_users: number;
  active_users: number;
  inactive_users: number;
  suspended_users: number;
  locked_users: number;
  recent_logins: number;
  users_by_role: Record<string, number>;
  users_by_company: Record<string, number>;
}

// Company Types
export interface Company {
  id: string;
  name: string;
  domain: string;
  status: 'active' | 'inactive' | 'suspended';
  settings?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface CompanyStats {
  total_companies: number;
  active_companies: number;
  inactive_companies: number;
  suspended_companies: number;
  total_users: number;
  companies_by_domain: Record<string, number>;
}

// Audit Log Types
export interface AuditLog {
  id: string;
  timestamp: string;
  user_id?: string;
  user_email?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  details?: Record<string, unknown>;
  ip_address?: string;
  user_agent?: string;
  status: 'success' | 'failure';
  error_message?: string;
}

export interface AuditLogStats {
  total_logs: number;
  success_count: number;
  failure_count: number;
  actions_by_type: Record<string, number>;
  actions_by_resource: Record<string, number>;
  actions_by_user: Record<string, number>;
  actions_by_hour: Record<string, number>;
}

// Notification Types
export type NotificationType = string;
export type NotificationPriority = 'low' | 'normal' | 'high' | 'critical';
export type NotificationStatus = 'unread' | 'read' | 'archived';

export interface Notification {
  id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  message: string;
  data?: Record<string, unknown>;
  priority: NotificationPriority;
  status: NotificationStatus;
  created_at: string;
  read_at?: string;
  expires_at?: string;
}

export interface NotificationPreferences {
  user_id: string;
  email_enabled: boolean;
  push_enabled: boolean;
  sms_enabled: boolean;
  notification_types?: Record<string, boolean>;
  quiet_hours_start?: string;
  quiet_hours_end?: string;
  min_priority?: NotificationPriority;
  created_at: string;
  updated_at: string;
}

export interface NotificationStats {
  total_notifications: number;
  unread_count: number;
  read_count: number;
  archived_count: number;
  notifications_by_type: Record<string, number>;
  notifications_by_priority: Record<NotificationPriority, number>;
  recent_notifications: number;
}

// WebSocket Types
export interface WebSocketMessage {
  type: 'notification' | 'system' | 'heartbeat';
  notification?: Notification;
  data?: unknown;
}

// Auth Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

// Pagination Types
export interface PaginationParams {
  page?: number;
  limit?: number;
  sort?: string;
  order?: 'asc' | 'desc';
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

// API Response Types
export interface ApiError {
  error: string;
  message: string;
  status_code: number;
}

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: ApiError;
}
