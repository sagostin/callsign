package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"callsign/middleware"
	"callsign/services/esl"

	"github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now (secured by JWT)
	},
}

// ConsoleMessage represents a message to/from the console WebSocket
type ConsoleMessage struct {
	Type      string `json:"type"`  // "log", "command", "response", "error", "status"
	Level     string `json:"level"` // DEBUG, INFO, WARNING, ERROR
	Timestamp string `json:"timestamp"`
	Module    string `json:"module"`
	Message   string `json:"message"`
	Command   string `json:"command"`
	Body      string `json:"body"`
}

// ConsoleClient represents a connected WebSocket client
type ConsoleClient struct {
	conn    *websocket.Conn
	send    chan ConsoleMessage
	manager *ConsoleManager
	mu      sync.Mutex
}

// ConsoleManager manages WebSocket connections for the FS console
type ConsoleManager struct {
	clients    map[*ConsoleClient]bool
	broadcast  chan ConsoleMessage
	register   chan *ConsoleClient
	unregister chan *ConsoleClient
	eslManager *esl.Manager
	mu         sync.RWMutex
}

// NewConsoleManager creates a new console manager
func NewConsoleManager(eslManager *esl.Manager) *ConsoleManager {
	return &ConsoleManager{
		clients:    make(map[*ConsoleClient]bool),
		broadcast:  make(chan ConsoleMessage, 100),
		register:   make(chan *ConsoleClient),
		unregister: make(chan *ConsoleClient),
		eslManager: eslManager,
	}
}

// Run starts the console manager
func (m *ConsoleManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()
			log.Infof("Console client connected. Total: %d", len(m.clients))

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
			}
			m.mu.Unlock()
			log.Infof("Console client disconnected. Total: %d", len(m.clients))

		case msg := <-m.broadcast:
			m.mu.RLock()
			for client := range m.clients {
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

// Broadcast sends a message to all connected clients
func (m *ConsoleManager) Broadcast(msg ConsoleMessage) {
	select {
	case m.broadcast <- msg:
	default:
		log.Warn("Console broadcast channel full")
	}
}

// ExecuteCommand executes a FreeSWITCH API command
func (m *ConsoleManager) ExecuteCommand(command string) (string, error) {
	if m.eslManager == nil || !m.eslManager.IsConnected() {
		return "", nil
	}
	return m.eslManager.API(command)
}

// FreeSwitchConsole handles WebSocket connections for live FS console
func (h *Handler) FreeSwitchConsole(ctx iris.Context) {
	// Verify user is system admin
	claims := ctx.Values().Get("claims").(*middleware.Claims)
	if claims.Role != "system_admin" {
		ctx.StatusCode(http.StatusForbidden)
		ctx.JSON(iris.Map{"error": "System admin access required"})
		return
	}

	// Get the underlying http.ResponseWriter and *http.Request
	w := ctx.ResponseWriter()
	r := ctx.Request()

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create console manager if not exists
	if h.ConsoleManager == nil {
		h.ConsoleManager = NewConsoleManager(h.ESLManager)
		go h.ConsoleManager.Run()
		go h.startLogStreaming()
	}

	client := &ConsoleClient{
		conn:    conn,
		send:    make(chan ConsoleMessage, 256),
		manager: h.ConsoleManager,
	}

	h.ConsoleManager.register <- client

	// Send connection status
	client.send <- ConsoleMessage{
		Type:      "status",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Connected to FreeSWITCH console",
	}

	// Start read/write goroutines
	go client.writePump()
	go client.readPump()
}

// startLogStreaming subscribes to FS log events and broadcasts them
func (h *Handler) startLogStreaming() {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		log.Warn("ESL not connected, log streaming not available")
		return
	}

	// Subscribe to log events
	err := h.ESLManager.SubscribeEvents("LOG", "CUSTOM", "CHANNEL_CREATE", "CHANNEL_DESTROY")
	if err != nil {
		log.Errorf("Failed to subscribe to log events: %v", err)
		return
	}

	log.Info("Started FreeSWITCH log streaming")

	// Process events
	events := h.ESLManager.Events()
	for ev := range events {
		eventName := ev.Get("Event-Name")

		var msg ConsoleMessage
		msg.Timestamp = time.Now().Format(time.RFC3339)
		msg.Type = "log"

		switch eventName {
		case "LOG":
			msg.Level = ev.Get("Log-Level")
			msg.Module = ev.Get("Log-File")
			msg.Message = ev.Body
			if msg.Level == "" {
				msg.Level = "INFO"
			}
		case "CHANNEL_CREATE":
			msg.Level = "INFO"
			msg.Module = "channel"
			msg.Message = "Channel created: " + ev.Get("Unique-ID") + " " + ev.Get("Caller-Caller-ID-Number")
		case "CHANNEL_DESTROY":
			msg.Level = "INFO"
			msg.Module = "channel"
			msg.Message = "Channel destroyed: " + ev.Get("Unique-ID")
		default:
			continue
		}

		if h.ConsoleManager != nil {
			h.ConsoleManager.Broadcast(msg)
		}
	}
}

// writePump pumps messages to the WebSocket connection
func (c *ConsoleClient) writePump() {
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
				log.Errorf("Failed to marshal message: %v", err)
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

// readPump pumps messages from the WebSocket connection
func (c *ConsoleClient) readPump() {
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
				log.Errorf("WebSocket error: %v", err)
			}
			break
		}

		var msg ConsoleMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Warnf("Invalid message: %v", err)
			continue
		}

		if msg.Type == "command" && msg.Command != "" {
			// Execute the command
			result, err := c.manager.ExecuteCommand(msg.Command)

			response := ConsoleMessage{
				Type:      "response",
				Timestamp: time.Now().Format(time.RFC3339),
				Command:   msg.Command,
			}

			if err != nil {
				response.Type = "error"
				response.Message = err.Error()
			} else {
				response.Body = result
			}

			c.send <- response
		}
	}
}
