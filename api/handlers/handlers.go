package handlers

import (
	"callsign/config"
	"callsign/middleware"
	"callsign/models"
	"callsign/services/cdr"
	"callsign/services/esl"
	"callsign/services/logging"
	"callsign/services/messaging"
	"callsign/services/websocket"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handler holds all HTTP handlers and their dependencies
type Handler struct {
	DB                  *gorm.DB
	Config              *config.Config
	Auth                *middleware.AuthMiddleware
	ESLManager          *esl.Manager
	ConsoleManager      *ConsoleManager
	NotificationManager *NotificationManager
	LogManager          *logging.LogManager
	WSHub               *websocket.Hub
	MsgManager          *messaging.Manager
	CHClient            *cdr.ClickHouseClient
}

// NewHandler creates a new Handler instance
func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		DB:     db,
		Config: cfg,
		Auth:   middleware.NewAuthMiddleware(cfg, db),
	}
}

// SetESLManager sets the ESL manager reference
func (h *Handler) SetESLManager(m *esl.Manager) {
	h.ESLManager = m
}

// SetWSHub sets the WebSocket hub reference
func (h *Handler) SetWSHub(hub *websocket.Hub) {
	h.WSHub = hub
}

// SetMsgManager sets the messaging manager reference
func (h *Handler) SetMsgManager(mgr *messaging.Manager) {
	h.MsgManager = mgr
}

// SetClickHouse sets the ClickHouse CDR client reference
func (h *Handler) SetClickHouse(ch *cdr.ClickHouseClient) {
	h.CHClient = ch
}

// reloadXML triggers a FreeSWITCH XML reload in the background.
// Safe to call even if ESL is not connected — it silently no-ops.
func (h *Handler) reloadXML() {
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		go h.ESLManager.ReloadXML()
	}
}

// reloadSofia triggers a Sofia profile rescan + XML reload in the background.
func (h *Handler) reloadSofia(profileName string) {
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		go func() {
			h.ESLManager.ReloadXML()
			if profileName != "" {
				h.ESLManager.SofiaRescan(profileName)
			}
		}()
	}
}

// reloadCallcenter triggers a mod_callcenter reload + XML reload in the background.
func (h *Handler) reloadCallcenter() {
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		go func() {
			h.ESLManager.ReloadXML()
			h.ESLManager.CallcenterReload()
		}()
	}
}

// reloadACL triggers an ACL reload in the background.
func (h *Handler) reloadACL() {
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		go h.ESLManager.ReloadACL()
	}
}

// SetLogManager sets the LogManager reference
func (h *Handler) SetLogManager(lm *logging.LogManager) {
	h.LogManager = lm
}

// =====================
// Helper utilities
// =====================

// getLocalsUint retrieves a uint value from Fiber locals with a default fallback
func getLocalsUint(c *fiber.Ctx, key string, defaultVal uint) uint {
	if val, ok := c.Locals(key).(uint); ok {
		return val
	}
	return defaultVal
}

// getLocalsString retrieves a string value from Fiber locals with a default fallback
func getLocalsString(c *fiber.Ctx, key string, defaultVal string) string {
	if val, ok := c.Locals(key).(string); ok {
		return val
	}
	return defaultVal
}

// =====================
// WebSocket
// =====================

// HandleWebSocket handles the general-purpose real-time event WebSocket endpoint
func (h *Handler) HandleWebSocket(c *fiber.Ctx) error {
	return fiberws.New(func(conn *fiberws.Conn) {
		// Extract token from query param for authentication
		token := conn.Query("token")
		if token == "" {
			conn.WriteJSON(fiber.Map{"error": "token required"})
			conn.Close()
			return
		}

		// Verify token
		claims, err := h.Auth.VerifyToken(token)
		if err != nil {
			conn.WriteJSON(fiber.Map{"error": "invalid token"})
			conn.Close()
			return
		}

		// Create hub client
		clientID := fmt.Sprintf("ws-%d-%d", claims.UserID, time.Now().UnixNano())
		var tenantID uint
		if claims.TenantID != nil {
			tenantID = *claims.TenantID
		}

		// Register with the WebSocket hub using gorilla-compatible type
		// The fiberws.Conn embeds fasthttp/websocket.Conn which is API-compatible
		// For the hub we need a gorilla websocket.Conn, but since the hub
		// uses its own broadcast mechanism, we bridge via a simple read loop
		conn.WriteJSON(fiber.Map{
			"type":    "connected",
			"user_id": claims.UserID,
		})

		_ = clientID
		_ = tenantID

		// Simple keepalive loop — real-time events are pushed via the hub's
		// broadcast mechanism to Notification WS. This endpoint is a fallback.
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})(c)
}

