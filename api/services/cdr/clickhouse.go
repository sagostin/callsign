package cdr

import (
	"callsign/config"
	"callsign/models"
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ClickHouseClient handles CDR storage in ClickHouse
type ClickHouseClient struct {
	conn    driver.Conn
	cfg     *config.Config
	enabled bool
}

// NewClickHouseClient creates a new ClickHouse client
func NewClickHouseClient(cfg *config.Config) *ClickHouseClient {
	return &ClickHouseClient{
		cfg:     cfg,
		enabled: cfg.ClickHouseEnabled,
	}
}

// Connect establishes connection to ClickHouse
func (c *ClickHouseClient) Connect() error {
	if !c.enabled {
		log.Info("ClickHouse is disabled")
		return nil
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", c.cfg.ClickHouseHost, c.cfg.ClickHousePort)},
		Auth: clickhouse.Auth{
			Database: c.cfg.ClickHouseDB,
			Username: c.cfg.ClickHouseUser,
			Password: c.cfg.ClickHousePass,
		},
		Debug: false,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Test connection
	if err := conn.Ping(context.Background()); err != nil {
		return fmt.Errorf("ClickHouse ping failed: %w", err)
	}

	c.conn = conn
	log.Info("Connected to ClickHouse")

	// Create tables if not exist
	return c.initSchema()
}

// initSchema creates the CDR table if it doesn't exist
func (c *ClickHouseClient) initSchema() error {
	ctx := context.Background()

	// CDR table
	cdrTable := `
	CREATE TABLE IF NOT EXISTS cdr (
		uuid UUID,
		tenant_id UInt32,
		caller_id_name String,
		caller_id_number String,
		destination_number String,
		dialed_number String,
		bridged_to String,
		start_time DateTime,
		answer_time Nullable(DateTime),
		end_time Nullable(DateTime),
		duration UInt32,
		billable_sec UInt32,
		direction LowCardinality(String),
		context String,
		hangup_cause LowCardinality(String),
		sip_code UInt16,
		gateway_name String,
		recorded Bool DEFAULT false,
		recording_path String,
		voicemail Bool DEFAULT false,
		conference Bool DEFAULT false,
		queue Bool DEFAULT false,
		queue_name String,
		rate Decimal(10, 6),
		cost Decimal(10, 4),
		extension String,
		extension_id UInt32,
		user_id UInt32,
		created_at DateTime DEFAULT now()
	) ENGINE = MergeTree()
	PARTITION BY toYYYYMM(start_time)
	ORDER BY (tenant_id, start_time, uuid)
	TTL start_time + INTERVAL 2 YEAR
	`

	if err := c.conn.Exec(ctx, cdrTable); err != nil {
		return fmt.Errorf("failed to create cdr table: %w", err)
	}

	// Hourly stats materialized view
	statsView := `
	CREATE MATERIALIZED VIEW IF NOT EXISTS cdr_hourly_stats
	ENGINE = SummingMergeTree()
	PARTITION BY toYYYYMM(hour)
	ORDER BY (tenant_id, hour, direction)
	AS SELECT
		tenant_id,
		toStartOfHour(start_time) AS hour,
		direction,
		count() AS call_count,
		countIf(answer_time IS NOT NULL) AS answered_count,
		sum(billable_sec) AS total_duration,
		sum(cost) AS total_cost
	FROM cdr
	GROUP BY tenant_id, hour, direction
	`

	if err := c.conn.Exec(ctx, statsView); err != nil {
		log.Warnf("Could not create stats view (may already exist): %v", err)
	}

	log.Info("ClickHouse schema initialized")
	return nil
}

// InsertCDR inserts a single CDR record
func (c *ClickHouseClient) InsertCDR(record *models.CallRecord) error {
	if !c.enabled || c.conn == nil {
		return nil
	}

	ctx := context.Background()
	query := `
	INSERT INTO cdr (
		uuid, tenant_id, caller_id_name, caller_id_number,
		destination_number, dialed_number, bridged_to,
		start_time, answer_time, end_time,
		duration, billable_sec, direction, context, hangup_cause, sip_code,
		gateway_name, recorded, recording_path, voicemail, conference,
		queue, queue_name, rate, cost, extension, extension_id, user_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return c.conn.Exec(ctx, query,
		record.UUID, record.TenantID, record.CallerIDName, record.CallerIDNumber,
		record.DestinationNumber, record.DialedNumber, record.BridgedTo,
		record.StartTime, record.AnswerTime, record.EndTime,
		record.Duration, record.BillableSec, record.Direction, record.Context,
		record.HangupCause, record.SIPCode, record.GatewayName,
		record.Recorded, record.RecordingPath, record.Voicemail, record.Conference,
		record.Queue, record.QueueName, record.Rate, record.Cost,
		record.Extension, record.ExtensionID, record.UserID,
	)
}

// BatchInsert inserts multiple CDR records
func (c *ClickHouseClient) BatchInsert(records []*models.CallRecord) error {
	if !c.enabled || c.conn == nil || len(records) == 0 {
		return nil
	}

	ctx := context.Background()
	batch, err := c.conn.PrepareBatch(ctx, `
		INSERT INTO cdr (
			uuid, tenant_id, caller_id_name, caller_id_number,
			destination_number, dialed_number, bridged_to,
			start_time, answer_time, end_time,
			duration, billable_sec, direction, context, hangup_cause, sip_code,
			gateway_name, recorded, recording_path, voicemail, conference,
			queue, queue_name, rate, cost, extension, extension_id, user_id
		)
	`)
	if err != nil {
		return err
	}

	for _, r := range records {
		err := batch.Append(
			r.UUID, r.TenantID, r.CallerIDName, r.CallerIDNumber,
			r.DestinationNumber, r.DialedNumber, r.BridgedTo,
			r.StartTime, r.AnswerTime, r.EndTime,
			r.Duration, r.BillableSec, r.Direction, r.Context,
			r.HangupCause, r.SIPCode, r.GatewayName,
			r.Recorded, r.RecordingPath, r.Voicemail, r.Conference,
			r.Queue, r.QueueName, r.Rate, r.Cost,
			r.Extension, r.ExtensionID, r.UserID,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

// Close closes the ClickHouse connection
func (c *ClickHouseClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsEnabled returns whether ClickHouse is enabled
func (c *ClickHouseClient) IsEnabled() bool {
	return c.enabled && c.conn != nil
}

// SyncJob handles periodic sync from PostgreSQL to ClickHouse
type SyncJob struct {
	db    *gorm.DB
	ch    *ClickHouseClient
	batch int
}

// NewSyncJob creates a new sync job
func NewSyncJob(db *gorm.DB, ch *ClickHouseClient) *SyncJob {
	return &SyncJob{
		db:    db,
		ch:    ch,
		batch: 1000,
	}
}

// Run executes the sync job
func (s *SyncJob) Run() error {
	if !s.ch.IsEnabled() {
		return nil
	}

	log.Info("Starting CDR sync to ClickHouse")
	start := time.Now()
	total := 0

	for {
		var records []*models.CallRecord
		result := s.db.Where("synced_to_click_house = ?", false).
			Limit(s.batch).
			Find(&records)

		if result.Error != nil {
			return result.Error
		}

		if len(records) == 0 {
			break
		}

		// Insert to ClickHouse
		if err := s.ch.BatchInsert(records); err != nil {
			log.Errorf("ClickHouse batch insert failed: %v", err)
			return err
		}

		// Mark as synced
		ids := make([]uint, len(records))
		for i, r := range records {
			ids[i] = r.ID
		}
		now := time.Now()
		s.db.Model(&models.CallRecord{}).Where("id IN ?", ids).Updates(map[string]interface{}{
			"synced_to_click_house": true,
			"synced_at":             now,
		})

		total += len(records)
		log.Infof("Synced %d CDR records to ClickHouse", len(records))
	}

	log.Infof("CDR sync completed: %d records in %v", total, time.Since(start))
	return nil
}

// StartPeriodicSync runs sync job on a schedule
func (s *SyncJob) StartPeriodicSync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := s.Run(); err != nil {
				log.Errorf("CDR sync job failed: %v", err)
			}
		}
	}()
}

// CleanupOldRecords removes PostgreSQL records older than retention period
func (s *SyncJob) CleanupOldRecords(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := s.db.Where("synced_to_click_house = ? AND created_at < ?", true, cutoff).
		Delete(&models.CallRecord{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Infof("Cleaned up %d old CDR records from PostgreSQL", result.RowsAffected)
	}
	return nil
}

// =====================
// ClickHouse Read Methods
// =====================

// CDRRow represents a CDR record returned from ClickHouse
type CDRRow struct {
	UUID              string     `json:"uuid"`
	TenantID          uint32     `json:"tenant_id"`
	CallerIDName      string     `json:"caller_id_name"`
	CallerIDNumber    string     `json:"caller_id_number"`
	DestinationNumber string     `json:"destination_number"`
	StartTime         time.Time  `json:"start_time"`
	AnswerTime        *time.Time `json:"answer_time,omitempty"`
	EndTime           *time.Time `json:"end_time,omitempty"`
	Duration          uint32     `json:"duration"`
	BillableSec       uint32     `json:"billable_sec"`
	Direction         string     `json:"direction"`
	HangupCause       string     `json:"hangup_cause"`
	GatewayName       string     `json:"gateway_name"`
	Recorded          bool       `json:"recorded"`
	RecordingPath     string     `json:"recording_path"`
	Extension         string     `json:"extension"`
	Cost              float64    `json:"cost"`
}

// HourlyStats represents an aggregated hourly stats row
type HourlyStats struct {
	TenantID      uint32    `json:"tenant_id"`
	Hour          time.Time `json:"hour"`
	Direction     string    `json:"direction"`
	CallCount     uint64    `json:"call_count"`
	AnsweredCount uint64    `json:"answered_count"`
	TotalDuration uint64    `json:"total_duration"`
	TotalCost     float64   `json:"total_cost"`
}

// ExtensionStats represents per-extension aggregated stats
type ExtensionStats struct {
	Extension     string  `json:"extension"`
	CallCount     uint64  `json:"call_count"`
	AnsweredCount uint64  `json:"answered_count"`
	TotalDuration uint64  `json:"total_duration"`
	AvgDuration   float64 `json:"avg_duration"`
	TotalCost     float64 `json:"total_cost"`
}

// KPIResult represents key performance indicators
type KPIResult struct {
	TotalCalls    uint64  `json:"total_calls"`
	AnsweredCalls uint64  `json:"answered_calls"`
	AnswerRate    float64 `json:"answer_rate"`
	AvgHandleTime float64 `json:"avg_handle_time"`
	TotalDuration uint64  `json:"total_duration"`
	TotalCost     float64 `json:"total_cost"`
	InboundCalls  uint64  `json:"inbound_calls"`
	OutboundCalls uint64  `json:"outbound_calls"`
}

// QueryCDR returns paginated CDR records from ClickHouse
func (c *ClickHouseClient) QueryCDR(tenantID uint, filters map[string]string, page, limit int) ([]CDRRow, int64, error) {
	if !c.enabled || c.conn == nil {
		return nil, 0, fmt.Errorf("ClickHouse not available")
	}

	ctx := context.Background()
	offset := (page - 1) * limit

	// Build WHERE clause
	where := "tenant_id = ?"
	args := []interface{}{uint32(tenantID)}

	if ext, ok := filters["extension"]; ok && ext != "" {
		where += " AND (caller_id_number = ? OR destination_number = ?)"
		args = append(args, ext, ext)
	}
	if start, ok := filters["start_date"]; ok && start != "" {
		where += " AND start_time >= ?"
		args = append(args, start)
	}
	if end, ok := filters["end_date"]; ok && end != "" {
		where += " AND start_time <= ?"
		args = append(args, end)
	}
	if dir, ok := filters["direction"]; ok && dir != "" {
		where += " AND direction = ?"
		args = append(args, dir)
	}

	// Count total
	var total uint64
	countQuery := fmt.Sprintf("SELECT count() FROM cdr WHERE %s", where)
	if err := c.conn.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch rows
	query := fmt.Sprintf(`
		SELECT uuid, tenant_id, caller_id_name, caller_id_number,
			destination_number, start_time, answer_time, end_time,
			duration, billable_sec, direction, hangup_cause,
			gateway_name, recorded, recording_path, extension, cost
		FROM cdr
		WHERE %s
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`, where)
	args = append(args, limit, offset)

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []CDRRow
	for rows.Next() {
		var r CDRRow
		if err := rows.Scan(
			&r.UUID, &r.TenantID, &r.CallerIDName, &r.CallerIDNumber,
			&r.DestinationNumber, &r.StartTime, &r.AnswerTime, &r.EndTime,
			&r.Duration, &r.BillableSec, &r.Direction, &r.HangupCause,
			&r.GatewayName, &r.Recorded, &r.RecordingPath, &r.Extension, &r.Cost,
		); err != nil {
			return nil, 0, err
		}
		results = append(results, r)
	}

	return results, int64(total), nil
}

// QueryHourlyStats returns aggregated hourly stats from the materialized view
func (c *ClickHouseClient) QueryHourlyStats(tenantID uint, from, to time.Time) ([]HourlyStats, error) {
	if !c.enabled || c.conn == nil {
		return nil, fmt.Errorf("ClickHouse not available")
	}

	ctx := context.Background()
	query := `
		SELECT tenant_id, hour, direction, call_count, answered_count, total_duration, total_cost
		FROM cdr_hourly_stats
		WHERE tenant_id = ? AND hour >= ? AND hour <= ?
		ORDER BY hour
	`

	rows, err := c.conn.Query(ctx, query, uint32(tenantID), from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []HourlyStats
	for rows.Next() {
		var s HourlyStats
		if err := rows.Scan(&s.TenantID, &s.Hour, &s.Direction, &s.CallCount,
			&s.AnsweredCount, &s.TotalDuration, &s.TotalCost); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}

// QueryExtensionStats returns per-extension aggregated stats
func (c *ClickHouseClient) QueryExtensionStats(tenantID uint, from, to time.Time) ([]ExtensionStats, error) {
	if !c.enabled || c.conn == nil {
		return nil, fmt.Errorf("ClickHouse not available")
	}

	ctx := context.Background()
	query := `
		SELECT extension,
			count() AS call_count,
			countIf(answer_time IS NOT NULL) AS answered_count,
			sum(billable_sec) AS total_duration,
			avg(billable_sec) AS avg_duration,
			sum(cost) AS total_cost
		FROM cdr
		WHERE tenant_id = ? AND start_time >= ? AND start_time <= ? AND extension != ''
		GROUP BY extension
		ORDER BY call_count DESC
	`

	rows, err := c.conn.Query(ctx, query, uint32(tenantID), from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ExtensionStats
	for rows.Next() {
		var s ExtensionStats
		if err := rows.Scan(&s.Extension, &s.CallCount, &s.AnsweredCount,
			&s.TotalDuration, &s.AvgDuration, &s.TotalCost); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}

// QueryKPI returns overall KPI metrics
func (c *ClickHouseClient) QueryKPI(tenantID uint, from, to time.Time) (*KPIResult, error) {
	if !c.enabled || c.conn == nil {
		return nil, fmt.Errorf("ClickHouse not available")
	}

	ctx := context.Background()
	query := `
		SELECT
			count() AS total_calls,
			countIf(answer_time IS NOT NULL) AS answered_calls,
			if(count() > 0, countIf(answer_time IS NOT NULL) * 100.0 / count(), 0) AS answer_rate,
			avgIf(billable_sec, answer_time IS NOT NULL) AS avg_handle_time,
			sum(billable_sec) AS total_duration,
			sum(cost) AS total_cost,
			countIf(direction = 'inbound') AS inbound_calls,
			countIf(direction = 'outbound') AS outbound_calls
		FROM cdr
		WHERE tenant_id = ? AND start_time >= ? AND start_time <= ?
	`

	var r KPIResult
	if err := c.conn.QueryRow(ctx, query, uint32(tenantID), from, to).Scan(
		&r.TotalCalls, &r.AnsweredCalls, &r.AnswerRate, &r.AvgHandleTime,
		&r.TotalDuration, &r.TotalCost, &r.InboundCalls, &r.OutboundCalls,
	); err != nil {
		return nil, err
	}
	return &r, nil
}
