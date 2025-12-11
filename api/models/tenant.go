package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tenant represents a tenant (organization/company) in the multi-tenant system
type Tenant struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Basic info
	Name        string `json:"name" gorm:"not null"`
	Domain      string `json:"domain" gorm:"uniqueIndex"`
	Description string `json:"description"`

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`

	// Service plan
	ProfileID *uint          `json:"profile_id"`
	Profile   *TenantProfile `json:"profile,omitempty" gorm:"foreignKey:ProfileID"`

	// Whitelabel settings
	WhitelabelEnabled bool   `json:"whitelabel_enabled" gorm:"default:false"`
	WhitelabelLogo    string `json:"whitelabel_logo"`
	WhitelabelName    string `json:"whitelabel_name"`
	WhitelabelPrimary string `json:"whitelabel_primary_color"`

	// SSL/TLS settings
	SSLEnabled   bool   `json:"ssl_enabled" gorm:"default:false"`
	SSLCert      string `json:"-"`
	SSLKey       string `json:"-"`
	SSLDomain    string `json:"ssl_domain"`
	SSLAutoRenew bool   `json:"ssl_auto_renew" gorm:"default:true"`

	// Settings stored as JSON
	Settings string `json:"settings" gorm:"type:jsonb;default:'{}'"`

	// Associated users (not loaded by default)
	Users []User `json:"-" gorm:"foreignKey:TenantID"`
}

// BeforeCreate generates UUID
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	t.UUID = uuid.New()
	return nil
}

// TenantProfile defines service plan limits for a tenant
type TenantProfile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Profile info
	Name        string `json:"name" gorm:"uniqueIndex;not null"`
	Description string `json:"description"`

	// Limits
	MaxExtensions     int `json:"max_extensions" gorm:"default:-1"` // -1 = unlimited
	MaxDevices        int `json:"max_devices" gorm:"default:-1"`
	MaxQueues         int `json:"max_queues" gorm:"default:-1"`
	MaxConferences    int `json:"max_conferences" gorm:"default:-1"`
	MaxRingGroups     int `json:"max_ring_groups" gorm:"default:-1"`
	MaxIVRMenus       int `json:"max_ivr_menus" gorm:"default:-1"`
	MaxVoicemailBoxes int `json:"max_voicemail_boxes" gorm:"default:-1"`
	MaxFaxServers     int `json:"max_fax_servers" gorm:"default:-1"`
	MaxUsers          int `json:"max_users" gorm:"default:-1"`

	// Call recording settings
	RecordingEnabled   bool `json:"recording_enabled" gorm:"default:true"`
	RecordingStorage   int  `json:"recording_storage_gb" gorm:"default:10"` // GB
	RecordingRetention int  `json:"recording_retention_days" gorm:"default:90"`

	// Feature flags
	FaxEnabled           bool `json:"fax_enabled" gorm:"default:true"`
	SMSEnabled           bool `json:"sms_enabled" gorm:"default:false"`
	WebRTCEnabled        bool `json:"webrtc_enabled" gorm:"default:true"`
	ConferencingEnabled  bool `json:"conferencing_enabled" gorm:"default:true"`
	CallBroadcastEnabled bool `json:"call_broadcast_enabled" gorm:"default:false"`

	// Associated tenants
	Tenants []Tenant `json:"-" gorm:"foreignKey:ProfileID"`
}

// BeforeCreate generates UUID
func (tp *TenantProfile) BeforeCreate(tx *gorm.DB) error {
	tp.UUID = uuid.New()
	return nil
}

// CheckLimit verifies if a tenant is within their profile limit
// Returns true if within limit, false if exceeded
func (tp *TenantProfile) CheckLimit(limitName string, currentCount int) bool {
	var limit int
	switch limitName {
	case "extensions":
		limit = tp.MaxExtensions
	case "devices":
		limit = tp.MaxDevices
	case "queues":
		limit = tp.MaxQueues
	case "conferences":
		limit = tp.MaxConferences
	case "ring_groups":
		limit = tp.MaxRingGroups
	case "ivr_menus":
		limit = tp.MaxIVRMenus
	case "voicemail_boxes":
		limit = tp.MaxVoicemailBoxes
	case "fax_servers":
		limit = tp.MaxFaxServers
	case "users":
		limit = tp.MaxUsers
	default:
		return true // Unknown limits default to allowed
	}

	// -1 means unlimited
	if limit < 0 {
		return true
	}

	return currentCount < limit
}
