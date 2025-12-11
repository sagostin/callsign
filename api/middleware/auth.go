package middleware

import (
	"callsign/config"
	"callsign/models"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID   uint            `json:"user_id"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Role     models.UserRole `json:"role"`
	TenantID *uint           `json:"tenant_id,omitempty"`
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
func (a *AuthMiddleware) RequireAuth() iris.Handler {
	return func(ctx iris.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Authorization header required"})
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := a.VerifyToken(tokenString)
		if err != nil {
			log.Warnf("Token verification failed: %v", err)
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Invalid or expired token"})
			return
		}

		// Store claims in context for downstream handlers
		ctx.Values().Set("claims", claims)
		ctx.Values().Set("user_id", claims.UserID)
		ctx.Values().Set("username", claims.Username)
		ctx.Values().Set("role", claims.Role)
		if claims.TenantID != nil {
			ctx.Values().Set("tenant_id", *claims.TenantID)
		}

		ctx.Next()
	}
}

// RequireRole is middleware that requires a specific role
func (a *AuthMiddleware) RequireRole(roles ...models.UserRole) iris.Handler {
	return func(ctx iris.Context) {
		claims, ok := ctx.Values().Get("claims").(*Claims)
		if !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Authentication required"})
			return
		}

		// System admin has access to everything
		if claims.Role == models.RoleSystemAdmin {
			ctx.Next()
			return
		}

		// Check if user has one of the required roles
		for _, role := range roles {
			if claims.Role == role {
				ctx.Next()
				return
			}
		}

		ctx.StatusCode(http.StatusForbidden)
		ctx.JSON(iris.Map{"error": "Insufficient permissions"})
	}
}

// RequireSystemAdmin is middleware that requires system admin role
func (a *AuthMiddleware) RequireSystemAdmin() iris.Handler {
	return a.RequireRole(models.RoleSystemAdmin)
}

// RequireTenantAdmin is middleware that requires tenant admin or higher
func (a *AuthMiddleware) RequireTenantAdmin() iris.Handler {
	return a.RequireRole(models.RoleSystemAdmin, models.RoleTenantAdmin)
}

// GetClaims extracts the claims from the context
func GetClaims(ctx iris.Context) *Claims {
	if claims, ok := ctx.Values().Get("claims").(*Claims); ok {
		return claims
	}
	return nil
}

// GetUserID extracts the user ID from the context
func GetUserID(ctx iris.Context) uint {
	return ctx.Values().GetUintDefault("user_id", 0)
}

// GetTenantID extracts the tenant ID from the context
func GetTenantID(ctx iris.Context) uint {
	return ctx.Values().GetUintDefault("tenant_id", 0)
}

// GetRole extracts the role from the context
func GetRole(ctx iris.Context) models.UserRole {
	if role, ok := ctx.Values().Get("role").(models.UserRole); ok {
		return role
	}
	return models.RoleUser
}

// RequirePermission is middleware that requires specific permissions
func (a *AuthMiddleware) RequirePermission(perms ...models.Permission) iris.Handler {
	return func(ctx iris.Context) {
		claims, ok := ctx.Values().Get("claims").(*Claims)
		if !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Authentication required"})
			return
		}

		// System admin has all permissions
		if claims.Role == models.RoleSystemAdmin {
			ctx.Next()
			return
		}

		// Get user from database to check permissions
		var user models.User
		if err := a.DB.First(&user, claims.UserID).Error; err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		// Check if user has any of the required permissions
		if user.HasAnyPermission(perms...) {
			ctx.Next()
			return
		}

		ctx.StatusCode(http.StatusForbidden)
		ctx.JSON(iris.Map{
			"error":    "Insufficient permissions",
			"required": perms,
		})
	}
}

// RequireAllPermissions is middleware that requires ALL specified permissions
func (a *AuthMiddleware) RequireAllPermissions(perms ...models.Permission) iris.Handler {
	return func(ctx iris.Context) {
		claims, ok := ctx.Values().Get("claims").(*Claims)
		if !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Authentication required"})
			return
		}

		// System admin has all permissions
		if claims.Role == models.RoleSystemAdmin {
			ctx.Next()
			return
		}

		// Get user to check permissions
		var user models.User
		if err := a.DB.First(&user, claims.UserID).Error; err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		// Check if user has ALL required permissions
		if user.HasAllPermissions(perms...) {
			ctx.Next()
			return
		}

		ctx.StatusCode(http.StatusForbidden)
		ctx.JSON(iris.Map{
			"error":    "Insufficient permissions",
			"required": perms,
		})
	}
}
