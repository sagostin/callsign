package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GetWebRTCConfig returns the WebRTC/SIP configuration needed by the browser softphone.
// This is called by the frontend sipService.js to discover the WSS URL, SIP domain,
// and ICE servers for SIP.js UserAgent initialization.
func (h *Handler) GetWebRTCConfig(c *fiber.Ctx) error {
	wssURL := h.Config.SIPWssURL
	sipDomain := h.Config.SIPDomain

	// Build STUN servers list
	stunServers := []string{}
	if h.Config.STUNServers != "" {
		for _, s := range strings.Split(h.Config.STUNServers, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				stunServers = append(stunServers, s)
			}
		}
	}

	// Build TURN servers list (format: url,username,credential per entry; entries separated by ;)
	type turnServer struct {
		URLs       string `json:"urls"`
		Username   string `json:"username,omitempty"`
		Credential string `json:"credential,omitempty"`
	}
	turnServers := []turnServer{}
	if h.Config.TURNServers != "" {
		for _, entry := range strings.Split(h.Config.TURNServers, ";") {
			parts := strings.SplitN(strings.TrimSpace(entry), ",", 3)
			if len(parts) >= 1 && parts[0] != "" {
				ts := turnServer{URLs: parts[0]}
				if len(parts) >= 2 {
					ts.Username = parts[1]
				}
				if len(parts) >= 3 {
					ts.Credential = parts[2]
				}
				turnServers = append(turnServers, ts)
			}
		}
	}

	return c.JSON(fiber.Map{
		"wss_url":      wssURL,
		"sip_domain":   sipDomain,
		"stun_servers": stunServers,
		"turn_servers": turnServers,
		"enabled":      wssURL != "" && sipDomain != "",
	})
}
