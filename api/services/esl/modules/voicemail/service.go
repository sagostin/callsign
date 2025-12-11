package voicemail

import (
	"callsign/models"
	"callsign/services/esl"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
)

const (
	ServiceName    = "voicemail"
	ServiceAddress = "127.0.0.2:9001"
)

// Service implements the voicemail ESL module
type Service struct {
	*esl.BaseService
	recordingDir string
}

// New creates a new voicemail service
func New() *Service {
	return &Service{
		BaseService:  esl.NewBaseService(ServiceName, ServiceAddress),
		recordingDir: "/var/lib/callsign/voicemail",
	}
}

// Init initializes the voicemail service
func (s *Service) Init(manager *esl.Manager) error {
	if err := s.BaseService.Init(manager); err != nil {
		return err
	}

	// Ensure recording directory exists
	if err := os.MkdirAll(s.recordingDir, 0755); err != nil {
		log.Warnf("Could not create voicemail directory: %v", err)
	}

	log.Info("Voicemail service initialized")
	return nil
}

// Handle processes incoming voicemail connections
func (s *Service) Handle(conn *eventsocket.Connection) {
	defer conn.Close()

	manager := s.Manager()
	if manager == nil {
		log.Error("Voicemail: manager not initialized")
		return
	}

	// Connect and get channel info
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Voicemail: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	callerName := ev.Get("Caller-Caller-ID-Name")
	dest := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")
	action := ev.Get("variable_voicemail_action") // "deposit" or "check"

	logger := log.WithFields(log.Fields{
		"uuid":      uuid,
		"caller":    callerID,
		"extension": dest,
		"domain":    domain,
		"action":    action,
	})
	logger.Info("Voicemail: handling request")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer the call
	conn.Execute("answer", "", true)

	// Determine action (deposit vs check)
	if action == "check" || dest == "*97" || dest == "*98" {
		s.handleCheck(conn, manager, uuid, callerID, domain, logger)
	} else {
		s.handleDeposit(conn, manager, uuid, callerID, callerName, dest, domain, logger)
	}
}

// handleDeposit handles leaving a voicemail message
func (s *Service) handleDeposit(conn *eventsocket.Connection, manager *esl.Manager, uuid, callerID, callerName, extension, domain string, logger *log.Entry) {
	db := manager.DB

	// Find the voicemail box
	var box models.VoicemailBox
	if err := db.Where("extension = ? AND enabled = ?", extension, true).First(&box).Error; err != nil {
		logger.Warnf("Voicemail box not found for extension %s", extension)
		conn.Execute("playback", "voicemail/vm-not_available.wav", true)
		conn.Execute("hangup", "", false)
		return
	}

	// Check if box is full
	if box.MaxMessages > 0 && (box.NewMessages+box.SavedMessages) >= box.MaxMessages {
		logger.Info("Voicemail box is full")
		conn.Execute("playback", "voicemail/vm-mailbox_full.wav", true)
		conn.Execute("hangup", "", false)
		return
	}

	// Play greeting
	if box.GreetingPath != "" && fileExists(box.GreetingPath) {
		conn.Execute("playback", box.GreetingPath, true)
	} else {
		conn.Execute("playback", "voicemail/vm-person.wav", true)
		conn.Execute("say", fmt.Sprintf("en number iterated %s", extension), true)
		conn.Execute("playback", "voicemail/vm-not_available.wav", true)
	}

	// Prompt to leave message
	if !box.SkipInstructions {
		conn.Execute("playback", "voicemail/vm-record_message.wav", true)
	}

	// Record message
	recordPath := s.getRecordingPath(box.TenantID, box.Extension, uuid)
	recordStart := time.Now()

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(recordPath), 0755)

	// Record with max length, silence detection
	maxSecs := box.MaxMessageSecs
	if maxSecs == 0 {
		maxSecs = 180
	}
	conn.Execute("record", fmt.Sprintf("%s %d 100 5", recordPath, maxSecs), true)

	recordEnd := time.Now()
	duration := int(recordEnd.Sub(recordStart).Seconds())

	// Get file size
	var fileSize int64
	if info, err := os.Stat(recordPath); err == nil {
		fileSize = info.Size()
	}

	// Minimum duration check (ignore short recordings)
	if duration < 3 {
		logger.Info("Recording too short, discarding")
		os.Remove(recordPath)
		conn.Execute("hangup", "", false)
		return
	}

	// Save to database
	message := &models.VoicemailMessage{
		BoxID:          box.ID,
		TenantID:       box.TenantID,
		CallerIDName:   callerName,
		CallerIDNumber: callerID,
		Duration:       duration,
		FilePath:       recordPath,
		FileSize:       fileSize,
		RecordedAt:     recordStart,
		IsNew:          true,
		ChannelUUID:    uuid,
	}

	if err := db.Create(message).Error; err != nil {
		logger.Errorf("Failed to save voicemail message: %v", err)
	} else {
		// Update box message count
		db.Model(&box).Update("new_messages", box.NewMessages+1)
		logger.Infof("Voicemail saved: %d seconds", duration)
	}

	// Thank the caller
	conn.Execute("playback", "voicemail/vm-goodbye.wav", true)
	conn.Execute("hangup", "", false)

	// TODO: Send email notification if configured
	// TODO: Send MWI notification to extension
}

// handleCheck handles checking voicemail messages
func (s *Service) handleCheck(conn *eventsocket.Connection, manager *esl.Manager, uuid, callerID, domain string, logger *log.Entry) {
	db := manager.DB

	// Find the voicemail box for this caller
	var box models.VoicemailBox
	if err := db.Where("extension = ? AND enabled = ?", callerID, true).First(&box).Error; err != nil {
		logger.Warnf("No voicemail box for caller %s", callerID)
		conn.Execute("playback", "voicemail/vm-not_available.wav", true)
		conn.Execute("hangup", "", false)
		return
	}

	// TODO: Authenticate with PIN
	// conn.Execute("playback", "voicemail/vm-enter_pass.wav", true)
	// conn.Execute("read", "4 4 voicemail/vm-enter_pass.wav pin 10000 #", true)

	// Get new messages
	var messages []models.VoicemailMessage
	db.Where("box_id = ? AND is_new = ?", box.ID, true).Order("created_at DESC").Find(&messages)

	// Announce message count
	count := len(messages)
	if count == 0 {
		conn.Execute("playback", "voicemail/vm-no_messages.wav", true)
	} else {
		conn.Execute("playback", "voicemail/vm-you_have.wav", true)
		conn.Execute("say", fmt.Sprintf("en number pronounced %d", count), true)
		if count == 1 {
			conn.Execute("playback", "voicemail/vm-message.wav", true)
		} else {
			conn.Execute("playback", "voicemail/vm-messages.wav", true)
		}

		// Play each message
		for i, msg := range messages {
			conn.Execute("playback", "voicemail/vm-message_number.wav", true)
			conn.Execute("say", fmt.Sprintf("en number pronounced %d", i+1), true)

			// Play the recording
			if fileExists(msg.FilePath) {
				conn.Execute("playback", msg.FilePath, true)
			}

			// Mark as read
			msg.MarkAsRead(db)

			// TODO: Add menu for save/delete/repeat
		}
	}

	conn.Execute("playback", "voicemail/vm-goodbye.wav", true)
	conn.Execute("hangup", "", false)
}

// getRecordingPath generates the file path for a recording
func (s *Service) getRecordingPath(tenantID uint, extension, uuid string) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s.wav", timestamp, uuid[:8])
	return filepath.Join(s.recordingDir, fmt.Sprintf("%d", tenantID), extension, filename)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
