package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// =====================
// User Portal Handlers
// =====================

func (h *Handler) GetUserDevices(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	// Get user's extension
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	// Find extensions assigned to this user
	var extensions []models.Extension
	if err := h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, middleware.GetTenantID(ctx)).Find(&extensions).Error; err != nil {
		ctx.JSON(iris.Map{"data": []interface{}{}})
		return
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

	ctx.JSON(iris.Map{"data": devices})
}

func (h *Handler) GetUserCallHistory(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	// Get user's extension number
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	// Fetch CDR for the user's extension
	var cdrs []models.CallRecord
	if err := h.DB.Where("tenant_id = ? AND (caller_id_number = ? OR destination_number = ?)",
		middleware.GetTenantID(ctx), user.Extension, user.Extension).
		Order("start_time DESC").
		Limit(100).
		Find(&cdrs).Error; err != nil {
		ctx.JSON(iris.Map{"data": []interface{}{}})
		return
	}

	ctx.JSON(iris.Map{"data": cdrs})
}

func (h *Handler) GetUserVoicemail(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	// Get user's extension
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	// Find voicemail box for user's extension
	var box models.VoicemailBox
	if err := h.DB.Where("extension = ? AND tenant_id = ?", user.Extension, middleware.GetTenantID(ctx)).
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

func (h *Handler) GetUserSettings(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	// Get associated extension settings
	var ext models.Extension
	h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, middleware.GetTenantID(ctx)).First(&ext)

	ctx.JSON(iris.Map{
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

func (h *Handler) UpdateUserSettings(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
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

	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
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
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update user"})
		return
	}

	// Update extension settings if user has one
	var ext models.Extension
	if err := h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, middleware.GetTenantID(ctx)).First(&ext).Error; err == nil {
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

	ctx.JSON(iris.Map{"message": "Settings updated"})
}

func (h *Handler) GetUserContacts(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var contacts []models.Contact
	if err := h.DB.Where("user_id = ? AND tenant_id = ?", claims.UserID, middleware.GetTenantID(ctx)).Find(&contacts).Error; err != nil {
		ctx.JSON(iris.Map{"data": []interface{}{}})
		return
	}

	ctx.JSON(iris.Map{"data": contacts})
}

func (h *Handler) CreateUserContact(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var contact models.Contact
	if err := ctx.ReadJSON(&contact); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	contact.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&contact).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create contact"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": contact, "message": "Contact created"})
}
