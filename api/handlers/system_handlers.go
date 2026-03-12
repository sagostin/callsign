package handlers

import (
	"callsign/handlers/freeswitch"
	"callsign/middleware"
	"callsign/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// =====================
// Tenants
// =====================

func (h *Handler) ListTenants(c *fiber.Ctx) error {
	var tenants []models.Tenant
	if err := h.DB.Preload("Profile").Find(&tenants).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve tenants"})
	}
	return c.JSON(fiber.Map{"data": tenants})
}

func (h *Handler) CreateTenant(c *fiber.Ctx) error {
	var input struct {
		Name        string `json:"name"`
		Domain      string `json:"domain"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		ProfileID   *uint  `json:"profile_id"`
		AdminEmail  string `json:"admin_email"`
		Settings    string `json:"settings"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	tenant := models.Tenant{
		Name:        input.Name,
		Domain:      input.Domain,
		Description: input.Description,
		Enabled:     input.Enabled,
		ProfileID:   input.ProfileID,
		Settings:    input.Settings,
	}

	if err := h.DB.Create(&tenant).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create tenant"})
	}

	// Provision FreeSWITCH resources (feature codes, park slots, dialplans)
	if err := models.ProvisionTenant(h.DB, &tenant); err != nil {
		log.WithError(err).WithField("tenant_id", tenant.ID).Warn("Tenant created but provisioning failed")
	}

	// Reload FreeSWITCH so new dialplan context is available
	h.reloadXML()

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": tenant})
}

func (h *Handler) GetTenant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tenant ID"})
	}

	var tenant models.Tenant
	if err := h.DB.Preload("Profile").First(&tenant, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	return c.JSON(tenant)
}

func (h *Handler) UpdateTenant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tenant ID"})
	}

	var tenant models.Tenant
	if err := h.DB.First(&tenant, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	if err := c.BodyParser(&tenant); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	tenant.ID = uint(id)
	if err := h.DB.Save(&tenant).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update tenant"})
	}

	return c.JSON(tenant)
}

func (h *Handler) DeleteTenant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tenant ID"})
	}

	// Deprovision FreeSWITCH resources before deleting tenant
	if err := models.DeprovisionTenant(h.DB, uint(id)); err != nil {
		log.WithError(err).WithField("tenant_id", id).Warn("Tenant deprovisioning failed")
	}

	if err := h.DB.Delete(&models.Tenant{}, id).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete tenant"})
	}

	// Reload FreeSWITCH so removed dialplans/directory entries are cleared
	h.reloadXML()

	return c.JSON(fiber.Map{"message": "Tenant deleted successfully"})
}

// =====================
// Tenant Profiles
// =====================

func (h *Handler) ListTenantProfiles(c *fiber.Ctx) error {
	var profiles []models.TenantProfile
	if err := h.DB.Find(&profiles).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve tenant profiles"})
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

	return c.JSON(fiber.Map{"data": result})
}

func (h *Handler) CreateTenantProfile(c *fiber.Ctx) error {
	var profile models.TenantProfile
	if err := c.BodyParser(&profile); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Create(&profile).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create tenant profile"})
	}

	return c.Status(http.StatusCreated).JSON(profile)
}

func (h *Handler) GetTenantProfile(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid profile ID"})
	}

	var profile models.TenantProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Profile not found"})
	}

	return c.JSON(profile)
}

func (h *Handler) UpdateTenantProfile(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid profile ID"})
	}

	var profile models.TenantProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Profile not found"})
	}

	if err := c.BodyParser(&profile); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	profile.ID = uint(id)
	if err := h.DB.Save(&profile).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update profile"})
	}

	return c.JSON(profile)
}

func (h *Handler) DeleteTenantProfile(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid profile ID"})
	}

	if err := h.DB.Delete(&models.TenantProfile{}, id).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete profile"})
	}

	return c.JSON(fiber.Map{"message": "Profile deleted successfully"})
}

// =====================
// Users (System Admin)
// =====================

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)

	var users []models.User
	query := h.DB
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Find(&users).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve users"})
	}
	return c.JSON(fiber.Map{"data": users})
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Hash password
	if password := c.FormValue("password"); password != "" {
		if err := user.SetPassword(password); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set password"})
		}
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(http.StatusCreated).JSON(user)
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	user.ID = uint(id)
	if err := h.DB.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
	}

	return c.JSON(user)
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	if err := h.DB.Delete(&models.User{}, id).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete user"})
	}

	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// =====================
// Gateways
// =====================

func (h *Handler) ListGateways(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)

	var gateways []models.Gateway
	query := h.DB
	if tenantID > 0 {
		// Tenant admin sees their gateways + system-wide gateways
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantID)
	}
	// System admin sees all gateways

	if err := query.Order("gateway_name").Find(&gateways).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve gateways"})
	}
	return c.JSON(fiber.Map{"data": gateways})
}

func (h *Handler) CreateGateway(c *fiber.Ctx) error {
	var gateway models.Gateway
	if err := c.BodyParser(&gateway); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Password has json:"-" so BodyParser skips it. Extract manually.
	var raw map[string]interface{}
	if err := json.Unmarshal(c.Body(), &raw); err == nil {
		if pw, ok := raw["password"].(string); ok && pw != "" {
			gateway.Password = pw
		}
	}

	// If tenant admin, scope to their tenant
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID > 0 {
		gateway.TenantID = &tenantID
	}

	if err := h.DB.Create(&gateway).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create gateway"})
	}

	// Trigger FreeSWITCH reload so new gateway is picked up
	h.reloadSofia("internal")
	return c.Status(http.StatusCreated).JSON(gateway)
}

func (h *Handler) GetGateway(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid gateway ID"})
	}

	tenantID := middleware.GetScopedTenantID(c)

	var gateway models.Gateway
	query := h.DB
	if tenantID > 0 {
		query = query.Where("(tenant_id = ? OR tenant_id IS NULL)", tenantID)
	}

	if err := query.First(&gateway, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Gateway not found"})
	}

	return c.JSON(gateway)
}

func (h *Handler) UpdateGateway(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid gateway ID"})
	}

	tenantID := middleware.GetScopedTenantID(c)

	var gateway models.Gateway
	query := h.DB
	if tenantID > 0 {
		// Tenant admins can only update their own gateways, not system-wide ones
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&gateway, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Gateway not found"})
	}

	// Stash existing password before BodyParser overwrites (Password has json:"-")
	existingPassword := gateway.Password

	if err := c.BodyParser(&gateway); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Password has json:"-" so BodyParser skips it. Extract manually from raw body.
	var raw map[string]interface{}
	if err := json.Unmarshal(c.Body(), &raw); err == nil {
		if pw, ok := raw["password"].(string); ok && pw != "" {
			gateway.Password = pw
		} else {
			gateway.Password = existingPassword // Preserve existing password
		}
	} else {
		gateway.Password = existingPassword
	}

	gateway.ID = uint(id)
	if tenantID > 0 {
		gateway.TenantID = &tenantID // Preserve tenant ownership
	}

	if err := h.DB.Save(&gateway).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update gateway"})
	}

	// Trigger FreeSWITCH reload so gateway changes are picked up
	h.reloadSofia("internal")
	return c.JSON(gateway)
}

func (h *Handler) DeleteGateway(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid gateway ID"})
	}

	tenantID := middleware.GetScopedTenantID(c)

	var gateway models.Gateway
	query := h.DB
	if tenantID > 0 {
		// Tenant admins can only delete their own gateways
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&gateway, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Gateway not found"})
	}

	if err := h.DB.Delete(&gateway).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete gateway"})
	}

	// Trigger FreeSWITCH reload so gateway removal is picked up
	h.reloadSofia("internal")
	return c.JSON(fiber.Map{"message": "Gateway deleted successfully"})
}

