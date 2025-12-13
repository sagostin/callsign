package handlers

import (
	"net/http"
	"strconv"

	"callsign/middleware"
	"callsign/models"

	"github.com/kataras/iris/v12"
)

// ListCallHandlingRules lists call handling rules for an extension
func (h *Handler) ListCallHandlingRules(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	extID, _ := strconv.Atoi(ctx.Params().Get("ext"))

	var rules []models.CallHandlingRule
	if err := h.DB.Where("extension_id = ? AND tenant_id = ?", extID, tenantID).
		Order("priority ASC, id ASC").
		Find(&rules).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to load call handling rules"})
		return
	}

	ctx.JSON(iris.Map{"data": rules})
}

// CreateCallHandlingRule creates a new call handling rule for an extension
func (h *Handler) CreateCallHandlingRule(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	extID, _ := strconv.Atoi(ctx.Params().Get("ext"))

	// Verify extension belongs to tenant
	var ext models.Extension
	if err := h.DB.Where("id = ? AND tenant_id = ?", extID, tenantID).First(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension not found"})
		return
	}

	var rule models.CallHandlingRule
	if err := ctx.ReadJSON(&rule); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
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
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create call handling rule"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": rule, "message": "Call handling rule created"})
}

// UpdateCallHandlingRule updates a call handling rule
func (h *Handler) UpdateCallHandlingRule(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	ruleID, _ := strconv.Atoi(ctx.Params().Get("ruleId"))

	var rule models.CallHandlingRule
	if err := h.DB.Where("id = ? AND tenant_id = ?", ruleID, tenantID).First(&rule).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call handling rule not found"})
		return
	}

	var input models.CallHandlingRule
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
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
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update call handling rule"})
		return
	}

	ctx.JSON(iris.Map{"data": rule, "message": "Call handling rule updated"})
}

// DeleteCallHandlingRule deletes a call handling rule
func (h *Handler) DeleteCallHandlingRule(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	ruleID, _ := strconv.Atoi(ctx.Params().Get("ruleId"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", ruleID, tenantID).
		Delete(&models.CallHandlingRule{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete call handling rule"})
		return
	}

	ctx.JSON(iris.Map{"message": "Call handling rule deleted"})
}

// ReorderCallHandlingRules updates priorities for drag-drop reordering
func (h *Handler) ReorderCallHandlingRules(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	extID, _ := strconv.Atoi(ctx.Params().Get("ext"))

	var input struct {
		RuleIDs []uint `json:"rule_ids"` // Ordered list of rule IDs
	}
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Update priorities in order
	for i, ruleID := range input.RuleIDs {
		h.DB.Model(&models.CallHandlingRule{}).
			Where("id = ? AND extension_id = ? AND tenant_id = ?", ruleID, extID, tenantID).
			Update("priority", (i+1)*10)
	}

	ctx.JSON(iris.Map{"message": "Rules reordered"})
}

// ============ Profile-level Call Handling Rules ============

// ListProfileCallHandlingRules lists call handling rules for a profile
func (h *Handler) ListProfileCallHandlingRules(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	profileID, _ := strconv.Atoi(ctx.Params().Get("id"))

	var rules []models.CallHandlingRule
	if err := h.DB.Where("profile_id = ? AND tenant_id = ?", profileID, tenantID).
		Order("priority ASC, id ASC").
		Find(&rules).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to load call handling rules"})
		return
	}

	ctx.JSON(iris.Map{"data": rules})
}

// CreateProfileCallHandlingRule creates a new call handling rule for a profile
func (h *Handler) CreateProfileCallHandlingRule(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	profileID, _ := strconv.Atoi(ctx.Params().Get("id"))

	// Verify profile belongs to tenant
	var profile models.ExtensionProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", profileID, tenantID).First(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension profile not found"})
		return
	}

	var rule models.CallHandlingRule
	if err := ctx.ReadJSON(&rule); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
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
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create call handling rule"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": rule, "message": "Call handling rule created"})
}

// UpdateProfileCallHandlingRule updates a profile call handling rule
func (h *Handler) UpdateProfileCallHandlingRule(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	ruleID, _ := strconv.Atoi(ctx.Params().Get("ruleId"))

	var rule models.CallHandlingRule
	if err := h.DB.Where("id = ? AND tenant_id = ?", ruleID, tenantID).First(&rule).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call handling rule not found"})
		return
	}

	var input models.CallHandlingRule
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
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
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update call handling rule"})
		return
	}

	ctx.JSON(iris.Map{"data": rule, "message": "Call handling rule updated"})
}

// DeleteProfileCallHandlingRule deletes a profile call handling rule
func (h *Handler) DeleteProfileCallHandlingRule(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	ruleID, _ := strconv.Atoi(ctx.Params().Get("ruleId"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", ruleID, tenantID).
		Delete(&models.CallHandlingRule{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete call handling rule"})
		return
	}

	ctx.JSON(iris.Map{"message": "Call handling rule deleted"})
}

// ReorderProfileCallHandlingRules updates priorities for a profile's rules
func (h *Handler) ReorderProfileCallHandlingRules(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	profileID, _ := strconv.Atoi(ctx.Params().Get("id"))

	var input struct {
		RuleIDs []uint `json:"rule_ids"`
	}
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	for i, ruleID := range input.RuleIDs {
		h.DB.Model(&models.CallHandlingRule{}).
			Where("id = ? AND profile_id = ? AND tenant_id = ?", ruleID, profileID, tenantID).
			Update("priority", (i+1)*10)
	}

	ctx.JSON(iris.Map{"message": "Rules reordered"})
}