// =====================
// Health & Status
// =====================

// Health returns the health status of the API
func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// =====================
// Authentication
// =====================

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
}

// resolveTenantDomain determines the tenant domain from the request.
// Priority: explicit domain field > Host header. Returns empty string
// for localhost / 127.x which means "no tenant scoping".
func (h *Handler) resolveTenantDomain(c *fiber.Ctx, explicit string) string {
	domain := strings.TrimSpace(explicit)
	if domain == "" {
		domain = c.Hostname()
	}
	// Strip port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	// Skip tenant scoping for local development
	if domain == "" || domain == "localhost" || strings.HasPrefix(domain, "127.") {
		return ""
	}
	return domain
}

// Login authenticates a user and returns a JWT token
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Resolve tenant from the connected domain
	var user models.User
	var userErr error

	if domain := h.resolveTenantDomain(c, req.Domain); domain != "" {
		var tenant models.Tenant
		if err := h.DB.Where("domain = ? AND enabled = true", domain).First(&tenant).Error; err != nil {
			log.WithField("domain", domain).Debug("Login: tenant not found for domain")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		userErr = h.DB.Where("(username = ? OR email = ?) AND tenant_id = ?", req.Username, req.Username, tenant.ID).First(&user).Error
	} else {
		// No tenant resolved (localhost) – global lookup for backward compat
		userErr = h.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error
	}

	if userErr != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if !user.CheckPassword(req.Password) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// Generate token
	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":        user.ID,
			"uuid":      user.UUID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"tenant_id": user.TenantID,
		},
	})
}

// AdminLogin authenticates an admin user
func (h *Handler) AdminLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	adminRoles := []models.UserRole{models.RoleSystemAdmin, models.RoleTenantAdmin}

	var user models.User
	var userErr error

	if domain := h.resolveTenantDomain(c, req.Domain); domain != "" {
		var tenant models.Tenant
		if err := h.DB.Where("domain = ? AND enabled = true", domain).First(&tenant).Error; err != nil {
			log.WithField("domain", domain).Debug("AdminLogin: tenant not found for domain")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials or insufficient permissions"})
		}
		// Tenant-scoped lookup for tenant_admin; system_admin can log in from any domain
		userErr = h.DB.Where("(username = ? OR email = ?) AND role IN ? AND (tenant_id = ? OR role = ?)",
			req.Username, req.Username, adminRoles, tenant.ID, models.RoleSystemAdmin).First(&user).Error
	} else {
		userErr = h.DB.Where("(username = ? OR email = ?) AND role IN ?", req.Username, req.Username, adminRoles).First(&user).Error
	}

	if userErr != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials or insufficient permissions"})
	}

	if !user.CheckPassword(req.Password) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// Generate token
	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":        user.ID,
			"uuid":      user.UUID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"tenant_id": user.TenantID,
		},
	})
}

// Register creates a new user account
func (h *Handler) Register(c *fiber.Ctx) error {
	// NOTE: Implement based on your registration requirements
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error": "Registration not implemented"})
}

// ExtensionLoginRequest represents a client/extension login payload
type ExtensionLoginRequest struct {
	Extension    string `json:"extension"`
	Password     string `json:"password"`
	Domain       string `json:"domain"`
	EndpointType string `json:"endpoint_type"` // web_client, mobile_app, desktop_app
	InstanceID   string `json:"instance_id"`   // Client-generated unique ID
	DeviceLabel  string `json:"device_label"`  // e.g. "Chrome Browser"
	AppVersion   string `json:"app_version"`
	OSInfo       string `json:"os_info"`
}

