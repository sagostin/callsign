package esl

import (
	"callsign/config"
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
	}
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

	// Register ESL services
	m.registerServices()

	// Start all registered services
	if err := m.Registry.StartAll(); err != nil {
		m.Stop()
		return fmt.Errorf("failed to start ESL services: %w", err)
	}

	m.running = true
	log.Info("ESL manager started")

	return nil
}

// registerServices registers the ESL socket servers
func (m *Manager) registerServices() {
	// Call control service
	m.Registry.Register("callcontrol", "127.0.0.1:9001", m.handleCallControl)

	// Voicemail service
	m.Registry.Register("voicemail", "127.0.0.2:9001", m.handleVoicemail)

	// Queue/Call Center service
	m.Registry.Register("queue", "127.0.0.3:9001", m.handleQueue)

	// Conference service
	m.Registry.Register("conference", "127.0.0.4:9001", m.handleConference)

	// BLF/Presence service
	m.Registry.Register("blf", "127.0.0.5:9001", m.handleBLF)

	// Feature codes service
	m.Registry.Register("featurecodes", "127.0.0.6:9001", m.handleFeatureCodes)
}

// handleCallControl handles general call control
func (m *Manager) handleCallControl(conn *eventsocket.Connection) {
	defer conn.Close()

	// Connect and get channel info
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Call control: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	dest := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")

	log.WithFields(log.Fields{
		"uuid":        uuid,
		"destination": dest,
		"domain":      domain,
	}).Info("Call control: handling call")

	// Subscribe to events for this channel
	conn.Send("linger")
	conn.Send("myevents")

	// TODO: Implement actual call routing logic
	// For now, just bridge to the destination

	// Example: Bridge to user
	dialString := fmt.Sprintf("user/%s@%s", dest, domain)
	conn.Execute("bridge", dialString, true)

	// Wait for hangup
	for {
		ev, err := conn.ReadEvent()
		if err != nil {
			break
		}

		if ev.Get("Event-Name") == "CHANNEL_HANGUP_COMPLETE" {
			break
		}
	}
}

// handleVoicemail handles voicemail operations
func (m *Manager) handleVoicemail(conn *eventsocket.Connection) {
	defer conn.Close()

	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Voicemail: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	dest := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")

	log.WithFields(log.Fields{
		"uuid":      uuid,
		"caller":    callerID,
		"extension": dest,
		"domain":    domain,
	}).Info("Voicemail: handling request")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer the call
	conn.Execute("answer", "", true)

	// TODO: Implement voicemail logic
	// - Check if this is deposit or retrieval
	// - For retrieval: authenticate user, play messages
	// - For deposit: record message, save to DB

	// For now, play a placeholder greeting
	conn.Execute("playback", "ivr/ivr-please_leave_message.wav", true)

	// Record message
	recordPath := fmt.Sprintf("/tmp/voicemail_%s.wav", uuid)
	conn.Execute("record", fmt.Sprintf("%s 60 100 5", recordPath), true)

	// Thank the caller
	conn.Execute("playback", "ivr/ivr-thank_you.wav", true)

	// Hangup
	conn.Execute("hangup", "", false)
}

// handleConference handles conference rooms
func (m *Manager) handleConference(conn *eventsocket.Connection) {
	defer conn.Close()

	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Conference: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	conferenceNum := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")

	log.WithFields(log.Fields{
		"uuid":       uuid,
		"caller":     callerID,
		"conference": conferenceNum,
		"domain":     domain,
	}).Info("Conference: joining caller")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer
	conn.Execute("answer", "", true)

	// TODO: Implement conference logic
	// - Look up conference by number/PIN
	// - Authenticate if required
	// - Join conference

	// For now, join a simple conference
	confName := fmt.Sprintf("%s-%s", domain, conferenceNum)
	conn.Execute("conference", fmt.Sprintf("%s@default", confName), true)

	// Wait for hangup
	for {
		ev, err := conn.ReadEvent()
		if err != nil {
			break
		}
		if ev.Get("Event-Name") == "CHANNEL_HANGUP_COMPLETE" {
			break
		}
	}
}

// Stop stops the ESL manager
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	// Stop services
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

