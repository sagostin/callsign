package models

import (
	"fmt"

	"gorm.io/gorm"
)

// FeatureCodeModule represents a selectable group of related feature codes.
// During tenant setup, admins pick which modules to enable. Enabled modules
// have their feature codes written into the tenant's feature_codes table,
// and the dialplan is regenerated on the next FreeSWITCH fetch.
type FeatureCodeModule string

const (
	FCModuleVoicemail   FeatureCodeModule = "voicemail"
	FCModuleCallForward FeatureCodeModule = "call_forward"
	FCModuleDND         FeatureCodeModule = "dnd"
	FCModuleCallFlow    FeatureCodeModule = "call_flow"
	FCModuleParking     FeatureCodeModule = "parking"
	FCModulePickup      FeatureCodeModule = "pickup"
	FCModuleIntercom    FeatureCodeModule = "intercom"
	FCModulePaging      FeatureCodeModule = "paging"
	FCModuleRecording   FeatureCodeModule = "recording"
	FCModuleTransfer    FeatureCodeModule = "transfer"
	FCModuleSpeedDial   FeatureCodeModule = "speed_dial"
	FCModuleConference  FeatureCodeModule = "conference"
	FCModuleQueue       FeatureCodeModule = "queue"
)

// featureCodeTemplate defines a single code within a module
type featureCodeTemplate struct {
	Code        string
	CodeRegex   string // Optional regex for variable codes like *70XX
	Name        string
	Description string
	Action      FeatureCodeAction
	ActionData  string
	Order       int
}

// moduleDefinition groups templates under a module name
type moduleDefinition struct {
	Module      FeatureCodeModule
	Label       string // Human-readable name
	Description string
	Templates   []featureCodeTemplate
}

// allModules defines every available feature code module and its default codes.
// Tenants can enable any subset. Once provisioned, codes can be customized.
var allModules = []moduleDefinition{
	{
		Module:      FCModuleVoicemail,
		Label:       "Voicemail",
		Description: "Check voicemail, direct-to-voicemail deposit",
		Templates: []featureCodeTemplate{
			{Code: "*97", Name: "Check Own Voicemail", Action: FCActionVoicemail, ActionData: "check_own", Order: 10},
			{Code: "*98", Name: "Check Any Voicemail", Action: FCActionVoicemail, ActionData: "check_any", Order: 11},
		},
	},
	{
		Module:      FCModuleCallForward,
		Label:       "Call Forwarding",
		Description: "Enable/disable unconditional call forwarding",
		Templates: []featureCodeTemplate{
			{Code: "*72", Name: "Enable Call Forward", Action: FCActionCallForward, ActionData: "enable", Order: 20},
			{Code: "*73", Name: "Disable Call Forward", Action: FCActionCallForward, ActionData: "disable", Order: 21},
		},
	},
	{
		Module:      FCModuleDND,
		Label:       "Do Not Disturb",
		Description: "Toggle DND status with BLF presence update",
		Templates: []featureCodeTemplate{
			{Code: "*78", Name: "Enable DND", Action: FCActionDND, ActionData: "enable", Order: 30},
			{Code: "*79", Name: "Disable DND", Action: FCActionDND, ActionData: "disable", Order: 31},
		},
	},
	{
		Module:      FCModuleCallFlow,
		Label:       "Call Flow / Day-Night",
		Description: "Toggle day/night mode for call routing",
		Templates: []featureCodeTemplate{
			{Code: "*67", Name: "Toggle Call Flow", Action: FCActionCallFlowToggle, Order: 40},
		},
	},
	{
		Module:      FCModuleParking,
		Label:       "Call Parking",
		Description: "Valet park, directed park, and park retrieval",
		Templates: []featureCodeTemplate{
			{Code: "*56", Name: "Valet Park", Action: FCActionPark, Order: 50},
			{Code: "*57", CodeRegex: `^\*57(\d{2})$`, Name: "Park to Slot", Action: FCActionParkSlot, Order: 51},
			{Code: "*58", CodeRegex: `^\*58(\d{2})$`, Name: "Retrieve from Slot", Action: FCActionParkRetrieve, Order: 52},
		},
	},
	{
		Module:      FCModulePickup,
		Label:       "Call Pickup",
		Description: "Pick up a ringing call from a group or specific extension",
		Templates: []featureCodeTemplate{
			{Code: "*8", Name: "Group Pickup", Action: FCActionPickup, Order: 60},
			{Code: "**", CodeRegex: `^\*\*(\d+)$`, Name: "Directed Pickup", Action: FCActionPickup, ActionData: "directed", Order: 61},
		},
	},
	{
		Module:      FCModuleIntercom,
		Label:       "Intercom",
		Description: "Auto-answer intercom to a specific extension",
		Templates: []featureCodeTemplate{
			{Code: "*80", CodeRegex: `^\*80(\d+)$`, Name: "Intercom", Action: FCActionIntercom, Order: 70},
		},
	},
	{
		Module:      FCModulePaging,
		Label:       "Paging",
		Description: "Page a group of extensions",
		Templates: []featureCodeTemplate{
			{Code: "*81", Name: "Page Group", Action: FCActionPageGroup, Order: 80},
		},
	},
	{
		Module:      FCModuleRecording,
		Label:       "Call Recording",
		Description: "Toggle call recording for your extension",
		Templates: []featureCodeTemplate{
			{Code: "*82", Name: "Enable Recording", Action: FCActionRecord, ActionData: "enable", Order: 90},
			{Code: "*83", Name: "Disable Recording", Action: FCActionRecord, ActionData: "disable", Order: 91},
		},
	},
	{
		Module:      FCModuleTransfer,
		Label:       "Transfer",
		Description: "Blind and attended transfer",
		Templates: []featureCodeTemplate{
			{Code: "*1", Name: "Blind Transfer", Action: FCActionTransfer, ActionData: "blind", Order: 100},
			{Code: "*2", Name: "Attended Transfer", Action: FCActionTransfer, ActionData: "attended", Order: 101},
		},
	},
	{
		Module:      FCModuleSpeedDial,
		Label:       "Speed Dial",
		Description: "Single-digit speed dial codes",
		Templates: []featureCodeTemplate{
			{Code: "*0", CodeRegex: `^\*0(\d)$`, Name: "Speed Dial", Action: FCActionSpeedDial, Order: 110},
		},
	},
	{
		Module:      FCModuleConference,
		Label:       "Conference Bridges",
		Description: "Conference bridge access codes",
		Templates: []featureCodeTemplate{
			{Code: "*85", Name: "Start Conference", Action: FCActionCustom, ActionData: "conference", Order: 120},
		},
	},
	{
		Module:      FCModuleQueue,
		Label:       "Call Queue / Agent",
		Description: "Agent login/logout for call queues",
		Templates: []featureCodeTemplate{
			{Code: "*90", Name: "Agent Login", Action: FCActionCustom, ActionData: "agent_login", Order: 130},
			{Code: "*91", Name: "Agent Logout", Action: FCActionCustom, ActionData: "agent_logout", Order: 131},
		},
	},
}

