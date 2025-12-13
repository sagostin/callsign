package handlers

import (
	"encoding/json"
	"sync"
	"time"

	"callsign/config"
	"callsign/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
)

// NotificationMessage represents a notification sent to clients
type NotificationMessage struct {
	Type       string      `json:"type"`                 // notification type: call_incoming, voicemail_new, message_new, system_alert, etc.
	Title      string      `json:"title,omitempty"`      // notification title
	Message    string      `json:"message,omitempty"`    // notification message
	Timestamp  string      `json:"timestamp"`            // ISO timestamp
	Data       interface{} `json:"data,omitempty"`       // additional data (call_id, voicemail_id, etc.)
	Persistent bool        `json:"persistent,omitempty"` // if true, keep in notification history
	Token      string      `json:"token,omitempty"`      // for auth message only
}

// NotificationClient represents a connected notification WebSocket client
type NotificationClient struct {
	conn          *websocket.Conn
	send          chan NotificationMessage
	manager       *NotificationManager
	authenticated bool
	userID        uint
	tenantID      uint
	role          string
	mu            sync.Mutex
}

// NotificationManager manages WebSocket connections for notifications
type NotificationManager struct {
	clients    map[*NotificationClient]bool
	broadcast  chan NotificationMessage
	register   chan *NotificationClient
	unregister chan *NotificationClient
	config     *config.Config
	mu         sync.RWMutex
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(cfg *config.Config) *NotificationManager {
	return &NotificationManager{
		clients:    make(map[*NotificationClient]bool),
		broadcast:  make(chan NotificationMessage, 100),
		register:   make(chan *NotificationClient),
		unregister: make(chan *NotificationClient),
		config:     cfg,
	}
}

// Run starts the notification manager
func (m *NotificationManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()
			log.Debugf("Notification client connected (pending auth). Total: %d", len(m.clients))

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
			}
			m.mu.Unlock()
			log.Debugf("Notification client disconnected. Total: %d", len(m.clients))

		case msg := <-m.broadcast:
			m.mu.RLock()
			for client := range m.clients {
				// Only send to authenticated clients
				if !client.authenticated {
					continue
				}
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(m.clients, client)
				}
			}
			m.mu.RUnlock()
		}
	}
}

// Broadcast sends a notification to all authenticated clients
func (m *NotificationManager) Broadcast(msg NotificationMessage) {
	msg.Timestamp = time.Now().Format(time.RFC3339)
	select {
	case m.broadcast <- msg:
	default:
		log.Warn("Notification broadcast channel full")
	}
}

// BroadcastToTenant sends a notification to all clients of a specific tenant
func (m *NotificationManager) BroadcastToTenant(tenantID uint, msg NotificationMessage) {
	msg.Timestamp = time.Now().Format(time.RFC3339)
	m.mu.RLock()
	defer m.mu.RUnlock()

	for client := range m.clients {
		if client.authenticated && client.tenantID == tenantID {
			select {
			case client.send <- msg:
			default:
				// Skip slow clients
			}
		}
	}
}

// BroadcastToUser sends a notification to a specific user
func (m *NotificationManager) BroadcastToUser(userID uint, msg NotificationMessage) {
	msg.Timestamp = time.Now().Format(time.RFC3339)
	m.mu.RLock()
	defer m.mu.RUnlock()

	for client := range m.clients {
		if client.authenticated && client.userID == userID {
			select {
			case client.send <- msg:
			default:
				// Skip slow clients
			}
		}
	}
}

