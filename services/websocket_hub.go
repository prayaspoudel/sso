package services

import (
	"encoding/json"
	"sync"
	"time"

	"sso/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	UserEmail     string
	Connection    *websocket.Conn
	Send          chan []byte
	Hub           *WebSocketHub
	IPAddress     string
	ConnectedAt   time.Time
	LastHeartbeat time.Time
	mu            sync.Mutex
}

// WebSocketHub maintains the set of active clients and broadcasts messages to clients
type WebSocketHub struct {
	// Registered clients mapped by user ID
	clients map[uuid.UUID]map[*WebSocketClient]bool

	// Broadcast channel for all clients
	broadcast chan []byte

	// Targeted message channel (user ID -> message)
	targeted chan *TargetedMessage

	// Register requests from clients
	Register chan *WebSocketClient

	// Unregister requests from clients
	Unregister chan *WebSocketClient

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// TargetedMessage represents a message targeted to specific user(s)
type TargetedMessage struct {
	UserIDs []uuid.UUID
	Message []byte
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[uuid.UUID]map[*WebSocketClient]bool),
		broadcast:  make(chan []byte, 256),
		targeted:   make(chan *TargetedMessage, 256),
		Register:   make(chan *WebSocketClient),
		Unregister: make(chan *WebSocketClient),
	}
}

// Run starts the hub's main loop
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case targeted := <-h.targeted:
			h.sendTargeted(targeted)
		}
	}
}

// registerClient registers a new client
func (h *WebSocketHub) registerClient(client *WebSocketClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*WebSocketClient]bool)
	}
	h.clients[client.UserID][client] = true
}

// unregisterClient unregisters a client
func (h *WebSocketHub) unregisterClient(client *WebSocketClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.UserID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)

			// Remove user entry if no more clients
			if len(clients) == 0 {
				delete(h.clients, client.UserID)
			}
		}
	}
}

// broadcastToAll sends a message to all connected clients
func (h *WebSocketHub) broadcastToAll(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				// Client's send buffer is full, close the connection
				go func(c *WebSocketClient) {
					h.Unregister <- c
				}(client)
			}
		}
	}
}

// sendTargeted sends a message to specific users
func (h *WebSocketHub) sendTargeted(targeted *TargetedMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userID := range targeted.UserIDs {
		if clients, ok := h.clients[userID]; ok {
			for client := range clients {
				select {
				case client.Send <- targeted.Message:
				default:
					// Client's send buffer is full, close the connection
					go func(c *WebSocketClient) {
						h.Unregister <- c
					}(client)
				}
			}
		}
	}
}

// BroadcastNotification broadcasts a notification to all connected clients
func (h *WebSocketHub) BroadcastNotification(notification *models.Notification) error {
	message := models.WebSocketMessage{
		Type:         "notification",
		Notification: notification,
		Timestamp:    time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- messageBytes
	return nil
}

// SendNotificationToUser sends a notification to a specific user
func (h *WebSocketHub) SendNotificationToUser(userID uuid.UUID, notification *models.Notification) error {
	message := models.WebSocketMessage{
		Type:         "notification",
		Notification: notification,
		Timestamp:    time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.targeted <- &TargetedMessage{
		UserIDs: []uuid.UUID{userID},
		Message: messageBytes,
	}

	return nil
}

// SendNotificationToUsers sends a notification to multiple users
func (h *WebSocketHub) SendNotificationToUsers(userIDs []uuid.UUID, notification *models.Notification) error {
	message := models.WebSocketMessage{
		Type:         "notification",
		Notification: notification,
		Timestamp:    time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.targeted <- &TargetedMessage{
		UserIDs: userIDs,
		Message: messageBytes,
	}

	return nil
}

// SendMessageToUser sends a custom message to a user
func (h *WebSocketHub) SendMessageToUser(userID uuid.UUID, messageType string, data map[string]interface{}) error {
	message := models.WebSocketMessage{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.targeted <- &TargetedMessage{
		UserIDs: []uuid.UUID{userID},
		Message: messageBytes,
	}

	return nil
}

// GetConnectedUserCount returns the number of connected users
func (h *WebSocketHub) GetConnectedUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetTotalConnectionCount returns the total number of connections
func (h *WebSocketHub) GetTotalConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// GetConnectedUsers returns a list of connected users
func (h *WebSocketHub) GetConnectedUsers() []models.WebSocketConnectionInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var connections []models.WebSocketConnectionInfo

	for userID, clients := range h.clients {
		for client := range clients {
			connections = append(connections, models.WebSocketConnectionInfo{
				UserID:        userID,
				UserEmail:     client.UserEmail,
				ConnectedAt:   client.ConnectedAt,
				LastHeartbeat: client.LastHeartbeat,
				IPAddress:     client.IPAddress,
			})
		}
	}

	return connections
}

// IsUserConnected checks if a user is currently connected
func (h *WebSocketHub) IsUserConnected(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// DisconnectUser disconnects all connections for a user
func (h *WebSocketHub) DisconnectUser(userID uuid.UUID) {
	h.mu.RLock()
	clients, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		for client := range clients {
			h.Unregister <- client
		}
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Connection.Close()
	}()

	c.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Connection.SetPongHandler(func(string) error {
		c.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.mu.Lock()
		c.LastHeartbeat = time.Now()
		c.mu.Unlock()
		return nil
	})

	for {
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close error
			}
			break
		}

		// Handle incoming messages (e.g., mark notification as read, send acknowledgment)
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *WebSocketClient) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages from the client
func (c *WebSocketClient) handleMessage(message []byte) {
	var wsMsg models.WebSocketMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		// Invalid message format
		return
	}

	switch wsMsg.Type {
	case "pong":
		// Update last heartbeat
		c.mu.Lock()
		c.LastHeartbeat = time.Now()
		c.mu.Unlock()

	case "mark_read":
		// Handle marking notification as read
		// This would need to call the notification service

	case "typing":
		// Handle typing indicator (if implementing chat)

	default:
		// Unknown message type
	}
}
