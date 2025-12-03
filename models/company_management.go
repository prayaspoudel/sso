package models

import (
	"time"
)

// CompanyListFilter represents filters for listing companies
type CompanyListFilter struct {
	Search       string `form:"search" json:"search"`               // Search by name or domain
	Status       string `form:"status" json:"status"`               // active, inactive, suspended
	Industry     string `form:"industry" json:"industry"`           // Filter by industry
	MinEmployees int    `form:"min_employees" json:"min_employees"` // Minimum employee count
	MaxEmployees int    `form:"max_employees" json:"max_employees"` // Maximum employee count
	SortBy       string `form:"sort_by" json:"sort_by"`             // name, created_at, user_count
	SortOrder    string `form:"sort_order" json:"sort_order"`       // asc, desc
	Page         int    `form:"page" json:"page"`                   // Page number (1-based)
	PageSize     int    `form:"page_size" json:"page_size"`         // Items per page
}

// CompanyListResponse represents paginated company list response
type CompanyListResponse struct {
	Companies []CompanyDetail `json:"companies"`
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"page_size"`
	TotalPage int             `json:"total_pages"`
}

// CompanyDetail represents detailed company information
type CompanyDetail struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Domain      string                 `json:"domain,omitempty"`
	Industry    string                 `json:"industry,omitempty"`
	Description string                 `json:"description,omitempty"`
	LogoURL     string                 `json:"logo_url,omitempty"`
	Website     string                 `json:"website,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Address     string                 `json:"address,omitempty"`
	Status      string                 `json:"status"`
	UserCount   int                    `json:"user_count"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty"`
}

// CompanyCreateRequest represents request to create a new company
type CompanyCreateRequest struct {
	Name        string                 `json:"name" binding:"required,min=2,max=100"`
	Domain      string                 `json:"domain" binding:"omitempty,max=100"`
	Industry    string                 `json:"industry" binding:"omitempty,max=50"`
	Description string                 `json:"description" binding:"omitempty,max=500"`
	LogoURL     string                 `json:"logo_url" binding:"omitempty,url,max=255"`
	Website     string                 `json:"website" binding:"omitempty,url,max=255"`
	Phone       string                 `json:"phone" binding:"omitempty,max=20"`
	Address     string                 `json:"address" binding:"omitempty,max=255"`
	Settings    map[string]interface{} `json:"settings"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CompanyUpdateRequest represents request to update a company
type CompanyUpdateRequest struct {
	Name        *string                `json:"name" binding:"omitempty,min=2,max=100"`
	Domain      *string                `json:"domain" binding:"omitempty,max=100"`
	Industry    *string                `json:"industry" binding:"omitempty,max=50"`
	Description *string                `json:"description" binding:"omitempty,max=500"`
	LogoURL     *string                `json:"logo_url" binding:"omitempty,url,max=255"`
	Website     *string                `json:"website" binding:"omitempty,url,max=255"`
	Phone       *string                `json:"phone" binding:"omitempty,max=20"`
	Address     *string                `json:"address" binding:"omitempty,max=255"`
	Status      *string                `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	Settings    map[string]interface{} `json:"settings"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CompanyStatusUpdateRequest represents request to update company status
type CompanyStatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive suspended"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

// UserCompanyAddRequest represents request to add a user to a company
type UserCompanyAddRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Role   string `json:"role" binding:"required,oneof=owner admin member viewer"`
}

// UserCompanyUpdateRequest represents request to update user role in company
type UserCompanyUpdateRequest struct {
	Role string `json:"role" binding:"required,oneof=owner admin member viewer"`
}

// CompanyUserDetail represents user information within a company
type CompanyUserDetail struct {
	UserID    string     `json:"user_id"`
	Email     string     `json:"email"`
	Name      string     `json:"name,omitempty"`
	Role      string     `json:"role"`
	JoinedAt  time.Time  `json:"joined_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	Status    string     `json:"status"`
}

// CompanyUsersResponse represents paginated company users response
type CompanyUsersResponse struct {
	Users     []CompanyUserDetail `json:"users"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
	TotalPage int                 `json:"total_pages"`
}

// CompanyStats represents company statistics
type CompanyStats struct {
	TotalCompanies         int64            `json:"total_companies"`
	ActiveCompanies        int64            `json:"active_companies"`
	InactiveCompanies      int64            `json:"inactive_companies"`
	SuspendedCompanies     int64            `json:"suspended_companies"`
	TotalUsers             int64            `json:"total_users"`
	AverageUsersPerCompany float64          `json:"average_users_per_company"`
	ByIndustry             map[string]int64 `json:"by_industry"`
	RecentCompanies        []CompanyDetail  `json:"recent_companies"`
}

// CompanyBulkActionRequest represents bulk actions on companies
type CompanyBulkActionRequest struct {
	Action     string   `json:"action" binding:"required,oneof=activate deactivate suspend delete export"`
	CompanyIDs []string `json:"company_ids" binding:"required,min=1,dive,uuid"`
	Reason     string   `json:"reason" binding:"omitempty,max=500"` // For status changes
}

// CompanyBulkActionResponse represents result of bulk action
type CompanyBulkActionResponse struct {
	Success    int               `json:"success"`
	Failed     int               `json:"failed"`
	Total      int               `json:"total"`
	Errors     map[string]string `json:"errors,omitempty"`      // company_id -> error message
	ExportData []CompanyDetail   `json:"export_data,omitempty"` // For export action
}

// CompanyInviteRequest represents request to invite user to company
type CompanyInviteRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Role      string `json:"role" binding:"required,oneof=admin member viewer"`
	Message   string `json:"message" binding:"omitempty,max=500"`
	ExpiresIn int    `json:"expires_in" binding:"omitempty,min=1,max=30"` // Days, default 7
}

// CompanyInviteResponse represents company invitation
type CompanyInviteResponse struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	InviteURL string    `json:"invite_url"`
}

// CompanyTransferRequest represents request to transfer company ownership
type CompanyTransferRequest struct {
	NewOwnerID string `json:"new_owner_id" binding:"required,uuid"`
	Password   string `json:"password" binding:"required"` // Require password confirmation
}

// CompanySettingsUpdateRequest represents request to update company settings
type CompanySettingsUpdateRequest struct {
	Settings map[string]interface{} `json:"settings" binding:"required"`
}

// CompanyMetadataUpdateRequest represents request to update company metadata
type CompanyMetadataUpdateRequest struct {
	Metadata map[string]interface{} `json:"metadata" binding:"required"`
}
