package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"
	"strconv"

	"github.com/kataras/iris/v12"
)

// =====================
// CDR / Call Records
// =====================

func (h *Handler) ListCDR(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	// Pagination
	page, _ := strconv.Atoi(ctx.URLParamDefault("page", "1"))
	limit, _ := strconv.Atoi(ctx.URLParamDefault("limit", "50"))
	offset := (page - 1) * limit

	// Build query
	query := h.DB.Where("tenant_id = ?", tenantID)

	// Filters
	if ext := ctx.URLParam("extension"); ext != "" {
		query = query.Where("caller_id_number = ? OR destination_number = ?", ext, ext)
	}
	if startDate := ctx.URLParam("start_date"); startDate != "" {
		query = query.Where("start_stamp >= ?", startDate)
	}
	if endDate := ctx.URLParam("end_date"); endDate != "" {
		query = query.Where("start_stamp <= ?", endDate)
	}
	if direction := ctx.URLParam("direction"); direction != "" {
		query = query.Where("direction = ?", direction)
	}

	var total int64
	query.Model(&models.CallRecord{}).Count(&total)

	var records []models.CallRecord
	query.Order("start_stamp DESC").Offset(offset).Limit(limit).Find(&records)

	ctx.JSON(iris.Map{
		"data":  records,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) GetCDR(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)
	id, _ := ctx.Params().GetUint("id")

	var record models.CallRecord
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&record).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Call record not found"})
		return
	}

	ctx.JSON(record)
}

func (h *Handler) ExportCDR(ctx iris.Context) {
	// Placeholder - implement CSV/Excel export
	ctx.StatusCode(http.StatusNotImplemented)
	ctx.JSON(iris.Map{"error": "Export not yet implemented"})
}

// =====================
// Audit Logs
// =====================

func (h *Handler) ListAuditLogs(ctx iris.Context) {
	tenantID := middleware.GetTenantID(ctx)

	// Pagination
	page, _ := strconv.Atoi(ctx.URLParamDefault("page", "1"))
	limit, _ := strconv.Atoi(ctx.URLParamDefault("limit", "50"))
	offset := (page - 1) * limit

	query := h.DB.Where("tenant_id = ?", tenantID)

	// Filters
	if action := ctx.URLParam("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	if userID := ctx.URLParam("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	query.Model(&models.AuditLog{}).Count(&total)

	var logs []models.AuditLog
	query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs)

	ctx.JSON(iris.Map{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