// handleQueue handles call center queue calls
func (m *Manager) handleQueue(conn *eventsocket.Connection) {
	defer conn.Close()

	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Queue: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	queueNum := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")

	log.WithFields(log.Fields{
		"uuid":   uuid,
		"caller": callerID,
		"queue":  queueNum,
		"domain": domain,
	}).Info("Queue: caller entering queue")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer
	conn.Execute("answer", "", true)

	// Set queue variables
	conn.Execute("set", "hangup_after_bridge=true", true)
	conn.Execute("set", "continue_on_fail=false", true)

	// Enter the queue (mod_callcenter)
	queueName := fmt.Sprintf("%s@%s", queueNum, domain)
	conn.Execute("callcenter", queueName, true)

	// Wait for hangup
	for {
		ev, err := conn.ReadEvent()
		if err != nil {
			break
		}
		if ev.Get("Event-Name") == "CHANNEL_HANGUP_COMPLETE" {
			break
		}
	}
}

// handleFeatureCodes handles feature code dialing (e.g., *97, *72)
func (m *Manager) handleFeatureCodes(conn *eventsocket.Connection) {
	defer conn.Close()

	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("FeatureCodes: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	code := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")
	tenantIDStr := ev.Get("variable_tenant_id")

	log.WithFields(log.Fields{
		"uuid":   uuid,
		"caller": callerID,
		"code":   code,
		"domain": domain,
	}).Info("FeatureCodes: processing feature code")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer
	conn.Execute("answer", "", true)

	// Parse tenant ID
	var tenantID uint = 1
	if tenantIDStr != "" {
		fmt.Sscanf(tenantIDStr, "%d", &tenantID)
	}

	// Look up feature code in database
	var fc struct {
		Code   string
		Action string
	}
	err = m.DB.Table("feature_codes").
		Select("code, action").
		Where("tenant_id = ? AND code = ? AND enabled = ?", tenantID, code, true).
		First(&fc).Error

	if err != nil {
		log.Warnf("Feature code not found: %s", code)
		conn.Execute("playback", "ivr/ivr-invalid_selection.wav", true)
		conn.Execute("hangup", "NORMAL_CLEARING", true)
		return
	}

	// Execute based on action type
	switch fc.Action {
	case "voicemail":
		if code == "*97" {
			conn.Execute("voicemail", fmt.Sprintf("check default %s %s", domain, callerID), true)
		} else {
			conn.Execute("voicemail", fmt.Sprintf("check default %s", domain), true)
		}

	case "call_forward":
		if code == "*72" {
			conn.Execute("playback", "ivr/ivr-enter_dest_number.wav", true)
			conn.Execute("read", "2 20 tone_stream://%(250,50,440) forward_dest 10000 #", true)
			// Would update DB here
			conn.Execute("playback", "ivr/ivr-call_forwarding_is_now_enabled.wav", true)
		} else {
			conn.Execute("playback", "ivr/ivr-call_forwarding_is_now_disabled.wav", true)
		}

	case "dnd":
		if code == "*78" {
			conn.Execute("playback", "ivr/ivr-dnd_activated.wav", true)
		} else {
			conn.Execute("playback", "ivr/ivr-dnd_deactivated.wav", true)
		}

	case "call_flow_toggle":
		conn.Execute("playback", "ivr/ivr-night_mode.wav", true)

	default:
		conn.Execute("playback", "tone_stream://%(100,0,600);%(100,0,800)", true)
	}

	conn.Execute("hangup", "NORMAL_CLEARING", true)
}

// GetActiveCalls returns the number of active calls
func (m *Manager) GetActiveCalls() int {
	return m.Sessions.Count()
}

// handleBLF handles BLF/presence subscription events
func (m *Manager) handleBLF(conn *eventsocket.Connection) {
	defer conn.Close()

	// Connect and subscribe to presence events
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("BLF: connect failed: %v", err)
		return
	}

	log.WithFields(log.Fields{
		"uuid": ev.Get("Unique-Id"),
	}).Info("BLF: Connection established")

	// Subscribe to PRESENCE_PROBE events
	_, err = conn.Send("event json PRESENCE_PROBE")
	if err != nil {
		log.Errorf("BLF: Failed to subscribe to presence events: %v", err)
		return
	}

	// Event loop
	for {
		ev, err := conn.ReadEvent()
		if err != nil {
			log.Debugf("BLF: Connection closed: %v", err)
			return
		}

		eventName := ev.Get("Event-Name")
		if eventName == "PRESENCE_PROBE" {
			proto := ev.Get("Proto")
			from := ev.Get("From")
			to := ev.Get("To")

			log.WithFields(log.Fields{
				"proto": proto,
				"from":  from,
				"to":    to,
			}).Debug("BLF: PRESENCE_PROBE received")

			// Handle the probe based on protocol
			m.handlePresenceProbe(conn, ev, proto, to)
		}
	}
}

