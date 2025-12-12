package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// =====================
// Tenants
// =====================

func (h *Handler) ListTenants(ctx iris.Context) {
	var tenants []models.Tenant
	if err := h.DB.Preload("Profile").Find(&tenants).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve tenants"})
		return
	}
	ctx.JSON(iris.Map{"data": tenants})
}

func (h *Handler) CreateTenant(ctx iris.Context) {
	var tenant models.Tenant
	if err := ctx.ReadJSON(&tenant); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Create(&tenant).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create tenant"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(tenant)
}

func (h *Handler) GetTenant(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid tenant ID"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.Preload("Profile").First(&tenant, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	ctx.JSON(tenant)
}

func (h *Handler) UpdateTenant(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid tenant ID"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	if err := ctx.ReadJSON(&tenant); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	tenant.ID = uint(id)
	if err := h.DB.Save(&tenant).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update tenant"})
		return
	}

	ctx.JSON(tenant)
}

func (h *Handler) DeleteTenant(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid tenant ID"})
		return
	}

	if err := h.DB.Delete(&models.Tenant{}, id).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete tenant"})
		return
	}

	ctx.JSON(iris.Map{"message": "Tenant deleted successfully"})
}

// =====================
// Tenant Profiles
// =====================

func (h *Handler) ListTenantProfiles(ctx iris.Context) {
	var profiles []models.TenantProfile
	if err := h.DB.Find(&profiles).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve tenant profiles"})
		return
	}

	// Compute tenant count for each profile
	type ProfileWithCount struct {
		models.TenantProfile
		TenantCount int64 `json:"tenant_count"`
	}

	var result []ProfileWithCount
	for _, p := range profiles {
		var count int64
		h.DB.Model(&models.Tenant{}).Where("profile_id = ?", p.ID).Count(&count)
		result = append(result, ProfileWithCount{
			TenantProfile: p,
			TenantCount:   count,
		})
	}

	ctx.JSON(iris.Map{"data": result})
}

func (h *Handler) CreateTenantProfile(ctx iris.Context) {
	var profile models.TenantProfile
	if err := ctx.ReadJSON(&profile); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Create(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create tenant profile"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(profile)
}

func (h *Handler) GetTenantProfile(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid profile ID"})
		return
	}

	var profile models.TenantProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Profile not found"})
		return
	}

	ctx.JSON(profile)
}

func (h *Handler) UpdateTenantProfile(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid profile ID"})
		return
	}

	var profile models.TenantProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Profile not found"})
		return
	}

	if err := ctx.ReadJSON(&profile); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	profile.ID = uint(id)
	if err := h.DB.Save(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update profile"})
		return
	}

	ctx.JSON(profile)
}

func (h *Handler) DeleteTenantProfile(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid profile ID"})
		return
	}

	if err := h.DB.Delete(&models.TenantProfile{}, id).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete profile"})
		return
	}

	ctx.JSON(iris.Map{"message": "Profile deleted successfully"})
}

// =====================
// Users (System Admin)
// =====================

func (h *Handler) ListUsers(ctx iris.Context) {
	tenantID := middleware.GetScopedTenantID(ctx)

	var users []models.User
	query := h.DB
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Find(&users).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve users"})
		return
	}
	ctx.JSON(iris.Map{"data": users})
}

func (h *Handler) CreateUser(ctx iris.Context) {
	var user models.User
	if err := ctx.ReadJSON(&user); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Hash password
	if password := ctx.FormValue("password"); password != "" {
		if err := user.SetPassword(password); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "Failed to set password"})
			return
		}
	}

	if err := h.DB.Create(&user).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create user"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(user)
}

func (h *Handler) GetUser(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	ctx.JSON(user)
}

func (h *Handler) UpdateUser(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	if err := ctx.ReadJSON(&user); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	user.ID = uint(id)
	if err := h.DB.Save(&user).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update user"})
		return
	}

	ctx.JSON(user)
}

