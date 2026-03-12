package conference

import (
	"callsign/models"
	"callsign/services/esl"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	ServiceName    = "conference"
	ServiceAddress = "127.0.0.4:9001"
)

// Service implements the conference room ESL module with control capabilities
type Service struct {
	*esl.BaseService
	db       *gorm.DB
	sessions sync.Map // map[string]*models.ConferenceSession - active sessions by name
}

// New creates a new conference service
func New(db *gorm.DB) *Service {
	return &Service{
		BaseService: esl.NewBaseService(ServiceName, ServiceAddress),
		db:          db,
	}
}

// Init initializes the conference service
func (s *Service) Init(manager *esl.Manager) error {
	if err := s.BaseService.Init(manager); err != nil {
		return err
	}
	log.Info("Conference service initialized with control capabilities")
	return nil
}

// Handle processes incoming conference connections
func (s *Service) Handle(conn *eventsocket.Connection) {
	defer conn.Close()

	manager := s.Manager()
	if manager == nil {
		log.Error("Conference: manager not initialized")
		return
	}

	// Connect and get channel info
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Conference: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	callerName := ev.Get("Caller-Caller-ID-Name")
	conferenceNum := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")
	tenantIDStr := ev.Get("variable_tenant_id")

	logger := log.WithFields(log.Fields{
		"uuid":       uuid,
		"caller":     callerID,
		"conference": conferenceNum,
		"domain":     domain,
	})
	logger.Info("Conference: incoming participant")

	conn.Send("linger")
	conn.Send("myevents")

	// Answer the call
	conn.Execute("answer", "", true)

	// Look up conference config from DB
	var conf models.Conference
	db := manager.DB
	if err := db.Where("extension = ? AND enabled = ?", conferenceNum, true).First(&conf).Error; err != nil {
		logger.Warnf("Conference not found for extension %s", conferenceNum)
		conn.Execute("playback", "ivr/ivr-call_cannot_be_completed_as_dialed.wav", true)
		conn.Execute("hangup", "", false)
		return
	}

	// ---------- Max Participant Enforcement ----------
	confName := fmt.Sprintf("%s-%s@default", domain, conferenceNum)
	if conf.MaxMembers > 0 {
		currentCount := s.getConferenceMemberCount(confName)
		if currentCount >= conf.MaxMembers {
			logger.Infof("Conference full (%d/%d)", currentCount, conf.MaxMembers)
			conn.Execute("playback", "ivr/ivr-conference_is_full.wav", true)
			conn.Execute("hangup", "NORMAL_CLEARING", false)
			return
		}
	}

	// ---------- PIN Authentication ----------
	isModerator := false
	if conf.PIN != "" || conf.ModeratorPIN != "" {
		authenticated := false
		for attempt := 0; attempt < 3; attempt++ {
			// Collect PIN
			conn.Execute("play_and_get_digits",
				"4 10 1 5000 # conference/conf-pin.wav ivr/ivr-that_was_an_invalid_entry.wav pin \\d+ 5000",
				true)

			ev, err := conn.Send("api uuid_getvar " + uuid + " pin")
			if err != nil {
				break
			}
			enteredPIN := strings.TrimSpace(ev.Body)
			if enteredPIN == "" || enteredPIN == "_undef_" {
				continue
			}

			// Check moderator PIN first
			if conf.ModeratorPIN != "" && enteredPIN == conf.ModeratorPIN {
				isModerator = true
				authenticated = true
				logger.Info("Conference: authenticated as moderator")
				break
			}

			// Check participant PIN
			if conf.PIN != "" && enteredPIN == conf.PIN {
				authenticated = true
				logger.Info("Conference: authenticated as participant")
				break
			}
		}

		if !authenticated {
			logger.Warn("Conference: PIN auth failed")
			conn.Execute("playback", "ivr/ivr-not_authorized.wav", true)
			conn.Execute("hangup", "NORMAL_CLEARING", false)
			return
		}
	}

	// Get or create session
	tenantID, _ := strconv.ParseUint(tenantIDStr, 10, 32)
	session := s.getOrCreateSession(confName, uint(tenantID))

	// ---------- Build Conference Flags ----------
	var flags []string
	if isModerator {
		flags = append(flags, "moderator")
	}
	if conf.MuteOnJoin && !isModerator {
		flags = append(flags, "mute")
	}
	if conf.WaitForModerator && !isModerator {
		flags = append(flags, "wait-mod")
	}

	// Set caller info
	conn.Execute("set", "conference_member_nospeak_relax=true", true)
	conn.Execute("set", fmt.Sprintf("effective_caller_id_name=%s", callerName), true)
	conn.Execute("set", fmt.Sprintf("effective_caller_id_number=%s", callerID), true)

	if isModerator {
		conn.Execute("set", "conference_member_flags=moderator", true)
	}
	if conf.MuteOnJoin && !isModerator {
		conn.Execute("set", "conference_member_flags=mute", true)
	}

	logger.Infof("Joining conference: %s (moderator=%v)", confName, isModerator)

	// ---------- Auto-Recording ----------
	if conf.RecordConference {
		recordPath := fmt.Sprintf("/var/lib/freeswitch/recordings/%s/conference_%s_%s.wav",
			domain, conferenceNum, time.Now().Format("20060102_150405"))
		conn.Execute("set", fmt.Sprintf("conference_auto_record=%s", recordPath), true)
		s.db.Model(session).Updates(map[string]interface{}{
			"recording":      true,
			"recording_path": recordPath,
		})
	}

	// Join the conference (blocks until hangup)
	confArg := confName
	if conf.ProfileName != "" && conf.ProfileName != "default" {
		confArg = fmt.Sprintf("%s@%s", confName, conf.ProfileName)
	}
	conn.Execute("conference", confArg, true)

	// Wait for hangup and track events
	for {
		ev, err := conn.ReadEvent()
		if err != nil {
			break
		}
		eventName := ev.Get("Event-Name")

		switch eventName {
		case "CHANNEL_HANGUP_COMPLETE":
			s.handleParticipantLeave(session, uuid)
			logger.Info("Conference: participant left")
			return
		case "CUSTOM":
			subclass := ev.Get("Event-Subclass")
			if strings.HasPrefix(subclass, "conference::") {
				s.handleConferenceEvent(session, ev)
			}
		}
	}
}

