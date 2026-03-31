package deploy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// htmlTagPattern matches HTML/script tags for XSS prevention
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// DefaultMaxJSONSize is the fallback max size for agent JSON files (24KB)
// The actual limit is fetched from backend via schema endpoint
const DefaultMaxJSONSize = 24 * 1024

// AgentConfig represents the agent configuration from JSON file
type AgentConfig struct {
	Name            string       `json:"name"`
	AgentID         string       `json:"agent_id"`
	Description     string       `json:"description"`
	Image           string       `json:"image,omitempty"`
	AgentType       string       `json:"agent_type"`
	Categories      []string     `json:"categories"`
	Capabilities    []Capability `json:"capabilities"`
	Commands        []Command    `json:"commands,omitempty"`
	NlpFallback     bool         `json:"nlp_fallback"`
	McpManifest     string       `json:"mcpManifest,omitempty"`
	MetadataVersion string       `json:"metadata_version,omitempty"`
}

// Capability represents an agent capability
type Capability struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CommandParameter represents a parameter for an agent command
type CommandParameter struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Required       bool   `json:"required"`
	Description    string `json:"description,omitempty"`
	MinValue       string `json:"minValue,omitempty"`
	IsBillingCount bool   `json:"isBillingCount,omitempty"`
}

// Command represents an agent command
type Command struct {
	Trigger      string             `json:"trigger"`
	Argument     string             `json:"argument,omitempty"`
	Description  string             `json:"description,omitempty"`
	Parameters   []CommandParameter `json:"parameters,omitempty"`
	StrictArg    *bool              `json:"strictArg,omitempty"`
	MinArgs      *int               `json:"minArgs,omitempty"`
	MaxArgs      *int               `json:"maxArgs,omitempty"`
	PricePerUnit float64            `json:"pricePerUnit,omitempty"`
	PriceType    string             `json:"priceType,omitempty"`
	TaskUnit     string             `json:"taskUnit,omitempty"`
	TimeUnit     string             `json:"timeUnit,omitempty"`
}

// MintResult is defined in chain.go with fields:
// TokenID, TxHash, AgentID, Status, ContractAddress, Message

// Minter handles the gasless minting flow
type Minter struct {
	config       *MintConfig
	httpClient   *HTTPClient
	walClient    *WALClient
	schemaCache  *SchemaCache
}

// MintConfig contains configuration for minting
type MintConfig struct {
	PrivateKey  string // Wallet private key (hex)
	BackendURL  string // Backend API URL
	RPCEndpoint string // Blockchain RPC endpoint
}

// NewMinter creates a new minter instance
func NewMinter(config *MintConfig) (*Minter, error) {
	// Apply defaults from environment
	if config.BackendURL == "" {
		config.BackendURL = os.Getenv("BACKEND_URL")
		if config.BackendURL == "" {
			config.BackendURL = "http://localhost:8080"
		}
	}

	if config.RPCEndpoint == "" {
		config.RPCEndpoint = os.Getenv("RPC_ENDPOINT")
	}

	if config.PrivateKey == "" {
		config.PrivateKey = os.Getenv("PRIVATE_KEY")
		if config.PrivateKey == "" {
			return nil, fmt.Errorf("private key is required")
		}
	}

	httpClient := NewHTTPClient(config.BackendURL)

	return &Minter{
		config:     config,
		httpClient: httpClient,
		walClient:  NewWALClient(),
	}, nil
}

// Mint loads an agent config from JSON file and mints/syncs the agent
func (m *Minter) Mint(jsonPath string) (*MintResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return m.MintWithContext(ctx, jsonPath)
}

