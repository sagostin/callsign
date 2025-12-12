package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BannedIP stores IPs that have been banned for security reasons
type BannedIP struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// IP address that was banned
	IP string `json:"ip" gorm:"index;not null"`

	// Source of the ban (e.g., "freeswitch-callsign", "manual")
	Source string `json:"source" gorm:"index;default:'fail2ban'"`

	// Reason for the ban
	Reason string `json:"reason"`

	// Count of failures before ban
	Failures int `json:"failures"`

	// When the IP was banned
	BannedAt time.Time `json:"banned_at"`

	// When the ban expires (null = permanent)
	ExpiresAt *time.Time `json:"expires_at"`

	// Status: banned, unbanned
	Status string `json:"status" gorm:"default:'banned'"`

	// Tenant-specific tracking (optional - for attacks targeting specific tenants)
	TenantID   *uint  `json:"tenant_id" gorm:"index"`
	Domain     string `json:"domain" gorm:"index"` // Domain being attacked
	Extension  string `json:"extension"`           // Extension/user being targeted
	UserAgent  string `json:"user_agent"`          // User-Agent from attacker
	TargetType string `json:"target_type"`         // "sip", "web", "api", "provisioning"
}

// BeforeCreate generates UUID
func (b *BannedIP) BeforeCreate(tx *gorm.DB) error {
	b.UUID = uuid.New()
	if b.BannedAt.IsZero() {
		b.BannedAt = time.Now()
	}
	return nil
}
