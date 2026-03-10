package messaging

import (
	"callsign/models"
	"context"
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// QueueWorker processes outbound messages from the message queue
type QueueWorker struct {
	db        *gorm.DB
	providers map[uint]SMSProvider // providerID -> provider instance
	cancel    context.CancelFunc
	stopped   chan struct{}
}

// NewQueueWorker creates a new queue worker
func NewQueueWorker(db *gorm.DB, providers map[uint]SMSProvider) *QueueWorker {
	return &QueueWorker{
		db:        db,
		providers: providers,
		stopped:   make(chan struct{}),
	}
}

// Start begins processing the message queue
func (w *QueueWorker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go func() {
		defer close(w.stopped)
		log.Info("Message queue worker started")

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("Message queue worker stopping")
				return
			case <-ticker.C:
				w.processBatch(ctx)
			}
		}
	}()
}

// Stop gracefully stops the queue worker
func (w *QueueWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
		<-w.stopped
	}
}

// processBatch fetches and processes pending queue items
func (w *QueueWorker) processBatch(ctx context.Context) {
	var items []models.MessageQueueItem

	// Fetch pending and retry-ready items
	now := time.Now()
	result := w.db.Where(
		"(status = 'pending' OR (status = 'retry' AND next_retry_at <= ?)) AND attempts < max_attempts",
		now,
	).Order("created_at ASC").Limit(10).Find(&items)

	if result.Error != nil {
		log.Errorf("Queue worker: failed to fetch items: %v", result.Error)
		return
	}

	for _, item := range items {
		select {
		case <-ctx.Done():
			return
		default:
			w.processItem(ctx, item)
		}
	}
}

// processItem sends a single message via the appropriate provider
func (w *QueueWorker) processItem(ctx context.Context, item models.MessageQueueItem) {
	logger := log.WithFields(log.Fields{
		"queue_item_id": item.ID,
		"from":          item.FromNumber,
		"to":            item.ToNumber,
		"provider_id":   item.ProviderID,
		"attempt":       item.Attempts + 1,
	})

	// Mark as processing
	w.db.Model(&item).Updates(map[string]interface{}{
		"status":   "processing",
		"attempts": item.Attempts + 1,
	})

	// Find provider
	provider, ok := w.providers[item.ProviderID]
	if !ok {
		logger.Error("Provider not found for queue item")
		w.failItem(&item, "provider not configured")
		return
	}

	// Build send request
	req := SendRequest{
		From: item.FromNumber,
		To:   item.ToNumber,
		Body: item.Body,
	}

	// If there are media attachments, load them
	if item.HasMedia {
		mediaURLs, err := w.loadMediaURLs(&item)
		if err != nil {
			logger.WithError(err).Error("Failed to load media for queue item")
			w.retryOrFail(&item, fmt.Sprintf("media load error: %v", err))
			return
		}
		req.MediaURLs = mediaURLs
	}

	// Send via provider
	var resp *SendResponse
	var err error
	if item.HasMedia && len(req.MediaURLs) > 0 {
		resp, err = provider.SendMMS(ctx, req)
	} else {
		resp, err = provider.SendSMS(ctx, req)
	}

	if err != nil {
		logger.WithError(err).Error("Failed to send message")
		w.retryOrFail(&item, err.Error())
		return
	}

	// Success
	now := time.Now()
	w.db.Model(&item).Updates(map[string]interface{}{
		"status":              "sent",
		"provider_message_id": resp.MessageID,
		"processed_at":        &now,
	})

	// Update the source message status
	w.updateSourceMessageStatus(&item, "sent", resp.MessageID)

	logger.WithField("provider_message_id", resp.MessageID).Info("Message sent successfully")
}

