package models

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FeatureCodeAction defines the action type for a feature code
type FeatureCodeAction string

const (
	FCActionVoicemail      FeatureCodeAction = "voicemail"
	FCActionCallForward    FeatureCodeAction = "call_forward"
	FCActionDND            FeatureCodeAction = "dnd"
	FCActionCallFlowToggle FeatureCodeAction = "call_flow_toggle"
	FCActionTransfer       FeatureCodeAction = "transfer"
	FCActionPark           FeatureCodeAction = "park"
	FCActionParkSlot       FeatureCodeAction = "park_slot"     // Park to specific slot
	FCActionParkRetrieve   FeatureCodeAction = "park_retrieve" // Retrieve from specific slot
	FCActionPickup         FeatureCodeAction = "pickup"
	FCActionIntercom       FeatureCodeAction = "intercom"
	FCActionPageGroup      FeatureCodeAction = "page_group"
	FCActionSpeedDial      FeatureCodeAction = "speed_dial"
	FCActionRecord         FeatureCodeAction = "record"
	FCActionCustom         FeatureCodeAction = "custom"
	FCActionWebhook        FeatureCodeAction = "webhook"
	FCActionLua            FeatureCodeAction = "lua"
)

// FeatureCode represents a configurable feature code
// These can be global (system_admin only) or tenant-specific
type FeatureCode struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership - null TenantID means global (system admin only)
	TenantID *uint `json:"tenant_id" gorm:"index"`

	// Code definition
	Code        string            `json:"code" gorm:"index;not null"` // e.g., "*55", "*70"
	CodeRegex   string            `json:"code_regex,omitempty"`       // Regex pattern, e.g., "^\*70(\d+)$" for *70[slot]
	Name        string            `json:"name" gorm:"not null"`
	Description string            `json:"description,omitempty"`
	Action      FeatureCodeAction `json:"action" gorm:"not null"`

	// Optional extension alias (e.g., ext 700 for park)
	Extension string `json:"extension,omitempty" gorm:"index"`

	// Ordering for dialplan matching (lower = higher priority)
	Order int `json:"order" gorm:"default:100"`

	// System code flag (managed by system admin only)
	IsGlobal bool `json:"is_global" gorm:"default:false"`

	// Context for dialplan
	Context string `json:"context" gorm:"default:'default'"`

	// Action data (depends on action type)
	ActionData   string `json:"action_data,omitempty"`   // JSON or dial string
	ActionParams string `json:"action_params,omitempty"` // Additional params JSON

	// For park action
	ParkLotName  string `json:"park_lot_name,omitempty"`           // Named parking lot
	ParkTimeout  int    `json:"park_timeout,omitempty"`            // Timeout in seconds
	ParkAnnounce bool   `json:"park_announce" gorm:"default:true"` // Announce slot number

	// For webhook action
	WebhookURL    string `json:"webhook_url,omitempty"`
	WebhookMethod string `json:"webhook_method,omitempty"` // GET, POST

	// For custom/lua action
	LuaScript string `json:"lua_script,omitempty"`

	// For transfer action
	TransferDest    string `json:"transfer_dest,omitempty"`
	TransferContext string `json:"transfer_context,omitempty"`

	// For group actions (pickup, page)
	GroupID *uint `json:"group_id,omitempty"`

	// Requires PIN?
	RequirePIN bool   `json:"require_pin" gorm:"default:false"`
	PIN        string `json:"-"` // Never expose

	// BLF hint for presence
	BLFHint string `json:"blf_hint,omitempty"` // e.g., "park+*70${slot}@${domain}"

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID and validates
func (fc *FeatureCode) BeforeCreate(tx *gorm.DB) error {
	fc.UUID = uuid.New()
	return fc.Validate()
}

// BeforeUpdate validates before update
func (fc *FeatureCode) BeforeUpdate(tx *gorm.DB) error {
	return fc.Validate()
}

// Validate checks feature code validity
func (fc *FeatureCode) Validate() error {
	// Validate code format (must start with * or #, or be a regex)
	if fc.CodeRegex == "" {
		if len(fc.Code) < 2 {
			return errors.New("feature code must be at least 2 characters")
		}
		if fc.Code[0] != '*' && fc.Code[0] != '#' {
			return errors.New("feature code must start with * or #")
		}
	}

	// Validate regex if provided
	if fc.CodeRegex != "" {
		if _, err := regexp.Compile(fc.CodeRegex); err != nil {
			return errors.New("invalid code_regex pattern: " + err.Error())
		}
	}

	return nil
}

