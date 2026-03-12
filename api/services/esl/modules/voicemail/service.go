package voicemail

import (
	"callsign/models"
	"callsign/services/esl"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	ServiceName    = "voicemail"
	ServiceAddress = "127.0.0.2:9001"

	maxPINAttempts = 3
)

// Service implements the voicemail ESL module with full IVR
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

// ========== Voicemail Deposit ==========

// handleDeposit handles leaving a voicemail message
func (s *Service) handleDeposit(conn *eventsocket.Connection, manager *esl.Manager,
	uuid, callerID, callerName, extension, domain string, logger *log.Entry) {

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

	// Send MWI notification
	s.sendMWI(manager, box.Extension, domain, box.NewMessages+1, box.SavedMessages)

	// Notify via WebSocket
	if manager.WSHub != nil {
		manager.NotifyCallEvent(box.TenantID, "voicemail_new", map[string]interface{}{
			"extension": box.Extension,
			"caller":    callerID,
			"duration":  duration,
		})
	}

	// Thank the caller
	conn.Execute("playback", "voicemail/vm-goodbye.wav", true)
	conn.Execute("hangup", "", false)
}

// ========== Voicemail Check (Full IVR) ==========

// handleCheck handles checking voicemail messages with full IVR
func (s *Service) handleCheck(conn *eventsocket.Connection, manager *esl.Manager,
	uuid, callerID, domain string, logger *log.Entry) {

	db := manager.DB

	// Find the voicemail box for this caller
	var box models.VoicemailBox
	if err := db.Where("extension = ? AND enabled = ?", callerID, true).First(&box).Error; err != nil {
		logger.Warnf("No voicemail box for caller %s", callerID)
		conn.Execute("playback", "voicemail/vm-not_available.wav", true)
		conn.Execute("hangup", "", false)
		return
	}

	// --- PIN Authentication ---
	if box.Password != "" {
		authenticated := false
		for attempt := 0; attempt < maxPINAttempts; attempt++ {
			pin := s.collectDigits(conn, uuid, "voicemail/vm-enter_pass.wav", 4, 8, 5000)
			if pin == box.Password {
				authenticated = true
				break
			}
			conn.Execute("playback", "voicemail/vm-fail_auth.wav", true)
		}
		if !authenticated {
			logger.Warn("PIN auth failed after 3 attempts")
			conn.Execute("playback", "voicemail/vm-goodbye.wav", true)
			conn.Execute("hangup", "", false)
			return
		}
	}

	// --- Main Menu IVR ---
	s.mainMenu(conn, manager, uuid, &box, domain, logger)
}

// mainMenu presents the main voicemail menu
func (s *Service) mainMenu(conn *eventsocket.Connection, manager *esl.Manager,
	uuid string, box *models.VoicemailBox, domain string, logger *log.Entry) {

	db := manager.DB

	for {
		// Refresh message counts
		var newCount, savedCount int64
		db.Model(&models.VoicemailMessage{}).Where("box_id = ? AND is_new = true", box.ID).Count(&newCount)
		db.Model(&models.VoicemailMessage{}).Where("box_id = ? AND is_new = false AND deleted_at IS NULL", box.ID).Count(&savedCount)

		// Announce counts
		conn.Execute("playback", "voicemail/vm-you_have.wav", true)
		conn.Execute("say", fmt.Sprintf("en number pronounced %d", newCount), true)
		if newCount == 1 {
			conn.Execute("playback", "voicemail/vm-new.wav", true)
			conn.Execute("playback", "voicemail/vm-message.wav", true)
		} else {
			conn.Execute("playback", "voicemail/vm-new.wav", true)
			conn.Execute("playback", "voicemail/vm-messages.wav", true)
		}
		if savedCount > 0 {
			conn.Execute("say", fmt.Sprintf("en number pronounced %d", savedCount), true)
			conn.Execute("playback", "voicemail/vm-saved.wav", true)
			if savedCount == 1 {
				conn.Execute("playback", "voicemail/vm-message.wav", true)
			} else {
				conn.Execute("playback", "voicemail/vm-messages.wav", true)
			}
		}

		// Main menu prompt:
		// 1 = listen to new messages
		// 2 = listen to saved messages
		// 5 = greeting management
		// * = exit
		conn.Execute("playback", "voicemail/vm-main_menu.wav", true)
		digit := s.collectDigits(conn, uuid, "silence_stream://2000", 1, 1, 5000)

		switch digit {
		case "1":
			s.listenMessages(conn, manager, uuid, box, domain, true, logger)
		case "2":
			s.listenMessages(conn, manager, uuid, box, domain, false, logger)
		case "5":
			s.greetingMenu(conn, manager, uuid, box, domain, logger)
		case "*":
			conn.Execute("playback", "voicemail/vm-goodbye.wav", true)
			conn.Execute("hangup", "", false)
			return
		default:
			conn.Execute("playback", "voicemail/vm-invalid_value.wav", true)
		}
	}
}

