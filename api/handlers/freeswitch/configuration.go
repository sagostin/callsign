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

	log.WithFields(log.Fields{
		"config":   configName,
		"hostname": hostname,
		"tag_name": req.TagName,
		"key_name": req.KeyName,
	}).Debug("Configuration request received")

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
	case "callcenter.conf":
		xml = h.buildCallcenterConfig()
	case "voicemail.conf":
		xml = h.buildVoicemailConfig()
	default:
		// Unknown config, let FreeSWITCH fall back to static file
		log.WithField("config", configName).Debug("Configuration not handled, falling back to static file")
		return ""
	}

	if xml != "" {
		log.WithField("config", configName).Debug("Configuration generated from database")
		h.Cache.Set(cacheKey, xml, CacheTTL.Configuration)
	}

	return xml
}

// buildSofiaConfig - sofia.conf is handled via static files
// SIP profiles are stored as individual XML files in sip_profiles/ directory
// and included via: <X-PRE-PROCESS cmd="include" data="sip_profiles/*.xml"/>
// Gateways are served dynamically via directory (purpose=gateways)
// Return empty to let FreeSWITCH use the static sofia.conf.xml
func (h *FSHandler) buildSofiaConfig(hostname string) string {
	log.Debug("sofia.conf requested - using static file (profiles on disk, gateways via directory)")
	return ""
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

// buildConferenceConfig generates conference.conf XML from database
func (h *FSHandler) buildConferenceConfig() string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="configuration">`)
	b.WriteString("\n")
	b.WriteString(`    <configuration name="conference.conf" description="Audio Conference">`)
	b.WriteString("\n")

	// Caller controls — always include defaults
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
	b.WriteString(`        <group name="moderator">`)
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

	// Always include the default profile
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
	b.WriteString(`          <param name="moderator-controls" value="moderator"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="comfort-noise" value="true"/>`)
	b.WriteString("\n")
	b.WriteString(`        </profile>`)
	b.WriteString("\n")

	// Generate profiles from DB conferences that have custom profile names
	var conferences []models.Conference
	h.DB.Where("enabled = ? AND profile_name != 'default'", true).Find(&conferences)

	// Track generated profile names to avoid duplicates
	generated := map[string]bool{"default": true}

	for _, conf := range conferences {
		profileName := conf.ProfileName
		if profileName == "" || generated[profileName] {
			continue
		}
		generated[profileName] = true

		b.WriteString(fmt.Sprintf(`        <profile name="%s">`, xmlEscape(profileName)))
		b.WriteString("\n")
		b.WriteString(`          <param name="rate" value="16000"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="interval" value="20"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="energy-level" value="100"/>`)
		b.WriteString("\n")

		// Music on hold
		moh := "local_stream://default"
		if conf.MusicOnHold != "" {
			moh = conf.MusicOnHold
		}
		b.WriteString(fmt.Sprintf(`          <param name="moh-sound" value="%s"/>`, xmlEscape(moh)))
		b.WriteString("\n")

		b.WriteString(`          <param name="caller-controls" value="default"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="moderator-controls" value="moderator"/>`)
		b.WriteString("\n")

		if conf.MaxMembers > 0 {
			b.WriteString(fmt.Sprintf(`          <param name="max-members" value="%d"/>`, conf.MaxMembers))
			b.WriteString("\n")
		}

		if conf.MuteOnJoin {
			b.WriteString(`          <param name="mute" value="true"/>`)
			b.WriteString("\n")
		}

		if conf.AnnounceCount {
			b.WriteString(`          <param name="announce-count" value="10"/>`)
			b.WriteString("\n")
		}

		if conf.PIN != "" {
			b.WriteString(fmt.Sprintf(`          <param name="pin" value="%s"/>`, xmlEscape(conf.PIN)))
			b.WriteString("\n")
		}

		if conf.ModeratorPIN != "" {
			b.WriteString(fmt.Sprintf(`          <param name="moderator-pin" value="%s"/>`, xmlEscape(conf.ModeratorPIN)))
			b.WriteString("\n")
		}

		if conf.RecordConference {
			b.WriteString(`          <param name="auto-record" value="/var/lib/callsign/recordings/conference/${conference_name}_${strftime(%Y%m%d-%H%M%S)}.wav"/>`)
			b.WriteString("\n")
		}

		b.WriteString(`          <param name="comfort-noise" value="true"/>`)
		b.WriteString("\n")

		if conf.EnterSound != "" {
			b.WriteString(fmt.Sprintf(`          <param name="enter-sound" value="%s"/>`, xmlEscape(conf.EnterSound)))
			b.WriteString("\n")
		}

		if conf.ExitSound != "" {
			b.WriteString(fmt.Sprintf(`          <param name="exit-sound" value="%s"/>`, xmlEscape(conf.ExitSound)))
			b.WriteString("\n")
		}

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

	// Tenant-specific music streams from database
	var mediaFiles []models.MediaFile
	h.DB.Where("type = ?", models.MediaTypeMusic).Find(&mediaFiles)

	// Group by tenant for tenant-specific streams
	tenantStreams := make(map[uint][]models.MediaFile)
	for _, mf := range mediaFiles {
		tenantStreams[mf.TenantID] = append(tenantStreams[mf.TenantID], mf)
	}

	for tenantID := range tenantStreams {
		streamName := fmt.Sprintf("tenant_%d", tenantID)
		streamPath := fmt.Sprintf("/var/lib/callsign/media/%d/music", tenantID)
		b.WriteString(fmt.Sprintf(`      <directory name="%s" path="%s">`, streamName, streamPath))
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
	}

	b.WriteString(`    </configuration>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}

