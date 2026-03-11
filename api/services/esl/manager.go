package esl

import (
	"callsign/config"
	"callsign/models"
	"callsign/services/websocket"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	WSHub     *websocket.Hub

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

	// Custom IVR/Auto-Attendant engine (walks flow graph via ESL)
	m.Registry.Register("ivr", "127.0.0.7:9001", m.handleIVR)
}

// handleCallControl handles general call control with full routing
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
	callerID := ev.Get("Caller-Caller-ID-Number")
	tenantIDStr := ev.Get("variable_tenant_id")
	ringGroupUUID := ev.Get("variable_ring_group_uuid")

	log.WithFields(log.Fields{
		"uuid":            uuid,
		"destination":     dest,
		"domain":          domain,
		"caller":          callerID,
		"ring_group_uuid": ringGroupUUID,
	}).Info("Call control: handling call")

	// Subscribe to events for this channel
	conn.Send("linger")
	conn.Send("myevents")

	// Parse tenant ID
	var tenantID uint = 1
	if tenantIDStr != "" {
		fmt.Sscanf(tenantIDStr, "%d", &tenantID)
	}

	// Broadcast call event via WebSocket
	if m.WSHub != nil {
		m.WSHub.NotifyCallEvent(tenantID, "ringing", map[string]interface{}{
			"uuid":        uuid,
			"caller":      callerID,
			"destination": dest,
			"domain":      domain,
		})
	}

	// Route based on what variables are set by the dialplan
	if ringGroupUUID != "" {
		// Ring group call — handle with strategy
		m.handleRingGroupCall(conn, ringGroupUUID, dest, domain, tenantID)
		return
	}

	// Check if this is an internal extension call
	var ext struct {
		ID                     uint
		Extension              string
		ForwardAllEnabled      bool
		ForwardAllDest         string
		ForwardBusyEnabled     bool
		ForwardBusyDest        string
		ForwardNoAnswerEnabled bool
		ForwardNoAnswerDest    string
		DoNotDisturb           bool
		VoicemailEnabled       bool
	}
	err = m.DB.Table("extensions").
		Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
		Where("tenants.domain = ? AND extensions.extension = ? AND extensions.enabled = ?", domain, dest, true).
		Select("extensions.id, extensions.extension, extensions.forward_all_enabled, extensions.forward_all_dest, " +
			"extensions.forward_busy_enabled, extensions.forward_busy_dest, " +
			"extensions.forward_no_answer_enabled, extensions.forward_no_answer_dest, " +
			"extensions.do_not_disturb, extensions.voicemail_enabled").
		First(&ext).Error

	if err == nil {
		// Found internal extension
		// Check DND
		if ext.DoNotDisturb {
			log.Infof("Call control: DND active for %s, sending to voicemail", dest)
			if ext.VoicemailEnabled {
				conn.Execute("answer", "", true)
				conn.Execute("voicemail", fmt.Sprintf("default %s %s", domain, dest), true)
			} else {
				conn.Execute("respond", "486 Busy Here", false)
			}
			return
		}

		// Check unconditional forwarding
		if ext.ForwardAllEnabled && ext.ForwardAllDest != "" {
			log.Infof("Call control: forwarding %s to %s (forward all)", dest, ext.ForwardAllDest)
			conn.Execute("set", "call_timeout=30", true)
			conn.Execute("bridge", fmt.Sprintf("user/%s@%s", ext.ForwardAllDest, domain), true)
			return
		}

		// Try to ring the extension with no-answer/busy fallback
		conn.Execute("set", "call_timeout=30", true)
		conn.Execute("set", "hangup_after_bridge=true", true)
		conn.Execute("set", "continue_on_fail=true", true)

		dialString := fmt.Sprintf("user/%s@%s", dest, domain)
		conn.Execute("bridge", dialString, true)

		// Check bridge result for forwarding
		resultEv, err := conn.ReadEvent()
		if err != nil {
			return
		}

		hangupCause := resultEv.Get("variable_hangup_cause")
		if hangupCause == "" {
			hangupCause = resultEv.Get("variable_originate_disposition")
		}

		switch hangupCause {
		case "USER_BUSY":
			if ext.ForwardBusyEnabled && ext.ForwardBusyDest != "" {
				log.Infof("Call control: busy forwarding %s to %s", dest, ext.ForwardBusyDest)
				conn.Execute("bridge", fmt.Sprintf("user/%s@%s", ext.ForwardBusyDest, domain), true)
				return
			}
			if ext.VoicemailEnabled {
				conn.Execute("answer", "", true)
				conn.Execute("voicemail", fmt.Sprintf("default %s %s", domain, dest), true)
			}

		case "NO_ANSWER", "ALLOTTED_TIMEOUT", "NO_USER_RESPONSE":
			if ext.ForwardNoAnswerEnabled && ext.ForwardNoAnswerDest != "" {
				log.Infof("Call control: no-answer forwarding %s to %s", dest, ext.ForwardNoAnswerDest)
				conn.Execute("bridge", fmt.Sprintf("user/%s@%s", ext.ForwardNoAnswerDest, domain), true)
				return
			}
			if ext.VoicemailEnabled {
				conn.Execute("answer", "", true)
				conn.Execute("voicemail", fmt.Sprintf("default %s %s", domain, dest), true)
			}
		}

		return
	}

	// Not an internal extension — try outbound routing
	m.handleOutboundCall(conn, dest, domain, tenantID)
}

