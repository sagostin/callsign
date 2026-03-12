package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

// =====================
// Call Broadcast Campaigns
// =====================

// ListBroadcasts returns all broadcast campaigns for the current tenant
func (h *Handler) ListBroadcasts(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var campaigns []models.BroadcastCampaign
	query := h.DB.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&campaigns).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve campaigns"})
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

	return c.JSON(result)
}

// CreateBroadcast creates a new broadcast campaign
func (h *Handler) CreateBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var campaign models.BroadcastCampaign
	if err := c.BodyParser(&campaign); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
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
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create campaign"})
	}

	return c.Status(http.StatusCreated).JSON(campaign)
}

// GetBroadcast returns a specific broadcast campaign
func (h *Handler) GetBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Campaign not found"})
	}

	return c.JSON(fiber.Map{
		"campaign":     campaign,
		"progress":     campaign.Progress(),
		"target_count": len(campaign.Recipients),
	})
}

// UpdateBroadcast updates an existing broadcast campaign
func (h *Handler) UpdateBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var existing models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&existing).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Campaign not found"})
	}

	// Don't allow editing a running campaign
	if existing.Status == models.BroadcastStatusRunning {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Cannot edit a running campaign. Pause it first."})
	}

	var updates models.BroadcastCampaign
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	updates.ID = existing.ID
	updates.TenantID = tenantID
	if err := h.DB.Model(&existing).Updates(updates).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update campaign"})
	}

	h.DB.First(&existing, id)
	return c.JSON(existing)
}

// DeleteBroadcast deletes a broadcast campaign
func (h *Handler) DeleteBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.BroadcastCampaign{})
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Campaign not found"})
	}

	return c.JSON(fiber.Map{"message": "Campaign deleted"})
}

// StartBroadcast starts (or resumes) a broadcast campaign
func (h *Handler) StartBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Campaign not found"})
	}

	if campaign.Status == models.BroadcastStatusRunning {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Campaign is already running"})
	}

	if len(campaign.Recipients) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Campaign has no recipients"})
	}

	now := time.Now()
	h.DB.Model(&campaign).Updates(map[string]interface{}{
		"status":     models.BroadcastStatusRunning,
		"started_at": now,
	})

	// TODO: Launch async broadcast worker via ESL
	// The worker would iterate recipients, originate calls via FreeSWITCH,
	// and play the recording, updating stats as calls complete.

	return c.JSON(fiber.Map{"message": "Campaign started", "started_at": now})
}

// StopBroadcast stops a running broadcast campaign
func (h *Handler) StopBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ? AND status = ?", id, tenantID, models.BroadcastStatusRunning).First(&campaign).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Running campaign not found"})
	}

	h.DB.Model(&campaign).Update("status", models.BroadcastStatusPaused)

	// TODO: Signal broadcast worker to stop originating new calls

	return c.JSON(fiber.Map{"message": "Campaign stopped"})
}

// GetBroadcastStats returns statistics for a broadcast campaign
func (h *Handler) GetBroadcastStats(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Campaign not found"})
	}

	return c.JSON(fiber.Map{
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