// GetGatewayStatus returns live gateway status from FreeSWITCH
func (h *Handler) GetGatewayStatus(c *fiber.Ctx) error {
	// Get all gateways from the database
	var gateways []models.Gateway
	if err := h.DB.Find(&gateways).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve gateways"})
	}

	// Get ESL client from context if available
	eslClient := c.Locals("esl_client")
	if eslClient == nil {
		// Return basic status from database if ESL not available
		statusMap := make(map[string]interface{})
		for _, gw := range gateways {
			statusMap[gw.GatewayName] = map[string]interface{}{
				"state":   "UNKNOWN",
				"enabled": gw.Enabled,
			}
		}
		return c.JSON(fiber.Map{"data": statusMap, "esl_connected": false})
	}

	// Query FreeSWITCH for real-time gateway status
	statusMap := make(map[string]interface{})
	for _, gw := range gateways {
		gwResult := map[string]interface{}{
			"state":    "UNKNOWN",
			"enabled":  gw.Enabled,
			"register": gw.Register,
		}

		if h.ESLManager != nil && h.ESLManager.IsConnected() {
			result, err := h.ESLManager.API(fmt.Sprintf("sofia status gateway %s", gw.GatewayName))
			if err == nil && !strings.Contains(result, "-ERR") {
				// Parse state from output
				for _, line := range strings.Split(result, "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "State") {
						parts := strings.Fields(line)
						if len(parts) >= 2 {
							gwResult["state"] = parts[len(parts)-1]
						}
					}
					if strings.HasPrefix(line, "Status") {
						parts := strings.Fields(line)
						if len(parts) >= 2 {
							gwResult["status_detail"] = strings.Join(parts[1:], " ")
						}
					}
					if strings.HasPrefix(line, "Ping-Time") {
						parts := strings.Fields(line)
						if len(parts) >= 2 {
							gwResult["ping_time"] = parts[len(parts)-1]
						}
					}
				}
			}
		} else {
			// Fallback to DB-based heuristic
			if !gw.Enabled {
				gwResult["state"] = "DISABLED"
			} else if gw.Register {
				gwResult["state"] = "TRYING"
			} else {
				gwResult["state"] = "ACTIVE"
			}
		}

		statusMap[gw.GatewayName] = gwResult
	}

	eslConnected := h.ESLManager != nil && h.ESLManager.IsConnected()
	return c.JSON(fiber.Map{"data": statusMap, "esl_connected": eslConnected})
}

// =====================
// Bridges
// =====================

func (h *Handler) ListBridges(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []interface{}{}, "message": "Not implemented"})
}

func (h *Handler) CreateBridge(c *fiber.Ctx) error {
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error": "Not implemented"})
}

func (h *Handler) GetBridge(c *fiber.Ctx) error {
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error": "Not implemented"})
}

func (h *Handler) UpdateBridge(c *fiber.Ctx) error {
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error": "Not implemented"})
}

func (h *Handler) DeleteBridge(c *fiber.Ctx) error {
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error": "Not implemented"})
}

// =====================
// SIP Profiles
// =====================
// SIP profiles MUST be written to disk as XML files because:
// - Sofia loads them via X-PRE-PROCESS from the sip_profiles directory
// - Gateways are served dynamically via directory (purpose=gateways)
//   when profiles have parse=true on their domain definitions

// writeSIPProfileToDisk writes a SIP profile XML file to the sip_profiles directory
func (h *Handler) writeSIPProfileToDisk(profile *models.SIPProfile) error {
	writer := freeswitch.NewProfileWriter(h.Config.SIPProfilesPath)

	// If settings/domains are not loaded, fetch them
	var settings []models.SIPProfileSetting
	var domains []models.SIPProfileDomain

	if len(profile.Settings) > 0 {
		settings = profile.Settings
	} else {
		h.DB.Where("sip_profile_uuid = ?", profile.UUID).Find(&settings)
	}

	if len(profile.Domains) > 0 {
		domains = profile.Domains
	} else {
		h.DB.Where("sip_profile_uuid = ?", profile.UUID).Find(&domains)
	}

	return writer.WriteProfile(profile, settings, domains)
}

func (h *Handler) ListSIPProfiles(c *fiber.Ctx) error {
	// Dedup cleanup: remove duplicate settings (same profile + setting_name)
	// This fixes any historical duplication from GORM auto-cascade creates
	h.DB.Exec(`
		DELETE FROM sip_profile_settings
		WHERE id NOT IN (
			SELECT MIN(id) FROM sip_profile_settings
			GROUP BY sip_profile_uuid, setting_name
		)
	`)

	var profiles []models.SIPProfile
	if err := h.DB.Preload("Settings").Preload("Domains").Order("profile_name").Find(&profiles).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve SIP profiles"})
	}
	return c.JSON(fiber.Map{"data": profiles})
}

// SyncSIPProfiles imports SIP profiles from disk XML files that don't exist in DB.
// DB is the source of truth — existing profiles are NOT overwritten.
// Only new profiles found on disk (not yet in DB) are imported.
func (h *Handler) SyncSIPProfiles(c *fiber.Ctx) error {
	profilesPath := h.Config.SIPProfilesPath
	if profilesPath == "" {
		profilesPath = "/etc/freeswitch/sip_profiles"
	}

	importer := freeswitch.NewProfileImporter(profilesPath, h.DB)

	log.WithField("path", profilesPath).Info("Syncing SIP profiles from disk")

	// Do NOT overwrite existing profiles — DB is source of truth
	if err := importer.SyncProfiles(false); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to sync profiles: " + err.Error()})
	}

	// Return updated list
	var profiles []models.SIPProfile
	h.DB.Preload("Settings").Preload("Domains").Order("profile_name").Find(&profiles)

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Sync complete. Found %d profiles in database.", len(profiles)),
		"data":    profiles,
		"path":    profilesPath,
	})
}

func (h *Handler) CreateSIPProfile(c *fiber.Ctx) error {
	var input models.SIPProfile
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Stash settings/domains from payload before creating profile
	incomingSettings := input.Settings
	incomingDomains := input.Domains
	input.Settings = nil
	input.Domains = nil

	// Check for soft-deleted profile with same name and permanently delete it
	var existingSoftDeleted models.SIPProfile
	if err := h.DB.Unscoped().Where("profile_name = ? AND deleted_at IS NOT NULL", input.ProfileName).First(&existingSoftDeleted).Error; err == nil {
		h.DB.Unscoped().Where("sip_profile_uuid = ?", existingSoftDeleted.UUID).Delete(&models.SIPProfileSetting{})
		h.DB.Unscoped().Where("sip_profile_uuid = ?", existingSoftDeleted.UUID).Delete(&models.SIPProfileDomain{})
		h.DB.Unscoped().Delete(&existingSoftDeleted)
	}

	// Create profile WITHOUT associations (prevents GORM auto-cascade duplication)
	if err := h.DB.Omit("Settings", "Domains").Create(&input).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create SIP profile"})
	}

	// Create settings explicitly
	if len(incomingSettings) > 0 {
		for i := range incomingSettings {
			incomingSettings[i].SIPProfileUUID = input.UUID
			incomingSettings[i].ID = 0
		}
		h.DB.Create(&incomingSettings)
		input.Settings = incomingSettings
	} else {
		// Add defaults if no settings provided
		var defaultSettings []models.SIPProfileSetting
		if input.ProfileName == "external" {
			defaultSettings = models.DefaultExternalProfileSettings()
		} else {
			defaultSettings = models.DefaultInternalProfileSettings()
		}
		for i := range defaultSettings {
			defaultSettings[i].SIPProfileUUID = input.UUID
		}
		if len(defaultSettings) > 0 {
			h.DB.Create(&defaultSettings)
			input.Settings = defaultSettings
		}
	}

	// Create domains explicitly
	if len(incomingDomains) > 0 {
		for i := range incomingDomains {
			incomingDomains[i].SIPProfileUUID = input.UUID
			incomingDomains[i].ID = 0
		}
		h.DB.Create(&incomingDomains)
		input.Domains = incomingDomains
	}

	// Write profile XML to disk (required for Sofia X-PRE-PROCESS)
	if err := h.writeSIPProfileToDisk(&input); err != nil {
		log.Printf("Failed to write SIP profile to disk: %v", err)
	}

	return c.Status(http.StatusCreated).JSON(input)
}

