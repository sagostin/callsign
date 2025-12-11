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

	// Play initial announcement if configured
	if queue.AnnounceSound != "" {
		conn.Execute("playback", queue.AnnounceSound, true)
	}

	// Build callcenter app string
	// Format: queue_name[@domain]
	ccQueue := queue.Name
	if domain != "" {
		ccQueue = fmt.Sprintf("%s@%s", queue.Name, domain)
	}

	logger.Infof("Joining caller to queue: %s", ccQueue)

	// Join the queue
	// This blocks until the call is answered or times out
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
		// Parse result to count waiting calls
		lines := strings.Split(result, "\n")
		stats.WaitingCalls = len(lines) - 1 // Subtract header
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
