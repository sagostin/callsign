package models

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SeedDefaultAdmin creates a default system admin if no users exist in the database
// This ensures there's always a way to log into a fresh installation
func SeedDefaultAdmin(db *gorm.DB) error {
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}

	// If users exist, don't seed
	if count > 0 {
		log.Debug("Users exist in database, skipping admin seed")
		return nil
	}

	log.Info("No users found in database, creating default system admin...")

	// Get admin credentials from environment or use defaults
	username := os.Getenv("DEFAULT_ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}

	email := os.Getenv("DEFAULT_ADMIN_EMAIL")
	if email == "" {
		email = "admin@localhost"
	}

	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if password == "" {
		password = "changeme123" // Should be changed immediately after first login
		log.Warn("Using default admin password - PLEASE CHANGE THIS IMMEDIATELY!")
	}

	// Create the admin user
	admin := &User{
		Username:  username,
		Email:     email,
		Role:      RoleSystemAdmin,
		FirstName: "System",
		LastName:  "Administrator",
	}

	// Set password (this will hash it)
	if err := admin.SetPassword(password); err != nil {
		return err
	}

	// Create in database
	if err := db.Create(admin).Error; err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"username": username,
		"email":    email,
		"role":     RoleSystemAdmin,
	}).Info("Default system admin created successfully")

	return nil
}

// SeedDefaultTenantProfile creates a default tenant profile if none exists
func SeedDefaultTenantProfile(db *gorm.DB) error {
	var count int64
	if err := db.Model(&TenantProfile{}).Count(&count).Error; err != nil {
		return err
	}

	// If profiles exist, don't seed
	if count > 0 {
		return nil
	}

	log.Info("No tenant profiles found, creating default profiles...")

	// Create default profiles
	profiles := []TenantProfile{
		{
			Name:             "Starter",
			Description:      "Basic plan for small businesses",
			MaxExtensions:    10,
			MaxDevices:       15,
			MaxQueues:        2,
			MaxRingGroups:    5,
			MaxIVRMenus:      3,
			MaxConferences:   2,
			RecordingEnabled: false,
			FaxEnabled:       false,
			SMSEnabled:       false,
		},
		{
			Name:             "Professional",
			Description:      "Professional plan with advanced features",
			MaxExtensions:    50,
			MaxDevices:       75,
			MaxQueues:        10,
			MaxRingGroups:    20,
			MaxIVRMenus:      15,
			MaxConferences:   10,
			RecordingEnabled: true,
			FaxEnabled:       true,
			SMSEnabled:       false,
		},
		{
			Name:             "Enterprise",
			Description:      "Unlimited plan for large organizations",
			MaxExtensions:    -1, // -1 = unlimited
			MaxDevices:       -1,
			MaxQueues:        -1,
			MaxRingGroups:    -1,
			MaxIVRMenus:      -1,
			MaxConferences:   -1,
			RecordingEnabled: true,
			FaxEnabled:       true,
			SMSEnabled:       true,
		},
	}

	for _, profile := range profiles {
		if err := db.Create(&profile).Error; err != nil {
			return err
		}
		log.Infof("Created tenant profile: %s", profile.Name)
	}

	return nil
}

// RunSeeds executes all database seeding functions
func RunSeeds(db *gorm.DB) error {
	log.Info("Running database seeds...")

	// Seed in order
	seedFuncs := []func(*gorm.DB) error{
		SeedDefaultAdmin,
		SeedDefaultTenantProfile,
		SeedDefaultOutboundRoutes,
		SeedDefaultSounds,
		SeedDefaultFeatureCodes,
		SeedDefaultChatplans,
	}

	for _, seedFunc := range seedFuncs {
		if err := seedFunc(db); err != nil {
			return err
		}
	}

	log.Info("Database seeding completed")
	return nil
}

