package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ==================
// Device Handlers
// ==================

// ListDevices returns all devices for the current tenant
func (h *Handler) ListDevices(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var devices []models.Device
	query := h.DB.Where("tenant_id = ?", tenantID).
		Preload("User").
		Preload("Template").
		Preload("Lines").
		Preload("Lines.Extension")

	// Optional filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if deviceType := c.Query("type"); deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}
	if manufacturer := c.Query("manufacturer"); manufacturer != "" {
		query = query.Where("manufacturer = ?", manufacturer)
	}
	if userID := c.Query("user_id"); userID != "" {
		if userID == "unassigned" {
			query = query.Where("user_id IS NULL")
		} else {
			query = query.Where("user_id = ?", userID)
		}
	}

	query.Order("created_at DESC").Find(&devices)

	return c.JSON(fiber.Map{"data": devices})
}

// CreateDevice creates a new device
func (h *Handler) CreateDevice(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var device models.Device
	if err := c.BodyParser(&device); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Normalize MAC address
	device.MAC = strings.ToLower(models.NormalizeMAC(device.MAC))
	if device.MAC == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "MAC address is required"})
	}

	// Check for duplicate MAC
	var existing models.Device
	if err := h.DB.Where("mac = ?", device.MAC).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Device with this MAC address already exists"})
	}

	device.TenantID = tenantID

	// Set default device type
	if device.DeviceType == "" {
		if device.Manufacturer != "" {
			device.DeviceType = models.DeviceTypeProvisioned
		} else {
			device.DeviceType = models.DeviceTypeGenericSIP
		}
	}

	if err := h.DB.Create(&device).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create device"})
	}

	// Create default Line 1 if extension is provided
	if device.Lines == nil || len(device.Lines) == 0 {
		line := models.DeviceLine{
			DeviceID:   device.ID,
			LineNumber: 1,
			Enabled:    true,
		}
		h.DB.Create(&line)
	}

	return c.JSON(fiber.Map{"data": device, "message": "Device created"})
}

// GetDevice returns a single device by ID or MAC
func (h *Handler) GetDevice(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	idOrMac := c.Params("id")

	var device models.Device
	query := h.DB.Where("tenant_id = ?", tenantID).
		Preload("User").
		Preload("Template").
		Preload("Lines").
		Preload("Lines.Extension")

	// Try by ID first, then by MAC
	if err := query.Where("id = ?", idOrMac).First(&device).Error; err != nil {
		// Try normalized MAC
		normalizedMAC := strings.ToLower(models.NormalizeMAC(idOrMac))
		if err := query.Where("mac = ?", normalizedMAC).First(&device).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Device not found"})
		}
	}

	// Return device data with ProvisionToken exposed
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":                  device.ID,
			"uuid":                device.UUID,
			"created_at":          device.CreatedAt,
			"updated_at":          device.UpdatedAt,
			"tenant_id":           device.TenantID,
			"mac":                 device.MAC,
			"name":                device.Name,
			"device_type":         device.DeviceType,
			"manufacturer":        device.Manufacturer,
			"model":               device.Model,
			"template_id":         device.TemplateID,
			"template":            device.Template,
			"profile_id":          device.ProfileID,
			"profile":             device.Profile,
			"user_id":             device.UserID,
			"user":                device.User,
			"lines":               device.Lines,
			"registration_prefix": device.RegistrationPrefix,
			"registration_user":   device.RegistrationUser,
			"sip_server":          device.SIPServer,
			"sip_proxy":           device.SIPProxy,
			"sip_transport":       device.SIPTransport,
			"sip_port":            device.SIPPort,
			"provision_url":       device.ProvisionURL,
			"provision_token":     device.ProvisionToken, // Exposed for admin
			"last_provision":      device.LastProvision,
			"provision_count":     device.ProvisionCount,
			"status":              device.Status,
			"registered":          device.Registered,
			"registration_ip":     device.RegistrationIP,
			"user_agent":          device.UserAgent,
			"last_seen":           device.LastSeen,
			"location_id":         device.LocationID,
			"enabled":             device.Enabled,
		},
	})
}

