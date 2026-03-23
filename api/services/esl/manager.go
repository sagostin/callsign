package esl

import (
	"callsign/config"
	"callsign/services/email"
	"callsign/services/tts"
	"callsign/services/websocket"
	"fmt"
	"sync"
	"time"

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
	Email     *email.Service

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

// SetEmailService sets the email service for notifications
func (m *Manager) SetEmailService(svc *email.Service) {
	m.Email = svc
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
		"PRESENCE_PROBE",
		"PRESENCE_IN",
		"MESSAGE_WAITING",
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

// ========== Presence / BLF / MWI Helpers ==========
// These methods send presence and MWI events to FreeSWITCH via the inbound ESL
// connection, allowing any module or API handler to update BLF lamp states and
// voicemail indicators without needing a direct reference to the blf package.

// SendPresenceEvent sends a PRESENCE_IN event to update a BLF lamp on phones.
// `on` controls whether the lamp is lit (confirmed) or off (terminated).
// `user` is the full presence identity (e.g., "dnd+1001@example.com").
// `proto` is the presence protocol (e.g., "sip", "dnd", "forward", "voicemail").
func (m *Manager) SendPresenceEvent(user, proto string, on bool) {
	m.mu.RLock()
	client := m.Client
	m.mu.RUnlock()

	if client == nil {
		return
	}

	answerState := "terminated"
	eventCount := "0"
	rpid := ""
	if on {
		answerState = "confirmed"
		eventCount = "1"
		rpid = "unknown"
	}

	event := fmt.Sprintf("sendevent PRESENCE_IN\n"+
		"proto: %s\n"+
		"event_type: presence\n"+
		"alt_event_type: dialog\n"+
		"Presence-Call-Direction: outbound\n"+
		"from: %s\n"+
		"login: %s\n"+
		"unique-id: %s\n"+
		"status: Active (1 waiting)\n"+
		"answer-state: %s\n"+
		"rpid: %s\n"+
		"event_count: %s\n\n",
		proto, user, user, fmt.Sprintf("%d", time.Now().UnixNano()),
		answerState, rpid, eventCount)

	client.Send(event)
}

// SendDNDPresence updates the BLF lamp for DND status on an extension.
func (m *Manager) SendDNDPresence(extension, domain string, enabled bool) {
	user := fmt.Sprintf("dnd+%s@%s", extension, domain)
	m.SendPresenceEvent(user, "dnd", enabled)
}

// SendForwardPresence updates the BLF lamp for call forward status.
func (m *Manager) SendForwardPresence(extension, domain string, enabled bool) {
	user := fmt.Sprintf("forward+%s@%s", extension, domain)
	m.SendPresenceEvent(user, "forward", enabled)
}

// SendExtensionPresence updates the BLF lamp for an extension's call state.
func (m *Manager) SendExtensionPresence(extension, domain string, busy bool) {
	user := fmt.Sprintf("%s@%s", extension, domain)
	m.SendPresenceEvent(user, "sip", busy)
}

// SendMWI sends a Message Waiting Indicator event to FreeSWITCH.
// This causes phones to light their voicemail lamp / show envelope icon.
func (m *Manager) SendMWI(extension, domain string, newMsgs, savedMsgs int) {
	m.mu.RLock()
	client := m.Client
	m.mu.RUnlock()

	if client == nil {
		return
	}

	waiting := "no"
	if newMsgs > 0 {
		waiting = "yes"
	}

	event := fmt.Sprintf("sendevent MESSAGE_WAITING\n"+
		"MWI-Messages-Waiting: %s\n"+
		"MWI-Message-Account: sip:%s@%s\n"+
		"MWI-Voice-Message: %d/%d (0/0)\n\n",
		waiting, extension, domain, newMsgs, savedMsgs)

	client.Send(event)

	log.WithFields(log.Fields{
		"extension": extension,
		"new":       newMsgs,
		"saved":     savedMsgs,
	}).Debug("MWI sent via ESL Manager")
}
