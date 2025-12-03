package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sso/models"
	"sso/repository"
	"sso/utils"

	"github.com/google/uuid"
)

const (
	// Account lockout settings
	MaxFailedAttempts = 5
	LockoutDuration   = 30 * time.Minute
	AttemptWindow     = 15 * time.Minute
)

type SecurityService struct {
	securityRepo *repository.SecurityRepository
	userRepo     *repository.UserRepository
}

func NewSecurityService(securityRepo *repository.SecurityRepository, userRepo *repository.UserRepository) *SecurityService {
	return &SecurityService{
		securityRepo: securityRepo,
		userRepo:     userRepo,
	}
}

// RecordLoginAttempt records a login attempt (successful or failed)
func (s *SecurityService) RecordLoginAttempt(ctx context.Context, email, ipAddress string, successful bool) error {
	attempt := &models.LoginAttempt{
		ID:         uuid.New(),
		Email:      email,
		IPAddress:  ipAddress,
		Successful: successful,
		CreatedAt:  time.Now(),
	}
	return s.securityRepo.RecordLoginAttempt(ctx, attempt)
}

// CheckAccountLockout checks if an account should be locked due to failed attempts
func (s *SecurityService) CheckAccountLockout(ctx context.Context, email, ipAddress string) error {
	// Get recent failed attempts
	failedAttempts, err := s.securityRepo.GetRecentFailedAttempts(ctx, email, AttemptWindow)
	if err != nil {
		return err
	}

	// If max attempts exceeded, lock the account
	if failedAttempts >= MaxFailedAttempts {
		// Get user
		user, err := s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return err
		}

		// Check if already locked
		isLocked, _ := s.securityRepo.IsAccountLocked(ctx, user.ID)
		if isLocked {
			return errors.New("account is already locked")
		}

		// Lock the account
		lockout := &models.AccountLockout{
			ID:          uuid.New(),
			UserID:      user.ID,
			LockedAt:    time.Now(),
			LockedUntil: time.Now().Add(LockoutDuration),
			Reason:      fmt.Sprintf("Too many failed login attempts (%d)", failedAttempts),
			CreatedAt:   time.Now(),
		}

		if err := s.securityRepo.LockAccount(ctx, lockout); err != nil {
			return err
		}

		return fmt.Errorf("account locked until %s due to too many failed login attempts",
			lockout.LockedUntil.Format(time.RFC3339))
	}

	return nil
}

// IsAccountLocked checks if an account is currently locked
func (s *SecurityService) IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, *models.AccountLockout, error) {
	isLocked, err := s.securityRepo.IsAccountLocked(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	if !isLocked {
		return false, nil, nil
	}

	lockout, err := s.securityRepo.GetAccountLockout(ctx, userID)
	return true, lockout, err
}

// UnlockAccount manually unlocks an account (admin function)
func (s *SecurityService) UnlockAccount(ctx context.Context, userID uuid.UUID) error {
	// Unlock the account
	if err := s.securityRepo.UnlockAccount(ctx, userID); err != nil {
		return err
	}

	// Get user email
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Clear failed login attempts
	return s.securityRepo.ClearLoginAttempts(ctx, user.Email)
}

// ValidatePasswordStrength validates password against requirements
func (s *SecurityService) ValidatePasswordStrength(password string) error {
	requirements := utils.DefaultPasswordRequirements()
	return utils.ValidatePassword(password, requirements)
}

// CheckCommonPassword checks if password is commonly used
func (s *SecurityService) CheckCommonPassword(password string) error {
	if utils.CheckCommonPasswords(password) {
		return errors.New("password is too common, please choose a stronger password")
	}
	return nil
}

// RBAC Functions

// AssignRole assigns a role to a user
func (s *SecurityService) AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	// Get role by name
	role, err := s.securityRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	return s.securityRepo.AssignRoleToUser(ctx, userID, role.ID)
}

// RemoveRole removes a role from a user
func (s *SecurityService) RemoveRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	role, err := s.securityRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	return s.securityRepo.RemoveRoleFromUser(ctx, userID, role.ID)
}

// GetUserRoles gets all roles assigned to a user
func (s *SecurityService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	return s.securityRepo.GetUserRoles(ctx, userID)
}

// CheckPermission checks if a user has a specific permission
func (s *SecurityService) CheckPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	// Check if user has super_admin role (has all permissions)
	roles, err := s.securityRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		// Super admin has all permissions
		if role.Name == "super_admin" {
			return true, nil
		}

		// Check if role has the specific permission or wildcard
		for _, perm := range role.Permissions {
			if perm == "*" || perm == permission {
				return true, nil
			}
		}
	}

	return false, nil
}

// CreateRole creates a new role
func (s *SecurityService) CreateRole(ctx context.Context, name, description string, permissions []string) (*models.Role, error) {
	role := &models.Role{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.securityRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

// ListRoles lists all available roles
func (s *SecurityService) ListRoles(ctx context.Context) ([]*models.Role, error) {
	return s.securityRepo.ListRoles(ctx)
}
