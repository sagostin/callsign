package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RingGroupStrategy defines call distribution strategies
type RingGroupStrategy string

const (
	RingStrategySimultaneous RingGroupStrategy = "simultaneous" // Ring all at once
	RingStrategySequence     RingGroupStrategy = "sequence"     // Ring one after another
	RingStrategyRandom       RingGroupStrategy = "random"       // Random order
	RingStrategyRoundRobin   RingGroupStrategy = "round-robin"  // Cycle through
	RingStrategyEnterprise   RingGroupStrategy = "enterprise"   // Ring all with per-member delays
	RingStrategyRollover     RingGroupStrategy = "rollover"     // Sequential without delay, pipe-separated
)

// RingGroup represents a ring group for call distribution
type RingGroup struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID  uint   `json:"tenant_id" gorm:"index;not null"`
	Name      string `json:"name" gorm:"not null"`
	Extension string `json:"extension" gorm:"index"` // Dial-in number

	// Strategy
	Strategy    RingGroupStrategy `json:"strategy" gorm:"default:'simultaneous'"`
	RingTimeout int               `json:"ring_timeout" gorm:"default:30"` // Seconds per destination

	// Caller ID
	CallerIDNamePrefix   string `json:"caller_id_name_prefix"` // Prefix to show on phones
	CallerIDNumberPrefix string `json:"caller_id_number_prefix"`

	// Timeouts & Failures
	TimeoutDestination     string `json:"timeout_destination"`      // Where to go on timeout
	TimeoutDestinationType string `json:"timeout_destination_type"` // extension, voicemail, ivr

	// Audio
	RingbackTone string `json:"ringback_tone"`
	MusicOnHold  string `json:"music_on_hold"`

	// Settings
	SkipBusyMembers    bool   `json:"skip_busy_members" gorm:"default:true"`
	DistinctiveRing    bool   `json:"distinctive_ring" gorm:"default:false"`
	DistinctRingName   string `json:"distinct_ring_name"` // Display name on phone, e.g., "Sales"
	AlertInfo          string `json:"alert_info"`         // SIP Alert-Info header value for ringtone
	FollowMeEnabled    bool   `json:"follow_me_enabled" gorm:"default:false"`
	MissedCallTracking bool   `json:"missed_call_tracking" gorm:"default:true"`

	// Greeting / Pre-ring audio
	GreetingPath string `json:"greeting_path"` // Audio file to play before ringing

	// Exit key — allows caller to press a key during ringing to jump to timeout destination
	ExitKey string `json:"exit_key"` // DTMF key, e.g., "1"

	// Call screening — record caller name and play to answering party
	CallScreenEnabled bool `json:"call_screen_enabled" gorm:"default:false"`

	// Ring group-level forwarding — forward entire group to a single destination
	ForwardEnabled     bool   `json:"forward_enabled" gorm:"default:false"`
	ForwardDestination string `json:"forward_destination"`

	// Missed call notification
	MissedCallEmail string `json:"missed_call_email"` // Email address for missed call alerts

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`

	// Relations
	Destinations []RingGroupDestination `json:"destinations,omitempty" gorm:"foreignKey:RingGroupID"`
}

// BeforeCreate generates UUID
func (r *RingGroup) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}

// RingGroupDestination represents a destination in a ring group
type RingGroupDestination struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Assignment
	RingGroupID uint `json:"ring_group_id" gorm:"index;not null"`
	ExtensionID uint `json:"extension_id" gorm:"index"`
	TenantID    uint `json:"tenant_id" gorm:"index;not null"`

	// Destination
	DestinationType string `json:"destination_type"` // extension, external, gateway
	Destination     string `json:"destination"`      // Extension number or external number

	// Timing (for sequence strategy)
	Delay   int `json:"delay" gorm:"default:0"`    // Seconds before ringing this dest
	Timeout int `json:"timeout" gorm:"default:30"` // Seconds to ring this dest

	// Order
	Priority int `json:"priority" gorm:"default:1"` // Lower = higher priority

	// Settings
	PromptConfirm bool `json:"prompt_confirm" gorm:"default:false"` // Press 1 to confirm
}
