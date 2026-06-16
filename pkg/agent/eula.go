package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/auth"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/version"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	eulaURL  = "https://cdn.teneo.pro/Teneo_Agent_SDK_End_User_License_Agreement_(EULA)_v1_1_0.pdf"
	rulesURL = "https://cdn.teneo.pro/Teneo_Agent_SDK_Public_Deployment_Rules_v1_0_0.pdf"
)

// eulaStatusResponse matches the server's EULAStatusResponse (handlers/eula.go)
type eulaStatusResponse struct {
	HasAccepted        bool         `json:"has_accepted"`
	CurrentEULA        *eulaDetails `json:"current_eula,omitempty"`
	RequiresAcceptance bool         `json:"requires_acceptance"`
}

// eulaDetails matches the server's db.EULAAgreement (no JSON tags = PascalCase)
type eulaDetails struct {
	Version     string `json:"Version"`
	ContentHash string `json:"ContentHash"`
	Title       string `json:"Title"`
}

// eulaAcceptRequest matches the server's AcceptEULARequest
type eulaAcceptRequest struct {
	WalletAddress       string         `json:"wallet_address"`
	EulaVersion         string         `json:"eula_version"`
	AcceptanceSignature string         `json:"acceptance_signature"`
	AcceptanceHash      string         `json:"acceptance_hash"`
	DeveloperCountry    string         `json:"developer_country"`
	AuxData             map[string]any `json:"aux_data,omitempty"`
}

// printEULALinks prints the EULA and deployment rules links at startup
func printEULALinks() {
	log.Println("========================================")
	log.Println("Teneo Agent SDK - Legal Documents")
	log.Printf("  EULA: %s", eulaURL)
	log.Printf("  Public Deployment Rules: %s", rulesURL)
	log.Println("========================================")
}

// checkAndAcceptEULA checks if EULA acceptance is required and auto-accepts it
func checkAndAcceptEULA(backendURL, privateKeyHex string) error {
	// Create auth manager for signing
	authManager, err := auth.NewManager(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	walletAddress := authManager.GetAddress()

	// Step 1: Check EULA status
	statusURL := fmt.Sprintf("%s/api/eula/status?wallet=%s", backendURL, walletAddress)
	resp, err := http.Get(statusURL)
	if err != nil {
		return fmt.Errorf("failed to check EULA status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read EULA status response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("EULA status check failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var status eulaStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("failed to parse EULA status: %w", err)
	}

	// Already accepted or no EULA required
	if !status.RequiresAcceptance {
		log.Printf("EULA already accepted or not required")
		return nil
	}

	// Need to accept — extract EULA details
	if status.CurrentEULA == nil {
		return fmt.Errorf("EULA acceptance required but no EULA details returned")
	}

	eulaVersion := status.CurrentEULA.Version
	contentHash := status.CurrentEULA.ContentHash

	log.Printf("EULA acceptance required (version: %s), auto-accepting...", eulaVersion)

	// Step 2: Build and sign the acceptance message
	currentHour := time.Now().Unix() / 3600
	message := fmt.Sprintf("I accept the Teneo End User License Agreement (EULA)\nVersion: %s\nContent Hash: %s\nTimestamp: %d",
		eulaVersion, contentHash, currentHour)

	signature, err := authManager.SignMessage(message)
	if err != nil {
		return fmt.Errorf("failed to sign EULA acceptance: %w", err)
	}

	// Compute acceptance_hash (keccak256 of the Ethereum-prefixed message)
	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	acceptanceHash := hexutil.Encode(crypto.Keccak256([]byte(prefixedMessage)))

	// Step 3: Submit acceptance
	acceptReq := eulaAcceptRequest{
		WalletAddress:       walletAddress,
		EulaVersion:         eulaVersion,
		AcceptanceSignature: signature,
		AcceptanceHash:      acceptanceHash,
		DeveloperCountry:    "US",
		AuxData: map[string]any{
			"accepted_via": "sdk",
			"sdk_version":  version.Version(),
		},
	}

	reqBody, err := json.Marshal(acceptReq)
	if err != nil {
		return fmt.Errorf("failed to marshal EULA accept request: %w", err)
	}

	acceptURL := fmt.Sprintf("%s/api/eula/accept", backendURL)
	acceptResp, err := http.Post(acceptURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to submit EULA acceptance: %w", err)
	}
	defer acceptResp.Body.Close()

	acceptBody, err := io.ReadAll(acceptResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read EULA accept response: %w", err)
	}

	if acceptResp.StatusCode != http.StatusOK {
		return fmt.Errorf("EULA acceptance failed (HTTP %d): %s", acceptResp.StatusCode, string(acceptBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(acceptBody, &result); err != nil {
		return fmt.Errorf("failed to parse EULA accept response: %w", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		return fmt.Errorf("EULA acceptance was not successful: %s", string(acceptBody))
	}

	log.Printf("EULA v%s accepted successfully", eulaVersion)
	return nil
}
