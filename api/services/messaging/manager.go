package messaging

import (
	"callsign/config"
	"callsign/models"
	"callsign/services/websocket"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Manager orchestrates all messaging operations
type Manager struct {
	DB         *gorm.DB
	Config     *config.Config
	Hub        *websocket.Hub
	Transcoder *Transcoder
	Queue      *QueueWorker
	providers  map[uint]SMSProvider // providerID -> provider
}

// NewManager creates a new messaging manager
func NewManager(db *gorm.DB, cfg *config.Config, hub *websocket.Hub) *Manager {
	// Build transcoding config from app config
	tcConfig := TranscodeConfig{
		MaxSizeKB:  cfg.MaxMMSSizeKB,
		MaxWidth:   1920,
		MaxHeight:  1920,
		FFmpegPath: cfg.FFmpegPath,
		TmpDir:     cfg.TranscodeTmpDir,
	}
	if tcConfig.MaxSizeKB == 0 {
		tcConfig.MaxSizeKB = 600
	}

	providers := make(map[uint]SMSProvider)

	m := &Manager{
		DB:         db,
		Config:     cfg,
		Hub:        hub,
		Transcoder: NewTranscoder(tcConfig),
		providers:  providers,
	}

	// Load providers from database and initialize
	m.loadProviders()

	// Create queue worker with loaded providers
	m.Queue = NewQueueWorker(db, providers)

	return m
}

// Start initializes and starts background workers
func (m *Manager) Start() {
	m.Queue.Start()
	log.Info("Messaging manager started")
}

// Stop gracefully shuts down the messaging manager
func (m *Manager) Stop() {
	m.Queue.Stop()
	log.Info("Messaging manager stopped")
}

// loadProviders loads messaging providers from the database and initializes their API clients
func (m *Manager) loadProviders() {
	var dbProviders []models.MessagingProvider
	m.DB.Where("enabled = true").Find(&dbProviders)

	for _, p := range dbProviders {
		switch p.Type {
		case models.ProviderTelnyx:
			// Use DB credentials or fall back to config
			apiKey := m.Config.TelnyxAPIKey
			msgProfile := m.Config.TelnyxMessagingProfile
			provider := NewTelnyxProvider(apiKey, msgProfile)
			m.providers[p.ID] = provider
			log.WithField("provider_id", p.ID).Info("Loaded Telnyx messaging provider")

		default:
			log.WithFields(log.Fields{
				"provider_id":   p.ID,
				"provider_type": p.Type,
			}).Warn("Unsupported messaging provider type")
		}
	}

	log.WithField("count", len(m.providers)).Info("Messaging providers loaded")
}

// GetProvider returns a provider by ID
func (m *Manager) GetProvider(providerID uint) (SMSProvider, bool) {
	p, ok := m.providers[providerID]
	return p, ok
}

// SendMessage is the primary API for sending outbound SMS/MMS
// It creates queue items and lets the queue worker handle delivery
func (m *Manager) SendMessage(tenantID uint, from, to, body string, media []MediaItem, providerID uint) error {
	hasMedia := len(media) > 0

	// If there's media, transcode if needed before queuing
	if hasMedia {
		for i, item := range media {
			if len(item.Data) > 0 && m.Transcoder.NeedsTranscoding(item.Data, item.ContentType) {
				result, err := m.Transcoder.Transcode(item.Data, item.ContentType)
				if err != nil {
					log.WithError(err).Warn("Media transcoding failed, sending original")
					continue
				}
				media[i].Data = result.Data
				media[i].ContentType = result.ContentType
			}
		}
	}

	// Enqueue for delivery
	return m.Queue.Enqueue(tenantID, providerID, from, to, body, hasMedia, nil, nil)
}

// RouteInboundSMS processes an inbound SMS and routes it to the correct thread
func (m *Manager) RouteInboundSMS(toNumber, fromNumber, body string, mediaURLs []string) error {
	// Find the Destination (DID) being messaged
	var dest models.Destination
	if err := m.DB.Where("destination_number = ? AND sms_enabled = true", toNumber).First(&dest).Error; err != nil {
		return fmt.Errorf("no SMS-enabled destination found for %s: %w", toNumber, err)
	}

	tenantID := dest.TenantID

	// Resolve contact from phone number
	var contact models.Contact
	contactFound := m.DB.Where("tenant_id = ? AND (phone = ? OR mobile_phone = ? OR phone_alt = ?)",
		tenantID, fromNumber, fromNumber, fromNumber).First(&contact).Error == nil

	senderName := fromNumber
	var contactID *uint
	if contactFound {
		senderName = contact.FullName()
		contactID = &contact.ID
	}

	switch dest.SMSMode {
	case "shared":
		return m.routeSharedSMS(tenantID, toNumber, fromNumber, body, senderName, contactID, mediaURLs)
	case "dedicated":
		return m.routeDedicatedSMS(tenantID, &dest, fromNumber, body, senderName, contactID, mediaURLs)
	default:
		return fmt.Errorf("SMS mode '%s' not supported for destination %s", dest.SMSMode, toNumber)
	}
}

// routeSharedSMS routes an inbound SMS to a group chat thread visible to all tenant users
func (m *Manager) routeSharedSMS(tenantID uint, toNumber, fromNumber, body, senderName string, contactID *uint, mediaURLs []string) error {
	// Find or create a group SMS thread for this conversation
	var thread models.ChatThread
	err := m.DB.Where(
		"tenant_id = ? AND channel = 'sms' AND is_group_sms = true AND group_sms_number = ? AND remote_number = ?",
		tenantID, toNumber, fromNumber,
	).First(&thread).Error

	if err == gorm.ErrRecordNotFound {
		thread = models.ChatThread{
			TenantID:       tenantID,
			Channel:        models.ChannelSMS,
			LocalNumber:    toNumber,
			RemoteNumber:   fromNumber,
			IsGroupSMS:     true,
			GroupSMSNumber: toNumber,
			ContactID:      contactID,
			Status:         "open",
		}
		if err := m.DB.Create(&thread).Error; err != nil {
			return fmt.Errorf("failed to create group SMS thread: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to find thread: %w", err)
	}

	// Create the chat message
	msg := models.ChatMessage{
		TenantID:    tenantID,
		ThreadID:    thread.ID,
		SenderType:  "contact",
		SenderName:  senderName,
		ContentType: "text",
		Body:        body,
		Status:      "delivered",
	}
	if contactID != nil {
		msg.SenderID = *contactID
	}

	if err := m.DB.Create(&msg).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Store media attachments
	m.storeMediaAttachments(tenantID, msg.ID, mediaURLs)

	// Update thread
	now := time.Now()
	m.DB.Model(&thread).Updates(map[string]interface{}{
		"last_message_at": &now,
		"status":          "open",
	})

	// Broadcast via WebSocket — visible to ALL users in tenant
	if m.Hub != nil {
		m.Hub.BroadcastToTenant(tenantID, websocket.EventChat, "new_message", map[string]interface{}{
			"thread_id":    thread.ID,
			"message_id":   msg.ID,
			"sender_name":  senderName,
			"sender_type":  "contact",
			"body":         body,
			"channel":      "sms",
			"is_group_sms": true,
			"from":         fromNumber,
			"to":           toNumber,
		})
	}

	return nil
}

// routeDedicatedSMS routes an inbound SMS to the specific extension that owns the number
func (m *Manager) routeDedicatedSMS(tenantID uint, dest *models.Destination, fromNumber, body, senderName string, contactID *uint, mediaURLs []string) error {
	// Find the extension assignment for this number
	var assignment models.SMSNumberAssignment
	if err := m.DB.Where("destination_id = ? AND tenant_id = ? AND enabled = true",
		dest.ID, tenantID).First(&assignment).Error; err != nil {
		return fmt.Errorf("no SMS assignment found for destination %d: %w", dest.ID, err)
	}

	// Find or create a private conversation for this extension + remote number
	var conv models.Conversation
	err := m.DB.Where(
		"tenant_id = ? AND extension_id = ? AND local_number = ? AND remote_number = ?",
		tenantID, assignment.ExtensionID, dest.DestinationNumber, fromNumber,
	).First(&conv).Error

	if err == gorm.ErrRecordNotFound {
		conv = models.Conversation{
			TenantID:     tenantID,
			ExtensionID:  assignment.ExtensionID,
			LocalNumber:  dest.DestinationNumber,
			RemoteNumber: fromNumber,
			Status:       "active",
			LastMessage:  time.Now(),
		}
		if err := m.DB.Create(&conv).Error; err != nil {
			return fmt.Errorf("failed to create conversation: %w", err)
		}
	}

	// Create the SMS message
	smsMsg := models.Message{
		TenantID:       tenantID,
		ExtensionID:    assignment.ExtensionID,
		ConversationID: &conv.ID,
		Type:           "sms",
		Direction:      "inbound",
		From:           fromNumber,
		To:             dest.DestinationNumber,
		Body:           body,
		Status:         "delivered",
	}
	if err := m.DB.Create(&smsMsg).Error; err != nil {
		return fmt.Errorf("failed to create SMS message: %w", err)
	}

	// Update conversation
	m.DB.Model(&conv).Updates(map[string]interface{}{
		"last_message": time.Now(),
		"unread_count": gorm.Expr("unread_count + 1"),
	})

	// Broadcast via WebSocket — only to the specific user's tenant (they filter by extension)
	if m.Hub != nil {
		m.Hub.NotifySMS(tenantID, map[string]interface{}{
			"conversation_id": conv.ID,
			"message_id":      smsMsg.ID,
			"extension_id":    assignment.ExtensionID,
			"from":            fromNumber,
			"to":              dest.DestinationNumber,
			"body":            body,
			"sender_name":     senderName,
			"direction":       "inbound",
		})
	}

	return nil
}

// storeMediaAttachments downloads and stores media from URLs as ChatAttachments
func (m *Manager) storeMediaAttachments(tenantID, messageID uint, mediaURLs []string) {
	for _, url := range mediaURLs {
		attachment := models.ChatAttachment{
			MessageID:   messageID,
			TenantID:    tenantID,
			FileName:    "media",
			ContentType: "application/octet-stream",
			ExternalURL: url,
		}
		if err := m.DB.Create(&attachment).Error; err != nil {
			log.WithError(err).Error("Failed to store media attachment")
		}
	}
}

// HandleStatusUpdate processes a delivery status webhook and updates message status
func (m *Manager) HandleStatusUpdate(status *StatusUpdate) error {
	// Find the queue item by provider message ID
	var item models.MessageQueueItem
	if err := m.DB.Where("provider_message_id = ?", status.MessageID).First(&item).Error; err != nil {
		log.WithField("provider_message_id", status.MessageID).Debug("Status update for unknown message, ignoring")
		return nil
	}

	// Update queue item status
	m.DB.Model(&item).Updates(map[string]interface{}{
		"status": status.Status,
	})

	// Update source message
	if item.MessageID != nil {
		updates := map[string]interface{}{"status": status.Status}
		if status.Status == "delivered" {
			now := time.Now()
			updates["delivered_at"] = &now
		}
		if status.Status == "failed" {
			updates["error_code"] = status.ErrorCode
			updates["error_message"] = status.ErrorMsg
		}
		m.DB.Model(&models.Message{}).Where("id = ?", *item.MessageID).Updates(updates)
	}

	if item.ChatMessageID != nil {
		updates := map[string]interface{}{"status": status.Status}
		if status.Status == "delivered" {
			now := time.Now()
			updates["delivered_at"] = &now
		}
		if status.Status == "failed" {
			updates["failed_reason"] = status.ErrorMsg
		}
		m.DB.Model(&models.ChatMessage{}).Where("id = ?", *item.ChatMessageID).Updates(updates)
	}

	// Broadcast status update via WebSocket
	if m.Hub != nil {
		m.Hub.BroadcastToTenant(item.TenantID, websocket.EventSMS, "status_update", map[string]interface{}{
			"queue_item_id":       item.ID,
			"provider_message_id": status.MessageID,
			"status":              status.Status,
			"error_code":          status.ErrorCode,
		})
	}

	return nil
}
