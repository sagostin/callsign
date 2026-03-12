package freeswitch

import (
	"callsign/config"
	"callsign/services/xmlcache"
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SystemProfiles are protected and cannot be deleted
var SystemProfiles = map[string]bool{
	"internal": true,
	"external": true,
}

// IsSystemProfile checks if a profile name is a protected system profile
func IsSystemProfile(name string) bool {
	return SystemProfiles[strings.ToLower(name)]
}

// XMLCurlRequest represents the parsed request from FreeSWITCH mod_xml_curl
type XMLCurlRequest struct {
	// Common fields
	Section  string `form:"section"`
	TagName  string `form:"tag_name"`
	KeyName  string `form:"key_name"`
	KeyValue string `form:"key_value"`
	Hostname string `form:"hostname"`

	// Directory fields
	User       string `form:"user"`
	Domain     string `form:"domain"`
	Action     string `form:"action"`
	Purpose    string `form:"purpose"`
	SIPProfile string `form:"sip_profile"`

	// SIP Auth fields
	SIPAuthUsername string `form:"sip_auth_username"`
	SIPAuthRealm    string `form:"sip_auth_realm"`
	SIPContactUser  string `form:"sip_contact_user"`
	SIPContactHost  string `form:"sip_contact_host"`
	SIPUserAgent    string `form:"sip_user_agent"`
	IP              string `form:"ip"`

	// Dialplan fields
	Context           string `form:"context"`
	DestinationNumber string `form:"destination_number"`
	CallerIDName      string `form:"caller_id_name"`
	CallerIDNumber    string `form:"caller_id_number"`
	ChannelUUID       string `form:"uuid"`
	ChannelName       string `form:"chan_name"`
	NetworkAddr       string `form:"network_addr"`
	ANI               string `form:"ani"`
	RDNIS             string `form:"rdnis"`
	Source            string `form:"source"`

	// FreeSWITCH info
	FreeSwitchHostname string `form:"FreeSWITCH-Hostname"`
	FreeSwitchIPv4     string `form:"FreeSWITCH-IPv4"`
}

// FSHandler handles FreeSWITCH XML CURL requests
type FSHandler struct {
	DB     *gorm.DB
	Config *config.Config
	Cache  *xmlcache.XMLCache
}

// NewFSHandler creates a new FreeSWITCH handler
func NewFSHandler(db *gorm.DB, cfg *config.Config) *FSHandler {
	return &FSHandler{
		DB:     db,
		Config: cfg,
		Cache:  xmlcache.New(),
	}
}

// HandleXMLCurl is the main entry point for mod_xml_curl requests
func (h *FSHandler) HandleXMLCurl(c *fiber.Ctx) error {
	// Parse multipart form data from FreeSWITCH
	// We manually extract only the fields we need to avoid schema warnings
	// from the 100+ extra fields FreeSWITCH sends
	req := XMLCurlRequest{
		// Common fields
		Section:  c.FormValue("section"),
		TagName:  c.FormValue("tag_name"),
		KeyName:  c.FormValue("key_name"),
		KeyValue: c.FormValue("key_value"),
		Hostname: c.FormValue("hostname"),

		// Directory fields
		User:       c.FormValue("user"),
		Domain:     c.FormValue("domain"),
		Action:     c.FormValue("action"),
		Purpose:    c.FormValue("purpose"),
		SIPProfile: c.FormValue("sip_profile"),

		// SIP Auth fields
		SIPAuthUsername: c.FormValue("sip_auth_username"),
		SIPAuthRealm:    c.FormValue("sip_auth_realm"),
		SIPContactUser:  c.FormValue("sip_contact_user"),
		SIPContactHost:  c.FormValue("sip_contact_host"),
		SIPUserAgent:    c.FormValue("sip_user_agent"),
		IP:              c.FormValue("ip"),

		// Dialplan fields
		Context:           c.FormValue("context"),
		DestinationNumber: c.FormValue("destination_number"),
		CallerIDName:      c.FormValue("caller_id_name"),
		CallerIDNumber:    c.FormValue("caller_id_number"),
		ChannelUUID:       c.FormValue("uuid"),
		ChannelName:       c.FormValue("chan_name"),
		NetworkAddr:       c.FormValue("network_addr"),
		ANI:               c.FormValue("ani"),
		RDNIS:             c.FormValue("rdnis"),
		Source:            c.FormValue("source"),

		// FreeSWITCH info
		FreeSwitchHostname: c.FormValue("FreeSWITCH-Hostname"),
		FreeSwitchIPv4:     c.FormValue("FreeSWITCH-IPv4"),
	}

	// Log the request with all relevant fields
	log.WithFields(log.Fields{
		"section":     req.Section,
		"tag_name":    req.TagName,
		"key_name":    req.KeyName,
		"key_value":   req.KeyValue,
		"user":        req.User,
		"domain":      req.Domain,
		"action":      req.Action,
		"purpose":     req.Purpose,
		"context":     req.Context,
		"dest":        req.DestinationNumber,
		"hostname":    req.FreeSwitchHostname,
		"sip_profile": req.SIPProfile,
	}).Debug("XML CURL request received")

	// Get hostname (use FreeSWITCH-Hostname or hostname field)
	hostname := req.FreeSwitchHostname
	if hostname == "" {
		hostname = req.Hostname
	}

	var xml string

	// Dispatch based on section
	switch req.Section {
	case "directory":
		xml = h.handleDirectory(&req)
	case "configuration":
		xml = h.handleConfiguration(&req, hostname)
	case "dialplan":
		xml = h.handleDialplan(&req)
	case "phrases":
		xml = h.handlePhrases(&req)
	default:
		log.Warnf("Unknown XML CURL section: %s", req.Section)
		return h.sendNotFound(c)
	}

	// If no XML was generated, return not found
	if xml == "" {
		return h.sendNotFound(c)
	}

	// Send XML response
	c.Set("Content-Type", "text/xml")
	return c.SendString(xml)
}

// sendNotFound sends a "not found" XML response
// This tells FreeSWITCH to fall back to static config files
func (h *FSHandler) sendNotFound(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/xml")
	return c.SendString(`<?xml version="1.0" encoding="UTF-8"?>
<document type="freeswitch/xml">
  <section name="result">
    <result status="not found"/>
  </section>
</document>`)
}

// FlushCache handles cache invalidation requests
func (h *FSHandler) FlushCache(c *fiber.Ctx) error {
	pattern := c.Query("pattern")
	key := c.Query("key")

	var count int

	if key != "" {
		h.Cache.Delete(key)
		count = 1
	} else if pattern != "" {
		count = h.Cache.DeleteByPattern(pattern)
	} else {
		h.Cache.Flush()
		count = -1 // Indicates full flush
	}

	return c.JSON(fiber.Map{
		"message":       "Cache flushed",
		"items_deleted": count,
	})
}

// CacheStats returns cache statistics
func (h *FSHandler) CacheStats(c *fiber.Ctx) error {
	return c.JSON(h.Cache.Stats())
}

// AuthMiddleware provides Basic Auth for FreeSWITCH endpoints
func FreeSwitchAuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for API key in config
		expectedKey := cfg.FreeSwitchAPIKey
		if expectedKey == "" {
			// No API key configured, allow all requests
			log.Debug("FreeSWITCH auth: No API key configured, allowing request")
			return c.Next()
		}

		// Allow localhost connections without auth for internal FreeSWITCH
		remoteAddr := c.IP()
		if strings.HasPrefix(remoteAddr, "127.0.0.1") ||
			strings.HasPrefix(remoteAddr, "::1") ||
			strings.HasPrefix(remoteAddr, "localhost") ||
			strings.HasPrefix(remoteAddr, "[::1]") {
			log.WithField("remote", remoteAddr).Debug("FreeSWITCH auth: Localhost connection, allowing request")
			return c.Next()
		}

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.WithFields(log.Fields{
				"remote": remoteAddr,
				"path":   c.Path(),
			}).Warn("FreeSWITCH auth: Missing Authorization header")
			return unauthorized(c)
		}

		// Parse Basic Auth
		const prefix = "Basic "
		if !strings.HasPrefix(authHeader, prefix) {
			log.Warn("FreeSWITCH auth: Invalid Authorization header format")
			return unauthorized(c)
		}

		decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
		if err != nil {
			log.Warn("FreeSWITCH auth: Failed to decode Authorization header")
			return unauthorized(c)
		}

		credentials := string(decoded)
		colonIdx := strings.Index(credentials, ":")
		if colonIdx < 0 {
			log.Warn("FreeSWITCH auth: Invalid credentials format")
			return unauthorized(c)
		}

		// Password is the API key
		password := credentials[colonIdx+1:]
		if password != expectedKey {
			log.WithField("remote", remoteAddr).Warn("FreeSWITCH auth: Invalid API key")
			return unauthorized(c)
		}

		log.Debug("FreeSWITCH auth: Authentication successful")
		return c.Next()
	}
}

