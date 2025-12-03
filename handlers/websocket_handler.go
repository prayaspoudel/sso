package handlers

import (
	"fmt"
	"net/http"
	"time"

	"sso/models"
	"sso/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	notificationService *services.NotificationService
	upgrader            websocket.Upgrader
}

func NewWebSocketHandler(notificationService *services.NotificationService) *WebSocketHandler {
	return &WebSocketHandler{
		notificationService: notificationService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking
				return true
			},
		},
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user from context (must be authenticated)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user email
	userEmailInterface, _ := c.Get("userEmail")
	userEmail, _ := userEmailInterface.(string)

	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Create WebSocket client
	client := &services.WebSocketClient{
		ID:            uuid.New(),
		UserID:        userID,
		UserEmail:     userEmail,
		Connection:    conn,
		Send:          make(chan []byte, 256),
		Hub:           h.notificationService.GetWebSocketHub(),
		IPAddress:     c.ClientIP(),
		ConnectedAt:   time.Now(),
		LastHeartbeat: time.Now(),
	}

	// Register client with hub
	client.Hub.Register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()
}

// GetUserIDFromContext is a helper to get user ID from gin context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return userID, true
}

// ListNotifications retrieves notifications for the authenticated user
func (h *WebSocketHandler) ListNotifications(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var filter models.NotificationFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure user can only see their own notifications
	filter.UserID = &userID

	response, err := h.notificationService.ListNotifications(&filter, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNotification retrieves a single notification
func (h *WebSocketHandler) GetNotification(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	notification, err := h.notificationService.GetNotification(notificationID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// CreateNotification creates a new notification (admin only)
func (h *WebSocketHandler) CreateNotification(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	var req models.NotificationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Check if user is admin

	notification, err := h.notificationService.CreateNotification(&req, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// BroadcastNotification broadcasts a notification to all users or specific groups (admin only)
func (h *WebSocketHandler) BroadcastNotification(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	var req models.NotificationBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Check if user is admin/super_admin

	count, err := h.notificationService.BroadcastNotification(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to broadcast notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Notification broadcast successfully",
		"recipients_count": count,
	})
}

// MarkAsRead marks a notification as read
func (h *WebSocketHandler) MarkAsRead(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.MarkAsRead(notificationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkMultipleAsRead marks multiple notifications as read
func (h *WebSocketHandler) MarkMultipleAsRead(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	var req models.NotificationMarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.notificationService.MarkMultipleAsRead(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications marked as read"})
}

// MarkAllAsRead marks all notifications as read for the authenticated user
func (h *WebSocketHandler) MarkAllAsRead(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	err := h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification deletes a notification
func (h *WebSocketHandler) DeleteNotification(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.DeleteNotification(notificationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// DeleteMultipleNotifications deletes multiple notifications
func (h *WebSocketHandler) DeleteMultipleNotifications(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	var req models.NotificationDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.notificationService.DeleteMultipleNotifications(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications deleted successfully"})
}

// GetUnreadCount gets the count of unread notifications
func (h *WebSocketHandler) GetUnreadCount(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// GetNotificationStats gets notification statistics
func (h *WebSocketHandler) GetNotificationStats(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	// Check if requesting stats for specific user
	userIDStr := c.Query("user_id")
	var targetUserID *uuid.UUID
	if userIDStr != "" {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		targetUserID = &parsedUserID
	}

	stats, err := h.notificationService.GetNotificationStats(targetUserID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetPreference gets notification preferences for the authenticated user
func (h *WebSocketHandler) GetPreference(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	preference, err := h.notificationService.GetPreference(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification preferences"})
		return
	}

	c.JSON(http.StatusOK, preference)
}

// UpdatePreference updates notification preferences
func (h *WebSocketHandler) UpdatePreference(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	var req models.NotificationPreferenceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preference, err := h.notificationService.UpdatePreference(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification preferences"})
		return
	}

	c.JSON(http.StatusOK, preference)
}

// GetConnectedUsers gets list of currently connected users (admin only)
func (h *WebSocketHandler) GetConnectedUsers(c *gin.Context) {
	// TODO: Check if user is admin

	connections := h.notificationService.GetConnectedUsers()

	c.JSON(http.StatusOK, gin.H{
		"connections":       connections,
		"total_connections": len(connections),
		"unique_users":      h.notificationService.GetWebSocketHub().GetConnectedUserCount(),
	})
}

// DisconnectUser disconnects a specific user (admin only)
func (h *WebSocketHandler) DisconnectUser(c *gin.Context) {
	// TODO: Check if user is admin

	userIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	h.notificationService.DisconnectUser(targetUserID)

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User %s disconnected", targetUserID)})
}

// SendTestNotification sends a test notification (for testing purposes)
func (h *WebSocketHandler) SendTestNotification(c *gin.Context) {
	userID, _ := getUserIDFromContext(c)

	err := h.notificationService.SendCustomNotification(
		userID,
		"Test Notification",
		"This is a test notification sent at "+time.Now().Format(time.RFC3339),
		models.NotificationPriorityNormal,
		map[string]interface{}{
			"test":      true,
			"timestamp": time.Now(),
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send test notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test notification sent successfully"})
}
