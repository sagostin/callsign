package middleware

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// TenantMiddleware provides tenant scoping functionality
type TenantMiddleware struct {
	DB *gorm.DB
}

// NewTenantMiddleware creates a new tenant middleware instance
func NewTenantMiddleware(db *gorm.DB) *TenantMiddleware {
	return &TenantMiddleware{DB: db}
}

// RequireTenant ensures requests are scoped to a specific tenant
func (t *TenantMiddleware) RequireTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := GetClaims(c)
		if claims == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}

		// System admins can optionally specify a tenant via query param or header
		if claims.Role == "system_admin" {
			// Check for X-Tenant-ID header or tenant_id query param
			tenantIDHeader := c.Get("X-Tenant-ID")
			tenantIDQuery := c.Query("tenant_id")

			tenantIDStr := tenantIDHeader
			if tenantIDStr == "" {
				tenantIDStr = tenantIDQuery
			}

			if tenantIDStr != "" {
				// Parse string to uint and store as uint (not string)
				var tenantID uint64
				if _, err := fmt.Sscanf(tenantIDStr, "%d", &tenantID); err == nil && tenantID > 0 {
					c.Locals("scoped_tenant_id", uint(tenantID))
				}
			}
			// System admins can proceed without tenant scope for global operations
			return c.Next()
		}

		// Regular users must have a tenant ID
		if claims.TenantID == nil {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Tenant association required"})
		}

		// Verify tenant exists and is active
		var count int64
		t.DB.Table("tenants").
			Where("id = ? AND enabled = ? AND deleted_at IS NULL", *claims.TenantID, true).
			Count(&count)

		if count == 0 {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Tenant not found or disabled"})
		}

		c.Locals("scoped_tenant_id", *claims.TenantID)
		return c.Next()
	}
}

// GetScopedTenantID returns the tenant ID for the current request
// This accounts for system admins who may be viewing a specific tenant
func GetScopedTenantID(c *fiber.Ctx) uint {
	if scopedID, ok := c.Locals("scoped_tenant_id").(uint); ok && scopedID > 0 {
		return scopedID
	}
	return GetTenantID(c)
}

// ValidateTenantAccess ensures the user can access resources for the specified tenant
func ValidateTenantAccess(c *fiber.Ctx, resourceTenantID uint) bool {
	claims := GetClaims(c)
	if claims == nil {
		return false
	}

	// System admins can access any tenant
	if claims.Role == "system_admin" {
		return true
	}

	// Users can only access their own tenant
	return claims.TenantID != nil && *claims.TenantID == resourceTenantID
}
