package handlers

import (
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// SMSHandler handles SMS/MMS-related API requests
type SMSHandler struct {
	DB *gorm.DB
}

// NewSMSHandler creates a new SMS handler
func NewSMSHandler(db *gorm.DB) *SMSHandler {
	return &SMSHandler{DB: db}
}

// ListConversations returns message conversations for a tenant
func (h *SMSHandler) ListConversations(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var conversations []models.Conversation
	h.DB.Where("tenant_id = ?", tenantID).
		Order("last_message DESC").
		Limit(50).
		Find(&conversations)

	ctx.JSON(conversations)
}

// GetConversation returns a specific conversation with messages
func (h *SMSHandler) GetConversation(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var conv models.Conversation
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages").First(&conv).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conversation not found"})
		return
	}

	ctx.JSON(conv)
}

// SendMessage sends a new SMS/MMS message
func (h *SMSHandler) SendMessage(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var msg models.Message
	if err := ctx.ReadJSON(&msg); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	msg.TenantID = tenantID
	msg.Direction = "outbound"
	msg.Status = "pending"
	if msg.Type == "" {
		msg.Type = "sms"
	}

	// TODO: Encrypt body before saving
	// msg.BodyEncrypted = encMgr.Encrypt(msg.Body)

	if err := h.DB.Create(&msg).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// TODO: Queue for delivery via provider

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(msg)
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
func (h *SoundHandler) ListSounds(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var sounds []models.Sound
	h.DB.Where("tenant_id = ? OR tenant_id = 0 OR tenant_id IS NULL", tenantID).
		Order("category, name").Find(&sounds)

	ctx.JSON(sounds)
}

// GetSound returns a specific sound
func (h *SoundHandler) GetSound(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var sound models.Sound
	if err := h.DB.Where("id = ? AND (tenant_id = ? OR tenant_id = 0)", id, tenantID).
		First(&sound).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Sound not found"})
		return
	}

	ctx.JSON(sound)
}

// CreateSound creates a new sound
func (h *SoundHandler) CreateSound(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var sound models.Sound
	if err := ctx.ReadJSON(&sound); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	sound.TenantID = tenantID
	if err := h.DB.Create(&sound).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(sound)
}

// DeleteSound deletes a sound
func (h *SoundHandler) DeleteSound(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	// Cannot delete system sounds
	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Sound{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Sound not found or cannot delete system sound"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
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
func (h *PhraseHandler) ListPhrases(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var phrases []models.Phrase
	h.DB.Where("tenant_id = ? OR tenant_id = 0 OR tenant_id IS NULL", tenantID).
		Order("macro_name").Find(&phrases)

	ctx.JSON(phrases)
}

// CreatePhrase creates a new phrase
func (h *PhraseHandler) CreatePhrase(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var phrase models.Phrase
	if err := ctx.ReadJSON(&phrase); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	phrase.TenantID = tenantID
	if err := h.DB.Create(&phrase).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(phrase)
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
func (h *ChatplanHandler) ListChatplans(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var chatplans []models.Chatplan
	h.DB.Where("tenant_id = ? OR tenant_id = 0", tenantID).
		Order("\"order\" ASC").Find(&chatplans)

	ctx.JSON(chatplans)
}

// CreateChatplan creates a new chatplan
func (h *ChatplanHandler) CreateChatplan(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var chatplan models.Chatplan
	if err := ctx.ReadJSON(&chatplan); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	chatplan.TenantID = tenantID
	if err := h.DB.Create(&chatplan).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(chatplan)
}

// UpdateChatplan updates a chatplan
func (h *ChatplanHandler) UpdateChatplan(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var existing models.Chatplan
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&existing).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Chatplan not found"})
		return
	}

	if err := ctx.ReadJSON(&existing); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	existing.TenantID = tenantID
	h.DB.Save(&existing)
	ctx.JSON(existing)
}

// DeleteChatplan deletes a chatplan
func (h *ChatplanHandler) DeleteChatplan(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Chatplan{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Chatplan not found"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
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
func (h *DefaultOutboundRouteHandler) ListRoutes(ctx iris.Context) {
	var routes []models.DefaultOutboundRoute
	h.DB.Order("\"order\" ASC").Find(&routes)
	ctx.JSON(routes)
}

// CreateRoute creates a new default route (system admin only)
func (h *DefaultOutboundRouteHandler) CreateRoute(ctx iris.Context) {
	var route models.DefaultOutboundRoute
	if err := ctx.ReadJSON(&route); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&route).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(route)
}

// UpdateRoute updates a default route
func (h *DefaultOutboundRouteHandler) UpdateRoute(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")

	var route models.DefaultOutboundRoute
	if err := h.DB.First(&route, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Route not found"})
		return
	}

	if err := ctx.ReadJSON(&route); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	h.DB.Save(&route)
	ctx.JSON(route)
}

// DeleteRoute deletes a default route
func (h *DefaultOutboundRouteHandler) DeleteRoute(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")

	result := h.DB.Delete(&models.DefaultOutboundRoute{}, id)
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Route not found"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
}
