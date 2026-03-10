package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"time"

	"github.com/kataras/iris/v12"
)

// =====================
// Hospitality Handlers
// =====================

// ListRooms returns all hotel rooms for the tenant
func (h *Handler) ListRooms(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var rooms []models.HotelRoom
	query := h.DB.Where("tenant_id = ?", tenantID)

	if status := ctx.URLParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if floor := ctx.URLParam("floor"); floor != "" {
		query = query.Where("floor = ?", floor)
	}

	query.Order("room_number ASC").Find(&rooms)

	ctx.JSON(iris.Map{"data": rooms})
}

// CreateRoom creates a new hotel room
func (h *Handler) CreateRoom(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var room models.HotelRoom
	if err := ctx.ReadJSON(&room); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	room.TenantID = tenantID

	if err := h.DB.Create(&room).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create room"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": room})
}

// GetRoom returns a single hotel room
func (h *Handler) GetRoom(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Room not found"})
		return
	}

	ctx.JSON(iris.Map{"data": room})
}

// UpdateRoom updates a hotel room
func (h *Handler) UpdateRoom(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Room not found"})
		return
	}

	if err := ctx.ReadJSON(&room); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	room.TenantID = tenantID
	h.DB.Save(&room)

	ctx.JSON(iris.Map{"data": room, "message": "Room updated"})
}

// DeleteRoom deletes a hotel room
func (h *Handler) DeleteRoom(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.HotelRoom{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Room not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Room deleted"})
}

// CheckInGuest checks a guest into a room
func (h *Handler) CheckInGuest(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Room not found"})
		return
	}

	var req struct {
		GuestName string `json:"guest_name"`
		CIDName   string `json:"cid_name"`
		CIDNumber string `json:"cid_number"`
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
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

	ctx.JSON(iris.Map{"data": room, "message": "Guest checked in"})
}

// CheckOutGuest checks a guest out of a room
func (h *Handler) CheckOutGuest(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Room not found"})
		return
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

	ctx.JSON(iris.Map{"data": room, "message": "Guest checked out"})
}

// ScheduleWakeupCall sets or clears a wakeup call for a room
func (h *Handler) ScheduleWakeupCall(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var room models.HotelRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&room).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Room not found"})
		return
	}

	var req struct {
		Enabled    bool       `json:"enabled"`
		WakeupTime *time.Time `json:"wakeup_time"`
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	room.WakeupEnabled = req.Enabled
	room.WakeupTime = req.WakeupTime

	h.DB.Save(&room)

	ctx.JSON(iris.Map{"data": room, "message": "Wakeup call updated"})
}
