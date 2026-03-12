package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gofiber/fiber/v2"
)

// =====================
// Paging Groups
// =====================

func (h *Handler) ListPageGroups(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var groups []models.PageGroup
	h.DB.Where("tenant_id = ?", tenantID).Preload("Destinations").Order("name").Find(&groups)

	return c.JSON(groups)
}

func (h *Handler) CreatePageGroup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var group models.PageGroup
	if err := c.BodyParser(&group); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	group.TenantID = tenantID

	if err := h.DB.Create(&group).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(group)
}

func (h *Handler) GetPageGroup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var group models.PageGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Destinations").First(&group).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Page group not found"})
	}

	return c.JSON(group)
}

func (h *Handler) UpdatePageGroup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var group models.PageGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&group).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Page group not found"})
	}

	if err := c.BodyParser(&group); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	group.TenantID = tenantID
	h.DB.Save(&group)
	return c.JSON(group)
}

func (h *Handler) DeletePageGroup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var group models.PageGroup
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&group).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Page group not found"})
	}

	h.DB.Delete(&group)
	c.Status(http.StatusNoContent)
	return nil
}

// =====================
// Provisioning Templates
// =====================

func (h *Handler) ListProvisioningTemplates(c *fiber.Ctx) error {
	var templates []models.ProvisioningTemplate

	// System templates (tenant_id is null) + tenant-specific
	tenantID := middleware.GetTenantID(c)
	h.DB.Where("tenant_id IS NULL OR tenant_id = ?", tenantID).Order("vendor, priority").Find(&templates)

	return c.JSON(templates)
}

func (h *Handler) CreateProvisioningTemplate(c *fiber.Ctx) error {
	var tmpl models.ProvisioningTemplate
	if err := c.BodyParser(&tmpl); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// System admin can create system templates (nil tenant_id)
	claims := middleware.GetClaims(c)
	if claims != nil && claims.Role != "system_admin" {
		tenantID := middleware.GetTenantID(c)
		tmpl.TenantID = &tenantID
	}

	if err := h.DB.Create(&tmpl).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(tmpl)
}

func (h *Handler) GetProvisioningTemplate(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	var tmpl models.ProvisioningTemplate
	if err := h.DB.Where("id = ? AND (tenant_id IS NULL OR tenant_id = ?)", id, tenantID).
		First(&tmpl).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Template not found"})
	}

	return c.JSON(tmpl)
}

func (h *Handler) UpdateProvisioningTemplate(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	var tmpl models.ProvisioningTemplate
	if err := h.DB.Where("id = ? AND (tenant_id IS NULL OR tenant_id = ?)", id, tenantID).
		First(&tmpl).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Template not found"})
	}

	if err := c.BodyParser(&tmpl); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	h.DB.Save(&tmpl)
	return c.JSON(tmpl)
}

func (h *Handler) DeleteProvisioningTemplate(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	tenantID := middleware.GetTenantID(c)

	var tmpl models.ProvisioningTemplate
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&tmpl).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Template not found"})
	}

	h.DB.Delete(&tmpl)
	c.Status(http.StatusNoContent)
	return nil
}

// ServeProvisioningConfig handles device config requests
// GET /provisioning/{mac}/{filename}
func (h *Handler) ServeProvisioningConfig(c *fiber.Ctx) error {
	mac := c.Params("mac")
	filename := c.Params("filename")

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
		c.Status(http.StatusNotFound)
		return c.SendString("Device not found")
	}

	// Find matching template
	var tmpl models.ProvisioningTemplate
	err = h.DB.Where("(tenant_id IS NULL OR tenant_id = ?) AND vendor = ? AND enabled = true",
		device.TenantID, device.Vendor).
		Order("priority ASC").
		First(&tmpl).Error

	if err != nil {
		c.Status(http.StatusNotFound)
		return c.SendString("No template found for this device")
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
		c.Status(http.StatusInternalServerError)
		return c.SendString("Template parse error: " + err.Error())
	}

	var output strings.Builder
	if err := t.Execute(&output, vars); err != nil {
		c.Status(http.StatusInternalServerError)
		return c.SendString("Template execution error: " + err.Error())
	}

	// Set content type based on file type
	switch tmpl.FileType {
	case "xml":
		c.Set("Content-Type", "application/xml")
	case "cfg":
		c.Set("Content-Type", "text/plain")
	default:
		c.Set("Content-Type", "application/octet-stream")
	}

	return c.SendString(output.String())
}

