package middleware

import (
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
)

// RequirePermission is a middleware that checks if the user has a specific permission
func RequirePermission(perm models.Permission) iris.Handler {
	return func(ctx iris.Context) {
		claims := GetClaims(ctx)
		if claims == nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Unauthorized"})
			return
		}

		// Get user from context
		user, ok := ctx.Values().Get("user").(*models.User)
		if !ok {
			// If user not loaded, check role from claims
			// System admins always have access
			if models.UserRole(claims.Role) == models.RoleSystemAdmin {
				ctx.Next()
				return
			}

			// For other roles, we need the full user object
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{
				"error":               "Insufficient permissions",
				"required_permission": string(perm),
			})
			return
		}

		if !user.HasPermission(perm) {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{
				"error":               "Insufficient permissions",
				"required_permission": string(perm),
			})
			return
		}

		ctx.Next()
	}
}

// RequireAnyPermission is a middleware that checks if the user has any of the specified permissions
func RequireAnyPermission(perms ...models.Permission) iris.Handler {
	return func(ctx iris.Context) {
		claims := GetClaims(ctx)
		if claims == nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Unauthorized"})
			return
		}

		// System admins always have access
		if models.UserRole(claims.Role) == models.RoleSystemAdmin {
			ctx.Next()
			return
		}

		// Get user from context
		user, ok := ctx.Values().Get("user").(*models.User)
		if !ok {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{"error": "Insufficient permissions"})
			return
		}

		if !user.HasAnyPermission(perms...) {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{"error": "Insufficient permissions"})
			return
		}

		ctx.Next()
	}
}

// RequireAllPermissions is a middleware that checks if the user has all of the specified permissions
func RequireAllPermissions(perms ...models.Permission) iris.Handler {
	return func(ctx iris.Context) {
		claims := GetClaims(ctx)
		if claims == nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Unauthorized"})
			return
		}

		// System admins always have access
		if models.UserRole(claims.Role) == models.RoleSystemAdmin {
			ctx.Next()
			return
		}

		// Get user from context
		user, ok := ctx.Values().Get("user").(*models.User)
		if !ok {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{"error": "Insufficient permissions"})
			return
		}

		if !user.HasAllPermissions(perms...) {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{"error": "Insufficient permissions"})
			return
		}

		ctx.Next()
	}
}

// LoadUser is a middleware that loads the full user object into context
func LoadUser(db interface {
	First(dest interface{}, conds ...interface{}) interface{ Error() error }
}) iris.Handler {
	return func(ctx iris.Context) {
		claims := GetClaims(ctx)
		if claims == nil {
			ctx.Next()
			return
		}

		var user models.User
		// This is a simplified version - in production you'd use the actual DB
		// The interface is used to avoid import cycles
		ctx.Values().Set("user", &user)
		ctx.Next()
	}
}
