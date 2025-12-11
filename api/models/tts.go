package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TTSProvider represents the Text-to-Speech provider type
type TTSProvider string

const (
	TTSElevenLabs TTSProvider = "elevenlabs"
	TTSOpenAI     TTSProvider = "openai"
	TTSGoogle     TTSProvider = "google"
	TTSAzure      TTSProvider = "azure"
	TTSEdge       TTSProvider = "edge"
	TTSPiper      TTSProvider = "piper"
	TTSCustom     TTSProvider = "custom"
)

// TTSVoice represents a configured TTS voice
type TTSVoice struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership - null = system voice
	TenantID *uint `json:"tenant_id" gorm:"index"`

	// Voice identification
	Name        string      `json:"name" gorm:"not null"`
	Description string      `json:"description,omitempty"`
	Provider    TTSProvider `json:"provider" gorm:"not null"`
	VoiceID     string      `json:"voice_id" gorm:"not null"` // Provider-specific voice ID
	Language    string      `json:"language" gorm:"default:'en-US'"`
	Gender      string      `json:"gender,omitempty"` // male, female, neutral
	Enabled     bool        `json:"enabled" gorm:"default:true"`
	IsDefault   bool        `json:"is_default" gorm:"default:false"`

	// Voice characteristics
	Speed float64 `json:"speed" gorm:"default:1.0"` // 0.5 - 2.0
	Pitch float64 `json:"pitch" gorm:"default:1.0"` // 0.5 - 2.0

	// Sample audio URL for preview
	SampleURL string `json:"sample_url,omitempty"`

	// Provider-specific settings as JSON
	Settings string `json:"settings,omitempty" gorm:"type:jsonb"`
}

func (v *TTSVoice) BeforeCreate(tx *gorm.DB) error {
	v.UUID = uuid.New()
	return nil
}

// GetSettings unmarshals provider-specific settings
func (v *TTSVoice) GetSettings() (map[string]interface{}, error) {
	if v.Settings == "" {
		return make(map[string]interface{}), nil
	}
	var settings map[string]interface{}
	err := json.Unmarshal([]byte(v.Settings), &settings)
	return settings, err
}

// TTSConfig holds provider-specific configuration
type TTSConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TenantID  *uint     `json:"tenant_id" gorm:"index"` // null = system default
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Provider settings
	Provider TTSProvider `json:"provider"`
	Enabled  bool        `json:"enabled" gorm:"default:true"`
	Priority int         `json:"priority" gorm:"default:0"` // For fallback ordering

	// API credentials (encrypted)
	APIKey       string `json:"-"`
	APIEndpoint  string `json:"api_endpoint,omitempty"`
	APISecretEnc string `json:"-" gorm:"column:api_secret_enc"` // Encrypted

	// Provider-specific settings as JSON
	Settings string `json:"settings,omitempty" gorm:"type:jsonb"`

	// Default voice for this provider
	DefaultVoiceID string `json:"default_voice_id,omitempty"`

	// Usage tracking
	UsageCharsMonth int64   `json:"usage_chars_month" gorm:"default:0"`
	UsageLimitChars int64   `json:"usage_limit_chars" gorm:"default:0"` // 0 = unlimited
	CostPerChar     float64 `json:"cost_per_char,omitempty"`
}

// TTSAudioFile represents a generated TTS audio file
type TTSAudioFile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership
	TenantID uint `json:"tenant_id" gorm:"index"`

	// Source reference
	SourceType string    `json:"source_type" gorm:"index"` // "greeting", "ivr", "announcement", "custom"
	SourceUUID uuid.UUID `json:"source_uuid,omitempty" gorm:"type:uuid;index"`

	// Input text
	Text     string `json:"text" gorm:"type:text"`
	SSML     string `json:"ssml,omitempty" gorm:"type:text"` // If using SSML
	Language string `json:"language" gorm:"default:'en-US'"`

	// Voice used
	Provider TTSProvider `json:"provider"`
	VoiceID  string      `json:"voice_id"`
	Speed    float64     `json:"speed" gorm:"default:1.0"`
	Pitch    float64     `json:"pitch" gorm:"default:1.0"`

	// Output file
	FilePath   string  `json:"file_path"`
	FileFormat string  `json:"file_format" gorm:"default:'wav'"` // wav, mp3, ogg
	Duration   float64 `json:"duration"`                         // seconds
	FileSize   int64   `json:"file_size"`                        // bytes

	// Cost tracking
	CharCount int     `json:"char_count"`
	Cost      float64 `json:"cost,omitempty"`

	// Caching
	TextHash string `json:"text_hash" gorm:"index"` // For cache lookup
}

