package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// HolidayList represents a collection of holiday dates for a tenant
type HolidayList struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Name     string `json:"name" gorm:"not null"`

	// Source type: manual or external URL
	ExternalURL string     `json:"external_url"` // ICS/iCal URL for syncing
	LastSynced  *time.Time `json:"last_synced"`

	// Dates stored as JSON array
	Dates pq.StringArray `json:"dates" gorm:"type:text[]"` // YYYY-MM-DD format

	// Whether to use this list
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (h *HolidayList) BeforeCreate(tx *gorm.DB) error {
	h.UUID = uuid.New()
	return nil
}

// HolidayDate represents a single holiday date within a list
type HolidayDate struct {
	Date string `json:"date"` // YYYY-MM-DD
	Name string `json:"name"` // e.g., "Christmas Day"
}
