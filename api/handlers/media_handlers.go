package handlers

import (
	"callsign/middleware"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// =====================
// System Media (Sounds & Music)
// =====================

type FileNode struct {
	Name       string      `json:"name"`
	Path       string      `json:"path"` // Relative to root
	Type       string      `json:"type"` // "directory" or "file"
	Size       int64       `json:"size,omitempty"`
	Children   []*FileNode `json:"children,omitempty"`
	Source     string      `json:"source,omitempty"`      // "system" or "tenant"
	IsOverride bool        `json:"is_override,omitempty"` // true if tenant file overrides system file
}

// ListSystemSounds returns the directory structure of /usr/share/freeswitch/sounds
func (h *Handler) ListSystemSounds(c *fiber.Ctx) error {
	root := "/usr/share/freeswitch/sounds"
	tree, err := buildFileTree(root, root)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan sounds directory: " + err.Error()})
	}

	// Filter out "music" directory from sounds listing as it has its own endpoint
	var filteredChildren []*FileNode
	for _, child := range tree.Children {
		if child.Name != "music" {
			filteredChildren = append(filteredChildren, child)
		}
	}
	tree.Children = filteredChildren

	return c.JSON(fiber.Map{"data": tree.Children})
}

// ListSystemMusic returns the directory structure of /usr/share/freeswitch/sounds/music
func (h *Handler) ListSystemMusic(c *fiber.Ctx) error {
	root := "/usr/share/freeswitch/sounds/music"
	tree, err := buildFileTree(root, root)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan music directory: " + err.Error()})
	}
	return c.JSON(fiber.Map{"data": tree.Children})
}

// StreamSystemSound serves a system sound file for playback
func (h *Handler) StreamSystemSound(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Path is required"})
	}

	// Security: prevent directory traversal
	if strings.Contains(path, "..") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid path"})
	}

	fullPath := filepath.Join("/usr/share/freeswitch/sounds", path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	return c.SendFile(fullPath)
}

// StreamSystemMusic serves a system music file for playback
func (h *Handler) StreamSystemMusic(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Path is required"})
	}

	if strings.Contains(path, "..") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid path"})
	}

	fullPath := filepath.Join("/usr/share/freeswitch/sounds/music", path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	return c.SendFile(fullPath)
}

func buildFileTree(root, currentPath string) (*FileNode, error) {
	info, err := os.Stat(currentPath)
	if err != nil {
		return nil, err
	}

	relPath := strings.TrimPrefix(currentPath, root)
	relPath = strings.TrimPrefix(relPath, "/")

	node := &FileNode{
		Name: info.Name(),
		Path: relPath,
		Type: "file",
		Size: info.Size(),
	}

	if info.IsDir() {
		node.Type = "directory"
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			// Skip hidden files
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			childPath := filepath.Join(currentPath, entry.Name())
			child, err := buildFileTree(root, childPath)
			if err != nil {
				continue
			}
			node.Children = append(node.Children, child)
		}
	} else {
		// Only include audio files
		ext := strings.ToLower(filepath.Ext(node.Name))
		if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
			return nil, fmt.Errorf("skipped non-audio file")
		}
	}

	return node, nil
}

// generateUniqueFilename checks if a file exists and adds a random suffix if needed
func generateUniqueFilename(dir, filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	// Check if original filename exists
	origPath := filepath.Join(dir, filename)
	if _, err := os.Stat(origPath); os.IsNotExist(err) {
		return filename // Original name is available
	}

	// Generate a short random suffix
	b := make([]byte, 4)
	rand.Read(b)
	suffix := hex.EncodeToString(b)

	newName := fmt.Sprintf("%s_%s%s", base, suffix, ext)
	return newName
}

// UploadSystemSound handles uploading a sound file to a specific path
func (h *Handler) UploadSystemSound(c *fiber.Ctx) error {
	// Path should be relative to /usr/share/freeswitch/sounds (e.g., "en/us/callie/ivr")
	targetPath := c.FormValue("path")
	if targetPath == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Target path is required"})
	}

	// Prevent directory traversal
	if strings.Contains(targetPath, "..") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid path"})
	}

	header, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Failed to read file"})
	}
	file, err := header.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	// Validate extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type. Only .wav, .mp3, .ogg allowed"})
	}

	fullPath := filepath.Join("/usr/share/freeswitch/sounds", targetPath)

	// Ensure directory exists
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create directory"})
	}

	// Generate unique filename if one already exists
	finalFilename := generateUniqueFilename(fullPath, header.Filename)
	dstPath := filepath.Join(fullPath, finalFilename)
	out, err := os.Create(dstPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create destination file"})
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "File uploaded successfully", "path": filepath.Join(targetPath, finalFilename), "filename": finalFilename})
}

