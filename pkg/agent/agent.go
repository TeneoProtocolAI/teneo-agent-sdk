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

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/auth"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/nft"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/version"
)

// Agent represents a Teneo agent instance
type Agent struct {
	config      *Config
	handler     types.AgentHandler
	nftManager  *nft.BusinessCardManager
	authManager *auth.Manager
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
	mu          sync.RWMutex
}

// NewAgent creates a new Teneo agent instance
func NewAgent(config *Config, handler types.AgentHandler) (*Agent, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		config:  config,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Initialize NFT manager if contract address is provided
	if config.NFTContractAddress != "" {
		nftManager, err := nft.NewBusinessCardManager(
			config.EthereumRPC,
			config.NFTContractAddress,
			config.PrivateKey,
		)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create NFT manager: %w", err)
		}
		agent.nftManager = nftManager
	}

	// Initialize auth manager
	authManager, err := auth.NewManager(config.PrivateKey)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}
	agent.authManager = authManager

	return agent, nil
}

// Run starts the agent and blocks until it's stopped
func (a *Agent) Run(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent is already running")
	}
	a.running = true
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
	}()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run initialization
	if err := a.initialize(); err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	log.Printf("🚀 Teneo Agent '%s' v%s starting up...", a.config.Name, a.config.Version)
	log.Printf("📋 Capabilities: %v", a.config.Capabilities)
	log.Printf("🔗 Owner: %s", a.config.OwnerAddress)

	// Start the agent main loop
	a.wg.Add(1)
	go a.mainLoop()

	// Wait for shutdown signal or context cancellation
	select {
	case <-sigChan:
		log.Println("🛑 Shutdown signal received, stopping agent...")
	case <-ctx.Done():
		log.Println("🛑 Context cancelled, stopping agent...")
	case <-a.ctx.Done():
		log.Println("🛑 Agent context cancelled, stopping agent...")
	}

	// Graceful shutdown
	a.cancel()
	a.wg.Wait()

	log.Println("✅ Agent stopped successfully")
	return nil
}

// initialize sets up the agent
func (a *Agent) initialize() error {
	// Call handler initialization if available
	if initializer, ok := a.handler.(types.AgentInitializer); ok {
		if err := initializer.Initialize(a.ctx, a.config); err != nil {
			return fmt.Errorf("handler initialization failed: %w", err)
		}
	}

	// Register agent with Teneo network if NFT manager is available
	if a.nftManager != nil {
		if err := a.registerWithNetwork(); err != nil {
			log.Printf("⚠️  Failed to register with network: %v", err)
			// Don't fail completely, agent can still work locally
		}
	}

	return nil
}

// registerWithNetwork registers the agent with the Teneo network
func (a *Agent) registerWithNetwork() error {
	log.Println("🔍 Checking agent registration...")

	// Check if agent already has an NFT business card
	businessCard, err := a.nftManager.GetAgentByOwner(a.ctx, a.config.OwnerAddress)
	if err != nil {
		log.Println("📄 No existing business card found, creating new one...")

		// Create mint request
		mintRequest := &types.MintRequest{
			Name:           a.config.Name,
			Description:    a.config.Description,
			Capabilities:   a.config.ResolveCapabilities(),
			ContactInfo:    a.config.ContactInfo,
			PricingModel:   a.config.PricingModel,
			InterfaceType:  a.config.InterfaceType,
			ResponseFormat: a.config.ResponseFormat,
			Version:        a.config.Version,
			SDKVersion:     version.Version(),
		}

		// Mint new business card
		businessCard, err = a.nftManager.MintAgentCard(a.ctx, mintRequest)
		if err != nil {
			return fmt.Errorf("failed to mint business card: %w", err)
		}

		log.Printf("✅ Agent registered with token ID: %s", businessCard.TokenID.String())
	} else {
		log.Printf("✅ Agent already registered with token ID: %s", businessCard.TokenID.String())
	}

	return nil
}

// mainLoop runs the main agent processing loop
func (a *Agent) mainLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Duration(a.config.TaskCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.processAvailableTasks()
		}
	}
}

// processAvailableTasks checks for and processes available tasks
func (a *Agent) processAvailableTasks() {
	// This would typically connect to the Teneo network to get tasks
	// For now, we'll simulate task processing
	if taskProvider, ok := a.handler.(types.TaskProvider); ok {
		tasks, err := taskProvider.GetAvailableTasks(a.ctx)
		if err != nil {
			log.Printf("❌ Failed to get available tasks: %v", err)
			return
		}

		for _, task := range tasks {
			select {
			case <-a.ctx.Done():
				return
			default:
				a.processTask(task)
			}
		}
	}
}

// processTask processes a single task
func (a *Agent) processTask(task types.Task) {
	log.Printf("🔄 Processing task: %s", task.ID)

	// Create task context with timeout
	taskCtx, cancel := context.WithTimeout(a.ctx, time.Duration(a.config.TaskTimeout)*time.Second)
	defer cancel()

	// Process the task
	result, err := a.handler.ProcessTask(taskCtx, task.Content)
	if err != nil {
		log.Printf("❌ Task %s failed: %v", task.ID, err)
		return
	}

	log.Printf("✅ Task %s completed successfully", task.ID)

	// Handle task result if handler supports it
	if resultHandler, ok := a.handler.(types.TaskResultHandler); ok {
		if err := resultHandler.HandleTaskResult(taskCtx, task.ID, result); err != nil {
			log.Printf("⚠️  Failed to handle task result: %v", err)
		}
	}
}

// GetBusinessCard returns the agent's business card information
func (a *Agent) GetBusinessCard() (*types.BusinessCard, error) {
	if a.nftManager == nil {
		return nil, fmt.Errorf("NFT manager not initialized")
	}

	return a.nftManager.GetAgentByOwner(a.ctx, a.config.OwnerAddress)
}

// UpdateMetadata updates the agent's metadata on the blockchain
func (a *Agent) UpdateMetadata(description, contactInfo, pricingModel, version string) error {
	if a.nftManager == nil {
		return fmt.Errorf("NFT manager not initialized")
	}

	return a.nftManager.UpdateAgentMetadata(a.ctx, description, contactInfo, pricingModel, version)
}

// SetActive sets the agent's active status
func (a *Agent) SetActive(active bool) error {
	if a.nftManager == nil {
		return fmt.Errorf("NFT manager not initialized")
	}

	return a.nftManager.SetAgentActive(a.ctx, active)
}

// GetCapabilities returns the agent's capabilities
func (a *Agent) GetCapabilities() []string {
	return a.config.Capabilities
}

// GetAddress returns the agent's Ethereum address
func (a *Agent) GetAddress() string {
	return a.config.OwnerAddress
}

// GetAuthToken generates an authentication token for the agent
func (a *Agent) GetAuthToken() (string, error) {
	return a.authManager.GenerateToken(a.config.OwnerAddress)
}

// Shutdown gracefully shuts down the agent
func (a *Agent) Shutdown() {
	a.cancel()
}

// Close releases resources used by the agent
func (a *Agent) Close() error {
	a.Shutdown()

	if a.nftManager != nil {
		a.nftManager.Close()
	}

	return nil
}

// IsRunning returns whether the agent is currently running
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}
