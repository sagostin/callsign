package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/kataras/iris/v12"
)

// ==================
// Device Handlers
// ==================

// ListDevices returns all devices for the current tenant
func (h *Handler) ListDevices(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var devices []models.Device
	query := h.DB.Where("tenant_id = ?", tenantID).
		Preload("User").
		Preload("Template").
		Preload("Lines").
		Preload("Lines.Extension")

	// Optional filters
	if status := ctx.URLParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if deviceType := ctx.URLParam("type"); deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}
	if manufacturer := ctx.URLParam("manufacturer"); manufacturer != "" {
		query = query.Where("manufacturer = ?", manufacturer)
	}
	if userID := ctx.URLParam("user_id"); userID != "" {
		if userID == "unassigned" {
			query = query.Where("user_id IS NULL")
		} else {
			query = query.Where("user_id = ?", userID)
		}
	}

	query.Order("created_at DESC").Find(&devices)

	ctx.JSON(iris.Map{"data": devices})
}

// CreateDevice creates a new device
func (h *Handler) CreateDevice(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var device models.Device
	if err := ctx.ReadJSON(&device); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Normalize MAC address
	device.MAC = strings.ToLower(models.NormalizeMAC(device.MAC))
	if device.MAC == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "MAC address is required"})
		return
	}

	// Check for duplicate MAC
	var existing models.Device
	if err := h.DB.Where("mac = ?", device.MAC).First(&existing).Error; err == nil {
		ctx.StatusCode(iris.StatusConflict)
		ctx.JSON(iris.Map{"error": "Device with this MAC address already exists"})
		return
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
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create device"})
		return
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

	ctx.JSON(iris.Map{"data": device, "message": "Device created"})
}

