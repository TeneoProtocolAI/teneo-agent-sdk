package nft

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	sdktypes "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// AgentMetadata represents the metadata for an agent NFT
type AgentMetadata struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Image        string                 `json:"image"`
	AgentID      string                 `json:"agent_id"`
	Capabilities []sdktypes.Capability  `json:"capabilities"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// IPFSUploadResponse represents the response from IPFS upload
type IPFSUploadResponse struct {
	Success  bool   `json:"success"`
	IpfsHash string `json:"ipfsHash"`
	PinSize  int64  `json:"pinSize"`
	Error    string `json:"error,omitempty"`
}

// MintSignatureRequest represents the request to get a mint signature
type MintSignatureRequest struct {
	To       string `json:"to"`
	TokenURI string `json:"tokenURI"`
	Nonce    uint64 `json:"nonce"`
	AgentID  string `json:"agent_id"`
}

// MintSignatureResponse represents the response with mint signature
type MintSignatureResponse struct {
	Signature string `json:"signature"`
	Nonce     uint64 `json:"nonce"`
}

// ContractConfigResponse represents the contract configuration
type ContractConfigResponse struct {
	ContractAddress string `json:"contract_address"`
	ChainID         string `json:"chain_id"`
	NetworkName     string `json:"network_name"`
}

const (
	mintStatusSynced         = "SYNCED"
	mintStatusMintRequired   = "MINT_REQUIRED"
	mintStatusResumeMint     = "RESUME_MINT"
	mintStatusUpdateRequired = "UPDATE_REQUIRED"
)

// MintOrResumeResult represents idempotent mint outcome.
type MintOrResumeResult struct {
	Status      string `json:"status"`
	TokenID     uint64 `json:"token_id"`
	MetadataURI string `json:"metadata_uri,omitempty"`
	TxHash      string `json:"tx_hash,omitempty"`
}

type sdkAgentPayload struct {
	Name             string                 `json:"name"`
	AgentID          string                 `json:"agent_id"`
	Description      string                 `json:"description"`
	ShortDescription string                 `json:"short_description,omitempty"`
	Image            string                 `json:"image,omitempty"`
	AgentType        string                 `json:"agent_type"`
	Capabilities     json.RawMessage        `json:"capabilities"`
	Commands         json.RawMessage        `json:"commands,omitempty"`
	NlpFallback      bool                   `json:"nlp_fallback"`
	Categories       json.RawMessage        `json:"categories,omitempty"`
	TutorialURL      string                 `json:"tutorial_url,omitempty"`
	FAQItems         json.RawMessage        `json:"faq_items,omitempty"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
	MetadataVersion  string                 `json:"metadata_version,omitempty"`
}

type sdkChallengeResponse struct {
	Challenge string `json:"challenge"`
}

type sdkVerifyResponse struct {
	SessionToken string `json:"session_token"`
}

type sdkSyncResponse struct {
	Status          string `json:"status"`
	TokenID         *int64 `json:"token_id,omitempty"`
	ContractAddress string `json:"contract_address,omitempty"`
	AgentID         string `json:"agent_id,omitempty"`
	CurrentHash     string `json:"current_hash,omitempty"`
	NewHash         string `json:"new_hash,omitempty"`
	Message         string `json:"message,omitempty"`
}

type sdkDeployResponse struct {
	Signature       string `json:"signature"`
	Nonce           uint64 `json:"nonce"`
	ContractAddress string `json:"contract_address"`
	ChainID         string `json:"chain_id"`
	IpfsHash        string `json:"ipfs_hash"`
	MetadataURI     string `json:"metadata_uri"`
	AgentID         string `json:"agent_id"`
	// Gasless fields (populated when server does the minting)
	TokenID int64  `json:"token_id,omitempty"`
	TxHash  string `json:"tx_hash,omitempty"`
	Gasless bool   `json:"gasless"`
}

type sdkUpdateResponse struct {
	Success     bool   `json:"success"`
	MetadataURI string `json:"metadata_uri,omitempty"`
	TxHash      string `json:"tx_hash,omitempty"`
	Message     string `json:"message,omitempty"`
}

// NFTMinter handles NFT minting operations
type NFTMinter struct {
	client          *ethclient.Client
	contractAddress common.Address
	backendURL      string
	chainID         *big.Int
	privateKey      *ecdsa.PrivateKey
	address         common.Address
	httpClient      *http.Client
}

