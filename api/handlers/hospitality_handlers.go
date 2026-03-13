package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// =====================
// Hospitality Handlers
// =====================

// ListRooms returns all hotel rooms for the tenant
func (h *Handler) ListRooms(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var rooms []models.HotelRoom
	query := h.DB.Where("tenant_id = ?", tenantID)

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if floor := c.Query("floor"); floor != "" {
		query = query.Where("floor = ?", floor)
	}

	query.Order("room_number ASC").Find(&rooms)

	return c.JSON(fiber.Map{"data": rooms})
}

// CreateRoom creates a new hotel room
func (h *Handler) CreateRoom(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var room models.HotelRoom
	if err := c.BodyParser(&room); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	room.TenantID = tenantID

	if err := h.DB.Create(&room).Error; err != nil {
		h.logError("HOSPITALITY", "CreateRoom: Failed to create room", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create room"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": room})
}

// GetRoom returns a single hotel room
func (h *Handler) GetRoom(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		h.logWarn("HOSPITALITY", "GetRoom: Room not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	return c.JSON(fiber.Map{"data": room})
}

// UpdateRoom updates a hotel room
func (h *Handler) UpdateRoom(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		h.logWarn("HOSPITALITY", "UpdateRoom: Room not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	if err := c.BodyParser(&room); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	room.TenantID = tenantID
	h.DB.Save(&room)

	return c.JSON(fiber.Map{"data": room, "message": "Room updated"})
}

// DeleteRoom deletes a hotel room
func (h *Handler) DeleteRoom(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.HotelRoom{})
	if result.RowsAffected == 0 {
		h.logWarn("HOSPITALITY", "DeleteRoom: Room not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	return c.JSON(fiber.Map{"message": "Room deleted"})
}

// CheckInGuest checks a guest into a room
func (h *Handler) CheckInGuest(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		h.logWarn("HOSPITALITY", "CheckInGuest: Room not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	var req struct {
		GuestName string `json:"guest_name"`
		CIDName   string `json:"cid_name"`
		CIDNumber string `json:"cid_number"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	now := time.Now()
	room.Status = "occupied"
	room.GuestName = req.GuestName
	room.CheckInTime = &now
	room.CheckOutTime = nil
	room.CIDName = req.CIDName
	room.CIDNumber = req.CIDNumber
	room.DNDEnabled = false
	room.WakeupEnabled = false
	room.WakeupTime = nil

	h.DB.Save(&room)

	return c.JSON(fiber.Map{"data": room, "message": "Guest checked in"})
}

// CheckOutGuest checks a guest out of a room
func (h *Handler) CheckOutGuest(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		h.logWarn("HOSPITALITY", "CheckOutGuest: Room not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	now := time.Now()
	room.Status = "vacant"
	room.GuestName = ""
	room.CheckOutTime = &now
	room.CIDName = ""
	room.CIDNumber = ""
	room.DNDEnabled = false
	room.WakeupEnabled = false
	room.WakeupTime = nil

	h.DB.Save(&room)

	return c.JSON(fiber.Map{"data": room, "message": "Guest checked out"})
}

// ScheduleWakeupCall sets or clears a wakeup call for a room
func (h *Handler) ScheduleWakeupCall(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		h.logWarn("HOSPITALITY", "ScheduleWakeupCall: Room not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	var req struct {
		Enabled    bool       `json:"enabled"`
		WakeupTime *time.Time `json:"wakeup_time"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	room.WakeupEnabled = req.Enabled
	room.WakeupTime = req.WakeupTime

	h.DB.Save(&room)

	return c.JSON(fiber.Map{"data": room, "message": "Wakeup call updated"})
}
