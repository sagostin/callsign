package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/kataras/iris/v12"
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
	ctx.JSON(iris.Map{"data": profiles})
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
	// Get counts
	var userCount, tenantCount int64
	h.DB.Model(&models.User{}).Count(&userCount)
	h.DB.Model(&models.Tenant{}).Count(&tenantCount)

	ctx.JSON(iris.Map{
		"users":   userCount,
		"tenants": tenantCount,
	})
}