func (h *Handler) GetSIPProfile(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid profile ID"})
	}

	var profile models.SIPProfile
	if err := h.DB.Preload("Settings").Preload("Domains").First(&profile, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "SIP profile not found"})
	}

	return c.JSON(profile)
}

func (h *Handler) UpdateSIPProfile(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid profile ID"})
	}

	var profile models.SIPProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "SIP profile not found"})
	}

	var input models.SIPProfile
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Use a transaction to ensure atomicity
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		input.ID = uint(id)
		input.UUID = profile.UUID // Ensure UUID doesn't change

		// Update profile fields ONLY, ignoring associations to prevent auto-save duplication
		if err := tx.Omit("Settings", "Domains").Save(&input).Error; err != nil {
			return err
		}

		// Update settings if provided
		// Note: We always replace all settings if the list is present
		if input.Settings != nil {
			// Delete existing settings using struct for correct column name resolution
			if err := tx.Where(&models.SIPProfileSetting{SIPProfileUUID: profile.UUID}).Delete(&models.SIPProfileSetting{}).Error; err != nil {
				return err
			}

			// Create new settings
			if len(input.Settings) > 0 {
				for i := range input.Settings {
					input.Settings[i].SIPProfileUUID = profile.UUID
					// Ensure ID is zero to force create
					input.Settings[i].ID = 0
				}
				if err := tx.Create(&input.Settings).Error; err != nil {
					return err
				}
			}
		}

		// Update domains if provided
		if input.Domains != nil {
			// Delete existing domains using struct for correct column name resolution
			if err := tx.Where(&models.SIPProfileDomain{SIPProfileUUID: profile.UUID}).Delete(&models.SIPProfileDomain{}).Error; err != nil {
				return err
			}

			// Create new domains
			if len(input.Domains) > 0 {
				for i := range input.Domains {
					input.Domains[i].SIPProfileUUID = profile.UUID
					input.Domains[i].ID = 0
				}
				if err := tx.Create(&input.Domains).Error; err != nil {
					return err
				}
			}
		}

		// Update the profile object with the new associations for response
		profile = input
		return nil
	})

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update SIP profile: " + err.Error()})
	}

	// Write updated profile XML to disk (required for Sofia X-PRE-PROCESS)
	if err := h.writeSIPProfileToDisk(&profile); err != nil {
		log.Printf("Failed to write SIP profile to disk: %v", err)
	}

	return c.JSON(profile)
}

func (h *Handler) DeleteSIPProfile(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid profile ID"})
	}

	var profile models.SIPProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "SIP profile not found"})
	}

	// Prevent deletion of system profiles (internal/external)
	if freeswitch.IsSystemProfile(profile.ProfileName) {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Cannot delete system profile: " + profile.ProfileName})
	}

	// Check for associated gateways
	var gatewayCount int64
	h.DB.Model(&models.Gateway{}).Where("sip_profile_uuid = ?", profile.UUID).Count(&gatewayCount)
	if gatewayCount > 0 {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error":   "Cannot delete profile",
			"message": fmt.Sprintf("Profile has %d associated gateway(s). Remove them first.", gatewayCount),
		})
	}

	// Check for associated domains
	var domainCount int64
	h.DB.Model(&models.SIPProfileDomain{}).Where("sip_profile_uuid = ?", profile.UUID).Count(&domainCount)
	if domainCount > 0 {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error":   "Cannot delete profile",
			"message": fmt.Sprintf("Profile has %d associated domain(s). Remove them first.", domainCount),
		})
	}

	// Delete related settings (safe to delete)
	h.DB.Where("sip_profile_uuid = ?", profile.UUID).Delete(&models.SIPProfileSetting{})

	if err := h.DB.Delete(&profile).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete SIP profile"})
	}

	// Delete profile XML file from disk
	writer := freeswitch.NewProfileWriter(h.Config.SIPProfilesPath)
	if err := writer.DeleteProfile(profile.ProfileName); err != nil {
		log.Printf("Failed to delete SIP profile from disk: %v", err)
	}

	return c.JSON(fiber.Map{"message": "SIP profile deleted successfully"})
}

// =====================
// Sofia Control Commands
// =====================

// GetSofiaStatus returns the status of all Sofia profiles from FreeSWITCH
func (h *Handler) GetSofiaStatus(c *fiber.Ctx) error {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH ESL not connected"})
	}

	result, err := h.ESLManager.API("sofia status")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get sofia status: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": result})
}

// GetSofiaProfileStatus returns the status of a specific Sofia profile
func (h *Handler) GetSofiaProfileStatus(c *fiber.Ctx) error {
	profileName := c.Params("name")
	if profileName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Profile name required"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH ESL not connected"})
	}

	result, err := h.ESLManager.API("sofia status profile " + profileName)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get profile status: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": result, "profile": profileName})
}

// GetSofiaProfileRegistrations returns the registrations for a specific Sofia profile
func (h *Handler) GetSofiaProfileRegistrations(c *fiber.Ctx) error {
	profileName := c.Params("name")
	if profileName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Profile name required"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH ESL not connected"})
	}

	result, err := h.ESLManager.API("sofia status profile " + profileName + " reg")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get registrations: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": result, "profile": profileName})
}

// GetSofiaGatewayStatus returns gateway status for a profile
func (h *Handler) GetSofiaGatewayStatus(c *fiber.Ctx) error {
	profileName := c.Params("name")
	if profileName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Profile name required"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH ESL not connected"})
	}

	result, err := h.ESLManager.API("sofia status gateway")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get gateway status: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": result})
}

// RestartSofiaProfile restarts a Sofia profile
// Uses BgAPI to avoid blocking — sofia restart can take 10-30s and may
// disrupt the ESL connection, causing synchronous API() calls to hang/502.
func (h *Handler) RestartSofiaProfile(c *fiber.Ctx) error {
	profileName := c.Params("name")
	if profileName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Profile name required"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH ESL not connected"})
	}

	jobUUID, err := h.ESLManager.BgAPI("sofia profile " + profileName + " restart")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to restart profile: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":  "Profile restart command sent",
		"data":     "Restart command queued (background job: " + jobUUID + ")",
		"profile":  profileName,
		"job_uuid": jobUUID,
	})
}

// StartSofiaProfile starts a stopped Sofia profile
// Uses BgAPI to avoid blocking — sofia start can take several seconds.
func (h *Handler) StartSofiaProfile(c *fiber.Ctx) error {
	profileName := c.Params("name")
	if profileName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Profile name required"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH ESL not connected"})
	}

	jobUUID, err := h.ESLManager.BgAPI("sofia profile " + profileName + " start")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start profile: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":  "Profile start command sent",
		"data":     "Start command queued (background job: " + jobUUID + ")",
		"profile":  profileName,
		"job_uuid": jobUUID,
	})
}

// StopSofiaProfile stops a running Sofia profile
// Uses BgAPI to avoid blocking — sofia stop can take several seconds.
func (h *Handler) StopSofiaProfile(c *fiber.Ctx) error {
	profileName := c.Params("name")
	if profileName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Profile name required"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH ESL not connected"})
	}

	jobUUID, err := h.ESLManager.BgAPI("sofia profile " + profileName + " stop")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to stop profile: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":  "Profile stop command sent",
		"data":     "Stop command queued (background job: " + jobUUID + ")",
		"profile":  profileName,
		"job_uuid": jobUUID,
	})
}

// ReloadSofiaXML reloads FreeSWITCH XML configuration
func (h *Handler) ReloadSofiaXML(c *fiber.Ctx) error {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH ESL not connected"})
	}

	result, err := h.ESLManager.API("reloadxml")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reload XML: " + err.Error()})
	}

	return c.JSON(fiber.Map{"message": "XML configuration reloaded", "data": result})
}

