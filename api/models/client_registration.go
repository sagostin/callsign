package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EndpointType represents the type of SIP endpoint registration
type EndpointType string

const (
	EndpointTypeDevice     EndpointType = "device"      // Physical SIP device (by MAC)
	EndpointTypeMobileApp  EndpointType = "mobile_app"  // Mobile application
	EndpointTypeDesktopApp EndpointType = "desktop_app" // Desktop application
	EndpointTypeWebClient  EndpointType = "web_client"  // Browser WebRTC client
)

// ClientRegistration tracks active SIP registrations across all endpoint types.
// Every device, app, and web client registers independently in the FreeSWITCH
// directory with its own SIP credentials. This model provides a single queryable
// table to answer "which endpoints are currently registered for extension X?"
type ClientRegistration struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant ownership
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Linked User (nil for unassigned devices)
	UserID *uint `json:"user_id" gorm:"index"`
	User   *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Linked Extension (nil for unassigned devices)
	ExtensionID *uint      `json:"extension_id" gorm:"index"`
	Extension   *Extension `json:"extension,omitempty" gorm:"foreignKey:ExtensionID"`

	// Linked Device (set only for physical device endpoints)
	DeviceID *uint   `json:"device_id" gorm:"index"`
	Device   *Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`

	// Endpoint classification
	EndpointType EndpointType `json:"endpoint_type" gorm:"type:varchar(20);index;not null"` // device, mobile_app, desktop_app, web_client

	// SIP Identity
	// Format: d_<mac> for devices, app_<user_short>_<instance> for apps, web_<user_short>_<session> for web
	RegistrationUser string `json:"registration_user" gorm:"uniqueIndex;not null"`
	RegistrationPass string `json:"-"`                        // SIP password (never expose)
	DisplayName      string `json:"display_name"`             // Caller display name
	InstanceID       string `json:"instance_id" gorm:"index"` // Client-generated unique instance ID

	// Registration Status (updated via FreeSWITCH events or polling)
	Status       string     `json:"status" gorm:"default:'provisioned'"` // provisioned, registered, expired, offline
	ContactIP    string     `json:"contact_ip"`                          // Last known registration IP
	ContactPort  int        `json:"contact_port"`                        // Last known registration port
	UserAgent    string     `json:"user_agent"`                          // SIP User-Agent header
	Transport    string     `json:"transport" gorm:"default:'udp'"`      // udp, tcp, tls, wss
	RegisteredAt *time.Time `json:"registered_at"`                       // When last REGISTER was received
	ExpiresAt    *time.Time `json:"expires_at"`                          // When registration expires
	LastSeen     *time.Time `json:"last_seen"`                           // Last activity timestamp

	// Capabilities
	AllowOutbound bool `json:"allow_outbound" gorm:"default:false"` // Can make outbound calls (inherited from extension)
	WebRTC        bool `json:"webrtc" gorm:"default:false"`         // Is a WebRTC endpoint (needs special NAT handling)

	// Metadata
	DeviceLabel string `json:"device_label"` // Human-readable label ("John's iPhone", "Lobby Phone")
	AppVersion  string `json:"app_version"`  // For app/web client: version string
	OSInfo      string `json:"os_info"`      // For app/web client: "iOS 17.2", "Chrome 120"

	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID and registration password
func (cr *ClientRegistration) BeforeCreate(tx *gorm.DB) error {
	cr.UUID = uuid.New()
	if cr.RegistrationPass == "" {
		cr.RegistrationPass = uuid.New().String()[:16]
	}
	return nil
}

// IsAssigned returns true if this registration is linked to a user/extension
func (cr *ClientRegistration) IsAssigned() bool {
	return cr.UserID != nil && cr.ExtensionID != nil
}

// IsRegistered returns true if the endpoint is currently registered with FreeSWITCH
func (cr *ClientRegistration) IsRegistered() bool {
	return cr.Status == "registered"
}

// IsExpired returns true if the registration has expired
func (cr *ClientRegistration) IsExpired() bool {
	if cr.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*cr.ExpiresAt)
}

// GenerateRegistrationUser builds the SIP registration username based on endpoint type
func GenerateRegistrationUser(endpointType EndpointType, identifier string, instanceID string) string {
	switch endpointType {
	case EndpointTypeDevice:
		return "d_" + NormalizeMAC(identifier)
	case EndpointTypeMobileApp:
		return "app_" + identifier + "_" + instanceID
	case EndpointTypeDesktopApp:
		return "app_" + identifier + "_" + instanceID
	case EndpointTypeWebClient:
		return "web_" + identifier + "_" + instanceID
	default:
		return "reg_" + identifier
	}
}
