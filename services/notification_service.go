package services

import (
	"fmt"
	"time"

	"sso/models"
	"sso/repository"

	"github.com/google/uuid"
)

type NotificationService struct {
	repo *repository.NotificationRepository
	hub  *WebSocketHub
}

func NewNotificationService(repo *repository.NotificationRepository, hub *WebSocketHub) *NotificationService {
	return &NotificationService{
		repo: repo,
		hub:  hub,
	}
}

// CreateNotification creates a new notification and sends it via WebSocket
func (s *NotificationService) CreateNotification(req *models.NotificationCreateRequest, senderID *uuid.UUID) (*models.Notification, error) {
	// Create notification in database
	notification, err := s.repo.CreateNotification(req)
	if err != nil {
		return nil, err
	}

	// Send via WebSocket if user is connected
	if notification.UserID != nil && s.hub.IsUserConnected(*notification.UserID) {
		s.hub.SendNotificationToUser(*notification.UserID, notification)
	}

	// TODO: Send via email/SMS based on user preferences

	return notification, nil
}

// CreateNotificationForUser is a helper to create notification for a specific user
func (s *NotificationService) CreateNotificationForUser(
	userID uuid.UUID,
	notifType models.NotificationType,
	title, message string,
	priority models.NotificationPriority,
	data map[string]interface{},
) (*models.Notification, error) {
	req := &models.NotificationCreateRequest{
		UserID:   &userID,
		Type:     notifType,
		Title:    title,
		Message:  message,
		Priority: priority,
		Data:     data,
	}

	return s.CreateNotification(req, nil)
}

// BroadcastNotification broadcasts a notification to all users or specific groups
func (s *NotificationService) BroadcastNotification(req *models.NotificationBroadcastRequest, senderID uuid.UUID) (int64, error) {
	// If target roles or companies specified, get specific users
	// For now, we'll broadcast to all connected users
	// TODO: Filter by roles and companies

	// Create notification without user_id (broadcast)
	notification, err := s.repo.CreateNotification(&models.NotificationCreateRequest{
		UserID:     nil, // nil means broadcast
		Type:       req.Type,
		Title:      req.Title,
		Message:    req.Message,
		Priority:   req.Priority,
		Data:       req.Data,
		ActionURL:  req.ActionURL,
		ActionText: req.ActionText,
		ExpiresAt:  req.ExpiresAt,
	})

	if err != nil {
		return 0, err
	}

	// Broadcast via WebSocket
	s.hub.BroadcastNotification(notification)

	// Return count of connected users
	return int64(s.hub.GetTotalConnectionCount()), nil
}

// ListNotifications retrieves notifications for a user with filtering
func (s *NotificationService) ListNotifications(filter *models.NotificationFilter, requesterID uuid.UUID) (*models.NotificationListResponse, error) {
	// Ensure user can only see their own notifications (unless admin)
	// TODO: Check if requester is admin
	if filter.UserID == nil || *filter.UserID != requesterID {
		filter.UserID = &requesterID
	}

	notifications, total, err := s.repo.ListNotifications(filter)
	if err != nil {
		return nil, err
	}

	// Get unread count
	unreadCount, _ := s.repo.GetUnreadCount(requesterID)

	// Calculate total pages
	pageSize := filter.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		Page:          filter.Page,
		PageSize:      pageSize,
		TotalPages:    totalPages,
		UnreadCount:   unreadCount,
	}, nil
}

// GetNotification retrieves a single notification
func (s *NotificationService) GetNotification(notificationID uuid.UUID, requesterID uuid.UUID) (*models.Notification, error) {
	notification, err := s.repo.GetNotificationByID(notificationID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this notification
	if notification.UserID != nil && *notification.UserID != requesterID {
		// TODO: Check if requester is admin
		return nil, fmt.Errorf("access denied")
	}

	return notification, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID uuid.UUID, requesterID uuid.UUID) error {
	// Verify notification belongs to user
	notification, err := s.repo.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}

	if notification.UserID != nil && *notification.UserID != requesterID {
		return fmt.Errorf("access denied")
	}

	return s.repo.MarkAsRead(notificationID)
}

// MarkMultipleAsRead marks multiple notifications as read
func (s *NotificationService) MarkMultipleAsRead(req *models.NotificationMarkReadRequest, requesterID uuid.UUID) error {
	// TODO: Verify all notifications belong to user
	return s.repo.MarkMultipleAsRead(req.NotificationIDs)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	return s.repo.MarkAllAsReadForUser(userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID uuid.UUID, requesterID uuid.UUID) error {
	// Verify notification belongs to user
	notification, err := s.repo.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}

	if notification.UserID != nil && *notification.UserID != requesterID {
		return fmt.Errorf("access denied")
	}

	return s.repo.DeleteNotification(notificationID)
}

