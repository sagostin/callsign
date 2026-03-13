package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ===========================
// Client Registration Handlers
// ===========================

// ListClientRegistrations returns all active registrations for the current tenant
func (h *Handler) ListClientRegistrations(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	// Optional filters
	endpointType := c.Query("type")        // device, mobile_app, desktop_app, web_client
	extensionID := c.Query("extension_id") // filter by extension
	status := c.Query("status")            // provisioned, registered, expired, offline

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
		h.logError("CLIENT_REG", "ListClientRegistrations: Failed to fetch registrations", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch registrations"})
	}

	return c.JSON(fiber.Map{"registrations": registrations, "total": len(registrations)})
}

// ListExtensionRegistrations returns all active registrations for a specific extension
func (h *Handler) ListExtensionRegistrations(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	extID := c.Params("id")

	// Verify the extension belongs to this tenant
	var ext models.Extension
	if err := h.DB.Where("id = ? AND tenant_id = ?", extID, tenantID).First(&ext).Error; err != nil {
		h.logWarn("CLIENT_REG", "ListExtensionRegistrations: Extension not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Extension not found"})
	}

	var registrations []models.ClientRegistration
	if err := h.DB.Where("extension_id = ? AND tenant_id = ? AND enabled = ?", extID, tenantID, true).
		Preload("Device").
		Order("endpoint_type ASC, created_at DESC").
		Find(&registrations).Error; err != nil {
		h.logError("CLIENT_REG", "ListExtensionRegistrations: Failed to fetch registrations", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch registrations"})
	}

	return c.JSON(fiber.Map{
		"extension":     ext.Extension,
		"registrations": registrations,
		"total":         len(registrations),
	})
}

// ProvisionClientRegistration creates a new app/web client registration
// This generates SIP credentials for the client to register with FreeSWITCH
func (h *Handler) ProvisionClientRegistration(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var req struct {
		EndpointType string `json:"endpoint_type"` // mobile_app, desktop_app, web_client
		InstanceID   string `json:"instance_id"`   // Client-generated unique ID
		DeviceLabel  string `json:"device_label"`  // "John's iPhone", "Chrome Browser"
		AppVersion   string `json:"app_version"`
		OSInfo       string `json:"os_info"`
		UserID       *uint  `json:"user_id"`      // Optional: link to user
		ExtensionID  *uint  `json:"extension_id"` // Optional: link to extension
	}

	if err := c.BodyParser(&req); err != nil {
		h.logWarn("CLIENT_REG", "ProvisionClientRegistration: Invalid request body", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate endpoint type
	endpointType := models.EndpointType(req.EndpointType)
	switch endpointType {
	case models.EndpointTypeMobileApp, models.EndpointTypeDesktopApp, models.EndpointTypeWebClient:
		// Valid
	default:
		h.logWarn("CLIENT_REG", "ProvisionClientRegistration: Invalid endpoint_type. Must be: mobile_app, desktop_app, or web_client", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid endpoint_type. Must be: mobile_app, desktop_app, or web_client"})
	}

	// Instance ID required for apps and web clients
	if req.InstanceID == "" {
		req.InstanceID = uuid.New().String()[:8]
	}

	// If user/extension is specified, verify they belong to this tenant
	if req.UserID != nil {
		var user models.User
		if err := h.DB.Where("id = ? AND tenant_id = ?", *req.UserID, tenantID).First(&user).Error; err != nil {
			h.logWarn("CLIENT_REG", "ProvisionClientRegistration: User not found in tenant", h.reqFields(c, nil))
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "User not found in tenant"})
		}
	}
	if req.ExtensionID != nil {
		var ext models.Extension
		if err := h.DB.Where("id = ? AND tenant_id = ?", *req.ExtensionID, tenantID).First(&ext).Error; err != nil {
			h.logWarn("CLIENT_REG", "ProvisionClientRegistration: Extension not found in tenant", h.reqFields(c, nil))
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Extension not found in tenant"})
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
			return c.JSON(fiber.Map{
				"registration":        existing,
				"sip_user":            existing.RegistrationUser,
				"sip_password":        existing.RegistrationPass,
				"already_provisioned": true,
			})
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
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create registration: " + err.Error()})
	}

	// Trigger FreeSWITCH XML reload for directory changes
	h.reloadXML()

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"registration": reg,
		"sip_user":     reg.RegistrationUser,
		"sip_password": reg.RegistrationPass,
	})
}

// DeleteClientRegistration removes a client registration (force-unregister)
func (h *Handler) DeleteClientRegistration(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	regID := c.Params("id")

	var reg models.ClientRegistration
	if err := h.DB.Where("id = ? AND tenant_id = ?", regID, tenantID).First(&reg).Error; err != nil {
		h.logWarn("CLIENT_REG", "DeleteClientRegistration: Registration not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Registration not found"})
	}

	if err := h.DB.Delete(&reg).Error; err != nil {
		h.logError("CLIENT_REG", "DeleteClientRegistration: Failed to delete registration", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete registration"})
	}

	// Trigger FreeSWITCH XML reload for directory changes
	h.reloadXML()

	return c.JSON(fiber.Map{"message": "Registration removed"})
}

// ListUnassignedRegistrations returns devices/clients that are registered but not assigned to a user
func (h *Handler) ListUnassignedRegistrations(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var registrations []models.ClientRegistration
	if err := h.DB.Where(
		"tenant_id = ? AND user_id IS NULL AND enabled = ?",
		tenantID, true,
	).Preload("Device").Order("created_at DESC").Find(&registrations).Error; err != nil {
		h.logError("CLIENT_REG", "ListUnassignedRegistrations: Failed to fetch registrations", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch registrations"})
	}

	return c.JSON(fiber.Map{"registrations": registrations, "total": len(registrations)})
}

// AssignRegistration assigns a client registration to a user/extension
func (h *Handler) AssignRegistration(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	regID := c.Params("id")

	var req struct {
		UserID      uint `json:"user_id"`
		ExtensionID uint `json:"extension_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		h.logWarn("CLIENT_REG", "AssignRegistration: Invalid request body", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var reg models.ClientRegistration
	if err := h.DB.Where("id = ? AND tenant_id = ?", regID, tenantID).First(&reg).Error; err != nil {
		h.logWarn("CLIENT_REG", "AssignRegistration: Registration not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Registration not found"})
	}

	// Verify user and extension belong to tenant
	var user models.User
	if err := h.DB.Where("id = ? AND tenant_id = ?", req.UserID, tenantID).First(&user).Error; err != nil {
		h.logWarn("CLIENT_REG", "AssignRegistration: User not found in tenant", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "User not found in tenant"})
	}

	var ext models.Extension
	if err := h.DB.Where("id = ? AND tenant_id = ?", req.ExtensionID, tenantID).First(&ext).Error; err != nil {
		h.logWarn("CLIENT_REG", "AssignRegistration: Extension not found in tenant", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Extension not found in tenant"})
	}

	reg.UserID = &req.UserID
	reg.ExtensionID = &req.ExtensionID
	reg.AllowOutbound = true

	if err := h.DB.Save(&reg).Error; err != nil {
		h.logError("CLIENT_REG", "AssignRegistration: Failed to assign registration", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign registration"})
	}

	// Trigger FreeSWITCH XML reload for directory changes
	h.reloadXML()

	return c.JSON(fiber.Map{"message": "Registration assigned", "registration": reg})
}
