package handlers

import (
	"callsign/models"
	"callsign/services/esl/modules/conference"
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// ConferenceHandler handles conference control API requests
type ConferenceHandler struct {
	DB      *gorm.DB
	Service *conference.Service
}

// NewConferenceHandler creates a new conference handler
func NewConferenceHandler(db *gorm.DB, svc *conference.Service) *ConferenceHandler {
	return &ConferenceHandler{
		DB:      db,
		Service: svc,
	}
}

// ListConferences returns all conference rooms for a tenant
func (h *ConferenceHandler) ListConferences(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var conferences []models.Conference
	h.DB.Where("tenant_id = ? AND enabled = ?", tenantID, true).Find(&conferences)

	ctx.JSON(conferences)
}

// GetConference returns a specific conference
func (h *ConferenceHandler) GetConference(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var conf models.Conference
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Members").First(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conference not found"})
		return
	}

	ctx.JSON(conf)
}

// CreateConference creates a new conference room
func (h *ConferenceHandler) CreateConference(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var conf models.Conference
	if err := ctx.ReadJSON(&conf); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	conf.TenantID = tenantID
	if err := h.DB.Create(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(conf)
}

// UpdateConference updates a conference room
func (h *ConferenceHandler) UpdateConference(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	var conf models.Conference
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&conf).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conference not found"})
		return
	}

	if err := ctx.ReadJSON(&conf); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	conf.TenantID = tenantID
	h.DB.Save(&conf)
	ctx.JSON(conf)
}

// DeleteConference deletes a conference room
func (h *ConferenceHandler) DeleteConference(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	id, _ := ctx.Params().GetUint("id")

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Conference{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Conference not found"})
		return
	}

	ctx.StatusCode(http.StatusNoContent)
}

// ========== Live Conference Control ==========

// ListLiveConferences returns active conferences from FreeSWITCH
func (h *ConferenceHandler) ListLiveConferences(ctx iris.Context) {
	if h.Service == nil {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "Conference service not available"})
		return
	}

	conferences, err := h.Service.ListLive()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(conferences)
}

// GetLiveConference returns live conference with members
func (h *ConferenceHandler) GetLiveConference(ctx iris.Context) {
	confName := ctx.Params().Get("name")

	if h.Service == nil {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "Conference service not available"})
		return
	}

	info, err := h.Service.GetLiveConference(confName)
	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(info)
}

// MuteMember mutes a conference member
func (h *ConferenceHandler) MuteMember(ctx iris.Context) {
	confName := ctx.Params().Get("name")
	memberID, _ := strconv.Atoi(ctx.Params().Get("member"))

	if err := h.Service.MuteMember(confName, memberID); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "mute", "member": memberID})
}

// UnmuteMember unmutes a conference member
func (h *ConferenceHandler) UnmuteMember(ctx iris.Context) {
	confName := ctx.Params().Get("name")
	memberID, _ := strconv.Atoi(ctx.Params().Get("member"))

	if err := h.Service.UnmuteMember(confName, memberID); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "unmute", "member": memberID})
}

// DeafMember makes a member deaf
func (h *ConferenceHandler) DeafMember(ctx iris.Context) {
	confName := ctx.Params().Get("name")
	memberID, _ := strconv.Atoi(ctx.Params().Get("member"))

	if err := h.Service.DeafMember(confName, memberID); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "deaf", "member": memberID})
}

// UndeafMember removes deaf from member
func (h *ConferenceHandler) UndeafMember(ctx iris.Context) {
	confName := ctx.Params().Get("name")
	memberID, _ := strconv.Atoi(ctx.Params().Get("member"))

	if err := h.Service.UndeafMember(confName, memberID); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "undeaf", "member": memberID})
}

// KickMember kicks a member from conference
func (h *ConferenceHandler) KickMember(ctx iris.Context) {
	confName := ctx.Params().Get("name")
	memberID, _ := strconv.Atoi(ctx.Params().Get("member"))

	if err := h.Service.KickMember(confName, memberID); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "kick", "member": memberID})
}

// LockConference locks the conference
func (h *ConferenceHandler) LockConference(ctx iris.Context) {
	confName := ctx.Params().Get("name")

	if err := h.Service.LockConference(confName); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "lock", "conference": confName})
}

// UnlockConference unlocks the conference
func (h *ConferenceHandler) UnlockConference(ctx iris.Context) {
	confName := ctx.Params().Get("name")

	if err := h.Service.UnlockConference(confName); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "unlock", "conference": confName})
}

// StartRecording starts recording the conference
func (h *ConferenceHandler) StartRecording(ctx iris.Context) {
	confName := ctx.Params().Get("name")

	var body struct {
		Path string `json:"path"`
	}
	ctx.ReadJSON(&body)

	if body.Path == "" {
		body.Path = "/var/lib/freeswitch/recordings/" + confName + "_" + time.Now().Format("20060102_150405") + ".wav"
	}

	if err := h.Service.StartRecording(confName, body.Path); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "recording_start", "path": body.Path})
}

// StopRecording stops recording
func (h *ConferenceHandler) StopRecording(ctx iris.Context) {
	confName := ctx.Params().Get("name")

	if err := h.Service.StopRecording(confName); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "recording_stop"})
}

// MuteAll mutes all participants
func (h *ConferenceHandler) MuteAll(ctx iris.Context) {
	confName := ctx.Params().Get("name")

	if err := h.Service.MuteAll(confName); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "mute_all"})
}

// UnmuteAll unmutes all participants
func (h *ConferenceHandler) UnmuteAll(ctx iris.Context) {
	confName := ctx.Params().Get("name")

	if err := h.Service.UnmuteAll(confName); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "unmute_all"})
}

// SetFloor gives floor to a member
func (h *ConferenceHandler) SetFloor(ctx iris.Context) {
	confName := ctx.Params().Get("name")
	memberID, _ := strconv.Atoi(ctx.Params().Get("member"))

	if err := h.Service.SetFloor(confName, memberID); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(iris.Map{"status": "ok", "action": "floor", "member": memberID})
}

// ========== Conference Stats ==========

// GetConferenceStats returns conference statistics
func (h *ConferenceHandler) GetConferenceStats(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	// Parse date range
	startStr := ctx.URLParam("start")
	endStr := ctx.URLParam("end")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			startDate = time.Now().AddDate(0, 0, -30) // Default: last 30 days
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			endDate = time.Now()
		}
	} else {
		endDate = time.Now()
	}

	stats, err := models.GetConferenceStats(h.DB, tenantID, startDate, endDate)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctx.JSON(stats)
}

// GetConferenceSessions returns conference session history
func (h *ConferenceHandler) GetConferenceSessions(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	confID, _ := ctx.Params().GetUint("id")

	var sessions []models.ConferenceSession
	query := h.DB.Where("tenant_id = ?", tenantID)

	if confID > 0 {
		query = query.Where("conference_id = ?", confID)
	}

	query.Order("start_time DESC").Limit(50).Find(&sessions)
	ctx.JSON(sessions)
}

// GetSessionParticipants returns participants for a session
func (h *ConferenceHandler) GetSessionParticipants(ctx iris.Context) {
	sessionID, _ := ctx.Params().GetUint("session_id")

	var participants []models.ConferenceParticipant
	h.DB.Where("session_id = ?", sessionID).Order("join_time ASC").Find(&participants)

	ctx.JSON(participants)
}
