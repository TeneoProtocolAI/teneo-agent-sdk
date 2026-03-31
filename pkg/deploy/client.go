package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/version"
)

// HTTPClient wraps HTTP operations for SDK deploy endpoints
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// ChallengeRequest is the request body for /api/sdk/auth/challenge
type ChallengeRequest struct {
	WalletAddress string `json:"wallet_address"`
}

// ChallengeResponse is the response from /api/sdk/auth/challenge
type ChallengeResponse struct {
	Challenge string `json:"challenge"`
	ExpiresAt int64  `json:"expires_at"`
}

// VerifyRequest is the request body for /api/sdk/auth/verify
type VerifyRequest struct {
	WalletAddress string `json:"wallet_address"`
	Challenge     string `json:"challenge"`
	Signature     string `json:"signature"`
}

// VerifyResponse is the response from /api/sdk/auth/verify
type VerifyResponse struct {
	SessionToken string `json:"session_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// DeployRequest is the request body for /api/sdk/agent/deploy
type DeployRequest struct {
	WalletAddress    string                 `json:"wallet_address"`
	AgentID          string                 `json:"agent_id"`
	AgentName        string                 `json:"agent_name"`
	Description      string                 `json:"description"`
	Image            string                 `json:"image,omitempty"`
	AgentType        string                 `json:"agent_type"`
	Capabilities     json.RawMessage        `json:"capabilities"`
	Commands         json.RawMessage        `json:"commands,omitempty"`
	NlpFallback      bool                   `json:"nlp_fallback"`
	Categories       json.RawMessage        `json:"categories,omitempty"`
	ShortDescription string                 `json:"short_description,omitempty"`
	TutorialURL      string                 `json:"tutorial_url,omitempty"`
	FAQItems         json.RawMessage        `json:"faq_items,omitempty"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
	ConfigHash       string                 `json:"config_hash"`
	MetadataVersion  string                 `json:"metadata_version,omitempty"`
}

// DeployResponse is the response from /api/sdk/agent/deploy
type DeployResponse struct {
	Signature       string `json:"signature"`
	Nonce           uint64 `json:"nonce"`
	ContractAddress string `json:"contract_address"`
	ChainID         string `json:"chain_id"`
	RPCURL          string `json:"rpc_url"`
	AgentID         string `json:"agent_id"`
	ConfigHash      string `json:"config_hash"`
	// Gasless minting fields (populated when server mints on behalf of SDK)
	TokenID     int64  `json:"token_id,omitempty"`
	TxHash      string `json:"tx_hash,omitempty"`
	MetadataURI string `json:"metadata_uri,omitempty"`
	Gasless     bool   `json:"gasless"`
}

// ConfirmMintRequest is the request body for /api/sdk/agent/confirm-mint.
type ConfirmMintRequest struct {
	AgentID          string          `json:"agent_id"`
	AgentName        string          `json:"agent_name"`
	WalletAddress    string          `json:"wallet_address"`
	TokenID          int64           `json:"token_id"`
	TxHash           string          `json:"tx_hash"`
	ConfigHash       string          `json:"config_hash"`
	Description      string          `json:"description,omitempty"`
	Image            string          `json:"image,omitempty"`
	AgentType        string          `json:"agent_type,omitempty"`
	Capabilities     json.RawMessage `json:"capabilities,omitempty"`
	Commands         json.RawMessage `json:"commands,omitempty"`
	NlpFallback      bool            `json:"nlp_fallback"`
	Categories       json.RawMessage `json:"categories,omitempty"`
	ShortDescription string          `json:"short_description,omitempty"`
	TutorialURL      string          `json:"tutorial_url,omitempty"`
	FAQItems         json.RawMessage `json:"faq_items,omitempty"`
	MetadataVersion  string          `json:"metadata_version,omitempty"`
}

// ConfirmMintResponse is the response from /api/sdk/agent/confirm-mint
type ConfirmMintResponse struct {
	Success     bool   `json:"success"`
	ID          string `json:"id"`
	Message     string `json:"message"`
	MetadataURI string `json:"metadata_uri,omitempty"`
}

// UpdateMetadataRequest is the request body for POST /api/sdk/agent/update
type UpdateMetadataRequest struct {
	WalletAddress    string          `json:"wallet_address"`
	AgentID          string          `json:"agent_id"`
	AgentName        string          `json:"agent_name"`
	Description      string          `json:"description"`
	Image            string          `json:"image,omitempty"`
	AgentType        string          `json:"agent_type"`
	Capabilities     json.RawMessage `json:"capabilities"`
	Commands         json.RawMessage `json:"commands,omitempty"`
	NlpFallback      bool            `json:"nlp_fallback"`
	Categories       json.RawMessage `json:"categories,omitempty"`
	ShortDescription string          `json:"short_description,omitempty"`
	TutorialURL      string          `json:"tutorial_url,omitempty"`
	FAQItems         json.RawMessage `json:"faq_items,omitempty"`
	ConfigHash       string          `json:"config_hash"`
	MetadataVersion  string          `json:"metadata_version,omitempty"`
}

