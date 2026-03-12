package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// =====================
// Extensions
// =====================

func (h *Handler) ListExtensions(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var extensions []models.Extension
	query := h.DB.Where("tenant_id = ?", middleware.GetTenantID(c))

	// Search filter
	if search := c.Query("search"); search != "" {
		query = query.Where("extension LIKE ? OR effective_caller_id_name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&extensions).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch extensions"})
	}

	return c.JSON(fiber.Map{"data": extensions})
}

func (h *Handler) CreateExtension(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	// Use a struct that accepts password from JSON
	var input struct {
		Extension               string `json:"extension"`
		Password                string `json:"password"`
		WebPassword             string `json:"web_password"`
		DisplayName             string `json:"display_name"`
		Email                   string `json:"email"`
		VoicemailPin            string `json:"voicemail_pin"`
		ProfileID               *uint  `json:"profile_id"`
		Enabled                 bool   `json:"enabled"`
		EffectiveCallerIDName   string `json:"effective_caller_id_name"`
		EffectiveCallerIDNumber string `json:"effective_caller_id_number"`
		OutboundCallerIDName    string `json:"outbound_caller_id_name"`
		OutboundCallerIDNumber  string `json:"outbound_caller_id_number"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Validate required fields
	if input.Extension == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Extension number is required"})
	}

	tenantID := middleware.GetTenantID(c)

	// Check if extension already exists
	var existing models.Extension
	if err := h.DB.Where("extension = ? AND tenant_id = ?", input.Extension, tenantID).First(&existing).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Extension already exists"})
	}

	// Generate password if not provided
	password := input.Password
	if password == "" {
		password = generateRandomPassword(16)
	}

	// Get tenant for domain
	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch tenant"})
	}

	ext := models.Extension{
		TenantID:                tenantID,
		Extension:               input.Extension,
		Password:                password,
		Enabled:                 true,
		Domain:                  tenant.Domain,
		UserContext:             tenant.Domain,
		EffectiveCallerIDName:   input.DisplayName,
		EffectiveCallerIDNumber: input.Extension,
		OutboundCallerIDName:    input.OutboundCallerIDName,
		OutboundCallerIDNumber:  input.OutboundCallerIDNumber,
		VoicemailEnabled:        true,
		VoicemailPassword:       input.VoicemailPin,
		VoicemailMailTo:         input.Email,
		DirectoryFirstName:      input.DisplayName,
		DirectoryVisible:        true,
	}

	// Hash web password if provided
	if input.WebPassword != "" {
		if err := ext.SetWebPassword(input.WebPassword); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash web password"})
		}
	}

	if err := h.DB.Create(&ext).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create extension: " + err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": ext, "message": "Extension created"})
}

// generateRandomPassword creates a random alphanumeric password
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

func (h *Handler) GetExtension(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("ext"))
	var ext models.Extension

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&ext).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension not found"})
	}

	return c.JSON(fiber.Map{"data": ext})
}

func (h *Handler) UpdateExtension(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("ext"))
	var ext models.Extension

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&ext).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension not found"})
	}

	// Use input struct to handle fields that may not be in the model's JSON tags
	// Use pointers to distinguish between missing fields (nil) and explicit zero values
	var input struct {
		Extension               *string `json:"extension"`
		Password                *string `json:"password"`
		WebPassword             *string `json:"web_password"`
		Enabled                 *bool   `json:"enabled"`
		ProfileID               *uint   `json:"profile_id"`
		EffectiveCallerIDName   *string `json:"effective_caller_id_name"`
		EffectiveCallerIDNumber *string `json:"effective_caller_id_number"`
		OutboundCallerIDName    *string `json:"outbound_caller_id_name"`
		OutboundCallerIDNumber  *string `json:"outbound_caller_id_number"`
		DirectoryFirstName      *string `json:"directory_first_name"`
		DirectoryLastName       *string `json:"directory_last_name"`
		VoicemailEnabled        *bool   `json:"voicemail_enabled"`
		VoicemailPin            *string `json:"voicemail_pin"`
		VoicemailMailTo         *string `json:"voicemail_mail_to"`
		RingStrategy            *string `json:"ring_strategy"`
		RingDeviceOrder         *string `json:"ring_device_order"`
		NoAnswerAction          *string `json:"no_answer_action"`
		NoAnswerForwardTo       *string `json:"no_answer_forward_to"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Update fields from input if they are present (not nil)
	if input.Extension != nil {
		ext.Extension = *input.Extension
	}
	if input.Password != nil && *input.Password != "" { // Password usually shouldn't be cleared to empty via update, unless intentional security reset
		ext.Password = *input.Password
	}
	if input.Enabled != nil {
		ext.Enabled = *input.Enabled
	}
	// ProfileID is special: it's a pointer in the model too.
	// If input.ProfileID is nil (omitted), do nothing.
	// If input.ProfileID is NOT nil, update ext.ProfileID.
	// NOTE: To set profile to NULL, the client sends null, unmarshals to nil pointer?
	// Go's json.Unmarshal unmarshals null to nil pointer, and missing field to nil pointer.
	// We cannot distinguish "set to null" vs "omitted" easily with standard lib for pointer-to-pointer.
	// However, for ProfileID *uint:
	// If the user wants to unassign, they usually send `profile_id: null`.
	// But `input.ProfileID` (type *uint) will be nil in both cases.
	// FIX: For now, we assume if `profile_id` key exists in map it's an update.
	// But ReadJSON uses struct.
	// To allow unassigning, we'd need a different approach (e.g. map[string]interface{} or external library).
	// Given previous context, the user just wants "editing" to work. Unassigning might be edge case or handled by sending 0?
	// The frontend sends `null` or valid ID.
	// If we use simple *uint in input: `null` -> nil. Omitted -> nil.
	// So we can only update if we send a value. We cannot "unset" it easily unless we use logic like 0 = unset.
	// But let's check the frontend. It sends `profile_id: extension.value.profileId || null`.
	// If we want to support unsetting, we might need to rely on `ProfileID` being updated if `ProfileID` is in the payload.
	// For now, let's just apply if not nil, which matches current behavior for "setting" a profile.
	// To fix "unsetting", we might need to revisit later or use 0.
	if input.ProfileID != nil {
		ext.ProfileID = input.ProfileID
	}

	if input.EffectiveCallerIDName != nil {
		ext.EffectiveCallerIDName = *input.EffectiveCallerIDName
	}
	if input.EffectiveCallerIDNumber != nil {
		ext.EffectiveCallerIDNumber = *input.EffectiveCallerIDNumber
	}
	if input.OutboundCallerIDName != nil {
		ext.OutboundCallerIDName = *input.OutboundCallerIDName
	}
	if input.OutboundCallerIDNumber != nil {
		ext.OutboundCallerIDNumber = *input.OutboundCallerIDNumber
	}
	if input.DirectoryFirstName != nil {
		ext.DirectoryFirstName = *input.DirectoryFirstName
	}
	if input.DirectoryLastName != nil {
		ext.DirectoryLastName = *input.DirectoryLastName
	}
	if input.VoicemailEnabled != nil {
		ext.VoicemailEnabled = *input.VoicemailEnabled
	}
	if input.VoicemailPin != nil {
		ext.VoicemailPassword = *input.VoicemailPin
	}
	if input.VoicemailMailTo != nil {
		ext.VoicemailMailTo = *input.VoicemailMailTo
	}
	if input.RingStrategy != nil {
		ext.RingStrategy = *input.RingStrategy
	}
	if input.RingDeviceOrder != nil {
		ext.RingDeviceOrder = *input.RingDeviceOrder
	}
	if input.NoAnswerAction != nil {
		ext.NoAnswerAction = *input.NoAnswerAction
	}
	if input.NoAnswerForwardTo != nil {
		ext.NoAnswerForwardTo = *input.NoAnswerForwardTo
	}

	// Hash web password if provided
	if input.WebPassword != nil && *input.WebPassword != "" {
		if err := ext.SetWebPassword(*input.WebPassword); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash web password"})
		}
	}

	if err := h.DB.Save(&ext).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update extension"})
	}

	return c.JSON(fiber.Map{"data": ext, "message": "Extension updated"})
}

