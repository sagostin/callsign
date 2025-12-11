package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogManager manages log templates and handles dispatching logs to Loki.
type LogManager struct {
	Templates  map[string]string
	LokiClient *LokiClient
	LogChannel chan *LogEntry
	wg         sync.WaitGroup
	closed     bool
	mu         sync.RWMutex
}

// LogEntry represents the structure of a log message.
type LogEntry struct {
	Message        string                 `json:"message,omitempty"`
	Error          error                  `json:"error,omitempty"`
	Type           string                 `json:"type,omitempty"`
	Level          logrus.Level           `json:"level,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
	Timestamp      time.Time              `json:"timestamp,omitempty"`
}

// LokiConfig holds Loki configuration
type LokiConfig struct {
	Enabled  bool
	PushURL  string
	Username string
	Password string
	Job      string
}

// LokiClient handles interactions with the Loki service.
type LokiClient struct {
	Config     LokiConfig
	httpClient *http.Client
}

// LokiPushData represents the data structure required by Loki's push API.
type LokiPushData struct {
	Streams []LokiStream `json:"streams"`
}

// LokiStream represents a stream of logs with the same labels in Loki.
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"` // Array of [timestamp, line] tuples
}

// NewLokiClient initializes a new Loki client.
func NewLokiClient(cfg LokiConfig) *LokiClient {
	return &LokiClient{
		Config: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // Prevent lockups with timeout
		},
	}
}

// IsEnabled returns whether Loki is enabled
func (c *LokiClient) IsEnabled() bool {
	return c.Config.Enabled && c.Config.PushURL != ""
}

