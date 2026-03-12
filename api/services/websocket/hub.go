package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
	EventChat         EventType = "chat"
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
	ID       string
	TenantID uint
	UserID   uint
	Conn     *websocket.Conn
	Topics   []string // Subscribed topics
	mu       sync.Mutex
}

// WriteJSON safely writes JSON to the client connection
func (cl *Client) WriteJSON(v interface{}) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.Conn.WriteJSON(v)
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

	if c, ok := h.clients[client.ID]; ok {
		c.Conn.Close()
	}
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

	// If tenant-specific, only send to that tenant's clients
	if event.TenantID > 0 {
		if clients, ok := h.tenants[event.TenantID]; ok {
			for _, client := range clients {
				if client.Conn != nil {
					client.mu.Lock()
					err := client.Conn.WriteMessage(websocket.TextMessage, data)
					client.mu.Unlock()
					if err != nil {
						log.Warnf("Write to client %s failed: %v", client.ID, err)
					}
				}
			}
		}
	} else {
		// Broadcast to all clients
		for _, client := range h.clients {
			if client.Conn != nil {
				client.mu.Lock()
				err := client.Conn.WriteMessage(websocket.TextMessage, data)
				client.mu.Unlock()
				if err != nil {
					log.Warnf("Write to client %s failed: %v", client.ID, err)
				}
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

// NotifyChatMessage notifies about a new chat message
func (h *Hub) NotifyChatMessage(tenantID uint, threadID uint, msgData map[string]interface{}) {
	msgData["thread_id"] = threadID
	h.BroadcastToTenant(tenantID, EventChat, "new_message", msgData)
}

// NotifyMessageStatus notifies about a message delivery status change
func (h *Hub) NotifyMessageStatus(tenantID uint, messageID uint, status string) {
	h.BroadcastToTenant(tenantID, EventSMS, "status_update", map[string]interface{}{
		"message_id": messageID,
		"status":     status,
	})
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
