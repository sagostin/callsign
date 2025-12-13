package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SIPProfile represents a FreeSWITCH Sofia SIP profile (internal, external, etc.)
type SIPProfile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Profile identification
	ProfileName string `json:"profile_name" gorm:"uniqueIndex;not null"` // e.g., "internal", "external"
	Alias       string `json:"alias"`                                    // Friendly display name, e.g., "Primary Internal", "Trunk Gateway"
	Description string `json:"description"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// Usage type determines context and auth behavior
	// - "internal" = registrations from extensions, uses tenant context, auth required
	// - "trunks" = system gateways only, public context, no auth (inbound from providers)
	// - "mixed" = both (default for backwards compatibility)
	UsageType string `json:"usage_type" gorm:"default:'mixed'"` // internal, trunks, mixed

	// Hostname filter (null = all hosts)
	Hostname string `json:"hostname"`

	// Settings relationship
	Settings []SIPProfileSetting `json:"settings" gorm:"foreignKey:SIPProfileUUID;references:UUID"`

	// Domains relationship
	Domains []SIPProfileDomain `json:"domains" gorm:"foreignKey:SIPProfileUUID;references:UUID"`
}

// BeforeCreate generates UUID
func (p *SIPProfile) BeforeCreate(tx *gorm.DB) error {
	p.UUID = uuid.New()
	return nil
}

// SIPProfileSetting represents individual settings for a SIP profile
type SIPProfileSetting struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	SIPProfileUUID uuid.UUID `json:"sip_profile_uuid" gorm:"type:uuid;index"`
	SettingName    string    `json:"setting_name" gorm:"not null"`
	SettingValue   string    `json:"setting_value"`
	Enabled        bool      `json:"enabled" gorm:"default:true"`
}

// SIPProfileDomain represents domains associated with a SIP profile
type SIPProfileDomain struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	SIPProfileUUID uuid.UUID `json:"sip_profile_uuid" gorm:"type:uuid;index"`
	DomainName     string    `json:"domain_name" gorm:"not null"` // e.g., "all" or specific domain
	Alias          bool      `json:"alias" gorm:"default:true"`
	Parse          bool      `json:"parse" gorm:"default:true"`
}

// SofiaGlobalSetting represents global Sofia settings
type SofiaGlobalSetting struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	SettingName  string `json:"setting_name" gorm:"uniqueIndex;not null"`
	SettingValue string `json:"setting_value"`
	Enabled      bool   `json:"enabled" gorm:"default:true"`
	Description  string `json:"description"`
}

// DefaultSIPProfileSettings returns common default settings for a new internal profile
func DefaultInternalProfileSettings() []SIPProfileSetting {
	return []SIPProfileSetting{
		{SettingName: "context", SettingValue: "public", Enabled: true},
		{SettingName: "rtp-timer-name", SettingValue: "soft", Enabled: true},
		{SettingName: "sip-ip", SettingValue: "$${local_ip_v4}", Enabled: true},
		{SettingName: "sip-port", SettingValue: "5060", Enabled: true},
		{SettingName: "rtp-ip", SettingValue: "$${local_ip_v4}", Enabled: true},
		{SettingName: "codec-prefs", SettingValue: "OPUS,G722,PCMU,PCMA,VP8", Enabled: true},
		{SettingName: "inbound-codec-negotiation", SettingValue: "generous", Enabled: true},
		{SettingName: "manage-presence", SettingValue: "true", Enabled: true},
		{SettingName: "auth-calls", SettingValue: "true", Enabled: true},
		{SettingName: "inbound-reg-force-matching-username", SettingValue: "true", Enabled: true},
		{SettingName: "nonce-ttl", SettingValue: "60", Enabled: true},
		{SettingName: "rtp-timeout-sec", SettingValue: "300", Enabled: true},
		{SettingName: "rtp-hold-timeout-sec", SettingValue: "1800", Enabled: true},
	}
}

// DefaultExternalProfileSettings returns common default settings for an external profile
func DefaultExternalProfileSettings() []SIPProfileSetting {
	return []SIPProfileSetting{
		{SettingName: "context", SettingValue: "public", Enabled: true},
		{SettingName: "rtp-timer-name", SettingValue: "soft", Enabled: true},
		{SettingName: "sip-ip", SettingValue: "$${local_ip_v4}", Enabled: true},
		{SettingName: "sip-port", SettingValue: "5080", Enabled: true},
		{SettingName: "rtp-ip", SettingValue: "$${local_ip_v4}", Enabled: true},
		{SettingName: "codec-prefs", SettingValue: "OPUS,G722,PCMU,PCMA", Enabled: true},
		{SettingName: "inbound-codec-negotiation", SettingValue: "generous", Enabled: true},
		{SettingName: "auth-calls", SettingValue: "false", Enabled: true},
	}
}
