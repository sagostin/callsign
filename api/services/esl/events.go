package esl

import (
	"strconv"
	"strings"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
)

// EventHandler defines a handler for FreeSWITCH events
type EventHandler func(event *eventsocket.Event, session *CallSession)

// EventProcessor processes FreeSWITCH events
type EventProcessor struct {
	client   *Client
	sessions *SessionManager
	handlers map[string][]EventHandler
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(client *Client, sessions *SessionManager) *EventProcessor {
	return &EventProcessor{
		client:   client,
		sessions: sessions,
		handlers: make(map[string][]EventHandler),
	}
}

// On registers a handler for an event type
func (p *EventProcessor) On(eventName string, handler EventHandler) {
	p.handlers[eventName] = append(p.handlers[eventName], handler)
}

// Start starts processing events from the client
func (p *EventProcessor) Start() {
	go p.processEvents()
}

func (p *EventProcessor) processEvents() {
	for {
		select {
		case event := <-p.client.Events():
			p.handleEvent(event)
		case err := <-p.client.Errors():
			log.Errorf("ESL error: %v", err)
		}
	}
}

func (p *EventProcessor) handleEvent(event *eventsocket.Event) {
	eventName := event.Get("Event-Name")
	if eventName == "" {
		return
	}

	// Get the channel UUID
	channelUUID := event.Get("Unique-ID")
	if channelUUID == "" {
		channelUUID = event.Get("Channel-UUID")
	}

	// Find associated session
	var session *CallSession
	if channelUUID != "" {
		session = p.sessions.GetByUUID(channelUUID)
	}

	// Log event
	log.WithFields(log.Fields{
		"event":       eventName,
		"uuid":        channelUUID,
		"has_session": session != nil,
	}).Debug("Processing event")

	// Call registered handlers
	if handlers, ok := p.handlers[eventName]; ok {
		for _, h := range handlers {
			h(event, session)
		}
	}

	// Call wildcard handlers
	if handlers, ok := p.handlers["*"]; ok {
		for _, h := range handlers {
			h(event, session)
		}
	}
}

// DefaultEventHandlers returns common event handlers
func DefaultEventHandlers(sessions *SessionManager) map[string]EventHandler {
	return map[string]EventHandler{
		"CHANNEL_CREATE": func(ev *eventsocket.Event, session *CallSession) {
			uuid := ev.Get("Unique-ID")
			direction := ev.Get("Call-Direction")
			profile := extractProfileFromChannelName(ev.Get("Channel-Name"))
			domain := ev.Get("variable_domain_name")

			if session == nil && direction == "inbound" {
				// Create new session for inbound calls
				session = sessions.Create(0, domain, uuid)

				ch := NewChannelState(uuid, profile, DirectionInbound)
				ch.CallerIDName = ev.Get("Caller-Caller-ID-Name")
				ch.CallerIDNumber = ev.Get("Caller-Caller-ID-Number")
				ch.Destination = ev.Get("Caller-Destination-Number")
				ch.Context = ev.Get("Caller-Context")

				session.SetALeg(ch)
				session.SetState(SessionStateRinging)

				log.WithFields(log.Fields{
					"uuid":        uuid,
					"caller":      ch.CallerIDNumber,
					"destination": ch.Destination,
					"domain":      domain,
				}).Info("New inbound call session")
			}
		},

		"CHANNEL_ANSWER": func(ev *eventsocket.Event, session *CallSession) {
			if session == nil {
				return
			}

			uuid := ev.Get("Unique-ID")
			now := parseEventTimestamp(ev)

			if session.ALeg != nil && session.ALeg.UUID == uuid {
				session.ALeg.AnsweredAt = &now
				session.SetState(SessionStateAnswered)
				log.Infof("A-leg answered: %s", uuid)
			} else if session.BLeg != nil && session.BLeg.UUID == uuid {
				session.BLeg.AnsweredAt = &now
				session.SetState(SessionStateBridged)
				log.Infof("B-leg answered, call bridged: %s", uuid)
			}
		},

		"CHANNEL_BRIDGE": func(ev *eventsocket.Event, session *CallSession) {
			if session == nil {
				return
			}

			otherLegUUID := ev.Get("Other-Leg-Unique-ID")
			if otherLegUUID != "" && session.BLeg == nil {
				// Register B-leg
				direction := DirectionOutbound
				if ev.Get("Other-Leg-Direction") == "inbound" {
					direction = DirectionInbound
				}

				bleg := NewChannelState(
					otherLegUUID,
					extractProfileFromChannelName(ev.Get("Other-Leg-Channel-Name")),
					direction,
				)
				bleg.CallerIDName = ev.Get("Other-Leg-Caller-ID-Name")
				bleg.CallerIDNumber = ev.Get("Other-Leg-Caller-ID-Number")
				bleg.Destination = ev.Get("Other-Leg-Destination-Number")

				session.SetBLeg(bleg)
				session.SetState(SessionStateBridged)
				sessions.RegisterBLeg(session.ALeg.UUID, otherLegUUID)

				log.WithFields(log.Fields{
					"a_uuid": session.ALeg.UUID,
					"b_uuid": otherLegUUID,
				}).Info("Call bridged")
			}
		},

		"CHANNEL_HANGUP_COMPLETE": func(ev *eventsocket.Event, session *CallSession) {
			if session == nil {
				return
			}

			uuid := ev.Get("Unique-ID")
			cause := ev.Get("Hangup-Cause")
			now := parseEventTimestamp(ev)

			if session.ALeg != nil && session.ALeg.UUID == uuid {
				session.ALeg.HangupAt = &now
				session.ALeg.HangupCause = cause
			} else if session.BLeg != nil && session.BLeg.UUID == uuid {
				session.BLeg.HangupAt = &now
				session.BLeg.HangupCause = cause
			}

			session.SetState(SessionStateHangup)

			log.WithFields(log.Fields{
				"uuid":     uuid,
				"cause":    cause,
				"duration": session.Duration().String(),
			}).Info("Call ended")

			// Remove session after hangup
			if session.ALeg != nil {
				sessions.Remove(session.ALeg.UUID)
			}
		},
	}
}

// Helper functions
func extractProfileFromChannelName(channelName string) string {
	// Channel names look like: sofia/internal/1001@domain
	parts := strings.Split(channelName, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

func parseEventTimestamp(ev *eventsocket.Event) (t time.Time) {
	// Event timestamps are in microseconds
	ts := ev.Get("Event-Date-Timestamp")
	if ts != "" {
		if usec, err := strconv.ParseInt(ts, 10, 64); err == nil {
			return time.UnixMicro(usec)
		}
	}
	return time.Now()
}
