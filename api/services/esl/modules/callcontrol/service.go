package callcontrol

import (
	"callsign/services/esl"
	"fmt"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
)

const (
	ServiceName    = "callcontrol"
	ServiceAddress = "127.0.0.1:9001"
)

// Service implements the general call control ESL module
type Service struct {
	*esl.BaseService
}

// New creates a new call control service
func New() *Service {
	return &Service{
		BaseService: esl.NewBaseService(ServiceName, ServiceAddress),
	}
}

// Init initializes the call control service
func (s *Service) Init(manager *esl.Manager) error {
	if err := s.BaseService.Init(manager); err != nil {
		return err
	}
	log.Info("Call control service initialized")
	return nil
}

// Handle processes incoming call control connections
func (s *Service) Handle(conn *eventsocket.Connection) {
	defer conn.Close()

	manager := s.Manager()
	if manager == nil {
		log.Error("Callcontrol: manager not initialized")
		return
	}

	// Connect and get channel info
	ev, err := conn.Send("connect")
	if err != nil {
		log.Errorf("Callcontrol: connect failed: %v", err)
		return
	}

	uuid := ev.Get("Unique-ID")
	callerID := ev.Get("Caller-Caller-ID-Number")
	dest := ev.Get("Caller-Destination-Number")
	domain := ev.Get("variable_domain_name")
	context := ev.Get("Caller-Context")

	logger := log.WithFields(log.Fields{
		"uuid":        uuid,
		"caller":      callerID,
		"destination": dest,
		"domain":      domain,
		"context":     context,
	})
	logger.Info("Callcontrol: handling call")

	conn.Send("linger")
	conn.Send("myevents")

	// Route based on destination
	// This is a basic implementation - extend with lookup logic

	// Check if destination is a local extension
	dialString := fmt.Sprintf("user/%s@%s", dest, domain)

	// Set some basic variables
	conn.Execute("set", "hangup_after_bridge=true", true)
	conn.Execute("set", "continue_on_fail=true", true)

	// Ring the destination
	conn.Execute("set", "ringback=${us-ring}", true)

	// Bridge the call
	logger.Infof("Bridging to: %s", dialString)
	conn.Execute("bridge", dialString, true)

	// If bridge fails, handle failure
	bridgeResult := ""
	for {
		ev, err := conn.ReadEvent()
		if err != nil {
			break
		}

		eventName := ev.Get("Event-Name")
		switch eventName {
		case "CHANNEL_BRIDGE":
			logger.Info("Call bridged successfully")
		case "CHANNEL_HANGUP_COMPLETE":
			bridgeResult = ev.Get("variable_bridge_hangup_cause")
			if bridgeResult == "" {
				bridgeResult = ev.Get("Hangup-Cause")
			}
			goto done
		}
	}

done:
	logger.WithField("result", bridgeResult).Info("Callcontrol: call ended")
}
