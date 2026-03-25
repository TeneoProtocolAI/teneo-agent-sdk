package deploy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DeployConfig contains all configuration for deploying an agent
type DeployConfig struct {
	// Backend Configuration
	BackendURL  string // Backend URL (default: from BACKEND_URL env or http://localhost:8080)
	RPCEndpoint string // Ethereum RPC endpoint (default: from RPC_ENDPOINT env)

	// Wallet Configuration
	PrivateKey string // Private key (hex, with or without 0x prefix)

	// Agent Configuration
	AgentID         string          // Unique agent identifier (lowercase, hyphens allowed)
	AgentName       string          // Display name for the agent
	Description     string          // Agent description
	Image           string          // Image URL or base64 data
	AgentType       string          // "command", "nlp", or "mcp"
	Capabilities    json.RawMessage // Agent capabilities array
	Commands        json.RawMessage // Agent commands (optional)
	NlpFallback     bool            // Enable NLP fallback
	Categories      json.RawMessage // Agent categories (optional)
	MetadataVersion string          // Metadata version (e.g. "2.4.0")

	// Profile metadata (optional, not included in config hash)
	ShortDescription string          // Brief one-line summary
	TutorialURL      string          // YouTube/video tutorial URL
	FAQItems         json.RawMessage // FAQ entries as [{question, answer}]

	// State Management
	StateFilePath string // Path to state file (default: .teneo-deploy-state.json)

	// Advanced Options
	MintPrice *big.Int // Custom mint price (default: 2 PEAQ)
}

// DeployResult contains the result of a successful deployment
type DeployResult struct {
	TokenID         uint64 `json:"token_id"`
	TxHash          string `json:"tx_hash"`
	ContractAddress string `json:"contract_address"`
	MetadataURI     string `json:"metadata_uri"`
	AgentID         string `json:"agent_id"`
	AlreadyMinted   bool   `json:"already_minted"`
}

// Deployer handles the full agent deployment flow
type Deployer struct {
	config        *DeployConfig
	httpClient    *HTTPClient
	chainClient   *ChainClient
	authenticator *Authenticator
	stateManager  *StateManager
	configHash    string
}

// NewDeployer creates a new deployer instance
func NewDeployer(config *DeployConfig) (*Deployer, error) {
	// Apply defaults
	if config.BackendURL == "" {
		if backendURL := os.Getenv("BACKEND_URL"); backendURL != "" {
			config.BackendURL = backendURL
		} else {
			config.BackendURL = "https://backend.developer.chatroom.teneo-protocol.ai"
		}
	}

	if config.RPCEndpoint == "" {
		if rpcEndpoint := os.Getenv("RPC_ENDPOINT"); rpcEndpoint != "" {
			config.RPCEndpoint = rpcEndpoint
		} else {
			config.RPCEndpoint = "https://peaq.api.onfinality.io/public"
		}
	}

	if config.PrivateKey == "" {
		if privateKey := os.Getenv("PRIVATE_KEY"); privateKey != "" {
			config.PrivateKey = privateKey
		} else {
			return nil, fmt.Errorf("private key is required")
		}
	}

	if config.StateFilePath == "" {
		config.StateFilePath = ".teneo-deploy-state.json"
	}

	if config.MetadataVersion == "" {
		config.MetadataVersion = "2.4.0"
	}

	if len(config.Commands) > 0 {
		normalizedCommands, err := normalizeCommandsJSON(config.Commands)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize commands: %w", err)
		}
		config.Commands = normalizedCommands
	}

	// Create HTTP client
	httpClient := NewHTTPClient(config.BackendURL)

	// Create authenticator
	authenticator, err := NewAuthenticator(config.PrivateKey, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Create state manager
	stateManager := NewStateManager(config.StateFilePath)

	// Compute config hash matching GenerateConfigHash logic
	configHash := computeConfigHash(config)

	return &Deployer{
		config:        config,
		httpClient:    httpClient,
		authenticator: authenticator,
		stateManager:  stateManager,
		configHash:    configHash,
	}, nil
}

