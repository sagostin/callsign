package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// FaxBox represents a virtual fax machine assigned to a tenant
type FaxBox struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID  uint   `json:"tenant_id" gorm:"index;not null"`
	Name      string `json:"name" gorm:"not null"` // e.g. "Sales Fax", "HR Confidential"
	DID       string `json:"did" gorm:"not null"`  // Phone number (10-digit after transforms)
	Extension string `json:"extension"`            // Internal extension number

	// Caller ID & Fax Header (from gofaxserver TenantNumber)
	CallerIDName string `json:"caller_id_name"` // CID name displayed on the fax
	Header       string `json:"header"`         // Header text at the top of fax pages

	// Notification settings (from gofaxserver notify pattern)
	NotifyEmails   pq.StringArray `json:"notify_emails" gorm:"type:text[]"` // Emails to notify on receive
	NotifyWebhook  string         `json:"notify_webhook"`                   // Webhook URL for notifications
	EmailOnReceive bool           `json:"email_on_receive" gorm:"default:true"`

	// Retention
	RetentionDays int  `json:"retention_days" gorm:"default:90"`
	Enabled       bool `json:"enabled" gorm:"default:true"`

	// Relations
	Endpoints []FaxEndpoint `json:"endpoints,omitempty" gorm:"foreignKey:FaxBoxID"`
	Jobs      []FaxJob      `json:"jobs,omitempty" gorm:"foreignKey:FaxBoxID"`
}

func (f *FaxBox) BeforeCreate(tx *gorm.DB) error {
	if f.UUID == uuid.Nil {
		f.UUID = uuid.New()
	}
	return nil
}

// FaxEndpoint represents a delivery target for fax routing
// Ported from gofaxserver's Endpoint model with same priority/bridge semantics
type FaxEndpoint struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Scope: "box" (specific fax box), "tenant" (all boxes in tenant), "global" (system-wide)
	Type   string `json:"type" gorm:"not null;default:'tenant'"` // box, tenant, global
	TypeID uint   `json:"type_id" gorm:"index"`                  // FaxBox ID or Tenant ID (0 for global)

	// For convenience when Type == "box"
	FaxBoxID *uint `json:"fax_box_id,omitempty" gorm:"index"`
	TenantID *uint `json:"tenant_id,omitempty" gorm:"index"`

	// Endpoint config (from gofaxserver)
	EndpointType string `json:"endpoint_type" gorm:"not null"` // gateway, webhook, email
	Endpoint     string `json:"endpoint" gorm:"not null"`      // gateway name, webhook URL, or email address(es)
	Priority     uint   `json:"priority" gorm:"default:0"`     // 0 = highest, 666 = disabled for delivery, 999 = fallback
	Bridge       bool   `json:"bridge" gorm:"default:false"`   // Enable fax bridging/transcoding mode

	Enabled bool `json:"enabled" gorm:"default:true"`
}

// FaxJob represents a fax send/receive job with full lifecycle tracking
// Ported from gofaxserver's FaxJob with DB persistence for results
type FaxJob struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint  `json:"tenant_id" gorm:"index;not null"`
	FaxBoxID *uint `json:"fax_box_id,omitempty" gorm:"index"` // nullable for inbound to unknown box

	// Direction
	Direction string `json:"direction" gorm:"not null;default:'inbound'"` // inbound, outbound

	// Call info (from gofaxserver FaxJob)
	CallerNumber string `json:"caller_number"`
	CalleeNumber string `json:"callee_number"`
	CallerIDName string `json:"caller_id_name"`
	Header       string `json:"header"` // Fax page header text

	// Status tracking (from gofaxserver queue phases)
	Status string `json:"status" gorm:"not null;default:'queued'"` // queued, routing, sending, receiving, bridging, complete, failed, waiting
	Phase  string `json:"phase"`                                   // ROUTED, BRIDGING, RECEIVING, SENDING, WAITING, DONE

	// File info
	FileName         string `json:"file_name"`     // Path to TIFF file on disk
	PDFFileName      string `json:"pdf_file_name"` // Path to converted PDF
	Pages            int    `json:"pages" gorm:"default:0"`
	TransferredPages int    `json:"transferred_pages" gorm:"default:0"`
	FileSize         int64  `json:"file_size"` // Bytes

	// Retry tracking (from gofaxserver queue retry logic)
	Attempts    int        `json:"attempts" gorm:"default:0"`
	MaxRetries  int        `json:"max_retries" gorm:"default:3"`
	LastError   string     `json:"last_error"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// Result details (from gofaxlib FaxResult)
	CallUUID     *uuid.UUID `json:"call_uuid,omitempty" gorm:"type:uuid"`
	Success      bool       `json:"success" gorm:"default:false"`
	HangupCause  string     `json:"hangup_cause"`
	RemoteID     string     `json:"remote_id"`     // Remote station ID
	LocalID      string     `json:"local_id"`      // Local station ID
	TransferRate uint       `json:"transfer_rate"` // bps
	ECM          bool       `json:"ecm"`           // Error Correction Mode used
	T38Status    string     `json:"t38_status"`    // negotiated, rejected, etc.
	ResultCode   int        `json:"result_code"`
	ResultText   string     `json:"result_text"`

	// Timestamps
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Source info (from gofaxserver)
	SourceType string `json:"source_type"` // gateway, api, webhook
	SourceID   string `json:"source_id"`   // gateway name, API user, etc.

	// Endpoint used
	EndpointType string `json:"endpoint_type"` // gateway, webhook, email
	EndpointName string `json:"endpoint_name"` // What endpoint was used for delivery

	// FaxBox relation
	FaxBox *FaxBox `json:"fax_box,omitempty" gorm:"foreignKey:FaxBoxID"`
}

func (f *FaxJob) BeforeCreate(tx *gorm.DB) error {
	if f.UUID == uuid.Nil {
		f.UUID = uuid.New()
	}
	return nil
}

// FaxPageResult stores per-page results from SpanDSP (from gofaxlib PageResult)
type FaxPageResult struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	FaxJobID  uint      `json:"fax_job_id" gorm:"index;not null"`

	Page             uint   `json:"page"`
	BadRows          uint   `json:"bad_rows"`
	LongestBadRowRun uint   `json:"longest_bad_row_run"`
	EncodingName     string `json:"encoding_name"`
	ImageSize        uint   `json:"image_size"` // bytes
	ImageWidth       uint   `json:"image_width"`
	ImageHeight      uint   `json:"image_height"`
	ImageResX        uint   `json:"image_res_x"`
	ImageResY        uint   `json:"image_res_y"`
}

// FaxNotificationLog tracks notification delivery (email, webhook) for a fax
type FaxNotificationLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	FaxJobID  uint      `json:"fax_job_id" gorm:"index;not null"`

	Type      string     `json:"type"`      // email, webhook
	Recipient string     `json:"recipient"` // Email address or webhook URL
	Status    string     `json:"status"`    // sent, failed, pending
	Attempts  int        `json:"attempts"`
	LastError string     `json:"last_error"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
}

// FaxStats provides aggregated fax statistics for a tenant or fax box
type FaxStats struct {
	TotalInbound       int64 `json:"total_inbound"`
	TotalOutbound      int64 `json:"total_outbound"`
	SuccessfulInbound  int64 `json:"successful_inbound"`
	SuccessfulOutbound int64 `json:"successful_outbound"`
	FailedInbound      int64 `json:"failed_inbound"`
	FailedOutbound     int64 `json:"failed_outbound"`
	PendingJobs        int64 `json:"pending_jobs"`
	TotalPages         int64 `json:"total_pages"`
}
