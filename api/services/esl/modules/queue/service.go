package queue

import (
	"callsign/models"
	"callsign/services/esl"
	"fmt"
	"strings"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
)

const (
	ServiceName    = "queue"
	ServiceAddress = "127.0.0.5:9001"
)

// Service implements the call queue ESL module
type Service struct {
	*esl.BaseService
}

// New creates a new queue service
func New() *Service {
	return &Service{
		BaseService: esl.NewBaseService(ServiceName, ServiceAddress),
	}
}

// Init initializes the queue service
func (s *Service) Init(manager *esl.Manager) error {
	if err := s.BaseService.Init(manager); err != nil {
		return err
	}

	// Sync agents to FreeSWITCH on startup
	go s.syncAgentsToFreeSWITCH()

	log.Info("Queue service initialized")
	return nil
}

// syncAgentsToFreeSWITCH syncs all agents from DB to FreeSWITCH mod_callcenter
func (s *Service) syncAgentsToFreeSWITCH() {
	manager := s.Manager()
	if manager == nil || manager.Client == nil {
		return
	}

	db := manager.DB
	var agents []models.QueueAgent
	db.Find(&agents)

	for _, agent := range agents {
		// Add agent to FreeSWITCH
		cmd := fmt.Sprintf("callcenter_config agent add %s callback", agent.AgentName)
		manager.Client.API(cmd)

		// Set contact
		if agent.Contact != "" {
			cmd = fmt.Sprintf("callcenter_config agent set contact %s %s", agent.AgentName, agent.Contact)
			manager.Client.API(cmd)
		}

		// Set status
		cmd = fmt.Sprintf("callcenter_config agent set status %s '%s'", agent.AgentName, agent.Status)
		manager.Client.API(cmd)
	}

	log.Infof("Synced %d queue agents to FreeSWITCH", len(agents))
}

// Handle processes incoming queue connections
func (s *Service) Handle(conn *eventsocket.Connection) {
	defer conn.Close()

	manager := s.Manager()
	if manager == nil {
		log.Error("Queue: manager not initialized")
		return
	}

	// Connect and get channel info
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Queue: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	callerName := ev.Get("Caller-Caller-ID-Name")
	dest := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")
	queueName := ev.Get("variable_queue_name") // May be set in dialplan

	// If no queue name, use destination
	if queueName == "" {
		queueName = dest
	}

	logger := log.WithFields(log.Fields{
		"uuid":   uuid,
		"caller": callerID,
		"queue":  queueName,
		"domain": domain,
	})
	logger.Info("Queue: incoming call")

	conn.Send("linger")
	conn.Send("myevents")

	// Find queue in database
	db := manager.DB
	var queue models.Queue
	if err := db.Where("(extension = ? OR name = ?) AND enabled = ?", queueName, queueName, true).First(&queue).Error; err != nil {
		logger.Warnf("Queue not found: %s", queueName)
		conn.Execute("playback", "ivr/ivr-call_cannot_be_completed_as_dialed.wav", true)
		conn.Execute("hangup", "", false)
		return
	}

	// Set caller ID info
	conn.Execute("set", fmt.Sprintf("effective_caller_id_name=%s", callerName), true)
	conn.Execute("set", fmt.Sprintf("effective_caller_id_number=%s", callerID), true)

	// Answer
	conn.Execute("answer", "", true)

	// ---------- Position Announcement ----------
	// Get queue stats to announce position before joining
	positionCount := s.getWaitingCount(queue.Name, domain)
	if positionCount >= 0 {
		pos := positionCount + 1 // This caller will be next after existing
		logger.Infof("Caller position in queue: %d", pos)

		// Set position as channel variable for later announcements
		conn.Execute("set", fmt.Sprintf("queue_position=%d", pos), true)

		// Play position announcement
		conn.Execute("playback", "ivr/ivr-you_are_number.wav", true)
		conn.Execute("say", fmt.Sprintf("en number pronounced %d", pos), true)
		conn.Execute("playback", "ivr/ivr-in_line.wav", true)
	}

	// ---------- Estimated Wait Time ----------
	if queue.AnnounceWaitTime {
		avgWait := s.getAverageWaitTime(queue.Name, domain)
		if avgWait > 0 {
			minutes := int(avgWait.Minutes())
			if minutes > 0 {
				conn.Execute("playback", "ivr/ivr-estimated_wait_time_is.wav", true)
				conn.Execute("say", fmt.Sprintf("en number pronounced %d", minutes), true)
				conn.Execute("playback", "ivr/ivr-minutes.wav", true)
			}
		}
	}

	// Play initial announcement if configured
	if queue.AnnounceSound != "" {
		conn.Execute("playback", queue.AnnounceSound, true)
	}

	// ---------- Callback Offer ----------
	if queue.CallbackEnabled && positionCount >= queue.CallbackThreshold {
		logger.Info("Queue: offering callback to caller")
		offered := s.offerCallback(conn, uuid, callerID, callerName, &queue, domain, logger)
		if offered {
			// Caller accepted callback — hang up, they'll be called back
			return
		}
		// Caller declined or timed out — continue to queue
	}

	// Build callcenter app string
	ccQueue := queue.Name
	if domain != "" {
		ccQueue = fmt.Sprintf("%s@%s", queue.Name, domain)
	}

	logger.Infof("Joining caller to queue: %s", ccQueue)

	// ---------- Periodic Position Announcements ----------
	// Set up periodic announcement if configured
	if queue.AnnouncePosition && queue.AnnounceFrequency > 0 {
		conn.Execute("set", fmt.Sprintf(
			"cc_queue_announce_position=%d", queue.AnnounceFrequency), true)
	}

	// Join the queue (blocks until answered or times out)
	conn.Execute("callcenter", ccQueue, true)

	// After callcenter app returns, log the result
	hangupCause := ""
	for {
		ev, err := conn.ReadEvent()
		if err != nil {
			break
		}

		eventName := ev.Get("Event-Name")
		if eventName == "CHANNEL_HANGUP_COMPLETE" {
			hangupCause = ev.Get("Hangup-Cause")
			break
		}
	}

	logger.WithField("hangup_cause", hangupCause).Info("Queue: call ended")
}

