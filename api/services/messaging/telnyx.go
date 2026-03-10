package messaging

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	telnyxAPIBase = "https://api.telnyx.com/v2"
)

// TelnyxProvider implements SMSProvider for the Telnyx Messaging v2 API
type TelnyxProvider struct {
	apiKey           string
	messagingProfile string
	httpClient       *http.Client
}

// NewTelnyxProvider creates a new Telnyx provider
func NewTelnyxProvider(apiKey, messagingProfile string) *TelnyxProvider {
	return &TelnyxProvider{
		apiKey:           apiKey,
		messagingProfile: messagingProfile,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *TelnyxProvider) Name() string {
	return "telnyx"
}

// telnyxSendRequest is the Telnyx v2 message send payload
type telnyxSendRequest struct {
	From               string   `json:"from"`
	To                 string   `json:"to"`
	Text               string   `json:"text,omitempty"`
	MediaURLs          []string `json:"media_urls,omitempty"`
	MessagingProfileID string   `json:"messaging_profile_id,omitempty"`
	Type               string   `json:"type"` // SMS or MMS
}

// telnyxResponse wraps the Telnyx API response
type telnyxResponse struct {
	Data struct {
		ID         string `json:"id"`
		RecordType string `json:"record_type"`
		Direction  string `json:"direction"`
		Type       string `json:"type"`
		Parts      int    `json:"parts"`
		Cost       *struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"cost"`
		To []struct {
			PhoneNumber string `json:"phone_number"`
			Status      string `json:"status"`
		} `json:"to"`
	} `json:"data"`
	Errors []struct {
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

// telnyxWebhookPayload represents a Telnyx webhook event
type telnyxWebhookPayload struct {
	Data struct {
		EventType string `json:"event_type"`
		ID        string `json:"id"`
		Payload   struct {
			ID        string `json:"id"`
			Direction string `json:"direction"`
			Type      string `json:"type"`
			From      struct {
				PhoneNumber string `json:"phone_number"`
			} `json:"from"`
			To []struct {
				PhoneNumber string `json:"phone_number"`
				Status      string `json:"status"`
			} `json:"to"`
			Text  string `json:"text"`
			Media []struct {
				URL         string `json:"url"`
				ContentType string `json:"content_type"`
				Size        int    `json:"size"`
			} `json:"media"`
			ReceivedAt  string `json:"received_at"`
			CompletedAt string `json:"completed_at"`
			Parts       int    `json:"parts"`
			Errors      []struct {
				Code   string `json:"code"`
				Title  string `json:"title"`
				Detail string `json:"detail"`
			} `json:"errors"`
		} `json:"payload"`
	} `json:"data"`
}

// SendSMS sends a text-only message via Telnyx
func (t *TelnyxProvider) SendSMS(ctx context.Context, req SendRequest) (*SendResponse, error) {
	return t.send(ctx, telnyxSendRequest{
		From:               req.From,
		To:                 req.To,
		Text:               req.Body,
		Type:               "SMS",
		MessagingProfileID: t.messagingProfile,
	})
}

// SendMMS sends a message with media via Telnyx
func (t *TelnyxProvider) SendMMS(ctx context.Context, req SendRequest) (*SendResponse, error) {
	return t.send(ctx, telnyxSendRequest{
		From:               req.From,
		To:                 req.To,
		Text:               req.Body,
		MediaURLs:          req.MediaURLs,
		Type:               "MMS",
		MessagingProfileID: t.messagingProfile,
	})
}

// send makes the actual API call to Telnyx
func (t *TelnyxProvider) send(ctx context.Context, body telnyxSendRequest) (*SendResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", telnyxAPIBase+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("telnyx API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var telResp telnyxResponse
	if err := json.Unmarshal(respBody, &telResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 || len(telResp.Errors) > 0 {
		errMsg := fmt.Sprintf("telnyx API error (status %d)", resp.StatusCode)
		if len(telResp.Errors) > 0 {
			errMsg = fmt.Sprintf("%s: %s - %s", errMsg, telResp.Errors[0].Title, telResp.Errors[0].Detail)
		}
		return nil, fmt.Errorf("telnyx API error: %s", errMsg)
	}

	log.WithFields(log.Fields{
		"message_id": telResp.Data.ID,
		"from":       body.From,
		"to":         body.To,
		"type":       body.Type,
		"parts":      telResp.Data.Parts,
	}).Info("Telnyx message sent")

	return &SendResponse{
		MessageID:   telResp.Data.ID,
		Status:      "queued",
		Segments:    telResp.Data.Parts,
		SubmittedAt: time.Now(),
	}, nil
}

// ParseInboundWebhook parses a Telnyx inbound message webhook
func (t *TelnyxProvider) ParseInboundWebhook(r *http.Request) (*InboundMessage, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read webhook body: %w", err)
	}
	// Reset body for potential re-read
	r.Body = io.NopCloser(bytes.NewReader(body))

	var webhook telnyxWebhookPayload
	if err := json.Unmarshal(body, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	payload := webhook.Data.Payload

	// Extract recipient number (first 'to' entry)
	toNumber := ""
	if len(payload.To) > 0 {
		toNumber = payload.To[0].PhoneNumber
	}

	// Extract media URLs
	var mediaURLs []string
	for _, m := range payload.Media {
		mediaURLs = append(mediaURLs, m.URL)
	}

	receivedAt := time.Now()
	if payload.ReceivedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, payload.ReceivedAt); err == nil {
			receivedAt = parsed
		}
	}

	return &InboundMessage{
		MessageID:  payload.ID,
		From:       payload.From.PhoneNumber,
		To:         toNumber,
		Body:       payload.Text,
		MediaURLs:  mediaURLs,
		ReceivedAt: receivedAt,
		Segments:   payload.Parts,
		Direction:  "inbound",
	}, nil
}

// ParseStatusWebhook parses a Telnyx delivery status webhook
func (t *TelnyxProvider) ParseStatusWebhook(r *http.Request) (*StatusUpdate, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read webhook body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var webhook telnyxWebhookPayload
	if err := json.Unmarshal(body, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	payload := webhook.Data.Payload

	// Map Telnyx event types to normalized statuses
	status := mapTelnyxEventToStatus(webhook.Data.EventType)

	var errorCode, errorMsg string
	if len(payload.Errors) > 0 {
		errorCode = payload.Errors[0].Code
		errorMsg = payload.Errors[0].Detail
	}

	return &StatusUpdate{
		MessageID: payload.ID,
		Status:    status,
		ErrorCode: errorCode,
		ErrorMsg:  errorMsg,
		UpdatedAt: time.Now(),
	}, nil
}

// VerifyWebhook verifies the Telnyx webhook signature
func (t *TelnyxProvider) VerifyWebhook(r *http.Request, secret string) error {
	if secret == "" {
		return nil // Skip verification if no secret configured
	}

	signature := r.Header.Get("telnyx-signature-ed25519")
	timestamp := r.Header.Get("telnyx-timestamp")

	if signature == "" || timestamp == "" {
		return fmt.Errorf("missing telnyx signature headers")
	}

	// Read body for verification
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	// HMAC-SHA256 verification
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

// mapTelnyxEventToStatus maps Telnyx webhook event types to normalized statuses
func mapTelnyxEventToStatus(eventType string) string {
	switch eventType {
	case "message.sent":
		return "sent"
	case "message.finalized":
		return "delivered"
	case "message.failed":
		return "failed"
	case "message.received":
		return "delivered"
	default:
		return "unknown"
	}
}
