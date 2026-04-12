package agent

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/alerting"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/auth"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/cache"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/deploy"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/health"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/network"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/nft"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// EnhancedAgent represents a fully functional Teneo network agent with all capabilities
type EnhancedAgent struct {
	config               *Config
	agentHandler         types.AgentHandler
	authManager          *auth.Manager
	networkClient        *network.NetworkClient
	protocolHandler      *network.ProtocolHandler
	taskCoordinator      *network.TaskCoordinator
	healthServer         *health.Server
	agentCache           cache.AgentCache
	alerter              *alerting.SlackAlerter
	backendURL           string
	agentID              string
	additionalHeaders    map[string]string
	submitForReviewOnRun bool
	running              bool
	startTime            time.Time
	mu                   sync.RWMutex
	ctx                  context.Context
	cancel               context.CancelFunc
}

// EnhancedAgentConfig represents configuration for the enhanced agent
type EnhancedAgentConfig struct {
	Config       *Config
	AgentHandler types.AgentHandler

	// NFT Minting Options (choose one: Deploy, Mint, or provide TokenID)
	Deploy  bool   // If true, use new secure deploy flow with database persistence
	Mint    bool   // If true, use legacy mint flow (no database persistence)
	TokenID uint64 // Required if Deploy and Mint are both false

	// Deploy-specific options
	AgentID       string // Required for Deploy, auto-generated from name if empty
	AgentType     string // Agent type: "command", "nlp", "mcp", "commandless" (default: "command")
	StateFilePath string // Path to state file for Deploy (default: .teneo-deploy-state.json)

	// Backend Configuration
	BackendURL  string // Default from env or "http://localhost:8080"
	RPCEndpoint string // Ethereum RPC endpoint

	// Optional: Submit agent for public visibility review after startup (defaults to false).
	// The agent goes through a review process (up to 72 hours) before becoming publicly visible.
	SubmitForReview bool
}

