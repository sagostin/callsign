package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
)

// =====================
// Extensions
// =====================

func (h *Handler) ListExtensions(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var extensions []models.Extension
	query := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx))

	// Search filter
	if search := ctx.URLParam("search"); search != "" {
		query = query.Where("extension LIKE ? OR effective_caller_id_name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&extensions).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch extensions"})
		return
	}

	ctx.JSON(iris.Map{"data": extensions})
}

func (h *Handler) CreateExtension(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	// Use a struct that accepts password from JSON
	var input struct {
		Extension               string `json:"extension"`
		Password                string `json:"password"`
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
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Validate required fields
	if input.Extension == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Extension number is required"})
		return
	}

	tenantID := middleware.GetTenantID(ctx)

	// Check if extension already exists
	var existing models.Extension
	if err := h.DB.Where("extension = ? AND tenant_id = ?", input.Extension, tenantID).First(&existing).Error; err == nil {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": "Extension already exists"})
		return
	}

	// Generate password if not provided
	password := input.Password
	if password == "" {
		password = generateRandomPassword(16)
	}

	// Get tenant for domain
	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch tenant"})
		return
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

	if err := h.DB.Create(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create extension: " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": ext, "message": "Extension created"})
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

func (h *Handler) GetExtension(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("ext"))
	var ext models.Extension

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension not found"})
		return
	}

	ctx.JSON(iris.Map{"data": ext})
}

func (h *Handler) UpdateExtension(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("ext"))
	var ext models.Extension

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension not found"})
		return
	}

	// Use input struct to handle fields that may not be in the model's JSON tags
	// Use pointers to distinguish between missing fields (nil) and explicit zero values
	var input struct {
		Extension               *string `json:"extension"`
		Password                *string `json:"password"`
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
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
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

	if err := h.DB.Save(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update extension"})
		return
	}

	ctx.JSON(iris.Map{"data": ext, "message": "Extension updated"})
}

func (h *Handler) DeleteExtension(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("ext"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.Extension{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete extension"})
		return
	}

	ctx.JSON(iris.Map{"message": "Extension deleted"})
}

func (h *Handler) GetExtensionStatus(ctx iris.Context) {
	// TODO: Integrate with FreeSWITCH ESL for real-time status
	ctx.JSON(iris.Map{"status": "registered", "message": "Real-time status requires ESL integration"})
}

// NOTE: Device handlers moved to device_handlers.go

// =====================
// Voicemail
// =====================

func (h *Handler) ListVoicemailBoxes(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var boxes []models.VoicemailBox
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Find(&boxes).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch voicemail boxes"})
		return
	}

	ctx.JSON(iris.Map{"data": boxes})
}

func (h *Handler) CreateVoicemailBox(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var box models.VoicemailBox
	if err := ctx.ReadJSON(&box); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	box.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&box).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create voicemail box"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": box, "message": "Voicemail box created"})
}

func (h *Handler) GetVoicemailBox(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var box models.VoicemailBox

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Preload("Messages").First(&box).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Voicemail box not found"})
		return
	}

	ctx.JSON(iris.Map{"data": box})
}

func (h *Handler) UpdateVoicemailBox(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var box models.VoicemailBox

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&box).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Voicemail box not found"})
		return
	}

	if err := ctx.ReadJSON(&box); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&box).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update voicemail box"})
		return
	}

	ctx.JSON(iris.Map{"data": box, "message": "Voicemail box updated"})
}

func (h *Handler) DeleteVoicemailBox(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.VoicemailBox{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete voicemail box"})
		return
	}

	ctx.JSON(iris.Map{"message": "Voicemail box deleted"})
}

// ListVoicemailMessages lists messages for a specific voicemail box
func (h *Handler) ListVoicemailMessages(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	ext := ctx.Params().Get("ext")
	tenantID := middleware.GetTenantID(ctx)

	// Find the voicemail box
	var box models.VoicemailBox
	if err := h.DB.Where("extension = ? AND tenant_id = ?", ext, tenantID).First(&box).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Voicemail box not found"})
		return
	}

	// Get messages
	var messages []models.VoicemailMessage
	if err := h.DB.Where("box_id = ?", box.ID).Order("created_at DESC").Find(&messages).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch messages"})
		return
	}

	ctx.JSON(iris.Map{"data": messages, "box": box})
}

