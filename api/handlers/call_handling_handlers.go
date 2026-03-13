package handlers

import (
	"net/http"
	"strconv"

	"callsign/middleware"
	"callsign/models"

	"github.com/gofiber/fiber/v2"
)

// ListCallHandlingRules lists call handling rules for an extension
func (h *Handler) ListCallHandlingRules(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	extID, _ := strconv.Atoi(c.Params("ext"))

	var rules []models.CallHandlingRule
	if err := h.DB.Where("extension_id = ? AND tenant_id = ?", extID, tenantID).
		Order("priority ASC, id ASC").
		Find(&rules).Error; err != nil {
		h.logError("CALL", "ListCallHandlingRules: Failed to load call handling rules", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load call handling rules"})
	}

	return c.JSON(fiber.Map{"data": rules})
}

// CreateCallHandlingRule creates a new call handling rule for an extension
func (h *Handler) CreateCallHandlingRule(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	extID, _ := strconv.Atoi(c.Params("ext"))

	// Verify extension belongs to tenant
	var ext models.Extension
	if err := h.DB.Where("id = ? AND tenant_id = ?", extID, tenantID).First(&ext).Error; err != nil {
		h.logWarn("CALL", "CreateCallHandlingRule: Extension not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension not found"})
	}

	var rule models.CallHandlingRule
	if err := c.BodyParser(&rule); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rule.TenantID = tenantID
	extIDUint := uint(extID)
	rule.ExtensionID = &extIDUint

	// Set default priority if not specified (append to end)
	if rule.Priority == 0 {
		var maxPriority int
		h.DB.Model(&models.CallHandlingRule{}).
			Where("extension_id = ?", extID).
			Select("COALESCE(MAX(priority), 0)").
			Scan(&maxPriority)
		rule.Priority = maxPriority + 10
	}

	if err := h.DB.Create(&rule).Error; err != nil {
		h.logError("CALL", "CreateCallHandlingRule: Failed to create call handling rule", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create call handling rule"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": rule, "message": "Call handling rule created"})
}

// UpdateCallHandlingRule updates a call handling rule
func (h *Handler) UpdateCallHandlingRule(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	ruleID, _ := strconv.Atoi(c.Params("ruleId"))

	var rule models.CallHandlingRule
	if err := h.DB.Where("id = ? AND tenant_id = ?", ruleID, tenantID).First(&rule).Error; err != nil {
		h.logWarn("CALL", "UpdateCallHandlingRule: Call handling rule not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call handling rule not found"})
	}

	var input models.CallHandlingRule
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Update fields
	rule.Name = input.Name
	rule.Description = input.Description
	rule.Priority = input.Priority
	rule.Enabled = input.Enabled
	rule.Events = input.Events
	rule.Conditions = input.Conditions
	rule.ActionType = input.ActionType
	rule.ActionTarget = input.ActionTarget
	rule.ActionParams = input.ActionParams

	if err := h.DB.Save(&rule).Error; err != nil {
		h.logError("CALL", "UpdateCallHandlingRule: Failed to update call handling rule", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update call handling rule"})
	}

	return c.JSON(fiber.Map{"data": rule, "message": "Call handling rule updated"})
}

