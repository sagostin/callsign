package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"net/http"

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
			activeCalls = parseFreeSwitchCalls(result)
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
			queueSummary = parseQueueList(result)
		}
	}

	return c.JSON(fiber.Map{
		"extensions":   panelData,
		"active_calls": activeCalls,
		"queues":       queueSummary,
	})
}

// parseFreeSwitchCalls parses the "show calls as json" output from FreeSWITCH
func parseFreeSwitchCalls(result string) []map[string]interface{} {
	return []map[string]interface{}{}
}

// parseQueueList parses callcenter queue list output
func parseQueueList(result string) []map[string]interface{} {
	return []map[string]interface{}{}
}