// MatchesDialedNumber checks if a dialed number matches this feature code
func (fc *FeatureCode) MatchesDialedNumber(dialed string) (bool, map[string]string) {
	captures := make(map[string]string)

	// First try regex match
	if fc.CodeRegex != "" {
		re, err := regexp.Compile(fc.CodeRegex)
		if err != nil {
			return false, nil
		}
		matches := re.FindStringSubmatch(dialed)
		if matches != nil {
			// Store captures
			for i, name := range re.SubexpNames() {
				if i > 0 && name != "" && i < len(matches) {
					captures[name] = matches[i]
				}
			}
			// Also store numbered captures
			for i := 1; i < len(matches); i++ {
				captures[string(rune('0'+i))] = matches[i]
			}
			return true, captures
		}
	}

	// Exact match
	if fc.Code == dialed {
		return true, captures
	}

	// Extension alias match
	if fc.Extension != "" && fc.Extension == dialed {
		return true, captures
	}

	return false, nil
}

// ToDialplanXML generates FreeSWITCH dialplan XML for this feature code
func (fc *FeatureCode) ToDialplanXML(domain string) string {
	condition := fc.Code
	if fc.CodeRegex != "" {
		condition = fc.CodeRegex
	}

	return `<extension name="` + fc.Name + `" continue="false">
  <condition field="destination_number" expression="` + condition + `">
    <action application="set" data="feature_code_uuid=` + fc.UUID.String() + `"/>
    <action application="socket" data="127.0.0.6:9001 async full"/>
  </condition>
</extension>`
}

