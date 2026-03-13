package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// =====================
// Admin Greeting Script Handlers
// =====================

// ListGreetingScripts returns all greeting scripts for the tenant
func (h *Handler) ListGreetingScripts(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		h.logWarn("GREETINGS", "ListGreetingScripts: Tenant context required", h.reqFields(c, nil))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	var scripts []models.GreetingScript
	query := h.DB.Where("tenant_id = ? AND user_id IS NULL", tenantID)

	// Filters
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if search := c.Query("search"); search != "" {
		search = "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	if err := query.Order("updated_at DESC").Find(&scripts).Error; err != nil {
		h.logError("GREETINGS", "ListGreetingScripts: Failed to fetch scripts", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch scripts"})
	}

	return c.JSON(fiber.Map{"data": scripts})
}

// GetGreetingScript returns a single greeting script
func (h *Handler) GetGreetingScript(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND tenant_id = ? AND user_id IS NULL", id, tenantID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Script not found"})
	}

	return c.JSON(fiber.Map{"data": script})
}

// CreateGreetingScript creates a new greeting script
func (h *Handler) CreateGreetingScript(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		h.logWarn("GREETINGS", "CreateGreetingScript: Tenant context required", h.reqFields(c, nil))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	var input struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ScriptText  string  `json:"script_text"`
		Category    string  `json:"category"`
		Provider    string  `json:"provider"`
		VoiceID     string  `json:"voice_id"`
		Speed       float64 `json:"speed"`
		Pitch       float64 `json:"pitch"`
		Language    string  `json:"language"`
		Generate    bool    `json:"generate"` // If true, generate audio immediately
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if input.Name == "" || input.ScriptText == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Name and script text are required"})
	}

	// Defaults
	if input.Provider == "" {
		input.Provider = "flite"
	}
	if input.VoiceID == "" {
		input.VoiceID = "default"
	}
	if input.Speed == 0 {
		input.Speed = 1.0
	}
	if input.Pitch == 0 {
		input.Pitch = 1.0
	}
	if input.Language == "" {
		input.Language = "en-US"
	}
	if input.Category == "" {
		input.Category = "custom"
	}

	script := models.GreetingScript{
		TenantID:    &tenantID,
		Name:        input.Name,
		Description: input.Description,
		ScriptText:  input.ScriptText,
		Category:    input.Category,
		Provider:    models.TTSProvider(input.Provider),
		VoiceID:     input.VoiceID,
		Speed:       input.Speed,
		Pitch:       input.Pitch,
		Language:    input.Language,
		Status:      models.GreetingStatusDraft,
	}

	if err := h.DB.Create(&script).Error; err != nil {
		h.logError("GREETINGS", "CreateGreetingScript: Failed to create", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create script"})
	}

	// Auto-generate if requested
	if input.Generate {
		h.generateGreetingAudio(&script, tenantID)
		h.DB.Save(&script)
	}

	h.logInfo("GREETINGS", "CreateGreetingScript: Created successfully", h.reqFields(c, map[string]interface{}{"script_id": script.ID}))
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": script})
}

