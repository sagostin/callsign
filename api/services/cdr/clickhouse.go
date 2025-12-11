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
