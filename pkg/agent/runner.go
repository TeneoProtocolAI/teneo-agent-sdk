package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Teneo-Protocol/teneo-agent-sdk/pkg/auth"
	"github.com/Teneo-Protocol/teneo-agent-sdk/pkg/health"
	"github.com/Teneo-Protocol/teneo-agent-sdk/pkg/network"
	"github.com/Teneo-Protocol/teneo-agent-sdk/pkg/types"
)

// EnhancedAgent represents a fully functional Teneo network agent with all capabilities
type EnhancedAgent struct {
	config          *Config
	agentHandler    types.AgentHandler
	authManager     *auth.Manager
	networkClient   *network.NetworkClient
	protocolHandler *network.ProtocolHandler
	taskCoordinator *network.TaskCoordinator
	healthServer    *health.Server
	running         bool
	startTime       time.Time
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// EnhancedAgentConfig represents configuration for the enhanced agent
type EnhancedAgentConfig struct {
	Config       *Config
	AgentHandler types.AgentHandler
}

// NewEnhancedAgent creates a new enhanced agent with network capabilities
func NewEnhancedAgent(config *EnhancedAgentConfig) (*EnhancedAgent, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if config.AgentHandler == nil {
		return nil, fmt.Errorf("agent handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	agent := &EnhancedAgent{
		config:       config.Config,
		agentHandler: config.AgentHandler,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Initialize authentication manager
	authManager, err := auth.NewManager(config.Config.PrivateKey)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}
	agent.authManager = authManager

	// Initialize network client
	networkConfig := &network.Config{
		WebSocketURL:     config.Config.WebSocketURL,
		ReconnectEnabled: config.Config.ReconnectEnabled,
		ReconnectDelay:   config.Config.ReconnectDelay,
		MaxReconnects:    config.Config.MaxReconnects,
		MessageTimeout:   config.Config.MessageTimeout,
		PingInterval:     config.Config.PingInterval,
		HandshakeTimeout: config.Config.HandshakeTimeout,
	}
	agent.networkClient = network.NewNetworkClient(networkConfig)

	// Initialize protocol handler
	agent.protocolHandler = network.NewProtocolHandler(
		agent.networkClient,
		authManager,
		config.Config.Name,
		config.Config.Capabilities,
		authManager.GetAddress(),
	)

	// Initialize task coordinator
	agent.taskCoordinator = network.NewTaskCoordinator(
		config.AgentHandler,
		agent.protocolHandler,
		config.Config.Capabilities,
	)

	// Initialize health server if enabled
	if config.Config.HealthEnabled {
		agentInfo := &health.AgentInfo{
			Name:         config.Config.Name,
			Version:      config.Config.Version,
			Wallet:       authManager.GetAddress(),
			Capabilities: config.Config.Capabilities,
			Description:  config.Config.Description,
		}

		agent.healthServer = health.NewServer(
			config.Config.HealthPort,
			agentInfo,
			agent,
		)
	}

	return agent, nil
}

// Start starts the enhanced agent with all its components
func (a *EnhancedAgent) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return fmt.Errorf("agent is already running")
	}

	a.startTime = time.Now()
	a.running = true

	log.Printf("🚀 Starting enhanced agent: %s v%s", a.config.Name, a.config.Version)
	log.Printf("💼 Wallet: %s", a.authManager.GetAddress())
	log.Printf("🔧 Capabilities: %v", a.config.Capabilities)

	// Initialize agent handler if it supports initialization
	if initializer, ok := a.agentHandler.(types.AgentInitializer); ok {
		if err := initializer.Initialize(a.ctx, a.config); err != nil {
			a.running = false
			return fmt.Errorf("failed to initialize agent handler: %w", err)
		}
	}

	// Start health server if enabled
	if a.healthServer != nil {
		go func() {
			log.Printf("🌐 Starting health monitoring on port %d", a.config.HealthPort)
			if err := a.healthServer.Start(); err != nil {
				log.Printf("❌ Health server error: %v", err)
			}
		}()
	}

	// Connect to network
	if err := a.networkClient.Connect(); err != nil {
		a.running = false
		return fmt.Errorf("failed to connect to network: %w", err)
	}

	// Start authentication process
	if err := a.protocolHandler.StartAuthentication(); err != nil {
		log.Printf("⚠️ Authentication failed, will retry: %v", err)
	}

	// Start periodic tasks
	go a.startPeriodicTasks()

	log.Printf("✅ Enhanced agent %s started successfully", a.config.Name)
	return nil
}

