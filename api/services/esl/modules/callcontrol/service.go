package callcontrol

import (
	"callsign/models"
	"callsign/services/esl"
	"fmt"
	"math/rand"
	"strings"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	ServiceName    = "callcontrol"
	ServiceAddress = "127.0.0.1:9001"
)

// Service implements the general call control ESL module.
// Handles: direct extension calls, ring group routing, outbound routing.
type Service struct {
	*esl.BaseService
}

// New creates a new call control service
func New() *Service {
	return &Service{
		BaseService: esl.NewBaseService(ServiceName, ServiceAddress),
	}
}

// Init initializes the call control service
func (s *Service) Init(manager *esl.Manager) error {
	if err := s.BaseService.Init(manager); err != nil {
		return err
	}
	log.Info("Call control service initialized")
	return nil
}

// callContext holds all context for a call being handled
type callContext struct {
	conn       *eventsocket.Connection
	manager    *esl.Manager
	db         *gorm.DB
	uuid       string
	callerID   string
	callerName string
	dest       string
	domain     string
	tenantID   uint
	logger     *log.Entry
}

// Handle processes incoming call control connections
func (s *Service) Handle(conn *eventsocket.Connection) {
	defer conn.Close()

	manager := s.Manager()
	if manager == nil {
		log.Error("Callcontrol: manager not initialized")
		return
	}

	// Connect and get channel info
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Callcontrol: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	callerName := ev.Get("Caller-Caller-ID-Name")
	dest := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")
	tenantIDStr := ev.Get("variable_tenant_id")
	ringGroupUUID := ev.Get("variable_ring_group_uuid")

	logger := log.WithFields(log.Fields{
		"uuid":        uuid,
		"caller":      callerID,
		"destination": dest,
		"domain":      domain,
	})
	logger.Info("Callcontrol: handling call")

	conn.Send("linger")
	conn.Send("myevents")

	ctx := &callContext{
		conn:       conn,
		manager:    manager,
		db:         manager.DB,
		uuid:       uuid,
		callerID:   callerID,
		callerName: callerName,
		dest:       dest,
		domain:     domain,
		tenantID:   parseTenantID(tenantIDStr),
		logger:     logger,
	}

	// Broadcast ringing event via WebSocket
	manager.NotifyCallEvent(ctx.tenantID, "ringing", map[string]interface{}{
		"uuid":        uuid,
		"caller":      callerID,
		"destination": dest,
		"domain":      domain,
	})

	// Route: ring group takes priority, then extension, then outbound
	if ringGroupUUID != "" {
		s.handleRingGroupCall(ctx, ringGroupUUID)
		return
	}

	s.handleExtensionCall(ctx)
}

// ========== Extension Call Routing ==========

