package models

import (
	"time"

	"gorm.io/gorm"
)

// Location represents an E911 location for emergency call routing.
// A location can be linked to a SystemNumber which becomes the E911 sending number
// for any extension/device assigned to this location.
type Location struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	TenantID  uint           `json:"tenant_id" gorm:"index"`
	Name      string         `json:"name"`
	Address1  string         `json:"address1"`
	Address2  string         `json:"address2,omitempty"`
	City      string         `json:"city"`
	State     string         `json:"state"`
	ZipCode   string         `json:"zip_code"`
	Country   string         `json:"country" gorm:"default:US"`
	CallerID  string         `json:"caller_id,omitempty"` // Manual E911 callback number (fallback)
	Notes     string         `json:"notes,omitempty"`
	IsDefault bool           `json:"is_default" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// E911 sending number — linked system number takes precedence over CallerID string
	SystemNumberID *uint         `json:"system_number_id" gorm:"index"`
	SystemNumber   *SystemNumber `json:"system_number,omitempty" gorm:"foreignKey:SystemNumberID"`
}
