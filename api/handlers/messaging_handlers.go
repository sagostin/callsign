package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// =====================
// Messaging (SMS/MMS)
// =====================

func (h *Handler) ListConversations(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var conversations []models.Conversation
	h.DB.Where("tenant_id = ?", tenantID).
		Order("last_message_at DESC").
		Limit(50).
		Find(&conversations)

	ctx.JSON(conversations)
}

func (h *Handler) GetConversation(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var conversation models.Conversation
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages").
		First(&conversation).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conversation not found"})
		return
	}

	ctx.JSON(conversation)
}

func (h *Handler) SendMessage(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var msg models.Message
	if err := ctx.ReadJSON(&msg); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	msg.TenantID = tenantID
	msg.Direction = "outbound"
	msg.Status = "pending"

	if err := h.DB.Create(&msg).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// TODO: Queue for delivery via messaging provider

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(msg)
}

// =====================
// Contacts
// =====================

func (h *Handler) ListContacts(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
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

func (h *Handler) CreateContact(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

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

func (h *Handler) GetContact(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Contact not found"})
		return
	}

	ctx.JSON(contact)
}

func (h *Handler) UpdateContact(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
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

func (h *Handler) DeleteContact(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Contact not found"})
		return
	}

	h.DB.Delete(&contact)
	ctx.StatusCode(http.StatusNoContent)
}

func (h *Handler) SyncContact(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Contact not found"})
		return
	}

	// TODO: Implement webhook sync logic
	ctx.JSON(iris.Map{"message": "Sync queued", "contact_id": id})
}

func (h *Handler) GetContactByPhone(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
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

// =====================
// Chat System
// =====================

func (h *Handler) ListChatThreads(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
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

func (h *Handler) CreateChatThread(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

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

func (h *Handler) GetChatThread(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
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

func (h *Handler) SendChatMessage(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
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

	if err := h.DB.Create(&msg).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Update thread's last message time
	h.DB.Model(&thread).Update("last_message_at", msg.CreatedAt)

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(msg)
}

func (h *Handler) ListChatRooms(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var rooms []models.ChatRoom
	h.DB.Where("tenant_id = ? AND archived = false", tenantID).Find(&rooms)

	ctx.JSON(rooms)
}

func (h *Handler) CreateChatRoom(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	claims := middleware.GetClaims(ctx)

	var room models.ChatRoom
	if err := ctx.ReadJSON(&room); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	room.TenantID = tenantID
	if claims != nil {
		room.CreatedByID = claims.UserID
	}

	if err := h.DB.Create(&room).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(room)
}

func (h *Handler) JoinChatRoom(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	roomID, _ := ctx.Params().GetUint("id")
	claims := middleware.GetClaims(ctx)

	var room models.ChatRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", roomID, tenantID).First(&room).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Room not found"})
		return
	}

	// Check if already member
	var existing models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND extension_id = ?", roomID, claims.UserID).First(&existing).Error; err == nil {
		ctx.JSON(iris.Map{"message": "Already a member"})
		return
	}

	member := models.ChatRoomMember{
		RoomID:      roomID,
		ExtensionID: claims.UserID,
		Role:        "member",
	}
	h.DB.Create(&member)

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(member)
}

func (h *Handler) ListChatQueues(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var queues []models.ChatQueue
	h.DB.Where("tenant_id = ?", tenantID).Preload("Agents").Find(&queues)

	ctx.JSON(queues)
}

func (h *Handler) CreateChatQueue(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

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