// handleRingGroupCall handles calls routed to ring groups
func (m *Manager) handleRingGroupCall(conn *eventsocket.Connection, ringGroupUUID, dest, domain string, tenantID uint) {
	var rg models.RingGroup
	if err := m.DB.Where("uuid = ? AND enabled = ?", ringGroupUUID, true).
		Preload("Destinations", func(db *gorm.DB) *gorm.DB {
			return db.Order("priority ASC")
		}).
		First(&rg).Error; err != nil {
		log.Errorf("Ring group %s not found: %v", ringGroupUUID, err)
		conn.Execute("hangup", "UNALLOCATED_NUMBER", false)
		return
	}

	if len(rg.Destinations) == 0 {
		log.Warnf("Ring group %s has no destinations", rg.Name)
		conn.Execute("hangup", "UNALLOCATED_NUMBER", false)
		return
	}

	conn.Execute("set", "hangup_after_bridge=true", true)
	conn.Execute("set", "continue_on_fail=true", true)

	if rg.RingbackTone != "" {
		conn.Execute("set", fmt.Sprintf("ringback=%s", rg.RingbackTone), true)
	}

	switch rg.Strategy {
	case models.RingStrategySimultaneous:
		// Ring all destinations at once
		var dialStrings []string
		for _, d := range rg.Destinations {
			ds := m.buildRingGroupDialString(d, domain)
			if ds != "" {
				dialStrings = append(dialStrings, ds)
			}
		}
		if len(dialStrings) > 0 {
			timeout := rg.RingTimeout
			if timeout <= 0 {
				timeout = 30
			}
			conn.Execute("set", fmt.Sprintf("call_timeout=%d", timeout), true)
			// Simultaneous: comma-separated
			conn.Execute("bridge", strings.Join(dialStrings, ","), true)
		}

	case models.RingStrategySequence:
		// Ring destinations one after another
		for _, d := range rg.Destinations {
			ds := m.buildRingGroupDialString(d, domain)
			if ds == "" {
				continue
			}
			timeout := d.Timeout
			if timeout <= 0 {
				timeout = rg.RingTimeout
			}
			if timeout <= 0 {
				timeout = 30
			}

			if d.Delay > 0 {
				conn.Execute("sleep", fmt.Sprintf("%d", d.Delay*1000), true)
			}

			conn.Execute("set", fmt.Sprintf("call_timeout=%d", timeout), true)
			conn.Execute("bridge", ds, true)

			// Check if bridge succeeded
			ev, err := conn.ReadEvent()
			if err != nil {
				return
			}
			cause := ev.Get("variable_originate_disposition")
			if cause == "SUCCESS" || cause == "" {
				return // Call connected
			}
		}

	case models.RingStrategyRandom:
		// Shuffle destinations and ring sequentially
		// Use a simple approach: try each destination with randomized order
		perm := make([]int, len(rg.Destinations))
		for i := range perm {
			perm[i] = i
		}
		// Fisher-Yates shuffle using UUID-based seed
		for i := len(perm) - 1; i > 0; i-- {
			j := int(rg.UUID[0]+rg.UUID[1]) % (i + 1)
			perm[i], perm[j] = perm[j], perm[i]
		}

		for _, idx := range perm {
			d := rg.Destinations[idx]
			ds := m.buildRingGroupDialString(d, domain)
			if ds == "" {
				continue
			}
			timeout := rg.RingTimeout
			if timeout <= 0 {
				timeout = 30
			}
			conn.Execute("set", fmt.Sprintf("call_timeout=%d", timeout), true)
			conn.Execute("bridge", ds, true)

			ev, err := conn.ReadEvent()
			if err != nil {
				return
			}
			cause := ev.Get("variable_originate_disposition")
			if cause == "SUCCESS" || cause == "" {
				return
			}
		}

	default:
		// Fallback: ring all simultaneously
		var dialStrings []string
		for _, d := range rg.Destinations {
			ds := m.buildRingGroupDialString(d, domain)
			if ds != "" {
				dialStrings = append(dialStrings, ds)
			}
		}
		if len(dialStrings) > 0 {
			conn.Execute("set", "call_timeout=30", true)
			conn.Execute("bridge", strings.Join(dialStrings, ","), true)
		}
	}

	// If we get here, all ring group destinations failed
	// Check for timeout destination
	if rg.TimeoutDestination != "" {
		action := rg.TimeoutDestinationType
		switch action {
		case "voicemail":
			conn.Execute("answer", "", true)
			conn.Execute("voicemail", fmt.Sprintf("default %s %s", domain, rg.TimeoutDestination), true)
		case "extension":
			conn.Execute("bridge", fmt.Sprintf("user/%s@%s", rg.TimeoutDestination, domain), true)
		default:
			conn.Execute("transfer", fmt.Sprintf("%s XML %s", rg.TimeoutDestination, domain), true)
		}
	}
}

