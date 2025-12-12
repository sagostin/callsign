package handlers

import (
	"callsign/config"
	"callsign/middleware"
	"callsign/models"
	"callsign/services/esl"
	"callsign/services/logging"
	"net/http"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// Handler holds all HTTP handlers and their dependencies
type Handler struct {
	DB             *gorm.DB
	Config         *config.Config
	Auth           *middleware.AuthMiddleware
	ESLManager     *esl.Manager
	ConsoleManager *ConsoleManager
	LogManager     *logging.LogManager
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

// SetLogManager sets the LogManager reference
func (h *Handler) SetLogManager(lm *logging.LogManager) {
	h.LogManager = lm
}

// =====================
// Health & Status
// =====================

// Health returns the health status of the API
func (h *Handler) Health(ctx iris.Context) {
	ctx.JSON(iris.Map{
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
}

// Login authenticates a user and returns a JWT token
func (h *Handler) Login(ctx iris.Context) {
	var req LoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	var user models.User
	if err := h.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(req.Password) {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// Generate token
	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(iris.Map{
		"token": token,
		"user": iris.Map{
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
func (h *Handler) AdminLogin(ctx iris.Context) {
	var req LoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	var user models.User
	if err := h.DB.Where("(username = ? OR email = ?) AND role IN ?", req.Username, req.Username, []models.UserRole{models.RoleSystemAdmin, models.RoleTenantAdmin}).First(&user).Error; err != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials or insufficient permissions"})
		return
	}

	if !user.CheckPassword(req.Password) {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid credentials"})
		return
	}

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// Generate token
	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(iris.Map{
		"token": token,
		"user": iris.Map{
			"id":        user.ID,
			"uuid":      user.UUID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"tenant_id": user.TenantID,
		},
	})
}

// Register creates a new user account
func (h *Handler) Register(ctx iris.Context) {
	// NOTE: Implement based on your registration requirements
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Registration not implemented"})
}

// RequestPasswordReset initiates a password reset
func (h *Handler) RequestPasswordReset(ctx iris.Context) {
	// NOTE: Implement password reset logic
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Password reset not implemented"})
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var user models.User
	if err := h.DB.Preload("Tenant").First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	ctx.JSON(iris.Map{
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
func (h *Handler) ChangePassword(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	if !user.CheckPassword(req.CurrentPassword) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Current password is incorrect"})
		return
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to set password"})
		return
	}

	if err := h.DB.Save(&user).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save password"})
		return
	}

	ctx.JSON(iris.Map{"message": "Password updated successfully"})
}

// Logout invalidates the current token
func (h *Handler) Logout(ctx iris.Context) {
	// NOTE: For JWT, logout is typically handled client-side
	// Optionally implement a token blacklist here
	ctx.JSON(iris.Map{"message": "Logged out successfully"})
}

// RefreshToken generates a new token for the authenticated user
func (h *Handler) RefreshToken(ctx iris.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Not authenticated"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	token, err := h.Auth.GenerateToken(&user)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(iris.Map{"token": token})
}
