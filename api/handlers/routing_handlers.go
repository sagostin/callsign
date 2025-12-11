package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
)

// =====================
// Feature Codes
// =====================

func (h *Handler) ListFeatureCodes(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var codes []models.FeatureCode
	h.DB.Where("tenant_id = ?", tenantID).Order("code").Find(&codes)

	ctx.JSON(codes)
}

func (h *Handler) ListSystemFeatureCodes(ctx iris.Context) {
	// Return the hardcoded system feature codes
	systemCodes := []map[string]interface{}{
		{"code": "*97", "action": "voicemail_check", "description": "Check Voicemail"},
		{"code": "*72", "action": "call_forward_enable", "description": "Enable Call Forward"},
		{"code": "*73", "action": "call_forward_disable", "description": "Disable Call Forward"},
		{"code": "*78", "action": "dnd_enable", "description": "Enable Do Not Disturb"},
		{"code": "*79", "action": "dnd_disable", "description": "Disable Do Not Disturb"},
		{"code": "*70", "action": "park", "description": "Park Call"},
		{"code": "*71", "action": "pickup", "description": "Pickup Parked Call"},
		{"code": "*0", "action": "intercom", "description": "Intercom"},
		{"code": "*1", "action": "blind_transfer", "description": "Blind Transfer"},
		{"code": "*2", "action": "attended_transfer", "description": "Attended Transfer"},
		{"code": "*67", "action": "caller_id_block", "description": "Block Caller ID"},
		{"code": "*82", "action": "caller_id_unblock", "description": "Unblock Caller ID"},
	}
	ctx.JSON(systemCodes)
}

func (h *Handler) CreateFeatureCode(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var code models.FeatureCode
	if err := ctx.ReadJSON(&code); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	code.TenantID = &tenantID

	if err := code.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&code).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(code)
}

func (h *Handler) GetFeatureCode(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var code models.FeatureCode
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&code).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Feature code not found"})
		return
	}

	ctx.JSON(code)
}

func (h *Handler) UpdateFeatureCode(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var code models.FeatureCode
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&code).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Feature code not found"})
		return
	}

	if err := ctx.ReadJSON(&code); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	code.TenantID = &tenantID
	h.DB.Save(&code)
	ctx.JSON(code)
}

func (h *Handler) DeleteFeatureCode(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var code models.FeatureCode
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&code).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Feature code not found"})
		return
	}

	h.DB.Delete(&code)
	ctx.StatusCode(http.StatusNoContent)
}

// =====================
// Time Conditions
// =====================

func (h *Handler) ListTimeConditions(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var conditions []models.TimeCondition
	h.DB.Where("tenant_id = ?", tenantID).Order("name").Find(&conditions)

	ctx.JSON(conditions)
}

func (h *Handler) CreateTimeCondition(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var condition models.TimeCondition
	if err := ctx.ReadJSON(&condition); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	condition.TenantID = tenantID

	if err := h.DB.Create(&condition).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(condition)
}

func (h *Handler) GetTimeCondition(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var condition models.TimeCondition
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&condition).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Time condition not found"})
		return
	}

	ctx.JSON(condition)
}

func (h *Handler) UpdateTimeCondition(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var condition models.TimeCondition
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&condition).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Time condition not found"})
		return
	}

	if err := ctx.ReadJSON(&condition); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	condition.TenantID = tenantID
	h.DB.Save(&condition)
	ctx.JSON(condition)
}

func (h *Handler) DeleteTimeCondition(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var condition models.TimeCondition
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&condition).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Time condition not found"})
		return
	}

	h.DB.Delete(&condition)
	ctx.StatusCode(http.StatusNoContent)
}

// =====================
// Call Flows
// =====================

func (h *Handler) ListCallFlows(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var flows []models.CallFlow
	h.DB.Where("tenant_id = ?", tenantID).Order("name").Find(&flows)

	ctx.JSON(flows)
}

func (h *Handler) CreateCallFlow(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var flow models.CallFlow
	if err := ctx.ReadJSON(&flow); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	flow.TenantID = tenantID

	if err := h.DB.Create(&flow).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(flow)
}

func (h *Handler) GetCallFlow(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var flow models.CallFlow
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&flow).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call flow not found"})
		return
	}

	ctx.JSON(flow)
}

func (h *Handler) UpdateCallFlow(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var flow models.CallFlow
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&flow).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call flow not found"})
		return
	}

	if err := ctx.ReadJSON(&flow); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	flow.TenantID = tenantID
	h.DB.Save(&flow)
	ctx.JSON(flow)
}

func (h *Handler) DeleteCallFlow(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var flow models.CallFlow
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&flow).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call flow not found"})
		return
	}

	h.DB.Delete(&flow)
	ctx.StatusCode(http.StatusNoContent)
}

func (h *Handler) ToggleCallFlow(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var flow models.CallFlow
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&flow).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call flow not found"})
		return
	}

	// Toggle status: day -> night -> holiday -> day
	switch flow.Status {
	case "day":
		flow.Status = "night"
	case "night":
		flow.Status = "holiday"
	default:
		flow.Status = "day"
	}

	h.DB.Save(&flow)
	ctx.JSON(flow)
}