// MintWithContext loads an agent config from JSON file and mints/syncs with context
func (m *Minter) MintWithContext(ctx context.Context, jsonPath string) (*MintResult, error) {
	log.Printf("📦 Loading agent config from: %s", jsonPath)

	// Step 1: Check file size (fast fail against default limit)
	fileInfo, err := os.Stat(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fileInfo.Size()
	if fileSize > DefaultMaxJSONSize {
		return nil, fmt.Errorf("JSON file too large (max %d bytes, got %d)", DefaultMaxJSONSize, fileSize)
	}

	// Step 2: Read file
	file, err := os.Open(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, DefaultMaxJSONSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Step 3: Parse JSON
	var config AgentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Step 4: Pre-validation (O(1) cheap checks)
	if err := m.preValidate(&config); err != nil {
		return nil, fmt.Errorf("pre-validation failed: %w", err)
	}

	// Step 5: Fetch and verify schema (with caching)
	schema, err := m.getSchema(ctx)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to fetch schema: %v (proceeding with local validation)", err)
	} else {
		log.Printf("📋 Schema version: %s, max JSON size: %d bytes", schema.SchemaVersion, schema.MaxJSONSize)

		// Validate file size against backend limit
		if schema.MaxJSONSize > 0 && int(fileSize) > schema.MaxJSONSize {
			return nil, fmt.Errorf("JSON file too large (backend limit: %d bytes, got %d)", schema.MaxJSONSize, fileSize)
		}
	}

	// Step 6: Full validation against schema
	if err := m.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	log.Printf("✅ Agent config validated: %s (%s)", config.Name, config.AgentID)

	// Step 7: Check WAL for pending operations
	wal, err := m.walClient.Load(config.AgentID)
	if err == nil && wal != nil && wal.PendingTxHash != "" {
		log.Printf("🔍 Found pending transaction in WAL: %s", wal.PendingTxHash)
		return m.recoverFromWAL(ctx, wal, &config)
	}

	// Step 8: Generate config hash
	configHash := GenerateConfigHash(&config)
	if len(configHash) >= 16 {
		log.Printf("🔐 Config hash: %s", configHash[:16]+"...")
	} else {
		log.Printf("🔐 Config hash: %s", configHash)
	}

	// Step 9: Proceed to sync
	schemaVersion := ""
	if schema != nil {
		schemaVersion = schema.SchemaVersion
	}

	return m.syncAndMint(ctx, &config, configHash, schemaVersion)
}

// preValidate performs cheap O(1) checks before full validation
func (m *Minter) preValidate(config *AgentConfig) error {
	// Check required top-level fields
	if config.Name == "" {
		return fmt.Errorf("name is required")
	}
	if config.AgentID == "" {
		return fmt.Errorf("agentId is required")
	}
	if config.AgentType == "" {
		return fmt.Errorf("agentType is required")
	}

	// Check agentId format
	for _, c := range config.AgentID {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return fmt.Errorf("agentId can only contain lowercase letters, numbers, and hyphens")
		}
	}

	return nil
}

// validateConfig performs full validation against schema
func (m *Minter) validateConfig(config *AgentConfig) error {
	// Name validation
	if len(config.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}
	if len(config.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}

	// XSS prevention: reject HTML tags in name and description
	if htmlTagPattern.MatchString(config.Name) {
		return fmt.Errorf("name must not contain HTML tags")
	}

	// AgentID validation
	if len(config.AgentID) > 64 {
		return fmt.Errorf("agentId must not exceed 64 characters")
	}

	// Description validation
	if len(config.Description) < 10 {
		return fmt.Errorf("description must be at least 10 characters")
	}
	if len(config.Description) > 2000 {
		return fmt.Errorf("description must not exceed 2000 characters")
	}
	if htmlTagPattern.MatchString(config.Description) {
		return fmt.Errorf("description must not contain HTML tags")
	}

	// AgentType validation
	validTypes := map[string]bool{"command": true, "nlp": true, "mcp": true, "commandless": true}
	if !validTypes[config.AgentType] {
		return fmt.Errorf("agentType must be 'command', 'nlp', 'mcp', or 'commandless'")
	}

	// Categories validation
	if len(config.Categories) < 1 {
		return fmt.Errorf("at least 1 category is required")
	}
	if len(config.Categories) > 2 {
		return fmt.Errorf("maximum 2 categories allowed")
	}

	// Capabilities validation
	if len(config.Capabilities) < 1 {
		return fmt.Errorf("at least 1 capability is required")
	}
	if len(config.Capabilities) > 50 {
		return fmt.Errorf("maximum 50 capabilities allowed")
	}

	for i, cap := range config.Capabilities {
		if cap.Name == "" {
			return fmt.Errorf("capability %d: name is required", i+1)
		}
		if len(cap.Name) > 100 {
			return fmt.Errorf("capability %d: name must not exceed 100 characters", i+1)
		}
		if len(cap.Description) > 500 {
			return fmt.Errorf("capability %d: description must not exceed 500 characters", i+1)
		}
	}

	// Commands validation (optional)
	if len(config.Commands) > 100 {
		return fmt.Errorf("maximum 100 commands allowed")
	}

	for i, cmd := range config.Commands {
		if cmd.Trigger == "" {
			return fmt.Errorf("command %d: trigger is required", i+1)
		}
		if len(cmd.Trigger) > 100 {
			return fmt.Errorf("command %d: trigger must not exceed 100 characters", i+1)
		}
		if len(cmd.Description) > 500 {
			return fmt.Errorf("command %d: description must not exceed 500 characters", i+1)
		}
	}

	// MCP manifest validation
	if config.AgentType == "mcp" && config.McpManifest == "" {
		return fmt.Errorf("mcpManifest is required for mcp agent type")
	}

	return nil
}

