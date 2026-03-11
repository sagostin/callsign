package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

// ===========================
// Client Registration Handlers
// ===========================

// ListClientRegistrations returns all active registrations for the current tenant
func (h *Handler) ListClientRegistrations(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	// Optional filters
	endpointType := ctx.URLParam("type")        // device, mobile_app, desktop_app, web_client
	extensionID := ctx.URLParam("extension_id") // filter by extension
	status := ctx.URLParam("status")            // provisioned, registered, expired, offline

	query := h.DB.Where("tenant_id = ?", tenantID).
		Preload("User").
		Preload("Extension").
		Preload("Device")

	if endpointType != "" {
		query = query.Where("endpoint_type = ?", endpointType)
	}
	if extensionID != "" {
		query = query.Where("extension_id = ?", extensionID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var registrations []models.ClientRegistration
	if err := query.Order("created_at DESC").Find(&registrations).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch registrations"})
		return
	}

	ctx.JSON(iris.Map{"registrations": registrations, "total": len(registrations)})
}

// ListExtensionRegistrations returns all active registrations for a specific extension
func (h *Handler) ListExtensionRegistrations(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	extID := ctx.Params().GetUintDefault("id", 0)

	// Verify the extension belongs to this tenant
	var ext models.Extension
	if err := h.DB.Where("id = ? AND tenant_id = ?", extID, tenantID).First(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Extension not found"})
		return
	}

	var registrations []models.ClientRegistration
	if err := h.DB.Where("extension_id = ? AND tenant_id = ? AND enabled = ?", extID, tenantID, true).
		Preload("Device").
		Order("endpoint_type ASC, created_at DESC").
		Find(&registrations).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch registrations"})
		return
	}

	ctx.JSON(iris.Map{
		"extension":     ext.Extension,
		"registrations": registrations,
		"total":         len(registrations),
	})
}

// ProvisionClientRegistration creates a new app/web client registration
// This generates SIP credentials for the client to register with FreeSWITCH
func (h *Handler) ProvisionClientRegistration(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var req struct {
		EndpointType string `json:"endpoint_type"` // mobile_app, desktop_app, web_client
		InstanceID   string `json:"instance_id"`   // Client-generated unique ID
		DeviceLabel  string `json:"device_label"`  // "John's iPhone", "Chrome Browser"
		AppVersion   string `json:"app_version"`
		OSInfo       string `json:"os_info"`
		UserID       *uint  `json:"user_id"`      // Optional: link to user
		ExtensionID  *uint  `json:"extension_id"` // Optional: link to extension
	}

	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Validate endpoint type
	endpointType := models.EndpointType(req.EndpointType)
	switch endpointType {
	case models.EndpointTypeMobileApp, models.EndpointTypeDesktopApp, models.EndpointTypeWebClient:
		// Valid
	default:
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid endpoint_type. Must be: mobile_app, desktop_app, or web_client"})
		return
	}

	// Instance ID required for apps and web clients
	if req.InstanceID == "" {
		req.InstanceID = uuid.New().String()[:8]
	}

	// If user/extension is specified, verify they belong to this tenant
	if req.UserID != nil {
		var user models.User
		if err := h.DB.Where("id = ? AND tenant_id = ?", *req.UserID, tenantID).First(&user).Error; err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "User not found in tenant"})
			return
		}
	}
	if req.ExtensionID != nil {
		var ext models.Extension
		if err := h.DB.Where("id = ? AND tenant_id = ?", *req.ExtensionID, tenantID).First(&ext).Error; err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "Extension not found in tenant"})
			return
		}
	}

	// Auto-resolve user/extension linkage: if user is given but not extension, find their extension
	if req.UserID != nil && req.ExtensionID == nil {
		var ext models.Extension
		if err := h.DB.Where("user_id = ? AND tenant_id = ? AND enabled = ?", *req.UserID, tenantID, true).First(&ext).Error; err == nil {
			req.ExtensionID = &ext.ID
		}
	}

	// Generate a short user identifier for the registration username
	userShort := uuid.New().String()[:8]
	if req.UserID != nil {
		var user models.User
		if err := h.DB.First(&user, *req.UserID).Error; err == nil {
			userShort = user.UUID.String()[:8]
		}
	}

	// Build registration user
	regUser := models.GenerateRegistrationUser(endpointType, userShort, req.InstanceID)

	// Check for existing registration with same instance ID for this user
	var existing models.ClientRegistration
	if req.UserID != nil {
		if err := h.DB.Where(
			"user_id = ? AND instance_id = ? AND endpoint_type = ? AND tenant_id = ?",
			*req.UserID, req.InstanceID, endpointType, tenantID,
		).First(&existing).Error; err == nil {
			// Return existing credentials
			ctx.JSON(iris.Map{
				"registration":        existing,
				"sip_user":            existing.RegistrationUser,
				"sip_password":        existing.RegistrationPass,
				"already_provisioned": true,
			})
			return
		}
	}

	// Create the registration
	reg := models.ClientRegistration{
		TenantID:         tenantID,
		UserID:           req.UserID,
		ExtensionID:      req.ExtensionID,
		EndpointType:     endpointType,
		RegistrationUser: regUser,
		InstanceID:       req.InstanceID,
		DisplayName:      req.DeviceLabel,
		DeviceLabel:      req.DeviceLabel,
		AppVersion:       req.AppVersion,
		OSInfo:           req.OSInfo,
		AllowOutbound:    req.ExtensionID != nil, // Can make outbound if linked to extension
		WebRTC:           endpointType == models.EndpointTypeWebClient,
		Status:           "provisioned",
		Enabled:          true,
	}

	if err := h.DB.Create(&reg).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create registration: " + err.Error()})
		return
	}

	// Trigger FreeSWITCH XML reload for directory changes
	h.reloadXML()

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{
		"registration": reg,
		"sip_user":     reg.RegistrationUser,
		"sip_password": reg.RegistrationPass,
	})
}

