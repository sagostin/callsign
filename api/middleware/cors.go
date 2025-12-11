package middleware

import (
	"callsign/config"
	"strings"

	"github.com/kataras/iris/v12"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing
func CORS(cfg *config.Config) iris.Handler {
	allowedOrigins := make(map[string]bool)
	for _, origin := range cfg.CORSOrigins {
		allowedOrigins[strings.TrimSpace(origin)] = true
	}

	return func(ctx iris.Context) {
		origin := ctx.GetHeader("Origin")

		// Check if origin is allowed
		if allowedOrigins["*"] || allowedOrigins[origin] {
			ctx.Header("Access-Control-Allow-Origin", origin)
		}

		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Tenant-ID")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		ctx.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if ctx.Method() == iris.MethodOptions {
			ctx.StatusCode(iris.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
