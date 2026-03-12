package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

// RequestLogger logs all incoming requests
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Get the real client IP
		clientIP := c.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = c.Get("X-Real-IP")
		}
		if clientIP == "" {
			clientIP = c.IP()
		}

		// Store client IP in context for other handlers
		c.Locals("client_ip", clientIP)

		// Process request
		err := c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get status code
		statusCode := c.Response().StatusCode()

		// Log the request
		fields := log.Fields{
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     statusCode,
			"duration":   duration.String(),
			"client_ip":  clientIP,
			"user_agent": c.Get("User-Agent"),
		}

		// Add user info if authenticated
		if userID, ok := c.Locals("user_id").(uint); ok && userID > 0 {
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

		return err
	}
}

// Recovery handles panics and returns a graceful error response
func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(log.Fields{
					"panic":  r,
					"path":   c.Path(),
					"method": c.Method(),
				}).Error("Panic recovered")

				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()

		return c.Next()
	}
}
