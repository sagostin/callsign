package middleware

import (
	"callsign/models"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

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

		// Log audit event after handler completes
		go logAuditEvent(db, c, tenantID, userID, username, role, newValue)

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

func logAuditEvent(db *gorm.DB, c *fiber.Ctx, tenantID, userID uint, username, role string, newValue interface{}) {
	method := c.Method()
	path := c.Path()
	status := c.Response().StatusCode()

	// Determine action
	var action models.AuditAction
	switch method {
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
	resource, resourceID := parseResourceFromPath(path)

	// Check for auth endpoints
	if strings.Contains(path, "/auth/login") {
		action = models.AuditActionLogin
		resource = "auth"
	} else if strings.Contains(path, "/auth/logout") {
		action = models.AuditActionLogout
		resource = "auth"
	}

	// Get old value if it was stored by handler
	oldValue := c.Locals("audit_old_value")

	// Create audit log
	entry := &models.AuditLog{
		TenantID:   tenantID,
		UserID:     userID,
		Username:   username,
		UserRole:   role,
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Success:    status >= 200 && status < 400,
	}

	if oldValue != nil {
		if data, err := json.Marshal(oldValue); err == nil {
			entry.OldValue = data
		}
	}
	if newValue != nil {
		if data, err := json.Marshal(newValue); err == nil {
			entry.NewValue = data
		}
	}

	// Check for error message
	if status >= 400 {
		if errMsg, ok := c.Locals("error").(string); ok {
			entry.Error = errMsg
		}
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