// UploadSystemMusic handles uploading a music file to a specific rate directory
func (h *Handler) UploadSystemMusic(c *fiber.Ctx) error {
	// Rate folder (e.g., "8000", "16000", "32000", "48000")
	rate := c.FormValue("rate")
	if rate != "8000" && rate != "16000" && rate != "32000" && rate != "48000" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid sample rate. Must be 8000, 16000, 32000, or 48000"})
	}

	header, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Failed to read file"})
	}
	file, err := header.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type. Only .wav, .mp3, .ogg allowed"})
	}

	// Optional folder/genre parameter
	folder := c.FormValue("folder")
	if folder != "" && strings.Contains(folder, "..") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid folder name"})
	}

	// Build path: /music/{rate} or /music/{rate}/{folder}
	fullPath := filepath.Join("/usr/share/freeswitch/sounds/music", rate)
	if folder != "" {
		fullPath = filepath.Join(fullPath, folder)
	}

	// Ensure directory exists
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create directory"})
	}

	// Generate unique filename if one already exists
	finalFilename := generateUniqueFilename(fullPath, header.Filename)
	dstPath := filepath.Join(fullPath, finalFilename)
	out, err := os.Create(dstPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create destination file"})
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	resultPath := filepath.Join(rate, finalFilename)
	if folder != "" {
		resultPath = filepath.Join(rate, folder, finalFilename)
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "File uploaded successfully", "path": resultPath, "filename": finalFilename})
}

// ListTenantSounds merges system sounds with tenant overrides
func (h *Handler) ListTenantSounds(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	systemRoot := "/usr/share/freeswitch/sounds"
	tenantRoot := fmt.Sprintf("/usr/share/freeswitch/sounds/tenants/%d", tenantID)

	// systemTree will be the base
	systemTree, err := buildFileTree(systemRoot, systemRoot)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan system sounds: " + err.Error()})
	}
	markSource(systemTree, "system")

	// Tenant tree (may not exist yet)
	tenantTree, err := buildFileTree(tenantRoot, tenantRoot)
	if err == nil && tenantTree != nil {
		markSource(tenantTree, "tenant")
		mergeTrees(systemTree, tenantTree)
	}

	// Filter out "music" directory and "tenants" directory from root if visible
	var filteredChildren []*FileNode
	for _, child := range systemTree.Children {
		if child.Name != "music" && child.Name != "tenants" {
			filteredChildren = append(filteredChildren, child)
		}
	}
	systemTree.Children = filteredChildren

	return c.JSON(fiber.Map{"data": systemTree.Children})
}

// Helper to recursively mark source on all nodes
func markSource(node *FileNode, source string) {
	node.Source = source
	for _, child := range node.Children {
		markSource(child, source)
	}
}

// mergeTrees overlays tenantNode onto systemNode
func mergeTrees(systemNode, tenantNode *FileNode) {
	// If matching file, tenant overrides system
	if systemNode.Type == "file" && tenantNode.Type == "file" {
		systemNode.Size = tenantNode.Size
		systemNode.Source = "tenant"
		systemNode.IsOverride = true
		return
	}

	// If directory, merge children
	if systemNode.Type == "directory" && tenantNode.Type == "directory" {
		// Map system children by name for quick lookup
		sysChildren := make(map[string]*FileNode)
		for _, child := range systemNode.Children {
			sysChildren[child.Name] = child
		}

		for _, tChild := range tenantNode.Children {
			if sChild, exists := sysChildren[tChild.Name]; exists {
				// Recursively merge
				mergeTrees(sChild, tChild)
			} else {
				// New tenant file/dir (shadowing nothing, just adding)
				systemNode.Children = append(systemNode.Children, tChild)
			}
		}

		// Sort children by name for consistent display
		sort.Slice(systemNode.Children, func(i, j int) bool {
			return systemNode.Children[i].Name < systemNode.Children[j].Name
		})
	}
}