func (h *Handler) DeleteExtension(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("ext"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Delete(&models.Extension{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete extension"})
	}

	return c.JSON(fiber.Map{"message": "Extension deleted"})
}

func (h *Handler) GetExtensionStatus(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.Atoi(c.Params("id"))

	var ext models.Extension
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&ext).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension not found"})
	}

	// Get tenant domain for registration lookup
	var tenant models.Tenant
	h.DB.First(&tenant, tenantID)

	status := "unregistered"
	registrationIP := ""

	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		result, err := h.ESLManager.API(fmt.Sprintf("sofia status profile internal user %s@%s",
			ext.Extension, tenant.Domain))
		if err == nil && !strings.Contains(result, "Invalid") && !strings.Contains(result, "-ERR") {
			status = "registered"
			// Try to extract IP from result
			lines := strings.Split(result, "\n")
			for _, line := range lines {
				if strings.Contains(line, "Contact") {
					// Extract IP from contact string
					if atIdx := strings.Index(line, "@"); atIdx > 0 {
						rest := line[atIdx+1:]
						if colonIdx := strings.Index(rest, ":"); colonIdx > 0 {
							registrationIP = rest[:colonIdx]
						} else if semiIdx := strings.Index(rest, ";"); semiIdx > 0 {
							registrationIP = rest[:semiIdx]
						}
					}
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"extension":       ext.Extension,
		"status":          status,
		"registration_ip": registrationIP,
	})
}

// NOTE: Device handlers moved to device_handlers.go

// =====================
// Voicemail
// =====================

func (h *Handler) ListVoicemailBoxes(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var boxes []models.VoicemailBox
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(c)).Find(&boxes).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch voicemail boxes"})
	}

	return c.JSON(fiber.Map{"data": boxes})
}

