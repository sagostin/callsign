package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// BroadcastStatus represents the status of a broadcast campaign
type BroadcastStatus string

const (
	BroadcastStatusDraft     BroadcastStatus = "draft"
	BroadcastStatusScheduled BroadcastStatus = "scheduled"
	BroadcastStatusRunning   BroadcastStatus = "running"
	BroadcastStatusPaused    BroadcastStatus = "paused"
	BroadcastStatusCompleted BroadcastStatus = "completed"
	BroadcastStatusCancelled BroadcastStatus = "cancelled"
)

// BroadcastCampaign represents a call broadcast campaign
type BroadcastCampaign struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	TenantID        uint            `json:"tenant_id" gorm:"index"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	Status          BroadcastStatus `json:"status" gorm:"default:draft"`
	RecordingID     *uint           `json:"recording_id,omitempty"`   // Link to audio library
	RecordingPath   string          `json:"recording_path,omitempty"` // Direct file path
	CallerID        string          `json:"caller_id"`
	ConcurrentLimit int             `json:"concurrent_limit" gorm:"default:5"`
	Timeout         int             `json:"timeout" gorm:"default:30"`       // Ring timeout per call in seconds
	RetryAttempts   int             `json:"retry_attempts" gorm:"default:0"` // Number of retries on failure
	RetryDelay      int             `json:"retry_delay" gorm:"default:300"`  // Delay between retries in seconds
	Recipients      pq.StringArray  `json:"recipients" gorm:"type:text[]"`   // List of phone numbers
	ScheduledAt     *time.Time      `json:"scheduled_at,omitempty"`
	StartedAt       *time.Time      `json:"started_at,omitempty"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`

	// Statistics (updated as campaign runs)
	TotalCalls    int `json:"total_calls" gorm:"default:0"`
	AnsweredCalls int `json:"answered_calls" gorm:"default:0"`
	FailedCalls   int `json:"failed_calls" gorm:"default:0"`
	BusyCalls     int `json:"busy_calls" gorm:"default:0"`
	NoAnswerCalls int `json:"no_answer_calls" gorm:"default:0"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Progress returns the campaign completion percentage
func (c *BroadcastCampaign) Progress() int {
	if c.TotalCalls == 0 || len(c.Recipients) == 0 {
		return 0
	}
	return int(float64(c.TotalCalls) / float64(len(c.Recipients)) * 100)
}
