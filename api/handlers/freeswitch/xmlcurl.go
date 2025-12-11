package freeswitch

import (
	"callsign/config"
	"callsign/services/xmlcache"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

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
	var req XMLCurlRequest

	// Parse form data (mod_xml_curl sends POST with form data)
	if err := ctx.ReadForm(&req); err != nil {
		log.Warnf("Failed to parse XML CURL request: %v", err)
		h.sendNotFound(ctx)
		return
	}

	// Log the request
	log.WithFields(log.Fields{
		"section":   req.Section,
		"key_value": req.KeyValue,
		"user":      req.User,
		"domain":    req.Domain,
		"action":    req.Action,
		"context":   req.Context,
		"dest":      req.DestinationNumber,
		"hostname":  req.FreeSwitchHostname,
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
			// No API key configured, allow request
			ctx.Next()
			return
		}

		// Get Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(ctx)
			return
		}

		// Parse Basic Auth
		const prefix = "Basic "
		if !strings.HasPrefix(authHeader, prefix) {
			unauthorized(ctx)
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
		if err != nil {
			unauthorized(ctx)
			return
		}

		credentials := string(decoded)
		colonIdx := strings.Index(credentials, ":")
		if colonIdx < 0 {
			unauthorized(ctx)
			return
		}

		// Password is the API key
		password := credentials[colonIdx+1:]
		if password != expectedKey {
			unauthorized(ctx)
			return
		}

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
