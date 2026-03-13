package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// =====================
// DB-Backed Media Handlers (Audio Library)
// =====================

// ListMediaFiles returns a list of media files for the tenant
func (h *Handler) ListMediaFiles(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		h.logWarn("MEDIA_DB", "ListMediaFiles: Tenant context required", h.reqFields(c, nil))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	var files []models.MediaFile
	query := h.DB.Where("tenant_id = ?", tenantID)

	// Filters
	if mediaType := c.Query("type"); mediaType != "" {
		query = query.Where("type = ?", mediaType)
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if search := c.Query("search"); search != "" {
		search = "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	if err := query.Find(&files).Error; err != nil {
		h.logError("MEDIA_DB", "ListMediaFiles: Failed to fetch files", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch files"})
	}

	// Auto-Sync: If no files found in DB, check disk for existing files
	// This helps with migration or manual file placement
	if len(files) == 0 && c.Query("search") == "" && c.Query("type") == "" {
		synced, err := h.syncTenantMediaFiles(tenantID)
		if err == nil && len(synced) > 0 {
			files = synced
		}
	}

	return c.JSON(fiber.Map{"data": files})
}

// syncTenantMediaFiles scans the tenant's media directory and adds missing files to DB
func (h *Handler) syncTenantMediaFiles(tenantID uint) ([]models.MediaFile, error) {
	var synced []models.MediaFile
	storageRoot := "/usr/share/freeswitch/sounds"

	// Define paths to scan based on our convention
	// We scan both the new convention 'tenants/{id}/media/{type}'
	// and potentially the root 'tenants/{id}' for legacy/flat files if needed.
	// For now, let's stick to the structured folders we expect: greeting, recording, music, custom

	dirs := []string{"greeting", "recording", "music", "custom", "queue"}

	for _, typeDir := range dirs {
		mediaType := models.MediaType(typeDir)
		if typeDir == "music" {
			mediaType = models.MediaTypeMusic
		}

		relDir := fmt.Sprintf("tenants/%d/media/%s", tenantID, typeDir)
		fullDir := filepath.Join(storageRoot, relDir)

		entries, err := os.ReadDir(fullDir)
		if err != nil {
			continue // Directory might not exist, skip
		}

		for _, entry := range entries {
			if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			// Check extension
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
				continue
			}

			// Check if exists in DB
			var count int64
			filename := entry.Name()
			h.DB.Model(&models.MediaFile{}).Where("tenant_id = ? AND filename = ? AND start_path = ?", tenantID, filename, relDir).Count(&count)
			// Actually just checking filename and tenant might be enough if we enforce unique filenames,
			// but let's check path too to be safe or just filename if we want to be loose.
			// Let's check filename for now to avoid duplicates if path varies slightly.
			h.DB.Model(&models.MediaFile{}).Where("tenant_id = ? AND filename = ?", tenantID, filename).Count(&count)

			if count == 0 {
				// Insert
				info, _ := entry.Info()
				newFile := models.MediaFile{
					TenantID:    tenantID,
					Name:        strings.TrimSuffix(filename, ext), // Use filename as name
					Description: "Auto-synced file",
					Type:        mediaType,
					Category:    "Imported",
					Filename:    filename,
					Path:        filepath.Join(relDir, filename),
					MimeType:    "audio/" + strings.TrimPrefix(ext, "."),
					Size:        info.Size(),
				}

				if err := h.DB.Create(&newFile).Error; err == nil {
					synced = append(synced, newFile)
				}
			}
		}
	}

	return synced, nil
}

// UploadMediaFile handles uploading a new media file
func (h *Handler) UploadMediaFile(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		h.logWarn("MEDIA_DB", "UploadMediaFile: Tenant context required", h.reqFields(c, nil))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	// Form fields
	name := c.FormValue("name")
	description := c.FormValue("description")
	mediaType := c.FormValue("type")
	category := c.FormValue("category")

	if name == "" {
		h.logWarn("MEDIA_DB", "UploadMediaFile: Name is required", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	// File upload
	header, err := c.FormFile("file")
	if err != nil {
		h.logWarn("MEDIA_DB", "UploadMediaFile: Failed to read file", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Failed to read file"})
	}
	file, err := header.Open()
	if err != nil {
		h.logError("MEDIA_DB", "UploadMediaFile: Failed to open file", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	// Validate extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		h.logWarn("MEDIA_DB", "UploadMediaFile: Invalid file type. Only .wav, .mp3, .ogg allowed", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type. Only .wav, .mp3, .ogg allowed"})
	}

	// Storage path: tenants/{id}/media/{type}/
	// Default type if empty
	if mediaType == "" {
		mediaType = "custom"
	}

	relPath := fmt.Sprintf("tenants/%d/media/%s", tenantID, mediaType)
	storageRoot := "/usr/share/freeswitch/sounds" // Or a separate storage dir
	fullDir := filepath.Join(storageRoot, relPath)

	if err := os.MkdirAll(fullDir, 0755); err != nil {
		h.logError("MEDIA_DB", "UploadMediaFile: Failed to create directory", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create directory"})
	}

	// Unique filename to prevent collisions? Or allow overwrite?
	// Let's prepend timestamp or use UUID
	// safeFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	safeFilename := header.Filename // simplistic for now
	dstPath := filepath.Join(fullDir, safeFilename)

	out, err := os.Create(dstPath)
	if err != nil {
		h.logError("MEDIA_DB", "UploadMediaFile: Failed to create destination file", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create destination file"})
	}
	defer out.Close()

	written, err := io.Copy(out, file)
	if err != nil {
		h.logError("MEDIA_DB", "UploadMediaFile: Failed to save file", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	// Create DB record
	mediaFile := models.MediaFile{
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		Type:        models.MediaType(mediaType),
		Category:    category,
		Filename:    safeFilename,
		Path:        filepath.Join(relPath, safeFilename),
		MimeType:    header.Header.Get("Content-Type"),
		Size:        written,
	}

	if err := h.DB.Create(&mediaFile).Error; err != nil {
		h.logError("MEDIA_DB", "UploadMediaFile: Failed to create database record", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create database record"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "File uploaded successfully", "data": mediaFile})
}

// UpdateMediaFile updates metadata for a media file
func (h *Handler) UpdateMediaFile(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		h.logWarn("MEDIA_DB", "UpdateMediaFile: Tenant context required", h.reqFields(c, nil))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var mediaFile models.MediaFile

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&mediaFile).Error; err != nil {
		h.logWarn("MEDIA_DB", "UpdateMediaFile: File not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	// Update fields
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Type        string `json:"type"`
	}

	if err := c.BodyParser(&input); err != nil {
		h.logWarn("MEDIA_DB", "UpdateMediaFile: Invalid input", h.reqFields(c, nil))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	mediaFile.Name = input.Name
	mediaFile.Description = input.Description
	mediaFile.Category = input.Category
	if input.Type != "" {
		mediaFile.Type = models.MediaType(input.Type)
	}

	if err := h.DB.Save(&mediaFile).Error; err != nil {
		h.logError("MEDIA_DB", "UpdateMediaFile: Failed to update record", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update record"})
	}

	return c.JSON(fiber.Map{"message": "Updated successfully", "data": mediaFile})
}

// DeleteMediaFile removes a media file
func (h *Handler) DeleteMediaFile(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		h.logWarn("MEDIA_DB", "DeleteMediaFile: Tenant context required", h.reqFields(c, nil))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var mediaFile models.MediaFile

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&mediaFile).Error; err != nil {
		h.logWarn("MEDIA_DB", "DeleteMediaFile: File not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	// Delete from disk
	storageRoot := "/usr/share/freeswitch/sounds"
	fullPath := filepath.Join(storageRoot, mediaFile.Path)

	// Prepare for deletion
	// Use transaction
	tx := h.DB.Begin()

	if err := tx.Delete(&mediaFile).Error; err != nil {
		tx.Rollback()
		h.logError("MEDIA_DB", "DeleteMediaFile: Failed to delete database record", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete database record"})
	}

	// Try removing file, but don't fail hard if missing (could be manually deleted)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		// Log warning?
		fmt.Printf("Warning: Failed to delete file %s: %v\n", fullPath, err)
	}

	tx.Commit()

	return c.JSON(fiber.Map{"message": "File deleted successfully"})
}

// StreamMediaFile serves an audio file for playback
func (h *Handler) StreamMediaFile(c *fiber.Ctx) error {
	// Try scoped tenant ID first (set by RequireTenant middleware)
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		// Fallback to direct tenant ID from JWT claims
		tenantID = middleware.GetTenantID(c)
	}
	if tenantID == 0 {
		h.logWarn("MEDIA_DB", "StreamMediaFile: Tenant context required", h.reqFields(c, nil))
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	id, _ := strconv.Atoi(c.Params("id"))
	var mediaFile models.MediaFile

	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&mediaFile).Error; err != nil {
		h.logWarn("MEDIA_DB", "StreamMediaFile: File not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	storageRoot := "/usr/share/freeswitch/sounds"
	fullPath := filepath.Join(storageRoot, mediaFile.Path)

	// Check file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		h.logWarn("MEDIA_DB", "StreamMediaFile: File not found on disk", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found on disk"})
	}

	// Determine content type
	contentType := mediaFile.MimeType
	if contentType == "" {
		ext := strings.ToLower(filepath.Ext(fullPath))
		switch ext {
		case ".wav":
			contentType = "audio/wav"
		case ".mp3":
			contentType = "audio/mpeg"
		case ".ogg":
			contentType = "audio/ogg"
		default:
			contentType = "application/octet-stream"
		}
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", mediaFile.Filename))

	return c.SendFile(fullPath)
}