// DeleteClientRegistration removes a client registration (force-unregister)
func (h *Handler) DeleteClientRegistration(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	regID := ctx.Params().GetUintDefault("id", 0)

	var reg models.ClientRegistration
	if err := h.DB.Where("id = ? AND tenant_id = ?", regID, tenantID).First(&reg).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Registration not found"})
		return
	}

	if err := h.DB.Delete(&reg).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete registration"})
		return
	}

	// Trigger FreeSWITCH XML reload for directory changes
	h.reloadXML()

	ctx.JSON(iris.Map{"message": "Registration removed"})
}

// ListUnassignedRegistrations returns devices/clients that are registered but not assigned to a user
func (h *Handler) ListUnassignedRegistrations(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var registrations []models.ClientRegistration
	if err := h.DB.Where(
		"tenant_id = ? AND user_id IS NULL AND enabled = ?",
		tenantID, true,
	).Preload("Device").Order("created_at DESC").Find(&registrations).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch registrations"})
		return
	}

	ctx.JSON(iris.Map{"registrations": registrations, "total": len(registrations)})
}

// AssignRegistration assigns a client registration to a user/extension
func (h *Handler) AssignRegistration(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	regID := ctx.Params().GetUintDefault("id", 0)

	var req struct {
		UserID      uint `json:"user_id"`
		ExtensionID uint `json:"extension_id"`
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	var reg models.ClientRegistration
	if err := h.DB.Where("id = ? AND tenant_id = ?", regID, tenantID).First(&reg).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Registration not found"})
		return
	}

	// Verify user and extension belong to tenant
	var user models.User
	if err := h.DB.Where("id = ? AND tenant_id = ?", req.UserID, tenantID).First(&user).Error; err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "User not found in tenant"})
		return
	}

	var ext models.Extension
	if err := h.DB.Where("id = ? AND tenant_id = ?", req.ExtensionID, tenantID).First(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Extension not found in tenant"})
		return
	}

	reg.UserID = &req.UserID
	reg.ExtensionID = &req.ExtensionID
	reg.AllowOutbound = true

	if err := h.DB.Save(&reg).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to assign registration"})
		return
	}

	// Trigger FreeSWITCH XML reload for directory changes
	h.reloadXML()

	ctx.JSON(iris.Map{"message": "Registration assigned", "registration": reg})
}
