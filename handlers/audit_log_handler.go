package handlers

import (
	"fmt"
	"net/http"
	"time"

	"sso/models"
	"sso/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditLogHandler struct {
	service *services.AuditLogService
}

func NewAuditLogHandler(service *services.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{service: service}
}

// ListAuditLogs godoc
// @Summary List audit logs
// @Description Get list of audit logs with filtering, sorting, and pagination
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param resource query string false "Filter by resource type"
// @Param ip_address query string false "Filter by IP address"
// @Param start_date query string false "Filter from date (RFC3339)"
// @Param end_date query string false "Filter to date (RFC3339)"
// @Param search query string false "Search in action, resource, or details"
// @Param sort_by query string false "Sort by field (created_at, action)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(50)
// @Success 200 {object} models.AuditLogListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs [get]
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var filter models.AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.ListAuditLogs(filter, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAuditLog godoc
// @Summary Get audit log
// @Description Get detailed information about a specific audit log
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param id path string true "Audit Log ID"
// @Success 200 {object} models.AuditLogDetail
// @Failure 404 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/{id} [get]
func (h *AuditLogHandler) GetAuditLog(c *gin.Context) {
	userID, _ := c.Get("user_id")
	logIDStr := c.Param("id")

	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audit log ID"})
		return
	}

	log, err := h.service.GetAuditLog(logID, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, log)
}

// GetAuditLogStats godoc
// @Summary Get audit log statistics
// @Description Get statistics about audit logs
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Success 200 {object} models.AuditLogStats
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/stats [get]
func (h *AuditLogHandler) GetAuditLogStats(c *gin.Context) {
	userID, _ := c.Get("user_id")

	stats, err := h.service.GetAuditLogStats(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAuditTimeline godoc
// @Summary Get audit timeline
// @Description Get timeline of audit events for a specific resource or user
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param request body models.AuditLogTimelineRequest true "Timeline parameters"
// @Success 200 {object} models.AuditLogTimelineResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/timeline [post]
func (h *AuditLogHandler) GetAuditTimeline(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.AuditLogTimelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.GetAuditTimeline(req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExportAuditLogs godoc
// @Summary Export audit logs
// @Description Export audit logs in CSV or JSON format
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param request body models.AuditLogExportRequest true "Export parameters"
// @Success 200 {object} models.AuditLogExportResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/export [post]
func (h *AuditLogHandler) ExportAuditLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.AuditLogExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.ExportAuditLogs(req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CleanupOldLogs godoc
// @Summary Cleanup old audit logs
// @Description Remove old audit logs based on retention policy
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param request body models.AuditLogCleanupRequest true "Cleanup parameters"
// @Success 200 {object} models.AuditLogCleanupResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/cleanup [post]
func (h *AuditLogHandler) CleanupOldLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.AuditLogCleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate older_than date
	if req.OlderThan.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "older_than date is required"})
		return
	}

	response, err := h.service.CleanupOldLogs(req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDistinctActions godoc
// @Summary Get distinct actions
// @Description Get all unique action types in audit logs
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Success 200 {object} map[string][]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/actions [get]
func (h *AuditLogHandler) GetDistinctActions(c *gin.Context) {
	userID, _ := c.Get("user_id")

	actions, err := h.service.GetDistinctActions(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"actions": actions})
}

// GetDistinctResources godoc
// @Summary Get distinct resources
// @Description Get all unique resource types in audit logs
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Success 200 {object} map[string][]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/resources [get]
func (h *AuditLogHandler) GetDistinctResources(c *gin.Context) {
	userID, _ := c.Get("user_id")

	resources, err := h.service.GetDistinctResources(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resources": resources})
}

// CompareAuditLogs godoc
// @Summary Compare audit logs
// @Description Compare two audit log entries to see changes
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param before_id query string true "Before audit log ID"
// @Param after_id query string true "After audit log ID"
// @Success 200 {object} models.AuditLogCompareResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/compare [get]
func (h *AuditLogHandler) CompareAuditLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")

	beforeIDStr := c.Query("before_id")
	afterIDStr := c.Query("after_id")

	beforeID, err := uuid.Parse(beforeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid before_id"})
		return
	}

	afterID, err := uuid.Parse(afterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid after_id"})
		return
	}

	response, err := h.service.CompareAuditLogs(beforeID, afterID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserActivitySummary godoc
// @Summary Get user activity summary
// @Description Get summary of a user's recent activity
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Param days query int false "Number of days (default 7)"
// @Success 200 {object} models.AuditLogTimelineResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /audit-logs/user-activity [get]
func (h *AuditLogHandler) GetUserActivitySummary(c *gin.Context) {
	requesterID, _ := c.Get("user_id")

	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	days := 7
	if daysParam := c.Query("days"); daysParam != "" {
		fmt.Sscanf(daysParam, "%d", &days)
	}

	req := models.AuditLogTimelineRequest{
		UserID:    &userID,
		StartDate: time.Now().AddDate(0, 0, -days),
		EndDate:   time.Now(),
		Limit:     100,
	}

	response, err := h.service.GetAuditTimeline(req, requesterID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
