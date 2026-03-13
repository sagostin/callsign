package models

import (
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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

// DefaultExternalProfileSettings returns the FusionPBX-style default settings for the external profile.
// Uses $${variable} notation for FreeSWITCH global variables.
func DefaultExternalProfileSettings() []SIPProfileSetting {
	return []SIPProfileSetting{
		// --- Disabled settings (present but not active, matching FusionPBX) ---
		{SettingName: "aggressive-nat-detection", SettingValue: "false", Enabled: false},
		{SettingName: "apply-register-acl", SettingValue: "providers", Enabled: false},
		{SettingName: "dbname", SettingValue: "share_presence", Enabled: false},
		{SettingName: "disable-srv503", SettingValue: "true", Enabled: false},
		{SettingName: "enable-100rel", SettingValue: "true", Enabled: false},
		{SettingName: "enable-3pcc", SettingValue: "true", Enabled: false},
		{SettingName: "enable-rfc-5626", SettingValue: "true", Enabled: false},
		{SettingName: "force-register-db-domain", SettingValue: "$${domain}", Enabled: false},
		{SettingName: "force-register-domain", SettingValue: "$${domain}", Enabled: false},
		{SettingName: "inbound-late-negotiation", SettingValue: "true", Enabled: false},
		{SettingName: "minimum-session-expires", SettingValue: "0", Enabled: false},
		{SettingName: "odbc-dsn", SettingValue: "$${dsn}", Enabled: false},
		{SettingName: "presence-hosts", SettingValue: "$${domain}", Enabled: false},
		{SettingName: "rtp-hold-timeout-sec", SettingValue: "1800", Enabled: false},
		{SettingName: "rtp-timeout-sec", SettingValue: "300", Enabled: false},
		{SettingName: "shutdown-on-fail", SettingValue: "true", Enabled: false},
		{SettingName: "tls-ciphers", SettingValue: "$${sip_tls_ciphers}", Enabled: false},

		// --- Enabled settings ---
		{SettingName: "apply-inbound-acl", SettingValue: "providers", Enabled: true},
		{SettingName: "apply-nat-acl", SettingValue: "nat.auto", Enabled: true},
		{SettingName: "auth-calls", SettingValue: "false", Enabled: true},
		{SettingName: "auth-subscriptions", SettingValue: "false", Enabled: true},
		{SettingName: "context", SettingValue: "public", Enabled: true},
		{SettingName: "debug", SettingValue: "0", Enabled: true},
		{SettingName: "dialplan", SettingValue: "XML", Enabled: true},
		{SettingName: "dtmf-duration", SettingValue: "2000", Enabled: true},
		{SettingName: "dtmf-type", SettingValue: "rfc2833", Enabled: true},
		{SettingName: "enable-timer", SettingValue: "false", Enabled: true},
		{SettingName: "ext-rtp-ip", SettingValue: "$${external_rtp_ip}", Enabled: true},
		{SettingName: "ext-sip-ip", SettingValue: "$${external_sip_ip}", Enabled: true},
		{SettingName: "hold-music", SettingValue: "$${hold_music}", Enabled: true},
		{SettingName: "inbound-codec-negotiation", SettingValue: "generous", Enabled: true},
		{SettingName: "inbound-codec-prefs", SettingValue: "$${global_codec_prefs}", Enabled: true},
		{SettingName: "local-network-acl", SettingValue: "localnet.auto", Enabled: true},
		{SettingName: "manage-presence", SettingValue: "true", Enabled: true},
		{SettingName: "media_hold_timeout", SettingValue: "1800", Enabled: true},
		{SettingName: "media_timeout", SettingValue: "300", Enabled: true},
		{SettingName: "nonce-ttl", SettingValue: "60", Enabled: true},
		{SettingName: "outbound-codec-prefs", SettingValue: "$${outbound_codec_prefs}", Enabled: true},
		{SettingName: "rfc2833-pt", SettingValue: "101", Enabled: true},
		{SettingName: "rtp-ip", SettingValue: "$${local_ip_v4}", Enabled: true},
		{SettingName: "rtp-timer-name", SettingValue: "soft", Enabled: true},
		{SettingName: "session-timeout", SettingValue: "0", Enabled: true},
		{SettingName: "sip-capture", SettingValue: "yes", Enabled: true},
		{SettingName: "sip-ip", SettingValue: "$${local_ip_v4}", Enabled: true},
		{SettingName: "sip-port", SettingValue: "5080", Enabled: true},
		{SettingName: "sip-trace", SettingValue: "no", Enabled: true},
		{SettingName: "suppress-cng", SettingValue: "true", Enabled: true},
		{SettingName: "tls", SettingValue: "$${external_ssl_enable}", Enabled: true},
		{SettingName: "tls-bind-params", SettingValue: "transport=tls", Enabled: true},
		{SettingName: "tls-cert-dir", SettingValue: "$${external_ssl_dir}", Enabled: true},
		{SettingName: "tls-only", SettingValue: "false", Enabled: true},
		{SettingName: "tls-passphrase", SettingValue: "", Enabled: true},
		{SettingName: "tls-sip-port", SettingValue: "5081", Enabled: true},
		{SettingName: "tls-verify-date", SettingValue: "false", Enabled: true},
		{SettingName: "tls-verify-depth", SettingValue: "2", Enabled: true},
		{SettingName: "tls-verify-in-subjects", SettingValue: "", Enabled: true},
		{SettingName: "tls-verify-policy", SettingValue: "none", Enabled: true},
		{SettingName: "tls-version", SettingValue: "$${sip_tls_version}", Enabled: true},
		{SettingName: "track-calls", SettingValue: "false", Enabled: true},
		{SettingName: "user-agent-string", SettingValue: "FreeSWITCH", Enabled: true},
		{SettingName: "zrtp-passthru", SettingValue: "true", Enabled: true},
	}
}

// DefaultInternalProfileSettings returns the FusionPBX-style default settings for the internal profile.
// Uses $${variable} notation for FreeSWITCH global variables.
func DefaultInternalProfileSettings() []SIPProfileSetting {
	return []SIPProfileSetting{
		// --- Disabled settings (present but not active, matching FusionPBX) ---
		{SettingName: "accept-blind-auth", SettingValue: "true", Enabled: false},
		{SettingName: "accept-blind-reg", SettingValue: "true", Enabled: false},
		{SettingName: "aggressive-nat-detection", SettingValue: "true", Enabled: false},
		{SettingName: "apply-inbound-acl", SettingValue: "domains", Enabled: false},
		{SettingName: "dbname", SettingValue: "share_presence", Enabled: false},
		{SettingName: "disable-srv503", SettingValue: "true", Enabled: false},
		{SettingName: "enable-100rel", SettingValue: "true", Enabled: false},
		{SettingName: "enable-3pcc", SettingValue: "true", Enabled: false},
		{SettingName: "enable-rfc-5626", SettingValue: "true", Enabled: false},
		{SettingName: "force-subscription-domain", SettingValue: "$${domain}", Enabled: false},
		{SettingName: "inbound-late-negotiation", SettingValue: "true", Enabled: false},
		{SettingName: "minimum-session-expires", SettingValue: "0", Enabled: false},
		{SettingName: "odbc-dsn", SettingValue: "$${dsn}", Enabled: false},
		{SettingName: "rtp-hold-timeout-sec", SettingValue: "1800", Enabled: false},
		{SettingName: "rtp-timeout-sec", SettingValue: "300", Enabled: false},
		{SettingName: "shutdown-on-fail", SettingValue: "true", Enabled: false},
		{SettingName: "tls-ciphers", SettingValue: "$${sip_tls_ciphers}", Enabled: false},

		// --- Enabled settings ---
		{SettingName: "apply-nat-acl", SettingValue: "nat.auto", Enabled: true},
		{SettingName: "auth-calls", SettingValue: "true", Enabled: true},
		{SettingName: "auth-subscriptions", SettingValue: "true", Enabled: true},
		{SettingName: "context", SettingValue: "public", Enabled: true},
		{SettingName: "debug", SettingValue: "0", Enabled: true},
		{SettingName: "dialplan", SettingValue: "XML", Enabled: true},
		{SettingName: "dtmf-duration", SettingValue: "2000", Enabled: true},
		{SettingName: "dtmf-type", SettingValue: "rfc2833", Enabled: true},
		{SettingName: "enable-timer", SettingValue: "false", Enabled: true},
		{SettingName: "ext-rtp-ip", SettingValue: "$${external_rtp_ip}", Enabled: true},
		{SettingName: "ext-sip-ip", SettingValue: "$${external_sip_ip}", Enabled: true},
		{SettingName: "force-register-db-domain", SettingValue: "$${domain}", Enabled: true},
		{SettingName: "force-register-domain", SettingValue: "$${domain}", Enabled: true},
		{SettingName: "hold-music", SettingValue: "$${hold_music}", Enabled: true},
		{SettingName: "inbound-codec-negotiation", SettingValue: "generous", Enabled: true},
		{SettingName: "inbound-codec-prefs", SettingValue: "$${global_codec_prefs}", Enabled: true},
		{SettingName: "inbound-reg-force-matching-username", SettingValue: "true", Enabled: true},
		{SettingName: "local-network-acl", SettingValue: "localnet.auto", Enabled: true},
		{SettingName: "manage-presence", SettingValue: "true", Enabled: true},
		{SettingName: "media_hold_timeout", SettingValue: "1800", Enabled: true},
		{SettingName: "media_timeout", SettingValue: "300", Enabled: true},
		{SettingName: "nonce-ttl", SettingValue: "60", Enabled: true},
		{SettingName: "outbound-codec-prefs", SettingValue: "$${outbound_codec_prefs}", Enabled: true},
		{SettingName: "presence-hosts", SettingValue: "$${domain}", Enabled: true},
		{SettingName: "rfc2833-pt", SettingValue: "101", Enabled: true},
		{SettingName: "rtp-ip", SettingValue: "$${local_ip_v4}", Enabled: true},
		{SettingName: "rtp-timer-name", SettingValue: "soft", Enabled: true},
		{SettingName: "session-timeout", SettingValue: "0", Enabled: true},
		{SettingName: "sip-capture", SettingValue: "yes", Enabled: true},
		{SettingName: "sip-ip", SettingValue: "$${local_ip_v4}", Enabled: true},
		{SettingName: "sip-port", SettingValue: "5060", Enabled: true},
		{SettingName: "sip-trace", SettingValue: "no", Enabled: true},
		{SettingName: "suppress-cng", SettingValue: "true", Enabled: true},
		{SettingName: "tls", SettingValue: "$${internal_ssl_enable}", Enabled: true},
		{SettingName: "tls-bind-params", SettingValue: "transport=tls", Enabled: true},
		{SettingName: "tls-cert-dir", SettingValue: "$${internal_ssl_dir}", Enabled: true},
		{SettingName: "tls-only", SettingValue: "false", Enabled: true},
		{SettingName: "tls-passphrase", SettingValue: "", Enabled: true},
		{SettingName: "tls-sip-port", SettingValue: "5061", Enabled: true},
		{SettingName: "tls-verify-date", SettingValue: "false", Enabled: true},
		{SettingName: "tls-verify-depth", SettingValue: "2", Enabled: true},
		{SettingName: "tls-verify-in-subjects", SettingValue: "", Enabled: true},
		{SettingName: "tls-verify-policy", SettingValue: "none", Enabled: true},
		{SettingName: "tls-version", SettingValue: "$${sip_tls_version}", Enabled: true},
		{SettingName: "track-calls", SettingValue: "false", Enabled: true},
		{SettingName: "user-agent-string", SettingValue: "FreeSWITCH", Enabled: true},
		{SettingName: "zrtp-passthru", SettingValue: "true", Enabled: true},
		// WebRTC / WebSocket support
		{SettingName: "ws-binding", SettingValue: ":5066", Enabled: false},
		{SettingName: "wss-binding", SettingValue: ":7443", Enabled: true},
		{SettingName: "apply-candidate-acl", SettingValue: "localnet.auto", Enabled: true},
	}
}

// EnsureDefaultProfiles creates the default "internal" and "external" SIP profiles
// in the database if they don't already exist, using FusionPBX-style defaults.
// This replaces the old disk-import approach — profiles are seeded from code.
func EnsureDefaultProfiles(db *gorm.DB) error {
	profiles := []struct {
		Name        string
		Description string
		UsageType   string
		Settings    func() []SIPProfileSetting
	}{
		{
			Name:        "internal",
			Description: "Internal SIP profile for extension registrations",
			UsageType:   "internal",
			Settings:    DefaultInternalProfileSettings,
		},
		{
			Name:        "external",
			Description: "External SIP profile for trunk/provider connections",
			UsageType:   "trunks",
			Settings:    DefaultExternalProfileSettings,
		},
	}

	for _, p := range profiles {
		var existing SIPProfile
		err := db.Where("profile_name = ?", p.Name).First(&existing).Error
		if err == nil {
			log.WithField("profile", p.Name).Debug("SIP profile already exists, skipping seed")
			continue
		}

		if err != gorm.ErrRecordNotFound {
			return err
		}

		// Also check for soft-deleted profile and permanently remove it
		var softDeleted SIPProfile
		if err := db.Unscoped().Where("profile_name = ? AND deleted_at IS NOT NULL", p.Name).First(&softDeleted).Error; err == nil {
			db.Unscoped().Where("sip_profile_uuid = ?", softDeleted.UUID).Delete(&SIPProfileSetting{})
			db.Unscoped().Where("sip_profile_uuid = ?", softDeleted.UUID).Delete(&SIPProfileDomain{})
			db.Unscoped().Delete(&softDeleted)
		}

		// Create profile
		profile := SIPProfile{
			ProfileName: p.Name,
			Description: p.Description,
			UsageType:   p.UsageType,
			Enabled:     true,
		}

		if err := db.Omit("Settings", "Domains").Create(&profile).Error; err != nil {
			log.WithError(err).WithField("profile", p.Name).Error("Failed to create default SIP profile")
			return err
		}

		// Create settings
		settings := p.Settings()
		for i := range settings {
			settings[i].SIPProfileUUID = profile.UUID
		}
		if err := db.Create(&settings).Error; err != nil {
			log.WithError(err).WithField("profile", p.Name).Error("Failed to create default SIP profile settings")
			return err
		}

		// Create default domain (all, with alias+parse for dynamic gateway loading)
		domain := SIPProfileDomain{
			SIPProfileUUID: profile.UUID,
			DomainName:     "all",
			Alias:          true,
			Parse:          true,
		}
		if err := db.Create(&domain).Error; err != nil {
			log.WithError(err).WithField("profile", p.Name).Error("Failed to create default SIP profile domain")
			return err
		}

		log.WithFields(log.Fields{
			"profile":  p.Name,
			"settings": len(settings),
		}).Info("Created default SIP profile from built-in defaults")
	}

	return nil
}
