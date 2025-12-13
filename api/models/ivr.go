package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// IVRMenu represents an IVR/Auto-Attendant menu
type IVRMenu struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Name     string `json:"name" gorm:"not null"`

	// Extension to reach this IVR
	Extension string `json:"extension" gorm:"index"`

	// Greetings
	GreetLong     string `json:"greet_long"`     // Initial greeting audio
	GreetShort    string `json:"greet_short"`    // Short greeting on retry
	InvalidSound  string `json:"invalid_sound"`  // Invalid option audio
	ExitSound     string `json:"exit_sound"`     // Exit audio
	TransferSound string `json:"transfer_sound"` // Transfer confirmation

	// Timeouts & Limits
	Timeout        int `json:"timeout" gorm:"default:10"`            // Seconds to wait for input
	MaxFailures    int `json:"max_failures" gorm:"default:3"`        // Max invalid attempts
	MaxTimeouts    int `json:"max_timeouts" gorm:"default:3"`        // Max timeout attempts
	DigitLen       int `json:"digit_len" gorm:"default:4"`           // Max digits to collect
	InterDigitTime int `json:"inter_digit_time" gorm:"default:2000"` // ms between digits

	// Features
	DirectDial     bool   `json:"direct_dial" gorm:"default:false"` // Allow extension dialing
	Ringback       string `json:"ringback"`                         // Ringback tone
	CallerIDPrefix string `json:"caller_id_prefix"`                 // Prefix for caller ID

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`

	// Relations
	Options []IVRMenuOption `json:"options,omitempty" gorm:"foreignKey:IVRMenuID"`
}

// BeforeCreate generates UUID
func (m *IVRMenu) BeforeCreate(tx *gorm.DB) error {
	m.UUID = uuid.New()
	return nil
}

// IVRMenuOption represents a DTMF option in an IVR menu
type IVRMenuOption struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Parent
	IVRMenuID uint `json:"ivr_menu_id" gorm:"index;not null"`
	TenantID  uint `json:"tenant_id" gorm:"index;not null"`

	// DTMF digit(s)
	Digits string `json:"digits" gorm:"not null"` // 1, 2, *, #, timeout, etc.

	// Action to execute
	Action      string `json:"action" gorm:"not null"` // transfer, ivr, voicemail, etc.
	ActionParam string `json:"action_param"`           // Destination/parameter

	// Display
	Description string `json:"description"`
	Order       int    `json:"order" gorm:"default:0"`

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// IVROptionAction constants
const (
	IVRActionTransfer  = "transfer"   // Transfer to extension
	IVRActionIVR       = "ivr"        // Go to another IVR
	IVRActionVoicemail = "voicemail"  // Send to voicemail
	IVRActionQueue     = "queue"      // Send to queue
	IVRActionRingGroup = "ring_group" // Send to ring group
	IVRActionPlayback  = "playback"   // Play audio
	IVRActionHangup    = "hangup"     // Hang up
	IVRActionRepeat    = "repeat"     // Repeat menu
	IVRActionExit      = "exit"       // Exit IVR
)

// TimeCondition represents time-based routing
type TimeCondition struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Name     string `json:"name" gorm:"not null"`

	// Extension to reach
	Extension string `json:"extension" gorm:"index"`

	// Schedule
	Timezone  string        `json:"timezone" gorm:"default:'America/New_York'"`
	Weekdays  pq.Int32Array `json:"weekdays" gorm:"type:integer[]"` // 0=Sun, 1=Mon, etc.
	StartTime string        `json:"start_time"`                     // HH:MM
	EndTime   string        `json:"end_time"`                       // HH:MM

	// Holiday Override - links to HolidayList for override
	HolidayListID    *uint  `json:"holiday_list_id" gorm:"index"` // Optional link to HolidayList
	HolidayDestType  string `json:"holiday_dest_type"`            // extension, ivr, queue, etc.
	HolidayDestValue string `json:"holiday_dest_value"`           // Destination ID/number when holiday matches

	// Legacy: inline holiday dates (deprecated in favor of HolidayListID)
	Holidays pq.StringArray `json:"holidays" gorm:"type:text[]"` // YYYY-MM-DD dates

	// Destinations - using extension/feature code format
	MatchDestType    string `json:"match_dest_type"`  // extension, ivr, queue, etc.
	MatchDestValue   string `json:"match_dest_value"` // Destination extension/number
	NoMatchDestType  string `json:"nomatch_dest_type"`
	NoMatchDestValue string `json:"nomatch_dest_value"`

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (t *TimeCondition) BeforeCreate(tx *gorm.DB) error {
	t.UUID = uuid.New()
	return nil
}

// CallFlow represents a multi-state toggle switch (day/night, or N custom states)
type CallFlow struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID    uint   `json:"tenant_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`

	// Extension/feature code to toggle
	Extension   string `json:"extension" gorm:"index"`
	FeatureCode string `json:"feature_code"` // *30

	// Current state index (0 = first state, 1 = second, etc.)
	CurrentState int `json:"current_state" gorm:"default:0"`

	// Destinations array - minimum 2 states, can add more
	// Example: [{"label": "Day Mode", "dest_type": "ivr", "dest_value": "Main Menu"},
	//           {"label": "Night Mode", "dest_type": "voicemail", "dest_value": "100"}]
	Destinations CallFlowDestinations `json:"destinations" gorm:"type:jsonb;default:'[]'"`

	// Optional: Sound to play when toggling
	ToggleSound string `json:"toggle_sound"`

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// CallFlowDestination represents a single toggle state/destination
type CallFlowDestination struct {
	Label     string `json:"label"`      // e.g., "Day Mode", "Night Mode", "Holiday"
	DestType  string `json:"dest_type"`  // ivr, queue, ring_group, extension, voicemail, external
	DestValue string `json:"dest_value"` // The target value
	Sound     string `json:"sound"`      // Optional sound to play when entering this state
}

// CallFlowDestinations is a slice for JSONB storage
type CallFlowDestinations []CallFlowDestination

// GORM Value/Scan for CallFlowDestinations
func (d CallFlowDestinations) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *CallFlowDestinations) Scan(value interface{}) error {
	if value == nil {
		*d = CallFlowDestinations{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, d)
}

// BeforeCreate generates UUID
func (c *CallFlow) BeforeCreate(tx *gorm.DB) error {
	c.UUID = uuid.New()
	return nil
}

// Recording represents an audio file in the library
type Recording struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// File info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	FilePath    string `json:"file_path" gorm:"not null"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	Duration    int    `json:"duration"` // seconds
	MimeType    string `json:"mime_type"`

	// Category
	Category string `json:"category"` // greeting, moh, ivr, prompt

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (r *Recording) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}

// NOTE: Contact model moved to chat.go with enhanced webhook/sync support
