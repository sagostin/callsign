package middleware

import (
	"time"

	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
)

// RequestLogger logs all incoming requests
func RequestLogger() iris.Handler {
	return func(ctx iris.Context) {
		start := time.Now()

		// Get the real client IP
		clientIP := ctx.GetHeader("X-Forwarded-For")
		if clientIP == "" {
			clientIP = ctx.GetHeader("X-Real-IP")
		}
		if clientIP == "" {
			clientIP = ctx.RemoteAddr()
		}

		// Store client IP in context for other handlers
		ctx.Values().Set("client_ip", clientIP)

		// Process request
		ctx.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get status code
		statusCode := ctx.GetStatusCode()

		// Log the request
		fields := log.Fields{
			"method":     ctx.Method(),
			"path":       ctx.Path(),
			"status":     statusCode,
			"duration":   duration.String(),
			"client_ip":  clientIP,
			"user_agent": ctx.GetHeader("User-Agent"),
		}

		// Add user info if authenticated
		if userID := ctx.Values().GetUintDefault("user_id", 0); userID > 0 {
			fields["user_id"] = userID
		}

		// Log level based on status code
		if statusCode >= 500 {
			log.WithFields(fields).Error("Request completed with server error")
		} else if statusCode >= 400 {
			log.WithFields(fields).Warn("Request completed with client error")
		} else {
			log.WithFields(fields).Info("Request completed")
		}
	}
}

// Recovery handles panics and returns a graceful error response
func Recovery() iris.Handler {
	return func(ctx iris.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(log.Fields{
					"panic":  r,
					"path":   ctx.Path(),
					"method": ctx.Method(),
				}).Error("Panic recovered")

				ctx.StatusCode(iris.StatusInternalServerError)
				ctx.JSON(iris.Map{
					"error": "Internal server error",
				})
			}
		}()

		ctx.Next()
	}
}
