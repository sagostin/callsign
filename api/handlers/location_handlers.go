package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// =====================
// E911 Location Management
// =====================

// ListLocations returns all E911 locations for the current tenant
func (h *Handler) ListLocations(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var locations []models.Location
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("is_default DESC, name ASC").Find(&locations).Error; err != nil {
		h.logError("LOCATION", "ListLocations: Failed to retrieve locations", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve locations"})
	}

	return c.JSON(locations)
}

// CreateLocation creates a new E911 location
func (h *Handler) CreateLocation(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var location models.Location
	if err := c.BodyParser(&location); err != nil {
		h.logWarn("LOCATION", "CreateLocation: Invalid request payload", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	location.TenantID = tenantID

	// If this is the first location or marked as default, handle default logic
	if location.IsDefault {
		h.DB.Model(&models.Location{}).Where("tenant_id = ?", tenantID).Update("is_default", false)
	}

	if err := h.DB.Create(&location).Error; err != nil {
		h.logError("LOCATION", "CreateLocation: Failed to create location", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create location"})
	}

	return c.Status(http.StatusCreated).JSON(location)
}

// GetLocation returns a specific E911 location
func (h *Handler) GetLocation(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		h.logWarn("LOCATION", "GetLocation: Invalid location ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid location ID"})
	}

	var location models.Location
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&location).Error; err != nil {
		h.logWarn("LOCATION", "GetLocation: Location not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Location not found"})
	}

	return c.JSON(location)
}

// UpdateLocation updates an existing E911 location
func (h *Handler) UpdateLocation(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		h.logWarn("LOCATION", "UpdateLocation: Invalid location ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid location ID"})
	}

	var existing models.Location
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&existing).Error; err != nil {
		h.logWarn("LOCATION", "UpdateLocation: Location not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Location not found"})
	}

	var updates models.Location
	if err := c.BodyParser(&updates); err != nil {
		h.logWarn("LOCATION", "UpdateLocation: Invalid request payload", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Handle default flag
	if updates.IsDefault && !existing.IsDefault {
		h.DB.Model(&models.Location{}).Where("tenant_id = ? AND id != ?", tenantID, id).Update("is_default", false)
	}

	// Apply updates
	updates.ID = existing.ID
	updates.TenantID = tenantID
	if err := h.DB.Model(&existing).Updates(updates).Error; err != nil {
		h.logError("LOCATION", "UpdateLocation: Failed to update location", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update location"})
	}

	h.DB.First(&existing, id)
	return c.JSON(existing)
}

// DeleteLocation deletes an E911 location
func (h *Handler) DeleteLocation(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		h.logWarn("LOCATION", "DeleteLocation: Invalid location ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid location ID"})
	}

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Location{})
	if result.RowsAffected == 0 {
		h.logWarn("LOCATION", "DeleteLocation: Location not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Location not found"})
	}

	return c.JSON(fiber.Map{"message": "Location deleted"})
}
