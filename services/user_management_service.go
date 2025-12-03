package services

import (
	"context"
	"errors"
	"time"

	"sso/models"
	"sso/repository"
	"sso/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserManagementService handles user management business logic
type UserManagementService struct {
	userRepo      *repository.UserManagementRepository
	userRepoBasic *repository.UserRepository
	emailService  *EmailService
}

// NewUserManagementService creates a new UserManagementService
func NewUserManagementService(
	userRepo *repository.UserManagementRepository,
	userRepoBasic *repository.UserRepository,
	emailService *EmailService,
) *UserManagementService {
	return &UserManagementService{
		userRepo:      userRepo,
		userRepoBasic: userRepoBasic,
		emailService:  emailService,
	}
}

// ListUsers retrieves paginated list of users with filters
func (s *UserManagementService) ListUsers(ctx context.Context, filter models.UserListFilter) (*models.UserListResponse, error) {
	users, totalCount, err := s.userRepo.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	totalPages := (totalCount + pageSize - 1) / pageSize

	return &models.UserListResponse{
		Users:      users,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *UserManagementService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.UserDetail, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// CreateUser creates a new user
func (s *UserManagementService) CreateUser(ctx context.Context, req models.UserCreateRequest) (*models.UserDetail, error) {
	// Check if user already exists
	existingUser, err := s.userRepoBasic.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Validate password strength
	if err := utils.ValidatePassword(req.Password, nil); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	now := time.Now()
	user := &models.User{
		ID:            uuid.New(),
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		CompanyID:     req.CompanyID,
		Role:          req.Role,
		IsActive:      req.IsActive,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Send welcome email if requested
	if req.SendWelcomeEmail && s.emailService != nil {
		go s.emailService.SendWelcomeEmail(ctx, user.Email, user.FirstName)
	}

	// Get full user details
	return s.userRepo.GetUserByID(ctx, user.ID)
}

// UpdateUser updates user information
func (s *UserManagementService) UpdateUser(ctx context.Context, userID uuid.UUID, req models.UserUpdateRequest) (*models.UserDetail, error) {
	// Check if user exists
	existingUser, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.CompanyID != nil {
		updates["company_id"] = *req.CompanyID
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return existingUser, nil
	}

	if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return nil, err
	}

	// Get updated user details
	return s.userRepo.GetUserByID(ctx, userID)
}

// UpdateUserProfile updates user profile (by user themselves)
func (s *UserManagementService) UpdateUserProfile(ctx context.Context, userID uuid.UUID, req models.UserProfileUpdateRequest) (*models.UserDetail, error) {
	updates := map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	}

	if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return nil, err
	}

	return s.userRepo.GetUserByID(ctx, userID)
}

// ChangeUserPassword changes user password
func (s *UserManagementService) ChangeUserPassword(ctx context.Context, userID uuid.UUID, req models.UserPasswordChangeRequest) error {
	// Get user
	user, err := s.userRepoBasic.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if err := utils.ValidatePassword(req.NewPassword, nil); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	updates := map[string]interface{}{
		"password_hash": string(hashedPassword),
	}

	if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return err
	}

	// Send notification email
	if s.emailService != nil {
		go s.emailService.SendPasswordChangedEmail(ctx, user.Email, user.FirstName)
	}

	return nil
}

// DeleteUser soft deletes a user
func (s *UserManagementService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.DeleteUser(ctx, userID)
}

// HardDeleteUser permanently deletes a user
func (s *UserManagementService) HardDeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.HardDeleteUser(ctx, userID)
}

// UpdateUserStatus updates user active status
func (s *UserManagementService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, req models.UserStatusUpdateRequest) error {
	return s.userRepo.UpdateUserStatus(ctx, userID, req.IsActive)
}

// UnlockUserAccount unlocks a locked user account
func (s *UserManagementService) UnlockUserAccount(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.UnlockUserAccount(ctx, userID)
}

// BulkAction performs bulk action on multiple users
func (s *UserManagementService) BulkAction(ctx context.Context, req models.UserBulkActionRequest) error {
	if len(req.UserIDs) == 0 {
		return errors.New("no user IDs provided")
	}

	switch req.Action {
	case "activate":
		return s.userRepo.BulkUpdateUsers(ctx, req.UserIDs, map[string]interface{}{
			"is_active": true,
		})
	case "deactivate":
		return s.userRepo.BulkUpdateUsers(ctx, req.UserIDs, map[string]interface{}{
			"is_active": false,
		})
	case "delete":
		return s.userRepo.BulkUpdateUsers(ctx, req.UserIDs, map[string]interface{}{
			"is_active": false,
		})
	case "unlock":
		for _, userID := range req.UserIDs {
			if err := s.userRepo.UnlockUserAccount(ctx, userID); err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("invalid action")
	}
}

// GetUserStats retrieves user statistics
func (s *UserManagementService) GetUserStats(ctx context.Context) (*models.UserStats, error) {
	return s.userRepo.GetUserStats(ctx)
}

// ValidateUserAccess checks if user has access to perform action
func (s *UserManagementService) ValidateUserAccess(ctx context.Context, actorID, targetUserID uuid.UUID) error {
	// Get actor user
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor == nil {
		return errors.New("actor user not found")
	}

	// Admin can access all users
	if actor.Role == "admin" || actor.Role == "super_admin" {
		return nil
	}

	// Manager can access users in same company
	if actor.Role == "manager" {
		targetUser, err := s.userRepo.GetUserByID(ctx, targetUserID)
		if err != nil {
			return err
		}
		if targetUser == nil {
			return errors.New("target user not found")
		}
		if actor.CompanyID == targetUser.CompanyID {
			return nil
		}
	}

	// User can only access their own data
	if actorID == targetUserID {
		return nil
	}

	return errors.New("access denied")
}