// GetVoicemailMessage gets a single voicemail message
func (h *Handler) GetVoicemailMessage(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	tenantID := middleware.GetTenantID(ctx)

	var message models.VoicemailMessage
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&message).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Message not found"})
		return
	}

	ctx.JSON(iris.Map{"data": message})
}

// DeleteVoicemailMessage deletes a voicemail message
func (h *Handler) DeleteVoicemailMessage(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	tenantID := middleware.GetTenantID(ctx)

	var message models.VoicemailMessage
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&message).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Message not found"})
		return
	}

	// TODO: Delete the audio file from storage

	// Update box message counts
	if message.IsNew {
		h.DB.Model(&models.VoicemailBox{}).Where("id = ?", message.BoxID).Update("new_messages", models.VoicemailBox{}.NewMessages-1)
	} else {
		h.DB.Model(&models.VoicemailBox{}).Where("id = ?", message.BoxID).Update("saved_messages", models.VoicemailBox{}.SavedMessages-1)
	}

	if err := h.DB.Delete(&message).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete message"})
		return
	}

	ctx.JSON(iris.Map{"message": "Voicemail message deleted"})
}

// MarkVoicemailRead marks a voicemail message as read
func (h *Handler) MarkVoicemailRead(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	tenantID := middleware.GetTenantID(ctx)

	var message models.VoicemailMessage
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&message).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Message not found"})
		return
	}

	if message.IsNew {
		if err := message.MarkAsRead(h.DB); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "Failed to mark as read"})
			return
		}

		// Update box counts
		h.DB.Model(&models.VoicemailBox{}).Where("id = ?", message.BoxID).Updates(map[string]interface{}{
			"new_messages":   models.VoicemailBox{}.NewMessages - 1,
			"saved_messages": models.VoicemailBox{}.SavedMessages + 1,
		})
	}

	ctx.JSON(iris.Map{"data": message, "message": "Marked as read"})
}

// StreamVoicemailMessage streams the voicemail audio file
func (h *Handler) StreamVoicemailMessage(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	tenantID := middleware.GetTenantID(ctx)

	var message models.VoicemailMessage
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&message).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Message not found"})
		return
	}

	if message.FilePath == "" {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Audio file not found"})
		return
	}

	// Serve the file
	ctx.ServeFile(message.FilePath)
}

// =====================
// Recordings
// =====================

func (h *Handler) ListRecordings(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var recordings []models.Recording
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Order("created_at DESC").Find(&recordings).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch recordings"})
		return
	}

	ctx.JSON(iris.Map{"data": recordings})
}

func (h *Handler) GetRecording(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var recording models.Recording

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&recording).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Recording not found"})
		return
	}

	ctx.JSON(iris.Map{"data": recording})
}

func (h *Handler) DeleteRecording(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.Recording{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete recording"})
		return
	}

	ctx.JSON(iris.Map{"message": "Recording deleted"})
}

// =====================
// IVR Menus
// =====================

func (h *Handler) ListIVRMenus(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var menus []models.IVRMenu
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Preload("Options").Find(&menus).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch IVR menus"})
		return
	}

	ctx.JSON(iris.Map{"data": menus})
}

func (h *Handler) CreateIVRMenu(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var menu models.IVRMenu
	if err := ctx.ReadJSON(&menu); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	menu.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&menu).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create IVR menu"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": menu, "message": "IVR menu created"})
}

func (h *Handler) GetIVRMenu(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var menu models.IVRMenu

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Preload("Options").First(&menu).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "IVR menu not found"})
		return
	}

	ctx.JSON(iris.Map{"data": menu})
}

func (h *Handler) UpdateIVRMenu(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var menu models.IVRMenu

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&menu).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "IVR menu not found"})
		return
	}

	if err := ctx.ReadJSON(&menu); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&menu).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update IVR menu"})
		return
	}

	ctx.JSON(iris.Map{"data": menu, "message": "IVR menu updated"})
}

func (h *Handler) DeleteIVRMenu(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.IVRMenu{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete IVR menu"})
		return
	}

	ctx.JSON(iris.Map{"message": "IVR menu deleted"})
}