const defaultBackendURL = "https://backend.developer.chatroom.teneo-protocol.ai"

// Mint reads PRIVATE_KEY from env and mints (or resumes) an agent from a JSON metadata file.
// Returns the token ID, tx hash, and metadata URI. This is the simplest way to deploy an agent.
func Mint(jsonFilePath string) (*MintOrResumeResult, error) {
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		return nil, fmt.Errorf("PRIVATE_KEY environment variable is required")
	}
	return MintWithKey(privateKey, jsonFilePath)
}

// MintWithKey mints (or resumes) an agent from a JSON metadata file using the provided private key.
func MintWithKey(privateKeyHex, jsonFilePath string) (*MintOrResumeResult, error) {
	minter, err := NewNFTMinter(privateKeyHex)
	if err != nil {
		return nil, err
	}
	return minter.MintOrResumeFromJSONFile(jsonFilePath)
}

// NewNFTMinter creates a new NFT minter. Only requires your private key.
// Backend URL and RPC are handled automatically.
func NewNFTMinter(privateKeyHex string) (*NFTMinter, error) {
	return NewNFTMinterWithConfig(defaultBackendURL, "", privateKeyHex)
}

// NewNFTMinterWithConfig creates a minter with custom backend URL and RPC endpoint.
// Most users should use NewNFTMinter instead.
func NewNFTMinterWithConfig(backendURL, rpcEndpoint, privateKeyHex string) (*NFTMinter, error) {
	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Get address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	// Create Ethereum client if RPC endpoint provided
	var ethClient *ethclient.Client
	if rpcEndpoint != "" {
		ethClient, err = ethclient.Dial(rpcEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
		}
	}

	return &NFTMinter{
		client:     ethClient,
		backendURL: backendURL,
		privateKey: privateKey,
		address:    address,
		httpClient: httpClient,
	}, nil
}

// MintAgent mints a new agent NFT
func (m *NFTMinter) MintAgent(metadata AgentMetadata) (uint64, error) {
	fmt.Println("   [Step 1/5] 🔍 Getting contract configuration...")
	// 1. Get contract configuration from backend
	config, err := m.getContractConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to get contract config: %w", err)
	}

	// Set contract address
	m.contractAddress = common.HexToAddress(config.ContractAddress)
	fmt.Printf("   ✅ Contract address: %s\n", config.ContractAddress)

	// Set chain ID
	chainID, ok := new(big.Int).SetString(config.ChainID, 10)
	if !ok {
		return 0, fmt.Errorf("invalid chain ID: %s", config.ChainID)
	}
	m.chainID = chainID
	fmt.Printf("   ✅ Chain ID: %s\n", config.ChainID)

	fmt.Println("\n   [Step 2/5] 📤 Uploading metadata to IPFS...")
	// 2. Send metadata to backend (backend handles IPFS upload via Pinata)
	ipfsHash, err := m.uploadMetadataToIPFS(metadata)
	if err != nil {
		return 0, fmt.Errorf("failed to send metadata to backend: %w", err)
	}
	fmt.Printf("   ✅ IPFS URI: %s\n", ipfsHash)

	fmt.Println("\n   [Step 3/5] 🔢 Getting nonce from contract...")
	// 3. Get current nonce from contract for this wallet
	nonce, err := m.getNonce(m.address)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	fmt.Printf("   ✅ Nonce: %d\n", nonce)

	fmt.Println("\n   [Step 4/5] 🔐 Requesting mint signature...")
	// 4. Request mint signature from backend (passing wallet address + IPFS URI + nonce)
	signature, err := m.requestMintSignature(m.address.Hex(), metadata.AgentID, ipfsHash, nonce)
	if err != nil {
		return 0, fmt.Errorf("failed to get mint signature: %w", err)
	}

	fmt.Println("\n   [Step 5/5] ⛓️  Executing blockchain transaction...")
	// 5. Execute mint transaction on-chain with the signature
	tokenID, err := m.executeMint(signature)
	if err != nil {
		return 0, fmt.Errorf("failed to execute mint: %w", err)
	}

	return tokenID, nil
}

