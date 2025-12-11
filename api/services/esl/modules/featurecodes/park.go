package featurecodes

import (
	"callsign/models"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// handlePark handles valet parking (auto slot assignment)
func handlePark(ctx *ExecutionContext) {
	fc := ctx.FeatureCode
	lotName := fc.ParkLotName
	if lotName == "" {
		lotName = "default"
	}

	// Find available slot
	slot, err := models.GetAvailableSlot(ctx.DB, ctx.TenantID, lotName)
	if err != nil {
		log.Warnf("No parking slots available")
		ctx.Conn.Execute("playback", "ivr/ivr-no_parking_slots_available.wav", true)
		return
	}

	// Park the current call
	ctx.Conn.Execute("valet_park", fmt.Sprintf("%s %d in %d %d %s",
		ctx.Domain,
		slot.SlotNumber,
		fc.ParkTimeout,
		fc.ParkTimeout,
		ctx.Domain,
	), true)

	// Update slot status
	slot.ParkCall(ctx.DB, ctx.UUID, ctx.CallerID, ctx.CallerName, ctx.CallerID)

	// Announce slot number
	if fc.ParkAnnounce {
		ctx.Conn.Execute("playback", "ivr/ivr-call_parked_at.wav", true)
		ctx.Conn.Execute("say", fmt.Sprintf("en number iterated %d", slot.SlotNumber), true)
	}

	// Send BLF update
	ctx.Service.sendPresenceNotify(ctx.Domain, fmt.Sprintf("park+*57%02d", slot.SlotNumber), "confirmed")

	log.WithFields(log.Fields{
		"slot":   slot.SlotNumber,
		"caller": ctx.CallerID,
	}).Info("Call parked")
}

// handleParkSlot parks to a specific slot (e.g., *7001 for slot 01)
func handleParkSlot(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	// Get slot number from regex capture or action_data
	slotStr := ctx.GetCapture("slot")
	if slotStr == "" {
		slotStr = ctx.GetCapture("1") // Numbered capture
	}
	if slotStr == "" {
		slotStr = fc.ActionData
	}

	if slotStr == "" {
		// Prompt for slot
		ctx.Conn.Execute("playback", "ivr/ivr-enter_slot_number.wav", true)
		ev, _ := ctx.Conn.Execute("read", "1 4 tone_stream://%(250,50,440) slot 5000 #", true)
		slotStr = ev.Get("variable_slot")
	}

	slotNum, err := strconv.Atoi(slotStr)
	if err != nil {
		ctx.Conn.Execute("playback", "ivr/ivr-invalid_slot.wav", true)
		return
	}

	lotName := fc.ParkLotName
	if lotName == "" {
		lotName = "default"
	}

	// Get or create slot
	slot, err := models.GetSlotByNumber(ctx.DB, ctx.TenantID, slotNum, lotName)
	if err != nil {
		// Create slot if doesn't exist
		slot = &models.ParkSlot{
			TenantID:   ctx.TenantID,
			Domain:     ctx.Domain,
			SlotNumber: slotNum,
			LotName:    lotName,
		}
		ctx.DB.Create(slot)
	}

	if slot.IsOccupied {
		ctx.Conn.Execute("playback", "ivr/ivr-slot_occupied.wav", true)
		return
	}

	// Park the call
	ctx.Conn.Execute("valet_park", fmt.Sprintf("%s %d in %d %d %s",
		ctx.Domain,
		slotNum,
		fc.ParkTimeout,
		fc.ParkTimeout,
		ctx.Domain,
	), true)

	// Update slot
	slot.ParkCall(ctx.DB, ctx.UUID, ctx.CallerID, ctx.CallerName, ctx.CallerID)

	// Announce
	ctx.Conn.Execute("playback", "ivr/ivr-call_parked_at.wav", true)
	ctx.Conn.Execute("say", fmt.Sprintf("en number iterated %d", slotNum), true)

	// BLF update
	ctx.Service.sendPresenceNotify(ctx.Domain, fmt.Sprintf("park+*57%02d", slotNum), "confirmed")

	log.WithFields(log.Fields{
		"slot":   slotNum,
		"caller": ctx.CallerID,
	}).Info("Call parked to specific slot")
}

// handleParkRetrieve retrieves from a parking slot
func handleParkRetrieve(ctx *ExecutionContext) {
	fc := ctx.FeatureCode

	// Get slot number from regex capture
	slotStr := ctx.GetCapture("slot")
	if slotStr == "" {
		slotStr = ctx.GetCapture("1")
	}
	if slotStr == "" {
		slotStr = fc.ActionData
	}

	if slotStr == "" {
		// Prompt for slot
		ctx.Conn.Execute("playback", "ivr/ivr-enter_slot_number.wav", true)
		ev, _ := ctx.Conn.Execute("read", "1 4 tone_stream://%(250,50,440) slot 5000 #", true)
		slotStr = ev.Get("variable_slot")
	}

	slotNum, err := strconv.Atoi(slotStr)
	if err != nil {
		ctx.Conn.Execute("playback", "ivr/ivr-invalid_slot.wav", true)
		return
	}

	lotName := fc.ParkLotName
	if lotName == "" {
		lotName = "default"
	}

	// Get slot
	slot, err := models.GetSlotByNumber(ctx.DB, ctx.TenantID, slotNum, lotName)
	if err != nil || !slot.IsOccupied {
		ctx.Conn.Execute("playback", "ivr/ivr-slot_not_found.wav", true)
		return
	}

	// Retrieve the call
	ctx.Conn.Execute("valet_park", fmt.Sprintf("%s %d out", ctx.Domain, slotNum), true)

	// Clear slot
	slot.RetrieveCall(ctx.DB)

	// BLF update
	ctx.Service.sendPresenceNotify(ctx.Domain, fmt.Sprintf("park+*57%02d", slotNum), "terminated")

	log.WithFields(log.Fields{
		"slot":   slotNum,
		"caller": ctx.CallerID,
	}).Info("Retrieved call from parking slot")
}
