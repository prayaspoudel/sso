package services

import (
	"database/sql"
	"errors"

	"sso/models"
	"sso/repository"
)

type CompanyManagementService struct {
	repo     *repository.CompanyManagementRepository
	userRepo *repository.UserManagementRepository
}

func NewCompanyManagementService(db *sql.DB) *CompanyManagementService {
	return &CompanyManagementService{
		repo:     repository.NewCompanyManagementRepository(db),
		userRepo: repository.NewUserManagementRepository(db),
	}
}

// ListCompanies retrieves companies with filtering
func (s *CompanyManagementService) ListCompanies(filter models.CompanyListFilter, requesterID string) (*models.CompanyListResponse, error) {
	// Check if requester has permission
	// For now, allow all authenticated users

	companies, total, err := s.repo.ListCompanies(filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &models.CompanyListResponse{
		Companies: companies,
		Total:     total,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		TotalPage: totalPages,
	}, nil
}

// GetCompany retrieves a company by ID
func (s *CompanyManagementService) GetCompany(companyID, requesterID string) (*models.CompanyDetail, error) {
	// Check if requester has access to this company
	// For now, allow all authenticated users

	company, err := s.repo.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("company not found")
		}
		return nil, err
	}

	return company, nil
}

// CreateCompany creates a new company
func (s *CompanyManagementService) CreateCompany(req models.CompanyCreateRequest, creatorID string) (*models.CompanyDetail, error) {
	// Validate request
	if req.Name == "" {
		return nil, errors.New("company name is required")
	}

	// Create company
	company, err := s.repo.CreateCompany(req)
	if err != nil {
		return nil, err
	}

	// Add creator as owner
	err = s.repo.AddUserToCompany(company.ID, creatorID, "owner")
	if err != nil {
		// Rollback company creation
		s.repo.DeleteCompany(company.ID)
		return nil, err
	}

	// TODO: Log audit

	return company, nil
}

// UpdateCompany updates a company
func (s *CompanyManagementService) UpdateCompany(companyID string, req models.CompanyUpdateRequest, updaterID string) error {
	// Check if company exists
	_, err := s.repo.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("company not found")
		}
		return err
	}

	// Check if updater has permission (must be owner or admin)
	role, err := s.repo.GetUserRoleInCompany(updaterID, companyID)
	if err != nil || (role != "owner" && role != "admin") {
		return errors.New("insufficient permissions to update company")
	}

	// Update company
	err = s.repo.UpdateCompany(companyID, req)
	if err != nil {
		return err
	}

	// TODO: Log audit

	return nil
}

// DeleteCompany soft deletes a company
func (s *CompanyManagementService) DeleteCompany(companyID, deleterID string) error {
	// Check if company exists
	_, err := s.repo.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("company not found")
		}
		return err
	}

	// Check if deleter has permission (must be owner)
	role, err := s.repo.GetUserRoleInCompany(deleterID, companyID)
	if err != nil || role != "owner" {
		return errors.New("only company owner can delete company")
	}

	// Delete company
	err = s.repo.DeleteCompany(companyID)
	if err != nil {
		return err
	}

	// TODO: Log audit

	return nil
}

// UpdateCompanyStatus updates company status
func (s *CompanyManagementService) UpdateCompanyStatus(companyID string, req models.CompanyStatusUpdateRequest, updaterID string) error {
	// Check if company exists
	_, err := s.repo.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("company not found")
		}
		return err
	}

	// Check if updater has permission (must be owner or admin)
	role, err := s.repo.GetUserRoleInCompany(updaterID, companyID)
	if err != nil || (role != "owner" && role != "admin") {
		return errors.New("insufficient permissions to update company status")
	}

	// Update status
	err = s.repo.UpdateCompanyStatus(companyID, req.Status)
	if err != nil {
		return err
	}

	// TODO: Log audit

	return nil
}

// AddUserToCompany adds a user to a company
func (s *CompanyManagementService) AddUserToCompany(companyID string, req models.UserCompanyAddRequest, adderID string) error {
	// Check if company exists
	_, err := s.repo.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("company not found")
		}
		return err
	}

	// TODO: Check if user exists when user management is integrated

	// Check if adder has permission (must be owner or admin)
	role, err := s.repo.GetUserRoleInCompany(adderID, companyID)
	if err != nil || (role != "owner" && role != "admin") {
		return errors.New("insufficient permissions to add users to company")
	}

	// Cannot add owner role unless you are owner
	if req.Role == "owner" && role != "owner" {
		return errors.New("only company owner can add other owners")
	}

	// Add user to company
	err = s.repo.AddUserToCompany(companyID, req.UserID, req.Role)
	if err != nil {
		return err
	}

	// TODO: Log audit

	return nil
}