func (a *TTSAudioFile) BeforeCreate(tx *gorm.DB) error {
	a.UUID = uuid.New()
	return nil
}

// SystemPhrase represents a pre-defined system phrase with TTS
type SystemPhrase struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Ownership - null = system phrase
	TenantID *uint `json:"tenant_id" gorm:"index"`

	// Phrase identification
	PhraseKey   string `json:"phrase_key" gorm:"index;not null"` // e.g., "vm_greeting", "ivr_welcome"
	Category    string `json:"category" gorm:"index"`            // voicemail, ivr, queue, etc.
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description,omitempty"`

	// Language versions
	Translations []PhraseTranslation `json:"translations" gorm:"foreignKey:PhraseID"`
}

func (p *SystemPhrase) BeforeCreate(tx *gorm.DB) error {
	p.UUID = uuid.New()
	return nil
}

// PhraseTranslation represents a language version of a phrase
type PhraseTranslation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PhraseID  uint      `json:"phrase_id" gorm:"index"`
	Language  string    `json:"language" gorm:"index;not null"` // en-US, es-MX, etc.
	Text      string    `json:"text" gorm:"type:text;not null"`
	SSML      string    `json:"ssml,omitempty" gorm:"type:text"`
	AudioPath string    `json:"audio_path,omitempty"` // Pre-generated audio
	VoiceID   string    `json:"voice_id,omitempty"`   // Voice used for audio
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultSystemPhrases returns the default system phrases
func DefaultSystemPhrases() []SystemPhrase {
	return []SystemPhrase{
		// Voicemail
		{
			PhraseKey: "vm_greeting_default",
			Category:  "voicemail",
			Name:      "Default Voicemail Greeting",
			Translations: []PhraseTranslation{
				{Language: "en-US", Text: "The person you are trying to reach is not available. Please leave a message after the tone."},
				{Language: "es-MX", Text: "La persona que intenta contactar no está disponible. Por favor deje un mensaje después del tono."},
			},
		},
		{
			PhraseKey: "vm_instructions",
			Category:  "voicemail",
			Name:      "Voicemail Instructions",
			Translations: []PhraseTranslation{
				{Language: "en-US", Text: "Press 1 to listen to your messages. Press 2 to record a new greeting. Press 3 for advanced options."},
				{Language: "es-MX", Text: "Presione 1 para escuchar sus mensajes. Presione 2 para grabar un nuevo saludo. Presione 3 para opciones avanzadas."},
			},
		},
		// IVR
		{
			PhraseKey: "ivr_welcome",
			Category:  "ivr",
			Name:      "IVR Welcome",
			Translations: []PhraseTranslation{
				{Language: "en-US", Text: "Thank you for calling. Please listen to the following options."},
				{Language: "es-MX", Text: "Gracias por llamar. Por favor escuche las siguientes opciones."},
			},
		},
		{
			PhraseKey: "ivr_invalid_option",
			Category:  "ivr",
			Name:      "IVR Invalid Option",
			Translations: []PhraseTranslation{
				{Language: "en-US", Text: "That is not a valid option. Please try again."},
				{Language: "es-MX", Text: "Esa no es una opción válida. Por favor intente de nuevo."},
			},
		},
		// Queue
		{
			PhraseKey: "queue_position",
			Category:  "queue",
			Name:      "Queue Position",
			Translations: []PhraseTranslation{
				{Language: "en-US", Text: "You are caller number {position} in the queue. Please hold."},
				{Language: "es-MX", Text: "Usted es el llamante número {position} en la cola. Por favor espere."},
			},
		},
		{
			PhraseKey: "queue_hold_music_interrupt",
			Category:  "queue",
			Name:      "Queue Hold Music Interrupt",
			Translations: []PhraseTranslation{
				{Language: "en-US", Text: "Thank you for your patience. An agent will be with you shortly."},
				{Language: "es-MX", Text: "Gracias por su paciencia. Un agente estará con usted en breve."},
			},
		},
		// Parking
		{
			PhraseKey: "park_slot_announcement",
			Category:  "parking",
			Name:      "Park Slot Announcement",
			Translations: []PhraseTranslation{
				{Language: "en-US", Text: "Your call has been parked in slot {slot}."},
				{Language: "es-MX", Text: "Su llamada ha sido estacionada en el espacio {slot}."},
			},
		},
	}
}
