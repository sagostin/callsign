package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NumberGroup groups system numbers for outbound routing to specific carriers.
// Each group has an ordered list of gateways with priority/weight for failover.
type NumberGroup struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Group identification
	Name        string `json:"name" gorm:"uniqueIndex;not null"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// Default gateway for this group
	DefaultGatewayID *uint    `json:"default_gateway_id" gorm:"index"`
	DefaultGateway   *Gateway `json:"default_gateway,omitempty" gorm:"foreignKey:DefaultGatewayID"`

	// Ordered list of gateways with priority/weight for failover/load-balancing
	GatewayPriorities GatewayPriorityList `json:"gateway_priorities" gorm:"type:jsonb;default:'[]'"`

	// Numbers in this group (not loaded by default)
	Numbers []SystemNumber `json:"-" gorm:"foreignKey:NumberGroupID"`
}

// BeforeCreate generates UUID
func (ng *NumberGroup) BeforeCreate(tx *gorm.DB) error {
	ng.UUID = uuid.New()
	return nil
}

// GatewayPriority represents a gateway entry in a number group's priority list
type GatewayPriority struct {
	GatewayID   uint   `json:"gateway_id"`
	GatewayName string `json:"gateway_name"` // For display
	Priority    int    `json:"priority"`     // Lower = higher priority
	Weight      int    `json:"weight"`       // For weighted load-balancing (default 1)
}

// GatewayPriorityList is a JSONB-stored slice of GatewayPriority
type GatewayPriorityList []GatewayPriority

func (g GatewayPriorityList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GatewayPriorityList) Scan(value interface{}) error {
	if value == nil {
		*g = GatewayPriorityList{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, g)
}

// SystemNumber represents a phone number in the central system pool.
// System admin adds numbers here and assigns them to tenants.
// Tenants cannot add their own numbers.
type SystemNumber struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// The phone number in E.164 format (e.g. +14155551234)
	PhoneNumber string `json:"phone_number" gorm:"uniqueIndex;not null"`

	// Display / metadata
	CallerIDName string `json:"caller_id_name"`
	Description  string `json:"description"`

	// Number group for outbound routing
	NumberGroupID *uint        `json:"number_group_id" gorm:"index"`
	NumberGroup   *NumberGroup `json:"number_group,omitempty" gorm:"foreignKey:NumberGroupID"`

	// Tenant assignment (null = unassigned / available)
	TenantID *uint   `json:"tenant_id" gorm:"index"`
	Tenant   *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`

	// Auto-created Destination when assigned to a tenant
	DestinationID *uint        `json:"destination_id" gorm:"index"`
	Destination   *Destination `json:"destination,omitempty" gorm:"foreignKey:DestinationID"`

	// Capabilities
	SMSEnabled   bool `json:"sms_enabled" gorm:"default:false"`
	MMSEnabled   bool `json:"mms_enabled" gorm:"default:false"`
	FaxEnabled   bool `json:"fax_enabled" gorm:"default:false"`
	E911Eligible bool `json:"e911_eligible" gorm:"default:true"`

	// Linked messaging number (SMS/MMS must exist on main list)
	MessagingNumberID *uint            `json:"messaging_number_id" gorm:"index"`
	MessagingNumber   *MessagingNumber `json:"messaging_number,omitempty" gorm:"foreignKey:MessagingNumberID"`

	// Status
	Status  string `json:"status" gorm:"default:'available'"` // available, assigned, reserved, porting
	Enabled bool   `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (sn *SystemNumber) BeforeCreate(tx *gorm.DB) error {
	sn.UUID = uuid.New()
	return nil
}

// SystemNumber status constants
const (
	NumberStatusAvailable = "available"
	NumberStatusAssigned  = "assigned"
	NumberStatusReserved  = "reserved"
	NumberStatusPorting   = "porting"
)
