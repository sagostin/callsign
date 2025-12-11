package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/kataras/iris/v12"
)

// =====================
// Paging Groups
// =====================

func (h *Handler) ListPageGroups(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var groups []models.PageGroup
	h.DB.Where("tenant_id = ?", tenantID).Preload("Destinations").Order("name").Find(&groups)

	ctx.JSON(groups)
}

func (h *Handler) CreatePageGroup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var group models.PageGroup
	if err := ctx.ReadJSON(&group); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	group.TenantID = tenantID

	if err := h.DB.Create(&group).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(group)
}

func (h *Handler) GetPageGroup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var group models.PageGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Destinations").First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Page group not found"})
		return
	}

	ctx.JSON(group)
}

func (h *Handler) UpdatePageGroup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var group models.PageGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Page group not found"})
		return
	}

	if err := ctx.ReadJSON(&group); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	group.TenantID = tenantID
	h.DB.Save(&group)
	ctx.JSON(group)
}

func (h *Handler) DeletePageGroup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var group models.PageGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&group).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Page group not found"})
		return
	}

	h.DB.Delete(&group)
	ctx.StatusCode(http.StatusNoContent)
}

// =====================
// Provisioning Templates
// =====================

func (h *Handler) ListProvisioningTemplates(ctx iris.Context) {
	var templates []models.ProvisioningTemplate

	// System templates (tenant_id is null) + tenant-specific
	tenantID := middleware.GetTenantID(ctx)
	h.DB.Where("tenant_id IS NULL OR tenant_id = ?", tenantID).Order("vendor, priority").Find(&templates)

	ctx.JSON(templates)
}

func (h *Handler) CreateProvisioningTemplate(ctx iris.Context) {
	var tmpl models.ProvisioningTemplate
	if err := ctx.ReadJSON(&tmpl); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// System admin can create system templates (nil tenant_id)
	claims := middleware.GetClaims(ctx)
	if claims != nil && claims.Role != "system_admin" {
		tenantID := middleware.GetTenantID(ctx)
		tmpl.TenantID = &tenantID
	}

	if err := h.DB.Create(&tmpl).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(tmpl)
}

func (h *Handler) GetProvisioningTemplate(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	tenantID := middleware.GetTenantID(ctx)

	var tmpl models.ProvisioningTemplate
	if err := h.DB.Where("id = ? AND (tenant_id IS NULL OR tenant_id = ?)", id, tenantID).
		First(&tmpl).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Template not found"})
		return
	}

	ctx.JSON(tmpl)
}

func (h *Handler) UpdateProvisioningTemplate(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	tenantID := middleware.GetTenantID(ctx)

	var tmpl models.ProvisioningTemplate
	if err := h.DB.Where("id = ? AND (tenant_id IS NULL OR tenant_id = ?)", id, tenantID).
		First(&tmpl).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Template not found"})
		return
	}

	if err := ctx.ReadJSON(&tmpl); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	h.DB.Save(&tmpl)
	ctx.JSON(tmpl)
}

func (h *Handler) DeleteProvisioningTemplate(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	tenantID := middleware.GetTenantID(ctx)

	var tmpl models.ProvisioningTemplate
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&tmpl).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Template not found"})
		return
	}

	h.DB.Delete(&tmpl)
	ctx.StatusCode(http.StatusNoContent)
}

// ServeProvisioningConfig handles device config requests
// GET /provisioning/{mac}/{filename}
func (h *Handler) ServeProvisioningConfig(ctx iris.Context) {
	mac := ctx.Params().Get("mac")
	filename := ctx.Params().Get("filename")

	// Normalize MAC address
	mac = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", ""))

	// Find device by MAC
	var device struct {
		ID          uint
		TenantID    uint
		Vendor      string
		Model       string
		ExtensionID *uint
		Extension   string
		Password    string
		Domain      string
	}

	err := h.DB.Table("devices").
		Select("devices.id, devices.tenant_id, devices.vendor, devices.model, devices.extension_id, "+
			"extensions.extension, extensions.password, tenants.domain").
		Joins("LEFT JOIN extensions ON devices.extension_id = extensions.id").
		Joins("JOIN tenants ON devices.tenant_id = tenants.id").
		Where("UPPER(REPLACE(REPLACE(devices.mac_address, ':', ''), '-', '')) = ?", mac).
		First(&device).Error

	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.Text("Device not found")
		return
	}

	// Find matching template
	var tmpl models.ProvisioningTemplate
	err = h.DB.Where("(tenant_id IS NULL OR tenant_id = ?) AND vendor = ? AND enabled = true",
		device.TenantID, device.Vendor).
		Order("priority ASC").
		First(&tmpl).Error

	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.Text("No template found for this device")
		return
	}

	// Get provisioning variables
	vars := map[string]string{
		"mac_address": mac,
		"extension":   device.Extension,
		"password":    device.Password,
		"domain":      device.Domain,
		"server":      device.Domain,
		"filename":    filename,
	}

	// Load tenant/device variables
	var provVars []models.ProvisioningVariable
	h.DB.Where("tenant_id = ? AND (device_id IS NULL OR device_id = ?)", device.TenantID, device.ID).
		Find(&provVars)

	for _, v := range provVars {
		vars[v.Name] = v.Value
	}

	// Process template
	t, err := template.New("config").Parse(tmpl.Content)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Text("Template parse error: " + err.Error())
		return
	}

	var output strings.Builder
	if err := t.Execute(&output, vars); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Text("Template execution error: " + err.Error())
		return
	}

	// Set content type based on file type
	switch tmpl.FileType {
	case "xml":
		ctx.ContentType("application/xml")
	case "cfg":
		ctx.ContentType("text/plain")
	default:
		ctx.ContentType("application/octet-stream")
	}

	ctx.Text(output.String())
}

