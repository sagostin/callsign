// Package fax provides the fax module for CallSign.
// Architecture ported from gofaxserver: Manager (state) → Router (routing) → Queue (processing).
// Uses gofaxlib for low-level FreeSWITCH/SpanDSP operations.

package fax

import (
	"callsign/config"
	"callsign/models"
	"callsign/services/fax/gofaxlib"
	"callsign/services/logging"
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Note: T.38/fallback logic is handled by gofaxlib.RetryStrategy
// which provides 5 progressive fallback levels per src/dst pair.

// FaxJobInternal is the in-memory representation of a fax job being processed
// (equivalent to gofaxserver's FaxJob struct, but backed by CallSign's models.FaxJob)
type FaxJobInternal struct {
	UUID           uuid.UUID
	CallUUID       uuid.UUID
	CallerIdNumber string
	CalleeNumber   string
	CallerIdName   string
	Header         string
	FileName       string
	Status         string
	Ts             time.Time

	SrcTenantID uint
	DstTenantID uint

	// Routing info from gofaxserver
	SourceType string // gateway, api, webhook
	SourceID   string // gateway name, API user, etc.
	Identifier string

	// Resolved endpoints
	Endpoints []*models.FaxEndpoint

	// Result from gofaxlib
	Result *gofaxlib.FaxResult

	// Retry tracking
	Attempts  int
	LastError string

	// DB job reference
	DBJobID uint
}

// Manager is the main fax service — equivalent to gofaxserver's Server struct
// but integrated into CallSign's architecture.
type Manager struct {
	DB         *gorm.DB
	Config     *config.Config
	LogManager *logging.LogManager

	// In-memory state (from gofaxserver Server struct)
	mu              sync.RWMutex
	FaxBoxes        map[uint]*models.FaxBox        // keyed by FaxBox.ID
	FaxBoxesByDID   map[string]*models.FaxBox      // keyed by DID number string
	TenantEndpoints map[uint][]*models.FaxEndpoint // keyed by tenant ID
	BoxEndpoints    map[uint][]*models.FaxEndpoint // keyed by fax box ID
	GlobalEndpoints []*models.FaxEndpoint

	// Progressive retry strategy (replaces simple T.38 flip-flop)
	RetryStrategy *gofaxlib.RetryStrategy

	// Routing channel (from gofaxserver)
	JobRouting chan *FaxJobInternal

	// Queue result channel
	QueueResults chan *FaxJobInternal

	// Fax configuration
	EnableT38     bool
	RetryAttempts int
	RetryDelay    time.Duration
	TempDir       string

	// ESL connection info (for FreeSWITCH)
	ESLAddr     string
	ESLPassword string

	// Upstream gateways
	UpstreamGateways []string
}

// NewManager creates a new fax Manager
func NewManager(db *gorm.DB, cfg *config.Config, logMgr *logging.LogManager) *Manager {
	return &Manager{
		DB:              db,
		Config:          cfg,
		LogManager:      logMgr,
		FaxBoxes:        make(map[uint]*models.FaxBox),
		FaxBoxesByDID:   make(map[string]*models.FaxBox),
		TenantEndpoints: make(map[uint][]*models.FaxEndpoint),
		BoxEndpoints:    make(map[uint][]*models.FaxEndpoint),
		RetryStrategy:   gofaxlib.NewRetryStrategy(),
		JobRouting:      make(chan *FaxJobInternal),
		QueueResults:    make(chan *FaxJobInternal),
		EnableT38:       true,
		RetryAttempts:   3,
		RetryDelay:      time.Minute,
		TempDir:         "/tmp/callsign-fax",
		ESLAddr:         cfg.FreeSwitchHost + ":" + cfg.FreeSwitchPort,
		ESLPassword:     cfg.FreeSwitchPassword,
	}
}

// Start initializes the fax module — loads data from DB, starts router and queue goroutines
func (m *Manager) Start() error {
	m.LogManager.Info("FAX", "Starting fax module...", nil)

	// Load all fax data from DB into memory (from gofaxserver loadTenants/loadEndpoints pattern)
	if err := m.loadFaxBoxes(); err != nil {
		return fmt.Errorf("failed to load fax boxes: %w", err)
	}
	if err := m.loadEndpoints(); err != nil {
		return fmt.Errorf("failed to load fax endpoints: %w", err)
	}

	m.LogManager.Info("FAX", fmt.Sprintf("Loaded %d fax boxes and %d endpoints",
		len(m.FaxBoxes), m.countEndpoints()), nil)

	// Start router goroutine (from gofaxserver Router.Start)
	go m.routerLoop()

	// Start queue result processor
	go m.processResults()

	m.LogManager.Info("FAX", "Fax module started successfully", nil)
	return nil
}

// loadFaxBoxes loads all fax boxes from DB into memory
func (m *Manager) loadFaxBoxes() error {
	var boxes []models.FaxBox
	if err := m.DB.Where("enabled = ?", true).Find(&boxes).Error; err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.FaxBoxes = make(map[uint]*models.FaxBox)
	m.FaxBoxesByDID = make(map[string]*models.FaxBox)
	for i := range boxes {
		box := &boxes[i]
		m.FaxBoxes[box.ID] = box
		m.FaxBoxesByDID[box.DID] = box
	}
	return nil
}

// loadEndpoints loads all fax endpoints from DB into memory maps
// (from gofaxserver loadEndpoints pattern)
func (m *Manager) loadEndpoints() error {
	var endpoints []models.FaxEndpoint
	if err := m.DB.Where("enabled = ?", true).Find(&endpoints).Error; err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.TenantEndpoints = make(map[uint][]*models.FaxEndpoint)
	m.BoxEndpoints = make(map[uint][]*models.FaxEndpoint)
	m.GlobalEndpoints = nil
	m.UpstreamGateways = nil

	for i := range endpoints {
		ep := &endpoints[i]
		switch ep.Type {
		case "global":
			m.GlobalEndpoints = append(m.GlobalEndpoints, ep)
			if ep.EndpointType == "gateway" {
				gwName := strings.Split(ep.Endpoint, ":")[0]
				m.UpstreamGateways = append(m.UpstreamGateways, gwName)
			}
		case "tenant":
			if ep.TenantID != nil {
				m.TenantEndpoints[*ep.TenantID] = append(m.TenantEndpoints[*ep.TenantID], ep)
			}
		case "box":
			if ep.FaxBoxID != nil {
				m.BoxEndpoints[*ep.FaxBoxID] = append(m.BoxEndpoints[*ep.FaxBoxID], ep)
			}
		}
	}
	return nil
}

// ReloadData reloads fax boxes and endpoints from DB (from gofaxserver ReloadData)
func (m *Manager) ReloadData() error {
	if err := m.loadFaxBoxes(); err != nil {
		return fmt.Errorf("failed to reload fax boxes: %w", err)
	}
	if err := m.loadEndpoints(); err != nil {
		return fmt.Errorf("failed to reload endpoints: %w", err)
	}
	return nil
}

func (m *Manager) countEndpoints() int {
	count := len(m.GlobalEndpoints)
	for _, eps := range m.TenantEndpoints {
		count += len(eps)
	}
	for _, eps := range m.BoxEndpoints {
		count += len(eps)
	}
	return count
}

// --- Retry Strategy helpers ---

// GetFaxParams returns the FreeSWITCH channel parameters for the next fax attempt
// based on the per-pair retry history (T.38 → G.711, V.17 → V.29, ECM fallback)
func (m *Manager) GetFaxParams(srcNum, dstNum string, bridgeEnabled bool) gofaxlib.FaxChannelParams {
	level := m.RetryStrategy.GetNextLevel(srcNum, dstNum)
	return gofaxlib.ParamsForLevel(level, bridgeEnabled)
}

// RecordFaxAttempt records the result of a fax attempt for retry tracking
func (m *Manager) RecordFaxAttempt(srcNum, dstNum string, params gofaxlib.FaxChannelParams, success bool) {
	m.RetryStrategy.RecordAttempt(srcNum, dstNum, params.RetryLevel, success)
}

// --- Endpoint Resolution (from gofaxserver getEndpointsForNumber) ---

// GetEndpointsForDID returns endpoints for a given DID number.
// Box-specific endpoints take priority, then tenant endpoints, then global.
// Priority 666 endpoints are excluded (same as gofaxserver).
func (m *Manager) GetEndpointsForDID(did string) ([]*models.FaxEndpoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filterValid := func(endpoints []*models.FaxEndpoint) []*models.FaxEndpoint {
		var valid []*models.FaxEndpoint
		for _, ep := range endpoints {
			if ep.Priority != 666 {
				valid = append(valid, ep)
			}
		}
		return valid
	}

	// 1. Check box-specific endpoints
	box, exists := m.FaxBoxesByDID[did]
	if exists {
		if eps, ok := m.BoxEndpoints[box.ID]; ok && len(eps) > 0 {
			filtered := filterValid(eps)
			if len(filtered) > 0 {
				return filtered, nil
			}
		}

		// 2. Check tenant-level endpoints
		if eps, ok := m.TenantEndpoints[box.TenantID]; ok && len(eps) > 0 {
			filtered := filterValid(eps)
			if len(filtered) > 0 {
				return filtered, nil
			}
		}
	}

	// 3. Global endpoints (upstream gateways)
	if len(m.GlobalEndpoints) > 0 {
		filtered := filterValid(m.GlobalEndpoints)
		if len(filtered) > 0 {
			return filtered, nil
		}
	}

	return nil, fmt.Errorf("no valid endpoints found for DID %s", did)
}

// GetFaxBoxByDID returns the fax box for a given DID
func (m *Manager) GetFaxBoxByDID(did string) (*models.FaxBox, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	box, exists := m.FaxBoxesByDID[did]
	if !exists {
		return nil, fmt.Errorf("fax box for DID %s not found", did)
	}
	return box, nil
}

// --- Router Loop (from gofaxserver Router.Start) ---

func (m *Manager) routerLoop() {
	for {
		job := <-m.JobRouting
		go m.routeFax(job)
	}
}

// routeFax handles routing for a single fax job (from gofaxserver routeFax)
func (m *Manager) routeFax(job *FaxJobInternal) {
	job.Ts = time.Now()

	// Resolve destination fax box
	dstBox, _ := m.GetFaxBoxByDID(job.CalleeNumber)
	if dstBox != nil {
		job.DstTenantID = dstBox.TenantID
	}

	// Resolve source fax box (if caller is also known)
	srcBox, _ := m.GetFaxBoxByDID(job.CallerIdNumber)
	if srcBox != nil {
		job.SrcTenantID = srcBox.TenantID
		job.CallerIdName = srcBox.CallerIDName
		job.Header = srcBox.Header
	}

	m.LogManager.Info("FAX.ROUTER", fmt.Sprintf("Routing fax: %s → %s", job.CallerIdNumber, job.CalleeNumber), map[string]interface{}{
		"uuid":        job.UUID.String(),
		"src_tenant":  job.SrcTenantID,
		"dst_tenant":  job.DstTenantID,
		"source_type": job.SourceType,
	})

	// Find endpoints for destination
	endpoints, err := m.GetEndpointsForDID(job.CalleeNumber)
	if err != nil || len(endpoints) == 0 {
		m.LogManager.Error("FAX.ROUTER", fmt.Sprintf("No endpoints for DID %s: %v", job.CalleeNumber, err), nil)
		job.Status = "failed"
		m.persistJobStatus(job)
		return
	}

	job.Endpoints = endpoints
	job.Status = "routing"
	m.persistJobStatus(job)

	// Process via queue (inline, since we're already in a goroutine)
	m.processFaxJob(job)
}

// --- Queue Processing (from gofaxserver queue.go processFax) ---

func (m *Manager) processFaxJob(job *FaxJobInternal) {
	// Group endpoints by type → priority (from gofaxserver)
	groupMap := make(map[string]map[uint][]*models.FaxEndpoint, 4)
	for _, ep := range job.Endpoints {
		if groupMap[ep.EndpointType] == nil {
			groupMap[ep.EndpointType] = make(map[uint][]*models.FaxEndpoint)
		}
		groupMap[ep.EndpointType][ep.Priority] = append(groupMap[ep.EndpointType][ep.Priority], ep)
	}

	maxAttempts := m.RetryAttempts
	baseDelay := m.RetryDelay
	maxDelay := 60 * time.Second

	// Process each endpoint type concurrently (from gofaxserver)
	var wg sync.WaitGroup
	for endpointType, prioMap := range groupMap {
		epType := endpointType
		typeMap := prioMap

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Sort priorities (999 always last, from gofaxserver)
			prios := make([]uint, 0, len(typeMap))
			for p := range typeMap {
				prios = append(prios, p)
			}
			sort.Slice(prios, func(i, j int) bool {
				a, b := prios[i], prios[j]
				switch {
				case a == 999 && b != 999:
					return false
				case b == 999 && a != 999:
					return true
				default:
					return a < b
				}
			})

			for _, prio := range prios {
				group := typeMap[prio]
				if len(group) == 0 {
					continue
				}

				switch epType {
				case "gateway":
					if m.processGatewayGroup(job, group, prio, maxAttempts, baseDelay, maxDelay) {
						return // success
					}
				case "webhook":
					if m.processWebhookGroup(job, group, prio, maxAttempts, baseDelay, maxDelay) {
						return // success
					}
				case "email":
					if m.processEmailGroup(job, group, prio) {
						return // success
					}
				default:
					log.Warnf("Unknown fax endpoint type: %s", epType)
				}
			}
		}()
	}
	wg.Wait()

	// Final status update
	if job.Result != nil && job.Result.Success {
		job.Status = "complete"
	} else if job.Status != "complete" {
		job.Status = "failed"
	}

	now := time.Now()
	m.DB.Model(&models.FaxJob{}).Where("id = ?", job.DBJobID).Updates(map[string]interface{}{
		"status":            job.Status,
		"success":           job.Result != nil && job.Result.Success,
		"completed_at":      &now,
		"hangup_cause":      safeField(job, "hangup_cause"),
		"remote_id":         safeField(job, "remote_id"),
		"local_id":          safeField(job, "local_id"),
		"transfer_rate":     safeUint(job, "transfer_rate"),
		"ecm":               job.Result != nil && job.Result.Ecm,
		"t38_status":        safeField(job, "t38_status"),
		"result_code":       safeInt(job, "result_code"),
		"result_text":       safeField(job, "result_text"),
		"pages":             safeUint(job, "total_pages"),
		"transferred_pages": safeUint(job, "transferred_pages"),
	})

	// Persist page results to DB
	if job.Result != nil {
		for _, pr := range job.Result.PageResults {
			m.DB.Create(&models.FaxPageResult{
				FaxJobID:         job.DBJobID,
				Page:             pr.Page,
				BadRows:          pr.BadRows,
				LongestBadRowRun: pr.LongestBadRowRun,
				EncodingName:     pr.EncodingName,
				ImageSize:        pr.ImageSize,
				ImageWidth:       pr.ImagePixelSize.X,
				ImageHeight:      pr.ImagePixelSize.Y,
				ImageResX:        pr.ImageResolution.X,
				ImageResY:        pr.ImageResolution.Y,
			})
		}
	}

	// Trigger notifications for completed fax
	go m.sendNotifications(job)

	m.QueueResults <- job
}

