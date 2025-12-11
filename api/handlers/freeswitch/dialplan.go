package freeswitch

import (
	"callsign/models"
	"callsign/services/xmlcache"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// handleDialplan processes dialplan section requests for call routing
func (h *FSHandler) handleDialplan(req *XMLCurlRequest) string {
	context := req.Context
	if context == "" {
		context = "default"
	}

	// Determine dialplan mode
	// "single" mode: lookup specific destination (inbound routes)
	// "multiple" mode: get all dialplans for context (internal calls)
	mode := "multiple"
	if context == "public" && req.DestinationNumber != "" {
		mode = "single"
	}

	var cacheKey string
	if mode == "single" {
		cacheKey = xmlcache.DialplanSingleKey(context, req.DestinationNumber)
	} else {
		cacheKey = xmlcache.DialplanKey(context)
	}

	// Check cache
	if cached, ok := h.Cache.Get(cacheKey); ok {
		log.Debugf("Dialplan cache hit for context=%s", context)
		return cached
	}

	var xml string
	if mode == "single" {
		xml = h.buildSingleDialplan(req)
	} else {
		xml = h.buildMultiDialplan(req)
	}

	if xml != "" {
		h.Cache.Set(cacheKey, xml, CacheTTL.Dialplan)
	}

	return xml
}

// buildSingleDialplan looks up a specific destination number (inbound route)
func (h *FSHandler) buildSingleDialplan(req *XMLCurlRequest) string {
	// Look up destination by number
	var dest models.Destination
	result := h.DB.Where(
		"(destination_number = ? OR destination_number_regex ~ ?) AND enabled = ?",
		req.DestinationNumber, "^"+req.DestinationNumber+"$", true,
	).Order("destination_order ASC").First(&dest)

	if result.Error != nil {
		// No destination found, try to find a matching dialplan by regex
		return h.findDialplanByPattern(req)
	}

	// Build XML for this destination
	return h.buildDestinationXML(&dest, req)
}

// buildMultiDialplan gets all dialplans for a context
func (h *FSHandler) buildMultiDialplan(req *XMLCurlRequest) string {
	var dialplans []models.Dialplan

	// Get dialplans for this context (and global ones)
	h.DB.Where(
		"dialplan_context IN ? AND enabled = ?",
		[]string{req.Context, "${domain_name}", "global"},
		true,
	).Order("dialplan_order ASC").Find(&dialplans)

	if len(dialplans) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="dialplan" description="Dialplan">`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`    <context name="%s">`, xmlEscape(req.Context)))
	b.WriteString("\n")

	for _, dp := range dialplans {
		if dp.DialplanXML != "" {
			// Use pre-generated XML
			b.WriteString(dp.DialplanXML)
			b.WriteString("\n")
		} else {
			// Generate XML from details
			xml := h.buildDialplanFromDetails(&dp)
			b.WriteString(xml)
		}
	}

	b.WriteString(`    </context>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}

// buildDestinationXML generates dialplan XML for an inbound destination
func (h *FSHandler) buildDestinationXML(dest *models.Destination, req *XMLCurlRequest) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="dialplan" description="Dialplan">`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`    <context name="%s">`, xmlEscape(req.Context)))
	b.WriteString("\n")

	// Build extension for this destination
	b.WriteString(fmt.Sprintf(`      <extension name="%s">`, xmlEscape(dest.DestinationNumber)))
	b.WriteString("\n")

	// Condition to match the destination number
	b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$">`,
		xmlEscape(dest.DestinationNumber)))
	b.WriteString("\n")

	// Caller ID manipulation
	if dest.CallerIDNamePrefix != "" {
		b.WriteString(fmt.Sprintf(`          <action application="set" data="effective_caller_id_name=%s${caller_id_name}"/>`,
			xmlEscape(dest.CallerIDNamePrefix)))
		b.WriteString("\n")
	}

	// Set recording if enabled
	if dest.RecordEnabled {
		b.WriteString(`          <action application="set" data="record_session=true"/>`)
		b.WriteString("\n")
	}

	// Set account code
	if dest.AccountCode != "" {
		b.WriteString(fmt.Sprintf(`          <action application="set" data="accountcode=%s"/>`,
			xmlEscape(dest.AccountCode)))
		b.WriteString("\n")
	}

	// Execute the destination action
	if dest.DestinationAction != "" {
		// Parse action: "transfer 1001 XML default" or "bridge user/1001@domain"
		action := dest.DestinationAction
		parts := strings.SplitN(action, " ", 2)
		if len(parts) == 2 {
			b.WriteString(fmt.Sprintf(`          <action application="%s" data="%s"/>`,
				xmlEscape(parts[0]), xmlEscape(parts[1])))
		} else {
			b.WriteString(fmt.Sprintf(`          <action application="%s"/>`, xmlEscape(action)))
		}
		b.WriteString("\n")
	}

	b.WriteString(`        </condition>`)
	b.WriteString("\n")
	b.WriteString(`      </extension>`)
	b.WriteString("\n")
	b.WriteString(`    </context>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}

// findDialplanByPattern searches for dialplans matching the destination number
func (h *FSHandler) findDialplanByPattern(req *XMLCurlRequest) string {
	// This would search dialplan_details for matching conditions
	// For now, return empty to fall back to static config
	return ""
}

// buildDialplanFromDetails generates XML from dialplan details (conditions/actions)
func (h *FSHandler) buildDialplanFromDetails(dp *models.Dialplan) string {
	// Load details
	var details []models.DialplanDetail
	h.DB.Where("dialplan_uuid = ? AND enabled = ?", dp.UUID, true).
		Order("detail_group ASC, detail_order ASC").
		Find(&details)

	if len(details) == 0 {
		return ""
	}

	var b strings.Builder

	continueStr := ""
	if dp.Continue {
		continueStr = ` continue="true"`
	}

	b.WriteString(fmt.Sprintf(`      <extension name="%s"%s>`, xmlEscape(dp.DialplanName), continueStr))
	b.WriteString("\n")

	// Group details by condition groups
	currentGroup := -1
	inCondition := false

	for _, detail := range details {
		// Start new condition group
		if detail.DetailGroup != currentGroup {
			if inCondition {
				b.WriteString(`        </condition>`)
				b.WriteString("\n")
			}
			currentGroup = detail.DetailGroup
			inCondition = false
		}

		switch detail.DetailType {
		case "condition":
			if inCondition {
				b.WriteString(`        </condition>`)
				b.WriteString("\n")
			}
			breakStr := ""
			if detail.ConditionBreak != "" {
				breakStr = fmt.Sprintf(` break="%s"`, xmlEscape(detail.ConditionBreak))
			}
			b.WriteString(fmt.Sprintf(`        <condition field="%s" expression="%s"%s>`,
				xmlEscape(detail.ConditionField),
				xmlEscape(detail.ConditionExpression),
				breakStr))
			b.WriteString("\n")
			inCondition = true

		case "action":
			b.WriteString(fmt.Sprintf(`          <action application="%s" data="%s"/>`,
				xmlEscape(detail.ActionApplication),
				xmlEscape(detail.ActionData)))
			b.WriteString("\n")

		case "anti-action":
			b.WriteString(fmt.Sprintf(`          <anti-action application="%s" data="%s"/>`,
				xmlEscape(detail.ActionApplication),
				xmlEscape(detail.ActionData)))
			b.WriteString("\n")
		}
	}

	if inCondition {
		b.WriteString(`        </condition>`)
		b.WriteString("\n")
	}

	b.WriteString(`      </extension>`)
	b.WriteString("\n")

	return b.String()
}

// handlePhrases processes phrase section requests
func (h *FSHandler) handlePhrases(req *XMLCurlRequest) string {
	// TODO: Implement phrase handling
	// For now, return empty to fall back to static config
	return ""
}
