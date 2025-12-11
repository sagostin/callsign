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

	err := db.AutoMigrate(
		// Core models
		&User{},
		&Tenant{},
		&TenantProfile{},

		// Extension/Directory models
		&Extension{},
		&ExtensionSetting{},

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

		// Paging
		&PageGroup{},
		&PageGroupDestination{},

		// Provisioning
		&ProvisioningTemplate{},
		&ProvisioningVariable{},

		// Call Recording & Transcription
		&CallRecording{},
		&Transcription{},
		&TranscriptionSegment{},
		&TranscriptionConfig{},
		&RecordingConfig{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database migrations completed")
	return nil
}
