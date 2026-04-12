package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/gorilla/websocket"
)

// NetworkClient handles WebSocket communication for Teneo agents
type NetworkClient struct {
	conn              *websocket.Conn
	url               string
	additionalHeaders http.Header
	messageHandlers   map[string]MessageHandler
	reconnector       *ReconnectionManager
	authenticated     bool
	running           bool
	reconnecting      int32 // atomic flag for reconnection state
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	sendChan          chan *types.Message
	receiveChan       chan *types.Message
	wg                sync.WaitGroup // For goroutine lifecycle management

	// Resilience components
	circuitBreaker *CircuitBreaker
	retryQueue     *MessageRetryQueue
	healthMonitor  *HealthMonitor
	supervisor     *GoroutineSupervisor
}

// MessageHandler defines the function signature for message handlers
type MessageHandler func(*types.Message) error

// Config represents network configuration
type Config struct {
	WebSocketURL       string
	ReconnectEnabled   bool
	ReconnectDelay     time.Duration
	MaxReconnects      int
	MessageTimeout     time.Duration
	PingInterval       time.Duration
	HandshakeTimeout   time.Duration
	AdditionalHeaders  map[string]string // Extra HTTP headers sent during WebSocket handshake
}

// DefaultNetworkConfig returns default network configuration
func DefaultNetworkConfig() *Config {
	return &Config{
		WebSocketURL:     "ws://localhost:8090/ws",
		ReconnectEnabled: true,
		ReconnectDelay:   5 * time.Second,
		MaxReconnects:    0, // 0 (or negative) = retry forever
		MessageTimeout:   30 * time.Second,
		PingInterval:     30 * time.Second,
		HandshakeTimeout: 10 * time.Second,
	}
}

// NewNetworkClient creates a new network client
func NewNetworkClient(config *Config) *NetworkClient {
	ctx, cancel := context.WithCancel(context.Background())

	// Build HTTP headers from AdditionalHeaders map
	var headers http.Header
	if len(config.AdditionalHeaders) > 0 {
		headers = make(http.Header, len(config.AdditionalHeaders))
		for k, v := range config.AdditionalHeaders {
			headers.Set(k, v)
		}
	}

	client := &NetworkClient{
		url:               config.WebSocketURL,
		additionalHeaders: headers,
		messageHandlers:   make(map[string]MessageHandler),
		authenticated:     false,
		running:           false,
		ctx:               ctx,
		cancel:            cancel,
		sendChan:          make(chan *types.Message, 100),
		receiveChan:       make(chan *types.Message, 100),
	}

	client.reconnector = &ReconnectionManager{
		enabled:     config.ReconnectEnabled,
		maxAttempts: config.MaxReconnects,
		delay:       config.ReconnectDelay,
		backoffFunc: exponentialBackoff,
	}

	// Initialize resilience components
	client.circuitBreaker = NewCircuitBreaker(3, 30*time.Second)
	client.circuitBreaker.SetStateChangeHandler(func(from, to CircuitState) {
		log.Printf("🔌 Circuit breaker state changed: %s → %s", from, to)
	})

	client.retryQueue = NewMessageRetryQueue(DefaultRetryPolicy(), client.sendMessageDirect)

	client.healthMonitor = NewHealthMonitor(10 * time.Second)
	client.healthMonitor.SetHealthCheckFunc(client.healthCheck)
	client.healthMonitor.SetStatusChangeHandler(func(old, new HealthStatus) {
		log.Printf("🏥 Health status changed: %s → %s", old, new)
	})

	client.supervisor = NewGoroutineSupervisor(ctx)

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

	conn, _, err := dialer.Dial(c.url, c.additionalHeaders)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	c.running = true
	c.authenticated = false

	// Set up pong handler to respond to server pings
	c.conn.SetPongHandler(func(appData string) error {
		// Reset read deadline when we receive a pong
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Set initial read deadline
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Register and start supervised goroutines
	c.registerGoroutines()
	if err := c.supervisor.Start(); err != nil {
		return fmt.Errorf("failed to start supervisor: %w", err)
	}

	// Start resilience components
	c.retryQueue.Start()
	c.healthMonitor.Start()
	c.healthMonitor.RecordConnectionEstablished()

	log.Printf("🔗 Connected to WebSocket server: %s", c.url)
	return nil
}

// Disconnect closes the WebSocket connection with graceful shutdown
func (c *NetworkClient) Disconnect() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}

	c.running = false
	c.authenticated = false
	oldConn := c.conn
	c.conn = nil
	c.mu.Unlock()

	// Stop resilience components
	c.supervisor.Stop()
	c.retryQueue.Stop()
	c.healthMonitor.Stop()
	c.healthMonitor.RecordConnectionLost()

	// Send close message
	if oldConn != nil {
		oldConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		oldConn.Close()
	}

	// Cancel context and wait for goroutines
	c.cancel()

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ All goroutines stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Println("⚠️ Timeout waiting for goroutines to stop")
	}

	log.Println("🔌 Disconnected from WebSocket server")
	return nil
}