// Deploy executes the full deployment flow with resilience and idempotency
func (d *Deployer) Deploy(ctx context.Context) (*DeployResult, error) {
	log.Println("🚀 Starting agent deployment...")

	// Load existing state
	state, err := d.stateManager.Load()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to load state file: %v", err)
	}

	// Create chain client for on-chain checks
	// We need contract info first, so we'll create it later if needed
	var chainClient *ChainClient

	// Check if we need to recover from partial deployment
	if state != nil {
		log.Printf("📋 Found existing state: status=%s, agentID=%s", state.Status, state.AgentID)

		// Verify agent ID matches
		if state.AgentID != d.config.AgentID {
			log.Printf("⚠️ State file is for different agent (%s vs %s), starting fresh", state.AgentID, d.config.AgentID)
			state = nil
		}
	}

	// Handle recovery scenarios
	if state != nil && state.ContractAddress != "" {
		chainClient, err = NewChainClient(d.config.RPCEndpoint, state.ContractAddress, state.ChainID, d.config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create chain client: %w", err)
		}
		defer chainClient.Close()

		// Check on-chain status
		hasAccess, err := chainClient.HasAccess(ctx)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to check on-chain access: %v", err)
		}

		if hasAccess {
			switch state.Status {
			case StatusConfirmed:
				// Fully complete
				log.Println("✅ Agent already deployed and confirmed")
				return &DeployResult{
					TokenID:         state.TokenID,
					TxHash:          state.TxHash,
					ContractAddress: state.ContractAddress,
					AgentID:         state.AgentID,
					AlreadyMinted:   true,
				}, nil

			case StatusMinted:
				// Minted but not confirmed - just need to confirm
				log.Println("📋 Agent minted but not confirmed, completing confirmation...")
				return d.confirmOnly(ctx, state)

			default:
				// Has access but state is pending - recover token ID and confirm
				log.Println("🔍 Agent has on-chain access, recovering token ID...")
				tokenID, err := chainClient.GetTokenID(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to recover token ID: %w", err)
				}
				state.TokenID = tokenID
				state.Status = StatusMinted
				d.stateManager.Save(state)
				return d.confirmOnly(ctx, state)
			}
		}

		// No on-chain access - need to re-deploy
		if state.Status == StatusPending {
			log.Println("📋 Pending deployment found, retrying...")
		}
	}

	// Fresh deployment or retry
	return d.fullDeploy(ctx)
}

// fullDeploy executes a complete deployment from scratch
func (d *Deployer) fullDeploy(ctx context.Context) (*DeployResult, error) {
	// Validate configuration
	if err := d.validateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Step 1: Authenticate
	log.Println("[Step 1/5] 🔐 Authenticating with backend...")
	sessionToken, _, err := d.authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	log.Println("   ✅ Authentication successful")

	// Step 2: Call deploy endpoint
	log.Println("[Step 2/5] 📤 Preparing deployment (uploading metadata, getting signature)...")
	deployResp, err := d.callDeploy(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("deploy preparation failed: %w", err)
	}
	if len(deployResp.ConfigHash) >= 16 {
		log.Printf("   ✅ Metadata stored, config hash: %s", deployResp.ConfigHash[:16]+"...")
	} else {
		log.Printf("   ✅ Metadata stored, config hash: %s", deployResp.ConfigHash)
	}
	log.Printf("   ✅ Contract: %s (Chain ID: %s)", deployResp.ContractAddress, deployResp.ChainID)

	// If server performed gasless minting, everything is done — skip steps 3-4
	if deployResp.Gasless && deployResp.TokenID <= 0 {
		return nil, fmt.Errorf("server returned gasless=true but invalid token_id=%d — this is a server error, not retrying to prevent double-mint", deployResp.TokenID)
	}
	if deployResp.Gasless {
		log.Printf("   ✅ Gasless mint! Token ID: %d, Tx: %s", deployResp.TokenID, deployResp.TxHash)

		state := &DeployState{
			AgentID:         d.config.AgentID,
			AgentName:       d.config.AgentName,
			WalletAddress:   d.authenticator.GetAddress(),
			Status:          StatusConfirmed,
			ContractAddress: deployResp.ContractAddress,
			ConfigHash:      deployResp.ConfigHash,
			ChainID:         deployResp.ChainID,
			TokenID:         uint64(deployResp.TokenID),
			TxHash:          deployResp.TxHash,
			CreatedAt:       time.Now().UTC(),
		}
		if err := d.stateManager.Save(state); err != nil {
			log.Printf("⚠️ Warning: Failed to save state: %v", err)
		}

		log.Println("[Step 5/5] ✅ Deployment complete (gasless)!")

		return &DeployResult{
			TokenID:         uint64(deployResp.TokenID),
			TxHash:          deployResp.TxHash,
			ContractAddress: deployResp.ContractAddress,
			MetadataURI:     deployResp.MetadataURI,
			AgentID:         d.config.AgentID,
			AlreadyMinted:   false,
		}, nil
	}

	// Server must perform gasless minting — client-side minting is not supported
	return nil, fmt.Errorf("server did not perform gasless minting (gasless=false in response) — ensure your backend supports gasless minting")
}

