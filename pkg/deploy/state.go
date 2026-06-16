package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DeployStatus represents the current status of a deployment
type DeployStatus string

const (
	StatusPending   DeployStatus = "pending"   // Metadata uploaded, awaiting mint
	StatusMinted    DeployStatus = "minted"    // NFT minted on-chain, awaiting confirmation
	StatusConfirmed DeployStatus = "confirmed" // Agent saved to database
)

// DeployState tracks the progress of a deployment operation
type DeployState struct {
	AgentID         string       `json:"agent_id"`
	AgentName       string       `json:"agent_name"`
	WalletAddress   string       `json:"wallet_address"`
	TokenID         uint64       `json:"token_id,omitempty"`
	TxHash          string       `json:"tx_hash,omitempty"`
	ContractAddress string       `json:"contract_address,omitempty"`
	RPCURL          string       `json:"rpc_url,omitempty"`
	ConfigHash      string       `json:"config_hash,omitempty"`
	Status          DeployStatus `json:"status"`
	SessionToken    string       `json:"session_token,omitempty"`
	SessionExpiry   int64        `json:"session_expiry,omitempty"`
	Nonce           uint64       `json:"nonce,omitempty"`
	ChainID         string       `json:"chain_id,omitempty"`
	Signature       string       `json:"signature,omitempty"`
	UpdatedAt       time.Time    `json:"updated_at"`
	CreatedAt       time.Time    `json:"created_at"`
	Error           string       `json:"error,omitempty"`
}

// StateManager handles persistent state storage for deploy operations
type StateManager struct {
	filePath string
	mu       sync.RWMutex
}

// NewStateManager creates a new state manager with the specified file path
func NewStateManager(filePath string) *StateManager {
	// Expand home directory if needed
	if filePath == "" {
		filePath = ".teneo-deploy-state.json"
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0700)
	}

	return &StateManager{
		filePath: filePath,
	}
}

// Load reads the state from disk
func (sm *StateManager) Load() (*DeployState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data, err := os.ReadFile(sm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No state file exists
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state DeployState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// Save writes the state to disk atomically
func (sm *StateManager) Save(state *DeployState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state.UpdatedAt = time.Now().UTC()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temp file first for atomic operation
	tempPath := sm.filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp state file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, sm.filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file on error
		return fmt.Errorf("failed to save state file: %w", err)
	}

	return nil
}

// Delete removes the state file
func (sm *StateManager) Delete() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if err := os.Remove(sm.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete state file: %w", err)
	}
	return nil
}

// UpdateStatus atomically updates the status and saves
func (sm *StateManager) UpdateStatus(status DeployStatus) error {
	state, err := sm.Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no state to update")
	}
	state.Status = status
	return sm.Save(state)
}

// SetMinted updates state after successful on-chain mint
func (sm *StateManager) SetMinted(tokenID uint64, txHash string) error {
	state, err := sm.Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no state to update")
	}
	state.TokenID = tokenID
	state.TxHash = txHash
	state.Status = StatusMinted
	return sm.Save(state)
}

// SetConfirmed updates state after successful database confirmation
func (sm *StateManager) SetConfirmed() error {
	state, err := sm.Load()
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no state to update")
	}
	state.Status = StatusConfirmed
	return sm.Save(state)
}

// CreateInitialState creates a new state for a fresh deployment
func (sm *StateManager) CreateInitialState(agentID, agentName, walletAddress string) (*DeployState, error) {
	now := time.Now().UTC()
	state := &DeployState{
		AgentID:       agentID,
		AgentName:     agentName,
		WalletAddress: walletAddress,
		Status:        StatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := sm.Save(state); err != nil {
		return nil, err
	}
	return state, nil
}

// IsSessionValid checks if the stored session token is still valid
func (state *DeployState) IsSessionValid() bool {
	if state.SessionToken == "" {
		return false
	}
	return time.Now().Unix() < state.SessionExpiry
}

// GetFilePath returns the path to the state file
func (sm *StateManager) GetFilePath() string {
	return sm.filePath
}
