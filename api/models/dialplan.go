package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Dialplan represents a dialplan entry for call routing
type Dialplan struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant association (null = global dialplan)
	TenantID *uint   `json:"tenant_id" gorm:"index"`
	Tenant   *Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Dialplan identification
	DialplanName    string `json:"dialplan_name" gorm:"not null"`
	DialplanContext string `json:"dialplan_context" gorm:"not null;default:'default'"`
	Description     string `json:"description"`
	Enabled         bool   `json:"enabled" gorm:"default:true"`

	// Ordering
	DialplanOrder int `json:"dialplan_order" gorm:"default:100"`

	// Pre-generated XML (like FusionPBX approach)
	// This is generated when the dialplan is saved via the API
	DialplanXML string `json:"dialplan_xml" gorm:"type:text"`

	// For continue processing
	Continue bool `json:"continue" gorm:"default:false"`

	// Details (conditions and actions)
	Details []DialplanDetail `json:"details" gorm:"foreignKey:DialplanUUID;references:UUID"`
}

// BeforeCreate generates UUID
func (d *Dialplan) BeforeCreate(tx *gorm.DB) error {
	d.UUID = uuid.New()
	return nil
}

// DialplanDetail represents conditions and actions within a dialplan
type DialplanDetail struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	DialplanUUID uuid.UUID `json:"dialplan_uuid" gorm:"type:uuid;index"`

	// Type: condition, action, anti-action
	DetailType string `json:"detail_type" gorm:"not null"` // condition, action, anti-action

	// For conditions
	ConditionField      string `json:"condition_field"`      // e.g., "destination_number"
	ConditionExpression string `json:"condition_expression"` // e.g., "^(\\d{10})$"
	ConditionBreak      string `json:"condition_break"`      // on-true, on-false, always, never

	// For actions
	ActionApplication string `json:"action_application"` // e.g., "bridge", "transfer"
	ActionData        string `json:"action_data"`        // e.g., "user/1001@${domain_name}"

	// Ordering
	DetailOrder int  `json:"detail_order" gorm:"default:10"`
	Enabled     bool `json:"enabled" gorm:"default:true"`

	// Grouping
	DetailGroup int `json:"detail_group" gorm:"default:0"`
}

// Destination represents an inbound DID/number destination
type Destination struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant association
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Number info
	DestinationNumber      string `json:"destination_number" gorm:"index;not null"` // The DID
	DestinationNumberRegex string `json:"destination_number_regex"`                 // Regex version
	Description            string `json:"description"`
	Enabled                bool   `json:"enabled" gorm:"default:true"`

	// Inbound routing - where calls TO this number go
	DestinationType   string `json:"destination_type"`   // extension, ivr, ring_group, etc.
	DestinationAction string `json:"destination_action"` // e.g., "transfer 1001 XML default"

	// Outbound gateway associations - which trunks can use this number for outbound
	// Empty = inbound only, numbers must be assigned to gateways for outbound use
	// Controlled by system admin only
	GatewayAssociations GatewayAssociations `json:"gateway_associations" gorm:"type:jsonb;default:'[]'"`
	OutboundMode        string              `json:"outbound_mode" gorm:"default:'disabled'"` // disabled, single, bridge, round_robin

	// Associated dialplan
	DialplanUUID *uuid.UUID `json:"dialplan_uuid" gorm:"type:uuid"`

	// Caller ID manipulation
	CallerIDNamePrefix   string `json:"caller_id_name_prefix"`
	CallerIDNumberPrefix string `json:"caller_id_number_prefix"`

	// Context
	Context string `json:"context" gorm:"default:'public'"`

	// Recording
	RecordEnabled bool `json:"record_enabled" gorm:"default:false"`

	// Account code
	AccountCode string `json:"account_code"`

	// Order
	DestinationOrder int `json:"destination_order" gorm:"default:100"`
}

// GatewayAssociation links a number to a gateway for outbound use
type GatewayAssociation struct {
	GatewayUUID uuid.UUID `json:"gateway_uuid"`
	GatewayName string    `json:"gateway_name"` // For display
	Priority    int       `json:"priority"`     // Lower = higher priority (for round-robin/failover)
	Weight      int       `json:"weight"`       // For weighted round-robin, default 1
}

// GatewayAssociations is a slice for JSONB storage
type GatewayAssociations []GatewayAssociation

// GORM Value/Scan for GatewayAssociations
func (g GatewayAssociations) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GatewayAssociations) Scan(value interface{}) error {
	if value == nil {
		*g = GatewayAssociations{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, g)
}

// BeforeCreate generates UUID
func (d *Destination) BeforeCreate(tx *gorm.DB) error {
	d.UUID = uuid.New()
	return nil
}
