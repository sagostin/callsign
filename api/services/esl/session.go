package esl

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionState represents the state of a call session
type SessionState string

const (
	SessionStateInitiating SessionState = "initiating"
	SessionStateRinging    SessionState = "ringing"
	SessionStateEarly      SessionState = "early" // Early media
	SessionStateAnswered   SessionState = "answered"
	SessionStateBridged    SessionState = "bridged"
	SessionStateHeld       SessionState = "held"
	SessionStateTransfer   SessionState = "transferring"
	SessionStateHangup     SessionState = "hangup"
)

// CallDirection represents the direction of a call leg
type CallDirection string

const (
	DirectionInbound  CallDirection = "inbound"
	DirectionOutbound CallDirection = "outbound"
)

// ChannelState represents the state of a single call leg
type ChannelState struct {
	UUID           string            `json:"uuid"`
	Profile        string            `json:"profile"` // internal, webrtc, public
	Direction      CallDirection     `json:"direction"`
	CallerIDName   string            `json:"caller_id_name"`
	CallerIDNumber string            `json:"caller_id_number"`
	Destination    string            `json:"destination"`
	Context        string            `json:"context"`
	CreatedAt      time.Time         `json:"created_at"`
	AnsweredAt     *time.Time        `json:"answered_at,omitempty"`
	HangupAt       *time.Time        `json:"hangup_at,omitempty"`
	HangupCause    string            `json:"hangup_cause,omitempty"`
	Variables      map[string]string `json:"variables"`
}

// NewChannelState creates a new channel state
func NewChannelState(uuid, profile string, direction CallDirection) *ChannelState {
	return &ChannelState{
		UUID:      uuid,
		Profile:   profile,
		Direction: direction,
		CreatedAt: time.Now(),
		Variables: make(map[string]string),
	}
}

// CallSession represents a B2BUA session with A-leg and B-leg
type CallSession struct {
	UUID       uuid.UUID      `json:"uuid"`
	TenantID   uint           `json:"tenant_id"`
	DomainName string         `json:"domain_name"`
	ALeg       *ChannelState  `json:"a_leg"`           // Originator
	BLeg       *ChannelState  `json:"b_leg,omitempty"` // Destination (nil until bridged)
	State      SessionState   `json:"state"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Metadata   map[string]any `json:"metadata"`
	mu         sync.RWMutex
}

// NewCallSession creates a new call session
func NewCallSession(tenantID uint, domainName string) *CallSession {
	return &CallSession{
		UUID:       uuid.New(),
		TenantID:   tenantID,
		DomainName: domainName,
		State:      SessionStateInitiating,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   make(map[string]any),
	}
}

// SetALeg sets the A-leg (originator) channel
func (s *CallSession) SetALeg(ch *ChannelState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ALeg = ch
	s.UpdatedAt = time.Now()
}

// SetBLeg sets the B-leg (destination) channel
func (s *CallSession) SetBLeg(ch *ChannelState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BLeg = ch
	s.UpdatedAt = time.Now()
}

// SetState updates the session state
func (s *CallSession) SetState(state SessionState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.State = state
	s.UpdatedAt = time.Now()
}

// GetState returns the current session state
func (s *CallSession) GetState() SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State
}

// SetMetadata sets a metadata value
func (s *CallSession) SetMetadata(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Metadata[key] = value
	s.UpdatedAt = time.Now()
}

// GetMetadata gets a metadata value
func (s *CallSession) GetMetadata(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.Metadata[key]
	return val, ok
}

// IsBridged returns true if both legs exist
func (s *CallSession) IsBridged() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ALeg != nil && s.BLeg != nil && s.State == SessionStateBridged
}

// Duration returns the call duration (from answer to now/hangup)
func (s *CallSession) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.ALeg == nil || s.ALeg.AnsweredAt == nil {
		return 0
	}

	if s.ALeg.HangupAt != nil {
		return s.ALeg.HangupAt.Sub(*s.ALeg.AnsweredAt)
	}

	return time.Since(*s.ALeg.AnsweredAt)
}

// SessionManager manages active call sessions
type SessionManager struct {
	sessions map[string]*CallSession // Key: A-leg UUID
	byBLeg   map[string]string       // B-leg UUID -> A-leg UUID
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*CallSession),
		byBLeg:   make(map[string]string),
	}
}

// Create creates a new session and tracks it
func (m *SessionManager) Create(tenantID uint, domainName string, alegUUID string) *CallSession {
	session := NewCallSession(tenantID, domainName)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[alegUUID] = session
	return session
}

// Get retrieves a session by A-leg UUID
func (m *SessionManager) Get(alegUUID string) *CallSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[alegUUID]
}

// GetByBLeg retrieves a session by B-leg UUID
func (m *SessionManager) GetByBLeg(blegUUID string) *CallSession {
	m.mu.RLock()
	alegUUID, ok := m.byBLeg[blegUUID]
	m.mu.RUnlock()

	if !ok {
		return nil
	}
	return m.Get(alegUUID)
}

// GetByUUID retrieves a session by either A-leg or B-leg UUID
func (m *SessionManager) GetByUUID(channelUUID string) *CallSession {
	if session := m.Get(channelUUID); session != nil {
		return session
	}
	return m.GetByBLeg(channelUUID)
}

// RegisterBLeg associates a B-leg UUID with a session
func (m *SessionManager) RegisterBLeg(alegUUID, blegUUID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byBLeg[blegUUID] = alegUUID
}

// Remove removes a session
func (m *SessionManager) Remove(alegUUID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := m.sessions[alegUUID]
	if session != nil && session.BLeg != nil {
		delete(m.byBLeg, session.BLeg.UUID)
	}

	delete(m.sessions, alegUUID)
}

// Count returns the number of active sessions
func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// All returns all active sessions
func (m *SessionManager) All() []*CallSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*CallSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}