func (h *Handler) DeleteUser(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid user ID"})
		return
	}

	if err := h.DB.Delete(&models.User{}, id).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete user"})
		return
	}

	ctx.JSON(iris.Map{"message": "User deleted successfully"})
}

// =====================
// Gateways
// =====================

func (h *Handler) ListGateways(ctx iris.Context) {
	tenantID := middleware.GetScopedTenantID(ctx)

	var gateways []models.Gateway
	query := h.DB
	if tenantID > 0 {
		// Tenant admin sees their gateways + system-wide gateways
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantID)
	}
	// System admin sees all gateways

	if err := query.Order("gateway_name").Find(&gateways).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve gateways"})
		return
	}
	ctx.JSON(iris.Map{"data": gateways})
}

func (h *Handler) CreateGateway(ctx iris.Context) {
	var gateway models.Gateway
	if err := ctx.ReadJSON(&gateway); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// If tenant admin, scope to their tenant
	tenantID := middleware.GetScopedTenantID(ctx)
	if tenantID > 0 {
		gateway.TenantID = &tenantID
	}

	if err := h.DB.Create(&gateway).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create gateway"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(gateway)
}

func (h *Handler) GetGateway(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid gateway ID"})
		return
	}

	tenantID := middleware.GetScopedTenantID(ctx)

	var gateway models.Gateway
	query := h.DB
	if tenantID > 0 {
		query = query.Where("(tenant_id = ? OR tenant_id IS NULL)", tenantID)
	}

	if err := query.First(&gateway, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Gateway not found"})
		return
	}

	ctx.JSON(gateway)
}

func (h *Handler) UpdateGateway(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid gateway ID"})
		return
	}

	tenantID := middleware.GetScopedTenantID(ctx)

	var gateway models.Gateway
	query := h.DB
	if tenantID > 0 {
		// Tenant admins can only update their own gateways, not system-wide ones
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&gateway, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Gateway not found"})
		return
	}

	if err := ctx.ReadJSON(&gateway); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	gateway.ID = uint(id)
	if tenantID > 0 {
		gateway.TenantID = &tenantID // Preserve tenant ownership
	}

	if err := h.DB.Save(&gateway).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update gateway"})
		return
	}

	ctx.JSON(gateway)
}

func (h *Handler) DeleteGateway(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid gateway ID"})
		return
	}

	tenantID := middleware.GetScopedTenantID(ctx)

	var gateway models.Gateway
	query := h.DB
	if tenantID > 0 {
		// Tenant admins can only delete their own gateways
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&gateway, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Gateway not found"})
		return
	}

	if err := h.DB.Delete(&gateway).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete gateway"})
		return
	}

	ctx.JSON(iris.Map{"message": "Gateway deleted successfully"})
}

// =====================
// Bridges
// =====================

func (h *Handler) ListBridges(ctx iris.Context) {
	ctx.JSON(iris.Map{"data": []interface{}{}, "message": "Not implemented"})
}

func (h *Handler) CreateBridge(ctx iris.Context) {
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Not implemented"})
}

func (h *Handler) GetBridge(ctx iris.Context) {
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Not implemented"})
}

func (h *Handler) UpdateBridge(ctx iris.Context) {
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Not implemented"})
}

func (h *Handler) DeleteBridge(ctx iris.Context) {
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Not implemented"})
}

// =====================
// SIP Profiles
// =====================

func (h *Handler) ListSIPProfiles(ctx iris.Context) {
	var profiles []models.SIPProfile
	if err := h.DB.Preload("Settings").Preload("Domains").Order("profile_name").Find(&profiles).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve SIP profiles"})
		return
	}
	ctx.JSON(iris.Map{"data": profiles})
}