// getSchema fetches the validation schema from backend
func (m *Minter) getSchema(ctx context.Context) (*SchemaResponse, error) {
	// Check cache
	if m.schemaCache != nil && time.Since(m.schemaCache.FetchedAt) < time.Hour {
		return m.schemaCache.Schema, nil
	}

	// Fetch from backend
	schema, err := m.httpClient.GetSchema()
	if err != nil {
		// Use stale cache if available
		if m.schemaCache != nil {
			log.Printf("⚠️ Using stale schema cache")
			return m.schemaCache.Schema, nil
		}
		return nil, err
	}

	// Update cache
	m.schemaCache = &SchemaCache{
		Schema:    schema,
		FetchedAt: time.Now(),
	}

	return schema, nil
}

// syncAndMint performs the sync and mint flow
func (m *Minter) syncAndMint(ctx context.Context, config *AgentConfig, configHash, schemaVersion string) (*MintResult, error) {
	// Create authenticator
	authenticator, err := NewAuthenticator(m.config.PrivateKey, m.httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Get challenge
	log.Println("🔐 Getting authentication challenge...")
	challenge, err := m.httpClient.GetChallenge(authenticator.GetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}

	// Sign challenge
	signature, err := authenticator.SignChallenge(challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to sign challenge: %w", err)
	}

	// Call sync endpoint
	log.Println("🔄 Syncing with backend...")
	syncResp, err := m.httpClient.Sync(&SyncRequest{
		Wallet:        authenticator.GetAddress(),
		AgentID:       config.AgentID,
		ConfigHash:    configHash,
		Challenge:     challenge,
		Signature:     signature,
		SchemaVersion: schemaVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}

	log.Printf("📋 Sync status: %s", syncResp.Status)

	switch syncResp.Status {
	case "SYNCED":
		log.Println("✅ Agent already synced!")
		if syncResp.TokenID == nil {
			return nil, fmt.Errorf("backend returned SYNCED status but no token_id")
		}
		return &MintResult{
			TokenID:         uint64(*syncResp.TokenID),
			AgentID:         config.AgentID,
			Status:          MintStatusAlreadyOwned,
			ContractAddress: syncResp.ContractAddress,
			Message:         "Agent synced successfully",
		}, nil

	case "UPDATE_REQUIRED":
		log.Printf("⚠️ Config changed (current: %s, new: %s), auto-updating...", syncResp.CurrentHash, syncResp.NewHash)
		return m.executeUpdate(ctx, config, configHash, syncResp)

	case "MINT_REQUIRED", "RESUME_MINT":
		log.Println("💰 Minting required, proceeding...")
		return m.executeMint(ctx, config, authenticator, configHash)

	default:
		return nil, fmt.Errorf("unexpected sync status: %s", syncResp.Status)
	}
}

// executeMint performs the actual minting operation
func (m *Minter) executeMint(ctx context.Context, config *AgentConfig, authenticator *Authenticator, configHash string) (*MintResult, error) {
	// Authenticate for deploy endpoint
	log.Println("🔐 Authenticating for deploy...")
	sessionToken, _, err := authenticator.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Convert config to deploy request
	capabilitiesJSON, _ := json.Marshal(config.Capabilities)
	commandsJSON, _ := json.Marshal(config.Commands)
	categoriesJSON, _ := json.Marshal(config.Categories)

	deployReq := &DeployRequest{
		WalletAddress:   authenticator.GetAddress(),
		AgentID:         config.AgentID,
		AgentName:       config.Name,
		Description:     config.Description,
		Image:           config.Image,
		AgentType:       config.AgentType,
		Capabilities:    capabilitiesJSON,
		Commands:        commandsJSON,
		NlpFallback:     config.NlpFallback,
		Categories:      categoriesJSON,
		ConfigHash:      configHash,
		MetadataVersion: config.MetadataVersion,
	}

	// Call deploy endpoint
	log.Println("📤 Storing metadata and getting mint signature...")
	deployResp, err := m.httpClient.Deploy(sessionToken, deployReq)
	if err != nil {
		return nil, fmt.Errorf("deploy failed: %w", err)
	}

	if len(deployResp.ConfigHash) >= 16 {
		log.Printf("✅ Deploy prepared, config hash: %s", deployResp.ConfigHash[:16]+"...")
	} else {
		log.Printf("✅ Deploy prepared, config hash: %s", deployResp.ConfigHash)
	}

	// If server performed gasless minting, everything is done
	if deployResp.Gasless && deployResp.TokenID <= 0 {
		return nil, fmt.Errorf("server returned gasless=true but invalid token_id=%d — this is a server error, not retrying to prevent double-mint", deployResp.TokenID)
	}
	if deployResp.Gasless {
		log.Printf("✅ Gasless mint! Token ID: %d, Tx: %s", deployResp.TokenID, deployResp.TxHash)
		m.walClient.Delete(config.AgentID)
		return &MintResult{
			TokenID:         uint64(deployResp.TokenID),
			AgentID:         config.AgentID,
			Status:          MintStatusMinted,
			ContractAddress: deployResp.ContractAddress,
			TxHash:          deployResp.TxHash,
			Message:         "Agent minted successfully (gasless)",
		}, nil
	}

	// Server must perform gasless minting — client-side minting is not supported
	return nil, fmt.Errorf("server did not perform gasless minting (gasless=false in response) — ensure your backend supports gasless minting")
}

// executeUpdate handles automatic metadata re-upload when config changes
func (m *Minter) executeUpdate(ctx context.Context, config *AgentConfig, configHash string, syncResp *SyncResponse) (*MintResult, error) {
	// 1. Create authenticator
	authenticator, err := NewAuthenticator(m.config.PrivateKey, m.httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	// 2. Authenticate to get session token
	log.Println("🔐 Authenticating for metadata update...")
	sessionToken, _, err := authenticator.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// 3. Convert config to UpdateMetadataRequest
	capabilitiesJSON, _ := json.Marshal(config.Capabilities)
	commandsJSON, _ := json.Marshal(config.Commands)
	categoriesJSON, _ := json.Marshal(config.Categories)

	updateReq := &UpdateMetadataRequest{
		WalletAddress:   authenticator.GetAddress(),
		AgentID:         config.AgentID,
		AgentName:       config.Name,
		Description:     config.Description,
		Image:           config.Image,
		AgentType:       config.AgentType,
		Capabilities:    capabilitiesJSON,
		Commands:        commandsJSON,
		NlpFallback:     config.NlpFallback,
		Categories:      categoriesJSON,
		ConfigHash:      configHash,
		MetadataVersion: config.MetadataVersion,
	}

	// 4. Call update endpoint
	log.Println("📤 Uploading updated metadata to IPFS and updating on-chain...")
	updateResp, err := m.httpClient.UpdateMetadata(sessionToken, updateReq)
	if err != nil {
		return nil, fmt.Errorf("metadata update failed: %w", err)
	}

	log.Printf("✅ Metadata updated: IPFS=%s, TxHash=%s", updateResp.IpfsHash, updateResp.TxHash)

	// 5. Re-sync to verify SYNCED status
	log.Println("🔄 Verifying update with re-sync...")
	// Get new challenge for re-sync
	challenge, err := m.httpClient.GetChallenge(authenticator.GetAddress())
	if err != nil {
		// Update succeeded, but re-sync failed - still return success
		log.Printf("⚠️ Re-sync challenge failed: %v (update was successful)", err)
		var tokenID uint64
		if syncResp.TokenID != nil {
			tokenID = uint64(*syncResp.TokenID)
		}
		return &MintResult{
			AgentID:         config.AgentID,
			TokenID:         tokenID,
			ContractAddress: syncResp.ContractAddress,
			Status:          MintStatusUpdated,
			TxHash:          updateResp.TxHash,
			Message:         "Agent metadata updated successfully",
		}, nil
	}

	signature, err := authenticator.SignChallenge(challenge)
	if err != nil {
		// Same - update succeeded
		log.Printf("⚠️ Re-sync sign failed: %v (update was successful)", err)
		var tokenID uint64
		if syncResp.TokenID != nil {
			tokenID = uint64(*syncResp.TokenID)
		}
		return &MintResult{
			AgentID:         config.AgentID,
			TokenID:         tokenID,
			ContractAddress: syncResp.ContractAddress,
			Status:          MintStatusUpdated,
			TxHash:          updateResp.TxHash,
			Message:         "Agent metadata updated successfully",
		}, nil
	}

	reSyncResp, err := m.httpClient.Sync(&SyncRequest{
		Wallet:     authenticator.GetAddress(),
		AgentID:    config.AgentID,
		ConfigHash: configHash,
		Challenge:  challenge,
		Signature:  signature,
	})

	if err != nil {
		log.Printf("⚠️ Re-sync failed: %v (update was successful)", err)
	} else {
		log.Printf("✅ Re-sync status: %s", reSyncResp.Status)
	}

	var tokenID uint64
	if syncResp.TokenID != nil {
		tokenID = uint64(*syncResp.TokenID)
	}
	// Prefer re-sync token ID if available
	if reSyncResp != nil && reSyncResp.TokenID != nil {
		tokenID = uint64(*reSyncResp.TokenID)
	}

	return &MintResult{
		AgentID:         config.AgentID,
		TokenID:         tokenID,
		ContractAddress: syncResp.ContractAddress,
		Status:          MintStatusUpdated,
		TxHash:          updateResp.TxHash,
		Message:         "Agent metadata updated successfully",
	}, nil
}

// recoverFromWAL recovers a pending mint operation from WAL
func (m *Minter) recoverFromWAL(ctx context.Context, wal *WALEntry, config *AgentConfig) (*MintResult, error) {
	log.Printf("🔄 Recovering from WAL state: %s", wal.State)

	// Use RPC URL from WAL (saved from deploy response), fallback to config
	rpcEndpoint := wal.RPCURL
	if rpcEndpoint == "" {
		rpcEndpoint = m.config.RPCEndpoint
	}

	// Create chain client
	chainClient, err := NewChainClient(rpcEndpoint, wal.ContractAddress, wal.ChainID, m.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain client: %w", err)
	}
	defer chainClient.Close()

	// Check transaction receipt
	if wal.PendingTxHash != "" {
		log.Printf("🔍 Checking transaction: %s", wal.PendingTxHash)

		receipt, err := chainClient.GetTransactionReceipt(ctx, wal.PendingTxHash)
		if err != nil {
			log.Printf("⚠️ Transaction not found or pending: %v", err)
			// Transaction might be pending or dropped - wait or retry
			return nil, fmt.Errorf("pending transaction status unknown, please check: %s", wal.PendingTxHash)
		}

		if receipt.Status == 1 {
			// Transaction succeeded
			log.Println("✅ Transaction confirmed!")

			tokenID := wal.PendingTokenID
			if tokenID == nil {
				// Extract from receipt logs
				extractedID, err := chainClient.ExtractTokenIDFromReceipt(receipt)
				if err != nil {
					return nil, fmt.Errorf("failed to extract token ID from receipt: %w", err)
				}
				tokenID = &extractedID
			}

			// Confirm with backend (IPFS upload + tokenURI update happens server-side)
			authenticator, err := NewAuthenticator(m.config.PrivateKey, m.httpClient)
			if err != nil {
				return nil, fmt.Errorf("failed to create authenticator: %w", err)
			}

			sessionToken, _, err := authenticator.Authenticate()
			if err != nil {
				log.Printf("⚠️ Warning: Failed to authenticate for confirm: %v", err)
			} else {
				confirmReq := &ConfirmMintRequest{
					AgentID:       config.AgentID,
					WalletAddress: wal.Wallet,
					TokenID:       int64(*tokenID),
					TxHash:        wal.PendingTxHash,
					ConfigHash:    wal.ConfigHash,
				}

				if _, err := m.httpClient.ConfirmMint(sessionToken, confirmReq); err != nil {
					log.Printf("⚠️ Warning: Confirm-mint failed: %v", err)
				}
			}

			// Clean up WAL
			m.walClient.Delete(config.AgentID)

			return &MintResult{
				TokenID:         *tokenID,
				AgentID:         config.AgentID,
				Status:          MintStatusMinted,
				ContractAddress: wal.ContractAddress,
				TxHash:          wal.PendingTxHash,
				Message:         "Recovered from pending transaction",
			}, nil
		}

		// Transaction failed - clean up and retry
		log.Println("❌ Transaction failed, cleaning up WAL...")
		m.walClient.Delete(config.AgentID)
	}

	// No pending transaction or it failed - start fresh
	return m.syncAndMint(ctx, config, wal.ConfigHash, "")
}

// GenerateConfigHash generates a canonical v4 hash of the agent config.
// v4 includes ALL command fields (description, argument, parameters, strictArg, minArgs,
// maxArgs, priceType, taskUnit, timeUnit) and capability descriptions — so any change
// to any field triggers an IPFS re-upload.
// Image is deliberately excluded — image changes are cosmetic, not functional.
func GenerateConfigHash(config *AgentConfig) string {
	// Sort capabilities alphabetically by name (include description)
	caps := make([]Capability, len(config.Capabilities))
	copy(caps, config.Capabilities)
	sort.Slice(caps, func(i, j int) bool {
		return caps[i].Name < caps[j].Name
	})
	capParts := make([]string, len(caps))
	for i, c := range caps {
		capParts[i] = c.Name + "~" + c.Description
	}

	// Sort categories
	categories := make([]string, len(config.Categories))
	copy(categories, config.Categories)
	sort.Strings(categories)

	// Build deterministic string (no image)
	parts := []string{
		"v4",
		config.AgentID,
		config.Name,
		config.Description,
		config.AgentType,
		strings.Join(capParts, ","),
		strconv.FormatBool(config.NlpFallback),
		strings.Join(categories, ","),
	}

	// Include full command data (sorted by trigger for determinism)
	if len(config.Commands) > 0 {
		cmds := make([]Command, len(config.Commands))
		copy(cmds, config.Commands)
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Trigger < cmds[j].Trigger
		})

		cmdParts := make([]string, len(cmds))
		for i, cmd := range cmds {
			// Sort parameters by name
			params := make([]CommandParameter, len(cmd.Parameters))
			copy(params, cmd.Parameters)
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
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Mint is a convenience function for simple minting
func MintAgent(jsonPath string, config *MintConfig) (*MintResult, error) {
	if config == nil {
		config = &MintConfig{}
	}

	minter, err := NewMinter(config)
	if err != nil {
		return nil, err
	}

	return minter.Mint(jsonPath)
}

// Abandon abandons an unminted agent reservation
func (m *Minter) Abandon(agentID string) error {
	// Create authenticator
	authenticator, err := NewAuthenticator(m.config.PrivateKey, m.httpClient)
	if err != nil {
		return fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Get challenge
	challenge, err := m.httpClient.GetChallenge(authenticator.GetAddress())
	if err != nil {
		return fmt.Errorf("failed to get challenge: %w", err)
	}

	// Sign challenge
	signature, err := authenticator.SignChallenge(challenge)
	if err != nil {
		return fmt.Errorf("failed to sign challenge: %w", err)
	}

	// Call abandon endpoint
	abandonReq := &AbandonRequest{
		Wallet:    authenticator.GetAddress(),
		AgentID:   agentID,
		Challenge: challenge,
		Signature: signature,
	}

	_, err = m.httpClient.Abandon(abandonReq)
	if err != nil {
		return fmt.Errorf("abandon failed: %w", err)
	}

	// Clean up WAL if exists
	m.walClient.Delete(agentID)

	log.Printf("✅ Reservation abandoned: %s", agentID)
	return nil
}

// AbandonAgent is a convenience function to abandon a reservation
func AbandonAgent(agentID string, config *MintConfig) error {
	if config == nil {
		config = &MintConfig{}
	}

	minter, err := NewMinter(config)
	if err != nil {
		return err
	}

	return minter.Abandon(agentID)
}
