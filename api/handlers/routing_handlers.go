package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ReorderItem represents a single item in a reorder request
type ReorderItem struct {
	ID    uint `json:"id"`
	Order int  `json:"order"`
}

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

func (h *Handler) ListInboundRoutes(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var routes []models.Dialplan
	// Inbound routes typically live in the 'public' context
	if err := h.DB.Where("tenant_id = ? AND dialplan_context = ?", tenantID, "public").
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).
		Order("dialplan_order ASC").Find(&routes).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve inbound routes"})
	}

	return c.JSON(fiber.Map{"data": routes})
}

func (h *Handler) CreateInboundRoute(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var route models.Dialplan
	if err := c.BodyParser(&route); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	route.TenantID = &tenantID
	if route.DialplanContext == "" {
		route.DialplanContext = "public"
	}

	if err := h.DB.Create(&route).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create inbound route"})
	}

	// Reload FreeSWITCH so new inbound route is active
	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(route)
}

func (h *Handler) GetInboundRoute(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var route models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ? AND dialplan_context = ?", id, tenantID, "public").
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).First(&route).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Inbound route not found"})
	}

	return c.JSON(fiber.Map{"data": route})
}

func (h *Handler) UpdateInboundRoute(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var route models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ? AND dialplan_context = ?", id, tenantID, "public").First(&route).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Inbound route not found"})
	}

	if err := c.BodyParser(&route); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	route.ID = uint(id)
	route.TenantID = &tenantID
	route.DialplanContext = "public"

	if err := h.DB.Save(&route).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inbound route"})
	}

	// Replace details if present
	if len(route.Details) > 0 {
		h.DB.Where("dialplan_uuid = ?", route.UUID).Delete(&models.DialplanDetail{})
		for i := range route.Details {
			route.Details[i].DialplanUUID = route.UUID
		}
		h.DB.Create(&route.Details)
	}

	// Reload with details
	h.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
		return db.Order("detail_order ASC")
	}).First(&route, id)

	h.reloadXML()
	return c.JSON(fiber.Map{"data": route, "message": "Inbound route updated"})
}

func (h *Handler) DeleteInboundRoute(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var route models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ? AND dialplan_context = ?", id, tenantID, "public").First(&route).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Inbound route not found"})
	}

	// Delete details first
	h.DB.Where("dialplan_uuid = ?", route.UUID).Delete(&models.DialplanDetail{})
	h.DB.Delete(&route)
	h.reloadXML()
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) ReorderInboundRoutes(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var items []ReorderItem
	if err := c.BodyParser(&items); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tx := h.DB.Begin()
	for _, item := range items {
		if err := tx.Model(&models.Dialplan{}).
			Where("id = ? AND tenant_id = ? AND dialplan_context = ?", item.ID, tenantID, "public").
			Update("dialplan_order", item.Order).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reorder routes"})
		}
	}
	tx.Commit()

	return c.JSON(fiber.Map{"message": "Inbound routes reordered"})
}

// =====================
// Outbound Routes (Dialplan context=default/domain)
// =====================

func (h *Handler) ListOutboundRoutes(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var routes []models.Dialplan
	// Outbound routes live in the 'default' context (or the tenant's domain context)
	if err := h.DB.Where("tenant_id = ? AND dialplan_context != ?", tenantID, "public").
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).
		Order("dialplan_order ASC").Find(&routes).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve outbound routes"})
	}

	return c.JSON(fiber.Map{"data": routes})
}

func (h *Handler) CreateOutboundRoute(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var route models.Dialplan
	if err := c.BodyParser(&route); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	route.TenantID = &tenantID
	if route.DialplanContext == "" {
		route.DialplanContext = "default"
	}

	if err := h.DB.Create(&route).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create outbound route"})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(route)
}

func (h *Handler) GetOutboundRoute(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var route models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ? AND dialplan_context != ?", id, tenantID, "public").
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).First(&route).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Outbound route not found"})
	}

	return c.JSON(fiber.Map{"data": route})
}

func (h *Handler) UpdateOutboundRoute(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var route models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ? AND dialplan_context != ?", id, tenantID, "public").First(&route).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Outbound route not found"})
	}

	originalContext := route.DialplanContext

	if err := c.BodyParser(&route); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	route.ID = uint(id)
	route.TenantID = &tenantID
	// Preserve context — don't allow switching to "public"
	if route.DialplanContext == "" || route.DialplanContext == "public" {
		route.DialplanContext = originalContext
	}

	if err := h.DB.Save(&route).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update outbound route"})
	}

	// Replace details if present
	if len(route.Details) > 0 {
		h.DB.Where("dialplan_uuid = ?", route.UUID).Delete(&models.DialplanDetail{})
		for i := range route.Details {
			route.Details[i].DialplanUUID = route.UUID
		}
		h.DB.Create(&route.Details)
	}

	// Reload with details
	h.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
		return db.Order("detail_order ASC")
	}).First(&route, id)

	h.reloadXML()
	return c.JSON(fiber.Map{"data": route, "message": "Outbound route updated"})
}

