package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// =====================
// User Portal Handlers
// =====================

func (h *Handler) GetUserDevices(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	// Get user's extension
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Find extensions assigned to this user
	var extensions []models.Extension
	if err := h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, middleware.GetTenantID(c)).Find(&extensions).Error; err != nil {
		return c.JSON(fiber.Map{"data": []interface{}{}})
	}

	// Map to device-like structure
	devices := make([]map[string]interface{}, 0)
	for _, ext := range extensions {
		if ext.DeviceUUID != "" {
			devices = append(devices, map[string]interface{}{
				"device_uuid": ext.DeviceUUID,
				"extension":   ext.Extension,
				"name":        ext.EffectiveCallerIDName,
			})
		}
	}

	return c.JSON(fiber.Map{"data": devices})
}

func (h *Handler) GetUserCallHistory(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	// Get user's extension number
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Fetch CDR for the user's extension
	var cdrs []models.CallRecord
	if err := h.DB.Where("tenant_id = ? AND (caller_id_number = ? OR destination_number = ?)",
		middleware.GetTenantID(c), user.Extension, user.Extension).
		Order("start_time DESC").
		Limit(100).
		Find(&cdrs).Error; err != nil {
		return c.JSON(fiber.Map{"data": []interface{}{}})
	}

	return c.JSON(fiber.Map{"data": cdrs})
}

func (h *Handler) GetUserVoicemail(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	// Get user's extension
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Find voicemail box for user's extension
	var box models.VoicemailBox
	if err := h.DB.Where("extension = ? AND tenant_id = ?", user.Extension, middleware.GetTenantID(c)).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&box).Error; err != nil {
		return c.JSON(fiber.Map{"data": nil, "messages": []interface{}{}})
	}

	return c.JSON(fiber.Map{
		"data":        box,
		"messages":    box.Messages,
		"new_count":   box.NewMessages,
		"saved_count": box.SavedMessages,
	})
}

func (h *Handler) GetUserSettings(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Get associated extension settings
	var ext models.Extension
	h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, middleware.GetTenantID(c)).First(&ext)

	return c.JSON(fiber.Map{
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.PhoneNumber,
			"extension":  user.Extension,
		},
		"call_settings": map[string]interface{}{
			"do_not_disturb":         ext.DoNotDisturb,
			"forward_all_enabled":    ext.ForwardAllEnabled,
			"forward_all_dest":       ext.ForwardAllDestination,
			"forward_busy_enabled":   ext.ForwardBusyEnabled,
			"forward_busy_dest":      ext.ForwardBusyDestination,
			"forward_no_answer":      ext.ForwardNoAnswerEnabled,
			"forward_no_answer_dest": ext.ForwardNoAnswerDestination,
			"voicemail_enabled":      ext.VoicemailEnabled,
			"follow_me_enabled":      ext.FollowMeEnabled,
			"record_inbound":         ext.RecordInbound,
			"record_outbound":        ext.RecordOutbound,
		},
	})
}

func (h *Handler) UpdateUserSettings(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var req struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		// Call settings
		DoNotDisturb           *bool   `json:"do_not_disturb"`
		ForwardAllEnabled      *bool   `json:"forward_all_enabled"`
		ForwardAllDestination  *string `json:"forward_all_destination"`
		ForwardBusyEnabled     *bool   `json:"forward_busy_enabled"`
		ForwardBusyDestination *string `json:"forward_busy_destination"`
		VoicemailEnabled       *bool   `json:"voicemail_enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Update user fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}

	if err := h.DB.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
	}

	// Update extension settings if user has one
	var ext models.Extension
	if err := h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, middleware.GetTenantID(c)).First(&ext).Error; err == nil {
		if req.DoNotDisturb != nil {
			ext.DoNotDisturb = *req.DoNotDisturb
		}
		if req.ForwardAllEnabled != nil {
			ext.ForwardAllEnabled = *req.ForwardAllEnabled
		}
		if req.ForwardAllDestination != nil {
			ext.ForwardAllDestination = *req.ForwardAllDestination
		}
		if req.ForwardBusyEnabled != nil {
			ext.ForwardBusyEnabled = *req.ForwardBusyEnabled
		}
		if req.ForwardBusyDestination != nil {
			ext.ForwardBusyDestination = *req.ForwardBusyDestination
		}
		if req.VoicemailEnabled != nil {
			ext.VoicemailEnabled = *req.VoicemailEnabled
		}
		h.DB.Save(&ext)
	}

	return c.JSON(fiber.Map{"message": "Settings updated"})
}

func (h *Handler) GetUserContacts(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var contacts []models.Contact
	if err := h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, middleware.GetTenantID(c)).Find(&contacts).Error; err != nil {
		return c.JSON(fiber.Map{"data": []interface{}{}})
	}

	return c.JSON(fiber.Map{"data": contacts})
}

func (h *Handler) CreateUserContact(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var contact models.Contact
	if err := c.BodyParser(&contact); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	contact.TenantID = middleware.GetTenantID(c)

	if err := h.DB.Create(&contact).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create contact"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": contact, "message": "Contact created"})
}
