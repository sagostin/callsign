package models_test

import (
	"callsign/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Tenant{},
		&models.TenantProfile{},
		&models.Extension{},
		&models.FeatureCode{},
		&models.Conference{},
		&models.Queue{},
		&models.RingGroup{},
		&models.DefaultOutboundRoute{},
		&models.Sound{},
	)
	require.NoError(t, err)

	return db
}

func TestUserPasswordHashing(t *testing.T) {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Test password setting
	err := user.SetPassword("testpassword123")
	assert.NoError(t, err)
	assert.NotEmpty(t, user.Password)
	assert.NotEqual(t, "testpassword123", user.Password)

	// Test password verification
	assert.True(t, user.CheckPassword("testpassword123"))
	assert.False(t, user.CheckPassword("wrongpassword"))
}

func TestUserPermissions(t *testing.T) {
	// System admin should have all permissions
	admin := &models.User{Role: models.RoleSystemAdmin}
	assert.True(t, admin.HasPermission(models.PermSystemManage))
	assert.True(t, admin.HasPermission(models.PermExtensionCreate))

	// Tenant admin should have tenant-level permissions
	tenantAdmin := &models.User{Role: models.RoleTenantAdmin}
	assert.True(t, tenantAdmin.HasPermission(models.PermExtensionCreate))
	assert.False(t, tenantAdmin.HasPermission(models.PermSystemManage))

	// Regular user should only have personal permissions
	user := &models.User{Role: models.RoleUser}
	assert.True(t, user.HasPermission(models.PermProfileView))
	assert.False(t, user.HasPermission(models.PermExtensionCreate))
}

func TestFeatureCodeValidation(t *testing.T) {
	tenantID := uint(1)

	// Valid custom feature code
	fc := &models.FeatureCode{
		TenantID: &tenantID,
		Code:     "*55",
		Name:     "Custom Code",
		Action:   models.FCActionWebhook,
		IsGlobal: false,
	}
	assert.NoError(t, fc.Validate())

	// Valid global feature code (no tenant, IsGlobal true)
	fcGlobal := &models.FeatureCode{
		TenantID: nil, // Global codes have nil TenantID
		Code:     "*97",
		Name:     "Check Voicemail",
		Action:   models.FCActionVoicemail,
		IsGlobal: true,
	}
	assert.NoError(t, fcGlobal.Validate())

	// Invalid code format
	fc3 := &models.FeatureCode{
		TenantID: &tenantID,
		Code:     "55", // Missing * or #
		Name:     "Invalid",
		Action:   models.FCActionCustom,
		IsGlobal: false,
	}
	assert.Error(t, fc3.Validate())
}

func TestFeatureCodeIsGlobal(t *testing.T) {
	tenantID := uint(1)

	// Global code (IsGlobal true, nil TenantID)
	global := &models.FeatureCode{
		TenantID: nil,
		Code:     "*97",
		IsGlobal: true,
	}
	assert.True(t, global.IsGlobal)
	assert.Nil(t, global.TenantID)

	// Tenant-specific code
	tenant := &models.FeatureCode{
		TenantID: &tenantID,
		Code:     "*55",
		IsGlobal: false,
	}
	assert.False(t, tenant.IsGlobal)
	assert.NotNil(t, tenant.TenantID)
}

func TestConferenceSessionDuration(t *testing.T) {
	now := time.Now()
	session := &models.ConferenceSession{
		StartTime: now.Add(-10 * time.Minute),
	}

	// Active session
	assert.True(t, session.IsActive())
	assert.GreaterOrEqual(t, session.Duration(), 600) // At least 10 minutes

	// Ended session
	endTime := now.Add(-5 * time.Minute)
	session.EndTime = &endTime
	assert.False(t, session.IsActive())
	assert.Equal(t, 300, session.Duration()) // Exactly 5 minutes
}

func TestDefaultOutboundRouteCRUD(t *testing.T) {
	db := setupTestDB(t)

	route := &models.DefaultOutboundRoute{
		Name:        "Test Route",
		DigitPrefix: "1",
		DigitMin:    11,
		DigitMax:    11,
		Order:       10,
		Enabled:     true,
	}

	// Create
	err := db.Create(route).Error
	assert.NoError(t, err)
	assert.NotZero(t, route.ID)
	assert.NotEmpty(t, route.UUID)

	// Read
	var found models.DefaultOutboundRoute
	err = db.First(&found, route.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Route", found.Name)

	// Update
	err = db.Model(&found).Update("name", "Updated Route").Error
	assert.NoError(t, err)

	// Delete
	err = db.Delete(&found).Error
	assert.NoError(t, err)
}

func TestSeedFunctions(t *testing.T) {
	db := setupTestDB(t)

	// Test seeding doesn't error on empty DB
	err := models.SeedDefaultOutboundRoutes(db)
	assert.NoError(t, err)

	// Verify routes created
	var count int64
	db.Model(&models.DefaultOutboundRoute{}).Count(&count)
	assert.Equal(t, int64(4), count) // Emergency, Local, Long Distance, International

	// Running again should not create duplicates
	err = models.SeedDefaultOutboundRoutes(db)
	assert.NoError(t, err)
	db.Model(&models.DefaultOutboundRoute{}).Count(&count)
	assert.Equal(t, int64(4), count)
}

func TestExtensionCRUD(t *testing.T) {
	db := setupTestDB(t)

	ext := &models.Extension{
		TenantID:  1,
		Extension: "1001",
		Password:  "secret",
		Enabled:   true,
	}

	// Create
	err := db.Create(ext).Error
	assert.NoError(t, err)
	assert.NotZero(t, ext.ID)

	// Read
	var found models.Extension
	err = db.Where("extension = ?", "1001").First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, uint(1), found.TenantID)

	// Update
	err = db.Model(&found).Update("enabled", false).Error
	assert.NoError(t, err)

	// Delete
	err = db.Delete(&found).Error
	assert.NoError(t, err)
}

func TestQueueCRUD(t *testing.T) {
	db := setupTestDB(t)

	queue := &models.Queue{
		TenantID: 1,
		Name:     "Support Queue",
		Strategy: models.QueueStrategyRingAll,
		Enabled:  true,
	}

	err := db.Create(queue).Error
	assert.NoError(t, err)
	assert.NotZero(t, queue.ID)
	assert.NotEmpty(t, queue.UUID)
}

func TestConferenceCRUD(t *testing.T) {
	db := setupTestDB(t)

	conf := &models.Conference{
		TenantID:  1,
		Name:      "Team Meeting",
		Extension: "8000",
		Enabled:   true,
	}

	err := db.Create(conf).Error
	assert.NoError(t, err)
	assert.NotZero(t, conf.ID)
	assert.NotEmpty(t, conf.UUID)
}
