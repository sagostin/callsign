package handlers

import (
	"callsign/models"
	"callsign/services/encryption"
	"callsign/services/messaging"
	"callsign/services/websocket"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ChatHandler handles chat-related API requests
type ChatHandler struct {
	DB         *gorm.DB
	WSHub      *websocket.Hub
	MsgManager *messaging.Manager
	EncMgr     *encryption.Manager
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *gorm.DB, wsHub *websocket.Hub, msgManager *messaging.Manager, encMgr *encryption.Manager) *ChatHandler {
	return &ChatHandler{DB: db, WSHub: wsHub, MsgManager: msgManager, EncMgr: encMgr}
}

// ==================== Chat Threads ====================

// ListThreads returns chat threads for a tenant
func (h *ChatHandler) ListThreads(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	channel := c.Query("channel", "")
	status := c.Query("status", "open")

	query := h.DB.Where("tenant_id = ?", tenantID)
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var threads []models.ChatThread
	query.Order("last_message_at DESC").Limit(50).Find(&threads)
	return c.JSON(threads)
}

// GetThread returns a thread with messages
func (h *ChatHandler) GetThread(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var thread models.ChatThread
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(100)
		}).
		Preload("Messages.Attachments").
		First(&thread).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Thread not found"})
	}

	return c.JSON(thread)
}

// CreateThread creates a new chat thread
func (h *ChatHandler) CreateThread(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var thread models.ChatThread
	if err := c.BodyParser(&thread); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	thread.TenantID = tenantID
	if err := h.DB.Create(&thread).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(thread)
}

// SendMessage sends a message to a thread
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	threadIDu64, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid thread ID"})
	}
	threadID := uint(threadIDu64)

	// Verify thread exists
	var thread models.ChatThread
	if err := h.DB.Where("id = ? AND tenant_id = ?", threadID, tenantID).First(&thread).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Thread not found"})
	}

	var msg models.ChatMessage
	if err := c.BodyParser(&msg); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	msg.TenantID = tenantID
	msg.ThreadID = threadID

	// Encrypt body for external channels (fail loud if encryption fails)
	if msg.Body != "" && h.EncMgr != nil && (thread.Channel == models.ChannelSMS || thread.Channel == models.ChannelMMS) {
		encrypted, err := h.EncMgr.Encrypt(msg.Body)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encrypt message"})
		}
		msg.BodyEncrypted = encrypted
		msg.BodyHash = h.EncMgr.HashForLookup(msg.Body)
		msg.Body = "" // Clear plaintext after encryption
	}

	if err := h.DB.Create(&msg).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Update thread's last message time
	h.DB.Model(&thread).Update("last_message_at", msg.CreatedAt)

	// Broadcast via WebSocket
	if h.WSHub != nil {
		h.WSHub.NotifyChatMessage(tenantID, threadID, map[string]interface{}{
			"id":          msg.ID,
			"sender_id":   msg.SenderID,
			"sender_type": msg.SenderType,
			"body":        msg.Body,
			"channel":     thread.Channel,
		})
	}

	// For SMS/MMS threads, queue outbound messages for delivery via provider
	if h.MsgManager != nil && (thread.Channel == "sms" || thread.Channel == "mms") && msg.SenderType == "extension" {
		go h.MsgManager.SendMessage(tenantID, thread.LocalNumber, thread.RemoteNumber, msg.Body, nil, 0)
	}

	return c.Status(http.StatusCreated).JSON(msg)
}

// ==================== Chat Rooms ====================

// ListRooms returns chat rooms for a tenant
func (h *ChatHandler) ListRooms(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	extensionID := getLocalsUint(c, "extension_id", 0)

	var rooms []models.ChatRoom
	h.DB.Joins("JOIN chat_room_members ON chat_room_members.room_id = chat_rooms.id").
		Where("chat_rooms.tenant_id = ? AND chat_room_members.extension_id = ?", tenantID, extensionID).
		Or("chat_rooms.tenant_id = ? AND chat_rooms.is_public = true", tenantID).
		Where("chat_rooms.archived = false").
		Find(&rooms)

	return c.JSON(rooms)
}

// CreateRoom creates a new chat room
func (h *ChatHandler) CreateRoom(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	extensionID := getLocalsUint(c, "extension_id", 0)

	var room models.ChatRoom
	if err := c.BodyParser(&room); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	room.TenantID = tenantID
	room.CreatedByID = extensionID

	if err := h.DB.Create(&room).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Add creator as owner
	member := models.ChatRoomMember{
		RoomID:      room.ID,
		ExtensionID: extensionID,
		Role:        "owner",
	}
	h.DB.Create(&member)

	return c.Status(http.StatusCreated).JSON(room)
}

