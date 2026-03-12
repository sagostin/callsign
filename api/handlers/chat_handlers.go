package handlers

import (
	"callsign/models"
	"callsign/services/messaging"
	"callsign/services/websocket"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ChatHandler handles chat-related API requests
type ChatHandler struct {
	DB         *gorm.DB
	WSHub      *websocket.Hub
	MsgManager *messaging.Manager
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *gorm.DB, wsHub *websocket.Hub, msgManager *messaging.Manager) *ChatHandler {
	return &ChatHandler{DB: db, WSHub: wsHub, MsgManager: msgManager}
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

	// TODO: Encrypt body for external channels
	// if thread.Channel == models.ChannelSMS || thread.Channel == models.ChannelMMS {
	//     msg.BodyEncrypted = encMgr.Encrypt(msg.Body)
	// }

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

	// TODO: Implement webhook sync logic
	// 1. Find associated ContactWebhook by contact.ExternalSource
	// 2. Fetch data from webhook URL with ExternalID
	// 3. Map fields and update contact

	return c.JSON(fiber.Map{"message": "Sync queued", "contact_id": id})
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

	// TODO: Apply field mapping and upsert contact
	// 1. Parse webhook.FieldMapping
	// 2. Extract fields from data
	// 3. Find or create contact by ExternalID
	// 4. Update contact fields
	// 5. Save

	return c.JSON(fiber.Map{"message": "Contact ingested", "source": source})
}
