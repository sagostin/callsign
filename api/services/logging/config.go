package logging

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Config holds logging configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, stderr, or file path
	Method string // "standard" (stdout only) or "loki" (stdout + Loki push)
	Loki   LokiConfig
}

// DefaultConfig returns default logging configuration
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
		Method: "standard",
		Loki: LokiConfig{
			Enabled: false,
			PushURL: "",
			Job:     "callsign-api",
		},
	}
}

// SetupLogrus configures the global logrus logger
func SetupLogrus(cfg Config) {
	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Set formatter
	if cfg.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logrus.Warnf("Failed to open log file %s, falling back to stdout: %v", cfg.Output, err)
			output = os.Stdout
		} else {
			output = file
		}
	}
	logrus.SetOutput(output)
}

// NewLogManagerFromConfig creates a LogManager from a Config
func NewLogManagerFromConfig(cfg Config) *LogManager {
	SetupLogrus(cfg)

	method := strings.ToLower(cfg.Method)
	if method == "" {
		method = "standard"
	}

	// Only create Loki client if method is "loki" AND Loki config is enabled
	if method == "loki" && cfg.Loki.Enabled {
		lokiClient := NewLokiClient(cfg.Loki)
		logrus.Infof("Loki logging enabled: %s (job: %s)", cfg.Loki.PushURL, cfg.Loki.Job)
		return NewLogManager(lokiClient, true)
	}

	logrus.Infof("Logging method: %s (stdout only)", method)
	return NewLogManager(nil, false)
}
