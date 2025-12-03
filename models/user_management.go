package models

import (
	"time"

	"github.com/google/uuid"
)

// UserListFilter represents filters for listing users
type UserListFilter struct {
	Search      string     // Search in name, email
	CompanyID   *uuid.UUID // Filter by company
	Role        *string    // Filter by role
	Status      *string    // active, inactive, suspended
	IsVerified  *bool      // Filter by email verification
	Has2FA      *bool      // Filter by 2FA enabled
	CreatedFrom *time.Time // Created after
	CreatedTo   *time.Time // Created before
	Page        int        // Page number (1-indexed)
	PageSize    int        // Items per page
	SortBy      string     // Field to sort by
	SortOrder   string     // asc or desc
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Users      []UserDetail `json:"users"`
	TotalCount int          `json:"total_count"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// UserDetail represents detailed user information
type UserDetail struct {
	ID                 uuid.UUID  `json:"id"`
	Email              string     `json:"email"`
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	CompanyID          uuid.UUID  `json:"company_id"`
	CompanyName        string     `json:"company_name,omitempty"`
	Role               string     `json:"role"`
	IsActive           bool       `json:"is_active"`
	EmailVerified      bool       `json:"email_verified"`
	TwoFactorEnabled   bool       `json:"two_factor_enabled"`
	LastLoginAt        *time.Time `json:"last_login_at"`
	LastLoginIP        string     `json:"last_login_ip,omitempty"`
	FailedLoginCount   int        `json:"failed_login_count"`
	AccountLockedAt    *time.Time `json:"account_locked_at"`
	SocialAccountCount int        `json:"social_account_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// UserCreateRequest represents request to create a new user
type UserCreateRequest struct {
	Email            string    `json:"email" binding:"required,email"`
	Password         string    `json:"password" binding:"required,min=8"`
	FirstName        string    `json:"first_name" binding:"required"`
	LastName         string    `json:"last_name" binding:"required"`
	CompanyID        uuid.UUID `json:"company_id" binding:"required"`
	Role             string    `json:"role" binding:"required"`
	IsActive         bool      `json:"is_active"`
	SendWelcomeEmail bool      `json:"send_welcome_email"`
}

// UserUpdateRequest represents request to update user
type UserUpdateRequest struct {
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	CompanyID *uuid.UUID `json:"company_id"`
	Role      *string    `json:"role"`
	IsActive  *bool      `json:"is_active"`
}

// UserProfileUpdateRequest represents request to update user profile (by user themselves)
type UserProfileUpdateRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	Timezone  string `json:"timezone"`
	Language  string `json:"language"`
	Avatar    string `json:"avatar"`
}

// UserPasswordChangeRequest represents password change request
type UserPasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UserStatusUpdateRequest represents request to update user status
type UserStatusUpdateRequest struct {
	IsActive bool   `json:"is_active" binding:"required"`
	Reason   string `json:"reason"`
}

// UserBulkActionRequest represents bulk action on users
type UserBulkActionRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" binding:"required,min=1"`
	Action  string      `json:"action" binding:"required"` // activate, deactivate, delete, unlock
	Reason  string      `json:"reason"`
}

// UserActivity represents user activity log
type UserActivity struct {
	ID          uuid.UUID         `json:"id"`
	UserID      uuid.UUID         `json:"user_id"`
	Action      string            `json:"action"`
	Description string            `json:"description"`
	IPAddress   string            `json:"ip_address"`
	UserAgent   string            `json:"user_agent"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers        int            `json:"total_users"`
	ActiveUsers       int            `json:"active_users"`
	InactiveUsers     int            `json:"inactive_users"`
	VerifiedUsers     int            `json:"verified_users"`
	Users2FA          int            `json:"users_2fa"`
	NewUsersToday     int            `json:"new_users_today"`
	NewUsersThisWeek  int            `json:"new_users_this_week"`
	NewUsersThisMonth int            `json:"new_users_this_month"`
	LoginsTodayCount  int            `json:"logins_today_count"`
	UsersByRole       map[string]int `json:"users_by_role"`
	UsersByCompany    map[string]int `json:"users_by_company"`
}

// UserExportFormat represents export format for users
type UserExportFormat string

const (
	UserExportFormatCSV  UserExportFormat = "csv"
	UserExportFormatJSON UserExportFormat = "json"
	UserExportFormatXLSX UserExportFormat = "xlsx"
)

// UserImportRequest represents bulk user import request
type UserImportRequest struct {
	Format      string            `json:"format" binding:"required"` // csv, json
	Data        string            `json:"data" binding:"required"`
	CompanyID   uuid.UUID         `json:"company_id" binding:"required"`
	DefaultRole string            `json:"default_role"`
	SendEmails  bool              `json:"send_emails"`
	Options     map[string]string `json:"options,omitempty"`
}

// UserImportResult represents result of bulk import
type UserImportResult struct {
	TotalRows    int           `json:"total_rows"`
	SuccessCount int           `json:"success_count"`
	FailureCount int           `json:"failure_count"`
	SkippedCount int           `json:"skipped_count"`
	Errors       []ImportError `json:"errors,omitempty"`
	CreatedUsers []UserDetail  `json:"created_users,omitempty"`
	Duration     string        `json:"duration"`
}

// ImportError represents an import error
type ImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}
