package router

import (
	"callsign/config"
	"callsign/handlers"
	"callsign/handlers/freeswitch"
	"callsign/middleware"
	"callsign/services/esl/modules/conference"
	"callsign/services/fax"
	"callsign/services/messaging"
	"callsign/services/websocket"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Router manages the Fiber application and routes
type Router struct {
	App               *fiber.App
	DB                *gorm.DB
	Config            *config.Config
	Auth              *middleware.AuthMiddleware
	Tenant            *middleware.TenantMiddleware
	Handler           *handlers.Handler
	FSHandler         *freeswitch.FSHandler
	ConferenceHandler *handlers.ConferenceHandler
	FaxHandler        *handlers.FaxHandler
	SMSNumberHandler  *handlers.SMSNumberHandler
	WebhookHandler    *handlers.WebhookHandler
	MsgManager        *messaging.Manager
	WSHub             *websocket.Hub
}

// NewRouter creates a new Router instance
func NewRouter(db *gorm.DB, cfg *config.Config) *Router {
	h := handlers.NewHandler(db, cfg)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize messaging manager
	msgManager := messaging.NewManager(db, cfg, wsHub)
	msgManager.Start()

	// Wire dependencies into handler
	h.SetWSHub(wsHub)
	h.SetMsgManager(msgManager)

	return &Router{
		App:               fiber.New(),
		DB:                db,
		Config:            cfg,
		Auth:              middleware.NewAuthMiddleware(cfg, db),
		Tenant:            middleware.NewTenantMiddleware(db),
		Handler:           h,
		FSHandler:         freeswitch.NewFSHandler(db, cfg),
		ConferenceHandler: handlers.NewConferenceHandler(db, nil),
		FaxHandler:        handlers.NewFaxHandler(h, nil),
		SMSNumberHandler:  handlers.NewSMSNumberHandler(db),
		WebhookHandler:    handlers.NewWebhookHandler(msgManager),
		MsgManager:        msgManager,
		WSHub:             wsHub,
	}
}

// internalKeyAuth validates the X-Internal-Key header for internal service access
func (r *Router) internalKeyAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-Internal-Key")

		// Check against configured internal key
		configuredKey := r.Config.InternalAPIKey
		if configuredKey == "" {
			configuredKey = "callsign-internal-key" // Default for development
		}

		if key == "" || key != configuredKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or missing internal API key"})
		}

		return c.Next()
	}
}

// SetFaxManager wires the fax manager into the FaxHandler after startup
func (r *Router) SetFaxManager(fm *fax.Manager) {
	r.FaxHandler.FaxManager = fm
}

// SetConferenceService wires the conference service into the ConferenceHandler after startup
func (r *Router) SetConferenceService(svc *conference.Service) {
	r.ConferenceHandler.Service = svc
}

