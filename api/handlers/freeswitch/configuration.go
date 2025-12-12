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
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="configuration">`)
	b.WriteString("\n")
	b.WriteString(`    <configuration name="acl.conf" description="Network Lists">`)
	b.WriteString("\n")
	b.WriteString(`      <network-lists>`)
	b.WriteString("\n")

	// Fetch all ACLs from database
	var acls []models.ACL
	h.DB.Where("enabled = ?", true).Preload("Nodes", "enabled = ?", true).Find(&acls)

	for _, acl := range acls {
		defaultAction := acl.Default
		if defaultAction == "" {
			defaultAction = "deny"
		}
		b.WriteString(fmt.Sprintf(`        <list name="%s" default="%s">`, xmlEscape(acl.Name), xmlEscape(defaultAction)))
		b.WriteString("\n")

		for _, node := range acl.Nodes {
			nodeType := node.Type
			if nodeType == "" {
				nodeType = "allow"
			}
			if node.CIDR != "" {
				b.WriteString(fmt.Sprintf(`          <node type="%s" cidr="%s"/>`, xmlEscape(nodeType), xmlEscape(node.CIDR)))
				b.WriteString("\n")
			} else if node.Domain != "" {
				b.WriteString(fmt.Sprintf(`          <node type="%s" domain="%s"/>`, xmlEscape(nodeType), xmlEscape(node.Domain)))
				b.WriteString("\n")
			}
		}

		// Auto-include enabled tenant domains for 'providers' and 'domains' ACLs
		// This matches FusionPBX behavior where these ACLs dynamically include active domains
		if acl.Name == "providers" || acl.Name == "domains" {
			var tenants []models.Tenant
			h.DB.Where("enabled = ?", true).Find(&tenants)
			for _, tenant := range tenants {
				if tenant.Domain != "" {
					b.WriteString(fmt.Sprintf(`          <node type="allow" domain="%s" description="%s"/>`,
						xmlEscape(tenant.Domain), xmlEscape(tenant.Name)))
					b.WriteString("\n")
				}
			}
		}

		b.WriteString(`        </list>`)
		b.WriteString("\n")
	}

	// If no ACLs in database, provide sensible defaults
	if len(acls) == 0 {
		b.WriteString(`        <list name="lan" default="allow">`)
		b.WriteString("\n")
		b.WriteString(`          <node type="allow" cidr="192.168.0.0/16"/>`)
		b.WriteString("\n")
		b.WriteString(`          <node type="allow" cidr="10.0.0.0/8"/>`)
		b.WriteString("\n")
		b.WriteString(`          <node type="allow" cidr="172.16.0.0/12"/>`)
		b.WriteString("\n")
		b.WriteString(`        </list>`)
		b.WriteString("\n")
		b.WriteString(`        <list name="loopback.auto" default="allow"/>`)
		b.WriteString("\n")
	}

	b.WriteString(`      </network-lists>`)
	b.WriteString("\n")
	b.WriteString(`    </configuration>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}

// buildIVRConfig generates ivr.conf XML for IVR menus
func (h *FSHandler) buildIVRConfig(req *XMLCurlRequest) string {
	// Get all enabled IVR menus
	var menus []models.IVRMenu
	h.DB.Where("enabled = ?", true).Preload("Options", "enabled = ?", true).Find(&menus)

	if len(menus) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="configuration">`)
	b.WriteString("\n")
	b.WriteString(`    <configuration name="ivr.conf" description="IVR Menus">`)
	b.WriteString("\n")
	b.WriteString(`      <menus>`)
	b.WriteString("\n")

	for _, menu := range menus {
		b.WriteString(fmt.Sprintf(`        <menu name="%s"`, xmlEscape(menu.UUID.String())))
		b.WriteString(fmt.Sprintf(` greet-long="%s"`, xmlEscape(menu.GreetLong)))
		if menu.GreetShort != "" {
			b.WriteString(fmt.Sprintf(` greet-short="%s"`, xmlEscape(menu.GreetShort)))
		}
		if menu.InvalidSound != "" {
			b.WriteString(fmt.Sprintf(` invalid-sound="%s"`, xmlEscape(menu.InvalidSound)))
		}
		if menu.ExitSound != "" {
			b.WriteString(fmt.Sprintf(` exit-sound="%s"`, xmlEscape(menu.ExitSound)))
		}
		b.WriteString(fmt.Sprintf(` timeout="%d"`, menu.Timeout))
		b.WriteString(fmt.Sprintf(` max-failures="%d"`, menu.MaxFailures))
		b.WriteString(fmt.Sprintf(` max-timeouts="%d"`, menu.MaxTimeouts))
		b.WriteString(fmt.Sprintf(` digit-len="%d"`, menu.DigitLen))
		b.WriteString(fmt.Sprintf(` inter-digit-timeout="%d"`, menu.InterDigitTime))
		b.WriteString(`>`)
		b.WriteString("\n")

		// Add options as entries
		for _, opt := range menu.Options {
			action := h.ivrOptionToAction(&opt)
			b.WriteString(fmt.Sprintf(`          <entry digits="%s" action="%s" param="%s"/>`,
				xmlEscape(opt.Digits),
				xmlEscape(action.App),
				xmlEscape(action.Param)))
			b.WriteString("\n")
		}

		b.WriteString(`        </menu>`)
		b.WriteString("\n")
	}

	b.WriteString(`      </menus>`)
	b.WriteString("\n")
	b.WriteString(`    </configuration>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}