// =====================
// System Settings & Status
// =====================

func (h *Handler) GetSystemSettings(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"server_host":     h.Config.ServerHost,
		"server_port":     h.Config.ServerPort,
		"db_host":         h.Config.DBHost,
		"freeswitch_host": h.Config.FreeSwitchHost,
	})
}

func (h *Handler) UpdateSystemSettings(c *fiber.Ctx) error {
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error": "Not implemented"})
}

func (h *Handler) GetSystemLogs(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []interface{}{}, "message": "Not implemented"})
}

func (h *Handler) GetSystemStatus(c *fiber.Ctx) error {
	status := fiber.Map{
		"status":   "operational",
		"database": "connected",
	}

	// Add FreeSWITCH status if ESL manager is available
	if h.ESLManager != nil {
		status["freeswitch"] = h.ESLManager.FreeSwitchStatus()
	} else {
		status["freeswitch"] = fiber.Map{
			"esl_connected": false,
			"esl_running":   false,
			"message":       "ESL manager not initialized",
		}
	}

	return c.JSON(status)
}

func (h *Handler) GetSystemStats(c *fiber.Ctx) error {
	// Get database counts
	var tenantCount, userCount, extensionCount, gatewayCount int64
	h.DB.Model(&models.Tenant{}).Count(&tenantCount)
	h.DB.Model(&models.User{}).Count(&userCount)
	h.DB.Model(&models.Extension{}).Count(&extensionCount)
	h.DB.Model(&models.Gateway{}).Count(&gatewayCount)

	// Live stats from FreeSWITCH ESL
	activeChannels := int64(0)
	eslConnected := h.ESLManager != nil && h.ESLManager.IsConnected()

	// Registration counters
	totalRegs := int64(0)
	trunkOnline := int64(0)
	trunkTotal := gatewayCount

	if eslConnected {
		// Get active channel count
		if result, err := h.ESLManager.API("show channels count"); err == nil {
			// Output: "N total." — parse the number
			resultStr := strings.TrimSpace(result)
			lines := strings.Split(resultStr, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "total") {
					parts := strings.Fields(line)
					if len(parts) > 0 {
						if n, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
							activeChannels = n
						}
					}
				}
			}
		}

		// Get registration counts from sofia status
		if result, err := h.ESLManager.API("sofia status"); err == nil {
			lines := strings.Split(result, "\n")
			for _, line := range lines {
				// Sofia status lines contain RUNNING and registration count
				if strings.Contains(line, "RUNNING") {
					fields := strings.Fields(line)
					// Typical line: "internal-ipv4  sip:mod_sofia@... RUNNING (0)"
					// The last field in parens is the registration count
					for _, f := range fields {
						if strings.HasPrefix(f, "(") && strings.HasSuffix(f, ")") {
							numStr := strings.Trim(f, "()")
							if n, err := strconv.ParseInt(numStr, 10, 64); err == nil {
								totalRegs += n
							}
						}
					}
				}
			}
		}

		// Get gateway status
		if result, err := h.ESLManager.API("sofia status gateway"); err == nil {
			lines := strings.Split(result, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "REGED") || strings.Contains(line, "NOREG") {
					if strings.Contains(line, "REGED") {
						trunkOnline++
					}
				}
			}
		}
	}

	// Count devices from DB (registered client registrations)
	var deskPhones, softphones, mobileApps int64
	h.DB.Model(&models.ClientRegistration{}).Where("endpoint_type = ? AND enabled = true", "desk_phone").Count(&deskPhones)
	h.DB.Model(&models.ClientRegistration{}).Where("endpoint_type = ? AND enabled = true", "web_client").Count(&softphones)
	h.DB.Model(&models.ClientRegistration{}).Where("endpoint_type = ? AND enabled = true", "mobile_app").Count(&mobileApps)

	deviceStats := fiber.Map{
		"desk_phones": fiber.Map{"total": deskPhones, "online": 0},
		"softphones":  fiber.Map{"total": softphones, "online": 0},
		"mobile":      fiber.Map{"total": mobileApps, "online": 0},
		"trunks":      fiber.Map{"total": trunkTotal, "online": trunkOnline},
	}

	// If ESL connected, set online counts from total registrations (best-effort split)
	if eslConnected && totalRegs > 0 {
		totalDevices := deskPhones + softphones + mobileApps
		if totalDevices > 0 {
			// Proportional split of actual registrations across device types
			deviceStats["desk_phones"] = fiber.Map{"total": deskPhones, "online": totalRegs * deskPhones / totalDevices}
			deviceStats["softphones"] = fiber.Map{"total": softphones, "online": totalRegs * softphones / totalDevices}
			deviceStats["mobile"] = fiber.Map{"total": mobileApps, "online": totalRegs * mobileApps / totalDevices}
		}
	}

	// System alerts — count recent error-level log entries
	var alertCount int64
	h.DB.Model(&models.AuditLog{}).Where("action = 'error' AND created_at > NOW() - INTERVAL '24 hours'").Count(&alertCount)

	return c.JSON(fiber.Map{
		"tenants":         tenantCount,
		"users":           userCount,
		"extensions":      extensionCount,
		"active_channels": activeChannels,
		"alerts":          alertCount,
		"gateways":        gatewayCount,
		"devices":         deviceStats,
		"esl_connected":   eslConnected,
		"registrations":   totalRegs,
	})
}

// =====================
// Messaging Providers
// =====================

func (h *Handler) ListMessagingProviders(c *fiber.Ctx) error {
	var providers []models.MessagingProvider
	// Only get system-level providers (tenant_id is null)
	if err := h.DB.Where("tenant_id IS NULL").Find(&providers).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve messaging providers"})
	}
	return c.JSON(fiber.Map{"data": providers})
}

func (h *Handler) CreateMessagingProvider(c *fiber.Ctx) error {
	var provider models.MessagingProvider
	if err := c.BodyParser(&provider); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// AuthToken and WebhookSecret have json:"-" so BodyParser skips them.
	var raw map[string]interface{}
	if err := json.Unmarshal(c.Body(), &raw); err == nil {
		if token, ok := raw["auth_token"].(string); ok && token != "" {
			provider.AuthToken = token
		}
		if secret, ok := raw["webhook_secret"].(string); ok && secret != "" {
			provider.WebhookSecret = secret
		}
	}

	// Ensure it's a system-level provider
	provider.TenantID = nil

	if err := h.DB.Create(&provider).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create messaging provider"})
	}

	return c.Status(http.StatusCreated).JSON(provider)
}

func (h *Handler) GetMessagingProvider(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid provider ID"})
	}

	var provider models.MessagingProvider
	if err := h.DB.Where("tenant_id IS NULL").First(&provider, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Messaging provider not found"})
	}

	return c.JSON(provider)
}

func (h *Handler) UpdateMessagingProvider(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid provider ID"})
	}

	var provider models.MessagingProvider
	if err := h.DB.Where("tenant_id IS NULL").First(&provider, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Messaging provider not found"})
	}

	// Stash existing secrets before BodyParser overwrites (these have json:"-")
	existingAuthToken := provider.AuthToken
	existingWebhookSecret := provider.WebhookSecret

	if err := c.BodyParser(&provider); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// AuthToken and WebhookSecret have json:"-" so BodyParser skips them.
	var raw map[string]interface{}
	if err := json.Unmarshal(c.Body(), &raw); err == nil {
		if token, ok := raw["auth_token"].(string); ok && token != "" {
			provider.AuthToken = token
		} else {
			provider.AuthToken = existingAuthToken
		}
		if secret, ok := raw["webhook_secret"].(string); ok && secret != "" {
			provider.WebhookSecret = secret
		} else {
			provider.WebhookSecret = existingWebhookSecret
		}
	} else {
		provider.AuthToken = existingAuthToken
		provider.WebhookSecret = existingWebhookSecret
	}

	provider.ID = uint(id)
	if err := h.DB.Save(&provider).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update messaging provider"})
	}

	return c.JSON(provider)
}

