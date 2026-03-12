package utils

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse represents a standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Success sends a successful response with data
func Success(c *fiber.Ctx, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage sends a successful response with a message
func SuccessWithMessage(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func Error(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}

// ValidationError sends a 400 Bad Request with the validation error
func ValidationError(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusBadRequest, message)
}

// NotFoundError sends a 404 Not Found error
func NotFoundError(c *fiber.Ctx, resource string) error {
	return Error(c, fiber.StatusNotFound, resource+" not found")
}

// UnauthorizedError sends a 401 Unauthorized error
func UnauthorizedError(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Unauthorized"
	}
	return Error(c, fiber.StatusUnauthorized, message)
}

// ForbiddenError sends a 403 Forbidden error
func ForbiddenError(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Forbidden"
	}
	return Error(c, fiber.StatusForbidden, message)
}

// InternalServerError sends a 500 Internal Server Error
func InternalServerError(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Internal server error"
	}
	return Error(c, fiber.StatusInternalServerError, message)
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// Paginated sends a paginated response
func Paginated(c *fiber.Ctx, data interface{}, page, limit int, total int64) error {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return c.JSON(PaginatedResponse{
		Success:    true,
		Data:       data,
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	})
}