// SendMessage sends a message through the WebSocket connection with retry support
func (c *NetworkClient) SendMessage(msg *types.Message) error {
	// Use circuit breaker
	return c.circuitBreaker.Call(func() error {
		err := c.sendMessageDirect(msg)
		if err != nil {
			// Queue for retry if failed
			c.retryQueue.Enqueue(msg, err)
			c.healthMonitor.RecordMessageFailed()
		}
		return err
	})
}

// sendMessageDirect sends a message directly without retry logic
func (c *NetworkClient) sendMessageDirect(msg *types.Message) error {
	c.mu.RLock()
	if !c.running {
		c.mu.RUnlock()
		return fmt.Errorf("client is not running")
	}
	c.mu.RUnlock()

	select {
	case c.sendChan <- msg:
		c.healthMonitor.RecordMessageSent()
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

			// Set read deadline before reading
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			_, messageData, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("❌ Read error: %v", err)
				if c.reconnector.enabled && atomic.CompareAndSwapInt32(&c.reconnecting, 0, 1) {
					go c.attemptReconnection()
				}
				return
			}

			var msg types.Message
			if err := json.Unmarshal(messageData, &msg); err != nil {
				log.Printf("❌ Failed to unmarshal message: %v", err)
				continue
			}

			// Record successful message receipt
			c.healthMonitor.RecordMessageReceived()

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
				if c.reconnector.enabled && atomic.CompareAndSwapInt32(&c.reconnecting, 0, 1) {
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
	defer atomic.StoreInt32(&c.reconnecting, 0) // Reset flag when done

	// Check without holding lock
	if !c.reconnector.ShouldReconnect() {
		log.Printf("❌ Max reconnection attempts reached, giving up")
		c.healthMonitor.RecordReconnectAttempt(false)
		return
	}

	// Increment attempts (minimal lock time)
	c.mu.Lock()
	c.reconnector.attempts++
	backoff := c.reconnector.NextBackoff()
	c.mu.Unlock()

	if c.reconnector.maxAttempts <= 0 {
		log.Printf("🔄 Reconnection attempt %d (∞) in %v...",
			c.reconnector.attempts, backoff)
	} else {
		log.Printf("🔄 Reconnection attempt %d/%d in %v...",
			c.reconnector.attempts, c.reconnector.maxAttempts, backoff)
	}

	// Sleep without holding lock
	time.Sleep(backoff)

	// Attempt reconnection
	if err := c.reconnect(); err != nil {
		log.Printf("❌ Reconnection failed: %v", err)
		c.healthMonitor.RecordReconnectAttempt(false)

		// Try again if we haven't exceeded max attempts
		if c.reconnector.ShouldReconnect() {
			atomic.StoreInt32(&c.reconnecting, 0) // Reset flag before next attempt
			go c.attemptReconnection()
		}
	} else {
		log.Printf("✅ Reconnected successfully")
		c.reconnector.Reset()
		c.healthMonitor.RecordReconnectAttempt(true)
		c.healthMonitor.RecordConnectionEstablished()
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

	// Cancel existing context and create new one for fresh goroutines
	c.cancel()
	c.ctx, c.cancel = context.WithCancel(context.Background())

	// Establish new connection
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(c.url, c.additionalHeaders)
	if err != nil {
		return fmt.Errorf("failed to reconnect to WebSocket: %w", err)
	}

	c.conn = conn
	c.running = true
	c.authenticated = false

	// Set up pong handler to respond to server pings
	c.conn.SetPongHandler(func(appData string) error {
		// Reset read deadline when we receive a pong
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Set initial read deadline
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Restart message processing goroutines
	go c.readMessages()
	go c.writeMessages()
	go c.processMessages()
	go c.pingPongHandler()

	log.Printf("🔗 Reconnected to WebSocket server: %s", c.url)
	return nil
}

// pingPongHandler handles WebSocket ping/pong to keep connection alive
func (c *NetworkClient) pingPongHandler() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic in pingPongHandler: %v", r)
		}
	}()

	pingInterval := 25 * time.Second
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			running := c.running
			c.mu.RUnlock()

			if !running || conn == nil {
				continue
			}

			// Send ping message
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Printf("⚠️ Ping failed: %v", err)
				// Trigger reconnection if ping fails
				if c.reconnector.enabled && atomic.CompareAndSwapInt32(&c.reconnecting, 0, 1) {
					go c.attemptReconnection()
				}
				return
			}
		}
	}
}

