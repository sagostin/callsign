package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Gateway represents a SIP trunk/gateway for outbound calling
type Gateway struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant association (null = system-wide gateway)
	TenantID *uint   `json:"tenant_id" gorm:"index"`
	Tenant   *Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Gateway identification
	GatewayName string `json:"gateway_name" gorm:"not null"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// Profile association
	ProfileName string `json:"profile_name" gorm:"default:'external'"` // Which SIP profile to use

	// Authentication
	Username     string `json:"username"`
	Password     string `json:"-"`             // Never expose
	AuthUsername string `json:"auth_username"` // If different from username
	Realm        string `json:"realm"`

	// Connection settings
	Proxy         string `json:"proxy" gorm:"not null"` // e.g., "sip.provider.com"
	RegisterProxy string `json:"register_proxy"`        // If different from proxy
	FromUser      string `json:"from_user"`
	FromDomain    string `json:"from_domain"`
	Extension     string `json:"extension"` // Extension to use for inbound

	// Transport
	Transport string `json:"transport" gorm:"default:'udp'"` // udp, tcp, tls

	// Registration
	Register          bool   `json:"register" gorm:"default:true"`
	RegisterTransport string `json:"register_transport"`
	ExpireSeconds     int    `json:"expire_seconds" gorm:"default:3600"`
	RetrySeconds      int    `json:"retry_seconds" gorm:"default:30"`

	// Caller ID
	CallerIDInFrom bool `json:"caller_id_in_from" gorm:"default:false"`

	// Codec preferences
	CodecPrefs string `json:"codec_prefs"`

	// Channels/limits
	Channels int `json:"channels" gorm:"default:0"` // 0 = unlimited

	// Ping (keepalive)
	Ping string `json:"ping"`

	// Contact params
	ContactParams string `json:"contact_params"`

	// Context for inbound calls
	Context string `json:"context" gorm:"default:'public'"`

	// Dial Format Configuration (for outbound calls)
	// These control how the system formats numbers when sending to this trunk
	DialFormat          string `json:"dial_format" gorm:"default:'e164'"`         // e164, 10d, 11d, custom
	Allow10Digit        bool   `json:"allow_10_digit" gorm:"default:true"`        // Allow 10-digit NANPA dialing
	Allow11Digit        bool   `json:"allow_11_digit" gorm:"default:true"`        // Allow 1+10 digit dialing
	InternationalPrefix string `json:"international_prefix" gorm:"default:'011'"` // Prefix for international (e.g., 011)
	InternationalFormat string `json:"international_format" gorm:"default:'011'"` // How to format outbound international
	StripPrefix         string `json:"strip_prefix"`                              // Prefix to strip before sending
	PrependPrefix       string `json:"prepend_prefix"`                            // Prefix to prepend before sending
	TechPrefix          string `json:"tech_prefix"`                               // Tech prefix for LCR

	// Rate limiting and routing
	Priority int    `json:"priority" gorm:"default:0"` // Lower = higher priority for LCR
	Weight   int    `json:"weight" gorm:"default:100"` // Weight for load balancing
	RouteTag string `json:"route_tag"`                 // Tag for dial plan routing

	// Status (read-only, updated by FreeSWITCH events)
	Status     string     `json:"status" gorm:"-"`
	LastStatus *time.Time `json:"last_status"`
}

// BeforeCreate generates UUID
func (g *Gateway) BeforeCreate(tx *gorm.DB) error {
	g.UUID = uuid.New()
	return nil
}

// GatewayToXMLParams converts gateway to FreeSWITCH XML params
func (g *Gateway) ToXMLParams() map[string]string {
	params := make(map[string]string)

	if g.Username != "" {
		params["username"] = g.Username
	}
	if g.Password != "" {
		params["password"] = g.Password
	}
	if g.AuthUsername != "" {
		params["auth-username"] = g.AuthUsername
	}
	if g.Realm != "" {
		params["realm"] = g.Realm
	}
	if g.Proxy != "" {
		params["proxy"] = g.Proxy
	}
	if g.RegisterProxy != "" {
		params["register-proxy"] = g.RegisterProxy
	}
	if g.FromUser != "" {
		params["from-user"] = g.FromUser
	}
	if g.FromDomain != "" {
		params["from-domain"] = g.FromDomain
	}
	if g.Extension != "" {
		params["extension"] = g.Extension
	}
	if g.Transport != "" {
		params["register-transport"] = g.Transport
	}
	if g.Register {
		params["register"] = "true"
	} else {
		params["register"] = "false"
	}
	if g.ExpireSeconds > 0 {
		params["expire-seconds"] = string(rune(g.ExpireSeconds))
	}
	if g.RetrySeconds > 0 {
		params["retry-seconds"] = string(rune(g.RetrySeconds))
	}
	if g.CallerIDInFrom {
		params["caller-id-in-from"] = "true"
	}
	if g.Ping != "" {
		params["ping"] = g.Ping
	}
	if g.ContactParams != "" {
		params["contact-params"] = g.ContactParams
	}
	if g.Context != "" {
		params["context"] = g.Context
	}

	return params
}
