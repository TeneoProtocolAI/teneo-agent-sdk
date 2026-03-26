package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SubmitForReview submits an agent for public visibility review on the Teneo network.
//
// Agents go through a review process before becoming publicly visible:
//
//	private → in_review → public (approved) or declined
//
// This is a standalone utility that can be called from any context — no running agent required.
// The agent must have been deployed, connected at least once, and be currently online.
//
// Parameters:
//   - backendURL: The Teneo backend URL (e.g. "https://backend.developer.chatroom.teneo-protocol.ai")
//   - agentID: The agent's ID as registered in the database
//   - creatorWallet: The Ethereum wallet address that owns the agent's NFT
//   - tokenID: The NFT token ID for on-chain ownership verification
//
// Example:
//
//	err := agent.SubmitForReview(
//	    "https://backend.developer.chatroom.teneo-protocol.ai",
//	    "my-agent-id",
//	    "0xYourWalletAddress",
//	    42,
//	)
//
// HTTP API equivalent (for non-Go clients):
//
//	POST {backendURL}/api/agents/{agent-id}/submit-for-review
//	Content-Type: application/json
//
//	{
//	    "creator_wallet": "0xYourWalletAddress",
//	    "token_id": 42
//	}
func SubmitForReview(backendURL, agentID, creatorWallet string, tokenID uint64) error {
	backendURL = strings.TrimRight(backendURL, "/")

	reqBody, err := json.Marshal(map[string]interface{}{
		"creator_wallet": creatorWallet,
		"token_id":       tokenID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/agents/%s/submit-for-review", backendURL, agentID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("submit for review failed: %s", errResp.Error)
		}
		return fmt.Errorf("submit for review failed with status %d", resp.StatusCode)
	}

	return nil
}

// WithdrawPublic withdraws a public agent back to private visibility on the Teneo network.
//
// Only agents with review status "public" can be withdrawn. After withdrawal the agent
// returns to "private" status and must be re-submitted for review to become public again.
//
// This is a standalone utility that can be called from any context — no running agent required.
//
// Parameters:
//   - backendURL: The Teneo backend URL (e.g. "https://backend.developer.chatroom.teneo-protocol.ai")
//   - agentID: The agent's ID as registered in the database
//   - creatorWallet: The Ethereum wallet address that owns the agent's NFT
//   - tokenID: The NFT token ID for on-chain ownership verification
//
// Example:
//
//	err := agent.WithdrawPublic(
//	    "https://backend.developer.chatroom.teneo-protocol.ai",
//	    "my-agent-id",
//	    "0xYourWalletAddress",
//	    42,
//	)
//
// HTTP API equivalent (for non-Go clients):
//
//	POST {backendURL}/api/agents/{agent-id}/withdraw-public
//	Content-Type: application/json
//
//	{
//	    "creator_wallet": "0xYourWalletAddress",
//	    "token_id": 42
//	}
func WithdrawPublic(backendURL, agentID, creatorWallet string, tokenID uint64) error {
	backendURL = strings.TrimRight(backendURL, "/")

	reqBody, err := json.Marshal(map[string]interface{}{
		"creator_wallet": creatorWallet,
		"token_id":       tokenID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/agents/%s/withdraw-public", backendURL, agentID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("withdraw public failed: %s", errResp.Error)
		}
		return fmt.Errorf("withdraw public failed with status %d", resp.StatusCode)
	}

	return nil
}