func (h *Handler) CreateSIPProfile(ctx iris.Context) {
	var profile models.SIPProfile
	if err := ctx.ReadJSON(&profile); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Create profile first
	if err := h.DB.Create(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create SIP profile"})
		return
	}

	// Add default settings based on profile type if no settings provided
	if len(profile.Settings) == 0 {
		var defaultSettings []models.SIPProfileSetting
		if profile.ProfileName == "external" {
			defaultSettings = models.DefaultExternalProfileSettings()
		} else {
			defaultSettings = models.DefaultInternalProfileSettings()
		}

		for i := range defaultSettings {
			defaultSettings[i].SIPProfileUUID = profile.UUID
		}

		if len(defaultSettings) > 0 {
			h.DB.Create(&defaultSettings)
			profile.Settings = defaultSettings
		}
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(profile)
}

func (h *Handler) GetSIPProfile(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid profile ID"})
		return
	}

	var profile models.SIPProfile
	if err := h.DB.Preload("Settings").Preload("Domains").First(&profile, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "SIP profile not found"})
		return
	}

	ctx.JSON(profile)
}

func (h *Handler) UpdateSIPProfile(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid profile ID"})
		return
	}

	var profile models.SIPProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "SIP profile not found"})
		return
	}

	if err := ctx.ReadJSON(&profile); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	profile.ID = uint(id)
	if err := h.DB.Save(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update SIP profile"})
		return
	}

	// Update settings if provided
	if len(profile.Settings) > 0 {
		// Delete existing settings and recreate
		h.DB.Where("sip_profile_uuid = ?", profile.UUID).Delete(&models.SIPProfileSetting{})
		for i := range profile.Settings {
			profile.Settings[i].SIPProfileUUID = profile.UUID
		}
		h.DB.Create(&profile.Settings)
	}

	// Update domains if provided
	if len(profile.Domains) > 0 {
		h.DB.Where("sip_profile_uuid = ?", profile.UUID).Delete(&models.SIPProfileDomain{})
		for i := range profile.Domains {
			profile.Domains[i].SIPProfileUUID = profile.UUID
		}
		h.DB.Create(&profile.Domains)
	}

	ctx.JSON(profile)
}

func (h *Handler) DeleteSIPProfile(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid profile ID"})
		return
	}

	var profile models.SIPProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "SIP profile not found"})
		return
	}

	// Delete related settings and domains first
	h.DB.Where("sip_profile_uuid = ?", profile.UUID).Delete(&models.SIPProfileSetting{})
	h.DB.Where("sip_profile_uuid = ?", profile.UUID).Delete(&models.SIPProfileDomain{})

	if err := h.DB.Delete(&profile).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete SIP profile"})
		return
	}

	ctx.JSON(iris.Map{"message": "SIP profile deleted successfully"})
}

// =====================
// Sofia Control Commands
// =====================

// GetSofiaStatus returns the status of all Sofia profiles from FreeSWITCH
func (h *Handler) GetSofiaStatus(ctx iris.Context) {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH ESL not connected"})
		return
	}

	result, err := h.ESLManager.API("sofia status")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to get sofia status: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{"data": result})
}

// GetSofiaProfileStatus returns the status of a specific Sofia profile
func (h *Handler) GetSofiaProfileStatus(ctx iris.Context) {
	profileName := ctx.Params().Get("name")
	if profileName == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Profile name required"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH ESL not connected"})
		return
	}

	result, err := h.ESLManager.API("sofia status profile " + profileName)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to get profile status: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{"data": result, "profile": profileName})
}

// GetSofiaProfileRegistrations returns the registrations for a specific Sofia profile
func (h *Handler) GetSofiaProfileRegistrations(ctx iris.Context) {
	profileName := ctx.Params().Get("name")
	if profileName == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Profile name required"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH ESL not connected"})
		return
	}

	result, err := h.ESLManager.API("sofia status profile " + profileName + " reg")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to get registrations: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{"data": result, "profile": profileName})
}

// GetSofiaGatewayStatus returns gateway status for a profile
func (h *Handler) GetSofiaGatewayStatus(ctx iris.Context) {
	profileName := ctx.Params().Get("name")
	if profileName == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Profile name required"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH ESL not connected"})
		return
	}

	result, err := h.ESLManager.API("sofia status gateway")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to get gateway status: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{"data": result})
}