// getConferenceMemberCount returns the current number of participants
func (s *Service) getConferenceMemberCount(confName string) int {
	manager := s.Manager()
	if manager == nil || manager.Client == nil {
		return 0
	}

	result, err := manager.Client.API(fmt.Sprintf("conference %s list count", confName))
	if err != nil {
		return 0
	}

	count := 0
	fmt.Sscanf(strings.TrimSpace(result), "%d", &count)
	return count
}

// getOrCreateSession gets or creates a conference session
func (s *Service) getOrCreateSession(confName string, tenantID uint) *models.ConferenceSession {
	if val, ok := s.sessions.Load(confName); ok {
		return val.(*models.ConferenceSession)
	}

	session := &models.ConferenceSession{
		ConferenceName: confName,
		TenantID:       tenantID,
		StartTime:      time.Now(),
	}

	if s.db != nil {
		s.db.Create(session)
	}

	s.sessions.Store(confName, session)
	return session
}

// handleParticipantLeave marks participant as left
func (s *Service) handleParticipantLeave(session *models.ConferenceSession, uuid string) {
	if s.db == nil {
		return
	}

	now := time.Now()
	s.db.Model(&models.ConferenceParticipant{}).
		Where("session_id = ? AND uuid = ? AND leave_time IS NULL", session.ID, uuid).
		Update("leave_time", now)

	// Update session member count
	var activeCount int64
	s.db.Model(&models.ConferenceParticipant{}).
		Where("session_id = ? AND leave_time IS NULL", session.ID).
		Count(&activeCount)

	s.db.Model(session).Update("member_count", activeCount)

	// End session if empty
	if activeCount == 0 {
		s.db.Model(session).Update("end_time", now)
		s.sessions.Delete(session.ConferenceName)
	}
}

