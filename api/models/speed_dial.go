package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SpeedDialGroup represents a group of speed dial entries with a common prefix
type SpeedDialGroup struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant association
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Group info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Prefix      string `json:"prefix" gorm:"not null"` // e.g., "*0", "*9"
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// Entries stored as JSON (for simplicity and ordering)
	Entries SpeedDialEntries `json:"entries" gorm:"type:jsonb;default:'[]'"`
}

// SpeedDialEntry represents a single speed dial slot
type SpeedDialEntry struct {
	Slot        int    `json:"slot"`        // Slot number (1-99, determines dial code e.g., *01, *02)
	Label       string `json:"label"`       // Display name
	Destination string `json:"destination"` // Phone number or extension to call
}

// SpeedDialEntries is a slice of entries that can be stored as JSON
type SpeedDialEntries []SpeedDialEntry

// GORM Value/Scan for SpeedDialEntries
func (e SpeedDialEntries) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *SpeedDialEntries) Scan(value interface{}) error {
	if value == nil {
		*e = SpeedDialEntries{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, e)
}

// BeforeCreate generates UUID
func (s *SpeedDialGroup) BeforeCreate(tx *gorm.DB) error {
	s.UUID = uuid.New()
	return nil
}
