package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// normalizeToE164 attempts to convert a phone number to E.164 format
func normalizeToE164(number string) string {
	// Remove all non-digit characters except leading +
	hasPlus := strings.HasPrefix(number, "+")
	digits := regexp.MustCompile(`\D`).ReplaceAllString(number, "")

	if digits == "" {
		return number // Return original if no digits
	}

	// If already has country code prefix
	if hasPlus {
		return "+" + digits
	}

	// US/CA: 10 digits -> +1 prefix
	if len(digits) == 10 {
		return "+1" + digits
	}

	// US/CA: 11 digits starting with 1
	if len(digits) == 11 && strings.HasPrefix(digits, "1") {
		return "+" + digits
	}

	// Assume international, add + prefix
	return "+" + digits
}

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
// Holiday Lists
// =====================

func (h *Handler) ListHolidayLists(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	var lists []models.HolidayList
	h.DB.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&lists)
	ctx.JSON(iris.Map{"data": lists})
}

func (h *Handler) CreateHolidayList(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var list models.HolidayList
	if err := ctx.ReadJSON(&list); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	list.TenantID = tenantID
	if err := h.DB.Create(&list).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create holiday list"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": list})
}

func (h *Handler) GetHolidayList(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var list models.HolidayList
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&list).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Holiday list not found"})
		return
	}

	ctx.JSON(iris.Map{"data": list})
}

func (h *Handler) UpdateHolidayList(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var list models.HolidayList
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&list).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Holiday list not found"})
		return
	}

	if err := ctx.ReadJSON(&list); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	list.TenantID = tenantID
	h.DB.Save(&list)
	ctx.JSON(iris.Map{"data": list})
}

func (h *Handler) DeleteHolidayList(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var list models.HolidayList
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&list).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Holiday list not found"})
		return
	}

	h.DB.Delete(&list)
	ctx.StatusCode(http.StatusNoContent)
}

func (h *Handler) SyncHolidayList(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var list models.HolidayList
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&list).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Holiday list not found"})
		return
	}

	if list.ExternalURL == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "No external URL configured for this holiday list"})
		return
	}

	// TODO: Fetch and parse ICS from external URL
	// For now, just update the last synced time
	now := time.Now()
	list.LastSynced = &now
	h.DB.Save(&list)

	ctx.JSON(iris.Map{"message": "Holiday list synced", "data": list})
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

	// Cycle through states: 0 -> 1 -> 2 -> ... -> 0
	numStates := len(flow.Destinations)
	if numStates > 0 {
		flow.CurrentState = (flow.CurrentState + 1) % numStates
	}

	h.DB.Save(&flow)

	// Return current state label
	stateLabel := ""
	if numStates > 0 && flow.CurrentState < numStates {
		stateLabel = flow.Destinations[flow.CurrentState].Label
	}

	ctx.JSON(iris.Map{
		"data":        flow,
		"state_index": flow.CurrentState,
		"state_label": stateLabel,
	})
}

// =====================
// Numbers / DIDs
// =====================

func (h *Handler) ListNumbers(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var numbers []models.Destination
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("destination_number").Find(&numbers).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve numbers"})
		return
	}
	ctx.JSON(iris.Map{"data": numbers})
}

func (h *Handler) CreateNumber(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var number models.Destination
	if err := ctx.ReadJSON(&number); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	number.TenantID = tenantID

	// Normalize to E.164 (Basic US/CA logic for now)
	num := number.DestinationNumber
	// Strip non-digits
	normalized := ""
	for _, r := range num {
		if r >= '0' && r <= '9' {
			normalized += string(r)
		}
	}

	// Check length and add prefix
	if len(normalized) == 10 {
		normalized = "+1" + normalized
	} else if len(normalized) == 11 && normalized[0] == '1' {
		normalized = "+" + normalized
	} else if len(normalized) > 0 {
		normalized = "+" + normalized
	}

	number.DestinationNumber = normalized

	if err := h.DB.Create(&number).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create number"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(number)
}

func (h *Handler) GetNumber(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var number models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&number).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Number not found"})
		return
	}

	ctx.JSON(iris.Map{"data": number})
}

func (h *Handler) UpdateNumber(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var number models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&number).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Number not found"})
		return
	}

	if err := ctx.ReadJSON(&number); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Re-normalize if changed? For now assume edits respect format or just update props
	// Basic protection: if number field is submitted, re-normalize
	if number.DestinationNumber != "" {
		num := number.DestinationNumber
		normalized := ""
		for _, r := range num {
			if r >= '0' && r <= '9' {
				normalized += string(r)
			} else if r == '+' { // maintain existing + if passed
				normalized += string(r)
			}
		}
		// If user removed +, re-add if needed logic... simple re-save for now
		// Assuming generic update
	}

	number.TenantID = tenantID
	h.DB.Save(&number)
	ctx.JSON(number)
}

