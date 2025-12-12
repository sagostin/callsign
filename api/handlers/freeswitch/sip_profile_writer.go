package freeswitch

import (
	"callsign/models"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Default SIP Profile path
const DefaultSIPProfilesPath = "/etc/freeswitch/sip_profiles"

// ProfileWriter handles writing SIP profile XML files to disk
// SIP profiles must be on disk because sofia.conf loads them via X-PRE-PROCESS
type ProfileWriter struct {
	ProfilesPath string
}

// NewProfileWriter creates a profile writer with the given path
func NewProfileWriter(path string) *ProfileWriter {
	if path == "" {
		path = DefaultSIPProfilesPath
	}
	return &ProfileWriter{ProfilesPath: path}
}

// WriteProfile generates and writes a SIP profile XML file to disk
func (w *ProfileWriter) WriteProfile(profile *models.SIPProfile, settings []models.SIPProfileSetting, domains []models.SIPProfileDomain) error {
	xml := w.GenerateProfileXML(profile, settings, domains)

	filename := profile.ProfileName + ".xml"
	fullPath := filepath.Join(w.ProfilesPath, filename)

	// Ensure directory exists
	if err := os.MkdirAll(w.ProfilesPath, 0755); err != nil {
		log.WithError(err).WithField("path", w.ProfilesPath).Error("Failed to create SIP profiles directory")
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(fullPath, []byte(xml), 0644); err != nil {
		log.WithError(err).WithField("path", fullPath).Error("Failed to write SIP profile XML")
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	log.WithFields(log.Fields{
		"profile": profile.ProfileName,
		"path":    fullPath,
	}).Info("SIP profile XML written to disk")

	return nil
}

// DeleteProfile removes a SIP profile XML file (protected profiles cannot be deleted)
func (w *ProfileWriter) DeleteProfile(profileName string) error {
	if IsSystemProfile(profileName) {
		return fmt.Errorf("cannot delete protected system profile: %s", profileName)
	}

	filename := profileName + ".xml"
	fullPath := filepath.Join(w.ProfilesPath, filename)

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete profile file: %w", err)
	}

	log.WithField("profile", profileName).Info("SIP profile XML deleted from disk")
	return nil
}

// GenerateProfileXML creates the XML content for a SIP profile
// Note: Gateways are served dynamically via directory (purpose=gateways)
// so we just include an empty gateways section here
func (w *ProfileWriter) GenerateProfileXML(profile *models.SIPProfile, settings []models.SIPProfileSetting, domains []models.SIPProfileDomain) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`<profile name="%s">`, xmlEscape(profile.ProfileName)))
	b.WriteString("\n")

	// Description comment if provided
	if profile.Description != "" {
		b.WriteString(fmt.Sprintf("\t<!-- %s -->\n", xmlEscape(profile.Description)))
	}

	// Aliases
	b.WriteString("\t<aliases>\n")
	b.WriteString("\t</aliases>\n\n")

	// Gateways section - empty because gateways are served via directory (purpose=gateways)
	// when the domain has parse=true
	b.WriteString("\t<!-- Gateways are served dynamically via directory (purpose=gateways) -->\n")
	b.WriteString("\t<gateways>\n")
	b.WriteString("\t</gateways>\n\n")

	// Domains with parse=true to enable dynamic gateway loading
	b.WriteString("\t<domains>\n")
	if len(domains) > 0 {
		for _, domain := range domains {
			aliasStr := "false"
			if domain.Alias {
				aliasStr = "true"
			}
			parseStr := "false"
			if domain.Parse {
				parseStr = "true"
			}
			b.WriteString(fmt.Sprintf("\t\t<domain name=\"%s\" alias=\"%s\" parse=\"%s\"/>\n",
				xmlEscape(domain.DomainName), aliasStr, parseStr))
		}
	} else {
		// Default domain configuration with parse=true for dynamic gateway loading
		b.WriteString("\t\t<domain name=\"all\" alias=\"true\" parse=\"true\"/>\n")
	}
	b.WriteString("\t</domains>\n\n")

	// Settings
	b.WriteString("\t<settings>\n")

	for _, s := range settings {
		if !s.Enabled {
			// Write as disabled param
			b.WriteString(fmt.Sprintf("\t\t<param name=\"%s\" value=\"%s\" enabled=\"false\"/>\n",
				xmlEscape(s.SettingName),
				xmlEscape(s.SettingValue)))
		} else {
			b.WriteString(fmt.Sprintf("\t\t<param name=\"%s\" value=\"%s\"/>\n",
				xmlEscape(s.SettingName),
				xmlEscape(s.SettingValue)))
		}
	}

	b.WriteString("\t</settings>\n")
	b.WriteString("</profile>\n")

	return b.String()
}

// WriteAllProfiles writes all enabled profiles to disk
func (w *ProfileWriter) WriteAllProfiles(profiles []models.SIPProfile) error {
	var errs []error

	for i := range profiles {
		if err := w.WriteProfile(&profiles[i], profiles[i].Settings, profiles[i].Domains); err != nil {
			errs = append(errs, fmt.Errorf("profile %s: %w", profiles[i].ProfileName, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to write %d profiles: %v", len(errs), errs)
	}

	return nil
}