// MintAgentFromJSONFile mints a new agent NFT from a JSON metadata file.
func (m *NFTMinter) MintAgentFromJSONFile(path string) (uint64, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read metadata file: %w", err)
	}
	result, err := m.MintOrResumeFromJSON(fileBytes)
	if err != nil {
		return 0, err
	}
	return result.TokenID, nil
}

// MintAgentFromJSON mints a new agent NFT from raw JSON metadata.
// This path is compatible with the backend template schema used by /api/ipfs/upload-metadata.
func (m *NFTMinter) MintAgentFromJSON(rawJSON []byte) (uint64, error) {
	result, err := m.MintOrResumeFromJSON(rawJSON)
	if err != nil {
		return 0, err
	}
	return result.TokenID, nil
}

// MintOrResumeFromJSONFile executes idempotent mint-or-login flow from file.
func (m *NFTMinter) MintOrResumeFromJSONFile(path string) (*MintOrResumeResult, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}
	return m.MintOrResumeFromJSON(fileBytes)
}

// MintOrResumeFromJSON executes idempotent mint flow using SDK sync/deploy/update endpoints.
func (m *NFTMinter) MintOrResumeFromJSON(rawJSON []byte) (*MintOrResumeResult, error) {
	if !json.Valid(rawJSON) {
		return nil, fmt.Errorf("invalid metadata json")
	}

	config, canonicalJSON, configHash, err := m.parsePayloadAndHash(rawJSON)
	if err != nil {
		return nil, err
	}
	_ = canonicalJSON

	fmt.Println("   [Step 1/4] 🔍 Syncing mint state...")
	syncResp, err := m.syncAgentState(config.AgentID, configHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sync agent state: %w", err)
	}

	if syncResp.TokenID != nil && *syncResp.TokenID > 0 && syncResp.Status == mintStatusSynced {
		fmt.Printf("   ✅ Already minted, token id: %d\n", *syncResp.TokenID)
		return &MintOrResumeResult{
			Status:  syncResp.Status,
			TokenID: uint64(*syncResp.TokenID),
		}, nil
	}

	fmt.Println("   [Step 2/4] 🔐 Authenticating SDK session...")
	sessionToken, err := m.authenticateSDKSession()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate sdk session: %w", err)
	}

	if syncResp.Status == mintStatusUpdateRequired {
		fmt.Println("   [Step 3/4] ♻️  Updating metadata/tokenURI...")
		updateResp, updateErr := m.callSDKUpdate(sessionToken, config, configHash)
		if updateErr != nil {
			return nil, fmt.Errorf("failed to update metadata: %w", updateErr)
		}
		if syncResp.TokenID == nil || *syncResp.TokenID <= 0 {
			return nil, fmt.Errorf("update response missing existing token id")
		}
		return &MintOrResumeResult{
			Status:      syncResp.Status,
			TokenID:     uint64(*syncResp.TokenID),
			MetadataURI: updateResp.MetadataURI,
			TxHash:      updateResp.TxHash,
		}, nil
	}

	if syncResp.Status != mintStatusMintRequired && syncResp.Status != mintStatusResumeMint {
		return nil, fmt.Errorf("unsupported sync status: %s", syncResp.Status)
	}

	fmt.Println("   [Step 3/4] 📦 Preparing deploy + mint tx...")
	deployResp, err := m.callSDKDeploy(sessionToken, config, configHash)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare deploy: %w", err)
	}

	// If server did gasless minting, skip client-side chain interaction
	if deployResp.Gasless {
		if deployResp.TokenID <= 0 {
			return nil, fmt.Errorf("server returned gasless=true but invalid token_id=%d — this is a server error, not retrying to prevent double-mint", deployResp.TokenID)
		}
		fmt.Printf("   ✅ Gasless mint! Token ID: %d, Tx: %s\n", deployResp.TokenID, deployResp.TxHash)
		return &MintOrResumeResult{
			Status:      syncResp.Status,
			TokenID:     uint64(deployResp.TokenID),
			MetadataURI: deployResp.MetadataURI,
			TxHash:      deployResp.TxHash,
		}, nil
	}

	// Server must perform gasless minting — client-side minting is not supported
	return nil, fmt.Errorf("server did not perform gasless minting (gasless=false in response) — ensure your backend supports gasless minting")
}

