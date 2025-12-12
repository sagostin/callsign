package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/kataras/iris/v12"
)

// =====================
// Extensions
// =====================

func (h *Handler) ListExtensions(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var extensions []models.Extension
	query := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx))

	// Search filter
	if search := ctx.URLParam("search"); search != "" {
		query = query.Where("extension LIKE ? OR effective_caller_id_name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&extensions).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch extensions"})
		return
	}

	ctx.JSON(iris.Map{"data": extensions})
}

func (h *Handler) CreateExtension(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var ext models.Extension
	if err := ctx.ReadJSON(&ext); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	ext.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create extension"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": ext, "message": "Extension created"})
}

func (h *Handler) GetExtension(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var ext models.Extension

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension not found"})
		return
	}

	ctx.JSON(iris.Map{"data": ext})
}

func (h *Handler) UpdateExtension(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var ext models.Extension

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension not found"})
		return
	}

	if err := ctx.ReadJSON(&ext); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update extension"})
		return
	}

	ctx.JSON(iris.Map{"data": ext, "message": "Extension updated"})
}

func (h *Handler) DeleteExtension(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.Extension{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete extension"})
		return
	}

	ctx.JSON(iris.Map{"message": "Extension deleted"})
}

func (h *Handler) GetExtensionStatus(ctx iris.Context) {
	// TODO: Integrate with FreeSWITCH ESL for real-time status
	ctx.JSON(iris.Map{"status": "registered", "message": "Real-time status requires ESL integration"})
}

// =====================
// Devices
// =====================

func (h *Handler) ListDevices(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	// For now, return devices from extensions that have device_uuid set
	var extensions []models.Extension
	h.DB.Where("tenant_id = ? AND device_uuid IS NOT NULL AND device_uuid != ''", middleware.GetTenantID(ctx)).Find(&extensions)

	devices := make([]map[string]interface{}, 0)
	for _, ext := range extensions {
		devices = append(devices, map[string]interface{}{
			"device_uuid": ext.DeviceUUID,
			"extension":   ext.Extension,
			"name":        ext.EffectiveCallerIDName,
		})
	}

	ctx.JSON(iris.Map{"data": devices})
}

func (h *Handler) CreateDevice(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var req struct {
		MAC        string `json:"mac"`
		Extension  string `json:"extension"`
		Name       string `json:"name"`
		TemplateID uint   `json:"template_id"`
	}

	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// TODO: Implement device provisioning logic
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "Device created", "mac": req.MAC})
}

func (h *Handler) GetDevice(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"data": map[string]interface{}{"id": id}, "message": "Device details"})
}

func (h *Handler) UpdateDevice(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"message": "Device updated", "id": id})
}

func (h *Handler) DeleteDevice(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"message": "Device deleted", "id": id})
}

// =====================
// Voicemail
// =====================

func (h *Handler) ListVoicemailBoxes(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var boxes []models.VoicemailBox
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Find(&boxes).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch voicemail boxes"})
		return
	}

	ctx.JSON(iris.Map{"data": boxes})
}

func (h *Handler) CreateVoicemailBox(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var box models.VoicemailBox
	if err := ctx.ReadJSON(&box); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	box.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&box).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create voicemail box"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": box, "message": "Voicemail box created"})
}

func (h *Handler) GetVoicemailBox(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var box models.VoicemailBox

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Preload("Messages").First(&box).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Voicemail box not found"})
		return
	}

	ctx.JSON(iris.Map{"data": box})
}

func (h *Handler) UpdateVoicemailBox(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var box models.VoicemailBox

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&box).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Voicemail box not found"})
		return
	}

	if err := ctx.ReadJSON(&box); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&box).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update voicemail box"})
		return
	}

	ctx.JSON(iris.Map{"data": box, "message": "Voicemail box updated"})
}

func (h *Handler) DeleteVoicemailBox(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.VoicemailBox{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete voicemail box"})
		return
	}

	ctx.JSON(iris.Map{"message": "Voicemail box deleted"})
}

// =====================
// Recordings
// =====================

func (h *Handler) ListRecordings(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var recordings []models.Recording
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Order("created_at DESC").Find(&recordings).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch recordings"})
		return
	}

	ctx.JSON(iris.Map{"data": recordings})
}

func (h *Handler) GetRecording(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var recording models.Recording

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&recording).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Recording not found"})
		return
	}

	ctx.JSON(iris.Map{"data": recording})
}

func (h *Handler) DeleteRecording(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.Recording{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete recording"})
		return
	}

	ctx.JSON(iris.Map{"message": "Recording deleted"})
}

// =====================
// IVR Menus
// =====================

func (h *Handler) ListIVRMenus(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var menus []models.IVRMenu
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Preload("Options").Find(&menus).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch IVR menus"})
		return
	}

	ctx.JSON(iris.Map{"data": menus})
}

func (h *Handler) CreateIVRMenu(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var menu models.IVRMenu
	if err := ctx.ReadJSON(&menu); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	menu.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&menu).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create IVR menu"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": menu, "message": "IVR menu created"})
}

func (h *Handler) GetIVRMenu(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var menu models.IVRMenu

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Preload("Options").First(&menu).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "IVR menu not found"})
		return
	}

	ctx.JSON(iris.Map{"data": menu})
}

