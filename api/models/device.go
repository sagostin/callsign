package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceType represents the type of device
type DeviceType string

const (
	DeviceTypeProvisioned DeviceType = "provisioned" // Auto-provisioned via HTTP
	DeviceTypeGenericSIP  DeviceType = "generic_sip" // Manual SIP registration
	DeviceTypeSoftphone   DeviceType = "softphone"   // WebRTC/mobile app
)

// DeviceManufacturer represents a configurable device manufacturer grouping
type DeviceManufacturer struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Identity
	Code        string `json:"code" gorm:"uniqueIndex;not null"` // e.g., "yealink", "poly", "generic"
	Name        string `json:"name" gorm:"not null"`             // Display name: "Yealink", "Poly"
	Description string `json:"description"`

	// Branding
	LogoURL string `json:"logo_url"` // URL to manufacturer logo
	Color   string `json:"color"`    // Brand color for UI

	// Device detection patterns
	UserAgentPattern string `json:"user_agent_pattern"` // Regex to detect devices
	MACPrefix        string `json:"mac_prefix"`         // Common MAC prefixes (comma-separated)

	// Sorting/display
	SortOrder int  `json:"sort_order" gorm:"default:100"`
	IsDefault bool `json:"is_default" gorm:"default:false"` // Show by default
	Enabled   bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (m *DeviceManufacturer) BeforeCreate(tx *gorm.DB) error {
	m.UUID = uuid.New()
	return nil
}

