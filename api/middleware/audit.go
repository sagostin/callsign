package middleware

import (
	"callsign/models"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// auditSnapshot holds all context values captured before the goroutine runs.
// Fiber recycles *fiber.Ctx after the handler returns, so we must not access
// the context from a goroutine — doing so causes nil-pointer panics.
type auditSnapshot struct {
	Method    string
	Path      string
	Status    int
	IP        string
	UserAgent string
	TenantID  uint
	UserID    uint
	Username  string
	Role      string
	NewValue  interface{}
	OldValue  interface{}
	Error     string
}

// AuditMiddleware creates a middleware that logs audit events
func AuditMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip for certain paths
		path := c.Path()
		if shouldSkipAudit(path) {
			return c.Next()
		}

		// Only audit write operations
		method := c.Method()
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			return c.Next()
		}

		// Get user info from context (set by auth middleware)
		claims, _ := c.Locals("claims").(*Claims)

		var tenantID, userID uint
		var username, role string

		if claims != nil {
			tenantID = GetTenantID(c) // Use helper to get scoped or actual tenant ID
			userID = claims.UserID
			username = claims.Username
			role = string(claims.Role)
		}

		// Get request body for new value
		var newValue interface{}
		if method == "POST" || method == "PUT" || method == "PATCH" {
			body := c.Body()
			if len(body) > 0 {
				json.Unmarshal(body, &newValue)
			}
		}

		// Store pre-request state
		c.Locals("audit_tenant_id", tenantID)
		c.Locals("audit_user_id", userID)
		c.Locals("audit_username", username)
		c.Locals("audit_role", role)
		c.Locals("audit_new_value", newValue)

		// Process request
		err := c.Next()

		// Capture everything we need from the context BEFORE launching the
		// goroutine.  Fiber recycles *fiber.Ctx objects after the handler
		// returns, so accessing c from a goroutine is a use-after-free bug
		// that causes nil-pointer panics (the 502s we were seeing).
		snap := auditSnapshot{
			Method:    method,
			Path:      path,
			Status:    c.Response().StatusCode(),
			IP:        c.IP(),
			UserAgent: c.Get("User-Agent"),
			TenantID:  tenantID,
			UserID:    userID,
			Username:  username,
			Role:      role,
			NewValue:  newValue,
			OldValue:  c.Locals("audit_old_value"),
		}
		if snap.Status >= 400 {
			if errMsg, ok := c.Locals("error").(string); ok {
				snap.Error = errMsg
			}
		}

		go logAuditEvent(db, snap)

		return err
	}
}

func shouldSkipAudit(path string) bool {
	// Skip health checks, static files, etc.
	skipPaths := []string{
		"/health",
		"/api/health",
		"/metrics",
		"/swagger",
		"/favicon",
	}

	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

func logAuditEvent(db *gorm.DB, s auditSnapshot) {
	// Determine action
	var action models.AuditAction
	switch s.Method {
	case "POST":
		action = models.AuditActionCreate
	case "PUT", "PATCH":
		action = models.AuditActionUpdate
	case "DELETE":
		action = models.AuditActionDelete
	default:
		return
	}

	// Determine resource from path
	resource, resourceID := parseResourceFromPath(s.Path)

	// Check for auth endpoints
	if strings.Contains(s.Path, "/auth/login") {
		action = models.AuditActionLogin
		resource = "auth"
	} else if strings.Contains(s.Path, "/auth/logout") {
		action = models.AuditActionLogout
		resource = "auth"
	}

	// Create audit log
	entry := &models.AuditLog{
		TenantID:   s.TenantID,
		UserID:     s.UserID,
		Username:   s.Username,
		UserRole:   s.Role,
		IPAddress:  s.IP,
		UserAgent:  s.UserAgent,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Success:    s.Status >= 200 && s.Status < 400,
	}

	if s.OldValue != nil {
		if data, err := json.Marshal(s.OldValue); err == nil {
			entry.OldValue = data
		}
	}
	if s.NewValue != nil {
		if data, err := json.Marshal(s.NewValue); err == nil {
			entry.NewValue = data
		}
	}

	// Check for error message
	if s.Status >= 400 {
		entry.Error = s.Error
	}

	db.Create(entry)
}

func parseResourceFromPath(path string) (resource, resourceID string) {
	// Parse /api/v1/resource/id format
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")

	if len(parts) > 0 {
		resource = parts[0]
	}
	if len(parts) > 1 {
		resourceID = parts[1]
	}

	return
}

// SetOldValue is a helper to store old value for audit logging
// Call this in handlers before updating a resource
func SetOldValue(c *fiber.Ctx, oldValue interface{}) {
	c.Locals("audit_old_value", oldValue)
}