func (m *NFTMinter) parsePayloadAndHash(rawJSON []byte) (*sdkAgentPayload, []byte, string, error) {
	var config sdkAgentPayload
	if err := json.Unmarshal(rawJSON, &config); err != nil {
		return nil, nil, "", fmt.Errorf("failed to parse metadata json: %w", err)
	}
	config.AgentID = strings.TrimSpace(config.AgentID)
	config.Name = strings.TrimSpace(config.Name)
	config.Description = strings.TrimSpace(config.Description)
	config.AgentType = strings.TrimSpace(config.AgentType)
	if config.AgentID == "" || config.Name == "" || config.Description == "" || config.AgentType == "" {
		return nil, nil, "", fmt.Errorf("metadata json missing required fields: agent_id, name, description, agent_type")
	}
	if len(config.Capabilities) == 0 {
		return nil, nil, "", fmt.Errorf("metadata json missing required field: capabilities")
	}
	if len(config.Categories) == 0 {
		return nil, nil, "", fmt.Errorf("metadata json missing required field: categories")
	}
	if config.MetadataVersion == "" {
		config.MetadataVersion = "2.4.0"
	}

	// Compute v3 canonical config hash (must match server's GenerateConfigHash)
	// Uses pipe-delimited string of specific fields, excludes image/properties/metadata_version
	configHash := m.generateCanonicalConfigHash(&config)

	canonicalJSON, err := json.Marshal(config)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to serialize metadata json: %w", err)
	}
	return &config, canonicalJSON, configHash, nil
}