// handleExtensionCall routes a call to an internal extension with DND,
// forwarding, and voicemail fallback.
func (s *Service) handleExtensionCall(ctx *callContext) {
	var ext models.Extension
	err := ctx.db.
		Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
		Where("tenants.domain = ? AND extensions.extension = ? AND extensions.enabled = ?",
			ctx.domain, ctx.dest, true).
		First(&ext).Error

	if err != nil {
		// Not an internal extension — try outbound routing
		s.handleOutboundCall(ctx)
		return
	}

	// --- DND ---
	if ext.DoNotDisturb {
		ctx.logger.Info("DND active, routing to voicemail")
		if ext.VoicemailEnabled {
			ctx.conn.Execute("answer", "", true)
			ctx.conn.Execute("voicemail", fmt.Sprintf("default %s %s", ctx.domain, ctx.dest), true)
		} else {
			ctx.conn.Execute("respond", "486 Busy Here", false)
		}
		return
	}

	// --- Unconditional forwarding ---
	if ext.ForwardAllEnabled && ext.ForwardAllDestination != "" {
		ctx.logger.Infof("Forward all to %s", ext.ForwardAllDestination)
		s.resolveAndBridge(ctx, ext.ForwardAllDestination, 1)
		return
	}

	// --- Follow-me: ring extension + external simultaneously ---
	if ext.FollowMeEnabled && ext.FollowMeDestination != "" {
		ctx.logger.Infof("Follow-me to %s", ext.FollowMeDestination)
		s.handleFollowMe(ctx, &ext)
		return
	}

	// --- Ring extension with fallback ---
	ctx.conn.Execute("set", fmt.Sprintf("call_timeout=%d", ext.CallTimeout), true)
	ctx.conn.Execute("set", "hangup_after_bridge=true", true)
	ctx.conn.Execute("set", "continue_on_fail=true", true)
	ctx.conn.Execute("set", "ringback=${us-ring}", true)

	// Set diversion header for call routing visibility
	ctx.conn.Execute("set", fmt.Sprintf("sip_h_Diversion=<sip:%s@%s>", ctx.dest, ctx.domain), true)

	dialString := fmt.Sprintf("user/%s@%s", ctx.dest, ctx.domain)
	ctx.conn.Execute("bridge", dialString, true)

	// Check bridge result
	cause := s.getBridgeResult(ctx)

	switch cause {
	case "USER_BUSY":
		if ext.ForwardBusyEnabled && ext.ForwardBusyDestination != "" {
			ctx.logger.Infof("Busy forward to %s", ext.ForwardBusyDestination)
			s.resolveAndBridge(ctx, ext.ForwardBusyDestination, 1)
			return
		}
		if ext.VoicemailEnabled {
			ctx.conn.Execute("answer", "", true)
			ctx.conn.Execute("voicemail", fmt.Sprintf("default %s %s", ctx.domain, ctx.dest), true)
		}

	case "NO_ANSWER", "ALLOTTED_TIMEOUT", "NO_USER_RESPONSE":
		if ext.ForwardNoAnswerEnabled && ext.ForwardNoAnswerDestination != "" {
			ctx.logger.Infof("No-answer forward to %s", ext.ForwardNoAnswerDestination)
			s.resolveAndBridge(ctx, ext.ForwardNoAnswerDestination, 1)
			return
		}
		if ext.VoicemailEnabled {
			ctx.conn.Execute("answer", "", true)
			ctx.conn.Execute("voicemail", fmt.Sprintf("default %s %s", ctx.domain, ctx.dest), true)
		}

	case "USER_NOT_REGISTERED":
		if ext.ForwardUserNotRegisteredEnabled && ext.ForwardUserNotRegisteredDestination != "" {
			ctx.logger.Infof("Not registered forward to %s", ext.ForwardUserNotRegisteredDestination)
			s.resolveAndBridge(ctx, ext.ForwardUserNotRegisteredDestination, 1)
			return
		}
		if ext.VoicemailEnabled {
			ctx.conn.Execute("answer", "", true)
			ctx.conn.Execute("voicemail", fmt.Sprintf("default %s %s", ctx.domain, ctx.dest), true)
		}
	}
}

// resolveAndBridge follows up to 3 levels of forward-all chains
func (s *Service) resolveAndBridge(ctx *callContext, dest string, depth int) {
	if depth > 3 {
		ctx.logger.Warn("Forward chain too deep, stopping")
		return
	}

	// Check if the forward destination itself has forwarding set
	var fwdExt models.Extension
	err := ctx.db.
		Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
		Where("tenants.domain = ? AND extensions.extension = ? AND extensions.enabled = ?",
			ctx.domain, dest, true).
		First(&fwdExt).Error

	if err == nil && fwdExt.ForwardAllEnabled && fwdExt.ForwardAllDestination != "" {
		ctx.logger.Infof("Following forward chain: %s -> %s", dest, fwdExt.ForwardAllDestination)
		s.resolveAndBridge(ctx, fwdExt.ForwardAllDestination, depth+1)
		return
	}

	ctx.conn.Execute("set", "call_timeout=30", true)
	ctx.conn.Execute("bridge", fmt.Sprintf("user/%s@%s", dest, ctx.domain), true)
}

