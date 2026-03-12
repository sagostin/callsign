package esl

import (
	"callsign/config"
	"callsign/services/tts"
	"callsign/services/websocket"
	"fmt"
	"sync"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Manager manages the ESL client, servers, and session tracking
type Manager struct {
	Config    *config.Config
	DB        *gorm.DB
	Client    *Client
	Sessions  *SessionManager
	Processor *EventProcessor
	Registry  *ServiceRegistry
	Modules   *ModuleRegistry
	WSHub     *websocket.Hub
	TTS       *tts.Service

	running bool
	mu      sync.RWMutex
}

// NewManager creates a new ESL manager
func NewManager(cfg *config.Config, db *gorm.DB) *Manager {
	return &Manager{
		Config:   cfg,
		DB:       db,
		Sessions: NewSessionManager(),
		Registry: NewServiceRegistry(),
		Modules:  NewModuleRegistry(),
	}
}

// SetWSHub sets the WebSocket hub for real-time event broadcasting
func (m *Manager) SetWSHub(hub *websocket.Hub) {
	m.WSHub = hub
}

// Start starts the ESL manager
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("ESL manager already running")
	}

	// Create and connect inbound client
	m.Client = NewClient(
		m.Config.FreeSwitchHost,
		8021, // TODO: make configurable
		m.Config.FreeSwitchPassword,
	)

	if err := m.Client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to FreeSWITCH: %w", err)
	}

	// Subscribe to events
	events := []string{
		"CHANNEL_CREATE",
		"CHANNEL_ANSWER",
		"CHANNEL_BRIDGE",
		"CHANNEL_UNBRIDGE",
		"CHANNEL_HANGUP_COMPLETE",
		"CHANNEL_STATE",
		"DTMF",
		"RECORD_START",
		"RECORD_STOP",
		"PLAYBACK_START",
		"PLAYBACK_STOP",
		"CUSTOM",
	}

	if err := m.Client.Subscribe(events...); err != nil {
		m.Client.Close()
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Create event processor with default handlers
	m.Processor = NewEventProcessor(m.Client, m.Sessions)

	handlers := DefaultEventHandlers(m.Sessions)
	for eventName, handler := range handlers {
		m.Processor.On(eventName, handler)
	}

	// Start processing events
	m.Client.StartEventLoop()
	m.Processor.Start()

	// Initialize and start all registered modules
	if err := m.Modules.InitAll(m); err != nil {
		m.Stop()
		return fmt.Errorf("failed to init ESL modules: %w", err)
	}
	if err := m.Modules.StartAll(); err != nil {
		m.Stop()
		return fmt.Errorf("failed to start ESL modules: %w", err)
	}

	m.running = true
	log.Info("ESL manager started")

	return nil
}

// RegisterModule registers an ESL service module. Call before Start().
func (m *Manager) RegisterModule(service Service) error {
	return m.Modules.Register(service)
}

// Stop stops the ESL manager
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	// Stop modules
	m.Modules.StopAll()

	// Stop legacy service registry
	m.Registry.StopAll()

	// Close client
	if m.Client != nil {
		m.Client.Close()
	}

	m.running = false
	log.Info("ESL manager stopped")
}

// IsRunning returns whether the manager is running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// IsConnected returns whether the ESL client is connected
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Client != nil && m.Client.IsConnected()
}

// API sends an API command to FreeSWITCH and returns the result
func (m *Manager) API(command string) (string, error) {
	m.mu.RLock()
	client := m.Client
	m.mu.RUnlock()

	if client == nil {
		return "", fmt.Errorf("not connected")
	}
	return client.API(command)
}

// BgAPI sends a background API command to FreeSWITCH and returns the job UUID.
// Use this for long-running commands (e.g. sofia profile restart) that may
// block or disrupt the ESL connection if run synchronously.
func (m *Manager) BgAPI(command string) (string, error) {
	m.mu.RLock()
	client := m.Client
	m.mu.RUnlock()

	if client == nil {
		return "", fmt.Errorf("not connected")
	}
	return client.BgAPI(command)
}

// ReloadXML sends a reloadxml command to FreeSWITCH
// Call this after config changes that affect xml_curl-served data
func (m *Manager) ReloadXML() error {
	_, err := m.API("reloadxml")
	if err != nil {
		log.Warnf("Failed to reload XML: %v", err)
	} else {
		log.Info("FreeSWITCH XML reloaded")
	}
	return err
}

// SofiaRescan rescans a Sofia profile to pick up new gateways.
// Uses BgAPI to avoid blocking the ESL socket — rescan can take several seconds.
func (m *Manager) SofiaRescan(profileName string) error {
	cmd := fmt.Sprintf("sofia profile %s rescan", profileName)
	jobUUID, err := m.BgAPI(cmd)
	if err != nil {
		log.Warnf("Failed to rescan Sofia profile %s: %v", profileName, err)
	} else {
		log.Infof("Sofia profile %s rescan queued (job %s)", profileName, jobUUID)
	}
	return err
}

// SofiaRestart restarts a Sofia profile (for profile-level config changes).
// Uses BgAPI to avoid blocking the ESL socket — restart can take 10-30s
// and will disrupt the connection if run synchronously.
func (m *Manager) SofiaRestart(profileName string) error {
	cmd := fmt.Sprintf("sofia profile %s restart", profileName)
	jobUUID, err := m.BgAPI(cmd)
	if err != nil {
		log.Warnf("Failed to restart Sofia profile %s: %v", profileName, err)
	} else {
		log.Infof("Sofia profile %s restart queued (job %s)", profileName, jobUUID)
	}
	return err
}

// CallcenterReload reloads mod_callcenter configuration
func (m *Manager) CallcenterReload() error {
	_, err := m.API("reload mod_callcenter")
	if err != nil {
		log.Warnf("Failed to reload mod_callcenter: %v", err)
	} else {
		log.Info("mod_callcenter reloaded")
	}
	return err
}

// ReloadACL reloads the access control list configuration
func (m *Manager) ReloadACL() error {
	_, err := m.API("reloadacl")
	if err != nil {
		log.Warnf("Failed to reload ACL: %v", err)
	} else {
		log.Info("ACL reloaded")
	}
	return err
}

// FreeSwitchStatus returns the status of the FreeSWITCH connection and system info
func (m *Manager) FreeSwitchStatus() map[string]interface{} {
	status := map[string]interface{}{
		"esl_connected": m.IsConnected(),
		"esl_running":   m.IsRunning(),
		"active_calls":  m.GetActiveCalls(),
	}

	if m.IsConnected() {
		if result, err := m.API("status"); err == nil {
			status["freeswitch_status"] = result
		}
		if result, err := m.API("sofia status"); err == nil {
			status["sofia_status"] = result
		}
	}

	return status
}

// SubscribeEvents subscribes to additional FreeSWITCH events
func (m *Manager) SubscribeEvents(events ...string) error {
	m.mu.RLock()
	client := m.Client
	m.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("not connected")
	}
	return client.Subscribe(events...)
}

// Events returns the event channel from the client
func (m *Manager) Events() <-chan *eventsocket.Event {
	m.mu.RLock()
	client := m.Client
	m.mu.RUnlock()

	if client == nil {
		return nil
	}
	return client.Events()
}

// GetActiveCalls returns the number of active calls
func (m *Manager) GetActiveCalls() int {
	return m.Sessions.Count()
}

// NotifyCallEvent broadcasts a call event through the WebSocket hub
func (m *Manager) NotifyCallEvent(tenantID uint, event string, data map[string]interface{}) {
	if m.WSHub != nil {
		m.WSHub.NotifyCallEvent(tenantID, event, data)
	}
}