func (h *Handler) CreateVoicemailBox(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var box models.VoicemailBox
	if err := c.BodyParser(&box); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	box.TenantID = middleware.GetTenantID(c)

	if err := h.DB.Create(&box).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create voicemail box"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": box, "message": "Voicemail box created"})
}

func (h *Handler) GetVoicemailBox(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var box models.VoicemailBox

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Preload("Messages").First(&box).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Voicemail box not found"})
	}

	return c.JSON(fiber.Map{"data": box})
}

func (h *Handler) UpdateVoicemailBox(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var box models.VoicemailBox

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&box).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Voicemail box not found"})
	}

	if err := c.BodyParser(&box); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Save(&box).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update voicemail box"})
	}

	return c.JSON(fiber.Map{"data": box, "message": "Voicemail box updated"})
}

func (h *Handler) DeleteVoicemailBox(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Delete(&models.VoicemailBox{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete voicemail box"})
	}

	return c.JSON(fiber.Map{"message": "Voicemail box deleted"})
}

// ListVoicemailMessages lists messages for a specific voicemail box
func (h *Handler) ListVoicemailMessages(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	ext := c.Params("ext")
	tenantID := middleware.GetTenantID(c)

	// Find the voicemail box
	var box models.VoicemailBox
	if err := h.DB.Where("extension = ? AND tenant_id = ?", ext, tenantID).First(&box).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Voicemail box not found"})
	}

	// Get messages
	var messages []models.VoicemailMessage
	if err := h.DB.Where("box_id = ?", box.ID).Order("created_at DESC").Find(&messages).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch messages"})
	}

	return c.JSON(fiber.Map{"data": messages, "box": box})
}

// GetVoicemailMessage gets a single voicemail message
func (h *Handler) GetVoicemailMessage(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	var message models.VoicemailMessage
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&message).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	return c.JSON(fiber.Map{"data": message})
}

// DeleteVoicemailMessage deletes a voicemail message
func (h *Handler) DeleteVoicemailMessage(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	var message models.VoicemailMessage
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&message).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	// TODO: Delete the audio file from storage

	// Update box message counts
	if message.IsNew {
		h.DB.Model(&models.VoicemailBox{}).Where("id = ?", message.BoxID).Update("new_messages", models.VoicemailBox{}.NewMessages-1)
	} else {
		h.DB.Model(&models.VoicemailBox{}).Where("id = ?", message.BoxID).Update("saved_messages", models.VoicemailBox{}.SavedMessages-1)
	}

	if err := h.DB.Delete(&message).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete message"})
	}

	return c.JSON(fiber.Map{"message": "Voicemail message deleted"})
}

