package handlers

import (
	"callsign/middleware"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kataras/iris/v12"
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
func (h *Handler) ListSystemSounds(ctx iris.Context) {
	root := "/usr/share/freeswitch/sounds"
	tree, err := buildFileTree(root, root)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to scan sounds directory: " + err.Error()})
		return
	}

	// Filter out "music" directory from sounds listing as it has its own endpoint
	var filteredChildren []*FileNode
	for _, child := range tree.Children {
		if child.Name != "music" {
			filteredChildren = append(filteredChildren, child)
		}
	}
	tree.Children = filteredChildren

	ctx.JSON(iris.Map{"data": tree.Children})
}

// ListSystemMusic returns the directory structure of /usr/share/freeswitch/sounds/music
func (h *Handler) ListSystemMusic(ctx iris.Context) {
	root := "/usr/share/freeswitch/sounds/music"
	tree, err := buildFileTree(root, root)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to scan music directory: " + err.Error()})
		return
	}
	ctx.JSON(iris.Map{"data": tree.Children})
}

func buildFileTree(root, currentPath string) (*FileNode, error) {
	info, err := os.Stat(currentPath)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		Name: info.Name(),
		Path: strings.TrimPrefix(currentPath, root),
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

// UploadSystemSound handles uploading a sound file to a specific path
func (h *Handler) UploadSystemSound(ctx iris.Context) {
	// Path should be relative to /usr/share/freeswitch/sounds (e.g., "en/us/callie/ivr")
	targetPath := ctx.FormValue("path")
	if targetPath == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Target path is required"})
		return
	}

	// Prevent directory traversal
	if strings.Contains(targetPath, "..") {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid path"})
		return
	}

	file, header, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Failed to read file"})
		return
	}
	defer file.Close()

	// Validate extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid file type. Only .wav, .mp3, .ogg allowed"})
		return
	}

	fullPath := filepath.Join("/usr/share/freeswitch/sounds", targetPath)

	// Ensure directory exists
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create directory"})
		return
	}

	dstPath := filepath.Join(fullPath, header.Filename)
	out, err := os.Create(dstPath)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create destination file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save file"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "File uploaded successfully", "path": filepath.Join(targetPath, header.Filename)})
}

// UploadSystemMusic handles uploading a music file to a specific rate directory
func (h *Handler) UploadSystemMusic(ctx iris.Context) {
	// Rate folder (e.g., "8000", "16000", "32000", "48000")
	rate := ctx.FormValue("rate")
	if rate != "8000" && rate != "16000" && rate != "32000" && rate != "48000" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid sample rate. Must be 8000, 16000, 32000, or 48000"})
		return
	}

	file, header, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Failed to read file"})
		return
	}
	defer file.Close()

	// Validate extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid file type. Only .wav, .mp3, .ogg allowed"})
		return
	}

	fullPath := filepath.Join("/usr/share/freeswitch/sounds/music", rate)

	// Ensure directory exists
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create directory"})
		return
	}

	dstPath := filepath.Join(fullPath, header.Filename)
	out, err := os.Create(dstPath)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create destination file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save file"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "File uploaded successfully", "path": filepath.Join(rate, header.Filename)})
}

// ListTenantSounds merges system sounds with tenant overrides
func (h *Handler) ListTenantSounds(ctx iris.Context) {
	tenantID := middleware.GetScopedTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Tenant context required"})
		return
	}

	systemRoot := "/usr/share/freeswitch/sounds"
	tenantRoot := fmt.Sprintf("/usr/share/freeswitch/sounds/tenants/%d", tenantID)

	// systemTree will be the base
	systemTree, err := buildFileTree(systemRoot, systemRoot)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to scan system sounds: " + err.Error()})
		return
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

	ctx.JSON(iris.Map{"data": systemTree.Children})
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
func (h *Handler) UploadTenantSound(ctx iris.Context) {
	tenantID := middleware.GetScopedTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Tenant context required"})
		return
	}

	// Path relative to sounds root (e.g., "en/us/callie/ivr")
	targetPath := ctx.FormValue("path")
	if targetPath == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Target path is required"})
		return
	}

	if strings.Contains(targetPath, "..") {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid path"})
		return
	}

	file, header, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Failed to read file"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid file type"})
		return
	}

	// Save to /usr/share/freeswitch/sounds/tenants/{id}/{path}
	baseDir := fmt.Sprintf("/usr/share/freeswitch/sounds/tenants/%d", tenantID)
	fullDirPath := filepath.Join(baseDir, targetPath)

	if err := os.MkdirAll(fullDirPath, 0755); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create tenant directory"})
		return
	}

	dstPath := filepath.Join(fullDirPath, header.Filename)
	out, err := os.Create(dstPath)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create destination file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save file"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "File uploaded successfully", "path": filepath.Join(targetPath, header.Filename), "is_override": true})
}