// RemoveUserFromCompany removes a user from a company
func (s *CompanyManagementService) RemoveUserFromCompany(companyID, userID, removerID string) error {
	// Check if company exists
	_, err := s.repo.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("company not found")
		}
		return err
	}

	// Check if remover has permission (must be owner or admin)
	removerRole, err := s.repo.GetUserRoleInCompany(removerID, companyID)
	if err != nil || (removerRole != "owner" && removerRole != "admin") {
		return errors.New("insufficient permissions to remove users from company")
	}

	// Get target user's role
	targetRole, err := s.repo.GetUserRoleInCompany(userID, companyID)
	if err != nil {
		return errors.New("user is not a member of this company")
	}

	// Cannot remove owner unless you are owner
	if targetRole == "owner" && removerRole != "owner" {
		return errors.New("only company owner can remove other owners")
	}

	// Cannot remove yourself if you're the only owner
	if userID == removerID && targetRole == "owner" {
		// Check if there are other owners
		users, _, err := s.repo.GetCompanyUsers(companyID, 1, 100)
		if err != nil {
			return err
		}
		ownerCount := 0
		for _, u := range users {
			if u.Role == "owner" {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return errors.New("cannot remove the only owner of the company")
		}
	}

	// Remove user from company
	err = s.repo.RemoveUserFromCompany(companyID, userID)
	if err != nil {
		return err
	}

	// TODO: Log audit

	return nil
}

// UpdateUserRoleInCompany updates user's role in a company
func (s *CompanyManagementService) UpdateUserRoleInCompany(companyID, userID string, req models.UserCompanyUpdateRequest, updaterID string) error {
	// Check if company exists
	_, err := s.repo.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("company not found")
		}
		return err
	}

	// Check if updater has permission (must be owner or admin)
	updaterRole, err := s.repo.GetUserRoleInCompany(updaterID, companyID)
	if err != nil || (updaterRole != "owner" && updaterRole != "admin") {
		return errors.New("insufficient permissions to update user roles")
	}

	// Get target user's current role
	targetRole, err := s.repo.GetUserRoleInCompany(userID, companyID)
	if err != nil {
		return errors.New("user is not a member of this company")
	}

	// Cannot change owner role unless you are owner
	if (targetRole == "owner" || req.Role == "owner") && updaterRole != "owner" {
		return errors.New("only company owner can change owner role")
	}

	// Update user role
	err = s.repo.UpdateUserRoleInCompany(companyID, userID, req.Role)
	if err != nil {
		return err
	}

	// TODO: Log audit

	return nil
}

// GetCompanyUsers retrieves users of a company
func (s *CompanyManagementService) GetCompanyUsers(companyID, requesterID string, page, pageSize int) (*models.CompanyUsersResponse, error) {
	// Check if company exists
	_, err := s.repo.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("company not found")
		}
		return nil, err
	}

	// Check if requester has access to this company
	_, err = s.repo.GetUserRoleInCompany(requesterID, companyID)
	if err != nil {
		return nil, errors.New("you don't have access to this company")
	}

	users, total, err := s.repo.GetCompanyUsers(companyID, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.CompanyUsersResponse{
		Users:     users,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPages,
	}, nil
}

// GetCompanyStats retrieves company statistics
func (s *CompanyManagementService) GetCompanyStats(requesterID string) (*models.CompanyStats, error) {
	// For now, allow all authenticated users
	// In production, restrict to super_admin

	stats, err := s.repo.GetCompanyStats()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// BulkActionCompanies performs bulk actions on companies
func (s *CompanyManagementService) BulkActionCompanies(req models.CompanyBulkActionRequest, requesterID string) (*models.CompanyBulkActionResponse, error) {
	// For now, allow all authenticated users
	// In production, restrict based on permissions

	var exportData []models.CompanyDetail

	// For export action, retrieve company data
	if req.Action == "export" {
		for _, companyID := range req.CompanyIDs {
			company, err := s.repo.GetCompanyByID(companyID)
			if err == nil {
				exportData = append(exportData, *company)
			}
		}

		return &models.CompanyBulkActionResponse{
			Success:    len(exportData),
			Failed:     len(req.CompanyIDs) - len(exportData),
			Total:      len(req.CompanyIDs),
			ExportData: exportData,
		}, nil
	}

	// Perform bulk action
	success, failed, errors, err := s.repo.BulkUpdateCompanies(req.Action, req.CompanyIDs)
	if err != nil {
		return nil, err
	}

	// TODO: Log audit

	return &models.CompanyBulkActionResponse{
		Success: success,
		Failed:  failed,
		Total:   len(req.CompanyIDs),
		Errors:  errors,
	}, nil
}