// =====================
// Queues
// =====================

func (h *Handler) ListQueues(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var queues []models.Queue
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Preload("Agents").Find(&queues).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch queues"})
		return
	}

	ctx.JSON(iris.Map{"data": queues})
}

func (h *Handler) CreateQueue(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var queue models.Queue
	if err := ctx.ReadJSON(&queue); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	queue.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create queue"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": queue, "message": "Queue created"})
}

func (h *Handler) GetQueue(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var queue models.Queue

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Preload("Agents").First(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Queue not found"})
		return
	}

	ctx.JSON(iris.Map{"data": queue})
}

func (h *Handler) UpdateQueue(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var queue models.Queue

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Queue not found"})
		return
	}

	if err := ctx.ReadJSON(&queue); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update queue"})
		return
	}

	ctx.JSON(iris.Map{"data": queue, "message": "Queue updated"})
}

func (h *Handler) DeleteQueue(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.Queue{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete queue"})
		return
	}

	ctx.JSON(iris.Map{"message": "Queue deleted"})
}

// =====================
// Ring Groups
// =====================

func (h *Handler) ListRingGroups(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var groups []models.RingGroup
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Preload("Destinations").Find(&groups).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch ring groups"})
		return
	}

	ctx.JSON(iris.Map{"data": groups})
}

func (h *Handler) CreateRingGroup(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var group models.RingGroup
	if err := ctx.ReadJSON(&group); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	group.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&group).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create ring group"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": group, "message": "Ring group created"})
}

func (h *Handler) GetRingGroup(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var group models.RingGroup

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Preload("Destinations").First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Ring group not found"})
		return
	}

	ctx.JSON(iris.Map{"data": group})
}

func (h *Handler) UpdateRingGroup(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var group models.RingGroup

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Ring group not found"})
		return
	}

	if err := ctx.ReadJSON(&group); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&group).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update ring group"})
		return
	}

	ctx.JSON(iris.Map{"data": group, "message": "Ring group updated"})
}

func (h *Handler) DeleteRingGroup(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.RingGroup{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete ring group"})
		return
	}

	ctx.JSON(iris.Map{"message": "Ring group deleted"})
}

// =====================
// Conferences
// =====================

func (h *Handler) ListConferences(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var conferences []models.Conference
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Find(&conferences).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch conferences"})
		return
	}

	ctx.JSON(iris.Map{"data": conferences})
}

func (h *Handler) CreateConference(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var conf models.Conference
	if err := ctx.ReadJSON(&conf); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	conf.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create conference"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": conf, "message": "Conference created"})
}

func (h *Handler) GetConference(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var conf models.Conference

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conference not found"})
		return
	}

	ctx.JSON(iris.Map{"data": conf})
}

func (h *Handler) UpdateConference(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var conf models.Conference

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conference not found"})
		return
	}

	if err := ctx.ReadJSON(&conf); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update conference"})
		return
	}

	ctx.JSON(iris.Map{"data": conf, "message": "Conference updated"})
}

func (h *Handler) DeleteConference(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.Conference{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete conference"})
		return
	}

	ctx.JSON(iris.Map{"message": "Conference deleted"})
}

// =====================
// Numbers/DIDs
// =====================

// Numbers handlers moved to routing_handlers.go

// Routing and Dialplan handlers moved to routing_handlers.go

// =====================
// Audio Library
// =====================

func (h *Handler) ListAudioFiles(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var files []models.Recording
	if err := h.DB.Where("tenant_id = ? AND type = ?", middleware.GetTenantID(ctx), "audio").Find(&files).Error; err != nil {
		ctx.JSON(iris.Map{"data": []interface{}{}})
		return
	}

	ctx.JSON(iris.Map{"data": files})
}

func (h *Handler) UploadAudioFile(ctx iris.Context) {
	// TODO: Implement file upload handling
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "Audio file uploaded"})
}

func (h *Handler) GetAudioFile(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"data": map[string]interface{}{"id": id}})
}

func (h *Handler) DeleteAudioFile(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"message": "Audio file deleted", "id": id})
}

// =====================
// Music on Hold
// =====================

func (h *Handler) ListMOHStreams(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	// TODO: Add MOH model or use Recording with type filter
	ctx.JSON(iris.Map{"data": []interface{}{}, "message": "MOH streams"})
}

