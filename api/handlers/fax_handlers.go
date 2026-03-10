package handlers

import (
	"callsign/models"
	"callsign/services/fax"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
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
func (fh *FaxHandler) ListFaxBoxes(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	if tenantID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var boxes []models.FaxBox
	query := fh.Handler.DB.Where("tenant_id = ?", tenantID)

	// Optional search
	if search := ctx.URLParam("search"); search != "" {
		query = query.Where("name ILIKE ? OR did ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&boxes).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to list fax boxes"})
		return
	}

	ctx.JSON(boxes)
}

// CreateFaxBox creates a new fax box
func (fh *FaxHandler) CreateFaxBox(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	if tenantID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Tenant ID required"})
		return
	}

	var box models.FaxBox
	if err := ctx.ReadJSON(&box); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	box.TenantID = tenantID
	box.UUID = uuid.New()

	if err := fh.Handler.DB.Create(&box).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create fax box: " + err.Error()})
		return
	}

	// Reload fax data in memory
	fh.FaxManager.ReloadData()

	ctx.StatusCode(iris.StatusCreated)
	ctx.JSON(box)
}

// GetFaxBox returns a specific fax box
func (fh *FaxHandler) GetFaxBox(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	boxID, err := ctx.Params().GetUint("boxId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid box ID"})
		return
	}

	var box models.FaxBox
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", boxID, tenantID).
		Preload("Endpoints").First(&box).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Fax box not found"})
		return
	}

	ctx.JSON(box)
}

// UpdateFaxBox updates an existing fax box
func (fh *FaxHandler) UpdateFaxBox(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	boxID, err := ctx.Params().GetUint("boxId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid box ID"})
		return
	}

	var existing models.FaxBox
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", boxID, tenantID).First(&existing).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Fax box not found"})
		return
	}

	var updates models.FaxBox
	if err := ctx.ReadJSON(&updates); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	updates.ID = existing.ID
	updates.TenantID = tenantID
	updates.UUID = existing.UUID

	if err := fh.Handler.DB.Model(&existing).Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update fax box"})
		return
	}

	fh.FaxManager.ReloadData()
	ctx.JSON(existing)
}

// DeleteFaxBox deletes a fax box
func (fh *FaxHandler) DeleteFaxBox(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	boxID, err := ctx.Params().GetUint("boxId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid box ID"})
		return
	}

	result := fh.Handler.DB.Where("id = ? AND tenant_id = ?", boxID, tenantID).Delete(&models.FaxBox{})
	if result.Error != nil || result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Fax box not found"})
		return
	}

	fh.FaxManager.ReloadData()
	ctx.JSON(iris.Map{"message": "Fax box deleted"})
}

// --- Fax Endpoint CRUD ---

// ListFaxEndpoints returns fax endpoints for the current tenant
func (fh *FaxHandler) ListFaxEndpoints(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var endpoints []models.FaxEndpoint
	if err := fh.Handler.DB.Where("tenant_id = ? OR type = 'global'", tenantID).Find(&endpoints).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to list fax endpoints"})
		return
	}
	ctx.JSON(endpoints)
}

// CreateFaxEndpoint creates a new fax endpoint
func (fh *FaxHandler) CreateFaxEndpoint(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var ep models.FaxEndpoint
	if err := ctx.ReadJSON(&ep); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}
	ep.TenantID = &tenantID

	if err := fh.Handler.DB.Create(&ep).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create endpoint: " + err.Error()})
		return
	}

	fh.FaxManager.ReloadData()
	ctx.StatusCode(iris.StatusCreated)
	ctx.JSON(ep)
}

// UpdateFaxEndpoint updates an existing fax endpoint
func (fh *FaxHandler) UpdateFaxEndpoint(ctx iris.Context) {
	epID, err := ctx.Params().GetUint("epId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid endpoint ID"})
		return
	}

	var existing models.FaxEndpoint
	if err := fh.Handler.DB.First(&existing, epID).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Endpoint not found"})
		return
	}

	var updates models.FaxEndpoint
	if err := ctx.ReadJSON(&updates); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	if err := fh.Handler.DB.Model(&existing).Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update endpoint"})
		return
	}

	fh.FaxManager.ReloadData()
	ctx.JSON(existing)
}

// DeleteFaxEndpoint deletes a fax endpoint
func (fh *FaxHandler) DeleteFaxEndpoint(ctx iris.Context) {
	epID, err := ctx.Params().GetUint("epId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid endpoint ID"})
		return
	}

	result := fh.Handler.DB.Delete(&models.FaxEndpoint{}, epID)
	if result.Error != nil || result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Endpoint not found"})
		return
	}

	fh.FaxManager.ReloadData()
	ctx.JSON(iris.Map{"message": "Endpoint deleted"})
}

// --- Fax Jobs ---