// =====================
// Device Call Control
// =====================

func (h *Handler) DeviceHangup(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	mac := c.Params("mac")

	// Find device's current call UUID
	deviceCallUUID, err := h.getDeviceActiveCallUUID(tenantID, mac)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "No active call found"})
	}

	// Use ESL to kill the call
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		h.ESLManager.API(fmt.Sprintf("uuid_kill %s", deviceCallUUID))
	}

	return c.JSON(fiber.Map{
		"message":   "Hangup command sent",
		"call_uuid": deviceCallUUID,
	})
}

func (h *Handler) DeviceTransfer(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	mac := c.Params("mac")

	var req struct {
		Destination string `json:"destination"`
		Type        string `json:"type"` // "blind" or "attended"
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	deviceCallUUID, err := h.getDeviceActiveCallUUID(tenantID, mac)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "No active call found"})
	}

	// Use ESL to transfer the call
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		if req.Type == "attended" {
			h.ESLManager.API(fmt.Sprintf("uuid_transfer %s -both %s XML default", deviceCallUUID, req.Destination))
		} else {
			h.ESLManager.API(fmt.Sprintf("uuid_transfer %s %s XML default", deviceCallUUID, req.Destination))
		}
	}

	return c.JSON(fiber.Map{
		"message":     "Transfer command sent",
		"call_uuid":   deviceCallUUID,
		"destination": req.Destination,
	})
}

func (h *Handler) DeviceHold(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	mac := c.Params("mac")

	var req struct {
		Hold bool `json:"hold"`
	}
	c.BodyParser(&req)

	deviceCallUUID, err := h.getDeviceActiveCallUUID(tenantID, mac)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "No active call found"})
	}

	// Use ESL to hold/unhold the call
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		if req.Hold {
			h.ESLManager.API(fmt.Sprintf("uuid_hold %s", deviceCallUUID))
		} else {
			h.ESLManager.API(fmt.Sprintf("uuid_hold off %s", deviceCallUUID))
		}
	}

	return c.JSON(fiber.Map{
		"message":   "Hold command sent",
		"call_uuid": deviceCallUUID,
		"hold":      req.Hold,
	})
}

func (h *Handler) DeviceDial(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	mac := c.Params("mac")

	var req struct {
		Number string `json:"number"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Device not found"})
	}

	// Use ESL to originate the call
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		dialString := fmt.Sprintf("user/%s@%s", device.Extension, device.Domain)
		bridge := fmt.Sprintf("bridge(sofia/external/%s)", req.Number)
		h.ESLManager.Client.Originate(dialString, bridge, "")
	}

	return c.JSON(fiber.Map{
		"message": "Dial command sent",
		"from":    device.Extension,
		"to":      req.Number,
	})
}

func (h *Handler) DeviceCallStatus(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	mac := c.Params("mac")

	deviceCallUUID, err := h.getDeviceActiveCallUUID(tenantID, mac)
	if err != nil {
		return c.JSON(fiber.Map{
			"active": false,
		})
	}

	// Get call details from ESL if available
	var direction, callerID, destination, callState string
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		direction, _ = h.ESLManager.API(fmt.Sprintf("uuid_getvar %s direction", deviceCallUUID))
		callerID, _ = h.ESLManager.API(fmt.Sprintf("uuid_getvar %s caller_id_number", deviceCallUUID))
		destination, _ = h.ESLManager.API(fmt.Sprintf("uuid_getvar %s destination_number", deviceCallUUID))
		callState, _ = h.ESLManager.API(fmt.Sprintf("uuid_getvar %s channel_call_state", deviceCallUUID))
	}

	return c.JSON(fiber.Map{
		"active":      true,
		"call_uuid":   deviceCallUUID,
		"direction":   strings.TrimSpace(direction),
		"caller_id":   strings.TrimSpace(callerID),
		"destination": strings.TrimSpace(destination),
		"call_state":  strings.TrimSpace(callState),
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