// ========== Ring Group Routing ==========

// handleRingGroupCall handles calls routed to ring groups with full feature support
func (s *Service) handleRingGroupCall(ctx *callContext, ringGroupUUID string) {
	var rg models.RingGroup
	if err := ctx.db.Where("uuid = ? AND enabled = ?", ringGroupUUID, true).
		Preload("Destinations", func(db *gorm.DB) *gorm.DB {
			return db.Order("priority ASC")
		}).
		First(&rg).Error; err != nil {
		ctx.logger.Errorf("Ring group %s not found: %v", ringGroupUUID, err)
		ctx.conn.Execute("hangup", "UNALLOCATED_NUMBER", false)
		return
	}

	ctx.logger = ctx.logger.WithField("ring_group", rg.Name)

	// --- Ring group-level forwarding ---
	if rg.ForwardEnabled && rg.ForwardDestination != "" {
		ctx.logger.Infof("Ring group forwarded to %s", rg.ForwardDestination)
		ctx.conn.Execute("set", "call_timeout=30", true)
		ctx.conn.Execute("bridge", fmt.Sprintf("user/%s@%s", rg.ForwardDestination, ctx.domain), true)
		return
	}

	if len(rg.Destinations) == 0 {
		ctx.logger.Warn("Ring group has no destinations")
		ctx.conn.Execute("hangup", "UNALLOCATED_NUMBER", false)
		return
	}

	// --- Caller ID prefix ---
	if rg.CallerIDNamePrefix != "" {
		ctx.conn.Execute("set", fmt.Sprintf("effective_caller_id_name=%s%s",
			rg.CallerIDNamePrefix, ctx.callerName), true)
	}
	if rg.CallerIDNumberPrefix != "" {
		ctx.conn.Execute("set", fmt.Sprintf("effective_caller_id_number=%s%s",
			rg.CallerIDNumberPrefix, ctx.callerID), true)
	}

	// --- Distinctive ring / Alert-Info ---
	if rg.AlertInfo != "" {
		ctx.conn.Execute("set", fmt.Sprintf("sip_h_Alert-Info=%s", rg.AlertInfo), true)
	}

	// --- Greeting playback ---
	if rg.GreetingPath != "" {
		ctx.conn.Execute("playback", rg.GreetingPath, true)
	}

	// --- Ringback tone ---
	ctx.conn.Execute("set", "hangup_after_bridge=true", true)
	ctx.conn.Execute("set", "continue_on_fail=true", true)
	if rg.RingbackTone != "" {
		ctx.conn.Execute("set", fmt.Sprintf("ringback=%s", rg.RingbackTone), true)
	} else {
		ctx.conn.Execute("set", "ringback=${us-ring}", true)
	}

	// --- Diversion header ---
	if rg.Extension != "" {
		ctx.conn.Execute("set", fmt.Sprintf("sip_h_Diversion=<sip:%s@%s>", rg.Extension, ctx.domain), true)
	}

	// --- Filter destinations: skip DND, check follow-me ---
	destinations := s.filterDestinations(ctx, &rg)
	if len(destinations) == 0 {
		ctx.logger.Warn("All destinations skipped (DND/unavailable)")
		s.handleRingGroupTimeout(ctx, &rg)
		return
	}

	// --- Fire custom RING_GROUPS event ---
	s.fireRingGroupEvent(ctx, &rg, "ringing")

	// --- Execute ring strategy ---
	connected := false
	switch rg.Strategy {
	case models.RingStrategySimultaneous:
		connected = s.ringSimultaneous(ctx, destinations, rg.RingTimeout)
	case models.RingStrategySequence:
		connected = s.ringSequence(ctx, destinations, rg.RingTimeout)
	case models.RingStrategyRandom:
		connected = s.ringRandom(ctx, destinations, rg.RingTimeout)
	case models.RingStrategyEnterprise:
		connected = s.ringEnterprise(ctx, destinations, rg.RingTimeout)
	case models.RingStrategyRollover:
		connected = s.ringRollover(ctx, destinations, rg.RingTimeout)
	case models.RingStrategyRoundRobin:
		connected = s.ringRoundRobin(ctx, destinations, rg.RingTimeout)
	default:
		connected = s.ringSimultaneous(ctx, destinations, rg.RingTimeout)
	}

	if !connected {
		// --- Missed call handling ---
		s.fireRingGroupEvent(ctx, &rg, "missed")

		if rg.MissedCallTracking {
			ctx.logger.Info("Ring group: missed call logged")
			// TODO: persist missed call record + email notification if rg.MissedCallEmail != ""
		}

		// --- Timeout destination ---
		s.handleRingGroupTimeout(ctx, &rg)
	} else {
		s.fireRingGroupEvent(ctx, &rg, "answered")
	}
}

