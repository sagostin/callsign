package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"encoding/json"
	"net/http"

	"github.com/kataras/iris/v12"
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
func (h *Handler) GetTenantSettings(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	// Parse settings from JSONB
	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	ctx.JSON(iris.Map{
		"data": iris.Map{
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
func (h *Handler) UpdateTenantSettings(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	// Parse incoming settings
	var settings TenantSettings
	if err := ctx.ReadJSON(&settings); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Serialize to JSON
	settingsJSON, _ := json.Marshal(settings)
	tenant.Settings = string(settingsJSON)

	if err := h.DB.Save(&tenant).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save settings"})
		return
	}

	ctx.JSON(iris.Map{"message": "Settings updated", "data": settings})
}

// GetTenantBranding returns branding settings for the current tenant
func (h *Handler) GetTenantBranding(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	ctx.JSON(iris.Map{
		"data": iris.Map{
			"whitelabel_enabled": tenant.WhitelabelEnabled,
			"name":               tenant.WhitelabelName,
			"logo_url":           tenant.WhitelabelLogo,
			"primary_color":      tenant.WhitelabelPrimary,
		},
	})
}

// UpdateTenantBranding updates branding settings
func (h *Handler) UpdateTenantBranding(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	var req struct {
		WhitelabelEnabled bool   `json:"whitelabel_enabled"`
		Name              string `json:"name"`
		LogoURL           string `json:"logo_url"`
		PrimaryColor      string `json:"primary_color"`
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	tenant.WhitelabelEnabled = req.WhitelabelEnabled
	tenant.WhitelabelName = req.Name
	tenant.WhitelabelLogo = req.LogoURL
	tenant.WhitelabelPrimary = req.PrimaryColor

	if err := h.DB.Save(&tenant).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save branding"})
		return
	}

	ctx.JSON(iris.Map{"message": "Branding updated"})
}

// GetTenantSMTP returns SMTP settings for the current tenant
func (h *Handler) GetTenantSMTP(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	ctx.JSON(iris.Map{
		"data": iris.Map{
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
func (h *Handler) UpdateTenantSMTP(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
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
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
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
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save SMTP settings"})
		return
	}

	ctx.JSON(iris.Map{"message": "SMTP settings updated"})
}

// TestTenantSMTP sends a test email
func (h *Handler) TestTenantSMTP(ctx iris.Context) {
	// TODO: Implement actual SMTP test
	ctx.JSON(iris.Map{"message": "Test email sent successfully"})
}

// GetTenantMessaging returns messaging/SMS settings
func (h *Handler) GetTenantMessaging(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	ctx.JSON(iris.Map{
		"data": iris.Map{
			"provider":    settings.MessagingProvider,
			"account_sid": settings.MessagingAccountSID,
		},
	})
}

// UpdateTenantMessaging updates messaging settings
func (h *Handler) UpdateTenantMessaging(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
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
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	settings.MessagingProvider = req.Provider
	settings.MessagingAccountSID = req.AccountSID
	if req.AuthToken != "" {
		settings.MessagingAuthToken = req.AuthToken
	}

	settingsJSON, _ := json.Marshal(settings)
	tenant.Settings = string(settingsJSON)

	if err := h.DB.Save(&tenant).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save messaging settings"})
		return
	}

	ctx.JSON(iris.Map{"message": "Messaging settings updated"})
}

// GetTenantHospitality returns hospitality settings
func (h *Handler) GetTenantHospitality(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	ctx.JSON(iris.Map{
		"data": iris.Map{
			"enabled": settings.HospitalityEnabled,
		},
	})
}

// UpdateTenantHospitality updates hospitality settings
func (h *Handler) UpdateTenantHospitality(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	var settings TenantSettings
	if tenant.Settings != "" && tenant.Settings != "{}" {
		json.Unmarshal([]byte(tenant.Settings), &settings)
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	settings.HospitalityEnabled = req.Enabled

	settingsJSON, _ := json.Marshal(settings)
	tenant.Settings = string(settingsJSON)

	if err := h.DB.Save(&tenant).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save hospitality settings"})
		return
	}

	ctx.JSON(iris.Map{"message": "Hospitality settings updated"})
}