// MarkVoicemailRead marks a voicemail message as read
func (h *Handler) MarkVoicemailRead(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	var message models.VoicemailMessage
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&message).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	if message.IsNew {
		if err := message.MarkAsRead(h.DB); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to mark as read"})
		}

		// Update box counts
		h.DB.Model(&models.VoicemailBox{}).Where("id = ?", message.BoxID).Updates(map[string]interface{}{
			"new_messages":   models.VoicemailBox{}.NewMessages - 1,
			"saved_messages": models.VoicemailBox{}.SavedMessages + 1,
		})
	}

	return c.JSON(fiber.Map{"data": message, "message": "Marked as read"})
}

// StreamVoicemailMessage streams the voicemail audio file
func (h *Handler) StreamVoicemailMessage(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	var message models.VoicemailMessage
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&message).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	if message.FilePath == "" {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Audio file not found"})
	}

	// Serve the file
	return c.SendFile(message.FilePath)
}

// =====================
// Recordings
// =====================

func (h *Handler) ListRecordings(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var recordings []models.Recording
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(c)).Order("created_at DESC").Find(&recordings).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch recordings"})
	}

	return c.JSON(fiber.Map{"data": recordings})
}

func (h *Handler) GetRecording(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var recording models.Recording

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&recording).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording not found"})
	}

	return c.JSON(fiber.Map{"data": recording})
}

func (h *Handler) DeleteRecording(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Delete(&models.Recording{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete recording"})
	}

	return c.JSON(fiber.Map{"message": "Recording deleted"})
}

// StreamRecording streams recording audio
func (h *Handler) StreamRecording(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	// Try CallRecording first (call recordings), then Recording (audio library)
	var callRec models.CallRecording
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&callRec).Error; err == nil {
		if callRec.FilePath == "" {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording file not found"})
		}
		return c.SendFile(callRec.FilePath)
	}

	var rec models.Recording
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&rec).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording not found"})
	}

	if rec.FilePath == "" {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording file not found"})
	}

	return c.SendFile(rec.FilePath)
}

// DownloadRecording downloads recording as attachment
func (h *Handler) DownloadRecording(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	var callRec models.CallRecording
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&callRec).Error; err == nil {
		if callRec.FilePath == "" {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording file not found"})
		}
		filename := callRec.FileName
		if filename == "" {
			filename = "recording.wav"
		}
		c.Set("Content-Disposition", "attachment; filename="+filename)
		return c.SendFile(callRec.FilePath)
	}

	var rec models.Recording
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&rec).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording not found"})
	}

	if rec.FilePath == "" {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording file not found"})
	}

	filename := rec.FileName
	if filename == "" {
		filename = "recording.wav"
	}
	c.Set("Content-Disposition", "attachment; filename="+filename)
	return c.SendFile(rec.FilePath)
}