func unauthorized(c *fiber.Ctx) error {
	c.Set("WWW-Authenticate", `Basic realm="FreeSWITCH API"`)
	return c.Status(http.StatusUnauthorized).SendString("Unauthorized")
}

// Default cache TTLs
var CacheTTL = struct {
	Configuration time.Duration
	Directory     time.Duration
	Dialplan      time.Duration
}{
	Configuration: 1 * time.Hour,
	Directory:     5 * time.Minute,
	Dialplan:      30 * time.Minute,
}

// ==================
// File Browser Handlers for Config Inspector
// ==================

// FileEntry represents a file or directory in the config file browser
type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size,omitempty"`
}

// ListConfigDirectory lists files and directories in the FreeSWITCH config directory
func (h *FSHandler) ListConfigDirectory(c *fiber.Ctx) error {
	basePath := h.Config.FreeSwitchConfPath
	relativePath := c.Query("path")

	// Clean and validate the path to prevent directory traversal
	if relativePath != "" {
		relativePath = filepath.Clean(relativePath)
		if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid path"})
		}
	}

	targetPath := filepath.Join(basePath, relativePath)

	// Verify path is within the base path
	if !strings.HasPrefix(targetPath, basePath) {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Check if path exists
	info, err := os.Stat(targetPath)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Path not found"})
	}

	if !info.IsDir() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Not a directory"})
	}

	// Read directory contents
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read directory"})
	}

	var files []FileEntry
	for _, entry := range entries {
		entryInfo, _ := entry.Info()
		file := FileEntry{
			Name:  entry.Name(),
			Path:  filepath.Join(relativePath, entry.Name()),
			IsDir: entry.IsDir(),
		}
		if entryInfo != nil && !entry.IsDir() {
			file.Size = entryInfo.Size()
		}
		files = append(files, file)
	}

	// Sort: directories first, then alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	return c.JSON(fiber.Map{
		"path":  relativePath,
		"files": files,
	})
}

// ReadConfigFile returns the content of a config file
func (h *FSHandler) ReadConfigFile(c *fiber.Ctx) error {
	basePath := h.Config.FreeSwitchConfPath
	relativePath := c.Query("path")

	if relativePath == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Path required"})
	}

	// Clean and validate the path
	relativePath = filepath.Clean(relativePath)
	if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid path"})
	}

	targetPath := filepath.Join(basePath, relativePath)

	// Verify path is within the base path
	if !strings.HasPrefix(targetPath, basePath) {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Check if file exists
	info, err := os.Stat(targetPath)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	if info.IsDir() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Cannot read directory"})
	}

	// Limit file size to prevent large memory usage (max 1MB)
	if info.Size() > 1024*1024 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "File too large"})
	}

	// Read file content
	content, err := os.ReadFile(targetPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read file"})
	}

	return c.JSON(fiber.Map{
		"path":    relativePath,
		"name":    info.Name(),
		"size":    info.Size(),
		"content": string(content),
	})
}
