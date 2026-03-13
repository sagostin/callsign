package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"

	"github.com/gofiber/fiber/v2"
)

// TenantSettings represents the JSONB settings column structure
type TenantSettings struct {
	// General
	SIPDomain        string `json:"sip_domain"`
	PortalDomain     string `json:"portal_domain"`
	Timezone         string `json:"timezone"`
	OperatorExt      string `json:"operator_ext"`
	FallbackCallerID string `json:"fallback_caller_id"`
	PanicEnabled     bool   `json:"panic_enabled"`

	// SMTP
	SMTPOverride   bool   `json:"smtp_override"`
	SMTPHost       string `json:"smtp_host"`
	SMTPPort       string `json:"smtp_port"`
	SMTPUsername   string `json:"smtp_username"`
	SMTPPassword   string `json:"smtp_password"`
	SMTPFromEmail  string `json:"smtp_from_email"`
	SMTPEncryption string `json:"smtp_encryption"`

	// Messaging
	MessagingProvider   string `json:"messaging_provider"`
	MessagingAccountSID string `json:"messaging_account_sid"`
	MessagingAuthToken  string `json:"messaging_auth_token"`

	// Hospitality
	HospitalityEnabled bool `json:"hospitality_enabled"`

	// SSL/Security
	ForceHTTPS bool `json:"force_https"`

	// User Limits
	VMLimit      int    `json:"vm_limit"`
	FaxRetention string `json:"fax_retention"`
}

// GetTenantSettings returns the current tenant's settings
func (h *Handler) GetTenantSettings(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "GetTenantSettings: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "GetTenantSettings: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	// Parse settings from JSONB
	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":                   tenant.ID,
			"uuid":                 tenant.UUID.String(),
			"name":                 tenant.Name,
			"domain":               tenant.Domain,
			"settings":             settings,
			"ssl_enabled":          tenant.SSLEnabled,
			"ssl_domain":           tenant.SSLDomain,
			"provisioning_secret":  tenant.ProvisioningSecret,
			"provisioning_enabled": tenant.ProvisioningEnabled,
		},
	})
}