// JoinRoom adds an extension to a room
func (h *ChatHandler) JoinRoom(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	roomIDu64, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid room ID"})
	}
	roomID := uint(roomIDu64)
	extensionID := getLocalsUint(c, "extension_id", 0)

	var room models.ChatRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", roomID, tenantID).First(&room).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	// Check if already member
	var existing models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND extension_id = ?", roomID, extensionID).First(&existing).Error; err == nil {
		return c.JSON(fiber.Map{"message": "Already a member"})
	}

	member := models.ChatRoomMember{
		RoomID:      roomID,
		ExtensionID: extensionID,
		Role:        "member",
	}
	h.DB.Create(&member)

	return c.Status(http.StatusCreated).JSON(member)
}

// ==================== Chat Queues ====================

// ListQueues returns chat queues for a tenant
func (h *ChatHandler) ListQueues(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var queues []models.ChatQueue
	h.DB.Where("tenant_id = ?", tenantID).Preload("Agents").Find(&queues)

	return c.JSON(queues)
}

// CreateQueue creates a new chat queue
func (h *ChatHandler) CreateQueue(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var queue models.ChatQueue
	if err := c.BodyParser(&queue); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	queue.TenantID = tenantID
	if err := h.DB.Create(&queue).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(queue)
}

// ==================== Contacts ====================

// ContactHandler handles contact-related API requests
type ContactHandler struct {
	DB *gorm.DB
}

// NewContactHandler creates a new contact handler
func NewContactHandler(db *gorm.DB) *ContactHandler {
	return &ContactHandler{DB: db}
}

// ListContacts returns contacts for a tenant
func (h *ContactHandler) ListContacts(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	search := c.Query("search")

	query := h.DB.Where("tenant_id = ?", tenantID)
	if search != "" {
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var contacts []models.Contact
	query.Order("last_name, first_name").Limit(100).Find(&contacts)
	return c.JSON(contacts)
}

// GetContact returns a specific contact
func (h *ContactHandler) GetContact(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	return c.JSON(contact)
}

// GetContactByPhone finds a contact by phone number
func (h *ContactHandler) GetContactByPhone(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	phone := c.Query("phone")

	var contact models.Contact
	if err := h.DB.Where("tenant_id = ? AND (phone = ? OR mobile_phone = ? OR phone_alt = ?)",
		tenantID, phone, phone, phone).First(&contact).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	return c.JSON(contact)
}

// CreateContact creates a new contact
func (h *ContactHandler) CreateContact(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var contact models.Contact
	if err := c.BodyParser(&contact); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	contact.TenantID = tenantID
	if err := h.DB.Create(&contact).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(contact)
}

// UpdateContact updates a contact
func (h *ContactHandler) UpdateContact(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	if err := c.BodyParser(&contact); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	contact.TenantID = tenantID
	h.DB.Save(&contact)
	return c.JSON(contact)
}

// SyncContact triggers a webhook sync for a contact
func (h *ContactHandler) SyncContact(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	// Guard: contact must have external source for webhook sync
	if contact.ExternalSource == "" || contact.ExternalID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Contact has no external source configured"})
	}

	// Find associated ContactWebhook by contact.ExternalSource
	var webhook models.ContactWebhook
	if err := h.DB.Where("tenant_id = ? AND source = ? AND enabled = true", tenantID, contact.ExternalSource).
		First(&webhook).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Webhook source not found"})
	}

	// Guard: webhook must have a fetch URL
	if webhook.FetchURL == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Webhook has no fetch URL configured"})
	}

	// Build fetch URL with ExternalID
	fetchURL := strings.ReplaceAll(webhook.FetchURL, "{{external_id}}", contact.ExternalID)

	// Create HTTP request
	req, err := http.NewRequest(webhook.FetchMethod, fetchURL, nil)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create webhook request"})
	}

	// Apply custom headers if configured
	if webhook.FetchHeaders != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(webhook.FetchHeaders), &headers); err == nil {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		}
	}

	// Execute request with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch webhook data"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": "Webhook returned non-200 status"})
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read webhook response"})
	}

	var webhookData map[string]interface{}
	if err := json.Unmarshal(body, &webhookData); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse webhook response"})
	}

	// Apply field mapping if configured
	if webhook.FieldMapping != "" {
		var fieldMapping map[string]string
		if err := json.Unmarshal([]byte(webhook.FieldMapping), &fieldMapping); err == nil {
			h.applyFieldMapping(&contact, webhookData, fieldMapping)
		}
	}

	// Update sync metadata
	now := time.Now()
	contact.LastSyncAt = &now

	// Save updated contact
	if err := h.DB.Save(&contact).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save contact"})
	}

	return c.JSON(fiber.Map{"message": "Sync completed", "contact_id": id})
}