// processGatewayGroup handles gateway endpoint group (from gofaxserver queue.go)
// Uses progressive retry strategy: T.38→G.711, V.17 baudrate fallback, ECM disable,
// bridge/transcoding for devices that can't handle T.38 natively.
func (m *Manager) processGatewayGroup(job *FaxJobInternal, group []*models.FaxEndpoint, prio uint, maxAttempts int, baseDelay, maxDelay time.Duration) bool {
	m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Processing gateway group (priority %d, %d endpoints)", prio, len(group)), map[string]interface{}{
		"uuid": job.UUID.String(),
	})

	for _, ep := range group {
		success := m.retryWithBackoff(context.Background(), maxAttempts, baseDelay, maxDelay, func(attempt int) (bool, bool) {
			job.Attempts++
			job.CallUUID = uuid.New()
			job.Status = "sending"
			m.persistJobStatus(job)

			// Get retry-level-aware fax parameters (T.38, V.17, ECM, bridge)
			faxParams := m.GetFaxParams(job.CallerIdNumber, job.CalleeNumber, ep.Bridge)

			m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Sending fax via gateway %s (attempt %d/%d) — %s",
				ep.Endpoint, attempt, maxAttempts, faxParams.String()), map[string]interface{}{
				"uuid":        job.UUID.String(),
				"call_uuid":   job.CallUUID.String(),
				"priority":    prio,
				"retry_level": faxParams.RetryLevel,
			})

			// Connect to FreeSWITCH and send fax via ESL
			conn, err := eventsocket.Dial(m.ESLAddr, m.ESLPassword)
			if err != nil {
				m.LogManager.Error("FAX.QUEUE", fmt.Sprintf("ESL connection failed: %v", err), nil)
				job.LastError = err.Error()
				m.RecordFaxAttempt(job.CallerIdNumber, job.CalleeNumber, faxParams, false)
				return false, true // retriable
			}
			defer conn.Close()

			// Build originate command using retry-strategy channel vars
			gwName := strings.Split(ep.Endpoint, ":")[0]
			channelVars := faxParams.ToFreeSwitchVars()
			originateVars := fmt.Sprintf(
				"{origination_caller_id_number=%s,origination_caller_id_name=%s,%s}",
				job.CallerIdNumber, job.CallerIdName, channelVars,
			)
			dialString := fmt.Sprintf("sofia/gateway/%s/%s", gwName, job.CalleeNumber)

			var faxApp string
			if faxParams.BridgeMode {
				// Bridge mode: transcode between T.38 upstream and G.711 device
				faxApp = fmt.Sprintf("bridge %s", dialString)
			} else {
				faxApp = fmt.Sprintf("txfax %s", job.FileName)
			}

			cmd := fmt.Sprintf("bgapi originate %s%s '%s'", originateVars, dialString, faxApp)
			result, err := conn.Send(cmd)
			if err != nil {
				m.LogManager.Error("FAX.QUEUE", fmt.Sprintf("Originate failed: %v", err), nil)
				job.LastError = err.Error()
				m.RecordFaxAttempt(job.CallerIdNumber, job.CalleeNumber, faxParams, false)
				return false, true
			}

			m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Originate result: %s", result.Body), map[string]interface{}{
				"uuid": job.UUID.String(),
			})

			// Record attempt — next retry will use appropriate fallback level
			// Full success/failure tracking happens via ESL events
			m.RecordFaxAttempt(job.CallerIdNumber, job.CalleeNumber, faxParams, true)
			return true, false
		})

		if success {
			return true
		}
	}
	return false
}

