package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/teneo/agent-sdk-go/pkg/types"
)

// NetworkClient handles WebSocket communication for Teneo agents
type NetworkClient struct {
	conn            *websocket.Conn
	url             string
	messageHandlers map[string]MessageHandler
	reconnector     *ReconnectionManager
	authenticated   bool
	running         bool
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	sendChan        chan *types.Message
	receiveChan     chan *types.Message
}

// MessageHandler defines the function signature for message handlers
type MessageHandler func(*types.Message) error

// Config represents network configuration
type Config struct {
	WebSocketURL     string
	ReconnectEnabled bool
	ReconnectDelay   time.Duration
	MaxReconnects    int
	MessageTimeout   time.Duration
	PingInterval     time.Duration
	HandshakeTimeout time.Duration
}

// DefaultNetworkConfig returns default network configuration
func DefaultNetworkConfig() *Config {
	return &Config{
		WebSocketURL:     "ws://localhost:8090/ws",
		ReconnectEnabled: true,
		ReconnectDelay:   5 * time.Second,
		MaxReconnects:    10,
		MessageTimeout:   30 * time.Second,
		PingInterval:     30 * time.Second,
		HandshakeTimeout: 10 * time.Second,
	}
}

// NewNetworkClient creates a new network client
func NewNetworkClient(config *Config) *NetworkClient {
	ctx, cancel := context.WithCancel(context.Background())

	client := &NetworkClient{
		url:             config.WebSocketURL,
		messageHandlers: make(map[string]MessageHandler),
		authenticated:   false,
		running:         false,
		ctx:             ctx,
		cancel:          cancel,
		sendChan:        make(chan *types.Message, 100),
		receiveChan:     make(chan *types.Message, 100),
	}

	client.reconnector = &ReconnectionManager{
		enabled:     config.ReconnectEnabled,
		maxAttempts: config.MaxReconnects,
		delay:       config.ReconnectDelay,
		backoffFunc: exponentialBackoff,
	}

	return client
}

// Connect establishes WebSocket connection
func (c *NetworkClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("client is already running")
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	c.running = true
	c.authenticated = false

	// Start message processing goroutines
	go c.readMessages()
	go c.writeMessages()
	go c.processMessages()

	log.Printf("🔗 Connected to WebSocket server: %s", c.url)
	return nil
}

// Disconnect closes the WebSocket connection
func (c *NetworkClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	c.running = false
	c.authenticated = false
	c.cancel()

	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		c.conn = nil
	}

	log.Println("🔌 Disconnected from WebSocket server")
	return nil
}

// SendMessage sends a message through the WebSocket connection
func (c *NetworkClient) SendMessage(msg *types.Message) error {
	c.mu.RLock()
	if !c.running {
		c.mu.RUnlock()
		return fmt.Errorf("client is not running")
	}
	c.mu.RUnlock()

	select {
	case c.sendChan <- msg:
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("client is shutting down")
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout")
	}
}

// SendRawData sends raw JSON data directly via WebSocket (for compatibility with server expectations)
func (c *NetworkClient) SendRawData(data []byte) error {
	c.mu.RLock()
	if !c.running || c.conn == nil {
		c.mu.RUnlock()
		return fmt.Errorf("client is not running or not connected")
	}
	conn := c.conn
	c.mu.RUnlock()

	return conn.WriteMessage(1, data) // 1 = TextMessage
}

// RegisterHandler registers a message handler for a specific message type
func (c *NetworkClient) RegisterHandler(msgType string, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageHandlers[msgType] = handler
}

// IsConnected returns whether the client is connected
func (c *NetworkClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running && c.conn != nil
}

// IsAuthenticated returns whether the client is authenticated
func (c *NetworkClient) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authenticated
}

// SetAuthenticated sets the authentication status
func (c *NetworkClient) SetAuthenticated(authenticated bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authenticated = authenticated
}

// readMessages reads messages from WebSocket connection
func (c *NetworkClient) readMessages() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic in readMessages: %v", r)
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if c.conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			_, messageData, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("❌ Read error: %v", err)
				if c.reconnector.enabled && c.reconnector.ShouldReconnect() {
					go c.attemptReconnection()
				}
				return
			}

			var msg types.Message
			if err := json.Unmarshal(messageData, &msg); err != nil {
				log.Printf("❌ Failed to unmarshal message: %v", err)
				continue
			}

			select {
			case c.receiveChan <- &msg:
			case <-c.ctx.Done():
				return
			}
		}
	}
}

// writeMessages writes messages to WebSocket connection
func (c *NetworkClient) writeMessages() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic in writeMessages: %v", r)
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.sendChan:
			if c.conn == nil {
				continue
			}

			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("❌ Failed to marshal message: %v", err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("❌ Write error: %v", err)
				if c.reconnector.enabled && c.reconnector.ShouldReconnect() {
					go c.attemptReconnection()
				}
				return
			}
		}
	}
}

// processMessages processes incoming messages
func (c *NetworkClient) processMessages() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic in processMessages: %v", r)
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.receiveChan:
			if handler, exists := c.messageHandlers[msg.Type]; exists {
				if err := handler(msg); err != nil {
					log.Printf("❌ Handler error for message type %s: %v", msg.Type, err)
				}
			} else {
				log.Printf("⚠️  No handler for message type: %s", msg.Type)
			}
		}
	}
}

// attemptReconnection attempts to reconnect to the WebSocket server
func (c *NetworkClient) attemptReconnection() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.reconnector.ShouldReconnect() {
		log.Printf("❌ Max reconnection attempts reached, giving up")
		return
	}

	c.reconnector.attempts++
	backoff := c.reconnector.NextBackoff()

	log.Printf("🔄 Reconnection attempt %d/%d in %v...",
		c.reconnector.attempts, c.reconnector.maxAttempts, backoff)

	time.Sleep(backoff)

	if err := c.reconnect(); err != nil {
		log.Printf("❌ Reconnection failed: %v", err)
		go c.attemptReconnection()
	} else {
		log.Printf("✅ Reconnected successfully")
		c.reconnector.Reset()
	}
}

// reconnect performs the actual reconnection
func (c *NetworkClient) reconnect() error {
	// Close existing connection
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.running = false
	c.authenticated = false

	// Create new connection
	return c.Connect()
}

// exponentialBackoff calculates exponential backoff delay
func exponentialBackoff(attempt int) time.Duration {
	delay := time.Duration(attempt) * 5 * time.Second
	if delay > 60*time.Second {
		delay = 60 * time.Second
	}
	return delay
}
