package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ChatChannel types
const (
	ChannelSMS      = "sms"
	ChannelMMS      = "mms"
	ChannelInternal = "internal" // Extension-to-extension
	ChannelWeb      = "web"      // Web chat widget
	ChannelRoom     = "room"     // Chat room
)

// ChatRoom represents a multi-user chat room
type ChatRoom struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Room info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Type        string `json:"type" gorm:"default:'group'"` // group, private, support

	// Settings
	IsPublic        bool `json:"is_public" gorm:"default:false"`
	MaxParticipants int  `json:"max_participants" gorm:"default:100"`
	RetentionDays   int  `json:"retention_days" gorm:"default:365"`

	// Ownership
	CreatedByID uint `json:"created_by_id" gorm:"index"`

	// Members
	Members []ChatRoomMember `json:"members,omitempty" gorm:"foreignKey:RoomID"`

	// Status
	Archived bool `json:"archived" gorm:"default:false"`
}

func (r *ChatRoom) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}

// ChatRoomMember represents room membership
type ChatRoomMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	RoomID      uint `json:"room_id" gorm:"index;not null"`
	ExtensionID uint `json:"extension_id" gorm:"index;not null"`

	// Role in room
	Role string `json:"role" gorm:"default:'member'"` // owner, admin, member

	// Notification settings
	Muted       bool       `json:"muted" gorm:"default:false"`
	LastReadAt  *time.Time `json:"last_read_at"`
	UnreadCount int        `json:"unread_count" gorm:"default:0"`
}

// ChatQueue represents a shared chat queue for support
type ChatQueue struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Queue info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`

	// Routing
	Strategy string `json:"strategy" gorm:"default:'round-robin'"` // round-robin, least-busy, broadcast

	// SLA settings
	MaxWaitSeconds   int `json:"max_wait_seconds" gorm:"default:300"`
	TargetResponseMs int `json:"target_response_ms" gorm:"default:60000"`

	// Agents
	Agents []ChatQueueAgent `json:"agents,omitempty" gorm:"foreignKey:QueueID"`

	// Auto-response
	WelcomeMessage string `json:"welcome_message"`
	AwayMessage    string `json:"away_message"`

	// Hours
	BusinessHoursID *uint `json:"business_hours_id" gorm:"index"`

	Enabled bool `json:"enabled" gorm:"default:true"`
}

func (q *ChatQueue) BeforeCreate(tx *gorm.DB) error {
	q.UUID = uuid.New()
	return nil
}

// ChatQueueAgent represents an agent in a chat queue
type ChatQueueAgent struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	QueueID     uint `json:"queue_id" gorm:"index;not null"`
	ExtensionID uint `json:"extension_id" gorm:"index;not null"`

	// Agent settings
	MaxConcurrentChats int  `json:"max_concurrent_chats" gorm:"default:5"`
	Priority           int  `json:"priority" gorm:"default:0"`
	Enabled            bool `json:"enabled" gorm:"default:true"`

	// Status
	Status         string     `json:"status" gorm:"default:'available'"` // available, busy, away, offline
	ActiveChats    int        `json:"active_chats" gorm:"default:0"`
	LastAssignedAt *time.Time `json:"last_assigned_at"`
}

// ChatThread represents a unified conversation thread
type ChatThread struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Thread type
	Channel string `json:"channel" gorm:"not null"` // sms, mms, internal, web, room

	// For SMS/MMS threads
	LocalNumber  string `json:"local_number,omitempty" gorm:"index"`
	RemoteNumber string `json:"remote_number,omitempty" gorm:"index"`

	// For internal/room threads
	RoomID *uint `json:"room_id,omitempty" gorm:"index"`

	// For queue-based threads
	QueueID    *uint `json:"queue_id,omitempty" gorm:"index"`
	AssignedTo *uint `json:"assigned_to,omitempty" gorm:"index"` // Extension ID

	// Contact link (for external conversations)
	ContactID *uint `json:"contact_id,omitempty" gorm:"index"`

	// Participants (for internal chats)
	Participants pq.Int64Array `json:"participants" gorm:"type:bigint[]"` // Extension IDs

	// Status
	Status        string     `json:"status" gorm:"default:'open'"`     // open, pending, resolved, closed
	Priority      string     `json:"priority" gorm:"default:'normal'"` // low, normal, high, urgent
	LastMessageAt *time.Time `json:"last_message_at"`
	ResolvedAt    *time.Time `json:"resolved_at"`

	// Metadata
	Subject string         `json:"subject"`
	Tags    pq.StringArray `json:"tags" gorm:"type:text[]"`

	// Messages
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:ThreadID"`
}

func (t *ChatThread) BeforeCreate(tx *gorm.DB) error {
	t.UUID = uuid.New()
	return nil
}

