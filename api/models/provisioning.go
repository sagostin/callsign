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

// TenantDeletionPreview contains counts of all resources that will be deleted
type TenantDeletionPreview struct {
	Extensions          int64 `json:"extensions"`
	Users               int64 `json:"users"`
	Devices             int64 `json:"devices"`
	VoicemailBoxes      int64 `json:"voicemail_boxes"`
	VoicemailMessages   int64 `json:"voicemail_messages"`
	Recordings          int64 `json:"recordings"`
	CallRecordings      int64 `json:"call_recordings"`
	IVRMenus            int64 `json:"ivr_menus"`
	Queues              int64 `json:"queues"`
	RingGroups          int64 `json:"ring_groups"`
	Conferences         int64 `json:"conferences"`
	CallFlows           int64 `json:"call_flows"`
	TimeConditions      int64 `json:"time_conditions"`
	FeatureCodes        int64 `json:"feature_codes"`
	Dialplans           int64 `json:"dialplans"`
	PageGroups          int64 `json:"page_groups"`
	SpeedDialGroups     int64 `json:"speed_dial_groups"`
	Contacts            int64 `json:"contacts"`
	Conversations       int64 `json:"conversations"`
	Messages            int64 `json:"messages"`
	Broadcasts          int64 `json:"broadcasts"`
	FaxBoxes            int64 `json:"fax_boxes"`
	FaxJobs             int64 `json:"fax_jobs"`
	HotelRooms          int64 `json:"hotel_rooms"`
	ChatThreads         int64 `json:"chat_threads"`
	CDRRecords          int64 `json:"cdr_records"`
	CallBlocks          int64 `json:"call_blocks"`
	HolidayLists        int64 `json:"holiday_lists"`
	Locations           int64 `json:"locations"`
	MediaFiles          int64 `json:"media_files"`
	AuditLogs           int64 `json:"audit_logs"`
	ClientRegistrations int64 `json:"client_registrations"`
}

// PreviewTenantDeletion returns counts of all resources belonging to a tenant
func PreviewTenantDeletion(db *gorm.DB, tenantID uint) TenantDeletionPreview {
	var p TenantDeletionPreview
	db.Model(&Extension{}).Where("tenant_id = ?", tenantID).Count(&p.Extensions)
	db.Model(&User{}).Where("tenant_id = ?", tenantID).Count(&p.Users)
	db.Model(&Device{}).Where("tenant_id = ?", tenantID).Count(&p.Devices)
	db.Model(&VoicemailBox{}).Where("tenant_id = ?", tenantID).Count(&p.VoicemailBoxes)
	db.Model(&VoicemailMessage{}).Where("tenant_id = ?", tenantID).Count(&p.VoicemailMessages)
	db.Model(&Recording{}).Where("tenant_id = ?", tenantID).Count(&p.Recordings)
	db.Model(&CallRecording{}).Where("tenant_id = ?", tenantID).Count(&p.CallRecordings)
	db.Model(&IVRMenu{}).Where("tenant_id = ?", tenantID).Count(&p.IVRMenus)
	db.Model(&Queue{}).Where("tenant_id = ?", tenantID).Count(&p.Queues)
	db.Model(&RingGroup{}).Where("tenant_id = ?", tenantID).Count(&p.RingGroups)
	db.Model(&Conference{}).Where("tenant_id = ?", tenantID).Count(&p.Conferences)
	db.Model(&CallFlow{}).Where("tenant_id = ?", tenantID).Count(&p.CallFlows)
	db.Model(&TimeCondition{}).Where("tenant_id = ?", tenantID).Count(&p.TimeConditions)
	db.Model(&FeatureCode{}).Where("tenant_id = ?", tenantID).Count(&p.FeatureCodes)
	db.Model(&Dialplan{}).Where("tenant_id = ?", tenantID).Count(&p.Dialplans)
	db.Model(&PageGroup{}).Where("tenant_id = ?", tenantID).Count(&p.PageGroups)
	db.Model(&SpeedDialGroup{}).Where("tenant_id = ?", tenantID).Count(&p.SpeedDialGroups)
	db.Model(&Contact{}).Where("tenant_id = ?", tenantID).Count(&p.Contacts)
	db.Model(&Conversation{}).Where("tenant_id = ?", tenantID).Count(&p.Conversations)
	db.Model(&Message{}).Where("tenant_id = ?", tenantID).Count(&p.Messages)
	db.Model(&BroadcastCampaign{}).Where("tenant_id = ?", tenantID).Count(&p.Broadcasts)
	db.Model(&FaxBox{}).Where("tenant_id = ?", tenantID).Count(&p.FaxBoxes)
	db.Model(&FaxJob{}).Where("tenant_id = ?", tenantID).Count(&p.FaxJobs)
	db.Model(&HotelRoom{}).Where("tenant_id = ?", tenantID).Count(&p.HotelRooms)
	db.Model(&ChatThread{}).Where("tenant_id = ?", tenantID).Count(&p.ChatThreads)
	db.Model(&CallRecord{}).Where("tenant_id = ?", tenantID).Count(&p.CDRRecords)
	db.Model(&CallBlock{}).Where("tenant_id = ?", tenantID).Count(&p.CallBlocks)
	db.Model(&HolidayList{}).Where("tenant_id = ?", tenantID).Count(&p.HolidayLists)
	db.Model(&Location{}).Where("tenant_id = ?", tenantID).Count(&p.Locations)
	db.Model(&MediaFile{}).Where("tenant_id = ?", tenantID).Count(&p.MediaFiles)
	db.Model(&AuditLog{}).Where("tenant_id = ?", tenantID).Count(&p.AuditLogs)
	db.Model(&ClientRegistration{}).Where("tenant_id = ?", tenantID).Count(&p.ClientRegistrations)
	return p
}

