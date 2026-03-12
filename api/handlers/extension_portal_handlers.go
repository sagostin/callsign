package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/kataras/iris/v12"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// =========================================
// Extension Portal Handlers
// =========================================
// These handlers are scoped to the logged-in extension (via ExtensionID in
// the JWT claims) rather than requiring a User model lookup. They are the
// primary data-access layer for the extension panel / web-client UI.

// resolveExtension loads the Extension from claims.ExtensionID, or falls
// back to looking it up via claims.UserID when the token was generated from
// a traditional user login.
func (h *Handler) resolveExtension(ctx iris.Context) (*models.Extension, bool) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return nil, false
	}

	var ext models.Extension
	tenantID := middleware.GetTenantID(ctx)

	// Prefer ExtensionID from JWT (extension login)
	if claims.ExtensionID != nil {
		if err := h.DB.Where("id = ? AND tenant_id = ?", *claims.ExtensionID, tenantID).First(&ext).Error; err == nil {
			return &ext, true
		}
	}

	// Fallback: look up by user_id (admin/user login with an assigned extension)
	if claims.UserID > 0 {
		if err := h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, tenantID).First(&ext).Error; err == nil {
			return &ext, true
		}
	}

	ctx.StatusCode(http.StatusNotFound)
	ctx.JSON(iris.Map{"error": "No extension associated with this session"})
	return nil, false
}

// GetExtensionDevices returns devices/registrations for the logged-in extension.
func (h *Handler) GetExtensionDevices(ctx iris.Context) {
	ext, ok := h.resolveExtension(ctx)
	if !ok {
		return
	}

	tenantID := middleware.GetTenantID(ctx)

	// Collect hardware devices linked via DeviceUUID
	devices := make([]map[string]interface{}, 0)
	if ext.DeviceUUID != "" {
		devices = append(devices, map[string]interface{}{
			"device_uuid": ext.DeviceUUID,
			"extension":   ext.Extension,
			"name":        ext.EffectiveCallerIDName,
		})
	}

	// Also include client registrations for this extension
	var registrations []models.ClientRegistration
	h.DB.Where("extension_id = ? AND tenant_id = ?", ext.ID, tenantID).Find(&registrations)
	for _, reg := range registrations {
		devices = append(devices, map[string]interface{}{
			"registration_id": reg.ID,
			"device_type":     reg.EndpointType,
			"user_agent":      reg.UserAgent,
			"extension":       ext.Extension,
			"name":            ext.EffectiveCallerIDName,
		})
	}

	ctx.JSON(iris.Map{"data": devices})
}