// DeleteCallHandlingRule deletes a call handling rule
func (h *Handler) DeleteCallHandlingRule(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	ruleID, _ := strconv.Atoi(c.Params("ruleId"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", ruleID, tenantID).
		Delete(&models.CallHandlingRule{}).Error; err != nil {
		h.logError("CALL", "DeleteCallHandlingRule: Failed to delete call handling rule", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete call handling rule"})
	}

	return c.JSON(fiber.Map{"message": "Call handling rule deleted"})
}

// ReorderCallHandlingRules updates priorities for drag-drop reordering
func (h *Handler) ReorderCallHandlingRules(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	extID, _ := strconv.Atoi(c.Params("ext"))

	var input struct {
		RuleIDs []uint `json:"rule_ids"` // Ordered list of rule IDs
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Update priorities in order
	for i, ruleID := range input.RuleIDs {
		h.DB.Model(&models.CallHandlingRule{}).
			Where("id = ? AND extension_id = ? AND tenant_id = ?", ruleID, extID, tenantID).
			Update("priority", (i+1)*10)
	}

	return c.JSON(fiber.Map{"message": "Rules reordered"})
}

// ============ Profile-level Call Handling Rules ============

// ListProfileCallHandlingRules lists call handling rules for a profile
func (h *Handler) ListProfileCallHandlingRules(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	profileID, _ := strconv.Atoi(c.Params("id"))

	var rules []models.CallHandlingRule
	if err := h.DB.Where("profile_id = ? AND tenant_id = ?", profileID, tenantID).
		Order("priority ASC, id ASC").
		Find(&rules).Error; err != nil {
		h.logError("CALL", "ListProfileCallHandlingRules: Failed to load call handling rules", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load call handling rules"})
	}

	return c.JSON(fiber.Map{"data": rules})
}

// CreateProfileCallHandlingRule creates a new call handling rule for a profile
func (h *Handler) CreateProfileCallHandlingRule(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	profileID, _ := strconv.Atoi(c.Params("id"))

	// Verify profile belongs to tenant
	var profile models.ExtensionProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", profileID, tenantID).First(&profile).Error; err != nil {
		h.logWarn("CALL", "CreateProfileCallHandlingRule: Extension profile not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension profile not found"})
	}

	var rule models.CallHandlingRule
	if err := c.BodyParser(&rule); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rule.TenantID = tenantID
	profileIDUint := uint(profileID)
	rule.ProfileID = &profileIDUint

	// Set default priority
	if rule.Priority == 0 {
		var maxPriority int
		h.DB.Model(&models.CallHandlingRule{}).
			Where("profile_id = ?", profileID).
			Select("COALESCE(MAX(priority), 0)").
			Scan(&maxPriority)
		rule.Priority = maxPriority + 10
	}

	if err := h.DB.Create(&rule).Error; err != nil {
		h.logError("CALL", "CreateProfileCallHandlingRule: Failed to create call handling rule", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create call handling rule"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": rule, "message": "Call handling rule created"})
}

// UpdateProfileCallHandlingRule updates a profile call handling rule
func (h *Handler) UpdateProfileCallHandlingRule(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	ruleID, _ := strconv.Atoi(c.Params("ruleId"))

	var rule models.CallHandlingRule
	if err := h.DB.Where("id = ? AND tenant_id = ?", ruleID, tenantID).First(&rule).Error; err != nil {
		h.logWarn("CALL", "UpdateProfileCallHandlingRule: Call handling rule not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call handling rule not found"})
	}

	var input models.CallHandlingRule
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rule.Name = input.Name
	rule.Description = input.Description
	rule.Priority = input.Priority
	rule.Enabled = input.Enabled
	rule.Events = input.Events
	rule.Conditions = input.Conditions
	rule.ActionType = input.ActionType
	rule.ActionTarget = input.ActionTarget
	rule.ActionParams = input.ActionParams

	if err := h.DB.Save(&rule).Error; err != nil {
		h.logError("CALL", "UpdateProfileCallHandlingRule: Failed to update call handling rule", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update call handling rule"})
	}

	return c.JSON(fiber.Map{"data": rule, "message": "Call handling rule updated"})
}

// DeleteProfileCallHandlingRule deletes a profile call handling rule
func (h *Handler) DeleteProfileCallHandlingRule(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	ruleID, _ := strconv.Atoi(c.Params("ruleId"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", ruleID, tenantID).
		Delete(&models.CallHandlingRule{}).Error; err != nil {
		h.logError("CALL", "DeleteProfileCallHandlingRule: Failed to delete call handling rule", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete call handling rule"})
	}

	return c.JSON(fiber.Map{"message": "Call handling rule deleted"})
}

// ReorderProfileCallHandlingRules updates priorities for a profile's rules
func (h *Handler) ReorderProfileCallHandlingRules(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	profileID, _ := strconv.Atoi(c.Params("id"))

	var input struct {
		RuleIDs []uint `json:"rule_ids"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	for i, ruleID := range input.RuleIDs {
		h.DB.Model(&models.CallHandlingRule{}).
			Where("id = ? AND profile_id = ? AND tenant_id = ?", ruleID, profileID, tenantID).
			Update("priority", (i+1)*10)
	}

	return c.JSON(fiber.Map{"message": "Rules reordered"})
}