// handlePresenceProbe processes a PRESENCE_PROBE event
func (m *Manager) handlePresenceProbe(conn *eventsocket.Connection, ev *eventsocket.Event, proto, to string) {
	// Parse user@domain
	parts := splitUserDomain(to)
	if parts[1] == "" {
		return
	}

	user := parts[0]
	domain := parts[1]

	switch proto {
	case "dnd":
		m.handleDNDProbe(conn, user, domain, to)
	case "forward":
		m.handleForwardProbe(conn, user, domain, to)
	case "voicemail":
		m.handleVoicemailProbe(conn, user, domain, to)
	case "flow":
		m.handleCallFlowProbe(conn, user, domain, to)
	default:
		// Standard extension BLF
		m.handleExtensionProbe(conn, user, domain, to)
	}
}

func (m *Manager) handleDNDProbe(conn *eventsocket.Connection, user, domain, to string) {
	// Query extension DND status from database
	var ext struct {
		DoNotDisturb bool
	}
	if err := m.DB.Table("extensions").
		Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
		Where("tenants.domain = ? AND extensions.extension = ?", domain, user).
		Select("do_not_disturb").First(&ext).Error; err != nil {
		return
	}

	sendPresenceIn(conn, to, ext.DoNotDisturb, "dnd")
}

func (m *Manager) handleForwardProbe(conn *eventsocket.Connection, user, domain, to string) {
	var ext struct {
		ForwardAllEnabled bool
	}
	if err := m.DB.Table("extensions").
		Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
		Where("tenants.domain = ? AND extensions.extension = ?", domain, user).
		Select("forward_all_enabled").First(&ext).Error; err != nil {
		return
	}

	sendPresenceIn(conn, to, ext.ForwardAllEnabled, "forward")
}

func (m *Manager) handleVoicemailProbe(conn *eventsocket.Connection, user, domain, to string) {
	var count int64
	m.DB.Table("voicemail_messages").
		Joins("JOIN voicemail_boxes ON voicemail_messages.voicemail_box_id = voicemail_boxes.id").
		Joins("JOIN tenants ON voicemail_boxes.tenant_id = tenants.id").
		Where("tenants.domain = ? AND voicemail_boxes.extension = ? AND voicemail_messages.is_read = false", domain, user).
		Count(&count)

	sendPresenceIn(conn, to, count > 0, "voicemail")
}

func (m *Manager) handleCallFlowProbe(conn *eventsocket.Connection, user, domain, to string) {
	var flow struct {
		Status string
	}
	if err := m.DB.Table("call_flows").
		Joins("JOIN tenants ON tenants.id = call_flows.tenant_id").
		Where("tenants.domain = ? AND (call_flows.extension = ? OR call_flows.feature_code = ?)", domain, user, user).
		Select("status").First(&flow).Error; err != nil {
		return
	}

	// Lamp on = night mode
	sendPresenceIn(conn, to, flow.Status != "day", "flow")
}

func (m *Manager) handleExtensionProbe(conn *eventsocket.Connection, user, domain, to string) {
	var presence struct {
		State string
	}
	if err := m.DB.Table("extension_presences").
		Joins("JOIN tenants ON tenants.id = extension_presences.tenant_id").
		Where("tenants.domain = ? AND extension_presences.extension = ?", domain, user).
		Select("state").First(&presence).Error; err != nil {
		return
	}

	isBusy := presence.State == "busy" || presence.State == "ringing" || presence.State == "onhold"
	sendPresenceIn(conn, to, isBusy, "sip")
}

// sendPresenceIn sends PRESENCE_IN event to control BLF lamp
func sendPresenceIn(conn *eventsocket.Connection, user string, on bool, proto string) {
	answerState := "terminated"
	eventCount := "0"
	if on {
		answerState = "confirmed"
		eventCount = "1"
	}

	event := fmt.Sprintf(`sendevent PRESENCE_IN
proto: %s
event_type: presence
alt_event_type: dialog
Presence-Call-Direction: outbound
from: %s
login: %s
status: Active (1 waiting)
answer-state: %s
event_count: %s

`, proto, user, user, answerState, eventCount)

	conn.Send(event)
}

func splitUserDomain(addr string) [2]string {
	for i, c := range addr {
		if c == '@' {
			return [2]string{addr[:i], addr[i+1:]}
		}
	}
	return [2]string{addr, ""}
}
