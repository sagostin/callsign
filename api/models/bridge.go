package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Bridge struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID  *uint   `json:"tenant_id" gorm:"index"`
	Tenant    *Tenant `json:"-" gorm:"foreignKey:TenantID"`
	Name      string  `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Enabled   bool    `json:"enabled" gorm:"default:true"`

	BridgeType string `json:"bridge_type" gorm:"default:'sip'"` // sip, webrtc, external
	Host       string `json:"host" gorm:"not null"`             // Primary host
	BackupHost string `json:"backup_host"`                      // Backup host
	Port       int    `json:"port" gorm:"default:5060"`
	Transport  string `json:"transport" gorm:"default:'udp'"`   // udp, tcp, tls

	Username string `json:"username"`
	Password string `json:"-"`
	AuthUser string `json:"auth_user"`

	FromUser   string `json:"from_user"`
	FromDomain string `json:"from_domain"`
	Extension  string `json:"extension"`

	Codec      string `json:"codec"`
	Context    string `json:"context" gorm:"default:'default'"`
	Expires    int    `json:"expires" gorm:"default:3600"`

	Status     string     `json:"status" gorm:"-"`
	LastStatus *time.Time `json:"last_status"`
}

func (b *Bridge) BeforeCreate(tx *gorm.DB) error {
	b.UUID = uuid.New()
	return nil
}