package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/lib/pq"
)

// =====================
// Call Broadcast Campaigns
// =====================

// ListBroadcasts returns all broadcast campaigns for the current tenant
func (h *Handler) ListBroadcasts(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var campaigns []models.BroadcastCampaign
	query := h.DB.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	// Filter by status if provided
	if status := ctx.URLParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&campaigns).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve campaigns"})
		return
	}

	// Enrich with progress percentage
	type CampaignResponse struct {
		models.BroadcastCampaign
		Progress    int `json:"progress"`
		TargetCount int `json:"target_count"`
	}

	result := make([]CampaignResponse, len(campaigns))
	for i, c := range campaigns {
		result[i] = CampaignResponse{
			BroadcastCampaign: c,
			Progress:          c.Progress(),
			TargetCount:       len(c.Recipients),
		}
	}

	ctx.JSON(result)
}

// CreateBroadcast creates a new broadcast campaign
func (h *Handler) CreateBroadcast(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	var campaign models.BroadcastCampaign
	if err := ctx.ReadJSON(&campaign); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	campaign.TenantID = tenantID
	if campaign.Status == "" {
		campaign.Status = models.BroadcastStatusDraft
	}

	// Parse recipients from newline-separated string if sent as a string field
	if len(campaign.Recipients) == 1 && strings.Contains(campaign.Recipients[0], "\n") {
		lines := strings.Split(campaign.Recipients[0], "\n")
		cleaned := make(pq.StringArray, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				cleaned = append(cleaned, line)
			}
		}
		campaign.Recipients = cleaned
	}

	if err := h.DB.Create(&campaign).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create campaign"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(campaign)
}

// GetBroadcast returns a specific broadcast campaign
func (h *Handler) GetBroadcast(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid campaign ID"})
		return
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Campaign not found"})
		return
	}

	ctx.JSON(iris.Map{
		"campaign":     campaign,
		"progress":     campaign.Progress(),
		"target_count": len(campaign.Recipients),
	})
}

// UpdateBroadcast updates an existing broadcast campaign
func (h *Handler) UpdateBroadcast(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid campaign ID"})
		return
	}

	var existing models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&existing).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Campaign not found"})
		return
	}

	// Don't allow editing a running campaign
	if existing.Status == models.BroadcastStatusRunning {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": "Cannot edit a running campaign. Pause it first."})
		return
	}

	var updates models.BroadcastCampaign
	if err := ctx.ReadJSON(&updates); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	updates.ID = existing.ID
	updates.TenantID = tenantID
	if err := h.DB.Model(&existing).Updates(updates).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update campaign"})
		return
	}

	h.DB.First(&existing, id)
	ctx.JSON(existing)
}

// DeleteBroadcast deletes a broadcast campaign
func (h *Handler) DeleteBroadcast(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid campaign ID"})
		return
	}

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.BroadcastCampaign{})
	if result.RowsAffected == 0 {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Campaign not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Campaign deleted"})
}

// StartBroadcast starts (or resumes) a broadcast campaign
func (h *Handler) StartBroadcast(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid campaign ID"})
		return
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Campaign not found"})
		return
	}

	if campaign.Status == models.BroadcastStatusRunning {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": "Campaign is already running"})
		return
	}

	if len(campaign.Recipients) == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Campaign has no recipients"})
		return
	}

	now := time.Now()
	h.DB.Model(&campaign).Updates(map[string]interface{}{
		"status":     models.BroadcastStatusRunning,
		"started_at": now,
	})

	// TODO: Launch async broadcast worker via ESL
	// The worker would iterate recipients, originate calls via FreeSWITCH,
	// and play the recording, updating stats as calls complete.

	ctx.JSON(iris.Map{"message": "Campaign started", "started_at": now})
}

// StopBroadcast stops a running broadcast campaign
func (h *Handler) StopBroadcast(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid campaign ID"})
		return
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ? AND status = ?", id, tenantID, models.BroadcastStatusRunning).First(&campaign).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Running campaign not found"})
		return
	}

	h.DB.Model(&campaign).Update("status", models.BroadcastStatusPaused)

	// TODO: Signal broadcast worker to stop originating new calls

	ctx.JSON(iris.Map{"message": "Campaign stopped"})
}

// GetBroadcastStats returns statistics for a broadcast campaign
func (h *Handler) GetBroadcastStats(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid campaign ID"})
		return
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Campaign not found"})
		return
	}

	ctx.JSON(iris.Map{
		"total_recipients": len(campaign.Recipients),
		"total_calls":      campaign.TotalCalls,
		"answered_calls":   campaign.AnsweredCalls,
		"failed_calls":     campaign.FailedCalls,
		"busy_calls":       campaign.BusyCalls,
		"no_answer_calls":  campaign.NoAnswerCalls,
		"progress":         campaign.Progress(),
		"status":           campaign.Status,
		"started_at":       campaign.StartedAt,
		"completed_at":     campaign.CompletedAt,
	})
}
