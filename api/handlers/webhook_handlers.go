package handlers

import (
	"callsign/services/messaging"
	"net/http"

	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
)

// WebhookHandler handles inbound carrier webhooks
type WebhookHandler struct {
	MsgManager *messaging.Manager
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(mgr *messaging.Manager) *WebhookHandler {
	return &WebhookHandler{MsgManager: mgr}
}

// TelnyxInbound handles inbound SMS/MMS webhooks from Telnyx
func (h *WebhookHandler) TelnyxInbound(ctx iris.Context) {
	if h.MsgManager == nil {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "Messaging not configured"})
		return
	}

	// Find the Telnyx provider
	var telnyxProvider messaging.SMSProvider
	for _, p := range h.getProviders() {
		if p.Name() == "telnyx" {
			telnyxProvider = p
			break
		}
	}

	if telnyxProvider == nil {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "Telnyx provider not configured"})
		return
	}

	// Verify webhook signature
	webhookSecret := h.MsgManager.Config.TelnyxWebhookSecret
	if webhookSecret != "" {
		if err := telnyxProvider.VerifyWebhook(ctx.Request(), webhookSecret); err != nil {
			log.WithError(err).Warn("Telnyx webhook verification failed")
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Invalid webhook signature"})
			return
		}
	}

	// Parse the inbound message
	inbound, err := telnyxProvider.ParseInboundWebhook(ctx.Request())
	if err != nil {
		log.WithError(err).Error("Failed to parse Telnyx inbound webhook")
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Failed to parse webhook"})
		return
	}

	// Route the inbound SMS
	if err := h.MsgManager.RouteInboundSMS(inbound.To, inbound.From, inbound.Body, inbound.MediaURLs); err != nil {
		log.WithError(err).Error("Failed to route inbound SMS")
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to process inbound message"})
		return
	}

	log.WithFields(log.Fields{
		"from":       inbound.From,
		"to":         inbound.To,
		"message_id": inbound.MessageID,
	}).Info("Inbound SMS processed")

	// Telnyx expects 200 OK
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"status": "ok"})
}

// TelnyxStatus handles delivery status webhooks from Telnyx
func (h *WebhookHandler) TelnyxStatus(ctx iris.Context) {
	if h.MsgManager == nil {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "Messaging not configured"})
		return
	}

	// Find the Telnyx provider
	var telnyxProvider messaging.SMSProvider
	for _, p := range h.getProviders() {
		if p.Name() == "telnyx" {
			telnyxProvider = p
			break
		}
	}

	if telnyxProvider == nil {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "Telnyx provider not configured"})
		return
	}

	// Parse the status update
	status, err := telnyxProvider.ParseStatusWebhook(ctx.Request())
	if err != nil {
		log.WithError(err).Error("Failed to parse Telnyx status webhook")
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Failed to parse webhook"})
		return
	}

	// Handle the status update
	if err := h.MsgManager.HandleStatusUpdate(status); err != nil {
		log.WithError(err).Error("Failed to handle status update")
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"status": "ok"})
}

// getProviders returns all configured providers from the manager
func (h *WebhookHandler) getProviders() []messaging.SMSProvider {
	// Iterate through known provider IDs
	// In a real system, we'd have a better way to enumerate providers
	var providers []messaging.SMSProvider
	for id := uint(1); id <= 100; id++ {
		if p, ok := h.MsgManager.GetProvider(id); ok {
			providers = append(providers, p)
		}
	}
	return providers
}