// DeleteTenantSound removes a tenant sound/override
func (h *Handler) DeleteTenantSound(ctx iris.Context) {
	tenantID := middleware.GetScopedTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Tenant context required"})
		return
	}

	// Full logical path relative to sounds root, e.g. "en/us/callie/ivr/welcome.wav"
	targetPath := ctx.URLParam("path")
	if targetPath == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Path required"})
		return
	}

	if strings.Contains(targetPath, "..") {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid path"})
		return
	}

	baseDir := fmt.Sprintf("/usr/share/freeswitch/sounds/tenants/%d", tenantID)
	fullPath := filepath.Join(baseDir, targetPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(iris.Map{"error": "File not found"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete file"})
		return
	}

	ctx.JSON(iris.Map{"message": "File deleted, revert to system sound if applicable"})
}

// ListTenantMusic merges system music with tenant overrides
func (h *Handler) ListTenantMusic(ctx iris.Context) {
	tenantID := middleware.GetScopedTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Tenant context required"})
		return
	}

	systemRoot := "/usr/share/freeswitch/sounds/music"
	tenantRoot := fmt.Sprintf("/usr/share/freeswitch/sounds/music/tenants/%d", tenantID)

	systemTree, err := buildFileTree(systemRoot, systemRoot)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to scan system music: " + err.Error()})
		return
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

	ctx.JSON(iris.Map{"data": systemTree.Children})
}

// UploadTenantMusic handles uploading a music file to a specific rate directory for a tenant
func (h *Handler) UploadTenantMusic(ctx iris.Context) {
	tenantID := middleware.GetScopedTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Tenant context required"})
		return
	}

	// Rate folder (e.g., "8000", "16000", "32000", "48000")
	rate := ctx.FormValue("rate")
	if rate != "8000" && rate != "16000" && rate != "32000" && rate != "48000" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid sample rate"})
		return
	}

	file, header, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Failed to read file"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".wav" && ext != ".mp3" && ext != ".ogg" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid file type"})
		return
	}

	// Save to /usr/share/freeswitch/sounds/music/tenants/{id}/{rate}
	baseDir := fmt.Sprintf("/usr/share/freeswitch/sounds/music/tenants/%d", tenantID)
	fullRatePath := filepath.Join(baseDir, rate)

	if err := os.MkdirAll(fullRatePath, 0755); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create tenant directory"})
		return
	}

	dstPath := filepath.Join(fullRatePath, header.Filename)
	out, err := os.Create(dstPath)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create destination file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save file"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(iris.Map{"message": "File uploaded successfully", "path": filepath.Join(rate, header.Filename), "is_override": true})
}

// DeleteTenantMusic removes a tenant music file
func (h *Handler) DeleteTenantMusic(ctx iris.Context) {
	tenantID := middleware.GetScopedTenantID(ctx)
	if tenantID == 0 {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Tenant context required"})
		return
	}

	// Path relative to music/tenant root (e.g. "8000/music.wav")
	targetPath := ctx.URLParam("path")
	if targetPath == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Path required"})
		return
	}

	if strings.Contains(targetPath, "..") {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid path"})
		return
	}

	baseDir := fmt.Sprintf("/usr/share/freeswitch/sounds/music/tenants/%d", tenantID)
	fullPath := filepath.Join(baseDir, targetPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(iris.Map{"error": "File not found"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete file"})
		return
	}

	ctx.JSON(iris.Map{"message": "File deleted"})
}