// ValidateToken validates a JWT token and returns claims
func (m *NotificationManager) ValidateToken(tokenString string) (*middleware.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &middleware.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*middleware.Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// NotificationWebSocket handles WebSocket connections for real-time notifications
func (h *Handler) NotificationWebSocket(ctx iris.Context) {
	w := ctx.ResponseWriter()
	r := ctx.Request()

	// Check for token in query param (fallback for initial connection)
	token := r.URL.Query().Get("token")

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create notification manager if not exists
	if h.NotificationManager == nil {
		h.NotificationManager = NewNotificationManager(h.Config)
		go h.NotificationManager.Run()
	}

	client := &NotificationClient{
		conn:          conn,
		send:          make(chan NotificationMessage, 64),
		manager:       h.NotificationManager,
		authenticated: false,
	}

	h.NotificationManager.register <- client

	// If token provided in query param, auto-authenticate
	if token != "" {
		claims, err := h.NotificationManager.ValidateToken(token)
		if err == nil {
			client.mu.Lock()
			client.authenticated = true
			client.userID = claims.UserID
			if claims.TenantID != nil {
				client.tenantID = *claims.TenantID
			}
			client.role = string(claims.Role)
			client.mu.Unlock()

			client.send <- NotificationMessage{
				Type:      "connected",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Connected to notification service",
			}
			log.Debugf("Notification client auto-authenticated: user=%d tenant=%d", claims.UserID, client.tenantID)
		}
	}

	// Send auth required if not authenticated
	if !client.authenticated {
		client.send <- NotificationMessage{
			Type:      "auth_required",
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "Send auth message with token to authenticate",
		}
	}

	// Start read/write goroutines
	go client.notificationWritePump()
	go client.notificationReadPump()
}

// notificationWritePump pumps messages to the WebSocket connection
func (c *NotificationClient) notificationWritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				log.Errorf("Failed to marshal notification: %v", err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// notificationReadPump pumps messages from the WebSocket connection
func (c *NotificationClient) notificationReadPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debugf("Notification WebSocket closed: %v", err)
			}
			break
		}

		var msg NotificationMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Warnf("Invalid notification message: %v", err)
			continue
		}

		// Handle authentication
		if msg.Type == "auth" && msg.Token != "" {
			claims, err := c.manager.ValidateToken(msg.Token)
			if err != nil {
				c.send <- NotificationMessage{
					Type:      "auth_error",
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   "Invalid or expired token",
				}
				c.conn.Close()
				return
			}

			c.mu.Lock()
			c.authenticated = true
			c.userID = claims.UserID
			if claims.TenantID != nil {
				c.tenantID = *claims.TenantID
			}
			c.role = string(claims.Role)
			c.mu.Unlock()

			c.send <- NotificationMessage{
				Type:      "connected",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Connected to notification service",
			}
			log.Debugf("Notification client authenticated: user=%d tenant=%d", claims.UserID, claims.TenantID)
			continue
		}

		// Handle ping (keep-alive from client)
		if msg.Type == "ping" {
			c.send <- NotificationMessage{
				Type:      "pong",
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}
	}
}

// SendCallNotification sends a call notification to relevant users
func (m *NotificationManager) SendCallNotification(tenantID uint, callType, callerID, callerNumber, callID string) {
	msg := NotificationMessage{
		Type:       callType, // call_incoming, call_missed
		Title:      "Incoming Call",
		Message:    "From: " + callerID,
		Persistent: true,
		Data: map[string]string{
			"call_id":       callID,
			"caller_id":     callerID,
			"caller_number": callerNumber,
		},
	}

	if callType == "call_missed" {
		msg.Title = "Missed Call"
	}

	m.BroadcastToTenant(tenantID, msg)
}

// SendVoicemailNotification sends a voicemail notification
func (m *NotificationManager) SendVoicemailNotification(userID, tenantID uint, callerID string, duration int, vmID, boxID uint) {
	msg := NotificationMessage{
		Type:       "voicemail_new",
		Title:      "New Voicemail",
		Message:    "From: " + callerID + " (" + string(rune(duration)) + "s)",
		Persistent: true,
		Data: map[string]interface{}{
			"voicemail_id": vmID,
			"box_id":       boxID,
			"caller_id":    callerID,
			"duration":     duration,
		},
	}
	m.BroadcastToUser(userID, msg)
}

// SendSystemAlert sends a system-wide alert
func (m *NotificationManager) SendSystemAlert(title, message string, persistent bool) {
	msg := NotificationMessage{
		Type:       "system_alert",
		Title:      title,
		Message:    message,
		Persistent: persistent,
	}
	m.Broadcast(msg)
}
