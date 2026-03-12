package models

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ProvisionTenant creates all default FreeSWITCH resources for a new tenant.
// This should be called after the Tenant row is created.
// It clones global feature codes, creates parking slots, and sets up a default
// internal dialplan so extension-to-extension calls work out of the box.
func ProvisionTenant(db *gorm.DB, tenant *Tenant) error {
	return db.Transaction(func(tx *gorm.DB) error {
		log.WithFields(log.Fields{
			"tenant_id": tenant.ID,
			"domain":    tenant.Domain,
		}).Info("Provisioning FreeSWITCH resources for tenant")

		// 1. Provision all feature code modules for this tenant
		if err := ProvisionFeatureCodes(tx, tenant.ID, nil); err != nil {
			return fmt.Errorf("provision feature codes: %w", err)
		}

		// 2. Create default parking slots
		if err := EnsureParkSlots(tx, tenant.ID, tenant.Domain, 10, 1); err != nil {
			return fmt.Errorf("create park slots: %w", err)
		}

		// 3. Create default internal dialplan for local extension routing
		if err := createDefaultInternalDialplan(tx, tenant); err != nil {
			return fmt.Errorf("create internal dialplan: %w", err)
		}

		// 4. Create default voicemail routing dialplan
		if err := createDefaultVoicemailDialplan(tx, tenant); err != nil {
			return fmt.Errorf("create voicemail dialplan: %w", err)
		}

		log.WithField("tenant_id", tenant.ID).Info("Tenant provisioning completed")
		return nil
	})
}

// DeprovisionTenant removes all tenant-specific FreeSWITCH resources.
// Call this before deleting the tenant row.
func DeprovisionTenant(db *gorm.DB, tenantID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		log.WithField("tenant_id", tenantID).Info("Deprovisioning FreeSWITCH resources for tenant")

		// Delete dialplan details for this tenant's dialplans first (FK dependency)
		var dialplanUUIDs []string
		tx.Model(&Dialplan{}).Where("tenant_id = ?", tenantID).Pluck("uuid", &dialplanUUIDs)
		if len(dialplanUUIDs) > 0 {
			if err := tx.Where("dialplan_uuid IN ?", dialplanUUIDs).Delete(&DialplanDetail{}).Error; err != nil {
				log.Warnf("Failed to delete dialplan details for tenant %d: %v", tenantID, err)
			}
		}

		// Delete tenant-scoped resources in dependency order
		deprovisionModels := []struct {
			name  string
			model interface{}
		}{
			{"park_slots", &ParkSlot{}},
			{"feature_codes", &FeatureCode{}},
			{"extension_presences", &ExtensionPresence{}},
			{"dialplans", &Dialplan{}},
			{"destinations", &Destination{}},
			{"voicemail_messages", &VoicemailMessage{}},
			{"voicemail_boxes", &VoicemailBox{}},
			{"ring_group_destinations", &RingGroupDestination{}},
			{"ring_groups", &RingGroup{}},
			{"queue_agents", &QueueAgent{}},
			{"queues", &Queue{}},
			{"ivr_menu_options", &IVRMenuOption{}},
			{"ivr_menus", &IVRMenu{}},
			{"conferences", &Conference{}},
			{"time_conditions", &TimeCondition{}},
			{"call_flows", &CallFlow{}},
			{"extensions", &Extension{}},
		}

		for _, tbl := range deprovisionModels {
			result := tx.Where("tenant_id = ?", tenantID).Delete(tbl.model)
			if result.Error != nil {
				log.Warnf("Failed to delete %s for tenant %d: %v", tbl.name, tenantID, result.Error)
			} else if result.RowsAffected > 0 {
				log.Debugf("Deleted %d %s for tenant %d", result.RowsAffected, tbl.name, tenantID)
			}
		}

		log.WithField("tenant_id", tenantID).Info("Tenant deprovisioning completed")
		return nil
	})
}

// createDefaultInternalDialplan creates a dialplan entry for the tenant's
// domain context that routes extension-to-extension calls.
func createDefaultInternalDialplan(tx *gorm.DB, tenant *Tenant) error {
	dp := &Dialplan{
		TenantID:        &tenant.ID,
		DialplanName:    "Local Extension",
		DialplanContext: tenant.Domain,
		Description:     "Routes calls between local extensions",
		Enabled:         true,
		DialplanOrder:   900, // High order = low priority (after feature codes, queues, etc.)
		Continue:        false,
		DialplanXML: fmt.Sprintf(`<extension name="local_extension_%s" continue="false">
  <condition field="destination_number" expression="^(\d{3,5})$">
    <action application="set" data="hangup_after_bridge=true"/>
    <action application="set" data="call_timeout=30"/>
    <action application="set" data="dialed_extension=$1"/>
    <action application="export" data="dialed_extension=$1"/>
    <action application="bridge" data="user/$1@%s"/>
  </condition>
</extension>`, tenant.Domain, tenant.Domain),
	}

	return tx.Create(dp).Error
}

// createDefaultVoicemailDialplan creates the default voicemail routing for
// a tenant — routes to the Go voicemail ESL service.
func createDefaultVoicemailDialplan(tx *gorm.DB, tenant *Tenant) error {
	dp := &Dialplan{
		TenantID:        &tenant.ID,
		DialplanName:    "Voicemail",
		DialplanContext: tenant.Domain,
		Description:     "Routes to voicemail service for unanswered calls",
		Enabled:         true,
		DialplanOrder:   950,
		Continue:        false,
		DialplanXML: `<extension name="voicemail" continue="false">
  <condition field="destination_number" expression="^vmain$">
    <action application="answer"/>
    <action application="sleep" data="1000"/>
    <action application="socket" data="127.0.0.2:9001 async full"/>
  </condition>
</extension>`,
	}

	return tx.Create(dp).Error
}
