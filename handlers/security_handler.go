package handlers

import (
	"net/http"

	"sso/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SecurityHandler struct {
	securityService *services.SecurityService
}

func NewSecurityHandler(securityService *services.SecurityService) *SecurityHandler {
	return &SecurityHandler{
		securityService: securityService,
	}
}

// UnlockAccountRequest represents the request to unlock an account
type UnlockAccountRequest struct {
	UserID string `json:"userId" binding:"required,uuid"`
}

// AssignRoleRequest represents the request to assign a role
type AssignRoleRequest struct {
	UserID   string `json:"userId" binding:"required,uuid"`
	RoleName string `json:"roleName" binding:"required"`
}

// CreateRoleRequest represents the request to create a new role
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" binding:"required"`
}

// UnlockAccount unlocks a locked user account (admin only)
// @Summary Unlock user account
// @Tags Security
// @Accept json
// @Produce json
// @Param request body UnlockAccountRequest true "Unlock Account Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/admin/unlock-account [post]
func (h *SecurityHandler) UnlockAccount(c *gin.Context) {
	var req UnlockAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.securityService.UnlockAccount(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlock account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account unlocked successfully",
	})
}

// AssignRole assigns a role to a user (admin only)
// @Summary Assign role to user
// @Tags Security
// @Accept json
// @Produce json
// @Param request body AssignRoleRequest true "Assign Role Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/admin/assign-role [post]
func (h *SecurityHandler) AssignRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.securityService.AssignRole(c.Request.Context(), userID, req.RoleName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role assigned successfully",
	})
}

// RemoveRole removes a role from a user (admin only)
// @Summary Remove role from user
// @Tags Security
// @Accept json
// @Produce json
// @Param request body AssignRoleRequest true "Remove Role Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/admin/remove-role [post]
func (h *SecurityHandler) RemoveRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.securityService.RemoveRole(c.Request.Context(), userID, req.RoleName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role removed successfully",
	})
}

// GetUserRoles gets all roles for a user
// @Summary Get user roles
// @Tags Security
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {array} models.Role
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/{userId}/roles [get]
func (h *SecurityHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roles, err := h.securityService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

// ListRoles lists all available roles
// @Summary List all roles
// @Tags Security
// @Produce json
// @Success 200 {array} models.Role
// @Failure 500 {object} map[string]string
// @Router /api/v1/roles [get]
func (h *SecurityHandler) ListRoles(c *gin.Context) {
	roles, err := h.securityService.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

// CreateRole creates a new role (super admin only)
// @Summary Create new role
// @Tags Security
// @Accept json
// @Produce json
// @Param request body CreateRoleRequest true "Create Role Request"
// @Success 201 {object} models.Role
// @Failure 400 {object} map[string]string
// @Router /api/v1/admin/roles [post]
func (h *SecurityHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.securityService.CreateRole(c.Request.Context(), req.Name, req.Description, req.Permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"role":    role,
		"message": "Role created successfully",
	})
}

// GetMyRoles gets the current user's roles
// @Summary Get current user's roles
// @Tags Security
// @Produce json
// @Success 200 {array} models.Role
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/my-roles [get]
func (h *SecurityHandler) GetMyRoles(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	roles, err := h.securityService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

// CheckPermission checks if the current user has a specific permission
// @Summary Check user permission
// @Tags Security
// @Produce json
// @Param permission query string true "Permission name"
// @Success 200 {object} map[string]bool
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/check-permission [get]
func (h *SecurityHandler) CheckPermission(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	permission := c.Query("permission")
	if permission == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Permission parameter required"})
		return
	}

	hasPermission, err := h.securityService.CheckPermission(c.Request.Context(), userID, permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasPermission": hasPermission,
		"permission":    permission,
	})
}