// SeedDefaultOutboundRoutes creates default outbound routing rules
func SeedDefaultOutboundRoutes(db *gorm.DB) error {
	var count int64
	if err := db.Model(&DefaultOutboundRoute{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Info("Creating default outbound routes...")

	routes := []DefaultOutboundRoute{
		{
			Name:        "Local/Toll-Free",
			Description: "7-11 digit local and toll-free calls",
			DigitPrefix: "",
			DigitMin:    7,
			DigitMax:    11,
			Order:       10,
			Enabled:     true,
		},
		{
			Name:        "Long Distance",
			Description: "1+ domestic long distance",
			DigitPrefix: "1",
			DigitMin:    11,
			DigitMax:    11,
			StripDigits: 0,
			Order:       20,
			Enabled:     true,
		},
		{
			Name:        "International",
			Description: "011+ international calls",
			DigitPrefix: "011",
			DigitMin:    10,
			DigitMax:    20,
			StripDigits: 0,
			Order:       30,
			Enabled:     true,
		},
		{
			Name:        "Emergency",
			Description: "911 emergency calls",
			DigitPrefix: "911",
			DigitMin:    3,
			DigitMax:    3,
			Order:       1,
			Enabled:     true,
		},
	}

	for _, route := range routes {
		if err := db.Create(&route).Error; err != nil {
			return err
		}
	}

	log.Info("Default outbound routes created")
	return nil
}

// SeedDefaultSounds creates default system sounds
func SeedDefaultSounds(db *gorm.DB) error {
	var count int64
	if err := db.Model(&Sound{}).Where("tenant_id IS NULL OR tenant_id = 0").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Info("Creating default system sounds...")

	sounds := []Sound{
		{Name: "silence", Category: "misc", Language: "en", FilePath: "/usr/share/freeswitch/sounds/en/us/callie/misc/8000/silence.wav"},
		{Name: "ring", Category: "misc", Language: "en", FilePath: "/usr/share/freeswitch/sounds/en/us/callie/ivr/8000/ivr-ring.wav"},
		{Name: "beep", Category: "misc", Language: "en", FilePath: "/usr/share/freeswitch/sounds/en/us/callie/tone/8000/beep.wav"},
		{Name: "ivr_greeting", Category: "ivr", Language: "en", FilePath: "/usr/share/freeswitch/sounds/en/us/callie/ivr/8000/ivr-welcome.wav"},
		{Name: "voicemail_greeting", Category: "voicemail", Language: "en", FilePath: "/usr/share/freeswitch/sounds/en/us/callie/voicemail/8000/vm-hello.wav"},
	}

	for _, sound := range sounds {
		sound.TenantID = 0 // System sounds
		sound.Enabled = true
		if err := db.Create(&sound).Error; err != nil {
			return err
		}
	}

	log.Info("Default system sounds created")
	return nil
}

// SeedDefaultFeatureCodes creates system feature codes for the default tenant
func SeedDefaultFeatureCodes(db *gorm.DB) error {
	// Check if any global feature codes exist
	var count int64
	if err := db.Model(&FeatureCode{}).Where("tenant_id IS NULL").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Info("Creating default feature codes...")

	// Default system feature codes (global, managed by system admin)
	defaults := []FeatureCode{
		{Code: "*97", CodeRegex: `^\*97$`, Name: "Voicemail Check", Action: FCActionVoicemail, Order: 10, IsGlobal: true},
		{Code: "*98", CodeRegex: `^\*98$`, Name: "Voicemail Direct", Action: FCActionVoicemail, Order: 11, IsGlobal: true},
		{Code: "*72", CodeRegex: `^\*72$`, Name: "Call Forward Enable", Action: FCActionCallForward, Order: 20, IsGlobal: true},
		{Code: "*73", CodeRegex: `^\*73$`, Name: "Call Forward Disable", Action: FCActionCallForward, ActionData: "disable", Order: 21, IsGlobal: true},
		{Code: "*78", CodeRegex: `^\*78$`, Name: "DND Enable", Action: FCActionDND, Order: 30, IsGlobal: true},
		{Code: "*79", CodeRegex: `^\*79$`, Name: "DND Disable", Action: FCActionDND, ActionData: "disable", Order: 31, IsGlobal: true},
		{Code: "*30", CodeRegex: `^\*30$`, Name: "Call Flow Toggle", Action: FCActionCallFlowToggle, Order: 40, IsGlobal: true},
		// Parking - auto slot
		{Code: "*70", CodeRegex: `^\*70$`, Name: "Valet Park", Action: FCActionPark, Order: 50, IsGlobal: true, ParkTimeout: 120},
		// Parking - specific slot (*7001-*7099)
		{Code: "*70XX", CodeRegex: `^\*70(?P<slot>\d{2})$`, Name: "Park to Slot", Action: FCActionParkSlot, Order: 51, IsGlobal: true, ParkTimeout: 120},
		// Retrieve from slot (*5701-*5799)
		{Code: "*57XX", CodeRegex: `^\*57(?P<slot>\d{2})$`, Name: "Retrieve from Slot", Action: FCActionParkRetrieve, Order: 52, IsGlobal: true, Extension: "5700"},
		// Pickup
		{Code: "*8", CodeRegex: `^\*8$`, Name: "Group Pickup", Action: FCActionPickup, ActionData: "group", Order: 60, IsGlobal: true},
		{Code: "**", CodeRegex: `^\*\*(\d+)$`, Name: "Directed Pickup", Action: FCActionPickup, Order: 61, IsGlobal: true},
		// Intercom
		{Code: "*0", CodeRegex: `^\*0(\d+)$`, Name: "Intercom", Action: FCActionIntercom, Order: 70, IsGlobal: true},
	}

	for _, fc := range defaults {
		fc.Enabled = true
		if err := db.Create(&fc).Error; err != nil {
			return err
		}
	}

	log.Info("Default feature codes created")
	return nil
}

// SeedDefaultChatplans creates default SMS/MMS routing rules
func SeedDefaultChatplans(db *gorm.DB) error {
	var count int64
	if err := db.Model(&Chatplan{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Info("Creating default chatplans...")

	chatplans := []Chatplan{
		{
			Name:      "Default Inbound",
			Context:   "default",
			Direction: "inbound",
			ToPattern: ".*",
			Action:    ChatplanActionForward,
			ForwardTo: "${extension}",
			Order:     100,
			Enabled:   true,
		},
		{
			Name:          "Auto-Reply OOO",
			Context:       "default",
			Direction:     "inbound",
			MessageMatch:  "(?i)hello|hi|help",
			Action:        ChatplanActionReply,
			ReplyTemplate: "Thanks for your message! We'll get back to you shortly.",
			Order:         50,
			Enabled:       false, // Disabled by default
		},
	}

	for _, cp := range chatplans {
		if err := db.Create(&cp).Error; err != nil {
			return err
		}
	}

	log.Info("Default chatplans created")
	return nil
}
