package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CallBlock represents a blocked caller for a tenant
type CallBlock struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Tenant association
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`

	// Block configuration
	Number    string `json:"number" gorm:"not null"`            // E.164 format
	MatchType string `json:"match_type" gorm:"default:'exact'"` // exact, prefix, regex
	Action    string `json:"action" gorm:"default:'reject'"`    // reject, busy, hangup

	// Status
	Enabled bool   `json:"enabled" gorm:"default:true"`
	Notes   string `json:"notes"`

	// Associated auto-generated dialplan
	DialplanUUID *uuid.UUID `json:"dialplan_uuid" gorm:"type:uuid"`
}

// BeforeCreate generates UUID
func (cb *CallBlock) BeforeCreate(tx *gorm.DB) error {
	cb.UUID = uuid.New()
	return nil
}
