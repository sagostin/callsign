package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditAction defines audit log action types
type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionRead   AuditAction = "read"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionLogin  AuditAction = "login"
	AuditActionLogout AuditAction = "logout"
	AuditActionExport AuditAction = "export"
	AuditActionImport AuditAction = "import"
	AuditActionCall   AuditAction = "call"
	AuditActionConfig AuditAction = "config"
)

// AuditLog represents a global audit log entry
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	// Who
	TenantID  uint   `json:"tenant_id" gorm:"index"`
	UserID    uint   `json:"user_id" gorm:"index"`
	Username  string `json:"username"`
	UserRole  string `json:"user_role"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`

	// What
	Action     AuditAction `json:"action" gorm:"index"`
	Resource   string      `json:"resource" gorm:"index"` // user, extension, conference, etc.
	ResourceID string      `json:"resource_id"`           // ID of affected resource

	// Details
	OldValue json.RawMessage `json:"old_value,omitempty" gorm:"type:jsonb"`
	NewValue json.RawMessage `json:"new_value,omitempty" gorm:"type:jsonb"`
	Metadata json.RawMessage `json:"metadata,omitempty" gorm:"type:jsonb"` // Additional context

	// Result
	Success bool   `json:"success" gorm:"default:true"`
	Error   string `json:"error,omitempty"`
}

// BeforeCreate generates UUID
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	a.UUID = uuid.New()
	return nil
}

// AuditLogEntry is a helper for creating audit entries
type AuditLogEntry struct {
	DB         *gorm.DB
	TenantID   uint
	UserID     uint
	Username   string
	UserRole   string
	IPAddress  string
	UserAgent  string
	Action     AuditAction
	Resource   string
	ResourceID string
}

// Log creates an audit log entry
func (e *AuditLogEntry) Log(success bool, oldVal, newVal interface{}, err error) error {
	log := &AuditLog{
		TenantID:   e.TenantID,
		UserID:     e.UserID,
		Username:   e.Username,
		UserRole:   e.UserRole,
		IPAddress:  e.IPAddress,
		UserAgent:  e.UserAgent,
		Action:     e.Action,
		Resource:   e.Resource,
		ResourceID: e.ResourceID,
		Success:    success,
	}

	if oldVal != nil {
		if data, err := json.Marshal(oldVal); err == nil {
			log.OldValue = data
		}
	}
	if newVal != nil {
		if data, err := json.Marshal(newVal); err == nil {
			log.NewValue = data
		}
	}
	if err != nil {
		log.Error = err.Error()
	}

	return e.DB.Create(log).Error
}

// CreateAuditLog is a convenience function to create audit logs
func CreateAuditLog(db *gorm.DB, tenantID, userID uint, username, role, ip, ua string,
	action AuditAction, resource, resourceID string, success bool, oldVal, newVal interface{}) error {

	entry := &AuditLogEntry{
		DB:         db,
		TenantID:   tenantID,
		UserID:     userID,
		Username:   username,
		UserRole:   role,
		IPAddress:  ip,
		UserAgent:  ua,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
	}
	return entry.Log(success, oldVal, newVal, nil)
}
