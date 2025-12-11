package esl

import (
	"fmt"
	"sync"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	log "github.com/sirupsen/logrus"
)

// Client represents an inbound ESL connection to FreeSWITCH
type Client struct {
	Host     string
	Port     int
	Password string
	conn     *eventsocket.Connection
	events   chan *eventsocket.Event
	errors   chan error
	done     chan struct{}
	mu       sync.RWMutex
	running  bool
}

// NewClient creates a new ESL client
func NewClient(host string, port int, password string) *Client {
	return &Client{
		Host:     host,
		Port:     port,
		Password: password,
		events:   make(chan *eventsocket.Event, 1000),
		errors:   make(chan error, 10),
		done:     make(chan struct{}),
	}
}

// Connect establishes connection to FreeSWITCH
func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	log.Infof("Connecting to FreeSWITCH ESL at %s", addr)

	conn, err := eventsocket.Dial(addr, c.Password)
	if err != nil {
		return fmt.Errorf("failed to connect to FreeSWITCH: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.running = true
	c.mu.Unlock()

	log.Info("Connected to FreeSWITCH ESL")

	return nil
}

// Subscribe subscribes to FreeSWITCH events
func (c *Client) Subscribe(eventTypes ...string) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	// Build event subscription string
	events := "plain"
	for _, et := range eventTypes {
		events += " " + et
	}

	_, err := conn.Send(fmt.Sprintf("event %s", events))
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	log.Infof("Subscribed to events: %v", eventTypes)

	return nil
}

// StartEventLoop starts reading events from FreeSWITCH
func (c *Client) StartEventLoop() {
	go c.eventLoop()
}

func (c *Client) eventLoop() {
	for {
		c.mu.RLock()
		conn := c.conn
		running := c.running
		c.mu.RUnlock()

		if !running || conn == nil {
			return
		}

		ev, err := conn.ReadEvent()
		if err != nil {
			select {
			case <-c.done:
				return
			default:
				c.errors <- err
				// Try to reconnect
				c.reconnect()
			}
			continue
		}

		select {
		case c.events <- ev:
		case <-c.done:
			return
		default:
			log.Warn("Event channel full, dropping event")
		}
	}
}

func (c *Client) reconnect() {
	c.mu.Lock()
	c.running = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(i+1) * time.Second)

		log.Infof("Attempting to reconnect to FreeSWITCH (attempt %d)", i+1)

		if err := c.Connect(); err != nil {
			log.Warnf("Reconnect failed: %v", err)
			continue
		}

		log.Info("Reconnected to FreeSWITCH")
		return
	}

	log.Error("Failed to reconnect to FreeSWITCH after 10 attempts")
}

// Events returns the event channel
func (c *Client) Events() <-chan *eventsocket.Event {
	return c.events
}

// Errors returns the error channel
func (c *Client) Errors() <-chan error {
	return c.errors
}

// Send sends a command to FreeSWITCH
func (c *Client) Send(command string) (*eventsocket.Event, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	return conn.Send(command)
}

// API executes an API command
func (c *Client) API(command string) (string, error) {
	ev, err := c.Send(fmt.Sprintf("api %s", command))
	if err != nil {
		return "", err
	}
	return ev.Body, nil
}

// BgAPI executes a background API command
func (c *Client) BgAPI(command string) (string, error) {
	ev, err := c.Send(fmt.Sprintf("bgapi %s", command))
	if err != nil {
		return "", err
	}
	return ev.Get("Job-UUID"), nil
}

// Originate starts a new call
func (c *Client) Originate(dialString, app, appArgs string) (string, error) {
	cmd := fmt.Sprintf("originate %s &%s(%s)", dialString, app, appArgs)
	return c.BgAPI(cmd)
}

// Close closes the ESL connection
func (c *Client) Close() {
	close(c.done)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// IsConnected returns connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && c.running
}
