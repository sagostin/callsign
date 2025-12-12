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
// Order of processing: 1) Call Blocks, 2) Feature Codes, 3) Destinations
func (h *FSHandler) buildSingleDialplan(req *XMLCurlRequest) string {
	// First, check if caller is blocked
	if blockXML := h.checkCallBlocks(req); blockXML != "" {
		return blockXML
	}

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

// checkCallBlocks checks if the caller is on a block list and returns reject XML if so
func (h *FSHandler) checkCallBlocks(req *XMLCurlRequest) string {
	callerNumber := req.CallerIDNumber
	if callerNumber == "" {
		return ""
	}

	// Find destination to get tenant ID
	var dest models.Destination
	if err := h.DB.Where("destination_number = ? AND enabled = ?", req.DestinationNumber, true).First(&dest).Error; err != nil {
		return "" // No destination, no block check
	}

	// Check for matching call blocks for this tenant
	var blocks []models.CallBlock
	h.DB.Where("tenant_id = ? AND enabled = ?", dest.TenantID, true).Find(&blocks)

	for _, block := range blocks {
		matched := false
		switch block.MatchType {
		case "exact":
			matched = callerNumber == block.Number || "+"+callerNumber == block.Number || callerNumber == "+"+block.Number
		case "prefix":
			matched = strings.HasPrefix(callerNumber, block.Number) || strings.HasPrefix(callerNumber, "+"+block.Number)
		case "regex":
			// Simple regex matching
			if strings.HasPrefix(callerNumber, block.Number) {
				matched = true
			}
		default:
			matched = callerNumber == block.Number
		}

		if matched {
			log.Infof("Call blocked: caller=%s matched block=%s (type=%s)", callerNumber, block.Number, block.MatchType)
			return h.buildBlockedCallerXML(req, &block)
		}
	}

	return ""
}

// buildBlockedCallerXML generates XML to reject a blocked caller
func (h *FSHandler) buildBlockedCallerXML(req *XMLCurlRequest, block *models.CallBlock) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="dialplan" description="Dialplan">`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`    <context name="%s">`, xmlEscape(req.Context)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`      <extension name="blocked_caller_%s">`, xmlEscape(block.Number)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`        <condition field="caller_id_number" expression="%s">`, xmlEscape(block.Number)))
	b.WriteString("\n")

	// Action based on block type
	switch block.Action {
	case "busy":
		b.WriteString(`          <action application="respond" data="486 Busy Here"/>`)
	case "hangup":
		b.WriteString(`          <action application="hangup" data="CALL_REJECTED"/>`)
	default: // reject
		b.WriteString(`          <action application="respond" data="603 Decline"/>`)
	}
	b.WriteString("\n")

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

// buildMultiDialplan gets all dialplans for a context including feature codes
func (h *FSHandler) buildMultiDialplan(req *XMLCurlRequest) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="dialplan" description="Dialplan">`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`    <context name="%s">`, xmlEscape(req.Context)))
	b.WriteString("\n")

	// 1. First, add global dialplans (lowest order, highest priority)
	var globalDialplans []models.Dialplan
	h.DB.Where(
		"dialplan_context = ? AND enabled = ?",
		"global",
		true,
	).Order("dialplan_order ASC").Find(&globalDialplans)

	for _, dp := range globalDialplans {
		if dp.DialplanXML != "" {
			b.WriteString(dp.DialplanXML)
			b.WriteString("\n")
		} else {
			xml := h.buildDialplanFromDetails(&dp)
			b.WriteString(xml)
		}
	}

	// 2. Add feature codes for this context (if internal/domain context)
	if req.Context != "public" {
		featureXML := h.buildFeatureCodeDialplans(req)
		if featureXML != "" {
			b.WriteString(featureXML)
		}

		// 2b. Add time conditions and call flows
		timeCondXML := h.buildTimeConditionDialplans(req)
		if timeCondXML != "" {
			b.WriteString(timeCondXML)
		}
	}

	// 3. Add context-specific and domain dialplans
	var dialplans []models.Dialplan
	h.DB.Where(
		"dialplan_context IN ? AND enabled = ?",
		[]string{req.Context, "${domain_name}"},
		true,
	).Order("dialplan_order ASC").Find(&dialplans)

	for _, dp := range dialplans {
		if dp.DialplanXML != "" {
			b.WriteString(dp.DialplanXML)
			b.WriteString("\n")
		} else {
			xml := h.buildDialplanFromDetails(&dp)
			b.WriteString(xml)
		}
	}

	// If nothing was added, return empty for fallback
	if len(globalDialplans) == 0 && len(dialplans) == 0 {
		return ""
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

// buildFeatureCodeDialplans generates dialplan XML for all feature codes
func (h *FSHandler) buildFeatureCodeDialplans(req *XMLCurlRequest) string {
	// Get domain from context (context is often the domain name)
	domain := req.Context
	if domain == "default" || domain == "" {
		domain = req.Domain
	}

	// Find tenant by domain
	var tenant models.Tenant
	if err := h.DB.Where("domain = ?", domain).First(&tenant).Error; err != nil {
		// No tenant found, just use global feature codes
		tenant.ID = 0
	}

	// Get feature codes (global + tenant-specific)
	var featureCodes []models.FeatureCode
	h.DB.Where(
		"(tenant_id IS NULL OR tenant_id = ?) AND enabled = ?",
		tenant.ID, true,
	).Order(`"order" ASC`).Find(&featureCodes)

	if len(featureCodes) == 0 {
		return ""
	}

	var b strings.Builder

	// Add comment for clarity
	b.WriteString(`      <!-- Feature Codes -->`)
	b.WriteString("\n")

	for _, fc := range featureCodes {
		// Use the model's ToDialplanXML method or build custom
		condition := fc.Code
		if fc.CodeRegex != "" {
			condition = fc.CodeRegex
		}

		b.WriteString(fmt.Sprintf(`      <extension name="fc_%s" continue="false">`, xmlEscape(fc.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="%s">`, xmlEscape(condition)))
		b.WriteString("\n")

		// Set feature code info as channel variables
		b.WriteString(fmt.Sprintf(`          <action application="set" data="feature_code_uuid=%s"/>`, fc.UUID.String()))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="feature_code_action=%s"/>`, string(fc.Action)))
		b.WriteString("\n")

		// Route to ESL socket for processing
		b.WriteString(`          <action application="socket" data="127.0.0.1:9001 async full"/>`)
		b.WriteString("\n")

		b.WriteString(`        </condition>`)
		b.WriteString("\n")
		b.WriteString(`      </extension>`)
		b.WriteString("\n")
	}

	return b.String()
}

// buildTimeConditionDialplans generates dialplan XML for time conditions and call flows
func (h *FSHandler) buildTimeConditionDialplans(req *XMLCurlRequest) string {
	// Get domain from context
	domain := req.Context
	if domain == "default" || domain == "" {
		domain = req.Domain
	}

	// Find tenant by domain
	var tenant models.Tenant
	if err := h.DB.Where("domain = ?", domain).First(&tenant).Error; err != nil {
		return ""
	}

	var b strings.Builder

	// 1. Generate Time Conditions
	var timeConditions []models.TimeCondition
	h.DB.Where("tenant_id = ? AND enabled = ?", tenant.ID, true).Find(&timeConditions)

	for _, tc := range timeConditions {
		if tc.Extension == "" {
			continue
		}

		b.WriteString(fmt.Sprintf(`      <!-- Time Condition: %s -->`, xmlEscape(tc.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`      <extension name="tc_%s" continue="false">`, tc.UUID.String()))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$">`, xmlEscape(tc.Extension)))
		b.WriteString("\n")

		// Build time expression
		// FreeSWITCH time format: wday, hour, minute, month, mday, year, yday
		timeExpr := h.buildTimeExpression(&tc)
		if timeExpr != "" {
			b.WriteString(fmt.Sprintf(`          <condition %s>`, timeExpr))
			b.WriteString("\n")

			// Match action (within time)
			matchAction := h.buildDestinationAction(tc.MatchDestType, tc.MatchDestValue, domain)
			b.WriteString(fmt.Sprintf(`            <action application="%s" data="%s"/>`,
				matchAction.App, xmlEscape(matchAction.Data)))
			b.WriteString("\n")

			// Anti-action (outside time)
			noMatchAction := h.buildDestinationAction(tc.NoMatchDestType, tc.NoMatchDestValue, domain)
			b.WriteString(fmt.Sprintf(`            <anti-action application="%s" data="%s"/>`,
				noMatchAction.App, xmlEscape(noMatchAction.Data)))
			b.WriteString("\n")

			b.WriteString(`          </condition>`)
			b.WriteString("\n")
		}

		b.WriteString(`        </condition>`)
		b.WriteString("\n")
		b.WriteString(`      </extension>`)
		b.WriteString("\n")
	}

	// 2. Generate Call Flows (Day/Night toggles)
	var callFlows []models.CallFlow
	h.DB.Where("tenant_id = ? AND enabled = ?", tenant.ID, true).Find(&callFlows)

	for _, cf := range callFlows {
		if cf.Extension == "" {
			continue
		}

		b.WriteString(fmt.Sprintf(`      <!-- Call Flow: %s -->`, xmlEscape(cf.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`      <extension name="cf_%s" continue="false">`, cf.UUID.String()))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$">`, xmlEscape(cf.Extension)))
		b.WriteString("\n")

		// Route based on current status
		var destType, destValue string
		switch cf.Status {
		case "night":
			destType = cf.NightDestType
			destValue = cf.NightDestValue
		case "holiday":
			destType = cf.HolidayDestType
			destValue = cf.HolidayDestValue
		default: // "day"
			destType = cf.DayDestType
			destValue = cf.DayDestValue
		}

		action := h.buildDestinationAction(destType, destValue, domain)
		b.WriteString(fmt.Sprintf(`          <action application="set" data="call_flow_uuid=%s"/>`, cf.UUID.String()))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="call_flow_status=%s"/>`, cf.Status))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="%s" data="%s"/>`,
			action.App, xmlEscape(action.Data)))
		b.WriteString("\n")

		b.WriteString(`        </condition>`)
		b.WriteString("\n")
		b.WriteString(`      </extension>`)
		b.WriteString("\n")
	}

	return b.String()
}

// buildTimeExpression creates a FreeSWITCH time expression from TimeCondition
func (h *FSHandler) buildTimeExpression(tc *models.TimeCondition) string {
	var parts []string

	// Weekdays (wday 1-7, Sunday=1)
	if len(tc.Weekdays) > 0 {
		// Convert to FreeSWITCH format (1-7)
		wdays := ""
		for i, d := range tc.Weekdays {
			if i > 0 {
				wdays += ","
			}
			wdays += fmt.Sprintf("%d", d+1) // 0-indexed to 1-indexed
		}
		parts = append(parts, fmt.Sprintf(`wday="%s"`, wdays))
	}

	// Time of day
	if tc.StartTime != "" && tc.EndTime != "" {
		// Format: hour="09:00-17:00"
		parts = append(parts, fmt.Sprintf(`time-of-day="%s-%s"`, tc.StartTime, tc.EndTime))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

// destinationAction holds parsed destination action
type destinationAction struct {
	App  string
	Data string
}

// buildDestinationAction converts destination type/value to FreeSWITCH action
func (h *FSHandler) buildDestinationAction(destType, destValue, domain string) destinationAction {
	switch destType {
	case "extension":
		return destinationAction{App: "transfer", Data: destValue + " XML " + domain}
	case "ivr":
		return destinationAction{App: "ivr", Data: destValue}
	case "voicemail":
		return destinationAction{App: "voicemail", Data: "default " + domain + " " + destValue}
	case "ring_group":
		return destinationAction{App: "transfer", Data: destValue + " XML " + domain}
	case "queue":
		return destinationAction{App: "callcenter", Data: destValue + "@" + domain}
	case "external":
		return destinationAction{App: "bridge", Data: "sofia/gateway/default/" + destValue}
	case "hangup":
		return destinationAction{App: "hangup", Data: "NORMAL_CLEARING"}
	case "playback":
		return destinationAction{App: "playback", Data: destValue}
	default:
		if destValue != "" {
			return destinationAction{App: "transfer", Data: destValue + " XML " + domain}
		}
		return destinationAction{App: "hangup", Data: "UNALLOCATED_NUMBER"}
	}
}

// handlePhrases processes phrase section requests
func (h *FSHandler) handlePhrases(req *XMLCurlRequest) string {
	// TODO: Implement phrase handling
	// For now, return empty to fall back to static config
	return ""
}