// processWebhookGroup handles webhook endpoint group (from gofaxserver queue.go)
func (m *Manager) processWebhookGroup(job *FaxJobInternal, group []*models.FaxEndpoint, prio uint, maxAttempts int, baseDelay, maxDelay time.Duration) bool {
	m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Processing webhook group (priority %d, %d endpoints)", prio, len(group)), map[string]interface{}{
		"uuid": job.UUID.String(),
	})

	// TODO: Implement webhook delivery (TIFF→PDF, base64, POST to endpoint URL)
	// This follows the same pattern as gofaxserver queue.go webhook processing
	log.Warn("Webhook fax delivery not yet implemented")
	return false
}

// processEmailGroup handles email endpoint group
func (m *Manager) processEmailGroup(job *FaxJobInternal, group []*models.FaxEndpoint, prio uint) bool {
	m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Processing email group (priority %d, %d endpoints)", prio, len(group)), map[string]interface{}{
		"uuid": job.UUID.String(),
	})

	// TODO: Implement email delivery (TIFF→PDF attachment, SMTP send)
	log.Warn("Email fax delivery not yet implemented")
	return false
}

// retryWithBackoff runs fn up to maxAttempts with exponential backoff + jitter
// (from gofaxserver queue.go)
func (m *Manager) retryWithBackoff(ctx context.Context, maxAttempts int, base, max time.Duration, fn func(attempt int) (ok bool, retriable bool)) bool {
	delay := base
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return false
		}
		ok, retriable := fn(attempt)
		if ok {
			return true
		}
		if !retriable || attempt == maxAttempts {
			return false
		}
		// ±10% jitter (from gofaxserver)
		jit := time.Duration(rand.Int63n(int64(delay)/5)) - delay/10
		sleep := delay + jit
		if sleep > max {
			sleep = max
		}
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return false
		}
		delay *= 2
		if delay > max {
			delay = max
		}
	}
	return false
}