// ExtensionLogin authenticates an extension user and returns a JWT + SIP credentials
func (h *Handler) ExtensionLogin(c *fiber.Ctx) error {
	var req ExtensionLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.Extension == "" || req.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Extension and password are required"})
	}

	// Resolve tenant from the connected domain
	domain := h.resolveTenantDomain(c, req.Domain)
	if domain == "" {
		// Extension login always requires a tenant context
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Tenant domain is required for extension login"})
	}

	var tenant models.Tenant
	if err := h.DB.Where("domain = ? AND enabled = true", domain).First(&tenant).Error; err != nil {
		log.WithField("domain", domain).Debug("ExtensionLogin: tenant not found for domain")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Look up extension within this tenant
	var ext models.Extension
	if err := h.DB.Where("extension = ? AND tenant_id = ? AND enabled = true", req.Extension, tenant.ID).First(&ext).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Verify web login password
	if !ext.CheckWebPassword(req.Password) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate JWT with extension context
	token, err := h.Auth.GenerateExtensionToken(&ext)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	// Determine endpoint type, default to web_client
	endpointType := models.EndpointTypeWebClient
	switch models.EndpointType(req.EndpointType) {
	case models.EndpointTypeMobileApp, models.EndpointTypeDesktopApp, models.EndpointTypeWebClient:
		endpointType = models.EndpointType(req.EndpointType)
	}

	// Auto-generate instance ID if not provided
	instanceID := req.InstanceID
	if instanceID == "" {
		instanceID = ext.UUID.String()[:8]
	}

	// Check for existing registration for this extension + instance + type
	var reg models.ClientRegistration
	reused := false
	if err := h.DB.Where(
		"extension_id = ? AND instance_id = ? AND endpoint_type = ? AND tenant_id = ?",
		ext.ID, instanceID, endpointType, tenant.ID,
	).First(&reg).Error; err == nil {
		reused = true
	} else {
		// Provision a new client registration
		regUser := models.GenerateRegistrationUser(endpointType, ext.Extension, instanceID)
		reg = models.ClientRegistration{
			TenantID:         tenant.ID,
			ExtensionID:      &ext.ID,
			EndpointType:     endpointType,
			RegistrationUser: regUser,
			InstanceID:       instanceID,
			DisplayName:      ext.EffectiveCallerIDName,
			DeviceLabel:      req.DeviceLabel,
			AppVersion:       req.AppVersion,
			OSInfo:           req.OSInfo,
			AllowOutbound:    true,
			WebRTC:           endpointType == models.EndpointTypeWebClient,
			Status:           "provisioned",
			Enabled:          true,
		}
		if ext.UserID != nil {
			reg.UserID = ext.UserID
		}

		if err := h.DB.Create(&reg).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to provision SIP credentials"})
		}
		h.reloadXML()
	}

	return c.JSON(fiber.Map{
		"token": token,
		"extension": fiber.Map{
			"id":        ext.ID,
			"uuid":      ext.UUID,
			"extension": ext.Extension,
			"tenant_id": ext.TenantID,
			"caller_id": ext.EffectiveCallerIDName,
		},
		"sip_user":     reg.RegistrationUser,
		"sip_password": reg.RegistrationPass,
		"sip_domain":   tenant.Domain,
		"reused":       reused,
	})
}

// RequestPasswordReset initiates a password reset
func (h *Handler) RequestPasswordReset(c *fiber.Ctx) error {
	// NOTE: Implement password reset logic
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error": "Password reset not implemented"})
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var user models.User
	if err := h.DB.Preload("Tenant").First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"id":           user.ID,
		"uuid":         user.UUID,
		"username":     user.Username,
		"email":        user.Email,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"phone_number": user.PhoneNumber,
		"extension":    user.Extension,
		"role":         user.Role,
		"tenant_id":    user.TenantID,
		"tenant":       user.Tenant,
		"last_login":   user.LastLogin,
		"created_at":   user.CreatedAt,
	})
}

// ChangePasswordRequest represents the change password payload
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword updates the authenticated user's password
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Current password is incorrect"})
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set password"})
	}

	if err := h.DB.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save password"})
	}

	return c.JSON(fiber.Map{"message": "Password updated successfully"})
}

// Logout invalidates the current token
func (h *Handler) Logout(c *fiber.Ctx) error {
	// NOTE: For JWT, logout is typically handled client-side
	// Optionally implement a token blacklist here
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

// RefreshToken generates a new token for the authenticated user
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{"token": token})
}