// handleConferenceEvent processes conference-specific events
func (s *Service) handleConferenceEvent(session *models.ConferenceSession, ev *eventsocket.Event) {
	action := ev.Get("Action")

	switch action {
	case "add-member":
		s.handleMemberAdd(session, ev)
	case "del-member":
		s.handleMemberDel(session, ev)
	case "start-talking":
		s.updateMemberState(session, ev, "talking", true)
	case "stop-talking":
		s.updateMemberState(session, ev, "talking", false)
	case "mute-member":
		s.updateMemberState(session, ev, "muted", true)
	case "unmute-member":
		s.updateMemberState(session, ev, "muted", false)
	case "lock":
		s.db.Model(session).Update("locked", true)
	case "unlock":
		s.db.Model(session).Update("locked", false)
	case "start-recording":
		s.db.Model(session).Updates(map[string]interface{}{
			"recording":      true,
			"recording_path": ev.Get("Path"),
		})
	case "stop-recording":
		s.db.Model(session).Update("recording", false)
	}
}

func (s *Service) handleMemberAdd(session *models.ConferenceSession, ev *eventsocket.Event) {
	if s.db == nil {
		return
	}

	memberID, _ := strconv.Atoi(ev.Get("Member-ID"))
	participant := &models.ConferenceParticipant{
		SessionID:      session.ID,
		TenantID:       session.TenantID,
		MemberID:       memberID,
		UUID:           ev.Get("Unique-ID"),
		CallerIDName:   ev.Get("Caller-Caller-ID-Name"),
		CallerIDNumber: ev.Get("Caller-Caller-ID-Number"),
		JoinTime:       time.Now(),
	}

	s.db.Create(participant)

	// Update session stats
	var count int64
	s.db.Model(&models.ConferenceParticipant{}).
		Where("session_id = ? AND leave_time IS NULL", session.ID).
		Count(&count)

	updates := map[string]interface{}{
		"member_count": count,
		"total_joins":  gorm.Expr("total_joins + 1"),
	}
	if int(count) > session.MaxMembers {
		updates["max_members"] = count
	}
	s.db.Model(session).Updates(updates)
}

func (s *Service) handleMemberDel(session *models.ConferenceSession, ev *eventsocket.Event) {
	uuid := ev.Get("Unique-ID")
	s.handleParticipantLeave(session, uuid)
}

func (s *Service) updateMemberState(session *models.ConferenceSession, ev *eventsocket.Event, field string, value bool) {
	if s.db == nil {
		return
	}
	uuid := ev.Get("Unique-ID")
	s.db.Model(&models.ConferenceParticipant{}).
		Where("session_id = ? AND uuid = ? AND leave_time IS NULL", session.ID, uuid).
		Update(field, value)
}

// ========== Conference Control Methods (called via API) ==========

// ListLive returns list of active conferences from FreeSWITCH
func (s *Service) ListLive() ([]models.LiveConferenceInfo, error) {
	manager := s.Manager()
	if manager == nil || manager.Client == nil {
		return nil, fmt.Errorf("ESL client not connected")
	}

	result, err := manager.Client.API("conference list")
	if err != nil {
		return nil, err
	}

	// Parse conference list
	return parseConferenceList(result), nil
}

// GetLiveConference returns live conference info
func (s *Service) GetLiveConference(confName string) (*models.LiveConferenceInfo, error) {
	manager := s.Manager()
	if manager == nil || manager.Client == nil {
		return nil, fmt.Errorf("ESL client not connected")
	}

	result, err := manager.Client.API(fmt.Sprintf("conference %s list", confName))
	if err != nil {
		return nil, err
	}

	info := parseConferenceMembers(confName, result)
	return info, nil
}