// Device represents a provisioned SIP endpoint
type Device struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Device Identity
	MAC  string `json:"mac" gorm:"uniqueIndex;not null"` // MAC address (normalized, no colons)
	Name string `json:"name"`                            // Friendly name

	// Device Type
	DeviceType   DeviceType `json:"device_type" gorm:"default:'provisioned'"`
	Manufacturer string     `json:"manufacturer"` // Yealink, Poly, Grandstream, etc.
	Model        string     `json:"model"`        // T54W, VVX450, etc.

	// Template (for provisioned devices)
	TemplateID *uint           `json:"template_id"`
	Template   *DeviceTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`

	// Profile (tenant-level device grouping with shared settings)
	ProfileID *uint          `json:"profile_id" gorm:"index"`
	Profile   *DeviceProfile `json:"profile,omitempty" gorm:"foreignKey:ProfileID"`

	// User Assignment - Devices are assigned to USERS, not extensions directly
	UserID *uint `json:"user_id" gorm:"index"`
	User   *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Lines/Accounts on this device
	Lines []DeviceLine `json:"lines,omitempty" gorm:"foreignKey:DeviceID"`

	// Device Registration (for directory)
	// Devices register with a prefix to differentiate from extensions
	RegistrationPrefix string `json:"registration_prefix" gorm:"default:'d_'"` // Prefix for device registration
	RegistrationUser   string `json:"registration_user"`                       // Auto-set: prefix + normalized MAC
	RegistrationPass   string `json:"-"`                                       // Device registration password

	// SIP Settings (for generic SIP devices)
	SIPServer    string `json:"sip_server"`
	SIPProxy     string `json:"sip_proxy"`
	SIPTransport string `json:"sip_transport" gorm:"default:'udp'"` // udp, tcp, tls
	SIPPort      int    `json:"sip_port" gorm:"default:5060"`

	// Provisioning
	ProvisionURL   string    `json:"provision_url"`   // Auto-generated provisioning URL
	ProvisionToken string    `json:"-"`               // Secret token for provisioning auth
	LastProvision  time.Time `json:"last_provision"`  // Last successful provision
	ProvisionCount int       `json:"provision_count"` // Number of times provisioned

	// Status (updated via FreeSWITCH events)
	Status         string    `json:"status" gorm:"default:'offline'"` // online, offline, ringing, busy
	Registered     bool      `json:"registered" gorm:"default:false"`
	RegistrationIP string    `json:"registration_ip"`
	UserAgent      string    `json:"user_agent"` // SIP User-Agent header
	LastSeen       time.Time `json:"last_seen"`

	// Location (for E911)
	LocationID *uint `json:"location_id"`

	// Settings
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID, provision token, and registration credentials
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	d.UUID = uuid.New()
	if d.ProvisionToken == "" {
		d.ProvisionToken = uuid.New().String()
	}

	// Set registration user: prefix + MAC
	if d.RegistrationPrefix == "" {
		d.RegistrationPrefix = "d_"
	}
	if d.RegistrationUser == "" && d.MAC != "" {
		d.RegistrationUser = d.RegistrationPrefix + NormalizeMAC(d.MAC)
	}

	// Generate registration password if not set
	if d.RegistrationPass == "" {
		d.RegistrationPass = uuid.New().String()[:16] // 16-char random password
	}

	return nil
}

// GetRegistrationUser returns the SIP registration username for this device
func (d *Device) GetRegistrationUser() string {
	if d.RegistrationUser != "" {
		return d.RegistrationUser
	}
	return d.RegistrationPrefix + NormalizeMAC(d.MAC)
}

// NormalizeMAC removes colons and converts to lowercase
func NormalizeMAC(mac string) string {
	result := ""
	for _, c := range mac {
		if c != ':' && c != '-' && c != '.' {
			result += string(c)
		}
	}
	return result
}

// DeviceLine represents a line/account on a device
type DeviceLine struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Parent device
	DeviceID uint `json:"device_id" gorm:"index;not null"`

	// Line number on the device (1, 2, 3, etc.)
	LineNumber int `json:"line_number" gorm:"default:1"`

	// Extension assignment (optional - can be unassigned)
	ExtensionID *uint      `json:"extension_id" gorm:"index"`
	Extension   *Extension `json:"extension,omitempty" gorm:"foreignKey:ExtensionID"`

	// Line label (displayed on phone)
	Label string `json:"label"`

	// Override credentials (for shared lines, BLF, etc.)
	// If empty, uses extension credentials
	UserID   string `json:"user_id_override"`   // SIP user ID override
	AuthUser string `json:"auth_user_override"` // SIP auth username override
	Password string `json:"-"`                  // SIP password override (encrypted)

	// Line type
	LineType string `json:"line_type" gorm:"default:'line'"` // line, blf, speed_dial, shared

	// For BLF/speed dial
	BLFExtension string `json:"blf_extension"` // Extension to monitor
	SpeedDial    string `json:"speed_dial"`    // Number to dial

	Enabled bool `json:"enabled" gorm:"default:true"`
}

// GetEffectiveCredentials returns the credentials for this line
// Uses extension credentials if no override is set
func (l *DeviceLine) GetEffectiveCredentials() (userID, authUser, password string) {
	if l.UserID != "" {
		userID = l.UserID
	} else if l.Extension != nil {
		userID = l.Extension.Extension
	}

	if l.AuthUser != "" {
		authUser = l.AuthUser
	} else {
		authUser = userID
	}

	if l.Password != "" {
		password = l.Password
	} else if l.Extension != nil {
		password = l.Extension.Password
	}

	return
}

// DeviceTemplate represents a provisioning template for devices
type DeviceTemplate struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Scope - NULL TenantID = system/global template
	TenantID *uint `json:"tenant_id" gorm:"index"`

	// Template Info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`

	// Device Type
	Manufacturer string `json:"manufacturer" gorm:"index"` // Yealink, Poly, Grandstream
	Model        string `json:"model"`                     // T54W, VVX450, etc.
	Family       string `json:"family"`                    // T5x, VVX4xx (for compatibility)

	// Template Content
	ConfigTemplate string `json:"config_template" gorm:"type:text"` // Template with variables
	ConfigType     string `json:"config_type" gorm:"default:'xml'"` // xml, cfg, json

	// Inheritance
	ParentID *uint           `json:"parent_id"`
	Parent   *DeviceTemplate `json:"parent,omitempty" gorm:"foreignKey:ParentID"`

	// Firmware
	FirmwareID *uint     `json:"firmware_id"`
	Firmware   *Firmware `json:"firmware,omitempty" gorm:"foreignKey:FirmwareID"`

	// Security/Verification
	UserAgentPattern string `json:"user_agent_pattern"` // Regex to match User-Agent header
	MACPattern       string `json:"mac_pattern"`        // Regex to match MAC address (e.g., ^001565 for Yealink)
	RequireHTTPS     bool   `json:"require_https" gorm:"default:false"`
	IPWhitelist      string `json:"ip_whitelist"` // Comma-separated IP/CIDR list

	// Usage count
	DeviceCount int64 `json:"device_count" gorm:"-"` // Computed field

	// Status
	IsSystem bool `json:"is_system" gorm:"default:false"` // System template (read-only for tenants)
	Enabled  bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (t *DeviceTemplate) BeforeCreate(tx *gorm.DB) error {
	t.UUID = uuid.New()
	return nil
}

