package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sso/models"
)

type CompanyManagementRepository struct {
	db *sql.DB
}

func NewCompanyManagementRepository(db *sql.DB) *CompanyManagementRepository {
	return &CompanyManagementRepository{db: db}
}

// ListCompanies retrieves companies with filtering, sorting, and pagination
func (r *CompanyManagementRepository) ListCompanies(filter models.CompanyListFilter) ([]models.CompanyDetail, int64, error) {
	// Build WHERE clause
	whereClauses := []string{"c.deleted_at IS NULL"}
	args := []interface{}{}
	argCount := 1

	if filter.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(c.name ILIKE $%d OR c.domain ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.status = $%d", argCount))
		args = append(args, filter.Status)
		argCount++
	}

	if filter.Industry != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.industry = $%d", argCount))
		args = append(args, filter.Industry)
		argCount++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT c.id)
		FROM companies c
		WHERE %s
	`, whereSQL)

	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	sortBy := "c.created_at"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "name":
			sortBy = "c.name"
		case "user_count":
			sortBy = "user_count"
		case "created_at":
			sortBy = "c.created_at"
		}
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	// Query companies with user count
	query := fmt.Sprintf(`
		SELECT 
			c.id,
			c.name,
			c.domain,
			c.industry,
			c.description,
			c.logo_url,
			c.website,
			c.phone,
			c.address,
			c.status,
			c.settings,
			c.metadata,
			c.created_at,
			c.updated_at,
			c.deleted_at,
			COALESCE(COUNT(DISTINCT uc.user_id), 0) as user_count
		FROM companies c
		LEFT JOIN user_companies uc ON c.id = uc.company_id AND uc.deleted_at IS NULL
		WHERE %s
		GROUP BY c.id
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, sortBy, sortOrder, argCount, argCount+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var companies []models.CompanyDetail
	for rows.Next() {
		var c models.CompanyDetail
		var settingsJSON, metadataJSON []byte

		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Domain,
			&c.Industry,
			&c.Description,
			&c.LogoURL,
			&c.Website,
			&c.Phone,
			&c.Address,
			&c.Status,
			&settingsJSON,
			&metadataJSON,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
			&c.UserCount,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(settingsJSON) > 0 {
			json.Unmarshal(settingsJSON, &c.Settings)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &c.Metadata)
		}

		companies = append(companies, c)
	}

	return companies, total, nil
}

// GetCompanyByID retrieves a company by ID
func (r *CompanyManagementRepository) GetCompanyByID(companyID string) (*models.CompanyDetail, error) {
	query := `
		SELECT 
			c.id,
			c.name,
			c.domain,
			c.industry,
			c.description,
			c.logo_url,
			c.website,
			c.phone,
			c.address,
			c.status,
			c.settings,
			c.metadata,
			c.created_at,
			c.updated_at,
			c.deleted_at,
			COALESCE(COUNT(DISTINCT uc.user_id), 0) as user_count
		FROM companies c
		LEFT JOIN user_companies uc ON c.id = uc.company_id AND uc.deleted_at IS NULL
		WHERE c.id = $1 AND c.deleted_at IS NULL
		GROUP BY c.id
	`

	var c models.CompanyDetail
	var settingsJSON, metadataJSON []byte

	err := r.db.QueryRow(query, companyID).Scan(
		&c.ID,
		&c.Name,
		&c.Domain,
		&c.Industry,
		&c.Description,
		&c.LogoURL,
		&c.Website,
		&c.Phone,
		&c.Address,
		&c.Status,
		&settingsJSON,
		&metadataJSON,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
		&c.UserCount,
	)
	if err != nil {
		return nil, err
	}

	if len(settingsJSON) > 0 {
		json.Unmarshal(settingsJSON, &c.Settings)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &c.Metadata)
	}

	return &c, nil
}

