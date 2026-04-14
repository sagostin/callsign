package handlers

import (
	"callsign/config"
	"callsign/middleware"
	"callsign/models"
	"callsign/services/broadcast"
	"callsign/services/cdr"
	"callsign/services/esl"
	"callsign/services/logging"
	"callsign/services/messaging"
	"callsign/services/websocket"
	"callsign/services/xmlcache"
	"crypto/rand"
	"encoding/base64"
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
	XMLCache            *xmlcache.XMLCache
	BroadcastWorker     *broadcast.BroadcastWorker
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

// SetXMLCache sets the XML cache reference for cache invalidation
func (h *Handler) SetXMLCache(cache *xmlcache.XMLCache) {
	h.XMLCache = cache
}

// SetBroadcastWorker sets the broadcast campaign worker reference
func (h *Handler) SetBroadcastWorker(worker *broadcast.BroadcastWorker) {
	h.BroadcastWorker = worker
}

// flushXMLCache invalidates the xmlcache so the next mod_xml_curl request
// from FreeSWITCH fetches fresh data from the database. Since all config
// (directory, dialplan, configuration, ACLs) is served dynamically via
// mod_xml_curl, no reloadxml command is needed — FreeSWITCH will pick up
// the changes on the next lookup. For cases where FreeSWITCH has cached
// stale data internally, use the manual "Reload XML" button in the UI.
func (h *Handler) flushXMLCache() {
	if h.XMLCache != nil {
		h.XMLCache.Flush()
		log.Debug("XML cache flushed")
	}
}

// reloadXML flushes the XML cache so FreeSWITCH picks up changes on next xml_curl request.
// No reloadxml ESL command is sent — all config is served dynamically via mod_xml_curl.
func (h *Handler) reloadXML() {
	h.flushXMLCache()
}

// reloadSofia flushes the XML cache and tells FreeSWITCH to reload XML and
// rescan the named Sofia profile so gateway/config changes take effect immediately.
func (h *Handler) reloadSofia(profileName string) {
	h.flushXMLCache()

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		log.Warn("reloadSofia: ESL not connected, skipping FreeSWITCH reload")
		return
	}

	// Reload XML so FreeSWITCH re-reads disk config
	if _, err := h.ESLManager.API("reloadxml"); err != nil {
		log.WithError(err).Warn("reloadSofia: failed to send reloadxml")
	}

	// Rescan the profile to pick up gateway changes without a full restart
	if profileName != "" {
		cmd := fmt.Sprintf("sofia profile %s rescan", profileName)
		if _, err := h.ESLManager.BgAPI(cmd); err != nil {
			log.WithError(err).WithField("profile", profileName).Warn("reloadSofia: failed to rescan profile")
		}
	}
}

// reloadCallcenter flushes the XML cache. Callcenter config is fetched dynamically.
func (h *Handler) reloadCallcenter() {
	h.flushXMLCache()
}

// reloadACL flushes the XML cache. ACL config is fetched dynamically via xml_curl.
func (h *Handler) reloadACL() {
	h.flushXMLCache()
}

// SetLogManager sets the LogManager reference
func (h *Handler) SetLogManager(lm *logging.LogManager) {
	h.LogManager = lm
}

// logRequest logs an API operation via the LogManager (if available).
func (h *Handler) logRequest(level log.Level, logType, message string, fields map[string]interface{}) {
	if h.LogManager != nil {
		h.LogManager.Log(level, logType, message, fields)
	}
}

// logInfo logs an info-level API operation.
func (h *Handler) logInfo(logType, message string, fields map[string]interface{}) {
	h.logRequest(log.InfoLevel, logType, message, fields)
}

// logWarn logs a warn-level API operation (e.g. validation failures, not-found).
func (h *Handler) logWarn(logType, message string, fields map[string]interface{}) {
	h.logRequest(log.WarnLevel, logType, message, fields)
}

// logError logs an error-level API operation (e.g. DB failures, internal errors).
func (h *Handler) logError(logType, message string, fields map[string]interface{}) {
	h.logRequest(log.ErrorLevel, logType, message, fields)
}

