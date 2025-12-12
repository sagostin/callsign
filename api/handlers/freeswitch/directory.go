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

	// Check if this is a device registration (starts with configurable prefix, default "d_")
	if strings.HasPrefix(req.User, "d_") || strings.HasPrefix(req.User, "dev_") {
		xml := h.handleDeviceRegistration(req)
		if xml != "" {
			h.Cache.Set(cacheKey, xml, CacheTTL.Directory)
			return xml
		}
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

// handleDeviceRegistration handles device SIP registration (with prefix like d_)
func (h *FSHandler) handleDeviceRegistration(req *XMLCurlRequest) string {
	// Look up device by registration user
	var device models.Device
	if err := h.DB.Where("registration_user = ? AND enabled = ?", req.User, true).
		Preload("Tenant").
		First(&device).Error; err != nil {
		log.Debugf("Device not found for registration: %s@%s", req.User, req.Domain)
		return ""
	}

	// Verify domain matches tenant
	var tenant models.Tenant
	h.DB.First(&tenant, device.TenantID)
	if tenant.Domain != req.Domain {
		log.Warnf("Device registration domain mismatch: %s vs %s", req.Domain, tenant.Domain)
		return ""
	}

	// Build device directory XML
	return h.buildDeviceDirectoryXML(&device, req)
}

// buildDeviceDirectoryXML generates directory XML for a device registration
func (h *FSHandler) buildDeviceDirectoryXML(device *models.Device, req *XMLCurlRequest) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="directory">`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`    <domain name="%s">`, xmlEscape(req.Domain)))
	b.WriteString("\n")

	b.WriteString(`      <groups>`)
	b.WriteString("\n")
	b.WriteString(`        <group name="devices">`)
	b.WriteString("\n")
	b.WriteString(`          <users>`)
	b.WriteString("\n")

	// Device user entry
	b.WriteString(fmt.Sprintf(`            <user id="%s">`, xmlEscape(device.RegistrationUser)))
	b.WriteString("\n")

	// Params
	b.WriteString(`              <params>`)
	b.WriteString("\n")

	// A1 hash for auth
	a1Hash := generateA1Hash(device.RegistrationUser, req.Domain, device.RegistrationPass)
	b.WriteString(fmt.Sprintf(`                <param name="a1-hash" value="%s"/>`, a1Hash))
	b.WriteString("\n")

	b.WriteString(`              </params>`)
	b.WriteString("\n")

	// Variables
	b.WriteString(`              <variables>`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`                <variable name="device_uuid" value="%s"/>`, device.UUID.String()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`                <variable name="device_mac" value="%s"/>`, xmlEscape(device.MAC)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`                <variable name="device_type" value="%s"/>`, string(device.DeviceType)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`                <variable name="user_context" value="%s"/>`, xmlEscape(req.Domain)))
	b.WriteString("\n")

	// Can't make outbound calls directly (device registration is for receiving only)
	// Outbound calls go through the assigned extension
	b.WriteString(`                <variable name="toll_allow" value=""/>`)
	b.WriteString("\n")
	b.WriteString(`                <variable name="is_device" value="true"/>`)
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

	// MWI account
	if ext.MWIAccount != "" {
		b.WriteString(fmt.Sprintf(`                <param name="MWI-Account" value="%s"/>`, xmlEscape(ext.MWIAccount)))
		b.WriteString("\n")
	}

	// Auth ACL
	if ext.AuthACL != "" {
		b.WriteString(fmt.Sprintf(`                <param name="auth-acl" value="%s"/>`, xmlEscape(ext.AuthACL)))
		b.WriteString("\n")
	}

	// Max registrations
	if ext.MaxRegistrations > 0 {
		b.WriteString(fmt.Sprintf(`                <param name="max-registrations-per-extension" value="%d"/>`, ext.MaxRegistrations))
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
	if ext.EmergencyCallerIDName != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="emergency_caller_id_name" value="%s"/>`, xmlEscape(ext.EmergencyCallerIDName)))
		b.WriteString("\n")
	}
	if ext.EmergencyCallerIDNumber != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="emergency_caller_id_number" value="%s"/>`, xmlEscape(ext.EmergencyCallerIDNumber)))
		b.WriteString("\n")
	}

	// Call settings
	b.WriteString(fmt.Sprintf(`                <variable name="call_timeout" value="%d"/>`, ext.CallTimeout))
	b.WriteString("\n")

	if ext.TollAllow != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="toll_allow" value="%s"/>`, xmlEscape(ext.TollAllow)))
		b.WriteString("\n")
	}

	// Hold music
	if ext.HoldMusic != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="hold_music" value="%s"/>`, xmlEscape(ext.HoldMusic)))
		b.WriteString("\n")
	}

	// Call groups
	if ext.CallGroup != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="call_group" value="%s"/>`, xmlEscape(ext.CallGroup)))
		b.WriteString("\n")
	}
	if ext.PickupGroup != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="pickup_group" value="%s"/>`, xmlEscape(ext.PickupGroup)))
		b.WriteString("\n")
	}

	// SIP/NAT settings
	if ext.SIPForceContact != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="sip-force-contact" value="%s"/>`, xmlEscape(ext.SIPForceContact)))
		b.WriteString("\n")
	}
	if ext.SIPForceExpires > 0 {
		b.WriteString(fmt.Sprintf(`                <variable name="sip-force-expires" value="%d"/>`, ext.SIPForceExpires))
		b.WriteString("\n")
	}

	// Media bypass
	if ext.BypassMedia != "" {
		switch ext.BypassMedia {
		case "bypass-media":
			b.WriteString(`                <variable name="bypass_media" value="true"/>`)
		case "bypass-media-after-bridge":
			b.WriteString(`                <variable name="bypass_media_after_bridge" value="true"/>`)
		case "proxy-media":
			b.WriteString(`                <variable name="proxy_media" value="true"/>`)
		}
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
	if ext.ForwardUserNotRegisteredEnabled {
		b.WriteString(`                <variable name="forward_user_not_registered_enabled" value="true"/>`)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`                <variable name="forward_user_not_registered_destination" value="%s"/>`, xmlEscape(ext.ForwardUserNotRegisteredDestination)))
		b.WriteString("\n")
	}

	// DND
	if ext.DoNotDisturb {
		b.WriteString(`                <variable name="do_not_disturb" value="true"/>`)
		b.WriteString("\n")
	}

	// Follow me
	if ext.FollowMeEnabled {
		b.WriteString(`                <variable name="follow_me_enabled" value="true"/>`)
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

	// Directory info
	directoryFullName := ""
	if ext.DirectoryFirstName != "" {
		directoryFullName = ext.DirectoryFirstName
		if ext.DirectoryLastName != "" {
			directoryFullName += " " + ext.DirectoryLastName
		}
	}
	if directoryFullName != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="directory_full_name" value="%s"/>`, xmlEscape(directoryFullName)))
		b.WriteString("\n")
	}
	if ext.DirectoryVisible {
		b.WriteString(`                <variable name="directory-visible" value="true"/>`)
		b.WriteString("\n")
	}
	if ext.DirectoryExtenVisible {
		b.WriteString(`                <variable name="directory-exten-visible" value="true"/>`)
		b.WriteString("\n")
	}

	// Limit (max concurrent calls)
	b.WriteString(fmt.Sprintf(`                <variable name="limit_max" value="%d"/>`, ext.LimitMax))
	b.WriteString("\n")
	if ext.LimitDestination != "" {
		b.WriteString(fmt.Sprintf(`                <variable name="limit_destination" value="%s"/>`, xmlEscape(ext.LimitDestination)))
		b.WriteString("\n")
	}

	// Standard export vars
	b.WriteString(`                <variable name="record_stereo" value="true"/>`)
	b.WriteString("\n")
	b.WriteString(`                <variable name="transfer_fallback_extension" value="operator"/>`)
	b.WriteString("\n")
	b.WriteString(`                <variable name="export_vars" value="domain_name,domain_uuid"/>`)
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
