package handlers

import (
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// ChatHandler handles chat-related API requests
type ChatHandler struct {
	DB *gorm.DB
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *gorm.DB) *ChatHandler {
	return &ChatHandler{DB: db}
}

// ==================== Chat Threads ====================

// ListThreads returns chat threads for a tenant
func (h *ChatHandler) ListThreads(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	channel := ctx.URLParamDefault("channel", "")
	status := ctx.URLParamDefault("status", "open")

	query := h.DB.Where("tenant_id = ?", tenantID)
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var threads []models.ChatThread
	query.Order("last_message_at DESC").Limit(50).Find(&threads)
	ctx.JSON(threads)
}

// GetThread returns a thread with messages
func (h *ChatHandler) GetThread(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var thread models.ChatThread
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(100)
		}).
		Preload("Messages.Attachments").
		First(&thread).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Thread not found"})
		return
	}

	ctx.JSON(thread)
}

// CreateThread creates a new chat thread
func (h *ChatHandler) CreateThread(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var thread models.ChatThread
	if err := ctx.ReadJSON(&thread); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	thread.TenantID = tenantID
	if err := h.DB.Create(&thread).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(thread)
}

// SendMessage sends a message to a thread
func (h *ChatHandler) SendMessage(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	threadID, _ := ctx.Params().GetUint("id")

	// Verify thread exists
	var thread models.ChatThread
	if err := h.DB.Where("id = ? AND tenant_id = ?", threadID, tenantID).First(&thread).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Thread not found"})
		return
	}

	var msg models.ChatMessage
	if err := ctx.ReadJSON(&msg); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	msg.TenantID = tenantID
	msg.ThreadID = threadID

	// TODO: Encrypt body for external channels
	// if thread.Channel == models.ChannelSMS || thread.Channel == models.ChannelMMS {
	//     msg.BodyEncrypted = encMgr.Encrypt(msg.Body)
	// }

	if err := h.DB.Create(&msg).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Update thread's last message time
	h.DB.Model(&thread).Update("last_message_at", msg.CreatedAt)

	// TODO: Broadcast via WebSocket
	// TODO: For SMS/MMS, queue for delivery

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(msg)
}

// ==================== Chat Rooms ====================

// ListRooms returns chat rooms for a tenant
func (h *ChatHandler) ListRooms(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	extensionID := ctx.Values().GetUintDefault("extension_id", 0)

	var rooms []models.ChatRoom
	h.DB.Joins("JOIN chat_room_members ON chat_room_members.room_id = chat_rooms.id").
		Where("chat_rooms.tenant_id = ? AND chat_room_members.extension_id = ?", tenantID, extensionID).
		Or("chat_rooms.tenant_id = ? AND chat_rooms.is_public = true", tenantID).
		Where("chat_rooms.archived = false").
		Find(&rooms)

	ctx.JSON(rooms)
}

// CreateRoom creates a new chat room
func (h *ChatHandler) CreateRoom(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	extensionID := ctx.Values().GetUintDefault("extension_id", 0)

	var room models.ChatRoom
	if err := ctx.ReadJSON(&room); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	room.TenantID = tenantID
	room.CreatedByID = extensionID

	if err := h.DB.Create(&room).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Add creator as owner
	member := models.ChatRoomMember{
		RoomID:      room.ID,
		ExtensionID: extensionID,
		Role:        "owner",
	}
	h.DB.Create(&member)

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(room)
}

// JoinRoom adds an extension to a room
func (h *ChatHandler) JoinRoom(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	roomID, _ := ctx.Params().GetUint("id")
	extensionID := ctx.Values().GetUintDefault("extension_id", 0)

	var room models.ChatRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", roomID, tenantID).First(&room).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Room not found"})
		return
	}

	// Check if already member
	var existing models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND extension_id = ?", roomID, extensionID).First(&existing).Error; err == nil {
		ctx.JSON(iris.Map{"message": "Already a member"})
		return
	}

	member := models.ChatRoomMember{
		RoomID:      roomID,
		ExtensionID: extensionID,
		Role:        "member",
	}
	h.DB.Create(&member)

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(member)
}

// ==================== Chat Queues ====================

// ListQueues returns chat queues for a tenant
func (h *ChatHandler) ListQueues(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var queues []models.ChatQueue
	h.DB.Where("tenant_id = ?", tenantID).Preload("Agents").Find(&queues)

	ctx.JSON(queues)
}

// CreateQueue creates a new chat queue
func (h *ChatHandler) CreateQueue(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var queue models.ChatQueue
	if err := ctx.ReadJSON(&queue); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	queue.TenantID = tenantID
	if err := h.DB.Create(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(queue)
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
func (h *ContactHandler) ListContacts(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	search := ctx.URLParam("search")

	query := h.DB.Where("tenant_id = ?", tenantID)
	if search != "" {
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var contacts []models.Contact
	query.Order("last_name, first_name").Limit(100).Find(&contacts)
	ctx.JSON(contacts)
}

// GetContact returns a specific contact
func (h *ContactHandler) GetContact(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Contact not found"})
		return
	}

	ctx.JSON(contact)
}

// GetContactByPhone finds a contact by phone number
func (h *ContactHandler) GetContactByPhone(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	phone := ctx.URLParam("phone")

	var contact models.Contact
	if err := h.DB.Where("tenant_id = ? AND (phone = ? OR mobile_phone = ? OR phone_alt = ?)",
		tenantID, phone, phone, phone).First(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Contact not found"})
		return
	}

	ctx.JSON(contact)
}

// CreateContact creates a new contact
func (h *ContactHandler) CreateContact(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var contact models.Contact
	if err := ctx.ReadJSON(&contact); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	contact.TenantID = tenantID
	if err := h.DB.Create(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(contact)
}

// UpdateContact updates a contact
func (h *ContactHandler) UpdateContact(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Contact not found"})
		return
	}

	if err := ctx.ReadJSON(&contact); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	contact.TenantID = tenantID
	h.DB.Save(&contact)
	ctx.JSON(contact)
}

// SyncContact triggers a webhook sync for a contact
func (h *ContactHandler) SyncContact(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Contact not found"})
		return
	}

	// TODO: Implement webhook sync logic
	// 1. Find associated ContactWebhook by contact.ExternalSource
	// 2. Fetch data from webhook URL with ExternalID
	// 3. Map fields and update contact

	ctx.JSON(iris.Map{"message": "Sync queued", "contact_id": id})
}

// WebhookIngestContact handles inbound webhook data for contacts
func (h *ContactHandler) WebhookIngestContact(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	source := ctx.Params().GetString("source")

	// Parse incoming data
	var data map[string]interface{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Find webhook config for this source
	var webhook models.ContactWebhook
	if err := h.DB.Where("tenant_id = ? AND source = ? AND enabled = true", tenantID, source).
		First(&webhook).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Webhook source not found"})
		return
	}

	// TODO: Apply field mapping and upsert contact
	// 1. Parse webhook.FieldMapping
	// 2. Extract fields from data
	// 3. Find or create contact by ExternalID
	// 4. Update contact fields
	// 5. Save

	ctx.JSON(iris.Map{"message": "Contact ingested", "source": source})
}