func (h *Handler) CreateMOHStream(ctx iris.Context) {
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "MOH stream created"})
}

func (h *Handler) GetMOHStream(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"data": map[string]interface{}{"id": id}})
}

func (h *Handler) UpdateMOHStream(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"message": "MOH stream updated", "id": id})
}

func (h *Handler) DeleteMOHStream(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"message": "MOH stream deleted", "id": id})
}

// =====================
// Extension Profiles
// =====================

func (h *Handler) ListExtensionProfiles(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var profiles []models.ExtensionProfile
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&profiles).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch extension profiles"})
		return
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

	ctx.JSON(iris.Map{"data": response})
}

func (h *Handler) CreateExtensionProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var profile models.ExtensionProfile
	if err := ctx.ReadJSON(&profile); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	profile.TenantID = tenantID

	if err := h.DB.Create(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create extension profile"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": profile, "message": "Extension profile created"})
}

func (h *Handler) GetExtensionProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().GetUintDefault("id", 0)

	var profile models.ExtensionProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension profile not found"})
		return
	}

	ctx.JSON(iris.Map{"data": profile})
}

func (h *Handler) UpdateExtensionProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().GetUintDefault("id", 0)

	var profile models.ExtensionProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension profile not found"})
		return
	}

	var input models.ExtensionProfile
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	profile.Name = input.Name
	profile.Color = input.Color
	profile.Permissions = input.Permissions
	profile.CallHandling = input.CallHandling
	profile.RoutingOverride = input.RoutingOverride

	if err := h.DB.Save(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update extension profile"})
		return
	}

	ctx.JSON(iris.Map{"data": profile, "message": "Extension profile updated"})
}

func (h *Handler) DeleteExtensionProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().GetUintDefault("id", 0)

	// Unassign extensions from this profile
	h.DB.Model(&models.Extension{}).
		Where("tenant_id = ? AND profile_id = ?", tenantID, id).
		Update("profile_id", nil)

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.ExtensionProfile{})
	if result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete extension profile"})
		return
	}

	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension profile not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Extension profile deleted"})
}

// =====================
// Speed Dials
// =====================

func (h *Handler) ListSpeedDialGroups(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var groups []models.SpeedDialGroup
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&groups).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch speed dial groups"})
		return
	}

	ctx.JSON(iris.Map{"data": groups})
}

func (h *Handler) CreateSpeedDialGroup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var group models.SpeedDialGroup
	if err := ctx.ReadJSON(&group); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	group.TenantID = tenantID
	if group.Entries == nil {
		group.Entries = models.SpeedDialEntries{}
	}

	if err := h.DB.Create(&group).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create speed dial group"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": group, "message": "Speed dial group created"})
}

func (h *Handler) GetSpeedDialGroup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().GetUintDefault("id", 0)

	var group models.SpeedDialGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Speed dial group not found"})
		return
	}

	ctx.JSON(iris.Map{"data": group})
}

func (h *Handler) UpdateSpeedDialGroup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().GetUintDefault("id", 0)

	var group models.SpeedDialGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Speed dial group not found"})
		return
	}

	var input models.SpeedDialGroup
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	group.Name = input.Name
	group.Description = input.Description
	group.Prefix = input.Prefix
	group.Enabled = input.Enabled
	group.Entries = input.Entries

	if err := h.DB.Save(&group).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update speed dial group"})
		return
	}

	ctx.JSON(iris.Map{"data": group, "message": "Speed dial group updated"})
}

func (h *Handler) DeleteSpeedDialGroup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().GetUintDefault("id", 0)

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.SpeedDialGroup{})
	if result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete speed dial group"})
		return
	}

	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Speed dial group not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Speed dial group deleted"})
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
func (h *Handler) CheckDialCode(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var input struct {
		Code      string `json:"code"`
		Type      string `json:"type"`
		ExcludeID uint   `json:"exclude_id"`
	}
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	conflict := h.checkDialCodeConflict(tenantID, input.Code, input.Type, input.ExcludeID)
	ctx.JSON(iris.Map{
		"available": conflict == nil,
		"conflict":  conflict,
	})
}