// Init sets up all routes and middleware
func (r *Router) Init() {
	// Global middleware
	r.App.Use(middleware.Recovery())
	r.App.Use(middleware.RequestLogger())
	r.App.Use(middleware.CORS(r.Config))

	// API base group
	api := r.App.Group("/api")

	// Health check (public)
	api.Get("/health", r.Handler.Health)

	// Public authentication routes
	auth := api.Group("/auth")
	auth.Post("/login", r.Handler.Login)
	auth.Post("/admin/login", r.Handler.AdminLogin)
	auth.Post("/extension/login", r.Handler.ExtensionLogin)
	auth.Post("/register", r.Handler.Register) // If self-registration is enabled
	auth.Post("/password/reset", r.Handler.RequestPasswordReset)

	// Public WebSocket routes (auth handled inside handler via first message)
	api.Get("/system/console", r.Handler.FreeSwitchConsole)
	api.Get("/ws/notifications", r.Handler.NotificationWebSocket)

	// Device provisioning (public, authenticated via tenant secret in URL)
	// URL format: /provision/{tenant_uuid}/{secret}/{mac}.cfg
	provision := api.Group("/provision")
	provision.Get("/:tenant/:secret/:mac", r.Handler.GetDeviceConfigSecure)

	// Internal routes (authenticated via X-Internal-Key header)
	// These are for internal services like fail2ban
	internal := api.Group("/internal")
	internal.Use(r.internalKeyAuth())
	internal.Post("/fail2ban/report", r.Handler.ReportBannedIP)

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(r.Auth.RequireAuth())
	protected.Use(middleware.AuditMiddleware(r.DB))

	// Auth routes (authenticated)
	protectedAuth := protected.Group("/auth")
	protectedAuth.Get("/me", r.Handler.GetProfile)
	protectedAuth.Put("/password", r.Handler.ChangePassword)
	protectedAuth.Post("/logout", r.Handler.Logout)
	protectedAuth.Post("/refresh", r.Handler.RefreshToken)

	// Tenant-scoped routes
	tenantScoped := protected.Group("")
	tenantScoped.Use(r.Tenant.RequireTenant())

	// Tenant Settings
	tenant := tenantScoped.Group("/tenant")
	tenant.Get("/settings", r.Handler.GetTenantSettings)
	tenant.Put("/settings", r.Handler.UpdateTenantSettings)
	tenant.Get("/branding", r.Handler.GetTenantBranding)
	tenant.Put("/branding", r.Handler.UpdateTenantBranding)
	tenant.Get("/smtp", r.Handler.GetTenantSMTP)
	tenant.Put("/smtp", r.Handler.UpdateTenantSMTP)
	tenant.Post("/smtp/test", r.Handler.TestTenantSMTP)
	tenant.Get("/messaging", r.Handler.GetTenantMessaging)
	tenant.Put("/messaging", r.Handler.UpdateTenantMessaging)
	tenant.Get("/hospitality", r.Handler.GetTenantHospitality)
	tenant.Put("/hospitality", r.Handler.UpdateTenantHospitality)

	// E911 Locations
	tenant.Get("/locations", r.Handler.ListLocations)
	tenant.Post("/locations", r.Handler.CreateLocation)
	tenant.Get("/locations/:id", r.Handler.GetLocation)
	tenant.Put("/locations/:id", r.Handler.UpdateLocation)
	tenant.Delete("/locations/:id", r.Handler.DeleteLocation)

	// Extensions
	extensions := tenantScoped.Group("/extensions")
	extensions.Get("/", r.Handler.ListExtensions)
	extensions.Post("/", r.Handler.CreateExtension)
	extensions.Get("/:ext", r.Handler.GetExtension)
	extensions.Put("/:ext", r.Handler.UpdateExtension)
	extensions.Delete("/:ext", r.Handler.DeleteExtension)
	extensions.Get("/:ext/status", r.Handler.GetExtensionStatus)

	// Call Handling Rules for Extension
	extensions.Get("/:ext/call-rules", r.Handler.ListCallHandlingRules)
	extensions.Post("/:ext/call-rules", r.Handler.CreateCallHandlingRule)
	extensions.Put("/:ext/call-rules/:ruleId", r.Handler.UpdateCallHandlingRule)
	extensions.Delete("/:ext/call-rules/:ruleId", r.Handler.DeleteCallHandlingRule)
	extensions.Post("/:ext/call-rules/reorder", r.Handler.ReorderCallHandlingRules)

	// Extension Profiles
	extProfiles := tenantScoped.Group("/extension-profiles")
	extProfiles.Get("/", r.Handler.ListExtensionProfiles)
	extProfiles.Post("/", r.Handler.CreateExtensionProfile)
	extProfiles.Get("/:id", r.Handler.GetExtensionProfile)
	extProfiles.Put("/:id", r.Handler.UpdateExtensionProfile)
	extProfiles.Delete("/:id", r.Handler.DeleteExtensionProfile)

	// Call Handling Rules for Profile
	extProfiles.Get("/:id/call-rules", r.Handler.ListProfileCallHandlingRules)
	extProfiles.Post("/:id/call-rules", r.Handler.CreateProfileCallHandlingRule)
	extProfiles.Put("/:id/call-rules/:ruleId", r.Handler.UpdateProfileCallHandlingRule)
	extProfiles.Delete("/:id/call-rules/:ruleId", r.Handler.DeleteProfileCallHandlingRule)
	extProfiles.Post("/:id/call-rules/reorder", r.Handler.ReorderProfileCallHandlingRules)

	// Devices
	devices := tenantScoped.Group("/devices")
	devices.Get("/", r.Handler.ListDevices)
	devices.Post("/", r.Handler.CreateDevice)
	devices.Get("/:id", r.Handler.GetDevice)
	devices.Put("/:id", r.Handler.UpdateDevice)
	devices.Delete("/:id", r.Handler.DeleteDevice)
	devices.Post("/:id/assign-user", r.Handler.AssignDeviceToUser)
	devices.Post("/:id/assign-profile", r.Handler.AssignDeviceToProfile)
	devices.Post("/:id/reprovision", r.Handler.ReprovisionDevice)
	devices.Put("/:id/lines", r.Handler.UpdateDeviceLines)

	// Client Registrations (apps, web clients, device registrations)
	registrations := tenantScoped.Group("/registrations")
	registrations.Get("/", r.Handler.ListClientRegistrations)
	registrations.Post("/provision", r.Handler.ProvisionClientRegistration)
	registrations.Delete("/:id", r.Handler.DeleteClientRegistration)
	registrations.Get("/unassigned", r.Handler.ListUnassignedRegistrations)
	registrations.Post("/:id/assign", r.Handler.AssignRegistration)
	registrations.Get("/extension/:id", r.Handler.ListExtensionRegistrations)

	// Device Profiles (tenant-level device grouping)
	deviceProfiles := tenantScoped.Group("/device-profiles")
	deviceProfiles.Get("/", r.Handler.ListDeviceProfiles)
	deviceProfiles.Post("/", r.Handler.CreateDeviceProfile)
	deviceProfiles.Get("/:id", r.Handler.GetDeviceProfile)
	deviceProfiles.Put("/:id", r.Handler.UpdateDeviceProfile)
	deviceProfiles.Delete("/:id", r.Handler.DeleteDeviceProfile)

	// Device Templates (tenant-level, includes system templates)
	deviceTemplates := tenantScoped.Group("/device-templates")
	deviceTemplates.Get("/", r.Handler.ListDeviceTemplates)
	deviceTemplates.Post("/", r.Handler.CreateDeviceTemplate)

	// Voicemail
	voicemail := tenantScoped.Group("/voicemail")
	voicemail.Get("/boxes", r.Handler.ListVoicemailBoxes)
	voicemail.Post("/boxes", r.Handler.CreateVoicemailBox)
	voicemail.Get("/boxes/:ext", r.Handler.GetVoicemailBox)
	voicemail.Put("/boxes/:ext", r.Handler.UpdateVoicemailBox)
	voicemail.Delete("/boxes/:ext", r.Handler.DeleteVoicemailBox)
	// Voicemail messages
	voicemail.Get("/boxes/:ext/messages", r.Handler.ListVoicemailMessages)
	voicemail.Get("/messages/:id", r.Handler.GetVoicemailMessage)
	voicemail.Delete("/messages/:id", r.Handler.DeleteVoicemailMessage)
	voicemail.Post("/messages/:id/read", r.Handler.MarkVoicemailRead)
	voicemail.Get("/messages/:id/stream", r.Handler.StreamVoicemailMessage)

	// Recordings
	recordings := tenantScoped.Group("/recordings")
	recordings.Get("/", r.Handler.ListRecordings)
	recordings.Get("/config", r.Handler.GetRecordingConfig)
	recordings.Get("/:id", r.Handler.GetRecording)
	recordings.Delete("/:id", r.Handler.DeleteRecording)
	recordings.Get("/:id/stream", r.Handler.StreamRecording)
	recordings.Get("/:id/download", r.Handler.DownloadRecording)
	recordings.Put("/:id/notes", r.Handler.UpdateRecordingNotes)
	recordings.Get("/:id/transcription", r.Handler.GetRecordingTranscription)

	// IVR Menus
	ivr := tenantScoped.Group("/ivr")
	ivr.Get("/menus", r.Handler.ListIVRMenus)
	ivr.Post("/menus", r.Handler.CreateIVRMenu)
	ivr.Get("/menus/:id", r.Handler.GetIVRMenu)
	ivr.Put("/menus/:id", r.Handler.UpdateIVRMenu)
	ivr.Delete("/menus/:id", r.Handler.DeleteIVRMenu)

	// Queues
	queues := tenantScoped.Group("/queues")
	queues.Get("/", r.Handler.ListQueues)
	queues.Post("/", r.Handler.CreateQueue)
	queues.Get("/:id", r.Handler.GetQueue)
	queues.Put("/:id", r.Handler.UpdateQueue)
	queues.Delete("/:id", r.Handler.DeleteQueue)
	// Queue Agent Management
	queues.Get("/:id/agents", r.Handler.ListQueueAgents)
	queues.Post("/:id/agents", r.Handler.AddQueueAgent)
	queues.Delete("/:id/agents/:agentId", r.Handler.RemoveQueueAgent)
	queues.Post("/:id/agents/:agentId/pause", r.Handler.PauseQueueAgent)
	queues.Post("/:id/agents/:agentId/unpause", r.Handler.UnpauseQueueAgent)

	// Ring Groups
	ringGroups := tenantScoped.Group("/ring-groups")
	ringGroups.Get("/", r.Handler.ListRingGroups)
	ringGroups.Post("/", r.Handler.CreateRingGroup)
	ringGroups.Get("/:id", r.Handler.GetRingGroup)
	ringGroups.Put("/:id", r.Handler.UpdateRingGroup)
	ringGroups.Delete("/:id", r.Handler.DeleteRingGroup)

	// Speed Dials
	speedDials := tenantScoped.Group("/speed-dials")
	speedDials.Get("/", r.Handler.ListSpeedDialGroups)
	speedDials.Post("/", r.Handler.CreateSpeedDialGroup)
	speedDials.Get("/:id", r.Handler.GetSpeedDialGroup)
	speedDials.Put("/:id", r.Handler.UpdateSpeedDialGroup)
	speedDials.Delete("/:id", r.Handler.DeleteSpeedDialGroup)

	// Conferences
	conferences := tenantScoped.Group("/conferences")
	conferences.Get("/", r.Handler.ListConferences)
	conferences.Post("/", r.Handler.CreateConference)
	conferences.Get("/:id", r.Handler.GetConference)
	conferences.Put("/:id", r.Handler.UpdateConference)
	conferences.Delete("/:id", r.Handler.DeleteConference)
	// Conference Stats & Sessions
	conferences.Get("/:id/stats", r.ConferenceHandler.GetConferenceStats)
	conferences.Get("/:id/sessions", r.ConferenceHandler.GetConferenceSessions)
	conferences.Get("/sessions/:sessionId/participants", r.ConferenceHandler.GetSessionParticipants)
	// Live Conference Control
	live := conferences.Group("/live")
	live.Get("/", r.ConferenceHandler.ListLiveConferences)
	live.Get("/:name", r.ConferenceHandler.GetLiveConference)
	live.Post("/:name/mute/:memberId", r.ConferenceHandler.MuteMember)
	live.Post("/:name/unmute/:memberId", r.ConferenceHandler.UnmuteMember)
	live.Post("/:name/deaf/:memberId", r.ConferenceHandler.DeafMember)
	live.Post("/:name/undeaf/:memberId", r.ConferenceHandler.UndeafMember)
	live.Post("/:name/kick/:memberId", r.ConferenceHandler.KickMember)
	live.Post("/:name/lock", r.ConferenceHandler.LockConference)
	live.Post("/:name/unlock", r.ConferenceHandler.UnlockConference)
	live.Post("/:name/record/start", r.ConferenceHandler.StartRecording)
	live.Post("/:name/record/stop", r.ConferenceHandler.StopRecording)
	live.Post("/:name/mute-all", r.ConferenceHandler.MuteAll)
	live.Post("/:name/unmute-all", r.ConferenceHandler.UnmuteAll)
	live.Post("/:name/floor/:memberId", r.ConferenceHandler.SetFloor)

	// Numbers/DIDs
	numbers := tenantScoped.Group("/numbers")
	numbers.Get("/", r.Handler.ListNumbers)
	numbers.Post("/", r.Handler.CreateNumber)
	numbers.Get("/:id", r.Handler.GetNumber)
	numbers.Put("/:id", r.Handler.UpdateNumber)
	numbers.Delete("/:id", r.Handler.DeleteNumber)
	// Tenant-level: assign/unassign number to location (E911)
	numbers.Post("/:id/location", r.Handler.AssignNumberToLocation)
	numbers.Delete("/:id/location", r.Handler.UnassignNumberFromLocation)

	// Routing
	routing := tenantScoped.Group("/routing")
	routing.Get("/inbound", r.Handler.ListInboundRoutes)
	routing.Post("/inbound", r.Handler.CreateInboundRoute)
	routing.Get("/inbound/:id", r.Handler.GetInboundRoute)
	routing.Put("/inbound/:id", r.Handler.UpdateInboundRoute)
	routing.Delete("/inbound/:id", r.Handler.DeleteInboundRoute)
	routing.Post("/inbound/reorder", r.Handler.ReorderInboundRoutes)
	routing.Get("/outbound", r.Handler.ListOutboundRoutes)
	routing.Post("/outbound", r.Handler.CreateOutboundRoute)
	routing.Get("/outbound/:id", r.Handler.GetOutboundRoute)
	routing.Put("/outbound/:id", r.Handler.UpdateOutboundRoute)
	routing.Delete("/outbound/:id", r.Handler.DeleteOutboundRoute)
	routing.Post("/outbound/reorder", r.Handler.ReorderOutboundRoutes)
	routing.Post("/outbound/defaults", r.Handler.CreateDefaultUSCANRoutes)

	// Call Blocks
	routing.Get("/blocks", r.Handler.ListCallBlocks)
	routing.Post("/blocks", r.Handler.CreateCallBlock)
	routing.Put("/blocks/:id", r.Handler.UpdateCallBlock)
	routing.Delete("/blocks/:id", r.Handler.DeleteCallBlock)

	// Debugger
	routing.Get("/debug", r.FSHandler.DebugDialplanTenant)

	// Dial Plans
	dialPlans := tenantScoped.Group("/dial-plans")
	dialPlans.Get("/", r.Handler.ListDialPlans)
	dialPlans.Post("/", r.Handler.CreateDialPlan)
	dialPlans.Get("/:id", r.Handler.GetDialPlan)
	dialPlans.Put("/:id", r.Handler.UpdateDialPlan)
	dialPlans.Delete("/:id", r.Handler.DeleteDialPlan)

	// Audio Library
	audioLibrary := tenantScoped.Group("/audio-library")
	audioLibrary.Get("/", r.Handler.ListMediaFiles)
	audioLibrary.Post("/", r.Handler.UploadMediaFile)
	audioLibrary.Put("/:id", r.Handler.UpdateMediaFile)
	audioLibrary.Delete("/:id", r.Handler.DeleteMediaFile)
	audioLibrary.Get("/:id/stream", r.Handler.StreamMediaFile)

	// Music on Hold
	moh := tenantScoped.Group("/music-on-hold")
	moh.Get("/", r.Handler.ListMOHStreams)
	moh.Post("/", r.Handler.CreateMOHStream)
	moh.Get("/:id", r.Handler.GetMOHStream)
	moh.Put("/:id", r.Handler.UpdateMOHStream)
	moh.Delete("/:id", r.Handler.DeleteMOHStream)

	// Feature Codes
	featureCodes := tenantScoped.Group("/feature-codes")
	featureCodes.Get("/", r.Handler.ListFeatureCodes)
	featureCodes.Get("/system", r.Handler.ListSystemFeatureCodes)
	featureCodes.Post("/", r.Handler.CreateFeatureCode)
	featureCodes.Get("/:id", r.Handler.GetFeatureCode)
	featureCodes.Put("/:id", r.Handler.UpdateFeatureCode)
	featureCodes.Delete("/:id", r.Handler.DeleteFeatureCode)

	// Time Conditions
	timeConditions := tenantScoped.Group("/time-conditions")
	timeConditions.Get("/", r.Handler.ListTimeConditions)
	timeConditions.Post("/", r.Handler.CreateTimeCondition)
	timeConditions.Get("/:id", r.Handler.GetTimeCondition)
	timeConditions.Put("/:id", r.Handler.UpdateTimeCondition)
	timeConditions.Delete("/:id", r.Handler.DeleteTimeCondition)

	// Holiday Lists
	holidays := tenantScoped.Group("/holidays")
	holidays.Get("/", r.Handler.ListHolidayLists)
	holidays.Post("/", r.Handler.CreateHolidayList)
	holidays.Get("/:id", r.Handler.GetHolidayList)
	holidays.Put("/:id", r.Handler.UpdateHolidayList)
	holidays.Delete("/:id", r.Handler.DeleteHolidayList)
	holidays.Post("/:id/sync", r.Handler.SyncHolidayList)

	// Call Flows
	callFlows := tenantScoped.Group("/call-flows")
	callFlows.Get("/", r.Handler.ListCallFlows)
	callFlows.Post("/", r.Handler.CreateCallFlow)
	callFlows.Get("/:id", r.Handler.GetCallFlow)
	callFlows.Put("/:id", r.Handler.UpdateCallFlow)
	callFlows.Delete("/:id", r.Handler.DeleteCallFlow)
	callFlows.Post("/:id/toggle", r.Handler.ToggleCallFlow)

	// CDR / Call Records
	cdr := tenantScoped.Group("/cdr")
	cdr.Get("/", r.Handler.ListCDR)
	cdr.Get("/:id", r.Handler.GetCDR)
	cdr.Get("/export", r.Handler.ExportCDR)

	// Audit Logs
	auditLogs := tenantScoped.Group("/audit-logs")
	auditLogs.Get("/", r.Handler.ListAuditLogs)

	// Dial Code Collision Check
	tenantScoped.Post("/check-dial-code", r.Handler.CheckDialCode)

	// Messaging (SMS/MMS)
	msgRoutes := tenantScoped.Group("/messaging")
	msgRoutes.Get("/conversations", r.Handler.ListConversations)
	msgRoutes.Get("/conversations/:id", r.Handler.GetConversation)
	msgRoutes.Post("/send", r.Handler.SendMessage)

	// SMS Number Management
	msgRoutes.Get("/numbers", r.SMSNumberHandler.ListSMSNumbers)
	msgRoutes.Put("/numbers/:id/sms", r.SMSNumberHandler.ConfigureSMSNumber)
	msgRoutes.Get("/numbers/:id/assignments", r.SMSNumberHandler.ListNumberAssignments)
	msgRoutes.Post("/numbers/:id/assignments", r.SMSNumberHandler.AssignNumber)
	msgRoutes.Delete("/numbers/:id/assignments/:assignId", r.SMSNumberHandler.UnassignNumber)

	// Contacts
	contacts := tenantScoped.Group("/contacts")
	contacts.Get("/", r.Handler.ListContacts)
	contacts.Post("/", r.Handler.CreateContact)
	contacts.Get("/:id", r.Handler.GetContact)
	contacts.Put("/:id", r.Handler.UpdateContact)
	contacts.Delete("/:id", r.Handler.DeleteContact)
	contacts.Post("/:id/sync", r.Handler.SyncContact)
	contacts.Get("/lookup", r.Handler.GetContactByPhone)

	// Chat System
	chat := tenantScoped.Group("/chat")
	chat.Get("/threads", r.Handler.ListChatThreads)
	chat.Post("/threads", r.Handler.CreateChatThread)
	chat.Get("/threads/:id", r.Handler.GetChatThread)
	chat.Post("/threads/:id/messages", r.Handler.SendChatMessage)

	chat.Get("/rooms", r.Handler.ListChatRooms)
	chat.Post("/rooms", r.Handler.CreateChatRoom)
	chat.Post("/rooms/:id/join", r.Handler.JoinChatRoom)

	chat.Get("/queues", r.Handler.ListChatQueues)
	chat.Post("/queues", r.Handler.CreateChatQueue)

	// Paging Groups
	paging := tenantScoped.Group("/page-groups")
	paging.Get("/", r.Handler.ListPageGroups)
	paging.Post("/", r.Handler.CreatePageGroup)
	paging.Get("/:id", r.Handler.GetPageGroup)
	paging.Put("/:id", r.Handler.UpdatePageGroup)
	paging.Delete("/:id", r.Handler.DeletePageGroup)

	// Device Call Control
	deviceControl := tenantScoped.Group("/devices")
	deviceControl.Post("/:mac/hangup", r.Handler.DeviceHangup)
	deviceControl.Post("/:mac/transfer", r.Handler.DeviceTransfer)
	deviceControl.Post("/:mac/hold", r.Handler.DeviceHold)
	deviceControl.Post("/:mac/dial", r.Handler.DeviceDial)
	deviceControl.Get("/:mac/call-status", r.Handler.DeviceCallStatus)

	// Provisioning Templates (tenant-level)
	provisioning := tenantScoped.Group("/provisioning-templates")
	provisioning.Get("/", r.Handler.ListProvisioningTemplates)
	provisioning.Post("/", r.Handler.CreateProvisioningTemplate)
	provisioning.Get("/:id", r.Handler.GetProvisioningTemplate)
	provisioning.Put("/:id", r.Handler.UpdateProvisioningTemplate)
	provisioning.Delete("/:id", r.Handler.DeleteProvisioningTemplate)

	// Tenant Media (Sounds & Music Overrides)
	media := tenantScoped.Group("/media")
	media.Get("/sounds", r.Handler.ListTenantSounds)
	media.Post("/sounds", r.Handler.UploadTenantSound)
	media.Delete("/sounds", r.Handler.DeleteTenantSound)

	media.Get("/music", r.Handler.ListTenantMusic)
	media.Post("/music", r.Handler.UploadTenantMusic)
	media.Delete("/music", r.Handler.DeleteTenantMusic)

	// Fax
	faxRoutes := tenantScoped.Group("/fax")
	// Fax Boxes
	faxRoutes.Get("/boxes", r.FaxHandler.ListFaxBoxes)
	faxRoutes.Post("/boxes", r.FaxHandler.CreateFaxBox)
	faxRoutes.Get("/boxes/:boxId", r.FaxHandler.GetFaxBox)
	faxRoutes.Put("/boxes/:boxId", r.FaxHandler.UpdateFaxBox)
	faxRoutes.Delete("/boxes/:boxId", r.FaxHandler.DeleteFaxBox)

	// Fax Jobs
	faxRoutes.Get("/jobs", r.FaxHandler.ListFaxJobs)
	faxRoutes.Get("/jobs/:jobId", r.FaxHandler.GetFaxJob)
	faxRoutes.Delete("/jobs/:jobId", r.FaxHandler.DeleteFaxJob)
	faxRoutes.Get("/jobs/:jobId/download", r.FaxHandler.DownloadFax)
	faxRoutes.Post("/jobs/:jobId/retry", r.FaxHandler.RetryFax)

	// Fax Actions
	faxRoutes.Post("/send", r.FaxHandler.SendFax)
	faxRoutes.Get("/active", r.FaxHandler.GetActiveFaxes)
	faxRoutes.Get("/stats", r.FaxHandler.GetFaxStats)

	// Fax Endpoints
	faxRoutes.Get("/endpoints", r.FaxHandler.ListFaxEndpoints)
	faxRoutes.Post("/endpoints", r.FaxHandler.CreateFaxEndpoint)
	faxRoutes.Put("/endpoints/:epId", r.FaxHandler.UpdateFaxEndpoint)
	faxRoutes.Delete("/endpoints/:epId", r.FaxHandler.DeleteFaxEndpoint)

	// Reports & Analytics
	reports := tenantScoped.Group("/reports")
	reports.Get("/call-volume", r.Handler.GetCallVolumeReport)
	reports.Get("/agent-performance", r.Handler.GetAgentPerformanceReport)
	reports.Get("/queue-stats", r.Handler.GetQueueStatsReport)
	reports.Get("/extension-usage", r.Handler.GetExtensionUsageReport)
	reports.Get("/kpi", r.Handler.GetKPIReport)
	reports.Get("/number-usage", r.Handler.GetNumberUsageReport)
	reports.Get("/export", r.Handler.ExportReport)

	// Hospitality (Hotel room management)
	hospitality := tenantScoped.Group("/hospitality")
	hospitality.Get("/rooms", r.Handler.ListRooms)
	hospitality.Post("/rooms", r.Handler.CreateRoom)
	hospitality.Get("/rooms/:id", r.Handler.GetRoom)
	hospitality.Put("/rooms/:id", r.Handler.UpdateRoom)
	hospitality.Delete("/rooms/:id", r.Handler.DeleteRoom)
	hospitality.Post("/rooms/:id/checkin", r.Handler.CheckInGuest)
	hospitality.Post("/rooms/:id/checkout", r.Handler.CheckOutGuest)
	hospitality.Post("/rooms/:id/wakeup", r.Handler.ScheduleWakeupCall)

	// Call Broadcast Campaigns
	broadcast := tenantScoped.Group("/broadcast")
	broadcast.Get("/", r.Handler.ListBroadcasts)
	broadcast.Post("/", r.Handler.CreateBroadcast)
	broadcast.Get("/:id", r.Handler.GetBroadcast)
	broadcast.Put("/:id", r.Handler.UpdateBroadcast)
	broadcast.Delete("/:id", r.Handler.DeleteBroadcast)
	broadcast.Post("/:id/start", r.Handler.StartBroadcast)
	broadcast.Post("/:id/stop", r.Handler.StopBroadcast)
	broadcast.Get("/:id/stats", r.Handler.GetBroadcastStats)

	// Operator Panel
	tenantScoped.Get("/operator-panel", r.Handler.GetOperatorPanelData)

	// Live Operations
	liveOps := tenantScoped.Group("/live")
	liveOps.Post("/recording/start", r.Handler.StartCallRecording)
	liveOps.Post("/recording/stop", r.Handler.StopCallRecording)
	liveOps.Get("/calls", r.Handler.GetActiveCallsData)
	liveOps.Get("/queue-stats", r.Handler.GetLiveQueueStats)
	liveOps.Post("/wakeup/schedule", r.Handler.ScheduleWakeupESL)
	liveOps.Get("/registrations", r.Handler.GetDeviceRegistrations)

	// Tenant admin routes
	tenantAdmin := protected.Group("")
	tenantAdmin.Use(r.Auth.RequireTenantAdmin())
	tenantAdmin.Use(r.Tenant.RequireTenant())

	// Tenant users management
	users := tenantAdmin.Group("/users")
	users.Get("/", r.Handler.ListUsers)
	users.Post("/", r.Handler.CreateUser)
	users.Get("/:id", r.Handler.GetUser)
	users.Put("/:id", r.Handler.UpdateUser)
	users.Delete("/:id", r.Handler.DeleteUser)

	// System admin routes
	system := protected.Group("/system")
	system.Use(r.Auth.RequireSystemAdmin())

	// Tenants management
	tenants := system.Group("/tenants")
	tenants.Get("/", r.Handler.ListTenants)
	tenants.Post("/", r.Handler.CreateTenant)
	tenants.Get("/:id", r.Handler.GetTenant)
	tenants.Put("/:id", r.Handler.UpdateTenant)
	tenants.Delete("/:id", r.Handler.DeleteTenant)

	// System Numbers (centralized pool)
	sysNumbers := system.Group("/numbers")
	sysNumbers.Get("/", r.Handler.ListSystemNumbers)
	sysNumbers.Post("/", r.Handler.CreateSystemNumber)
	sysNumbers.Get("/:id", r.Handler.GetSystemNumber)
	sysNumbers.Put("/:id", r.Handler.UpdateSystemNumber)
	sysNumbers.Delete("/:id", r.Handler.DeleteSystemNumber)
	sysNumbers.Post("/:id/assign", r.Handler.AssignNumberToTenant)
	sysNumbers.Post("/:id/unassign", r.Handler.UnassignNumber)

	// Number Groups (outbound routing groups)
	numberGroups := system.Group("/number-groups")
	numberGroups.Get("/", r.Handler.ListNumberGroups)
	numberGroups.Post("/", r.Handler.CreateNumberGroup)
	numberGroups.Get("/:id", r.Handler.GetNumberGroup)
	numberGroups.Put("/:id", r.Handler.UpdateNumberGroup)
	numberGroups.Delete("/:id", r.Handler.DeleteNumberGroup)
	numberGroups.Post("/:id/reorder-gateways", r.Handler.ReorderGroupGateways)

	// Tenant Profiles
	profiles := system.Group("/tenant-profiles")
	profiles.Get("/", r.Handler.ListTenantProfiles)
	profiles.Post("/", r.Handler.CreateTenantProfile)
	profiles.Get("/:id", r.Handler.GetTenantProfile)
	profiles.Put("/:id", r.Handler.UpdateTenantProfile)
	profiles.Delete("/:id", r.Handler.DeleteTenantProfile)

	// Gateways
	gateways := system.Group("/gateways")
	gateways.Get("/", r.Handler.ListGateways)
	gateways.Post("/", r.Handler.CreateGateway)
	gateways.Get("/status", r.Handler.GetGatewayStatus)  // Must be before /:id
	gateways.Post("/reorder", r.Handler.ReorderGateways) // Must be before /:id
	gateways.Get("/:id", r.Handler.GetGateway)
	gateways.Put("/:id", r.Handler.UpdateGateway)
	gateways.Delete("/:id", r.Handler.DeleteGateway)

	// Bridges
	bridges := system.Group("/bridges")
	bridges.Get("/", r.Handler.ListBridges)
	bridges.Post("/", r.Handler.CreateBridge)
	bridges.Get("/:id", r.Handler.GetBridge)
	bridges.Put("/:id", r.Handler.UpdateBridge)
	bridges.Delete("/:id", r.Handler.DeleteBridge)

	// SIP Profiles
	sipProfiles := system.Group("/sip-profiles")
	sipProfiles.Get("/", r.Handler.ListSIPProfiles)
	sipProfiles.Post("/", r.Handler.CreateSIPProfile)
	sipProfiles.Post("/sync", r.Handler.SyncSIPProfiles) // Import from disk
	sipProfiles.Get("/:id", r.Handler.GetSIPProfile)
	sipProfiles.Put("/:id", r.Handler.UpdateSIPProfile)
	sipProfiles.Delete("/:id", r.Handler.DeleteSIPProfile)

	// Sofia Control (live FreeSWITCH commands)
	sofia := system.Group("/sofia")
	sofia.Get("/status", r.Handler.GetSofiaStatus)
	sofia.Get("/profiles/:name/status", r.Handler.GetSofiaProfileStatus)
	sofia.Get("/profiles/:name/registrations", r.Handler.GetSofiaProfileRegistrations)
	sofia.Get("/profiles/:name/gateways", r.Handler.GetSofiaGatewayStatus)
	sofia.Post("/profiles/:name/restart", r.Handler.RestartSofiaProfile)
	sofia.Post("/profiles/:name/start", r.Handler.StartSofiaProfile)
	sofia.Post("/profiles/:name/stop", r.Handler.StopSofiaProfile)
	sofia.Post("/reload-xml", r.Handler.ReloadSofiaXML)

	// System settings
	settings := system.Group("/settings")
	settings.Get("/", r.Handler.GetSystemSettings)
	settings.Put("/", r.Handler.UpdateSystemSettings)

	// System logs
	logs := system.Group("/logs")
	logs.Get("/", r.Handler.GetSystemLogs)

	// Messaging providers (system-level)
	msgProviders := system.Group("/messaging-providers")
	msgProviders.Get("/", r.Handler.ListMessagingProviders)
	msgProviders.Post("/", r.Handler.CreateMessagingProvider)
	msgProviders.Get("/:id", r.Handler.GetMessagingProvider)
	msgProviders.Put("/:id", r.Handler.UpdateMessagingProvider)
	msgProviders.Delete("/:id", r.Handler.DeleteMessagingProvider)

	// Messaging numbers (per-provider phone numbers)
	msgNumbers := system.Group("/messaging-numbers")
	msgNumbers.Get("/", r.Handler.ListMessagingNumbers)
	msgNumbers.Post("/", r.Handler.CreateMessagingNumber)
	msgNumbers.Put("/:id", r.Handler.UpdateMessagingNumber)
	msgNumbers.Delete("/:id", r.Handler.DeleteMessagingNumber)

	// Global dial plans
	dialplans := system.Group("/dialplans")
	dialplans.Get("/", r.Handler.ListGlobalDialplans)
	dialplans.Post("/", r.Handler.CreateGlobalDialplan)
	dialplans.Get("/:id", r.Handler.GetGlobalDialplan)
	dialplans.Put("/:id", r.Handler.UpdateGlobalDialplan)
	dialplans.Delete("/:id", r.Handler.DeleteGlobalDialplan)

	// Access Control Lists (ACLs)
	acls := system.Group("/acls")
	acls.Get("/", r.Handler.ListACLs)
	acls.Post("/", r.Handler.CreateACL)
	acls.Get("/:id", r.Handler.GetACL)
	acls.Put("/:id", r.Handler.UpdateACL)
	acls.Delete("/:id", r.Handler.DeleteACL)
	// ACL nodes (entries)
	acls.Post("/:id/nodes", r.Handler.CreateACLNode)
	acls.Put("/:id/nodes/:nodeId", r.Handler.UpdateACLNode)
	acls.Delete("/:id/nodes/:nodeId", r.Handler.DeleteACLNode)

	// System Media
	sysMedia := system.Group("/media")
	sysMedia.Get("/sounds", r.Handler.ListSystemSounds)
	sysMedia.Post("/sounds", r.Handler.UploadSystemSound)
	sysMedia.Get("/sounds/stream", r.Handler.StreamSystemSound)
	sysMedia.Get("/music", r.Handler.ListSystemMusic)
	sysMedia.Post("/music", r.Handler.UploadSystemMusic)
	sysMedia.Get("/music/stream", r.Handler.StreamSystemMusic)

	// System status
	system.Get("/status", r.Handler.GetSystemStatus)
	system.Get("/stats", r.Handler.GetSystemStats)

	// Security - Banned IPs
	security := system.Group("/security")
	security.Get("/banned-ips", r.Handler.ListBannedIPs)
	security.Post("/banned-ips", r.Handler.ReportBannedIP)
	security.Delete("/banned-ips/:ip", r.Handler.UnbanIP)

	// Device Templates (system-level master templates)
	sysDeviceTemplates := system.Group("/device-templates")
	sysDeviceTemplates.Get("/", r.Handler.ListSystemDeviceTemplates)
	sysDeviceTemplates.Post("/", r.Handler.CreateSystemDeviceTemplate)
	sysDeviceTemplates.Get("/:id", r.Handler.GetSystemDeviceTemplate)
	sysDeviceTemplates.Put("/:id", r.Handler.UpdateSystemDeviceTemplate)
	sysDeviceTemplates.Delete("/:id", r.Handler.DeleteSystemDeviceTemplate)

	// Device Manufacturers (configurable groupings)
	manufacturers := system.Group("/device-manufacturers")
	manufacturers.Get("/", r.Handler.ListDeviceManufacturers)
	manufacturers.Post("/", r.Handler.CreateDeviceManufacturer)
	manufacturers.Put("/:id", r.Handler.UpdateDeviceManufacturer)
	manufacturers.Delete("/:id", r.Handler.DeleteDeviceManufacturer)

	// Firmware Management
	firmware := system.Group("/firmware")
	firmware.Get("/", r.Handler.ListFirmware)
	firmware.Post("/", r.Handler.CreateFirmware)
	firmware.Get("/:id", r.Handler.GetFirmware)
	firmware.Put("/:id", r.Handler.UpdateFirmware)
	firmware.Delete("/:id", r.Handler.DeleteFirmware)
	firmware.Post("/:id/upload", r.Handler.UploadFirmware)
	firmware.Post("/:id/set-default", r.Handler.SetDefaultFirmware)

	// Config Inspector (System)
	system.Get("/xml/debug", r.FSHandler.DebugXML)
	system.Get("/config/files", r.FSHandler.ListConfigDirectory)
	system.Get("/config/file", r.FSHandler.ReadConfigFile)

	// User portal routes
	user := protected.Group("/user")
	user.Get("/devices", r.Handler.GetUserDevices)
	user.Get("/call-history", r.Handler.GetUserCallHistory)
	user.Get("/voicemail", r.Handler.GetUserVoicemail)
	user.Get("/settings", r.Handler.GetUserSettings)
	user.Put("/settings", r.Handler.UpdateUserSettings)
	user.Get("/contacts", r.Handler.GetUserContacts)
	user.Post("/contacts", r.Handler.CreateUserContact)

	// Extension portal routes (for extension panel / web client)
	extPortal := protected.Group("/extension/portal")
	extPortal.Use(r.Tenant.RequireTenant())
	extPortal.Get("/devices", r.Handler.GetExtensionDevices)
	extPortal.Get("/call-history", r.Handler.GetExtensionCallHistory)
	extPortal.Get("/voicemail", r.Handler.GetExtensionVoicemail)
	extPortal.Get("/settings", r.Handler.GetExtensionSettings)
	extPortal.Put("/settings", r.Handler.UpdateExtensionSettings)
	extPortal.Put("/password", r.Handler.ChangeExtensionPassword)
	extPortal.Get("/contacts", r.Handler.GetExtensionContacts)
	extPortal.Post("/contacts", r.Handler.CreateExtensionContact)

	// FreeSWITCH XML CURL endpoints (inside /api for consistency with Caddy routing)
	fs := r.App.Group("/api/freeswitch")
	fs.Use(freeswitch.FreeSwitchAuthMiddleware(r.Config))
	// Individual section handlers - FreeSWITCH mod_xml_curl calls these
	fs.Post("/directory", r.FSHandler.HandleXMLCurl)     // sip_auth, registration
	fs.Post("/dialplan", r.FSHandler.HandleXMLCurl)      // call routing
	fs.Post("/configuration", r.FSHandler.HandleXMLCurl) // sofia, event_socket, etc

	// Legacy combined handler (for backwards compatibility)
	fs.Post("/xmlapi", r.FSHandler.HandleXMLCurl)

	// CDR handler - receives POST from mod_xml_cdr
	// Config in FreeSWITCH: <param name="url" value="http://127.0.0.1:8080/api/freeswitch/cdr"/>
	fs.Post("/cdr", r.FSHandler.HandleXMLCDR)

	// Cache management endpoints
	fs.Get("/cache/flush", r.FSHandler.FlushCache)
	fs.Get("/cache/stats", r.FSHandler.CacheStats)

	// WebSocket endpoint for real-time events
	r.App.Get("/api/ws", r.Handler.HandleWebSocket)

	// Telnyx Webhooks (public — verified via webhook signature, no JWT)
	webhooks := r.App.Group("/api/webhooks")
	webhooks.Post("/telnyx/inbound", r.WebhookHandler.TelnyxInbound)
	webhooks.Post("/telnyx/status", r.WebhookHandler.TelnyxStatus)

	// Device Provisioning endpoint (public - devices authenticate via MAC)
	provisioningPub := r.App.Group("/provisioning")
	provisioningPub.Get("/:mac/:filename", r.Handler.ServeProvisioningConfig)

	log.Info("All routes loaded successfully")
}

// Listen starts the HTTP server
func (r *Router) Listen(addr string) {
	log.Infof("Starting server on %s", addr)
	if err := r.App.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
