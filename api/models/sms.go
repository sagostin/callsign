package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Message represents an SMS/MMS message (encrypted at rest)
type Message struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID    uint `json:"tenant_id" gorm:"index;not null"`
	ExtensionID uint `json:"extension_id" gorm:"index"`

	// Type: sms or mms
	Type string `json:"type" gorm:"default:'sms'"` // sms, mms

	// Message info
	Direction string `json:"direction" gorm:"not null"` // inbound, outbound
	From      string `json:"from" gorm:"index"`
	To        string `json:"to" gorm:"index"`

	// Encrypted content
	BodyEncrypted string `json:"-" gorm:"column:body_encrypted;type:text"`
	BodyHash      string `json:"-" gorm:"column:body_hash;index"`

	// Decrypted content (NOT persisted)
	Body string `json:"body" gorm:"-"`

	// Provider info (agnostic)
	ProviderID  uint   `json:"provider_id" gorm:"index"`
	ExternalID  string `json:"external_id"`  // Provider message ID
	ExternalRef string `json:"external_ref"` // Provider reference

	// Status
	Status       string     `json:"status" gorm:"default:'pending'"` // pending, queued, sent, delivered, failed, undeliverable
	ErrorCode    string     `json:"error_code,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	ReadAt       *time.Time `json:"read_at,omitempty"`

	// Cost tracking
	Segments int     `json:"segments" gorm:"default:1"`
	Cost     float64 `json:"cost,omitempty"`

	// Relations
	Media []MessageMedia `json:"media,omitempty" gorm:"foreignKey:MessageID"`
}

// BeforeCreate generates UUID
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	m.UUID = uuid.New()
	return nil
}

// MessageMedia represents MMS media attachments (stored as base64)
type MessageMedia struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	MessageID uint `json:"message_id" gorm:"index;not null"`
	TenantID  uint `json:"tenant_id" gorm:"index;not null"`

	// Media info
	ContentType string `json:"content_type" gorm:"not null"` // image/jpeg, video/mp4, etc.
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`

	// Storage (base64 encoded for encryption/portability)
	DataBase64      string `json:"-" gorm:"type:text"` // Base64 encoded media
	DataHash        string `json:"-" gorm:"index"`     // For deduplication
	ThumbnailBase64 string `json:"-" gorm:"type:text"` // Thumbnail for images/video

	// External URL (if stored in CDN/S3)
	ExternalURL string `json:"external_url,omitempty"`

	// Decrypted content (NOT persisted)
	Data      []byte `json:"data,omitempty" gorm:"-"`
	Thumbnail []byte `json:"thumbnail,omitempty" gorm:"-"`
}

// Conversation represents a message thread
type Conversation struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID    uint `json:"tenant_id" gorm:"index;not null"`
	ExtensionID uint `json:"extension_id" gorm:"index"`

	// Participants
	LocalNumber  string `json:"local_number" gorm:"index"`
	RemoteNumber string `json:"remote_number" gorm:"index"`

	// Status
	UnreadCount int       `json:"unread_count" gorm:"default:0"`
	LastMessage time.Time `json:"last_message"`
	Status      string    `json:"status" gorm:"default:'active'"` // active, archived, blocked

	// Relations
	Messages []Message `json:"messages,omitempty" gorm:"foreignKey:TenantID;references:TenantID"`
}

// BeforeCreate generates UUID
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	c.UUID = uuid.New()
	return nil
}

// MessagingProvider represents an SMS/MMS provider configuration
type MessagingProvider struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership (NULL = system, otherwise tenant-specific)
	TenantID *uint `json:"tenant_id" gorm:"index"`

	// Provider info
	Name     string `json:"name" gorm:"not null"`      // Display name
	Type     string `json:"type" gorm:"not null"`      // twilio, signalwire, bandwidth, telnyx, plivo, nexmo, custom
	Priority int    `json:"priority" gorm:"default:0"` // Lower = higher priority

	// API Credentials (encrypted)
	APIKeyEncrypted    string `json:"-" gorm:"column:api_key_encrypted"`
	APISecretEncrypted string `json:"-" gorm:"column:api_secret_encrypted"`
	AccountSID         string `json:"account_sid,omitempty"`
	AuthToken          string `json:"-"` // Not persisted, decrypted from APISecretEncrypted

	// Endpoints
	BaseURL        string `json:"base_url,omitempty"`        // API base URL
	SendEndpoint   string `json:"send_endpoint,omitempty"`   // Send message endpoint
	StatusEndpoint string `json:"status_endpoint,omitempty"` // Status callback

	// Webhook verification
	WebhookSecret   string `json:"-" gorm:"column:webhook_secret"` // For verifying inbound webhooks
	VerifySignature bool   `json:"verify_signature" gorm:"default:true"`

	// Capabilities
	SupportsSMS bool `json:"supports_sms" gorm:"default:true"`
	SupportsMMS bool `json:"supports_mms" gorm:"default:false"`
	SupportsRCS bool `json:"supports_rcs" gorm:"default:false"`

	// Phone numbers associated with this provider
	PhoneNumbers pq.StringArray `json:"phone_numbers" gorm:"type:text[]"`

	// Rate limiting
	RateLimitPerSecond int `json:"rate_limit_per_second" gorm:"default:10"`
	RateLimitPerMinute int `json:"rate_limit_per_minute" gorm:"default:100"`

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (p *MessagingProvider) BeforeCreate(tx *gorm.DB) error {
	p.UUID = uuid.New()
	return nil
}

