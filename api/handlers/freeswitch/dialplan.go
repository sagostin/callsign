package freeswitch

import (
	"callsign/models"
	"callsign/services/xmlcache"
	"fmt"
	"regexp"
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
			// Real regex matching
			if re, err := regexp.Compile(block.Number); err == nil {
				matched = re.MatchString(callerNumber)
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
	hasContent := false

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
		hasContent = true
	}

	// 2. Add feature codes for this context (if internal/domain context)
	if req.Context != "public" {
		featureXML := h.buildFeatureCodeDialplans(req)
		if featureXML != "" {
			b.WriteString(featureXML)
			hasContent = true
		}

		// 2b. Add time conditions and call flows
		timeCondXML := h.buildTimeConditionDialplans(req)
		if timeCondXML != "" {
			b.WriteString(timeCondXML)
			hasContent = true
		}

		// 2c. Add ring group routing
		ringGroupXML := h.buildRingGroupDialplans(req)
		if ringGroupXML != "" {
			b.WriteString(ringGroupXML)
			hasContent = true
		}

		// 2d. Add conference room routing
		conferenceXML := h.buildConferenceDialplans(req)
		if conferenceXML != "" {
			b.WriteString(conferenceXML)
			hasContent = true
		}

		// 2e. Add queue routing
		queueXML := h.buildQueueDialplans(req)
		if queueXML != "" {
			b.WriteString(queueXML)
			hasContent = true
		}

		// 2f. Add outbound routes (gateway routing for PSTN calls)
		outboundXML := h.buildOutboundRouteDialplans(req)
		if outboundXML != "" {
			b.WriteString(outboundXML)
			hasContent = true
		}

		// 2g. Add per-extension routing (rings all registered endpoints for each extension)
		extensionXML := h.buildExtensionDialplans(req)
		if extensionXML != "" {
			b.WriteString(extensionXML)
			hasContent = true
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
		hasContent = true
	}

	// If nothing was added anywhere, return empty for fallback
	if !hasContent {
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
	// Search existing dialplans for matching conditions
	var dialplans []models.Dialplan
	h.DB.Where("enabled = ?", true).Preload("Details", "enabled = ? AND detail_type = 'condition'", true).Find(&dialplans)

	for _, dp := range dialplans {
		for _, detail := range dp.Details {
			if detail.ConditionField == "destination_number" && detail.ConditionExpression != "" {
				// Actually test the regex against the destination number
				re, err := regexp.Compile(detail.ConditionExpression)
				if err != nil {
					continue // Invalid regex, skip
				}
				if !re.MatchString(req.DestinationNumber) {
					continue // Doesn't match
				}
				if dp.DialplanXML != "" {
					// Return the pre-generated XML wrapped in document
					var b strings.Builder
					b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
					b.WriteString("\n")
					b.WriteString(`<document type="freeswitch/xml">`)
					b.WriteString("\n")
					b.WriteString(`  <section name="dialplan" description="Dialplan">`)
					b.WriteString("\n")
					b.WriteString(fmt.Sprintf(`    <context name="%s">`, xmlEscape(req.Context)))
					b.WriteString("\n")
					b.WriteString(dp.DialplanXML)
					b.WriteString("\n")
					b.WriteString(`    </context>`)
					b.WriteString("\n")
					b.WriteString(`  </section>`)
					b.WriteString("\n")
					b.WriteString(`</document>`)
					return b.String()
				}
			}
		}
	}

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
		// No tenant found — no feature codes to generate
		return ""
	}

	// Get feature codes (strictly per-tenant, no global fallback)
	var featureCodes []models.FeatureCode
	h.DB.Where(
		"tenant_id = ? AND enabled = ?",
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

		// Route to featurecodes ESL socket for processing
		b.WriteString(`          <action application="socket" data="127.0.0.6:9001 async full"/>`)
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

		// Build time expression
		// FreeSWITCH time format: wday, hour, minute, month, mday, year, yday
		timeExpr := h.buildTimeExpression(&tc)
		if timeExpr != "" {
			// FreeSWITCH requires sibling conditions, not nested ones.
			// First condition matches destination_number with break="never" so processing continues.
			// Second condition tests the time expression with actions/anti-actions.
			b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$" break="never"/>`, xmlEscape(tc.Extension)))
			b.WriteString("\n")

			b.WriteString(fmt.Sprintf(`        <condition %s>`, timeExpr))
			b.WriteString("\n")

			// Match action (within time)
			matchAction := h.buildDestinationAction(tc.MatchDestType, tc.MatchDestValue, domain)
			b.WriteString(fmt.Sprintf(`          <action application="%s" data="%s"/>`,
				matchAction.App, xmlEscape(matchAction.Data)))
			b.WriteString("\n")

			// Anti-action (outside time)
			noMatchAction := h.buildDestinationAction(tc.NoMatchDestType, tc.NoMatchDestValue, domain)
			b.WriteString(fmt.Sprintf(`          <anti-action application="%s" data="%s"/>`,
				noMatchAction.App, xmlEscape(noMatchAction.Data)))
			b.WriteString("\n")

			b.WriteString(`        </condition>`)
			b.WriteString("\n")
		} else {
			// No time expression, just match by destination and route
			b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$">`, xmlEscape(tc.Extension)))
			b.WriteString("\n")
			matchAction := h.buildDestinationAction(tc.MatchDestType, tc.MatchDestValue, domain)
			b.WriteString(fmt.Sprintf(`          <action application="%s" data="%s"/>`,
				matchAction.App, xmlEscape(matchAction.Data)))
			b.WriteString("\n")
			b.WriteString(`        </condition>`)
			b.WriteString("\n")
		}

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

		// Route based on current state index
		var destType, destValue string
		if len(cf.Destinations) > 0 && cf.CurrentState < len(cf.Destinations) {
			dest := cf.Destinations[cf.CurrentState]
			destType = dest.DestType
			destValue = dest.DestValue
		}

		action := h.buildDestinationAction(destType, destValue, domain)
		b.WriteString(fmt.Sprintf(`          <action application="set" data="call_flow_uuid=%s"/>`, cf.UUID.String()))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="call_flow_state=%d"/>`, cf.CurrentState))
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

// buildRingGroupDialplans generates dialplan entries for ring groups
func (h *FSHandler) buildRingGroupDialplans(req *XMLCurlRequest) string {
	domain := req.Context
	if domain == "default" || domain == "" {
		domain = req.Domain
	}

	var tenant models.Tenant
	if err := h.DB.Where("domain = ?", domain).First(&tenant).Error; err != nil {
		return ""
	}

	var ringGroups []models.RingGroup
	h.DB.Where("tenant_id = ? AND enabled = ? AND extension != ''", tenant.ID, true).Find(&ringGroups)

	if len(ringGroups) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(`      <!-- Ring Groups -->`)
	b.WriteString("\n")

	for _, rg := range ringGroups {
		b.WriteString(fmt.Sprintf(`      <extension name="rg_%s" continue="false">`, xmlEscape(rg.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$">`, xmlEscape(rg.Extension)))
		b.WriteString("\n")

		// Set ring group info as variables
		b.WriteString(fmt.Sprintf(`          <action application="set" data="ring_group_uuid=%s"/>`, rg.UUID.String()))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="ring_group_name=%s"/>`, xmlEscape(rg.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="ring_group_strategy=%s"/>`, xmlEscape(string(rg.Strategy))))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="tenant_id=%d"/>`, tenant.ID))
		b.WriteString("\n")

		// Caller ID prefix
		if rg.CallerIDNamePrefix != "" {
			b.WriteString(fmt.Sprintf(`          <action application="set" data="effective_caller_id_name=%s${caller_id_name}"/>`, xmlEscape(rg.CallerIDNamePrefix)))
			b.WriteString("\n")
		}

		b.WriteString(`          <action application="set" data="hangup_after_bridge=true"/>`)
		b.WriteString("\n")
		b.WriteString(`          <action application="set" data="continue_on_fail=true"/>`)
		b.WriteString("\n")

		// Route to callcontrol ESL socket for ring group handling
		b.WriteString(`          <action application="socket" data="127.0.0.1:9001 async full"/>`)
		b.WriteString("\n")

		b.WriteString(`        </condition>`)
		b.WriteString("\n")
		b.WriteString(`      </extension>`)
		b.WriteString("\n")
	}

	return b.String()
}

// buildConferenceDialplans generates dialplan entries for conference rooms
func (h *FSHandler) buildConferenceDialplans(req *XMLCurlRequest) string {
	domain := req.Context
	if domain == "default" || domain == "" {
		domain = req.Domain
	}

	var tenant models.Tenant
	if err := h.DB.Where("domain = ?", domain).First(&tenant).Error; err != nil {
		return ""
	}

	var conferences []models.Conference
	h.DB.Where("tenant_id = ? AND enabled = ? AND extension != ''", tenant.ID, true).Find(&conferences)

	if len(conferences) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(`      <!-- Conference Rooms -->`)
	b.WriteString("\n")

	for _, conf := range conferences {
		b.WriteString(fmt.Sprintf(`      <extension name="conf_%s" continue="false">`, xmlEscape(conf.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$">`, xmlEscape(conf.Extension)))
		b.WriteString("\n")

		b.WriteString(fmt.Sprintf(`          <action application="set" data="conference_uuid=%s"/>`, conf.UUID.String()))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="tenant_id=%d"/>`, tenant.ID))
		b.WriteString("\n")

		// Route to conference ESL socket
		b.WriteString(`          <action application="socket" data="127.0.0.1:9001 async full"/>`)
		b.WriteString("\n")

		b.WriteString(`        </condition>`)
		b.WriteString("\n")
		b.WriteString(`      </extension>`)
		b.WriteString("\n")
	}

	return b.String()
}

// buildQueueDialplans generates dialplan entries for call center queues
func (h *FSHandler) buildQueueDialplans(req *XMLCurlRequest) string {
	domain := req.Context
	if domain == "default" || domain == "" {
		domain = req.Domain
	}

	var tenant models.Tenant
	if err := h.DB.Where("domain = ?", domain).First(&tenant).Error; err != nil {
		return ""
	}

	var queues []models.Queue
	h.DB.Where("tenant_id = ? AND enabled = ? AND extension != ''", tenant.ID, true).Find(&queues)

	if len(queues) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(`      <!-- Call Center Queues -->`)
	b.WriteString("\n")

	for _, q := range queues {
		b.WriteString(fmt.Sprintf(`      <extension name="queue_%s" continue="false">`, xmlEscape(q.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$">`, xmlEscape(q.Extension)))
		b.WriteString("\n")

		b.WriteString(fmt.Sprintf(`          <action application="set" data="queue_name=%s"/>`, xmlEscape(q.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="tenant_id=%d"/>`, tenant.ID))
		b.WriteString("\n")

		// Route to queue ESL socket
		b.WriteString(`          <action application="socket" data="127.0.0.3:9001 async full"/>`)
		b.WriteString("\n")

		b.WriteString(`        </condition>`)
		b.WriteString("\n")
		b.WriteString(`      </extension>`)
		b.WriteString("\n")
	}

	return b.String()
}

// buildOutboundRouteDialplans generates dialplan entries for outbound calls via gateways
func (h *FSHandler) buildOutboundRouteDialplans(req *XMLCurlRequest) string {
	var routes []models.DefaultOutboundRoute
	h.DB.Where("enabled = ?", true).Order("\"order\" ASC").Find(&routes)

	if len(routes) == 0 {
		return ""
	}

	// Build a map of gateways by ID
	var gateways []models.Gateway
	h.DB.Where("enabled = ?", true).Find(&gateways)
	gatewayMap := make(map[uint]*models.Gateway)
	for i := range gateways {
		gatewayMap[gateways[i].ID] = &gateways[i]
	}

	var b strings.Builder
	b.WriteString(`      <!-- Outbound Routes -->`)
	b.WriteString("\n")

	for _, route := range routes {
		gw := gatewayMap[route.GatewayID]
		if gw == nil {
			continue // No gateway assigned, skip
		}

		// Build regex pattern from route config
		var pattern string
		if route.DialString != "" {
			// Custom dial string pattern
			pattern = route.DialString
		} else {
			// Build from digit prefix/min/max
			if route.DigitPrefix != "" {
				digitRange := fmt.Sprintf("{%d,%d}", route.DigitMin-len(route.DigitPrefix), route.DigitMax-len(route.DigitPrefix))
				pattern = fmt.Sprintf("^%s(\\d%s)$", route.DigitPrefix, digitRange)
			} else {
				pattern = fmt.Sprintf("^(\\d{%d,%d})$", route.DigitMin, route.DigitMax)
			}
		}

		b.WriteString(fmt.Sprintf(`      <extension name="outbound_%s" continue="true">`, xmlEscape(route.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="%s">`, xmlEscape(pattern)))
		b.WriteString("\n")

		// Set outbound caller ID and variables
		b.WriteString(`          <action application="set" data="hangup_after_bridge=true"/>`)
		b.WriteString("\n")
		b.WriteString(`          <action application="set" data="continue_on_fail=true"/>`)
		b.WriteString("\n")

		// Build the bridge destination
		// $1 is the captured digits from the regex
		dialDest := "$1"
		if route.StripDigits > 0 {
			// Stripping is handled by regex capture group
			dialDest = "$1"
		}
		if route.PrependDigits != "" {
			dialDest = route.PrependDigits + dialDest
		}

		bridgeStr := fmt.Sprintf("sofia/gateway/%s/%s", gw.GatewayName, dialDest)

		// Add failover gateway if configured
		if route.Gateway2ID != nil {
			gw2 := gatewayMap[*route.Gateway2ID]
			if gw2 != nil {
				bridgeStr += fmt.Sprintf("|sofia/gateway/%s/%s", gw2.GatewayName, dialDest)
			}
		}

		b.WriteString(fmt.Sprintf(`          <action application="bridge" data="%s"/>`, xmlEscape(bridgeStr)))
		b.WriteString("\n")

		b.WriteString(`        </condition>`)
		b.WriteString("\n")
		b.WriteString(`      </extension>`)
		b.WriteString("\n")
	}

	return b.String()
}

