package featurecodes

import (
	"callsign/models"
	"fmt"
	"time"
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

// handleRecord handles call recording toggle (*82 enable, *83 disable)
func handleRecord(ctx *ExecutionContext) {
	action := ctx.FeatureCode.ActionData // "disable" = turn off, else toggle/enable

	var ext models.Extension
	if err := ctx.DB.Where("extension = ? AND tenant_id = ?", ctx.CallerID, ctx.TenantID).
		First(&ext).Error; err != nil {
		ctx.Conn.Execute("playback", "ivr/ivr-invalid_selection.wav", true)
		return
	}

	if action == "disable" {
		// Disable recording
		ctx.DB.Model(&ext).Updates(map[string]interface{}{
			"record_inbound":  false,
			"record_outbound": false,
		})
		ctx.Conn.Execute("playback", "ivr/ivr-recording_disabled.wav", true)

		// Stop active recording if this extension has a bridged call
		s := ctx.Service
		manager := s.Manager()
		if manager != nil && manager.Client != nil {
			// Find any active channel for this extension and stop recording
			manager.Client.API(fmt.Sprintf(
				"uuid_record %s stop all", ctx.UUID,
			))
		}
	} else {
		// Enable recording (toggle: if already on, turn off)
		newInbound := !ext.RecordInbound
		newOutbound := !ext.RecordOutbound
		ctx.DB.Model(&ext).Updates(map[string]interface{}{
			"record_inbound":  newInbound,
			"record_outbound": newOutbound,
		})

		if newInbound {
			ctx.Conn.Execute("playback", "ivr/ivr-recording_enabled.wav", true)

			// Start recording on active call if currently bridged
			s := ctx.Service
			manager := s.Manager()
			if manager != nil && manager.Client != nil {
				recordPath := fmt.Sprintf(
					"/var/lib/freeswitch/recordings/%s/%s_%s.wav",
					ctx.Domain, ctx.CallerID,
					time.Now().Format("20060102_150405"),
				)
				manager.Client.API(fmt.Sprintf(
					"uuid_record %s start %s", ctx.UUID, recordPath,
				))
			}
		} else {
			ctx.Conn.Execute("playback", "ivr/ivr-recording_disabled.wav", true)
		}
	}

	// Flush directory cache so FS picks up changes on next call
	ctx.Service.clearDirectoryCache(ctx.Domain)
}

// handleQueueLogin handles agent login feature code (*90 by default)
func handleQueueLogin(ctx *ExecutionContext) {
	var agent models.QueueAgent
	// Find agent by matching the caller's extension within this tenant
	if err := ctx.DB.Where("tenant_id = ? AND contact LIKE ?", ctx.TenantID, "%"+ctx.CallerID+"%").
		First(&agent).Error; err != nil {
		// No agent found — try by agent_name matching extension@domain
		agentName := fmt.Sprintf("%s@%s", ctx.CallerID, ctx.Domain)
		if err2 := ctx.DB.Where("tenant_id = ? AND agent_name = ?", ctx.TenantID, agentName).
			First(&agent).Error; err2 != nil {
			ctx.Conn.Execute("playback", "ivr/ivr-not_authorized.wav", true)
			return
		}
	}

	// Update status in DB
	agent.Status = models.AgentStatusAvailable
	ctx.DB.Save(&agent)

	// Sync to FreeSWITCH
	manager := ctx.Service.Manager()
	if manager != nil && manager.Client != nil {
		manager.Client.API(fmt.Sprintf("callcenter_config agent set status %s 'Available'", agent.AgentName))
	}

	ctx.Conn.Execute("playback", "ivr/ivr-agent_logged_in.wav", true)
}

// handleQueueLogout handles agent logout feature code (*91 by default)
func handleQueueLogout(ctx *ExecutionContext) {
	var agent models.QueueAgent
	if err := ctx.DB.Where("tenant_id = ? AND contact LIKE ?", ctx.TenantID, "%"+ctx.CallerID+"%").
		First(&agent).Error; err != nil {
		agentName := fmt.Sprintf("%s@%s", ctx.CallerID, ctx.Domain)
		if err2 := ctx.DB.Where("tenant_id = ? AND agent_name = ?", ctx.TenantID, agentName).
			First(&agent).Error; err2 != nil {
			ctx.Conn.Execute("playback", "ivr/ivr-not_authorized.wav", true)
			return
		}
	}

	// Update status in DB
	agent.Status = models.AgentStatusLoggedOut
	ctx.DB.Save(&agent)

	// Sync to FreeSWITCH
	manager := ctx.Service.Manager()
	if manager != nil && manager.Client != nil {
		manager.Client.API(fmt.Sprintf("callcenter_config agent set status %s 'Logged Out'", agent.AgentName))
	}

	ctx.Conn.Execute("playback", "ivr/ivr-agent_logged_out.wav", true)
}

// handleSpeedDial handles speed dial feature code (*0X)
func handleSpeedDial(ctx *ExecutionContext) {
	// The speed dial digit is captured from the regex (e.g., *01 → capture "1")
	digit := ctx.GetCapture("1")
	if digit == "" {
		ctx.Conn.Execute("playback", "ivr/ivr-invalid_selection.wav", true)
		return
	}

	// Speed dial destinations are stored in FeatureCode.ActionData for the
	// matching *0X code — the tenant admin sets the destination when customizing
	// the feature code after provisioning.
	dest := ctx.FeatureCode.ActionData
	if dest == "" {
		ctx.Conn.Execute("playback", "ivr/ivr-that_was_an_invalid_entry.wav", true)
		return
	}

	// Confirm and transfer
	ctx.Conn.Execute("playback", "tone_stream://%(100,0,600);%(100,0,800)", true)
	ctx.Conn.Execute("transfer", fmt.Sprintf("%s XML %s", dest, ctx.Domain), true)
}

// handleConference handles conference access via feature code (*85)
// Transfers the call to the conference module's ESL handler which handles
// PIN auth, max participants, moderator detection, etc.
func handleConference(ctx *ExecutionContext) {
	// The conference feature code might have a capture for a specific room
	roomNum := ctx.GetCapture("1")
	if roomNum == "" {
		// Default: use a conference name derived from the caller's extension
		roomNum = ctx.CallerID
	}

	// Transfer to the conference extension, which is routed via the dialplan
	// to the conference ESL module (127.0.0.4:9001) for full handling
	ctx.Conn.Execute("transfer", fmt.Sprintf("%s XML %s", roomNum, ctx.Domain), true)
}