func (h *Handler) UpdateIVRMenu(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var menu models.IVRMenu

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&menu).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "IVR menu not found"})
		return
	}

	if err := ctx.ReadJSON(&menu); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&menu).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update IVR menu"})
		return
	}

	ctx.JSON(iris.Map{"data": menu, "message": "IVR menu updated"})
}

func (h *Handler) DeleteIVRMenu(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.IVRMenu{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete IVR menu"})
		return
	}

	ctx.JSON(iris.Map{"message": "IVR menu deleted"})
}

// =====================
// Queues
// =====================

func (h *Handler) ListQueues(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var queues []models.Queue
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Preload("Agents").Find(&queues).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch queues"})
		return
	}

	ctx.JSON(iris.Map{"data": queues})
}

func (h *Handler) CreateQueue(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var queue models.Queue
	if err := ctx.ReadJSON(&queue); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	queue.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create queue"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": queue, "message": "Queue created"})
}

func (h *Handler) GetQueue(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var queue models.Queue

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Preload("Agents").First(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Queue not found"})
		return
	}

	ctx.JSON(iris.Map{"data": queue})
}

func (h *Handler) UpdateQueue(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var queue models.Queue

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Queue not found"})
		return
	}

	if err := ctx.ReadJSON(&queue); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&queue).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update queue"})
		return
	}

	ctx.JSON(iris.Map{"data": queue, "message": "Queue updated"})
}

func (h *Handler) DeleteQueue(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.Queue{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete queue"})
		return
	}

	ctx.JSON(iris.Map{"message": "Queue deleted"})
}

// =====================
// Ring Groups
// =====================

func (h *Handler) ListRingGroups(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var groups []models.RingGroup
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Preload("Destinations").Find(&groups).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch ring groups"})
		return
	}

	ctx.JSON(iris.Map{"data": groups})
}

func (h *Handler) CreateRingGroup(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var group models.RingGroup
	if err := ctx.ReadJSON(&group); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	group.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&group).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create ring group"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": group, "message": "Ring group created"})
}

func (h *Handler) GetRingGroup(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var group models.RingGroup

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Preload("Destinations").First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Ring group not found"})
		return
	}

	ctx.JSON(iris.Map{"data": group})
}

func (h *Handler) UpdateRingGroup(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var group models.RingGroup

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Ring group not found"})
		return
	}

	if err := ctx.ReadJSON(&group); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&group).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update ring group"})
		return
	}

	ctx.JSON(iris.Map{"data": group, "message": "Ring group updated"})
}

func (h *Handler) DeleteRingGroup(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.RingGroup{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete ring group"})
		return
	}

	ctx.JSON(iris.Map{"message": "Ring group deleted"})
}

// =====================
// Conferences
// =====================

func (h *Handler) ListConferences(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var conferences []models.Conference
	if err := h.DB.Where("tenant_id = ?", middleware.GetTenantID(ctx)).Find(&conferences).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch conferences"})
		return
	}

	ctx.JSON(iris.Map{"data": conferences})
}

func (h *Handler) CreateConference(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var conf models.Conference
	if err := ctx.ReadJSON(&conf); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	conf.TenantID = middleware.GetTenantID(ctx)

	if err := h.DB.Create(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create conference"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": conf, "message": "Conference created"})
}

func (h *Handler) GetConference(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var conf models.Conference

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conference not found"})
		return
	}

	ctx.JSON(iris.Map{"data": conf})
}

func (h *Handler) UpdateConference(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))
	var conf models.Conference

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).First(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conference not found"})
		return
	}

	if err := ctx.ReadJSON(&conf); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update conference"})
		return
	}

	ctx.JSON(iris.Map{"data": conf, "message": "Conference updated"})
}

func (h *Handler) DeleteConference(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	id, _ := strconv.Atoi(ctx.Params().Get("id"))

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, middleware.GetTenantID(ctx)).Delete(&models.Conference{}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete conference"})
		return
	}

	ctx.JSON(iris.Map{"message": "Conference deleted"})
}

// =====================
// Numbers/DIDs
// =====================

// Numbers handlers moved to routing_handlers.go

// Routing and Dialplan handlers moved to routing_handlers.go

// =====================
// Audio Library
// =====================

func (h *Handler) ListAudioFiles(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var files []models.Recording
	if err := h.DB.Where("tenant_id = ? AND type = ?", middleware.GetTenantID(ctx), "audio").Find(&files).Error; err != nil {
		ctx.JSON(iris.Map{"data": []interface{}{}})
		return
	}

	ctx.JSON(iris.Map{"data": files})
}

func (h *Handler) UploadAudioFile(ctx iris.Context) {
	// TODO: Implement file upload handling
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "Audio file uploaded"})
}

func (h *Handler) GetAudioFile(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"data": map[string]interface{}{"id": id}})
}

func (h *Handler) DeleteAudioFile(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"message": "Audio file deleted", "id": id})
}

// =====================
// Music on Hold
// =====================

func (h *Handler) ListMOHStreams(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	// TODO: Add MOH model or use Recording with type filter
	ctx.JSON(iris.Map{"data": []interface{}{}, "message": "MOH streams"})
}

func (h *Handler) CreateMOHStream(ctx iris.Context) {
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "MOH stream created"})
}

func (h *Handler) GetMOHStream(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"data": map[string]interface{}{"id": id}})
}

func (h *Handler) UpdateMOHStream(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"message": "MOH stream updated", "id": id})
}

func (h *Handler) DeleteMOHStream(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(iris.Map{"message": "MOH stream deleted", "id": id})
}
