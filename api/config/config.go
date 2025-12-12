package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	ServerHost string
	ServerPort string

	// Database settings
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT settings
	JWTSecret     string
	JWTExpiration int // hours

	// CORS settings
	CORSOrigins []string

	// FreeSWITCH settings
	FreeSwitchHost     string
	FreeSwitchPort     string
	FreeSwitchPassword string
	FreeSwitchAPIKey   string // API key for XML CURL authentication

	// ESL Service Addresses (configurable)
	ESLCallControlAddr string
	ESLVoicemailAddr   string
	ESLConferenceAddr  string
	ESLQueueAddr       string

	// ClickHouse settings (for CDR storage)
	ClickHouseEnabled bool
	ClickHouseHost    string
	ClickHousePort    string
	ClickHouseDB      string
	ClickHouseUser    string
	ClickHousePass    string

	// Logging settings
	LogLevel  string // debug, info, warn, error
	LogFormat string // json, text

	// Loki settings
	LokiEnabled  bool
	LokiPushURL  string
	LokiUsername string
	LokiPassword string
	LokiJob      string

	// Internal API settings
	InternalAPIKey string // Key for internal service auth (fail2ban, etc.)

	// Storage Paths
	FirmwarePath     string // Path for firmware file storage
	MediaBasePath    string // Base path for media files (sounds, music)
	ProvisioningPath string // Path for provisioning config files
	SIPProfilesPath  string // Path for FreeSWITCH SIP profile XML files
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server - matches API_HOST, API_PORT from docker-compose
		ServerHost: getEnv("API_HOST", "0.0.0.0"),
		ServerPort: getEnv("API_PORT", "8080"),

		// Database - matches POSTGRES_* from docker-compose
		DBHost:     getEnv("POSTGRES_HOST", "localhost"),
		DBPort:     getEnv("POSTGRES_PORT", "5432"),
		DBUser:     getEnv("POSTGRES_USER", "callsign"),
		DBPassword: getEnv("POSTGRES_PASSWORD", ""),
		DBName:     getEnv("POSTGRES_DB", "callsign"),
		DBSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),

		// JWT
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiration: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),

		// CORS
		CORSOrigins: []string{getEnv("CORS_ORIGINS", "*")},

		// FreeSWITCH - matches FREESWITCH_* from docker-compose
		FreeSwitchHost:     getEnv("FREESWITCH_HOST", "127.0.0.1"),
		FreeSwitchPort:     getEnv("FREESWITCH_ESL_PORT", "8021"),
		FreeSwitchPassword: getEnv("FREESWITCH_ESL_PASSWORD", "ClueCon"),
		FreeSwitchAPIKey:   getEnv("FREESWITCH_API_KEY", ""),

		// ESL Service Addresses
		ESLCallControlAddr: getEnv("ESL_CALLCONTROL_ADDR", "127.0.0.1:9001"),
		ESLVoicemailAddr:   getEnv("ESL_VOICEMAIL_ADDR", "127.0.0.2:9001"),
		ESLConferenceAddr:  getEnv("ESL_CONFERENCE_ADDR", "127.0.0.4:9001"),
		ESLQueueAddr:       getEnv("ESL_QUEUE_ADDR", "127.0.0.5:9001"),

		// ClickHouse - matches CLICKHOUSE_* from docker-compose
		ClickHouseEnabled: getEnvAsBool("CLICKHOUSE_ENABLED", false),
		ClickHouseHost:    getEnv("CLICKHOUSE_HOST", "127.0.0.1"),
		ClickHousePort:    getEnv("CLICKHOUSE_PORT", "9000"),
		ClickHouseDB:      getEnv("CLICKHOUSE_DB", "callsign"),
		ClickHouseUser:    getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePass:    getEnv("CLICKHOUSE_PASSWORD", ""),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "text"),

		// Loki - matches LOKI_URL from docker-compose
		LokiEnabled:  getEnvAsBool("LOKI_ENABLED", false),
		LokiPushURL:  getEnv("LOKI_URL", ""),
		LokiUsername: getEnv("LOKI_USERNAME", ""),
		LokiPassword: getEnv("LOKI_PASSWORD", ""),
		LokiJob:      getEnv("LOKI_JOB", "callsign-api"),

		// Internal API
		InternalAPIKey: getEnv("INTERNAL_API_KEY", "callsign-internal-key"),

		// Storage Paths
		FirmwarePath:     getEnv("FIRMWARE_PATH", "/usr/share/freeswitch/firmware"),
		MediaBasePath:    getEnv("MEDIA_PATH", "/usr/share/freeswitch/sounds"),
		ProvisioningPath: getEnv("PROVISIONING_PATH", "/var/lib/freeswitch/provisioning"),
		SIPProfilesPath:  getEnv("SIP_PROFILES_PATH", "/etc/freeswitch/sip_profiles"),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as integer or returns a default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvAsBool retrieves an environment variable as boolean or returns a default
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
