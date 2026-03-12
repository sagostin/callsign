package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Config holds SMTP configuration
type Config struct {
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	FromAddress string
	FromName    string
	Enabled     bool
}

// LoadFromEnv loads email config from environment variables
func LoadFromEnv() *Config {
	cfg := &Config{
		SMTPHost:    os.Getenv("SMTP_HOST"),
		SMTPPort:    os.Getenv("SMTP_PORT"),
		SMTPUser:    os.Getenv("SMTP_USER"),
		SMTPPass:    os.Getenv("SMTP_PASS"),
		FromAddress: os.Getenv("SMTP_FROM_ADDRESS"),
		FromName:    os.Getenv("SMTP_FROM_NAME"),
	}
	if cfg.SMTPPort == "" {
		cfg.SMTPPort = "587"
	}
	if cfg.FromName == "" {
		cfg.FromName = "CallSign PBX"
	}
	cfg.Enabled = cfg.SMTPHost != "" && cfg.FromAddress != ""
	return cfg
}

// Service provides email sending capabilities
type Service struct {
	config *Config
}

// New creates a new email service
func New(cfg *Config) *Service {
	return &Service{config: cfg}
}

// IsEnabled returns whether the email service is configured
func (s *Service) IsEnabled() bool {
	return s.config != nil && s.config.Enabled
}

// SendVoicemailNotification sends a voicemail notification email with .wav attachment
func (s *Service) SendVoicemailNotification(to, extension, callerID, callerName string, duration int, wavFilePath string) error {
	if !s.IsEnabled() {
		return nil
	}

	subject := fmt.Sprintf("New Voicemail from %s (%s) - %d seconds", callerName, callerID, duration)

	// Build text body
	body := fmt.Sprintf(
		"You have a new voicemail on extension %s.\n\n"+
			"From: %s (%s)\n"+
			"Duration: %d seconds\n\n"+
			"The voicemail recording is attached to this email.\n",
		extension, callerName, callerID, duration,
	)

	// Send with attachment
	return s.sendWithAttachment(to, subject, body, wavFilePath)
}

// sendWithAttachment sends an email with a file attachment
func (s *Service) sendWithAttachment(to, subject, body, attachmentPath string) error {
	cfg := s.config
	from := fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddress)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Headers
	headers := make(textproto.MIMEHeader)
	headers.Set("From", from)
	headers.Set("To", to)
	headers.Set("Subject", subject)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary()))

	// Write headers
	for k, v := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
	}
	buf.WriteString("\r\n")

	// Text part
	textPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"text/plain; charset=utf-8"},
	})
	if err != nil {
		return fmt.Errorf("create text part: %w", err)
	}
	textPart.Write([]byte(body))

	// Attachment part (if file exists)
	if attachmentPath != "" {
		fileData, err := os.ReadFile(attachmentPath)
		if err != nil {
			log.WithError(err).Warnf("Could not read voicemail attachment: %s", attachmentPath)
			// Continue without attachment
		} else {
			filename := filepath.Base(attachmentPath)
			attachPart, err := writer.CreatePart(textproto.MIMEHeader{
				"Content-Type":              {mime.TypeByExtension(filepath.Ext(filename))},
				"Content-Transfer-Encoding": {"base64"},
				"Content-Disposition":       {fmt.Sprintf("attachment; filename=%q", filename)},
			})
			if err == nil {
				encoded := base64.StdEncoding.EncodeToString(fileData)
				// Wrap at 76 chars per line
				for i := 0; i < len(encoded); i += 76 {
					end := i + 76
					if end > len(encoded) {
						end = len(encoded)
					}
					attachPart.Write([]byte(encoded[i:end] + "\r\n"))
				}
			}
		}
	}

	writer.Close()

	// Send via SMTP
	addr := net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort)
	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	}

	err = smtp.SendMail(addr, auth, cfg.FromAddress, []string{to}, buf.Bytes())
	if err != nil {
		log.WithError(err).Errorf("Failed to send voicemail email to %s", to)
		return err
	}

	log.WithFields(log.Fields{"to": to, "subject": subject}).Info("Voicemail email sent")
	return nil
}