// generateCanonicalConfigHash produces the v3 config hash matching the server algorithm.
// Includes: agentId, name, description, agentType, capabilities, nlpFallback, categories, command triggers+prices.
// Image is deliberately excluded — image changes are cosmetic, not functional.
// generateCanonicalConfigHash produces the v4 config hash matching the server algorithm.
// v4 includes ALL command fields (description, argument, parameters, strictArg, minArgs,
// maxArgs, priceType, taskUnit, timeUnit) and capability descriptions — so any change
// to any field triggers an IPFS re-upload.
// Image is deliberately excluded — image changes are cosmetic, not functional.
func (m *NFTMinter) generateCanonicalConfigHash(config *sdkAgentPayload) string {
	// Parse capabilities from JSON (name + description)
	type capEntry struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	var caps []capEntry
	if len(config.Capabilities) > 0 {
		_ = json.Unmarshal(config.Capabilities, &caps)
	}
	sort.Slice(caps, func(i, j int) bool {
		return caps[i].Name < caps[j].Name
	})
	capParts := make([]string, len(caps))
	for i, c := range caps {
		capParts[i] = c.Name + "~" + c.Description
	}

	// Parse categories from JSON
	var categories []string
	if len(config.Categories) > 0 {
		_ = json.Unmarshal(config.Categories, &categories)
	}
	sort.Strings(categories)

	// Build deterministic string
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

	// Parse and include full command data (sorted by trigger)
	type paramEntry struct {
		Name           string `json:"name"`
		Type           string `json:"type"`
		Required       bool   `json:"required"`
		Description    string `json:"description"`
		MinValue       string `json:"minValue"`
		IsBillingCount bool   `json:"isBillingCount"`
	}
	type cmdEntry struct {
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
	var cmds []cmdEntry
	if len(config.Commands) > 0 {
		_ = json.Unmarshal(config.Commands, &cmds)
	}
	if len(cmds) > 0 {
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Trigger < cmds[j].Trigger
		})
		cmdParts := make([]string, len(cmds))
		for i, cmd := range cmds {
			// Parse parameters for this command
			var params []paramEntry
			if len(cmd.Parameters) > 0 {
				_ = json.Unmarshal(cmd.Parameters, &params)
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
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (m *NFTMinter) syncAgentState(agentID, configHash string) (*sdkSyncResponse, error) {
	challenge, err := m.requestSDKChallenge()
	if err != nil {
		return nil, err
	}
	signature, err := m.signSDKChallenge(challenge)
	if err != nil {
		return nil, err
	}

	req := map[string]interface{}{
		"wallet":      m.address.Hex(),
		"agent_id":    agentID,
		"config_hash": configHash,
		"challenge":   challenge,
		"signature":   signature,
	}
	var resp sdkSyncResponse
	if err := m.postJSON("/api/sdk/agent/sync", req, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *NFTMinter) authenticateSDKSession() (string, error) {
	challenge, err := m.requestSDKChallenge()
	if err != nil {
		return "", err
	}
	signature, err := m.signSDKChallenge(challenge)
	if err != nil {
		return "", err
	}

	req := map[string]string{
		"wallet_address": m.address.Hex(),
		"challenge":      challenge,
		"signature":      signature,
	}
	var resp sdkVerifyResponse
	if err := m.postJSON("/api/sdk/auth/verify", req, nil, &resp); err != nil {
		return "", err
	}
	if resp.SessionToken == "" {
		return "", fmt.Errorf("empty sdk session token")
	}
	return resp.SessionToken, nil
}

func (m *NFTMinter) requestSDKChallenge() (string, error) {
	req := map[string]string{"wallet_address": m.address.Hex()}
	var resp sdkChallengeResponse
	if err := m.postJSON("/api/sdk/auth/challenge", req, nil, &resp); err != nil {
		return "", err
	}
	if resp.Challenge == "" {
		return "", fmt.Errorf("empty sdk challenge")
	}
	return resp.Challenge, nil
}

func (m *NFTMinter) signSDKChallenge(challenge string) (string, error) {
	message := "Teneo SDK auth: " + challenge
	hash := accounts.TextHash([]byte(message))
	sig, err := crypto.Sign(hash, m.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign sdk challenge: %w", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}
	return hexutil.Encode(sig), nil
}

func (m *NFTMinter) callSDKDeploy(sessionToken string, config *sdkAgentPayload, configHash string) (*sdkDeployResponse, error) {
	req := map[string]interface{}{
		"wallet_address":    m.address.Hex(),
		"agent_id":          config.AgentID,
		"agent_name":        config.Name,
		"description":       config.Description,
		"short_description": config.ShortDescription,
		"image":             config.Image,
		"agent_type":        config.AgentType,
		"capabilities":      json.RawMessage(config.Capabilities),
		"commands":          json.RawMessage(config.Commands),
		"nlp_fallback":      config.NlpFallback,
		"categories":        json.RawMessage(config.Categories),
		"tutorial_url":      config.TutorialURL,
		"faq_items":         json.RawMessage(config.FAQItems),
		"properties":        config.Properties,
		"config_hash":       configHash,
		"metadata_version":  config.MetadataVersion,
	}
	headers := map[string]string{"X-SDK-Session-Token": sessionToken}
	var resp sdkDeployResponse
	if err := m.postJSON("/api/sdk/agent/deploy", req, headers, &resp); err != nil {
		return nil, err
	}
	if resp.Signature == "" && !resp.Gasless {
		return nil, fmt.Errorf("deploy response missing signature")
	}
	return &resp, nil
}

func (m *NFTMinter) callSDKUpdate(sessionToken string, config *sdkAgentPayload, configHash string) (*sdkUpdateResponse, error) {
	req := map[string]interface{}{
		"wallet_address":    m.address.Hex(),
		"agent_id":          config.AgentID,
		"agent_name":        config.Name,
		"description":       config.Description,
		"short_description": config.ShortDescription,
		"image":             config.Image,
		"agent_type":        config.AgentType,
		"capabilities":      json.RawMessage(config.Capabilities),
		"commands":          json.RawMessage(config.Commands),
		"nlp_fallback":      config.NlpFallback,
		"categories":        json.RawMessage(config.Categories),
		"tutorial_url":      config.TutorialURL,
		"faq_items":         json.RawMessage(config.FAQItems),
		"config_hash":       configHash,
		"metadata_version":  config.MetadataVersion,
	}
	headers := map[string]string{"X-SDK-Session-Token": sessionToken}
	var resp sdkUpdateResponse
	if err := m.postJSON("/api/sdk/agent/update", req, headers, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("sdk update failed: %s", resp.Message)
	}
	return &resp, nil
}

func (m *NFTMinter) callSDKConfirmMint(
	sessionToken string,
	config *sdkAgentPayload,
	configHash, metadataURI string,
	tokenID uint64,
	txHash, contractAddress string,
) error {
	req := map[string]interface{}{
		"agent_id":             config.AgentID,
		"agent_name":           config.Name,
		"wallet_address":       m.address.Hex(),
		"token_id":             tokenID,
		"tx_hash":              txHash,
		"metadata_uri":         metadataURI,
		"description":          config.Description,
		"short_description":    config.ShortDescription,
		"image_url":            config.Image,
		"agent_type":           config.AgentType,
		"nlp_fallback":         config.NlpFallback,
		"capabilities":         json.RawMessage(config.Capabilities),
		"commands":             json.RawMessage(config.Commands),
		"nft_contract_address": contractAddress,
		"categories":           json.RawMessage(config.Categories),
		"tutorial_url":         config.TutorialURL,
		"faq_items":            json.RawMessage(config.FAQItems),
		"config_hash":          configHash,
		"metadata_version":     config.MetadataVersion,
	}
	headers := map[string]string{"X-SDK-Session-Token": sessionToken}
	var resp map[string]interface{}
	return m.postJSON("/api/sdk/agent/confirm-mint", req, headers, &resp)
}

func (m *NFTMinter) postJSON(path string, payload interface{}, headers map[string]string, out interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	url := strings.TrimRight(m.backendURL, "/") + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("backend returned status %d: %s", resp.StatusCode, extractErrorMessage(respBody))
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	return nil
}

func extractErrorMessage(body []byte) string {
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if msg, ok := parsed["message"].(string); ok && msg != "" {
			return msg
		}
		if errMsg, ok := parsed["error"].(string); ok && errMsg != "" {
			return errMsg
		}
	}
	return string(body)
}

// uploadMetadataToIPFS sends agent metadata to backend which handles IPFS upload
func (m *NFTMinter) uploadMetadataToIPFS(metadata AgentMetadata) (string, error) {
	// The backend handles the actual IPFS upload via Pinata
	// We just send the metadata to the backend endpoint

	// Prepare request body with agent metadata
	body, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create request to backend
	// Ensure backend URL doesn't have trailing slash
	backendURL := strings.TrimRight(m.backendURL, "/")
	req, err := http.NewRequest("POST", backendURL+"/api/ipfs/upload-metadata", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send metadata to backend
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to backend: %w", err)
	}
	defer resp.Body.Close()

	// Read response from backend
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read backend response: %w", err)
	}

	// Parse response - backend returns IPFS hash after uploading via Pinata
	var uploadResp IPFSUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse backend response: %w", err)
	}

	if !uploadResp.Success {
		return "", fmt.Errorf("backend upload failed: %s", uploadResp.Error)
	}

	// Return IPFS URI that backend created
	return fmt.Sprintf("ipfs://%s", uploadResp.IpfsHash), nil
}

// uploadRawMetadataToIPFS uploads raw JSON metadata to backend endpoint.
func (m *NFTMinter) uploadRawMetadataToIPFS(rawJSON []byte) (string, error) {
	backendURL := strings.TrimRight(m.backendURL, "/")
	req, err := http.NewRequest("POST", backendURL+"/api/ipfs/upload-metadata", bytes.NewBuffer(rawJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to backend: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read backend response: %w", err)
	}

	var generic map[string]interface{}
	if err := json.Unmarshal(respBody, &generic); err != nil {
		return "", fmt.Errorf("failed to parse backend response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if errMsg, ok := generic["error"].(string); ok && errMsg != "" {
			return "", fmt.Errorf("backend upload failed: %s", errMsg)
		}
		return "", fmt.Errorf("backend upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if success, ok := generic["success"].(bool); !ok || !success {
		if errMsg, ok := generic["error"].(string); ok && errMsg != "" {
			return "", fmt.Errorf("backend upload failed: %s", errMsg)
		}
		return "", fmt.Errorf("backend upload failed")
	}

	if ipfsURL, ok := generic["ipfs_url"].(string); ok && ipfsURL != "" {
		return ipfsURL, nil
	}
	if ipfsHash, ok := generic["ipfs_hash"].(string); ok && ipfsHash != "" {
		return fmt.Sprintf("ipfs://%s", ipfsHash), nil
	}
	if ipfsHash, ok := generic["ipfsHash"].(string); ok && ipfsHash != "" {
		return fmt.Sprintf("ipfs://%s", ipfsHash), nil
	}

	return "", fmt.Errorf("backend upload succeeded but no ipfs uri/hash returned")
}

// getContractConfig gets the contract configuration from backend
func (m *NFTMinter) getContractConfig() (*ContractConfigResponse, error) {
	// Ensure backend URL doesn't have trailing slash
	backendURL := strings.TrimRight(m.backendURL, "/")
	endpoint := backendURL + "/api/contract/config"

	fmt.Printf("   📡 Fetching contract config from: %s\n", endpoint)

	// Create request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("backend returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response for debugging if it looks like HTML
	if len(respBody) > 0 && respBody[0] == '<' {
		preview := string(respBody)
		if len(preview) > 100 {
			preview = preview[:100]
		}
		return nil, fmt.Errorf("backend returned HTML instead of JSON. Response starts with: %s", preview)
	}

	// Parse response
	var config ContractConfigResponse
	if err := json.Unmarshal(respBody, &config); err != nil {
		// Include part of the response in error for debugging
		preview := string(respBody)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		return nil, fmt.Errorf("failed to parse response: %w. Response: %s", err, preview)
	}

	return &config, nil
}

// getNonce gets the current nonce for an address from the contract
func (m *NFTMinter) getNonce(address common.Address) (uint64, error) {
	if m.client == nil {
		// If no Ethereum client, assume nonce is 0 (first mint)
		return 0, nil
	}

	// Parse the contract ABI
	contractABI, err := ParseABI()
	if err != nil {
		return 0, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the nonces method call
	data, err := contractABI.Pack("nonces", address)
	if err != nil {
		return 0, fmt.Errorf("failed to pack nonces call: %w", err)
	}

	// Call the contract
	result, err := m.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &m.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call nonces: %w", err)
	}

	// Unpack the result
	var nonce *big.Int
	err = contractABI.UnpackIntoInterface(&nonce, "nonces", result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack nonce: %w", err)
	}

	return nonce.Uint64(), nil
}

// requestMintSignature requests a mint signature from the backend
func (m *NFTMinter) requestMintSignature(to, agentID string, tokenURI string, nonce uint64) (string, error) {
	// Show progress
	fmt.Printf("   📝 Requesting mint signature from backend...\n")

	// Prepare request
	// Note: tokenURI is not used in signature generation but sent for compatibility
	reqBody := MintSignatureRequest{
		To:       to,
		TokenURI: "", // Empty as per backend expectation
		Nonce:    nonce,
		AgentID:  agentID,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	// Ensure backend URL doesn't have trailing slash
	backendURL := strings.TrimRight(m.backendURL, "/")
	endpoint := backendURL + "/api/signature/generate-mint"

	fmt.Printf("   📡 Sending request to: %s\n", endpoint)
	fmt.Printf("   📦 Request data: to=%s, agentID=%s, nonce=%d, tokenURI=\"\" (not used in signature)\n", to, agentID, nonce)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("backend returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Log the response status
	fmt.Printf("   📨 Response status: %d\n", resp.StatusCode)

	// Check if response is HTML (error page)
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || (len(respBody) > 0 && respBody[0] == '<') {
		preview := string(respBody)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}

		// Try to extract meaningful error from HTML
		errorMsg := "Backend returned HTML instead of JSON. "
		if strings.Contains(preview, "404") || strings.Contains(preview, "Not Found") {
			errorMsg += "The mint signature endpoint may not be available at this URL. "
		} else if strings.Contains(preview, "502") || strings.Contains(preview, "Bad Gateway") {
			errorMsg += "The backend server may be down or unreachable. "
		}

		fmt.Printf("   ❌ Error: %s\n", errorMsg)
		fmt.Printf("   📄 HTML Response preview:\n%s\n", preview)

		return "", fmt.Errorf("%sPlease check the backend URL configuration", errorMsg)
	}

	// Don't log raw response as it contains sensitive signature data

	// Parse response
	var sigResp MintSignatureResponse
	if err := json.Unmarshal(respBody, &sigResp); err != nil {
		// Check for common JSON parsing errors
		if strings.Contains(err.Error(), "invalid character 'p'") {
			// This often means we got a plain text error starting with 'p'
			if strings.HasPrefix(string(respBody), "property") || strings.HasPrefix(string(respBody), "please") {
				return "", fmt.Errorf("backend error: %s", string(respBody))
			}
		}

		// Include part of the response in error for debugging
		preview := string(respBody)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return "", fmt.Errorf("failed to parse JSON response: %w. Response: %s", err, preview)
	}

	// Validate signature response
	if sigResp.Signature == "" {
		return "", fmt.Errorf("backend returned empty signature")
	}

	fmt.Printf("   ✅ Received signature successfully\n")
	fmt.Printf("   ✅ Nonce confirmed: %d\n", sigResp.Nonce)
	return sigResp.Signature, nil
}

// executeMint executes the mint transaction on the blockchain
func (m *NFTMinter) executeMint(signature string) (uint64, error) {
	tokenID, _, err := m.executeMintWithTxHash(signature)
	if err != nil {
		return 0, err
	}
	return tokenID, nil
}

func (m *NFTMinter) executeMintWithTxHash(signature string) (uint64, string, error) {
	if m.client == nil {
		return 0, "", fmt.Errorf("ethereum client not initialized")
	}

	// Parse the contract ABI
	contractABI, err := ParseABI()
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Decode signature from hex
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return 0, "", fmt.Errorf("failed to decode signature: %w", err)
	}

	// Pack the mint method call
	data, err := contractABI.Pack("mint", m.address, sigBytes)
	if err != nil {
		return 0, "", fmt.Errorf("failed to pack mint call: %w", err)
	}

	// Get the current gas price
	gasPrice, err := m.client.SuggestGasPrice(context.Background())
	if err != nil {
		return 0, "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Get the nonce for the transaction
	nonce, err := m.client.PendingNonceAt(context.Background(), m.address)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get account nonce: %w", err)
	}

	// Create the transaction
	tx := types.NewTransaction(
		nonce,
		m.contractAddress,
		DefaultMintPrice(), // 0.01 ETH mint price
		uint64(300000),     // Gas limit
		gasPrice,
		data,
	)

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(m.chainID), m.privateKey)
	if err != nil {
		return 0, "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = m.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return 0, "", fmt.Errorf("failed to send transaction: %w", err)
	}

	fmt.Printf("Mint transaction sent: %s\n", signedTx.Hash().Hex())
	txHash := signedTx.Hash().Hex()

	// Wait for transaction receipt
	receipt, err := m.WaitForTransaction(context.Background(), signedTx)
	if err != nil {
		return 0, txHash, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	// Extract token ID from logs
	// The Transfer event has the token ID as the third topic
	for _, log := range receipt.Logs {
		if len(log.Topics) >= 4 && log.Address == m.contractAddress {
			// Transfer event signature
			transferEventSig := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
			if log.Topics[0] == transferEventSig {
				// Token ID is in the third topic
				tokenID := new(big.Int).SetBytes(log.Topics[3].Bytes())
				return tokenID.Uint64(), txHash, nil
			}
		}
	}

	// If we couldn't find the token ID in logs, return an error
	return 0, txHash, fmt.Errorf("could not extract token ID from transaction logs")
}

// GenerateMetadataHash generates a SHA256 hash of agent metadata
func GenerateMetadataHash(metadata AgentMetadata) string {
	// Create deterministic string representation
	data := fmt.Sprintf("%s:%s:%s:%s",
		metadata.Name,
		metadata.Description,
		metadata.Image,
		strings.Join(sdktypes.CapabilityNames(metadata.Capabilities), ","))

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SendMetadataHashToBackend sends the metadata hash for an existing agent
func (m *NFTMinter) SendMetadataHashToBackend(hash string, tokenID uint64, walletAddress string) error {
	// TODO: Implement backend endpoint for metadata hash submission
	// This would send the hash along with the token ID to verify ownership

	reqBody := map[string]interface{}{
		"hash":          hash,
		"tokenId":       tokenID,
		"walletAddress": walletAddress,
	}

	_, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// For now, we'll just log this operation
	// In production, this would make an actual HTTP request
	fmt.Printf("Would send metadata hash: %s for token ID: %d\n", hash, tokenID)

	return nil
}

// GetAddress returns the address associated with the minter
func (m *NFTMinter) GetAddress() common.Address {
	return m.address
}

// WaitForTransaction waits for a transaction to be confirmed
func (m *NFTMinter) WaitForTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	// Wait for transaction receipt
	receipt, err := bind.WaitMined(ctx, m.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	// Check if transaction was successful
	if receipt.Status == 0 {
		return nil, fmt.Errorf("transaction failed")
	}

	return receipt, nil
}