// UpdateTenantSettings updates the current tenant's settings
func (h *Handler) UpdateTenantSettings(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "UpdateTenantSettings: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "UpdateTenantSettings: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	// Parse incoming settings
	// We use a struct that includes both the settings JSONB fields and the top-level tenant fields
	var req struct {
		TenantSettings         // Embed settings
		TenantName     *string `json:"name"`
		Domain         *string `json:"domain"`
		SSLDomain      *string `json:"ssl_domain"`
		SSLEnabled     *bool   `json:"ssl_enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.logWarn("SETTINGS", "UpdateTenantSettings: Invalid request body", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	log.Printf("UpdateTenantSettings: Payload received: %+v", req)

	// Update top-level tenant fields if provided
	if req.TenantName != nil {
		log.Printf("UpdateTenantSettings: Updating Name to %s", *req.TenantName)
		tenant.Name = *req.TenantName
	}
	if req.Domain != nil {
		tenant.Domain = *req.Domain
	}
	if req.SSLDomain != nil {
		tenant.SSLDomain = *req.SSLDomain
	}
	if req.SSLEnabled != nil {
		tenant.SSLEnabled = *req.SSLEnabled
		// Also update the setting inside JSONB for consistency if needed, though usually SSLEnabled is on the model
		req.TenantSettings.ForceHTTPS = *req.SSLEnabled // Optional: link force https? No, keep separate.
	}

	// Update the JSONB settings
	settingsJSON, _ := json.Marshal(req.TenantSettings)
	tenant.Settings = string(settingsJSON)

	if err := h.DB.Save(&tenant).Error; err != nil {
		h.logError("SETTINGS", "UpdateTenantSettings: Failed to save settings", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save settings"})
	}

	return c.JSON(fiber.Map{"message": "Settings updated", "data": req})
}

// GetTenantBranding returns branding settings for the current tenant
func (h *Handler) GetTenantBranding(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "GetTenantBranding: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "GetTenantBranding: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"whitelabel_enabled": tenant.WhitelabelEnabled,
			"name":               tenant.WhitelabelName,
			"logo_url":           tenant.WhitelabelLogo,
			"primary_color":      tenant.WhitelabelPrimary,
		},
	})
}

// UpdateTenantBranding updates branding settings
func (h *Handler) UpdateTenantBranding(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "UpdateTenantBranding: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "UpdateTenantBranding: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	var req struct {
		WhitelabelEnabled bool   `json:"whitelabel_enabled"`
		Name              string `json:"name"`
		LogoURL           string `json:"logo_url"`
		PrimaryColor      string `json:"primary_color"`
	}
	if err := c.BodyParser(&req); err != nil {
		h.logWarn("SETTINGS", "UpdateTenantBranding: Invalid request body", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tenant.WhitelabelEnabled = req.WhitelabelEnabled
	tenant.WhitelabelName = req.Name
	tenant.WhitelabelLogo = req.LogoURL
	tenant.WhitelabelPrimary = req.PrimaryColor

	if err := h.DB.Save(&tenant).Error; err != nil {
		h.logError("SETTINGS", "UpdateTenantBranding: Failed to save branding", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save branding"})
	}

	return c.JSON(fiber.Map{"message": "Branding updated"})
}

// GetTenantSMTP returns SMTP settings for the current tenant
func (h *Handler) GetTenantSMTP(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "GetTenantSMTP: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "GetTenantSMTP: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"override":   settings.SMTPOverride,
			"host":       settings.SMTPHost,
			"port":       settings.SMTPPort,
			"username":   settings.SMTPUsername,
			"from_email": settings.SMTPFromEmail,
			"encryption": settings.SMTPEncryption,
		},
	})
}

// UpdateTenantSMTP updates SMTP settings
func (h *Handler) UpdateTenantSMTP(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "UpdateTenantSMTP: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "UpdateTenantSMTP: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	var req struct {
		Override   bool   `json:"override"`
		Host       string `json:"host"`
		Port       string `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		FromEmail  string `json:"from_email"`
		Encryption string `json:"encryption"`
	}
	if err := c.BodyParser(&req); err != nil {
		h.logWarn("SETTINGS", "UpdateTenantSMTP: Invalid request body", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	settings.SMTPOverride = req.Override
	settings.SMTPHost = req.Host
	settings.SMTPPort = req.Port
	settings.SMTPUsername = req.Username
	if req.Password != "" {
		settings.SMTPPassword = req.Password
	}
	settings.SMTPFromEmail = req.FromEmail
	settings.SMTPEncryption = req.Encryption

	settingsJSON, _ := json.Marshal(settings)
	tenant.Settings = string(settingsJSON)

	if err := h.DB.Save(&tenant).Error; err != nil {
		h.logError("SETTINGS", "UpdateTenantSMTP: Failed to save SMTP settings", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save SMTP settings"})
	}

	return c.JSON(fiber.Map{"message": "SMTP settings updated"})
}

// TestTenantSMTP sends a test email
func (h *Handler) TestTenantSMTP(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	claims := middleware.GetClaims(c)

	// Get tenant SMTP settings
	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "TestTenantSMTP: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	// Parse settings to get SMTP config
	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	if settings.SMTPHost == "" {
		h.logWarn("SETTINGS", "TestTenantSMTP: SMTP not configured", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "SMTP not configured"})
	}

	// Build SMTP address
	smtpPort := settings.SMTPPort
	if smtpPort == "" {
		smtpPort = "587"
	}
	smtpAddr := settings.SMTPHost + ":" + smtpPort

	// Determine recipient — use current user's email or a fallback
	recipient := ""
	if claims != nil && claims.Email != "" {
		recipient = claims.Email
	} else {
		recipient = settings.SMTPFromEmail
	}

	if recipient == "" {
		h.logWarn("SETTINGS", "TestTenantSMTP: No recipient email available", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "No recipient email available"})
	}

	// Compose test message
	from := settings.SMTPFromEmail
	if from == "" {
		from = settings.SMTPUsername
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: CallSign SMTP Test\r\n\r\n"+
		"This is a test email from CallSign PBX.\r\nYour SMTP settings are working correctly.\r\n",
		from, recipient)

	// Try to send
	var auth smtp.Auth
	if settings.SMTPUsername != "" {
		auth = smtp.PlainAuth("", settings.SMTPUsername, settings.SMTPPassword, settings.SMTPHost)
	}

	err := smtp.SendMail(smtpAddr, auth, from, []string{recipient}, []byte(msg))
	if err != nil {
		return c.Status(http.StatusBadGateway).JSON(fiber.Map{
			"error":   "SMTP test failed",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Test email sent successfully",
		"recipient": recipient,
	})
}

// GetTenantMessaging returns messaging/SMS settings
func (h *Handler) GetTenantMessaging(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "GetTenantMessaging: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "GetTenantMessaging: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"provider":    settings.MessagingProvider,
			"account_sid": settings.MessagingAccountSID,
		},
	})
}

// UpdateTenantMessaging updates messaging settings
func (h *Handler) UpdateTenantMessaging(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "UpdateTenantMessaging: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "UpdateTenantMessaging: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	var req struct {
		Provider   string `json:"provider"`
		AccountSID string `json:"account_sid"`
		AuthToken  string `json:"auth_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		h.logWarn("SETTINGS", "UpdateTenantMessaging: Invalid request body", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	settings.MessagingProvider = req.Provider
	settings.MessagingAccountSID = req.AccountSID
	if req.AuthToken != "" {
		settings.MessagingAuthToken = req.AuthToken
	}

	settingsJSON, _ := json.Marshal(settings)
	tenant.Settings = string(settingsJSON)

	if err := h.DB.Save(&tenant).Error; err != nil {
		h.logError("SETTINGS", "UpdateTenantMessaging: Failed to save messaging settings", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save messaging settings"})
	}

	return c.JSON(fiber.Map{"message": "Messaging settings updated"})
}

// GetTenantHospitality returns hospitality settings
func (h *Handler) GetTenantHospitality(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "GetTenantHospitality: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "GetTenantHospitality: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"enabled": settings.HospitalityEnabled,
		},
	})
}

// UpdateTenantHospitality updates hospitality settings
func (h *Handler) UpdateTenantHospitality(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		h.logWarn("SETTINGS", "UpdateTenantHospitality: Tenant ID required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		h.logWarn("SETTINGS", "UpdateTenantHospitality: Tenant not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.BodyParser(&req); err != nil {
		h.logWarn("SETTINGS", "UpdateTenantHospitality: Invalid request body", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	settings.HospitalityEnabled = req.Enabled

	settingsJSON, _ := json.Marshal(settings)
	tenant.Settings = string(settingsJSON)

	if err := h.DB.Save(&tenant).Error; err != nil {
		h.logError("SETTINGS", "UpdateTenantHospitality: Failed to save hospitality settings", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save hospitality settings"})
	}

	return c.JSON(fiber.Map{"message": "Hospitality settings updated"})
}
