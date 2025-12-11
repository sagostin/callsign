package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TranscriptionStatus represents the status of a transcription job
type TranscriptionStatus string

const (
	TranscriptionPending    TranscriptionStatus = "pending"
	TranscriptionProcessing TranscriptionStatus = "processing"
	TranscriptionCompleted  TranscriptionStatus = "completed"
	TranscriptionFailed     TranscriptionStatus = "failed"
)

// TranscriptionProvider represents supported transcription providers
type TranscriptionProvider string

const (
	TranscriptionProviderWhisper   TranscriptionProvider = "whisper"    // Local Whisper
	TranscriptionProviderOpenAI    TranscriptionProvider = "openai"     // OpenAI API
	TranscriptionProviderGoogleSTT TranscriptionProvider = "google_stt" // Google Speech-to-Text
	TranscriptionProviderAWS       TranscriptionProvider = "aws_transcribe"
	TranscriptionProviderAzure     TranscriptionProvider = "azure_speech"
	TranscriptionProviderDeepgram  TranscriptionProvider = "deepgram"
)

// CallRecording represents a call recording
type CallRecording struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Call Reference
	CallRecordID uint   `json:"call_record_id" gorm:"index"`          // Reference to CDR
	CallUUID     string `json:"call_uuid" gorm:"index;size:64"`       // FreeSWITCH call UUID
	ConferenceID *uint  `json:"conference_id,omitempty" gorm:"index"` // If conference recording

	// Recording Details
	Direction    string    `json:"direction"` // inbound, outbound, local
	CallerNumber string    `json:"caller_number"`
	CallerName   string    `json:"caller_name,omitempty"`
	CalleeNumber string    `json:"callee_number"`
	CalleeName   string    `json:"callee_name,omitempty"`
	ExtensionID  *uint     `json:"extension_id,omitempty" gorm:"index"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time,omitempty"`
	Duration     int       `json:"duration"` // Seconds

	// File Information
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`   // Full path to recording
	FileSize   int64  `json:"file_size"`   // Bytes
	FileFormat string `json:"file_format"` // wav, mp3, ogg
	SampleRate int    `json:"sample_rate"` // 8000, 16000, etc.
	Channels   int    `json:"channels"`    // 1=mono, 2=stereo

	// Storage
	StorageType string `json:"storage_type"`          // local, s3, gcs, azure
	StorageURL  string `json:"storage_url,omitempty"` // URL if cloud storage

	// Retention
	RetentionDays int       `json:"retention_days"`       // Days to keep
	ExpiresAt     time.Time `json:"expires_at,omitempty"` // Auto-delete date
	Archived      bool      `json:"archived" gorm:"default:false"`

	// Access Control
	IsConfidential bool `json:"is_confidential" gorm:"default:false"`

	// Transcription
	TranscriptionID     *uint               `json:"transcription_id,omitempty"`
	TranscriptionStatus TranscriptionStatus `json:"transcription_status" gorm:"default:'pending'"`
	Transcription       *Transcription      `json:"transcription,omitempty" gorm:"foreignKey:ID;references:TranscriptionID"`

	// Tags/Notes
	Tags  string `json:"tags,omitempty"` // Comma-separated tags
	Notes string `json:"notes,omitempty"`
}

// BeforeCreate generates UUID
func (c *CallRecording) BeforeCreate(tx *gorm.DB) error {
	c.UUID = uuid.New()
	return nil
}

// Transcription represents a transcription of a call recording
type Transcription struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint `json:"tenant_id" gorm:"index;not null"`

	// Recording Reference
	CallRecordingID uint `json:"call_recording_id" gorm:"index;not null"`

	// Status
	Status           TranscriptionStatus   `json:"status" gorm:"default:'pending'"`
	Provider         TranscriptionProvider `json:"provider"`
	ProviderJobID    string                `json:"provider_job_id,omitempty"` // External job ID
	ErrorMessage     string                `json:"error_message,omitempty"`
	ProcessingTimeMs int                   `json:"processing_time_ms,omitempty"` // Time to transcribe

	// Transcription Content
	FullText     string `json:"full_text" gorm:"type:text"` // Complete transcription
	FullTextJSON string `json:"-" gorm:"type:text"`         // JSON with timestamps

	// Analytics/Metadata
	Language     string  `json:"language"`   // Detected language
	Confidence   float64 `json:"confidence"` // Overall confidence 0-1
	WordCount    int     `json:"word_count"`
	SpeakerCount int     `json:"speaker_count"` // Diarization

	// Sentiment Analysis (optional)
	SentimentScore float64 `json:"sentiment_score,omitempty"` // -1 to 1
	SentimentLabel string  `json:"sentiment_label,omitempty"` // positive, negative, neutral

	// Keywords/Topics (optional)
	Keywords string `json:"keywords,omitempty"` // Extracted keywords
	Topics   string `json:"topics,omitempty"`   // Detected topics

	// Summary (optional)
	Summary string `json:"summary,omitempty" gorm:"type:text"` // AI-generated summary

	// Segments (for playback sync)
	Segments []TranscriptionSegment `json:"segments,omitempty" gorm:"foreignKey:TranscriptionID"`
}

// BeforeCreate generates UUID
func (t *Transcription) BeforeCreate(tx *gorm.DB) error {
	t.UUID = uuid.New()
	return nil
}

// TranscriptionSegment represents a segment of transcription with timing
type TranscriptionSegment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TranscriptionID uint `json:"transcription_id" gorm:"index;not null"`

	// Timing (milliseconds)
	StartMs int `json:"start_ms"`
	EndMs   int `json:"end_ms"`

	// Content
	Text       string  `json:"text"`
	Speaker    string  `json:"speaker,omitempty"` // Speaker label (Speaker 1, Agent, etc.)
	Confidence float64 `json:"confidence"`

	// Word-level timing (JSON array)
	WordsJSON string `json:"-" gorm:"type:text"` // [{"word":"hello","start":0,"end":500}]
}

// TranscriptionConfig represents tenant-level transcription settings
type TranscriptionConfig struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"uniqueIndex;not null"`

	// Provider Settings
	Provider        TranscriptionProvider `json:"provider" gorm:"default:'whisper'"`
	APIKeyEncrypted string                `json:"-"`                      // Encrypted API key
	APIEndpoint     string                `json:"api_endpoint,omitempty"` // Custom endpoint

	// Auto-Transcription Rules
	AutoTranscribe       bool `json:"auto_transcribe" gorm:"default:false"`
	TranscribeInbound    bool `json:"transcribe_inbound" gorm:"default:true"`
	TranscribeOutbound   bool `json:"transcribe_outbound" gorm:"default:true"`
	TranscribeVoicemail  bool `json:"transcribe_voicemail" gorm:"default:true"`
	TranscribeConference bool `json:"transcribe_conference" gorm:"default:false"`
	MinDurationSeconds   int  `json:"min_duration_seconds" gorm:"default:10"` // Only transcribe if > X seconds

	// Quality Settings
	Language        string `json:"language" gorm:"default:'en'"`       // Preferred language
	EnableDiarize   bool   `json:"enable_diarize" gorm:"default:true"` // Speaker diarization
	EnableSentiment bool   `json:"enable_sentiment" gorm:"default:false"`
	EnableSummary   bool   `json:"enable_summary" gorm:"default:false"`

	// Whisper-specific
	WhisperModel  string `json:"whisper_model" gorm:"default:'base'"` // tiny, base, small, medium, large
	WhisperDevice string `json:"whisper_device" gorm:"default:'cpu'"` // cpu, cuda
}