// NewEnhancedAgent creates a new enhanced agent with network capabilities
func NewEnhancedAgent(config *EnhancedAgentConfig) (*EnhancedAgent, error) {
	// Show EULA and deployment rules links at startup
	printEULALinks()

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
		} else if config.Config.WebSocketURL != "" {
			// Derive backend URL from WebSocket URL (strip /ws, wss->https, ws->http)
			derived := strings.TrimSuffix(config.Config.WebSocketURL, "/ws")
			derived = strings.Replace(derived, "wss://", "https://", 1)
			derived = strings.Replace(derived, "ws://", "http://", 1)
			config.BackendURL = derived
		} else {
			config.BackendURL = "https://backend.developer.chatroom.teneo-protocol.ai"
		}
	}

	// Set default RPC endpoint if not provided
	if config.RPCEndpoint == "" {
		if rpcEndpoint := os.Getenv("RPC_ENDPOINT"); rpcEndpoint != "" {
			config.RPCEndpoint = rpcEndpoint
		}
	}

	// Handle NFT deployment/minting
	if config.Deploy {
		// Use the new secure deploy flow with authentication and database persistence
		log.Printf("🚀 Deploying agent using secure SDK flow: %s", config.Config.Name)

		agentID := config.AgentID
		if agentID == "" {
			agentID = config.Config.AgentID
		}
		if agentID == "" {
			return nil, fmt.Errorf("agent_id is required: set AgentID on EnhancedAgentConfig or Config.AgentID")
		}

		// Build capabilities JSON
		capabilitiesJSON, err := buildCapabilitiesJSON(config.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to build capabilities JSON: %w", err)
		}

		// Determine agent type (default to "command" for backwards compatibility)
		agentType := config.AgentType
		if agentType == "" {
			agentType = "command"
		}

		// Create deploy configuration
		deployCfg := deploy.DeployConfig{
			BackendURL:       config.BackendURL,
			RPCEndpoint:      config.RPCEndpoint,
			PrivateKey:       config.Config.PrivateKey,
			AgentID:          agentID,
			AgentName:        config.Config.Name,
			Description:      config.Config.Description,
			Image:            config.Config.Image,
			AgentType:        agentType,
			Capabilities:     capabilitiesJSON,
			ShortDescription: config.Config.ShortDescription,
			TutorialURL:      config.Config.TutorialURL,
			StateFilePath:    config.StateFilePath,
			MetadataVersion:  "2.4.0",
		}

		// Execute deployment
		result, err := deploy.DeployAgent(deployCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to deploy agent: %w", err)
		}

		config.TokenID = result.TokenID
		if result.AlreadyMinted {
			log.Printf("✅ Agent was already deployed - Token ID: %d", result.TokenID)
		} else {
			log.Printf("✅ Successfully deployed agent - Token ID: %d, Tx: %s", result.TokenID, result.TxHash)
		}

		// Store token ID in environment and config for future use
		os.Setenv("NFT_TOKEN_ID", fmt.Sprintf("%d", result.TokenID))
		config.Config.NFTTokenID = fmt.Sprintf("%d", result.TokenID)
	} else if config.Mint {
		// Legacy mint flag — redirect to gasless deploy flow
		log.Printf("🎨 Minting NFT for agent (gasless): %s", config.Config.Name)
		minter, err := nft.NewNFTMinter(config.Config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create NFT minter: %w", err)
		}

		agentID := config.AgentID
		if agentID == "" {
			agentID = config.Config.AgentID
		}
		if agentID == "" {
			return nil, fmt.Errorf("agent_id is required: set AgentID on EnhancedAgentConfig or Config.AgentID")
		}
		metadata := nft.AgentMetadata{
			Name:         config.Config.Name,
			Description:  config.Config.Description,
			Image:        config.Config.Image,
			Capabilities: config.Config.ResolveCapabilities(),
			AgentID:      agentID,
		}

		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}

		result, err := minter.MintOrResumeFromJSON(metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to mint NFT: %w", err)
		}

		config.TokenID = result.TokenID
		log.Printf("✅ Successfully minted NFT with token ID: %d", result.TokenID)

		os.Setenv("NFT_TOKEN_ID", fmt.Sprintf("%d", result.TokenID))
		config.Config.NFTTokenID = fmt.Sprintf("%d", result.TokenID)
	} else {
		// Verify TokenID is set
		if config.TokenID == 0 {
			// Try to load from environment
			if tokenIDStr := os.Getenv("NFT_TOKEN_ID"); tokenIDStr != "" {
				// fmt.Sscanf returns count of items parsed (should be 1 for success)
				if n, err := fmt.Sscanf(tokenIDStr, "%d", &config.TokenID); err != nil || n != 1 {
					return nil, fmt.Errorf("invalid NFT_TOKEN_ID in environment: %s", tokenIDStr)
				}
			} else {
				return nil, fmt.Errorf("TokenID must be provided when Mint is false")
			}
		}

		// Propagate TokenID to config so WebSocket auth includes it
		config.Config.NFTTokenID = fmt.Sprintf("%d", config.TokenID)

		// Generate and send metadata hash
		metadata := nft.AgentMetadata{
			Name:         config.Config.Name,
			Description:  config.Config.Description,
			Image:        config.Config.Image,
			Capabilities: config.Config.ResolveCapabilities(),
			AgentID:      config.AgentID,
		}

		hash := nft.GenerateMetadataHash(metadata)
		log.Printf("📋 Using existing NFT token ID: %d with metadata hash: %s", config.TokenID, hash)

		// Send metadata hash to backend
		minter, err := nft.NewNFTMinterWithConfig(config.BackendURL, config.RPCEndpoint, config.Config.PrivateKey)
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

	// Auto-accept EULA if ACCEPT_EULA=true
	if strings.EqualFold(os.Getenv("ACCEPT_EULA"), "true") {
		log.Printf("📋 Checking EULA acceptance status...")
		if err := checkAndAcceptEULA(config.BackendURL, config.Config.PrivateKey); err != nil {
			return nil, fmt.Errorf("EULA acceptance failed: %w", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Resolve agent ID
	resolvedAgentID := config.AgentID
	if resolvedAgentID == "" {
		resolvedAgentID = config.Config.AgentID
	}

	agent := &EnhancedAgent{
		config:               config.Config,
		agentHandler:         config.AgentHandler,
		backendURL:           config.BackendURL,
		agentID:              resolvedAgentID,
		additionalHeaders:    config.Config.AdditionalHeaders,
		submitForReviewOnRun: config.SubmitForReview,
		ctx:                  ctx,
		cancel:               cancel,
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
		WebSocketURL:      config.Config.WebSocketURL,
		ReconnectEnabled:  config.Config.ReconnectEnabled,
		ReconnectDelay:    config.Config.ReconnectDelay,
		MaxReconnects:     config.Config.MaxReconnects,
		MessageTimeout:    config.Config.MessageTimeout,
		PingInterval:      config.Config.PingInterval,
		HandshakeTimeout:  config.Config.HandshakeTimeout,
		AdditionalHeaders: config.Config.AdditionalHeaders,
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

	// Initialize Slack alerter if configured
	if config.Config.SlackWebhookURL != "" {
		agent.alerter = alerting.NewSlackAlerter(alerting.SlackConfig{
			WebhookURL:      config.Config.SlackWebhookURL,
			AgentName:       config.Config.Name,
			AgentWallet:     authManager.GetAddress(),
			ThrottleSeconds: config.Config.SlackAlertThrottleSeconds,
		})
		agent.taskCoordinator.SetAlerter(agent.alerter)
		log.Printf("📢 Slack alerting enabled")
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

	// Close cache connection
	if a.agentCache != nil {
		if err := a.agentCache.Close(); err != nil {
			log.Printf("⚠️ Error closing cache connection: %v", err)
		}
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
	// Panic recovery with Slack alerting
	defer func() {
		if r := recover(); r != nil {
			reason := fmt.Sprintf("panic: %v", r)
			log.Printf("💀 Agent crashed: %s", reason)
			if a.alerter != nil {
				a.alerter.SendAgentCrash(reason, 1)
			}
		}
	}()

	if err := a.Start(); err != nil {
		return err
	}

	if a.submitForReviewOnRun {
		// Wait for registration to complete before submitting for review
		select {
		case <-a.protocolHandler.Registered():
			result, err := a.SubmitForReviewDetailed()
			if err != nil {
				log.Printf("⚠️ Failed to submit agent for review: %v", err)
			} else {
				switch result.Status {
				case "submitted":
					log.Printf("✅ Agent submitted for review")
				case "resubmitted_for_review":
					log.Printf("✅ Agent re-submitted for review after metadata changes")
				case "no_changes":
					log.Printf("ℹ️ Agent already in review with no metadata changes")
				case "already_public":
					log.Printf("ℹ️ Agent is already public")
				default:
					if result.Message != "" {
						log.Printf("ℹ️ Submit for review result: %s", result.Message)
					}
				}
			}
		case <-time.After(30 * time.Second):
			log.Printf("⚠️ Timed out waiting for registration — skipping submit for review")
		case <-a.ctx.Done():
		}
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("📡 Received interrupt signal")

	return a.Stop()
}

// SubmitForReview submits the agent for public visibility review on the Teneo network.
// The agent must have been deployed, connected at least once, and be currently online.
// Review can take up to 72 hours. The agent must stay online during review.
func (a *EnhancedAgent) SubmitForReviewDetailed() (*SubmitForReviewResult, error) {
	if a.agentID == "" {
		return nil, fmt.Errorf("agent ID is required for submit-for-review: set AgentID on EnhancedAgentConfig or Config.AgentID")
	}
	tokenID, err := a.getTokenID()
	if err != nil {
		return nil, err
	}
	return SubmitForReviewDetailed(a.backendURL, a.agentID, a.authManager.GetAddress(), tokenID, a.additionalHeaders)
}

// SubmitForReview submits the agent for public visibility review on the Teneo network.
// The agent must have been deployed, connected at least once, and be currently online.
// Review can take up to 72 hours. The agent must stay online during review.
func (a *EnhancedAgent) SubmitForReview() error {
	_, err := a.SubmitForReviewDetailed()
	return err
}

// WithdrawPublic withdraws a public agent back to private visibility.
// Only works on agents that are currently public.
func (a *EnhancedAgent) WithdrawPublic() error {
	if a.agentID == "" {
		return fmt.Errorf("agent ID is required for withdraw-public: set AgentID on EnhancedAgentConfig or Config.AgentID")
	}
	tokenID, err := a.getTokenID()
	if err != nil {
		return err
	}
	return WithdrawPublic(a.backendURL, a.agentID, a.authManager.GetAddress(), tokenID)
}

func (a *EnhancedAgent) getTokenID() (uint64, error) {
	if a.config.NFTTokenID == "" {
		return 0, fmt.Errorf("NFT token ID not set — agent must be deployed first")
	}
	var tokenID uint64
	if _, err := fmt.Sscanf(a.config.NFTTokenID, "%d", &tokenID); err != nil {
		return 0, fmt.Errorf("invalid NFT token ID %q: %w", a.config.NFTTokenID, err)
	}
	return tokenID, nil
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

// buildCapabilitiesJSON converts config capabilities to JSON.
// Uses CapabilityDetails (with descriptions) if available, otherwise
// falls back to Capabilities string slice for backward compatibility.
func buildCapabilitiesJSON(config *Config) ([]byte, error) {
	return json.Marshal(config.ResolveCapabilities())
}