// confirmOnly handles the case where we need to confirm an already-minted NFT
func (d *Deployer) confirmOnly(ctx context.Context, state *DeployState) (*DeployResult, error) {
	// Check if session is still valid
	sessionToken := state.SessionToken
	if !state.IsSessionValid() {
		log.Println("🔐 Re-authenticating...")
		var err error
		sessionToken, state.SessionExpiry, err = d.authenticate(ctx)
		if err != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", err)
		}
		state.SessionToken = sessionToken
		d.stateManager.Save(state)
	}

	log.Println("[Confirm] 💾 Confirming with backend...")
	confirmResp, err := d.confirmMint(ctx, sessionToken, state)
	if err != nil {
		if errors.Is(err, ErrSessionExpired) {
			log.Println("   ⚠️ Session expired, re-authenticating...")
			sessionToken, state.SessionExpiry, err = d.authenticate(ctx)
			if err != nil {
				return nil, fmt.Errorf("re-authentication failed: %w", err)
			}
			state.SessionToken = sessionToken
			d.stateManager.Save(state)

			confirmResp, err = d.confirmMint(ctx, sessionToken, state)
			if err != nil {
				return nil, fmt.Errorf("confirm-mint failed: %w", err)
			}
		} else {
			return nil, fmt.Errorf("confirm-mint failed: %w", err)
		}
	}

	state.Status = StatusConfirmed
	d.stateManager.Save(state)

	log.Println("✅ Agent confirmed successfully!")

	return &DeployResult{
		TokenID:         state.TokenID,
		TxHash:          state.TxHash,
		ContractAddress: state.ContractAddress,
		MetadataURI:     confirmResp.MetadataURI,
		AgentID:         state.AgentID,
		AlreadyMinted:   true,
	}, nil
}

// authenticate performs the challenge-response authentication
func (d *Deployer) authenticate(ctx context.Context) (string, int64, error) {
	return d.authenticator.Authenticate()
}

// callDeploy calls the deploy endpoint
func (d *Deployer) callDeploy(ctx context.Context, sessionToken string) (*DeployResponse, error) {
	req := &DeployRequest{
		WalletAddress:    d.authenticator.GetAddress(),
		AgentID:          d.config.AgentID,
		AgentName:        d.config.AgentName,
		Description:      d.config.Description,
		Image:            d.config.Image,
		AgentType:        d.config.AgentType,
		Capabilities:     d.config.Capabilities,
		Commands:         d.config.Commands,
		NlpFallback:      d.config.NlpFallback,
		Categories:       d.config.Categories,
		ShortDescription: d.config.ShortDescription,
		TutorialURL:      d.config.TutorialURL,
		FAQItems:         d.config.FAQItems,
		ConfigHash:       d.configHash,
		MetadataVersion:  d.config.MetadataVersion,
	}

	return d.httpClient.Deploy(sessionToken, req)
}

// confirmMint calls the confirm-mint endpoint.
// Metadata is retrieved from pending_metadata stored at deploy time — we only
// send identifiers and the tx proof.
func (d *Deployer) confirmMint(ctx context.Context, sessionToken string, state *DeployState) (*ConfirmMintResponse, error) {
	// Validate token ID fits in int64 before conversion
	if state.TokenID > math.MaxInt64 {
		return nil, fmt.Errorf("token ID %d exceeds int64 maximum", state.TokenID)
	}

	req := &ConfirmMintRequest{
		AgentID:          state.AgentID,
		AgentName:        d.config.AgentName,
		WalletAddress:    state.WalletAddress,
		TokenID:          int64(state.TokenID),
		TxHash:           state.TxHash,
		ConfigHash:       d.configHash,
		Description:      d.config.Description,
		Image:            d.config.Image,
		AgentType:        d.config.AgentType,
		Capabilities:     d.config.Capabilities,
		Commands:         d.config.Commands,
		NlpFallback:      d.config.NlpFallback,
		Categories:       d.config.Categories,
		ShortDescription: d.config.ShortDescription,
		TutorialURL:      d.config.TutorialURL,
		FAQItems:         d.config.FAQItems,
		MetadataVersion:  d.config.MetadataVersion,
	}

	return d.httpClient.ConfirmMint(sessionToken, req)
}

// validateConfig validates the deployment configuration
func (d *Deployer) validateConfig() error {
	if d.config.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}
	if d.config.AgentName == "" {
		return fmt.Errorf("agent_name is required")
	}
	if d.config.Description == "" {
		return fmt.Errorf("description is required")
	}
	if d.config.AgentType == "" {
		return fmt.Errorf("agent_type is required")
	}

	// Validate agent ID format (lowercase, hyphens, numbers only)
	for _, c := range d.config.AgentID {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return fmt.Errorf("agent_id can only contain lowercase letters, numbers, and hyphens")
		}
	}

	// Validate agent type
	validTypes := map[string]bool{"command": true, "nlp": true, "mcp": true, "commandless": true}
	if !validTypes[d.config.AgentType] {
		return fmt.Errorf("agent_type must be 'command', 'nlp', 'mcp', or 'commandless'")
	}

	// RPCEndpoint is only needed for client-side minting (non-gasless).
	// If the server supports gasless minting, the SDK never connects to RPC.
	// Validation is deferred to runtime if client-side minting is needed.

	return nil
}