// reqFields builds a common fields map for request logging, combining request
// context (method, path, client IP) with any additional fields provided.
func (h *Handler) reqFields(c *fiber.Ctx, extra map[string]interface{}) map[string]interface{} {
	fields := map[string]interface{}{
		"method": c.Method(),
		"path":   c.Path(),
		"ip":     c.IP(),
	}
	for k, v := range extra {
		fields[k] = v
	}
	return fields
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
		h.logWarn("AUTH", "Login: invalid request payload", h.reqFields(c, map[string]interface{}{"error": err.Error()}))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Resolve tenant from the connected domain
	var user models.User
	var userErr error

	if domain := h.resolveTenantDomain(c, req.Domain); domain != "" {
		var tenant models.Tenant
		if err := h.DB.Where("domain = ? AND enabled = true", domain).First(&tenant).Error; err != nil {
			h.logWarn("AUTH", "Login: tenant not found for domain", h.reqFields(c, map[string]interface{}{"domain": domain, "username": req.Username}))
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		userErr = h.DB.Where("(username = ? OR email = ?) AND tenant_id = ?", req.Username, req.Username, tenant.ID).First(&user).Error
	} else {
		// No tenant resolved (localhost) – global lookup for backward compat
		userErr = h.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error
	}

	if userErr != nil {
		h.logWarn("AUTH", "Login: user not found", h.reqFields(c, map[string]interface{}{"username": req.Username, "domain": req.Domain}))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if !user.CheckPassword(req.Password) {
		h.logWarn("AUTH", "Login: invalid password", h.reqFields(c, map[string]interface{}{"username": req.Username, "user_id": user.ID}))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// Generate token
	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		h.logError("AUTH", "Login: failed to generate token", h.reqFields(c, map[string]interface{}{"error": err.Error(), "user_id": user.ID}))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	h.logInfo("AUTH", "Login: successful", h.reqFields(c, map[string]interface{}{"user_id": user.ID, "username": user.Username, "role": user.Role}))
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
		h.logWarn("AUTH", "AdminLogin: invalid request payload", h.reqFields(c, map[string]interface{}{"error": err.Error()}))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	adminRoles := []models.UserRole{models.RoleSystemAdmin, models.RoleTenantAdmin}

	var user models.User
	var userErr error

	if domain := h.resolveTenantDomain(c, req.Domain); domain != "" {
		var tenant models.Tenant
		if err := h.DB.Where("domain = ? AND enabled = true", domain).First(&tenant).Error; err != nil {
			// No tenant for this domain — allow global system_admin login only
			h.logInfo("AUTH", "AdminLogin: no tenant for domain, trying global system_admin lookup", h.reqFields(c, map[string]interface{}{"domain": domain, "username": req.Username}))
			userErr = h.DB.Where("(username = ? OR email = ?) AND role = ?",
				req.Username, req.Username, models.RoleSystemAdmin).First(&user).Error
		} else {
			// Tenant-scoped lookup for tenant_admin; system_admin can log in from any domain
			userErr = h.DB.Where("(username = ? OR email = ?) AND role IN ? AND (tenant_id = ? OR role = ?)",
				req.Username, req.Username, adminRoles, tenant.ID, models.RoleSystemAdmin).First(&user).Error
		}
	} else {
		userErr = h.DB.Where("(username = ? OR email = ?) AND role IN ?", req.Username, req.Username, adminRoles).First(&user).Error
	}

	if userErr != nil {
		h.logWarn("AUTH", "AdminLogin: admin user not found or insufficient permissions", h.reqFields(c, map[string]interface{}{"username": req.Username, "domain": req.Domain}))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials or insufficient permissions"})
	}

	if !user.CheckPassword(req.Password) {
		h.logWarn("AUTH", "AdminLogin: invalid password", h.reqFields(c, map[string]interface{}{"username": req.Username, "user_id": user.ID}))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// Generate token
	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		h.logError("AUTH", "AdminLogin: failed to generate token", h.reqFields(c, map[string]interface{}{"error": err.Error(), "user_id": user.ID}))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	h.logInfo("AUTH", "AdminLogin: successful", h.reqFields(c, map[string]interface{}{"user_id": user.ID, "username": user.Username, "role": user.Role}))
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

// RegisterRequest represents a user registration payload
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Register creates a new user account
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		h.logWarn("AUTH", "Register: invalid request payload", h.reqFields(c, map[string]interface{}{"error": err.Error()}))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Guard: validate required fields are present
	if req.Email == "" || req.Password == "" || req.Username == "" {
		h.logWarn("AUTH", "Register: missing required fields", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Email, password, and username are required"})
	}

	// Guard: validate email format
	if !isValidEmail(req.Email) {
		h.logWarn("AUTH", "Register: invalid email format", h.reqFields(c, map[string]interface{}{"email": req.Email}))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email format"})
	}

	// Guard: validate password strength
	if len(req.Password) < 8 {
		h.logWarn("AUTH", "Register: password too short", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}

	// Guard: check if user with email already exists
	var existingUser models.User
	if err := h.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		h.logWarn("AUTH", "Register: user already exists", h.reqFields(c, map[string]interface{}{"email": req.Email, "username": req.Username}))
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "User with this email or username already exists"})
	}

	// Create new user with hashed password
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      models.RoleUser, // Default role for self-registration
	}

	if err := user.SetPassword(req.Password); err != nil {
		h.logError("AUTH", "Register: failed to hash password", h.reqFields(c, map[string]interface{}{"error": err.Error()}))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process registration"})
	}

	if err := h.DB.Create(&user).Error; err != nil {
		h.logError("AUTH", "Register: failed to create user", h.reqFields(c, map[string]interface{}{"error": err.Error()}))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	h.logInfo("AUTH", "Register: successful", h.reqFields(c, map[string]interface{}{"user_id": user.ID, "username": user.Username, "email": user.Email}))
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user": fiber.Map{
			"id":         user.ID,
			"uuid":       user.UUID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}

// isValidEmail performs basic email format validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Basic format check: contains @ and has domain portion
	atIndex := strings.Index(email, "@")
	if atIndex < 1 || atIndex == len(email)-1 {
		return false
	}
	domain := email[atIndex+1:]
	if domain == "" || !strings.Contains(domain, ".") {
		return false
	}
	return true
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
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Fail fast: validate email presence
	if strings.TrimSpace(req.Email) == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Email is required"})
	}

	// Normalize and validate email format
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !isValidEmail(email) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email format"})
	}

	// Look up user by email (silently fail to prevent email enumeration)
	var user models.User
	if err := h.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.JSON(fiber.Map{"message": "If an account exists, a reset link has been sent"})
	}

	// Generate cryptographically secure reset token
	token, err := generateSecureToken(32)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate reset token"})
	}

	// TODO: Store token in database with expiration (password_reset_tokens table)
	// Log token for testing purposes
	h.logInfo("Auth", "PasswordReset", map[string]interface{}{
		"user_id": user.ID,
		"email":   email,
		"token":   token,
	})

	// TODO: Send reset email via email service
	// For testing, return token in response
	return c.JSON(fiber.Map{
		"message": "If an account exists, a reset link has been sent",
		"token":   token, // Remove in production - email the token instead
	})
}

// generateSecureToken creates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
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
