package nft

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// testPrivateKey is a throwaway key for testing — never use in production
const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewNFTMinter(t *testing.T) {
	minter, err := NewNFTMinter(testPrivateKey)
	if err != nil {
		t.Fatalf("NewNFTMinter failed: %v", err)
	}
	if minter.backendURL != defaultBackendURL {
		t.Errorf("expected backendURL=%q, got %q", defaultBackendURL, minter.backendURL)
	}
	if minter.client != nil {
		t.Error("expected nil ethclient for gasless minter")
	}
	if minter.privateKey == nil {
		t.Error("expected privateKey to be set")
	}
	if minter.httpClient == nil {
		t.Error("expected httpClient to be set")
	}
	if minter.httpClient.Timeout != 120*time.Second {
		t.Errorf("expected 120s timeout, got %v", minter.httpClient.Timeout)
	}

	// Verify address is derived correctly
	expectedKey, _ := crypto.HexToECDSA(testPrivateKey)
	expectedAddr := crypto.PubkeyToAddress(expectedKey.PublicKey)
	if minter.address != expectedAddr {
		t.Errorf("expected address %s, got %s", expectedAddr.Hex(), minter.address.Hex())
	}
}

func TestNewNFTMinter_InvalidKey(t *testing.T) {
	_, err := NewNFTMinter("not-a-valid-key")
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
	if !strings.Contains(err.Error(), "invalid private key") {
		t.Errorf("expected 'invalid private key' error, got: %v", err)
	}
}

