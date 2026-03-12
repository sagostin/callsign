package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// =====================
// Live Call Recording Control
// =====================

// StartCallRecording starts recording an active call via ESL
func (h *Handler) StartCallRecording(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var req struct {
		UUID string `json:"uuid"` // Channel UUID to record
	}
	if err := c.BodyParser(&req); err != nil || req.UUID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Channel UUID required"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH not connected"})
	}

	// Generate recording path
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("call_%s_%s.wav", req.UUID[:8], timestamp)
	recordPath := filepath.Join("/var/lib/callsign/recordings", fmt.Sprintf("%d", tenantID), filename)

	// Start recording via FreeSWITCH ESL
	cmd := fmt.Sprintf("uuid_record %s start %s", req.UUID, recordPath)
	_, err := h.ESLManager.Client.API(cmd)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start recording: " + err.Error()})
	}

	// Save recording record to DB
	recording := &models.CallRecording{
		TenantID:  tenantID,
		CallUUID:  req.UUID,
		FilePath:  recordPath,
		StartTime: time.Now(),
		Notes:     "live-recording-in-progress",
	}
	h.DB.Create(recording)

	return c.JSON(fiber.Map{
		"message":      "Recording started",
		"recording_id": recording.ID,
		"file_path":    recordPath,
	})
}

// StopCallRecording stops recording an active call via ESL
func (h *Handler) StopCallRecording(c *fiber.Ctx) error {
	var req struct {
		UUID string `json:"uuid"`
	}
	if err := c.BodyParser(&req); err != nil || req.UUID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Channel UUID required"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH not connected"})
	}

	// Stop recording via ESL
	cmd := fmt.Sprintf("uuid_record %s stop all", req.UUID)
	_, err := h.ESLManager.Client.API(cmd)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to stop recording: " + err.Error()})
	}

	// Update recording record in DB
	now := time.Now()
	h.DB.Model(&models.CallRecording{}).
		Where("call_uuid = ? AND notes = ?", req.UUID, "live-recording-in-progress").
		Updates(map[string]interface{}{
			"notes":    "completed",
			"end_time": now,
		})

	return c.JSON(fiber.Map{"message": "Recording stopped"})
}

// GetActiveCallsData returns live channel data from FreeSWITCH
func (h *Handler) GetActiveCallsData(c *fiber.Ctx) error {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.JSON(fiber.Map{"calls": []interface{}{}, "count": 0})
	}

	result, err := h.ESLManager.Client.API("show channels as json")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get active calls"})
	}

	return c.JSON(fiber.Map{"raw": result})
}

// GetLiveQueueStats returns real-time queue statistics
func (h *Handler) GetLiveQueueStats(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	// Get all queues for tenant
	var queues []models.Queue
	h.DB.Where("tenant_id = ?", tenantID).Find(&queues)

	type QueueStats struct {
		ID              uint   `json:"id"`
		Name            string `json:"name"`
		Extension       string `json:"extension"`
		WaitingCalls    int    `json:"waiting_calls"`
		AvailableAgents int    `json:"available_agents"`
		TotalAgents     int    `json:"total_agents"`
		ActiveCalls     int    `json:"active_calls"`
	}

	stats := make([]QueueStats, len(queues))
	for i, q := range queues {
		qs := QueueStats{
			ID:        q.ID,
			Name:      q.Name,
			Extension: q.Extension,
		}

		// Get live stats from FreeSWITCH if available
		if h.ESLManager != nil && h.ESLManager.IsConnected() {
			// Get agent count
			result, err := h.ESLManager.Client.API(fmt.Sprintf("callcenter_config tier list agents %s", q.Name))
			if err == nil && result != "" {
				lines := countOutputLines(result)
				qs.TotalAgents = lines
			}

			// Get waiting calls
			result, err = h.ESLManager.Client.API(fmt.Sprintf("callcenter_config queue list members %s", q.Name))
			if err == nil && result != "" {
				lines := countOutputLines(result)
				qs.WaitingCalls = lines
			}
		}

		stats[i] = qs
	}

	return c.JSON(stats)
}

// WakeupCallSchedule represents a wake-up call schedule request
type WakeupCallSchedule struct {
	RoomExtension string `json:"room_extension"`
	Time          string `json:"time"` // HH:MM format
	Date          string `json:"date"` // YYYY-MM-DD
	RecordingID   *uint  `json:"recording_id,omitempty"`
}

// ScheduleWakeupESL schedules a wake-up call via FreeSWITCH sched_api
func (h *Handler) ScheduleWakeupESL(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var req WakeupCallSchedule
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "FreeSWITCH not connected"})
	}

	// Parse the target time
	targetTime, err := time.Parse("2006-01-02 15:04", req.Date+" "+req.Time)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date/time format"})
	}

	// Calculate seconds from now
	delay := int(time.Until(targetTime).Seconds())
	if delay <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Scheduled time must be in the future"})
	}

	// Get a recording path for the wakeup message
	audioPath := "voicemail/vm-wakeup_call.wav" // Default
	if req.RecordingID != nil {
		var rec models.Recording
		if err := h.DB.Where("id = ? AND tenant_id = ?", *req.RecordingID, tenantID).First(&rec).Error; err == nil {
			audioPath = rec.FilePath
		}
	}

	// Schedule via FreeSWITCH sched_api
	// This creates a scheduled originate that calls the room extension and plays the recording
	groupCall := fmt.Sprintf("originate {ignore_early_media=true}user/%s &playback(%s)", req.RoomExtension, audioPath)
	schedID := fmt.Sprintf("wakeup_%d_%s", tenantID, req.RoomExtension)
	cmd := fmt.Sprintf("sched_api +%d %s %s", delay, schedID, groupCall)

	_, err = h.ESLManager.Client.API(cmd)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to schedule wake-up call: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":    "Wake-up call scheduled",
		"sched_id":   schedID,
		"target":     req.RoomExtension,
		"scheduled":  targetTime,
		"delay_secs": delay,
	})
}

// GetDeviceRegistrations returns SIP registration status from FreeSWITCH Sofia
func (h *Handler) GetDeviceRegistrations(c *fiber.Ctx) error {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		return c.JSON(fiber.Map{"registrations": []interface{}{}, "connected": false})
	}

	// Query Sofia for all registrations
	result, err := h.ESLManager.Client.API("sofia status profile internal reg")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to query registrations"})
	}

	return c.JSON(fiber.Map{
		"raw":       result,
		"connected": true,
	})
}

// countOutputLines counts non-empty lines (excluding header) from FreeSWITCH output
func countOutputLines(result string) int {
	count := 0
	for i, c := range result {
		if c == '\n' && i > 0 {
			count++
		}
	}
	if count > 0 {
		count-- // subtract header line
	}
	return count
}

// strToUint converts a string to uint safely
func strToUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}
