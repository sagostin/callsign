package freeswitch

import (
	"callsign/models"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ProfileSyncer handles syncing SIP profiles from database to disk XML files.
// Profiles are seeded from built-in defaults (see models.EnsureDefaultProfiles),
// NOT imported from disk. Disk files are written because Sofia loads them via X-PRE-PROCESS.
type ProfileSyncer struct {
	ProfilesPath string
	DB           *gorm.DB
}

// NewProfileSyncer creates a new profile syncer
func NewProfileSyncer(profilesPath string, db *gorm.DB) *ProfileSyncer {
	return &ProfileSyncer{
		ProfilesPath: profilesPath,
		DB:           db,
	}
}

// SyncProfilesToFiles writes all enabled database profiles to XML files on disk.
// Call this after any profile modification or on startup.
func (s *ProfileSyncer) SyncProfilesToFiles() error {
	var profiles []models.SIPProfile
	if err := s.DB.Preload("Settings").Preload("Domains").Where("enabled = ?", true).Find(&profiles).Error; err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	writer := NewProfileWriter(s.ProfilesPath)

	for _, profile := range profiles {
		if err := writer.WriteProfile(&profile, profile.Settings, profile.Domains); err != nil {
			log.WithError(err).WithField("profile", profile.ProfileName).Warn("Failed to write profile file")
			continue
		}
	}

	log.WithField("count", len(profiles)).Debug("Synced profiles to disk")
	return nil
}

// EnsureProfilesPath creates the profiles directory if it doesn't exist
func (s *ProfileSyncer) EnsureProfilesPath() error {
	return os.MkdirAll(s.ProfilesPath, 0755)
}