func (h *Handler) DeleteMessagingProvider(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid provider ID"})
	}

	var provider models.MessagingProvider
	if err := h.DB.Where("tenant_id IS NULL").First(&provider, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Messaging provider not found"})
	}

	if err := h.DB.Delete(&provider).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete messaging provider"})
	}

	c.Status(http.StatusNoContent)
	return nil
}

// =====================
// Messaging Numbers
// =====================

func (h *Handler) ListMessagingNumbers(c *fiber.Ctx) error {
	var numbers []models.MessagingNumber
	query := h.DB.Preload("Provider").Order("phone_number ASC")

	// Optionally filter by provider
	if providerID := c.Query("provider_id"); providerID != "" {
		query = query.Where("provider_id = ?", providerID)
	}

	if err := query.Find(&numbers).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve messaging numbers"})
	}
	return c.JSON(fiber.Map{"data": numbers})
}

func (h *Handler) CreateMessagingNumber(c *fiber.Ctx) error {
	var number models.MessagingNumber
	if err := c.BodyParser(&number); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Verify provider exists
	var provider models.MessagingProvider
	if err := h.DB.First(&provider, number.ProviderID).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid provider ID"})
	}

	if err := h.DB.Create(&number).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create messaging number"})
	}

	return c.Status(http.StatusCreated).JSON(number)
}

func (h *Handler) UpdateMessagingNumber(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	var number models.MessagingNumber
	if err := h.DB.First(&number, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Messaging number not found"})
	}

	if err := c.BodyParser(&number); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Save(&number).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update messaging number"})
	}

	return c.JSON(number)
}

func (h *Handler) DeleteMessagingNumber(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	var number models.MessagingNumber
	if err := h.DB.First(&number, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Messaging number not found"})
	}

	if err := h.DB.Delete(&number).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete messaging number"})
	}

	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) ListGlobalDialplans(c *fiber.Ctx) error {
	var dialplans []models.Dialplan
	// Only get global dialplans (tenant_id is null)
	if err := h.DB.Where("tenant_id IS NULL").Preload("Details").Order("dialplan_order ASC").Find(&dialplans).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve dial plans"})
	}
	return c.JSON(fiber.Map{"data": dialplans})
}

func (h *Handler) CreateGlobalDialplan(c *fiber.Ctx) error {
	var dialplan models.Dialplan
	if err := c.BodyParser(&dialplan); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Ensure it's a global dialplan
	dialplan.TenantID = nil

	if err := h.DB.Create(&dialplan).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create dial plan"})
	}

	return c.Status(http.StatusCreated).JSON(dialplan)
}

func (h *Handler) GetGlobalDialplan(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid dial plan ID"})
	}

	var dialplan models.Dialplan
	if err := h.DB.Where("tenant_id IS NULL").Preload("Details").First(&dialplan, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Dial plan not found"})
	}

	return c.JSON(dialplan)
}

func (h *Handler) UpdateGlobalDialplan(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid dial plan ID"})
	}

	var dialplan models.Dialplan
	if err := h.DB.Where("tenant_id IS NULL").First(&dialplan, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Dial plan not found"})
	}

	var input models.Dialplan
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		input.ID = uint(id)
		input.UUID = dialplan.UUID
		input.TenantID = nil // Keep as global dialplan

		// Update dialplan fields only, skip associations to prevent auto-save duplication
		if err := tx.Omit("Details").Save(&input).Error; err != nil {
			return err
		}

		// Replace details if provided (delete old + create new)
		if input.Details != nil {
			if err := tx.Where("dialplan_uuid = ?", dialplan.UUID).Delete(&models.DialplanDetail{}).Error; err != nil {
				return err
			}
			if len(input.Details) > 0 {
				for i := range input.Details {
					input.Details[i].DialplanUUID = dialplan.UUID
					input.Details[i].ID = 0 // Force create
				}
				if err := tx.Create(&input.Details).Error; err != nil {
					return err
				}
			}
		}

		dialplan = input
		return nil
	})

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update dial plan: " + err.Error()})
	}

	// Reload FreeSWITCH so updated dialplan is active
	h.reloadXML()

	return c.JSON(dialplan)
}

func (h *Handler) DeleteGlobalDialplan(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid dial plan ID"})
	}

	var dialplan models.Dialplan
	if err := h.DB.Where("tenant_id IS NULL").First(&dialplan, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Dial plan not found"})
	}

	// Also delete related details
	h.DB.Where("dialplan_uuid = ?", dialplan.UUID).Delete(&models.DialplanDetail{})

	if err := h.DB.Delete(&dialplan).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete dial plan"})
	}

	c.Status(http.StatusNoContent)
	return nil
}

// =====================
// Access Control Lists (ACLs)
// =====================

func (h *Handler) ListACLs(c *fiber.Ctx) error {
	var acls []models.ACL
	if err := h.DB.Preload("Nodes", func(db *gorm.DB) *gorm.DB {
		return db.Order("priority ASC")
	}).Find(&acls).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve ACLs"})
	}
	return c.JSON(fiber.Map{"data": acls})
}

func (h *Handler) CreateACL(c *fiber.Ctx) error {
	var acl models.ACL
	if err := c.BodyParser(&acl); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Stash nodes before creating to prevent GORM auto-cascade
	incomingNodes := acl.Nodes
	acl.Nodes = nil

	if err := h.DB.Omit("Nodes").Create(&acl).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create ACL"})
	}

	// Create nodes explicitly
	if len(incomingNodes) > 0 {
		for i := range incomingNodes {
			incomingNodes[i].ACLUUID = acl.UUID
			incomingNodes[i].ID = 0
		}
		h.DB.Create(&incomingNodes)
		acl.Nodes = incomingNodes
	}

	h.reloadACL()
	return c.Status(http.StatusCreated).JSON(acl)
}

func (h *Handler) GetACL(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ACL ID"})
	}

	var acl models.ACL
	if err := h.DB.Preload("Nodes", func(db *gorm.DB) *gorm.DB {
		return db.Order("priority ASC")
	}).First(&acl, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "ACL not found"})
	}

	return c.JSON(acl)
}

func (h *Handler) UpdateACL(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ACL ID"})
	}

	var acl models.ACL
	if err := h.DB.First(&acl, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "ACL not found"})
	}

	var input models.ACL
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		// Update ACL fields only, omit nodes to prevent cascade
		input.ID = uint(id)
		input.UUID = acl.UUID // Preserve UUID
		if err := tx.Omit("Nodes").Save(&input).Error; err != nil {
			return err
		}

		// Replace nodes if provided in payload
		if input.Nodes != nil {
			// Delete existing nodes
			if err := tx.Where("acl_uuid = ?", acl.UUID).Delete(&models.ACLNode{}).Error; err != nil {
				return err
			}
			// Create new nodes
			if len(input.Nodes) > 0 {
				for i := range input.Nodes {
					input.Nodes[i].ACLUUID = acl.UUID
					input.Nodes[i].ID = 0
				}
				if err := tx.Create(&input.Nodes).Error; err != nil {
					return err
				}
			}
		}

		acl = input
		return nil
	})

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update ACL: " + err.Error()})
	}

	h.reloadACL()
	return c.JSON(acl)
}

func (h *Handler) DeleteACL(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ACL ID"})
	}

	var acl models.ACL
	if err := h.DB.First(&acl, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "ACL not found"})
	}

	// Delete associated nodes
	h.DB.Where("acl_uuid = ?", acl.UUID).Delete(&models.ACLNode{})

	if err := h.DB.Delete(&acl).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete ACL"})
	}

	h.reloadACL()
	return c.JSON(fiber.Map{"message": "ACL deleted successfully"})
}

