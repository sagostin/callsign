package featurecodes

import (
	"callsign/models"
	"fmt"
)

// handleVoicemail handles voicemail check feature codes
func handleVoicemail(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	if fc.Code == "*97" {
		// Check own voicemail
		ctx.Conn.Execute("voicemail", fmt.Sprintf("check default %s %s", ctx.Domain, ctx.CallerID), true)
	} else if fc.Code == "*98" {
		// Check voicemail with mailbox prompt
		ctx.Conn.Execute("voicemail", fmt.Sprintf("check default %s", ctx.Domain), true)
	} else {
		// Direct to specific mailbox from action_data
		if fc.ActionData != "" {
			ctx.Conn.Execute("voicemail", fmt.Sprintf("check default %s %s", ctx.Domain, fc.ActionData), true)
		} else {
			ctx.Conn.Execute("voicemail", fmt.Sprintf("check default %s", ctx.Domain), true)
		}
	}
}

// handleCallForward handles call forward enable/disable
func handleCallForward(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	// Determine if enabling or disabling
	isEnable := fc.Code == "*72" || fc.ActionData == "enable"

	if isEnable {
		// Prompt for destination
		ctx.Conn.Execute("playback", "ivr/ivr-enter_dest_number.wav", true)
		ev, err := ctx.Conn.Execute("read", "2 20 tone_stream://%(250,50,440);%(250,50,440) forward_dest 10000 #", true)
		if err != nil {
			return
		}
		dest := ev.Get("variable_forward_dest")
		if dest != "" {
			// Update extension in database
			ctx.DB.Model(&models.Extension{}).
				Where("extension = ? AND tenant_id = ?", ctx.CallerID, ctx.TenantID).
				Updates(map[string]interface{}{
					"forward_all_enabled":     true,
					"forward_all_destination": dest,
				})
			ctx.Service.clearDirectoryCache(ctx.Domain)
			models.UpdatePresence(ctx.DB, ctx.TenantID, ctx.CallerID, models.PresenceState("forwarded"))
			ctx.Conn.Execute("playback", "ivr/ivr-call_forwarding_is_now_enabled.wav", true)
		}
	} else {
		// Disable
		ctx.DB.Model(&models.Extension{}).
			Where("extension = ? AND tenant_id = ?", ctx.CallerID, ctx.TenantID).
			Update("forward_all_enabled", false)
		ctx.Service.clearDirectoryCache(ctx.Domain)
		ctx.Conn.Execute("playback", "ivr/ivr-call_forwarding_is_now_disabled.wav", true)
	}
}

// handleDND handles Do Not Disturb toggle
func handleDND(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	// Determine enable/disable
	enable := fc.Code == "*78" || fc.ActionData == "enable"

	ctx.DB.Model(&models.Extension{}).
		Where("extension = ? AND tenant_id = ?", ctx.CallerID, ctx.TenantID).
		Update("do_not_disturb", enable)

	// Update BLF presence
	if enable {
		models.UpdatePresence(ctx.DB, ctx.TenantID, ctx.CallerID, models.PresenceDND)
	} else {
		models.UpdatePresence(ctx.DB, ctx.TenantID, ctx.CallerID, models.PresenceAvailable)
	}

	ctx.Service.clearDirectoryCache(ctx.Domain)

	if enable {
		ctx.Conn.Execute("playback", "ivr/ivr-dnd_activated.wav", true)
	} else {
		ctx.Conn.Execute("playback", "ivr/ivr-dnd_deactivated.wav", true)
	}
}

// handleCallFlowToggle toggles day/night mode
func handleCallFlowToggle(ctx *ExecutionContext) {
	// Toggle call flow status
	var callFlow struct{ Status string }
	ctx.DB.Model(&models.Dialplan{}).
		Select("status").
		Where("tenant_id = ? AND app_uuid = ?", ctx.TenantID, "call_flow").
		First(&callFlow)

	newStatus := "day"
	if callFlow.Status == "day" {
		newStatus = "night"
	}

	ctx.DB.Model(&models.Dialplan{}).
		Where("tenant_id = ? AND app_uuid = ?", ctx.TenantID, "call_flow").
		Update("status", newStatus)

	ctx.Service.clearDialplanCache(ctx.Domain)

	if newStatus == "night" {
		ctx.Conn.Execute("playback", "ivr/ivr-night_mode.wav", true)
	} else {
		ctx.Conn.Execute("playback", "ivr/ivr-day_mode.wav", true)
	}
}

// handleRecord handles call recording toggle
func handleRecord(ctx *ExecutionContext) {
	// Toggle recording for the user's next call
	ctx.DB.Model(&models.Extension{}).
		Where("extension = ? AND tenant_id = ?", ctx.CallerID, ctx.TenantID).
		Updates(map[string]interface{}{
			"record_inbound":  true,
			"record_outbound": true,
		})

	ctx.Conn.Execute("playback", "ivr/ivr-recording_enabled.wav", true)
}