// MuteMember mutes a conference member
func (s *Service) MuteMember(confName string, memberID int) error {
	return s.conferenceAction(confName, "mute", strconv.Itoa(memberID))
}

// UnmuteMember unmutes a conference member
func (s *Service) UnmuteMember(confName string, memberID int) error {
	return s.conferenceAction(confName, "unmute", strconv.Itoa(memberID))
}

// DeafMember makes a member deaf
func (s *Service) DeafMember(confName string, memberID int) error {
	return s.conferenceAction(confName, "deaf", strconv.Itoa(memberID))
}

// UndeafMember removes deaf from member
func (s *Service) UndeafMember(confName string, memberID int) error {
	return s.conferenceAction(confName, "undeaf", strconv.Itoa(memberID))
}

// KickMember kicks a member from conference
func (s *Service) KickMember(confName string, memberID int) error {
	return s.conferenceAction(confName, "kick", strconv.Itoa(memberID))
}

// LockConference locks the conference
func (s *Service) LockConference(confName string) error {
	return s.conferenceAction(confName, "lock", "")
}

// UnlockConference unlocks the conference
func (s *Service) UnlockConference(confName string) error {
	return s.conferenceAction(confName, "unlock", "")
}

// StartRecording starts recording the conference
func (s *Service) StartRecording(confName, path string) error {
	return s.conferenceAction(confName, "recording", "start", path)
}

// StopRecording stops recording
func (s *Service) StopRecording(confName string) error {
	return s.conferenceAction(confName, "recording", "stop")
}

// SetFloor gives floor to a member
func (s *Service) SetFloor(confName string, memberID int) error {
	return s.conferenceAction(confName, "floor", strconv.Itoa(memberID))
}

// MuteAll mutes all participants
func (s *Service) MuteAll(confName string) error {
	return s.conferenceAction(confName, "mute", "all")
}

// UnmuteAll unmutes all participants
func (s *Service) UnmuteAll(confName string) error {
	return s.conferenceAction(confName, "unmute", "all")
}

// conferenceAction executes a conference API command
func (s *Service) conferenceAction(confName string, args ...string) error {
	manager := s.Manager()
	if manager == nil || manager.Client == nil {
		return fmt.Errorf("ESL client not connected")
	}

	// Build command: "conference confName arg1 arg2 ..."
	cmd := fmt.Sprintf("conference %s %s", confName, strings.Join(args, " "))
	_, err := manager.Client.API(cmd)
	return err
}

// Helper functions to parse FreeSWITCH conference output

func parseConferenceList(output string) []models.LiveConferenceInfo {
	var conferences []models.LiveConferenceInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Conference") {
			continue
		}
		// Format: "conference_name (members) (flags)"
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			info := models.LiveConferenceInfo{Name: parts[0]}
			if len(parts) >= 2 {
				info.MemberCount, _ = strconv.Atoi(strings.Trim(parts[1], "()"))
			}
			conferences = append(conferences, info)
		}
	}
	return conferences
}

func parseConferenceMembers(confName, output string) *models.LiveConferenceInfo {
	info := &models.LiveConferenceInfo{
		Name:    confName,
		Members: []models.LiveConferenceMember{},
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Parse member line: id;uuid;caller_id_name;caller_id_number;flags
		parts := strings.Split(line, ";")
		if len(parts) >= 4 {
			member := models.LiveConferenceMember{
				CallerIDName:   parts[2],
				CallerIDNumber: parts[3],
			}
			member.ID, _ = strconv.Atoi(parts[0])
			member.UUID = parts[1]

			if len(parts) >= 5 {
				flags := parts[4]
				member.Muted = strings.Contains(flags, "mute")
				member.Deaf = strings.Contains(flags, "deaf")
				member.Talking = strings.Contains(flags, "talking")
				member.Floor = strings.Contains(flags, "floor")
				member.Video = strings.Contains(flags, "video")
			}

			info.Members = append(info.Members, member)
		}
	}

	info.MemberCount = len(info.Members)
	return info
}