// filterDestinations checks each destination's DND status, optionally merges follow-me
func (s *Service) filterDestinations(ctx *callContext, rg *models.RingGroup) []models.RingGroupDestination {
	var filtered []models.RingGroupDestination

	for _, d := range rg.Destinations {
		if d.DestinationType != "extension" {
			// External destinations are always included
			filtered = append(filtered, d)
			continue
		}

		// Check if extension has DND enabled
		if rg.SkipBusyMembers {
			var ext struct {
				DoNotDisturb    bool
				FollowMeEnabled bool
				FollowMeDest    string
			}
			err := ctx.db.Table("extensions").
				Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
				Where("tenants.domain = ? AND extensions.extension = ?", ctx.domain, d.Destination).
				Select("do_not_disturb, follow_me_enabled, follow_me_destination").
				First(&ext).Error

			if err == nil {
				if ext.DoNotDisturb {
					ctx.logger.Debugf("Skipping %s (DND)", d.Destination)
					continue
				}

				// Merge follow-me destinations if enabled
				if rg.FollowMeEnabled && ext.FollowMeEnabled && ext.FollowMeDest != "" {
					ctx.logger.Debugf("Merging follow-me for %s -> %s", d.Destination, ext.FollowMeDest)
					fmDest := models.RingGroupDestination{
						RingGroupID:     d.RingGroupID,
						DestinationType: "external",
						Destination:     ext.FollowMeDest,
						Delay:           d.Delay + 5, // Delay follow-me by 5s
						Timeout:         d.Timeout,
						Priority:        d.Priority + 1,
						PromptConfirm:   true, // Always confirm external
					}
					filtered = append(filtered, d, fmDest)
					continue
				}
			}
		}

		filtered = append(filtered, d)
	}

	return filtered
}

// ========== Ring Strategies ==========

// ringSimultaneous rings all destinations at once (comma-separated)
func (s *Service) ringSimultaneous(ctx *callContext, dests []models.RingGroupDestination, timeout int) bool {
	if timeout <= 0 {
		timeout = 30
	}

	var dialStrings []string
	for _, d := range dests {
		ds := s.buildDialString(d, ctx.domain)
		if ds != "" {
			dialStrings = append(dialStrings, ds)
		}
	}
	if len(dialStrings) == 0 {
		return false
	}

	ctx.conn.Execute("set", fmt.Sprintf("call_timeout=%d", timeout), true)
	ctx.conn.Execute("bridge", strings.Join(dialStrings, ","), true)
	return s.getBridgeResult(ctx) == "" || s.getBridgeResult(ctx) == "SUCCESS"
}

// ringSequence rings destinations one after another
func (s *Service) ringSequence(ctx *callContext, dests []models.RingGroupDestination, defaultTimeout int) bool {
	for _, d := range dests {
		ds := s.buildDialString(d, ctx.domain)
		if ds == "" {
			continue
		}

		timeout := d.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		if timeout <= 0 {
			timeout = 30
		}

		if d.Delay > 0 {
			ctx.conn.Execute("sleep", fmt.Sprintf("%d", d.Delay*1000), true)
		}

		ctx.conn.Execute("set", fmt.Sprintf("call_timeout=%d", timeout), true)
		ctx.conn.Execute("bridge", ds, true)

		cause := s.getBridgeResult(ctx)
		if cause == "" || cause == "SUCCESS" {
			return true
		}
	}
	return false
}