func (h *Handler) DeleteOutboundRoute(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var route models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ? AND dialplan_context != ?", id, tenantID, "public").First(&route).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Outbound route not found"})
	}

	// Delete details first
	h.DB.Where("dialplan_uuid = ?", route.UUID).Delete(&models.DialplanDetail{})
	h.DB.Delete(&route)
	h.reloadXML()
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) ReorderOutboundRoutes(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var items []ReorderItem
	if err := c.BodyParser(&items); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tx := h.DB.Begin()
	for _, item := range items {
		if err := tx.Model(&models.Dialplan{}).
			Where("id = ? AND tenant_id = ? AND dialplan_context != ?", item.ID, tenantID, "public").
			Update("dialplan_order", item.Order).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reorder routes"})
		}
	}
	tx.Commit()

	return c.JSON(fiber.Map{"message": "Outbound routes reordered"})
}

// =====================
// Generic Dial Plans (CRUD)
// =====================

func (h *Handler) ListDialPlans(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var dialplans []models.Dialplan
	// List all dialplans for this tenant
	if err := h.DB.Where("tenant_id = ?", tenantID).
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).
		Order("dialplan_order ASC").Find(&dialplans).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve dial plans"})
	}
	return c.JSON(dialplans)
}

func (h *Handler) CreateDialPlan(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var dialplan models.Dialplan
	if err := c.BodyParser(&dialplan); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	dialplan.TenantID = &tenantID

	if err := h.DB.Create(&dialplan).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create dial plan"})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(dialplan)
}

func (h *Handler) GetDialPlan(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var dialplan models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("detail_order ASC")
		}).
		First(&dialplan).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Dial plan not found"})
	}

	return c.JSON(dialplan)
}

func (h *Handler) UpdateDialPlan(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var dialplan models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&dialplan).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Dial plan not found"})
	}

	if err := c.BodyParser(&dialplan); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	dialplan.ID = uint(id)
	dialplan.TenantID = &tenantID

	if err := h.DB.Save(&dialplan).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update dial plan"})
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

	h.reloadXML()
	return c.JSON(dialplan)
}

func (h *Handler) DeleteDialPlan(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var dialplan models.Dialplan
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&dialplan).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Dial plan not found"})
	}

	// Delete details first
	h.DB.Where("dialplan_uuid = ?", dialplan.UUID).Delete(&models.DialplanDetail{})

	h.DB.Delete(&dialplan)
	h.reloadXML()
	c.Status(http.StatusNoContent)
	return nil
}

// =====================
// Feature Codes
// =====================

func (h *Handler) ListFeatureCodes(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var codes []models.FeatureCode
	h.DB.Where("tenant_id = ?", tenantID).Order("code").Find(&codes)

	return c.JSON(codes)
}

func (h *Handler) ListSystemFeatureCodes(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	modules := models.ListAvailableModules(h.DB, tenantID)
	return c.JSON(fiber.Map{"data": modules})
}

func (h *Handler) CreateFeatureCode(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var code models.FeatureCode
	if err := c.BodyParser(&code); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	code.TenantID = &tenantID

	if err := code.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.DB.Create(&code).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(code)
}

func (h *Handler) GetFeatureCode(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var code models.FeatureCode
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&code).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Feature code not found"})
	}

	return c.JSON(code)
}

func (h *Handler) UpdateFeatureCode(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var code models.FeatureCode
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&code).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Feature code not found"})
	}

	if err := c.BodyParser(&code); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	code.TenantID = &tenantID
	h.DB.Save(&code)
	h.reloadXML()
	return c.JSON(code)
}

func (h *Handler) DeleteFeatureCode(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var code models.FeatureCode
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&code).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Feature code not found"})
	}

	h.DB.Delete(&code)
	h.reloadXML()
	c.Status(http.StatusNoContent)
	return nil
}

// =====================
// Time Conditions
// =====================

func (h *Handler) ListTimeConditions(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var conditions []models.TimeCondition
	h.DB.Where("tenant_id = ?", tenantID).Order("name").Find(&conditions)

	return c.JSON(conditions)
}

func (h *Handler) CreateTimeCondition(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var condition models.TimeCondition
	if err := c.BodyParser(&condition); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	condition.TenantID = tenantID

	if err := h.DB.Create(&condition).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(condition)
}

func (h *Handler) GetTimeCondition(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var condition models.TimeCondition
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&condition).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Time condition not found"})
	}

	return c.JSON(condition)
}

func (h *Handler) UpdateTimeCondition(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var condition models.TimeCondition
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&condition).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Time condition not found"})
	}

	if err := c.BodyParser(&condition); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	condition.TenantID = tenantID
	h.DB.Save(&condition)
	h.reloadXML()
	return c.JSON(condition)
}

func (h *Handler) DeleteTimeCondition(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var condition models.TimeCondition
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&condition).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Time condition not found"})
	}

	h.DB.Delete(&condition)
	h.reloadXML()
	c.Status(http.StatusNoContent)
	return nil
}

// =====================
// Holiday Lists
// =====================