// RecordingConfig represents tenant-level recording settings
type RecordingConfig struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TenantID uint `json:"tenant_id" gorm:"uniqueIndex;not null"`

	// Recording Rules
	RecordInbound    bool `json:"record_inbound" gorm:"default:false"`
	RecordOutbound   bool `json:"record_outbound" gorm:"default:false"`
	RecordLocal      bool `json:"record_local" gorm:"default:false"` // Extension to extension
	RecordConference bool `json:"record_conference" gorm:"default:true"`
	RecordVoicemail  bool `json:"record_voicemail" gorm:"default:true"`

	// Format Settings
	Format       string `json:"format" gorm:"default:'wav'"` // wav, mp3, ogg
	SampleRate   int    `json:"sample_rate" gorm:"default:16000"`
	Channels     int    `json:"channels" gorm:"default:2"`            // 1=mono, 2=stereo
	StereoLayout string `json:"stereo_layout" gorm:"default:'mixed'"` // mixed, split (leg a/b)

	// Storage Settings
	StorageType       string `json:"storage_type" gorm:"default:'local'"` // local, s3, gcs
	StoragePath       string `json:"storage_path"`                        // Local path or bucket
	StorageRegion     string `json:"storage_region,omitempty"`            // Cloud region
	StorageKeyEncrypt string `json:"-"`                                   // Encrypted access key

	// Retention
	RetentionDays       int  `json:"retention_days" gorm:"default:90"`
	AutoDeleteExpired   bool `json:"auto_delete_expired" gorm:"default:true"`
	ArchiveBeforeDelete bool `json:"archive_before_delete" gorm:"default:false"`

	// Legal/Compliance
	AnnouncementEnabled bool   `json:"announcement_enabled" gorm:"default:false"` // "This call may be recorded"
	AnnouncementFile    string `json:"announcement_file,omitempty"`
	RequireConsent      bool   `json:"require_consent" gorm:"default:false"` // Press 1 to consent
	BeeponStart         bool   `json:"beep_on_start" gorm:"default:false"`

	// Encryption
	EncryptRecordings bool   `json:"encrypt_recordings" gorm:"default:false"`
	EncryptionKeyID   string `json:"encryption_key_id,omitempty"`
}