// ChatMessage represents a single chat message
type ChatMessage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`
	ThreadID uint `json:"thread_id" gorm:"index;not null"`

	// Sender
	SenderType string `json:"sender_type" gorm:"not null"` // extension, contact, system, bot
	SenderID   uint   `json:"sender_id"`                   // Extension ID or Contact ID

	// Content (encrypted for SMS/external)
	ContentType   string `json:"content_type" gorm:"default:'text'"` // text, image, video, file, audio
	Body          string `json:"body" gorm:"-"`                      // Decrypted (not persisted)
	BodyEncrypted string `json:"-" gorm:"type:text"`
	BodyHash      string `json:"-" gorm:"index"`

	// Media attachments
	Attachments []ChatAttachment `json:"attachments,omitempty" gorm:"foreignKey:MessageID"`

	// Delivery status (for SMS)
	Status       string     `json:"status" gorm:"default:'sent'"` // pending, sent, delivered, read, failed
	DeliveredAt  *time.Time `json:"delivered_at"`
	ReadAt       *time.Time `json:"read_at"`
	FailedReason string     `json:"failed_reason,omitempty"`

	// Reply/Thread
	ReplyToID *uint `json:"reply_to_id,omitempty" gorm:"index"`

	// Reactions
	Reactions pq.StringArray `json:"reactions" gorm:"type:text[]"`

	// Provider info (for SMS/MMS)
	ProviderID string `json:"provider_id,omitempty"`
	ExternalID string `json:"external_id,omitempty"`
}

func (m *ChatMessage) BeforeCreate(tx *gorm.DB) error {
	m.UUID = uuid.New()
	return nil
}

// ChatAttachment represents media attached to messages
type ChatAttachment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	MessageID uint `json:"message_id" gorm:"index;not null"`
	TenantID  uint `json:"tenant_id" gorm:"index;not null"`

	// File info
	FileName    string `json:"file_name" gorm:"not null"`
	ContentType string `json:"content_type" gorm:"not null"`
	FileSize    int64  `json:"file_size"`

	// Storage (base64 or external)
	DataBase64      string `json:"-" gorm:"type:text"`
	ThumbnailBase64 string `json:"-" gorm:"type:text"`
	ExternalURL     string `json:"external_url,omitempty"`

	// Hash for deduplication
	DataHash string `json:"-" gorm:"index"`
}

// Contact represents a shared system contact
type Contact struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Basic info
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Company     string `json:"company"`
	Title       string `json:"title"`
	Avatar      string `json:"avatar"` // base64 or URL

	// Contact methods
	Email       string `json:"email" gorm:"index"`
	Phone       string `json:"phone" gorm:"index"`
	PhoneAlt    string `json:"phone_alt"`
	MobilePhone string `json:"mobile_phone"`

	// Address
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`

	// External integrations
	ExternalID     string `json:"external_id" gorm:"index"`       // ID from external system
	ExternalSource string `json:"external_source"`                // patient-portal, crm, etc.
	ExternalData   string `json:"external_data" gorm:"type:text"` // JSON blob

	// Webhook population
	LastSyncAt  *time.Time `json:"last_sync_at"`
	SyncEnabled bool       `json:"sync_enabled" gorm:"default:true"`
	WebhookURL  string     `json:"webhook_url,omitempty"`

	// Custom fields (JSON)
	CustomFields string `json:"custom_fields" gorm:"type:text"`

	// Tags & Groups
	Tags   pq.StringArray `json:"tags" gorm:"type:text[]"`
	Groups pq.StringArray `json:"groups" gorm:"type:text[]"`

	// Notes
	Notes string `json:"notes" gorm:"type:text"`

	// Preferences
	PreferredChannel  string `json:"preferred_channel"` // sms, email, phone
	DoNotContact      bool   `json:"do_not_contact" gorm:"default:false"`
	OptOutSMS         bool   `json:"opt_out_sms" gorm:"default:false"`
	OptOutEmail       bool   `json:"opt_out_email" gorm:"default:false"`
	PreferredLanguage string `json:"preferred_language" gorm:"default:'en'"`
	Timezone          string `json:"timezone"`

	// Status
	Status string `json:"status" gorm:"default:'active'"` // active, archived, blocked
}

func (c *Contact) BeforeCreate(tx *gorm.DB) error {
	c.UUID = uuid.New()
	return nil
}

// FullName returns the contact's full name
func (c *Contact) FullName() string {
	if c.DisplayName != "" {
		return c.DisplayName
	}
	if c.FirstName == "" && c.LastName == "" {
		return c.Phone
	}
	return c.FirstName + " " + c.LastName
}

// ContactWebhook defines a webhook source for contact data
type ContactWebhook struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Webhook info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Source      string `json:"source" gorm:"not null"` // patient-portal, salesforce, hubspot, custom

	// Endpoint for fetching data
	FetchURL     string `json:"fetch_url"`
	FetchMethod  string `json:"fetch_method" gorm:"default:'GET'"`
	FetchHeaders string `json:"fetch_headers" gorm:"type:text"` // JSON

	// Authentication
	AuthType   string `json:"auth_type"`          // none, basic, bearer, oauth2
	AuthConfig string `json:"-" gorm:"type:text"` // Encrypted JSON

	// Field mapping (JSON)
	FieldMapping string `json:"field_mapping" gorm:"type:text"`

	// Sync settings
	SyncInterval   int        `json:"sync_interval" gorm:"default:3600"` // seconds
	LastSyncAt     *time.Time `json:"last_sync_at"`
	LastSyncStatus string     `json:"last_sync_status"`

	Enabled bool `json:"enabled" gorm:"default:true"`
}

func (w *ContactWebhook) BeforeCreate(tx *gorm.DB) error {
	w.UUID = uuid.New()
	return nil
}

// ChatReadReceipt tracks who has read messages
type ChatReadReceipt struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	MessageID   uint      `json:"message_id" gorm:"index;not null"`
	ExtensionID uint      `json:"extension_id" gorm:"index;not null"`
	ReadAt      time.Time `json:"read_at"`
}

// ChatTypingIndicator for real-time typing status (not persisted, use WebSocket)
type ChatTypingIndicator struct {
	ThreadID    uint `json:"thread_id"`
	ExtensionID uint `json:"extension_id"`
	IsTyping    bool `json:"is_typing"`
}
