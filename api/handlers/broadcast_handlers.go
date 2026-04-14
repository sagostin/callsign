package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"
	"strings"

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
		h.logError("BROADCAST", "ListBroadcasts: Failed to retrieve campaigns", h.reqFields(c, nil))
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
		h.logWarn("BROADCAST", "CreateBroadcast: Invalid request payload", h.reqFields(c, nil))
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
		h.logError("BROADCAST", "CreateBroadcast: Failed to create campaign", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create campaign"})
	}

	return c.Status(http.StatusCreated).JSON(campaign)
}

// GetBroadcast returns a specific broadcast campaign
func (h *Handler) GetBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		h.logWarn("BROADCAST", "GetBroadcast: Invalid campaign ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		h.logWarn("BROADCAST", "GetBroadcast: Campaign not found", h.reqFields(c, nil))
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
		h.logWarn("BROADCAST", "UpdateBroadcast: Invalid campaign ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var existing models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&existing).Error; err != nil {
		h.logWarn("BROADCAST", "UpdateBroadcast: Campaign not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Campaign not found"})
	}

	// Don't allow editing a running campaign
	if existing.Status == models.BroadcastStatusRunning {
		h.logWarn("BROADCAST", "UpdateBroadcast: Cannot edit a running campaign. Pause it first.", h.reqFields(c, nil))
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Cannot edit a running campaign. Pause it first."})
	}

	var updates models.BroadcastCampaign
	if err := c.BodyParser(&updates); err != nil {
		h.logWarn("BROADCAST", "UpdateBroadcast: Invalid request payload", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	updates.ID = existing.ID
	updates.TenantID = tenantID
	if err := h.DB.Model(&existing).Updates(updates).Error; err != nil {
		h.logError("BROADCAST", "UpdateBroadcast: Failed to update campaign", h.reqFields(c, nil))
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
		h.logWarn("BROADCAST", "DeleteBroadcast: Invalid campaign ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	result := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.BroadcastCampaign{})
	if result.RowsAffected == 0 {
		h.logWarn("BROADCAST", "DeleteBroadcast: Campaign not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Campaign not found"})
	}

	return c.JSON(fiber.Map{"message": "Campaign deleted"})
}

// StartBroadcast starts (or resumes) a broadcast campaign
func (h *Handler) StartBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		h.logWarn("BROADCAST", "StartBroadcast: Invalid campaign ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		h.logWarn("BROADCAST", "StartBroadcast: Campaign not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Campaign not found"})
	}

	if campaign.Status == models.BroadcastStatusRunning {
		h.logWarn("BROADCAST", "StartBroadcast: Campaign is already running", h.reqFields(c, nil))
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "Campaign is already running"})
	}

	if len(campaign.Recipients) == 0 {
		h.logWarn("BROADCAST", "StartBroadcast: Campaign has no recipients", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Campaign has no recipients"})
	}

	// Check that broadcast worker is available
	if h.BroadcastWorker == nil {
		h.logError("BROADCAST", "StartBroadcast: Broadcast worker not configured", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Broadcast worker not configured"})
	}

	// Launch async broadcast worker - it handles status update internally
	if err := h.BroadcastWorker.StartCampaign(uint(id)); err != nil {
		h.logError("BROADCAST", "StartBroadcast: Failed to start campaign: "+err.Error(), h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start campaign"})
	}

	return c.JSON(fiber.Map{"message": "Campaign started"})
}

// StopBroadcast stops a running broadcast campaign
func (h *Handler) StopBroadcast(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		h.logWarn("BROADCAST", "StopBroadcast: Invalid campaign ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ? AND status = ?", id, tenantID, models.BroadcastStatusRunning).First(&campaign).Error; err != nil {
		h.logWarn("BROADCAST", "StopBroadcast: Running campaign not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Running campaign not found"})
	}

	// Check that broadcast worker is available
	if h.BroadcastWorker == nil {
		h.logError("BROADCAST", "StopBroadcast: Broadcast worker not configured", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Broadcast worker not configured"})
	}

	// Signal broadcast worker to stop
	if err := h.BroadcastWorker.StopCampaign(uint(id)); err != nil {
		h.logWarn("BROADCAST", "StopBroadcast: "+err.Error(), h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to stop campaign"})
	}

	return c.JSON(fiber.Map{"message": "Campaign stopped"})
}

// GetBroadcastStats returns statistics for a broadcast campaign
func (h *Handler) GetBroadcastStats(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		h.logWarn("BROADCAST", "GetBroadcastStats: Invalid campaign ID", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid campaign ID"})
	}

	var campaign models.BroadcastCampaign
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&campaign).Error; err != nil {
		h.logWarn("BROADCAST", "GetBroadcastStats: Campaign not found", h.reqFields(c, nil))
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
