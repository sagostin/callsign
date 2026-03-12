package handlers

import (
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
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
func (h *SMSNumberHandler) ListSMSNumbers(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var numbers []models.Destination
	h.DB.Where("tenant_id = ? AND sms_enabled = true", tenantID).
		Order("destination_number").Find(&numbers)

	return c.JSON(fiber.Map{"data": numbers})
}

// ConfigureSMSNumber enables/disables SMS and sets the mode for a DID
func (h *SMSNumberHandler) ConfigureSMSNumber(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var dest models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&dest).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number not found"})
	}

	var input struct {
		SMSEnabled    *bool  `json:"sms_enabled"`
		SMSMode       string `json:"sms_mode"` // disabled, shared, dedicated
		SMSProviderID *uint  `json:"sms_provider_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid SMS mode. Must be: disabled, shared, or dedicated"})
		}
	}
	if input.SMSProviderID != nil {
		// Verify provider exists
		var provider models.MessagingProvider
		if err := h.DB.First(&provider, *input.SMSProviderID).Error; err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Messaging provider not found"})
		}
		updates["sms_provider_id"] = input.SMSProviderID
	}

	h.DB.Model(&dest).Updates(updates)

	// Reload
	h.DB.First(&dest, id)
	return c.JSON(fiber.Map{"data": dest})
}

// ListNumberAssignments returns extension assignments for an SMS number
func (h *SMSNumberHandler) ListNumberAssignments(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	// Verify the destination exists and belongs to tenant
	var dest models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&dest).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number not found"})
	}

	var assignments []models.SMSNumberAssignment
	h.DB.Where("destination_id = ? AND tenant_id = ?", id, tenantID).Find(&assignments)

	return c.JSON(fiber.Map{"data": assignments})
}

// AssignNumber assigns a dedicated SMS number to an extension
func (h *SMSNumberHandler) AssignNumber(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	// Verify the destination exists and is SMS-enabled
	var dest models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ? AND sms_enabled = true", id, tenantID).First(&dest).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "SMS-enabled number not found"})
	}

	var input struct {
		ExtensionID uint `json:"extension_id"`
		IsDefault   bool `json:"is_default"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if input.ExtensionID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "extension_id is required"})
	}

	// Check for existing assignment
	var existing models.SMSNumberAssignment
	if err := h.DB.Where("destination_id = ? AND extension_id = ? AND tenant_id = ?",
		id, input.ExtensionID, tenantID).First(&existing).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "This number is already assigned to this extension"})
	}

	// If setting as default, unset other defaults for this extension
	if input.IsDefault {
		h.DB.Model(&models.SMSNumberAssignment{}).
			Where("extension_id = ? AND tenant_id = ?", input.ExtensionID, tenantID).
			Update("is_default", false)
	}

	assignment := models.SMSNumberAssignment{
		TenantID:      tenantID,
		DestinationID: uint(id),
		ExtensionID:   input.ExtensionID,
		IsDefault:     input.IsDefault,
		Enabled:       true,
	}

	if err := h.DB.Create(&assignment).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": assignment})
}

// UnassignNumber removes an SMS number assignment
func (h *SMSNumberHandler) UnassignNumber(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	assignID, _ := strconv.ParseUint(c.Params("assignId"), 10, 64)

	result := h.DB.Where("id = ? AND tenant_id = ?", assignID, tenantID).
		Delete(&models.SMSNumberAssignment{})

	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Assignment not found"})
	}

	c.Status(http.StatusNoContent)
	return nil
}
