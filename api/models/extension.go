package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Extension represents a phone extension/user in the system
type Extension struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant association
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Basic extension info
	Extension   string `json:"extension" gorm:"index;not null"` // e.g., "1001"
	NumberAlias string `json:"number_alias"`                    // Alternative number
	Password    string `json:"-" gorm:"not null"`               // SIP password (never expose)
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// User context (domain for call routing)
	UserContext string `json:"user_context"` // Usually the tenant domain
	Domain      string `json:"domain"`       // SIP domain

	// Caller ID
	EffectiveCallerIDName   string `json:"effective_caller_id_name"`
	EffectiveCallerIDNumber string `json:"effective_caller_id_number"`
	OutboundCallerIDName    string `json:"outbound_caller_id_name"`
	OutboundCallerIDNumber  string `json:"outbound_caller_id_number"`
	EmergencyCallerIDName   string `json:"emergency_caller_id_name"`
	EmergencyCallerIDNumber string `json:"emergency_caller_id_number"`

	// Call settings
	CallTimeout int    `json:"call_timeout" gorm:"default:30"`
	TollAllow   string `json:"toll_allow"` // e.g., "domestic,local,emergency"

	// Call forwarding
	ForwardAllEnabled                   bool   `json:"forward_all_enabled" gorm:"default:false"`
	ForwardAllDestination               string `json:"forward_all_destination"`
	ForwardBusyEnabled                  bool   `json:"forward_busy_enabled" gorm:"default:false"`
	ForwardBusyDestination              string `json:"forward_busy_destination"`
	ForwardNoAnswerEnabled              bool   `json:"forward_no_answer_enabled" gorm:"default:false"`
	ForwardNoAnswerDestination          string `json:"forward_no_answer_destination"`
	ForwardUserNotRegisteredEnabled     bool   `json:"forward_user_not_registered_enabled" gorm:"default:false"`
	ForwardUserNotRegisteredDestination string `json:"forward_user_not_registered_destination"`

	// Do Not Disturb
	DoNotDisturb bool `json:"do_not_disturb" gorm:"default:false"`

	// Voicemail
	VoicemailEnabled  bool   `json:"voicemail_enabled" gorm:"default:true"`
	VoicemailPassword string `json:"-"` // Voicemail PIN
	VoicemailMailTo   string `json:"voicemail_mail_to"`
	VoicemailFile     string `json:"voicemail_file"` // Greeting file

	// Follow Me
	FollowMeEnabled     bool   `json:"follow_me_enabled" gorm:"default:false"`
	FollowMeDestination string `json:"follow_me_destination"`

	// Recording
	RecordInbound  bool `json:"record_inbound" gorm:"default:false"`
	RecordOutbound bool `json:"record_outbound" gorm:"default:false"`

	// Limit/Concurrent calls
	LimitMax         int    `json:"limit_max" gorm:"default:5"`
	LimitDestination string `json:"limit_destination"`

	// Account code for billing
	AccountCode string `json:"account_code"`

	// MWI (Message Waiting Indicator) account
	MWIAccount string `json:"mwi_account"`

	// Associated user (optional)
	UserID *uint `json:"user_id" gorm:"index"`
	User   *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Device association
	DeviceUUID string `json:"device_uuid"`
}

// BeforeCreate generates UUID
func (e *Extension) BeforeCreate(tx *gorm.DB) error {
	e.UUID = uuid.New()
	return nil
}

// GetDialString returns the FreeSWITCH dial string for this extension
func (e *Extension) GetDialString() string {
	return "${sofia_contact(" + e.Extension + "@" + e.Domain + ")}"
}

// ExtensionSetting stores additional key-value settings for an extension
type ExtensionSetting struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ExtensionUUID uuid.UUID `json:"extension_uuid" gorm:"type:uuid;index"`
	SettingName   string    `json:"setting_name"`
	SettingValue  string    `json:"setting_value"`
	Enabled       bool      `json:"enabled" gorm:"default:true"`
}
