package middleware

import (
	"callsign/config"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing
func CORS(cfg *config.Config) fiber.Handler {
	allowedOrigins := make(map[string]bool)
	for _, origin := range cfg.CORSOrigins {
		allowedOrigins[strings.TrimSpace(origin)] = true
	}

	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		// Check if origin is allowed
		if allowedOrigins["*"] || allowedOrigins[origin] {
			c.Set("Access-Control-Allow-Origin", origin)
		}

		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Tenant-ID")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}