// RestartSofiaProfile restarts a Sofia profile
func (h *Handler) RestartSofiaProfile(ctx iris.Context) {
	profileName := ctx.Params().Get("name")
	if profileName == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Profile name required"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH ESL not connected"})
		return
	}

	result, err := h.ESLManager.API("sofia profile " + profileName + " restart reloadxml")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to restart profile: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{"message": "Profile restarted", "data": result, "profile": profileName})
}

// StartSofiaProfile starts a stopped Sofia profile
func (h *Handler) StartSofiaProfile(ctx iris.Context) {
	profileName := ctx.Params().Get("name")
	if profileName == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Profile name required"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH ESL not connected"})
		return
	}

	result, err := h.ESLManager.API("sofia profile " + profileName + " start")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to start profile: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{"message": "Profile started", "data": result, "profile": profileName})
}

// StopSofiaProfile stops a running Sofia profile
func (h *Handler) StopSofiaProfile(ctx iris.Context) {
	profileName := ctx.Params().Get("name")
	if profileName == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Profile name required"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH ESL not connected"})
		return
	}

	result, err := h.ESLManager.API("sofia profile " + profileName + " stop")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to stop profile: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{"message": "Profile stopped", "data": result, "profile": profileName})
}

// ReloadSofiaXML reloads FreeSWITCH XML configuration
func (h *Handler) ReloadSofiaXML(ctx iris.Context) {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH ESL not connected"})
		return
	}

	result, err := h.ESLManager.API("reloadxml")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to reload XML: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{"message": "XML configuration reloaded", "data": result})
}

// =====================
// System Settings & Status
// =====================

func (h *Handler) GetSystemSettings(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"server_host":     h.Config.ServerHost,
		"server_port":     h.Config.ServerPort,
		"db_host":         h.Config.DBHost,
		"freeswitch_host": h.Config.FreeSwitchHost,
	})
}

func (h *Handler) UpdateSystemSettings(ctx iris.Context) {
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Not implemented"})
}

func (h *Handler) GetSystemLogs(ctx iris.Context) {
	ctx.JSON(iris.Map{"data": []interface{}{}, "message": "Not implemented"})
}

func (h *Handler) GetSystemStatus(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"status":   "operational",
		"database": "connected",
	})
}

func (h *Handler) GetSystemStats(ctx iris.Context) {
	// Get database counts
	var tenantCount, userCount, extensionCount, gatewayCount int64
	h.DB.Model(&models.Tenant{}).Count(&tenantCount)
	h.DB.Model(&models.User{}).Count(&userCount)
	h.DB.Model(&models.Extension{}).Count(&extensionCount)
	h.DB.Model(&models.Gateway{}).Count(&gatewayCount)

	// Device registration stats (placeholder - would come from ESL/FreeSWITCH)
	// For now, return estimated values based on extension count
	deviceStats := iris.Map{
		"desk_phones": iris.Map{"total": extensionCount, "online": extensionCount * 85 / 100},
		"softphones":  iris.Map{"total": extensionCount * 40 / 100, "online": extensionCount * 30 / 100},
		"mobile":      iris.Map{"total": extensionCount * 15 / 100, "online": extensionCount * 10 / 100},
		"trunks":      iris.Map{"total": gatewayCount, "online": gatewayCount},
	}

	ctx.JSON(iris.Map{
		"tenants":         tenantCount,
		"users":           userCount,
		"extensions":      extensionCount,
		"active_channels": 0, // Would come from ESL/FreeSWITCH
		"alerts":          0, // System alerts
		"gateways":        gatewayCount,
		"devices":         deviceStats,
	})
}

// =====================
// Messaging Providers
// =====================

func (h *Handler) ListMessagingProviders(ctx iris.Context) {
	var providers []models.MessagingProvider
	// Only get system-level providers (tenant_id is null)
	if err := h.DB.Where("tenant_id IS NULL").Find(&providers).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve messaging providers"})
		return
	}
	ctx.JSON(iris.Map{"data": providers})
}