// offerCallback offers the caller a callback and schedules it if accepted.
// Returns true if callback was accepted (caller should hang up).
func (s *Service) offerCallback(
	conn *eventsocket.Connection,
	uuid, callerID, callerName string,
	queue *models.Queue,
	domain string,
	logger *log.Entry,
) bool {
	// Play: "Press 1 to receive a callback when an agent is available. Press 2 to continue waiting."
	conn.Execute("play_and_get_digits",
		"1 1 1 5000 # ivr/ivr-press_one_to_accept.wav silence_stream://250 callback_choice \\d 5000", true)

	ev, err := conn.Send("api uuid_getvar " + uuid + " callback_choice")
	if err != nil {
		return false
	}

	choice := strings.TrimSpace(ev.Body)
	if choice != "1" {
		return false
	}

	logger.Info("Queue: caller accepted callback")

	// Schedule callback in DB
	manager := s.Manager()
	if manager != nil {
		callback := &models.QueueCallback{
			QueueID:        queue.ID,
			TenantID:       queue.TenantID,
			CallerNumber:   callerID,
			CallerName:     callerName,
			CallbackNumber: callerID,
			Status:         "pending",
			RequestedAt:    time.Now(),
		}
		manager.DB.Create(callback)
	}

	// Confirm to caller
	conn.Execute("playback", "ivr/ivr-you_will_be_called_back.wav", true)
	conn.Execute("sleep", "500", true)
	conn.Execute("hangup", "NORMAL_CLEARING", false)
	return true
}

// getWaitingCount returns number of callers waiting in queue
func (s *Service) getWaitingCount(queueName, domain string) int {
	manager := s.Manager()
	if manager == nil || manager.Client == nil {
		return -1
	}

	ccQueue := queueName
	if domain != "" {
		ccQueue = fmt.Sprintf("%s@%s", queueName, domain)
	}

	result, err := manager.Client.API(fmt.Sprintf("callcenter_config queue list members %s", ccQueue))
	if err != nil {
		return -1
	}

	// Count non-empty, non-header lines
	lines := strings.Split(result, "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "name|") && line != "+OK" {
			count++
		}
	}
	return count
}

