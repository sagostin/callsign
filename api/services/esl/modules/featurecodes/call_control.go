package featurecodes

import (
	"callsign/models"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// handlePickup handles call pickup (group and directed)
func handlePickup(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	// Check for directed pickup with extension in code (e.g., **1001)
	ext := ctx.GetCapture("ext")
	if ext == "" {
		ext = ctx.GetCapture("1")
	}

	if ext != "" {
		// Directed pickup of specific extension
		ctx.Conn.Execute("intercept", fmt.Sprintf("-bleg %s@%s", ext, ctx.Domain), true)
		log.WithField("target", ext).Info("Directed pickup")
		return
	}

	// Check if this is group pickup (*8) or directed (**)
	if fc.Code == "*8" || fc.ActionData == "group" {
		// Group pickup - pickup any ringing call in the same pickup group
		ctx.Conn.Execute("pickup", fmt.Sprintf("%s@%s", ctx.CallerID, ctx.Domain), true)
		log.Info("Group pickup")
	} else {
		// Directed pickup - prompt for extension
		ctx.Conn.Execute("playback", "ivr/ivr-enter_ext.wav", true)
		ev, _ := ctx.Conn.Execute("read", "2 6 tone_stream://%(250,50,440) ext 5000 #", true)
		targetExt := ev.Get("variable_ext")
		if targetExt != "" {
			ctx.Conn.Execute("intercept", fmt.Sprintf("-bleg %s@%s", targetExt, ctx.Domain), true)
			log.WithField("target", targetExt).Info("Directed pickup")
		}
	}
}

// handleIntercom handles intercom calls (auto-answer)
func handleIntercom(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	// Get target extension from capture or prompt
	ext := ctx.GetCapture("ext")
	if ext == "" {
		ext = ctx.GetCapture("1")
	}
	if ext == "" {
		ext = fc.ActionData
	}

	if ext == "" {
		ctx.Conn.Execute("playback", "ivr/ivr-enter_ext.wav", true)
		ev, _ := ctx.Conn.Execute("read", "2 6 tone_stream://%(250,50,440) ext 5000 #", true)
		ext = ev.Get("variable_ext")
	}

	if ext != "" {
		// Set auto-answer headers
		ctx.Conn.Execute("set", "sip_auto_answer=true", true)
		ctx.Conn.Execute("set", "sip_h_Alert-Info=<http://0.0.0.0>;info=alert-autoanswer;delay=0", true)
		ctx.Conn.Execute("set", "sip_h_Call-Info=<sip:0.0.0.0>;answer-after=0", true)

		// Bridge to target
		ctx.Conn.Execute("bridge", fmt.Sprintf("user/%s@%s", ext, ctx.Domain), true)

		log.WithField("target", ext).Info("Intercom call")
	}
}

// handlePageGroup handles paging to a group
func handlePageGroup(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	// Get group ID from feature code or capture
	var groupID uint
	if fc.GroupID != nil {
		groupID = *fc.GroupID
	}

	// Get paging group from database
	var group models.PageGroup
	if err := ctx.DB.Where("id = ? AND tenant_id = ?", groupID, ctx.TenantID).First(&group).Error; err != nil {
		ctx.Conn.Execute("playback", "ivr/ivr-invalid_selection.wav", true)
		return
	}

	// Set paging headers
	ctx.Conn.Execute("set", "sip_auto_answer=true", true)
	ctx.Conn.Execute("set", "sip_h_Alert-Info=<http://0.0.0.0>;info=alert-autoanswer;delay=0", true)

	// Get members
	var members []models.PageGroupDestination
	ctx.DB.Where("page_group_id = ?", groupID).Find(&members)

	// Build dial string for all members
	dialStrings := []string{}
	for _, member := range members {
		dialStrings = append(dialStrings, fmt.Sprintf("user/%s@%s", member.Destination, ctx.Domain))
	}

	if len(dialStrings) > 0 {
		// Page all members simultaneously
		ctx.Conn.Execute("bridge", fmt.Sprintf("{ignore_early_media=true}%s",
			joinDialStrings(dialStrings, ",")), true)
	}
}

// joinDialStrings joins dial strings with separator
func joinDialStrings(strings []string, sep string) string {
	result := ""
	for i, s := range strings {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
