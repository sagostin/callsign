package freeswitch

import (
	"callsign/models"
	"callsign/services/xmlcache"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// handleDirectory processes directory section requests (SIP auth, user lookup)
func (h *FSHandler) handleDirectory(req *XMLCurlRequest) string {
	// Handle different purposes
	switch req.Purpose {
	case "gateways":
		return h.handleDirectoryGateways(req)
	case "network-list":
		return h.handleDirectoryNetworkList(req)
	default:
		// Default: user authentication or lookup
		return h.handleDirectoryUser(req)
	}
}

// handleDirectoryUser handles SIP registration/authentication requests
func (h *FSHandler) handleDirectoryUser(req *XMLCurlRequest) string {
	if req.User == "" || req.Domain == "" {
		return ""
	}

	// Check cache first
	cacheKey := xmlcache.DirectoryKey(req.Domain, req.User)
	if cached, ok := h.Cache.Get(cacheKey); ok {
		log.Debugf("Directory cache hit for %s@%s", req.User, req.Domain)
		return cached
	}

	// Query extension from database
	var ext models.Extension
	result := h.DB.Where(
		"(extension = ? OR number_alias = ?) AND domain = ? AND enabled = ?",
		req.User, req.User, req.Domain, true,
	).First(&ext)

	if result.Error != nil {
		log.Debugf("Extension not found: %s@%s", req.User, req.Domain)
		return ""
	}

	// Build XML response
	xml := h.buildDirectoryXML(&ext, req)

	// Cache the result
	h.Cache.Set(cacheKey, xml, CacheTTL.Directory)

	return xml
}

// buildDirectoryXML generates the directory XML for an extension
func (h *FSHandler) buildDirectoryXML(ext *models.Extension, req *XMLCurlRequest) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="directory">`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`    <domain name="%s">`, xmlEscape(req.Domain)))
	b.WriteString("\n")

	// Domain params
	b.WriteString(`      <params>`)
	b.WriteString("\n")
	dialString := fmt.Sprintf("{presence_id=${dialed_user}@${dialed_domain}}${sofia_contact(${dialed_user}@${dialed_domain})}")
	b.WriteString(fmt.Sprintf(`        <param name="dial-string" value="%s"/>`, xmlEscape(dialString)))
	b.WriteString("\n")
	b.WriteString(`      </params>`)
	b.WriteString("\n")

	// User groups
	b.WriteString(`      <groups>`)
	b.WriteString("\n")
	b.WriteString(`        <group name="default">`)
	b.WriteString("\n")
	b.WriteString(`          <users>`)
	b.WriteString("\n")

	// User entry with cacheable attribute for FreeSWITCH caching
	b.WriteString(fmt.Sprintf(`            <user id="%s" cacheable="60000">`, xmlEscape(ext.Extension)))
	b.WriteString("\n")

	// User params
	b.WriteString(`              <params>`)
	b.WriteString("\n")

	// Use a1-hash for secure password transmission
	a1Hash := generateA1Hash(ext.Extension, req.Domain, ext.Password)
	b.WriteString(fmt.Sprintf(`                <param name="a1-hash" value="%s"/>`, a1Hash))
	b.WriteString("\n")

	// Voicemail settings
	if ext.VoicemailEnabled {
		b.WriteString(`                <param name="vm-enabled" value="true"/>`)
		b.WriteString("\n")
	}

	b.WriteString(`              </params>`)
	b.WriteString("\n")

	// User variables
	b.WriteString(`              <variables>`)
	b.WriteString("\n")

	// Extension/user info
	b.WriteString(fmt.Sprintf(`                <variable name="extension_uuid" value="%s"/>`, ext.UUID.String()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`                <variable name="extension" value="%s"/>`, xmlEscape(ext.Extension)))
	b.WriteString("\n")

	// Context
	userContext := ext.UserContext
	if userContext == "" {
		userContext = req.Domain
	}
	b.WriteString(fmt.Sprintf(`                <variable name="user_context" value="%s"/>`, xmlEscape(userContext)))
	b.WriteString("\n")

	// Caller ID
	if ext.EffectiveCallerIDName != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="effective_caller_id_name" value="%s"/>`, xmlEscape(ext.EffectiveCallerIDName)))
		b.WriteString("\n")
	}
	if ext.EffectiveCallerIDNumber != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="effective_caller_id_number" value="%s"/>`, xmlEscape(ext.EffectiveCallerIDNumber)))
		b.WriteString("\n")
	}
	if ext.OutboundCallerIDName != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="outbound_caller_id_name" value="%s"/>`, xmlEscape(ext.OutboundCallerIDName)))
		b.WriteString("\n")
	}
	if ext.OutboundCallerIDNumber != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="outbound_caller_id_number" value="%s"/>`, xmlEscape(ext.OutboundCallerIDNumber)))
		b.WriteString("\n")
	}

	// Call settings
	b.WriteString(fmt.Sprintf(`                <variable name="call_timeout" value="%d"/>`, ext.CallTimeout))
	b.WriteString("\n")

	if ext.TollAllow != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="toll_allow" value="%s"/>`, xmlEscape(ext.TollAllow)))
		b.WriteString("\n")
	}

	// Call forwarding
	if ext.ForwardAllEnabled {
		b.WriteString(`                <variable name="forward_all_enabled" value="true"/>`)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`                <variable name="forward_all_destination" value="%s"/>`, xmlEscape(ext.ForwardAllDestination)))
		b.WriteString("\n")
	}
	if ext.ForwardBusyEnabled {
		b.WriteString(`                <variable name="forward_busy_enabled" value="true"/>`)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`                <variable name="forward_busy_destination" value="%s"/>`, xmlEscape(ext.ForwardBusyDestination)))
		b.WriteString("\n")
	}
	if ext.ForwardNoAnswerEnabled {
		b.WriteString(`                <variable name="forward_no_answer_enabled" value="true"/>`)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`                <variable name="forward_no_answer_destination" value="%s"/>`, xmlEscape(ext.ForwardNoAnswerDestination)))
		b.WriteString("\n")
	}

	// DND
	if ext.DoNotDisturb {
		b.WriteString(`                <variable name="do_not_disturb" value="true"/>`)
		b.WriteString("\n")
	}

	// Recording
	if ext.RecordInbound {
		b.WriteString(`                <variable name="record_inbound" value="true"/>`)
		b.WriteString("\n")
	}
	if ext.RecordOutbound {
		b.WriteString(`                <variable name="record_outbound" value="true"/>`)
		b.WriteString("\n")
	}

	// Account code
	if ext.AccountCode != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="accountcode" value="%s"/>`, xmlEscape(ext.AccountCode)))
		b.WriteString("\n")
	}

	// Limit (max concurrent calls)
	b.WriteString(fmt.Sprintf(`                <variable name="limit_max" value="%d"/>`, ext.LimitMax))
	b.WriteString("\n")

	b.WriteString(`              </variables>`)
	b.WriteString("\n")
	b.WriteString(`            </user>`)
	b.WriteString("\n")
	b.WriteString(`          </users>`)
	b.WriteString("\n")
	b.WriteString(`        </group>`)
	b.WriteString("\n")
	b.WriteString(`      </groups>`)
	b.WriteString("\n")
	b.WriteString(`    </domain>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}

// handleDirectoryGateways returns gateway directory info when purpose=gateways
func (h *FSHandler) handleDirectoryGateways(req *XMLCurlRequest) string {
	// For now, return not found - gateways are handled in configuration
	return ""
}

// handleDirectoryNetworkList returns network list info
func (h *FSHandler) handleDirectoryNetworkList(req *XMLCurlRequest) string {
	// For now, return not found
	return ""
}

// generateA1Hash creates MD5 hash of user:realm:password for secure SIP auth
func generateA1Hash(user, realm, password string) string {
	data := user + ":" + realm + ":" + password
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// xmlEscape escapes special XML characters
func xmlEscape(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))
	return b.String()
}