// UpdateDevice updates a device
func (h *Handler) UpdateDevice(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var device models.Device
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Device not found"})
	}

	// Parse input with ProvisionToken exposed
	var input struct {
		models.Device
		ProvisionToken string `json:"provision_token"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update allowed fields
	updates := map[string]interface{}{
		"name":               input.Name,
		"device_type":        input.DeviceType,
		"manufacturer":       input.Manufacturer,
		"model":              input.Model,
		"template_id":        input.TemplateID,
		"user_id":            input.UserID,
		"sip_server":         input.SIPServer,
		"sip_proxy":          input.SIPProxy,
		"sip_transport":      input.SIPTransport,
		"sip_port":           input.SIPPort,
		"early_media":        input.EarlyMedia,
		"supported_codecs":   input.SupportedCodecs,
		"t38_enabled":        input.T38Enabled,
		"encryption_enabled": input.EncryptionEnabled,
		"location_id":        input.LocationID,
		"enabled":            input.Enabled,
	}

	// Only update token if provided and non-empty (allow generation)
	if input.ProvisionToken != "" {
		updates["provision_token"] = input.ProvisionToken
	}

	h.DB.Model(&device).Updates(updates)

	// Reload with associations
	h.DB.Preload("User").Preload("Template").Preload("Lines").Preload("Lines.Extension").First(&device)

	return c.JSON(fiber.Map{"data": device, "message": "Device updated"})
}

// DeleteDevice deletes a device
func (h *Handler) DeleteDevice(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	// Delete lines first
	h.DB.Where("device_id = ?", id).Delete(&models.DeviceLine{})

	// Delete device
	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Device{})
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Device not found"})
	}

	return c.JSON(fiber.Map{"message": "Device deleted"})
}

// AssignDeviceToUser assigns a device to a user
func (h *Handler) AssignDeviceToUser(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var input struct {
		UserID *uint `json:"user_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result := h.DB.Model(&models.Device{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("user_id", input.UserID)

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Device not found"})
	}

	return c.JSON(fiber.Map{"message": "Device assigned to user"})
}

// ==================
// Device Line Handlers
// ==================

// UpdateDeviceLines updates lines for a device
func (h *Handler) UpdateDeviceLines(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	deviceID := c.Params("id")

	// Verify device belongs to tenant
	var device models.Device
	if err := h.DB.Where("id = ? AND tenant_id = ?", deviceID, tenantID).First(&device).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Device not found"})
	}

	var lines []models.DeviceLine
	if err := c.BodyParser(&lines); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Delete existing lines
	h.DB.Where("device_id = ?", device.ID).Delete(&models.DeviceLine{})

	// Create new lines
	for i := range lines {
		lines[i].DeviceID = device.ID
		if lines[i].LineNumber == 0 {
			lines[i].LineNumber = i + 1
		}
		h.DB.Create(&lines[i])
	}

	// Reload
	h.DB.Where("device_id = ?", device.ID).Preload("Extension").Find(&lines)

	return c.JSON(fiber.Map{"data": lines, "message": "Device lines updated"})
}

// ==================
// Provisioning Handlers
// ==================