// UpdateGreetingScript updates an existing greeting script
func (h *Handler) UpdateGreetingScript(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND tenant_id = ? AND user_id IS NULL", id, tenantID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Script not found"})
	}

	var input struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ScriptText  string  `json:"script_text"`
		Category    string  `json:"category"`
		Provider    string  `json:"provider"`
		VoiceID     string  `json:"voice_id"`
		Speed       float64 `json:"speed"`
		Pitch       float64 `json:"pitch"`
		Language    string  `json:"language"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if input.Name != "" {
		script.Name = input.Name
	}
	script.Description = input.Description
	if input.ScriptText != "" {
		script.ScriptText = input.ScriptText
	}
	if input.Category != "" {
		script.Category = input.Category
	}
	if input.Provider != "" {
		script.Provider = models.TTSProvider(input.Provider)
	}
	if input.VoiceID != "" {
		script.VoiceID = input.VoiceID
	}
	if input.Speed > 0 {
		script.Speed = input.Speed
	}
	if input.Pitch > 0 {
		script.Pitch = input.Pitch
	}
	if input.Language != "" {
		script.Language = input.Language
	}

	if err := h.DB.Save(&script).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update script"})
	}

	return c.JSON(fiber.Map{"data": script})
}

// DeleteGreetingScript deletes a greeting script and its generated file
func (h *Handler) DeleteGreetingScript(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND tenant_id = ? AND user_id IS NULL", id, tenantID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Script not found"})
	}

	// Delete generated file if exists
	if script.FilePath != "" {
		os.Remove(script.FilePath)
	}

	if err := h.DB.Delete(&script).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete script"})
	}

	return c.JSON(fiber.Map{"message": "Script deleted successfully"})
}

// GenerateGreeting triggers audio generation for a greeting script
func (h *Handler) GenerateGreeting(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Script not found"})
	}

	h.generateGreetingAudio(&script, tenantID)

	if err := h.DB.Save(&script).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save generated file info"})
	}

	if script.Status == models.GreetingStatusError {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": script.Error, "data": script})
	}

	return c.JSON(fiber.Map{"message": "Audio generated successfully", "data": script})
}

// StreamGreeting serves the generated audio file for playback
func (h *Handler) StreamGreeting(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		// Fallback to JWT tenant
		tenantID = middleware.GetTenantID(c)
	}
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Script not found"})
	}

	if script.FilePath == "" || script.Status != models.GreetingStatusReady {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "No audio file generated yet"})
	}

	if _, err := os.Stat(script.FilePath); os.IsNotExist(err) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Audio file not found on disk"})
	}

	c.Set("Content-Type", "audio/wav")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", script.FileName))
	return c.SendFile(script.FilePath)
}

// ListTTSVoices returns available TTS voices
func (h *Handler) ListTTSVoices(c *fiber.Ctx) error {
	// Return a curated list of available voices per provider
	voices := []fiber.Map{
		// Flite voices (local, always available)
		{"provider": "flite", "voice_id": "default", "name": "Default (Flite)", "language": "en-US", "gender": "male"},
		{"provider": "flite", "voice_id": "kal", "name": "Kal (Flite)", "language": "en-US", "gender": "male"},
		{"provider": "flite", "voice_id": "slt", "name": "SLT (Flite)", "language": "en-US", "gender": "female"},
		{"provider": "flite", "voice_id": "rms", "name": "RMS (Flite)", "language": "en-US", "gender": "male"},
		{"provider": "flite", "voice_id": "awb", "name": "AWB (Flite)", "language": "en-US", "gender": "male"},
	}

	// Add ElevenLabs voices if API key is configured
	if h.Config.ElevenLabsAPIKey != "" {
		voices = append(voices,
			fiber.Map{"provider": "elevenlabs", "voice_id": "21m00Tcm4TlvDq8ikWAM", "name": "Rachel (ElevenLabs)", "language": "en-US", "gender": "female"},
			fiber.Map{"provider": "elevenlabs", "voice_id": "AZnzlk1XvdvUeBnXmlld", "name": "Domi (ElevenLabs)", "language": "en-US", "gender": "female"},
			fiber.Map{"provider": "elevenlabs", "voice_id": "EXAVITQu4vr4xnSDxMaL", "name": "Bella (ElevenLabs)", "language": "en-US", "gender": "female"},
			fiber.Map{"provider": "elevenlabs", "voice_id": "ErXwobaYiN019PkySvjV", "name": "Antoni (ElevenLabs)", "language": "en-US", "gender": "male"},
			fiber.Map{"provider": "elevenlabs", "voice_id": "MF3mGyEYCl7XYWbV9V6O", "name": "Elli (ElevenLabs)", "language": "en-US", "gender": "female"},
			fiber.Map{"provider": "elevenlabs", "voice_id": "TxGEqnHWrfWFTfGW9XjX", "name": "Josh (ElevenLabs)", "language": "en-US", "gender": "male"},
			fiber.Map{"provider": "elevenlabs", "voice_id": "VR6AewLTigWG4xSOukaG", "name": "Arnold (ElevenLabs)", "language": "en-US", "gender": "male"},
			fiber.Map{"provider": "elevenlabs", "voice_id": "pNInz6obpgDQGcFmaJgB", "name": "Adam (ElevenLabs)", "language": "en-US", "gender": "male"},
		)
	}

	// Add OpenAI voices if API key is configured
	if h.Config.OpenAIAPIKey != "" {
		voices = append(voices,
			fiber.Map{"provider": "openai", "voice_id": "alloy", "name": "Alloy (OpenAI)", "language": "en-US", "gender": "neutral"},
			fiber.Map{"provider": "openai", "voice_id": "echo", "name": "Echo (OpenAI)", "language": "en-US", "gender": "male"},
			fiber.Map{"provider": "openai", "voice_id": "fable", "name": "Fable (OpenAI)", "language": "en-US", "gender": "neutral"},
			fiber.Map{"provider": "openai", "voice_id": "onyx", "name": "Onyx (OpenAI)", "language": "en-US", "gender": "male"},
			fiber.Map{"provider": "openai", "voice_id": "nova", "name": "Nova (OpenAI)", "language": "en-US", "gender": "female"},
			fiber.Map{"provider": "openai", "voice_id": "shimmer", "name": "Shimmer (OpenAI)", "language": "en-US", "gender": "female"},
		)
	}

	// Also return DB-stored voices
	var dbVoices []models.TTSVoice
	h.DB.Where("enabled = true").Find(&dbVoices)

	return c.JSON(fiber.Map{
		"data":      voices,
		"db_voices": dbVoices,
		"providers": h.getAvailableProviders(),
	})
}

// getAvailableProviders returns which TTS providers are configured
func (h *Handler) getAvailableProviders() []fiber.Map {
	providers := []fiber.Map{
		{"id": "flite", "name": "Flite (Local)", "available": true},
	}
	providers = append(providers, fiber.Map{
		"id": "elevenlabs", "name": "ElevenLabs", "available": h.Config.ElevenLabsAPIKey != "",
	})
	providers = append(providers, fiber.Map{
		"id": "openai", "name": "OpenAI TTS", "available": h.Config.OpenAIAPIKey != "",
	})
	return providers
}

// =====================
// User Greeting Handlers (Voicemail)
// =====================

// ListUserGreetings returns the current user's greeting scripts
func (h *Handler) ListUserGreetings(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var scripts []models.GreetingScript
	query := h.DB.Where("user_id = ?", claims.UserID)
	if claims.TenantID != nil {
		query = query.Where("tenant_id = ?", *claims.TenantID)
	}

	if err := query.Order("updated_at DESC").Find(&scripts).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch greetings"})
	}

	return c.JSON(fiber.Map{"data": scripts})
}

// CreateUserGreeting creates a user's voicemail greeting
func (h *Handler) CreateUserGreeting(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var input struct {
		Name       string  `json:"name"`
		ScriptText string  `json:"script_text"`
		Provider   string  `json:"provider"`
		VoiceID    string  `json:"voice_id"`
		Speed      float64 `json:"speed"`
		Language   string  `json:"language"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if input.ScriptText == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Script text is required"})
	}

	if input.Name == "" {
		input.Name = "Voicemail Greeting"
	}
	if input.Provider == "" {
		input.Provider = "flite"
	}
	if input.VoiceID == "" {
		input.VoiceID = "default"
	}
	if input.Speed == 0 {
		input.Speed = 1.0
	}
	if input.Language == "" {
		input.Language = "en-US"
	}

	userID := claims.UserID
	script := models.GreetingScript{
		TenantID:   claims.TenantID,
		UserID:     &userID,
		Name:       input.Name,
		ScriptText: input.ScriptText,
		Category:   "voicemail",
		Provider:   models.TTSProvider(input.Provider),
		VoiceID:    input.VoiceID,
		Speed:      input.Speed,
		Pitch:      1.0,
		Language:   input.Language,
		Status:     models.GreetingStatusDraft,
	}

	if err := h.DB.Create(&script).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create greeting"})
	}

	// Auto-generate
	var tenantID uint
	if claims.TenantID != nil {
		tenantID = *claims.TenantID
	}
	h.generateGreetingAudio(&script, tenantID)
	h.DB.Save(&script)

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": script})
}

