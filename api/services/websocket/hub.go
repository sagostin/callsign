package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/neffos"
	log "github.com/sirupsen/logrus"
)

// EventType represents the type of real-time event
type EventType string

const (
	EventPresence     EventType = "presence"
	EventConference   EventType = "conference"
	EventCall         EventType = "call"
	EventQueue        EventType = "queue"
	EventStats        EventType = "stats"
	EventNotification EventType = "notification"
	EventSMS          EventType = "sms"
	EventVoicemail    EventType = "voicemail"
	EventRegistration EventType = "registration"
)

// Event represents a real-time event
type Event struct {
	Type      EventType              `json:"type"`
	Action    string                 `json:"action"`
	TenantID  uint                   `json:"tenant_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	ID         string
	TenantID   uint
	UserID     uint
	Connection *neffos.Conn
	Topics     []string // Subscribed topics
}

// Hub manages WebSocket clients and broadcasting
type Hub struct {
	clients    map[string]*Client
	tenants    map[uint][]*Client // Clients grouped by tenant
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		tenants:    make(map[uint][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Event, 100),
	}
}

// Run starts the hub main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)

		case client := <-h.unregister:
			h.removeClient(client)

		case event := <-h.broadcast:
			h.broadcastEvent(event)
		}
	}
}

func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client
	h.tenants[client.TenantID] = append(h.tenants[client.TenantID], client)

	log.WithFields(log.Fields{
		"client_id": client.ID,
		"tenant_id": client.TenantID,
		"user_id":   client.UserID,
	}).Info("WebSocket client connected")
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client.ID)

	// Remove from tenant list
	if clients, ok := h.tenants[client.TenantID]; ok {
		for i, c := range clients {
			if c.ID == client.ID {
				h.tenants[client.TenantID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
	}

	log.WithFields(log.Fields{
		"client_id": client.ID,
	}).Info("WebSocket client disconnected")
}

func (h *Hub) broadcastEvent(event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(event)
	if err != nil {
		log.Errorf("Failed to marshal event: %v", err)
		return
	}

	msg := neffos.Message{
		Namespace: "default",
		Event:     string(event.Type),
		Body:      data,
	}

	// If tenant-specific, only send to that tenant's clients
	if event.TenantID > 0 {
		if clients, ok := h.tenants[event.TenantID]; ok {
			for _, client := range clients {
				if client.Connection != nil {
					client.Connection.Write(msg)
				}
			}
		}
	} else {
		// Broadcast to all clients
		for _, client := range h.clients {
			if client.Connection != nil {
				client.Connection.Write(msg)
			}
		}
	}
}

// Broadcast sends an event to relevant clients
func (h *Hub) Broadcast(event *Event) {
	event.Timestamp = time.Now()
	h.broadcast <- event
}

// BroadcastToTenant sends an event to all clients of a tenant
func (h *Hub) BroadcastToTenant(tenantID uint, eventType EventType, action string, data map[string]interface{}) {
	h.Broadcast(&Event{
		Type:     eventType,
		Action:   action,
		TenantID: tenantID,
		Data:     data,
	})
}

// NotifyPresenceChange notifies about BLF/presence updates
func (h *Hub) NotifyPresenceChange(tenantID uint, extension, state string) {
	h.BroadcastToTenant(tenantID, EventPresence, "update", map[string]interface{}{
		"extension": extension,
		"state":     state,
	})
}

// NotifyCallEvent notifies about call state changes
func (h *Hub) NotifyCallEvent(tenantID uint, action string, callData map[string]interface{}) {
	h.BroadcastToTenant(tenantID, EventCall, action, callData)
}

// NotifyConferenceEvent notifies about conference changes
func (h *Hub) NotifyConferenceEvent(tenantID uint, action string, confData map[string]interface{}) {
	h.BroadcastToTenant(tenantID, EventConference, action, confData)
}

// NotifyStats broadcasts statistics update
func (h *Hub) NotifyStats(tenantID uint, stats map[string]interface{}) {
	h.BroadcastToTenant(tenantID, EventStats, "update", stats)
}

// NotifySMS notifies about incoming SMS
func (h *Hub) NotifySMS(tenantID uint, smsData map[string]interface{}) {
	h.BroadcastToTenant(tenantID, EventSMS, "incoming", smsData)
}

// Handler handles WebSocket connections
type Handler struct {
	Hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{Hub: hub}
}

// HandleConnection handles new WebSocket connections using Neffos
// This is a simplified implementation - use proper Neffos server setup in production
func (h *Handler) HandleConnection(ctx iris.Context) {
	// For now, just return an error - proper WebSocket setup requires neffos server
	ctx.StatusCode(501)
	ctx.JSON(map[string]string{"error": "WebSocket requires neffos server configuration"})
}

// GetNeffosEvents returns the events map for Neffos server configuration
func (h *Handler) GetNeffosEvents() neffos.Namespaces {
	return neffos.Namespaces{
		"default": neffos.Events{
			neffos.OnNamespaceConnected: func(c *neffos.NSConn, msg neffos.Message) error {
				tenantID := c.Conn.Get("tenant_id")
				userID := c.Conn.Get("user_id")

				client := &Client{
					ID:         c.Conn.ID(),
					TenantID:   tenantID.(uint),
					UserID:     userID.(uint),
					Connection: c.Conn,
				}
				h.Hub.register <- client
				return nil
			},
			neffos.OnNamespaceDisconnect: func(c *neffos.NSConn, msg neffos.Message) error {
				h.Hub.unregister <- &Client{ID: c.Conn.ID()}
				return nil
			},
			"subscribe": func(c *neffos.NSConn, msg neffos.Message) error {
				log.Infof("Client %s subscribed to: %s", c.Conn.ID(), string(msg.Body))
				return nil
			},
			"ping": func(c *neffos.NSConn, msg neffos.Message) error {
				c.Emit("pong", []byte(`{"type":"pong"}`))
				return nil
			},
		},
	}
}
