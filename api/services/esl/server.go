package esl

import (
	"fmt"
	"net"
	"sync"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
)

// Server represents an outbound ESL server that FreeSWITCH connects to
type Server struct {
	Address  string
	handler  ConnectionHandler
	listener net.Listener
	done     chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	running  bool
}

// ConnectionHandler handles incoming ESL connections from FreeSWITCH
type ConnectionHandler func(conn *eventsocket.Connection)

// NewServer creates a new outbound ESL server
func NewServer(address string, handler ConnectionHandler) *Server {
	return &Server{
		Address: address,
		handler: handler,
		done:    make(chan struct{}),
	}
}

// ListenAndServe starts the ESL server
func (s *Server) ListenAndServe() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	log.Infof("Starting ESL server on %s", s.Address)

	err := eventsocket.ListenAndServe(s.Address, s.handleConnection)
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return err
	}

	return nil
}

// Start starts the server in a goroutine
func (s *Server) Start() error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.ListenAndServe(); err != nil {
			select {
			case <-s.done:
				// Expected shutdown
			default:
				log.Errorf("ESL server error: %v", err)
			}
		}
	}()
	return nil
}

func (s *Server) handleConnection(conn *eventsocket.Connection) {
	log.Debugf("Incoming ESL connection from FreeSWITCH")

	s.wg.Add(1)
	defer s.wg.Done()

	// Call the user-provided handler
	s.handler(conn)
}

// Stop stops the ESL server
func (s *Server) Stop() {
	close(s.done)

	s.mu.Lock()
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	s.mu.Unlock()

	s.wg.Wait()
	log.Info("ESL server stopped")
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// ServiceRegistry manages multiple ESL servers for different services
type ServiceRegistry struct {
	servers map[string]*Server
	mu      sync.RWMutex
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		servers: make(map[string]*Server),
	}
}

// Register registers a service handler
func (r *ServiceRegistry) Register(name, address string, handler ConnectionHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	server := NewServer(address, handler)
	r.servers[name] = server

	log.Infof("Registered ESL service: %s at %s", name, address)
	return nil
}

// StartAll starts all registered services
func (r *ServiceRegistry) StartAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, server := range r.servers {
		log.Infof("Starting ESL service: %s", name)
		if err := server.Start(); err != nil {
			return fmt.Errorf("failed to start %s: %w", name, err)
		}
	}

	return nil
}

// StopAll stops all registered services
func (r *ServiceRegistry) StopAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, server := range r.servers {
		log.Infof("Stopping ESL service: %s", name)
		server.Stop()
	}
}

// Get returns a specific server by name
func (r *ServiceRegistry) Get(name string) *Server {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.servers[name]
}
