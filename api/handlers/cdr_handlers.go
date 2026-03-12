package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// =====================
// CDR / Call Records
// =====================

func (h *Handler) ListCDR(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset := (page - 1) * limit

	// Build query
	query := h.DB.Where("tenant_id = ?", tenantID)

	// Filters
	if ext := c.Query("extension"); ext != "" {
		query = query.Where("caller_id_number = ? OR destination_number = ?", ext, ext)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("start_stamp >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("start_stamp <= ?", endDate)
	}
	if direction := c.Query("direction"); direction != "" {
		query = query.Where("direction = ?", direction)
	}

	var total int64
	query.Model(&models.CallRecord{}).Count(&total)

	var records []models.CallRecord
	query.Order("start_stamp DESC").Offset(offset).Limit(limit).Find(&records)

	return c.JSON(fiber.Map{
		"data":  records,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) GetCDR(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var record models.CallRecord
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&record).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Call record not found"})
	}

	return c.JSON(record)
}

func (h *Handler) ExportCDR(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	query := h.DB.Where("tenant_id = ?", tenantID)

	// Apply same filters as ListCDR
	if ext := c.Query("extension"); ext != "" {
		query = query.Where("caller_id_number = ? OR destination_number = ?", ext, ext)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("start_stamp >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("start_stamp <= ?", endDate)
	}
	if direction := c.Query("direction"); direction != "" {
		query = query.Where("direction = ?", direction)
	}

	var records []models.CallRecord
	query.Order("start_stamp DESC").Limit(10000).Find(&records)

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=call-records.csv")

	// Write CSV header
	c.WriteString("Date,Direction,Caller ID,Destination,Duration (s),Billable (s),Hangup Cause,Recording\n")

	for _, r := range records {
		line := r.StartTime.Format("2006-01-02 15:04:05") + "," +
			string(r.Direction) + "," +
			r.CallerIDNumber + "," +
			r.DestinationNumber + "," +
			strconv.Itoa(r.Duration) + "," +
			strconv.Itoa(r.BillableSec) + "," +
			r.HangupCause + "," +
			r.RecordingPath + "\n"
		c.WriteString(line)
	}

	return nil
}

// =====================
// Audit Logs
// =====================

func (h *Handler) ListAuditLogs(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset := (page - 1) * limit

	query := h.DB.Model(&models.AuditLog{})

	// System admins with no tenant see all logs; tenant admins see only their tenant
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Filters
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if category := c.Query("category"); category != "" {
		// Map UI categories to action types
		switch category {
		case "security":
			query = query.Where("action IN ?", []string{"login", "logout"})
		case "configuration":
			query = query.Where("action IN ?", []string{"create", "update", "delete", "config", "import", "export"})
		case "user":
			query = query.Where("resource IN ?", []string{"user", "extension"})
		case "telephony":
			query = query.Where("resource IN ?", []string{"call", "gateway", "trunk", "route", "ivr", "queue"})
		}
	}
	if severity := c.Query("severity"); severity != "" {
		// Map severity to success/action
		switch severity {
		case "critical":
			query = query.Where("success = false")
		case "warning":
			query = query.Where("action IN ?", []string{"delete", "config"})
		case "info":
			query = query.Where("success = true")
		}
	}

	var total int64
	query.Count(&total)

	var logs []models.AuditLog
	query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs)

	return c.JSON(fiber.Map{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
