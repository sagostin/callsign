package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CallHandlingRule defines a condition-based call routing rule
// Rules can be attached to Extensions (per-user) or ExtensionProfiles (shared defaults)
type CallHandlingRule struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant ownership
	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Rule ownership - one of these should be set
	ExtensionID *uint `json:"extension_id" gorm:"index"` // Per-extension rule
	ProfileID   *uint `json:"profile_id" gorm:"index"`   // Per-profile rule (applies to all extensions with this profile)

	// Rule metadata
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Priority    int    `json:"priority" gorm:"default:100"` // Lower = higher priority, evaluated first
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// Events - triggers for when to evaluate this rule
	Events CallHandlingEvents `json:"events" gorm:"type:jsonb;default:'{}'"`

	// Conditions - additional criteria that must match
	Conditions CallHandlingConditions `json:"conditions" gorm:"type:jsonb;default:'[]'"`

	// Action - what to do when rule matches
	ActionType   string                   `json:"action_type" gorm:"not null"`                  // forward, voicemail, find_me, reject, ring_devices
	ActionTarget string                   `json:"action_target"`                                // Forward number, voicemail box, etc.
	ActionParams CallHandlingActionParams `json:"action_params" gorm:"type:jsonb;default:'{}'"` // Additional settings
}

// CallHandlingEvents defines when a rule should be evaluated
type CallHandlingEvents struct {
	OnPhone     bool `json:"on_phone"`    // User is currently on another call (busy)
	NoAnswer    bool `json:"no_answer"`   // Ring timeout reached with no answer
	AnyCall     bool `json:"any_call"`    // Any incoming call (always evaluate)
	Unavailable bool `json:"unavailable"` // DND mode, offline, or unregistered
}

// CallHandlingCondition defines a single condition to match
type CallHandlingCondition struct {
	Type  string `json:"type"`  // presence, caller_id, caller_name, date_range, time_of_day, day_of_week, holiday_list
	Op    string `json:"op"`    // equals, not_equals, contains, starts_with, ends_with, regex, in, not_in, between
	Value any    `json:"value"` // The value(s) to match against
}

// CallHandlingConditions is a slice for JSONB storage
type CallHandlingConditions []CallHandlingCondition

// CallHandlingActionParams holds additional settings for the action
type CallHandlingActionParams struct {
	// For forward action
	RingTimeout int `json:"ring_timeout"` // Seconds to ring before giving up

	// For voicemail action
	GreetingID   string `json:"greeting_id"`   // Which greeting to play
	RecordOption string `json:"record_option"` // always, optional, none

	// For find_me action
	FindMeMode    string   `json:"find_me_mode"`    // simultaneous, sequential
	FindMeNumbers []string `json:"find_me_numbers"` // Numbers to ring
	FindMeDelay   int      `json:"find_me_delay"`   // Delay between sequential rings (seconds)

	// For ring_devices action (override default device ringing)
	RingDevices struct {
		Softphone bool `json:"softphone"`
		DeskPhone bool `json:"desk_phone"`
		Mobile    bool `json:"mobile"`
	} `json:"ring_devices"`

	// For reject action
	RejectReason string `json:"reject_reason"` // busy, unavailable, declined
}

// BeforeCreate generates UUID
func (r *CallHandlingRule) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}

// GORM Value/Scan for CallHandlingEvents
func (e CallHandlingEvents) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *CallHandlingEvents) Scan(value interface{}) error {
	if value == nil {
		*e = CallHandlingEvents{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, e)
}

// GORM Value/Scan for CallHandlingConditions
func (c CallHandlingConditions) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CallHandlingConditions) Scan(value interface{}) error {
	if value == nil {
		*c = CallHandlingConditions{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, c)
}

// GORM Value/Scan for CallHandlingActionParams
func (p CallHandlingActionParams) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *CallHandlingActionParams) Scan(value interface{}) error {
	if value == nil {
		*p = CallHandlingActionParams{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, p)
}
