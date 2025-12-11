package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ACL represents an Access Control List for network-based access control
type ACL struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// ACL identification
	Name        string `json:"name" gorm:"uniqueIndex;not null"` // e.g., "lan", "domains", "trunks"
	Description string `json:"description"`
	Default     string `json:"default" gorm:"default:'deny'"` // "allow" or "deny"
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// Nodes (entries in this ACL)
	Nodes []ACLNode `json:"nodes" gorm:"foreignKey:ACLUUID;references:UUID"`
}

// BeforeCreate generates UUID
func (a *ACL) BeforeCreate(tx *gorm.DB) error {
	a.UUID = uuid.New()
	return nil
}

// ACLNode represents an individual entry in an ACL (IP/CIDR allow/deny rule)
type ACLNode struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	ACLUUID uuid.UUID `json:"acl_uuid" gorm:"type:uuid;index;not null"`

	// Node details
	Type        string `json:"type" gorm:"default:'allow'"` // "allow" or "deny"
	CIDR        string `json:"cidr"`                        // IP address or CIDR block (e.g., "192.168.1.0/24")
	Domain      string `json:"domain"`                      // Domain name (alternative to CIDR)
	Description string `json:"description"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`
	Priority    int    `json:"priority" gorm:"default:100"` // Lower = higher priority
}

// DefaultACLs returns default ACL lists for a new installation
func DefaultACLs() []ACL {
	return []ACL{
		{
			Name:        "lan",
			Description: "Local Area Network (RFC1918 addresses)",
			Default:     "allow",
			Enabled:     true,
			Nodes: []ACLNode{
				{Type: "allow", CIDR: "192.168.0.0/16", Description: "Class C Private", Enabled: true, Priority: 10},
				{Type: "allow", CIDR: "10.0.0.0/8", Description: "Class A Private", Enabled: true, Priority: 20},
				{Type: "allow", CIDR: "172.16.0.0/12", Description: "Class B Private", Enabled: true, Priority: 30},
			},
		},
		{
			Name:        "loopback",
			Description: "Loopback addresses",
			Default:     "allow",
			Enabled:     true,
			Nodes: []ACLNode{
				{Type: "allow", CIDR: "127.0.0.0/8", Description: "IPv4 Loopback", Enabled: true, Priority: 10},
				{Type: "allow", CIDR: "::1/128", Description: "IPv6 Loopback", Enabled: true, Priority: 20},
			},
		},
		{
			Name:        "nat",
			Description: "NAT addresses",
			Default:     "allow",
			Enabled:     true,
			Nodes:       []ACLNode{},
		},
		{
			Name:        "domains",
			Description: "Trusted domains for SIP registrations",
			Default:     "deny",
			Enabled:     true,
			Nodes:       []ACLNode{},
		},
		{
			Name:        "trunks",
			Description: "Trusted SIP trunk provider IPs",
			Default:     "deny",
			Enabled:     true,
			Nodes:       []ACLNode{},
		},
	}
}
