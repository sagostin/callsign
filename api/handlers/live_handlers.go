package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
)

// =====================
// Live Call Recording Control
// =====================

// StartCallRecording starts recording an active call via ESL
func (h *Handler) StartCallRecording(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var req struct {
		UUID string `json:"uuid"` // Channel UUID to record
	}
	if err := ctx.ReadJSON(&req); err != nil || req.UUID == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Channel UUID required"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH not connected"})
		return
	}

	// Generate recording path
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("call_%s_%s.wav", req.UUID[:8], timestamp)
	recordPath := filepath.Join("/var/lib/callsign/recordings", fmt.Sprintf("%d", tenantID), filename)

	// Start recording via FreeSWITCH ESL
	cmd := fmt.Sprintf("uuid_record %s start %s", req.UUID, recordPath)
	_, err := h.ESLManager.Client.API(cmd)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to start recording: " + err.Error()})
		return
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

	ctx.JSON(iris.Map{
		"message":      "Recording started",
		"recording_id": recording.ID,
		"file_path":    recordPath,
	})
}

// StopCallRecording stops recording an active call via ESL
func (h *Handler) StopCallRecording(ctx iris.Context) {
	var req struct {
		UUID string `json:"uuid"`
	}
	if err := ctx.ReadJSON(&req); err != nil || req.UUID == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Channel UUID required"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH not connected"})
		return
	}

	// Stop recording via ESL
	cmd := fmt.Sprintf("uuid_record %s stop all", req.UUID)
	_, err := h.ESLManager.Client.API(cmd)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to stop recording: " + err.Error()})
		return
	}

	// Update recording record in DB
	now := time.Now()
	h.DB.Model(&models.CallRecording{}).
		Where("call_uuid = ? AND notes = ?", req.UUID, "live-recording-in-progress").
		Updates(map[string]interface{}{
			"notes":    "completed",
			"end_time": now,
		})

	ctx.JSON(iris.Map{"message": "Recording stopped"})
}

// GetActiveCallsData returns live channel data from FreeSWITCH
func (h *Handler) GetActiveCallsData(ctx iris.Context) {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.JSON(iris.Map{"calls": []interface{}{}, "count": 0})
		return
	}

	result, err := h.ESLManager.Client.API("show channels as json")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to get active calls"})
		return
	}

	ctx.JSON(iris.Map{"raw": result})
}

// GetLiveQueueStats returns real-time queue statistics
func (h *Handler) GetLiveQueueStats(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

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

	ctx.JSON(stats)
}

// WakeupCallSchedule represents a wake-up call schedule request
type WakeupCallSchedule struct {
	RoomExtension string `json:"room_extension"`
	Time          string `json:"time"` // HH:MM format
	Date          string `json:"date"` // YYYY-MM-DD
	RecordingID   *uint  `json:"recording_id,omitempty"`
}

// ScheduleWakeupESL schedules a wake-up call via FreeSWITCH sched_api
func (h *Handler) ScheduleWakeupESL(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var req WakeupCallSchedule
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "FreeSWITCH not connected"})
		return
	}

	// Parse the target time
	targetTime, err := time.Parse("2006-01-02 15:04", req.Date+" "+req.Time)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid date/time format"})
		return
	}

	// Calculate seconds from now
	delay := int(time.Until(targetTime).Seconds())
	if delay <= 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Scheduled time must be in the future"})
		return
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
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to schedule wake-up call: " + err.Error()})
		return
	}

	ctx.JSON(iris.Map{
		"message":    "Wake-up call scheduled",
		"sched_id":   schedID,
		"target":     req.RoomExtension,
		"scheduled":  targetTime,
		"delay_secs": delay,
	})
}

// GetDeviceRegistrations returns SIP registration status from FreeSWITCH Sofia
func (h *Handler) GetDeviceRegistrations(ctx iris.Context) {
	if h.ESLManager == nil || !h.ESLManager.IsConnected() {
		ctx.JSON(iris.Map{"registrations": []interface{}{}, "connected": false})
		return
	}

	// Query Sofia for all registrations
	result, err := h.ESLManager.Client.API("sofia status profile internal reg")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to query registrations"})
		return
	}

	ctx.JSON(iris.Map{
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
