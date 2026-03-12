package handlers

import (
	"callsign/models"
	"callsign/services/fax"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// FaxHandler handles fax-related API requests
type FaxHandler struct {
	Handler    *Handler
	FaxManager *fax.Manager
}

// NewFaxHandler creates a new FaxHandler
func NewFaxHandler(h *Handler, fm *fax.Manager) *FaxHandler {
	return &FaxHandler{Handler: h, FaxManager: fm}
}

// --- Fax Box CRUD ---

// ListFaxBoxes returns all fax boxes for the current tenant
func (fh *FaxHandler) ListFaxBoxes(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	if tenantID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var boxes []models.FaxBox
	query := fh.Handler.DB.Where("tenant_id = ?", tenantID)

	// Optional search
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR did ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&boxes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list fax boxes"})
	}

	return c.JSON(boxes)
}

// CreateFaxBox creates a new fax box
func (fh *FaxHandler) CreateFaxBox(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	if tenantID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tenant ID required"})
	}

	var box models.FaxBox
	if err := c.BodyParser(&box); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	box.TenantID = tenantID
	box.UUID = uuid.New()

	if err := fh.Handler.DB.Create(&box).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create fax box: " + err.Error()})
	}

	// Reload fax data in memory
	fh.FaxManager.ReloadData()

	return c.Status(fiber.StatusCreated).JSON(box)
}

// GetFaxBox returns a specific fax box
func (fh *FaxHandler) GetFaxBox(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	boxID, err := strconv.ParseUint(c.Params("boxId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid box ID"})
	}

	var box models.FaxBox
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", boxID, tenantID).
		Preload("Endpoints").First(&box).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Fax box not found"})
	}

	return c.JSON(box)
}

// UpdateFaxBox updates an existing fax box
func (fh *FaxHandler) UpdateFaxBox(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	boxID, err := strconv.ParseUint(c.Params("boxId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid box ID"})
	}

	var existing models.FaxBox
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", boxID, tenantID).First(&existing).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Fax box not found"})
	}

	var updates models.FaxBox
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	updates.ID = existing.ID
	updates.TenantID = tenantID
	updates.UUID = existing.UUID

	if err := fh.Handler.DB.Model(&existing).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update fax box"})
	}

	fh.FaxManager.ReloadData()
	return c.JSON(existing)
}

// DeleteFaxBox deletes a fax box
func (fh *FaxHandler) DeleteFaxBox(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	boxID, err := strconv.ParseUint(c.Params("boxId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid box ID"})
	}

	result := fh.Handler.DB.Where("id = ? AND tenant_id = ?", boxID, tenantID).Delete(&models.FaxBox{})
	if result.Error != nil || result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Fax box not found"})
	}

	fh.FaxManager.ReloadData()
	return c.JSON(fiber.Map{"message": "Fax box deleted"})
}

// --- Fax Endpoint CRUD ---

// ListFaxEndpoints returns fax endpoints for the current tenant
func (fh *FaxHandler) ListFaxEndpoints(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var endpoints []models.FaxEndpoint
	if err := fh.Handler.DB.Where("tenant_id = ? OR type = 'global'", tenantID).Find(&endpoints).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list fax endpoints"})
	}
	return c.JSON(endpoints)
}

// CreateFaxEndpoint creates a new fax endpoint
func (fh *FaxHandler) CreateFaxEndpoint(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var ep models.FaxEndpoint
	if err := c.BodyParser(&ep); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	ep.TenantID = &tenantID

	if err := fh.Handler.DB.Create(&ep).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create endpoint: " + err.Error()})
	}

	fh.FaxManager.ReloadData()
	return c.Status(fiber.StatusCreated).JSON(ep)
}

// UpdateFaxEndpoint updates an existing fax endpoint
func (fh *FaxHandler) UpdateFaxEndpoint(c *fiber.Ctx) error {
	epID, err := strconv.ParseUint(c.Params("epId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid endpoint ID"})
	}

	var existing models.FaxEndpoint
	if err := fh.Handler.DB.First(&existing, epID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Endpoint not found"})
	}

	var updates models.FaxEndpoint
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := fh.Handler.DB.Model(&existing).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update endpoint"})
	}

	fh.FaxManager.ReloadData()
	return c.JSON(existing)
}

// DeleteFaxEndpoint deletes a fax endpoint
func (fh *FaxHandler) DeleteFaxEndpoint(c *fiber.Ctx) error {
	epID, err := strconv.ParseUint(c.Params("epId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid endpoint ID"})
	}

	result := fh.Handler.DB.Delete(&models.FaxEndpoint{}, epID)
	if result.Error != nil || result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Endpoint not found"})
	}

	fh.FaxManager.ReloadData()
	return c.JSON(fiber.Map{"message": "Endpoint deleted"})
}

// --- Fax Jobs ---

// ListFaxJobs returns fax jobs for the current tenant
func (fh *FaxHandler) ListFaxJobs(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var jobs []models.FaxJob
	query := fh.Handler.DB.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	// Filters
	if direction := c.Query("direction"); direction != "" {
		query = query.Where("direction = ?", direction)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if boxID := c.Query("box_id"); boxID != "" {
		query = query.Where("fax_box_id = ?", boxID)
	}

	// Pagination
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	query = query.Limit(limit).Offset(offset)

	if err := query.Find(&jobs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list fax jobs"})
	}

	return c.JSON(jobs)
}

// GetFaxJob returns a specific fax job with page results
func (fh *FaxHandler) GetFaxJob(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	jobID, err := strconv.ParseUint(c.Params("jobId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid job ID"})
	}

	var job models.FaxJob
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", jobID, tenantID).First(&job).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Fax job not found"})
	}

	// Load page results
	var pageResults []models.FaxPageResult
	fh.Handler.DB.Where("fax_job_id = ?", job.ID).Find(&pageResults)

	// Load notification logs
	var notifications []models.FaxNotificationLog
	fh.Handler.DB.Where("fax_job_id = ?", job.ID).Find(&notifications)

	return c.JSON(fiber.Map{
		"job":           job,
		"page_results":  pageResults,
		"notifications": notifications,
	})
}