// ProviderType constants
const (
	ProviderTwilio     = "twilio"
	ProviderSignalWire = "signalwire"
	ProviderBandwidth  = "bandwidth"
	ProviderTelnyx     = "telnyx"
	ProviderPlivo      = "plivo"
	ProviderNexmo      = "nexmo"
	ProviderVonage     = "vonage"
	ProviderCustom     = "custom"
)

// Chatplan represents a chatplan routing rule (SMS dialplan)
type Chatplan struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint   `json:"tenant_id" gorm:"index"` // NULL = global
	Name     string `json:"name" gorm:"not null"`
	Context  string `json:"context" gorm:"default:'default'"`

	// Matching
	Direction    string `json:"direction" gorm:"default:'inbound'"` // inbound, outbound
	FromPattern  string `json:"from_pattern"`                       // Regex pattern
	ToPattern    string `json:"to_pattern"`                         // Regex pattern
	MessageMatch string `json:"message_match"`                      // Regex for message body

	// Action
	Action      string `json:"action" gorm:"not null"` // forward, reply, webhook, lua, queue
	ActionParam string `json:"action_param"`

	// For webhook action
	WebhookURL     string `json:"webhook_url,omitempty"`
	WebhookMethod  string `json:"webhook_method,omitempty"`
	WebhookHeaders string `json:"webhook_headers,omitempty"` // JSON headers

	// For reply action
	ReplyTemplate string `json:"reply_template,omitempty"`

	// For forward action
	ForwardTo string `json:"forward_to,omitempty"`

	// Priority & Status
	Order   int  `json:"order" gorm:"default:0"`
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (c *Chatplan) BeforeCreate(tx *gorm.DB) error {
	c.UUID = uuid.New()
	return nil
}

// ChatplanAction constants
const (
	ChatplanActionForward = "forward"
	ChatplanActionReply   = "reply"
	ChatplanActionWebhook = "webhook"
	ChatplanActionLua     = "lua"
	ChatplanActionQueue   = "queue"
)

// Phrase represents a TTS/audio phrase definition
type Phrase struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID  uint   `json:"tenant_id" gorm:"index"` // NULL = system phrase
	MacroName string `json:"macro_name" gorm:"index;not null"`
	Language  string `json:"language" gorm:"default:'en'"`
	Module    string `json:"module"` // mod_say_en

	Type    string `json:"type" gorm:"default:'tts'"` // tts, file, prompt
	Content string `json:"content"`

	Enabled bool `json:"enabled" gorm:"default:true"`
}

func (p *Phrase) BeforeCreate(tx *gorm.DB) error {
	p.UUID = uuid.New()
	return nil
}

// Sound represents an audio file
type Sound struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID    uint   `json:"tenant_id" gorm:"index"` // NULL = system sound
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Language    string `json:"language" gorm:"default:'en'"`
	Category    string `json:"category"` // ivr, moh, digits, misc

	FilePath   string `json:"file_path" gorm:"not null"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	Duration   int    `json:"duration"`
	SampleRate int    `json:"sample_rate" gorm:"default:8000"`
	Format     string `json:"format" gorm:"default:'wav'"`

	Enabled bool `json:"enabled" gorm:"default:true"`
}

func (s *Sound) BeforeCreate(tx *gorm.DB) error {
	s.UUID = uuid.New()
	return nil
}

// DefaultOutboundRoute represents system-level outbound routing
type DefaultOutboundRoute struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`

	DigitPrefix string `json:"digit_prefix"`
	DigitMin    int    `json:"digit_min" gorm:"default:7"`
	DigitMax    int    `json:"digit_max" gorm:"default:15"`

	GatewayID     uint   `json:"gateway_id" gorm:"index"`
	Gateway2ID    *uint  `json:"gateway_2_id" gorm:"index"` // Failover
	DialString    string `json:"dial_string"`
	StripDigits   int    `json:"strip_digits" gorm:"default:0"`
	PrependDigits string `json:"prepend_digits"`

	Order   int  `json:"order" gorm:"default:0"`
	Enabled bool `json:"enabled" gorm:"default:true"`
	Default bool `json:"default" gorm:"default:false"`
}

func (r *DefaultOutboundRoute) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}
