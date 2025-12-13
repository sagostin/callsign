package main

import (
	"callsign/config"
	"callsign/handlers/freeswitch"
	"callsign/models"
	"callsign/router"
	"callsign/services/esl"
	"callsign/services/logging"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables first
	if err := godotenv.Load(); err != nil {
		log.Debug("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize logging with Loki support
	logConfig := logging.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Output: "stdout",
		Loki: logging.LokiConfig{
			Enabled:  cfg.LokiEnabled,
			PushURL:  cfg.LokiPushURL,
			Username: cfg.LokiUsername,
			Password: cfg.LokiPassword,
			Job:      cfg.LokiJob,
		},
	}
	logManager := logging.NewLogManagerFromConfig(logConfig)
	defer logManager.Close()

	logManager.Info("STARTUP", "Starting CallSign API Server...", nil)

	// Initialize database connection
	db, err := models.InitDB(cfg)
	if err != nil {
		logManager.Error("STARTUP", "Failed to connect to database: "+err.Error(), nil)
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	if err := models.AutoMigrate(db); err != nil {
		logManager.Error("STARTUP", "Failed to run database migrations: "+err.Error(), nil)
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Run database seeds (creates default admin if no users exist)
	if err := models.RunSeeds(db); err != nil {
		logManager.Error("STARTUP", "Failed to run database seeds: "+err.Error(), nil)
		log.Fatalf("Failed to run database seeds: %v", err)
	}

	// Import SIP profiles from disk on first boot (if DB is empty)
	// After import, profiles are managed via DB and synced back to files
	profileImporter := freeswitch.NewProfileImporter(cfg.SIPProfilesPath, db)
	// Pass false to not overwrite existing profiles on boot (safe default)
	if err := profileImporter.SyncProfiles(false); err != nil {
		logManager.Warn("STARTUP", "Failed to import SIP profiles: "+err.Error(), nil)
		log.Warnf("Failed to import SIP profiles: %v", err)
	}

	// Initialize and start ESL Manager
	eslManager := esl.NewManager(cfg, db)
	go func() {
		if err := eslManager.Start(); err != nil {
			logManager.Error("ESL", "Failed to start ESL manager: "+err.Error(), nil)
			log.Errorf("Failed to start ESL manager: %v", err)
		} else {
			logManager.Info("ESL", "ESL manager started successfully", nil)
		}
	}()
	defer eslManager.Stop()

	// Initialize router with ESL manager reference
	r := router.NewRouter(db, cfg)
	r.Init()
	r.Handler.SetESLManager(eslManager)
	r.Handler.SetLogManager(logManager)

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logManager.Info("SHUTDOWN", "Server shutting down...", nil)
		eslManager.Stop()
		logManager.Close() // Ensure logs are flushed to Loki
		os.Exit(0)
	}()

	// Start server
	addr := cfg.ServerHost + ":" + cfg.ServerPort
	logManager.Info("STARTUP", "Server listening on "+addr, map[string]interface{}{
		"host": cfg.ServerHost,
		"port": cfg.ServerPort,
	})
	r.Listen(addr)
}