// retryOrFail either schedules a retry or marks as permanently failed
func (w *QueueWorker) retryOrFail(item *models.MessageQueueItem, errMsg string) {
	if item.Attempts >= item.MaxAttempts {
		w.failItem(item, errMsg)
		return
	}

	// Exponential backoff: 30s, 2m, 10m
	delays := []time.Duration{30 * time.Second, 2 * time.Minute, 10 * time.Minute}
	delay := delays[0]
	if item.Attempts-1 < len(delays) {
		delay = delays[item.Attempts-1]
	} else {
		delay = delays[len(delays)-1]
	}
	// Apply jitter
	jitter := time.Duration(float64(delay) * 0.1 * math.Abs(float64(time.Now().UnixNano()%100)/100.0))
	nextRetry := time.Now().Add(delay + jitter)

	w.db.Model(item).Updates(map[string]interface{}{
		"status":        "retry",
		"last_error":    errMsg,
		"next_retry_at": &nextRetry,
	})
}

// failItem permanently marks an item as failed
func (w *QueueWorker) failItem(item *models.MessageQueueItem, errMsg string) {
	now := time.Now()
	w.db.Model(item).Updates(map[string]interface{}{
		"status":       "failed",
		"last_error":   errMsg,
		"processed_at": &now,
	})

	w.updateSourceMessageStatus(item, "failed", "")
}

// updateSourceMessageStatus updates the original Message or ChatMessage status
func (w *QueueWorker) updateSourceMessageStatus(item *models.MessageQueueItem, status, providerMsgID string) {
	if item.MessageID != nil {
		updates := map[string]interface{}{"status": status}
		if providerMsgID != "" {
			updates["external_id"] = providerMsgID
		}
		if status == "sent" {
			now := time.Now()
			updates["sent_at"] = &now
		}
		w.db.Model(&models.Message{}).Where("id = ?", *item.MessageID).Updates(updates)
	}

	if item.ChatMessageID != nil {
		updates := map[string]interface{}{"status": status}
		if providerMsgID != "" {
			updates["external_id"] = providerMsgID
		}
		w.db.Model(&models.ChatMessage{}).Where("id = ?", *item.ChatMessageID).Updates(updates)
	}
}

// loadMediaURLs loads media attachment URLs for a queue item
func (w *QueueWorker) loadMediaURLs(item *models.MessageQueueItem) ([]string, error) {
	var urls []string

	if item.ChatMessageID != nil {
		var attachments []models.ChatAttachment
		if err := w.db.Where("message_id = ?", *item.ChatMessageID).Find(&attachments).Error; err != nil {
			return nil, err
		}
		for _, att := range attachments {
			if att.ExternalURL != "" {
				urls = append(urls, att.ExternalURL)
			}
			// If no external URL, we'd need to serve attachments via our API
			// and provide that URL to Telnyx. For now, skip inline base64.
		}
	}

	if item.MessageID != nil {
		var media []models.MessageMedia
		if err := w.db.Where("message_id = ?", *item.MessageID).Find(&media).Error; err != nil {
			return nil, err
		}
		for _, m := range media {
			if m.ExternalURL != "" {
				urls = append(urls, m.ExternalURL)
			}
		}
	}

	return urls, nil
}

// Enqueue adds a new message to the outbound queue
func (w *QueueWorker) Enqueue(tenantID, providerID uint, from, to, body string, hasMedia bool, messageID, chatMessageID *uint) error {
	item := models.MessageQueueItem{
		TenantID:      tenantID,
		ProviderID:    providerID,
		FromNumber:    from,
		ToNumber:      to,
		Body:          body,
		HasMedia:      hasMedia,
		Status:        "pending",
		MaxAttempts:   3,
		MessageID:     messageID,
		ChatMessageID: chatMessageID,
	}

	if err := w.db.Create(&item).Error; err != nil {
		return fmt.Errorf("failed to enqueue message: %w", err)
	}

	log.WithFields(log.Fields{
		"queue_item_id": item.ID,
		"from":          from,
		"to":            to,
	}).Info("Message enqueued for delivery")

	return nil
}
