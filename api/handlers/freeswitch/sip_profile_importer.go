package freeswitch

import (
	"callsign/models"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ProfileImporter handles importing SIP profiles from XML files into the database
type ProfileImporter struct {
	ProfilesPath string
	DB           *gorm.DB
}

// NewProfileImporter creates a new profile importer
func NewProfileImporter(profilesPath string, db *gorm.DB) *ProfileImporter {
	return &ProfileImporter{
		ProfilesPath: profilesPath,
		DB:           db,
	}
}

// XMLProfile represents a parsed SIP profile from XML
type XMLProfile struct {
	XMLName  xml.Name    `xml:"profile"`
	Name     string      `xml:"name,attr"`
	Aliases  XMLAliases  `xml:"aliases"`
	Gateways XMLGateways `xml:"gateways"`
	Domains  XMLDomains  `xml:"domains"`
	Settings XMLSettings `xml:"settings"`
}

type XMLAliases struct {
	Aliases []XMLAlias `xml:"alias"`
}

type XMLAlias struct {
	Name string `xml:"name,attr"`
}

type XMLGateways struct {
	Gateways []XMLGateway `xml:"gateway"`
}

type XMLGateway struct {
	Name   string     `xml:"name,attr"`
	Params []XMLParam `xml:"param"`
}

type XMLDomains struct {
	Domains []XMLDomain `xml:"domain"`
}

type XMLDomain struct {
	Name  string `xml:"name,attr"`
	Alias string `xml:"alias,attr"`
	Parse string `xml:"parse,attr"`
}

type XMLSettings struct {
	Params []XMLParam `xml:"param"`
}

type XMLParam struct {
	Name    string `xml:"name,attr"`
	Value   string `xml:"value,attr"`
	Enabled string `xml:"enabled,attr"` // Optional: "true", "false", or empty
}

// ImportProfilesOnBoot imports SIP profiles from disk that don't exist in the database
// This allows adding new profiles by placing XML files in the sip_profiles directory
// Once imported, the database becomes the source of truth and profiles are synced back to files
func (i *ProfileImporter) ImportProfilesOnBoot() error {
	log.Info("Checking for new SIP profiles to import from disk...")

	// Find all XML files in sip_profiles directory
	files, err := filepath.Glob(filepath.Join(i.ProfilesPath, "*.xml"))
	if err != nil {
		return fmt.Errorf("failed to list profile files: %w", err)
	}

	if len(files) == 0 {
		log.Debug("No SIP profile XML files found in " + i.ProfilesPath)
		return nil
	}

	// Get existing profile names from DB
	var existingProfiles []models.SIPProfile
	i.DB.Select("profile_name").Find(&existingProfiles)
	existingNames := make(map[string]bool)
	for _, p := range existingProfiles {
		existingNames[p.ProfileName] = true
	}

	imported := 0
	for _, file := range files {
		// Skip certain files that aren't profiles
		baseName := filepath.Base(file)
		if strings.HasPrefix(baseName, ".") || strings.Contains(baseName, "example") {
			continue
		}

		profile, err := i.parseProfileFile(file)
		if err != nil {
			log.WithError(err).WithField("file", file).Warn("Failed to parse profile file")
			continue
		}

		if profile == nil {
			continue
		}

		// Skip if already exists in database
		if existingNames[profile.Name] {
			log.WithField("profile", profile.Name).Debug("Profile already exists in database, skipping")
			continue
		}

		if err := i.importProfile(profile); err != nil {
			log.WithError(err).WithField("profile", profile.Name).Warn("Failed to import profile")
			continue
		}

		imported++
		log.WithField("profile", profile.Name).Info("Imported new SIP profile from file")
	}

	if imported > 0 {
		log.WithField("count", imported).Info("Completed SIP profile import")
	} else {
		log.Debug("No new SIP profiles to import")
	}
	return nil
}

// parseProfileFile reads and parses an XML profile file
func (i *ProfileImporter) parseProfileFile(filePath string) (*XMLProfile, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Handle FreeSWITCH variables like $${var} - replace with placeholder or defaults
	content := string(data)
	content = i.expandVariables(content)

	var profile XMLProfile
	if err := xml.Unmarshal([]byte(content), &profile); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	if profile.Name == "" {
		return nil, fmt.Errorf("profile has no name attribute")
	}

	return &profile, nil
}

// expandVariables replaces FreeSWITCH $${var} variables with defaults
func (i *ProfileImporter) expandVariables(content string) string {
	// Common FreeSWITCH variables and their defaults
	replacements := map[string]string{
		"$${local_ip_v4}":          "auto",
		"$${external_rtp_ip}":      "auto-nat",
		"$${external_sip_ip}":      "auto-nat",
		"$${domain}":               "$${domain}",
		"$${hold_music}":           "local_stream://moh",
		"$${global_codec_prefs}":   "OPUS,G722,PCMU,PCMA",
		"$${outbound_codec_prefs}": "OPUS,G722,PCMU,PCMA",
		"$${internal_ssl_enable}":  "false",
		"$${external_ssl_enable}":  "false",
		"$${internal_ssl_dir}":     "",
		"$${external_ssl_dir}":     "",
		"$${sip_tls_version}":      "tlsv1.2",
		"$${sip_tls_ciphers}":      "",
		"$${presence_privacy}":     "false",
		"$${dsn}":                  "",
		"$${recordings_dir}":       "/var/lib/freeswitch/recordings",
	}

	for variable, defaultVal := range replacements {
		content = strings.ReplaceAll(content, variable, defaultVal)
	}

	return content
}

// importProfile creates database records for a parsed profile
func (i *ProfileImporter) importProfile(xmlProfile *XMLProfile) error {
	// Start a transaction
	return i.DB.Transaction(func(tx *gorm.DB) error {
		// Create the profile
		profile := &models.SIPProfile{
			ProfileName: xmlProfile.Name,
			Description: fmt.Sprintf("Imported from %s.xml", xmlProfile.Name),
			Enabled:     true,
		}

		if err := tx.Create(profile).Error; err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		// Import settings
		for _, param := range xmlProfile.Settings.Params {
			// Check if setting is enabled
			enabled := true
			if param.Enabled == "false" {
				enabled = false
			}

			setting := &models.SIPProfileSetting{
				SIPProfileUUID: profile.UUID,
				SettingName:    param.Name,
				SettingValue:   param.Value,
				Enabled:        enabled,
			}

			if err := tx.Create(setting).Error; err != nil {
				log.WithError(err).WithField("setting", param.Name).Warn("Failed to import setting")
			}
		}

		// Import domains
		for _, xmlDomain := range xmlProfile.Domains.Domains {
			domain := &models.SIPProfileDomain{
				SIPProfileUUID: profile.UUID,
				DomainName:     xmlDomain.Name,
				Alias:          xmlDomain.Alias == "true",
				Parse:          xmlDomain.Parse == "true",
			}

			if err := tx.Create(domain).Error; err != nil {
				log.WithError(err).WithField("domain", xmlDomain.Name).Warn("Failed to import domain")
			}
		}

		// Note: We don't import static gateways from XML
		// Gateways should be created via the UI/API and served dynamically

		log.WithFields(log.Fields{
			"profile":  xmlProfile.Name,
			"settings": len(xmlProfile.Settings.Params),
			"domains":  len(xmlProfile.Domains.Domains),
		}).Debug("Profile imported to database")

		return nil
	})
}

// SyncProfilesToFiles writes all database profiles to XML files
// Call this after any profile modification
func (i *ProfileImporter) SyncProfilesToFiles() error {
	var profiles []models.SIPProfile
	if err := i.DB.Preload("Settings").Preload("Domains").Find(&profiles).Error; err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	writer := NewProfileWriter(i.ProfilesPath)

	for _, profile := range profiles {
		if !profile.Enabled {
			// Remove disabled profile files
			if err := writer.DeleteProfile(profile.ProfileName); err != nil {
				log.WithError(err).WithField("profile", profile.ProfileName).Warn("Failed to delete disabled profile file")
			}
			continue
		}

		if err := writer.WriteProfile(&profile, profile.Settings, profile.Domains); err != nil {
			log.WithError(err).WithField("profile", profile.ProfileName).Warn("Failed to write profile file")
			continue
		}
	}

	log.WithField("count", len(profiles)).Debug("Synced profiles to disk")
	return nil
}

// EnsureProfilesPath creates the profiles directory if it doesn't exist
func (i *ProfileImporter) EnsureProfilesPath() error {
	return os.MkdirAll(i.ProfilesPath, 0755)
}