// DeprovisionTenant removes ALL tenant-specific resources.
// Call this before deleting the tenant row.
func DeprovisionTenant(db *gorm.DB, tenantID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		log.WithField("tenant_id", tenantID).Info("Deprovisioning all resources for tenant")

		// Delete dialplan details for this tenant's dialplans first (FK dependency)
		var dialplanUUIDs []string
		tx.Model(&Dialplan{}).Where("tenant_id = ?", tenantID).Pluck("uuid", &dialplanUUIDs)
		if len(dialplanUUIDs) > 0 {
			if err := tx.Where("dialplan_uuid IN ?", dialplanUUIDs).Delete(&DialplanDetail{}).Error; err != nil {
				log.Warnf("Failed to delete dialplan details for tenant %d: %v", tenantID, err)
			}
		}

		// Delete all tenant-scoped resources in dependency order.
		// Child records (FK dependencies) are deleted before parents.
		deprovisionModels := []struct {
			name  string
			model interface{}
		}{
			// --- FreeSWITCH provisioning ---
			{"park_slots", &ParkSlot{}},
			{"feature_codes", &FeatureCode{}},
			{"extension_presences", &ExtensionPresence{}},
			// --- Dialplan & routing ---
			{"dialplans", &Dialplan{}},
			{"destinations", &Destination{}},
			{"call_blocks", &CallBlock{}},
			// --- Voicemail ---
			{"voicemail_messages", &VoicemailMessage{}},
			{"voicemail_boxes", &VoicemailBox{}},
			// --- Ring groups ---
			{"ring_group_destinations", &RingGroupDestination{}},
			{"ring_groups", &RingGroup{}},
			// --- Queues ---
			{"queue_agents", &QueueAgent{}},
			{"queues", &Queue{}},
			// --- IVR ---
			{"ivr_menu_options", &IVRMenuOption{}},
			{"ivr_menus", &IVRMenu{}},
			// --- Time / schedule ---
			{"time_conditions", &TimeCondition{}},
			{"call_flow_destinations", &CallFlowDestination{}},
			{"call_flows", &CallFlow{}},
			{"holiday_lists", &HolidayList{}},
			// --- Conferences ---
			{"conference_participants", &ConferenceParticipant{}},
			{"conference_sessions", &ConferenceSession{}},
			{"conference_members", &ConferenceMember{}},
			{"conferences", &Conference{}},
			// --- Recordings ---
			{"transcription_segments", &TranscriptionSegment{}},
			{"transcriptions", &Transcription{}},
			{"call_recordings", &CallRecording{}},
			{"recordings", &Recording{}},
			// --- Devices ---
			{"device_lines", &DeviceLine{}},
			{"devices", &Device{}},
			{"device_profiles", &DeviceProfile{}},
			{"client_registrations", &ClientRegistration{}},
			// --- Contact & messaging ---
			{"contacts", &Contact{}},
			{"message_media", &MessageMedia{}},
			{"messages", &Message{}},
			{"conversations", &Conversation{}},
			{"sms_number_assignments", &SMSNumberAssignment{}},
			{"message_queue_items", &MessageQueueItem{}},
			// --- Chat ---
			{"chat_messages", &ChatMessage{}},
			{"chat_room_members", &ChatRoomMember{}},
			{"chat_threads", &ChatThread{}},
			{"chat_rooms", &ChatRoom{}},
			{"chat_queues", &ChatQueue{}},
			// --- Paging ---
			{"page_group_destinations", &PageGroupDestination{}},
			{"page_groups", &PageGroup{}},
			// --- Speed dials ---
			{"speed_dial_entries", &SpeedDialEntry{}},
			{"speed_dial_groups", &SpeedDialGroup{}},
			// --- Fax ---
			{"fax_jobs", &FaxJob{}},
			{"fax_endpoints", &FaxEndpoint{}},
			{"fax_boxes", &FaxBox{}},
			// --- Hospitality ---
			{"hotel_rooms", &HotelRoom{}},
			// --- Broadcasts ---
			{"broadcasts", &BroadcastCampaign{}},
			// --- CDR / Audit ---
			{"cdr_records", &CallRecord{}},
			{"audit_logs", &AuditLog{}},
			// --- Media ---
			{"media_files", &MediaFile{}},
			// --- Provisioning ---
			{"provisioning_variables", &ProvisioningVariable{}},
			{"provisioning_templates", &ProvisioningTemplate{}},
			// --- Profiles ---
			{"call_handling_rules", &CallHandlingRule{}},
			{"extension_profiles", &ExtensionProfile{}},
			// --- Locations ---
			{"locations", &Location{}},
			// --- Extensions (last since many things reference them) ---
			{"extensions", &Extension{}},
			// --- Users ---
			{"users", &User{}},
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