// ringRandom shuffles destinations and rings sequentially
func (s *Service) ringRandom(ctx *callContext, dests []models.RingGroupDestination, timeout int) bool {
	shuffled := make([]models.RingGroupDestination, len(dests))
	copy(shuffled, dests)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return s.ringSequence(ctx, shuffled, timeout)
}

// ringEnterprise rings all simultaneously but with per-member delays
// Uses FreeSWITCH enterprise originate (:_: separator)
func (s *Service) ringEnterprise(ctx *callContext, dests []models.RingGroupDestination, timeout int) bool {
	if timeout <= 0 {
		timeout = 30
	}

	var legs []string
	for _, d := range dests {
		ds := s.buildDialString(d, ctx.domain)
		if ds == "" {
			continue
		}

		// Per-leg variables for enterprise originate
		legTimeout := d.Timeout
		if legTimeout <= 0 {
			legTimeout = timeout
		}
		legVars := fmt.Sprintf("[originate_timeout=%d", legTimeout)
		if d.Delay > 0 {
			legVars += fmt.Sprintf(",originate_delay_start=%d", d.Delay)
		}
		legVars += "]"

		legs = append(legs, legVars+ds)
	}
	if len(legs) == 0 {
		return false
	}

	// Enterprise: each leg gets its own originate thread
	ctx.conn.Execute("set", fmt.Sprintf("call_timeout=%d", timeout), true)
	ctx.conn.Execute("bridge", strings.Join(legs, ":_:"), true)

	cause := s.getBridgeResult(ctx)
	return cause == "" || cause == "SUCCESS"
}

// ringRollover rings sequentially using pipe separator (no per-leg delay)
func (s *Service) ringRollover(ctx *callContext, dests []models.RingGroupDestination, timeout int) bool {
	if timeout <= 0 {
		timeout = 30
	}

	var dialStrings []string
	for _, d := range dests {
		ds := s.buildDialString(d, ctx.domain)
		if ds != "" {
			dialStrings = append(dialStrings, ds)
		}
	}
	if len(dialStrings) == 0 {
		return false
	}

	// Rollover: pipe-separated = try next leg when previous fails
	ctx.conn.Execute("set", fmt.Sprintf("call_timeout=%d", timeout), true)
	ctx.conn.Execute("bridge", strings.Join(dialStrings, "|"), true)

	cause := s.getBridgeResult(ctx)
	return cause == "" || cause == "SUCCESS"
}

// ringRoundRobin distributes calls evenly across agents by rotating through the list.
// It shuffles destinations once per call to ensure fair distribution while maintaining
// a simple rotation-based approach.
func (s *Service) ringRoundRobin(ctx *callContext, dests []models.RingGroupDestination, timeout int) bool {
	if len(dests) == 0 {
		return false
	}
	// Rotate the list to distribute evenly across calls
	rotated := s.rotateDestinations(dests)
	return s.ringSequence(ctx, rotated, timeout)
}

// rotateDestinations rotates the destination list by one position to implement round-robin.
func (s *Service) rotateDestinations(dests []models.RingGroupDestination) []models.RingGroupDestination {
	if len(dests) <= 1 {
		return dests
	}
	rotated := make([]models.RingGroupDestination, len(dests))
	copy(rotated, dests[1:])
	rotated[len(rotated)-1] = dests[0]
	return rotated
}

// ========== Dial String Building ==========

