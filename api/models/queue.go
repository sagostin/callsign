package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QueueStrategy defines call distribution strategies
type QueueStrategy string

const (
	QueueStrategyRingAll        QueueStrategy = "ring-all"
	QueueStrategyLongestIdle    QueueStrategy = "longest-idle-agent"
	QueueStrategyRoundRobin     QueueStrategy = "round-robin"
	QueueStrategyTopDown        QueueStrategy = "top-down"
	QueueStrategyAgentWithLeast QueueStrategy = "agent-with-least-talk-time"
	QueueStrategyRandom         QueueStrategy = "random"
)

// AgentStatus defines agent availability states
type AgentStatus string

const (
	AgentStatusLoggedOut         AgentStatus = "Logged Out"
	AgentStatusAvailable         AgentStatus = "Available"
	AgentStatusAvailableOnDemand AgentStatus = "Available (On Demand)"
	AgentStatusOnBreak           AgentStatus = "On Break"
)

// AgentState defines agent call states
type AgentState string

const (
	AgentStateIdle      AgentState = "Idle"
	AgentStateWaiting   AgentState = "Waiting"
	AgentStateReceiving AgentState = "Receiving"
	AgentStateInCall    AgentState = "In a queue call"
)

// Queue represents a call center queue
type Queue struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID  uint   `json:"tenant_id" gorm:"index;not null"`
	Name      string `json:"name" gorm:"not null"`
	Extension string `json:"extension" gorm:"index"` // Queue dial number

	// Strategy
	Strategy        QueueStrategy `json:"strategy" gorm:"default:'longest-idle-agent'"`
	MohSound        string        `json:"moh_sound"`        // Music on hold
	TimeBasedScore  string        `json:"time_based_score"` // queue, system
	TierRulesApply  bool          `json:"tier_rules_apply" gorm:"default:true"`
	TierRuleWaitSec int           `json:"tier_rule_wait_sec" gorm:"default:30"`

	// Timeouts
	MaxWaitTime                   int `json:"max_wait_time" gorm:"default:0"` // 0 = unlimited
	MaxWaitTimeNoAgent            int `json:"max_wait_time_no_agent" gorm:"default:90"`
	MaxWaitTimeNoAgentTimeReached int `json:"max_wait_time_no_agent_time_reached" gorm:"default:0"`

	// Agent settings
	WrapUpTime        int `json:"wrap_up_time" gorm:"default:10"`      // Seconds after call
	RejectDelayTime   int `json:"reject_delay_time" gorm:"default:60"` // After reject
	BusyDelayTime     int `json:"busy_delay_time" gorm:"default:60"`   // After busy
	NoAnswerDelayTime int `json:"no_answer_delay_time" gorm:"default:60"`

	// Announcements
	AnnounceSound     string `json:"announce_sound"`     // Periodic announcement
	AnnounceFrequency int    `json:"announce_frequency"` // Seconds between
	AnnouncePosition  bool   `json:"announce_position" gorm:"default:true"`

	// Exit destinations
	ExitAction      string `json:"exit_action"`      // hangup, transfer
	ExitDestination string `json:"exit_destination"` // Transfer destination

	// Status
	Enabled bool `json:"enabled" gorm:"default:true"`

	// Relations
	Agents []QueueAgent `json:"agents,omitempty" gorm:"foreignKey:QueueID"`
}

// BeforeCreate generates UUID
func (q *Queue) BeforeCreate(tx *gorm.DB) error {
	q.UUID = uuid.New()
	return nil
}

// QueueAgent represents an agent assigned to a queue (tier)
type QueueAgent struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Assignment
	QueueID     uint `json:"queue_id" gorm:"index;not null"`
	ExtensionID uint `json:"extension_id" gorm:"index"`
	TenantID    uint `json:"tenant_id" gorm:"index;not null"`

	// Agent identity (for callcenter_config)
	AgentName string `json:"agent_name" gorm:"not null"` // e.g., "1001@domain.com"
	Contact   string `json:"contact"`                    // Dial string

	// Tier settings
	TierLevel    int `json:"tier_level" gorm:"default:1"`    // 1 = highest priority
	TierPosition int `json:"tier_position" gorm:"default:1"` // Order within tier

	// Status (synced with FreeSWITCH)
	Status        AgentStatus `json:"status" gorm:"default:'Logged Out'"`
	State         AgentState  `json:"state" gorm:"default:'Idle'"`
	LastBridgeEnd *time.Time  `json:"last_bridge_end,omitempty"`

	// Settings
	MaxNoAnswer       int `json:"max_no_answer" gorm:"default:3"`
	WrapUpTime        int `json:"wrap_up_time"` // Override queue setting
	NoAnswerDelayTime int `json:"no_answer_delay_time"`
}

// QueueMember represents a caller waiting in queue (runtime, not persisted)
type QueueMember struct {
	UUID           string    `json:"uuid"`
	QueueName      string    `json:"queue_name"`
	CallerIDName   string    `json:"caller_id_name"`
	CallerIDNumber string    `json:"caller_id_number"`
	JoinedAt       time.Time `json:"joined_at"`
	State          string    `json:"state"` // Waiting, Trying, Answered
	ServingAgent   string    `json:"serving_agent,omitempty"`
}

// QueueStatistics represents real-time queue stats
type QueueStatistics struct {
	QueueID         uint      `json:"queue_id"`
	QueueName       string    `json:"queue_name"`
	WaitingCalls    int       `json:"waiting_calls"`
	AvailableAgents int       `json:"available_agents"`
	TotalAgents     int       `json:"total_agents"`
	TotalCallsToday int       `json:"total_calls_today"`
	AbandonedToday  int       `json:"abandoned_today"`
	AvgWaitTime     int       `json:"avg_wait_time_sec"`
	AvgTalkTime     int       `json:"avg_talk_time_sec"`
	LongestWait     int       `json:"longest_wait_sec"`
	LastUpdated     time.Time `json:"last_updated"`
}
