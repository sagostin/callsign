package freeswitch

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

	"github.com/kataras/iris/v12"
)

// DebugXML allows system admins to simulate any XML CURL request
func (h *FSHandler) DebugXML(ctx iris.Context) {
	section := ctx.URLParam("section")
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
		req.Context = ctx.URLParam("context")
		if req.Context == "" {
			req.Context = "default"
		}
		req.DestinationNumber = ctx.URLParam("destination_number")
		req.Domain = ctx.URLParam("domain")

	case "directory":
		req.User = ctx.URLParam("user")
		req.Domain = ctx.URLParam("domain")
		req.Action = ctx.URLParam("action") // e.g. message-count, sip_auth

	case "configuration":
		req.KeyName = "name"
		req.KeyValue = ctx.URLParam("config_name") // e.g. sofia.conf
		// Some configs need other params, simplistic for now
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

	ctx.JSON(iris.Map{
		"xml":     xml,
		"request": req,
	})
}

// DebugDialplanTenant allows tenant admins to debug their own routing
func (h *FSHandler) DebugDialplanTenant(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	// Get tenant domain to force correct context/domain
	var tenant models.Tenant
	if err := h.DB.First(&tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Tenant not found"})
		return
	}

	// Force domain to tenant's domain
	domain := tenant.Domain
	context := ctx.URLParam("context")

	// Tenant can only debug 'public' (inbound) or their domain context (outbound/internal)
	if context != "public" && context != domain {
		// Default to domain context if invalid or empty
		context = domain
	}

	dest := ctx.URLParam("destination_number")

	req := XMLCurlRequest{
		Section:            "dialplan",
		Context:            context,
		DestinationNumber:  dest,
		Domain:             domain,
		FreeSwitchHostname: "debug-simulator",
	}

	xml := h.handleDialplan(&req)

	ctx.JSON(iris.Map{
		"xml":                xml,
		"destination_number": dest,
		"context":            context,
		"domain":             domain,
	})
}