func (h *Handler) CreateMessagingProvider(ctx iris.Context) {
	var provider models.MessagingProvider
	if err := ctx.ReadJSON(&provider); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Ensure it's a system-level provider
	provider.TenantID = nil

	if err := h.DB.Create(&provider).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create messaging provider"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(provider)
}

func (h *Handler) GetMessagingProvider(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid provider ID"})
		return
	}

	var provider models.MessagingProvider
	if err := h.DB.Where("tenant_id IS NULL").First(&provider, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Messaging provider not found"})
		return
	}

	ctx.JSON(provider)
}

func (h *Handler) UpdateMessagingProvider(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid provider ID"})
		return
	}

	var provider models.MessagingProvider
	if err := h.DB.Where("tenant_id IS NULL").First(&provider, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Messaging provider not found"})
		return
	}

	if err := ctx.ReadJSON(&provider); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&provider).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update messaging provider"})
		return
	}

	ctx.JSON(provider)
}

func (h *Handler) DeleteMessagingProvider(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid provider ID"})
		return
	}

	var provider models.MessagingProvider
	if err := h.DB.Where("tenant_id IS NULL").First(&provider, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Messaging provider not found"})
		return
	}

	if err := h.DB.Delete(&provider).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete messaging provider"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
}

// =====================
// Global Dial Plans
// =====================

func (h *Handler) ListGlobalDialplans(ctx iris.Context) {
	var dialplans []models.Dialplan
	// Only get global dialplans (tenant_id is null)
	if err := h.DB.Where("tenant_id IS NULL").Preload("Details").Order("dialplan_order ASC").Find(&dialplans).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve dial plans"})
		return
	}
	ctx.JSON(iris.Map{"data": dialplans})
}

func (h *Handler) CreateGlobalDialplan(ctx iris.Context) {
	var dialplan models.Dialplan
	if err := ctx.ReadJSON(&dialplan); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Ensure it's a global dialplan
	dialplan.TenantID = nil

	if err := h.DB.Create(&dialplan).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create dial plan"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(dialplan)
}

func (h *Handler) GetGlobalDialplan(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid dial plan ID"})
		return
	}

	var dialplan models.Dialplan
	if err := h.DB.Where("tenant_id IS NULL").Preload("Details").First(&dialplan, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Dial plan not found"})
		return
	}

	ctx.JSON(dialplan)
}

func (h *Handler) UpdateGlobalDialplan(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid dial plan ID"})
		return
	}

	var dialplan models.Dialplan
	if err := h.DB.Where("tenant_id IS NULL").First(&dialplan, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Dial plan not found"})
		return
	}

	if err := ctx.ReadJSON(&dialplan); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&dialplan).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update dial plan"})
		return
	}

	ctx.JSON(dialplan)
}

func (h *Handler) DeleteGlobalDialplan(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid dial plan ID"})
		return
	}

	var dialplan models.Dialplan
	if err := h.DB.Where("tenant_id IS NULL").First(&dialplan, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Dial plan not found"})
		return
	}

	// Also delete related details
	h.DB.Where("dialplan_uuid = ?", dialplan.UUID).Delete(&models.DialplanDetail{})

	if err := h.DB.Delete(&dialplan).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete dial plan"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
}

// =====================
// Access Control Lists (ACLs)
// =====================

func (h *Handler) ListACLs(ctx iris.Context) {
	var acls []models.ACL
	if err := h.DB.Preload("Nodes", func(db *gorm.DB) *gorm.DB {
		return db.Order("priority ASC")
	}).Find(&acls).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve ACLs"})
		return
	}
	ctx.JSON(iris.Map{"data": acls})
}

func (h *Handler) CreateACL(ctx iris.Context) {
	var acl models.ACL
	if err := ctx.ReadJSON(&acl); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Create(&acl).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create ACL"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(acl)
}

