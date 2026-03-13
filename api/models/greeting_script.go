package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GreetingScriptStatus represents the generation status
type GreetingScriptStatus string

const (
	GreetingStatusDraft      GreetingScriptStatus = "draft"
	GreetingStatusGenerating GreetingScriptStatus = "generating"
	GreetingStatusReady      GreetingScriptStatus = "ready"
	GreetingStatusError      GreetingScriptStatus = "error"
)

// GreetingScript represents a saved TTS greeting script that can be
// generated / regenerated into an audio file for use in IVRs, voicemail, etc.
type GreetingScript struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID *uint `json:"tenant_id" gorm:"index"` // null = system-level script
	UserID   *uint `json:"user_id" gorm:"index"`   // null = admin-created (not user-owned)

	// Script content
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description,omitempty"`
	ScriptText  string `json:"script_text" gorm:"type:text;not null"`  // Multi-line script
	Category    string `json:"category" gorm:"index;default:'custom'"` // ivr, voicemail, announcement, queue, custom

	// Voice configuration (saved for regeneration)
	Provider TTSProvider `json:"provider" gorm:"default:'flite'"`
	VoiceID  string      `json:"voice_id" gorm:"default:'default'"`
	Speed    float64     `json:"speed" gorm:"default:1.0"`
	Pitch    float64     `json:"pitch" gorm:"default:1.0"`
	Language string      `json:"language" gorm:"default:'en-US'"`

	// Generated output
	FilePath    string     `json:"file_path,omitempty"`
	FileName    string     `json:"file_name,omitempty"`
	FileSize    int64      `json:"file_size,omitempty"`
	Duration    float64    `json:"duration,omitempty"` // seconds
	GeneratedAt *time.Time `json:"generated_at,omitempty"`

	// Status
	Status GreetingScriptStatus `json:"status" gorm:"default:'draft'"`
	Error  string               `json:"error,omitempty"`
}

// BeforeCreate generates UUID
func (g *GreetingScript) BeforeCreate(tx *gorm.DB) error {
	g.UUID = uuid.New()
	return nil
}