// ========== Public API ==========

// FeatureCodeModuleInfo is returned by the API for module listing
type FeatureCodeModuleInfo struct {
	Module      FeatureCodeModule `json:"module"`
	Label       string            `json:"label"`
	Description string            `json:"description"`
	Enabled     bool              `json:"enabled"`
	CodeCount   int               `json:"code_count"` // Number of codes in this module
}

// ListAvailableModules returns all available feature code modules with their
// enabled status for a given tenant.
func ListAvailableModules(db *gorm.DB, tenantID uint) []FeatureCodeModuleInfo {
	// Get currently provisioned modules for this tenant
	enabledModules := make(map[string]bool)
	var codes []FeatureCode
	db.Where("tenant_id = ?", tenantID).Select("action, action_data").Find(&codes)

	// Determine which modules are enabled by checking if any code from
	// each module exists for this tenant
	for _, m := range allModules {
		for _, t := range m.Templates {
			for _, c := range codes {
				if c.Action == t.Action {
					enabledModules[string(m.Module)] = true
				}
			}
		}
	}

	var result []FeatureCodeModuleInfo
	for _, m := range allModules {
		result = append(result, FeatureCodeModuleInfo{
			Module:      m.Module,
			Label:       m.Label,
			Description: m.Description,
			Enabled:     enabledModules[string(m.Module)],
			CodeCount:   len(m.Templates),
		})
	}
	return result
}

// ProvisionFeatureCodes creates feature codes for the specified modules for a tenant.
// If the tenant already has codes for a module, those are skipped (not overwritten).
// Pass nil or empty for modules to provision ALL modules.
func ProvisionFeatureCodes(db *gorm.DB, tenantID uint, modules []FeatureCodeModule) error {
	// Build set of requested modules
	requested := make(map[FeatureCodeModule]bool)
	if len(modules) == 0 {
		// Provision all modules
		for _, m := range allModules {
			requested[m.Module] = true
		}
	} else {
		for _, m := range modules {
			requested[m] = true
		}
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, mod := range allModules {
			if !requested[mod.Module] {
				continue
			}

			for _, tmpl := range mod.Templates {
				// Skip if this code already exists for this tenant
				var count int64
				tx.Model(&FeatureCode{}).
					Where("tenant_id = ? AND code = ?", tenantID, tmpl.Code).
					Count(&count)
				if count > 0 {
					continue
				}

				fc := FeatureCode{
					TenantID:    &tenantID,
					Code:        tmpl.Code,
					CodeRegex:   tmpl.CodeRegex,
					Name:        tmpl.Name,
					Description: tmpl.Description,
					Action:      tmpl.Action,
					ActionData:  tmpl.ActionData,
					Order:       tmpl.Order,
					Enabled:     true,
				}
				if err := tx.Create(&fc).Error; err != nil {
					return fmt.Errorf("failed to create feature code %s for tenant %d: %w",
						tmpl.Code, tenantID, err)
				}
			}
		}
		return nil
	})
}

// DeprovisionFeatureCodes removes all feature codes for specified modules.
// Passing nil removes ALL feature codes for the tenant.
func DeprovisionFeatureCodes(db *gorm.DB, tenantID uint, modules []FeatureCodeModule) error {
	if len(modules) == 0 {
		// Remove all feature codes for tenant
		return db.Where("tenant_id = ?", tenantID).Delete(&FeatureCode{}).Error
	}

	// Collect all action types belonging to the requested modules
	actions := make(map[FeatureCodeAction]bool)
	for _, mod := range allModules {
		for _, m := range modules {
			if mod.Module == m {
				for _, t := range mod.Templates {
					actions[t.Action] = true
				}
			}
		}
	}

	var actionList []FeatureCodeAction
	for a := range actions {
		actionList = append(actionList, a)
	}

	return db.Where("tenant_id = ? AND action IN ?", tenantID, actionList).Delete(&FeatureCode{}).Error
}

// GetModuleForAction maps a feature code action to its parent module
func GetModuleForAction(action FeatureCodeAction) FeatureCodeModule {
	for _, mod := range allModules {
		for _, t := range mod.Templates {
			if t.Action == action {
				return mod.Module
			}
		}
	}
	return ""
}
