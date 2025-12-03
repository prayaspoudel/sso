package handlers

import (
	"net/http"
	"strconv"

	"sso/models"
	"sso/services"

	"github.com/gin-gonic/gin"
)

type CompanyManagementHandler struct {
	service *services.CompanyManagementService
}

func NewCompanyManagementHandler(service *services.CompanyManagementService) *CompanyManagementHandler {
	return &CompanyManagementHandler{service: service}
}

// ListCompanies godoc
// @Summary List companies
// @Description Get list of companies with filtering, sorting, and pagination
// @Tags Company Management
// @Accept json
// @Produce json
// @Param search query string false "Search by name or domain"
// @Param status query string false "Filter by status (active, inactive, suspended)"
// @Param industry query string false "Filter by industry"
// @Param sort_by query string false "Sort by field (name, created_at, user_count)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} models.CompanyListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /companies [get]
func (h *CompanyManagementHandler) ListCompanies(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var filter models.CompanyListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.ListCompanies(filter, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCompany godoc
// @Summary Get company
// @Description Get detailed information about a company
// @Tags Company Management
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} models.CompanyDetail
// @Failure 404 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /companies/{id} [get]
func (h *CompanyManagementHandler) GetCompany(c *gin.Context) {
	userID, _ := c.Get("user_id")
	companyID := c.Param("id")

	company, err := h.service.GetCompany(companyID, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, company)
}

// CreateCompany godoc
// @Summary Create company
// @Description Create a new company
// @Tags Company Management
// @Accept json
// @Produce json
// @Param request body models.CompanyCreateRequest true "Company details"
// @Success 201 {object} models.CompanyDetail
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /companies [post]
func (h *CompanyManagementHandler) CreateCompany(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.CompanyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company, err := h.service.CreateCompany(req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, company)
}

// UpdateCompany godoc
// @Summary Update company
// @Description Update company information
// @Tags Company Management
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param request body models.CompanyUpdateRequest true "Updated company details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /companies/{id} [put]
func (h *CompanyManagementHandler) UpdateCompany(c *gin.Context) {
	userID, _ := c.Get("user_id")
	companyID := c.Param("id")

	var req models.CompanyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateCompany(companyID, req, userID.(string))
	if err != nil {
		if err.Error() == "insufficient permissions to update company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company updated successfully"})
}

// DeleteCompany godoc
// @Summary Delete company
// @Description Soft delete a company
// @Tags Company Management
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /companies/{id} [delete]
func (h *CompanyManagementHandler) DeleteCompany(c *gin.Context) {
	userID, _ := c.Get("user_id")
	companyID := c.Param("id")

	err := h.service.DeleteCompany(companyID, userID.(string))
	if err != nil {
		if err.Error() == "only company owner can delete company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}

// UpdateCompanyStatus godoc
// @Summary Update company status
// @Description Update the status of a company
// @Tags Company Management
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param request body models.CompanyStatusUpdateRequest true "Status update details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security BearerAuth
// @Router /companies/{id}/status [put]
func (h *CompanyManagementHandler) UpdateCompanyStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	companyID := c.Param("id")

	var req models.CompanyStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateCompanyStatus(companyID, req, userID.(string))
	if err != nil {
		if err.Error() == "insufficient permissions to update company status" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company status updated successfully"})
}

// AddUserToCompany godoc
// @Summary Add user to company
// @Description Add a user to a company with specified role
// @Tags Company Management
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param request body models.UserCompanyAddRequest true "User details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security BearerAuth
// @Router /companies/{id}/users [post]
func (h *CompanyManagementHandler) AddUserToCompany(c *gin.Context) {
	userID, _ := c.Get("user_id")
	companyID := c.Param("id")

	var req models.UserCompanyAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.AddUserToCompany(companyID, req, userID.(string))
	if err != nil {
		if err.Error() == "insufficient permissions to add users to company" || err.Error() == "only company owner can add other owners" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to company successfully"})
}

// RemoveUserFromCompany godoc
// @Summary Remove user from company
// @Description Remove a user from a company
// @Tags Company Management
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /companies/{id}/users/{user_id} [delete]
func (h *CompanyManagementHandler) RemoveUserFromCompany(c *gin.Context) {
	userID, _ := c.Get("user_id")
	companyID := c.Param("id")
	targetUserID := c.Param("user_id")

	err := h.service.RemoveUserFromCompany(companyID, targetUserID, userID.(string))
	if err != nil {
		if err.Error() == "insufficient permissions to remove users from company" || err.Error() == "only company owner can remove other owners" || err.Error() == "cannot remove the only owner of the company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from company successfully"})
}

// UpdateUserRoleInCompany godoc
// @Summary Update user role in company
// @Description Update a user's role within a company
// @Tags Company Management
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param user_id path string true "User ID"
// @Param request body models.UserCompanyUpdateRequest true "Role update details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security BearerAuth
// @Router /companies/{id}/users/{user_id}/role [put]
func (h *CompanyManagementHandler) UpdateUserRoleInCompany(c *gin.Context) {
	userID, _ := c.Get("user_id")
	companyID := c.Param("id")
	targetUserID := c.Param("user_id")

	var req models.UserCompanyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateUserRoleInCompany(companyID, targetUserID, req, userID.(string))
	if err != nil {
		if err.Error() == "insufficient permissions to update user roles" || err.Error() == "only company owner can change owner role" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// GetCompanyUsers godoc
// @Summary Get company users
// @Description Get list of users in a company
// @Tags Company Management
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} models.CompanyUsersResponse
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /companies/{id}/users [get]
func (h *CompanyManagementHandler) GetCompanyUsers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	companyID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.service.GetCompanyUsers(companyID, userID.(string), page, pageSize)
	if err != nil {
		if err.Error() == "you don't have access to this company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCompanyStats godoc
// @Summary Get company statistics
// @Description Get statistics about companies
// @Tags Company Management
// @Accept json
// @Produce json
// @Success 200 {object} models.CompanyStats
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /companies/stats [get]
func (h *CompanyManagementHandler) GetCompanyStats(c *gin.Context) {
	userID, _ := c.Get("user_id")

	stats, err := h.service.GetCompanyStats(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// BulkActionCompanies godoc
// @Summary Bulk action on companies
// @Description Perform bulk actions on multiple companies
// @Tags Company Management
// @Accept json
// @Produce json
// @Param request body models.CompanyBulkActionRequest true "Bulk action details"
// @Success 200 {object} models.CompanyBulkActionResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /companies/bulk [post]
func (h *CompanyManagementHandler) BulkActionCompanies(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.CompanyBulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.BulkActionCompanies(req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