// GetDeviceConfigSecure returns provisioning config with tenant secret verification
// URL: /provision/{tenant_uuid}/{secret}/{mac}.cfg
func (h *Handler) GetDeviceConfigSecure(c *fiber.Ctx) error {
	tenantUUID := c.Params("tenant")
	secret := c.Params("secret")
	mac := c.Params("mac")
	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	// Remove extension if present (.cfg, .xml, etc.)
	if idx := strings.LastIndex(mac, "."); idx > 0 {
		mac = mac[:idx]
	}
	normalizedMAC := strings.ToLower(models.NormalizeMAC(mac))

	// Verify tenant exists and secret matches
	var tenant models.Tenant
	if err := h.DB.Where("uuid = ? AND enabled = ?", tenantUUID, true).First(&tenant).Error; err != nil {
		// Log brute force attempt - invalid tenant
		if h.LogManager != nil {
			h.LogManager.Warn("PROVISION_BRUTE_FORCE", "Provisioning attempt with invalid tenant UUID", map[string]interface{}{
				"reason":      "invalid_tenant",
				"ip":          clientIP,
				"tenant_uuid": tenantUUID,
				"mac":         mac,
				"user_agent":  userAgent,
			})
		}

		c.Status(fiber.StatusNotFound)
		return c.SendString("Invalid provisioning URL")
	}

	// Check if provisioning is enabled for tenant
	if !tenant.ProvisioningEnabled {
		if h.LogManager != nil {
			h.LogManager.Info("PROVISION_DENIED", "Provisioning disabled for tenant", map[string]interface{}{
				"reason": "disabled",
				"ip":     clientIP,
				"tenant": tenant.Domain,
				"mac":    mac,
			})
		}

		c.Status(fiber.StatusForbidden)
		return c.SendString("Provisioning disabled for this tenant")
	}

	// Verify secret
	if tenant.ProvisioningSecret != secret {
		// Log brute force attempt - wrong secret
		if h.LogManager != nil {
			h.LogManager.Warn("PROVISION_BRUTE_FORCE", "Provisioning attempt with invalid secret", map[string]interface{}{
				"reason":     "invalid_secret",
				"ip":         clientIP,
				"tenant":     tenant.Domain,
				"mac":        mac,
				"user_agent": userAgent,
			})
		}

		c.Status(fiber.StatusForbidden)
		return c.SendString("Invalid provisioning URL")
	}

	// Find device
	var device models.Device
	if err := h.DB.Where("mac = ? AND tenant_id = ?", normalizedMAC, tenant.ID).
		Preload("Template").
		Preload("Profile").
		Preload("Lines").
		Preload("Lines.Extension").
		First(&device).Error; err != nil {
		c.Status(fiber.StatusNotFound)
		return c.SendString("Device not found")
	}

	if !device.Enabled {
		c.Status(fiber.StatusForbidden)
		return c.SendString("Device is disabled")
	}

	// Get template (from device, profile, or default for manufacturer)
	var tmpl models.DeviceTemplate
	if device.TemplateID != nil {
		h.DB.First(&tmpl, *device.TemplateID)
	} else if device.Profile != nil && device.Profile.TemplateID != nil {
		h.DB.First(&tmpl, *device.Profile.TemplateID)
	} else {
		// Try to find default template for manufacturer/model
		h.DB.Where("manufacturer = ? AND (model = ? OR model = '' OR model IS NULL) AND enabled = ?",
			device.Manufacturer, device.Model, true).
			Order("model DESC").
			First(&tmpl)
	}

	if tmpl.ID == 0 || tmpl.ConfigTemplate == "" {
		c.Status(fiber.StatusNotFound)
		return c.SendString("No template available for this device")
	}

	// Verify User-Agent if pattern is set
	if tmpl.UserAgentPattern != "" {
		userAgent := c.Get("User-Agent")
		matched, _ := regexp.MatchString(tmpl.UserAgentPattern, userAgent)
		if !matched {
			c.Status(fiber.StatusForbidden)
			return c.SendString("Device not authorized")
		}
	}

	// Verify MAC pattern if set
	if tmpl.MACPattern != "" {
		matched, _ := regexp.MatchString(tmpl.MACPattern, normalizedMAC)
		if !matched {
			c.Status(fiber.StatusForbidden)
			return c.SendString("Device model mismatch")
		}
	}

	return h.renderDeviceConfig(c, &device, &tmpl, &tenant)
}

// GetDeviceConfig returns the provisioning configuration for a device (legacy, no auth)
func (h *Handler) GetDeviceConfig(c *fiber.Ctx) error {
	mac := c.Params("mac")
	normalizedMAC := strings.ToLower(models.NormalizeMAC(mac))

	var device models.Device
	if err := h.DB.Where("mac = ?", normalizedMAC).
		Preload("Template").
		Preload("Profile").
		Preload("Lines").
		Preload("Lines.Extension").
		First(&device).Error; err != nil {
		c.Status(fiber.StatusNotFound)
		return c.SendString("Device not found")
	}

	if !device.Enabled {
		c.Status(fiber.StatusForbidden)
		return c.SendString("Device is disabled")
	}

	// Get template
	var tmpl models.DeviceTemplate
	if device.TemplateID != nil {
		h.DB.First(&tmpl, *device.TemplateID)
	} else {
		// Try to find default template for manufacturer/model
		h.DB.Where("manufacturer = ? AND (model = ? OR model = '' OR model IS NULL) AND enabled = ?",
			device.Manufacturer, device.Model, true).
			Order("model DESC").
			First(&tmpl)
	}

	if tmpl.ID == 0 || tmpl.ConfigTemplate == "" {
		c.Status(fiber.StatusNotFound)
		return c.SendString("No template available for this device")
	}

	// Get tenant for domain
	var tenant models.Tenant
	h.DB.First(&tenant, device.TenantID)

	return h.renderDeviceConfig(c, &device, &tmpl, &tenant)
}

