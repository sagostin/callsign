package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/kataras/iris/v12"
)

// =====================
// E911 Location Management
// =====================

// ListLocations returns all E911 locations for the current tenant
func (h *Handler) ListLocations(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var locations []models.Location
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("is_default DESC, name ASC").Find(&locations).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve locations"})
		return
	}

	ctx.JSON(locations)
}

// CreateLocation creates a new E911 location
func (h *Handler) CreateLocation(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var location models.Location
	if err := ctx.ReadJSON(&location); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	location.TenantID = tenantID

	// If this is the first location or marked as default, handle default logic
	if location.IsDefault {
		h.DB.Model(&models.Location{}).Where("tenant_id = ?", tenantID).Update("is_default", false)
	}

	if err := h.DB.Create(&location).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create location"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(location)
}

// GetLocation returns a specific E911 location
func (h *Handler) GetLocation(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid location ID"})
		return
	}

	var location models.Location
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&location).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Location not found"})
		return
	}

	ctx.JSON(location)
}

// UpdateLocation updates an existing E911 location
func (h *Handler) UpdateLocation(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid location ID"})
		return
	}

	var existing models.Location
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&existing).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Location not found"})
		return
	}

	var updates models.Location
	if err := ctx.ReadJSON(&updates); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Handle default flag
	if updates.IsDefault && !existing.IsDefault {
		h.DB.Model(&models.Location{}).Where("tenant_id = ? AND id != ?", tenantID, id).Update("is_default", false)
	}

	// Apply updates
	updates.ID = existing.ID
	updates.TenantID = tenantID
	if err := h.DB.Model(&existing).Updates(updates).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update location"})
		return
	}

	h.DB.First(&existing, id)
	ctx.JSON(existing)
}

// DeleteLocation deletes an E911 location
func (h *Handler) DeleteLocation(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid location ID"})
		return
	}

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Location{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Location not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Location deleted"})
}
