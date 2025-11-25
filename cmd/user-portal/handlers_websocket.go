package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/notifications"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now, should be restricted in production
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WSMessage represents a generic WebSocket message
type WSMessage struct {
	Type      string          `json:"type"`
	MeetingID string          `json:"meeting_id,omitempty"`
	UserID    string          `json:"user_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// WSClient represents a connected WebSocket client
type WSClient struct {
	ID        string
	UserID    string
	Conn      *websocket.Conn
	Send      chan WSMessage
	MeetingID string
	Portal    *UserPortal
}

// WSHub manages all WebSocket connections
type WSHub struct {
	// Registered clients by meeting ID
	clients map[string]map[*WSClient]bool

	// Register requests from clients
	register chan *WSClient

	// Unregister requests from clients
	unregister chan *WSClient

	// Broadcast messages to meeting participants
	broadcast chan WSMessage

	mu sync.RWMutex
}

// NewWSHub creates a new WebSocket hub
func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[string]map[*WSClient]bool),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan WSMessage, 256),
	}
}

// Run starts the WebSocket hub
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.MeetingID] == nil {
				h.clients[client.MeetingID] = make(map[*WSClient]bool)
			}
			h.clients[client.MeetingID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.MeetingID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.MeetingID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// ВАЖНО: Исправление гонки данных (race condition)
			// Нельзя изменять map под read lock - нужно собрать клиенты для удаления,
			// затем освободить read lock и удалить их под write lock
			h.mu.RLock()
			meetingID := message.MeetingID
			var failedClients []*WSClient
			if clients, ok := h.clients[meetingID]; ok {
				for client := range clients {
					select {
					case client.Send <- message:
						// Успешно отправлено
					default:
						// Канал заблокирован - клиент не успевает обрабатывать сообщения
						// Собираем его для удаления
						failedClients = append(failedClients, client)
					}
				}
			}
			h.mu.RUnlock()

			// Очищаем неудачные клиенты под write lock
			// CRITICAL: Это должно быть ПОСЛЕ RUnlock(), иначе возникает гонка данных
			if len(failedClients) > 0 {
				h.mu.Lock()
				for _, client := range failedClients {
					if clients, ok := h.clients[meetingID]; ok {
						if _, ok := clients[client]; ok {
							close(client.Send)
							delete(clients, client)
						}
					}
				}
				h.mu.Unlock()
			}
		}
	}
}

// BroadcastToMeeting sends a message to all clients in a meeting
func (h *WSHub) BroadcastToMeeting(meetingID string, msgType string, data interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	message := WSMessage{
		Type:      msgType,
		MeetingID: meetingID,
		Data:      dataJSON,
		Timestamp: time.Now(),
	}

	h.broadcast <- message
	return nil
}

// GetClientsInMeeting returns count of clients in a meeting
func (h *WSHub) GetClientsInMeeting(meetingID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[meetingID]; ok {
		return len(clients)
	}
	return 0
}

// ReadPump reads messages from the WebSocket connection
func (c *WSClient) ReadPump() {
	defer func() {
		c.Portal.wsHub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg WSMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Portal.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		msg.UserID = c.UserID
		msg.MeetingID = c.MeetingID
		msg.Timestamp = time.Now()

		// Handle different message types
		c.handleMessage(msg)
	}
}

// WritePump writes messages to the WebSocket connection
func (c *WSClient) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteJSON(message)
			if err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *WSClient) handleMessage(msg WSMessage) {
	switch msg.Type {
	case "screen_share_start":
		// Broadcast to all participants that screen sharing started
		c.Portal.wsHub.BroadcastToMeeting(c.MeetingID, "screen_share_started", map[string]string{
			"user_id": c.UserID,
		})

	case "screen_share_stop":
		// Broadcast to all participants that screen sharing stopped
		c.Portal.wsHub.BroadcastToMeeting(c.MeetingID, "screen_share_stopped", map[string]string{
			"user_id": c.UserID,
		})

	case "ping":
		// Respond with pong
		c.Send <- WSMessage{
			Type:      "pong",
			Timestamp: time.Now(),
		}

	default:
		c.Portal.logger.Info("Unknown WebSocket message type: " + msg.Type)
	}
}

// handleWebSocket handles WebSocket connections for meetings
// @Summary WebSocket endpoint for meeting real-time communication
// @Description Establishes WebSocket connection for real-time meeting events
// @Tags meetings
// @Accept json
// @Produce json
// @Param meeting_id path string true "Meeting ID"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BearerAuth
// @Router /api/v1/meetings/{meeting_id}/ws [get]
func (up *UserPortal) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract user claims from context (set by auth middleware)
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.logger.Error("User claims not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID.String()

	// Extract meeting ID from path: /api/v1/meetings/{meetingId}/ws
	pathSuffix := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	segments := strings.Split(pathSuffix, "/")
	if len(segments) < 2 || segments[0] == "" {
		up.logger.Error("Meeting ID not provided")
		http.Error(w, "Meeting ID required", http.StatusBadRequest)
		return
	}
	meetingID := segments[0]

	// Verify meeting exists
	meetingUUID, err := uuid.Parse(meetingID)
	if err != nil {
		up.logger.Error("Invalid meeting ID format: " + err.Error())
		http.Error(w, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	_, err = up.meetingRepo.GetMeetingByID(meetingUUID)
	if err != nil {
		up.logger.Error("Failed to get meeting: " + err.Error())
		http.Error(w, "Meeting not found", http.StatusNotFound)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		up.logger.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := &WSClient{
		ID:        r.Header.Get("Sec-WebSocket-Key"),
		UserID:    userID,
		Conn:      conn,
		Send:      make(chan WSMessage, 256),
		MeetingID: meetingID,
		Portal:    up,
	}

	// Register client
	up.wsHub.register <- client

	up.logger.Infof("WebSocket client connected: user=%s, meeting=%s", client.UserID, meetingID)

	// Send welcome message
	client.Send <- WSMessage{
		Type:      "connected",
		Data:      json.RawMessage(`{"message":"Connected to meeting"}`),
		Timestamp: time.Now(),
	}

	// Start pumps in goroutines
	go client.WritePump()
	go client.ReadPump()
}

// NotificationClient represents a WebSocket client for user notifications
type NotificationClient struct {
	ID             string
	UserID         uuid.UUID
	Conn           *websocket.Conn
	Send           chan *notifications.Notification
	Portal         *UserPortal
	Subscriber     *notifications.Subscriber
}

// handleNotificationsWebSocket handles WebSocket connections for real-time notifications
// @Summary WebSocket endpoint for real-time notifications
// @Description Establishes WebSocket connection for receiving real-time updates about meetings, recordings, transcriptions, etc.
// @Tags notifications
// @Accept json
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BearerAuth
// @Router /api/v1/notifications/ws [get]
func (up *UserPortal) handleNotificationsWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract user claims from context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.logger.Error("User claims not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		up.logger.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	// Subscribe to notifications
	subscriber := up.notificationService.Subscribe(claims.UserID, notifications.SubscriptionFilters{
		// Subscribe to all events by default - client can filter
	})

	// Create notification client
	client := &NotificationClient{
		ID:         uuid.New().String(),
		UserID:     claims.UserID,
		Conn:       conn,
		Send:       make(chan *notifications.Notification, 256),
		Portal:     up,
		Subscriber: subscriber,
	}

	up.logger.Infof("Notification WebSocket client connected: user=%s", claims.UserID)

	// Send welcome message
	welcomeNotification := &notifications.Notification{
		ID:        uuid.New().String(),
		Type:      "system.connected",
		Timestamp: time.Now(),
		Message:   "Connected to notification service",
	}
	client.Send <- welcomeNotification

	// Start pumps
	go client.writeNotificationPump()
	go client.readNotificationPump()

	// Forward notifications from service to client
	go client.forwardNotifications()
}

// writeNotificationPump writes notifications to the WebSocket connection
func (nc *NotificationClient) writeNotificationPump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		nc.Conn.Close()
	}()

	for {
		select {
		case notification, ok := <-nc.Send:
			nc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				nc.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := nc.Conn.WriteJSON(notification)
			if err != nil {
				return
			}

		case <-ticker.C:
			nc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := nc.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readNotificationPump reads messages from the WebSocket connection
func (nc *NotificationClient) readNotificationPump() {
	defer func() {
		nc.Portal.notificationService.Unsubscribe(nc.Subscriber.ID)
		nc.Conn.Close()
	}()

	nc.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	nc.Conn.SetPongHandler(func(string) error {
		nc.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg map[string]interface{}
		err := nc.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				nc.Portal.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Handle ping/pong and subscription updates
		if msgType, ok := msg["type"].(string); ok {
			if msgType == "ping" {
				pong := &notifications.Notification{
					ID:        uuid.New().String(),
					Type:      "system.pong",
					Timestamp: time.Now(),
				}
				nc.Send <- pong
			}
		}
	}
}

// forwardNotifications forwards notifications from the service to the WebSocket client
func (nc *NotificationClient) forwardNotifications() {
	for notification := range nc.Subscriber.Channel {
		select {
		case nc.Send <- notification:
			// Successfully forwarded
		default:
			// Client is slow, drop notification
			nc.Portal.logger.Error("Dropped notification for slow client: user=" + nc.UserID.String())
		}
	}
}
