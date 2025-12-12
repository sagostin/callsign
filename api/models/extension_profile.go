package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExtensionProfile defines permission sets and call handling rules for extensions
type ExtensionProfile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant association
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Profile info
	Name  string `json:"name" gorm:"not null"`
	Color string `json:"color" gorm:"default:'#6366f1'"` // Hex color for UI

	// Permissions stored as JSON
	Permissions ExtensionPermissions `json:"permissions" gorm:"type:jsonb;default:'{}'"`

	// Call handling overrides
	CallHandling CallHandlingOverride `json:"call_handling" gorm:"type:jsonb;default:'{}'"`

	// Custom routing override (description)
	RoutingOverride string `json:"routing_override"`

	// Extension count is computed, not stored
}

// ExtensionPermissions defines what an extension can do
type ExtensionPermissions struct {
	Outbound      bool `json:"outbound"`
	International bool `json:"international"`
	Recording     bool `json:"recording"`
	Portal        bool `json:"portal"`
	Voicemail     bool `json:"voicemail"`
}

// CallHandlingOverride defines ring strategy overrides
type CallHandlingOverride struct {
	OverrideStrategy bool   `json:"override_strategy"`
	Strategy         string `json:"strategy"` // simultaneous, sequential
	OverrideDevices  bool   `json:"override_devices"`
	Devices          struct {
		Softphone bool `json:"softphone"`
		DeskPhone bool `json:"desk_phone"`
		Mobile    bool `json:"mobile"`
	} `json:"devices"`
}

// GORM Value/Scan for ExtensionPermissions
func (p ExtensionPermissions) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ExtensionPermissions) Scan(value interface{}) error {
	if value == nil {
		*p = ExtensionPermissions{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, p)
}

// GORM Value/Scan for CallHandlingOverride
func (c CallHandlingOverride) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CallHandlingOverride) Scan(value interface{}) error {
	if value == nil {
		*c = CallHandlingOverride{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, c)
}

// BeforeCreate generates UUID
func (ep *ExtensionProfile) BeforeCreate(tx *gorm.DB) error {
	ep.UUID = uuid.New()
	return nil
}
