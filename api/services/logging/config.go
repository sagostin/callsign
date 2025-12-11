package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Config holds logging configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, stderr, or file path
	Loki   LokiConfig
}

// DefaultConfig returns default logging configuration
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
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

	var lokiClient *LokiClient
	if cfg.Loki.Enabled {
		lokiClient = NewLokiClient(cfg.Loki)
		logrus.Infof("Loki logging enabled: %s (job: %s)", cfg.Loki.PushURL, cfg.Loki.Job)
	} else {
		lokiClient = NewLokiClient(LokiConfig{Enabled: false})
		logrus.Debug("Loki logging disabled")
	}

	return NewLogManager(lokiClient)
}