// --- Persistence helpers ---

func (m *Manager) persistJobStatus(job *FaxJobInternal) {
	updates := map[string]interface{}{
		"status":   job.Status,
		"attempts": job.Attempts,
	}
	if job.LastError != "" {
		updates["last_error"] = job.LastError
	}
	m.DB.Model(&models.FaxJob{}).Where("id = ?", job.DBJobID).Updates(updates)
}

func (m *Manager) processResults() {
	for job := range m.QueueResults {
		m.LogManager.Info("FAX.RESULT", fmt.Sprintf("Fax %s completed: success=%v", job.UUID.String(), job.Result != nil && job.Result.Success), nil)
	}
}

// sendNotifications sends email/webhook notifications for a completed fax
func (m *Manager) sendNotifications(job *FaxJobInternal) {
	if job.DstTenantID == 0 {
		return
	}

	// Find the destination fax box
	box, _ := m.GetFaxBoxByDID(job.CalleeNumber)
	if box == nil {
		return
	}

	// Email notifications
	if box.EmailOnReceive && len(box.NotifyEmails) > 0 {
		for _, email := range box.NotifyEmails {
			m.DB.Create(&models.FaxNotificationLog{
				FaxJobID:  job.DBJobID,
				Type:      "email",
				Recipient: email,
				Status:    "pending",
			})
			// TODO: Actually send email via tenant SMTP config
		}
	}

	// Webhook notifications
	if box.NotifyWebhook != "" {
		m.DB.Create(&models.FaxNotificationLog{
			FaxJobID:  job.DBJobID,
			Type:      "webhook",
			Recipient: box.NotifyWebhook,
			Status:    "pending",
		})
		// TODO: POST to webhook URL
	}
}

// --- Safe field helpers for DB updates ---

func safeField(job *FaxJobInternal, field string) string {
	if job.Result == nil {
		return ""
	}
	switch field {
	case "hangup_cause":
		return job.Result.HangupCause
	case "remote_id":
		return job.Result.RemoteID
	case "local_id":
		return job.Result.LocalID
	case "t38_status":
		return job.Result.T38Status
	case "result_text":
		return job.Result.ResultText
	}
	return ""
}

func safeUint(job *FaxJobInternal, field string) uint {
	if job.Result == nil {
		return 0
	}
	switch field {
	case "transfer_rate":
		return job.Result.TransferRate
	case "total_pages":
		return job.Result.TotalPages
	case "transferred_pages":
		return job.Result.TransferredPages
	}
	return 0
}

func safeInt(job *FaxJobInternal, field string) int {
	if job.Result == nil {
		return 0
	}
	switch field {
	case "result_code":
		return job.Result.ResultCode
	}
	return 0
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
