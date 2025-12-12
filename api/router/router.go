package router

import (
	"callsign/config"
	"callsign/handlers"
	"callsign/handlers/freeswitch"
	"callsign/middleware"

	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Router manages the Iris application and routes
type Router struct {
	App       *iris.Application
	DB        *gorm.DB
	Config    *config.Config
	Auth      *middleware.AuthMiddleware
	Tenant    *middleware.TenantMiddleware
	Handler   *handlers.Handler
	FSHandler *freeswitch.FSHandler
}

// NewRouter creates a new Router instance
func NewRouter(db *gorm.DB, cfg *config.Config) *Router {
	return &Router{
		App:       iris.New(),
		DB:        db,
		Config:    cfg,
		Auth:      middleware.NewAuthMiddleware(cfg, db),
		Tenant:    middleware.NewTenantMiddleware(db),
		Handler:   handlers.NewHandler(db, cfg),
		FSHandler: freeswitch.NewFSHandler(db, cfg),
	}
}

// internalKeyAuth validates the X-Internal-Key header for internal service access
func (r *Router) internalKeyAuth() iris.Handler {
	return func(ctx iris.Context) {
		key := ctx.GetHeader("X-Internal-Key")

		// Check against configured internal key
		configuredKey := r.Config.InternalAPIKey
		if configuredKey == "" {
			configuredKey = "callsign-internal-key" // Default for development
		}

		if key == "" || key != configuredKey {
			ctx.StatusCode(iris.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Invalid or missing internal API key"})
			return
		}

		ctx.Next()
	}
}

// Init sets up all routes and middleware
func (r *Router) Init() {
	// Global middleware
	r.App.Use(middleware.Recovery())
	r.App.Use(middleware.RequestLogger())
	r.App.Use(middleware.CORS(r.Config))

	// API base party
	api := r.App.Party("/api")
	{
		// Health check (public)
		api.Get("/health", r.Handler.Health)

		// Public authentication routes
		auth := api.Party("/auth")
		{
			auth.Post("/login", r.Handler.Login)
			auth.Post("/admin/login", r.Handler.AdminLogin)
			auth.Post("/register", r.Handler.Register) // If self-registration is enabled
			auth.Post("/password/reset", r.Handler.RequestPasswordReset)
		}

		// Public WebSocket routes (auth handled inside handler via first message)
		api.Get("/system/console", r.Handler.FreeSwitchConsole)

		// Device provisioning (public, authenticated via tenant secret in URL)
		// URL format: /provision/{tenant_uuid}/{secret}/{mac}.cfg
		provision := api.Party("/provision")
		{
			provision.Get("/{tenant}/{secret}/{mac}", r.Handler.GetDeviceConfigSecure)
		}

		// Internal routes (authenticated via X-Internal-Key header)
		// These are for internal services like fail2ban
		internal := api.Party("/internal")
		internal.Use(r.internalKeyAuth())
		{
			internal.Post("/fail2ban/report", r.Handler.ReportBannedIP)
		}

		// Protected routes (require authentication)
		protected := api.Party("")
		protected.Use(r.Auth.RequireAuth())
		{
			// Auth routes (authenticated)
			protectedAuth := protected.Party("/auth")
			{
				protectedAuth.Get("/me", r.Handler.GetProfile)
				protectedAuth.Put("/password", r.Handler.ChangePassword)
				protectedAuth.Post("/logout", r.Handler.Logout)
				protectedAuth.Post("/refresh", r.Handler.RefreshToken)
			}

			// Tenant-scoped routes
			tenantScoped := protected.Party("")
			tenantScoped.Use(r.Tenant.RequireTenant())
			{
				// Extensions
				extensions := tenantScoped.Party("/extensions")
				{
					extensions.Get("/", r.Handler.ListExtensions)
					extensions.Post("/", r.Handler.CreateExtension)
					extensions.Get("/{ext}", r.Handler.GetExtension)
					extensions.Put("/{ext}", r.Handler.UpdateExtension)
					extensions.Delete("/{ext}", r.Handler.DeleteExtension)
					extensions.Get("/{ext}/status", r.Handler.GetExtensionStatus)
				}

				// Extension Profiles
				extProfiles := tenantScoped.Party("/extension-profiles")
				{
					extProfiles.Get("/", r.Handler.ListExtensionProfiles)
					extProfiles.Post("/", r.Handler.CreateExtensionProfile)
					extProfiles.Get("/{id}", r.Handler.GetExtensionProfile)
					extProfiles.Put("/{id}", r.Handler.UpdateExtensionProfile)
					extProfiles.Delete("/{id}", r.Handler.DeleteExtensionProfile)
				}

				// Devices
				devices := tenantScoped.Party("/devices")
				{
					devices.Get("/", r.Handler.ListDevices)
					devices.Post("/", r.Handler.CreateDevice)
					devices.Get("/{id}", r.Handler.GetDevice)
					devices.Put("/{id}", r.Handler.UpdateDevice)
					devices.Delete("/{id}", r.Handler.DeleteDevice)
					devices.Post("/{id}/assign-user", r.Handler.AssignDeviceToUser)
					devices.Post("/{id}/assign-profile", r.Handler.AssignDeviceToProfile)
					devices.Post("/{id}/reprovision", r.Handler.ReprovisionDevice)
					devices.Put("/{id}/lines", r.Handler.UpdateDeviceLines)
				}

				// Device Profiles (tenant-level device grouping)
				deviceProfiles := tenantScoped.Party("/device-profiles")
				{
					deviceProfiles.Get("/", r.Handler.ListDeviceProfiles)
					deviceProfiles.Post("/", r.Handler.CreateDeviceProfile)
					deviceProfiles.Get("/{id}", r.Handler.GetDeviceProfile)
					deviceProfiles.Put("/{id}", r.Handler.UpdateDeviceProfile)
					deviceProfiles.Delete("/{id}", r.Handler.DeleteDeviceProfile)
				}

				// Device Templates (tenant-level, includes system templates)
				deviceTemplates := tenantScoped.Party("/device-templates")
				{
					deviceTemplates.Get("/", r.Handler.ListDeviceTemplates)
					deviceTemplates.Post("/", r.Handler.CreateDeviceTemplate)
				}

				// Voicemail
				voicemail := tenantScoped.Party("/voicemail")
				{
					voicemail.Get("/boxes", r.Handler.ListVoicemailBoxes)
					voicemail.Post("/boxes", r.Handler.CreateVoicemailBox)
					voicemail.Get("/boxes/{ext}", r.Handler.GetVoicemailBox)
					voicemail.Put("/boxes/{ext}", r.Handler.UpdateVoicemailBox)
					voicemail.Delete("/boxes/{ext}", r.Handler.DeleteVoicemailBox)
				}

				// Recordings
				recordings := tenantScoped.Party("/recordings")
				{
					recordings.Get("/", r.Handler.ListRecordings)
					recordings.Get("/{id}", r.Handler.GetRecording)
					recordings.Delete("/{id}", r.Handler.DeleteRecording)
				}

				// IVR Menus
				ivr := tenantScoped.Party("/ivr")
				{
					ivr.Get("/menus", r.Handler.ListIVRMenus)
					ivr.Post("/menus", r.Handler.CreateIVRMenu)
					ivr.Get("/menus/{id}", r.Handler.GetIVRMenu)
					ivr.Put("/menus/{id}", r.Handler.UpdateIVRMenu)
					ivr.Delete("/menus/{id}", r.Handler.DeleteIVRMenu)
				}

				// Queues
				queues := tenantScoped.Party("/queues")
				{
					queues.Get("/", r.Handler.ListQueues)
					queues.Post("/", r.Handler.CreateQueue)
					queues.Get("/{id}", r.Handler.GetQueue)
					queues.Put("/{id}", r.Handler.UpdateQueue)
					queues.Delete("/{id}", r.Handler.DeleteQueue)
				}

				// Ring Groups
				ringGroups := tenantScoped.Party("/ring-groups")
				{
					ringGroups.Get("/", r.Handler.ListRingGroups)
					ringGroups.Post("/", r.Handler.CreateRingGroup)
					ringGroups.Get("/{id}", r.Handler.GetRingGroup)
					ringGroups.Put("/{id}", r.Handler.UpdateRingGroup)
					ringGroups.Delete("/{id}", r.Handler.DeleteRingGroup)
				}

				// Conferences
				conferences := tenantScoped.Party("/conferences")
				{
					conferences.Get("/", r.Handler.ListConferences)
					conferences.Post("/", r.Handler.CreateConference)
					conferences.Get("/{id}", r.Handler.GetConference)
					conferences.Put("/{id}", r.Handler.UpdateConference)
					conferences.Delete("/{id}", r.Handler.DeleteConference)
				}

				// Numbers/DIDs
				numbers := tenantScoped.Party("/numbers")
				{
					numbers.Get("/", r.Handler.ListNumbers)
					numbers.Post("/", r.Handler.CreateNumber)
					numbers.Get("/{id}", r.Handler.GetNumber)
					numbers.Put("/{id}", r.Handler.UpdateNumber)
					numbers.Delete("/{id}", r.Handler.DeleteNumber)
				}

				// Routing
				routing := tenantScoped.Party("/routing")
				{
					routing.Get("/inbound", r.Handler.ListInboundRoutes)
					routing.Post("/inbound", r.Handler.CreateInboundRoute)
					routing.Get("/outbound", r.Handler.ListOutboundRoutes)
					routing.Post("/outbound", r.Handler.CreateOutboundRoute)
					routing.Post("/outbound/defaults", r.Handler.CreateDefaultUSCANRoutes)

					// Call Blocks
					routing.Get("/blocks", r.Handler.ListCallBlocks)
					routing.Post("/blocks", r.Handler.CreateCallBlock)
					routing.Put("/blocks/{id}", r.Handler.UpdateCallBlock)
					routing.Delete("/blocks/{id}", r.Handler.DeleteCallBlock)
				}

				// Dial Plans
				dialPlans := tenantScoped.Party("/dial-plans")
				{
					dialPlans.Get("/", r.Handler.ListDialPlans)
					dialPlans.Post("/", r.Handler.CreateDialPlan)
					dialPlans.Get("/{id}", r.Handler.GetDialPlan)
					dialPlans.Put("/{id}", r.Handler.UpdateDialPlan)
					dialPlans.Delete("/{id}", r.Handler.DeleteDialPlan)
				}

				// Audio Library
				audioLibrary := tenantScoped.Party("/audio-library")
				{
					audioLibrary.Get("/", r.Handler.ListMediaFiles)
					audioLibrary.Post("/", r.Handler.UploadMediaFile)
					audioLibrary.Put("/{id}", r.Handler.UpdateMediaFile)
					audioLibrary.Delete("/{id}", r.Handler.DeleteMediaFile)
					audioLibrary.Get("/{id}/stream", r.Handler.StreamMediaFile)
				}

				// Music on Hold
				moh := tenantScoped.Party("/music-on-hold")
				{
					moh.Get("/", r.Handler.ListMOHStreams)
					moh.Post("/", r.Handler.CreateMOHStream)
					moh.Get("/{id}", r.Handler.GetMOHStream)
					moh.Put("/{id}", r.Handler.UpdateMOHStream)
					moh.Delete("/{id}", r.Handler.DeleteMOHStream)
				}

				// Feature Codes
				featureCodes := tenantScoped.Party("/feature-codes")
				{
					featureCodes.Get("/", r.Handler.ListFeatureCodes)
					featureCodes.Get("/system", r.Handler.ListSystemFeatureCodes)
					featureCodes.Post("/", r.Handler.CreateFeatureCode)
					featureCodes.Get("/{id}", r.Handler.GetFeatureCode)
					featureCodes.Put("/{id}", r.Handler.UpdateFeatureCode)
					featureCodes.Delete("/{id}", r.Handler.DeleteFeatureCode)
				}

				// Time Conditions
				timeConditions := tenantScoped.Party("/time-conditions")
				{
					timeConditions.Get("/", r.Handler.ListTimeConditions)
					timeConditions.Post("/", r.Handler.CreateTimeCondition)
					timeConditions.Get("/{id}", r.Handler.GetTimeCondition)
					timeConditions.Put("/{id}", r.Handler.UpdateTimeCondition)
					timeConditions.Delete("/{id}", r.Handler.DeleteTimeCondition)
				}

				// Holiday Lists
				holidays := tenantScoped.Party("/holidays")
				{
					holidays.Get("/", r.Handler.ListHolidayLists)
					holidays.Post("/", r.Handler.CreateHolidayList)
					holidays.Get("/{id}", r.Handler.GetHolidayList)
					holidays.Put("/{id}", r.Handler.UpdateHolidayList)
					holidays.Delete("/{id}", r.Handler.DeleteHolidayList)
					holidays.Post("/{id}/sync", r.Handler.SyncHolidayList)
				}

				// Call Flows
				callFlows := tenantScoped.Party("/call-flows")
				{
					callFlows.Get("/", r.Handler.ListCallFlows)
					callFlows.Post("/", r.Handler.CreateCallFlow)
					callFlows.Get("/{id}", r.Handler.GetCallFlow)
					callFlows.Put("/{id}", r.Handler.UpdateCallFlow)
					callFlows.Delete("/{id}", r.Handler.DeleteCallFlow)
					callFlows.Post("/{id}/toggle", r.Handler.ToggleCallFlow)
				}

				// CDR / Call Records
				cdr := tenantScoped.Party("/cdr")
				{
					cdr.Get("/", r.Handler.ListCDR)
					cdr.Get("/{id}", r.Handler.GetCDR)
					cdr.Get("/export", r.Handler.ExportCDR)
				}

				// Audit Logs
				auditLogs := tenantScoped.Party("/audit-logs")
				{
					auditLogs.Get("/", r.Handler.ListAuditLogs)
				}

				// Messaging (SMS/MMS)
				messaging := tenantScoped.Party("/messaging")
				{
					messaging.Get("/conversations", r.Handler.ListConversations)
					messaging.Get("/conversations/{id}", r.Handler.GetConversation)
					messaging.Post("/send", r.Handler.SendMessage)
				}

				// Contacts
				contacts := tenantScoped.Party("/contacts")
				{
					contacts.Get("/", r.Handler.ListContacts)
					contacts.Post("/", r.Handler.CreateContact)
					contacts.Get("/{id}", r.Handler.GetContact)
					contacts.Put("/{id}", r.Handler.UpdateContact)
					contacts.Delete("/{id}", r.Handler.DeleteContact)
					contacts.Post("/{id}/sync", r.Handler.SyncContact)
					contacts.Get("/lookup", r.Handler.GetContactByPhone)
				}

				// Chat System
				chat := tenantScoped.Party("/chat")
				{
					chat.Get("/threads", r.Handler.ListChatThreads)
					chat.Post("/threads", r.Handler.CreateChatThread)
					chat.Get("/threads/{id}", r.Handler.GetChatThread)
					chat.Post("/threads/{id}/messages", r.Handler.SendChatMessage)

					chat.Get("/rooms", r.Handler.ListChatRooms)
					chat.Post("/rooms", r.Handler.CreateChatRoom)
					chat.Post("/rooms/{id}/join", r.Handler.JoinChatRoom)

					chat.Get("/queues", r.Handler.ListChatQueues)
					chat.Post("/queues", r.Handler.CreateChatQueue)
				}

				// Paging Groups
				paging := tenantScoped.Party("/page-groups")
				{
					paging.Get("/", r.Handler.ListPageGroups)
					paging.Post("/", r.Handler.CreatePageGroup)
					paging.Get("/{id}", r.Handler.GetPageGroup)
					paging.Put("/{id}", r.Handler.UpdatePageGroup)
					paging.Delete("/{id}", r.Handler.DeletePageGroup)
				}

				// Device Call Control
				deviceControl := tenantScoped.Party("/devices")
				{
					deviceControl.Post("/{mac}/hangup", r.Handler.DeviceHangup)
					deviceControl.Post("/{mac}/transfer", r.Handler.DeviceTransfer)
					deviceControl.Post("/{mac}/hold", r.Handler.DeviceHold)
					deviceControl.Post("/{mac}/dial", r.Handler.DeviceDial)
					deviceControl.Get("/{mac}/call-status", r.Handler.DeviceCallStatus)
				}

				// Provisioning Templates (tenant-level)
				provisioning := tenantScoped.Party("/provisioning-templates")
				{
					provisioning.Get("/", r.Handler.ListProvisioningTemplates)
					provisioning.Post("/", r.Handler.CreateProvisioningTemplate)
					provisioning.Get("/{id}", r.Handler.GetProvisioningTemplate)
					provisioning.Put("/{id}", r.Handler.UpdateProvisioningTemplate)
					provisioning.Delete("/{id}", r.Handler.DeleteProvisioningTemplate)
				}

				// Tenant Media (Sounds & Music Overrides)
				media := tenantScoped.Party("/media")
				{
					media.Get("/sounds", r.Handler.ListTenantSounds)
					media.Post("/sounds", r.Handler.UploadTenantSound)
					media.Delete("/sounds", r.Handler.DeleteTenantSound)

					media.Get("/music", r.Handler.ListTenantMusic)
					media.Post("/music", r.Handler.UploadTenantMusic)
					media.Delete("/music", r.Handler.DeleteTenantMusic)
				}
			}

			// Tenant admin routes
			tenantAdmin := protected.Party("")
			tenantAdmin.Use(r.Auth.RequireTenantAdmin())
			tenantAdmin.Use(r.Tenant.RequireTenant())
			{
				// Tenant users management
				users := tenantAdmin.Party("/users")
				{
					users.Get("/", r.Handler.ListUsers)
					users.Post("/", r.Handler.CreateUser)
					users.Get("/{id}", r.Handler.GetUser)
					users.Put("/{id}", r.Handler.UpdateUser)
					users.Delete("/{id}", r.Handler.DeleteUser)
				}
			}

			// System admin routes
			system := protected.Party("/system")
			system.Use(r.Auth.RequireSystemAdmin())
			{
				// Tenants management
				tenants := system.Party("/tenants")
				{
					tenants.Get("/", r.Handler.ListTenants)
					tenants.Post("/", r.Handler.CreateTenant)
					tenants.Get("/{id}", r.Handler.GetTenant)
					tenants.Put("/{id}", r.Handler.UpdateTenant)
					tenants.Delete("/{id}", r.Handler.DeleteTenant)
				}

				// System Numbers (All Tenants)
				system.Get("/numbers", r.Handler.ListAllNumbers)

				// Tenant Profiles
				profiles := system.Party("/tenant-profiles")
				{
					profiles.Get("/", r.Handler.ListTenantProfiles)
					profiles.Post("/", r.Handler.CreateTenantProfile)
					profiles.Get("/{id}", r.Handler.GetTenantProfile)
					profiles.Put("/{id}", r.Handler.UpdateTenantProfile)
					profiles.Delete("/{id}", r.Handler.DeleteTenantProfile)
				}

				// Gateways
				gateways := system.Party("/gateways")
				{
					gateways.Get("/", r.Handler.ListGateways)
					gateways.Post("/", r.Handler.CreateGateway)
					gateways.Get("/{id}", r.Handler.GetGateway)
					gateways.Put("/{id}", r.Handler.UpdateGateway)
					gateways.Delete("/{id}", r.Handler.DeleteGateway)
				}

				// Bridges
				bridges := system.Party("/bridges")
				{
					bridges.Get("/", r.Handler.ListBridges)
					bridges.Post("/", r.Handler.CreateBridge)
					bridges.Get("/{id}", r.Handler.GetBridge)
					bridges.Put("/{id}", r.Handler.UpdateBridge)
					bridges.Delete("/{id}", r.Handler.DeleteBridge)
				}

				// SIP Profiles
				sipProfiles := system.Party("/sip-profiles")
				{
					sipProfiles.Get("/", r.Handler.ListSIPProfiles)
					sipProfiles.Post("/", r.Handler.CreateSIPProfile)
					sipProfiles.Get("/{id}", r.Handler.GetSIPProfile)
					sipProfiles.Put("/{id}", r.Handler.UpdateSIPProfile)
					sipProfiles.Delete("/{id}", r.Handler.DeleteSIPProfile)
				}

				// System settings
				settings := system.Party("/settings")
				{
					settings.Get("/", r.Handler.GetSystemSettings)
					settings.Put("/", r.Handler.UpdateSystemSettings)
				}

				// System logs
				logs := system.Party("/logs")
				{
					logs.Get("/", r.Handler.GetSystemLogs)
				}

				// Messaging providers (system-level)
				messaging := system.Party("/messaging-providers")
				{
					messaging.Get("/", r.Handler.ListMessagingProviders)
					messaging.Post("/", r.Handler.CreateMessagingProvider)
					messaging.Get("/{id}", r.Handler.GetMessagingProvider)
					messaging.Put("/{id}", r.Handler.UpdateMessagingProvider)
					messaging.Delete("/{id}", r.Handler.DeleteMessagingProvider)
				}

				// Global dial plans
				dialplans := system.Party("/dialplans")
				{
					dialplans.Get("/", r.Handler.ListGlobalDialplans)
					dialplans.Post("/", r.Handler.CreateGlobalDialplan)
					dialplans.Get("/{id}", r.Handler.GetGlobalDialplan)
					dialplans.Put("/{id}", r.Handler.UpdateGlobalDialplan)
					dialplans.Delete("/{id}", r.Handler.DeleteGlobalDialplan)
				}

				// Access Control Lists (ACLs)
				acls := system.Party("/acls")
				{
					acls.Get("/", r.Handler.ListACLs)
					acls.Post("/", r.Handler.CreateACL)
					acls.Get("/{id}", r.Handler.GetACL)
					acls.Put("/{id}", r.Handler.UpdateACL)
					acls.Delete("/{id}", r.Handler.DeleteACL)
					// ACL nodes (entries)
					acls.Post("/{id}/nodes", r.Handler.CreateACLNode)
					acls.Put("/{id}/nodes/{nodeId}", r.Handler.UpdateACLNode)
					acls.Delete("/{id}/nodes/{nodeId}", r.Handler.DeleteACLNode)
				}

				// System Media
				media := system.Party("/media")
				{
					media.Get("/sounds", r.Handler.ListSystemSounds)
					media.Post("/sounds", r.Handler.UploadSystemSound)
					media.Get("/sounds/stream", r.Handler.StreamSystemSound)
					media.Get("/music", r.Handler.ListSystemMusic)
					media.Post("/music", r.Handler.UploadSystemMusic)
					media.Get("/music/stream", r.Handler.StreamSystemMusic)
				}

				// System status
				system.Get("/status", r.Handler.GetSystemStatus)
				system.Get("/stats", r.Handler.GetSystemStats)

				// Security - Banned IPs
				security := system.Party("/security")
				{
					security.Get("/banned-ips", r.Handler.ListBannedIPs)
					security.Post("/banned-ips", r.Handler.ReportBannedIP)
					security.Delete("/banned-ips/{ip}", r.Handler.UnbanIP)
				}

				// Device Templates (system-level master templates)
				deviceTemplates := system.Party("/device-templates")
				{
					deviceTemplates.Get("/", r.Handler.ListSystemDeviceTemplates)
					deviceTemplates.Post("/", r.Handler.CreateSystemDeviceTemplate)
					deviceTemplates.Get("/{id}", r.Handler.GetSystemDeviceTemplate)
					deviceTemplates.Put("/{id}", r.Handler.UpdateSystemDeviceTemplate)
					deviceTemplates.Delete("/{id}", r.Handler.DeleteSystemDeviceTemplate)
				}

				// Firmware Management
				firmware := system.Party("/firmware")
				{
					firmware.Get("/", r.Handler.ListFirmware)
					firmware.Post("/", r.Handler.CreateFirmware)
					firmware.Get("/{id}", r.Handler.GetFirmware)
					firmware.Put("/{id}", r.Handler.UpdateFirmware)
					firmware.Delete("/{id}", r.Handler.DeleteFirmware)
					firmware.Post("/{id}/upload", r.Handler.UploadFirmware)
					firmware.Post("/{id}/set-default", r.Handler.SetDefaultFirmware)
				}

			}

			// User portal routes
			user := protected.Party("/user")
			{
				user.Get("/devices", r.Handler.GetUserDevices)
				user.Get("/call-history", r.Handler.GetUserCallHistory)
				user.Get("/voicemail", r.Handler.GetUserVoicemail)
				user.Get("/settings", r.Handler.GetUserSettings)
				user.Put("/settings", r.Handler.UpdateUserSettings)
				user.Get("/contacts", r.Handler.GetUserContacts)
				user.Post("/contacts", r.Handler.CreateUserContact)
			}
		}
	}

	// FreeSWITCH XML CURL endpoints (inside /api for consistency with Caddy routing)
	fs := r.App.Party("/api/freeswitch")
	fs.Use(freeswitch.FreeSwitchAuthMiddleware(r.Config))
	{
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
	}

	// Device Provisioning endpoint (public - devices authenticate via MAC)
	provisioning := r.App.Party("/provisioning")
	{
		provisioning.Get("/{mac}/{filename}", r.Handler.ServeProvisioningConfig)
	}

	log.Info("All routes loaded successfully")
}

// Listen starts the HTTP server
func (r *Router) Listen(addr string) {
	log.Infof("Starting server on %s", addr)
	if err := r.App.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