// renderDeviceConfig renders the provisioning template
func (h *Handler) renderDeviceConfig(c *fiber.Ctx, device *models.Device, tmpl *models.DeviceTemplate, tenant *models.Tenant) error {
	// Build provisioning variables
	vars := models.ProvisioningVariables{
		MAC:          device.MAC,
		DeviceName:   device.Name,
		Model:        device.Model,
		Manufacturer: device.Manufacturer,
		Server:       tenant.Domain,
		Domain:       tenant.Domain,
		TenantName:   tenant.Name,
		TenantDomain: tenant.Domain,
		Timestamp:    time.Now(),
		Lines:        make(map[int]models.LineVariables),
	}

	// Add line variables
	for _, line := range device.Lines {
		if !line.Enabled {
			continue
		}

		userID, authUser, password := line.GetEffectiveCredentials()

		displayName := line.Label
		if displayName == "" && line.Extension != nil {
			displayName = line.Extension.EffectiveCallerIDName
			if displayName == "" {
				displayName = line.Extension.Extension
			}
		}

		vars.Lines[line.LineNumber] = models.LineVariables{
			LineNumber:  line.LineNumber,
			Extension:   userID,
			DisplayName: displayName,
			Label:       line.Label,
			UserID:      userID,
			AuthUser:    authUser,
			Password:    password,
			Server:      tenant.Domain,
			Enabled:     line.Enabled,
		}
	}

	// Render template
	t, err := template.New("config").Parse(tmpl.ConfigTemplate)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.SendString("Template parse error: " + err.Error())
	}

	// Set content type based on config type
	switch tmpl.ConfigType {
	case "xml":
		c.Set("Content-Type", "application/xml")
	case "cfg":
		c.Set("Content-Type", "text/plain")
	case "json":
		c.Set("Content-Type", "application/json")
	default:
		c.Set("Content-Type", "text/plain")
	}

	if err := t.Execute(c.Response().BodyWriter(), vars); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.SendString("Template execution error: " + err.Error())
	}

	// Update last provision time
	h.DB.Model(device).Updates(map[string]interface{}{
		"last_provision":  time.Now(),
		"provision_count": device.ProvisionCount + 1,
	})

	return nil
}

// ReprovisionDevice triggers a re-provision (SIP NOTIFY)
func (h *Handler) ReprovisionDevice(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var device models.Device
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Device not found"})
	}

	// Send SIP NOTIFY to trigger re-provision via ESL
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		// Get the device's extension and domain for the flush command
		var ext models.Extension
		var tenant models.Tenant
		if device.Lines != nil && len(device.Lines) > 0 {
			h.DB.Preload("Lines.Extension").First(&device, device.ID)
		}
		h.DB.First(&tenant, device.TenantID)

		if err := h.DB.Where("id = ?", device.Lines[0].ExtensionID).First(&ext).Error; err == nil {
			cmd := fmt.Sprintf("sofia profile internal flush_inbound_reg %s@%s reboot",
				ext.Extension, tenant.Domain)
			h.ESLManager.API(cmd)
		}
	}

	return c.JSON(fiber.Map{"message": "Reprovision triggered"})
}

// ==================
// Device Template Handlers (Tenant)
// ==================

// ListDeviceTemplates returns templates available to tenant
func (h *Handler) ListDeviceTemplates(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var templates []models.DeviceTemplate

	// Get tenant-specific + system templates
	h.DB.Where("tenant_id = ? OR tenant_id IS NULL", tenantID).
		Where("enabled = ?", true).
		Find(&templates)

	// Add device counts
	for i := range templates {
		h.DB.Model(&models.Device{}).Where("template_id = ?", templates[i].ID).Count(&templates[i].DeviceCount)
	}

	return c.JSON(fiber.Map{"data": templates})
}

