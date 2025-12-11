package freeswitch

import (
	"callsign/models"
	"callsign/services/xmlcache"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// handleConfiguration processes configuration section requests (sofia.conf, ivr.conf, etc.)
func (h *FSHandler) handleConfiguration(req *XMLCurlRequest, hostname string) string {
	configName := req.KeyValue

	// Check cache first
	cacheKey := xmlcache.ConfigKey(hostname, configName)
	if cached, ok := h.Cache.Get(cacheKey); ok {
		log.Debugf("Configuration cache hit for %s", configName)
		return cached
	}

	var xml string

	// Dispatch to specific configuration handler
	switch configName {
	case "sofia.conf":
		xml = h.buildSofiaConfig(hostname)
	case "acl.conf":
		xml = h.buildACLConfig()
	case "ivr.conf":
		xml = h.buildIVRConfig(req)
	case "conference.conf":
		xml = h.buildConferenceConfig()
	case "local_stream.conf":
		xml = h.buildLocalStreamConfig()
	default:
		// Unknown config, let FreeSWITCH fall back to static file
		log.Debugf("Unknown configuration requested: %s", configName)
		return ""
	}

	if xml != "" {
		h.Cache.Set(cacheKey, xml, CacheTTL.Configuration)
	}

	return xml
}

// buildSofiaConfig generates sofia.conf XML with SIP profiles and gateways
func (h *FSHandler) buildSofiaConfig(hostname string) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="configuration">`)
	b.WriteString("\n")
	b.WriteString(`    <configuration name="sofia.conf" description="Sofia SIP Endpoint">`)
	b.WriteString("\n")

	// Global settings
	b.WriteString(`      <global_settings>`)
	b.WriteString("\n")

	var globalSettings []models.SofiaGlobalSetting
	h.DB.Where("enabled = ?", true).Find(&globalSettings)

	for _, s := range globalSettings {
		b.WriteString(fmt.Sprintf(`        <param name="%s" value="%s"/>`, xmlEscape(s.SettingName), xmlEscape(s.SettingValue)))
		b.WriteString("\n")
	}

	// Default global settings if none in DB
	if len(globalSettings) == 0 {
		b.WriteString(`        <param name="log-level" value="0"/>`)
		b.WriteString("\n")
		b.WriteString(`        <param name="debug-presence" value="0"/>`)
		b.WriteString("\n")
	}

	b.WriteString(`      </global_settings>`)
	b.WriteString("\n")

	// Profiles
	b.WriteString(`      <profiles>`)
	b.WriteString("\n")

	var profiles []models.SIPProfile
	h.DB.Where("enabled = ? AND (hostname = ? OR hostname = '' OR hostname IS NULL)", true, hostname).
		Preload("Settings", "enabled = ?", true).
		Preload("Domains").
		Find(&profiles)

	for _, profile := range profiles {
		b.WriteString(fmt.Sprintf(`        <profile name="%s">`, xmlEscape(profile.ProfileName)))
		b.WriteString("\n")

		// Gateways for this profile
		b.WriteString(`          <gateways>`)
		b.WriteString("\n")

		var gateways []models.Gateway
		h.DB.Where("profile_name = ? AND enabled = ?", profile.ProfileName, true).Find(&gateways)

		for _, gw := range gateways {
			b.WriteString(fmt.Sprintf(`            <gateway name="%s">`, gw.UUID.String()))
			b.WriteString("\n")

			// Gateway params
			if gw.Username != "" {
				b.WriteString(fmt.Sprintf(`              <param name="username" value="%s"/>`, xmlEscape(gw.Username)))
				b.WriteString("\n")
			}
			if gw.Password != "" {
				b.WriteString(fmt.Sprintf(`              <param name="password" value="%s"/>`, xmlEscape(gw.Password)))
				b.WriteString("\n")
			}
			if gw.AuthUsername != "" {
				b.WriteString(fmt.Sprintf(`              <param name="auth-username" value="%s"/>`, xmlEscape(gw.AuthUsername)))
				b.WriteString("\n")
			}
			if gw.Realm != "" {
				b.WriteString(fmt.Sprintf(`              <param name="realm" value="%s"/>`, xmlEscape(gw.Realm)))
				b.WriteString("\n")
			}
			if gw.Proxy != "" {
				b.WriteString(fmt.Sprintf(`              <param name="proxy" value="%s"/>`, xmlEscape(gw.Proxy)))
				b.WriteString("\n")
			}
			if gw.RegisterProxy != "" {
				b.WriteString(fmt.Sprintf(`              <param name="register-proxy" value="%s"/>`, xmlEscape(gw.RegisterProxy)))
				b.WriteString("\n")
			}
			if gw.FromUser != "" {
				b.WriteString(fmt.Sprintf(`              <param name="from-user" value="%s"/>`, xmlEscape(gw.FromUser)))
				b.WriteString("\n")
			}
			if gw.FromDomain != "" {
				b.WriteString(fmt.Sprintf(`              <param name="from-domain" value="%s"/>`, xmlEscape(gw.FromDomain)))
				b.WriteString("\n")
			}
			if gw.Extension != "" {
				b.WriteString(fmt.Sprintf(`              <param name="extension" value="%s"/>`, xmlEscape(gw.Extension)))
				b.WriteString("\n")
			}
			if gw.Transport != "" {
				b.WriteString(fmt.Sprintf(`              <param name="register-transport" value="%s"/>`, xmlEscape(gw.Transport)))
				b.WriteString("\n")
			}
			if gw.Register {
				b.WriteString(`              <param name="register" value="true"/>`)
			} else {
				b.WriteString(`              <param name="register" value="false"/>`)
			}
			b.WriteString("\n")
			if gw.ExpireSeconds > 0 {
				b.WriteString(fmt.Sprintf(`              <param name="expire-seconds" value="%d"/>`, gw.ExpireSeconds))
				b.WriteString("\n")
			}
			if gw.RetrySeconds > 0 {
				b.WriteString(fmt.Sprintf(`              <param name="retry-seconds" value="%d"/>`, gw.RetrySeconds))
				b.WriteString("\n")
			}
			if gw.Ping != "" {
				b.WriteString(fmt.Sprintf(`              <param name="ping" value="%s"/>`, xmlEscape(gw.Ping)))
				b.WriteString("\n")
			}
			if gw.Context != "" {
				b.WriteString(fmt.Sprintf(`              <param name="context" value="%s"/>`, xmlEscape(gw.Context)))
				b.WriteString("\n")
			}

			b.WriteString(`            </gateway>`)
			b.WriteString("\n")
		}

		b.WriteString(`          </gateways>`)
		b.WriteString("\n")

		// Domains
		b.WriteString(`          <domains>`)
		b.WriteString("\n")

		for _, domain := range profile.Domains {
			aliasStr := "false"
			if domain.Alias {
				aliasStr = "true"
			}
			parseStr := "false"
			if domain.Parse {
				parseStr = "true"
			}
			b.WriteString(fmt.Sprintf(`            <domain name="%s" alias="%s" parse="%s"/>`,
				xmlEscape(domain.DomainName), aliasStr, parseStr))
			b.WriteString("\n")
		}

		// Default domain if none defined
		if len(profile.Domains) == 0 {
			b.WriteString(`            <domain name="all" alias="true" parse="true"/>`)
			b.WriteString("\n")
		}

		b.WriteString(`          </domains>`)
		b.WriteString("\n")

		// Settings
		b.WriteString(`          <settings>`)
		b.WriteString("\n")

		for _, setting := range profile.Settings {
			b.WriteString(fmt.Sprintf(`            <param name="%s" value="%s"/>`,
				xmlEscape(setting.SettingName), xmlEscape(setting.SettingValue)))
			b.WriteString("\n")
		}

		b.WriteString(`          </settings>`)
		b.WriteString("\n")
		b.WriteString(`        </profile>`)
		b.WriteString("\n")
	}

	b.WriteString(`      </profiles>`)
	b.WriteString("\n")
	b.WriteString(`    </configuration>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}