// buildExtensionDialplans generates per-extension routing that sends calls
// to the ESL socket for multi-device ringing across all registered endpoints.
// For each extension in the tenant, this creates a dialplan entry that:
//  1. Sets extension metadata (UUID, ring strategy, tenant_id)
//  2. Routes to the ESL call control socket which handles ringing all
//     registered endpoints (devices, apps, web clients) per the extension's
//     ring strategy (simultaneous or sequential)
func (h *FSHandler) buildExtensionDialplans(req *XMLCurlRequest) string {
	domain := req.Context
	if domain == "default" || domain == "" {
		domain = req.Domain
	}

	var tenant models.Tenant
	if err := h.DB.Where("domain = ?", domain).First(&tenant).Error; err != nil {
		return ""
	}

	var extensions []models.Extension
	h.DB.Where("tenant_id = ? AND enabled = ?", tenant.ID, true).Find(&extensions)

	if len(extensions) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(`      <!-- Extension Routing (multi-device) -->`)
	b.WriteString("\n")

	for _, ext := range extensions {
		b.WriteString(fmt.Sprintf(`      <extension name="ext_%s" continue="false">`, xmlEscape(ext.Extension)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`        <condition field="destination_number" expression="^%s$">`, xmlEscape(ext.Extension)))
		b.WriteString("\n")

		// Set extension metadata
		b.WriteString(fmt.Sprintf(`          <action application="set" data="extension_uuid=%s"/>`, ext.UUID.String()))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="dialed_extension=%s"/>`, xmlEscape(ext.Extension)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="ring_strategy=%s"/>`, xmlEscape(ext.RingStrategy)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="tenant_id=%d"/>`, tenant.ID))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(`          <action application="set" data="call_timeout=%d"/>`, ext.CallTimeout))
		b.WriteString("\n")

		// Caller ID for the extension
		if ext.EffectiveCallerIDName != "" {
			b.WriteString(fmt.Sprintf(`          <action application="set" data="callee_id_name=%s"/>`, xmlEscape(ext.EffectiveCallerIDName)))
			b.WriteString("\n")
		}
		if ext.EffectiveCallerIDNumber != "" {
			b.WriteString(fmt.Sprintf(`          <action application="set" data="callee_id_number=%s"/>`, xmlEscape(ext.EffectiveCallerIDNumber)))
			b.WriteString("\n")
		}

		b.WriteString(`          <action application="set" data="hangup_after_bridge=true"/>`)
		b.WriteString("\n")
		b.WriteString(`          <action application="set" data="continue_on_fail=true"/>`)
		b.WriteString("\n")

		// Route to ESL socket for smart multi-device ringing
		// The ESL handler queries all ClientRegistrations for this extension
		// and builds the appropriate dial string based on ring_strategy
		b.WriteString(`          <action application="socket" data="127.0.0.1:9001 async full"/>`)
		b.WriteString("\n")

		b.WriteString(`        </condition>`)
		b.WriteString("\n")
		b.WriteString(`      </extension>`)
		b.WriteString("\n")
	}

	return b.String()
}

// handlePhrases processes phrase section requests from FreeSWITCH mod_xml_curl
// It looks up phrase macros in the phrases table and generates proper XML for playback
func (h *FSHandler) handlePhrases(req *XMLCurlRequest) string {
	// Guard: Only handle macro lookups
	if req.KeyName != "macro" {
		return ""
	}

	// Guard: Must have a macro name
	macroName := req.KeyValue
	if macroName == "" {
		return ""
	}

	// Look up phrase by macro name
	var phrase models.Phrase
	result := h.DB.Where("macro_name = ? AND enabled = ?", macroName, true).First(&phrase)
	if result.Error != nil {
		// Phrase not found - fall back to static config
		return ""
	}

	// Generate phrase XML
	return h.buildPhraseXML(&phrase)
}

// buildPhraseXML generates FreeSWITCH phrase macro XML for a phrase
func (h *FSHandler) buildPhraseXML(phrase *models.Phrase) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("\n")
	b.WriteString(`<document type="freeswitch/xml">`)
	b.WriteString("\n")
	b.WriteString(`  <section name="phrases">`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`    <macros>`))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`      <macro name="%s">`, xmlEscape(phrase.MacroName)))
	b.WriteString("\n")
	b.WriteString(`        <input pattern="(.*)">`)
	b.WriteString("\n")
	b.WriteString(`          <match>`)
	b.WriteString("\n")

	// Generate action based on phrase type
	switch phrase.Type {
	case "tts":
		// TTS: speak using the configured module and content
		// Format: module|voice|text
		module := phrase.Module
		if module == "" {
			module = "flite"
		}
		b.WriteString(fmt.Sprintf(`            <action function="speak" data="%s|ssf|%s"/>`,
			xmlEscape(module), xmlEscape(phrase.Content)))
	case "file":
		// Audio file: play-file
		b.WriteString(fmt.Sprintf(`            <action function="play-file" data="%s"/>`,
			xmlEscape(phrase.Content)))
	case "prompt":
		// Prompts use the phrase system - might be TTS or file depending on setup
		// Treat as play-file with the content as the file path
		b.WriteString(fmt.Sprintf(`            <action function="play-file" data="%s"/>`,
			xmlEscape(phrase.Content)))
	default:
		// Default to TTS for backwards compatibility
		b.WriteString(fmt.Sprintf(`            <action function="speak" data="flite|ssf|%s"/>`,
			xmlEscape(phrase.Content)))
	}

	b.WriteString("\n")
	b.WriteString(`          </match>`)
	b.WriteString("\n")
	b.WriteString(`        </input>`)
	b.WriteString("\n")
	b.WriteString(`      </macro>`)
	b.WriteString("\n")
	b.WriteString(`    </macros>`)
	b.WriteString("\n")
	b.WriteString(`  </section>`)
	b.WriteString("\n")
	b.WriteString(`</document>`)

	return b.String()
}
