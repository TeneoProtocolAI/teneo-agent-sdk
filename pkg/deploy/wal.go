package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WAL State constants
const (
	WALStateIdle       = "IDLE"
	WALStateMinting    = "MINTING"
	WALStateConfirming = "CONFIRMING"
)

// WALEntry represents a Write-Ahead Log entry for crash recovery
type WALEntry struct {
	AgentID         string    `json:"agent_id"`
	Wallet          string    `json:"wallet"`
	State           string    `json:"state"` // IDLE, MINTING, CONFIRMING
	PendingTxHash   string    `json:"pending_tx_hash,omitempty"`
	PendingTokenID  *uint64   `json:"pending_token_id,omitempty"`
	ContractAddress string    `json:"contract_address,omitempty"`
	ChainID         string    `json:"chain_id,omitempty"`
	RPCURL          string    `json:"rpc_url,omitempty"`
	Signature       string    `json:"signature,omitempty"`
	ConfigHash      string    `json:"config_hash,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// WALClient handles Write-Ahead Log operations
type WALClient struct {
	walDir string
}

// NewWALClient creates a new WAL client
func NewWALClient() *WALClient {
	// Default WAL directory: ~/.teneo/wal/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	walDir := filepath.Join(homeDir, ".teneo", "wal")

	return &WALClient{
		walDir: walDir,
	}
}

// NewWALClientWithDir creates a WAL client with custom directory
func NewWALClientWithDir(walDir string) *WALClient {
	return &WALClient{
		walDir: walDir,
	}
}

// getPath returns the WAL file path for an agent
func (w *WALClient) getPath(agentID string) string {
	return filepath.Join(w.walDir, agentID+".json")
}

// ensureDir ensures the WAL directory exists
func (w *WALClient) ensureDir() error {
	return os.MkdirAll(w.walDir, 0700)
}

// Load loads a WAL entry for an agent
func (w *WALClient) Load(agentID string) (*WALEntry, error) {
	path := w.getPath(agentID)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No WAL exists
		}
		return nil, fmt.Errorf("failed to read WAL file: %w", err)
	}

	var entry WALEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse WAL file: %w", err)
	}

	return &entry, nil
}

// Save saves a WAL entry
func (w *WALClient) Save(entry *WALEntry) error {
	if err := w.ensureDir(); err != nil {
		return fmt.Errorf("failed to create WAL directory: %w", err)
	}

	entry.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal WAL entry: %w", err)
	}

	path := w.getPath(entry.AgentID)

	// Write atomically using temp file + rename
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write WAL temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename WAL temp file: %w", err)
	}

	return nil
}

// Delete removes a WAL entry
func (w *WALClient) Delete(agentID string) error {
	path := w.getPath(agentID)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete WAL file: %w", err)
	}

	return nil
}

// Exists checks if a WAL entry exists
func (w *WALClient) Exists(agentID string) bool {
	path := w.getPath(agentID)
	_, err := os.Stat(path)
	return err == nil
}

// List lists all WAL entries
func (w *WALClient) List() ([]*WALEntry, error) {
	if err := w.ensureDir(); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	entries, err := os.ReadDir(w.walDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAL directory: %w", err)
	}

	var walEntries []*WALEntry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		agentID := entry.Name()[:len(entry.Name())-5] // Remove .json
		walEntry, err := w.Load(agentID)
		if err != nil {
			continue // Skip invalid entries
		}

		if walEntry != nil {
			walEntries = append(walEntries, walEntry)
		}
	}

	return walEntries, nil
}

// CleanupOld removes WAL entries older than the specified duration
func (w *WALClient) CleanupOld(maxAge time.Duration) (int, error) {
	entries, err := w.List()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	deleted := 0

	for _, entry := range entries {
		if now.Sub(entry.UpdatedAt) > maxAge {
			if err := w.Delete(entry.AgentID); err == nil {
				deleted++
			}
		}
	}

	return deleted, nil
}