func (h *Handler) CreateACLNode(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ACL ID"})
	}

	var acl models.ACL
	if err := h.DB.First(&acl, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "ACL not found"})
	}

	var node models.ACLNode
	if err := c.BodyParser(&node); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	node.ACLUUID = acl.UUID

	if err := h.DB.Create(&node).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create ACL node"})
	}

	return c.Status(http.StatusCreated).JSON(node)
}

func (h *Handler) UpdateACLNode(c *fiber.Ctx) error {
	nodeId, err := strconv.Atoi(c.Params("nodeId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node ID"})
	}

	var node models.ACLNode
	if err := h.DB.First(&node, nodeId).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "ACL node not found"})
	}

	if err := c.BodyParser(&node); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Save(&node).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update ACL node"})
	}

	return c.JSON(node)
}

func (h *Handler) DeleteACLNode(c *fiber.Ctx) error {
	nodeId, err := strconv.Atoi(c.Params("nodeId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid node ID"})
	}

	var node models.ACLNode
	if err := h.DB.First(&node, nodeId).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "ACL node not found"})
	}

	if err := h.DB.Delete(&node).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete ACL node"})
	}

	return c.JSON(fiber.Map{"message": "ACL node deleted successfully"})
}

// =====================
// System Numbers (All Tenants)
// =====================

func (h *Handler) ListAllNumbers(c *fiber.Ctx) error {
	var numbers []models.Destination

	// Get all numbers across all tenants with tenant info
	if err := h.DB.Preload("Tenant").Order("destination_number ASC").Find(&numbers).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve numbers"})
	}

	return c.JSON(fiber.Map{"data": numbers})
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
func (h *Handler) ReportBannedIP(c *fiber.Ctx) error {
	var req BannedIPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Action == "unban" {
		// Mark IP as unbanned
		h.DB.Model(&models.BannedIP{}).
			Where("ip = ? AND status = ?", req.IP, "banned").
			Update("status", "unbanned")
		return c.JSON(fiber.Map{"message": "IP unbanned", "ip": req.IP})
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
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to record ban"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "IP banned", "data": bannedIP})
}

// ListBannedIPs returns all banned IPs
func (h *Handler) ListBannedIPs(c *fiber.Ctx) error {
	var bannedIPs []models.BannedIP

	query := h.DB.Order("banned_at DESC")

	// Filter by status
	status := c.Query("status", "banned")
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&bannedIPs).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve banned IPs"})
	}

	return c.JSON(fiber.Map{"data": bannedIPs})
}

// UnbanIP manually unbans an IP address
func (h *Handler) UnbanIP(c *fiber.Ctx) error {
	ip := c.Params("ip")

	result := h.DB.Model(&models.BannedIP{}).
		Where("ip = ? AND status = ?", ip, "banned").
		Update("status", "unbanned")

	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "IP not found or already unbanned"})
	}

	return c.JSON(fiber.Map{"message": "IP unbanned", "ip": ip})
}

// =====================
// System Device Templates
// =====================

// ListSystemDeviceTemplates returns all system (global) device templates
func (h *Handler) ListSystemDeviceTemplates(c *fiber.Ctx) error {
	var templates []models.DeviceTemplate

	query := h.DB.Where("tenant_id IS NULL")

	// Filter by manufacturer
	if manufacturer := c.Query("manufacturer"); manufacturer != "" {
		query = query.Where("manufacturer = ?", manufacturer)
	}

	query.Order("manufacturer, model, name").Find(&templates)

	// Add device counts
	for i := range templates {
		h.DB.Model(&models.Device{}).Where("template_id = ?", templates[i].ID).Count(&templates[i].DeviceCount)
	}

	return c.JSON(fiber.Map{"data": templates})
}

// CreateSystemDeviceTemplate creates a new system template
func (h *Handler) CreateSystemDeviceTemplate(c *fiber.Ctx) error {
	var tmpl models.DeviceTemplate
	if err := c.BodyParser(&tmpl); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// System templates have no tenant
	tmpl.TenantID = nil
	tmpl.IsSystem = true

	if err := h.DB.Create(&tmpl).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create template"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": tmpl, "message": "System template created"})
}

// GetSystemDeviceTemplate returns a single system template
func (h *Handler) GetSystemDeviceTemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	var tmpl models.DeviceTemplate
	if err := h.DB.Where("id = ? AND tenant_id IS NULL", id).
		Preload("Firmware").
		First(&tmpl).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Template not found"})
	}

	return c.JSON(fiber.Map{"data": tmpl})
}

// UpdateSystemDeviceTemplate updates a system template
func (h *Handler) UpdateSystemDeviceTemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	var tmpl models.DeviceTemplate
	if err := h.DB.Where("id = ? AND tenant_id IS NULL", id).First(&tmpl).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Template not found"})
	}

	var input models.DeviceTemplate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
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

	return c.JSON(fiber.Map{"data": tmpl, "message": "Template updated"})
}

// DeleteSystemDeviceTemplate deletes a system template
func (h *Handler) DeleteSystemDeviceTemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	// Check if any devices are using this template
	var count int64
	h.DB.Model(&models.Device{}).Where("template_id = ?", id).Count(&count)
	if count > 0 {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Template is in use by devices", "device_count": count})
	}

	result := h.DB.Where("id = ? AND tenant_id IS NULL", id).Delete(&models.DeviceTemplate{})
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Template not found"})
	}

	return c.JSON(fiber.Map{"message": "Template deleted"})
}

// =====================
// Firmware Management
// =====================

// ListFirmware returns all firmware files
func (h *Handler) ListFirmware(c *fiber.Ctx) error {
	var firmware []models.Firmware

	query := h.DB.Where("enabled = ?", true)

	// Filter by manufacturer
	if manufacturer := c.Query("manufacturer"); manufacturer != "" {
		query = query.Where("manufacturer = ?", manufacturer)
	}

	// Filter by model/family
	if model := c.Query("model"); model != "" {
		query = query.Where("model = ? OR family = ?", model, model)
	}

	query.Order("manufacturer, model, version DESC").Find(&firmware)

	return c.JSON(fiber.Map{"data": firmware})
}

// CreateFirmware creates a new firmware entry
func (h *Handler) CreateFirmware(c *fiber.Ctx) error {
	var fw models.Firmware
	if err := c.BodyParser(&fw); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.DB.Create(&fw).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create firmware entry"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": fw, "message": "Firmware created"})
}

// GetFirmware returns a single firmware entry
func (h *Handler) GetFirmware(c *fiber.Ctx) error {
	id := c.Params("id")

	var fw models.Firmware
	if err := h.DB.First(&fw, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Firmware not found"})
	}

	return c.JSON(fiber.Map{"data": fw})
}

// UpdateFirmware updates a firmware entry
func (h *Handler) UpdateFirmware(c *fiber.Ctx) error {
	id := c.Params("id")

	var fw models.Firmware
	if err := h.DB.First(&fw, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Firmware not found"})
	}

	var input models.Firmware
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
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

	return c.JSON(fiber.Map{"data": fw, "message": "Firmware updated"})
}

// DeleteFirmware deletes a firmware entry
func (h *Handler) DeleteFirmware(c *fiber.Ctx) error {
	id := c.Params("id")

	// Check if any templates are using this firmware
	var count int64
	h.DB.Model(&models.DeviceTemplate{}).Where("firmware_id = ?", id).Count(&count)
	if count > 0 {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Firmware is referenced by templates", "template_count": count})
	}

	result := h.DB.Delete(&models.Firmware{}, id)
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Firmware not found"})
	}

	return c.JSON(fiber.Map{"message": "Firmware deleted"})
}

// SetDefaultFirmware sets a firmware as the default for its model
func (h *Handler) SetDefaultFirmware(c *fiber.Ctx) error {
	id := c.Params("id")

	var fw models.Firmware
	if err := h.DB.First(&fw, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Firmware not found"})
	}

	// Clear default flag for same manufacturer/model
	h.DB.Model(&models.Firmware{}).
		Where("manufacturer = ? AND (model = ? OR family = ?)", fw.Manufacturer, fw.Model, fw.Family).
		Update("is_default", false)

	// Set this one as default
	h.DB.Model(&fw).Update("is_default", true)

	return c.JSON(fiber.Map{"message": "Firmware set as default"})
}

