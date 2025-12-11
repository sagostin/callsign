package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PageGroup represents a paging group for intercom/announcements
type PageGroup struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Name     string `json:"name" gorm:"not null"` // "All Phones", "Warehouse"

	// Dialing
	Extension   string `json:"extension" gorm:"index"` // Dial code, e.g., "*60"
	Description string `json:"description"`

	// Auto-Answer Settings
	AlertInfo  string `json:"alert_info" gorm:"default:'Ring Answer'"` // "Ring Answer", "Auto Answer"
	AutoAnswer string `json:"auto_answer" gorm:"default:'call_info'"`  // "call_info" or "sip_auto_answer"

	// Caller Options
	Mute           bool `json:"mute" gorm:"default:false"`      // Mute caller (announcement only)
	CheckBusy      bool `json:"check_busy" gorm:"default:true"` // Skip busy destinations
	IncludeCaller  bool `json:"include_caller" gorm:"default:false"`
	IsModerator    bool `json:"is_moderator" gorm:"default:true"`      // Page initiator is moderator
	EndConfOnLeave bool `json:"end_conf_on_leave" gorm:"default:true"` // End when moderator leaves

	// Delay/Recording Mode
	AllowRecordFirst bool `json:"allow_record_first" gorm:"default:false"` // Record message before paging
	RecordingMaxLen  int  `json:"recording_max_len" gorm:"default:90"`     // Max recording length in seconds
	SilenceThreshold int  `json:"silence_threshold" gorm:"default:200"`
	SilenceSeconds   int  `json:"silence_seconds" gorm:"default:3"`

	// Security
	PinNumber string `json:"pin_number,omitempty"` // Optional PIN protection (comma-separated for multiple)

	// Ringback/Audio
	RingbackTone string `json:"ringback_tone"`

	// Caller ID
	CallerIDName   string `json:"caller_id_name"`
	CallerIDNumber string `json:"caller_id_number"`

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`

	// Relations
	Destinations []PageGroupDestination `json:"destinations,omitempty" gorm:"foreignKey:PageGroupID"`
}

// BeforeCreate generates UUID
func (p *PageGroup) BeforeCreate(tx *gorm.DB) error {
	p.UUID = uuid.New()
	return nil
}

// PageGroupDestination represents a destination in a paging group
type PageGroupDestination struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Assignment
	PageGroupID uint `json:"page_group_id" gorm:"index;not null"`
	TenantID    uint `json:"tenant_id" gorm:"index;not null"`

	// Destination
	DestinationType string `json:"destination_type" gorm:"default:'extension'"` // extension, external
	ExtensionID     uint   `json:"extension_id" gorm:"index"`
	Destination     string `json:"destination"` // Extension number or range (e.g., "100-199")

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// ProvisioningTemplate represents a device provisioning template
type ProvisioningTemplate struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership (nil = system template)
	TenantID *uint `json:"tenant_id" gorm:"index"`

	// Template Info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Vendor      string `json:"vendor" gorm:"index;not null"` // "polycom", "yealink", "grandstream", etc.
	Model       string `json:"model"`                        // "*" for all models, or specific model
	FilePattern string `json:"file_pattern"`                 // e.g., "cfg.xml", "*.cfg"
	FileName    string `json:"file_name"`                    // The actual file name to serve
	FileType    string `json:"file_type"`                    // "cfg", "xml", "txt"

	// Template Content
	Content string `json:"content" gorm:"type:text"` // Template content with {{variables}}

	// Priority
	Priority  int  `json:"priority" gorm:"default:10"` // Lower = higher priority
	IsDefault bool `json:"is_default" gorm:"default:false"`
	Enabled   bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (p *ProvisioningTemplate) BeforeCreate(tx *gorm.DB) error {
	p.UUID = uuid.New()
	return nil
}

// ProvisioningVariable represents tenant/device-specific provisioning variables
type ProvisioningVariable struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint  `json:"tenant_id" gorm:"index;not null"`
	DeviceID *uint `json:"device_id" gorm:"index"` // nil = tenant default

	// Variable
	Name  string `json:"name"`  // e.g., "ntp_server", "admin_password"
	Value string `json:"value"` // Variable value
}

// DeviceVendor constants
const (
	VendorPolycom     = "polycom"
	VendorYealink     = "yealink"
	VendorGrandstream = "grandstream"
	VendorCisco       = "cisco"
	VendorFanvil      = "fanvil"
	VendorSnom        = "snom"
	VendorLinksys     = "linksys"
	VendorUniversal   = "universal"
)

// SupportedVendors returns list of supported device vendors
func SupportedVendors() []string {
	return []string{
		VendorPolycom,
		VendorYealink,
		VendorGrandstream,
		VendorCisco,
		VendorFanvil,
		VendorSnom,
		VendorLinksys,
		VendorUniversal,
	}
}
