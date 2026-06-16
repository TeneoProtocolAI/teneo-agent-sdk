package deploy

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWALClient_SaveAndLoad(t *testing.T) {
	// Create temp WAL directory
	tmpDir := t.TempDir()
	walClient := NewWALClientWithDir(tmpDir)

	entry := &WALEntry{
		AgentID:         "test-agent",
		Wallet:          "0x1234567890abcdef",
		State:           WALStateMinting,
		PendingTxHash:   "0xabcdef123456",
		ContractAddress: "0xcontract",
		ChainID:         "3338",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save
	err := walClient.Save(entry)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := walClient.Load("test-agent")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("Load() returned nil")
	}

	// Verify fields
	if loaded.AgentID != entry.AgentID {
		t.Errorf("AgentID = %v, want %v", loaded.AgentID, entry.AgentID)
	}
	if loaded.Wallet != entry.Wallet {
		t.Errorf("Wallet = %v, want %v", loaded.Wallet, entry.Wallet)
	}
	if loaded.State != entry.State {
		t.Errorf("State = %v, want %v", loaded.State, entry.State)
	}
	if loaded.PendingTxHash != entry.PendingTxHash {
		t.Errorf("PendingTxHash = %v, want %v", loaded.PendingTxHash, entry.PendingTxHash)
	}
}

func TestWALClient_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	walClient := NewWALClientWithDir(tmpDir)

	// Load non-existent
	loaded, err := walClient.Load("non-existent-agent")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded != nil {
		t.Error("Load() should return nil for non-existent entry")
	}
}

func TestWALClient_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	walClient := NewWALClientWithDir(tmpDir)

	entry := &WALEntry{
		AgentID:   "delete-test",
		Wallet:    "0x123",
		State:     WALStateMinting,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save
	if err := walClient.Save(entry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify exists
	if !walClient.Exists("delete-test") {
		t.Error("Entry should exist after save")
	}

	// Delete
	if err := walClient.Delete("delete-test"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	if walClient.Exists("delete-test") {
		t.Error("Entry should not exist after delete")
	}
}

func TestWALClient_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	walClient := NewWALClientWithDir(tmpDir)

	// Should not exist initially
	if walClient.Exists("test-exists") {
		t.Error("Entry should not exist initially")
	}

	// Create entry
	entry := &WALEntry{
		AgentID:   "test-exists",
		Wallet:    "0x123",
		State:     WALStateIdle,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := walClient.Save(entry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Should exist now
	if !walClient.Exists("test-exists") {
		t.Error("Entry should exist after save")
	}
}

func TestWALClient_List(t *testing.T) {
	tmpDir := t.TempDir()
	walClient := NewWALClientWithDir(tmpDir)

	// Create multiple entries
	entries := []string{"agent-1", "agent-2", "agent-3"}
	for _, agentID := range entries {
		entry := &WALEntry{
			AgentID:   agentID,
			Wallet:    "0x123",
			State:     WALStateMinting,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := walClient.Save(entry); err != nil {
			t.Fatalf("Save(%s) error = %v", agentID, err)
		}
	}

	// List all
	list, err := walClient.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != len(entries) {
		t.Errorf("List() returned %d entries, want %d", len(list), len(entries))
	}
}

func TestWALClient_CleanupOld(t *testing.T) {
	tmpDir := t.TempDir()
	walClient := NewWALClientWithDir(tmpDir)

	// Create old entry by directly writing JSON file (bypassing Save which updates timestamp)
	oldEntry := &WALEntry{
		AgentID:   "old-agent",
		Wallet:    "0x123",
		State:     WALStateMinting,
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-2 * time.Hour), // This will be preserved
	}

	// Write directly to file to preserve the old UpdatedAt
	oldEntryJSON := `{
		"agent_id": "old-agent",
		"wallet": "0x123",
		"state": "MINTING",
		"created_at": "` + oldEntry.CreatedAt.Format(time.RFC3339) + `",
		"updated_at": "` + oldEntry.UpdatedAt.Format(time.RFC3339) + `"
	}`
	walPath := filepath.Join(tmpDir, "old-agent.json")
	if err := os.WriteFile(walPath, []byte(oldEntryJSON), 0644); err != nil {
		t.Fatalf("Failed to write old entry: %v", err)
	}

	// Verify old entry was written correctly
	loaded, err := walClient.Load("old-agent")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("Old entry should exist")
	}

	// Create new entry (this one uses Save, so UpdatedAt will be now)
	newEntry := &WALEntry{
		AgentID:   "new-agent",
		Wallet:    "0x456",
		State:     WALStateMinting,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := walClient.Save(newEntry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Cleanup entries older than 1 hour
	deleted, err := walClient.CleanupOld(1 * time.Hour)
	if err != nil {
		t.Fatalf("CleanupOld() error = %v", err)
	}

	if deleted != 1 {
		t.Errorf("CleanupOld() deleted %d entries, want 1", deleted)
	}

	// Verify old is gone, new remains
	if walClient.Exists("old-agent") {
		t.Error("Old agent should have been cleaned up")
	}
	if !walClient.Exists("new-agent") {
		t.Error("New agent should still exist")
	}
}

func TestWALClient_TokenID(t *testing.T) {
	tmpDir := t.TempDir()
	walClient := NewWALClientWithDir(tmpDir)

	tokenID := uint64(12345)
	entry := &WALEntry{
		AgentID:        "token-test",
		Wallet:         "0x123",
		State:          WALStateConfirming,
		PendingTokenID: &tokenID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := walClient.Save(entry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := walClient.Load("token-test")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.PendingTokenID == nil {
		t.Fatal("PendingTokenID should not be nil")
	}
	if *loaded.PendingTokenID != tokenID {
		t.Errorf("PendingTokenID = %d, want %d", *loaded.PendingTokenID, tokenID)
	}
}

func TestWALStates(t *testing.T) {
	// Verify constants
	if WALStateIdle != "IDLE" {
		t.Errorf("WALStateIdle = %q, want %q", WALStateIdle, "IDLE")
	}
	if WALStateMinting != "MINTING" {
		t.Errorf("WALStateMinting = %q, want %q", WALStateMinting, "MINTING")
	}
	if WALStateConfirming != "CONFIRMING" {
		t.Errorf("WALStateConfirming = %q, want %q", WALStateConfirming, "CONFIRMING")
	}
}
