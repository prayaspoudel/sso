package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogFilter represents filters for audit log search
type AuditLogFilter struct {
	UserID     string    `form:"user_id" json:"user_id"`       // Filter by user ID
	Action     string    `form:"action" json:"action"`         // Filter by action (e.g., "user.create", "company.update")
	Resource   string    `form:"resource" json:"resource"`     // Filter by resource type (user, company, role, etc.)
	IPAddress  string    `form:"ip_address" json:"ip_address"` // Filter by IP address
	StartDate  time.Time `form:"start_date" json:"start_date"` // Filter from date
	EndDate    time.Time `form:"end_date" json:"end_date"`     // Filter to date
	SearchTerm string    `form:"search" json:"search"`         // Search in details
	SortBy     string    `form:"sort_by" json:"sort_by"`       // Sort by field (created_at, action)
	SortOrder  string    `form:"sort_order" json:"sort_order"` // Sort order (asc, desc)
	Page       int       `form:"page" json:"page"`             // Page number (1-based)
	PageSize   int       `form:"page_size" json:"page_size"`   // Items per page
}

// AuditLogListResponse represents paginated audit log list response
type AuditLogListResponse struct {
	Logs      []AuditLogDetail `json:"logs"`
	Total     int64            `json:"total"`
	Page      int              `json:"page"`
	PageSize  int              `json:"page_size"`
	TotalPage int              `json:"total_pages"`
}