// buildRingGroupDialString builds a FreeSWITCH dial string for a ring group destination
func (m *Manager) buildRingGroupDialString(d models.RingGroupDestination, domain string) string {
	switch d.DestinationType {
	case "extension":
		return fmt.Sprintf("user/%s@%s", d.Destination, domain)
	case "external":
		return fmt.Sprintf("sofia/gateway/default/%s", d.Destination)
	case "gateway":
		return fmt.Sprintf("sofia/gateway/%s/%s", d.Destination, d.Destination)
	default:
		if d.Destination != "" {
			return fmt.Sprintf("user/%s@%s", d.Destination, domain)
		}
		return ""
	}
}

// handleOutboundCall routes calls to PSTN via gateways
func (m *Manager) handleOutboundCall(conn *eventsocket.Connection, dest, domain string, tenantID uint) {
	// Find matching outbound route
	var routes []models.DefaultOutboundRoute
	m.DB.Where("enabled = ?", true).Order("\"order\" ASC").Find(&routes)

	for _, route := range routes {
		// Simple prefix matching
		if route.DigitPrefix != "" && !strings.HasPrefix(dest, route.DigitPrefix) {
			continue
		}

		// Check digit length
		if len(dest) < route.DigitMin || len(dest) > route.DigitMax {
			continue
		}

		// Found a matching route — get the gateway
		var gw models.Gateway
		if err := m.DB.Where("id = ? AND enabled = ?", route.GatewayID, true).First(&gw).Error; err != nil {
			continue
		}

		// Build dial destination
		dialDest := dest
		if route.StripDigits > 0 && len(dialDest) > route.StripDigits {
			dialDest = dialDest[route.StripDigits:]
		}
		if route.PrependDigits != "" {
			dialDest = route.PrependDigits + dialDest
		}

		log.WithFields(log.Fields{
			"route":   route.Name,
			"gateway": gw.GatewayName,
			"dest":    dialDest,
		}).Info("Call control: outbound route matched")

		conn.Execute("set", "hangup_after_bridge=true", true)

		bridgeStr := fmt.Sprintf("sofia/gateway/%s/%s", gw.GatewayName, dialDest)
		conn.Execute("bridge", bridgeStr, true)

		// If primary fails and failover is configured
		if route.Gateway2ID != nil {
			ev, err := conn.ReadEvent()
			if err != nil {
				return
			}
			cause := ev.Get("variable_originate_disposition")
			if cause != "SUCCESS" && cause != "" {
				var gw2 models.Gateway
				if err := m.DB.Where("id = ? AND enabled = ?", *route.Gateway2ID, true).First(&gw2).Error; err == nil {
					failoverStr := fmt.Sprintf("sofia/gateway/%s/%s", gw2.GatewayName, dialDest)
					conn.Execute("bridge", failoverStr, true)
				}
			}
		}

		return
	}

	// No matching outbound route
	log.Warnf("Call control: no outbound route for %s", dest)
	conn.Execute("respond", "404 Not Found", false)
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

	// Parse tenant ID
	var tenantID uint = 1
	if tidStr := ev.Get("variable_tenant_id"); tidStr != "" {
		fmt.Sscanf(tidStr, "%d", &tenantID)
	}

	// Determine if this is voicemail retrieval (*97) or deposit
	isRetrieval := dest == "*97" || dest == "*98"

	if isRetrieval {
		// Voicemail retrieval — authenticate with extension and PIN
		conn.Execute("playback", "voicemail/vm-enter_id.wav", true)
		result, _ := conn.Execute("read", "3 8 voicemail/vm-enter_id.wav # 5000", true)
		extNum := ""
		if result != nil {
			extNum = result.Get("variable_read_result")
		}
		if extNum == "" {
			extNum = callerID // Default to caller's extension
		}

		// Look up voicemail box for PIN verification
		var box models.VoicemailBox
		if err := m.DB.Where("extension = ? AND tenant_id = ? AND enabled = true", extNum, tenantID).
			First(&box).Error; err != nil {
			conn.Execute("playback", "voicemail/vm-goodbye.wav", true)
			conn.Execute("hangup", "", false)
			return
		}

		// PIN authentication
		if box.Password != "" {
			conn.Execute("playback", "voicemail/vm-enter_pass.wav", true)
			pinResult, _ := conn.Execute("read", "4 8 voicemail/vm-enter_pass.wav # 5000", true)
			pin := ""
			if pinResult != nil {
				pin = pinResult.Get("variable_read_result")
			}
			if pin != box.Password {
				conn.Execute("playback", "voicemail/vm-fail_auth.wav", true)
				conn.Execute("hangup", "", false)
				return
			}
		}

		// Play message count and iterate through messages
		var count int64
		m.DB.Model(&models.VoicemailMessage{}).Where("box_id = ? AND is_new = true", box.ID).Count(&count)

		if count == 0 {
			conn.Execute("playback", "voicemail/vm-no_more_messages.wav", true)
		} else {
			conn.Execute("say", fmt.Sprintf("en number pronounced %d", count), true)
			conn.Execute("playback", "voicemail/vm-messages.wav", true)

			// Play messages
			var messages []models.VoicemailMessage
			m.DB.Where("box_id = ? AND is_new = true", box.ID).Order("created_at DESC").Find(&messages)
			for i := range messages {
				if messages[i].FilePath != "" {
					conn.Execute("playback", messages[i].FilePath, true)
					// Mark as read
					now := time.Now()
					m.DB.Model(&messages[i]).Updates(map[string]interface{}{"is_new": false, "read_at": &now})
				}
			}
		}

		conn.Execute("playback", "voicemail/vm-goodbye.wav", true)
		conn.Execute("hangup", "", false)

	} else {
		// Voicemail deposit — record a message
		// Look up voicemail box for destination extension
		var box models.VoicemailBox
		err := m.DB.Where("extension = ? AND tenant_id = ? AND enabled = true", dest, tenantID).
			First(&box).Error

		// Play greeting
		if err == nil && box.GreetingPath != "" {
			conn.Execute("playback", box.GreetingPath, true)
		} else {
			conn.Execute("playback", "ivr/ivr-please_leave_message.wav", true)
		}

		// Record beep + message
		conn.Execute("playback", "tone_stream://%(200,0,800)", true)

		recordDir := "/var/lib/callsign/voicemail"
		recordPath := fmt.Sprintf("%s/%s_%s.wav", recordDir, dest, uuid)
		maxSecs := 180
		if box.MaxMessageSecs > 0 {
			maxSecs = box.MaxMessageSecs
		}
		conn.Execute("record", fmt.Sprintf("%s %d 100 5", recordPath, maxSecs), true)

		// Save voicemail to database
		if box.ID > 0 {
			vm := models.VoicemailMessage{
				BoxID:          box.ID,
				TenantID:       tenantID,
				CallerIDNumber: callerID,
				CallerIDName:   ev.Get("Caller-Caller-ID-Name"),
				FilePath:       recordPath,
				RecordedAt:     time.Now(),
				IsNew:          true,
				ChannelUUID:    uuid,
			}
			if err := m.DB.Create(&vm).Error; err != nil {
				log.Errorf("Voicemail: failed to save: %v", err)
			}

			// Update box message count
			m.DB.Model(&box).UpdateColumn("new_messages", gorm.Expr("new_messages + 1"))

			// Broadcast MWI (Message Waiting Indicator) via WebSocket
			if m.WSHub != nil {
				m.WSHub.BroadcastToTenant(tenantID, websocket.EventVoicemail, "new_message", map[string]interface{}{
					"box_id":       box.ID,
					"extension_id": box.ExtensionID,
					"caller":       callerID,
					"extension":    dest,
				})
			}
		}

		conn.Execute("playback", "ivr/ivr-thank_you.wav", true)
		conn.Execute("hangup", "", false)
	}
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

	// Broadcast conference join via WebSocket
	if m.WSHub != nil {
		var confTenantID uint = 1
		if tenantIDStr := ev.Get("variable_tenant_id"); tenantIDStr != "" {
			fmt.Sscanf(tenantIDStr, "%d", &confTenantID)
		}
		m.WSHub.NotifyConferenceEvent(confTenantID, "member_join", map[string]interface{}{
			"uuid":       uuid,
			"caller":     callerID,
			"conference": conferenceNum,
		})
	}

	conn.Send("linger")
	conn.Send("myevents")

	// Answer
	conn.Execute("answer", "", true)

	// Look up conference by number and check for PIN
	var conf models.Conference
	confFound := m.DB.Where("extension = ? AND enabled = ?", conferenceNum, true).
		Joins("JOIN tenants ON tenants.id = conferences.tenant_id AND tenants.domain = ?", domain).
		First(&conf).Error == nil

	if confFound && conf.PIN != "" {
		// Authenticate with PIN
		conn.Execute("playback", "conference/conf-pin.wav", true)
		pinResult, _ := conn.Execute("read", "4 8 conference/conf-pin.wav # 5000", true)
		pin := ""
		if pinResult != nil {
			pin = pinResult.Get("variable_read_result")
		}
		if pin != conf.PIN {
			conn.Execute("playback", "conference/conf-bad-pin.wav", true)
			conn.Execute("hangup", "", false)
			return
		}
	}

	// Build conference flags
	confProfile := "default"
	if confFound {
		if conf.MaxMembers > 0 {
			conn.Execute("set", fmt.Sprintf("conference_max_members=%d", conf.MaxMembers), true)
		}
		if conf.RecordConference {
			conn.Execute("set", "conference_auto_record=true", true)
		}
		if conf.MuteOnJoin {
			confProfile = "default+mute"
		}
	}

	// Join conference
	confName := fmt.Sprintf("%s-%s", domain, conferenceNum)
	conn.Execute("conference", fmt.Sprintf("%s@%s", confName, confProfile), true)

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

// SofiaRescan rescans a Sofia profile to pick up new gateways
func (m *Manager) SofiaRescan(profileName string) error {
	cmd := fmt.Sprintf("sofia profile %s rescan", profileName)
	_, err := m.API(cmd)
	if err != nil {
		log.Warnf("Failed to rescan Sofia profile %s: %v", profileName, err)
	} else {
		log.Infof("Sofia profile %s rescanned", profileName)
	}
	return err
}

// SofiaRestart restarts a Sofia profile (for profile-level config changes)
func (m *Manager) SofiaRestart(profileName string) error {
	cmd := fmt.Sprintf("sofia profile %s restart reloadxml", profileName)
	_, err := m.API(cmd)
	if err != nil {
		log.Warnf("Failed to restart Sofia profile %s: %v", profileName, err)
	} else {
		log.Infof("Sofia profile %s restarted", profileName)
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
		// Try to get FreeSWITCH system status
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

	// Broadcast queue event via WebSocket
	if m.WSHub != nil {
		var qTenantID uint = 1
		if tenantIDStr := ev.Get("variable_tenant_id"); tenantIDStr != "" {
			fmt.Sscanf(tenantIDStr, "%d", &qTenantID)
		}
		m.WSHub.BroadcastToTenant(qTenantID, websocket.EventQueue, "caller_enter", map[string]interface{}{
			"uuid":   uuid,
			"caller": callerID,
			"queue":  queueNum,
		})
	}

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

// handleIVR handles IVR/Auto-Attendant calls using the custom flow engine
func (m *Manager) handleIVR(conn *eventsocket.Connection) {
	defer conn.Close()

	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("IVR: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	callerName := ev.Get("Caller-Caller-ID-Name")
	dest := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")

	logger := log.WithFields(log.Fields{
		"uuid":   uuid,
		"caller": callerID,
		"dest":   dest,
		"domain": domain,
	})
	logger.Info("IVR: handling call")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer the call
	conn.Execute("answer", "", true)

	// Look up the IVR menu by extension
	var menu models.IVRMenu
	if err := m.DB.Where("extension = ? AND enabled = ?", dest, true).
		Preload("Options").First(&menu).Error; err != nil {
		logger.Errorf("IVR: no menu found for extension %s", dest)
		conn.Execute("playback", "ivr/ivr-invalid_entry.wav", true)
		conn.Execute("hangup", "", false)
		return
	}

	logger = logger.WithField("ivr_menu", menu.Name)

	// If we have visual flow data, execute the flow graph
	if len(menu.FlowData.Nodes) > 0 {
		logger.Info("IVR: executing visual flow graph")
		m.executeIVRFlow(conn, &menu, uuid, callerID, callerName, dest, domain, logger)
	} else {
		// Legacy: use IVRMenuOption rows
		logger.Info("IVR: executing legacy menu options")
		m.executeIVRLegacy(conn, &menu, uuid, callerID, dest, domain, logger)
	}
}

// executeIVRFlow walks the visual flow graph stored in menu.FlowData
func (m *Manager) executeIVRFlow(conn *eventsocket.Connection, menu *models.IVRMenu,
	uuid, callerID, callerName, dest, domain string, logger *log.Entry) {

	nodes := menu.FlowData.Nodes
	connections := menu.FlowData.Connections

	// Build maps
	nodeMap := make(map[string]*models.IVRFlowNode)
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}
	connMap := make(map[string]string)
	for _, c := range connections {
		connMap[c.SourceID+":"+c.SourceOutput] = c.TargetID
	}

	// Context variables for ${var} substitution
	vars := map[string]string{
		"caller_id":   callerID,
		"caller_name": callerName,
		"destination": dest,
		"domain":      domain,
		"ivr_name":    menu.Name,
	}

	// Find start node
	var current *models.IVRFlowNode
	for _, n := range nodeMap {
		if n.Type == "ivr_start" {
			current = n
			break
		}
	}
	if current == nil && len(nodes) > 0 {
		current = &nodes[0]
	}

	// Walk the graph (max 100 iterations for safety)
	for step := 0; step < 100 && current != nil; step++ {
		logger.WithFields(log.Fields{"step": step, "node": current.Type, "id": current.ID}).Debug("IVR: exec node")

		output := m.execIVRNode(conn, current, uuid, domain, vars, logger)
		if output == "__hangup__" || output == "" {
			break
		}

		// Find next node via connection map
		nextID := connMap[current.ID+":"+output]
		if nextID == "" {
			nextID = connMap[current.ID+":next"]
		}
		if nextID == "" {
			nextID = connMap[current.ID+":"]
		}
		current = nodeMap[nextID]
	}

	conn.Execute("hangup", "", false)
}

// execIVRNode executes a single flow node and returns the output port name
func (m *Manager) execIVRNode(conn *eventsocket.Connection, node *models.IVRFlowNode,
	uuid, domain string, vars map[string]string, logger *log.Entry) string {

	cfg := node.Config
	getStr := func(key, def string) string {
		if v, ok := cfg[key]; ok {
			if s, ok := v.(string); ok {
				return resolveIVRVars(s, vars)
			}
		}
		return def
	}
	getInt := func(key string, def int) int {
		if v, ok := cfg[key]; ok {
			if f, ok := v.(float64); ok {
				return int(f)
			}
		}
		return def
	}

	switch node.Type {
	case "ivr_start":
		return "next"

	case "gather":
		maxDigits := getInt("maxDigits", 1)
		timeout := getInt("timeout", 10)
		maxRetries := getInt("maxRetries", 3)
		promptType := getStr("promptType", "tts")
		var prompt string
		if promptType == "audio" {
			prompt = getStr("audioFile", "silence_stream://250")
		} else {
			prompt = fmt.Sprintf("say:%s", getStr("ttsText", "Please make your selection"))
		}
		invalidSound := getStr("invalidSound", "ivr/ivr-invalid_entry.wav")

		for attempt := 0; attempt < maxRetries; attempt++ {
			cmd := fmt.Sprintf("1 %d 1 %d # %s %s digits \\d+ %d",
				maxDigits, timeout*1000, prompt, invalidSound, timeout*1000)
			conn.Execute("play_and_get_digits", cmd, true)

			ev, err := conn.Send("api uuid_getvar " + uuid + " digits")
			if err != nil || strings.TrimSpace(ev.Body) == "" || ev.Body == "_undef_" {
				continue
			}
			digits := strings.TrimSpace(ev.Body)
			vars["caller_input"] = digits
			vars["gathered_digits"] = digits
			logger.WithField("digits", digits).Info("IVR: gathered digits")
			return "match"
		}
		return "timeout"

	case "play_audio":
		file := getStr("audioFile", "")
		if file != "" {
			conn.Execute("playback", file, true)
		}
		return "next"

	case "play_tts":
		text := getStr("text", "")
		if text != "" {
			engine := getStr("engine", "flite")
			voice := getStr("voice", "default")
			conn.Execute("speak", fmt.Sprintf("%s|%s|%s", engine, voice, text), true)
		}
		return "next"

	case "say_digits":
		value := getStr("value", "")
		if value != "" {
			conn.Execute("say", fmt.Sprintf("en number iterated %s", value), true)
		}
		return "next"

	case "web_request":
		method := getStr("method", "GET")
		url := getStr("url", "")
		if url == "" {
			return "error"
		}
		respVar := getStr("responseVar", "api_response")
		timeout := getInt("timeout", 5)
		bodyStr := getStr("body", "")

		client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
		var body io.Reader
		if bodyStr != "" && method != "GET" {
			body = strings.NewReader(bodyStr)
		}
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return "error"
		}
		if method != "GET" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := client.Do(req)
		if err != nil {
			vars[respVar] = ""
			return "error"
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		vars[respVar] = string(respBody)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return "success"
		}
		return "error"

	case "condition":
		variable := getStr("variable", "")
		op := getStr("operator", "==")
		value := getStr("value", "")
		result := false
		switch op {
		case "==":
			result = variable == value
		case "!=":
			result = variable != value
		case "contains":
			result = strings.Contains(variable, value)
		case ">":
			result = variable > value
		case "<":
			result = variable < value
		}
		if result {
			return "true"
		}
		return "false"

	case "set_variable":
		name := getStr("name", "")
		value := getStr("value", "")
		if name != "" {
			vars[name] = value
		}
		return "next"

	case "extension":
		ext := getStr("extension", "")
		if ext != "" {
			conn.Execute("transfer", fmt.Sprintf("%s XML %s", ext, domain), false)
		}
		return "__hangup__"

	case "queue":
		q := getStr("queueId", "")
		if q != "" {
			conn.Execute("transfer", fmt.Sprintf("%s XML %s", q, domain), false)
		}
		return "__hangup__"

	case "ring_group":
		rg := getStr("groupId", "")
		if rg != "" {
			conn.Execute("transfer", fmt.Sprintf("%s XML %s", rg, domain), false)
		}
		return "__hangup__"

	case "ivr_menu":
		menuID := getStr("menuId", "")
		if menuID != "" {
			conn.Execute("transfer", fmt.Sprintf("%s XML %s", menuID, domain), false)
		}
		return "__hangup__"

	case "external":
		number := getStr("number", "")
		if number != "" {
			conn.Execute("bridge", fmt.Sprintf("sofia/gateway/default/%s", number), true)
		}
		return "__hangup__"

	case "voicemail":
		mailbox := getStr("mailboxId", vars["destination"])
		conn.Execute("transfer", fmt.Sprintf("*99%s XML %s", mailbox, domain), false)
		return "__hangup__"

	case "hangup":
		conn.Execute("hangup", "", false)
		return "__hangup__"

	case "send_sms":
		logger.Info("IVR: SMS sending requested (provider integration needed)")
		return "sent"

	default:
		logger.Warnf("IVR: unknown node type: %s", node.Type)
		return "next"
	}
}

// executeIVRLegacy runs a traditional IVR using IVRMenuOption rows
func (m *Manager) executeIVRLegacy(conn *eventsocket.Connection, menu *models.IVRMenu,
	uuid, callerID, dest, domain string, logger *log.Entry) {

	for attempt := 0; attempt < menu.MaxFailures+menu.MaxTimeouts; attempt++ {
		greeting := menu.GreetLong
		if attempt > 0 && menu.GreetShort != "" {
			greeting = menu.GreetShort
		}
		if greeting != "" {
			conn.Execute("playback", greeting, true)
		}

		timeout := menu.Timeout
		if timeout == 0 {
			timeout = 10
		}
		maxDigits := menu.DigitLen
		if maxDigits == 0 {
			maxDigits = 1
		}

		cmd := fmt.Sprintf("1 %d 1 %d # silence_stream://250 %s digits \\d+ %d",
			maxDigits, timeout*1000, menu.InvalidSound, timeout*1000)
		conn.Execute("play_and_get_digits", cmd, true)

		ev, err := conn.Send("api uuid_getvar " + uuid + " digits")
		if err != nil {
			break
		}
		digits := strings.TrimSpace(ev.Body)
		if digits == "" || digits == "_undef_" {
			continue
		}

		for _, opt := range menu.Options {
			if !opt.Enabled || opt.Digits != digits {
				continue
			}
			logger.WithFields(log.Fields{"digits": digits, "action": opt.Action}).Info("IVR legacy: matched")
			switch opt.Action {
			case models.IVRActionPlayback:
				conn.Execute("playback", opt.ActionParam, true)
				continue
			case models.IVRActionRepeat:
				continue
			case models.IVRActionHangup:
				conn.Execute("hangup", "", false)
			default:
				conn.Execute("transfer", fmt.Sprintf("%s XML %s", opt.ActionParam, domain), false)
			}
			return
		}

		if menu.InvalidSound != "" {
			conn.Execute("playback", menu.InvalidSound, true)
		}
	}

	if menu.ExitSound != "" {
		conn.Execute("playback", menu.ExitSound, true)
	}
	conn.Execute("hangup", "", false)
}

// resolveIVRVars replaces ${var} placeholders with values from the vars map
func resolveIVRVars(input string, vars map[string]string) string {
	result := input
	for k, v := range vars {
		result = strings.ReplaceAll(result, "${"+k+"}", v)
	}
	return result
}
