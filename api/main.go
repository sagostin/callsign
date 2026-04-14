package main

import (
	"callsign/config"
	"callsign/handlers/freeswitch"
	"callsign/models"
	"callsign/router"
	"callsign/services/broadcast"
	"callsign/services/cdr"
	emailsvc "callsign/services/email"
	"callsign/services/esl"
	"callsign/services/esl/modules/blf"
	"callsign/services/esl/modules/callcontrol"
	conferencemod "callsign/services/esl/modules/conference"
	"callsign/services/esl/modules/featurecodes"
	"callsign/services/esl/modules/ivr"
	"callsign/services/esl/modules/queue"
	"callsign/services/esl/modules/voicemail"
	"callsign/services/fax"
	"callsign/services/logging"
	"callsign/services/tts"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"

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
		Method: cfg.LogMethod,
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

	// Seed default SIP profiles (internal, external) from built-in FusionPBX-style defaults
	// if they don't already exist in the database. No disk import — DB is source of truth.
	if err := models.EnsureDefaultProfiles(db); err != nil {
		logManager.Warn("STARTUP", "Failed to seed default SIP profiles: "+err.Error(), nil)
		log.Warnf("Failed to seed default SIP profiles: %v", err)
	}

	// Sync DB profiles to disk (sofia.conf loads profiles via X-PRE-PROCESS from sip_profiles/*.xml)
	profileSyncer := freeswitch.NewProfileSyncer(cfg.SIPProfilesPath, db)
	if err := profileSyncer.SyncProfilesToFiles(); err != nil {
		logManager.Warn("STARTUP", "Failed to sync SIP profiles to disk: "+err.Error(), nil)
		log.Warnf("Failed to sync SIP profiles to disk: %v", err)
	}

	// Initialize ESL Manager
	eslManager := esl.NewManager(cfg, db)

	// Register ESL modules — these handle incoming calls from FreeSWITCH
	// via outbound ESL sockets. Each module listens on its own loopback address.
	eslManager.RegisterModule(callcontrol.New())
	eslManager.RegisterModule(voicemail.New())
	eslManager.RegisterModule(queue.New())
	eslManager.RegisterModule(ivr.New())

	// Conference & feature codes modules need DB access for live-control APIs
	confService := conferencemod.New(db)
	eslManager.RegisterModule(confService)
	eslManager.RegisterModule(featurecodes.New(db))

	// Initialize BLF/Presence service — handles PRESENCE_PROBE events from FreeSWITCH
	// to update BLF lamp states (DND, forward, voicemail, call flow, agent, extension presence)
	blfService := blf.New(db)

	// Start ESL manager (connects to FreeSWITCH, inits + starts all modules)
	go func() {
		if err := eslManager.Start(); err != nil {
			logManager.Error("ESL", "Failed to start ESL manager: "+err.Error(), nil)
			log.Errorf("Failed to start ESL manager: %v", err)
		} else {
			logManager.Info("ESL", "ESL manager started successfully", nil)

			// Wire BLF service to handle PRESENCE_PROBE events from the ESL event processor
			eslManager.Processor.On("PRESENCE_PROBE", func(event *eventsocket.Event, session *esl.CallSession) {
				if eslManager.Client != nil {
					conn := eslManager.Client.Conn()
					if conn != nil {
						blfService.HandlePresenceProbe(conn, event)
					}
				}
			})

			// Update BLF presence when extensions answer/hang up calls
			eslManager.Processor.On("CHANNEL_ANSWER", func(event *eventsocket.Event, session *esl.CallSession) {
				ext := event.Get("Caller-Caller-ID-Number")
				domain := event.Get("variable_domain_name")
				if ext != "" && domain != "" {
					eslManager.SendExtensionPresence(ext, domain, true)
				}
			})

			eslManager.Processor.On("CHANNEL_HANGUP_COMPLETE", func(event *eventsocket.Event, session *esl.CallSession) {
				ext := event.Get("Caller-Caller-ID-Number")
				domain := event.Get("variable_domain_name")
				if ext != "" && domain != "" {
					eslManager.SendExtensionPresence(ext, domain, false)
				}
			})
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

	// Initialize email service for voicemail notifications
	emailCfg := emailsvc.LoadFromEnv()
	if emailCfg.Enabled {
		eslManager.SetEmailService(emailsvc.New(emailCfg))
		logManager.Info("STARTUP", "Email service initialized (SMTP: "+emailCfg.SMTPHost+")", nil)
	} else {
		logManager.Info("STARTUP", "Email service not configured (set SMTP_HOST + SMTP_FROM_ADDRESS)", nil)
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

	// Initialize broadcast campaign worker
	broadcastWorker := broadcast.NewBroadcastWorker(db, eslManager)
	r.Handler.SetBroadcastWorker(broadcastWorker)

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