// UploadFirmware handles firmware file upload
func (h *Handler) UploadFirmware(c *fiber.Ctx) error {
	// Get firmware ID
	id := c.Params("id")

	var fw models.Firmware
	if err := h.DB.First(&fw, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Firmware not found"})
	}

	// Get file
	header, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
	}
	file, err := header.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	// Use configured firmware path or default
	firmwareBasePath := h.Config.FirmwarePath
	if firmwareBasePath == "" {
		firmwareBasePath = "/usr/share/freeswitch/firmware"
	}

	// Sanitize manufacturer name for directory
	safeManufacturer := strings.ToLower(strings.ReplaceAll(fw.Manufacturer, " ", "_"))
	safeManufacturer = strings.ReplaceAll(safeManufacturer, "/", "_")

	// Create directory structure: firmware_path/manufacturer/model/
	safeModel := strings.ToLower(strings.ReplaceAll(fw.Model, " ", "_"))
	safeModel = strings.ReplaceAll(safeModel, "/", "_")
	dirPath := filepath.Join(firmwareBasePath, safeManufacturer, safeModel)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create firmware directory"})
	}

	// Sanitize filename
	safeFilename := strings.ReplaceAll(header.Filename, "..", "_")
	fullPath := filepath.Join(dirPath, safeFilename)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create file"})
	}
	defer dst.Close()

	// Copy file content
	written, err := io.Copy(dst, file)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	// Relative path for storage (from firmware base)
	relativePath := filepath.Join(safeManufacturer, safeModel, safeFilename)

	// Update firmware record
	h.DB.Model(&fw).Updates(map[string]interface{}{
		"file_path": relativePath,
		"file_name": safeFilename,
		"file_size": written,
	})

	return c.JSON(fiber.Map{
		"message":   "Firmware file uploaded",
		"path":      relativePath,
		"full_path": fullPath,
		"size":      written,
	})
}

// =====================
// System Numbers (centralized pool)
// =====================

func (h *Handler) ListSystemNumbers(c *fiber.Ctx) error {
	var numbers []models.SystemNumber

	query := h.DB.Preload("Tenant").Preload("NumberGroup").Order("phone_number ASC")

	// Optional filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if tenantID := c.Query("tenant_id"); tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if groupID := c.Query("group_id"); groupID != "" {
		query = query.Where("number_group_id = ?", groupID)
	}

	if err := query.Find(&numbers).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve system numbers"})
	}

	return c.JSON(fiber.Map{"data": numbers})
}

func (h *Handler) CreateSystemNumber(c *fiber.Ctx) error {
	var number models.SystemNumber
	if err := c.BodyParser(&number); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Normalize to E.164
	number.PhoneNumber = normalizeToE164(number.PhoneNumber)

	// Check for duplicates
	var existing models.SystemNumber
	if err := h.DB.Where("phone_number = ?", number.PhoneNumber).First(&existing).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "This number already exists in the system"})
	}

	// Set initial status
	if number.TenantID != nil {
		number.Status = models.NumberStatusAssigned
	} else {
		number.Status = models.NumberStatusAvailable
	}

	if err := h.DB.Create(&number).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create system number"})
	}

	// If assigned to a tenant on creation, auto-create Destination
	if number.TenantID != nil {
		dest := models.Destination{
			TenantID:          *number.TenantID,
			DestinationNumber: number.PhoneNumber,
			Description:       number.Description,
			Enabled:           true,
			Context:           "public",
		}
		if err := h.DB.Create(&dest).Error; err == nil {
			h.DB.Model(&number).Update("destination_id", dest.ID)
		}
	}

	// Reload with associations
	h.DB.Preload("Tenant").Preload("NumberGroup").First(&number, number.ID)

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": number})
}

func (h *Handler) GetSystemNumber(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	var number models.SystemNumber
	if err := h.DB.Preload("Tenant").Preload("NumberGroup").Preload("Destination").First(&number, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "System number not found"})
	}

	return c.JSON(fiber.Map{"data": number})
}

func (h *Handler) UpdateSystemNumber(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	var number models.SystemNumber
	if err := h.DB.First(&number, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "System number not found"})
	}

	if err := c.BodyParser(&number); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	number.ID = uint(id)
	if err := h.DB.Save(&number).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update system number"})
	}

	h.DB.Preload("Tenant").Preload("NumberGroup").First(&number, id)
	return c.JSON(fiber.Map{"data": number})
}

func (h *Handler) DeleteSystemNumber(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	var number models.SystemNumber
	if err := h.DB.First(&number, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "System number not found"})
	}

	// Clean up associated Destination if exists
	if number.DestinationID != nil {
		h.DB.Delete(&models.Destination{}, *number.DestinationID)
	}

	// Clean up location references
	h.DB.Model(&models.Location{}).Where("system_number_id = ?", number.ID).Update("system_number_id", nil)

	h.DB.Delete(&number)
	return c.JSON(fiber.Map{"message": "System number deleted"})
}

// AssignNumberToTenant assigns a system number to a tenant and creates a Destination
func (h *Handler) AssignNumberToTenant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	var input struct {
		TenantID uint `json:"tenant_id"`
	}
	if err := c.BodyParser(&input); err != nil || input.TenantID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "tenant_id is required"})
	}

	var number models.SystemNumber
	if err := h.DB.First(&number, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "System number not found"})
	}

	// Verify tenant exists
	var tenant models.Tenant
	if err := h.DB.First(&tenant, input.TenantID).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant not found"})
	}

	// If already assigned to a different tenant, clean up old Destination
	if number.TenantID != nil && *number.TenantID != input.TenantID && number.DestinationID != nil {
		h.DB.Delete(&models.Destination{}, *number.DestinationID)
	}

	// Create Destination for the tenant
	dest := models.Destination{
		TenantID:          input.TenantID,
		DestinationNumber: number.PhoneNumber,
		Description:       number.Description,
		Enabled:           true,
		Context:           "public",
	}
	if err := h.DB.Create(&dest).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create destination"})
	}

	// Update the system number
	h.DB.Model(&number).Updates(map[string]interface{}{
		"tenant_id":      input.TenantID,
		"destination_id": dest.ID,
		"status":         models.NumberStatusAssigned,
	})

	h.reloadXML()

	h.DB.Preload("Tenant").Preload("NumberGroup").Preload("Destination").First(&number, id)
	return c.JSON(fiber.Map{"data": number, "message": "Number assigned to tenant"})
}

// UnassignNumber removes a system number from a tenant
func (h *Handler) UnassignNumber(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	var number models.SystemNumber
	if err := h.DB.First(&number, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "System number not found"})
	}

	if number.TenantID == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Number is not assigned to any tenant"})
	}

	// Remove location references for this number
	h.DB.Model(&models.Location{}).Where("system_number_id = ?", number.ID).Update("system_number_id", nil)

	// Delete the Destination
	if number.DestinationID != nil {
		h.DB.Delete(&models.Destination{}, *number.DestinationID)
	}

	// Clear assignment
	h.DB.Model(&number).Updates(map[string]interface{}{
		"tenant_id":      nil,
		"destination_id": nil,
		"status":         models.NumberStatusAvailable,
	})

	h.reloadXML()

	return c.JSON(fiber.Map{"message": "Number unassigned from tenant"})
}

// =====================
// Number Groups
// =====================

func (h *Handler) ListNumberGroups(c *fiber.Ctx) error {
	var groups []models.NumberGroup
	if err := h.DB.Preload("DefaultGateway").Preload("MessagingProvider").Preload("RoutingRules", func(db *gorm.DB) *gorm.DB {
		return db.Order("priority ASC")
	}).Order("name ASC").Find(&groups).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve number groups"})
	}

	// Add number count for each group
	type GroupWithCount struct {
		models.NumberGroup
		NumberCount int64 `json:"number_count"`
	}
	var result []GroupWithCount
	for _, g := range groups {
		var count int64
		h.DB.Model(&models.SystemNumber{}).Where("number_group_id = ?", g.ID).Count(&count)
		result = append(result, GroupWithCount{NumberGroup: g, NumberCount: count})
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *Handler) CreateNumberGroup(c *fiber.Ctx) error {
	var group models.NumberGroup
	if err := c.BodyParser(&group); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.DB.Create(&group).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create number group"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": group})
}