// SendFax creates a new outbound fax job and enqueues it
func (fh *FaxHandler) SendFax(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var req struct {
		FaxBoxID     uint   `json:"fax_box_id"`
		Destination  string `json:"destination"`
		CallerIDName string `json:"caller_id_name,omitempty"`
		// File will be uploaded as multipart form
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate fax box belongs to tenant
	var box models.FaxBox
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", req.FaxBoxID, tenantID).First(&box).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Fax box not found"})
	}

	// Create fax job in DB
	now := time.Now()
	job := models.FaxJob{
		UUID:         uuid.New(),
		TenantID:     tenantID,
		FaxBoxID:     &box.ID,
		Direction:    "outbound",
		CallerNumber: box.DID,
		CalleeNumber: req.Destination,
		CallerIDName: req.CallerIDName,
		Header:       box.Header,
		Status:       "queued",
		MaxRetries:   3,
		StartedAt:    &now,
		SourceType:   "api",
	}
	if job.CallerIDName == "" {
		job.CallerIDName = box.CallerIDName
	}

	if err := fh.Handler.DB.Create(&job).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create fax job"})
	}

	// Enqueue for routing/sending
	internalJob := &fax.FaxJobInternal{
		UUID:           job.UUID,
		CallerIdNumber: job.CallerNumber,
		CalleeNumber:   job.CalleeNumber,
		CallerIdName:   job.CallerIDName,
		Header:         job.Header,
		Status:         "queued",
		SourceType:     "api",
		DBJobID:        job.ID,
	}
	go func() {
		fh.FaxManager.JobRouting <- internalJob
	}()

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Fax queued for sending",
		"job_id":  job.ID,
		"uuid":    job.UUID.String(),
	})
}

// RetryFax requeues a failed fax job
func (fh *FaxHandler) RetryFax(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	jobID, err := strconv.ParseUint(c.Params("jobId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid job ID"})
	}

	var job models.FaxJob
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ? AND status = ?", jobID, tenantID, "failed").First(&job).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Failed fax job not found"})
	}

	// Reset status
	job.Status = "queued"
	job.Attempts = 0
	job.LastError = ""
	fh.Handler.DB.Save(&job)

	// Re-enqueue
	internalJob := &fax.FaxJobInternal{
		UUID:           job.UUID,
		CallerIdNumber: job.CallerNumber,
		CalleeNumber:   job.CalleeNumber,
		CallerIdName:   job.CallerIDName,
		Header:         job.Header,
		Status:         "queued",
		SourceType:     job.SourceType,
		DBJobID:        job.ID,
	}
	go func() {
		fh.FaxManager.JobRouting <- internalJob
	}()

	return c.JSON(fiber.Map{"message": "Fax requeued", "job_id": job.ID})
}

// DeleteFaxJob deletes a fax job record
func (fh *FaxHandler) DeleteFaxJob(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	jobID, err := strconv.ParseUint(c.Params("jobId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid job ID"})
	}

	result := fh.Handler.DB.Where("id = ? AND tenant_id = ?", jobID, tenantID).Delete(&models.FaxJob{})
	if result.Error != nil || result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Fax job not found"})
	}

	return c.JSON(fiber.Map{"message": "Fax job deleted"})
}

// DownloadFax returns the fax file for download
func (fh *FaxHandler) DownloadFax(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)
	jobID, err := strconv.ParseUint(c.Params("jobId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid job ID"})
	}

	var job models.FaxJob
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", jobID, tenantID).First(&job).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Fax job not found"})
	}

	// Prefer PDF if available, otherwise TIFF
	filePath := job.PDFFileName
	contentType := "application/pdf"
	if filePath == "" {
		filePath = job.FileName
		contentType = "image/tiff"
	}

	if filePath == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No fax file available"})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=fax-%s.%s", job.UUID.String()[:8], c.Query("format", "pdf")))
	return c.SendFile(filePath)
}

// GetActiveFaxes returns currently active/in-progress fax jobs
func (fh *FaxHandler) GetActiveFaxes(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var jobs []models.FaxJob
	fh.Handler.DB.Where("tenant_id = ? AND status IN ?", tenantID,
		[]string{"queued", "routing", "sending", "receiving", "bridging", "waiting"}).
		Order("created_at DESC").Find(&jobs)

	return c.JSON(jobs)
}

// GetFaxStats returns fax statistics for the current tenant
func (fh *FaxHandler) GetFaxStats(c *fiber.Ctx) error {
	tenantID := getLocalsUint(c, "tenant_id", 0)

	var stats models.FaxStats

	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'inbound'", tenantID).Count((*int64)(&stats.TotalInbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'outbound'", tenantID).Count((*int64)(&stats.TotalOutbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'inbound' AND success = true", tenantID).Count((*int64)(&stats.SuccessfulInbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'outbound' AND success = true", tenantID).Count((*int64)(&stats.SuccessfulOutbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'inbound' AND status = 'failed'", tenantID).Count((*int64)(&stats.FailedInbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'outbound' AND status = 'failed'", tenantID).Count((*int64)(&stats.FailedOutbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND status IN ?", tenantID, []string{"queued", "routing", "sending", "receiving"}).Count((*int64)(&stats.PendingJobs))

	return c.JSON(stats)
}
