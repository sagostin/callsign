package middleware

import (
	"fmt"
	"net/http"

	"github.com/kataras/iris/v12"
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
func (t *TenantMiddleware) RequireTenant() iris.Handler {
	return func(ctx iris.Context) {
		claims := GetClaims(ctx)
		if claims == nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Authentication required"})
			return
		}

		// System admins can optionally specify a tenant via query param or header
		if claims.Role == "system_admin" {
			// Check for X-Tenant-ID header or tenant_id query param
			tenantIDHeader := ctx.GetHeader("X-Tenant-ID")
			tenantIDQuery := ctx.URLParam("tenant_id")

			tenantIDStr := tenantIDHeader
			if tenantIDStr == "" {
				tenantIDStr = tenantIDQuery
			}

			if tenantIDStr != "" {
				// Parse string to uint and store as uint (not string)
				var tenantID uint64
				if _, err := fmt.Sscanf(tenantIDStr, "%d", &tenantID); err == nil && tenantID > 0 {
					ctx.Values().Set("scoped_tenant_id", uint(tenantID))
				}
			}
			// System admins can proceed without tenant scope for global operations
			ctx.Next()
			return
		}

		// Regular users must have a tenant ID
		if claims.TenantID == nil {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{"error": "Tenant association required"})
			return
		}

		// Verify tenant exists and is active
		var count int64
		t.DB.Table("tenants").
			Where("id = ? AND enabled = ? AND deleted_at IS NULL", *claims.TenantID, true).
			Count(&count)

		if count == 0 {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{"error": "Tenant not found or disabled"})
			return
		}

		ctx.Values().Set("scoped_tenant_id", *claims.TenantID)
		ctx.Next()
	}
}

// GetScopedTenantID returns the tenant ID for the current request
// This accounts for system admins who may be viewing a specific tenant
func GetScopedTenantID(ctx iris.Context) uint {
	if scopedID := ctx.Values().GetUintDefault("scoped_tenant_id", 0); scopedID > 0 {
		return scopedID
	}
	return GetTenantID(ctx)
}

// ValidateTenantAccess ensures the user can access resources for the specified tenant
func ValidateTenantAccess(ctx iris.Context, resourceTenantID uint) bool {
	claims := GetClaims(ctx)
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