// DeployAgent is a convenience function for simple deployments
func DeployAgent(cfg DeployConfig) (*DeployResult, error) {
	deployer, err := NewDeployer(&cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return deployer.Deploy(ctx)
}

// generateAgentID generates a valid agent ID from a name
func generateAgentID(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")

	// Remove invalid characters
	var result strings.Builder
	for _, c := range id {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}

	// Remove leading/trailing hyphens and collapse multiple hyphens
	id = result.String()
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}
	id = strings.Trim(id, "-")

	return id
}

// computeConfigHash computes the v4 config hash from DeployConfig, matching the
// canonical format used by GenerateConfigHash in mint.go and the backend.
// v4 includes ALL command fields and capability descriptions.
func computeConfigHash(config *DeployConfig) string {
	// Parse capabilities from JSON (name + description)
	type capObj struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	var caps []capObj
	if len(config.Capabilities) > 0 {
		json.Unmarshal(config.Capabilities, &caps)
	}
	sort.Slice(caps, func(i, j int) bool {
		return caps[i].Name < caps[j].Name
	})
	capParts := make([]string, len(caps))
	for i, c := range caps {
		capParts[i] = c.Name + "~" + c.Description
	}

	// Parse categories from JSON and sort
	var categories []string
	if len(config.Categories) > 0 {
		json.Unmarshal(config.Categories, &categories)
	}
	sort.Strings(categories)

	parts := []string{
		"v4",
		config.AgentID,
		config.AgentName,
		config.Description,
		config.AgentType,
		strings.Join(capParts, ","),
		strconv.FormatBool(config.NlpFallback),
		strings.Join(categories, ","),
	}

	// Parse and include full command data (sorted by trigger)
	type paramObj struct {
		Name           string `json:"name"`
		Type           string `json:"type"`
		Required       bool   `json:"required"`
		Description    string `json:"description"`
		MinValue       string `json:"minValue"`
		IsBillingCount bool   `json:"isBillingCount"`
	}
	type cmdObj struct {
		Trigger      string          `json:"trigger"`
		Argument     string          `json:"argument"`
		Description  string          `json:"description"`
		Parameters   json.RawMessage `json:"parameters"`
		StrictArg    *bool           `json:"strictArg"`
		MinArgs      *int            `json:"minArgs"`
		MaxArgs      *int            `json:"maxArgs"`
		PricePerUnit float64         `json:"pricePerUnit"`
		PriceType    string          `json:"priceType"`
		TaskUnit     string          `json:"taskUnit"`
		TimeUnit     string          `json:"timeUnit"`
	}
	var cmds []cmdObj
	commandsJSON := config.Commands
	if len(commandsJSON) > 0 {
		if normalizedCommands, err := normalizeCommandsJSON(commandsJSON); err == nil {
			commandsJSON = normalizedCommands
		}
		json.Unmarshal(commandsJSON, &cmds)
	}
	if len(cmds) > 0 {
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Trigger < cmds[j].Trigger
		})
		cmdParts := make([]string, len(cmds))
		for i, cmd := range cmds {
			// Parse parameters for this command
			var params []paramObj
			if len(cmd.Parameters) > 0 {
				json.Unmarshal(cmd.Parameters, &params)
			}
			sort.Slice(params, func(a, b int) bool {
				return params[a].Name < params[b].Name
			})
			paramParts := make([]string, len(params))
			for j, p := range params {
				paramParts[j] = p.Name + "~" + p.Type + "~" + strconv.FormatBool(p.Required) + "~" + p.Description + "~" + p.MinValue + "~" + strconv.FormatBool(p.IsBillingCount)
			}

			strict := "false"
			if cmd.StrictArg != nil && *cmd.StrictArg {
				strict = "true"
			}
			minArgs := "0"
			if cmd.MinArgs != nil {
				minArgs = strconv.Itoa(*cmd.MinArgs)
			}
			maxArgs := "0"
			if cmd.MaxArgs != nil {
				maxArgs = strconv.Itoa(*cmd.MaxArgs)
			}

			cmdParts[i] = cmd.Trigger + ":" +
				cmd.Argument + ":" +
				cmd.Description + ":" +
				strconv.FormatFloat(cmd.PricePerUnit, 'f', -1, 64) + ":" +
				cmd.PriceType + ":" +
				cmd.TaskUnit + ":" +
				cmd.TimeUnit + ":" +
				strict + ":" +
				minArgs + ":" +
				maxArgs + ":" +
				strings.Join(paramParts, ";")
		}
		parts = append(parts, strings.Join(cmdParts, ","))
	}

	data := strings.Join(parts, "|")
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