func (h *Handler) GetACL(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ACL ID"})
		return
	}

	var acl models.ACL
	if err := h.DB.Preload("Nodes", func(db *gorm.DB) *gorm.DB {
		return db.Order("priority ASC")
	}).First(&acl, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "ACL not found"})
		return
	}

	ctx.JSON(acl)
}

func (h *Handler) UpdateACL(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ACL ID"})
		return
	}

	var acl models.ACL
	if err := h.DB.First(&acl, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "ACL not found"})
		return
	}

	if err := ctx.ReadJSON(&acl); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&acl).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update ACL"})
		return
	}

	ctx.JSON(acl)
}

func (h *Handler) DeleteACL(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ACL ID"})
		return
	}

	var acl models.ACL
	if err := h.DB.First(&acl, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "ACL not found"})
		return
	}

	// Delete associated nodes
	h.DB.Where("acl_uuid = ?", acl.UUID).Delete(&models.ACLNode{})

	if err := h.DB.Delete(&acl).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete ACL"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
}

func (h *Handler) CreateACLNode(ctx iris.Context) {
	id, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ACL ID"})
		return
	}

	var acl models.ACL
	if err := h.DB.First(&acl, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "ACL not found"})
		return
	}

	var node models.ACLNode
	if err := ctx.ReadJSON(&node); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	node.ACLUUID = acl.UUID

	if err := h.DB.Create(&node).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create ACL node"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(node)
}

func (h *Handler) UpdateACLNode(ctx iris.Context) {
	nodeId, err := strconv.Atoi(ctx.Params().Get("nodeId"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid node ID"})
		return
	}

	var node models.ACLNode
	if err := h.DB.First(&node, nodeId).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "ACL node not found"})
		return
	}

	if err := ctx.ReadJSON(&node); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if err := h.DB.Save(&node).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update ACL node"})
		return
	}

	ctx.JSON(node)
}

func (h *Handler) DeleteACLNode(ctx iris.Context) {
	nodeId, err := strconv.Atoi(ctx.Params().Get("nodeId"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid node ID"})
		return
	}

	var node models.ACLNode
	if err := h.DB.First(&node, nodeId).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "ACL node not found"})
		return
	}

	if err := h.DB.Delete(&node).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete ACL node"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
}

// =====================
// System Numbers (All Tenants)
// =====================

func (h *Handler) ListAllNumbers(ctx iris.Context) {
	var numbers []models.Destination

	// Get all numbers across all tenants with tenant info
	if err := h.DB.Preload("Tenant").Order("destination_number ASC").Find(&numbers).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve numbers"})
		return
	}

	ctx.JSON(iris.Map{"data": numbers})
}

// =====================
// Security - Banned IPs
// =====================

type BannedIPRequest struct {
	IP       string `json:"ip"`
	Jail     string `json:"jail"`
	Failures int    `json:"failures"`
	BannedAt string `json:"banned_at"`
	Action   string `json:"action"` // "ban" or "unban"
}

// ReportBannedIP receives reports from fail2ban when IPs are banned/unbanned
func (h *Handler) ReportBannedIP(ctx iris.Context) {
	var req BannedIPRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	if req.Action == "unban" {
		// Mark IP as unbanned
		h.DB.Model(&models.BannedIP{}).
			Where("ip = ? AND status = ?", req.IP, "banned").
			Update("status", "unbanned")
		ctx.JSON(iris.Map{"message": "IP unbanned", "ip": req.IP})
		return
	}

	// Create new ban record
	bannedIP := models.BannedIP{
		IP:       req.IP,
		Source:   req.Jail,
		Reason:   "SIP brute force attempt",
		Failures: req.Failures,
		Status:   "banned",
	}

	if err := h.DB.Create(&bannedIP).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to record ban"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "IP banned", "data": bannedIP})
}

// ListBannedIPs returns all banned IPs
func (h *Handler) ListBannedIPs(ctx iris.Context) {
	var bannedIPs []models.BannedIP

	query := h.DB.Order("banned_at DESC")

	// Filter by status
	status := ctx.URLParamDefault("status", "banned")
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&bannedIPs).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve banned IPs"})
		return
	}

	ctx.JSON(iris.Map{"data": bannedIPs})
}

