package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

func (p *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if p.Token == "" {
		p.Token = uuid.New().String()
	}
	return nil
}

func (p *PasswordResetToken) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

func (p *PasswordResetToken) IsUsed() bool {
	return p.UsedAt != nil
}
