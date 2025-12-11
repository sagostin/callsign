package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRole defines the role types for users
type UserRole string

const (
	RoleSystemAdmin UserRole = "system_admin"
	RoleTenantAdmin UserRole = "tenant_admin"
	RoleUser        UserRole = "user"
)

// User represents a system user
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Authentication
	Username string `json:"username" gorm:"uniqueIndex;not null"`
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"` // Never expose password in JSON

	// Role and permissions
	Role        UserRole `json:"role" gorm:"type:varchar(50);default:'user'"`
	Permissions string   `json:"permissions,omitempty" gorm:"type:text"` // Comma-separated permissions

	// Tenant association (null for system admins)
	TenantID *uint   `json:"tenant_id" gorm:"index"`
	Tenant   *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`

	// API access
	APIKey    string     `json:"-" gorm:"uniqueIndex"`
	LastLogin *time.Time `json:"last_login"`

	// Profile
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Extension   string `json:"extension"`    // Associated extension number
	ExtensionID *uint  `json:"extension_id"` // Link to Extension model
}

// BeforeCreate generates UUID and hashes password before creating user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.UUID = uuid.New()

	// Generate API key
	u.APIKey = uuid.New().String()

	return nil
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies a password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsSystemAdmin checks if the user has system admin privileges
func (u *User) IsSystemAdmin() bool {
	return u.Role == RoleSystemAdmin
}

// IsTenantAdmin checks if the user has tenant admin privileges
func (u *User) IsTenantAdmin() bool {
	return u.Role == RoleTenantAdmin || u.Role == RoleSystemAdmin
}

// CanAccessTenant checks if a user can access a specific tenant
func (u *User) CanAccessTenant(tenantID uint) bool {
	if u.IsSystemAdmin() {
		return true
	}
	return u.TenantID != nil && *u.TenantID == tenantID
}

// HasCustomPermission checks if user has a permission in their custom Permissions field
// This is for user-specific permission overrides, not role-based permissions
func (u *User) HasCustomPermission(permission string) bool {
	if u.Permissions == "" {
		return false
	}
	perms := strings.Split(u.Permissions, ",")
	for _, p := range perms {
		if strings.TrimSpace(p) == permission || strings.TrimSpace(p) == "*" {
			return true
		}
	}
	return false
}
