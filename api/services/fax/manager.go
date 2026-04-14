// Package fax provides the fax module for CallSign.
// Architecture ported from gofaxserver: Manager (state) → Router (routing) → Queue (processing).
// Uses gofaxlib for low-level FreeSWITCH/SpanDSP operations.

package fax

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"callsign/config"
	"callsign/models"
	"callsign/services/fax/gofaxlib"
	"callsign/services/logging"
	"context"

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
// Each endpoint in the group is a webhook URL to POST the fax document to.
// The fax TIFF is base64-encoded and sent as JSON along with fax metadata.
func (m *Manager) processWebhookGroup(job *FaxJobInternal, group []*models.FaxEndpoint, prio uint, maxAttempts int, baseDelay, maxDelay time.Duration) bool {
	if len(group) == 0 {
		return false
	}

	m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Processing webhook group (priority %d, %d endpoints)", prio, len(group)), map[string]interface{}{
		"uuid": job.UUID.String(),
	})

	// Determine the fax file to send (prefer PDF over TIFF)
	faxFilePath := job.FileName
	if job.Result != nil && job.Result.TotalPages > 0 {
		var dbJob models.FaxJob
		if m.DB.First(&dbJob, job.DBJobID).Error == nil && dbJob.PDFFileName != "" {
			faxFilePath = dbJob.PDFFileName
		}
	}

	for _, ep := range group {
		webhookURL := ep.Endpoint
		if webhookURL == "" {
			m.LogManager.Error("FAX.QUEUE", "Webhook endpoint has empty URL", map[string]interface{}{
				"uuid":  job.UUID.String(),
				"ep_id": ep.ID,
			})
			continue
		}

		delivered := m.retryWithBackoff(context.Background(), maxAttempts, baseDelay, maxDelay, func(attempt int) (bool, bool) {
			m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Delivering fax via webhook (attempt %d/%d): %s", attempt, maxAttempts, webhookURL), map[string]interface{}{
				"uuid":        job.UUID.String(),
				"webhook_url": webhookURL,
			})

			if sendErr := m.sendFaxToWebhook(webhookURL, faxFilePath, job); sendErr != nil {
				m.LogManager.Error("FAX.QUEUE", fmt.Sprintf("Webhook delivery failed: %v", sendErr), map[string]interface{}{
					"uuid":        job.UUID.String(),
					"webhook_url": webhookURL,
					"attempt":     attempt,
				})
				job.LastError = sendErr.Error()
				return false, true // retriable
			}

			m.LogManager.Info("FAX.QUEUE", "Webhook delivery successful", map[string]interface{}{
				"uuid":        job.UUID.String(),
				"webhook_url": webhookURL,
			})
			return true, false
		})

		if delivered {
			return true
		}
	}

	return false
}