// GetDevice returns a single device by ID or MAC
func (h *Handler) GetDevice(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	idOrMac := ctx.Params().Get("id")

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
			ctx.StatusCode(iris.StatusNotFound)
			ctx.JSON(iris.Map{"error": "Device not found"})
			return
		}
	}

	// Return device data with ProvisionToken exposed
	ctx.JSON(iris.Map{
		"data": iris.Map{
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
func (h *Handler) UpdateDevice(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var device models.Device
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Device not found"})
		return
	}

	// Parse input with ProvisionToken exposed
	var input struct {
		models.Device
		ProvisionToken string `json:"provision_token"`
	}
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Update allowed fields
	updates := map[string]interface{}{
		"name":          input.Name,
		"device_type":   input.DeviceType,
		"manufacturer":  input.Manufacturer,
		"model":         input.Model,
		"template_id":   input.TemplateID,
		"user_id":       input.UserID,
		"sip_server":    input.SIPServer,
		"sip_proxy":     input.SIPProxy,
		"sip_transport": input.SIPTransport,
		"sip_port":      input.SIPPort,
		"location_id":   input.LocationID,
		"enabled":       input.Enabled,
	}

	// Only update token if provided and non-empty (allow generation)
	if input.ProvisionToken != "" {
		updates["provision_token"] = input.ProvisionToken
	}

	h.DB.Model(&device).Updates(updates)

	// Reload with associations
	h.DB.Preload("User").Preload("Template").Preload("Lines").Preload("Lines.Extension").First(&device)

	ctx.JSON(iris.Map{"data": device, "message": "Device updated"})
}

// DeleteDevice deletes a device
func (h *Handler) DeleteDevice(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	// Delete lines first
	h.DB.Where("device_id = ?", id).Delete(&models.DeviceLine{})

	// Delete device
	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Device{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Device not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Device deleted"})
}

// AssignDeviceToUser assigns a device to a user
func (h *Handler) AssignDeviceToUser(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var input struct {
		UserID *uint `json:"user_id"`
	}
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	result := h.DB.Model(&models.Device{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("user_id", input.UserID)

	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Device not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Device assigned to user"})
}

// ==================
// Device Line Handlers
// ==================

// UpdateDeviceLines updates lines for a device
func (h *Handler) UpdateDeviceLines(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	deviceID := ctx.Params().Get("id")

	// Verify device belongs to tenant
	var device models.Device
	if err := h.DB.Where("id = ? AND tenant_id = ?", deviceID, tenantID).First(&device).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Device not found"})
		return
	}

	var lines []models.DeviceLine
	if err := ctx.ReadJSON(&lines); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
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

	ctx.JSON(iris.Map{"data": lines, "message": "Device lines updated"})
}

// ==================
// Provisioning Handlers
// ==================

// GetDeviceConfigSecure returns provisioning config with tenant secret verification
// URL: /provision/{tenant_uuid}/{secret}/{mac}.cfg
func (h *Handler) GetDeviceConfigSecure(ctx iris.Context) {
	tenantUUID := ctx.Params().Get("tenant")
	secret := ctx.Params().Get("secret")
	mac := ctx.Params().Get("mac")
	clientIP := ctx.RemoteAddr()
	userAgent := ctx.GetHeader("User-Agent")

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

		ctx.StatusCode(iris.StatusNotFound)
		ctx.Text("Invalid provisioning URL")
		return
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

		ctx.StatusCode(iris.StatusForbidden)
		ctx.Text("Provisioning disabled for this tenant")
		return
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

		ctx.StatusCode(iris.StatusForbidden)
		ctx.Text("Invalid provisioning URL")
		return
	}

	// Find device
	var device models.Device
	if err := h.DB.Where("mac = ? AND tenant_id = ?", normalizedMAC, tenant.ID).
		Preload("Template").
		Preload("Profile").
		Preload("Lines").
		Preload("Lines.Extension").
		First(&device).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.Text("Device not found")
		return
	}

	if !device.Enabled {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Text("Device is disabled")
		return
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
		ctx.StatusCode(iris.StatusNotFound)
		ctx.Text("No template available for this device")
		return
	}

	// Verify User-Agent if pattern is set
	if tmpl.UserAgentPattern != "" {
		userAgent := ctx.GetHeader("User-Agent")
		matched, _ := regexp.MatchString(tmpl.UserAgentPattern, userAgent)
		if !matched {
			ctx.StatusCode(iris.StatusForbidden)
			ctx.Text("Device not authorized")
			return
		}
	}

	// Verify MAC pattern if set
	if tmpl.MACPattern != "" {
		matched, _ := regexp.MatchString(tmpl.MACPattern, normalizedMAC)
		if !matched {
			ctx.StatusCode(iris.StatusForbidden)
			ctx.Text("Device model mismatch")
			return
		}
	}

	h.renderDeviceConfig(ctx, &device, &tmpl, &tenant)
}

// GetDeviceConfig returns the provisioning configuration for a device (legacy, no auth)
func (h *Handler) GetDeviceConfig(ctx iris.Context) {
	mac := ctx.Params().Get("mac")
	normalizedMAC := strings.ToLower(models.NormalizeMAC(mac))

	var device models.Device
	if err := h.DB.Where("mac = ?", normalizedMAC).
		Preload("Template").
		Preload("Profile").
		Preload("Lines").
		Preload("Lines.Extension").
		First(&device).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.Text("Device not found")
		return
	}

	if !device.Enabled {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Text("Device is disabled")
		return
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
		ctx.StatusCode(iris.StatusNotFound)
		ctx.Text("No template available for this device")
		return
	}

	// Get tenant for domain
	var tenant models.Tenant
	h.DB.First(&tenant, device.TenantID)

	h.renderDeviceConfig(ctx, &device, &tmpl, &tenant)
}

// renderDeviceConfig renders the provisioning template
func (h *Handler) renderDeviceConfig(ctx iris.Context, device *models.Device, tmpl *models.DeviceTemplate, tenant *models.Tenant) {
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
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.Text("Template parse error: " + err.Error())
		return
	}

	// Set content type based on config type
	switch tmpl.ConfigType {
	case "xml":
		ctx.ContentType("application/xml")
	case "cfg":
		ctx.ContentType("text/plain")
	case "json":
		ctx.ContentType("application/json")
	default:
		ctx.ContentType("text/plain")
	}

	if err := t.Execute(ctx.ResponseWriter(), vars); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.Text("Template execution error: " + err.Error())
		return
	}

	// Update last provision time
	h.DB.Model(device).Updates(map[string]interface{}{
		"last_provision":  time.Now(),
		"provision_count": device.ProvisionCount + 1,
	})
}