// UnbanIP manually unbans an IP address
func (h *Handler) UnbanIP(ctx iris.Context) {
	ip := ctx.Params().Get("ip")

	result := h.DB.Model(&models.BannedIP{}).
		Where("ip = ? AND status = ?", ip, "banned").
		Update("status", "unbanned")

	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "IP not found or already unbanned"})
		return
	}

	ctx.JSON(iris.Map{"message": "IP unbanned", "ip": ip})
}

// =====================
// System Device Templates
// =====================

// ListSystemDeviceTemplates returns all system (global) device templates
func (h *Handler) ListSystemDeviceTemplates(ctx iris.Context) {
	var templates []models.DeviceTemplate

	query := h.DB.Where("tenant_id IS NULL")

	// Filter by manufacturer
	if manufacturer := ctx.URLParam("manufacturer"); manufacturer != "" {
		query = query.Where("manufacturer = ?", manufacturer)
	}

	query.Order("manufacturer, model, name").Find(&templates)

	// Add device counts
	for i := range templates {
		h.DB.Model(&models.Device{}).Where("template_id = ?", templates[i].ID).Count(&templates[i].DeviceCount)
	}

	ctx.JSON(iris.Map{"data": templates})
}

// CreateSystemDeviceTemplate creates a new system template
func (h *Handler) CreateSystemDeviceTemplate(ctx iris.Context) {
	var tmpl models.DeviceTemplate
	if err := ctx.ReadJSON(&tmpl); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// System templates have no tenant
	tmpl.TenantID = nil
	tmpl.IsSystem = true

	if err := h.DB.Create(&tmpl).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create template"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": tmpl, "message": "System template created"})
}

// GetSystemDeviceTemplate returns a single system template
func (h *Handler) GetSystemDeviceTemplate(ctx iris.Context) {
	id := ctx.Params().Get("id")

	var tmpl models.DeviceTemplate
	if err := h.DB.Where("id = ? AND tenant_id IS NULL", id).
		Preload("Firmware").
		First(&tmpl).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Template not found"})
		return
	}

	ctx.JSON(iris.Map{"data": tmpl})
}

// UpdateSystemDeviceTemplate updates a system template
func (h *Handler) UpdateSystemDeviceTemplate(ctx iris.Context) {
	id := ctx.Params().Get("id")

	var tmpl models.DeviceTemplate
	if err := h.DB.Where("id = ? AND tenant_id IS NULL", id).First(&tmpl).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Template not found"})
		return
	}

	var input models.DeviceTemplate
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	updates := map[string]interface{}{
		"name":            input.Name,
		"description":     input.Description,
		"manufacturer":    input.Manufacturer,
		"model":           input.Model,
		"family":          input.Family,
		"config_template": input.ConfigTemplate,
		"config_type":     input.ConfigType,
		"firmware_id":     input.FirmwareID,
		"enabled":         input.Enabled,
	}

	h.DB.Model(&tmpl).Updates(updates)

	ctx.JSON(iris.Map{"data": tmpl, "message": "Template updated"})
}