// DeleteMultipleNotifications deletes multiple notifications
func (s *NotificationService) DeleteMultipleNotifications(req *models.NotificationDeleteRequest, requesterID uuid.UUID) error {
	// TODO: Verify all notifications belong to user
	return s.repo.DeleteMultipleNotifications(req.NotificationIDs)
}

// GetUnreadCount gets unread notification count for a user
func (s *NotificationService) GetUnreadCount(userID uuid.UUID) (int64, error) {
	return s.repo.GetUnreadCount(userID)
}

// GetNotificationStats gets notification statistics
func (s *NotificationService) GetNotificationStats(userID *uuid.UUID, requesterID uuid.UUID) (*models.NotificationStats, error) {
	// If requesting stats for specific user, verify access
	if userID != nil && *userID != requesterID {
		// TODO: Check if requester is admin
		return nil, fmt.Errorf("access denied")
	}

	// If no user specified and not admin, return stats for requester
	if userID == nil {
		userID = &requesterID
	}

	return s.repo.GetNotificationStats(userID)
}

// GetPreference gets notification preferences for a user
func (s *NotificationService) GetPreference(userID uuid.UUID) (*models.NotificationPreference, error) {
	return s.repo.GetOrCreatePreference(userID)
}

// UpdatePreference updates notification preferences
func (s *NotificationService) UpdatePreference(userID uuid.UUID, req *models.NotificationPreferenceUpdateRequest) (*models.NotificationPreference, error) {
	return s.repo.UpdatePreference(userID, req)
}

// CleanupExpiredNotifications cleans up expired notifications
func (s *NotificationService) CleanupExpiredNotifications() (int64, error) {
	return s.repo.CleanupExpiredNotifications()
}

// NotifyUserCreated sends notification when a user is created
func (s *NotificationService) NotifyUserCreated(user *models.User, creatorID uuid.UUID) error {
	_, err := s.CreateNotificationForUser(
		user.ID,
		models.NotificationTypeUserCreated,
		"Welcome to the System",
		fmt.Sprintf("Your account has been created successfully. Welcome %s!", user.FirstName),
		models.NotificationPriorityNormal,
		map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	)
	return err
}

// NotifyUserUpdated sends notification when a user is updated
func (s *NotificationService) NotifyUserUpdated(userID uuid.UUID, changes []string) error {
	_, err := s.CreateNotificationForUser(
		userID,
		models.NotificationTypeUserUpdated,
		"Profile Updated",
		fmt.Sprintf("Your profile has been updated. Changes: %v", changes),
		models.NotificationPriorityNormal,
		map[string]interface{}{
			"user_id": userID,
			"changes": changes,
		},
	)
	return err
}

// NotifyPasswordChanged sends notification when password is changed
func (s *NotificationService) NotifyPasswordChanged(userID uuid.UUID) error {
	_, err := s.CreateNotificationForUser(
		userID,
		models.NotificationTypeSecurityPasswordChange,
		"Password Changed",
		"Your password has been changed successfully. If you didn't make this change, please contact support immediately.",
		models.NotificationPriorityHigh,
		map[string]interface{}{
			"user_id": userID,
		},
	)
	return err
}

// NotifyLoginFailed sends notification when login fails
func (s *NotificationService) NotifyLoginFailed(userID uuid.UUID, ipAddress string) error {
	_, err := s.CreateNotificationForUser(
		userID,
		models.NotificationTypeSecurityLoginFailed,
		"Failed Login Attempt",
		fmt.Sprintf("A failed login attempt was detected from IP: %s", ipAddress),
		models.NotificationPriorityHigh,
		map[string]interface{}{
			"user_id":    userID,
			"ip_address": ipAddress,
		},
	)
	return err
}

// NotifyAccountLocked sends notification when account is locked
func (s *NotificationService) NotifyAccountLocked(userID uuid.UUID) error {
	_, err := s.CreateNotificationForUser(
		userID,
		models.NotificationTypeSecurityAccountLocked,
		"Account Locked",
		"Your account has been locked due to multiple failed login attempts. Please contact support to unlock.",
		models.NotificationPriorityCritical,
		map[string]interface{}{
			"user_id": userID,
		},
	)
	return err
}

// Notify2FAEnabled sends notification when 2FA is enabled
func (s *NotificationService) Notify2FAEnabled(userID uuid.UUID) error {
	_, err := s.CreateNotificationForUser(
		userID,
		models.NotificationTypeSecurity2FAEnabled,
		"Two-Factor Authentication Enabled",
		"Two-factor authentication has been enabled on your account for enhanced security.",
		models.NotificationPriorityNormal,
		map[string]interface{}{
			"user_id": userID,
		},
	)
	return err
}