// getAverageWaitTime calculates average wait time from recent calls
func (s *Service) getAverageWaitTime(queueName, domain string) time.Duration {
	manager := s.Manager()
	if manager == nil {
		return 0
	}

	// Look at last 10 answered calls from this queue to calculate average
	var avgSeconds float64
	manager.DB.Raw(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (answered_at - created_at))), 0) 
		FROM queue_calls 
		WHERE queue_name = ? AND status = 'answered' 
		AND created_at > NOW() - INTERVAL '1 hour'
		LIMIT 10
	`, queueName).Scan(&avgSeconds)

	return time.Duration(avgSeconds) * time.Second
}

// AddAgent adds an agent to a queue in both DB and FreeSWITCH
func (s *Service) AddAgent(queueID uint, extensionID uint, agentName, contact string) error {
	manager := s.Manager()
	if manager == nil {
		return fmt.Errorf("manager not initialized")
	}

	db := manager.DB

	// Get queue
	var queue models.Queue
	if err := db.First(&queue, queueID).Error; err != nil {
		return fmt.Errorf("queue not found: %w", err)
	}

	// Create agent record
	agent := &models.QueueAgent{
		QueueID:     queueID,
		ExtensionID: extensionID,
		TenantID:    queue.TenantID,
		AgentName:   agentName,
		Contact:     contact,
		Status:      models.AgentStatusLoggedOut,
		State:       models.AgentStateIdle,
	}

	if err := db.Create(agent).Error; err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Sync to FreeSWITCH
	if manager.Client != nil && manager.Client.IsConnected() {
		cmd := fmt.Sprintf("callcenter_config agent add %s callback", agentName)
		manager.Client.API(cmd)

		if contact != "" {
			cmd = fmt.Sprintf("callcenter_config agent set contact %s %s", agentName, contact)
			manager.Client.API(cmd)
		}

		// Add tier (link agent to queue)
		cmd = fmt.Sprintf("callcenter_config tier add %s %s %d %d",
			queue.Name, agentName, agent.TierLevel, agent.TierPosition)
		manager.Client.API(cmd)
	}

	return nil
}

// LoginAgent logs an agent in via ESL (used by feature code *90)
func (s *Service) LoginAgent(tenantID uint, extension string) error {
	manager := s.Manager()
	if manager == nil {
		return fmt.Errorf("manager not initialized")
	}

	db := manager.DB

	// Find agent by extension
	var agent models.QueueAgent
	if err := db.Where("tenant_id = ? AND contact LIKE ?", tenantID, "%"+extension+"%").
		First(&agent).Error; err != nil {
		return fmt.Errorf("agent not found for extension %s: %w", extension, err)
	}

	return s.SetAgentStatus(agent.ID, models.AgentStatusAvailable)
}

// LogoutAgent logs an agent out via ESL (used by feature code *91)
func (s *Service) LogoutAgent(tenantID uint, extension string) error {
	manager := s.Manager()
	if manager == nil {
		return fmt.Errorf("manager not initialized")
	}

	db := manager.DB

	var agent models.QueueAgent
	if err := db.Where("tenant_id = ? AND contact LIKE ?", tenantID, "%"+extension+"%").
		First(&agent).Error; err != nil {
		return fmt.Errorf("agent not found for extension %s: %w", extension, err)
	}

	return s.SetAgentStatus(agent.ID, models.AgentStatusLoggedOut)
}

// SetAgentStatus updates an agent's status in both DB and FreeSWITCH
func (s *Service) SetAgentStatus(agentID uint, status models.AgentStatus) error {
	manager := s.Manager()
	if manager == nil {
		return fmt.Errorf("manager not initialized")
	}

	db := manager.DB

	var agent models.QueueAgent
	if err := db.First(&agent, agentID).Error; err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	// Update in DB
	agent.Status = status
	if err := db.Save(&agent).Error; err != nil {
		return err
	}

	// Sync to FreeSWITCH
	if manager.Client != nil && manager.Client.IsConnected() {
		cmd := fmt.Sprintf("callcenter_config agent set status %s '%s'", agent.AgentName, status)
		manager.Client.API(cmd)
	}

	return nil
}

// ProcessCallbacks checks for pending callbacks and originates calls
func (s *Service) ProcessCallbacks() {
	manager := s.Manager()
	if manager == nil || manager.Client == nil {
		return
	}

	db := manager.DB
	var callbacks []models.QueueCallback
	db.Where("status = 'pending'").Order("requested_at ASC").Limit(5).Find(&callbacks)

	for _, cb := range callbacks {
		// Check if agents are available for this queue
		var queue models.Queue
		if err := db.First(&queue, cb.QueueID).Error; err != nil {
			continue
		}

		// Get available agent count
		ccQueue := queue.Name
		result, err := manager.Client.API(fmt.Sprintf("callcenter_config queue %s count agents Available", ccQueue))
		if err != nil {
			continue
		}

		var available int
		fmt.Sscanf(strings.TrimSpace(result), "%d", &available)
		if available == 0 {
			continue
		}

		// Originate callback
		log.WithFields(log.Fields{
			"callback_id": cb.ID,
			"number":      cb.CallbackNumber,
			"queue":       queue.Name,
		}).Info("Queue: originating callback")

		// Mark as in-progress
		db.Model(&cb).Update("status", "calling")

		// Originate: call the customer and bridge to queue
		originateCmd := fmt.Sprintf(
			"originate {origination_caller_id_name='%s',origination_caller_id_number='%s'}sofia/gateway/default/%s &callcenter(%s)",
			queue.Name, "callback", cb.CallbackNumber, ccQueue,
		)
		manager.Client.API(originateCmd)
	}
}

// GetQueueStats retrieves real-time queue statistics from FreeSWITCH
func (s *Service) GetQueueStats(queueName string) (*models.QueueStatistics, error) {
	manager := s.Manager()
	if manager == nil {
		return nil, fmt.Errorf("manager not initialized")
	}

	stats := &models.QueueStatistics{
		QueueName:   queueName,
		LastUpdated: time.Now(),
	}

	if manager.Client == nil || !manager.Client.IsConnected() {
		return stats, nil
	}

	// Get waiting members
	result, err := manager.Client.API(fmt.Sprintf("callcenter_config queue list members %s", queueName))
	if err == nil {
		lines := strings.Split(result, "\n")
		stats.WaitingCalls = len(lines) - 1
	}

	// Get agent counts
	result, err = manager.Client.API(fmt.Sprintf("callcenter_config queue %s count agents Available", queueName))
	if err == nil {
		fmt.Sscanf(strings.TrimSpace(result), "%d", &stats.AvailableAgents)
	}

	result, err = manager.Client.API(fmt.Sprintf("callcenter_config tier list agents %s", queueName))
	if err == nil {
		lines := strings.Split(result, "\n")
		stats.TotalAgents = len(lines) - 1
	}

	return stats, nil
}