// AuditLogDetail represents detailed audit log information
type AuditLogDetail struct {
	ID        uuid.UUID              `json:"id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	UserEmail string                 `json:"user_email,omitempty"` // Joined from users table
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// AuditLogCreateRequest represents request to create an audit log
type AuditLogCreateRequest struct {
	UserID    *uuid.UUID             `json:"user_id"`
	Action    string                 `json:"action" binding:"required"`
	Resource  string                 `json:"resource" binding:"required"`
	Details   map[string]interface{} `json:"details"`
	IPAddress string                 `json:"ip_address"`
}

// AuditLogStats represents audit log statistics
type AuditLogStats struct {
	TotalLogs      int64            `json:"total_logs"`
	LogsToday      int64            `json:"logs_today"`
	LogsThisWeek   int64            `json:"logs_this_week"`
	LogsThisMonth  int64            `json:"logs_this_month"`
	TopActions     map[string]int64 `json:"top_actions"`      // Top 10 actions
	TopResources   map[string]int64 `json:"top_resources"`    // Top 10 resources
	TopUsers       []AuditUserStat  `json:"top_users"`        // Top 10 users by activity
	ActivityByHour map[string]int64 `json:"activity_by_hour"` // Activity distribution by hour
	ActivityByDay  map[string]int64 `json:"activity_by_day"`  // Last 7 days activity
}

// AuditUserStat represents user activity statistics
type AuditUserStat struct {
	UserID    uuid.UUID `json:"user_id"`
	UserEmail string    `json:"user_email"`
	LogCount  int64     `json:"log_count"`
}

// AuditLogExportRequest represents request to export audit logs
type AuditLogExportRequest struct {
	Format     string         `json:"format" binding:"required,oneof=csv json"`
	Filter     AuditLogFilter `json:"filter"`
	MaxRecords int            `json:"max_records"` // Limit number of records (default 10000)
}

// AuditLogExportResponse represents audit log export response
type AuditLogExportResponse struct {
	FileName    string    `json:"file_name"`
	FileURL     string    `json:"file_url,omitempty"`
	FileData    string    `json:"file_data,omitempty"` // Base64 encoded for small files
	RecordCount int       `json:"record_count"`
	Format      string    `json:"format"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuditLogRetentionPolicy represents retention policy configuration
type AuditLogRetentionPolicy struct {
	ID              uuid.UUID `json:"id"`
	Resource        string    `json:"resource"`         // Resource type or "*" for all
	RetentionDays   int       `json:"retention_days"`   // Days to keep logs
	ArchiveEnabled  bool      `json:"archive_enabled"`  // Archive before deletion
	ArchiveLocation string    `json:"archive_location"` // S3 bucket, file path, etc.
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AuditLogRetentionPolicyRequest represents request to create/update retention policy
type AuditLogRetentionPolicyRequest struct {
	Resource        string `json:"resource" binding:"required"`
	RetentionDays   int    `json:"retention_days" binding:"required,min=1,max=3650"`
	ArchiveEnabled  bool   `json:"archive_enabled"`
	ArchiveLocation string `json:"archive_location"`
	Enabled         bool   `json:"enabled"`
}

// AuditLogCleanupRequest represents request to cleanup old audit logs
type AuditLogCleanupRequest struct {
	Resource  string    `json:"resource"`   // Specific resource or empty for all
	OlderThan time.Time `json:"older_than"` // Delete logs older than this date
	Archive   bool      `json:"archive"`    // Archive before deletion
	DryRun    bool      `json:"dry_run"`    // Don't actually delete, just count
}

// AuditLogCleanupResponse represents audit log cleanup response
type AuditLogCleanupResponse struct {
	DeletedCount  int64     `json:"deleted_count"`
	ArchivedCount int64     `json:"archived_count,omitempty"`
	DryRun        bool      `json:"dry_run"`
	ExecutedAt    time.Time `json:"executed_at"`
}

// AuditLogTimelineRequest represents request for audit timeline
type AuditLogTimelineRequest struct {
	UserID     *uuid.UUID `json:"user_id"`
	Resource   string     `json:"resource"`
	ResourceID string     `json:"resource_id"` // Specific resource instance
	StartDate  time.Time  `json:"start_date"`
	EndDate    time.Time  `json:"end_date"`
	Limit      int        `json:"limit"` // Max number of events
}

// AuditLogTimelineResponse represents audit timeline response
type AuditLogTimelineResponse struct {
	Events    []AuditLogDetail `json:"events"`
	Total     int64            `json:"total"`
	StartDate time.Time        `json:"start_date"`
	EndDate   time.Time        `json:"end_date"`
}

// AuditLogDiff represents changes between two audit log entries
type AuditLogDiff struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// AuditLogCompareResponse represents comparison between audit logs
type AuditLogCompareResponse struct {
	BeforeLog AuditLogDetail `json:"before_log"`
	AfterLog  AuditLogDetail `json:"after_log"`
	Changes   []AuditLogDiff `json:"changes"`
}

// AuditActionType constants
const (
	// User actions
	AuditActionUserCreate         = "user.create"
	AuditActionUserUpdate         = "user.update"
	AuditActionUserDelete         = "user.delete"
	AuditActionUserLogin          = "user.login"
	AuditActionUserLogout         = "user.logout"
	AuditActionUserPasswordChange = "user.password_change"
	AuditActionUserStatusChange   = "user.status_change"

	// Company actions
	AuditActionCompanyCreate       = "company.create"
	AuditActionCompanyUpdate       = "company.update"
	AuditActionCompanyDelete       = "company.delete"
	AuditActionCompanyStatusChange = "company.status_change"
	AuditActionCompanyAddUser      = "company.add_user"
	AuditActionCompanyRemoveUser   = "company.remove_user"
	AuditActionCompanyUpdateRole   = "company.update_role"

	// Role actions
	AuditActionRoleCreate = "role.create"
	AuditActionRoleUpdate = "role.update"
	AuditActionRoleDelete = "role.delete"
	AuditActionRoleAssign = "role.assign"
	AuditActionRoleRevoke = "role.revoke"

	// Permission actions
	AuditActionPermissionCreate = "permission.create"
	AuditActionPermissionUpdate = "permission.update"
	AuditActionPermissionDelete = "permission.delete"
	AuditActionPermissionAssign = "permission.assign"
	AuditActionPermissionRevoke = "permission.revoke"

	// Authentication actions
	AuditActionAuthLogin         = "auth.login"
	AuditActionAuthLogout        = "auth.logout"
	AuditActionAuthFailed        = "auth.failed"
	AuditActionAuth2FAEnable     = "auth.2fa_enable"
	AuditActionAuth2FADisable    = "auth.2fa_disable"
	AuditActionAuthPasswordReset = "auth.password_reset"

	// Session actions
	AuditActionSessionCreate = "session.create"
	AuditActionSessionRevoke = "session.revoke"
	AuditActionSessionExpire = "session.expire"

	// OAuth actions
	AuditActionOAuthAuthorize = "oauth.authorize"
	AuditActionOAuthToken     = "oauth.token"
	AuditActionOAuthRevoke    = "oauth.revoke"

	// System actions
	AuditActionSystemConfig  = "system.config"
	AuditActionSystemBackup  = "system.backup"
	AuditActionSystemRestore = "system.restore"
)

// AuditResourceType constants
const (
	AuditResourceUser       = "user"
	AuditResourceCompany    = "company"
	AuditResourceRole       = "role"
	AuditResourcePermission = "permission"
	AuditResourceSession    = "session"
	AuditResourceOAuth      = "oauth"
	AuditResourceSystem     = "system"
	AuditResourceAuth       = "auth"
)
