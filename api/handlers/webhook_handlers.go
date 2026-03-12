package handlers

import (
	"bytes"
	"callsign/services/messaging"
	"net/http"

	"github.com/gofiber/fiber/v2"
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

// fiberToHTTPRequest converts a Fiber context into a standard *http.Request.
// This is needed because provider webhook methods expect *http.Request
// but Fiber uses fasthttp under the hood.
func fiberToHTTPRequest(c *fiber.Ctx) (*http.Request, error) {
	req, err := http.NewRequest(c.Method(), c.OriginalURL(), bytes.NewReader(c.Body()))
	if err != nil {
		return nil, err
	}
	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})
	return req, nil
}

// TelnyxInbound handles inbound SMS/MMS webhooks from Telnyx
func (h *WebhookHandler) TelnyxInbound(c *fiber.Ctx) error {
	if h.MsgManager == nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "Messaging not configured"})
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
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "Telnyx provider not configured"})
	}

	// Convert Fiber request to standard http.Request for provider interface
	httpReq, err := fiberToHTTPRequest(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to convert request"})
	}

	// Verify webhook signature
	webhookSecret := h.MsgManager.Config.TelnyxWebhookSecret
	if webhookSecret != "" {
		if err := telnyxProvider.VerifyWebhook(httpReq, webhookSecret); err != nil {
			log.WithError(err).Warn("Telnyx webhook verification failed")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid webhook signature"})
		}
	}

	// Parse the inbound message
	inbound, err := telnyxProvider.ParseInboundWebhook(httpReq)
	if err != nil {
		log.WithError(err).Error("Failed to parse Telnyx inbound webhook")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Failed to parse webhook"})
	}

	// Route the inbound SMS
	if err := h.MsgManager.RouteInboundSMS(inbound.To, inbound.From, inbound.Body, inbound.MediaURLs); err != nil {
		log.WithError(err).Error("Failed to route inbound SMS")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process inbound message"})
	}

	log.WithFields(log.Fields{
		"from":       inbound.From,
		"to":         inbound.To,
		"message_id": inbound.MessageID,
	}).Info("Inbound SMS processed")

	// Telnyx expects 200 OK
	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "ok"})
}

// TelnyxStatus handles delivery status webhooks from Telnyx
func (h *WebhookHandler) TelnyxStatus(c *fiber.Ctx) error {
	if h.MsgManager == nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "Messaging not configured"})
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
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "Telnyx provider not configured"})
	}

	// Convert Fiber request to standard http.Request for provider interface
	httpReq, err := fiberToHTTPRequest(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to convert request"})
	}

	// Parse the status update
	status, err := telnyxProvider.ParseStatusWebhook(httpReq)
	if err != nil {
		log.WithError(err).Error("Failed to parse Telnyx status webhook")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Failed to parse webhook"})
	}

	// Handle the status update
	if err := h.MsgManager.HandleStatusUpdate(status); err != nil {
		log.WithError(err).Error("Failed to handle status update")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "ok"})
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
