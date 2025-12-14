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

	"github.com/kataras/iris/v12"
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
func (h *FSHandler) HandleXMLCurl(ctx iris.Context) {
	// Parse multipart form data from FreeSWITCH
	// We manually extract only the fields we need to avoid schema warnings
	// from the 100+ extra fields FreeSWITCH sends
	req := XMLCurlRequest{
		// Common fields
		Section:  ctx.FormValue("section"),
		TagName:  ctx.FormValue("tag_name"),
		KeyName:  ctx.FormValue("key_name"),
		KeyValue: ctx.FormValue("key_value"),
		Hostname: ctx.FormValue("hostname"),

		// Directory fields
		User:       ctx.FormValue("user"),
		Domain:     ctx.FormValue("domain"),
		Action:     ctx.FormValue("action"),
		Purpose:    ctx.FormValue("purpose"),
		SIPProfile: ctx.FormValue("sip_profile"),

		// SIP Auth fields
		SIPAuthUsername: ctx.FormValue("sip_auth_username"),
		SIPAuthRealm:    ctx.FormValue("sip_auth_realm"),
		SIPContactUser:  ctx.FormValue("sip_contact_user"),
		SIPContactHost:  ctx.FormValue("sip_contact_host"),
		SIPUserAgent:    ctx.FormValue("sip_user_agent"),
		IP:              ctx.FormValue("ip"),

		// Dialplan fields
		Context:           ctx.FormValue("context"),
		DestinationNumber: ctx.FormValue("destination_number"),
		CallerIDName:      ctx.FormValue("caller_id_name"),
		CallerIDNumber:    ctx.FormValue("caller_id_number"),
		ChannelUUID:       ctx.FormValue("uuid"),
		ChannelName:       ctx.FormValue("chan_name"),
		NetworkAddr:       ctx.FormValue("network_addr"),
		ANI:               ctx.FormValue("ani"),
		RDNIS:             ctx.FormValue("rdnis"),
		Source:            ctx.FormValue("source"),

		// FreeSWITCH info
		FreeSwitchHostname: ctx.FormValue("FreeSWITCH-Hostname"),
		FreeSwitchIPv4:     ctx.FormValue("FreeSWITCH-IPv4"),
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
		h.sendNotFound(ctx)
		return
	}

	// If no XML was generated, return not found
	if xml == "" {
		h.sendNotFound(ctx)
		return
	}

	// Send XML response
	ctx.ContentType("text/xml")
	ctx.WriteString(xml)
}