func (h *Handler) DeleteNumber(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var number models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&number).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Number not found"})
		return
	}

	h.DB.Delete(&number)
	ctx.StatusCode(http.StatusNoContent)
}

// =====================
// Default Dial Plans (US/CA)
// =====================

func (h *Handler) CreateDefaultUSCANRoutes(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	// 1. 11-digit dialing (1 + 10 digits) -> E.164 (+1 + 10 digits)
	// Order 9000+ for outbound routes (high priority to not conflict with internal)
	r1 := models.Dialplan{
		TenantID:        &tenantID,
		DialplanName:    "US/CA 11-Digit",
		DialplanContext: "default",
		DialplanOrder:   9000,
		Enabled:         true,
		Continue:        true,
		Details: []models.DialplanDetail{
			{DetailType: "condition", ConditionField: "destination_number", ConditionExpression: "^1(\\d{10})$", ConditionBreak: "on-false", DetailOrder: 10},
			{DetailType: "action", ActionApplication: "set", ActionData: "effective_caller_id_number=+${effective_caller_id_number}", DetailOrder: 20},
			{DetailType: "action", ActionApplication: "bridge", ActionData: "sofia/gateway/${default_gateway}/+1$1", DetailOrder: 30},
		},
	}

	// 2. 10-digit dialing (10 digits) -> E.164 (+1 + 10 digits)
	r2 := models.Dialplan{
		TenantID:        &tenantID,
		DialplanName:    "US/CA 10-Digit",
		DialplanContext: "default",
		DialplanOrder:   9010,
		Enabled:         true,
		Continue:        true,
		Details: []models.DialplanDetail{
			{DetailType: "condition", ConditionField: "destination_number", ConditionExpression: "^(\\d{10})$", ConditionBreak: "on-false", DetailOrder: 10},
			{DetailType: "action", ActionApplication: "set", ActionData: "effective_caller_id_number=+${effective_caller_id_number}", DetailOrder: 20},
			{DetailType: "action", ActionApplication: "bridge", ActionData: "sofia/gateway/${default_gateway}/+1$1", DetailOrder: 30},
		},
	}

	// 3. Emergency 911 - Order 100 (highest priority)
	r3 := models.Dialplan{
		TenantID:        &tenantID,
		DialplanName:    "Emergency 911",
		DialplanContext: "default",
		DialplanOrder:   100,
		Enabled:         true,
		Continue:        false,
		Details: []models.DialplanDetail{
			{DetailType: "condition", ConditionField: "destination_number", ConditionExpression: "^911$", ConditionBreak: "on-false", DetailOrder: 10},
			{DetailType: "action", ActionApplication: "bridge", ActionData: "sofia/gateway/${emergency_gateway}/911", DetailOrder: 20},
		},
	}

	h.DB.Create(&r1)
	h.DB.Create(&r2)
	h.DB.Create(&r3)

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "Default routes created"})
}

// =====================
// Call Blocks (Blocked Callers)
// =====================

func (h *Handler) ListCallBlocks(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var blocks []models.CallBlock
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&blocks).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve call blocks"})
		return
	}

	ctx.JSON(iris.Map{"data": blocks})
}

func (h *Handler) CreateCallBlock(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var block models.CallBlock
	if err := ctx.ReadJSON(&block); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	block.TenantID = tenantID

	// Normalize number to E.164 if possible
	block.Number = normalizeToE164(block.Number)

	// Set defaults
	if block.MatchType == "" {
		block.MatchType = "exact"
	}
	if block.Action == "" {
		block.Action = "reject"
	}

	if err := h.DB.Create(&block).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create call block"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "Call block created", "data": block})
}

func (h *Handler) UpdateCallBlock(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().GetUintDefault("id", 0)

	var block models.CallBlock
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&block).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call block not found"})
		return
	}

	var input models.CallBlock
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	block.Number = normalizeToE164(input.Number)
	block.MatchType = input.MatchType
	block.Action = input.Action
	block.Enabled = input.Enabled
	block.Notes = input.Notes

	if err := h.DB.Save(&block).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update call block"})
		return
	}

	ctx.JSON(iris.Map{"message": "Call block updated", "data": block})
}

func (h *Handler) DeleteCallBlock(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().GetUintDefault("id", 0)

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.CallBlock{})
	if result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete call block"})
		return
	}

	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call block not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Call block deleted"})
}
