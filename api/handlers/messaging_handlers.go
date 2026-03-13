package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// =====================
// Messaging (SMS/MMS)
// =====================

func (h *Handler) ListConversations(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var conversations []models.Conversation
	h.DB.Where("tenant_id = ?", tenantID).
		Order("last_message_at DESC").
		Limit(50).
		Find(&conversations)

	return c.JSON(conversations)
}

func (h *Handler) GetConversation(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var conversation models.Conversation
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages").
		First(&conversation).Error; err != nil {
		h.logWarn("MESSAGING", "GetConversation: Conversation not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Conversation not found"})
	}

	return c.JSON(conversation)
}

func (h *Handler) SendMessage(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var msg models.Message
	if err := c.BodyParser(&msg); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	msg.TenantID = tenantID
	msg.Direction = "outbound"
	msg.Status = "pending"

	if err := h.DB.Create(&msg).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Queue for delivery via messaging provider
	if h.MsgManager != nil {
		go h.MsgManager.SendMessage(tenantID, msg.From, msg.To, msg.Body, nil, msg.ProviderID)
	}

	return c.Status(http.StatusCreated).JSON(msg)
}

// =====================
// Contacts
// =====================

func (h *Handler) ListContacts(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
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

func (h *Handler) CreateContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

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

func (h *Handler) GetContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "GetContact: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	return c.JSON(contact)
}

func (h *Handler) UpdateContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "UpdateContact: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	if err := c.BodyParser(&contact); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	contact.TenantID = tenantID
	h.DB.Save(&contact)
	return c.JSON(contact)
}

func (h *Handler) DeleteContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "DeleteContact: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	h.DB.Delete(&contact)
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) SyncContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "SyncContact: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	// TODO: Implement webhook sync logic
	return c.JSON(fiber.Map{"message": "Sync queued", "contact_id": id})
}

func (h *Handler) GetContactByPhone(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	phone := c.Query("phone")

	var contact models.Contact
	if err := h.DB.Where("tenant_id = ? AND (phone = ? OR mobile_phone = ? OR phone_alt = ?)",
		tenantID, phone, phone, phone).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "GetContactByPhone: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	return c.JSON(contact)
}

// =====================
// Chat System
// =====================

func (h *Handler) ListChatThreads(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
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

func (h *Handler) CreateChatThread(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

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

func (h *Handler) GetChatThread(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var thread models.ChatThread
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(100)
		}).
		Preload("Messages.Attachments").
		First(&thread).Error; err != nil {
		h.logWarn("MESSAGING", "GetChatThread: Thread not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Thread not found"})
	}

	return c.JSON(thread)
}

func (h *Handler) SendChatMessage(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	threadIDu64, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	threadID := uint(threadIDu64)

	// Verify thread exists
	var thread models.ChatThread
	if err := h.DB.Where("id = ? AND tenant_id = ?", threadID, tenantID).First(&thread).Error; err != nil {
		h.logWarn("MESSAGING", "SendChatMessage: Thread not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Thread not found"})
	}

	var msg models.ChatMessage
	if err := c.BodyParser(&msg); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	msg.TenantID = tenantID
	msg.ThreadID = threadID

	if err := h.DB.Create(&msg).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Update thread's last message time
	h.DB.Model(&thread).Update("last_message_at", msg.CreatedAt)

	return c.Status(http.StatusCreated).JSON(msg)
}

func (h *Handler) ListChatRooms(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var rooms []models.ChatRoom
	h.DB.Where("tenant_id = ? AND archived = false", tenantID).Find(&rooms)

	return c.JSON(rooms)
}

func (h *Handler) CreateChatRoom(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	claims := middleware.GetClaims(c)

	var room models.ChatRoom
	if err := c.BodyParser(&room); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	room.TenantID = tenantID
	if claims != nil {
		room.CreatedByID = claims.UserID
	}

	if err := h.DB.Create(&room).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(room)
}

func (h *Handler) JoinChatRoom(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	roomIDu64, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	roomID := uint(roomIDu64)
	claims := middleware.GetClaims(c)

	var room models.ChatRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", roomID, tenantID).First(&room).Error; err != nil {
		h.logWarn("MESSAGING", "JoinChatRoom: Room not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	// Check if already member
	var existing models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND extension_id = ?", roomID, claims.UserID).First(&existing).Error; err == nil {
		return c.JSON(fiber.Map{"message": "Already a member"})
	}

	member := models.ChatRoomMember{
		RoomID:      roomID,
		ExtensionID: claims.UserID,
		Role:        "member",
	}
	h.DB.Create(&member)

	return c.Status(http.StatusCreated).JSON(member)
}

func (h *Handler) ListChatQueues(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var queues []models.ChatQueue
	h.DB.Where("tenant_id = ?", tenantID).Preload("Agents").Find(&queues)

	return c.JSON(queues)
}

func (h *Handler) CreateChatQueue(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

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
