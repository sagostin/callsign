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
// Each group has an ordered list of gateways with priority/weight for failover,
// outbound routing rules with regex matching, and an optional SMS provider.
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

	// SMS/Messaging provider for this group (default for all numbers)
	// Individual numbers can override via SystemNumber.MessagingNumberID
	MessagingProviderID *uint              `json:"messaging_provider_id" gorm:"index"`
	MessagingProvider   *MessagingProvider `json:"messaging_provider,omitempty" gorm:"foreignKey:MessagingProviderID"`

	// Outbound routing rules (voice) — regex-based route processing
	RoutingRules []OutboundRoutingRule `json:"routing_rules,omitempty" gorm:"foreignKey:NumberGroupID"`

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

// OutboundRoutingRule defines a regex-based outbound route for a number group.
// Rules are evaluated in priority order (lower = first). Each rule matches
// the dialed number against a regex pattern and applies transformations
// before routing to a specific gateway.
type OutboundRoutingRule struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Parent number group
	NumberGroupID uint         `json:"number_group_id" gorm:"index;not null"`
	NumberGroup   *NumberGroup `json:"-" gorm:"foreignKey:NumberGroupID"`

	// Rule identification
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// Matching: regex tested against the dialed number
	Pattern    string `json:"pattern" gorm:"not null"`                         // e.g. ^\+?1?(\d{10})$
	MatchField string `json:"match_field" gorm:"default:'destination_number'"` // destination_number, caller_id_number

	// Priority (lower = evaluated first)
	Priority int `json:"priority" gorm:"default:100"`
	Weight   int `json:"weight" gorm:"default:1"`

	// Number transformations applied before sending to gateway
	StripDigits int    `json:"strip_digits" gorm:"default:0"`     // Strip N leading digits
	Prepend     string `json:"prepend"`                           // Prepend after stripping
	Prefix      string `json:"prefix"`                            // e.g. international prefix 011
	DialFormat  string `json:"dial_format" gorm:"default:'e164'"` // e164, 11d, 10d, custom

	// Gateway routing — which trunk to send this call through
	GatewayID   *uint    `json:"gateway_id" gorm:"index"`
	Gateway     *Gateway `json:"gateway,omitempty" gorm:"foreignKey:GatewayID"`
	GatewayName string   `json:"gateway_name"` // For display / fallback

	// Behavior
	ContinueOnFail bool `json:"continue_on_fail" gorm:"default:true"` // Try next rule on bridge failure
}

// BeforeCreate generates UUID
func (r *OutboundRoutingRule) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}
