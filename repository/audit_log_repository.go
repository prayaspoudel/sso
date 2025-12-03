package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sso/models"

	"github.com/google/uuid"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// CreateAuditLog creates a new audit log entry
func (r *AuditLogRepository) CreateAuditLog(log models.AuditLogCreateRequest) error {
	detailsJSON, _ := json.Marshal(log.Details)

	query := `
		INSERT INTO audit_logs (user_id, action, resource, details, ip_address)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(query, log.UserID, log.Action, log.Resource, detailsJSON, log.IPAddress)
	return err
}

// ListAuditLogs retrieves audit logs with filtering, sorting, and pagination
func (r *AuditLogRepository) ListAuditLogs(filter models.AuditLogFilter) ([]models.AuditLogDetail, int64, error) {
	// Build WHERE clause
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argCount := 1

	if filter.UserID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("al.user_id = $%d", argCount))
		args = append(args, filter.UserID)
		argCount++
	}

	if filter.Action != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("al.action = $%d", argCount))
		args = append(args, filter.Action)
		argCount++
	}

	if filter.Resource != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("al.resource = $%d", argCount))
		args = append(args, filter.Resource)
		argCount++
	}

	if filter.IPAddress != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("al.ip_address = $%d", argCount))
		args = append(args, filter.IPAddress)
		argCount++
	}

	if !filter.StartDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("al.created_at >= $%d", argCount))
		args = append(args, filter.StartDate)
		argCount++
	}

	if !filter.EndDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("al.created_at <= $%d", argCount))
		args = append(args, filter.EndDate)
		argCount++
	}

	if filter.SearchTerm != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(al.action ILIKE $%d OR al.resource ILIKE $%d OR al.details::text ILIKE $%d)", argCount, argCount, argCount))
		args = append(args, "%"+filter.SearchTerm+"%")
		argCount++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM audit_logs al
		WHERE %s
	`, whereSQL)

	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	sortBy := "al.created_at"
	if filter.SortBy == "action" {
		sortBy = "al.action"
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
		filter.PageSize = 50
	}
	offset := (filter.Page - 1) * filter.PageSize

	// Query audit logs with user email
	query := fmt.Sprintf(`
		SELECT 
			al.id,
			al.user_id,
			u.email,
			al.action,
			al.resource,
			al.details,
			al.ip_address,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, sortBy, sortOrder, argCount, argCount+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLogDetail
	for rows.Next() {
		var log models.AuditLogDetail
		var detailsJSON []byte
		var userEmail sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&userEmail,
			&log.Action,
			&log.Resource,
			&detailsJSON,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if userEmail.Valid {
			log.UserEmail = userEmail.String
		}

		if len(detailsJSON) > 0 {
			json.Unmarshal(detailsJSON, &log.Details)
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetAuditLogByID retrieves an audit log by ID
func (r *AuditLogRepository) GetAuditLogByID(logID uuid.UUID) (*models.AuditLogDetail, error) {
	query := `
		SELECT 
			al.id,
			al.user_id,
			u.email,
			al.action,
			al.resource,
			al.details,
			al.ip_address,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.id = $1
	`

	var log models.AuditLogDetail
	var detailsJSON []byte
	var userEmail sql.NullString

	err := r.db.QueryRow(query, logID).Scan(
		&log.ID,
		&log.UserID,
		&userEmail,
		&log.Action,
		&log.Resource,
		&detailsJSON,
		&log.IPAddress,
		&log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if userEmail.Valid {
		log.UserEmail = userEmail.String
	}

	if len(detailsJSON) > 0 {
		json.Unmarshal(detailsJSON, &log.Details)
	}

	return &log, nil
}

// GetAuditLogStats retrieves audit log statistics
func (r *AuditLogRepository) GetAuditLogStats() (*models.AuditLogStats, error) {
	var stats models.AuditLogStats

	// Total logs
	err := r.db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&stats.TotalLogs)
	if err != nil {
		return nil, err
	}

	// Logs today
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM audit_logs 
		WHERE created_at >= CURRENT_DATE
	`).Scan(&stats.LogsToday)
	if err != nil {
		return nil, err
	}

	// Logs this week
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM audit_logs 
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
	`).Scan(&stats.LogsThisWeek)
	if err != nil {
		return nil, err
	}

	// Logs this month
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM audit_logs 
		WHERE created_at >= DATE_TRUNC('month', CURRENT_DATE)
	`).Scan(&stats.LogsThisMonth)
	if err != nil {
		return nil, err
	}

	// Top actions
	stats.TopActions = make(map[string]int64)
	rows, err := r.db.Query(`
		SELECT action, COUNT(*) as count
		FROM audit_logs
		GROUP BY action
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int64
		if err := rows.Scan(&action, &count); err != nil {
			continue
		}
		stats.TopActions[action] = count
	}

	// Top resources
	stats.TopResources = make(map[string]int64)
	rows, err = r.db.Query(`
		SELECT resource, COUNT(*) as count
		FROM audit_logs
		GROUP BY resource
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var resource string
		var count int64
		if err := rows.Scan(&resource, &count); err != nil {
			continue
		}
		stats.TopResources[resource] = count
	}

	// Top users
	rows, err = r.db.Query(`
		SELECT al.user_id, u.email, COUNT(*) as count
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.user_id IS NOT NULL
		GROUP BY al.user_id, u.email
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat models.AuditUserStat
		var email sql.NullString
		if err := rows.Scan(&stat.UserID, &email, &stat.LogCount); err != nil {
			continue
		}
		if email.Valid {
			stat.UserEmail = email.String
		}
		stats.TopUsers = append(stats.TopUsers, stat)
	}

	// Activity by hour (last 24 hours)
	stats.ActivityByHour = make(map[string]int64)
	rows, err = r.db.Query(`
		SELECT 
			TO_CHAR(created_at, 'HH24:00') as hour,
			COUNT(*) as count
		FROM audit_logs
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY hour
		ORDER BY hour
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hour string
		var count int64
		if err := rows.Scan(&hour, &count); err != nil {
			continue
		}
		stats.ActivityByHour[hour] = count
	}

	// Activity by day (last 7 days)
	stats.ActivityByDay = make(map[string]int64)
	rows, err = r.db.Query(`
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM-DD') as day,
			COUNT(*) as count
		FROM audit_logs
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
		GROUP BY day
		ORDER BY day
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var count int64
		if err := rows.Scan(&day, &count); err != nil {
			continue
		}
		stats.ActivityByDay[day] = count
	}

	return &stats, nil
}

// GetAuditTimeline retrieves audit logs for a specific timeline
func (r *AuditLogRepository) GetAuditTimeline(req models.AuditLogTimelineRequest) ([]models.AuditLogDetail, int64, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argCount := 1

	if req.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.user_id = $%d", argCount))
		args = append(args, req.UserID)
		argCount++
	}

	if req.Resource != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("al.resource = $%d", argCount))
		args = append(args, req.Resource)
		argCount++
	}

	if req.ResourceID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("al.details->>'resource_id' = $%d", argCount))
		args = append(args, req.ResourceID)
		argCount++
	}

	if !req.StartDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("al.created_at >= $%d", argCount))
		args = append(args, req.StartDate)
		argCount++
	}

	if !req.EndDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("al.created_at <= $%d", argCount))
		args = append(args, req.EndDate)
		argCount++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs al WHERE %s", whereSQL)
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Default limit
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	// Query timeline
	query := fmt.Sprintf(`
		SELECT 
			al.id,
			al.user_id,
			u.email,
			al.action,
			al.resource,
			al.details,
			al.ip_address,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE %s
		ORDER BY al.created_at DESC
		LIMIT $%d
	`, whereSQL, argCount)

	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLogDetail
	for rows.Next() {
		var log models.AuditLogDetail
		var detailsJSON []byte
		var userEmail sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&userEmail,
			&log.Action,
			&log.Resource,
			&detailsJSON,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if userEmail.Valid {
			log.UserEmail = userEmail.String
		}

		if len(detailsJSON) > 0 {
			json.Unmarshal(detailsJSON, &log.Details)
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// DeleteOldAuditLogs deletes audit logs older than specified date
func (r *AuditLogRepository) DeleteOldAuditLogs(resource string, olderThan time.Time) (int64, error) {
	var query string
	var args []interface{}

	if resource != "" {
		query = "DELETE FROM audit_logs WHERE resource = $1 AND created_at < $2"
		args = []interface{}{resource, olderThan}
	} else {
		query = "DELETE FROM audit_logs WHERE created_at < $1"
		args = []interface{}{olderThan}
	}

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// CountOldAuditLogs counts audit logs that would be deleted
func (r *AuditLogRepository) CountOldAuditLogs(resource string, olderThan time.Time) (int64, error) {
	var query string
	var args []interface{}
	var count int64

	if resource != "" {
		query = "SELECT COUNT(*) FROM audit_logs WHERE resource = $1 AND created_at < $2"
		args = []interface{}{resource, olderThan}
	} else {
		query = "SELECT COUNT(*) FROM audit_logs WHERE created_at < $1"
		args = []interface{}{olderThan}
	}

	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetDistinctActions returns all unique action types
func (r *AuditLogRepository) GetDistinctActions() ([]string, error) {
	query := `
		SELECT DISTINCT action 
		FROM audit_logs 
		ORDER BY action
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []string
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			continue
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// GetDistinctResources returns all unique resource types
func (r *AuditLogRepository) GetDistinctResources() ([]string, error) {
	query := `
		SELECT DISTINCT resource 
		FROM audit_logs 
		ORDER BY resource
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []string
	for rows.Next() {
		var resource string
		if err := rows.Scan(&resource); err != nil {
			continue
		}
		resources = append(resources, resource)
	}

	return resources, nil
}
