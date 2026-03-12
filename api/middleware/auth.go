package middleware

import (
	"callsign/config"
	"callsign/models"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID      uint            `json:"user_id"`
	Username    string          `json:"username"`
	Email       string          `json:"email"`
	Role        models.UserRole `json:"role"`
	TenantID    *uint           `json:"tenant_id,omitempty"`
	ExtensionID *uint           `json:"extension_id,omitempty"`
	jwt.RegisteredClaims
}

// AuthMiddleware provides JWT authentication functionality
type AuthMiddleware struct {
	Config *config.Config
	DB     *gorm.DB
}

// NewAuthMiddleware creates a new authentication middleware instance
func NewAuthMiddleware(cfg *config.Config, db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{
		Config: cfg,
		DB:     db,
	}
}

// GenerateToken creates a new JWT token for a user
func (a *AuthMiddleware) GenerateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(a.Config.JWTExpiration) * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		TenantID: user.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "callsign",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.Config.JWTSecret))
}

// GenerateExtensionToken creates a JWT token for an extension-based login session
func (a *AuthMiddleware) GenerateExtensionToken(ext *models.Extension) (string, error) {
	expirationTime := time.Now().Add(time.Duration(a.Config.JWTExpiration) * time.Hour)
	tenantID := ext.TenantID

	claims := &Claims{
		UserID:      0, // No User model association
		Username:    ext.Extension,
		Email:       "",
		Role:        models.RoleUser,
		TenantID:    &tenantID,
		ExtensionID: &ext.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "callsign",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.Config.JWTSecret))
}

// VerifyToken validates a JWT token and returns the claims
func (a *AuthMiddleware) VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.Config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// RequireAuth is middleware that requires a valid JWT token
func (a *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header required"})
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization header format"})
		}

		tokenString := parts[1]
		claims, err := a.VerifyToken(tokenString)
		if err != nil {
			log.Warnf("Token verification failed: %v", err)
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		// Store claims in context for downstream handlers
		c.Locals("claims", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		if claims.TenantID != nil {
			c.Locals("tenant_id", *claims.TenantID)
		}

		return c.Next()
	}
}

// RequireRole is middleware that requires a specific role
func (a *AuthMiddleware) RequireRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*Claims)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}

		// System admin has access to everything
		if claims.Role == models.RoleSystemAdmin {
			return c.Next()
		}

		// Check if user has one of the required roles
		for _, role := range roles {
			if claims.Role == role {
				return c.Next()
			}
		}

		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}
}

// RequireSystemAdmin is middleware that requires system admin role
func (a *AuthMiddleware) RequireSystemAdmin() fiber.Handler {
	return a.RequireRole(models.RoleSystemAdmin)
}

// RequireTenantAdmin is middleware that requires tenant admin or higher
func (a *AuthMiddleware) RequireTenantAdmin() fiber.Handler {
	return a.RequireRole(models.RoleSystemAdmin, models.RoleTenantAdmin)
}

// GetClaims extracts the claims from the context
func GetClaims(c *fiber.Ctx) *Claims {
	if claims, ok := c.Locals("claims").(*Claims); ok {
		return claims
	}
	return nil
}

// GetUserID extracts the user ID from the context
func GetUserID(c *fiber.Ctx) uint {
	if id, ok := c.Locals("user_id").(uint); ok {
		return id
	}
	return 0
}

// GetTenantID extracts the tenant ID from the context
// It first checks scoped_tenant_id (set from X-Tenant-ID header for system_admin)
// then falls back to tenant_id from JWT claims
func GetTenantID(c *fiber.Ctx) uint {
	// Check scoped tenant ID first (system admin with X-Tenant-ID header)
	if scopedID, ok := c.Locals("scoped_tenant_id").(uint); ok && scopedID > 0 {
		return scopedID
	}
	if id, ok := c.Locals("tenant_id").(uint); ok {
		return id
	}
	return 0
}

// GetExtensionID extracts the extension ID from the JWT claims.
// Returns 0 when the token was not generated via extension login.
func GetExtensionID(c *fiber.Ctx) uint {
	claims := GetClaims(c)
	if claims != nil && claims.ExtensionID != nil {
		return *claims.ExtensionID
	}
	return 0
}

// GetRole extracts the role from the context
func GetRole(c *fiber.Ctx) models.UserRole {
	if role, ok := c.Locals("role").(models.UserRole); ok {
		return role
	}
	return models.RoleUser
}

// RequirePermission is middleware that requires specific permissions
func (a *AuthMiddleware) RequirePermission(perms ...models.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*Claims)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}

		// System admin has all permissions
		if claims.Role == models.RoleSystemAdmin {
			return c.Next()
		}

		// Get user from database to check permissions
		var user models.User
		if err := a.DB.First(&user, claims.UserID).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
		}

		// Check if user has any of the required permissions
		if user.HasAnyPermission(perms...) {
			return c.Next()
		}

		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error":    "Insufficient permissions",
			"required": perms,
		})
	}
}

// RequireAllPermissions is middleware that requires ALL specified permissions
func (a *AuthMiddleware) RequireAllPermissions(perms ...models.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*Claims)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}

		// System admin has all permissions
		if claims.Role == models.RoleSystemAdmin {
			return c.Next()
		}

		// Get user to check permissions
		var user models.User
		if err := a.DB.First(&user, claims.UserID).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
		}

		// Check if user has ALL required permissions
		if user.HasAllPermissions(perms...) {
			return c.Next()
		}

		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error":    "Insufficient permissions",
			"required": perms,
		})
	}
}
