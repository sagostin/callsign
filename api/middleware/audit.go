package middleware

import (
	"bytes"
	"callsign/models"
	"encoding/json"
	"io"
	"strings"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// AuditMiddleware creates a middleware that logs audit events
func AuditMiddleware(db *gorm.DB) iris.Handler {
	return func(ctx iris.Context) {
		// Skip for certain paths
		path := ctx.Path()
		if shouldSkipAudit(path) {
			ctx.Next()
			return
		}

		// Only audit write operations
		method := ctx.Method()
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			ctx.Next()
			return
		}

		// Get user info from context (set by auth middleware)
		claims, _ := ctx.Values().Get("claims").(*Claims)

		var tenantID, userID uint
		var username, role string

		if claims != nil {
			tenantID = GetTenantID(ctx) // Use helper to get scoped or actual tenant ID
			userID = claims.UserID
			username = claims.Username
			role = string(claims.Role)
		}

		// Get request body for new value
		var newValue interface{}
		if method == "POST" || method == "PUT" || method == "PATCH" {
			// Read body bytes
			body, _ := ctx.GetBody()
			if len(body) > 0 {
				json.Unmarshal(body, &newValue)

				// Restore body for downstream handlers
				ctx.Request().Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		// Store pre-request state
		ctx.Values().Set("audit_tenant_id", tenantID)
		ctx.Values().Set("audit_user_id", userID)
		ctx.Values().Set("audit_username", username)
		ctx.Values().Set("audit_role", role)
		ctx.Values().Set("audit_new_value", newValue)

		// Process request
		ctx.Next()

		// Log audit event after handler completes
		go logAuditEvent(db, ctx, tenantID, userID, username, role, newValue)
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

func logAuditEvent(db *gorm.DB, ctx iris.Context, tenantID, userID uint, username, role string, newValue interface{}) {
	method := ctx.Method()
	path := ctx.Path()
	status := ctx.GetStatusCode()

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
	oldValue := ctx.Values().Get("audit_old_value")

	// Create audit log
	entry := &models.AuditLog{
		TenantID:   tenantID,
		UserID:     userID,
		Username:   username,
		UserRole:   role,
		IPAddress:  ctx.RemoteAddr(),
		UserAgent:  ctx.GetHeader("User-Agent"),
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
		if errMsg, ok := ctx.Values().Get("error").(string); ok {
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
func SetOldValue(ctx iris.Context, oldValue interface{}) {
	ctx.Values().Set("audit_old_value", oldValue)
}
