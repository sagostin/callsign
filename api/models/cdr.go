package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CallDirection indicates inbound or outbound
type CallDirection string

const (
	CallDirectionInbound  CallDirection = "inbound"
	CallDirectionOutbound CallDirection = "outbound"
	CallDirectionLocal    CallDirection = "local"
)

// CallRecord represents a Call Detail Record (CDR)
// Stored in PostgreSQL as live buffer, synced to ClickHouse
type CallRecord struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"` // FreeSWITCH call UUID
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`

	// Tenant
	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Caller Info
	CallerIDName   string `json:"caller_id_name"`
	CallerIDNumber string `json:"caller_id_number" gorm:"index"`

	// Destination Info
	DestinationNumber string `json:"destination_number" gorm:"index"`
	DialedNumber      string `json:"dialed_number"` // Original dialed

	// Call Legs
	AlegUUID  string `json:"a_leg_uuid" gorm:"index"`
	BlegUUID  string `json:"b_leg_uuid"`
	BridgedTo string `json:"bridged_to"` // Destination extension/number

	// Timestamps
	StartTime  time.Time  `json:"start_time" gorm:"index"`
	AnswerTime *time.Time `json:"answer_time,omitempty"`
	BridgeTime *time.Time `json:"bridge_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`

	// Duration
	Duration      int `json:"duration"`       // Total seconds
	BillableSec   int `json:"billable_sec"`   // Answered seconds
	ProgressSec   int `json:"progress_sec"`   // Ring time
	ProgressMedia int `json:"progress_media"` // Media progress

	// Call Info
	Direction   CallDirection `json:"direction" gorm:"index;default:'inbound'"`
	Context     string        `json:"context"`
	HangupCause string        `json:"hangup_cause"`
	SIPCode     int           `json:"sip_code"`

	// Routing
	GatewayName    string `json:"gateway_name"`
	SIPProfileName string `json:"sip_profile_name"`

	// Features Used
	Recorded      bool   `json:"recorded" gorm:"default:false"`
	RecordingPath string `json:"recording_path,omitempty"`
	Voicemail     bool   `json:"voicemail" gorm:"default:false"`
	Conference    bool   `json:"conference" gorm:"default:false"`
	Queue         bool   `json:"queue" gorm:"default:false"`
	QueueName     string `json:"queue_name,omitempty"`

	// Cost (if applicable)
	Rate     float64 `json:"rate" gorm:"type:decimal(10,6)"`
	Cost     float64 `json:"cost" gorm:"type:decimal(10,4)"`
	Currency string  `json:"currency" gorm:"default:'USD'"`

	// Extension info
	ExtensionID uint   `json:"extension_id,omitempty"`
	Extension   string `json:"extension"`
	UserID      uint   `json:"user_id,omitempty"`

	// Sync status
	SyncedToClickHouse bool       `json:"synced" gorm:"default:false;index"`
	SyncedAt           *time.Time `json:"synced_at,omitempty"`
}

// BeforeCreate ensures UUID is set if not provided from FreeSWITCH
func (c *CallRecord) BeforeCreate(tx *gorm.DB) error {
	if c.UUID == uuid.Nil {
		c.UUID = uuid.New()
	}
	return nil
}

// CallStatsSummary represents aggregated call statistics
type CallStatsSummary struct {
	TenantID           uint    `json:"tenant_id"`
	Date               string  `json:"date"` // YYYY-MM-DD
	TotalCalls         int     `json:"total_calls"`
	AnsweredCalls      int     `json:"answered_calls"`
	MissedCalls        int     `json:"missed_calls"`
	InboundCalls       int     `json:"inbound_calls"`
	OutboundCalls      int     `json:"outbound_calls"`
	TotalDuration      int     `json:"total_duration_sec"`
	AvgDuration        float64 `json:"avg_duration_sec"`
	TotalCost          float64 `json:"total_cost"`
	UniqueCallers      int     `json:"unique_callers"`
	UniqueDestinations int     `json:"unique_destinations"`
}

// GetCallStats calculates call statistics for a tenant and date range
func GetCallStats(db *gorm.DB, tenantID uint, startDate, endDate time.Time) (*CallStatsSummary, error) {
	var stats CallStatsSummary
	stats.TenantID = tenantID
	stats.Date = startDate.Format("2006-01-02")

	var count int64

	// Total calls
	db.Model(&CallRecord{}).Where("tenant_id = ? AND start_time BETWEEN ? AND ?",
		tenantID, startDate, endDate).Count(&count)
	stats.TotalCalls = int(count)

	// Answered calls
	db.Model(&CallRecord{}).Where("tenant_id = ? AND start_time BETWEEN ? AND ? AND answer_time IS NOT NULL",
		tenantID, startDate, endDate).Count(&count)
	stats.AnsweredCalls = int(count)

	stats.MissedCalls = stats.TotalCalls - stats.AnsweredCalls

	// Direction splits
	db.Model(&CallRecord{}).Where("tenant_id = ? AND start_time BETWEEN ? AND ? AND direction = ?",
		tenantID, startDate, endDate, CallDirectionInbound).Count(&count)
	stats.InboundCalls = int(count)
	db.Model(&CallRecord{}).Where("tenant_id = ? AND start_time BETWEEN ? AND ? AND direction = ?",
		tenantID, startDate, endDate, CallDirectionOutbound).Count(&count)
	stats.OutboundCalls = int(count)

	// Duration sums
	var totalDuration struct{ Sum int }
	db.Model(&CallRecord{}).Select("COALESCE(SUM(billable_sec), 0) as sum").
		Where("tenant_id = ? AND start_time BETWEEN ? AND ?", tenantID, startDate, endDate).
		Scan(&totalDuration)
	stats.TotalDuration = totalDuration.Sum

	if stats.AnsweredCalls > 0 {
		stats.AvgDuration = float64(stats.TotalDuration) / float64(stats.AnsweredCalls)
	}

	// Cost sum
	var totalCost struct{ Sum float64 }
	db.Model(&CallRecord{}).Select("COALESCE(SUM(cost), 0) as sum").
		Where("tenant_id = ? AND start_time BETWEEN ? AND ?", tenantID, startDate, endDate).
		Scan(&totalCost)
	stats.TotalCost = totalCost.Sum

	return &stats, nil
}
