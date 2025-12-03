package services

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sso/models"
	"sso/repository"

	"github.com/google/uuid"
)

type AuditLogService struct {
	repo *repository.AuditLogRepository
}

func NewAuditLogService(db *sql.DB) *AuditLogService {
	return &AuditLogService{
		repo: repository.NewAuditLogRepository(db),
	}
}

// LogActivity creates a new audit log entry
func (s *AuditLogService) LogActivity(userID *uuid.UUID, action, resource, ipAddress string, details map[string]interface{}) error {
	req := models.AuditLogCreateRequest{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Details:   details,
		IPAddress: ipAddress,
	}

	return s.repo.CreateAuditLog(req)
}

// ListAuditLogs retrieves audit logs with filtering
func (s *AuditLogService) ListAuditLogs(filter models.AuditLogFilter, requesterID string) (*models.AuditLogListResponse, error) {
	// TODO: Check if requester has permission to view audit logs
	// For now, allow all authenticated users

	logs, total, err := s.repo.ListAuditLogs(filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &models.AuditLogListResponse{
		Logs:      logs,
		Total:     total,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		TotalPage: totalPages,
	}, nil
}

// GetAuditLog retrieves a single audit log by ID
func (s *AuditLogService) GetAuditLog(logID uuid.UUID, requesterID string) (*models.AuditLogDetail, error) {
	// TODO: Check if requester has permission to view audit logs

	log, err := s.repo.GetAuditLogByID(logID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("audit log not found")
		}
		return nil, err
	}

	return log, nil
}

// GetAuditLogStats retrieves audit log statistics
func (s *AuditLogService) GetAuditLogStats(requesterID string) (*models.AuditLogStats, error) {
	// TODO: Check if requester has permission (should be super_admin)

	stats, err := s.repo.GetAuditLogStats()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetAuditTimeline retrieves audit timeline for a resource
func (s *AuditLogService) GetAuditTimeline(req models.AuditLogTimelineRequest, requesterID string) (*models.AuditLogTimelineResponse, error) {
	// TODO: Check if requester has permission to view timeline

	logs, total, err := s.repo.GetAuditTimeline(req)
	if err != nil {
		return nil, err
	}

	return &models.AuditLogTimelineResponse{
		Events:    logs,
		Total:     total,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}, nil
}

// ExportAuditLogs exports audit logs in specified format
func (s *AuditLogService) ExportAuditLogs(req models.AuditLogExportRequest, requesterID string) (*models.AuditLogExportResponse, error) {
	// TODO: Check if requester has permission to export logs

	// Set default max records
	if req.MaxRecords <= 0 {
		req.MaxRecords = 10000
	}

	// Override page size for export
	req.Filter.PageSize = req.MaxRecords
	req.Filter.Page = 1

	// Get logs
	logs, _, err := s.repo.ListAuditLogs(req.Filter)
	if err != nil {
		return nil, err
	}

	var fileData string
	var fileName string

	switch req.Format {
	case "json":
		fileName = fmt.Sprintf("audit_logs_%s.json", time.Now().Format("20060102_150405"))
		jsonData, err := json.MarshalIndent(logs, "", "  ")
		if err != nil {
			return nil, err
		}
		fileData = string(jsonData)

	case "csv":
		fileName = fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("20060102_150405"))
		csvData, err := s.convertToCSV(logs)
		if err != nil {
			return nil, err
		}
		fileData = csvData

	default:
		return nil, errors.New("unsupported export format")
	}

	return &models.AuditLogExportResponse{
		FileName:    fileName,
		FileData:    fileData,
		RecordCount: len(logs),
		Format:      req.Format,
		CreatedAt:   time.Now(),
	}, nil
}