// CreateCompany creates a new company
func (r *CompanyManagementRepository) CreateCompany(req models.CompanyCreateRequest) (*models.CompanyDetail, error) {
	settingsJSON, _ := json.Marshal(req.Settings)
	metadataJSON, _ := json.Marshal(req.Metadata)

	query := `
		INSERT INTO companies (
			name, domain, industry, description, logo_url, website, phone, address,
			status, settings, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	var company models.CompanyDetail
	company.Name = req.Name
	company.Domain = req.Domain
	company.Industry = req.Industry
	company.Description = req.Description
	company.LogoURL = req.LogoURL
	company.Website = req.Website
	company.Phone = req.Phone
	company.Address = req.Address
	company.Status = "active"
	company.Settings = req.Settings
	company.Metadata = req.Metadata

	err := r.db.QueryRow(
		query,
		req.Name, req.Domain, req.Industry, req.Description,
		req.LogoURL, req.Website, req.Phone, req.Address,
		"active", settingsJSON, metadataJSON,
	).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &company, nil
}

// UpdateCompany updates a company
func (r *CompanyManagementRepository) UpdateCompany(companyID string, req models.CompanyUpdateRequest) error {
	setClauses := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *req.Name)
		argCount++
	}

	if req.Domain != nil {
		setClauses = append(setClauses, fmt.Sprintf("domain = $%d", argCount))
		args = append(args, *req.Domain)
		argCount++
	}

	if req.Industry != nil {
		setClauses = append(setClauses, fmt.Sprintf("industry = $%d", argCount))
		args = append(args, *req.Industry)
		argCount++
	}

	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *req.Description)
		argCount++
	}

	if req.LogoURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("logo_url = $%d", argCount))
		args = append(args, *req.LogoURL)
		argCount++
	}

	if req.Website != nil {
		setClauses = append(setClauses, fmt.Sprintf("website = $%d", argCount))
		args = append(args, *req.Website)
		argCount++
	}

	if req.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argCount))
		args = append(args, *req.Phone)
		argCount++
	}

	if req.Address != nil {
		setClauses = append(setClauses, fmt.Sprintf("address = $%d", argCount))
		args = append(args, *req.Address)
		argCount++
	}

	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *req.Status)
		argCount++
	}

	if req.Settings != nil {
		settingsJSON, _ := json.Marshal(req.Settings)
		setClauses = append(setClauses, fmt.Sprintf("settings = $%d", argCount))
		args = append(args, settingsJSON)
		argCount++
	}

	if req.Metadata != nil {
		metadataJSON, _ := json.Marshal(req.Metadata)
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argCount))
		args = append(args, metadataJSON)
		argCount++
	}

	if len(setClauses) == 0 {
		return nil // No updates
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++

	args = append(args, companyID)

	query := fmt.Sprintf(`
		UPDATE companies
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "), argCount)

	_, err := r.db.Exec(query, args...)
	return err
}

// DeleteCompany soft deletes a company
func (r *CompanyManagementRepository) DeleteCompany(companyID string) error {
	query := `
		UPDATE companies
		SET deleted_at = $1, status = 'inactive'
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, time.Now(), companyID)
	return err
}

// UpdateCompanyStatus updates company status
func (r *CompanyManagementRepository) UpdateCompanyStatus(companyID, status string) error {
	query := `
		UPDATE companies
		SET status = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, status, time.Now(), companyID)
	return err
}

// AddUserToCompany adds a user to a company
func (r *CompanyManagementRepository) AddUserToCompany(companyID, userID, role string) error {
	query := `
		INSERT INTO user_companies (user_id, company_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, company_id) 
		DO UPDATE SET role = $3, deleted_at = NULL, updated_at = NOW()
	`

	_, err := r.db.Exec(query, userID, companyID, role)
	return err
}

// RemoveUserFromCompany removes a user from a company
func (r *CompanyManagementRepository) RemoveUserFromCompany(companyID, userID string) error {
	query := `
		UPDATE user_companies
		SET deleted_at = $1
		WHERE company_id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, time.Now(), companyID, userID)
	return err
}

// UpdateUserRoleInCompany updates user's role in a company
func (r *CompanyManagementRepository) UpdateUserRoleInCompany(companyID, userID, role string) error {
	query := `
		UPDATE user_companies
		SET role = $1, updated_at = $2
		WHERE company_id = $3 AND user_id = $4 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, role, time.Now(), companyID, userID)
	return err
}

