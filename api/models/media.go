package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaType represents the type of audio file
type MediaType string

const (
	MediaTypeGreeting  MediaType = "greeting"
	MediaTypeRecording MediaType = "recording"
	MediaTypeMusic     MediaType = "music"
	MediaTypeQueue     MediaType = "queue"
	MediaTypeSystem    MediaType = "system"
	MediaTypeCustom    MediaType = "custom"
)

// MediaFile represents an audio file stored in the system (tenant-specific)
type MediaFile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Metadata
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Type        MediaType `json:"type" gorm:"index;default:'custom'"`
	Category    string    `json:"category"` // Generic category tag (IVR, Voicemail, etc.)

	// File Info
	Filename string `json:"filename" gorm:"not null"`
	Path     string `json:"path" gorm:"not null"` // Relative to storage root
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// BeforeCreate generates UUID
func (m *MediaFile) BeforeCreate(tx *gorm.DB) error {
	m.UUID = uuid.New()
	return nil
}