// ReprovisionDevice triggers a re-provision (SIP NOTIFY)
func (h *Handler) ReprovisionDevice(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var device models.Device
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Device not found"})
		return
	}

	// TODO: Send SIP NOTIFY to trigger re-provision
	// This would be done via ESL:
	// sofia profile internal flush_inbound_reg <user>@<domain> reboot

	ctx.JSON(iris.Map{"message": "Reprovision triggered"})
}

// ==================
// Device Template Handlers (Tenant)
// ==================

// ListDeviceTemplates returns templates available to tenant
func (h *Handler) ListDeviceTemplates(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var templates []models.DeviceTemplate

	// Get tenant-specific + system templates
	h.DB.Where("tenant_id = ? OR tenant_id IS NULL", tenantID).
		Where("enabled = ?", true).
		Find(&templates)

	// Add device counts
	for i := range templates {
		h.DB.Model(&models.Device{}).Where("template_id = ?", templates[i].ID).Count(&templates[i].DeviceCount)
	}

	ctx.JSON(iris.Map{"data": templates})
}

// CreateDeviceTemplate creates a tenant-specific template
func (h *Handler) CreateDeviceTemplate(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var tmpl models.DeviceTemplate
	if err := ctx.ReadJSON(&tmpl); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	tmpl.TenantID = &tenantID
	tmpl.IsSystem = false // Tenant can't create system templates

	if err := h.DB.Create(&tmpl).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create template"})
		return
	}

	ctx.JSON(iris.Map{"data": tmpl, "message": "Template created"})
}

// ==================
// Device Profile Handlers (Tenant)
// ==================

// ListDeviceProfiles returns all device profiles for a tenant
func (h *Handler) ListDeviceProfiles(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

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

	ctx.JSON(iris.Map{"data": profiles})
}

// CreateDeviceProfile creates a new device profile
func (h *Handler) CreateDeviceProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var profile models.DeviceProfile
	if err := ctx.ReadJSON(&profile); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	profile.TenantID = tenantID

	if err := h.DB.Create(&profile).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create profile"})
		return
	}

	ctx.JSON(iris.Map{"data": profile, "message": "Device profile created"})
}

// GetDeviceProfile returns a single device profile
func (h *Handler) GetDeviceProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var profile models.DeviceProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Template").
		Preload("Firmware").
		First(&profile).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Profile not found"})
		return
	}

	ctx.JSON(iris.Map{"data": profile})
}

// UpdateDeviceProfile updates a device profile
func (h *Handler) UpdateDeviceProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	var profile models.DeviceProfile
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&profile).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Profile not found"})
		return
	}

	var input models.DeviceProfile
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
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

	ctx.JSON(iris.Map{"data": profile, "message": "Device profile updated"})
}

// DeleteDeviceProfile deletes a device profile
func (h *Handler) DeleteDeviceProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id := ctx.Params().Get("id")

	// Check if any devices are using this profile
	var count int64
	h.DB.Model(&models.Device{}).Where("profile_id = ?", id).Count(&count)
	if count > 0 {
		ctx.StatusCode(iris.StatusConflict)
		ctx.JSON(iris.Map{"error": "Profile is in use by devices", "device_count": count})
		return
	}

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.DeviceProfile{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Profile not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Device profile deleted"})
}

// AssignDeviceToProfile assigns a device to a profile
func (h *Handler) AssignDeviceToProfile(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	deviceID := ctx.Params().Get("id")

	var input struct {
		ProfileID *uint `json:"profile_id"`
	}
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Verify profile belongs to tenant (if provided)
	if input.ProfileID != nil {
		var profile models.DeviceProfile
		if err := h.DB.Where("id = ? AND tenant_id = ?", *input.ProfileID, tenantID).First(&profile).Error; err != nil {
			ctx.StatusCode(iris.StatusNotFound)
			ctx.JSON(iris.Map{"error": "Profile not found"})
			return
		}
	}

	result := h.DB.Model(&models.Device{}).
		Where("id = ? AND tenant_id = ?", deviceID, tenantID).
		Update("profile_id", input.ProfileID)

	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Device not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Device assigned to profile"})
}