// convertToCSV converts audit logs to CSV format
func (s *AuditLogService) convertToCSV(logs []models.AuditLogDetail) (string, error) {
	var b strings.Builder
	writer := csv.NewWriter(&b)

	// Write header
	header := []string{"ID", "User ID", "User Email", "Action", "Resource", "IP Address", "Details", "Created At"}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	// Write rows
	for _, log := range logs {
		detailsJSON, _ := json.Marshal(log.Details)
		userID := ""
		if log.UserID != nil {
			userID = log.UserID.String()
		}

		row := []string{
			log.ID.String(),
			userID,
			log.UserEmail,
			log.Action,
			log.Resource,
			log.IPAddress,
			string(detailsJSON),
			log.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return b.String(), nil
}

// CleanupOldLogs removes old audit logs based on retention policy
func (s *AuditLogService) CleanupOldLogs(req models.AuditLogCleanupRequest, requesterID string) (*models.AuditLogCleanupResponse, error) {
	// TODO: Check if requester has permission (should be super_admin)

	var deletedCount int64
	var err error

	if req.DryRun {
		// Just count what would be deleted
		deletedCount, err = s.repo.CountOldAuditLogs(req.Resource, req.OlderThan)
		if err != nil {
			return nil, err
		}
	} else {
		// TODO: Implement archive functionality if req.Archive is true

		// Actually delete
		deletedCount, err = s.repo.DeleteOldAuditLogs(req.Resource, req.OlderThan)
		if err != nil {
			return nil, err
		}
	}

	return &models.AuditLogCleanupResponse{
		DeletedCount: deletedCount,
		DryRun:       req.DryRun,
		ExecutedAt:   time.Now(),
	}, nil
}

// GetDistinctActions returns all unique action types
func (s *AuditLogService) GetDistinctActions(requesterID string) ([]string, error) {
	// TODO: Check permissions

	return s.repo.GetDistinctActions()
}

// GetDistinctResources returns all unique resource types
func (s *AuditLogService) GetDistinctResources(requesterID string) ([]string, error) {
	// TODO: Check permissions

	return s.repo.GetDistinctResources()
}

// CompareAuditLogs compares two audit log entries
func (s *AuditLogService) CompareAuditLogs(beforeID, afterID uuid.UUID, requesterID string) (*models.AuditLogCompareResponse, error) {
	// TODO: Check permissions

	beforeLog, err := s.repo.GetAuditLogByID(beforeID)
	if err != nil {
		return nil, fmt.Errorf("before log not found: %w", err)
	}

	afterLog, err := s.repo.GetAuditLogByID(afterID)
	if err != nil {
		return nil, fmt.Errorf("after log not found: %w", err)
	}

	// Compare details
	changes := s.compareDetails(beforeLog.Details, afterLog.Details)

	return &models.AuditLogCompareResponse{
		BeforeLog: *beforeLog,
		AfterLog:  *afterLog,
		Changes:   changes,
	}, nil
}

// compareDetails compares two detail maps and returns differences
func (s *AuditLogService) compareDetails(before, after map[string]interface{}) []models.AuditLogDiff {
	var diffs []models.AuditLogDiff

	// Check all keys in before
	for key, beforeVal := range before {
		afterVal, exists := after[key]
		if !exists {
			diffs = append(diffs, models.AuditLogDiff{
				Field:    key,
				OldValue: beforeVal,
				NewValue: nil,
			})
		} else if fmt.Sprintf("%v", beforeVal) != fmt.Sprintf("%v", afterVal) {
			diffs = append(diffs, models.AuditLogDiff{
				Field:    key,
				OldValue: beforeVal,
				NewValue: afterVal,
			})
		}
	}

	// Check for new keys in after
	for key, afterVal := range after {
		if _, exists := before[key]; !exists {
			diffs = append(diffs, models.AuditLogDiff{
				Field:    key,
				OldValue: nil,
				NewValue: afterVal,
			})
		}
	}

	return diffs
}

// ScheduleCleanup sets up automatic cleanup based on retention policies
func (s *AuditLogService) ScheduleCleanup(retentionDays int) error {
	// TODO: Implement scheduled cleanup using cron or similar
	// This would run periodically to clean up old logs
	return errors.New("scheduled cleanup not yet implemented")
}
