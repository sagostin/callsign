package featurecodes

import (
	"callsign/models"
	"callsign/services/esl"
	"fmt"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	ServiceName    = "featurecodes"
	ServiceAddress = "127.0.0.6:9001"
)

// Service implements the feature code ESL module
type Service struct {
	*esl.BaseService
	db *gorm.DB
}

// New creates a new feature code service
func New(db *gorm.DB) *Service {
	return &Service{
		BaseService: esl.NewBaseService(ServiceName, ServiceAddress),
		db:          db,
	}
}

// Init initializes the feature code service
func (s *Service) Init(manager *esl.Manager) error {
	if err := s.BaseService.Init(manager); err != nil {
		return err
	}
	log.Info("Feature codes service initialized")
	return nil
}

// Handle processes incoming feature code calls
func (s *Service) Handle(conn *eventsocket.Connection) {
	defer conn.Close()

	// Connect and get channel info
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("FeatureCodes: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	destNumber := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")
	tenantIDStr := ev.Get("variable_tenant_id")
	callerName := ev.Get("Caller-Caller-ID-Name")

	logger := log.WithFields(log.Fields{
		"uuid":   uuid,
		"caller": callerID,
		"code":   destNumber,
		"domain": domain,
	})
	logger.Info("FeatureCodes: processing feature code")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer the call
	conn.Execute("answer", "", true)

	// Parse tenant ID
	var tenantID uint = 1
	if tenantIDStr != "" {
		fmt.Sscanf(tenantIDStr, "%d", &tenantID)
	}

	// Look up feature code (supports regex matching)
	fc, err := models.GetFeatureCode(s.db, tenantID, destNumber)
	if err != nil {
		logger.Warnf("Feature code not found: %s", destNumber)
		conn.Execute("playback", "ivr/ivr-invalid_selection.wav", true)
		conn.Execute("hangup", "NORMAL_CLEARING", true)
		return
	}

	// Get regex captures if any
	_, captures := fc.MatchesDialedNumber(destNumber)

	// Create execution context
	ctx := &ExecutionContext{
		Conn:        conn,
		FeatureCode: fc,
		CallerID:    callerID,
		CallerName:  callerName,
		Domain:      domain,
		TenantID:    tenantID,
		Captures:    captures,
		UUID:        uuid,
		DB:          s.db,
		Service:     s,
	}

	// Execute the feature code action
	s.executeFeatureCode(ctx)
}

// ExecutionContext holds all context for feature code execution
type ExecutionContext struct {
	Conn        *eventsocket.Connection
	FeatureCode *models.FeatureCode
	CallerID    string
	CallerName  string
	Domain      string
	TenantID    uint
	Captures    map[string]string // Regex captures from code match
	UUID        string
	DB          *gorm.DB
	Service     *Service
}

// GetCapture returns a captured value from regex match
func (ctx *ExecutionContext) GetCapture(name string) string {
	if ctx.Captures == nil {
		return ""
	}
	return ctx.Captures[name]
}

// executeFeatureCode routes to the appropriate handler
func (s *Service) executeFeatureCode(ctx *ExecutionContext) {
	logger := log.WithFields(log.Fields{
		"code":   ctx.FeatureCode.Code,
		"action": ctx.FeatureCode.Action,
	})
	logger.Info("Executing feature code")

	switch ctx.FeatureCode.Action {
	case models.FCActionVoicemail:
		handleVoicemail(ctx)

	case models.FCActionCallForward:
		handleCallForward(ctx)

	case models.FCActionDND:
		handleDND(ctx)

	case models.FCActionCallFlowToggle:
		handleCallFlowToggle(ctx)

	case models.FCActionPark:
		handlePark(ctx)

	case models.FCActionParkSlot:
		handleParkSlot(ctx)

	case models.FCActionParkRetrieve:
		handleParkRetrieve(ctx)

	case models.FCActionPickup:
		handlePickup(ctx)

	case models.FCActionIntercom:
		handleIntercom(ctx)

	case models.FCActionPageGroup:
		handlePageGroup(ctx)

	case models.FCActionTransfer:
		handleTransfer(ctx)

	case models.FCActionWebhook:
		handleWebhook(ctx)

	case models.FCActionLua:
		handleLua(ctx)

	case models.FCActionRecord:
		handleRecord(ctx)

	case models.FCActionCustom:
		handleCustom(ctx)

	default:
		ctx.Conn.Execute("playback", "ivr/ivr-invalid_selection.wav", true)
	}

	ctx.Conn.Execute("hangup", "NORMAL_CLEARING", true)
}

// clearDirectoryCache clears the directory cache for a domain
func (s *Service) clearDirectoryCache(domain string) {
	manager := s.Manager()
	if manager != nil && manager.Client != nil {
		manager.Client.API(fmt.Sprintf("xml_flush_cache directory %s", domain))
	}
}

// clearDialplanCache clears the dialplan cache for a domain
func (s *Service) clearDialplanCache(domain string) {
	manager := s.Manager()
	if manager != nil && manager.Client != nil {
		manager.Client.API(fmt.Sprintf("xml_flush_cache dialplan %s", domain))
	}
}

// sendPresenceNotify sends a BLF presence update
func (s *Service) sendPresenceNotify(domain, user, state string) {
	manager := s.Manager()
	if manager != nil && manager.Client != nil {
		// Send presence event
		manager.Client.API(fmt.Sprintf(
			"presence in %s@%s|%s",
			user, domain, state,
		))
	}
}
