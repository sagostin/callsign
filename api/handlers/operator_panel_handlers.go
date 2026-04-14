package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// =====================
// Operator Panel
// =====================

// GetOperatorPanelData returns real-time extension states and active calls
// for the operator panel view
func (h *Handler) GetOperatorPanelData(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	// Get all extensions with their presence state
	var extensions []models.Extension
	if err := h.DB.Where("tenant_id = ?", tenantID).
		Order("extension ASC").Find(&extensions).Error; err != nil {
		h.logError("OPERATOR", "GetOperatorPanelData: Failed to retrieve extensions", h.reqFields(c, nil))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve extensions"})
	}

	// Get active call data from ESL if available
	var activeCalls []map[string]interface{}
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		result, err := h.ESLManager.Client.API("show calls as json")
		if err == nil && result != "" {
			activeCalls, _ = parseFreeSwitchCalls(result)
		}
	}

	// Get presence states from DB
	var presenceStates []models.ExtensionPresence
	h.DB.Where("tenant_id = ?", tenantID).Find(&presenceStates)

	// Build presence map
	presenceMap := make(map[string]models.PresenceState)
	for _, p := range presenceStates {
		presenceMap[p.Extension] = p.State
	}

	// Build extension panel data
	type PanelExtension struct {
		ID        uint                 `json:"id"`
		Extension string               `json:"extension"`
		Name      string               `json:"name"`
		CallerID  string               `json:"caller_id"`
		Enabled   bool                 `json:"enabled"`
		Presence  models.PresenceState `json:"presence"`
	}

	panelData := make([]PanelExtension, len(extensions))
	for i, ext := range extensions {
		presence := presenceMap[ext.Extension]
		if presence == "" {
			presence = models.PresenceOffline
		}
		panelData[i] = PanelExtension{
			ID:        ext.ID,
			Extension: ext.Extension,
			Name:      ext.EffectiveCallerIDName,
			CallerID:  ext.EffectiveCallerIDNumber,
			Enabled:   ext.Enabled,
			Presence:  presence,
		}
	}

	// Get queue summary if available
	var queueSummary []map[string]interface{}
	if h.ESLManager != nil && h.ESLManager.IsConnected() {
		result, err := h.ESLManager.Client.API("callcenter_config queue list")
		if err == nil && result != "" {
			queueSummary, _ = parseQueueList(result)
		}
	}

	return c.JSON(fiber.Map{
		"extensions":   panelData,
		"active_calls": activeCalls,
		"queues":       queueSummary,
	})
}

// parseFreeSwitchCalls parses the "show calls as json" output from FreeSWITCH
func parseFreeSwitchCalls(data string) ([]map[string]interface{}, error) {
	if data == "" {
		return []map[string]interface{}{}, nil
	}

	var calls []map[string]interface{}
	if err := json.Unmarshal([]byte(data), &calls); err != nil {
		return []map[string]interface{}{}, nil
	}
	return calls, nil
}

// parseQueueList parses callcenter queue list output
func parseQueueList(data string) ([]map[string]interface{}, error) {
	if data == "" {
		return []map[string]interface{}{}, nil
	}

	lines := strings.Split(data, "\n")
	var results []map[string]interface{}

	for _, line := range lines {
		// Skip headers and separators
		if strings.HasPrefix(line, " queues:") || strings.HasPrefix(line, "+--") || strings.HasPrefix(line, "| Name") {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "|") {
			continue
		}

		// Parse pipe-separated fields
		fields := strings.Split(line, "|")
		if len(fields) < 9 {
			continue
		}

		// Trim spaces from each field
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}

		queue := map[string]interface{}{
			"name":      fields[1],
			"strategy":  fields[2],
			"max_wait":  fields[3],
			"total_act": fields[4],
			"ram_act":   fields[5],
			"ram_queue": fields[6],
			"waiting":   fields[7],
			"calls":     fields[8],
		}
		results = append(results, queue)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}