func (h *Handler) ListHolidayLists(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	var lists []models.HolidayList
	h.DB.Where("tenant_id = ?", tenantID).Order("name ASC").Find(&lists)
	return c.JSON(fiber.Map{"data": lists})
}

func (h *Handler) CreateHolidayList(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var list models.HolidayList
	if err := c.BodyParser(&list); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	list.TenantID = tenantID
	if err := h.DB.Create(&list).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create holiday list"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": list})
}

func (h *Handler) GetHolidayList(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var list models.HolidayList
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&list).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Holiday list not found"})
	}

	return c.JSON(fiber.Map{"data": list})
}

func (h *Handler) UpdateHolidayList(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var list models.HolidayList
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&list).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Holiday list not found"})
	}

	if err := c.BodyParser(&list); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	list.TenantID = tenantID
	h.DB.Save(&list)
	return c.JSON(fiber.Map{"data": list})
}

func (h *Handler) DeleteHolidayList(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var list models.HolidayList
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&list).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Holiday list not found"})
	}

	h.DB.Delete(&list)
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) SyncHolidayList(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var list models.HolidayList
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&list).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Holiday list not found"})
	}

	if list.ExternalURL == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "No external URL configured for this holiday list"})
	}

	// TODO: Fetch and parse ICS from external URL
	// For now, just update the last synced time
	now := time.Now()
	list.LastSynced = &now
	h.DB.Save(&list)

	return c.JSON(fiber.Map{"message": "Holiday list synced", "data": list})
}

// =====================
// Call Flows
// =====================

func (h *Handler) ListCallFlows(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var flows []models.CallFlow
	h.DB.Where("tenant_id = ?", tenantID).Order("name").Find(&flows)

	return c.JSON(flows)
}

func (h *Handler) CreateCallFlow(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var flow models.CallFlow
	if err := c.BodyParser(&flow); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	flow.TenantID = tenantID

	if err := h.DB.Create(&flow).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(flow)
}

func (h *Handler) GetCallFlow(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var flow models.CallFlow
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&flow).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call flow not found"})
	}

	return c.JSON(flow)
}

func (h *Handler) UpdateCallFlow(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var flow models.CallFlow
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&flow).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call flow not found"})
	}

	if err := c.BodyParser(&flow); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	flow.TenantID = tenantID
	h.DB.Save(&flow)
	h.reloadXML()
	return c.JSON(flow)
}

func (h *Handler) DeleteCallFlow(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var flow models.CallFlow
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&flow).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call flow not found"})
	}

	h.DB.Delete(&flow)
	h.reloadXML()
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) ToggleCallFlow(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var flow models.CallFlow
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&flow).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call flow not found"})
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

	h.reloadXML()

	return c.JSON(fiber.Map{
		"data":        flow,
		"state_index": flow.CurrentState,
		"state_label": stateLabel,
	})
}

// =====================
// Numbers / DIDs
// =====================

func (h *Handler) ListNumbers(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var numbers []models.Destination
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("destination_number").Find(&numbers).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve numbers"})
	}
	return c.JSON(fiber.Map{"data": numbers})
}

func (h *Handler) CreateNumber(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var number models.Destination
	if err := c.BodyParser(&number); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create number"})
	}

	h.reloadXML()
	return c.Status(http.StatusCreated).JSON(number)
}

func (h *Handler) GetNumber(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var number models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&number).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number not found"})
	}

	return c.JSON(fiber.Map{"data": number})
}

func (h *Handler) UpdateNumber(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var number models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&number).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number not found"})
	}

	if err := c.BodyParser(&number); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
	h.reloadXML()
	return c.JSON(number)
}

func (h *Handler) DeleteNumber(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var number models.Destination
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&number).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number not found"})
	}

	h.DB.Delete(&number)
	h.reloadXML()
	c.Status(http.StatusNoContent)
	return nil
}

// =====================
// Default Dial Plans (US/CA)
// =====================

func (h *Handler) CreateDefaultUSCANRoutes(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

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

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Default routes created"})
}

// =====================
// Call Blocks (Blocked Callers)
// =====================

func (h *Handler) ListCallBlocks(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var blocks []models.CallBlock
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&blocks).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve call blocks"})
	}

	return c.JSON(fiber.Map{"data": blocks})
}

func (h *Handler) CreateCallBlock(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var block models.CallBlock
	if err := c.BodyParser(&block); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create call block"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Call block created", "data": block})
}

func (h *Handler) UpdateCallBlock(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var block models.CallBlock
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&block).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call block not found"})
	}

	var input models.CallBlock
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	block.Number = normalizeToE164(input.Number)
	block.MatchType = input.MatchType
	block.Action = input.Action
	block.Enabled = input.Enabled
	block.Notes = input.Notes

	if err := h.DB.Save(&block).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update call block"})
	}

	return c.JSON(fiber.Map{"message": "Call block updated", "data": block})
}

func (h *Handler) DeleteCallBlock(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.CallBlock{})
	if result.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete call block"})
	}

	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call block not found"})
	}

	return c.JSON(fiber.Map{"message": "Call block deleted"})
}
