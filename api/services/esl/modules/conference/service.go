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

	// Build conference name
	confName := fmt.Sprintf("%s-%s@default", domain, conferenceNum)

	// Get or create session
	tenantID, _ := strconv.ParseUint(tenantIDStr, 10, 32)
	session := s.getOrCreateSession(confName, uint(tenantID))

	// Set caller info for conference
	conn.Execute("set", "conference_member_nospeak_relax=true", true)
	conn.Execute("set", fmt.Sprintf("effective_caller_id_name=%s", callerName), true)
	conn.Execute("set", fmt.Sprintf("effective_caller_id_number=%s", callerID), true)

	logger.Infof("Joining conference: %s", confName)

	// Join the conference (blocks until hangup)
	conn.Execute("conference", confName, true)

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