// Stop gracefully stops the enhanced agent
func (a *EnhancedAgent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	log.Printf("🛑 Stopping enhanced agent: %s", a.config.Name)

	a.running = false
	a.cancel()

	// Cancel all active tasks
	a.taskCoordinator.CancelAllTasks()

	// Stop health server
	if a.healthServer != nil {
		if err := a.healthServer.Stop(); err != nil {
			log.Printf("⚠️ Error stopping health server: %v", err)
		}
	}

	// Disconnect from network
	if err := a.networkClient.Disconnect(); err != nil {
		log.Printf("⚠️ Error disconnecting from network: %v", err)
	}

	// Cleanup agent handler if it supports cleanup
	if cleaner, ok := a.agentHandler.(types.AgentCleaner); ok {
		if err := cleaner.Cleanup(a.ctx); err != nil {
			log.Printf("⚠️ Error cleaning up agent handler: %v", err)
		}
	}

	log.Printf("✅ Enhanced agent %s stopped successfully", a.config.Name)
	return nil
}

// Run runs the agent until interrupted
func (a *EnhancedAgent) Run() error {
	if err := a.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("📡 Received interrupt signal")

	return a.Stop()
}

// startPeriodicTasks starts periodic maintenance tasks
func (a *EnhancedAgent) startPeriodicTasks() {
	// Send periodic pings
	pingTicker := time.NewTicker(a.config.PingInterval)
	defer pingTicker.Stop()

	// Health checks
	healthTicker := time.NewTicker(30 * time.Second)
	defer healthTicker.Stop()

	// Status reporting
	statusTicker := time.NewTicker(5 * time.Minute)
	defer statusTicker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-pingTicker.C:
			if a.networkClient.IsConnected() && a.networkClient.IsAuthenticated() {
				if err := a.protocolHandler.SendPing(); err != nil {
					log.Printf("⚠️ Failed to send ping: %v", err)
				}
			}
		case <-healthTicker.C:
			// Perform health checks
			a.performHealthCheck()
		case <-statusTicker.C:
			// Log status
			a.logStatus()
		}
	}
}

// performHealthCheck performs periodic health checks
func (a *EnhancedAgent) performHealthCheck() {
	if !a.networkClient.IsConnected() {
		log.Printf("⚠️ Network disconnected, attempting reconnection...")
		if err := a.networkClient.Connect(); err != nil {
			log.Printf("❌ Reconnection failed: %v", err)
		}
	}

	if a.networkClient.IsConnected() && !a.networkClient.IsAuthenticated() {
		log.Printf("⚠️ Not authenticated, attempting authentication...")
		if err := a.protocolHandler.StartAuthentication(); err != nil {
			log.Printf("❌ Authentication failed: %v", err)
		}
	}
}

// logStatus logs the current agent status
func (a *EnhancedAgent) logStatus() {
	activeTasks := a.taskCoordinator.GetActiveTaskCount()
	uptime := time.Since(a.startTime)

	log.Printf("📊 Status - Connected: %v, Authenticated: %v, Active Tasks: %d, Uptime: %v",
		a.networkClient.IsConnected(),
		a.networkClient.IsAuthenticated(),
		activeTasks,
		uptime.Round(time.Second),
	)
}

// IsConnected implements the health.StatusGetter interface
func (a *EnhancedAgent) IsConnected() bool {
	return a.networkClient.IsConnected()
}

// IsAuthenticated implements the health.StatusGetter interface
func (a *EnhancedAgent) IsAuthenticated() bool {
	return a.networkClient.IsAuthenticated()
}

// GetActiveTaskCount implements the health.StatusGetter interface
func (a *EnhancedAgent) GetActiveTaskCount() int {
	return a.taskCoordinator.GetActiveTaskCount()
}

// GetUptime implements the health.StatusGetter interface
func (a *EnhancedAgent) GetUptime() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.running {
		return 0
	}

	return time.Since(a.startTime)
}

// GetConfig returns the agent configuration
func (a *EnhancedAgent) GetConfig() *Config {
	return a.config
}

// GetNetworkClient returns the network client
func (a *EnhancedAgent) GetNetworkClient() *network.NetworkClient {
	return a.networkClient
}

// GetTaskCoordinator returns the task coordinator
func (a *EnhancedAgent) GetTaskCoordinator() *network.TaskCoordinator {
	return a.taskCoordinator
}

// GetAuthManager returns the auth manager
func (a *EnhancedAgent) GetAuthManager() *auth.Manager {
	return a.authManager
}

// IsRunning returns whether the agent is currently running
func (a *EnhancedAgent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// UpdateCapabilities updates the agent's capabilities at runtime
func (a *EnhancedAgent) UpdateCapabilities(capabilities []string) {
	a.config.Capabilities = capabilities
	a.taskCoordinator.UpdateCapabilities(capabilities)

	if a.healthServer != nil {
		agentInfo := &health.AgentInfo{
			Name:         a.config.Name,
			Version:      a.config.Version,
			Wallet:       a.authManager.GetAddress(),
			Capabilities: capabilities,
			Description:  a.config.Description,
		}
		a.healthServer.UpdateAgentInfo(agentInfo)
	}

	log.Printf("🔄 Updated capabilities: %v", capabilities)
}
