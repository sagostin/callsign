package blf

import (
	"callsign/models"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service handles BLF (Busy Lamp Field) presence monitoring and subscription
type Service struct {
	DB *gorm.DB

	// Track active subscriptions: user@domain -> list of subscribers
	subscriptions map[string][]string
	mu            sync.RWMutex

	// WebSocket broadcast function (optional)
	broadcastFunc func(tenantID uint, event string, data interface{})
}

// New creates a new BLF service
func New(db *gorm.DB) *Service {
	return &Service{
		DB:            db,
		subscriptions: make(map[string][]string),
	}
}

// SetBroadcastFunc sets the WebSocket broadcast function
func (s *Service) SetBroadcastFunc(fn func(tenantID uint, event string, data interface{})) {
	s.broadcastFunc = fn
}

// HandlePresenceProbe handles PRESENCE_PROBE events from FreeSWITCH
// These occur when a device subscribes to BLF status for an extension
func (s *Service) HandlePresenceProbe(conn *eventsocket.Connection, ev *eventsocket.Event) error {
	proto := ev.Get("Proto")
	from := ev.Get("From")
	to := ev.Get("To")
	expiresStr := ev.Get("Expires")

	log.WithFields(log.Fields{
		"proto":   proto,
		"from":    from,
		"to":      to,
		"expires": expiresStr,
	}).Debug("BLF PRESENCE_PROBE received")

	// Parse user@domain from to field
	user, domain := parseUserDomain(to)
	if domain == "" {
		return nil
	}

	// Handle different protocols
	switch proto {
	case "dnd":
		return s.handleDNDProbe(conn, user, domain, to)
	case "forward":
		return s.handleForwardProbe(conn, user, domain, to)
	case "voicemail":
		return s.handleVoicemailProbe(conn, user, domain, to)
	case "flow":
		return s.handleCallFlowProbe(conn, user, domain, to)
	case "agent":
		return s.handleAgentProbe(conn, user, domain, to)
	default:
		// Standard extension presence - check registration/call state
		return s.handleExtensionProbe(conn, user, domain, to)
	}
}

// handleDNDProbe checks DND status for an extension
func (s *Service) handleDNDProbe(conn *eventsocket.Connection, user, domain, fullUser string) error {
	// Strip dnd+ prefix if present
	user = strings.TrimPrefix(user, "dnd+")

	var ext models.Extension
	if err := s.DB.Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
		Where("tenants.domain = ? AND (extensions.extension = ? OR extensions.number_alias = ?)",
			domain, user, user).
		First(&ext).Error; err != nil {
		log.Debugf("BLF: Extension not found for DND: %s@%s", user, domain)
		return nil
	}

	// Fire presence event - check if DND (DoNotDisturb) is enabled
	return s.turnLamp(conn, ext.DoNotDisturb, fullUser, "dnd")
}

// handleForwardProbe checks call forward status
func (s *Service) handleForwardProbe(conn *eventsocket.Connection, user, domain, fullUser string) error {
	// Strip forward+ prefix
	user = strings.TrimPrefix(user, "forward+")
	// Check if there's a specific number to check (format: extension/number)
	var targetNumber string
	if idx := strings.Index(user, "/"); idx > 0 {
		targetNumber = user[idx+1:]
		user = user[:idx]
	}

	var ext models.Extension
	if err := s.DB.Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
		Where("tenants.domain = ? AND (extensions.extension = ? OR extensions.number_alias = ?)",
			domain, user, user).
		First(&ext).Error; err != nil {
		return nil
	}

	isForwarded := ext.ForwardAllEnabled
	if isForwarded && targetNumber != "" {
		// Check if forward destination matches
		isForwarded = ext.ForwardAllDestination == targetNumber
	}

	return s.turnLamp(conn, isForwarded, fullUser, "forward")
}

// handleVoicemailProbe checks for unread voicemail
func (s *Service) handleVoicemailProbe(conn *eventsocket.Connection, user, domain, fullUser string) error {
	user = strings.TrimPrefix(user, "voicemail+")

	// Count unread voicemails for this extension
	var count int64
	s.DB.Table("voicemail_messages").
		Joins("JOIN voicemail_boxes ON voicemail_messages.voicemail_box_id = voicemail_boxes.id").
		Joins("JOIN tenants ON voicemail_boxes.tenant_id = tenants.id").
		Where("tenants.domain = ? AND voicemail_boxes.extension = ? AND voicemail_messages.read = false",
			domain, user).
		Count(&count)

	hasUnread := count > 0
	return s.turnLamp(conn, hasUnread, fullUser, "voicemail")
}

// handleCallFlowProbe checks call flow (day/night) status
func (s *Service) handleCallFlowProbe(conn *eventsocket.Connection, user, domain, fullUser string) error {
	user = strings.TrimPrefix(user, "flow+")

	var flow models.CallFlow
	if err := s.DB.Joins("JOIN tenants ON tenants.id = call_flows.tenant_id").
		Where("tenants.domain = ? AND (call_flows.extension = ? OR call_flows.feature_code = ?)",
			domain, user, user).
		First(&flow).Error; err != nil {
		return nil
	}

	// Lamp on = not at first state (0)
	isNotDefault := flow.CurrentState > 0
	return s.turnLamp(conn, isNotDefault, fullUser, "flow")

}

// handleAgentProbe checks call center agent status
func (s *Service) handleAgentProbe(conn *eventsocket.Connection, user, domain, fullUser string) error {
	user = strings.TrimPrefix(user, "agent+")

	// Query agent status from FreeSWITCH if available
	result, err := conn.Send(fmt.Sprintf("api callcenter_config agent get status %s@%s", user, domain))
	if err != nil {
		return nil
	}

	isAvailable := strings.Contains(string(result.Body), "Available")
	return s.turnLamp(conn, isAvailable, fullUser, "agent")
}

// handleExtensionProbe handles standard extension BLF
func (s *Service) handleExtensionProbe(conn *eventsocket.Connection, user, domain, fullUser string) error {
	// Get extension presence from database
	var presence models.ExtensionPresence
	if err := s.DB.Joins("JOIN tenants ON tenants.id = extension_presences.tenant_id").
		Where("tenants.domain = ? AND extension_presences.extension = ?", domain, user).
		First(&presence).Error; err != nil {
		// Not found - set to offline
		return s.turnLamp(conn, false, fullUser, "sip")
	}

	isBusy := presence.State == models.PresenceBusy ||
		presence.State == models.PresenceRinging ||
		presence.State == models.PresenceOnHold

	return s.turnLamp(conn, isBusy, fullUser, "sip")
}

// turnLamp sends a PRESENCE_IN event to control the BLF lamp
func (s *Service) turnLamp(conn *eventsocket.Connection, on bool, user, proto string) error {
	answerState := "terminated"
	eventCount := "0"
	rpid := ""
	if on {
		answerState = "confirmed"
		eventCount = "1"
		rpid = "unknown"
	}

	// Build SENDEVENT command for PRESENCE_IN
	event := fmt.Sprintf(`sendevent PRESENCE_IN
proto: %s
event_type: presence
alt_event_type: dialog
Presence-Call-Direction: outbound
from: %s
login: %s
unique-id: %s
status: Active (1 waiting)
answer-state: %s
rpid: %s
event_count: %s

`, proto, user, user, generateUUID(), answerState, rpid, eventCount)

	_, err := conn.Send(event)
	if err != nil {
		log.WithError(err).Error("BLF: Failed to send PRESENCE_IN")
		return err
	}

	log.WithFields(log.Fields{
		"user":  user,
		"proto": proto,
		"on":    on,
	}).Debug("BLF: Lamp state updated")

	return nil
}

// NotifyDNDChange notifies all subscribers of a DND status change
func (s *Service) NotifyDNDChange(conn *eventsocket.Connection, extension, domain string, enabled bool) error {
	user := fmt.Sprintf("dnd+%s@%s", extension, domain)
	return s.turnLamp(conn, enabled, user, "dnd")
}

// NotifyForwardChange notifies all subscribers of a call forward change
func (s *Service) NotifyForwardChange(conn *eventsocket.Connection, extension, domain string, enabled bool, destination string) error {
	// Notify general forward status
	user := fmt.Sprintf("forward+%s@%s", extension, domain)
	if err := s.turnLamp(conn, enabled, user, "forward"); err != nil {
		return err
	}

	// If specific destination, also notify that BLF
	if enabled && destination != "" {
		user = fmt.Sprintf("forward+%s/%s@%s", extension, destination, domain)
		return s.turnLamp(conn, true, user, "forward")
	}
	return nil
}

// NotifyVoicemailChange notifies subscribers when voicemail status changes
func (s *Service) NotifyVoicemailChange(conn *eventsocket.Connection, extension, domain string, hasUnread bool) error {
	user := fmt.Sprintf("voicemail+%s@%s", extension, domain)
	return s.turnLamp(conn, hasUnread, user, "voicemail")
}

// NotifyCallFlowChange notifies subscribers when call flow status changes
func (s *Service) NotifyCallFlowChange(conn *eventsocket.Connection, featureCode, domain string, isNightMode bool) error {
	user := fmt.Sprintf("flow+%s@%s", featureCode, domain)
	return s.turnLamp(conn, isNightMode, user, "flow")
}

// NotifyPresenceChange notifies subscribers of extension presence change
func (s *Service) NotifyPresenceChange(conn *eventsocket.Connection, extension, domain string, state models.PresenceState) error {
	user := fmt.Sprintf("%s@%s", extension, domain)
	isBusy := state == models.PresenceBusy || state == models.PresenceRinging || state == models.PresenceOnHold
	return s.turnLamp(conn, isBusy, user, "sip")
}

// BroadcastToWebSocket sends presence update to WebSocket clients
func (s *Service) BroadcastToWebSocket(tenantID uint, extension string, state models.PresenceState) {
	if s.broadcastFunc != nil {
		s.broadcastFunc(tenantID, "presence", map[string]interface{}{
			"extension": extension,
			"state":     state,
		})
	}
}

// Helper functions

func parseUserDomain(addr string) (user, domain string) {
	// Handle formats like "dnd+1001@example.com" or "1001@example.com"
	parts := strings.SplitN(addr, "@", 2)
	if len(parts) != 2 {
		return addr, ""
	}
	return parts[0], parts[1]
}

func generateUUID() string {
	// Simple UUID generation
	b := make([]byte, 16)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// PresenceEvent represents a presence event for JSON serialization
type PresenceEvent struct {
	Extension string               `json:"extension"`
	Domain    string               `json:"domain"`
	State     models.PresenceState `json:"state"`
	Proto     string               `json:"proto"`
	DND       bool                 `json:"dnd,omitempty"`
	Forwarded bool                 `json:"forwarded,omitempty"`
	ForwardTo string               `json:"forward_to,omitempty"`
}

// ToJSON serializes the event
func (e *PresenceEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
