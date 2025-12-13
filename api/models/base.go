package models

import (
	"callsign/config"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Database connection established")
	return db, nil
}

// AutoMigrate runs database migrations for all models
func AutoMigrate(db *gorm.DB) error {
	log.Info("Running database migrations...")

	// Pre-migration: Handle special column type conversions
	// that GORM can't auto-migrate (e.g., text[] -> jsonb)
	if err := runPreMigrations(db); err != nil {
		log.Warnf("Pre-migration step failed (may be expected on fresh install): %v", err)
	}

	err := db.AutoMigrate(
		// Core models
		&User{},
		&Tenant{},
		&TenantProfile{},

		// Extension/Directory models
		&Extension{},
		&ExtensionSetting{},
		&ExtensionProfile{},

		// SIP/Sofia models
		&SIPProfile{},
		&SIPProfileSetting{},
		&SIPProfileDomain{},
		&SofiaGlobalSetting{},
		&Gateway{},
		&ACL{},
		&ACLNode{},

		// Dialplan models
		&Dialplan{},
		&DialplanDetail{},
		&Destination{},

		// Voicemail models
		&VoicemailBox{},
		&VoicemailMessage{},

		// Queue models
		&Queue{},
		&QueueAgent{},

		// Conference models
		&Conference{},
		&ConferenceMember{},
		&ConferenceSession{},
		&ConferenceParticipant{},

		// Ring group models
		&RingGroup{},
		&RingGroupDestination{},

		// Feature codes & Presence
		&FeatureCode{},
		&ExtensionPresence{},

		// IVR & Routing
		&IVRMenu{},
		&IVRMenuOption{},
		&TimeCondition{},
		&HolidayList{},
		&CallFlow{},
		&Recording{},
		&Contact{},

		// SMS & Chatplan - Conversation MUST come before Message
		&Conversation{},
		&Message{},
		&MessageMedia{},
		&MessagingProvider{},
		&Chatplan{},
		&Phrase{},
		&Sound{},
		&DefaultOutboundRoute{},

		// Unified Chat System
		&ChatRoom{},
		&ChatRoomMember{},
		&ChatQueue{},
		&ChatQueueAgent{},
		&ChatThread{},
		&ChatMessage{},
		&ChatAttachment{},
		&ContactWebhook{},
		&ChatReadReceipt{},

		// Audit & CDR
		&AuditLog{},
		&CallRecord{},
		&BannedIP{},
		&PageGroup{},
		&PageGroupDestination{},

		// Provisioning
		&ProvisioningTemplate{},
		&ProvisioningVariable{},

		// Device Management
		&Device{},
		&DeviceLine{},
		&DeviceTemplate{},
		&DeviceProfile{},
		&Firmware{},

		// Call Recording & Transcription
		&CallRecording{},

		// Call Recording & Transcription
		&CallRecording{},
		&Transcription{},
		&TranscriptionSegment{},
		&TranscriptionConfig{},
		&RecordingConfig{},
		&MediaFile{},

		// Speed Dials
		&SpeedDialGroup{},

		// Call Handling Rules
		&CallHandlingRule{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database migrations completed")
	return nil
}

// runPreMigrations handles column type conversions that GORM can't auto-migrate
func runPreMigrations(db *gorm.DB) error {
	// Check if holiday_lists table exists and has dates column as text[]
	var columnType string
	result := db.Raw(`
		SELECT data_type 
		FROM information_schema.columns 
		WHERE table_name = 'holiday_lists' AND column_name = 'dates'
	`).Scan(&columnType)

	if result.Error != nil || columnType == "" {
		// Table or column doesn't exist yet, skip pre-migration
		return nil
	}

	// If the column is still ARRAY type, convert it to JSONB
	if columnType == "ARRAY" {
		log.Info("Converting holiday_lists.dates from text[] to jsonb...")

		// Convert text[] to jsonb by wrapping each date in a JSON object
		// Old format: ['2024-12-25', '2024-01-01']
		// New format: [{"date": "2024-12-25", "name": ""}, {"date": "2024-01-01", "name": ""}]
		err := db.Exec(`
			ALTER TABLE holiday_lists 
			ALTER COLUMN dates TYPE JSONB 
			USING (
				SELECT COALESCE(
					jsonb_agg(jsonb_build_object('date', elem, 'name', '')),
					'[]'::jsonb
				)
				FROM unnest(dates) AS elem
			)
		`).Error

		if err != nil {
			// Try simpler approach - just drop the column if conversion fails
			log.Warnf("Complex conversion failed, trying simple approach: %v", err)
			err = db.Exec(`ALTER TABLE holiday_lists DROP COLUMN IF EXISTS dates`).Error
			if err != nil {
				return err
			}
			// GORM will recreate the column with correct type
		}

		log.Info("holiday_lists.dates conversion completed")
	}

	return nil
}