// sendFaxToWebhook POSTs the fax document to the configured webhook URL.
// The request body is JSON containing base64-encoded fax data and metadata.
func (m *Manager) sendFaxToWebhook(webhookURL, faxFilePath string, job *FaxJobInternal) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}

	// Read fax file content
	var faxData []byte
	var readErr error
	if faxFilePath != "" {
		faxData, readErr = os.ReadFile(faxFilePath)
		if readErr != nil {
			return fmt.Errorf("failed to read fax file %s: %w", faxFilePath, readErr)
		}
	} else {
		return fmt.Errorf("no fax file available for delivery")
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(faxData)

	// Determine content type from file extension
	contentType := "image/tiff"
	if strings.HasSuffix(strings.ToLower(faxFilePath), ".pdf") {
		contentType = "application/pdf"
	}

	// Build request payload
	payload := map[string]interface{}{
		"fax_uuid":       job.UUID.String(),
		"caller_number":  job.CallerIdNumber,
		"callee_number":  job.CalleeNumber,
		"caller_id_name": job.CallerIdName,
		"header":         job.Header,
		"file_name":      filepath.Base(faxFilePath),
		"content_type":   contentType,
		"data":           encoded, // base64-encoded fax document
	}

	if job.Result != nil {
		payload["success"] = job.Result.Success
		payload["pages"] = job.Result.TotalPages
		payload["transfer_rate"] = job.Result.TransferRate
		payload["ecm"] = job.Result.Ecm
		payload["result_code"] = job.Result.ResultCode
		payload["result_text"] = job.Result.ResultText
	}

	payloadBytes, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", jsonErr)
	}

	// Create HTTP request
	req, reqErr := http.NewRequest("POST", webhookURL, bytes.NewReader(payloadBytes))
	if reqErr != nil {
		return fmt.Errorf("failed to create webhook request: %w", reqErr)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CallSign-Fax/1.0")

	// Send request with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, respErr := client.Do(req)
	if respErr != nil {
		return fmt.Errorf("webhook request failed: %w", respErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

// processEmailGroup handles email endpoint group — sends fax as email attachment
func (m *Manager) processEmailGroup(job *FaxJobInternal, group []*models.FaxEndpoint, prio uint) bool {
	if len(group) == 0 {
		return false
	}

	m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Processing email group (priority %d, %d endpoints)", prio, len(group)), map[string]interface{}{
		"uuid": job.UUID.String(),
	})

	// Gather email addresses from all endpoints in the group
	var emailRecipients []string
	for _, ep := range group {
		if ep.Endpoint == "" {
			continue
		}
		// Endpoint field may contain multiple emails separated by commas
		emails := strings.Split(ep.Endpoint, ",")
		for _, e := range emails {
			e = strings.TrimSpace(e)
			if e != "" {
				emailRecipients = append(emailRecipients, e)
			}
		}
	}

	if len(emailRecipients) == 0 {
		m.LogManager.Error("FAX.QUEUE", "No email recipients found in endpoint group", map[string]interface{}{
			"uuid": job.UUID.String(),
		})
		return false
	}

	// Get tenant SMTP settings for sending
	tenantSMTP := m.getTenantSMTPSettings(job.DstTenantID)
	if tenantSMTP == nil {
		m.LogManager.Error("FAX.QUEUE", "Tenant SMTP not configured", map[string]interface{}{
			"uuid":      job.UUID.String(),
			"tenant_id": job.DstTenantID,
		})
		return false
	}

	// Determine the fax file to send (prefer PDF over TIFF)
	faxFilePath := job.FileName
	var dbJob models.FaxJob
	if m.DB.First(&dbJob, job.DBJobID).Error == nil && dbJob.PDFFileName != "" {
		faxFilePath = dbJob.PDFFileName
	}

	// Build email subject and body
	subject := fmt.Sprintf("Fax from %s to %s", job.CallerIdNumber, job.CalleeNumber)
	body := fmt.Sprintf(
		"You have received a fax.\n\n"+
			"From: %s\n"+
			"To: %s\n"+
			"Date: %s\n"+
			"The fax document is attached to this email.\n",
		job.CallerIdNumber, job.CalleeNumber, time.Now().Format("2006-01-02 15:04:05"),
	)

	// Send to each recipient
	var lastErr error
	for _, recipient := range emailRecipients {
		if sendErr := m.sendFaxEmail(tenantSMTP, recipient, subject, body, faxFilePath, job); sendErr != nil {
			m.LogManager.Error("FAX.QUEUE", fmt.Sprintf("Failed to send fax email to %s: %v", recipient, sendErr), map[string]interface{}{
				"uuid":      job.UUID.String(),
				"recipient": recipient,
			})
			lastErr = sendErr
			continue
		}
		m.LogManager.Info("FAX.QUEUE", fmt.Sprintf("Fax email sent to %s", recipient), map[string]interface{}{
			"uuid":      job.UUID.String(),
			"recipient": recipient,
		})
	}

	return lastErr == nil
}

// tenantSMTPConfig holds SMTP settings for a tenant
type tenantSMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	FromAddr string
	FromName string
	UseTLS   bool
}

// getTenantSMTPSettings retrieves SMTP configuration for a tenant
func (m *Manager) getTenantSMTPSettings(tenantID uint) *tenantSMTPConfig {
	if tenantID == 0 {
		return nil
	}

	var tenant models.Tenant
	if err := m.DB.First(&tenant, tenantID).Error; err != nil {
		return nil
	}

	// Parse settings JSON
	var settings struct {
		SMTPOverride   bool   `json:"smtp_override"`
		SMTPHost       string `json:"smtp_host"`
		SMTPPort       string `json:"smtp_port"`
		SMTPUsername   string `json:"smtp_username"`
		SMTPPassword   string `json:"smtp_password"`
		SMTPFromEmail  string `json:"smtp_from_email"`
		SMTPEncryption string `json:"smtp_encryption"`
	}

	if tenant.Settings != "" && tenant.Settings != "{}" {
		if err := json.Unmarshal([]byte(tenant.Settings), &settings); err != nil {
			return nil
		}
	}

	// If SMTP not overridden and not configured, skip
	if !settings.SMTPOverride && settings.SMTPHost == "" {
		return nil
	}

	port := settings.SMTPPort
	if port == "" {
		port = "587"
	}

	useTLS := false
	if strings.ToLower(settings.SMTPEncryption) == "tls" || strings.ToLower(settings.SMTPEncryption) == "ssl" {
		useTLS = true
	}

	return &tenantSMTPConfig{
		Host:     settings.SMTPHost,
		Port:     port,
		Username: settings.SMTPUsername,
		Password: settings.SMTPPassword,
		FromAddr: settings.SMTPFromEmail,
		FromName: "CallSign Fax",
		UseTLS:   useTLS,
	}
}