// CreateDeviceTemplate creates a tenant-specific template
func (h *Handler) CreateDeviceTemplate(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var tmpl models.DeviceTemplate
	if err := c.BodyParser(&tmpl); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tmpl.TenantID = &tenantID
	tmpl.IsSystem = false // Tenant can't create system templates

	if err := h.DB.Create(&tmpl).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create template"})
	}

	return c.JSON(fiber.Map{"data": tmpl, "message": "Template created"})
}

// ==================
// Device Profile Handlers (Tenant)
// ==================

// ListDeviceProfiles returns all device profiles for a tenant
func (h *Handler) ListDeviceProfiles(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var profiles []models.DeviceProfile
	h.DB.Where("tenant_id = ?", tenantID).
		Preload("Template").
		Preload("Firmware").
		Order("name ASC").
		Find(&profiles)

	// Add device counts
	for i := range profiles {
		h.DB.Model(&models.Device{}).Where("profile_id = ?", profiles[i].ID).Count(&profiles[i].DeviceCount)
	}

	return c.JSON(fiber.Map{"data": profiles})
}

// CreateDeviceProfile creates a new device profile
func (h *Handler) CreateDeviceProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var profile models.DeviceProfile
	if err := c.BodyParser(&profile); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	profile.TenantID = tenantID

	if err := h.DB.Create(&profile).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create profile"})
	}

	return c.JSON(fiber.Map{"data": profile, "message": "Device profile created"})
}

// GetDeviceProfile returns a single device profile
func (h *Handler) GetDeviceProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var profile models.DeviceProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Template").
		Preload("Firmware").
		First(&profile).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Profile not found"})
	}

	return c.JSON(fiber.Map{"data": profile})
}

// UpdateDeviceProfile updates a device profile
func (h *Handler) UpdateDeviceProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	var profile models.DeviceProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&profile).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Profile not found"})
	}

	var input models.DeviceProfile
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	updates := map[string]interface{}{
		"name":                 input.Name,
		"description":          input.Description,
		"color":                input.Color,
		"manufacturer":         input.Manufacturer,
		"model":                input.Model,
		"template_id":          input.TemplateID,
		"config_overrides":     input.ConfigOverrides,
		"timezone":             input.Timezone,
		"language":             input.Language,
		"date_format":          input.DateFormat,
		"time_format":          input.TimeFormat,
		"ringtone_url":         input.RingtoneURL,
		"background_url":       input.BackgroundURL,
		"ntp_server":           input.NTPServer,
		"syslog_server":        input.SyslogServer,
		"default_volume":       input.DefaultVolume,
		"vad_enabled":          input.VADEnabled,
		"echo_cancellation":    input.EchoCancellation,
		"early_media":          input.EarlyMedia,
		"supported_codecs":     input.SupportedCodecs,
		"t38_enabled":          input.T38Enabled,
		"encryption_enabled":   input.EncryptionEnabled,
		"directory_enabled":    input.DirectoryEnabled,
		"call_waiting_enabled": input.CallWaitingEnabled,
		"call_record_enabled":  input.CallRecordEnabled,
		"auto_answer_enabled":  input.AutoAnswerEnabled,
		"blf_enabled":          input.BLFEnabled,
		"firmware_id":          input.FirmwareID,
		"is_default":           input.IsDefault,
		"enabled":              input.Enabled,
	}

	h.DB.Model(&profile).Updates(updates)

	// If setting as default, clear other defaults
	if input.IsDefault {
		h.DB.Model(&models.DeviceProfile{}).
			Where("tenant_id = ? AND id != ?", tenantID, profile.ID).
			Update("is_default", false)
	}

	return c.JSON(fiber.Map{"data": profile, "message": "Device profile updated"})
}

// DeleteDeviceProfile deletes a device profile
func (h *Handler) DeleteDeviceProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id := c.Params("id")

	// Check if any devices are using this profile
	var count int64
	h.DB.Model(&models.Device{}).Where("profile_id = ?", id).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Profile is in use by devices", "device_count": count})
	}

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.DeviceProfile{})
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Profile not found"})
	}

	return c.JSON(fiber.Map{"message": "Device profile deleted"})
}