func (h *Handler) GetNumberGroup(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	var group models.NumberGroup
	if err := h.DB.Preload("DefaultGateway").Preload("MessagingProvider").Preload("RoutingRules", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Gateway").Order("priority ASC")
	}).First(&group, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number group not found"})
	}

	return c.JSON(fiber.Map{"data": group})
}

func (h *Handler) UpdateNumberGroup(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	var group models.NumberGroup
	if err := h.DB.First(&group, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number group not found"})
	}

	if err := c.BodyParser(&group); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	group.ID = uint(id)
	if err := h.DB.Save(&group).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update number group"})
	}

	return c.JSON(fiber.Map{"data": group})
}

func (h *Handler) DeleteNumberGroup(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	// Unlink numbers from this group first
	h.DB.Model(&models.SystemNumber{}).Where("number_group_id = ?", id).Update("number_group_id", nil)

	if err := h.DB.Delete(&models.NumberGroup{}, id).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete number group"})
	}

	return c.JSON(fiber.Map{"message": "Number group deleted"})
}

// ReorderGroupGateways updates the gateway priority list for a number group
func (h *Handler) ReorderGroupGateways(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	var group models.NumberGroup
	if err := h.DB.First(&group, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number group not found"})
	}

	var input struct {
		GatewayPriorities models.GatewayPriorityList `json:"gateway_priorities"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	h.DB.Model(&group).Update("gateway_priorities", input.GatewayPriorities)

	return c.JSON(fiber.Map{"data": group, "message": "Gateway priorities updated"})
}

// =====================
// Outbound Routing Rules (per Number Group)
// =====================

// ListRoutingRules returns all routing rules for a number group
func (h *Handler) ListRoutingRules(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	// Verify group exists
	var group models.NumberGroup
	if err := h.DB.First(&group, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number group not found"})
	}

	var rules []models.OutboundRoutingRule
	if err := h.DB.Where("number_group_id = ?", id).Preload("Gateway").Order("priority ASC").Find(&rules).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve routing rules"})
	}

	return c.JSON(fiber.Map{"data": rules})
}

// CreateRoutingRule creates a new routing rule for a number group
func (h *Handler) CreateRoutingRule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	// Verify group exists
	var group models.NumberGroup
	if err := h.DB.First(&group, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number group not found"})
	}

	var rule models.OutboundRoutingRule
	if err := c.BodyParser(&rule); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Force parent association
	rule.NumberGroupID = uint(id)

	// Validate required fields
	if rule.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Rule name is required"})
	}
	if rule.Pattern == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Pattern is required"})
	}

	// Auto-fill gateway name if ID is provided
	if rule.GatewayID != nil && rule.GatewayName == "" {
		var gw models.Gateway
		if err := h.DB.First(&gw, *rule.GatewayID).Error; err == nil {
			rule.GatewayName = gw.GatewayName
		}
	}

	if err := h.DB.Create(&rule).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create routing rule"})
	}

	// Reload with gateway
	h.DB.Preload("Gateway").First(&rule, rule.ID)

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": rule})
}

// UpdateRoutingRule updates a routing rule
func (h *Handler) UpdateRoutingRule(c *fiber.Ctx) error {
	groupID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	ruleID, err := strconv.Atoi(c.Params("ruleId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rule ID"})
	}

	var rule models.OutboundRoutingRule
	if err := h.DB.Where("id = ? AND number_group_id = ?", ruleID, groupID).First(&rule).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Routing rule not found"})
	}

	if err := c.BodyParser(&rule); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Preserve identity
	rule.ID = uint(ruleID)
	rule.NumberGroupID = uint(groupID)

	// Auto-fill gateway name if ID changed
	if rule.GatewayID != nil && rule.GatewayName == "" {
		var gw models.Gateway
		if err := h.DB.First(&gw, *rule.GatewayID).Error; err == nil {
			rule.GatewayName = gw.GatewayName
		}
	}

	if err := h.DB.Save(&rule).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update routing rule"})
	}

	h.DB.Preload("Gateway").First(&rule, ruleID)
	return c.JSON(fiber.Map{"data": rule})
}

// DeleteRoutingRule deletes a routing rule
func (h *Handler) DeleteRoutingRule(c *fiber.Ctx) error {
	groupID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	ruleID, err := strconv.Atoi(c.Params("ruleId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rule ID"})
	}

	result := h.DB.Where("id = ? AND number_group_id = ?", ruleID, groupID).Delete(&models.OutboundRoutingRule{})
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Routing rule not found"})
	}

	return c.JSON(fiber.Map{"message": "Routing rule deleted"})
}

// ReorderRoutingRules bulk-updates priority for routing rules in a number group
func (h *Handler) ReorderRoutingRules(c *fiber.Ctx) error {
	groupID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	var items []struct {
		ID       uint `json:"id"`
		Priority int  `json:"priority"`
	}
	if err := c.BodyParser(&items); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	tx := h.DB.Begin()
	for _, item := range items {
		if err := tx.Model(&models.OutboundRoutingRule{}).
			Where("id = ? AND number_group_id = ?", item.ID, groupID).
			Update("priority", item.Priority).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reorder rules"})
		}
	}
	tx.Commit()

	return c.JSON(fiber.Map{"message": "Routing rules reordered"})
}

// ReorderGateways bulk-updates gateway priority/weight for trunk ordering
func (h *Handler) ReorderGateways(c *fiber.Ctx) error {
	var items []struct {
		ID       uint `json:"id"`
		Priority int  `json:"priority"`
		Weight   int  `json:"weight"`
	}
	if err := c.BodyParser(&items); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	tx := h.DB.Begin()
	for _, item := range items {
		if err := tx.Model(&models.Gateway{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
			"priority": item.Priority,
			"weight":   item.Weight,
		}).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reorder gateways"})
		}
	}
	tx.Commit()

	return c.JSON(fiber.Map{"message": "Gateway order updated"})
}

// AssignNumberToLocation links a tenant's assigned system number to a location for E911
func (h *Handler) AssignNumberToLocation(c *fiber.Ctx) error {
	numID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	tenantID := middleware.GetTenantID(c)

	var input struct {
		LocationID uint `json:"location_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Verify the system number is assigned to this tenant
	var number models.SystemNumber
	if err := h.DB.Where("id = ? AND tenant_id = ?", numID, tenantID).First(&number).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number not found or not assigned to your tenant"})
	}

	// Verify the location belongs to this tenant
	var location models.Location
	if err := h.DB.Where("id = ? AND tenant_id = ?", input.LocationID, tenantID).First(&location).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Location not found"})
	}

	// Link the number to the location
	h.DB.Model(&location).Update("system_number_id", number.ID)

	return c.JSON(fiber.Map{"message": "Number assigned to location", "data": location})
}

// UnassignNumberFromLocation removes the number-location link
func (h *Handler) UnassignNumberFromLocation(c *fiber.Ctx) error {
	numID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number ID"})
	}

	tenantID := middleware.GetTenantID(c)

	// Verify the system number is assigned to this tenant
	var number models.SystemNumber
	if err := h.DB.Where("id = ? AND tenant_id = ?", numID, tenantID).First(&number).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Number not found or not assigned to your tenant"})
	}

	// Remove from any locations in this tenant
	h.DB.Model(&models.Location{}).
		Where("system_number_id = ? AND tenant_id = ?", number.ID, tenantID).
		Update("system_number_id", nil)

	return c.JSON(fiber.Map{"message": "Number unassigned from location"})
}