// ivrActionResult holds parsed IVR action
type ivrActionResult struct {
	App   string
	Param string
}

// ivrOptionToAction converts an IVRMenuOption to FreeSWITCH action
func (h *FSHandler) ivrOptionToAction(opt *models.IVRMenuOption) ivrActionResult {
	switch opt.Action {
	case "transfer":
		return ivrActionResult{App: "menu-exec-app", Param: "transfer " + opt.ActionParam + " XML ${domain_name}"}
	case "ivr":
		return ivrActionResult{App: "menu-sub", Param: opt.ActionParam}
	case "voicemail":
		return ivrActionResult{App: "menu-exec-app", Param: "voicemail default ${domain_name} " + opt.ActionParam}
	case "ring_group":
		return ivrActionResult{App: "menu-exec-app", Param: "transfer " + opt.ActionParam + " XML ${domain_name}"}
	case "queue":
		return ivrActionResult{App: "menu-exec-app", Param: "callcenter " + opt.ActionParam + "@${domain_name}"}
	case "playback":
		return ivrActionResult{App: "menu-exec-app", Param: "playback " + opt.ActionParam}
	case "hangup":
		return ivrActionResult{App: "menu-exit", Param: ""}
	case "repeat":
		return ivrActionResult{App: "menu-top", Param: ""}
	case "exit":
		return ivrActionResult{App: "menu-exit", Param: ""}
	default:
		return ivrActionResult{App: "menu-exec-app", Param: opt.Action + " " + opt.ActionParam}
	}
}

// buildConferenceConfig generates conference.conf XML
func (h *FSHandler) buildConferenceConfig() string {
	// For now, return a default conference config
	// TODO: Build from database when conference models are created
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="configuration">`)
	b.WriteString("\n")
	b.WriteString(`    <configuration name="conference.conf" description="Audio Conference">`)
	b.WriteString("\n")
	b.WriteString(`      <caller-controls>`)
	b.WriteString("\n")
	b.WriteString(`        <group name="default">`)
	b.WriteString("\n")
	b.WriteString(`          <control action="mute" digits="0"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="deaf mute" digits="*"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="energy up" digits="9"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="energy equ" digits="8"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="energy dn" digits="7"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="vol talk up" digits="3"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="vol talk zero" digits="2"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="vol talk dn" digits="1"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="vol listen up" digits="6"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="vol listen zero" digits="5"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="vol listen dn" digits="4"/>`)
	b.WriteString("\n")
	b.WriteString(`          <control action="hangup" digits="#"/>`)
	b.WriteString("\n")
	b.WriteString(`        </group>`)
	b.WriteString("\n")
	b.WriteString(`      </caller-controls>`)
	b.WriteString("\n")
	b.WriteString(`      <profiles>`)
	b.WriteString("\n")
	b.WriteString(`        <profile name="default">`)
	b.WriteString("\n")
	b.WriteString(`          <param name="rate" value="16000"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="interval" value="20"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="energy-level" value="100"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="moh-sound" value="local_stream://default"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="caller-controls" value="default"/>`)
	b.WriteString("\n")
	b.WriteString(`        </profile>`)
	b.WriteString("\n")
	b.WriteString(`      </profiles>`)
	b.WriteString("\n")
	b.WriteString(`    </configuration>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}

// buildLocalStreamConfig generates local_stream.conf XML for music on hold
func (h *FSHandler) buildLocalStreamConfig() string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="configuration">`)
	b.WriteString("\n")
	b.WriteString(`    <configuration name="local_stream.conf" description="Local Streams">`)
	b.WriteString("\n")

	// Default stream pointing to standard FreeSWITCH music location
	b.WriteString(`      <directory name="default" path="$${sounds_dir}/music/default">`)
	b.WriteString("\n")
	b.WriteString(`        <param name="rate" value="8000"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="shuffle" value="true"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="channels" value="1"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="interval" value="20"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="timer-name" value="soft"/>`)
	b.WriteString("\n")
	b.WriteString(`      </directory>`)
	b.WriteString("\n")

	// Additional rate directories
	rates := []string{"8000", "16000", "32000", "48000"}
	for _, rate := range rates {
		b.WriteString(fmt.Sprintf(`      <directory name="moh/%s" path="$${sounds_dir}/music/%s">`, rate, rate))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <param name="rate" value="%s"/>`, rate))
		b.WriteString("\n")
		b.WriteString(`        <param name="shuffle" value="true"/>`)
		b.WriteString("\n")
		b.WriteString(`        <param name="channels" value="1"/>`)
		b.WriteString("\n")
		b.WriteString(`        <param name="interval" value="20"/>`)
		b.WriteString("\n")
		b.WriteString(`        <param name="timer-name" value="soft"/>`)
		b.WriteString("\n")
		b.WriteString(`      </directory>`)
		b.WriteString("\n")
	}

	// TODO: Add tenant-specific streams from database

	b.WriteString(`    </configuration>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}