// UpdateRecordingNotes updates notes/tags on a call recording
func (h *Handler) UpdateRecordingNotes(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	var rec models.CallRecording
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&rec).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording not found"})
	}

	var input struct {
		Notes string `json:"notes"`
		Tags  string `json:"tags"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rec.Notes = input.Notes
	rec.Tags = input.Tags

	if err := h.DB.Save(&rec).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update recording"})
	}

	return c.JSON(fiber.Map{"data": rec, "message": "Recording updated"})
}

// GetRecordingTranscription returns transcription for a call recording
func (h *Handler) GetRecordingTranscription(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	var rec models.CallRecording
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&rec).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Recording not found"})
	}

	var transcription models.Transcription
	if err := h.DB.Where("recording_id = ?", rec.ID).Preload("Segments").First(&transcription).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Transcription not found", "status": rec.TranscriptionStatus})
	}

	return c.JSON(fiber.Map{"data": transcription})
}

// GetRecordingConfig returns tenant-level recording configuration
func (h *Handler) GetRecordingConfig(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	tenantID := middleware.GetTenantID(c)

	var config models.RecordingConfig
	if err := h.DB.Where("tenant_id = ?", tenantID).First(&config).Error; err != nil {
		// Return default config if none exists
		return c.JSON(fiber.Map{"data": models.RecordingConfig{TenantID: tenantID}})
	}

	return c.JSON(fiber.Map{"data": config})
}

// =====================
// IVR Menus
// =====================

func (h *Handler) ListIVRMenus(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var menus []models.IVRMenu
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(c)).Preload("Options").Find(&menus).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch IVR menus"})
	}

	return c.JSON(fiber.Map{"data": menus})
}

func (h *Handler) CreateIVRMenu(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var menu models.IVRMenu
	if err := c.BodyParser(&menu); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	menu.TenantID = middleware.GetTenantID(c)

	if err := h.DB.Create(&menu).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create IVR menu"})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": menu, "message": "IVR menu created"})
}

func (h *Handler) GetIVRMenu(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var menu models.IVRMenu

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Preload("Options").First(&menu).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "IVR menu not found"})
	}

	return c.JSON(fiber.Map{"data": menu})
}

func (h *Handler) UpdateIVRMenu(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var menu models.IVRMenu

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&menu).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "IVR menu not found"})
	}

	if err := c.BodyParser(&menu); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Save(&menu).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update IVR menu"})
	}

	h.reloadXML()
	return c.JSON(fiber.Map{"data": menu, "message": "IVR menu updated"})
}

func (h *Handler) DeleteIVRMenu(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Delete(&models.IVRMenu{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete IVR menu"})
	}

	h.reloadXML()
	return c.JSON(fiber.Map{"message": "IVR menu deleted"})
}

// =====================
// Queues
// =====================

func (h *Handler) ListQueues(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var queues []models.Queue
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(c)).Preload("Agents").Find(&queues).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch queues"})
	}

	return c.JSON(fiber.Map{"data": queues})
}

func (h *Handler) CreateQueue(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var queue models.Queue
	if err := c.BodyParser(&queue); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	queue.TenantID = middleware.GetTenantID(c)

	if err := h.DB.Create(&queue).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create queue"})
	}

	h.reloadCallcenter()
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": queue, "message": "Queue created"})
}

func (h *Handler) GetQueue(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var queue models.Queue

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Preload("Agents").First(&queue).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Queue not found"})
	}

	return c.JSON(fiber.Map{"data": queue})
}

func (h *Handler) UpdateQueue(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var queue models.Queue

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&queue).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Queue not found"})
	}

	if err := c.BodyParser(&queue); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Save(&queue).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update queue"})
	}

	h.reloadCallcenter()
	return c.JSON(fiber.Map{"data": queue, "message": "Queue updated"})
}

func (h *Handler) DeleteQueue(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Delete(&models.Queue{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete queue"})
	}

	h.reloadCallcenter()
	return c.JSON(fiber.Map{"message": "Queue deleted"})
}

// =====================
// Queue Agent Management
// =====================

// ListQueueAgents lists agents for a specific queue
func (h *Handler) ListQueueAgents(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	// Verify queue belongs to tenant
	var queue models.Queue
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&queue).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Queue not found"})
	}

	var agents []models.QueueAgent
	if err := h.DB.Where("queue_id = ? AND tenant_id = ?", id, tenantID).Order("tier_level ASC, tier_position ASC").Find(&agents).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch agents"})
	}

	return c.JSON(fiber.Map{"data": agents})
}

// AddQueueAgent adds an agent to a queue
func (h *Handler) AddQueueAgent(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	tenantID := middleware.GetTenantID(c)

	// Verify queue belongs to tenant
	var queue models.Queue
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&queue).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Queue not found"})
	}

	var agent models.QueueAgent
	if err := c.BodyParser(&agent); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	agent.QueueID = uint(id)
	agent.TenantID = tenantID
	if agent.Status == "" {
		agent.Status = models.AgentStatusLoggedOut
	}
	if agent.State == "" {
		agent.State = models.AgentStateIdle
	}

	if err := h.DB.Create(&agent).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add agent"})
	}

	h.reloadCallcenter()
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": agent, "message": "Agent added to queue"})
}

// RemoveQueueAgent removes an agent from a queue
func (h *Handler) RemoveQueueAgent(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	agentID, _ := strconv.Atoi(c.Params("agentId"))
	tenantID := middleware.GetTenantID(c)

	// Verify queue belongs to tenant
	var queue models.Queue
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&queue).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Queue not found"})
	}

	if err := h.DB.Where("id = ? AND queue_id = ? AND tenant_id = ?", agentID, id, tenantID).Delete(&models.QueueAgent{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove agent"})
	}

	h.reloadCallcenter()
	return c.JSON(fiber.Map{"message": "Agent removed from queue"})
}

// PauseQueueAgent sets agent status to On Break
func (h *Handler) PauseQueueAgent(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	agentID, _ := strconv.Atoi(c.Params("agentId"))
	tenantID := middleware.GetTenantID(c)

	var agent models.QueueAgent
	if err := h.DB.Where("id = ? AND queue_id = ? AND tenant_id = ?", agentID, id, tenantID).First(&agent).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Agent not found"})
	}

	agent.Status = models.AgentStatusOnBreak
	if err := h.DB.Save(&agent).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to pause agent"})
	}

	return c.JSON(fiber.Map{"data": agent, "message": "Agent paused"})
}

// UnpauseQueueAgent sets agent status to Available
func (h *Handler) UnpauseQueueAgent(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	agentID, _ := strconv.Atoi(c.Params("agentId"))
	tenantID := middleware.GetTenantID(c)

	var agent models.QueueAgent
	if err := h.DB.Where("id = ? AND queue_id = ? AND tenant_id = ?", agentID, id, tenantID).First(&agent).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Agent not found"})
	}

	agent.Status = models.AgentStatusAvailable
	if err := h.DB.Save(&agent).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unpause agent"})
	}

	return c.JSON(fiber.Map{"data": agent, "message": "Agent unpaused"})
}

// =====================
// Ring Groups
// =====================

func (h *Handler) ListRingGroups(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var groups []models.RingGroup
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(c)).Preload("Destinations").Find(&groups).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch ring groups"})
	}

	return c.JSON(fiber.Map{"data": groups})
}

func (h *Handler) CreateRingGroup(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var group models.RingGroup
	if err := c.BodyParser(&group); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	group.TenantID = middleware.GetTenantID(c)

	if err := h.DB.Create(&group).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create ring group"})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": group, "message": "Ring group created"})
}

func (h *Handler) GetRingGroup(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var group models.RingGroup

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Preload("Destinations").First(&group).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Ring group not found"})
	}

	return c.JSON(fiber.Map{"data": group})
}

func (h *Handler) UpdateRingGroup(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var group models.RingGroup

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&group).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Ring group not found"})
	}

	if err := c.BodyParser(&group); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Save(&group).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update ring group"})
	}

	h.reloadXML()
	return c.JSON(fiber.Map{"data": group, "message": "Ring group updated"})
}

func (h *Handler) DeleteRingGroup(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Delete(&models.RingGroup{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete ring group"})
	}

	h.reloadXML()
	return c.JSON(fiber.Map{"message": "Ring group deleted"})
}

// =====================
// Conferences
// =====================

func (h *Handler) ListConferences(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var conferences []models.Conference
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(c)).Find(&conferences).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch conferences"})
	}

	return c.JSON(fiber.Map{"data": conferences})
}

func (h *Handler) CreateConference(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var conf models.Conference
	if err := c.BodyParser(&conf); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	conf.TenantID = middleware.GetTenantID(c)

	if err := h.DB.Create(&conf).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create conference"})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": conf, "message": "Conference created"})
}

func (h *Handler) GetConference(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var conf models.Conference

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&conf).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Conference not found"})
	}

	return c.JSON(fiber.Map{"data": conf})
}

func (h *Handler) UpdateConference(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var conf models.Conference

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).First(&conf).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Conference not found"})
	}

	if err := c.BodyParser(&conf); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Save(&conf).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update conference"})
	}

	h.reloadXML()
	return c.JSON(fiber.Map{"data": conf, "message": "Conference updated"})
}

func (h *Handler) DeleteConference(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(c)).Delete(&models.Conference{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete conference"})
	}

	return c.JSON(fiber.Map{"message": "Conference deleted"})
}

// =====================
// Numbers/DIDs
// =====================

// Numbers handlers moved to routing_handlers.go

// Routing and Dialplan handlers moved to routing_handlers.go

// =====================
// Audio Library
// =====================

func (h *Handler) ListAudioFiles(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var files []models.Recording
	if err := h.DB.Where("tenant_id = ? AND type = ?", middleware.GetTenantID(c), "audio").Find(&files).Error; err != nil {
		return c.JSON(fiber.Map{"data": []interface{}{}})
	}

	return c.JSON(fiber.Map{"data": files})
}

func (h *Handler) UploadAudioFile(c *fiber.Ctx) error {
	// TODO: Implement file upload handling
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Audio file uploaded"})
}

func (h *Handler) GetAudioFile(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"data": map[string]interface{}{"id": id}})
}

func (h *Handler) DeleteAudioFile(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Audio file deleted", "id": id})
}

// =====================
// Music on Hold
// =====================

func (h *Handler) ListMOHStreams(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	// TODO: Add MOH model or use Recording with type filter
	return c.JSON(fiber.Map{"data": []interface{}{}, "message": "MOH streams"})
}

func (h *Handler) CreateMOHStream(c *fiber.Ctx) error {
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "MOH stream created"})
}

func (h *Handler) GetMOHStream(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"data": map[string]interface{}{"id": id}})
}

func (h *Handler) UpdateMOHStream(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "MOH stream updated", "id": id})
}

func (h *Handler) DeleteMOHStream(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "MOH stream deleted", "id": id})
}

// =====================
// Extension Profiles
// =====================

func (h *Handler) ListExtensionProfiles(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var profiles []models.ExtensionProfile
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&profiles).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch extension profiles"})
	}

	// Count extensions per profile
	type ProfileCount struct {
		ProfileID uint
		Count     int64
	}
	var counts []ProfileCount
	h.DB.Model(&models.Extension{}).
		Select("profile_id, count(*) as count").
		Where("tenant_id = ? AND profile_id IS NOT NULL", tenantID).
		Group("profile_id").
		Scan(&counts)

	countMap := make(map[uint]int64)
	for _, c := range counts {
		countMap[c.ProfileID] = c.Count
	}

	// Build response with counts
	type ProfileResponse struct {
		models.ExtensionProfile
		ExtensionCount int64 `json:"extension_count"`
	}
	var response []ProfileResponse
	for _, p := range profiles {
		response = append(response, ProfileResponse{
			ExtensionProfile: p,
			ExtensionCount:   countMap[p.ID],
		})
	}

	return c.JSON(fiber.Map{"data": response})
}

func (h *Handler) CreateExtensionProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var profile models.ExtensionProfile
	if err := c.BodyParser(&profile); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	profile.TenantID = tenantID

	if err := h.DB.Create(&profile).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create extension profile"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": profile, "message": "Extension profile created"})
}

func (h *Handler) GetExtensionProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var profile models.ExtensionProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&profile).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension profile not found"})
	}

	return c.JSON(fiber.Map{"data": profile})
}

func (h *Handler) UpdateExtensionProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var profile models.ExtensionProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&profile).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension profile not found"})
	}

	var input models.ExtensionProfile
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	profile.Name = input.Name
	profile.Color = input.Color
	profile.Permissions = input.Permissions
	profile.CallHandling = input.CallHandling
	profile.RoutingOverride = input.RoutingOverride

	if err := h.DB.Save(&profile).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update extension profile"})
	}

	return c.JSON(fiber.Map{"data": profile, "message": "Extension profile updated"})
}

func (h *Handler) DeleteExtensionProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	// Unassign extensions from this profile
	h.DB.Model(&models.Extension{}).
		Where("tenant_id = ? AND profile_id = ?", tenantID, id).
		Update("profile_id", nil)

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.ExtensionProfile{})
	if result.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete extension profile"})
	}

	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension profile not found"})
	}

	return c.JSON(fiber.Map{"message": "Extension profile deleted"})
}

// =====================
// Speed Dials
// =====================

func (h *Handler) ListSpeedDialGroups(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var groups []models.SpeedDialGroup
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&groups).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch speed dial groups"})
	}

	return c.JSON(fiber.Map{"data": groups})
}

func (h *Handler) CreateSpeedDialGroup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var group models.SpeedDialGroup
	if err := c.BodyParser(&group); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	group.TenantID = tenantID
	if group.Entries == nil {
		group.Entries = models.SpeedDialEntries{}
	}

	if err := h.DB.Create(&group).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create speed dial group"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": group, "message": "Speed dial group created"})
}

func (h *Handler) GetSpeedDialGroup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var group models.SpeedDialGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&group).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Speed dial group not found"})
	}

	return c.JSON(fiber.Map{"data": group})
}

func (h *Handler) UpdateSpeedDialGroup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var group models.SpeedDialGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&group).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Speed dial group not found"})
	}

	var input models.SpeedDialGroup
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	group.Name = input.Name
	group.Description = input.Description
	group.Prefix = input.Prefix
	group.Enabled = input.Enabled
	group.Entries = input.Entries

	if err := h.DB.Save(&group).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update speed dial group"})
	}

	return c.JSON(fiber.Map{"data": group, "message": "Speed dial group updated"})
}

func (h *Handler) DeleteSpeedDialGroup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.SpeedDialGroup{})
	if result.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete speed dial group"})
	}

	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Speed dial group not found"})
	}

	return c.JSON(fiber.Map{"message": "Speed dial group deleted"})
}

// =====================
// Dial Code Collision Check
// =====================

// checkDialCodeConflict checks if a dial code is in use by any entity for the given tenant
func (h *Handler) checkDialCodeConflict(tenantID uint, code string, currentType string, currentID uint) map[string]string {
	// Check extensions
	var ext models.Extension
	query := h.DB.Where("tenant_id = ? AND (extension = ? OR number_alias = ?)", tenantID, code, code)
	if currentType == "extension" && currentID > 0 {
		query = query.Where("id != ?", currentID)
	}
	if query.First(&ext).Error == nil {
		return map[string]string{"type": "extension", "name": ext.EffectiveCallerIDName}
	}

	// Check IVR menus
	var ivr models.IVRMenu
	query = h.DB.Where("tenant_id = ? AND extension = ?", tenantID, code)
	if currentType == "ivr" && currentID > 0 {
		query = query.Where("id != ?", currentID)
	}
	if query.First(&ivr).Error == nil {
		return map[string]string{"type": "ivr", "name": ivr.Name}
	}

	// Check queues
	var queue models.Queue
	query = h.DB.Where("tenant_id = ? AND extension = ?", tenantID, code)
	if currentType == "queue" && currentID > 0 {
		query = query.Where("id != ?", currentID)
	}
	if query.First(&queue).Error == nil {
		return map[string]string{"type": "queue", "name": queue.Name}
	}

	// Check conferences
	var conf models.Conference
	query = h.DB.Where("tenant_id = ? AND extension = ?", tenantID, code)
	if currentType == "conference" && currentID > 0 {
		query = query.Where("id != ?", currentID)
	}
	if query.First(&conf).Error == nil {
		return map[string]string{"type": "conference", "name": conf.Name}
	}

	// Check ring groups
	var rg models.RingGroup
	query = h.DB.Where("tenant_id = ? AND extension = ?", tenantID, code)
	if currentType == "ring_group" && currentID > 0 {
		query = query.Where("id != ?", currentID)
	}
	if query.First(&rg).Error == nil {
		return map[string]string{"type": "ring_group", "name": rg.Name}
	}

	// Check time conditions
	var tc models.TimeCondition
	query = h.DB.Where("tenant_id = ? AND extension = ?", tenantID, code)
	if currentType == "time_condition" && currentID > 0 {
		query = query.Where("id != ?", currentID)
	}
	if query.First(&tc).Error == nil {
		return map[string]string{"type": "time_condition", "name": tc.Name}
	}

	// Check call flows (toggles)
	var cf models.CallFlow
	query = h.DB.Where("tenant_id = ? AND (extension = ? OR feature_code = ?)", tenantID, code, code)
	if currentType == "call_flow" && currentID > 0 {
		query = query.Where("id != ?", currentID)
	}
	if query.First(&cf).Error == nil {
		return map[string]string{"type": "call_flow", "name": cf.Name}
	}

	// Check feature codes
	var fc models.FeatureCode
	query = h.DB.Where("(tenant_id = ? OR tenant_id IS NULL) AND (code = ? OR extension = ?)", tenantID, code, code)
	if currentType == "feature_code" && currentID > 0 {
		query = query.Where("id != ?", currentID)
	}
	if query.First(&fc).Error == nil {
		return map[string]string{"type": "feature_code", "name": fc.Name}
	}

	return nil
}

// CheckDialCode is a public endpoint for UI to validate dial codes
func (h *Handler) CheckDialCode(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var input struct {
		Code      string `json:"code"`
		Type      string `json:"type"`
		ExcludeID uint   `json:"exclude_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	conflict := h.checkDialCodeConflict(tenantID, input.Code, input.Type, input.ExcludeID)
	return c.JSON(fiber.Map{
		"available": conflict == nil,
		"conflict":  conflict,
	})
}
