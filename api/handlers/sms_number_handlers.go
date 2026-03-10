package handlers

import (
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// SMSNumberHandler handles SMS number management API requests
type SMSNumberHandler struct {
	DB *gorm.DB
}

// NewSMSNumberHandler creates a new SMS number handler
func NewSMSNumberHandler(db *gorm.DB) *SMSNumberHandler {
	return &SMSNumberHandler{DB: db}
}

// ListSMSNumbers returns SMS-enabled DIDs for the tenant
func (h *SMSNumberHandler) ListSMSNumbers(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var numbers []models.Destination
	h.DB.Where("tenant_id = ? AND sms_enabled = true", tenantID).
		Order("destination_number").Find(&numbers)

	ctx.JSON(iris.Map{"data": numbers})
}

// ConfigureSMSNumber enables/disables SMS and sets the mode for a DID
func (h *SMSNumberHandler) ConfigureSMSNumber(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var dest models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&dest).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Number not found"})
		return
	}

	var input struct {
		SMSEnabled    *bool  `json:"sms_enabled"`
		SMSMode       string `json:"sms_mode"` // disabled, shared, dedicated
		SMSProviderID *uint  `json:"sms_provider_id"`
	}
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.SMSEnabled != nil {
		updates["sms_enabled"] = *input.SMSEnabled
	}
	if input.SMSMode != "" {
		// Validate mode
		switch input.SMSMode {
		case "disabled", "shared", "dedicated":
			updates["sms_mode"] = input.SMSMode
		default:
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "Invalid SMS mode. Must be: disabled, shared, or dedicated"})
			return
		}
	}
	if input.SMSProviderID != nil {
		// Verify provider exists
		var provider models.MessagingProvider
		if err := h.DB.First(&provider, *input.SMSProviderID).Error; err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "Messaging provider not found"})
			return
		}
		updates["sms_provider_id"] = input.SMSProviderID
	}

	h.DB.Model(&dest).Updates(updates)

	// Reload
	h.DB.First(&dest, id)
	ctx.JSON(iris.Map{"data": dest})
}

// ListNumberAssignments returns extension assignments for an SMS number
func (h *SMSNumberHandler) ListNumberAssignments(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	// Verify the destination exists and belongs to tenant
	var dest models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&dest).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Number not found"})
		return
	}

	var assignments []models.SMSNumberAssignment
	h.DB.Where("destination_id = ? AND tenant_id = ?", id, tenantID).Find(&assignments)

	ctx.JSON(iris.Map{"data": assignments})
}

// AssignNumber assigns a dedicated SMS number to an extension
func (h *SMSNumberHandler) AssignNumber(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	// Verify the destination exists and is SMS-enabled
	var dest models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ? AND sms_enabled = true", id, tenantID).First(&dest).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "SMS-enabled number not found"})
		return
	}

	var input struct {
		ExtensionID uint `json:"extension_id"`
		IsDefault   bool `json:"is_default"`
	}
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	if input.ExtensionID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "extension_id is required"})
		return
	}

	// Check for existing assignment
	var existing models.SMSNumberAssignment
	if err := h.DB.Where("destination_id = ? AND extension_id = ? AND tenant_id = ?",
		id, input.ExtensionID, tenantID).First(&existing).Error; err == nil {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": "This number is already assigned to this extension"})
		return
	}

	// If setting as default, unset other defaults for this extension
	if input.IsDefault {
		h.DB.Model(&models.SMSNumberAssignment{}).
			Where("extension_id = ? AND tenant_id = ?", input.ExtensionID, tenantID).
			Update("is_default", false)
	}

	assignment := models.SMSNumberAssignment{
		TenantID:      tenantID,
		DestinationID: id,
		ExtensionID:   input.ExtensionID,
		IsDefault:     input.IsDefault,
		Enabled:       true,
	}

	if err := h.DB.Create(&assignment).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": assignment})
}

// UnassignNumber removes an SMS number assignment
func (h *SMSNumberHandler) UnassignNumber(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	assignID, _ := ctx.Params().GetUint("assignId")

	result := h.DB.Where("id = ? AND tenant_id = ?", assignID, tenantID).
		Delete(&models.SMSNumberAssignment{})

	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Assignment not found"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
}