// ListFaxJobs returns fax jobs for the current tenant
func (fh *FaxHandler) ListFaxJobs(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var jobs []models.FaxJob
	query := fh.Handler.DB.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	// Filters
	if direction := ctx.URLParam("direction"); direction != "" {
		query = query.Where("direction = ?", direction)
	}
	if status := ctx.URLParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if boxID := ctx.URLParam("box_id"); boxID != "" {
		query = query.Where("fax_box_id = ?", boxID)
	}

	// Pagination
	limit, _ := strconv.Atoi(ctx.URLParamDefault("limit", "50"))
	offset, _ := strconv.Atoi(ctx.URLParamDefault("offset", "0"))
	query = query.Limit(limit).Offset(offset)

	if err := query.Find(&jobs).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to list fax jobs"})
		return
	}

	ctx.JSON(jobs)
}

// GetFaxJob returns a specific fax job with page results
func (fh *FaxHandler) GetFaxJob(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	jobID, err := ctx.Params().GetUint("jobId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid job ID"})
		return
	}

	var job models.FaxJob
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", jobID, tenantID).First(&job).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Fax job not found"})
		return
	}

	// Load page results
	var pageResults []models.FaxPageResult
	fh.Handler.DB.Where("fax_job_id = ?", job.ID).Find(&pageResults)

	// Load notification logs
	var notifications []models.FaxNotificationLog
	fh.Handler.DB.Where("fax_job_id = ?", job.ID).Find(&notifications)

	ctx.JSON(iris.Map{
		"job":           job,
		"page_results":  pageResults,
		"notifications": notifications,
	})
}

// SendFax creates a new outbound fax job and enqueues it
func (fh *FaxHandler) SendFax(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var req struct {
		FaxBoxID     uint   `json:"fax_box_id"`
		Destination  string `json:"destination"`
		CallerIDName string `json:"caller_id_name,omitempty"`
		// File will be uploaded as multipart form
	}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Validate fax box belongs to tenant
	var box models.FaxBox
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", req.FaxBoxID, tenantID).First(&box).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Fax box not found"})
		return
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
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create fax job"})
		return
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

	ctx.StatusCode(iris.StatusAccepted)
	ctx.JSON(iris.Map{
		"message": "Fax queued for sending",
		"job_id":  job.ID,
		"uuid":    job.UUID.String(),
	})
}

// RetryFax requeues a failed fax job
func (fh *FaxHandler) RetryFax(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	jobID, err := ctx.Params().GetUint("jobId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid job ID"})
		return
	}

	var job models.FaxJob
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ? AND status = ?", jobID, tenantID, "failed").First(&job).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Failed fax job not found"})
		return
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

	ctx.JSON(iris.Map{"message": "Fax requeued", "job_id": job.ID})
}

// DeleteFaxJob deletes a fax job record
func (fh *FaxHandler) DeleteFaxJob(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	jobID, err := ctx.Params().GetUint("jobId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid job ID"})
		return
	}

	result := fh.Handler.DB.Where("id = ? AND tenant_id = ?", jobID, tenantID).Delete(&models.FaxJob{})
	if result.Error != nil || result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Fax job not found"})
		return
	}

	ctx.JSON(iris.Map{"message": "Fax job deleted"})
}

// DownloadFax returns the fax file for download
func (fh *FaxHandler) DownloadFax(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)
	jobID, err := ctx.Params().GetUint("jobId")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid job ID"})
		return
	}

	var job models.FaxJob
	if err := fh.Handler.DB.Where("id = ? AND tenant_id = ?", jobID, tenantID).First(&job).Error; err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Fax job not found"})
		return
	}

	// Prefer PDF if available, otherwise TIFF
	filePath := job.PDFFileName
	contentType := "application/pdf"
	if filePath == "" {
		filePath = job.FileName
		contentType = "image/tiff"
	}

	if filePath == "" {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "No fax file available"})
		return
	}

	ctx.ContentType(contentType)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=fax-%s.%s", job.UUID.String()[:8], ctx.URLParamDefault("format", "pdf")))
	ctx.ServeFile(filePath)
}

// GetActiveFaxes returns currently active/in-progress fax jobs
func (fh *FaxHandler) GetActiveFaxes(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var jobs []models.FaxJob
	fh.Handler.DB.Where("tenant_id = ? AND status IN ?", tenantID,
		[]string{"queued", "routing", "sending", "receiving", "bridging", "waiting"}).
		Order("created_at DESC").Find(&jobs)

	ctx.JSON(jobs)
}

// GetFaxStats returns fax statistics for the current tenant
func (fh *FaxHandler) GetFaxStats(ctx iris.Context) {
	tenantID := ctx.Values().GetUintDefault("tenant_id", 0)

	var stats models.FaxStats

	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'inbound'", tenantID).Count((*int64)(&stats.TotalInbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'outbound'", tenantID).Count((*int64)(&stats.TotalOutbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'inbound' AND success = true", tenantID).Count((*int64)(&stats.SuccessfulInbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'outbound' AND success = true", tenantID).Count((*int64)(&stats.SuccessfulOutbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'inbound' AND status = 'failed'", tenantID).Count((*int64)(&stats.FailedInbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND direction = 'outbound' AND status = 'failed'", tenantID).Count((*int64)(&stats.FailedOutbound))
	fh.Handler.DB.Model(&models.FaxJob{}).Where("tenant_id = ? AND status IN ?", tenantID, []string{"queued", "routing", "sending", "receiving"}).Count((*int64)(&stats.PendingJobs))

	ctx.JSON(stats)
}
