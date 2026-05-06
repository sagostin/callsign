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

	body := fmt.Sprintf(
		"You have a new voicemail on extension %s.\n\n"+
			"From: %s (%s)\n"+
			"Duration: %d seconds\n\n"+
			"The voicemail recording is attached to this email.\n",
		extension, callerName, callerID, duration,
	)

	return s.sendWithAttachment(to, subject, body, wavFilePath)
}

// SendMissedCallNotification sends a missed call notification email for a ring group
func (s *Service) SendMissedCallNotification(to, ringGroupName, ringGroupExt, callerID, callerName string) error {
	if !s.IsEnabled() {
		return nil
	}

	subject := fmt.Sprintf("Missed Call on %s (%s)", ringGroupName, ringGroupExt)

	body := fmt.Sprintf(
		"A call to %s (%s) was not answered.\n\n"+
			"From: %s (%s)\n\n"+
			"Please follow up with the caller as needed.\n",
		ringGroupName, ringGroupExt, callerName, callerID,
	)

	return s.send(to, subject, body)
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(to, resetToken string) error {
	if !s.IsEnabled() {
		return nil
	}

	subject := "Password Reset Request - CallSign PBX"

	body := fmt.Sprintf(
		"You requested a password reset for your CallSign PBX account.\n\n"+
			"Your reset token is: %s\n\n"+
			"If you did not request this, please ignore this email.\n",
		resetToken,
	)

	return s.send(to, subject, body)
}

// send sends a plain text email
func (s *Service) send(to, subject, body string) error {
	cfg := s.config
	from := fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddress)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		from, to, subject, body)

	addr := net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort)
	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	}

	err := smtp.SendMail(addr, auth, cfg.FromAddress, []string{to}, []byte(msg))
	if err != nil {
		log.WithError(err).Errorf("Failed to send email to %s", to)
		return err
	}

	log.WithFields(log.Fields{"to": to, "subject": subject}).Info("Email sent")
	return nil
}

// sendWithAttachment sends an email with a file attachment
func (s *Service) sendWithAttachment(to, subject, body, attachmentPath string) error {
	cfg := s.config
	from := fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddress)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	headers := make(textproto.MIMEHeader)
	headers.Set("From", from)
	headers.Set("To", to)
	headers.Set("Subject", subject)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary()))

	for k, v := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
	}
	buf.WriteString("\r\n")

	textPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"text/plain; charset=utf-8"},
	})
	if err != nil {
		return fmt.Errorf("create text part: %w", err)
	}
	textPart.Write([]byte(body))

	if attachmentPath != "" {
		fileData, err := os.ReadFile(attachmentPath)
		if err != nil {
			log.WithError(err).Warnf("Could not read voicemail attachment: %s", attachmentPath)
		} else {
			filename := filepath.Base(attachmentPath)
			attachPart, err := writer.CreatePart(textproto.MIMEHeader{
				"Content-Type":              {mime.TypeByExtension(filepath.Ext(filename))},
				"Content-Transfer-Encoding": {"base64"},
				"Content-Disposition":      {fmt.Sprintf("attachment; filename=%q", filename)},
			})
			if err == nil {
				encoded := base64.StdEncoding.EncodeToString(fileData)
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