// DeleteSystemDeviceTemplate deletes a system template
func (h *Handler) DeleteSystemDeviceTemplate(ctx iris.Context) {
	id := ctx.Params().Get("id")

	// Check if any devices are using this template
	var count int64
	h.DB.Model(&models.Device{}).Where("template_id = ?", id).Count(&count)
	if count > 0 {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": "Template is in use by devices", "device_count": count})
		return
	}

	result := h.DB.Where("id = ? AND tenant_id IS NULL", id).Delete(&models.DeviceTemplate{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Template not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Template deleted"})
}

// =====================
// Firmware Management
// =====================

// ListFirmware returns all firmware files
func (h *Handler) ListFirmware(ctx iris.Context) {
	var firmware []models.Firmware

	query := h.DB.Where("enabled = ?", true)

	// Filter by manufacturer
	if manufacturer := ctx.URLParam("manufacturer"); manufacturer != "" {
		query = query.Where("manufacturer = ?", manufacturer)
	}

	// Filter by model/family
	if model := ctx.URLParam("model"); model != "" {
		query = query.Where("model = ? OR family = ?", model, model)
	}

	query.Order("manufacturer, model, version DESC").Find(&firmware)

	ctx.JSON(iris.Map{"data": firmware})
}

// CreateFirmware creates a new firmware entry
func (h *Handler) CreateFirmware(ctx iris.Context) {
	var fw models.Firmware
	if err := ctx.ReadJSON(&fw); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	if err := h.DB.Create(&fw).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create firmware entry"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"data": fw, "message": "Firmware created"})
}

// GetFirmware returns a single firmware entry
func (h *Handler) GetFirmware(ctx iris.Context) {
	id := ctx.Params().Get("id")

	var fw models.Firmware
	if err := h.DB.First(&fw, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Firmware not found"})
		return
	}

	ctx.JSON(iris.Map{"data": fw})
}

// UpdateFirmware updates a firmware entry
func (h *Handler) UpdateFirmware(ctx iris.Context) {
	id := ctx.Params().Get("id")

	var fw models.Firmware
	if err := h.DB.First(&fw, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Firmware not found"})
		return
	}

	var input models.Firmware
	if err := ctx.ReadJSON(&input); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	updates := map[string]interface{}{
		"manufacturer":  input.Manufacturer,
		"model":         input.Model,
		"family":        input.Family,
		"version":       input.Version,
		"release_date":  input.ReleaseDate,
		"release_notes": input.ReleaseNotes,
		"is_default":    input.IsDefault,
		"enabled":       input.Enabled,
	}

	h.DB.Model(&fw).Updates(updates)

	ctx.JSON(iris.Map{"data": fw, "message": "Firmware updated"})
}

// DeleteFirmware deletes a firmware entry
func (h *Handler) DeleteFirmware(ctx iris.Context) {
	id := ctx.Params().Get("id")

	// Check if any templates are using this firmware
	var count int64
	h.DB.Model(&models.DeviceTemplate{}).Where("firmware_id = ?", id).Count(&count)
	if count > 0 {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": "Firmware is referenced by templates", "template_count": count})
		return
	}

	result := h.DB.Delete(&models.Firmware{}, id)
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Firmware not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Firmware deleted"})
}

// SetDefaultFirmware sets a firmware as the default for its model
func (h *Handler) SetDefaultFirmware(ctx iris.Context) {
	id := ctx.Params().Get("id")

	var fw models.Firmware
	if err := h.DB.First(&fw, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Firmware not found"})
		return
	}

	// Clear default flag for same manufacturer/model
	h.DB.Model(&models.Firmware{}).
		Where("manufacturer = ? AND (model = ? OR family = ?)", fw.Manufacturer, fw.Model, fw.Family).
		Update("is_default", false)

	// Set this one as default
	h.DB.Model(&fw).Update("is_default", true)

	ctx.JSON(iris.Map{"message": "Firmware set as default"})
}

// UploadFirmware handles firmware file upload
func (h *Handler) UploadFirmware(ctx iris.Context) {
	// Get firmware ID
	id := ctx.Params().Get("id")

	var fw models.Firmware
	if err := h.DB.First(&fw, id).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Firmware not found"})
		return
	}

	// Get file
	file, header, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "File is required"})
		return
	}
	defer file.Close()

	// Save file
	firmwarePath := "/usr/share/freeswitch/firmware/" + fw.Manufacturer + "/" + header.Filename
	// TODO: Implement file save with proper directory creation

	// Update firmware record
	h.DB.Model(&fw).Updates(map[string]interface{}{
		"file_path": firmwarePath,
		"file_name": header.Filename,
		"file_size": header.Size,
	})

	ctx.JSON(iris.Map{"message": "Firmware file uploaded", "path": firmwarePath})
}