// Firmware represents device firmware files
type Firmware struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Device compatibility
	Manufacturer string `json:"manufacturer" gorm:"index;not null"` // Yealink, Poly
	Model        string `json:"model"`                              // Specific model
	Family       string `json:"family"`                             // Model family

	// Version info
	Version     string    `json:"version" gorm:"not null"`
	ReleaseDate time.Time `json:"release_date"`

	// File info
	FilePath string `json:"file_path" gorm:"not null"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	Checksum string `json:"checksum"` // SHA256

	// Release notes
	ReleaseNotes string `json:"release_notes" gorm:"type:text"`

	// Status
	IsDefault bool `json:"is_default" gorm:"default:false"` // Default for new devices
	Enabled   bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (f *Firmware) BeforeCreate(tx *gorm.DB) error {
	f.UUID = uuid.New()
	return nil
}

// ProvisioningVariables holds variables for template rendering
type ProvisioningVariables struct {
	// Device info
	MAC          string
	DeviceName   string
	Model        string
	Manufacturer string

	// Server info
	Server     string
	Domain     string
	ServerIP   string
	ServerPort int
	Transport  string

	// Lines (indexed by line number)
	Lines map[int]LineVariables

	// Firmware
	FirmwareURL     string
	FirmwareVersion string

	// Tenant info
	TenantName   string
	TenantDomain string

	// Misc
	ProvisionURL string
	Timestamp    time.Time
}

// LineVariables holds variables for a single line
type LineVariables struct {
	LineNumber  int
	Extension   string
	DisplayName string
	Label       string
	UserID      string
	AuthUser    string
	Password    string
	Server      string
	Enabled     bool
}

// DeviceProfile represents a tenant-level device configuration profile
// Allows grouping devices with shared settings and overrides
type DeviceProfile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Profile Info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Color       string `json:"color" gorm:"default:'#3b82f6'"` // For UI display

	// Device matching
	Manufacturer string `json:"manufacturer"` // Limit to specific manufacturer
	Model        string `json:"model"`        // Limit to specific model

	// Base template override (overrides device's own template)
	TemplateID *uint           `json:"template_id"`
	Template   *DeviceTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`

	// Configuration overrides (JSON)
	// These settings override the template defaults
	ConfigOverrides string `json:"config_overrides" gorm:"type:text"` // JSON key-value pairs

	// Common settings
	Timezone      string `json:"timezone" gorm:"default:'America/New_York'"`
	Language      string `json:"language" gorm:"default:'en'"`
	DateFormat    string `json:"date_format" gorm:"default:'MM/DD/YYYY'"`
	TimeFormat    string `json:"time_format" gorm:"default:'12hour'"` // 12hour, 24hour
	RingtoneURL   string `json:"ringtone_url"`
	BackgroundURL string `json:"background_url"`

	// Network settings
	NTPServer    string `json:"ntp_server"`
	SyslogServer string `json:"syslog_server"`

	// Audio settings
	DefaultVolume    int  `json:"default_volume" gorm:"default:7"`
	VADEnabled       bool `json:"vad_enabled" gorm:"default:true"`
	EchoCancellation bool `json:"echo_cancellation" gorm:"default:true"`

	// Feature toggles
	DirectoryEnabled   bool `json:"directory_enabled" gorm:"default:true"`
	CallWaitingEnabled bool `json:"call_waiting_enabled" gorm:"default:true"`
	CallRecordEnabled  bool `json:"call_record_enabled" gorm:"default:false"`
	AutoAnswerEnabled  bool `json:"auto_answer_enabled" gorm:"default:false"`
	BLFEnabled         bool `json:"blf_enabled" gorm:"default:true"`

	// Firmware
	FirmwareID *uint     `json:"firmware_id"`
	Firmware   *Firmware `json:"firmware,omitempty" gorm:"foreignKey:FirmwareID"`

	// Usage stats
	DeviceCount int64 `json:"device_count" gorm:"-"` // Computed

	// Status
	IsDefault bool `json:"is_default" gorm:"default:false"` // Applied to new devices
	Enabled   bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (p *DeviceProfile) BeforeCreate(tx *gorm.DB) error {
	p.UUID = uuid.New()
	return nil
}