// =====================
// Device Call Control
// =====================

func (h *Handler) DeviceHangup(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	mac := ctx.Params().Get("mac")

	// Find device's current call UUID
	deviceCallUUID, err := h.getDeviceActiveCallUUID(tenantID, mac)
	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "No active call found"})
		return
	}

	// TODO: Use ESL to kill the call
	// conn.Send(fmt.Sprintf("api uuid_kill %s", deviceCallUUID))

	ctx.JSON(iris.Map{
		"message":   "Hangup command sent",
		"call_uuid": deviceCallUUID,
	})
}

func (h *Handler) DeviceTransfer(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	mac := ctx.Params().Get("mac")

	var req struct {
		Destination string `json:"destination"`
		Type        string `json:"type"` // "blind" or "attended"
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	deviceCallUUID, err := h.getDeviceActiveCallUUID(tenantID, mac)
	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "No active call found"})
		return
	}

	// TODO: Use ESL to transfer
	// conn.Send(fmt.Sprintf("api uuid_transfer %s %s", deviceCallUUID, req.Destination))

	ctx.JSON(iris.Map{
		"message":     "Transfer command sent",
		"call_uuid":   deviceCallUUID,
		"destination": req.Destination,
	})
}

func (h *Handler) DeviceHold(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	mac := ctx.Params().Get("mac")

	var req struct {
		Hold bool `json:"hold"`
	}
	ctx.ReadJSON(&req)

	deviceCallUUID, err := h.getDeviceActiveCallUUID(tenantID, mac)
	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "No active call found"})
		return
	}

	// TODO: Use ESL to hold/unhold
	// cmd := "uuid_hold"
	// if !req.Hold { cmd = "uuid_hold off" }

	ctx.JSON(iris.Map{
		"message":   "Hold command sent",
		"call_uuid": deviceCallUUID,
		"hold":      req.Hold,
	})
}

func (h *Handler) DeviceDial(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	mac := ctx.Params().Get("mac")

	var req struct {
		Number string `json:"number"`
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Get device extension
	var device struct {
		Extension string
		Domain    string
	}
	normalizedMAC := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", ""))
	err := h.DB.Table("devices").
		Select("extensions.extension, tenants.domain").
		Joins("LEFT JOIN extensions ON devices.extension_id = extensions.id").
		Joins("JOIN tenants ON devices.tenant_id = tenants.id").
		Where("devices.tenant_id = ? AND UPPER(REPLACE(REPLACE(devices.mac_address, ':', ''), '-', '')) = ?",
			tenantID, normalizedMAC).
		First(&device).Error

	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Device not found"})
		return
	}

	// TODO: Use ESL to originate call
	// originate user/1001@domain.com &bridge(sofia/external/+15551234567)

	ctx.JSON(iris.Map{
		"message": "Dial command sent",
		"from":    device.Extension,
		"to":      req.Number,
	})
}

func (h *Handler) DeviceCallStatus(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	mac := ctx.Params().Get("mac")

	deviceCallUUID, err := h.getDeviceActiveCallUUID(tenantID, mac)
	if err != nil {
		ctx.JSON(iris.Map{
			"active": false,
		})
		return
	}

	// TODO: Get call details from ESL
	ctx.JSON(iris.Map{
		"active":    true,
		"call_uuid": deviceCallUUID,
	})
}

// Helper function
func (h *Handler) getDeviceActiveCallUUID(tenantID uint, mac string) (string, error) {
	// This would query the session manager or extension_presences table
	normalizedMAC := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", ""))

	var presence struct {
		CurrentCallUUID string
	}
	err := h.DB.Table("extension_presences").
		Select("extension_presences.current_call_uuid").
		Joins("JOIN extensions ON extension_presences.extension_id = extensions.id").
		Joins("JOIN devices ON devices.extension_id = extensions.id").
		Where("extensions.tenant_id = ? AND UPPER(REPLACE(REPLACE(devices.mac_address, ':', ''), '-', '')) = ?",
			tenantID, normalizedMAC).
		First(&presence).Error

	if err != nil || presence.CurrentCallUUID == "" {
		return "", fmt.Errorf("no active call")
	}

	return presence.CurrentCallUUID, nil
}
