package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VoicemailBox represents a voicemail box for an extension
type VoicemailBox struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID    uint   `json:"tenant_id" gorm:"index;not null"`
	ExtensionID uint   `json:"extension_id" gorm:"index"`
	Extension   string `json:"extension" gorm:"index;not null"` // Extension number

	// Authentication
	Password string `json:"-"` // PIN (stored hashed)

	// Notifications
	Email              string `json:"email"`
	AttachFile         bool   `json:"attach_file" gorm:"default:true"`
	TranscriptionEmail bool   `json:"transcription_email" gorm:"default:false"`

	// Settings
	Enabled          bool   `json:"enabled" gorm:"default:true"`
	GreetingPath     string `json:"greeting_path"` // Custom greeting file
	GreetingType     string `json:"greeting_type"` // name, unavailable, busy
	MaxMessages      int    `json:"max_messages" gorm:"default:50"`
	MaxMessageSecs   int    `json:"max_message_secs" gorm:"default:180"`
	SkipInstructions bool   `json:"skip_instructions" gorm:"default:false"`

	// Stats
	NewMessages   int `json:"new_messages" gorm:"default:0"`
	SavedMessages int `json:"saved_messages" gorm:"default:0"`

	// Relations
	Messages []VoicemailMessage `json:"messages,omitempty" gorm:"foreignKey:BoxID"`
}

// BeforeCreate generates UUID
func (v *VoicemailBox) BeforeCreate(tx *gorm.DB) error {
	v.UUID = uuid.New()
	return nil
}

// VoicemailMessage represents a voicemail message
type VoicemailMessage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	BoxID    uint `json:"box_id" gorm:"index;not null"`
	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Caller Info
	CallerIDName   string `json:"caller_id_name"`
	CallerIDNumber string `json:"caller_id_number"`

	// Message Info
	Duration      int       `json:"duration"` // Seconds
	FilePath      string    `json:"file_path"`
	FileSize      int64     `json:"file_size"`
	Transcription string    `json:"transcription,omitempty"`
	RecordedAt    time.Time `json:"recorded_at"`

	// Status
	IsNew       bool       `json:"is_new" gorm:"default:true;index"`
	IsUrgent    bool       `json:"is_urgent" gorm:"default:false"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	ForwardedTo string     `json:"forwarded_to,omitempty"`

	// FreeSWITCH reference
	ChannelUUID string `json:"channel_uuid"` // For correlation
}

// BeforeCreate generates UUID
func (v *VoicemailMessage) BeforeCreate(tx *gorm.DB) error {
	v.UUID = uuid.New()
	return nil
}

// MarkAsRead marks the message as read
func (v *VoicemailMessage) MarkAsRead(db *gorm.DB) error {
	now := time.Now()
	v.IsNew = false
	v.ReadAt = &now
	return db.Save(v).Error
}