// GetFeatureCode looks up a feature code for a tenant (including global codes)
func GetFeatureCode(db *gorm.DB, tenantID uint, code string) (*FeatureCode, error) {
	var fc FeatureCode

	// Try tenant-specific first, then global
	err := db.Where("(tenant_id = ? OR tenant_id IS NULL) AND enabled = ?", tenantID, true).
		Order("tenant_id DESC NULLS LAST, \"order\" ASC"). // Tenant-specific takes priority
		Find(&[]FeatureCode{}).Error                       // Get all matching

	if err != nil {
		return nil, err
	}

	// Find matching code (supports regex)
	var codes []FeatureCode
	db.Where("(tenant_id = ? OR tenant_id IS NULL) AND enabled = ?", tenantID, true).
		Order("\"order\" ASC").
		Find(&codes)

	for _, candidate := range codes {
		if matches, _ := candidate.MatchesDialedNumber(code); matches {
			fc = candidate
			return &fc, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// ListFeatureCodes returns all feature codes for a tenant (including global)
func ListFeatureCodes(db *gorm.DB, tenantID *uint) ([]FeatureCode, error) {
	var codes []FeatureCode
	query := db.Order("\"order\" ASC")

	if tenantID != nil {
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL") // Global only
	}

	err := query.Find(&codes).Error
	return codes, err
}

// =====================
// Park Slot Management
// =====================

// ParkSlot represents a call parking slot
type ParkSlot struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UpdatedAt time.Time `json:"updated_at"`

	// Ownership
	TenantID uint   `json:"tenant_id" gorm:"index;not null"`
	Domain   string `json:"domain" gorm:"index;not null"`

	// Slot info
	SlotNumber int    `json:"slot_number" gorm:"not null"`
	LotName    string `json:"lot_name" gorm:"default:'default'"`
	Extension  string `json:"extension,omitempty"` // e.g., *5701 for slot 01

	// Current state
	IsOccupied bool      `json:"is_occupied" gorm:"default:false"`
	CallUUID   string    `json:"call_uuid,omitempty"`
	ParkedBy   string    `json:"parked_by,omitempty"` // Extension that parked
	ParkedAt   time.Time `json:"parked_at,omitempty"`
	Timeout    int       `json:"timeout" gorm:"default:120"` // Seconds before recall

	// Caller info
	CallerIDName   string `json:"caller_id_name,omitempty"`
	CallerIDNumber string `json:"caller_id_number,omitempty"`

	// BLF
	BLFState string `json:"blf_state" gorm:"default:'idle'"` // idle, parked
}

// TableName for ParkSlot
func (ParkSlot) TableName() string {
	return "park_slots"
}

// GetAvailableSlot finds the next available parking slot
func GetAvailableSlot(db *gorm.DB, tenantID uint, lotName string) (*ParkSlot, error) {
	var slot ParkSlot
	if lotName == "" {
		lotName = "default"
	}

	err := db.Where("tenant_id = ? AND lot_name = ? AND is_occupied = ?", tenantID, lotName, false).
		Order("slot_number ASC").
		First(&slot).Error

	return &slot, err
}

// GetSlotByNumber gets a specific parking slot
func GetSlotByNumber(db *gorm.DB, tenantID uint, slotNumber int, lotName string) (*ParkSlot, error) {
	var slot ParkSlot
	if lotName == "" {
		lotName = "default"
	}

	err := db.Where("tenant_id = ? AND lot_name = ? AND slot_number = ?", tenantID, lotName, slotNumber).
		First(&slot).Error

	return &slot, err
}

// ParkCall parks a call in a slot
func (ps *ParkSlot) ParkCall(db *gorm.DB, callUUID, parkedBy, callerName, callerNumber string) error {
	ps.IsOccupied = true
	ps.CallUUID = callUUID
	ps.ParkedBy = parkedBy
	ps.ParkedAt = time.Now()
	ps.CallerIDName = callerName
	ps.CallerIDNumber = callerNumber
	ps.BLFState = "parked"

	return db.Save(ps).Error
}

// RetrieveCall clears a parked call from the slot
func (ps *ParkSlot) RetrieveCall(db *gorm.DB) error {
	ps.IsOccupied = false
	ps.CallUUID = ""
	ps.ParkedBy = ""
	ps.CallerIDName = ""
	ps.CallerIDNumber = ""
	ps.BLFState = "idle"

	return db.Save(ps).Error
}

// EnsureParkSlots creates parking slots for a domain if they don't exist
func EnsureParkSlots(db *gorm.DB, tenantID uint, domain string, count int, startNumber int) error {
	for i := 0; i < count; i++ {
		slotNum := startNumber + i
		var existing ParkSlot
		err := db.Where("tenant_id = ? AND slot_number = ? AND lot_name = ?", tenantID, slotNum, "default").
			First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			slot := &ParkSlot{
				TenantID:   tenantID,
				Domain:     domain,
				SlotNumber: slotNum,
				LotName:    "default",
				Extension:  "*57" + fmt.Sprintf("%02d", slotNum),
			}
			if err := db.Create(slot).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// =====================
// Presence / BLF
// =====================

// PresenceState represents the BLF/presence state of an extension
type PresenceState string

const (
	PresenceIdle      PresenceState = "idle"
	PresenceRinging   PresenceState = "ringing"
	PresenceBusy      PresenceState = "busy"
	PresenceOnHold    PresenceState = "onhold"
	PresenceDND       PresenceState = "dnd"
	PresenceOffline   PresenceState = "offline"
	PresenceParked    PresenceState = "parked"
	PresenceAvailable PresenceState = "available"
)

// ExtensionPresence represents the current presence/BLF state of an extension
type ExtensionPresence struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UpdatedAt time.Time `json:"updated_at"`

	// Identity
	TenantID    uint   `json:"tenant_id" gorm:"index;not null"`
	ExtensionID uint   `json:"extension_id" gorm:"uniqueIndex;not null"`
	Extension   string `json:"extension" gorm:"index;not null"`

	// Presence state
	State         PresenceState `json:"state" gorm:"default:'offline'"`
	StatusMessage string        `json:"status_message,omitempty"`

	// Call info (when busy/ringing)
	CurrentCallUUID string `json:"current_call_uuid,omitempty"`
	CallerIDName    string `json:"caller_id_name,omitempty"`
	CallerIDNumber  string `json:"caller_id_number,omitempty"`
	Direction       string `json:"direction,omitempty"` // inbound/outbound

	// Features
	DND           bool `json:"dnd" gorm:"default:false"`
	CallForwarded bool `json:"call_forwarded" gorm:"default:false"`
	Registered    bool `json:"registered" gorm:"default:false"`

	// Last activity
	LastRegistration time.Time `json:"last_registration,omitempty"`
	LastCall         time.Time `json:"last_call,omitempty"`
}

// UpdatePresence updates the presence state for an extension
func UpdatePresence(db *gorm.DB, tenantID uint, extension string, state PresenceState) error {
	return db.Model(&ExtensionPresence{}).
		Where("tenant_id = ? AND extension = ?", tenantID, extension).
		Updates(map[string]interface{}{
			"state":      state,
			"updated_at": time.Now(),
		}).Error
}

// GetPresence gets the current presence for an extension
func GetPresence(db *gorm.DB, tenantID uint, extension string) (*ExtensionPresence, error) {
	var p ExtensionPresence
	err := db.Where("tenant_id = ? AND extension = ?", tenantID, extension).First(&p).Error
	return &p, err
}

// NotifyPresenceChange sends a NOTIFY for BLF subscription
// This should be called whenever presence changes
func NotifyPresenceChange(db *gorm.DB, tenantID uint, extension string, state PresenceState) error {
	// Update DB
	if err := UpdatePresence(db, tenantID, extension, state); err != nil {
		return err
	}

	// The actual SIP NOTIFY would be sent via ESL:
	// event PRESENCE_IN 'proto=sip|status=...|...'
	// This is handled by the ESL manager

	return nil
}
