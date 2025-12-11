package utils

import (
	"github.com/kataras/iris/v12"
)

// APIResponse represents a standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Success sends a successful response with data
func Success(ctx iris.Context, data interface{}) {
	ctx.JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage sends a successful response with a message
func SuccessWithMessage(ctx iris.Context, message string, data interface{}) {
	ctx.JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func Error(ctx iris.Context, statusCode int, message string) {
	ctx.StatusCode(statusCode)
	ctx.JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}

// ValidationError sends a 400 Bad Request with the validation error
func ValidationError(ctx iris.Context, message string) {
	Error(ctx, iris.StatusBadRequest, message)
}

// NotFoundError sends a 404 Not Found error
func NotFoundError(ctx iris.Context, resource string) {
	Error(ctx, iris.StatusNotFound, resource+" not found")
}

// UnauthorizedError sends a 401 Unauthorized error
func UnauthorizedError(ctx iris.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	Error(ctx, iris.StatusUnauthorized, message)
}

// ForbiddenError sends a 403 Forbidden error
func ForbiddenError(ctx iris.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	Error(ctx, iris.StatusForbidden, message)
}

// InternalServerError sends a 500 Internal Server Error
func InternalServerError(ctx iris.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(ctx, iris.StatusInternalServerError, message)
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
func Paginated(ctx iris.Context, data interface{}, page, limit int, total int64) {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	ctx.JSON(PaginatedResponse{
		Success:    true,
		Data:       data,
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	})
}