// UploadTenantSound handles uploading a sound file to the tenant directory
func (h *Handler) UploadTenantSound(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	// Path relative to sounds root (e.g., "en/us/callie/ivr")
	targetPath := c.FormValue("path")
	if targetPath == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Target path is required"})
	}

	if strings.Contains(targetPath, "..") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid path"})
	}

	header, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Failed to read file"})
	}
	file, err := header.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type"})
	}

	// Save to /usr/share/freeswitch/sounds/tenants/{id}/{path}
	baseDir := fmt.Sprintf("/usr/share/freeswitch/sounds/tenants/%d", tenantID)
	fullDirPath := filepath.Join(baseDir, targetPath)

	if err := os.MkdirAll(fullDirPath, 0755); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create tenant directory"})
	}

	// Generate unique filename if one already exists
	finalFilename := generateUniqueFilename(fullDirPath, header.Filename)
	dstPath := filepath.Join(fullDirPath, finalFilename)
	out, err := os.Create(dstPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create destination file"})
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "File uploaded successfully", "path": filepath.Join(targetPath, finalFilename), "filename": finalFilename, "is_override": true})
}

// DeleteTenantSound removes a tenant sound/override
func (h *Handler) DeleteTenantSound(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	// Full logical path relative to sounds root, e.g. "en/us/callie/ivr/welcome.wav"
	targetPath := c.Query("path")
	if targetPath == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Path required"})
	}

	if strings.Contains(targetPath, "..") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid path"})
	}

	baseDir := fmt.Sprintf("/usr/share/freeswitch/sounds/tenants/%d", tenantID)
	fullPath := filepath.Join(baseDir, targetPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete file"})
	}

	return c.JSON(fiber.Map{"message": "File deleted, revert to system sound if applicable"})
}

// ListTenantMusic merges system music with tenant overrides
func (h *Handler) ListTenantMusic(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	systemRoot := "/usr/share/freeswitch/sounds/music"
	tenantRoot := fmt.Sprintf("/usr/share/freeswitch/sounds/music/tenants/%d", tenantID)

	systemTree, err := buildFileTree(systemRoot, systemRoot)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan system music: " + err.Error()})
	}
	markSource(systemTree, "system")

	tenantTree, err := buildFileTree(tenantRoot, tenantRoot)
	if err == nil && tenantTree != nil {
		markSource(tenantTree, "tenant")
		mergeTrees(systemTree, tenantTree)
	}

	// Unlike sounds, we don't need to filter specific dirs from music root usually
	// But let's filter "tenants" if it exists in root
	var filteredChildren []*FileNode
	for _, child := range systemTree.Children {
		if child.Name != "tenants" {
			filteredChildren = append(filteredChildren, child)
		}
	}
	systemTree.Children = filteredChildren

	return c.JSON(fiber.Map{"data": systemTree.Children})
}

// UploadTenantMusic handles uploading a music file to a specific rate directory for a tenant
func (h *Handler) UploadTenantMusic(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	// Rate folder (e.g., "8000", "16000", "32000", "48000")
	rate := c.FormValue("rate")
	if rate != "8000" && rate != "16000" && rate != "32000" && rate != "48000" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid sample rate"})
	}

	header, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Failed to read file"})
	}
	file, err := header.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type"})
	}

	// Optional folder/genre parameter
	folder := c.FormValue("folder")
	if folder != "" && strings.Contains(folder, "..") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid folder name"})
	}

	// Save to /usr/share/freeswitch/sounds/music/tenants/{id}/{rate} or {rate}/{folder}
	baseDir := fmt.Sprintf("/usr/share/freeswitch/sounds/music/tenants/%d", tenantID)
	fullRatePath := filepath.Join(baseDir, rate)
	if folder != "" {
		fullRatePath = filepath.Join(fullRatePath, folder)
	}

	if err := os.MkdirAll(fullRatePath, 0755); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create tenant directory"})
	}

	// Generate unique filename if one already exists
	finalFilename := generateUniqueFilename(fullRatePath, header.Filename)
	dstPath := filepath.Join(fullRatePath, finalFilename)
	out, err := os.Create(dstPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create destination file"})
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	resultPath := filepath.Join(rate, finalFilename)
	if folder != "" {
		resultPath = filepath.Join(rate, folder, finalFilename)
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "File uploaded successfully", "path": resultPath, "filename": finalFilename, "is_override": true})
}

// DeleteTenantMusic removes a tenant music file
func (h *Handler) DeleteTenantMusic(c *fiber.Ctx) error {
	tenantID := middleware.GetScopedTenantID(c)
	if tenantID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Tenant context required"})
	}

	// Path relative to music/tenant root (e.g. "8000/music.wav")
	targetPath := c.Query("path")
	if targetPath == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Path required"})
	}

	if strings.Contains(targetPath, "..") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid path"})
	}

	baseDir := fmt.Sprintf("/usr/share/freeswitch/sounds/music/tenants/%d", tenantID)
	fullPath := filepath.Join(baseDir, targetPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete file"})
	}

	return c.JSON(fiber.Map{"message": "File deleted"})
}
