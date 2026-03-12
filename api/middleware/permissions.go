package middleware

import (
	"callsign/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission is a middleware that checks if the user has a specific permission
func RequirePermission(perm models.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := GetClaims(c)
		if claims == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		// Get user from context
		user, ok := c.Locals("user").(*models.User)
		if !ok {
			// If user not loaded, check role from claims
			// System admins always have access
			if models.UserRole(claims.Role) == models.RoleSystemAdmin {
				return c.Next()
			}

			// For other roles, we need the full user object
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error":               "Insufficient permissions",
				"required_permission": string(perm),
			})
		}

		if !user.HasPermission(perm) {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error":               "Insufficient permissions",
				"required_permission": string(perm),
			})
		}

		return c.Next()
	}
}

// RequireAnyPermission is a middleware that checks if the user has any of the specified permissions
func RequireAnyPermission(perms ...models.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := GetClaims(c)
		if claims == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		// System admins always have access
		if models.UserRole(claims.Role) == models.RoleSystemAdmin {
			return c.Next()
		}

		// Get user from context
		user, ok := c.Locals("user").(*models.User)
		if !ok {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
		}

		if !user.HasAnyPermission(perms...) {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
		}

		return c.Next()
	}
}

// RequireAllPermissions is a middleware that checks if the user has all of the specified permissions
func RequireAllPermissions(perms ...models.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := GetClaims(c)
		if claims == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		// System admins always have access
		if models.UserRole(claims.Role) == models.RoleSystemAdmin {
			return c.Next()
		}

		// Get user from context
		user, ok := c.Locals("user").(*models.User)
		if !ok {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
		}

		if !user.HasAllPermissions(perms...) {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
		}

		return c.Next()
	}
}

// LoadUser is a middleware that loads the full user object into context
func LoadUser(db interface {
	First(dest interface{}, conds ...interface{}) interface{ Error() error }
}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := GetClaims(c)
		if claims == nil {
			return c.Next()
		}

		var user models.User
		// This is a simplified version - in production you'd use the actual DB
		// The interface is used to avoid import cycles
		c.Locals("user", &user)
		return c.Next()
	}
}
