package models

import (
	"time"

	"gorm.io/gorm"
)

// HotelRoom represents a room in a hospitality environment
type HotelRoom struct {
	gorm.Model
	TenantID      uint       `json:"tenant_id" gorm:"index"`
	RoomNumber    string     `json:"room_number" gorm:"index"`
	RoomType      string     `json:"room_type"` // single, double, suite
	Extension     string     `json:"extension"` // room phone extension
	ExtensionID   *uint      `json:"extension_id"`
	Floor         string     `json:"floor"`
	Status        string     `json:"status" gorm:"default:'vacant'"` // vacant, occupied, maintenance
	GuestName     string     `json:"guest_name"`
	CheckInTime   *time.Time `json:"check_in_time"`
	CheckOutTime  *time.Time `json:"check_out_time"`
	WakeupTime    *time.Time `json:"wakeup_time"`
	WakeupEnabled bool       `json:"wakeup_enabled"`
	DNDEnabled    bool       `json:"dnd_enabled"`
	CIDName       string     `json:"cid_name"`   // outbound caller ID name
	CIDNumber     string     `json:"cid_number"` // outbound caller ID number
	Notes         string     `json:"notes"`
	Enabled       bool       `json:"enabled" gorm:"default:true"`
}