// buildDialString creates a FreeSWITCH dial string for a ring group destination
func (s *Service) buildDialString(d models.RingGroupDestination, domain string) string {
	switch d.DestinationType {
	case "extension":
		return fmt.Sprintf("user/%s@%s", d.Destination, domain)
	case "external":
		return fmt.Sprintf("loopback/%s/%s", d.Destination, domain)
	case "gateway":
		return fmt.Sprintf("sofia/gateway/%s/%s", d.Destination, d.Destination)
	case "sip_uri":
		return d.Destination
	default:
		if d.Destination != "" {
			return fmt.Sprintf("user/%s@%s", d.Destination, domain)
		}
		return ""
	}
}

// ========== Outbound Routing ==========

// handleOutboundCall routes calls to PSTN via outbound routes and gateways
func (s *Service) handleOutboundCall(ctx *callContext) {
	// --- Toll-allow enforcement ---
	var callerExt models.Extension
	callerFound := ctx.db.
		Joins("JOIN tenants ON tenants.id = extensions.tenant_id").
		Where("tenants.domain = ? AND extensions.extension = ?", ctx.domain, ctx.callerID).
		First(&callerExt).Error == nil

	var routes []models.DefaultOutboundRoute
	ctx.db.Where("enabled = ?", true).Order("\"order\" ASC").Find(&routes)

	for _, route := range routes {
		if route.DigitPrefix != "" && !strings.HasPrefix(ctx.dest, route.DigitPrefix) {
			continue
		}
		if len(ctx.dest) < route.DigitMin || len(ctx.dest) > route.DigitMax {
			continue
		}

		// Check toll-allow: if the caller has toll restrictions,
		// verify their allowed classes include this route's class
		if callerFound && callerExt.TollAllow != "" && route.TollAllow != "" {
			allowed := strings.Split(callerExt.TollAllow, ",")
			routeClass := route.TollAllow
			if !containsString(allowed, routeClass) {
				ctx.logger.Infof("Toll-allow denied: caller %s (allow=%s) route %s (class=%s)",
					ctx.callerID, callerExt.TollAllow, route.Name, routeClass)
				continue
			}
		}

		var gw models.Gateway
		if err := ctx.db.Where("id = ? AND enabled = ?", route.GatewayID, true).First(&gw).Error; err != nil {
			continue
		}

		dialDest := ctx.dest
		if route.StripDigits > 0 && len(dialDest) > route.StripDigits {
			dialDest = dialDest[route.StripDigits:]
		}
		if route.PrependDigits != "" {
			dialDest = route.PrependDigits + dialDest
		}

		ctx.logger.WithFields(log.Fields{
			"route":   route.Name,
			"gateway": gw.GatewayName,
			"dest":    dialDest,
		}).Info("Outbound route matched")

		ctx.conn.Execute("set", "hangup_after_bridge=true", true)
		bridgeStr := fmt.Sprintf("sofia/gateway/%s/%s", gw.GatewayName, dialDest)
		ctx.conn.Execute("bridge", bridgeStr, true)

		// Failover gateway
		if route.Gateway2ID != nil {
			cause := s.getBridgeResult(ctx)
			if cause != "" && cause != "SUCCESS" {
				var gw2 models.Gateway
				if err := ctx.db.Where("id = ? AND enabled = ?", *route.Gateway2ID, true).First(&gw2).Error; err == nil {
					ctx.conn.Execute("bridge", fmt.Sprintf("sofia/gateway/%s/%s", gw2.GatewayName, dialDest), true)
				}
			}
		}

		return
	}

	ctx.logger.Warn("No outbound route matched")
	ctx.conn.Execute("respond", "404 Not Found", false)
}

// ========== Helpers ==========