func TestNewNFTMinter_EmptyKey(t *testing.T) {
	_, err := NewNFTMinter("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestNewNFTMinter_With0xPrefix(t *testing.T) {
	minter, err := NewNFTMinter("0x" + testPrivateKey)
	if err != nil {
		t.Fatalf("NewNFTMinter with 0x prefix failed: %v", err)
	}
	if minter.backendURL != defaultBackendURL {
		t.Errorf("expected backendURL=%q, got %q", defaultBackendURL, minter.backendURL)
	}
}

func TestParsePayloadAndHash(t *testing.T) {
	minter, _ := NewNFTMinter(testPrivateKey)

	metadata := `{
		"name": "Test Agent",
		"agent_id": "test-agent",
		"description": "A test agent for unit testing",
		"agent_type": "command",
		"capabilities": [{"name": "test_cap", "description": "test"}],
		"commands": [
			{"trigger": "ping", "description": "Ping test", "pricePerUnit": 0, "priceType": "task-transaction", "taskUnit": "per-query"},
			{"trigger": "analyze", "description": "Analyze test", "pricePerUnit": 0.05, "priceType": "task-transaction", "taskUnit": "per-query"}
		],
		"nlp_fallback": false,
		"categories": ["Utilities"],
		"metadata_version": "2.4.0"
	}`

	config, canonicalJSON, configHash, err := minter.parsePayloadAndHash([]byte(metadata))
	if err != nil {
		t.Fatalf("parsePayloadAndHash failed: %v", err)
	}

	// Verify parsed fields
	if config.AgentID != "test-agent" {
		t.Errorf("expected agent_id=test-agent, got %s", config.AgentID)
	}
	if config.Name != "Test Agent" {
		t.Errorf("expected name=Test Agent, got %s", config.Name)
	}
	if config.AgentType != "command" {
		t.Errorf("expected agent_type=command, got %s", config.AgentType)
	}
	if config.NlpFallback != false {
		t.Error("expected nlp_fallback=false")
	}

	// Verify canonical JSON is valid
	if !json.Valid(canonicalJSON) {
		t.Error("canonical JSON is not valid")
	}

	// Verify config hash is a 64-char hex string
	if len(configHash) != 64 {
		t.Errorf("expected 64-char hash, got %d chars: %s", len(configHash), configHash)
	}

	// Verify hash is deterministic — same input produces same output
	_, _, configHash2, err := minter.parsePayloadAndHash([]byte(metadata))
	if err != nil {
		t.Fatalf("second parsePayloadAndHash failed: %v", err)
	}
	if configHash != configHash2 {
		t.Errorf("hash not deterministic:\n  hash1: %s\n  hash2: %s", configHash, configHash2)
	}
}

func TestParsePayloadAndHash_MissingFields(t *testing.T) {
	minter, _ := NewNFTMinter(testPrivateKey)

	tests := []struct {
		name string
		json string
	}{
		{"missing agent_id", `{"name":"A","description":"desc enough","agent_type":"command","capabilities":[{"name":"x"}],"categories":["Utilities"]}`},
		{"missing name", `{"agent_id":"a","description":"desc enough","agent_type":"command","capabilities":[{"name":"x"}],"categories":["Utilities"]}`},
		{"missing description", `{"agent_id":"a","name":"A","agent_type":"command","capabilities":[{"name":"x"}],"categories":["Utilities"]}`},
		{"missing agent_type", `{"agent_id":"a","name":"A","description":"desc enough","capabilities":[{"name":"x"}],"categories":["Utilities"]}`},
		{"missing capabilities", `{"agent_id":"a","name":"A","description":"desc enough","agent_type":"command","categories":["Utilities"]}`},
		{"missing categories", `{"agent_id":"a","name":"A","description":"desc enough","agent_type":"command","capabilities":[{"name":"x"}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := minter.parsePayloadAndHash([]byte(tt.json))
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestHashDeterminism(t *testing.T) {
	minter, _ := NewNFTMinter(testPrivateKey)

	// Same metadata, different whitespace/ordering — hash should be stable
	metadata1 := `{"name":"Agent","agent_id":"a","description":"desc test here","agent_type":"command","capabilities":[{"name":"b"},{"name":"a"}],"nlp_fallback":false,"categories":["Z","A"],"commands":[{"trigger":"z","pricePerUnit":0.1},{"trigger":"a","pricePerUnit":0.2}]}`
	metadata2 := `{  "name" : "Agent" , "agent_id" : "a" , "description" : "desc test here" , "agent_type" : "command" , "capabilities" : [ { "name" : "b" } , { "name" : "a" } ] , "nlp_fallback" : false , "categories" : [ "Z" , "A" ] , "commands" : [ { "trigger" : "z" , "pricePerUnit" : 0.1 } , { "trigger" : "a" , "pricePerUnit" : 0.2 } ] }`

	_, _, hash1, err := minter.parsePayloadAndHash([]byte(metadata1))
	if err != nil {
		t.Fatalf("hash1 failed: %v", err)
	}
	_, _, hash2, err := minter.parsePayloadAndHash([]byte(metadata2))
	if err != nil {
		t.Fatalf("hash2 failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hashes should be identical regardless of whitespace:\n  hash1: %s\n  hash2: %s", hash1, hash2)
	}
}

func TestGaslessDeployFlow(t *testing.T) {
	// Mock server that simulates the full gasless minting flow
	callCount := map[string]int{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount[r.URL.Path]++

		switch r.URL.Path {
		case "/api/sdk/agent/sync":
			// Step 1: Agent not found → MINT_REQUIRED
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "MINT_REQUIRED",
				"message": "Agent not found, minting required",
			})

		case "/api/sdk/auth/challenge":
			// Step 2a: Return challenge
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"challenge": "test-challenge-12345",
			})

		case "/api/sdk/auth/verify":
			// Step 2b: Verify signature, return session token
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"session_token": "test-session-token",
			})

		case "/api/sdk/agent/deploy":
			// Step 3: Gasless mint — server returns token directly
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			// Verify required fields are present
			requiredFields := []string{"agent_id", "agent_name", "description", "agent_type", "config_hash"}
			for _, f := range requiredFields {
				if _, ok := req[f]; !ok {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("missing %s", f)})
					return
				}
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"contract_address": "0xd8493cc411D5d0daa58dd7d6C0A22baEA9fbb3e5",
				"chain_id":         "3338",
				"agent_id":         req["agent_id"],
				"config_hash":      req["config_hash"],
				"token_id":         42,
				"tx_hash":          "0xabc123def456",
				"metadata_uri":     "ipfs://QmTest123",
				"gasless":          true,
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	// Create minter pointing to mock server
	minter, err := NewNFTMinterWithConfig(server.URL, "", testPrivateKey)
	if err != nil {
		t.Fatalf("NewNFTMinter failed: %v", err)
	}

	metadata := `{
		"name": "Test Gasless Agent",
		"agent_id": "test-gasless-agent",
		"description": "Agent for testing gasless minting flow",
		"agent_type": "command",
		"capabilities": [{"name": "test", "description": "test capability"}],
		"commands": [{"trigger": "ping", "description": "Ping", "pricePerUnit": 0, "priceType": "task-transaction", "taskUnit": "per-query"}],
		"nlp_fallback": false,
		"categories": ["Utilities"],
		"metadata_version": "2.4.0"
	}`

	result, err := minter.MintOrResumeFromJSON([]byte(metadata))
	if err != nil {
		t.Fatalf("MintOrResumeFromJSON failed: %v", err)
	}

	// Verify result
	if result.TokenID != 42 {
		t.Errorf("expected token_id=42, got %d", result.TokenID)
	}
	if result.TxHash != "0xabc123def456" {
		t.Errorf("expected tx_hash=0xabc123def456, got %s", result.TxHash)
	}
	if result.MetadataURI != "ipfs://QmTest123" {
		t.Errorf("expected metadata_uri=ipfs://QmTest123, got %s", result.MetadataURI)
	}

	// Verify correct endpoints were called
	// sync calls challenge once (for signed sync request), then authenticateSDKSession calls it again
	if callCount["/api/sdk/agent/sync"] != 1 {
		t.Errorf("expected 1 sync call, got %d", callCount["/api/sdk/agent/sync"])
	}
	if callCount["/api/sdk/auth/challenge"] != 2 {
		t.Errorf("expected 2 challenge calls (1 for sync, 1 for session), got %d", callCount["/api/sdk/auth/challenge"])
	}
	if callCount["/api/sdk/auth/verify"] != 1 {
		t.Errorf("expected 1 verify call, got %d", callCount["/api/sdk/auth/verify"])
	}
	if callCount["/api/sdk/agent/deploy"] != 1 {
		t.Errorf("expected 1 deploy call, got %d", callCount["/api/sdk/agent/deploy"])
	}
	// confirm-mint should NOT be called for gasless
	if callCount["/api/sdk/agent/confirm-mint"] != 0 {
		t.Errorf("confirm-mint should not be called for gasless, got %d calls", callCount["/api/sdk/agent/confirm-mint"])
	}
}

func TestSyncedFlow(t *testing.T) {
	// Mock server: agent already minted and synced
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/sdk/auth/challenge":
			// syncAgentState needs a challenge to sign the sync request
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"challenge": "test-challenge-synced",
			})
		case "/api/sdk/agent/sync":
			tokenID := int64(99)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "SYNCED",
				"token_id": tokenID,
				"agent_id": "existing-agent",
				"message":  "Agent synced successfully",
			})
		default:
			t.Errorf("unexpected call to %s — synced flow should only call challenge + sync", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	minter, _ := NewNFTMinterWithConfig(server.URL, "", testPrivateKey)

	metadata := `{
		"name": "Existing Agent",
		"agent_id": "existing-agent",
		"description": "Already minted agent for testing",
		"agent_type": "command",
		"capabilities": [{"name": "test", "description": "test"}],
		"nlp_fallback": false,
		"categories": ["Utilities"]
	}`

	result, err := minter.MintOrResumeFromJSON([]byte(metadata))
	if err != nil {
		t.Fatalf("MintOrResumeFromJSON failed: %v", err)
	}

	if result.TokenID != 99 {
		t.Errorf("expected token_id=99, got %d", result.TokenID)
	}
	if result.Status != "SYNCED" {
		t.Errorf("expected status=SYNCED, got %s", result.Status)
	}
}

func TestGaslessDeployFlow_InvalidTokenID(t *testing.T) {
	// Server returns gasless=true but token_id=0 — must error, not silently fall through
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/sdk/agent/sync":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "MINT_REQUIRED",
				"message": "Agent not found",
			})
		case "/api/sdk/auth/challenge":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"challenge": "test"})
		case "/api/sdk/auth/verify":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"session_token": "tok"})
		case "/api/sdk/agent/deploy":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"contract_address": "0xd8493cc411D5d0daa58dd7d6C0A22baEA9fbb3e5",
				"chain_id":         "3338",
				"agent_id":         "test",
				"config_hash":      "abc",
				"token_id":         0,
				"gasless":          true,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	minter, _ := NewNFTMinterWithConfig(server.URL, "", testPrivateKey)
	metadata := `{"name":"A","agent_id":"a","description":"test agent desc","agent_type":"command","capabilities":[{"name":"x"}],"nlp_fallback":false,"categories":["Utilities"]}`

	_, err := minter.MintOrResumeFromJSON([]byte(metadata))
	if err == nil {
		t.Fatal("expected error for gasless=true with token_id=0")
	}
	if !strings.Contains(err.Error(), "double-mint") {
		t.Errorf("expected double-mint prevention error, got: %v", err)
	}
}

func TestNonGaslessDeployRejected(t *testing.T) {
	// Server returns gasless=false (non-gasless) — must error, not attempt client-side minting
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/sdk/agent/sync":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "MINT_REQUIRED",
				"message": "Agent not found",
			})
		case "/api/sdk/auth/challenge":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"challenge": "test-challenge"})
		case "/api/sdk/auth/verify":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"session_token": "tok"})
		case "/api/sdk/agent/deploy":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"gasless":          false,
				"signature":        "0xsig123",
				"contract_address": "0xContract",
				"chain_id":         "421614",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	minter, _ := NewNFTMinterWithConfig(server.URL, "", testPrivateKey)
	metadata := `{"name":"A","agent_id":"a","description":"test agent desc","agent_type":"command","capabilities":[{"name":"x"}],"nlp_fallback":false,"categories":["Utilities"]}`

	_, err := minter.MintOrResumeFromJSON([]byte(metadata))
	if err == nil {
		t.Fatal("expected error for non-gasless deploy response")
	}
	if !strings.Contains(err.Error(), "gasless") {
		t.Errorf("expected gasless-related error, got: %v", err)
	}
}

func TestMint_MissingEnvVar(t *testing.T) {
	t.Setenv("PRIVATE_KEY", "")
	_, err := Mint("nonexistent.json")
	if err == nil {
		t.Fatal("expected error when PRIVATE_KEY is not set")
	}
	if !strings.Contains(err.Error(), "PRIVATE_KEY") {
		t.Errorf("expected PRIVATE_KEY error, got: %v", err)
	}
}

func TestMintWithKey_InvalidFile(t *testing.T) {
	_, err := MintWithKey(testPrivateKey, "nonexistent-file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestMintWithKey_InvalidKey(t *testing.T) {
	_, err := MintWithKey("bad-key", "nonexistent.json")
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

// --- helpers ---

// testHashStartsWithV4 verifies the hash function uses v4 prefix
func testHashStartsWithV4(t *testing.T) {
	t.Helper()
	minter, _ := NewNFTMinter(testPrivateKey)
	metadata := `{
		"name": "V4 Test",
		"agent_id": "v4-test",
		"description": "Testing v4 hash",
		"agent_type": "command",
		"capabilities": [{"name": "cap1", "description": "desc1"}],
		"commands": [],
		"nlp_fallback": false,
		"categories": ["AI"],
		"metadata_version": "2.4.0"
	}`
	_, _, hash, err := minter.parsePayloadAndHash([]byte(metadata))
	if err != nil {
		t.Fatalf("parsePayloadAndHash failed: %v", err)
	}
	if len(hash) != 64 {
		t.Errorf("expected 64-char hash, got %d", len(hash))
	}
}
