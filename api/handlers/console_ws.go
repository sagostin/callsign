package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"callsign/config"
	"callsign/middleware"
	"callsign/services/esl"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins - auth handled via first message
	},
}

// ConsoleMessage represents a message to/from the console WebSocket
type ConsoleMessage struct {
	Type      string `json:"type"`  // "auth", "log", "command", "response", "error", "status"
	Level     string `json:"level"` // DEBUG, INFO, WARNING, ERROR
	Timestamp string `json:"timestamp"`
	Module    string `json:"module"`
	Message   string `json:"message"`
	Command   string `json:"command"`
	Body      string `json:"body"`
	Token     string `json:"token,omitempty"` // For auth message
}

// ConsoleClient represents a connected WebSocket client
type ConsoleClient struct {
	conn          *websocket.Conn
	send          chan ConsoleMessage
	manager       *ConsoleManager
	authenticated bool
	mu            sync.Mutex
}

// ConsoleManager manages WebSocket connections for the FS console
type ConsoleManager struct {
	clients    map[*ConsoleClient]bool
	broadcast  chan ConsoleMessage
	register   chan *ConsoleClient
	unregister chan *ConsoleClient
	eslManager *esl.Manager
	config     *config.Config
	mu         sync.RWMutex
}

// NewConsoleManager creates a new console manager
func NewConsoleManager(eslManager *esl.Manager, cfg *config.Config) *ConsoleManager {
	return &ConsoleManager{
		clients:    make(map[*ConsoleClient]bool),
		broadcast:  make(chan ConsoleMessage, 100),
		register:   make(chan *ConsoleClient),
		unregister: make(chan *ConsoleClient),
		eslManager: eslManager,
		config:     cfg,
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
			log.Infof("Console client connected (pending auth). Total: %d", len(m.clients))

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

// Broadcast sends a message to all connected authenticated clients
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

// ValidateToken validates a JWT token and returns claims
func (m *ConsoleManager) ValidateToken(tokenString string) (*middleware.Claims, error) {
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

// FreeSwitchConsole handles WebSocket connections for live FS console
// Auth is done via first message after connection (not query param)
func (h *Handler) FreeSwitchConsole(ctx iris.Context) {
	// Get the underlying http.ResponseWriter and *http.Request
	w := ctx.ResponseWriter()
	r := ctx.Request()

	// Upgrade to WebSocket (unauthenticated - auth via first message)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create console manager if not exists
	if h.ConsoleManager == nil {
		h.ConsoleManager = NewConsoleManager(h.ESLManager, h.Config)
		go h.ConsoleManager.Run()
		go h.startLogStreaming()
	}

	client := &ConsoleClient{
		conn:          conn,
		send:          make(chan ConsoleMessage, 256),
		manager:       h.ConsoleManager,
		authenticated: false,
	}

	h.ConsoleManager.register <- client

	// Send auth required message
	authMsg := ConsoleMessage{
		Type:      "auth_required",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Send auth message with token to authenticate",
	}
	data, _ := json.Marshal(authMsg)
	conn.WriteMessage(websocket.TextMessage, data)

	// Start read/write goroutines
	go client.writePump()
	go client.readPump()
}

// startLogStreaming subscribes to FS log events and broadcasts them
// Falls back to reading the log file directly if ESL is not connected
func (h *Handler) startLogStreaming() {
	// Try ESL first
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		// Subscribe to log events
		err := h.ESLManager.SubscribeEvents("LOG", "CUSTOM", "CHANNEL_CREATE", "CHANNEL_DESTROY")
		if err != nil {
			log.Errorf("Failed to subscribe to log events: %v", err)
		} else {
			log.Info("Started FreeSWITCH log streaming via ESL")
			go h.streamESLEvents()
			return
		}
	}

	// Fallback: tail the log file directly
	log.Info("ESL not available, falling back to log file streaming")
	go h.tailLogFile("/var/log/freeswitch/freeswitch.log")
}

// streamESLEvents processes ESL events and broadcasts them
func (h *Handler) streamESLEvents() {
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

// tailLogFile reads new lines from the FreeSWITCH log file and broadcasts them
func (h *Handler) tailLogFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Warnf("Could not open log file %s: %v", path, err)
		// Broadcast error message
		if h.ConsoleManager != nil {
			h.ConsoleManager.Broadcast(ConsoleMessage{
				Type:      "log",
				Level:     "WARNING",
				Timestamp: time.Now().Format(time.RFC3339),
				Module:    "logtail",
				Message:   "Log file not available: " + path,
			})
		}
		return
	}
	defer file.Close()

	// Seek to end of file
	file.Seek(0, 2)

	reader := bufio.NewReader(file)
	log.Infof("Started tailing log file: %s", path)

	if h.ConsoleManager != nil {
		h.ConsoleManager.Broadcast(ConsoleMessage{
			Type:      "log",
			Level:     "INFO",
			Timestamp: time.Now().Format(time.RFC3339),
			Module:    "logtail",
			Message:   "Started streaming from " + path,
		})
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse log level from FreeSWITCH log format
		level := "INFO"
		if strings.Contains(line, "[DEBUG]") {
			level = "DEBUG"
		} else if strings.Contains(line, "[INFO]") {
			level = "INFO"
		} else if strings.Contains(line, "[NOTICE]") {
			level = "NOTICE"
		} else if strings.Contains(line, "[WARNING]") || strings.Contains(line, "[WARN]") {
			level = "WARNING"
		} else if strings.Contains(line, "[ERR]") || strings.Contains(line, "[ERROR]") {
			level = "ERROR"
		} else if strings.Contains(line, "[CRIT]") || strings.Contains(line, "[ALERT]") {
			level = "ERROR"
		}

		if h.ConsoleManager != nil {
			h.ConsoleManager.Broadcast(ConsoleMessage{
				Type:      "log",
				Level:     level,
				Timestamp: time.Now().Format(time.RFC3339),
				Module:    "freeswitch",
				Message:   line,
			})
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

		// Handle authentication
		if msg.Type == "auth" && msg.Token != "" {
			claims, err := c.manager.ValidateToken(msg.Token)
			if err != nil {
				c.send <- ConsoleMessage{
					Type:      "auth_error",
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   "Invalid or expired token",
				}
				c.conn.Close()
				return
			}

			if claims.Role != "system_admin" {
				c.send <- ConsoleMessage{
					Type:      "auth_error",
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   "System admin access required",
				}
				c.conn.Close()
				return
			}

			c.mu.Lock()
			c.authenticated = true
			c.mu.Unlock()

			c.send <- ConsoleMessage{
				Type:      "auth_success",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Connected to FreeSWITCH console",
			}
			log.Infof("Console client authenticated: %s", claims.Username)
			continue
		}

		// Require authentication for all other messages
		if !c.authenticated {
			c.send <- ConsoleMessage{
				Type:      "error",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Authentication required",
			}
			continue
		}

		// Handle commands
		if msg.Type == "command" && msg.Command != "" {
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
