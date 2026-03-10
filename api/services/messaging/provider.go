package messaging

import (
	"context"
	"net/http"
	"time"
)

// SendRequest represents a request to send an SMS/MMS message
type SendRequest struct {
	From      string      `json:"from"`       // E.164 from number
	To        string      `json:"to"`         // E.164 to number
	Body      string      `json:"body"`       // Message text
	MediaURLs []string    `json:"media_urls"` // MMS media URLs
	Media     []MediaItem `json:"media"`      // MMS media items (if uploading)
}

// MediaItem represents a single media attachment
type MediaItem struct {
	ContentType string `json:"content_type"` // image/jpeg, video/mp4, etc.
	Data        []byte `json:"data"`         // Raw media bytes
	FileName    string `json:"file_name"`
	URL         string `json:"url"` // External URL (alternative to Data)
}

// SendResponse represents the result of sending a message
type SendResponse struct {
	MessageID   string    `json:"message_id"` // Provider message ID
	Status      string    `json:"status"`     // queued, sent, failed
	Segments    int       `json:"segments"`   // Number of SMS segments
	Cost        float64   `json:"cost"`       // Cost in USD
	SubmittedAt time.Time `json:"submitted_at"`
}

// InboundMessage represents an incoming SMS/MMS message from a carrier
type InboundMessage struct {
	MessageID  string      `json:"message_id"` // Provider message ID
	From       string      `json:"from"`       // E.164 sender number
	To         string      `json:"to"`         // E.164 recipient number (our DID)
	Body       string      `json:"body"`       // Message text
	MediaURLs  []string    `json:"media_urls"` // MMS media URLs
	Media      []MediaItem `json:"media"`      // Downloaded media
	ReceivedAt time.Time   `json:"received_at"`
	Segments   int         `json:"segments"`
	Direction  string      `json:"direction"` // inbound
}

// StatusUpdate represents a delivery status callback from a carrier
type StatusUpdate struct {
	MessageID string    `json:"message_id"` // Provider message ID
	Status    string    `json:"status"`     // queued, sent, delivered, failed, undeliverable
	ErrorCode string    `json:"error_code"`
	ErrorMsg  string    `json:"error_msg"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SMSProvider defines the interface for SMS/MMS carrier integrations
type SMSProvider interface {
	// Name returns the provider name (e.g., "telnyx", "twilio")
	Name() string

	// SendSMS sends a text-only SMS message
	SendSMS(ctx context.Context, req SendRequest) (*SendResponse, error)

	// SendMMS sends an MMS message with media attachments
	SendMMS(ctx context.Context, req SendRequest) (*SendResponse, error)

	// ParseInboundWebhook parses an inbound message webhook from the carrier
	ParseInboundWebhook(r *http.Request) (*InboundMessage, error)

	// ParseStatusWebhook parses a delivery status webhook from the carrier
	ParseStatusWebhook(r *http.Request) (*StatusUpdate, error)

	// VerifyWebhook verifies the webhook signature (returns error if invalid)
	VerifyWebhook(r *http.Request, secret string) error
}
