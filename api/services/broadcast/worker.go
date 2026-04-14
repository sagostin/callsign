// Package broadcast provides the broadcast campaign worker service.
package broadcast

import (
	"callsign/models"
	"callsign/services/esl"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// recipientInfo holds information about a recipient for the broadcast.
type recipientInfo struct {
	ID          uint
	PhoneNumber string
	Status      string
}

// campaignContext provides cancellation support for campaign runs.
type campaignContext struct {
	done chan struct{}
}

func (c *campaignContext) Done() <-chan struct{} { return c.done }
func (c *campaignContext) Err() error {
	select {
	case <-c.done:
		return fmt.Errorf("cancelled")
	default:
		return nil
	}
}

// BroadcastWorker handles async origination of broadcast campaign calls.
type BroadcastWorker struct {
	db  *gorm.DB
	esl *esl.Manager

	// Track running campaigns for stopping
	runningCampaigns map[uint]cancelFunc
	workerMutex      sync.Mutex
}

// cancelFunc is a function that cancels a campaign run.
type cancelFunc func()

// NewBroadcastWorker creates a new broadcast worker.
func NewBroadcastWorker(db *gorm.DB, eslManager *esl.Manager) *BroadcastWorker {
	return &BroadcastWorker{
		db:               db,
		esl:              eslManager,
		runningCampaigns: make(map[uint]cancelFunc),
	}
}

// StartCampaign loads a campaign and starts originating calls to all recipients.
// It runs asynchronously - the HTTP handler returns immediately.
func (w *BroadcastWorker) StartCampaign(campaignID uint) error {
	// Load campaign from DB
	var campaign models.BroadcastCampaign
	if err := w.db.First(&campaign, campaignID).Error; err != nil {
		return fmt.Errorf("failed to load campaign: %w", err)
	}

	// Guard: early exit if already running
	if campaign.Status == models.BroadcastStatusRunning {
		return fmt.Errorf("campaign %d is already running", campaignID)
	}

	// Guard: early exit if no recipients
	if len(campaign.Recipients) == 0 {
		return fmt.Errorf("campaign has no recipients")
	}

	// Update status to running
	now := time.Now()
	if err := w.db.Model(&campaign).Updates(map[string]interface{}{
		"status":     models.BroadcastStatusRunning,
		"started_at": now,
	}).Error; err != nil {
		return fmt.Errorf("failed to update campaign status: %w", err)
	}

	// Prepare recipients as a slice of structs for the run loop
	recipients := make([]recipientInfo, len(campaign.Recipients))
	for i, phone := range campaign.Recipients {
		recipients[i] = recipientInfo{
			ID:          uint(i + 1),
			PhoneNumber: phone,
			Status:      "pending",
		}
	}

	// Create cancellable context for this campaign
	ctx, cancel := contextWithCancel()

	// Register cancel function
	w.workerMutex.Lock()
	w.runningCampaigns[campaignID] = cancel
	w.workerMutex.Unlock()

	// Run the campaign loop asynchronously
	go w.runLoop(campaignID, recipients, ctx)

	return nil
}

// StopCampaign signals the worker to stop originating new calls for a campaign.
func (w *BroadcastWorker) StopCampaign(campaignID uint) error {
	w.workerMutex.Lock()
	cancel, exists := w.runningCampaigns[campaignID]
	w.workerMutex.Unlock()

	if !exists {
		return fmt.Errorf("campaign %d is not running", campaignID)
	}

	// Signal cancellation
	cancel()

	// Remove from tracking
	w.workerMutex.Lock()
	delete(w.runningCampaigns, campaignID)
	w.workerMutex.Unlock()

	// Update status to paused
	w.db.Model(&models.BroadcastCampaign{}).Where("id = ?", campaignID).
		Update("status", models.BroadcastStatusPaused)

	log.Infof("Broadcast campaign %d stop signaled", campaignID)
	return nil
}

// runLoop iterates through recipients and originates calls via ESL.
func (w *BroadcastWorker) runLoop(campaignID uint, recipients []recipientInfo, ctx *campaignContext) {
	defer func() {
		// Cleanup: remove from running campaigns and update status if not already stopped
		w.workerMutex.Lock()
		delete(w.runningCampaigns, campaignID)
		w.workerMutex.Unlock()

		// Check if campaign was fully completed
		var campaign models.BroadcastCampaign
		if err := w.db.First(&campaign, campaignID).Error; err == nil {
			if campaign.Status == models.BroadcastStatusRunning {
				now := time.Now()
				w.db.Model(&campaign).Updates(map[string]interface{}{
					"status":       models.BroadcastStatusCompleted,
					"completed_at": now,
				})
				log.Infof("Broadcast campaign %d completed", campaignID)
			}
		}
	}()

	// Load campaign for recording path and config
	var campaign models.BroadcastCampaign
	if err := w.db.First(&campaign, campaignID).Error; err != nil {
		log.Errorf("Broadcast campaign %d: failed to load: %v", campaignID, err)
		return
	}

	// Validate recording path exists
	if campaign.RecordingPath == "" {
		log.Errorf("Broadcast campaign %d: no recording path configured", campaignID)
		return
	}

	log.Infof("Broadcast campaign %d: starting loop for %d recipients", campaignID, len(recipients))

	// Semaphore for concurrent call limiting
	sem := make(chan struct{}, campaign.ConcurrentLimit)
	if campaign.ConcurrentLimit <= 0 {
		sem = make(chan struct{}, 5) // Default limit of 5
	}

	for i, recipient := range recipients {
		// Check for cancellation
		select {
		case <-ctx.Done():
			log.Infof("Broadcast campaign %d: cancelled by user", campaignID)
			return
		default:
		}

		// Update progress
		w.db.Model(&models.BroadcastCampaign{}).Where("id = ?", campaignID).
			Update("total_calls", i+1)

		// Acquire semaphore slot
		sem <- struct{}{}

		// Originate call asynchronously
		go func(recipient recipientInfo) {
			defer func() { <-sem }() // Release slot
			w.originateCall(&campaign, recipient)
		}(recipient)
	}

	// Wait for all calls to complete
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	log.Infof("Broadcast campaign %d: all calls originated, waiting for completion", campaignID)
}

// originateCall originates a single call and handles the result.
func (w *BroadcastWorker) originateCall(campaign *models.BroadcastCampaign, recipient recipientInfo) {
	// Build the originate command with campaign and recipient metadata
	// Format: bgapi originate {campaign_id=X,recipient_id=Y,ignore_early_media=true}endpoint &playback(path)
	cmd := fmt.Sprintf(
		"bgapi originate {campaign_id=%d,recipient_id=%d,ignore_early_media=true,campaign_name=%s}%s %s &playback(%s)",
		campaign.ID,
		recipient.ID,
		campaign.Name,
		recipient.PhoneNumber,
		campaign.CallerID, // Using caller ID as the destination/extension to bridge to
		campaign.RecordingPath,
	)

	log.WithFields(log.Fields{
		"campaign_id":  campaign.ID,
		"recipient_id": recipient.ID,
		"phone_number": recipient.PhoneNumber,
	}).Debug("Originating broadcast call")

	result, err := w.esl.BgAPI(cmd)
	if err != nil {
		log.Errorf("Broadcast call failed to originate: %v", err)
		w.updateRecipientStats(campaign.ID, "failed")
		return
	}

	log.WithFields(log.Fields{
		"campaign_id": campaign.ID,
		"job_uuid":    result,
		"phone":       recipient.PhoneNumber,
	}).Debug("Broadcast call originated")

	// Note: In a full implementation, we would track call state events
	// (CHANNEL_ANSWER, CHANNEL_HANGUP_COMPLETE) to update actual answer stats.
	// For now, we mark as answered since the originate was successful.
	// A more robust implementation would subscribe to ESL events and track UUIDs.
	w.updateRecipientStats(campaign.ID, "answered")
}

// updateRecipientStats updates the campaign's call statistics in the database.
func (w *BroadcastWorker) updateRecipientStats(campaignID uint, result string) {
	updates := map[string]interface{}{}
	switch result {
	case "answered":
		updates["answered_calls"] = gorm.Expr("answered_calls + 1")
	case "failed":
		updates["failed_calls"] = gorm.Expr("failed_calls + 1")
	case "busy":
		updates["busy_calls"] = gorm.Expr("busy_calls + 1")
	case "no_answer":
		updates["no_answer_calls"] = gorm.Expr("no_answer_calls + 1")
	}

	if len(updates) > 0 {
		w.db.Model(&models.BroadcastCampaign{}).Where("id = ?", campaignID).Updates(updates)
	}
}

// contextWithCancel creates a basic context with cancel support.
func contextWithCancel() (*campaignContext, cancelFunc) {
	ctx := &campaignContext{done: make(chan struct{})}
	cancel := cancelFunc(func() {
		close(ctx.done)
	})
	return ctx, cancel
}

// IsCampaignRunning returns whether a campaign is currently being processed.
func (w *BroadcastWorker) IsCampaignRunning(campaignID uint) bool {
	w.workerMutex.Lock()
	defer w.workerMutex.Unlock()
	_, exists := w.runningCampaigns[campaignID]
	return exists
}
