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
	"net/http"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
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
// Health & Status
// =====================

// Health returns the health status of the API
func (h *Handler) Health(ctx iris.Context) {
	ctx.JSON(iris.Map{
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
func (h *Handler) resolveTenantDomain(ctx iris.Context, explicit string) string {
	domain := strings.TrimSpace(explicit)
	if domain == "" {
		domain = ctx.Host()
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
func (h *Handler) Login(ctx iris.Context) {
	var req LoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Resolve tenant from the connected domain
	var user models.User
	var userErr error

	if domain := h.resolveTenantDomain(ctx, req.Domain); domain != "" {
		var tenant models.Tenant
		if err := h.DB.Where("domain = ? AND enabled = true", domain).First(&tenant).Error; err != nil {
			log.WithField("domain", domain).Debug("Login: tenant not found for domain")
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Invalid credentials"})
			return
		}
		userErr = h.DB.Where("(username = ? OR email = ?) AND tenant_id = ?", req.Username, req.Username, tenant.ID).First(&user).Error
	} else {
		// No tenant resolved (localhost) – global lookup for backward compat
		userErr = h.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error
	}

	if userErr != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(req.Password) {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// Generate token
	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(iris.Map{
		"token": token,
		"user": iris.Map{
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
func (h *Handler) AdminLogin(ctx iris.Context) {
	var req LoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	adminRoles := []models.UserRole{models.RoleSystemAdmin, models.RoleTenantAdmin}

	var user models.User
	var userErr error

	if domain := h.resolveTenantDomain(ctx, req.Domain); domain != "" {
		var tenant models.Tenant
		if err := h.DB.Where("domain = ? AND enabled = true", domain).First(&tenant).Error; err != nil {
			log.WithField("domain", domain).Debug("AdminLogin: tenant not found for domain")
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Invalid credentials or insufficient permissions"})
			return
		}
		// Tenant-scoped lookup for tenant_admin; system_admin can log in from any domain
		userErr = h.DB.Where("(username = ? OR email = ?) AND role IN ? AND (tenant_id = ? OR role = ?)",
			req.Username, req.Username, adminRoles, tenant.ID, models.RoleSystemAdmin).First(&user).Error
	} else {
		userErr = h.DB.Where("(username = ? OR email = ?) AND role IN ?", req.Username, req.Username, adminRoles).First(&user).Error
	}

	if userErr != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials or insufficient permissions"})
		return
	}

	if !user.CheckPassword(req.Password) {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// Generate token
	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(iris.Map{
		"token": token,
		"user": iris.Map{
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
func (h *Handler) Register(ctx iris.Context) {
	// NOTE: Implement based on your registration requirements
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Registration not implemented"})
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
func (h *Handler) ExtensionLogin(ctx iris.Context) {
	var req ExtensionLoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if req.Extension == "" || req.Password == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Extension and password are required"})
		return
	}

	// Resolve tenant from the connected domain
	domain := h.resolveTenantDomain(ctx, req.Domain)
	if domain == "" {
		// Extension login always requires a tenant context
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant domain is required for extension login"})
		return
	}

	var tenant models.Tenant
	if err := h.DB.Where("domain = ? AND enabled = true", domain).First(&tenant).Error; err != nil {
		log.WithField("domain", domain).Debug("ExtensionLogin: tenant not found for domain")
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	// Look up extension within this tenant
	var ext models.Extension
	if err := h.DB.Where("extension = ? AND tenant_id = ? AND enabled = true", req.Extension, tenant.ID).First(&ext).Error; err != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	// Verify web login password
	if !ext.CheckWebPassword(req.Password) {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	// Generate JWT with extension context
	token, err := h.Auth.GenerateExtensionToken(&ext)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
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
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "Failed to provision SIP credentials"})
			return
		}
		h.reloadXML()
	}

	ctx.JSON(iris.Map{
		"token": token,
		"extension": iris.Map{
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
func (h *Handler) RequestPasswordReset(ctx iris.Context) {
	// NOTE: Implement password reset logic
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Password reset not implemented"})
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var user models.User
	if err := h.DB.Preload("Tenant").First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	ctx.JSON(iris.Map{
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
func (h *Handler) ChangePassword(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	if !user.CheckPassword(req.CurrentPassword) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Current password is incorrect"})
		return
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to set password"})
		return
	}

	if err := h.DB.Save(&user).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save password"})
		return
	}

	ctx.JSON(iris.Map{"message": "Password updated successfully"})
}

// Logout invalidates the current token
func (h *Handler) Logout(ctx iris.Context) {
	// NOTE: For JWT, logout is typically handled client-side
	// Optionally implement a token blacklist here
	ctx.JSON(iris.Map{"message": "Logged out successfully"})
}

// RefreshToken generates a new token for the authenticated user
func (h *Handler) RefreshToken(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(iris.Map{"token": token})
}
