package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SMSNumberAssignment maps a DID (Destination) to an extension for dedicated SMS
type SMSNumberAssignment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// The SMS-enabled DID
	DestinationID uint `json:"destination_id" gorm:"index;not null"`

	// The user/extension this number is dedicated to
	ExtensionID uint `json:"extension_id" gorm:"index;not null"`

	// Whether this is the user's default outbound SMS number
	IsDefault bool `json:"is_default" gorm:"default:false"`

	Enabled bool `json:"enabled" gorm:"default:true"`
}

func (a *SMSNumberAssignment) BeforeCreate(tx *gorm.DB) error {
	a.UUID = uuid.New()
	return nil
}

// MessageQueueItem represents an outbound message in the delivery queue
type MessageQueueItem struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Source message (one of these will be set)
	MessageID     *uint `json:"message_id" gorm:"index"`      // FK to Message (SMS model)
	ChatMessageID *uint `json:"chat_message_id" gorm:"index"` // FK to ChatMessage

	// Provider to use
	ProviderID uint `json:"provider_id" gorm:"index;not null"`

	// Destination info
	FromNumber string `json:"from_number" gorm:"not null"`
	ToNumber   string `json:"to_number" gorm:"not null"`
	Body       string `json:"body" gorm:"type:text"`
	HasMedia   bool   `json:"has_media" gorm:"default:false"`

	// Delivery status
	Status      string `json:"status" gorm:"default:'pending'"` // pending, processing, sent, delivered, failed, retry
	Attempts    int    `json:"attempts" gorm:"default:0"`
	MaxAttempts int    `json:"max_attempts" gorm:"default:3"`

	// Retry scheduling
	NextRetryAt *time.Time `json:"next_retry_at"`
	LastError   string     `json:"last_error" gorm:"type:text"`

	// Provider response
	ProviderMessageID string `json:"provider_message_id"`
	ProviderResponse  string `json:"provider_response" gorm:"type:text"` // Raw JSON response

	// Processing metadata
	ProcessedAt *time.Time `json:"processed_at"`
}

func (q *MessageQueueItem) BeforeCreate(tx *gorm.DB) error {
	q.UUID = uuid.New()
	return nil
}

// MediaTranscodeJob tracks image/video transcoding for MMS delivery
type MediaTranscodeJob struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Source reference
	AttachmentID *uint `json:"attachment_id" gorm:"index"` // FK to ChatAttachment or MessageMedia

	// Source media
	SourceContentType string `json:"source_content_type" gorm:"not null"` // image/jpeg, video/mp4, etc
	SourceSizeBytes   int64  `json:"source_size_bytes"`
	SourceURL         string `json:"source_url"`         // If from external URL
	SourceData        string `json:"-" gorm:"type:text"` // Base64 original if stored locally

	// Output
	OutputContentType string `json:"output_content_type"`
	OutputSizeBytes   int64  `json:"output_size_bytes"`
	OutputData        string `json:"-" gorm:"type:text"` // Base64 transcoded result
	OutputURL         string `json:"output_url"`         // If stored externally
	ThumbnailData     string `json:"-" gorm:"type:text"` // Base64 thumbnail

	// Transcoding config
	TargetSizeKB int `json:"target_size_kb" gorm:"default:600"` // Carrier limit (Tier 2 = 600KB)
	MaxWidth     int `json:"max_width" gorm:"default:1920"`
	MaxHeight    int `json:"max_height" gorm:"default:1920"`
	Quality      int `json:"quality" gorm:"default:85"`

	// Status
	Status       string `json:"status" gorm:"default:'pending'"` // pending, processing, complete, failed
	ErrorMessage string `json:"error_message"`

	// Debug info
	FFmpegCommand string `json:"ffmpeg_command,omitempty" gorm:"type:text"`

	ProcessedAt *time.Time `json:"processed_at"`
}

func (j *MediaTranscodeJob) BeforeCreate(tx *gorm.DB) error {
	j.UUID = uuid.New()
	return nil
}
