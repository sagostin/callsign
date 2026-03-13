package handlers

import (
	"fmt"
	"net/http"
	"time"

	"callsign/middleware"

	"github.com/gofiber/fiber/v2"
)

// =====================
// Reports & Analytics
// =====================

// GetCallVolumeReport returns call volume statistics grouped by time interval
func (h *Handler) GetCallVolumeReport(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	// Parse query params
	interval := c.Query("interval", "hour") // hour, day, week
	startDate := c.Query("start", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDate := c.Query("end", time.Now().Format("2006-01-02"))
	direction := c.Query("direction") // inbound, outbound, or empty for both

	// Build query using CDR data
	var groupBy string
	switch interval {
	case "day":
		groupBy = "DATE(start_stamp)"
	case "week":
		groupBy = "DATE_TRUNC('week', start_stamp)"
	default:
		groupBy = "DATE_TRUNC('hour', start_stamp)"
	}

	type VolumeRow struct {
		Period     time.Time `json:"period"`
		TotalCalls int64     `json:"total_calls"`
		Answered   int64     `json:"answered"`
		Missed     int64     `json:"missed"`
		AvgDur     float64   `json:"avg_duration_seconds"`
	}

	var rows []VolumeRow
	query := h.DB.Table("xml_cdrs").
		Select(fmt.Sprintf("%s as period, COUNT(*) as total_calls, "+
			"SUM(CASE WHEN hangup_cause = 'NORMAL_CLEARING' THEN 1 ELSE 0 END) as answered, "+
			"SUM(CASE WHEN hangup_cause != 'NORMAL_CLEARING' THEN 1 ELSE 0 END) as missed, "+
			"AVG(duration) as avg_dur", groupBy)).
		Where("tenant_id = ? AND start_stamp BETWEEN ? AND ?", tenantID, startDate, endDate)

	if direction != "" {
		query = query.Where("direction = ?", direction)
	}

	query.Group(groupBy).Order("period ASC").Find(&rows)

	return c.JSON(fiber.Map{"data": rows, "interval": interval, "start": startDate, "end": endDate})
}

// GetAgentPerformanceReport returns agent-level performance metrics
func (h *Handler) GetAgentPerformanceReport(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	startDate := c.Query("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.Query("end", time.Now().Format("2006-01-02"))

	type AgentMetrics struct {
		AgentName     string  `json:"agent_name"`
		ExtensionID   uint    `json:"extension_id"`
		TotalCalls    int64   `json:"total_calls"`
		Answered      int64   `json:"answered"`
		Missed        int64   `json:"missed"`
		AvgTalkTime   float64 `json:"avg_talk_time_seconds"`
		TotalTalkTime float64 `json:"total_talk_time_seconds"`
	}

	var metrics []AgentMetrics
	h.DB.Table("xml_cdrs").
		Select("extensions.extension as agent_name, xml_cdrs.extension_id, "+
			"COUNT(*) as total_calls, "+
			"SUM(CASE WHEN hangup_cause = 'NORMAL_CLEARING' THEN 1 ELSE 0 END) as answered, "+
			"SUM(CASE WHEN hangup_cause != 'NORMAL_CLEARING' THEN 1 ELSE 0 END) as missed, "+
			"AVG(CASE WHEN billsec > 0 THEN billsec ELSE NULL END) as avg_talk_time, "+
			"SUM(billsec) as total_talk_time").
		Joins("LEFT JOIN extensions ON xml_cdrs.extension_id = extensions.id").
		Where("xml_cdrs.tenant_id = ? AND xml_cdrs.start_stamp BETWEEN ? AND ?",
			tenantID, startDate, endDate).
		Group("extensions.extension, xml_cdrs.extension_id").
		Order("total_calls DESC").
		Find(&metrics)

	return c.JSON(fiber.Map{"data": metrics, "start": startDate, "end": endDate})
}

// GetQueueStatsReport returns queue-level statistics
func (h *Handler) GetQueueStatsReport(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	type QueueStats struct {
		QueueID      uint    `json:"queue_id"`
		QueueName    string  `json:"queue_name"`
		TotalAgents  int64   `json:"total_agents"`
		ActiveAgents int64   `json:"active_agents"`
		PausedAgents int64   `json:"paused_agents"`
		AvgWaitTime  float64 `json:"avg_wait_time_seconds"`
	}

	var stats []QueueStats
	h.DB.Table("queues").
		Select("queues.id as queue_id, queues.name as queue_name, "+
			"COUNT(queue_agents.id) as total_agents, "+
			"SUM(CASE WHEN queue_agents.status = 'active' THEN 1 ELSE 0 END) as active_agents, "+
			"SUM(CASE WHEN queue_agents.status = 'paused' THEN 1 ELSE 0 END) as paused_agents").
		Joins("LEFT JOIN queue_agents ON queues.id = queue_agents.queue_id").
		Where("queues.tenant_id = ?", tenantID).
		Group("queues.id, queues.name").
		Order("queues.name").
		Find(&stats)

	return c.JSON(fiber.Map{"data": stats})
}

// GetExtensionUsageReport returns per-extension call usage statistics
func (h *Handler) GetExtensionUsageReport(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	startDate := c.Query("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.Query("end", time.Now().Format("2006-01-02"))

	// Try ClickHouse first for better performance on large datasets
	if h.CHClient != nil && h.CHClient.IsEnabled() {
		from, _ := time.Parse("2006-01-02", startDate)
		to, _ := time.Parse("2006-01-02", endDate)
		to = to.Add(24*time.Hour - time.Second) // End of day

		stats, err := h.CHClient.QueryExtensionStats(tenantID, from, to)
		if err == nil {
			return c.JSON(fiber.Map{"data": stats, "source": "clickhouse", "start": startDate, "end": endDate})
		}
		// Fallback to PostgreSQL
	}

	type ExtUsage struct {
		Extension     string  `json:"extension"`
		InboundCalls  int64   `json:"inbound_calls"`
		OutboundCalls int64   `json:"outbound_calls"`
		TotalMinutes  float64 `json:"total_minutes"`
	}

	var usage []ExtUsage
	h.DB.Table("xml_cdrs").
		Select("extensions.extension, "+
			"SUM(CASE WHEN xml_cdrs.direction = 'inbound' THEN 1 ELSE 0 END) as inbound_calls, "+
			"SUM(CASE WHEN xml_cdrs.direction = 'outbound' THEN 1 ELSE 0 END) as outbound_calls, "+
			"SUM(xml_cdrs.billsec) / 60.0 as total_minutes").
		Joins("LEFT JOIN extensions ON xml_cdrs.extension_id = extensions.id").
		Where("xml_cdrs.tenant_id = ? AND xml_cdrs.start_stamp BETWEEN ? AND ?",
			tenantID, startDate, endDate).
		Group("extensions.extension").
		Order("total_minutes DESC").
		Find(&usage)

	return c.JSON(fiber.Map{"data": usage, "source": "postgresql", "start": startDate, "end": endDate})
}

// GetKPIReport returns key performance indicators
func (h *Handler) GetKPIReport(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	startDate := c.Query("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.Query("end", time.Now().Format("2006-01-02"))

	// Try ClickHouse first for better performance on large datasets
	if h.CHClient != nil && h.CHClient.IsEnabled() {
		from, _ := time.Parse("2006-01-02", startDate)
		to, _ := time.Parse("2006-01-02", endDate)
		to = to.Add(24*time.Hour - time.Second)

		kpi, err := h.CHClient.QueryKPI(tenantID, from, to)
		if err == nil {
			return c.JSON(fiber.Map{"data": kpi, "source": "clickhouse", "start": startDate, "end": endDate})
		}
		// Fallback to PostgreSQL
	}

	type KPIs struct {
		TotalCalls     int64   `json:"total_calls"`
		AnsweredCalls  int64   `json:"answered_calls"`
		MissedCalls    int64   `json:"missed_calls"`
		ASR            float64 `json:"asr_percent"`
		ACD            float64 `json:"acd_seconds"`
		TotalMinutes   float64 `json:"total_minutes"`
		AvgCallsPerDay float64 `json:"avg_calls_per_day"`
	}

	var kpis KPIs
	h.DB.Table("xml_cdrs").
		Select("COUNT(*) as total_calls, "+
			"SUM(CASE WHEN hangup_cause = 'NORMAL_CLEARING' THEN 1 ELSE 0 END) as answered_calls, "+
			"SUM(CASE WHEN hangup_cause != 'NORMAL_CLEARING' THEN 1 ELSE 0 END) as missed_calls, "+
			"CASE WHEN COUNT(*) > 0 THEN (SUM(CASE WHEN hangup_cause = 'NORMAL_CLEARING' THEN 1 ELSE 0 END)::float / COUNT(*)::float) * 100 ELSE 0 END as asr, "+
			"AVG(CASE WHEN billsec > 0 THEN billsec ELSE NULL END) as acd, "+
			"SUM(billsec) / 60.0 as total_minutes").
		Where("tenant_id = ? AND start_stamp BETWEEN ? AND ?", tenantID, startDate, endDate).
		First(&kpis)

	// Calculate avg calls per day
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	days := end.Sub(start).Hours() / 24
	if days > 0 {
		kpis.AvgCallsPerDay = float64(kpis.TotalCalls) / days
	}

	return c.JSON(fiber.Map{"data": kpis, "source": "postgresql", "start": startDate, "end": endDate})
}

// GetNumberUsageReport returns phone number/DID utilization statistics
func (h *Handler) GetNumberUsageReport(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	startDate := c.Query("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.Query("end", time.Now().Format("2006-01-02"))

	type NumberUsage struct {
		Number       string  `json:"number"`
		InboundCalls int64   `json:"inbound_calls"`
		TotalMinutes float64 `json:"total_minutes"`
	}

	var usage []NumberUsage
	h.DB.Table("xml_cdrs").
		Select("destination_number as number, "+
			"COUNT(*) as inbound_calls, "+
			"SUM(billsec) / 60.0 as total_minutes").
		Where("tenant_id = ? AND direction = 'inbound' AND start_stamp BETWEEN ? AND ?",
			tenantID, startDate, endDate).
		Group("destination_number").
		Order("inbound_calls DESC").
		Limit(100).
		Find(&usage)

	return c.JSON(fiber.Map{"data": usage, "start": startDate, "end": endDate})
}

// ExportReport exports report data as CSV
func (h *Handler) ExportReport(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	reportType := c.Query("type", "call-volume")
	startDate := c.Query("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.Query("end", time.Now().Format("2006-01-02"))

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-report-%s.csv", reportType, startDate))

	// CSV header + data based on type
	switch reportType {
	case "call-volume":
		c.WriteString("Date,Total Calls,Answered,Missed,Avg Duration\n")
		type Row struct {
			Period     string
			TotalCalls int64
			Answered   int64
			Missed     int64
			AvgDur     float64
		}
		var rows []Row
		h.DB.Table("xml_cdrs").
			Select("DATE(start_stamp) as period, COUNT(*) as total_calls, "+
				"SUM(CASE WHEN hangup_cause = 'NORMAL_CLEARING' THEN 1 ELSE 0 END) as answered, "+
				"SUM(CASE WHEN hangup_cause != 'NORMAL_CLEARING' THEN 1 ELSE 0 END) as missed, "+
				"AVG(duration) as avg_dur").
			Where("tenant_id = ? AND start_stamp BETWEEN ? AND ?", tenantID, startDate, endDate).
			Group("DATE(start_stamp)").Order("period ASC").
			Find(&rows)
		for _, r := range rows {
			c.WriteString(fmt.Sprintf("%s,%d,%d,%d,%.1f\n", r.Period, r.TotalCalls, r.Answered, r.Missed, r.AvgDur))
		}
	default:
		h.logWarn("REPORT", "ExportReport: Unknown report type", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Unknown report type"})
	}
	return nil
}
