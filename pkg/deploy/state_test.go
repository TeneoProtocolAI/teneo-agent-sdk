package deploy

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateManager_CreateAndLoad(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "test-state.json")

	sm := NewStateManager(statePath)

	// Create initial state
	state, err := sm.CreateInitialState("test-agent", "Test Agent", "0x742d35Cc6634C0532925a3b844Bc9e7595f2b21D")
	if err != nil {
		t.Fatalf("failed to create initial state: %v", err)
	}

	if state.AgentID != "test-agent" {
		t.Errorf("expected agent_id 'test-agent', got '%s'", state.AgentID)
	}
	if state.Status != StatusPending {
		t.Errorf("expected status 'pending', got '%s'", state.Status)
	}

	// Load state
	loaded, err := sm.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loaded.AgentID != state.AgentID {
		t.Errorf("loaded agent_id mismatch: got '%s', want '%s'", loaded.AgentID, state.AgentID)
	}
	if loaded.Status != state.Status {
		t.Errorf("loaded status mismatch: got '%s', want '%s'", loaded.Status, state.Status)
	}
}

func TestStateManager_LoadNonexistent(t *testing.T) {
	sm := NewStateManager("/nonexistent/path/state.json")

	state, err := sm.Load()
	if err != nil {
		t.Errorf("expected nil error for nonexistent file, got %v", err)
	}
	if state != nil {
		t.Error("expected nil state for nonexistent file")
	}
}

func TestStateManager_SetMinted(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "test-state.json")

	sm := NewStateManager(statePath)

	// Create initial state
	_, err := sm.CreateInitialState("test-agent", "Test Agent", "0x742d35Cc6634C0532925a3b844Bc9e7595f2b21D")
	if err != nil {
		t.Fatalf("failed to create initial state: %v", err)
	}

	// Set minted
	if err := sm.SetMinted(12345, "0xabcdef"); err != nil {
		t.Fatalf("failed to set minted: %v", err)
	}

	// Load and verify
	loaded, err := sm.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loaded.TokenID != 12345 {
		t.Errorf("expected token_id 12345, got %d", loaded.TokenID)
	}
	if loaded.TxHash != "0xabcdef" {
		t.Errorf("expected tx_hash '0xabcdef', got '%s'", loaded.TxHash)
	}
	if loaded.Status != StatusMinted {
		t.Errorf("expected status 'minted', got '%s'", loaded.Status)
	}
}

func TestStateManager_SetConfirmed(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "test-state.json")

	sm := NewStateManager(statePath)

	// Create and mint
	sm.CreateInitialState("test-agent", "Test Agent", "0x742d35Cc6634C0532925a3b844Bc9e7595f2b21D")
	sm.SetMinted(12345, "0xabcdef")

	// Set confirmed
	if err := sm.SetConfirmed(); err != nil {
		t.Fatalf("failed to set confirmed: %v", err)
	}

	// Load and verify
	loaded, err := sm.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loaded.Status != StatusConfirmed {
		t.Errorf("expected status 'confirmed', got '%s'", loaded.Status)
	}
}

func TestStateManager_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "test-state.json")

	sm := NewStateManager(statePath)

	// Create state
	sm.CreateInitialState("test-agent", "Test Agent", "0x742d35Cc6634C0532925a3b844Bc9e7595f2b21D")

	// Verify file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("state file should exist")
	}

	// Delete
	if err := sm.Delete(); err != nil {
		t.Fatalf("failed to delete state: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("state file should be deleted")
	}
}

func TestDeployState_IsSessionValid(t *testing.T) {
	tests := []struct {
		name     string
		state    DeployState
		expected bool
	}{
		{
			name:     "no token",
			state:    DeployState{},
			expected: false,
		},
		{
			name: "expired token",
			state: DeployState{
				SessionToken:  "valid-token",
				SessionExpiry: time.Now().Add(-1 * time.Hour).Unix(),
			},
			expected: false,
		},
		{
			name: "valid token",
			state: DeployState{
				SessionToken:  "valid-token",
				SessionExpiry: time.Now().Add(1 * time.Hour).Unix(),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsSessionValid(); got != tt.expected {
				t.Errorf("IsSessionValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateManager_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "test-state.json")

	sm := NewStateManager(statePath)

	// Create initial state
	state, _ := sm.CreateInitialState("test-agent", "Test Agent", "0x742d35Cc6634C0532925a3b844Bc9e7595f2b21D")

	// Verify temp file is cleaned up
	tmpPath := statePath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful write")
	}

	// Verify state was written correctly
	loaded, _ := sm.Load()
	if loaded.AgentID != state.AgentID {
		t.Error("state was not written correctly")
	}
}