// sendFaxEmail sends a fax document as an email attachment
func (m *Manager) sendFaxEmail(cfg *tenantSMTPConfig, to, subject, body, attachmentPath string, job *FaxJobInternal) error {
	if cfg == nil || cfg.Host == "" {
		return fmt.Errorf("SMTP configuration is empty")
	}

	var attachData []byte
	var attachErr error
	if attachmentPath != "" {
		attachData, attachErr = os.ReadFile(attachmentPath)
		if attachErr != nil {
			m.LogManager.Warn("FAX.QUEUE", fmt.Sprintf("Could not read fax attachment: %s", attachmentPath), nil)
			// Continue without attachment — body text is still useful
		}
	}

	// Build multipart email
	var msg bytes.Buffer
	writer := multipart.NewWriter(&msg)

	// Email headers
	headers := make(textproto.MIMEHeader)
	headers.Set("From", fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddr))
	headers.Set("To", to)
	headers.Set("Subject", subject)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary()))

	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
	}
	msg.WriteString("\r\n")

	// Text body part
	textPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"text/plain; charset=utf-8"},
	})
	if err != nil {
		return fmt.Errorf("create text part: %w", err)
	}
	textPart.Write([]byte(body))

	// Attachment part (if we have fax data)
	if len(attachData) > 0 {
		filename := filepath.Base(attachmentPath)
		if filename == "" {
			filename = fmt.Sprintf("fax-%s.tiff", job.UUID.String()[:8])
		}

		var contentType string
		switch {
		case strings.HasSuffix(strings.ToLower(filename), ".pdf"):
			contentType = "application/pdf"
		case strings.HasSuffix(strings.ToLower(filename), ".tif"):
			contentType = "image/tiff"
		case strings.HasSuffix(strings.ToLower(filename), ".tiff"):
			contentType = "image/tiff"
		default:
			contentType = "application/octet-stream"
		}

		attachPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {contentType},
			"Content-Transfer-Encoding": {"base64"},
			"Content-Disposition":       {fmt.Sprintf("attachment; filename=%q", filename)},
		})
		if err != nil {
			return fmt.Errorf("create attachment part: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(attachData)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			attachPart.Write([]byte(encoded[i:end] + "\r\n"))
		}
	}

	writer.Close()

	// Send via SMTP
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	var sendErr error
	if cfg.UseTLS {
		// Connect with TLS
		tlsConfig := &tls.Config{
			ServerName: cfg.Host,
		}
		conn, connErr := tls.Dial("tcp", addr, tlsConfig)
		if connErr != nil {
			return fmt.Errorf("TLS connection failed: %w", connErr)
		}
		client, clientErr := smtp.NewClient(conn, cfg.Host)
		if clientErr != nil {
			return fmt.Errorf("SMTP client creation failed: %w", clientErr)
		}
		defer client.Close()

		if auth != nil {
			if authErr := client.Auth(auth); authErr != nil {
				return fmt.Errorf("SMTP auth failed: %w", authErr)
			}
		}
		if fromErr := client.Mail(cfg.FromAddr); fromErr != nil {
			return fmt.Errorf("SMTP MAIL from failed: %w", fromErr)
		}
		if rcptErr := client.Rcpt(to); rcptErr != nil {
			return fmt.Errorf("SMTP RCPT to failed: %w", rcptErr)
		}
		w, wErr := client.Data()
		if wErr != nil {
			return fmt.Errorf("SMTP DATA failed: %w", wErr)
		}
		_, wErr = w.Write(msg.Bytes())
		if wErr != nil {
			return fmt.Errorf("SMTP write failed: %w", wErr)
		}
		wErr = w.Close()
		if wErr != nil {
			return fmt.Errorf("SMTP data close failed: %w", wErr)
		}
		client.Quit()
	} else {
		sendErr = smtp.SendMail(addr, auth, cfg.FromAddr, []string{to}, msg.Bytes())
		if sendErr != nil {
			return fmt.Errorf("SMTP send failed: %w", sendErr)
		}
	}

	return nil
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