// exponentialBackoff calculates exponential backoff delay
func exponentialBackoff(attempt int) time.Duration {
	delay := time.Duration(attempt) * 5 * time.Second
	if delay > 60*time.Second {
		delay = 60 * time.Second
	}
	return delay
}

// getConn returns the connection safely
func (c *NetworkClient) getConn() *websocket.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// healthCheck performs a health check for the health monitor
func (c *NetworkClient) healthCheck() error {
	conn := c.getConn()
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	c.mu.RLock()
	connected := c.running
	authenticated := c.authenticated
	c.mu.RUnlock()

	if !connected {
		return fmt.Errorf("not connected")
	}

	if !authenticated {
		return fmt.Errorf("not authenticated")
	}

	return nil
}

// registerGoroutines registers all goroutines with the supervisor
func (c *NetworkClient) registerGoroutines() {
	policy := DefaultRestartPolicy()

	// Register read messages goroutine
	c.supervisor.Register("read-messages", "Message Reader",
		func(ctx context.Context) error {
			c.wg.Add(1)
			defer c.wg.Done()
			c.readMessages()
			return nil
		}, policy)

	// Register write messages goroutine
	c.supervisor.Register("write-messages", "Message Writer",
		func(ctx context.Context) error {
			c.wg.Add(1)
			defer c.wg.Done()
			c.writeMessages()
			return nil
		}, policy)

	// Register process messages goroutine
	c.supervisor.Register("process-messages", "Message Processor",
		func(ctx context.Context) error {
			c.wg.Add(1)
			defer c.wg.Done()
			c.processMessages()
			return nil
		}, policy)

	// Register ping/pong handler
	c.supervisor.Register("ping-pong", "Ping/Pong Handler",
		func(ctx context.Context) error {
			c.wg.Add(1)
			defer c.wg.Done()
			c.pingPongHandler()
			return nil
		}, policy)
}

// GetHealthReport returns a health report for the connection
func (c *NetworkClient) GetHealthReport() string {
	return c.healthMonitor.GetHealthReport()
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (c *NetworkClient) GetCircuitBreakerStats() CircuitBreakerStats {
	return c.circuitBreaker.GetStats()
}

// GetRetryQueueMetrics returns retry queue metrics
func (c *NetworkClient) GetRetryQueueMetrics() RetryMetrics {
	return c.retryQueue.GetMetrics()
}

// GetSupervisorStatus returns the status of all supervised goroutines
func (c *NetworkClient) GetSupervisorStatus() map[string]GoroutineStatus {
	return c.supervisor.GetStatus()
}