// PushLog sends a log entry to Loki.
func (c *LokiClient) PushLog(labels map[string]string, timestamp time.Time, line string) error {
	if !c.IsEnabled() {
		return nil // Silently skip if disabled
	}

	payload := LokiPushData{
		Streams: []LokiStream{
			{
				Stream: labels,
				Values: [][2]string{
					{strconv.FormatInt(timestamp.UnixNano(), 10), line},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.Config.PushURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Config.Username != "" && c.Config.Password != "" {
		req.SetBasicAuth(c.Config.Username, c.Config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from Loki: %d", resp.StatusCode)
	}

	return nil
}

// NewLogManager initializes a new LogManager.
func NewLogManager(lokiClient *LokiClient) *LogManager {
	lm := &LogManager{
		Templates:  make(map[string]string),
		LokiClient: lokiClient,
		LogChannel: make(chan *LogEntry, 1000), // Buffered to prevent blocking
		closed:     false,
	}
	lm.LoadTemplates()
	lm.wg.Add(1)
	go lm.processLogChannel()
	return lm
}

// LoadTemplates loads default log templates
func (lm *LogManager) LoadTemplates() {
	templates := map[string]string{
		"GENERIC_ERROR":       "An error occurred: %v",
		"UNEXPECTED_ERROR":    "Unexpected error: %v",
		"UNHANDLED_EXCEPTION": "Unhandled exception: %v",
		"REQUEST_RECEIVED":    "Request received: %s %s",
		"REQUEST_COMPLETED":   "Request completed: %s %s [%d] %s",
		"DB_ERROR":            "Database error: %v",
		"AUTH_FAILED":         "Authentication failed: %s",
		"AUTH_SUCCESS":        "Authentication success: %s",
		"CONFIG_LOADED":       "Configuration loaded successfully",
		"SERVER_STARTED":      "Server started on %s",
		"SERVER_SHUTDOWN":     "Server shutting down",
	}

	for name, template := range templates {
		lm.Templates[strings.ToUpper(name)] = template
	}
}

// AddTemplate adds a new log template to the manager.
func (lm *LogManager) AddTemplate(name, template string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.Templates[strings.ToUpper(name)] = template
}

// formatTemplate formats a template with provided arguments.
func (lm *LogManager) formatTemplate(templateName string, args ...interface{}) string {
	lm.mu.RLock()
	template, exists := lm.Templates[strings.ToUpper(templateName)]
	lm.mu.RUnlock()

	if !exists {
		// If no template found, use templateName as the format string directly
		return fmt.Sprintf(templateName, args...)
	}
	return fmt.Sprintf(template, args...)
}

// Log creates and sends a log entry
func (lm *LogManager) Log(level logrus.Level, logType string, message string, fields map[string]interface{}) {
	entry := &LogEntry{
		Message:        message,
		Type:           strings.ToUpper(logType),
		Level:          level,
		AdditionalData: fields,
		Timestamp:      time.Now(),
	}
	lm.SendLog(entry)
}

// LogTemplate creates and sends a log using a template
func (lm *LogManager) LogTemplate(level logrus.Level, logType string, templateName string, fields map[string]interface{}, args ...interface{}) {
	message := lm.formatTemplate(templateName, args...)
	lm.Log(level, logType, message, fields)
}

// Convenience methods for different log levels
func (lm *LogManager) Info(logType, message string, fields map[string]interface{}) {
	lm.Log(logrus.InfoLevel, logType, message, fields)
}

func (lm *LogManager) Debug(logType, message string, fields map[string]interface{}) {
	lm.Log(logrus.DebugLevel, logType, message, fields)
}

func (lm *LogManager) Warn(logType, message string, fields map[string]interface{}) {
	lm.Log(logrus.WarnLevel, logType, message, fields)
}

func (lm *LogManager) Error(logType, message string, fields map[string]interface{}) {
	lm.Log(logrus.ErrorLevel, logType, message, fields)
}

func (lm *LogManager) ErrorWithErr(logType string, err error, fields map[string]interface{}) {
	entry := &LogEntry{
		Message:        err.Error(),
		Error:          err,
		Type:           strings.ToUpper(logType),
		Level:          logrus.ErrorLevel,
		AdditionalData: fields,
		Timestamp:      time.Now(),
	}
	lm.SendLog(entry)
}

// SendLog sends a log to Loki asynchronously via the log channel.
func (lm *LogManager) SendLog(log *LogEntry) {
	// Always print locally first
	log.Print()

	// Check if closed before sending to channel
	lm.mu.RLock()
	closed := lm.closed
	lm.mu.RUnlock()

	if closed {
		return
	}

	// Non-blocking send to prevent lockups
	select {
	case lm.LogChannel <- log:
		// Sent successfully
	default:
		// Channel full, log warning locally but don't block
		logrus.Warn("Log channel full, dropping log entry")
	}
}

// processLogChannel processes logs from the channel and sends them to Loki.
func (lm *LogManager) processLogChannel() {
	defer lm.wg.Done()

	for log := range lm.LogChannel {
		if lm.LokiClient == nil || !lm.LokiClient.IsEnabled() {
			continue
		}

		labels := map[string]string{
			"job":   lm.LokiClient.Config.Job,
			"type":  log.Type,
			"level": log.Level.String(),
		}

		logLine := log.String()
		if err := lm.LokiClient.PushLog(labels, log.Timestamp, logLine); err != nil {
			// Only log error locally, don't create infinite loop
			logrus.Debugf("Failed to send log to Loki: %v", err)
		}
	}
}

// Print outputs the log locally (stdout via logrus).
func (le *LogEntry) Print() {
	logEntry := logrus.WithFields(logrus.Fields{
		"type": le.Type,
	})

	for key, value := range le.AdditionalData {
		logEntry = logEntry.WithField(key, value)
	}

	if le.Error != nil {
		logEntry = logEntry.WithError(le.Error)
	}

	switch le.Level {
	case logrus.ErrorLevel:
		logEntry.Error(le.Message)
	case logrus.WarnLevel:
		logEntry.Warn(le.Message)
	case logrus.DebugLevel:
		logEntry.Debug(le.Message)
	case logrus.TraceLevel:
		logEntry.Trace(le.Message)
	default:
		logEntry.Info(le.Message)
	}
}

// String serializes the LogEntry into JSON for Loki.
func (le *LogEntry) String() string {
	// Create a copy without the error (which doesn't serialize well)
	data := map[string]interface{}{
		"message":   le.Message,
		"type":      le.Type,
		"level":     le.Level.String(),
		"timestamp": le.Timestamp.Format(time.RFC3339),
	}

	if le.Error != nil {
		data["error"] = le.Error.Error()
	}

	for k, v := range le.AdditionalData {
		data[k] = v
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Error serializing log: %v", err)
	}
	return string(jsonData)
}

// Close gracefully shuts down the log manager and waits for the log channel to empty.
func (lm *LogManager) Close() {
	lm.mu.Lock()
	if lm.closed {
		lm.mu.Unlock()
		return
	}
	lm.closed = true
	lm.mu.Unlock()

	close(lm.LogChannel)
	lm.wg.Wait()
}

// AddField adds a new field to an already built log entry.
func (le *LogEntry) AddField(key string, value interface{}) *LogEntry {
	if le.AdditionalData == nil {
		le.AdditionalData = make(map[string]interface{})
	}
	le.AdditionalData[key] = value
	return le
}
