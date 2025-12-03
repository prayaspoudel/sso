package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sso/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// CreateNotification creates a new notification
func (r *NotificationRepository) CreateNotification(req *models.NotificationCreateRequest) (*models.Notification, error) {
	notification := &models.Notification{
		ID:         uuid.New(),
		UserID:     req.UserID,
		Type:       req.Type,
		Title:      req.Title,
		Message:    req.Message,
		Priority:   req.Priority,
		Status:     models.NotificationStatusUnread,
		Data:       req.Data,
		ActionURL:  req.ActionURL,
		ActionText: req.ActionText,
		ExpiresAt:  req.ExpiresAt,
		CreatedAt:  time.Now(),
	}

	// Set default priority if not provided
	if notification.Priority == "" {
		notification.Priority = models.NotificationPriorityNormal
	}

	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO notifications (id, user_id, type, title, message, priority, status, data, action_url, action_text, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.Exec(query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Priority,
		notification.Status,
		dataJSON,
		notification.ActionURL,
		notification.ActionText,
		notification.ExpiresAt,
		notification.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return notification, nil
}

// ListNotifications retrieves notifications with filtering
func (r *NotificationRepository) ListNotifications(filter *models.NotificationFilter) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	// Build WHERE clause
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argCount := 1

	if filter.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argCount))
		args = append(args, filter.UserID)
		argCount++
	}

	if filter.Type != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("type = $%d", argCount))
		args = append(args, filter.Type)
		argCount++
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCount))
		args = append(args, filter.Status)
		argCount++
	}

	if filter.Priority != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("priority = $%d", argCount))
		args = append(args, filter.Priority)
		argCount++
	}

	if filter.StartDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argCount))
		args = append(args, filter.StartDate)
		argCount++
	}

	if filter.EndDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argCount))
		args = append(args, filter.EndDate)
		argCount++
	}

	if filter.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(title ILIKE $%d OR message ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	// Add expires_at check (exclude expired notifications)
	whereClauses = append(whereClauses, "(expires_at IS NULL OR expires_at > NOW())")

	whereClause := strings.Join(whereClauses, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s", whereClause)
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	sortBy := "created_at"
	if filter.SortBy == "priority" {
		sortBy = "priority"
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Build pagination
	page := 1
	if filter.Page > 0 {
		page = filter.Page
	}

	pageSize := 20
	if filter.PageSize > 0 && filter.PageSize <= 100 {
		pageSize = filter.PageSize
	}

	offset := (page - 1) * pageSize

	// Query notifications
	query := fmt.Sprintf(`
		SELECT id, user_id, type, title, message, priority, status, data, action_url, action_text, read_at, created_at, expires_at
		FROM notifications
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argCount, argCount+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var notification models.Notification
		var dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.Priority,
			&notification.Status,
			&dataJSON,
			&notification.ActionURL,
			&notification.ActionText,
			&notification.ReadAt,
			&notification.CreatedAt,
			&notification.ExpiresAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(dataJSON) > 0 {
			json.Unmarshal(dataJSON, &notification.Data)
		}

		notifications = append(notifications, notification)
	}

	return notifications, total, nil
}

// GetNotificationByID retrieves a notification by ID
func (r *NotificationRepository) GetNotificationByID(notificationID uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	var dataJSON []byte

	query := `
		SELECT id, user_id, type, title, message, priority, status, data, action_url, action_text, read_at, created_at, expires_at
		FROM notifications
		WHERE id = $1
	`

	err := r.db.QueryRow(query, notificationID).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Priority,
		&notification.Status,
		&dataJSON,
		&notification.ActionURL,
		&notification.ActionText,
		&notification.ReadAt,
		&notification.CreatedAt,
		&notification.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification not found")
	}
	if err != nil {
		return nil, err
	}

	if len(dataJSON) > 0 {
		json.Unmarshal(dataJSON, &notification.Data)
	}

	return &notification, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(notificationID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE notifications
		SET status = $1, read_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, models.NotificationStatusRead, now, notificationID)
	return err
}

// MarkMultipleAsRead marks multiple notifications as read
func (r *NotificationRepository) MarkMultipleAsRead(notificationIDs []uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE notifications
		SET status = $1, read_at = $2
		WHERE id = ANY($3)
	`

	_, err := r.db.Exec(query, models.NotificationStatusRead, now, pq.Array(notificationIDs))
	return err
}

// MarkAllAsReadForUser marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsReadForUser(userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE notifications
		SET status = $1, read_at = $2
		WHERE user_id = $3 AND status = $4
	`

	_, err := r.db.Exec(query, models.NotificationStatusRead, now, userID, models.NotificationStatusUnread)
	return err
}

// DeleteNotification deletes a notification
func (r *NotificationRepository) DeleteNotification(notificationID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.db.Exec(query, notificationID)
	return err
}

// DeleteMultipleNotifications deletes multiple notifications
func (r *NotificationRepository) DeleteMultipleNotifications(notificationIDs []uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = ANY($1)`
	_, err := r.db.Exec(query, pq.Array(notificationIDs))
	return err
}

// GetUnreadCount gets the count of unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND status = $2 AND (expires_at IS NULL OR expires_at > NOW())
	`

	err := r.db.Get(&count, query, userID, models.NotificationStatusUnread)
	return count, err
}

// GetNotificationStats retrieves notification statistics
func (r *NotificationRepository) GetNotificationStats(userID *uuid.UUID) (*models.NotificationStats, error) {
	stats := &models.NotificationStats{
		ByType:     make(map[models.NotificationType]int64),
		ByPriority: make(map[models.NotificationPriority]int64),
		ByStatus:   make(map[models.NotificationStatus]int64),
	}

	whereClause := "1=1"
	var args []interface{}
	if userID != nil {
		whereClause = "user_id = $1"
		args = append(args, userID)
	}

	// Total notifications
	query := fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s", whereClause)
	r.db.Get(&stats.TotalNotifications, query, args...)

	// Unread notifications
	query = fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s AND status = '%s'", whereClause, models.NotificationStatusUnread)
	r.db.Get(&stats.UnreadNotifications, query, args...)

	// Read notifications
	query = fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s AND status = '%s'", whereClause, models.NotificationStatusRead)
	r.db.Get(&stats.ReadNotifications, query, args...)

	// Archived notifications
	query = fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s AND status = '%s'", whereClause, models.NotificationStatusArchived)
	r.db.Get(&stats.ArchivedNotifications, query, args...)

	// Notifications today
	query = fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s AND created_at >= CURRENT_DATE", whereClause)
	r.db.Get(&stats.NotificationsToday, query, args...)

	// Notifications this week
	query = fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s AND created_at >= DATE_TRUNC('week', CURRENT_DATE)", whereClause)
	r.db.Get(&stats.NotificationsThisWeek, query, args...)

	// Notifications this month
	query = fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s AND created_at >= DATE_TRUNC('month', CURRENT_DATE)", whereClause)
	r.db.Get(&stats.NotificationsThisMonth, query, args...)

	// By type
	query = fmt.Sprintf("SELECT type, COUNT(*) as count FROM notifications WHERE %s GROUP BY type", whereClause)
	rows, _ := r.db.Query(query, args...)
	defer rows.Close()
	for rows.Next() {
		var notifType models.NotificationType
		var count int64
		rows.Scan(&notifType, &count)
		stats.ByType[notifType] = count
	}

	// By priority
	query = fmt.Sprintf("SELECT priority, COUNT(*) as count FROM notifications WHERE %s GROUP BY priority", whereClause)
	rows, _ = r.db.Query(query, args...)
	defer rows.Close()
	for rows.Next() {
		var priority models.NotificationPriority
		var count int64
		rows.Scan(&priority, &count)
		stats.ByPriority[priority] = count
	}

	// By status
	query = fmt.Sprintf("SELECT status, COUNT(*) as count FROM notifications WHERE %s GROUP BY status", whereClause)
	rows, _ = r.db.Query(query, args...)
	defer rows.Close()
	for rows.Next() {
		var status models.NotificationStatus
		var count int64
		rows.Scan(&status, &count)
		stats.ByStatus[status] = count
	}

	return stats, nil
}

// CleanupExpiredNotifications deletes expired notifications
func (r *NotificationRepository) CleanupExpiredNotifications() (int64, error) {
	query := `
		DELETE FROM notifications
		WHERE expires_at IS NOT NULL AND expires_at <= NOW()
	`

	result, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	return count, nil
}

// GetOrCreatePreference gets or creates notification preferences for a user
func (r *NotificationRepository) GetOrCreatePreference(userID uuid.UUID) (*models.NotificationPreference, error) {
	// Try to get existing preference
	var pref models.NotificationPreference
	query := `
		SELECT id, user_id, email_enabled, sms_enabled, push_enabled, websocket_enabled, 
		       enabled_types, min_priority, quiet_hours_start, quiet_hours_end, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	err := r.db.QueryRow(query, userID).Scan(
		&pref.ID,
		&pref.UserID,
		&pref.EmailEnabled,
		&pref.SMSEnabled,
		&pref.PushEnabled,
		&pref.WebSocketEnabled,
		pq.Array(&pref.EnabledTypes),
		&pref.MinPriority,
		&pref.QuietHoursStart,
		&pref.QuietHoursEnd,
		&pref.CreatedAt,
		&pref.UpdatedAt,
	)

	if err == nil {
		return &pref, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create default preference
	pref = models.NotificationPreference{
		ID:               uuid.New(),
		UserID:           userID,
		EmailEnabled:     true,
		SMSEnabled:       false,
		PushEnabled:      true,
		WebSocketEnabled: true,
		EnabledTypes:     []models.NotificationType{}, // Empty means all types
		MinPriority:      models.NotificationPriorityNormal,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	insertQuery := `
		INSERT INTO notification_preferences (id, user_id, email_enabled, sms_enabled, push_enabled, websocket_enabled, 
		                                     enabled_types, min_priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Exec(insertQuery,
		pref.ID,
		pref.UserID,
		pref.EmailEnabled,
		pref.SMSEnabled,
		pref.PushEnabled,
		pref.WebSocketEnabled,
		pq.Array(pref.EnabledTypes),
		pref.MinPriority,
		pref.CreatedAt,
		pref.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &pref, nil
}

// UpdatePreference updates notification preferences
func (r *NotificationRepository) UpdatePreference(userID uuid.UUID, req *models.NotificationPreferenceUpdateRequest) (*models.NotificationPreference, error) {
	// Get existing preference
	pref, err := r.GetOrCreatePreference(userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.EmailEnabled != nil {
		pref.EmailEnabled = *req.EmailEnabled
	}
	if req.SMSEnabled != nil {
		pref.SMSEnabled = *req.SMSEnabled
	}
	if req.PushEnabled != nil {
		pref.PushEnabled = *req.PushEnabled
	}
	if req.WebSocketEnabled != nil {
		pref.WebSocketEnabled = *req.WebSocketEnabled
	}
	if req.EnabledTypes != nil {
		pref.EnabledTypes = req.EnabledTypes
	}
	if req.MinPriority != "" {
		pref.MinPriority = req.MinPriority
	}
	if req.QuietHoursStart != nil {
		pref.QuietHoursStart = req.QuietHoursStart
	}
	if req.QuietHoursEnd != nil {
		pref.QuietHoursEnd = req.QuietHoursEnd
	}

	pref.UpdatedAt = time.Now()

	// Update in database
	query := `
		UPDATE notification_preferences
		SET email_enabled = $1, sms_enabled = $2, push_enabled = $3, websocket_enabled = $4,
		    enabled_types = $5, min_priority = $6, quiet_hours_start = $7, quiet_hours_end = $8, updated_at = $9
		WHERE user_id = $10
	`

	_, err = r.db.Exec(query,
		pref.EmailEnabled,
		pref.SMSEnabled,
		pref.PushEnabled,
		pref.WebSocketEnabled,
		pq.Array(pref.EnabledTypes),
		pref.MinPriority,
		pref.QuietHoursStart,
		pref.QuietHoursEnd,
		pref.UpdatedAt,
		userID,
	)

	if err != nil {
		return nil, err
	}

	return pref, nil
}