// listenMessages plays messages and provides per-message DTMF controls
func (s *Service) listenMessages(conn *eventsocket.Connection, manager *esl.Manager,
	uuid string, box *models.VoicemailBox, domain string, isNew bool, logger *log.Entry) {

	db := manager.DB

	var messages []models.VoicemailMessage
	if isNew {
		db.Where("box_id = ? AND is_new = true", box.ID).Order("created_at ASC").Find(&messages)
	} else {
		db.Where("box_id = ? AND is_new = false AND deleted_at IS NULL", box.ID).Order("created_at ASC").Find(&messages)
	}

	if len(messages) == 0 {
		conn.Execute("playback", "voicemail/vm-no_messages.wav", true)
		return
	}

	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		// Announce message number and envelope info
		conn.Execute("playback", "voicemail/vm-message_number.wav", true)
		conn.Execute("say", fmt.Sprintf("en number pronounced %d", i+1), true)

		// Announce caller and timestamp
		if msg.CallerIDNumber != "" {
			conn.Execute("playback", "voicemail/vm-from.wav", true)
			conn.Execute("say", fmt.Sprintf("en number iterated %s", msg.CallerIDNumber), true)
		}
		conn.Execute("playback", "voicemail/vm-received.wav", true)
		conn.Execute("say", fmt.Sprintf("en current_date_time pronounced %d",
			msg.RecordedAt.Unix()), true)

		// Play message
		if fileExists(msg.FilePath) {
			conn.Execute("playback", msg.FilePath, true)
		}

		// Per-message DTMF menu
		// 1 = replay, 2 = save, 7 = delete, 8 = forward, 9 = return call, # = next
	msgMenu:
		for {
			conn.Execute("playback", "voicemail/vm-listen_options.wav", true)
			digit := s.collectDigits(conn, uuid, "silence_stream://3000", 1, 1, 5000)

			switch digit {
			case "1": // Replay
				if fileExists(msg.FilePath) {
					conn.Execute("playback", msg.FilePath, true)
				}
				continue msgMenu

			case "2": // Save
				msg.MarkAsRead(db)
				db.Model(&models.VoicemailBox{}).Where("id = ?", box.ID).
					Updates(map[string]interface{}{
						"new_messages":   max(0, box.NewMessages-1),
						"saved_messages": box.SavedMessages + 1,
					})
				conn.Execute("playback", "voicemail/vm-saved.wav", true)
				// Update MWI
				s.sendMWI(manager, box.Extension, domain, max(0, box.NewMessages-1), box.SavedMessages+1)
				break msgMenu

			case "7": // Delete
				db.Delete(&msg)
				if msg.IsNew {
					db.Model(&models.VoicemailBox{}).Where("id = ?", box.ID).
						Update("new_messages", max(0, box.NewMessages-1))
				} else {
					db.Model(&models.VoicemailBox{}).Where("id = ?", box.ID).
						Update("saved_messages", max(0, box.SavedMessages-1))
				}
				// Remove file
				os.Remove(msg.FilePath)
				conn.Execute("playback", "voicemail/vm-deleted.wav", true)
				// Update MWI
				s.sendMWI(manager, box.Extension, domain, max(0, box.NewMessages-1), box.SavedMessages)
				break msgMenu

			case "8": // Forward to another extension
				conn.Execute("playback", "voicemail/vm-forward_enter_ext.wav", true)
				fwdExt := s.collectDigits(conn, uuid, "silence_stream://500", 2, 8, 5000)
				if fwdExt != "" {
					s.forwardMessage(db, &msg, fwdExt, box.TenantID)
					conn.Execute("playback", "voicemail/vm-forwarded.wav", true)
				}
				break msgMenu

			case "9": // Return call
				conn.Execute("playback", "voicemail/vm-return_call.wav", true)
				if msg.CallerIDNumber != "" {
					conn.Execute("transfer",
						fmt.Sprintf("%s XML %s", msg.CallerIDNumber, domain), false)
					return
				}
				break msgMenu

			case "#": // Next message
				if msg.IsNew {
					msg.MarkAsRead(db)
					db.Model(&models.VoicemailBox{}).Where("id = ?", box.ID).
						Updates(map[string]interface{}{
							"new_messages":   max(0, box.NewMessages-1),
							"saved_messages": box.SavedMessages + 1,
						})
					s.sendMWI(manager, box.Extension, domain, max(0, box.NewMessages-1), box.SavedMessages+1)
				}
				break msgMenu

			default:
				conn.Execute("playback", "voicemail/vm-invalid_value.wav", true)
			}
		}
	}

	conn.Execute("playback", "voicemail/vm-no_more_messages.wav", true)
}

// ========== Greeting Management ==========

