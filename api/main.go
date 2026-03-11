package main

import (
	"callsign/config"
	"callsign/handlers/freeswitch"
	"callsign/models"
	"callsign/router"
	"callsign/services/cdr"
	"callsign/services/esl"
	conferencemod "callsign/services/esl/modules/conference"
	"callsign/services/fax"
	"callsign/services/logging"
	"callsign/services/tts"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Initialize TTS caching service
	ttsService, err := tts.NewService(cfg, db)
	if err != nil {
		logManager.Warn("STARTUP", "Failed to init TTS cache (inline TTS will be used): "+err.Error(), nil)
		log.Warnf("TTS cache init failed: %v", err)
	} else {
		eslManager.TTS = ttsService
		// Warm the cache with system phrases in background
		go ttsService.Init()
		logManager.Info("STARTUP", "TTS cache service initialized", nil)
	}

	// Initialize fax manager (routing, queue processing, retry strategy)
	faxManager := fax.NewManager(db, cfg, logManager)
	go func() {
		if err := faxManager.Start(); err != nil {
			logManager.Error("FAX", "Failed to start fax manager: "+err.Error(), nil)
			log.Errorf("Failed to start fax manager: %v", err)
		} else {
			logManager.Info("FAX", "Fax manager started successfully", nil)
		}
	}()

	// Initialize conference service (live control via ESL)
	confService := conferencemod.New(db)
	go func() {
		// Wait briefly for ESL to connect before initializing
		time.Sleep(2 * time.Second)
		if err := confService.Init(eslManager); err != nil {
			logManager.Error("CONFERENCE", "Failed to init conference service: "+err.Error(), nil)
			log.Errorf("Failed to init conference service: %v", err)
		} else {
			logManager.Info("CONFERENCE", "Conference service initialized", nil)
		}
	}()

	// Initialize ClickHouse CDR storage
	chClient := cdr.NewClickHouseClient(cfg)
	if err := chClient.Connect(); err != nil {
		logManager.Warn("STARTUP", "ClickHouse connect failed (CDR sync disabled): "+err.Error(), nil)
		log.Warnf("ClickHouse connect failed: %v", err)
	} else if chClient.IsEnabled() {
		logManager.Info("STARTUP", "ClickHouse connected — CDR sync enabled", nil)

		// Start periodic PG → ClickHouse sync (every 5 minutes)
		syncJob := cdr.NewSyncJob(db, chClient)
		syncJob.StartPeriodicSync(5 * time.Minute)

		// Start daily PG cleanup (remove synced records older than 90 days)
		go func() {
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				if err := syncJob.CleanupOldRecords(90); err != nil {
					log.Errorf("CDR cleanup failed: %v", err)
				}
			}
		}()
	}
	defer chClient.Close()

	// Initialize router with ESL manager reference
	r := router.NewRouter(db, cfg)
	r.Init()
	r.Handler.SetESLManager(eslManager)
	r.Handler.SetLogManager(logManager)
	r.Handler.SetClickHouse(chClient)

	// Wire fax manager and conference service into their handlers
	r.SetFaxManager(faxManager)
	r.SetConferenceService(confService)

	// Wire WebSocket hub to ESL manager for real-time event broadcasting
	// (call events, voicemail MWI, conference join/leave, queue events)
	eslManager.SetWSHub(r.WSHub)

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logManager.Info("SHUTDOWN", "Server shutting down...", nil)
		eslManager.Stop()
		chClient.Close()
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