// handleRingGroupTimeout executes the timeout destination for a ring group
func (s *Service) handleRingGroupTimeout(ctx *callContext, rg *models.RingGroup) {
	if rg.TimeoutDestination == "" {
		return
	}

	ctx.logger.Infof("Ring group timeout -> %s (%s)", rg.TimeoutDestination, rg.TimeoutDestinationType)

	switch rg.TimeoutDestinationType {
	case "voicemail":
		ctx.conn.Execute("answer", "", true)
		ctx.conn.Execute("voicemail", fmt.Sprintf("default %s %s", ctx.domain, rg.TimeoutDestination), true)
	case "extension":
		ctx.conn.Execute("set", "call_timeout=30", true)
		ctx.conn.Execute("bridge", fmt.Sprintf("user/%s@%s", rg.TimeoutDestination, ctx.domain), true)
	case "ivr":
		ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", rg.TimeoutDestination, ctx.domain), true)
	default:
		ctx.conn.Execute("transfer", fmt.Sprintf("%s XML %s", rg.TimeoutDestination, ctx.domain), true)
	}
}

// getBridgeResult reads events until bridge result is known
func (s *Service) getBridgeResult(ctx *callContext) string {
	ev, err := ctx.conn.ReadEvent()
	if err != nil {
		return "NETWORK_ERROR"
	}

	cause := ev.Get("variable_originate_disposition")
	if cause == "" {
		cause = ev.Get("variable_hangup_cause")
	}
	if cause == "" {
		cause = ev.Get("variable_bridge_hangup_cause")
	}
	return cause
}

// fireRingGroupEvent fires a custom FreeSWITCH event for ring group state changes
func (s *Service) fireRingGroupEvent(ctx *callContext, rg *models.RingGroup, status string) {
	if ctx.manager.Client == nil {
		return
	}
	event := fmt.Sprintf(`sendevent CUSTOM
Event-Subclass: RING_GROUPS
Ring-Group-UUID: %s
Ring-Group-Name: %s
Ring-Group-Extension: %s
Ring-Group-Status: %s
Caller-Caller-ID-Number: %s
Caller-Caller-ID-Name: %s
Unique-ID: %s

`, rg.UUID, rg.Name, rg.Extension, status, ctx.callerID, ctx.callerName, ctx.uuid)
	ctx.manager.Client.Send(event)
}

func parseTenantID(s string) uint {
	var id uint = 1
	if s != "" {
		fmt.Sscanf(s, "%d", &id)
	}
	return id
}

// ========== Follow-Me ==========

// handleFollowMe rings the extension and its follow-me destination simultaneously.
// The follow-me leg gets a "Press 1 to accept" confirm prompt if PromptConfirm is set.
func (s *Service) handleFollowMe(ctx *callContext, ext *models.Extension) {
	timeout := ext.CallTimeout
	if timeout <= 0 {
		timeout = 30
	}

	ctx.conn.Execute("set", fmt.Sprintf("call_timeout=%d", timeout), true)
	ctx.conn.Execute("set", "hangup_after_bridge=true", true)
	ctx.conn.Execute("set", "continue_on_fail=true", true)
	ctx.conn.Execute("set", "ringback=${us-ring}", true)

	// Build dial string: local extension + follow-me with confirm
	localLeg := fmt.Sprintf("user/%s@%s", ext.Extension, ctx.domain)
	followMeLeg := fmt.Sprintf(
		"[group_confirm_key=1,group_confirm_file=ivr/ivr-accept_press_one.wav,confirm_timeout=5]loopback/%s/%s",
		ext.FollowMeDestination, ctx.domain)

	// Ring both simultaneously — first to answer wins
	dialString := localLeg + "," + followMeLeg
	ctx.conn.Execute("bridge", dialString, true)

	// Fallback to voicemail if both failed
	cause := s.getBridgeResult(ctx)
	if cause != "" && cause != "SUCCESS" {
		if ext.VoicemailEnabled {
			ctx.conn.Execute("answer", "", true)
			ctx.conn.Execute("voicemail", fmt.Sprintf("default %s %s", ctx.domain, ext.Extension), true)
		}
	}
}

// containsString checks if a string slice contains a value (trimming whitespace)
func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if strings.TrimSpace(s) == strings.TrimSpace(val) {
			return true
		}
	}
	return false
}