// greetingMenu manages greeting recordings
func (s *Service) greetingMenu(conn *eventsocket.Connection, manager *esl.Manager,
	uuid string, box *models.VoicemailBox, domain string, logger *log.Entry) {

	db := manager.DB

	for {
		// 1 = record new greeting
		// 2 = listen to current greeting
		// 3 = delete greeting (use default)
		// * = return to main menu
		conn.Execute("playback", "voicemail/vm-greeting_options.wav", true)
		digit := s.collectDigits(conn, uuid, "silence_stream://3000", 1, 1, 5000)

		switch digit {
		case "1": // Record new greeting
			conn.Execute("playback", "voicemail/vm-record_greeting.wav", true)
			conn.Execute("playback", "tone_stream://%(200,0,800)", true)

			greetPath := s.getGreetingPath(box.TenantID, box.Extension)
			os.MkdirAll(filepath.Dir(greetPath), 0755)
			conn.Execute("record", fmt.Sprintf("%s 60 100 5", greetPath), true)

			// Confirm: 1 = keep, 2 = re-record, 3 = delete
			conn.Execute("playback", "voicemail/vm-review_recording.wav", true)
			review := s.collectDigits(conn, uuid, "silence_stream://3000", 1, 1, 5000)
			if review == "1" || review == "" {
				db.Model(box).Update("greeting_path", greetPath)
				conn.Execute("playback", "voicemail/vm-saved.wav", true)
			} else if review == "3" {
				os.Remove(greetPath)
				conn.Execute("playback", "voicemail/vm-deleted.wav", true)
			}
			// review == "2" loops back

		case "2": // Listen to current
			if box.GreetingPath != "" && fileExists(box.GreetingPath) {
				conn.Execute("playback", box.GreetingPath, true)
			} else {
				conn.Execute("playback", "voicemail/vm-no_greeting.wav", true)
			}

		case "3": // Delete greeting (use default)
			if box.GreetingPath != "" {
				os.Remove(box.GreetingPath)
				db.Model(box).Update("greeting_path", "")
				conn.Execute("playback", "voicemail/vm-deleted.wav", true)
			}

		case "*":
			return

		default:
			conn.Execute("playback", "voicemail/vm-invalid_value.wav", true)
		}
	}
}

// ========== MWI / BLF ==========

// sendMWI sends a Message Waiting Indicator event to FreeSWITCH
func (s *Service) sendMWI(manager *esl.Manager, extension, domain string, newMsgs, savedMsgs int) {
	if manager.Client == nil {
		return
	}

	// FreeSWITCH MWI event format
	event := fmt.Sprintf(`sendevent MESSAGE_WAITING
MWI-Messages-Waiting: %s
MWI-Message-Account: sip:%s@%s
MWI-Voice-Message: %d/%d (0/0)

`, boolToYesNo(newMsgs > 0), extension, domain, newMsgs, savedMsgs)

	manager.Client.Send(event)

	log.WithFields(log.Fields{
		"extension": extension,
		"new":       newMsgs,
		"saved":     savedMsgs,
	}).Debug("MWI sent")
}

// ========== Message Forwarding ==========

// forwardMessage copies a voicemail message to another extension's box
func (s *Service) forwardMessage(gdb *gorm.DB, msg *models.VoicemailMessage, targetExt string, tenantID uint) {

	var targetBox models.VoicemailBox
	if err := gdb.Where("extension = ? AND tenant_id = ? AND enabled = true",
		targetExt, tenantID).First(&targetBox).Error; err != nil {
		return
	}

	// Copy message record pointing to same file
	fwd := &models.VoicemailMessage{
		BoxID:          targetBox.ID,
		TenantID:       tenantID,
		CallerIDName:   msg.CallerIDName,
		CallerIDNumber: msg.CallerIDNumber,
		Duration:       msg.Duration,
		FilePath:       msg.FilePath, // Shared file reference
		FileSize:       msg.FileSize,
		RecordedAt:     msg.RecordedAt,
		IsNew:          true,
		ForwardedTo:    targetExt,
		ChannelUUID:    msg.ChannelUUID,
	}
	gdb.Create(fwd)

	// Update target box count
	gdb.Model(&targetBox).Update("new_messages", targetBox.NewMessages+1)

	// Update original message
	gdb.Model(msg).Update("forwarded_to", targetExt)
}

// ========== DTMF Helpers ==========

// collectDigits plays a prompt and collects DTMF digits
func (s *Service) collectDigits(conn *eventsocket.Connection, uuid, prompt string,
	minDigits, maxDigits, timeoutMs int) string {

	cmd := fmt.Sprintf("%d %d 1 %d # %s silence_stream://250 digits \\d+ %d",
		minDigits, maxDigits, timeoutMs, prompt, timeoutMs)
	conn.Execute("play_and_get_digits", cmd, true)

	ev, err := conn.Send("api uuid_getvar " + uuid + " digits")
	if err != nil {
		return ""
	}
	result := strings.TrimSpace(ev.Body)
	if result == "_undef_" {
		return ""
	}
	return result
}

// ========== Utility ==========

// getRecordingPath generates the file path for a recording
func (s *Service) getRecordingPath(tenantID uint, extension, uuid string) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s.wav", timestamp, uuid[:8])
	return filepath.Join(s.recordingDir, fmt.Sprintf("%d", tenantID), extension, filename)
}

// getGreetingPath generates the file path for a greeting
func (s *Service) getGreetingPath(tenantID uint, extension string) string {
	return filepath.Join(s.recordingDir, fmt.Sprintf("%d", tenantID), extension, "greeting.wav")
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
