package esl

import (
	"fmt"
	"sync"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
)

// Service defines the interface for ESL service modules
// Each module implements this interface to handle specific call types
type Service interface {
	// Name returns the unique name of the service
	Name() string

	// Address returns the listen address (e.g., "127.0.0.1:9001")
	Address() string

	// Init initializes the service with access to the manager
	Init(manager *Manager) error

	// Handle processes an incoming socket connection from FreeSWITCH
	Handle(conn *eventsocket.Connection)

	// Shutdown gracefully shuts down the service
	Shutdown()
}

// BaseService provides common functionality for services
type BaseService struct {
	name    string
	address string
	manager *Manager
	mu      sync.RWMutex
}

// NewBaseService creates a new base service
func NewBaseService(name, address string) *BaseService {
	return &BaseService{
		name:    name,
		address: address,
	}
}

func (s *BaseService) Name() string    { return s.name }
func (s *BaseService) Address() string { return s.address }

func (s *BaseService) Init(manager *Manager) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manager = manager
	return nil
}

func (s *BaseService) Manager() *Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.manager
}

func (s *BaseService) Shutdown() {
	// Override in subclasses if needed
}

// ModuleRegistry manages service modules
type ModuleRegistry struct {
	services map[string]Service
	servers  map[string]*Server
	mu       sync.RWMutex
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		services: make(map[string]Service),
		servers:  make(map[string]*Server),
	}
}

// Register adds a service module to the registry
func (r *ModuleRegistry) Register(service Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := service.Name()
	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	r.services[name] = service
	log.Infof("Registered ESL module: %s at %s", name, service.Address())
	return nil
}

// InitAll initializes all registered services
func (r *ModuleRegistry) InitAll(manager *Manager) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, service := range r.services {
		log.Infof("Initializing ESL module: %s", name)
		if err := service.Init(manager); err != nil {
			return fmt.Errorf("failed to init %s: %w", name, err)
		}
	}
	return nil
}

// StartAll starts socket servers for all registered services
func (r *ModuleRegistry) StartAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, service := range r.services {
		server := NewServer(service.Address(), service.Handle)
		r.servers[name] = server

		log.Infof("Starting ESL module server: %s at %s", name, service.Address())
		if err := server.Start(); err != nil {
			return fmt.Errorf("failed to start %s: %w", name, err)
		}
	}
	return nil
}

// StopAll stops all service servers
func (r *ModuleRegistry) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, service := range r.services {
		log.Infof("Stopping ESL module: %s", name)
		service.Shutdown()
	}

	for name, server := range r.servers {
		log.Infof("Stopping ESL server: %s", name)
		server.Stop()
	}
}

// Get returns a service by name
func (r *ModuleRegistry) Get(name string) Service {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.services[name]
}

// List returns all registered service names
func (r *ModuleRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}