// buildACLConfig generates acl.conf XML for access control lists
func (h *FSHandler) buildACLConfig() string {
	// For now, return a basic ACL config
	// TODO: Build from database
	return `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<document type="freeswitch/xml">
  <section name="configuration">
    <configuration name="acl.conf" description="Network Lists">
      <network-lists>
        <list name="lan" default="allow">
          <node type="allow" cidr="192.168.0.0/16"/>
          <node type="allow" cidr="10.0.0.0/8"/>
          <node type="allow" cidr="172.16.0.0/12"/>
        </list>
        <list name="loopback.auto" default="allow"/>
      </network-lists>
    </configuration>
  </section>
</document>`
}

// buildIVRConfig generates ivr.conf XML for IVR menus
func (h *FSHandler) buildIVRConfig(req *XMLCurlRequest) string {
	// IVR menu requests include Menu-Name header
	// For now, return empty to fall back to static config
	// TODO: Build from database when IVR models are created
	return ""
}

// buildConferenceConfig generates conference.conf XML
func (h *FSHandler) buildConferenceConfig() string {
	// TODO: Build from database
	return ""
}

// buildLocalStreamConfig generates local_stream.conf XML for music on hold
func (h *FSHandler) buildLocalStreamConfig() string {
	// TODO: Build from database
	return ""
}
