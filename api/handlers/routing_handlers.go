package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// =====================
// Inbound Routes (Dialplan context=public)
// =====================

func (h *Handler) ListInboundRoutes(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var routes []models.Dialplan
	// Inbound routes typically live in the 'public' context
	if err := h.DB.Where("tenant_id = ? AND dialplan_context = ?", tenantID, "public").
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).
		Order("dialplan_order ASC").Find(&routes).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve inbound routes"})
		return
	}

	ctx.JSON(iris.Map{"data": routes})
}

func (h *Handler) CreateInboundRoute(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var route models.Dialplan
	if err := ctx.ReadJSON(&route); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	route.TenantID = &tenantID
	if route.DialplanContext == "" {
		route.DialplanContext = "public"
	}

	if err := h.DB.Create(&route).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create inbound route"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(route)
}

// =====================
// Outbound Routes (Dialplan context=default/domain)
// =====================

func (h *Handler) ListOutboundRoutes(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var routes []models.Dialplan
	// Outbound routes live in the 'default' context (or the tenant's domain context)
	if err := h.DB.Where("tenant_id = ? AND dialplan_context != ?", tenantID, "public").
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).
		Order("dialplan_order ASC").Find(&routes).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve outbound routes"})
		return
	}

	ctx.JSON(iris.Map{"data": routes})
}

func (h *Handler) CreateOutboundRoute(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var route models.Dialplan
	if err := ctx.ReadJSON(&route); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	route.TenantID = &tenantID
	if route.DialplanContext == "" {
		route.DialplanContext = "default"
	}

	if err := h.DB.Create(&route).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create outbound route"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(route)
}

// =====================
// Generic Dial Plans (CRUD)
// =====================

func (h *Handler) ListDialPlans(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var dialplans []models.Dialplan
	// List all dialplans for this tenant
	if err := h.DB.Where("tenant_id = ?", tenantID).
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).
		Order("dialplan_order ASC").Find(&dialplans).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve dial plans"})
		return
	}
	ctx.JSON(dialplans)
}

func (h *Handler) CreateDialPlan(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var dialplan models.Dialplan
	if err := ctx.ReadJSON(&dialplan); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	dialplan.TenantID = &tenantID

	if err := h.DB.Create(&dialplan).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create dial plan"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(dialplan)
}

func (h *Handler) GetDialPlan(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var dialplan models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).
		First(&dialplan).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Dial plan not found"})
		return
	}

	ctx.JSON(dialplan)
}

func (h *Handler) UpdateDialPlan(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var dialplan models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&dialplan).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Dial plan not found"})
		return
	}

	if err := ctx.ReadJSON(&dialplan); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	dialplan.ID = id
	dialplan.TenantID = &tenantID

	if err := h.DB.Save(&dialplan).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update dial plan"})
		return
	}

	// Update details if present
	if len(dialplan.Details) > 0 {
		// Replace existing details
		h.DB.Where("dialplan_uuid = ?", dialplan.UUID).Delete(&models.DialplanDetail{})
		for i := range dialplan.Details {
			dialplan.Details[i].DialplanUUID = dialplan.UUID
		}
		h.DB.Create(&dialplan.Details)
	}

	// Reload with details
	h.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
		return db.Order("detail_order ASC")
	}).First(&dialplan, id)

	ctx.JSON(dialplan)
}

func (h *Handler) DeleteDialPlan(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var dialplan models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&dialplan).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Dial plan not found"})
		return
	}

	// Delete details first
	h.DB.Where("dialplan_uuid = ?", dialplan.UUID).Delete(&models.DialplanDetail{})

	h.DB.Delete(&dialplan)
	ctx.StatusCode(http.StatusNoContent)
}

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