// sendNotFound sends a "not found" XML response
// This tells FreeSWITCH to fall back to static config files
func (h *FSHandler) sendNotFound(ctx iris.Context) {
	ctx.ContentType("text/xml")
	ctx.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<document type="freeswitch/xml">
  <section name="result">
    <result status="not found"/>
  </section>
</document>`)
}

// FlushCache handles cache invalidation requests
func (h *FSHandler) FlushCache(ctx iris.Context) {
	pattern := ctx.URLParam("pattern")
	key := ctx.URLParam("key")

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

	ctx.JSON(iris.Map{
		"message":       "Cache flushed",
		"items_deleted": count,
	})
}

// CacheStats returns cache statistics
func (h *FSHandler) CacheStats(ctx iris.Context) {
	ctx.JSON(h.Cache.Stats())
}

// AuthMiddleware provides Basic Auth for FreeSWITCH endpoints
func FreeSwitchAuthMiddleware(cfg *config.Config) iris.Handler {
	return func(ctx iris.Context) {
		// Check for API key in config
		expectedKey := cfg.FreeSwitchAPIKey
		if expectedKey == "" {
			// No API key configured, allow all requests
			log.Debug("FreeSWITCH auth: No API key configured, allowing request")
			ctx.Next()
			return
		}

		// Allow localhost connections without auth for internal FreeSWITCH
		remoteAddr := ctx.RemoteAddr()
		if strings.HasPrefix(remoteAddr, "127.0.0.1") ||
			strings.HasPrefix(remoteAddr, "::1") ||
			strings.HasPrefix(remoteAddr, "localhost") ||
			strings.HasPrefix(remoteAddr, "[::1]") {
			log.WithField("remote", remoteAddr).Debug("FreeSWITCH auth: Localhost connection, allowing request")
			ctx.Next()
			return
		}

		// Get Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			log.WithFields(log.Fields{
				"remote": remoteAddr,
				"path":   ctx.Path(),
			}).Warn("FreeSWITCH auth: Missing Authorization header")
			unauthorized(ctx)
			return
		}

		// Parse Basic Auth
		const prefix = "Basic "
		if !strings.HasPrefix(authHeader, prefix) {
			log.Warn("FreeSWITCH auth: Invalid Authorization header format")
			unauthorized(ctx)
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
		if err != nil {
			log.Warn("FreeSWITCH auth: Failed to decode Authorization header")
			unauthorized(ctx)
			return
		}

		credentials := string(decoded)
		colonIdx := strings.Index(credentials, ":")
		if colonIdx < 0 {
			log.Warn("FreeSWITCH auth: Invalid credentials format")
			unauthorized(ctx)
			return
		}

		// Password is the API key
		password := credentials[colonIdx+1:]
		if password != expectedKey {
			log.WithField("remote", remoteAddr).Warn("FreeSWITCH auth: Invalid API key")
			unauthorized(ctx)
			return
		}

		log.Debug("FreeSWITCH auth: Authentication successful")
		ctx.Next()
	}
}

func unauthorized(ctx iris.Context) {
	ctx.Header("WWW-Authenticate", `Basic realm="FreeSWITCH API"`)
	ctx.StatusCode(http.StatusUnauthorized)
	ctx.WriteString("Unauthorized")
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
func (h *FSHandler) ListConfigDirectory(ctx iris.Context) {
	basePath := h.Config.FreeSwitchConfPath
	relativePath := ctx.URLParam("path")

	// Clean and validate the path to prevent directory traversal
	if relativePath != "" {
		relativePath = filepath.Clean(relativePath)
		if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "Invalid path"})
			return
		}
	}

	targetPath := filepath.Join(basePath, relativePath)

	// Verify path is within the base path
	if !strings.HasPrefix(targetPath, basePath) {
		ctx.StatusCode(http.StatusForbidden)
		ctx.JSON(iris.Map{"error": "Access denied"})
		return
	}

	// Check if path exists
	info, err := os.Stat(targetPath)
	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Path not found"})
		return
	}

	if !info.IsDir() {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Not a directory"})
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to read directory"})
		return
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

	ctx.JSON(iris.Map{
		"path":  relativePath,
		"files": files,
	})
}

// ReadConfigFile returns the content of a config file
func (h *FSHandler) ReadConfigFile(ctx iris.Context) {
	basePath := h.Config.FreeSwitchConfPath
	relativePath := ctx.URLParam("path")

	if relativePath == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Path required"})
		return
	}

	// Clean and validate the path
	relativePath = filepath.Clean(relativePath)
	if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid path"})
		return
	}

	targetPath := filepath.Join(basePath, relativePath)

	// Verify path is within the base path
	if !strings.HasPrefix(targetPath, basePath) {
		ctx.StatusCode(http.StatusForbidden)
		ctx.JSON(iris.Map{"error": "Access denied"})
		return
	}

	// Check if file exists
	info, err := os.Stat(targetPath)
	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "File not found"})
		return
	}

	if info.IsDir() {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Cannot read directory"})
		return
	}

	// Limit file size to prevent large memory usage (max 1MB)
	if info.Size() > 1024*1024 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "File too large"})
		return
	}

	// Read file content
	content, err := os.ReadFile(targetPath)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to read file"})
		return
	}

	ctx.JSON(iris.Map{
		"path":    relativePath,
		"name":    info.Name(),
		"size":    info.Size(),
		"content": string(content),
	})
}