// GetExtensionCallHistory returns CDR records for the logged-in extension.
func (h *Handler) GetExtensionCallHistory(ctx iris.Context) {
	ext, ok := h.resolveExtension(ctx)
	if !ok {
		return
	}

	limit := 100
	if l := ctx.URLParam("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}

	tenantID := middleware.GetTenantID(ctx)

	var cdrs []models.CallRecord
	if err := h.DB.Where("tenant_id = ? AND (caller_id_number = ? OR destination_number = ?)",
		tenantID, ext.Extension, ext.Extension).
		Order("start_time DESC").
		Limit(limit).
		Find(&cdrs).Error; err != nil {
		ctx.JSON(iris.Map{"data": []interface{}{}})
		return
	}

	ctx.JSON(iris.Map{"data": cdrs})
}

// GetExtensionVoicemail returns the voicemail box and messages for the extension.
func (h *Handler) GetExtensionVoicemail(ctx iris.Context) {
	ext, ok := h.resolveExtension(ctx)
	if !ok {
		return
	}

	tenantID := middleware.GetTenantID(ctx)

	var box models.VoicemailBox
	if err := h.DB.Where("extension = ? AND tenant_id = ?", ext.Extension, tenantID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&box).Error; err != nil {
		ctx.JSON(iris.Map{"data": nil, "messages": []interface{}{}})
		return
	}

	ctx.JSON(iris.Map{
		"data":        box,
		"messages":    box.Messages,
		"new_count":   box.NewMessages,
		"saved_count": box.SavedMessages,
	})
}

// GetExtensionSettings returns the call settings for the logged-in extension.
func (h *Handler) GetExtensionSettings(ctx iris.Context) {
	ext, ok := h.resolveExtension(ctx)
	if !ok {
		return
	}

	ctx.JSON(iris.Map{
		"extension": map[string]interface{}{
			"id":        ext.ID,
			"extension": ext.Extension,
			"name":      ext.EffectiveCallerIDName,
		},
		"profile": map[string]interface{}{
			"first_name":               ext.DirectoryFirstName,
			"last_name":                ext.DirectoryLastName,
			"email":                    ext.VoicemailMailTo,
			"outbound_caller_id_name":  ext.OutboundCallerIDName,
			"effective_caller_id_name": ext.EffectiveCallerIDName,
		},
		"call_settings": map[string]interface{}{
			"do_not_disturb":            ext.DoNotDisturb,
			"forward_all_enabled":       ext.ForwardAllEnabled,
			"forward_all_dest":          ext.ForwardAllDestination,
			"forward_busy_enabled":      ext.ForwardBusyEnabled,
			"forward_busy_dest":         ext.ForwardBusyDestination,
			"forward_no_answer_enabled": ext.ForwardNoAnswerEnabled,
			"forward_no_answer_dest":    ext.ForwardNoAnswerDestination,
			"voicemail_enabled":         ext.VoicemailEnabled,
			"voicemail_mail_to":         ext.VoicemailMailTo,
			"follow_me_enabled":         ext.FollowMeEnabled,
			"record_inbound":            ext.RecordInbound,
			"record_outbound":           ext.RecordOutbound,
			"ring_strategy":             ext.RingStrategy,
			"no_answer_action":          ext.NoAnswerAction,
			"no_answer_forward_to":      ext.NoAnswerForwardTo,
			"outbound_caller_id_name":   ext.OutboundCallerIDName,
		},
	})
}

// UpdateExtensionSettings updates the call settings for the logged-in extension.
func (h *Handler) UpdateExtensionSettings(ctx iris.Context) {
	ext, ok := h.resolveExtension(ctx)
	if !ok {
		return
	}

	var req struct {
		// Profile
		FirstName            *string `json:"first_name"`
		LastName             *string `json:"last_name"`
		Email                *string `json:"email"`
		OutboundCallerIDName *string `json:"outbound_caller_id_name"`
		// Call settings
		DoNotDisturb           *bool   `json:"do_not_disturb"`
		FollowMeEnabled        *bool   `json:"follow_me_enabled"`
		ForwardAllEnabled      *bool   `json:"forward_all_enabled"`
		ForwardAllDestination  *string `json:"forward_all_destination"`
		ForwardBusyEnabled     *bool   `json:"forward_busy_enabled"`
		ForwardBusyDestination *string `json:"forward_busy_destination"`
		ForwardNoAnswerEnabled *bool   `json:"forward_no_answer_enabled"`
		ForwardNoAnswerDest    *string `json:"forward_no_answer_dest"`
		VoicemailEnabled       *bool   `json:"voicemail_enabled"`
		VoicemailMailTo        *string `json:"voicemail_mail_to"`
		RecordInbound          *bool   `json:"record_inbound"`
		RecordOutbound         *bool   `json:"record_outbound"`
		RingStrategy           *string `json:"ring_strategy"`
		NoAnswerAction         *string `json:"no_answer_action"`
		NoAnswerForwardTo      *string `json:"no_answer_forward_to"`
	}

	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Profile fields
	if req.FirstName != nil {
		ext.DirectoryFirstName = *req.FirstName
	}
	if req.LastName != nil {
		ext.DirectoryLastName = *req.LastName
	}
	if req.Email != nil {
		ext.VoicemailMailTo = *req.Email
	}
	if req.OutboundCallerIDName != nil {
		ext.OutboundCallerIDName = *req.OutboundCallerIDName
	}
	// Call settings
	if req.DoNotDisturb != nil {
		ext.DoNotDisturb = *req.DoNotDisturb
	}
	if req.FollowMeEnabled != nil {
		ext.FollowMeEnabled = *req.FollowMeEnabled
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
	if req.ForwardNoAnswerEnabled != nil {
		ext.ForwardNoAnswerEnabled = *req.ForwardNoAnswerEnabled
	}
	if req.ForwardNoAnswerDest != nil {
		ext.ForwardNoAnswerDestination = *req.ForwardNoAnswerDest
	}
	if req.VoicemailEnabled != nil {
		ext.VoicemailEnabled = *req.VoicemailEnabled
	}
	if req.VoicemailMailTo != nil {
		ext.VoicemailMailTo = *req.VoicemailMailTo
	}
	if req.RecordInbound != nil {
		ext.RecordInbound = *req.RecordInbound
	}
	if req.RecordOutbound != nil {
		ext.RecordOutbound = *req.RecordOutbound
	}
	if req.RingStrategy != nil {
		ext.RingStrategy = *req.RingStrategy
	}
	if req.NoAnswerAction != nil {
		ext.NoAnswerAction = *req.NoAnswerAction
	}
	if req.NoAnswerForwardTo != nil {
		ext.NoAnswerForwardTo = *req.NoAnswerForwardTo
	}

	if err := h.DB.Save(ext).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update settings"})
		return
	}

	ctx.JSON(iris.Map{"message": "Settings updated"})
}

// ChangeExtensionPassword changes the web password for the logged-in extension.
func (h *Handler) ChangeExtensionPassword(ctx iris.Context) {
	ext, ok := h.resolveExtension(ctx)
	if !ok {
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	if req.NewPassword == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "New password is required"})
		return
	}

	// If web password is set, verify current password
	if ext.WebPassword != "" {
		if bcrypt.CompareHashAndPassword([]byte(ext.WebPassword), []byte(req.CurrentPassword)) != nil {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{"error": "Current password is incorrect"})
			return
		}
	}

	if err := ext.SetWebPassword(req.NewPassword); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to hash password"})
		return
	}

	if err := h.DB.Save(ext).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update password"})
		return
	}

	ctx.JSON(iris.Map{"message": "Password updated successfully"})
}

// GetExtensionContacts returns personal contacts for the extension's tenant.
func (h *Handler) GetExtensionContacts(ctx iris.Context) {
	_, ok := h.resolveExtension(ctx)
	if !ok {
		return
	}

	tenantID := middleware.GetTenantID(ctx)

	var contacts []models.Contact
	if err := h.DB.Where("tenant_id = ?", tenantID).Find(&contacts).Error; err != nil {
		ctx.JSON(iris.Map{"data": []interface{}{}})
		return
	}

	ctx.JSON(iris.Map{"data": contacts})
}

// CreateExtensionContact creates a personal contact scoped to the extension.
func (h *Handler) CreateExtensionContact(ctx iris.Context) {
	ext, ok := h.resolveExtension(ctx)
	if !ok {
		return
	}

	var contact models.Contact
	if err := ctx.ReadJSON(&contact); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	contact.TenantID = middleware.GetTenantID(ctx)
	// Associate the contact with the extension if the model supports it
	_ = ext // available for future extension_id association

	if err := h.DB.Create(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create contact"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": contact, "message": "Contact created"})
}