// buildCallcenterConfig generates callcenter.conf XML for mod_callcenter
// This is critical — queues, agents, and tiers are all served dynamically from the DB
func (h *FSHandler) buildCallcenterConfig() string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="configuration">`)
	b.WriteString("\n")
	b.WriteString(`    <configuration name="callcenter.conf" description="Call Center">`)
	b.WriteString("\n")

	// Global settings
	b.WriteString(`      <settings>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="truncate-tiers-on-load" value="true"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="truncate-agents-on-load" value="true"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="odbc-dsn" value=""/>`)
	b.WriteString("\n")
	b.WriteString(`      </settings>`)
	b.WriteString("\n")

	// Queues
	b.WriteString(`      <queues>`)
	b.WriteString("\n")

	var queues []models.Queue
	h.DB.Where("enabled = ?", true).Preload("Agents").Find(&queues)

	// Also get tenant domains for building queue names
	tenantDomains := make(map[uint]string)
	var tenants []models.Tenant
	h.DB.Where("enabled = ?", true).Find(&tenants)
	for _, t := range tenants {
		tenantDomains[t.ID] = t.Domain
	}

	for _, q := range queues {
		// Queue name format: queue_name@domain (matches FusionPBX convention)
		queueName := q.Name
		if domain, ok := tenantDomains[q.TenantID]; ok && domain != "" {
			queueName = fmt.Sprintf("%s@%s", q.Name, domain)
		}

		b.WriteString(fmt.Sprintf(`        <queue name="%s">`, xmlEscape(queueName)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <param name="strategy" value="%s"/>`, xmlEscape(string(q.Strategy))))
		b.WriteString("\n")

		// Music on hold
		moh := "local_stream://default"
		if q.MohSound != "" {
			moh = q.MohSound
		}
		b.WriteString(fmt.Sprintf(`          <param name="moh-sound" value="%s"/>`, xmlEscape(moh)))
		b.WriteString("\n")

		// Time-based score
		tbs := "queue"
		if q.TimeBasedScore != "" {
			tbs = q.TimeBasedScore
		}
		b.WriteString(fmt.Sprintf(`          <param name="time-base-score" value="%s"/>`, xmlEscape(tbs)))
		b.WriteString("\n")

		// Tier settings
		if q.TierRulesApply {
			b.WriteString(`          <param name="tier-rules-apply" value="true"/>`)
		} else {
			b.WriteString(`          <param name="tier-rules-apply" value="false"/>`)
		}
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <param name="tier-rule-wait-second" value="%d"/>`, q.TierRuleWaitSec))
		b.WriteString("\n")
		b.WriteString(`          <param name="tier-rule-wait-multiply-level" value="true"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="tier-rule-no-agent-no-wait" value="false"/>`)
		b.WriteString("\n")

		// Timeouts
		if q.MaxWaitTime > 0 {
			b.WriteString(fmt.Sprintf(`          <param name="max-wait-time" value="%d"/>`, q.MaxWaitTime))
			b.WriteString("\n")
		}
		if q.MaxWaitTimeNoAgent > 0 {
			b.WriteString(fmt.Sprintf(`          <param name="max-wait-time-with-no-agent" value="%d"/>`, q.MaxWaitTimeNoAgent))
			b.WriteString("\n")
		}
		if q.MaxWaitTimeNoAgentTimeReached > 0 {
			b.WriteString(fmt.Sprintf(`          <param name="max-wait-time-with-no-agent-time-reached" value="%d"/>`, q.MaxWaitTimeNoAgentTimeReached))
			b.WriteString("\n")
		}

		// Agent delay settings
		b.WriteString(fmt.Sprintf(`          <param name="discard-abandoned-after" value="%d"/>`, q.WrapUpTime))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <param name="rejected-delay-time" value="%d"/>`, q.RejectDelayTime))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <param name="busy-delay-time" value="%d"/>`, q.BusyDelayTime))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <param name="no-answer-delay-time" value="%d"/>`, q.NoAnswerDelayTime))
		b.WriteString("\n")

		// Announcements
		if q.AnnounceSound != "" {
			b.WriteString(fmt.Sprintf(`          <param name="announce-sound" value="%s"/>`, xmlEscape(q.AnnounceSound)))
			b.WriteString("\n")
		}
		if q.AnnounceFrequency > 0 {
			b.WriteString(fmt.Sprintf(`          <param name="announce-frequency" value="%d"/>`, q.AnnounceFrequency))
			b.WriteString("\n")
		}
		if q.AnnouncePosition {
			b.WriteString(`          <param name="announce-position" value="true"/>`)
			b.WriteString("\n")
		}

		b.WriteString(`        </queue>`)
		b.WriteString("\n")
	}

	b.WriteString(`      </queues>`)
	b.WriteString("\n")

	// Agents
	b.WriteString(`      <agents>`)
	b.WriteString("\n")

	var allAgents []models.QueueAgent
	h.DB.Find(&allAgents)

	for _, agent := range allAgents {
		contact := agent.Contact
		if contact == "" {
			// Default contact: build from agent name
			contact = fmt.Sprintf("user/%s", agent.AgentName)
		}

		b.WriteString(fmt.Sprintf(`        <agent name="%s" type="callback" contact="%s" status="%s" max-no-answer="%d" wrap-up-time="%d" reject-delay-time="0" busy-delay-time="0" no-answer-delay-time="%d"/>`,
			xmlEscape(agent.AgentName),
			xmlEscape(contact),
			xmlEscape(string(agent.Status)),
			agent.MaxNoAnswer,
			agent.WrapUpTime,
			agent.NoAnswerDelayTime))
		b.WriteString("\n")
	}

	b.WriteString(`      </agents>`)
	b.WriteString("\n")

	// Tiers (agent-to-queue mappings)
	b.WriteString(`      <tiers>`)
	b.WriteString("\n")

	for _, q := range queues {
		queueName := q.Name
		if domain, ok := tenantDomains[q.TenantID]; ok && domain != "" {
			queueName = fmt.Sprintf("%s@%s", q.Name, domain)
		}

		for _, agent := range q.Agents {
			b.WriteString(fmt.Sprintf(`        <tier agent="%s" queue="%s" level="%d" position="%d"/>`,
				xmlEscape(agent.AgentName),
				xmlEscape(queueName),
				agent.TierLevel,
				agent.TierPosition))
			b.WriteString("\n")
		}
	}

	b.WriteString(`      </tiers>`)
	b.WriteString("\n")

	b.WriteString(`    </configuration>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	log.WithField("queues", len(queues)).Debug("Generated callcenter.conf from database")

	return b.String()
}

// buildVoicemailConfig generates voicemail.conf XML
func (h *FSHandler) buildVoicemailConfig() string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="configuration">`)
	b.WriteString("\n")
	b.WriteString(`    <configuration name="voicemail.conf" description="Voicemail">`)
	b.WriteString("\n")

	// Global settings
	b.WriteString(`      <settings>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="file-extension" value="wav"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="record-threshold" value="200"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="record-silence-hits" value="5"/>`)
	b.WriteString("\n")
	b.WriteString(`        <param name="record-silence-threshold" value="200"/>`)
	b.WriteString("\n")
	b.WriteString(`      </settings>`)
	b.WriteString("\n")

	// Generate a profile per tenant domain
	var tenants []models.Tenant
	h.DB.Where("enabled = ?", true).Find(&tenants)

	b.WriteString(`      <profiles>`)
	b.WriteString("\n")

	// Default profile (fallback)
	b.WriteString(`        <profile name="default">`)
	b.WriteString("\n")
	b.WriteString(`          <param name="file-extension" value="wav"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="terminator-key" value="#"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="max-login-attempts" value="3"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="digit-timeout" value="10000"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="max-record-len" value="180"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="max-retries" value="3"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="tone-spec" value="%(1000, 0, 640)"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="callback-dialplan" value="XML"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="callback-context" value="default"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="play-new-messages-key" value="1"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="play-saved-messages-key" value="2"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="login-keys" value="0"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="operator-extension" value="operator XML default"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="record-title" value="name_and_number"/>`)
	b.WriteString("\n")
	b.WriteString(`          <param name="storage-dir" value="/var/lib/callsign/voicemail"/>`)
	b.WriteString("\n")
	b.WriteString(`        </profile>`)
	b.WriteString("\n")

	// Per-tenant profiles (named by domain)
	for _, tenant := range tenants {
		if tenant.Domain == "" {
			continue
		}
		b.WriteString(fmt.Sprintf(`        <profile name="%s">`, xmlEscape(tenant.Domain)))
		b.WriteString("\n")
		b.WriteString(`          <param name="file-extension" value="wav"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="terminator-key" value="#"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="max-login-attempts" value="3"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="digit-timeout" value="10000"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="max-record-len" value="180"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="max-retries" value="3"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="tone-spec" value="%(1000, 0, 640)"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="callback-dialplan" value="XML"/>`)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <param name="callback-context" value="%s"/>`, xmlEscape(tenant.Domain)))
		b.WriteString("\n")
		b.WriteString(`          <param name="play-new-messages-key" value="1"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="play-saved-messages-key" value="2"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="login-keys" value="0"/>`)
		b.WriteString("\n")
		b.WriteString(`          <param name="record-title" value="name_and_number"/>`)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <param name="storage-dir" value="/var/lib/callsign/voicemail/%d"/>`, tenant.ID))
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

	log.WithField("tenants", len(tenants)).Debug("Generated voicemail.conf from database")

	return b.String()
}
