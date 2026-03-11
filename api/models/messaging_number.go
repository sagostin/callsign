package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessagingNumber maps a phone number to a messaging provider (for outbound)
// and optionally to a tenant/extension (for inbound routing).
// Inbound routing is carrier-agnostic — the number determines the destination
// regardless of which carrier delivered the message.
// Outbound uses the ProviderID to decide which carrier to send through.
type MessagingNumber struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// The phone number in E.164 format (e.g. +14155551234)
	PhoneNumber string `json:"phone_number" gorm:"uniqueIndex;not null"`

	// Outbound carrier — which provider to use when sending FROM this number
	ProviderID uint               `json:"provider_id" gorm:"index;not null"`
	Provider   *MessagingProvider `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`

	// Inbound routing target (carrier-agnostic)
	TenantID    *uint `json:"tenant_id" gorm:"index"`
	ExtensionID *uint `json:"extension_id" gorm:"index"`

	// Number capabilities
	SMSEnabled   bool `json:"sms_enabled" gorm:"default:true"`
	MMSEnabled   bool `json:"mms_enabled" gorm:"default:false"`
	VoiceEnabled bool `json:"voice_enabled" gorm:"default:false"`

	// Display info
	FriendlyName string `json:"friendly_name"`
	Description  string `json:"description"`

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (n *MessagingNumber) BeforeCreate(tx *gorm.DB) error {
	n.UUID = uuid.New()
	return nil
}