// AssignDeviceToProfile assigns a device to a profile
func (h *Handler) AssignDeviceToProfile(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	deviceID := c.Params("id")

	var input struct {
		ProfileID *uint `json:"profile_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Verify profile belongs to tenant (if provided)
	if input.ProfileID != nil {
		var profile models.DeviceProfile
		if err := h.DB.Where("id = ? AND tenant_id = ?", *input.ProfileID, tenantID).First(&profile).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Profile not found"})
		}
	}

	result := h.DB.Model(&models.Device{}).
		Where("id = ? AND tenant_id = ?", deviceID, tenantID).
		Update("profile_id", input.ProfileID)

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Device not found"})
	}

	return c.JSON(fiber.Map{"message": "Device assigned to profile"})
}

// ==================
// Device Manufacturer Handlers (System Admin)
// ==================

// ListDeviceManufacturers returns all device manufacturers
func (h *Handler) ListDeviceManufacturers(c *fiber.Ctx) error {
	var manufacturers []models.DeviceManufacturer
	h.DB.Where("enabled = ?", true).Order("sort_order ASC, name ASC").Find(&manufacturers)

	// Get template counts per manufacturer
	for i := range manufacturers {
		var count int64
		h.DB.Model(&models.DeviceTemplate{}).
			Where("LOWER(manufacturer) = ?", manufacturers[i].Code).
			Count(&count)
		// We can add this as a computed field in response
		manufacturers[i].SortOrder = int(count) // Reusing SortOrder for count (temporary)
	}

	return c.JSON(fiber.Map{"data": manufacturers})
}

// CreateDeviceManufacturer creates a new manufacturer (system admin only)
func (h *Handler) CreateDeviceManufacturer(c *fiber.Ctx) error {
	var mfg models.DeviceManufacturer
	if err := c.BodyParser(&mfg); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Normalize code to lowercase
	mfg.Code = strings.ToLower(mfg.Code)
	if mfg.Code == "" || mfg.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Code and name are required"})
	}

	// Check for duplicate code
	var existing models.DeviceManufacturer
	if err := h.DB.Where("code = ?", mfg.Code).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Manufacturer code already exists"})
	}

	if err := h.DB.Create(&mfg).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create manufacturer"})
	}

	return c.JSON(fiber.Map{"data": mfg, "message": "Manufacturer created"})
}

// UpdateDeviceManufacturer updates a manufacturer
func (h *Handler) UpdateDeviceManufacturer(c *fiber.Ctx) error {
	id := c.Params("id")

	var mfg models.DeviceManufacturer
	if err := h.DB.Where("id = ?", id).First(&mfg).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Manufacturer not found"})
	}

	var input models.DeviceManufacturer
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	updates := map[string]interface{}{
		"name":               input.Name,
		"description":        input.Description,
		"logo_url":           input.LogoURL,
		"color":              input.Color,
		"user_agent_pattern": input.UserAgentPattern,
		"mac_prefix":         input.MACPrefix,
		"sort_order":         input.SortOrder,
		"is_default":         input.IsDefault,
		"enabled":            input.Enabled,
	}

	// Only update code if changing and not in use
	if input.Code != "" && input.Code != mfg.Code {
		var count int64
		h.DB.Model(&models.DeviceTemplate{}).Where("LOWER(manufacturer) = ?", mfg.Code).Count(&count)
		if count > 0 {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Cannot change code while templates exist"})
		}
		updates["code"] = strings.ToLower(input.Code)
	}

	h.DB.Model(&mfg).Updates(updates)

	return c.JSON(fiber.Map{"data": mfg, "message": "Manufacturer updated"})
}

// DeleteDeviceManufacturer deletes a manufacturer
func (h *Handler) DeleteDeviceManufacturer(c *fiber.Ctx) error {
	id := c.Params("id")

	// Check if any templates are using this manufacturer
	var mfg models.DeviceManufacturer
	if err := h.DB.Where("id = ?", id).First(&mfg).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Manufacturer not found"})
	}

	var count int64
	h.DB.Model(&models.DeviceTemplate{}).Where("LOWER(manufacturer) = ?", mfg.Code).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Manufacturer has templates", "template_count": count})
	}

	h.DB.Delete(&mfg)

	return c.JSON(fiber.Map{"message": "Manufacturer deleted"})
}
