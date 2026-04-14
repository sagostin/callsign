package handlers

import (
	"callsign/models"
	"callsign/services/encryption"
	"callsign/services/messaging"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SMSHandler handles SMS/MMS-related API requests
type SMSHandler struct {
	DB         *gorm.DB
	MsgManager *messaging.Manager
	EncMgr     *encryption.Manager
}

// NewSMSHandler creates a new SMS handler
func NewSMSHandler(db *gorm.DB, msgManager *messaging.Manager, encMgr *encryption.Manager) *SMSHandler {
	return &SMSHandler{DB: db, MsgManager: msgManager, EncMgr: encMgr}
}

// ListConversations returns message conversations for a tenant
func (h *SMSHandler) ListConversations(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var conversations []models.Conversation
	h.DB.Where("tenant_id = ?", tenantID).
		Order("last_message DESC").
		Limit(50).
		Find(&conversations)

	return c.JSON(conversations)
}

// GetConversation returns a specific conversation with messages
func (h *SMSHandler) GetConversation(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var conv models.Conversation
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages").First(&conv).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Conversation not found"})
	}

	return c.JSON(conv)
}

// SendMessage sends a new SMS/MMS message
func (h *SMSHandler) SendMessage(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var msg models.Message
	if err := c.BodyParser(&msg); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	msg.TenantID = tenantID
	msg.Direction = "outbound"
	msg.Status = "pending"
	if msg.Type == "" {
		msg.Type = "sms"
	}

	// Encrypt body before saving (fail loud if encryption fails)
	if msg.Body != "" && h.EncMgr != nil {
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

	// Queue for delivery via provider
	if h.MsgManager != nil {
		go h.MsgManager.SendMessage(tenantID, msg.From, msg.To, msg.Body, nil, 0)
	}

	return c.Status(http.StatusCreated).JSON(msg)
}

// SoundHandler handles sound/audio file API requests
type SoundHandler struct {
	DB *gorm.DB
}

// NewSoundHandler creates a new sound handler
func NewSoundHandler(db *gorm.DB) *SoundHandler {
	return &SoundHandler{DB: db}
}

// ListSounds returns all sounds for a tenant (+ system sounds)
func (h *SoundHandler) ListSounds(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var sounds []models.Sound
	h.DB.Where("tenant_id = ? OR tenant_id = 0 OR tenant_id IS NULL", tenantID).
		Order("category, name").Find(&sounds)

	return c.JSON(sounds)
}

// GetSound returns a specific sound
func (h *SoundHandler) GetSound(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var sound models.Sound
	if err := h.DB.Where("id = ? AND (tenant_id = ? OR tenant_id = 0)", id, tenantID).
		First(&sound).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Sound not found"})
	}

	return c.JSON(sound)
}

// CreateSound creates a new sound
func (h *SoundHandler) CreateSound(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var sound models.Sound
	if err := c.BodyParser(&sound); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	sound.TenantID = tenantID
	if err := h.DB.Create(&sound).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(sound)
}

// DeleteSound deletes a sound
func (h *SoundHandler) DeleteSound(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	// Cannot delete system sounds
	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Sound{})
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Sound not found or cannot delete system sound"})
	}

	c.Status(http.StatusNoContent)
	return nil
}

// PhraseHandler handles phrase API requests
type PhraseHandler struct {
	DB *gorm.DB
}

// NewPhraseHandler creates a new phrase handler
func NewPhraseHandler(db *gorm.DB) *PhraseHandler {
	return &PhraseHandler{DB: db}
}

// ListPhrases returns all phrases
func (h *PhraseHandler) ListPhrases(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var phrases []models.Phrase
	h.DB.Where("tenant_id = ? OR tenant_id = 0 OR tenant_id IS NULL", tenantID).
		Order("macro_name").Find(&phrases)

	return c.JSON(phrases)
}

// CreatePhrase creates a new phrase
func (h *PhraseHandler) CreatePhrase(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var phrase models.Phrase
	if err := c.BodyParser(&phrase); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	phrase.TenantID = tenantID
	if err := h.DB.Create(&phrase).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(phrase)
}

// ChatplanHandler handles chatplan API requests
type ChatplanHandler struct {
	DB *gorm.DB
}

// NewChatplanHandler creates a new chatplan handler
func NewChatplanHandler(db *gorm.DB) *ChatplanHandler {
	return &ChatplanHandler{DB: db}
}

// ListChatplans returns all chatplans for a tenant
func (h *ChatplanHandler) ListChatplans(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var chatplans []models.Chatplan
	h.DB.Where("tenant_id = ? OR tenant_id = 0", tenantID).
		Order("\"order\" ASC").Find(&chatplans)

	return c.JSON(chatplans)
}

// CreateChatplan creates a new chatplan
func (h *ChatplanHandler) CreateChatplan(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var chatplan models.Chatplan
	if err := c.BodyParser(&chatplan); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	chatplan.TenantID = tenantID
	if err := h.DB.Create(&chatplan).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(chatplan)
}

// UpdateChatplan updates a chatplan
func (h *ChatplanHandler) UpdateChatplan(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var existing models.Chatplan
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&existing).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Chatplan not found"})
	}

	if err := c.BodyParser(&existing); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	existing.TenantID = tenantID
	h.DB.Save(&existing)
	return c.JSON(existing)
}

// DeleteChatplan deletes a chatplan
func (h *ChatplanHandler) DeleteChatplan(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Chatplan{})
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Chatplan not found"})
	}

	c.Status(http.StatusNoContent)
	return nil
}

// DefaultOutboundRouteHandler handles system-level outbound routing
type DefaultOutboundRouteHandler struct {
	DB *gorm.DB
}

// NewDefaultOutboundRouteHandler creates a new handler
func NewDefaultOutboundRouteHandler(db *gorm.DB) *DefaultOutboundRouteHandler {
	return &DefaultOutboundRouteHandler{DB: db}
}

// ListRoutes returns all default outbound routes
func (h *DefaultOutboundRouteHandler) ListRoutes(c *fiber.Ctx) error {
	var routes []models.DefaultOutboundRoute
	h.DB.Order("\"order\" ASC").Find(&routes)
	return c.JSON(routes)
}

// CreateRoute creates a new default route (system admin only)
func (h *DefaultOutboundRouteHandler) CreateRoute(c *fiber.Ctx) error {
	var route models.DefaultOutboundRoute
	if err := c.BodyParser(&route); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.DB.Create(&route).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(route)
}

// UpdateRoute updates a default route
func (h *DefaultOutboundRouteHandler) UpdateRoute(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var route models.DefaultOutboundRoute
	if err := h.DB.First(&route, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Route not found"})
	}

	if err := c.BodyParser(&route); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	h.DB.Save(&route)
	return c.JSON(route)
}

// DeleteRoute deletes a default route
func (h *DefaultOutboundRouteHandler) DeleteRoute(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	result := h.DB.Delete(&models.DefaultOutboundRoute{}, id)
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Route not found"})
	}

	c.Status(http.StatusNoContent)
	return nil
}
