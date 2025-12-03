package agent

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/auth"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/cache"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/health"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/network"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/nft"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/ethereum/go-ethereum/crypto"
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
	agentCache      cache.AgentCache
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

	// NFT Minting Options
	Mint    bool   // If true, mint new NFT; if false, use TokenID
	TokenID uint64 // Required if Mint is false

	// Backend Configuration
	BackendURL  string // Default from env or "http://localhost:8080"
	RPCEndpoint string // Ethereum RPC endpoint
}

// NewEnhancedAgent creates a new enhanced agent with network capabilities
func NewEnhancedAgent(config *EnhancedAgentConfig) (*EnhancedAgent, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if config.AgentHandler == nil {
		return nil, fmt.Errorf("agent handler is required")
	}

	// Set default backend URL if not provided
	if config.BackendURL == "" {
		if backendURL := os.Getenv("BACKEND_URL"); backendURL != "" {
			config.BackendURL = backendURL
		} else {
			config.BackendURL = "http://localhost:8080"
		}
	}

	// Set default RPC endpoint if not provided
	if config.RPCEndpoint == "" {
		if rpcEndpoint := os.Getenv("RPC_ENDPOINT"); rpcEndpoint != "" {
			config.RPCEndpoint = rpcEndpoint
		}
	}

	// Handle NFT minting or verification
	if config.Mint {
		// Create NFT minter
		minter, err := nft.NewNFTMinter(config.BackendURL, config.RPCEndpoint, config.Config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create NFT minter: %w", err)
		}

		// Generate agent ID from name
		agentID := generateAgentID(config.Config.Name)

		// Prepare metadata
		metadata := nft.AgentMetadata{
			Name:         config.Config.Name,
			Description:  config.Config.Description,
			Image:        config.Config.Image,
			Capabilities: config.Config.Capabilities,
			AgentID:      agentID,
		}

		log.Printf("🎨 Minting NFT for agent: %s", config.Config.Name)

		// Mint NFT - this will:
		// 1. Send metadata to backend (backend uploads to IPFS)
		// 2. Get signature from backend
		// 3. Execute on-chain mint transaction
		tokenID, err := minter.MintAgent(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to mint NFT: %w", err)
		}

		config.TokenID = tokenID
		log.Printf("✅ Successfully minted NFT with token ID: %d", tokenID)

		// Store token ID in environment for future use
		os.Setenv("NFT_TOKEN_ID", fmt.Sprintf("%d", tokenID))
	} else {
		// Verify TokenID is set
		if config.TokenID == 0 {
			// Try to load from environment
			if tokenIDStr := os.Getenv("NFT_TOKEN_ID"); tokenIDStr != "" {
				if tokenID, err := fmt.Sscanf(tokenIDStr, "%d", &config.TokenID); err != nil || tokenID != 1 {
					return nil, fmt.Errorf("invalid NFT_TOKEN_ID in environment: %s", tokenIDStr)
				}
			} else {
				return nil, fmt.Errorf("TokenID must be provided when Mint is false")
			}
		}

		// Generate and send metadata hash
		metadata := nft.AgentMetadata{
			Name:         config.Config.Name,
			Description:  config.Config.Description,
			Image:        config.Config.Image,
			Capabilities: config.Config.Capabilities,
			AgentID:      generateAgentID(config.Config.Name),
		}

		hash := nft.GenerateMetadataHash(metadata)
		log.Printf("📋 Using existing NFT token ID: %d with metadata hash: %s", config.TokenID, hash)

		// Send metadata hash to backend
		minter, err := nft.NewNFTMinter(config.BackendURL, config.RPCEndpoint, config.Config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create NFT minter: %w", err)
		}

		walletAddress := getAddressFromPrivateKey(config.Config.PrivateKey)
		err = minter.SendMetadataHashToBackend(hash, config.TokenID, walletAddress)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to send metadata hash to backend: %v", err)
			// This is not critical, so we continue
		}
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
		config.Config.NFTTokenID,
		config.Config.Room,
	)

	// Initialize task coordinator
	agent.taskCoordinator = network.NewTaskCoordinator(
		config.AgentHandler,
		agent.protocolHandler,
		config.Config.Capabilities,
	)

	// Set rate limit if configured
	if config.Config.RateLimitPerMinute > 0 {
		agent.taskCoordinator.SetRateLimit(config.Config.RateLimitPerMinute)
	}

	// Initialize Redis cache if enabled
	if config.Config.RedisEnabled {
		log.Printf("🗄️  Initializing Redis cache at %s", config.Config.RedisAddress)

		// Set default key prefix if not provided
		keyPrefix := config.Config.RedisKeyPrefix
		if keyPrefix == "" {
			keyPrefix = fmt.Sprintf("teneo:agent:%s:", strings.ReplaceAll(strings.ToLower(config.Config.Name), " ", "_"))
		}

		redisConfig := &cache.RedisConfig{
			Address:   config.Config.RedisAddress,
			Username:  config.Config.RedisUsername,
			Password:  config.Config.RedisPassword,
			DB:        config.Config.RedisDB,
			KeyPrefix: keyPrefix,
			UseTLS:    config.Config.RedisUseTLS,
		}

		redisCache, err := cache.NewRedisCache(redisConfig)
		if err != nil {
			// Log error but don't fail - cache is optional
			log.Printf("⚠️  Failed to initialize Redis cache: %v (continuing without cache)", err)
			agent.agentCache = &cache.NoOpCache{}
		} else {
			agent.agentCache = redisCache
			log.Printf("✅ Redis cache initialized successfully with prefix: %s", keyPrefix)
		}
	} else {
		// Use no-op cache when Redis is disabled
		agent.agentCache = &cache.NoOpCache{}
	}

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

	// Connect to network with retry logic
	connectRetries := 3
	var connectErr error
	for i := 0; i < connectRetries; i++ {
		if err := a.networkClient.Connect(); err != nil {
			connectErr = err
			log.Printf("⚠️ Connection attempt %d/%d failed: %v", i+1, connectRetries, err)
			if i < connectRetries-1 {
				time.Sleep(time.Duration(i+1) * 2 * time.Second)
			}
		} else {
			connectErr = nil
			break
		}
	}

	if connectErr != nil {
		a.running = false
		return fmt.Errorf("failed to connect to network after %d attempts: %w", connectRetries, connectErr)
	}

	// Start authentication process with retry
	authRetries := 3
	var authErr error
	for i := 0; i < authRetries; i++ {
		if err := a.protocolHandler.StartAuthentication(); err != nil {
			authErr = err
			log.Printf("⚠️ Authentication attempt %d/%d failed: %v", i+1, authRetries, err)
			if i < authRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		} else {
			authErr = nil
			break
		}
	}

	if authErr != nil {
		log.Printf("⚠️ Authentication failed after %d attempts, will retry periodically: %v", authRetries, authErr)
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


        // NOW cancel the context
        a.cancel()
	// Close cache connection
	if a.agentCache != nil {
		if err := a.agentCache.Close(); err != nil {
			log.Printf("⚠️ Error closing cache connection: %v", err)
		}
	}

	// Cleanup agent handler with fresh context
	if cleaner, ok := a.agentHandler.(types.AgentCleaner); ok {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if err := cleaner.Cleanup(cleanupCtx); err != nil {
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

// GetCache returns the agent cache instance
// This allows agent implementations to access the cache for persistent storage
func (a *EnhancedAgent) GetCache() cache.AgentCache {
	return a.agentCache
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

// generateAgentID generates a unique agent ID from the agent name
func generateAgentID(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	agentID := strings.ToLower(name)
	agentID = strings.ReplaceAll(agentID, " ", "-")
	// Remove any characters that aren't lowercase letters, numbers, or hyphens
	result := ""
	for _, char := range agentID {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result += string(char)
		}
	}
	return result
}

// getAddressFromPrivateKey derives the Ethereum address from a private key
func getAddressFromPrivateKey(privateKeyHex string) string {
	// Import crypto package
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return ""
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return ""
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex()
}