// UpdateUserGreeting updates a user greeting
func (h *Handler) UpdateUserGreeting(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND user_id = ?", id, claims.UserID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Greeting not found"})
	}

	var input struct {
		Name       string  `json:"name"`
		ScriptText string  `json:"script_text"`
		Provider   string  `json:"provider"`
		VoiceID    string  `json:"voice_id"`
		Speed      float64 `json:"speed"`
		Language   string  `json:"language"`
		Regenerate bool    `json:"regenerate"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if input.Name != "" {
		script.Name = input.Name
	}
	if input.ScriptText != "" {
		script.ScriptText = input.ScriptText
	}
	if input.Provider != "" {
		script.Provider = models.TTSProvider(input.Provider)
	}
	if input.VoiceID != "" {
		script.VoiceID = input.VoiceID
	}
	if input.Speed > 0 {
		script.Speed = input.Speed
	}
	if input.Language != "" {
		script.Language = input.Language
	}

	if input.Regenerate {
		var tenantID uint
		if claims.TenantID != nil {
			tenantID = *claims.TenantID
		}
		h.generateGreetingAudio(&script, tenantID)
	}

	if err := h.DB.Save(&script).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update greeting"})
	}

	return c.JSON(fiber.Map{"data": script})
}

// DeleteUserGreeting deletes a user's greeting
func (h *Handler) DeleteUserGreeting(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND user_id = ?", id, claims.UserID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Greeting not found"})
	}

	if script.FilePath != "" {
		os.Remove(script.FilePath)
	}

	if err := h.DB.Delete(&script).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete greeting"})
	}

	return c.JSON(fiber.Map{"message": "Greeting deleted successfully"})
}

// ActivateUserGreeting sets a greeting as the active voicemail greeting
func (h *Handler) ActivateUserGreeting(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND user_id = ?", id, claims.UserID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Greeting not found"})
	}

	if script.Status != models.GreetingStatusReady || script.FilePath == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Greeting audio is not ready"})
	}

	// Find the user's voicemail box and update greeting path
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Update voicemail box greeting if it exists
	if user.Extension != "" && claims.TenantID != nil {
		h.DB.Model(&models.VoicemailBox{}).
			Where("extension = ? AND tenant_id = ?", user.Extension, *claims.TenantID).
			Updates(map[string]interface{}{
				"greeting_path": script.FilePath,
				"greeting_type": "custom",
			})
	}

	return c.JSON(fiber.Map{
		"message": "Greeting activated as voicemail greeting",
		"data":    script,
	})
}

// StreamUserGreeting streams a user's generated greeting audio
func (h *Handler) StreamUserGreeting(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var script models.GreetingScript
	if err := h.DB.Where("id = ? AND user_id = ?", id, claims.UserID).First(&script).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Greeting not found"})
	}

	if script.FilePath == "" || script.Status != models.GreetingStatusReady {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "No audio file generated"})
	}

	if _, err := os.Stat(script.FilePath); os.IsNotExist(err) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Audio file not found on disk"})
	}

	c.Set("Content-Type", "audio/wav")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", script.FileName))
	return c.SendFile(script.FilePath)
}

// =====================
// Audio Generation Core
// =====================

// generateGreetingAudio synthesizes audio from script text using the configured TTS engine.
// Updates the script in-place with file path, status, etc.
func (h *Handler) generateGreetingAudio(script *models.GreetingScript, tenantID uint) {
	script.Status = models.GreetingStatusGenerating
	script.Error = ""

	// Build output path
	storageRoot := h.Config.MediaBasePath
	if storageRoot == "" {
		storageRoot = "/usr/share/freeswitch/sounds"
	}

	var greetingsDir string
	if tenantID > 0 {
		greetingsDir = filepath.Join(storageRoot, "tenants", strconv.FormatUint(uint64(tenantID), 10), "greetings")
	} else {
		greetingsDir = filepath.Join(storageRoot, "greetings")
	}

	if err := os.MkdirAll(greetingsDir, 0755); err != nil {
		script.Status = models.GreetingStatusError
		script.Error = fmt.Sprintf("Failed to create directory: %v", err)
		return
	}

	// Generate unique filename
	fileName := fmt.Sprintf("%s.wav", uuid.New().String()[:12])
	outPath := filepath.Join(greetingsDir, fileName)

	// Synthesize based on provider
	var err error
	switch string(script.Provider) {
	case "flite":
		err = synthesizeFlite(script.ScriptText, script.VoiceID, outPath)
	case "elevenlabs":
		err = synthesizeElevenLabs(script.ScriptText, script.VoiceID, outPath, h.Config.ElevenLabsAPIKey)
	case "openai":
		err = synthesizeOpenAI(script.ScriptText, script.VoiceID, outPath, h.Config.OpenAIAPIKey)
	default:
		// Default to flite
		err = synthesizeFlite(script.ScriptText, script.VoiceID, outPath)
	}

	if err != nil {
		script.Status = models.GreetingStatusError
		script.Error = err.Error()
		return
	}

	// Get file info
	info, statErr := os.Stat(outPath)
	if statErr != nil {
		script.Status = models.GreetingStatusError
		script.Error = fmt.Sprintf("Generated file not found: %v", statErr)
		return
	}

	// Remove old file if different path
	if script.FilePath != "" && script.FilePath != outPath {
		os.Remove(script.FilePath)
	}

	now := time.Now()
	script.FilePath = outPath
	script.FileName = fileName
	script.FileSize = info.Size()
	script.GeneratedAt = &now
	script.Status = models.GreetingStatusReady
}

// synthesizeFlite generates audio using the local flite TTS engine
func synthesizeFlite(text, voice, outPath string) error {
	args := []string{"-t", text, "-o", outPath}
	if voice != "" && voice != "default" {
		args = append([]string{"-voice", voice}, args...)
	}
	cmd := exec.Command("flite", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("flite synthesis failed: %s: %w", string(out), err)
	}
	return nil
}

// synthesizeElevenLabs generates audio using the ElevenLabs API
func synthesizeElevenLabs(text, voiceID, outPath, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("ElevenLabs API key not configured")
	}

	// Use curl to call ElevenLabs API (simple approach, no extra Go HTTP deps needed)
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	payload := fmt.Sprintf(`{"text":"%s","model_id":"eleven_monolingual_v1","voice_settings":{"stability":0.5,"similarity_boost":0.5}}`,
		strings.ReplaceAll(text, `"`, `\"`))

	cmd := exec.Command("curl", "-s",
		"-X", "POST", url,
		"-H", fmt.Sprintf("xi-api-key: %s", apiKey),
		"-H", "Content-Type: application/json",
		"-H", "Accept: audio/mpeg",
		"-d", payload,
		"-o", outPath,
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ElevenLabs API call failed: %s: %w", string(out), err)
	}

	// Verify output file exists and has non-zero size
	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("ElevenLabs returned empty response")
	}

	return nil
}

// synthesizeOpenAI generates audio using the OpenAI TTS API
func synthesizeOpenAI(text, voice, outPath, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("OpenAI API key not configured")
	}

	payload := fmt.Sprintf(`{"model":"tts-1","input":"%s","voice":"%s","response_format":"wav"}`,
		strings.ReplaceAll(text, `"`, `\"`), voice)

	cmd := exec.Command("curl", "-s",
		"-X", "POST", "https://api.openai.com/v1/audio/speech",
		"-H", fmt.Sprintf("Authorization: Bearer %s", apiKey),
		"-H", "Content-Type: application/json",
		"-d", payload,
		"-o", outPath,
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("OpenAI TTS API call failed: %s: %w", string(out), err)
	}

	// Verify output file exists and has content
	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("OpenAI TTS returned empty response")
	}

	return nil
}
