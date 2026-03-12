package freeswitch

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// DebugXML allows system admins to simulate any XML CURL request
func (h *FSHandler) DebugXML(c *fiber.Ctx) error {
	section := c.Query("section")
	if section == "" {
		section = "dialplan" // Default
	}

	// Construct request based on section
	req := XMLCurlRequest{
		Section:            section,
		FreeSwitchHostname: "debug-simulator",
	}

	// Populate specific fields based on section
	switch section {
	case "dialplan":
		req.Context = c.Query("context")
		if req.Context == "" {
			req.Context = "default"
		}
		req.DestinationNumber = c.Query("destination_number")
		req.Domain = c.Query("domain")

	case "directory":
		req.User = c.Query("user")
		req.Domain = c.Query("domain")
		req.Action = c.Query("action")          // e.g. message-count, sip_auth
		req.Purpose = c.Query("purpose")        // e.g. gateways, network-list
		req.SIPProfile = c.Query("sip_profile") // e.g. internal, external

	case "configuration":
		req.KeyName = "name"
		req.KeyValue = c.Query("config_name") // e.g. sofia.conf
		// Fallback: also accept key_value param (used by ConfigInspector dynamic entries)
		if req.KeyValue == "" {
			req.KeyValue = c.Query("key_value")
		}
	}

	// Dispatch to handler logic
	// We reuse handleDialplan, handleDirectory, handleConfiguration etc via the main switch in HandleXMLCurl
	// But those are private methods on FSHandler usually, or we can just reproduce the switch here.
	// Looking at xmlcurl.go, they are methods on FSHandler.

	var xml string
	switch section {
	case "directory":
		xml = h.handleDirectory(&req)
	case "configuration":
		xml = h.handleConfiguration(&req, req.FreeSwitchHostname)
	case "dialplan":
		xml = h.handleDialplan(&req)
	case "phrases":
		xml = h.handlePhrases(&req)
	default:
		xml = "<!-- Unknown section -->"
	}

	return c.JSON(fiber.Map{
		"xml":     xml,
		"request": req,
	})
}

// DebugDialplanTenant allows tenant admins to debug their own routing
func (h *FSHandler) DebugDialplanTenant(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	// Get tenant domain to force correct context/domain
	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}

	// Force domain to tenant's domain
	domain := tenant.Domain
	context := c.Query("context")

	// Tenant can only debug 'public' (inbound) or their domain context (outbound/internal)
	if context != "public" && context != domain {
		// Default to domain context if invalid or empty
		context = domain
	}

	dest := c.Query("destination_number")

	req := XMLCurlRequest{
		Section:            "dialplan",
		Context:            context,
		DestinationNumber:  dest,
		Domain:             domain,
		FreeSwitchHostname: "debug-simulator",
	}

	xml := h.handleDialplan(&req)

	return c.JSON(fiber.Map{
		"xml":                xml,
		"destination_number": dest,
		"context":            context,
		"domain":             domain,
	})
}