// NotifyCompanyCreated sends notification about new company
func (s *NotificationService) NotifyCompanyCreated(companyID uuid.UUID, companyName string, adminIDs []uuid.UUID) error {
	for _, adminID := range adminIDs {
		s.CreateNotificationForUser(
			adminID,
			models.NotificationTypeCompanyCreated,
			"New Company Created",
			fmt.Sprintf("Company '%s' has been created successfully.", companyName),
			models.NotificationPriorityNormal,
			map[string]interface{}{
				"company_id":   companyID,
				"company_name": companyName,
			},
		)
	}
	return nil
}

// NotifyCompanyUserAdded sends notification when user is added to company
func (s *NotificationService) NotifyCompanyUserAdded(userID, companyID uuid.UUID, companyName string, role string) error {
	_, err := s.CreateNotificationForUser(
		userID,
		models.NotificationTypeCompanyUserAdded,
		"Added to Company",
		fmt.Sprintf("You have been added to '%s' with role: %s", companyName, role),
		models.NotificationPriorityNormal,
		map[string]interface{}{
			"user_id":      userID,
			"company_id":   companyID,
			"company_name": companyName,
			"role":         role,
		},
	)
	return err
}

// NotifySessionExpired sends notification when session expires
func (s *NotificationService) NotifySessionExpired(userID uuid.UUID) error {
	// Only send if user is connected
	if s.hub.IsUserConnected(userID) {
		_, err := s.CreateNotificationForUser(
			userID,
			models.NotificationTypeSessionExpired,
			"Session Expired",
			"Your session has expired. Please log in again.",
			models.NotificationPriorityNormal,
			map[string]interface{}{
				"user_id": userID,
			},
		)
		return err
	}
	return nil
}

// SendCustomNotification sends a custom notification
func (s *NotificationService) SendCustomNotification(
	userID uuid.UUID,
	title, message string,
	priority models.NotificationPriority,
	data map[string]interface{},
) error {
	_, err := s.CreateNotificationForUser(
		userID,
		models.NotificationTypeSystemAlert,
		title,
		message,
		priority,
		data,
	)
	return err
}

// GetWebSocketHub returns the WebSocket hub (for handler to upgrade connections)
func (s *NotificationService) GetWebSocketHub() *WebSocketHub {
	return s.hub
}

// GetConnectedUsers returns list of connected users
func (s *NotificationService) GetConnectedUsers() []models.WebSocketConnectionInfo {
	return s.hub.GetConnectedUsers()
}

// DisconnectUser disconnects a specific user
func (s *NotificationService) DisconnectUser(userID uuid.UUID) {
	s.hub.DisconnectUser(userID)
}

// ScheduleCleanup schedules periodic cleanup of expired notifications
func (s *NotificationService) ScheduleCleanup() {
	// Run cleanup every hour
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			count, err := s.CleanupExpiredNotifications()
			if err == nil && count > 0 {
				// Log cleanup
				fmt.Printf("Cleaned up %d expired notifications\n", count)
			}
		}
	}()
}

// ShouldSendNotification checks if notification should be sent based on user preferences
func (s *NotificationService) ShouldSendNotification(
	userID uuid.UUID,
	notifType models.NotificationType,
	priority models.NotificationPriority,
) (bool, error) {
	pref, err := s.repo.GetOrCreatePreference(userID)
	if err != nil {
		return false, err
	}

	// Check if notification type is enabled (empty array means all types enabled)
	if len(pref.EnabledTypes) > 0 {
		found := false
		for _, enabledType := range pref.EnabledTypes {
			if enabledType == notifType {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	// Check minimum priority
	priorityOrder := map[models.NotificationPriority]int{
		models.NotificationPriorityLow:      1,
		models.NotificationPriorityNormal:   2,
		models.NotificationPriorityHigh:     3,
		models.NotificationPriorityCritical: 4,
	}

	if priorityOrder[priority] < priorityOrder[pref.MinPriority] {
		return false, nil
	}

	// Check quiet hours
	if pref.QuietHoursStart != nil && pref.QuietHoursEnd != nil {
		now := time.Now()
		startTime := pref.QuietHoursStart.Hour()*60 + pref.QuietHoursStart.Minute()
		endTime := pref.QuietHoursEnd.Hour()*60 + pref.QuietHoursEnd.Minute()
		currentTime := now.Hour()*60 + now.Minute()

		if startTime <= currentTime && currentTime <= endTime {
			// During quiet hours, only send critical notifications
			if priority != models.NotificationPriorityCritical {
				return false, nil
			}
		}
	}

	return true, nil
}