// WebhookIngestContact handles inbound webhook data for contacts
func (h *ContactHandler) WebhookIngestContact(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	source := c.Params("source")

	// Parse incoming data
	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Find webhook config for this source
	var webhook models.ContactWebhook
	if err := h.DB.Where("tenant_id = ? AND source = ? AND enabled = true", tenantID, source).
		First(&webhook).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Webhook source not found"})
	}

	// Extract external_id from data using field mapping if present
	externalID := ""
	if webhook.FieldMapping != "" {
		var fieldMapping map[string]string
		if err := json.Unmarshal([]byte(webhook.FieldMapping), &fieldMapping); err == nil {
			if extIDField, ok := fieldMapping["external_id"]; ok {
				if val, ok := data[extIDField].(string); ok {
					externalID = val
				}
			}
		}
	}

	// Fallback: try common field names for external_id
	if externalID == "" {
		if val, ok := data["external_id"].(string); ok {
			externalID = val
		} else if val, ok := data["id"].(string); ok {
			externalID = val
		}
	}

	// Guard: external_id is required for upsert
	if externalID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "External ID not found in webhook data"})
	}

	// Find or create contact by ExternalID
	var contact models.Contact
	findErr := h.DB.Where("tenant_id = ? AND external_source = ? AND external_id = ?", tenantID, source, externalID).First(&contact).Error

	if findErr != nil {
		// Create new contact if not found
		if findErr == gorm.ErrRecordNotFound {
			contact = models.Contact{
				TenantID:       tenantID,
				ExternalSource: source,
				ExternalID:     externalID,
				Status:         "active",
			}
		} else {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}
	}

	// Apply field mapping if configured
	if webhook.FieldMapping != "" {
		var fieldMapping map[string]string
		if err := json.Unmarshal([]byte(webhook.FieldMapping), &fieldMapping); err == nil {
			h.applyFieldMapping(&contact, data, fieldMapping)
		}
	}

	// Store raw external data as JSON blob
	rawJSON, _ := json.Marshal(data)
	contact.ExternalData = string(rawJSON)

	// Update sync metadata
	now := time.Now()
	contact.LastSyncAt = &now

	// Save contact (create or update)
	if err := h.DB.Save(&contact).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save contact"})
	}

	return c.Status(http.StatusCreated).JSON(contact)
}

// applyFieldMapping maps webhook data fields to contact fields based on mapping config
func (h *ContactHandler) applyFieldMapping(contact *models.Contact, data map[string]interface{}, fieldMapping map[string]string) {
	// Direct field assignments based on common mapping keys
	fieldTypes := map[string]*string{
		"first_name":         &contact.FirstName,
		"last_name":          &contact.LastName,
		"display_name":       &contact.DisplayName,
		"company":            &contact.Company,
		"title":              &contact.Title,
		"email":              &contact.Email,
		"phone":              &contact.Phone,
		"phone_alt":          &contact.PhoneAlt,
		"mobile_phone":       &contact.MobilePhone,
		"address1":           &contact.Address1,
		"address2":           &contact.Address2,
		"city":               &contact.City,
		"state":              &contact.State,
		"postal_code":        &contact.PostalCode,
		"country":            &contact.Country,
		"external_id":        &contact.ExternalID,
		"external_data":      &contact.ExternalData,
		"notes":              &contact.Notes,
		"preferred_channel":  &contact.PreferredChannel,
		"preferred_language": &contact.PreferredLanguage,
		"timezone":           &contact.Timezone,
	}

	for webhookField, contactField := range fieldMapping {
		// Skip special fields
		if webhookField == "external_id" || webhookField == "external_data" {
			continue
		}

		if targetPtr, exists := fieldTypes[contactField]; exists {
			if val, ok := data[webhookField].(string); ok {
				*targetPtr = val
			}
		}
	}
}
