package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// User notifications
	NotificationTypeUserCreated      NotificationType = "user.created"
	NotificationTypeUserUpdated      NotificationType = "user.updated"
	NotificationTypeUserDeleted      NotificationType = "user.deleted"
	NotificationTypeUserStatusChange NotificationType = "user.status_changed"
	NotificationTypeUserLogin        NotificationType = "user.login"
	NotificationTypeUserLogout       NotificationType = "user.logout"

	// Company notifications
	NotificationTypeCompanyCreated      NotificationType = "company.created"
	NotificationTypeCompanyUpdated      NotificationType = "company.updated"
	NotificationTypeCompanyDeleted      NotificationType = "company.deleted"
	NotificationTypeCompanyStatusChange NotificationType = "company.status_changed"
	NotificationTypeCompanyUserAdded    NotificationType = "company.user_added"
	NotificationTypeCompanyUserRemoved  NotificationType = "company.user_removed"

	// Role notifications
	NotificationTypeRoleCreated  NotificationType = "role.created"
	NotificationTypeRoleUpdated  NotificationType = "role.updated"
	NotificationTypeRoleDeleted  NotificationType = "role.deleted"
	NotificationTypeRoleAssigned NotificationType = "role.assigned"
	NotificationTypeRoleRevoked  NotificationType = "role.revoked"

	// Security notifications
	NotificationTypeSecurityPasswordChange NotificationType = "security.password_changed"
	NotificationTypeSecurityLoginFailed    NotificationType = "security.login_failed"
	NotificationTypeSecurityAccountLocked  NotificationType = "security.account_locked"
	NotificationTypeSecurity2FAEnabled     NotificationType = "security.2fa_enabled"
	NotificationTypeSecurity2FADisabled    NotificationType = "security.2fa_disabled"

	// Session notifications
	NotificationTypeSessionCreated NotificationType = "session.created"
	NotificationTypeSessionExpired NotificationType = "session.expired"
	NotificationTypeSessionRevoked NotificationType = "session.revoked"

	// System notifications
	NotificationTypeSystemAlert        NotificationType = "system.alert"
	NotificationTypeSystemMaintenance  NotificationType = "system.maintenance"
	NotificationTypeSystemUpdate       NotificationType = "system.update"
	NotificationTypeSystemBackupFailed NotificationType = "system.backup_failed"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusUnread   NotificationStatus = "unread"
	NotificationStatusRead     NotificationStatus = "read"
	NotificationStatusArchived NotificationStatus = "archived"
)

// Notification represents a notification in the system
type Notification struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty" db:"user_id"` // nil for broadcast
	Type       NotificationType       `json:"type" db:"type"`
	Title      string                 `json:"title" db:"title"`
	Message    string                 `json:"message" db:"message"`
	Priority   NotificationPriority   `json:"priority" db:"priority"`
	Status     NotificationStatus     `json:"status" db:"status"`
	Data       map[string]interface{} `json:"data,omitempty" db:"data"` // Additional data as JSONB
	ActionURL  string                 `json:"action_url,omitempty" db:"action_url"`
	ActionText string                 `json:"action_text,omitempty" db:"action_text"`
	ReadAt     *time.Time             `json:"read_at,omitempty" db:"read_at"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
}

// NotificationCreateRequest represents a request to create a notification
type NotificationCreateRequest struct {
	UserID     *uuid.UUID             `json:"user_id,omitempty"` // nil for broadcast
	Type       NotificationType       `json:"type" binding:"required"`
	Title      string                 `json:"title" binding:"required,min=1,max=200"`
	Message    string                 `json:"message" binding:"required,min=1,max=1000"`
	Priority   NotificationPriority   `json:"priority"`
	Data       map[string]interface{} `json:"data,omitempty"`
	ActionURL  string                 `json:"action_url,omitempty"`
	ActionText string                 `json:"action_text,omitempty"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
}

// NotificationFilter represents filters for listing notifications
type NotificationFilter struct {
	UserID    *uuid.UUID           `json:"user_id,omitempty" form:"user_id"`
	Type      NotificationType     `json:"type,omitempty" form:"type"`
	Status    NotificationStatus   `json:"status,omitempty" form:"status"`
	Priority  NotificationPriority `json:"priority,omitempty" form:"priority"`
	StartDate *time.Time           `json:"start_date,omitempty" form:"start_date"`
	EndDate   *time.Time           `json:"end_date,omitempty" form:"end_date"`
	Search    string               `json:"search,omitempty" form:"search"`
	SortBy    string               `json:"sort_by,omitempty" form:"sort_by"`       // created_at, priority
	SortOrder string               `json:"sort_order,omitempty" form:"sort_order"` // asc, desc
	Page      int                  `json:"page,omitempty" form:"page"`
	PageSize  int                  `json:"page_size,omitempty" form:"page_size"`
}

