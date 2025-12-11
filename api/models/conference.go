package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Conference represents a conference room
type Conference struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID  uint   `json:"tenant_id" gorm:"index;not null"`
	Name      string `json:"name" gorm:"not null"`
	Extension string `json:"extension" gorm:"index"` // Dial-in number

	// FreeSWITCH profile
	ProfileName string `json:"profile_name" gorm:"default:'default'"` // FS conference profile

	// Authentication
	PIN          string `json:"pin,omitempty"`           // Participant PIN
	ModeratorPIN string `json:"moderator_pin,omitempty"` // Moderator PIN

	// Settings
	MaxMembers       int            `json:"max_members" gorm:"default:0"` // 0 = unlimited
	WaitForModerator bool           `json:"wait_for_moderator" gorm:"default:false"`
	MuteOnJoin       bool           `json:"mute_on_join" gorm:"default:false"`
	AnnounceCount    bool           `json:"announce_count" gorm:"default:true"`
	RecordConference bool           `json:"record_conference" gorm:"default:false"`
	Flags            pq.StringArray `json:"flags" gorm:"type:text[]"` // Additional FS flags

	// Audio
	MusicOnHold string `json:"music_on_hold"`
	EnterSound  string `json:"enter_sound"`
	ExitSound   string `json:"exit_sound"`

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`

	// Relations
	Members []ConferenceMember `json:"members,omitempty" gorm:"foreignKey:ConferenceID"`
}

// BeforeCreate generates UUID
func (c *Conference) BeforeCreate(tx *gorm.DB) error {
	c.UUID = uuid.New()
	return nil
}

// ConferenceMemberRole defines member roles
type ConferenceMemberRole string

const (
	ConferenceRoleModerator   ConferenceMemberRole = "moderator"
	ConferenceRoleParticipant ConferenceMemberRole = "participant"
)

// ConferenceMember represents a user who can join a conference with specific permissions
type ConferenceMember struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Assignment
	ConferenceID uint `json:"conference_id" gorm:"index;not null"`
	ExtensionID  uint `json:"extension_id" gorm:"index"`
	TenantID     uint `json:"tenant_id" gorm:"index;not null"`

	// Identity
	Extension string               `json:"extension"`
	Role      ConferenceMemberRole `json:"role" gorm:"default:'participant'"`

	// Permissions
	CanMuteOthers  bool `json:"can_mute_others" gorm:"default:false"`
	CanKickOthers  bool `json:"can_kick_others" gorm:"default:false"`
	CanLockConf    bool `json:"can_lock_conf" gorm:"default:false"`
	CanUnlockConf  bool `json:"can_unlock_conf" gorm:"default:false"`
	CanBarge       bool `json:"can_barge" gorm:"default:false"` // Listen to conf silently
	CanStartRecord bool `json:"can_start_record" gorm:"default:false"`
	CanStopRecord  bool `json:"can_stop_record" gorm:"default:false"`
	CanDeafOthers  bool `json:"can_deaf_others" gorm:"default:false"` // Make others deaf
	CanSetFloor    bool `json:"can_set_floor" gorm:"default:false"`   // Set who has floor

	// Settings
	StartMuted bool `json:"start_muted" gorm:"default:false"`
	StartDeaf  bool `json:"start_deaf" gorm:"default:false"`
}

// SetModeratorPermissions grants all moderator permissions
func (m *ConferenceMember) SetModeratorPermissions() {
	m.Role = ConferenceRoleModerator
	m.CanMuteOthers = true
	m.CanKickOthers = true
	m.CanLockConf = true
	m.CanUnlockConf = true
	m.CanBarge = true
	m.CanStartRecord = true
	m.CanStopRecord = true
	m.CanDeafOthers = true
	m.CanSetFloor = true
}

// ConferenceAction represents actions that can be performed on a conference
type ConferenceAction string

const (
	ConfActionLock    ConferenceAction = "lock"
	ConfActionUnlock  ConferenceAction = "unlock"
	ConfActionMute    ConferenceAction = "mute"
	ConfActionUnmute  ConferenceAction = "unmute"
	ConfActionDeaf    ConferenceAction = "deaf"
	ConfActionUndeaf  ConferenceAction = "undeaf"
	ConfActionKick    ConferenceAction = "kick"
	ConfActionHangup  ConferenceAction = "hup"
	ConfActionRecord  ConferenceAction = "recording start"
	ConfActionStopRec ConferenceAction = "recording stop"
	ConfActionFloor   ConferenceAction = "floor"
)

// ConferenceSession represents an active or past conference session
type ConferenceSession struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Conference reference
	ConferenceID   uint   `json:"conference_id" gorm:"index;not null"`
	ConferenceName string `json:"conference_name"` // FS conference name
	TenantID       uint   `json:"tenant_id" gorm:"index;not null"`

	// Session times
	StartTime time.Time  `json:"start_time" gorm:"index"`
	EndTime   *time.Time `json:"end_time,omitempty"`

	// Session state
	MemberCount int  `json:"member_count" gorm:"default:0"`
	MaxMembers  int  `json:"max_members" gorm:"default:0"` // Peak concurrent
	Locked      bool `json:"locked" gorm:"default:false"`

	// Recording
	Recording     bool   `json:"recording" gorm:"default:false"`
	RecordingPath string `json:"recording_path,omitempty"`

	// Stats
	TotalJoins      int `json:"total_joins" gorm:"default:0"`       // Total joins (including rejoins)
	UniqueMemberIDs int `json:"unique_member_ids" gorm:"default:0"` // Unique participants

	// Relations
	Participants []ConferenceParticipant `json:"participants,omitempty" gorm:"foreignKey:SessionID"`
}

// BeforeCreate generates UUID
func (s *ConferenceSession) BeforeCreate(tx *gorm.DB) error {
	s.UUID = uuid.New()
	return nil
}

// IsActive returns true if session is still active
func (s *ConferenceSession) IsActive() bool {
	return s.EndTime == nil
}

// Duration returns session duration in seconds
func (s *ConferenceSession) Duration() int {
	end := time.Now()
	if s.EndTime != nil {
		end = *s.EndTime
	}
	return int(end.Sub(s.StartTime).Seconds())
}

// ConferenceParticipant represents a live or past participant in a session
type ConferenceParticipant struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Session reference
	SessionID uint `json:"session_id" gorm:"index;not null"`
	TenantID  uint `json:"tenant_id" gorm:"index;not null"`

	// FreeSWITCH identifiers
	MemberID int    `json:"member_id"`         // FS conference member ID
	UUID     string `json:"uuid" gorm:"index"` // Channel UUID

	// Caller info
	CallerIDName   string `json:"caller_id_name"`
	CallerIDNumber string `json:"caller_id_number"`
	ExtensionID    uint   `json:"extension_id,omitempty"`

	// Participation times
	JoinTime  time.Time  `json:"join_time"`
	LeaveTime *time.Time `json:"leave_time,omitempty"`

	// State (synced from FreeSWITCH)
	Muted     bool `json:"muted" gorm:"default:false"`
	Deaf      bool `json:"deaf" gorm:"default:false"`
	Video     bool `json:"video" gorm:"default:false"`
	Floor     bool `json:"floor" gorm:"default:false"`     // Has speaking floor
	Moderator bool `json:"moderator" gorm:"default:false"` // Joined as moderator
	Talking   bool `json:"talking" gorm:"default:false"`   // Currently speaking

	// Energy/volume levels
	EnergyLevel int `json:"energy_level" gorm:"default:0"`
	VolumeIn    int `json:"volume_in" gorm:"default:0"`
	VolumeOut   int `json:"volume_out" gorm:"default:0"`
}

// IsActive returns true if participant is still in conference
func (p *ConferenceParticipant) IsActive() bool {
	return p.LeaveTime == nil
}

// Duration returns participant duration in seconds
func (p *ConferenceParticipant) Duration() int {
	end := time.Now()
	if p.LeaveTime != nil {
		end = *p.LeaveTime
	}
	return int(end.Sub(p.JoinTime).Seconds())
}

// ConferenceStats represents aggregated conference statistics
type ConferenceStats struct {
	TenantID           uint    `json:"tenant_id"`
	Date               string  `json:"date"` // YYYY-MM-DD
	TotalSessions      int     `json:"total_sessions"`
	TotalMinutes       int     `json:"total_minutes"`
	TotalParticipants  int     `json:"total_participants"`
	UniqueParticipants int     `json:"unique_participants"`
	PeakConcurrent     int     `json:"peak_concurrent"`
	AvgSessionDuration float64 `json:"avg_session_duration_min"`
	AvgParticipants    float64 `json:"avg_participants_per_session"`
	RecordedSessions   int     `json:"recorded_sessions"`
}

// GetConferenceStats calculates stats for a tenant and date range
func GetConferenceStats(db *gorm.DB, tenantID uint, startDate, endDate time.Time) (*ConferenceStats, error) {
	stats := &ConferenceStats{
		TenantID: tenantID,
		Date:     startDate.Format("2006-01-02"),
	}

	var count int64

	// Total sessions
	db.Model(&ConferenceSession{}).Where(
		"tenant_id = ? AND start_time BETWEEN ? AND ?",
		tenantID, startDate, endDate,
	).Count(&count)
	stats.TotalSessions = int(count)

	// Total participants
	db.Model(&ConferenceParticipant{}).
		Joins("JOIN conference_sessions ON conference_participants.session_id = conference_sessions.id").
		Where("conference_sessions.tenant_id = ? AND conference_participants.join_time BETWEEN ? AND ?",
			tenantID, startDate, endDate).
		Count(&count)
	stats.TotalParticipants = int(count)

	// Peak concurrent (max member_count from any session)
	var maxMembers struct{ Max int }
	db.Model(&ConferenceSession{}).Select("COALESCE(MAX(max_members), 0) as max").
		Where("tenant_id = ? AND start_time BETWEEN ? AND ?",
			tenantID, startDate, endDate).
		Scan(&maxMembers)
	stats.PeakConcurrent = maxMembers.Max

	// Recorded sessions
	db.Model(&ConferenceSession{}).Where(
		"tenant_id = ? AND start_time BETWEEN ? AND ? AND recording = true",
		tenantID, startDate, endDate,
	).Count(&count)
	stats.RecordedSessions = int(count)

	// Calculate averages
	if stats.TotalSessions > 0 {
		stats.AvgParticipants = float64(stats.TotalParticipants) / float64(stats.TotalSessions)
	}

	return stats, nil
}

// LiveConferenceInfo represents real-time conference data from FreeSWITCH
type LiveConferenceInfo struct {
	Name        string                 `json:"name"`
	MemberCount int                    `json:"member_count"`
	Locked      bool                   `json:"locked"`
	Recording   bool                   `json:"recording"`
	RunTime     int                    `json:"run_time_seconds"`
	Members     []LiveConferenceMember `json:"members"`
}

// LiveConferenceMember represents a live member from FreeSWITCH
type LiveConferenceMember struct {
	ID             int    `json:"id"`
	UUID           string `json:"uuid"`
	CallerIDName   string `json:"caller_id_name"`
	CallerIDNumber string `json:"caller_id_number"`
	Muted          bool   `json:"muted"`
	Deaf           bool   `json:"deaf"`
	Talking        bool   `json:"talking"`
	Floor          bool   `json:"floor"`
	Video          bool   `json:"video"`
	EnergyLevel    int    `json:"energy_level"`
}