// GetCompanyUsers retrieves users of a company with pagination
func (r *CompanyManagementRepository) GetCompanyUsers(companyID string, page, pageSize int) ([]models.CompanyUserDetail, int64, error) {
	// Count total
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM user_companies uc
		WHERE uc.company_id = $1 AND uc.deleted_at IS NULL
	`
	err := r.db.QueryRow(countQuery, companyID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Query users
	query := `
		SELECT 
			u.id,
			u.email,
			u.name,
			uc.role,
			uc.created_at,
			uc.updated_at,
			u.last_login_at,
			u.status
		FROM user_companies uc
		JOIN users u ON uc.user_id = u.id
		WHERE uc.company_id = $1 AND uc.deleted_at IS NULL
		ORDER BY uc.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, companyID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.CompanyUserDetail
	for rows.Next() {
		var u models.CompanyUserDetail
		err := rows.Scan(
			&u.UserID,
			&u.Email,
			&u.Name,
			&u.Role,
			&u.JoinedAt,
			&u.UpdatedAt,
			&u.LastLogin,
			&u.Status,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

// GetCompanyStats retrieves company statistics
func (r *CompanyManagementRepository) GetCompanyStats() (*models.CompanyStats, error) {
	var stats models.CompanyStats

	// Total companies by status
	statusQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'active') as active,
			COUNT(*) FILTER (WHERE status = 'inactive') as inactive,
			COUNT(*) FILTER (WHERE status = 'suspended') as suspended
		FROM companies
		WHERE deleted_at IS NULL
	`

	err := r.db.QueryRow(statusQuery).Scan(
		&stats.TotalCompanies,
		&stats.ActiveCompanies,
		&stats.InactiveCompanies,
		&stats.SuspendedCompanies,
	)
	if err != nil {
		return nil, err
	}

	// Total users across all companies
	userQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM user_companies
		WHERE deleted_at IS NULL
	`
	err = r.db.QueryRow(userQuery).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	// Average users per company
	if stats.TotalCompanies > 0 {
		stats.AverageUsersPerCompany = float64(stats.TotalUsers) / float64(stats.TotalCompanies)
	}

	// Companies by industry
	stats.ByIndustry = make(map[string]int64)
	industryQuery := `
		SELECT industry, COUNT(*)
		FROM companies
		WHERE deleted_at IS NULL AND industry IS NOT NULL AND industry != ''
		GROUP BY industry
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`
	rows, err := r.db.Query(industryQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var industry string
		var count int64
		if err := rows.Scan(&industry, &count); err != nil {
			continue
		}
		stats.ByIndustry[industry] = count
	}

	// Recent companies
	recent, _, err := r.ListCompanies(models.CompanyListFilter{
		Page:      1,
		PageSize:  5,
		SortBy:    "created_at",
		SortOrder: "desc",
	})
	if err != nil {
		return nil, err
	}
	stats.RecentCompanies = recent

	return &stats, nil
}

// BulkUpdateCompanies performs bulk actions on companies
func (r *CompanyManagementRepository) BulkUpdateCompanies(action string, companyIDs []string) (int, int, map[string]string, error) {
	success := 0
	failed := 0
	errors := make(map[string]string)

	for _, companyID := range companyIDs {
		var err error

		switch action {
		case "activate":
			err = r.UpdateCompanyStatus(companyID, "active")
		case "deactivate":
			err = r.UpdateCompanyStatus(companyID, "inactive")
		case "suspend":
			err = r.UpdateCompanyStatus(companyID, "suspended")
		case "delete":
			err = r.DeleteCompany(companyID)
		}

		if err != nil {
			failed++
			errors[companyID] = err.Error()
		} else {
			success++
		}
	}

	return success, failed, errors, nil
}

// GetUserRoleInCompany retrieves user's role in a company
func (r *CompanyManagementRepository) GetUserRoleInCompany(userID, companyID string) (string, error) {
	var role string
	query := `
		SELECT role
		FROM user_companies
		WHERE user_id = $1 AND company_id = $2 AND deleted_at IS NULL
	`

	err := r.db.QueryRow(query, userID, companyID).Scan(&role)
	if err != nil {
		return "", err
	}

	return role, nil
}