// NotificationListResponse represents a paginated list of notifications
type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int64          `json:"total"`
	Page          int            `json:"page"`
	PageSize      int            `json:"page_size"`
	TotalPages    int            `json:"total_pages"`
	UnreadCount   int64          `json:"unread_count"`
}

// NotificationMarkReadRequest represents a request to mark notifications as read
type NotificationMarkReadRequest struct {
	NotificationIDs []uuid.UUID `json:"notification_ids" binding:"required,min=1"`
}

// NotificationMarkAllReadRequest represents a request to mark all notifications as read
type NotificationMarkAllReadRequest struct {
	UserID *uuid.UUID `json:"user_id" binding:"required"`
}

// NotificationDeleteRequest represents a request to delete notifications
type NotificationDeleteRequest struct {
	NotificationIDs []uuid.UUID `json:"notification_ids" binding:"required,min=1"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalNotifications     int64                          `json:"total_notifications"`
	UnreadNotifications    int64                          `json:"unread_notifications"`
	ReadNotifications      int64                          `json:"read_notifications"`
	ArchivedNotifications  int64                          `json:"archived_notifications"`
	NotificationsToday     int64                          `json:"notifications_today"`
	NotificationsThisWeek  int64                          `json:"notifications_this_week"`
	NotificationsThisMonth int64                          `json:"notifications_this_month"`
	ByType                 map[NotificationType]int64     `json:"by_type"`
	ByPriority             map[NotificationPriority]int64 `json:"by_priority"`
	ByStatus               map[NotificationStatus]int64   `json:"by_status"`
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	ID               uuid.UUID            `json:"id" db:"id"`
	UserID           uuid.UUID            `json:"user_id" db:"user_id"`
	EmailEnabled     bool                 `json:"email_enabled" db:"email_enabled"`
	SMSEnabled       bool                 `json:"sms_enabled" db:"sms_enabled"`
	PushEnabled      bool                 `json:"push_enabled" db:"push_enabled"`
	WebSocketEnabled bool                 `json:"websocket_enabled" db:"websocket_enabled"`
	EnabledTypes     []NotificationType   `json:"enabled_types" db:"enabled_types"` // Array of enabled notification types
	MinPriority      NotificationPriority `json:"min_priority" db:"min_priority"`   // Minimum priority to receive
	QuietHoursStart  *time.Time           `json:"quiet_hours_start,omitempty" db:"quiet_hours_start"`
	QuietHoursEnd    *time.Time           `json:"quiet_hours_end,omitempty" db:"quiet_hours_end"`
	CreatedAt        time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at" db:"updated_at"`
}

// NotificationPreferenceUpdateRequest represents a request to update notification preferences
type NotificationPreferenceUpdateRequest struct {
	EmailEnabled     *bool                `json:"email_enabled,omitempty"`
	SMSEnabled       *bool                `json:"sms_enabled,omitempty"`
	PushEnabled      *bool                `json:"push_enabled,omitempty"`
	WebSocketEnabled *bool                `json:"websocket_enabled,omitempty"`
	EnabledTypes     []NotificationType   `json:"enabled_types,omitempty"`
	MinPriority      NotificationPriority `json:"min_priority,omitempty"`
	QuietHoursStart  *time.Time           `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd    *time.Time           `json:"quiet_hours_end,omitempty"`
}

// NotificationBroadcastRequest represents a request to broadcast a notification
type NotificationBroadcastRequest struct {
	Type            NotificationType       `json:"type" binding:"required"`
	Title           string                 `json:"title" binding:"required,min=1,max=200"`
	Message         string                 `json:"message" binding:"required,min=1,max=1000"`
	Priority        NotificationPriority   `json:"priority"`
	Data            map[string]interface{} `json:"data,omitempty"`
	ActionURL       string                 `json:"action_url,omitempty"`
	ActionText      string                 `json:"action_text,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	TargetRoles     []string               `json:"target_roles,omitempty"`     // Optional: specific roles only
	TargetCompanies []uuid.UUID            `json:"target_companies,omitempty"` // Optional: specific companies only
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type         string                 `json:"type"` // notification, ping, pong, error, auth
	Notification *Notification          `json:"notification,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// WebSocketAuthMessage represents an authentication message for WebSocket
type WebSocketAuthMessage struct {
	Token string `json:"token" binding:"required"`
}

// WebSocketConnectionInfo represents information about a WebSocket connection
type WebSocketConnectionInfo struct {
	UserID        uuid.UUID `json:"user_id"`
	UserEmail     string    `json:"user_email"`
	ConnectedAt   time.Time `json:"connected_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	IPAddress     string    `json:"ip_address"`
}

// NotificationDeliveryLog represents a log of notification delivery
type NotificationDeliveryLog struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	NotificationID uuid.UUID  `json:"notification_id" db:"notification_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Channel        string     `json:"channel" db:"channel"` // email, sms, push, websocket
	Status         string     `json:"status" db:"status"`   // pending, sent, failed, delivered
	Error          string     `json:"error,omitempty" db:"error"`
	SentAt         *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}