// UpdateMetadataResponse is the response from POST /api/sdk/agent/update
type UpdateMetadataResponse struct {
	Success     bool   `json:"success"`
	IpfsHash    string `json:"ipfs_hash,omitempty"`
	MetadataURI string `json:"metadata_uri,omitempty"`
	TxHash      string `json:"tx_hash,omitempty"`
	Message     string `json:"message"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// NewHTTPClient creates a new HTTP client for SDK endpoints
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// RequestChallenge requests an authentication challenge from the backend
func (c *HTTPClient) RequestChallenge(walletAddress string) (*ChallengeResponse, error) {
	reqBody := ChallengeRequest{WalletAddress: walletAddress}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal challenge request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/sdk/auth/challenge",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to request challenge: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read challenge response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("challenge request failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("challenge request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ChallengeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse challenge response: %w", err)
	}

	return &result, nil
}

// VerifySignature verifies the signed challenge and returns a session token
func (c *HTTPClient) VerifySignature(walletAddress, challenge, signature string) (*VerifyResponse, error) {
	reqBody := VerifyRequest{
		WalletAddress: walletAddress,
		Challenge:     challenge,
		Signature:     signature,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal verify request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/sdk/auth/verify",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read verify response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("authentication failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("authentication failed")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("verify request failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("verify request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result VerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse verify response: %w", err)
	}

	return &result, nil
}

// Deploy calls the deploy endpoint to prepare for minting
func (c *HTTPClient) Deploy(sessionToken string, req *DeployRequest) (*DeployResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deploy request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/api/sdk/agent/deploy",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create deploy request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-SDK-Session-Token", sessionToken)
	httpReq.Header.Set("X-SDK-Version", version.Version())

	// Use a longer timeout for deploy — gasless minting can take up to 3 minutes
	deployClient := &http.Client{Timeout: 180 * time.Second}
	resp, err := deployClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call deploy endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read deploy response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrSessionExpired
	}

	if resp.StatusCode == http.StatusConflict {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("conflict: %s", errResp.Error)
		}
		return nil, fmt.Errorf("agent already exists")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("deploy failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("deploy failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result DeployResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse deploy response: %w", err)
	}

	return &result, nil
}

// ConfirmMint confirms the mint and saves the agent to the database
func (c *HTTPClient) ConfirmMint(sessionToken string, req *ConfirmMintRequest) (*ConfirmMintResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal confirm request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/api/sdk/agent/confirm-mint",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create confirm request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-SDK-Session-Token", sessionToken)
	httpReq.Header.Set("X-SDK-Version", version.Version())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call confirm-mint endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read confirm response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrSessionExpired
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("confirm-mint failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("confirm-mint failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ConfirmMintResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse confirm response: %w", err)
	}

	return &result, nil
}

// UpdateMetadata calls the update endpoint to re-upload metadata and update on-chain tokenURI
func (c *HTTPClient) UpdateMetadata(sessionToken string, req *UpdateMetadataRequest) (*UpdateMetadataResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/api/sdk/agent/update",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create update request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-SDK-Session-Token", sessionToken)
	httpReq.Header.Set("X-SDK-Version", version.Version())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call update endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read update response: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("update failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrSessionExpired
	}

	if resp.StatusCode == http.StatusForbidden {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("update forbidden: %s", errResp.Error)
		}
		return nil, fmt.Errorf("update forbidden")
	}

	if resp.StatusCode == http.StatusNotFound {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("agent not found: %s", errResp.Error)
		}
		return nil, fmt.Errorf("agent not found")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded, please wait and retry")
	}

	if resp.StatusCode == http.StatusInternalServerError {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("update server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode == http.StatusServiceUnavailable {
		return nil, fmt.Errorf("update service unavailable, please try again later")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("update failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result UpdateMetadataResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse update response: %w", err)
	}

	return &result, nil
}

// ErrSessionExpired indicates the session token has expired
var ErrSessionExpired = fmt.Errorf("session expired")

// ErrGaslessMintingDisabled indicates gasless minting is disabled
var ErrGaslessMintingDisabled = fmt.Errorf("gasless minting is temporarily disabled")

// ErrSchemaOutdated indicates the schema version is outdated
var ErrSchemaOutdated = fmt.Errorf("schema version outdated")

// SchemaResponse is the response from GET /api/sdk/schema
type SchemaResponse struct {
	Schema        json.RawMessage `json:"schema"`
	SchemaVersion string          `json:"schema_version"`
	Signature     string          `json:"signature"`
	MaxJSONSize   int             `json:"max_json_size"` // Max allowed JSON size in bytes
}

// SchemaCache caches the schema response
type SchemaCache struct {
	Schema    *SchemaResponse
	FetchedAt time.Time
}

// SyncRequest is the request body for POST /api/sdk/agent/sync
type SyncRequest struct {
	Wallet        string `json:"wallet"`
	AgentID       string `json:"agent_id"`
	ConfigHash    string `json:"config_hash"`
	Challenge     string `json:"challenge"`
	Signature     string `json:"signature"`
	SchemaVersion string `json:"schema_version,omitempty"`
}

// SyncResponse is the response from POST /api/sdk/agent/sync
type SyncResponse struct {
	Status             string `json:"status"` // SYNCED, MINT_REQUIRED, RESUME_MINT, UPDATE_REQUIRED, AUTO_CONFIRMED
	TokenID            *int64 `json:"token_id,omitempty"`
	ContractAddress    string `json:"contract_address,omitempty"`
	AgentID            string `json:"agent_id,omitempty"`
	Creator            string `json:"creator,omitempty"`
	CurrentHash        string `json:"current_hash,omitempty"`
	NewHash            string `json:"new_hash,omitempty"`
	Message            string `json:"message,omitempty"`
	HasPendingMetadata bool   `json:"has_pending_metadata,omitempty"`
	RPCURL             string `json:"rpc_url,omitempty"`
	ConfigHash         string `json:"config_hash,omitempty"`
}

// GetSchema fetches the validation schema from the backend
func (c *HTTPClient) GetSchema() (*SchemaResponse, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/sdk/schema")
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("get schema failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("get schema failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result SchemaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse schema response: %w", err)
	}

	return &result, nil
}

// GetChallenge requests a challenge for authentication (used by sync flow)
func (c *HTTPClient) GetChallenge(walletAddress string) (string, error) {
	resp, err := c.RequestChallenge(walletAddress)
	if err != nil {
		return "", err
	}
	return resp.Challenge, nil
}

// Sync calls the sync endpoint to check agent status
func (c *HTTPClient) Sync(req *SyncRequest) (*SyncResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/api/sdk/agent/sync",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Cache-Control", "no-store")
	httpReq.Header.Set("X-SDK-Version", version.Version())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call sync endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sync response: %w", err)
	}

	// Handle specific error codes
	if resp.StatusCode == http.StatusServiceUnavailable {
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if errResp["error"] == "HEADLESS_MINTING_DISABLED" || errResp["error"] == "GASLESS_MINTING_DISABLED" {
				return nil, ErrGaslessMintingDisabled
			}
		}
		return nil, fmt.Errorf("service unavailable: %s", string(body))
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if errResp["error"] == "SCHEMA_OUTDATED" {
				return nil, ErrSchemaOutdated
			}
		}
		return nil, fmt.Errorf("bad request: %s", string(body))
	}

	if resp.StatusCode == http.StatusUnauthorized {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("authentication failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("authentication failed")
	}

	if resp.StatusCode == http.StatusForbidden {
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			msg := errResp["message"]
			if msg != nil {
				return nil, fmt.Errorf("forbidden: %v", msg)
			}
		}
		return nil, fmt.Errorf("forbidden: %s", string(body))
	}

	if resp.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("conflict: agent ID was just reserved by another request, retry")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if errResp["error"] == "MAX_RESERVATIONS" {
				return nil, fmt.Errorf("MAX_RESERVATIONS: %v", errResp["message"])
			}
		}
		return nil, fmt.Errorf("rate limit exceeded, please wait and retry")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("sync failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("sync failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result SyncResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse sync response: %w", err)
	}

	return &result, nil
}

// AbandonRequest is the request body for POST /api/sdk/agent/abandon
type AbandonRequest struct {
	Wallet    string `json:"wallet"`
	AgentID   string `json:"agent_id"`
	Challenge string `json:"challenge"`
	Signature string `json:"signature"`
}

// AbandonResponse is the response from POST /api/sdk/agent/abandon
type AbandonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	AgentID string `json:"agent_id"`
}

// Abandon calls the abandon endpoint to delete an unminted reservation
func (c *HTTPClient) Abandon(req *AbandonRequest) (*AbandonResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal abandon request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/api/sdk/agent/abandon",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create abandon request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-SDK-Version", version.Version())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call abandon endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read abandon response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("reservation not found or already minted")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("abandon failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("abandon failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result AbandonResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse abandon response: %w", err)
	}

	return &result, nil
}
